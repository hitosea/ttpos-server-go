package tasks

import (
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"ttpos-server-go/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DailySalesOutboundSummaryTask 每日销售出库汇总定时任务
// 每小时执行一次，检查门店是否到达营业结束时间，如果是则统计当天销售出库记录并写入汇总表
type DailySalesOutboundSummaryTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewDailySalesOutboundSummaryTask 创建每日销售出库汇总任务
func NewDailySalesOutboundSummaryTask(dbm *database.DBManager, cache cache.Cache) *DailySalesOutboundSummaryTask {
	return &DailySalesOutboundSummaryTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行定时任务
func (t *DailySalesOutboundSummaryTask) Execute() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("每日销售出库汇总定时任务,发生panic: %v", zap.Any("panic", r))
		}
	}()
	logger.Logger.Info("开始执行每日销售出库汇总定时任务")

	start := time.Now()
	// 分布式锁,避免多个节点同时执行
	lock.NewSystemLock().LockUuid(lock.DailySalesOutboundSummaryLock)
	defer lock.NewSystemLock().UnlockUuid(lock.DailySalesOutboundSummaryLock)
	spend := time.Since(start)
	logger.Logger.Info("分布式锁耗时: %v", zap.Duration("spend", spend))
	if spend > 1*time.Second {
		logger.Logger.Warn("其他节点已经处理该任务，本节点跳过")
		return
	}

	// 获取所有门店
	companies, err := t.getAllCompanies()
	if err != nil {
		logger.Logger.Error("获取门店列表失败: %v", zap.Error(err))
		return
	}

	logger.Logger.Info("找到 %d 个门店，开始检查营业结束时间", zap.Int("company_count", len(companies)))

	for _, company := range companies {
		if err := t.ProcessCompany(company); err != nil {
			logger.Logger.Error("处理门店 %s (%d) 失败: %v", zap.String("company_name", company.Name), zap.Uint64("company_uuid", company.Uuid), zap.Error(err))
			continue
		}
	}

	logger.Logger.Info("每日销售出库汇总定时任务执行完成")
}

// getAllCompanies 获取所有门店
func (t *DailySalesOutboundSummaryTask) getAllCompanies() ([]*model.Company, error) {
	var companies []*model.Company
	db := t.dbm.GetDB(0)

	// 查询所有未删除且启用的门店，包含设置信息
	err := db.Model(&model.Company{}).
		Joins("LEFT JOIN ttpos_company_setting ON ttpos_company.uuid = ttpos_company_setting.company_uuid").
		Where("ttpos_company.delete_time = 0 AND ttpos_company.status = 1").
		Preload("CompanySetting").
		Find(&companies).Error

	return companies, err
}

// processCompany 处理单个门店
func (t *DailySalesOutboundSummaryTask) ProcessCompany(company *model.Company) error {
	// 判断是否是erp商品,不是的不处理
	if !company.IsOpenErp() {
		return nil
	}

	// 获取门店营业时间
	openingHours, err := t.getOpeningHours(company.Uuid)
	if err != nil {
		return fmt.Errorf("获取门店营业时间失败: %w", err)
	}

	// 检查是否到达营业结束时间
	isBusinessEndTime, startTime, endTime := t.isBusinessEndTime(company, openingHours)
	if !isBusinessEndTime {
		logger.Logger.Info("门店 %s 未到达营业结束时间，跳过处理", zap.String("company_name", company.Name))
		return nil
	}

	// 判断该营业时段是否已经统计过
	year := time.Now().Format("20060102") // 20251023
	openingYearHours := fmt.Sprintf("%s %s", year, openingHours)
	db := t.dbm.GetDB(company.Uuid)
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)
	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(1), // 出库
		warehouseLogRepo.WhereScene(1),   // 销售出库
		func(db *gorm.DB) *gorm.DB {
			return db.Where("opening_hours = ?", openingYearHours)
		},
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return fmt.Errorf("获取营业时段记录失败: %w", err)
	}
	if len(existingLogs) > 0 {
		logger.Logger.Info(fmt.Sprintf("门店 %s 该营业时段%s已统计过，跳过处理", company.Name, openingYearHours))
		return nil
	}

	logger.Logger.Info("门店 %s 已到达营业结束时间，开始统计销售出库记录", zap.String("company_name", company.Name))

	// 统计当天销售出库记录
	outboundRecords, err := t.getDailySalesOutboundRecords(company.Uuid, startTime, endTime)
	if err != nil {
		return fmt.Errorf("获取销售出库记录失败: %w", err)
	}

	if len(outboundRecords) == 0 {
		logger.Logger.Info("门店 %s 今日无销售出库记录", zap.String("company_name", company.Name))
		return nil
	}

	// 生成汇总记录并写入数据库
	if err := t.saveOutboundSummaryRecords(company.Uuid, outboundRecords, openingYearHours); err != nil {
		return fmt.Errorf("保存出库汇总记录失败: %w", err)
	}

	logger.Logger.Info("门店 %s 销售出库汇总完成，共处理 %d 条记录", zap.String("company_name", company.Name), zap.Int("outbound_record_count", len(outboundRecords)))
	return nil
}

// getOpeningHours 获取门店营业时间
func (t *DailySalesOutboundSummaryTask) getOpeningHours(companyUuid uint64) (string, error) {
	setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	ctx := context.NewContext()
	ctx.SetCompanyUuid(companyUuid)
	businessSetting, err := setting.GetBusinessSetting(ctx)
	if err != nil {
		return "", err
	}
	if businessSetting.OpeningHours != "" {
		return businessSetting.OpeningHours, nil
	}

	// 返回默认值
	return "00:00-23:59", nil
}

// isBusinessEndTime 判断是否到达营业结束时间
func (t *DailySalesOutboundSummaryTask) isBusinessEndTime(company *model.Company, openingHours string) (bool, int64, int64) {
	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}
	logger.Logger.Info(fmt.Sprintf("时区: %s", timezone))
	timeUtil := utils.Timezone(timezone)

	// 获取营业时间的开始和结束时间戳
	startTime, endTime := timeUtil.OpeningHoursStartEndUnix(openingHours, utils.WithOpeningHoursType(1))
	logger.Logger.Info(fmt.Sprintf("openingHours: %s,营业开始时间: %s(%d),营业结束时间: %s(%d)", openingHours, timeUtil.FormatUnixTime(startTime, "2006-01-02 15:04:05"), startTime, timeUtil.FormatUnixTime(endTime, "2006-01-02 15:04:05"), endTime))
	now := timeUtil.Now().Unix()

	// 如果当前时间在营业结束时间之后，则认为已到达营业结束时间
	logger.Logger.Info(fmt.Sprintf("当前时间: %s(%d), 营业结束时间: %s(%d)", timeUtil.FormatUnixTime(now, "2006-01-02 15:04:05"), now, timeUtil.FormatUnixTime(endTime, "2006-01-02 15:04:05"), endTime))
	return now >= endTime, startTime, endTime
}

// getDailySalesOutboundRecords 获取当天销售出库记录
func (t *DailySalesOutboundSummaryTask) getDailySalesOutboundRecords(companyUuid uint64, startTime int64, endTime int64) ([]*OutboundRecord, error) {
	db := t.dbm.GetDB(companyUuid)

	// 使用 repository 方法查询出库单明细
	saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(db)
	saleOrderMaterials, err := saleOrderMaterialRepo.GetSaleOrderMaterialByCreateTimeBetween(
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	// 按仓库和物料分组汇总数量
	recordMap := make(map[string]*OutboundRecord)
	for _, item := range saleOrderMaterials {
		key := fmt.Sprintf("%d_%d", item.WarehouseUuid, item.MaterialUuid)
		if record, exists := recordMap[key]; exists {
			record.TotalNum += item.Num
		} else {
			materialName := ""
			materialBaseUnitUuid := uint64(0)
			materialBaseUnitName := ""

			// 从关联的物料信息中获取数据
			if item.Material != nil {
				materialName = item.Material.Name
				materialBaseUnitUuid = item.Material.UnitUuid
			}

			// 这里需要额外查询单位名称，因为关联查询比较复杂
			if materialBaseUnitUuid > 0 {
				var unitName string
				materialBaseUnitName = item.Material.Unit.Name
				materialBaseUnitName = unitName
			}

			recordMap[key] = &OutboundRecord{
				Uuid:                 item.Uuid,
				WarehouseUuid:        item.WarehouseUuid,
				MaterialUuid:         item.MaterialUuid,
				TotalNum:             item.Num,
				Valuation:            item.Material.Valuation,
				SupplierUuid:         item.Material.SupplierUuid,
				MaterialName:         materialName,
				MaterialBaseUnitUuid: materialBaseUnitUuid,
				MaterialBaseUnitName: materialBaseUnitName,
			}
		}
	}

	// 转换为切片返回
	var records []*OutboundRecord
	for _, record := range recordMap {
		records = append(records, record)
	}

	return records, nil
}

// OutboundRecord 出库记录汇总
type OutboundRecord struct {
	Uuid                 uint64  `json:"uuid"` // 出库记录ID
	WarehouseUuid        uint64  `json:"warehouse_uuid"`
	MaterialUuid         uint64  `json:"material_uuid"`
	TotalNum             float64 `json:"total_num"`
	Valuation            float64 `json:"valuation"` // 估值率
	SupplierUuid         uint64  `json:"supplier_uuid"`
	MaterialName         string  `json:"material_name"`
	MaterialBaseUnitUuid uint64  `json:"material_base_unit_uuid"`
	MaterialBaseUnitName string  `json:"material_base_unit_name"`
}

// saveOutboundSummaryRecords 保存出库汇总记录到ttpos_warehouse_in_out_log表
func (t *DailySalesOutboundSummaryTask) saveOutboundSummaryRecords(companyUuid uint64, records []*OutboundRecord, openingHours string) error {
	db := t.dbm.GetDB(companyUuid)

	// 生成出库单号
	orderNo, err := t.generateOrderNo(companyUuid)
	if err != nil {
		return fmt.Errorf("生成出库单号失败: %w", err)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 使用 repository 方法创建记录
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)

		uuids := make([]uint64, 0)
		for _, record := range records {
			uuids = append(uuids, record.Uuid)
			logRecord := &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeOut, // 出库
				Scene:                constant.WarehouseInOutLogSceneSale,  // 销售出库
				WarehouseUuid:        record.WarehouseUuid,
				MaterialUuid:         record.MaterialUuid,
				MaterialName:         record.MaterialName,
				MaterialBaseUnitUuid: record.MaterialBaseUnitUuid,
				MaterialBaseUnitName: record.MaterialBaseUnitName,
				Num:                  record.TotalNum,
				Price:                record.Valuation,
				Amount:               decimal.NewFromFloat(record.TotalNum).Mul(decimal.NewFromFloat(record.Valuation)).Round(2).InexactFloat64(),
				SupplierUuid:         record.SupplierUuid,
				OrderNo:              orderNo,
				OpeningHours:         openingHours,
			}

			if err := warehouseLogRepo.Create(logRecord); err != nil {
				logger.Logger.Error("保存出库记录失败: warehouse_uuid=%d, material_uuid=%d, error=%v", zap.Uint64("warehouse_uuid", record.WarehouseUuid), zap.Uint64("material_uuid", record.MaterialUuid), zap.Error(err))
				continue
			}
		}
		// 更新销售订单原料的统计状态
		saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(tx)
		if err := saleOrderMaterialRepo.UpdateSaleOrderMaterialIsSummarized(uuids); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// generateOrderNo 生成出库单号，格式：SSCK + YYYYMMDD + 4位序号
func (t *DailySalesOutboundSummaryTask) generateOrderNo(companyUuid uint64) (string, error) {
	db := t.dbm.GetDB(companyUuid)

	// 获取今天的日期
	now := time.Now()
	dateStr := now.Format("20060102")

	// 使用 repository 方法查询今天已有的SSCK开头的出库单号
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)
	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(1), // 出库
		warehouseLogRepo.WhereScene(1),   // 销售出库
		warehouseLogRepo.WhereCreateTimeBetween(
			int(now.Truncate(24*time.Hour).Unix()),
			int(now.Truncate(24*time.Hour).Add(24*time.Hour).Add(-time.Second).Unix()),
		),
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return "", err
	}

	// 解析最大序号
	sequence := 1
	for _, log := range existingLogs {
		if len(log.OrderNo) >= 16 {
			seqStr := log.OrderNo[12:16] // 取最后4位作为序号
			if seq, err := strconv.Atoi(seqStr); err == nil && seq >= sequence {
				sequence = seq + 1
			}
		}
	}

	// 生成4位序号，不足补0
	sequenceStr := fmt.Sprintf("%04d", sequence)

	return "SSCK" + dateStr + sequenceStr, nil
}
