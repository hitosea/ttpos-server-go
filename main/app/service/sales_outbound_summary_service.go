package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ISalesOutboundSummarySrv 销售出库汇总服务接口
type ISalesOutboundSummarySrv interface {
	// RegenerateSalesOutboundSummary 重新生成指定日期的销售出库汇总记录
	// ctx: gin.Context（可为 nil，用于命令行环境）
	// companyUuid: 门店UUID
	// date: 日期，格式：YYYY-MM-DD
	RegenerateSalesOutboundSummary(ctx *gin.Context, companyUuid uint64, date string) (*resp.RegenerateSalesOutboundSummaryResp, error)

	// RegenerateOrderMaterial 重新生成指定订单的材料用料记录
	// ctx: gin.Context（可为 nil，用于命令行环境）
	// companyUuid: 门店UUID
	// saleOrderUuid: 订单UUID
	RegenerateOrderMaterial(ctx *gin.Context, companyUuid uint64, saleOrderUuid uint64) (*resp.RegenerateOrderMaterialResp, error)
}

// salesOutboundSummarySrv 销售出库汇总服务实现
type salesOutboundSummarySrv struct {
	dbm        *database.DBManager
	settingSrv setting.ISrv
	cache      cache.Cache
}

// NewSalesOutboundSummarySrv 创建销售出库汇总服务实例
func NewSalesOutboundSummarySrv(
	dbm *database.DBManager,
	settingSrv setting.ISrv,
	cache cache.Cache,
) ISalesOutboundSummarySrv {
	return &salesOutboundSummarySrv{
		dbm:        dbm,
		settingSrv: settingSrv,
		cache:      cache,
	}
}

// RegenerateSalesOutboundSummary 重新生成指定日期的销售出库汇总记录
func (s *salesOutboundSummarySrv) RegenerateSalesOutboundSummary(
	ctx *gin.Context,
	companyUuid uint64,
	date string,
) (*resp.RegenerateSalesOutboundSummaryResp, error) {
	startTime := time.Now()

	// 1. 获取分布式锁
	lockKey := fmt.Sprintf("regenerate_sales_outbound_summary:%d:%s", companyUuid, date)
	systemLock := lock.NewSystemLock()
	if !systemLock.TryLockUuidString(lockKey) {
		return nil, errors.New("操作进行中，请稍后再试")
	}
	defer systemLock.UnlockUuidString(lockKey)

	// 2. 获取门店信息
	companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(0))
	company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "门店不存在")
	}

	// 3. 获取门店时区并解析日期
	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}
	timeUtil := utils.Timezone(timezone)
	targetDate, err := timeUtil.FormatTimeToTime(date)
	if err != nil {
		return nil, errors.WithMessage(err, "日期格式错误，应为 YYYY-MM-DD")
	}

	// 5. 获取营业时段配置
	openingHours, err := s.getOpeningHours(companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取营业时段失败")
	}

	// 6. 计算时间范围
	startTimeUnix, endTimeUnix := s.calculateTimeRange(company, openingHours, targetDate)

	// 7. 构建 opening_hours 字符串
	dateStr := targetDate.Format("20060102")
	openingYearHours := fmt.Sprintf("%s %s", dateStr, openingHours)

	// 8. 删除旧记录
	deletedCount, err := s.deleteOldRecords(companyUuid, dateStr)
	if err != nil {
		return nil, errors.WithMessage(err, "删除旧记录失败")
	}

	// 9. 重新生成记录
	generatedCount, err := s.generateNewRecords(companyUuid, startTimeUnix, endTimeUnix, openingYearHours)
	if err != nil {
		return nil, errors.WithMessage(err, "生成新记录失败")
	}

	durationMs := time.Since(startTime).Milliseconds()

	return &resp.RegenerateSalesOutboundSummaryResp{
		DeletedCount:   deletedCount,
		GeneratedCount: generatedCount,
		DurationMs:     durationMs,
	}, nil
}

// deleteOldRecords 删除指定日期的旧销售出库汇总记录
func (s *salesOutboundSummarySrv) deleteOldRecords(companyUuid uint64, dateStr string) (int, error) {
	db := s.dbm.GetDB(companyUuid)
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)

	// 查询该日期所有营业时段的销售出库汇总记录
	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(constant.WarehouseInOutLogLogTypeOut), // 出库
		warehouseLogRepo.WhereScene(constant.WarehouseInOutLogSceneSale),    // 销售出库
		repository.CommonRepo.WhereBySoftDelete(),                           // 未删除的记录
		func(db *gorm.DB) *gorm.DB {
			return db.Where("opening_hours LIKE ?", dateStr+"%")
		},
	}

	oldLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return 0, errors.WithMessage(err, "查询旧记录失败")
	}

	if len(oldLogs) == 0 {
		return 0, nil
	}

	// 软删除旧记录
	deletedCount := 0
	for _, log := range oldLogs {
		if err := warehouseLogRepo.Delete(log.Uuid); err != nil {
			logger.Logger.Warn("删除旧汇总记录失败", zap.Uint64("uuid", log.Uuid), zap.Error(err))
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// generateNewRecords 重新生成销售出库汇总记录
func (s *salesOutboundSummarySrv) generateNewRecords(
	companyUuid uint64,
	startTimeUnix int64,
	endTimeUnix int64,
	openingYearHours string,
) (int, error) {
	// 获取销售出库记录
	outboundRecords, err := s.getDailySalesOutboundRecords(companyUuid, startTimeUnix, endTimeUnix)
	if err != nil {
		return 0, errors.WithMessage(err, "获取销售出库记录失败")
	}

	if len(outboundRecords) == 0 {
		return 0, nil
	}

	// 保存汇总记录
	if err := s.saveOutboundSummaryRecords(companyUuid, outboundRecords, openingYearHours); err != nil {
		return 0, errors.WithMessage(err, "保存出库汇总记录失败")
	}

	return len(outboundRecords), nil
}

// getOpeningHours 获取门店营业时间（从 DailySalesOutboundSummaryTask 提取）
func (s *salesOutboundSummarySrv) getOpeningHours(companyUuid uint64) (string, error) {
	ctx := context.NewContext()
	ctx.SetCompanyUuid(companyUuid)
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return "", err
	}
	if businessSetting.OpeningHours != "" {
		return businessSetting.OpeningHours, nil
	}

	// 返回默认值
	return "00:00-23:59", nil
}

// calculateTimeRange 计算指定日期的营业时间范围（从 DailySalesOutboundSummaryTask 提取并修改）
func (s *salesOutboundSummarySrv) calculateTimeRange(
	company *model.Company,
	openingHours string,
	targetDate time.Time,
) (int64, int64) {
	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}

	// 加载时区
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.Local
	}

	// 解析营业时间格式 HH:MM-HH:MM
	timeRanges := strings.Split(openingHours, "-")
	if len(timeRanges) != 2 {
		// 格式错误，返回目标日期整天
		startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
		endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 0, loc)
		return startTime.Unix(), endTime.Unix()
	}

	startTimeStr := strings.TrimSpace(timeRanges[0])
	endTimeStr := strings.TrimSpace(timeRanges[1])

	// 解析开始时间
	startHour, startMin := 0, 0
	startTimeParts := strings.Split(startTimeStr, ":")
	if len(startTimeParts) >= 1 {
		startHour, _ = strconv.Atoi(startTimeParts[0])
	}
	if len(startTimeParts) >= 2 {
		startMin, _ = strconv.Atoi(startTimeParts[1])
	}

	// 解析结束时间
	endHour, endMin := 0, 0
	endTimeParts := strings.Split(endTimeStr, ":")
	if len(endTimeParts) >= 1 {
		endHour, _ = strconv.Atoi(endTimeParts[0])
	}
	if len(endTimeParts) >= 2 {
		endMin, _ = strconv.Atoi(endTimeParts[1])
	}

	// 构建目标日期的开始和结束时间
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), startHour, startMin, 0, 0, loc)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), endHour, endMin, 59, 0, loc)

	// 如果结束时间小于开始时间，说明跨天了，结束时间加一天
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	return startTime.Unix(), endTime.Unix()
}

// salesOutboundRecord 销售出库记录汇总（从 DailySalesOutboundSummaryTask 提取）
type salesOutboundRecord struct {
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

// getDailySalesOutboundRecords 获取当天销售出库记录（从 DailySalesOutboundSummaryTask 提取）
func (s *salesOutboundSummarySrv) getDailySalesOutboundRecords(
	companyUuid uint64,
	startTime int64,
	endTime int64,
) ([]*salesOutboundRecord, error) {
	db := s.dbm.GetDB(companyUuid)

	// 使用 repository 方法查询出库单明细（包含已统计的，因为要重新生成）
	saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(db)
	saleOrderMaterials, err := saleOrderMaterialRepo.GetSaleOrderMaterialByCreateTimeBetweenAll(
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	// 按仓库和物料分组汇总数量
	recordMap := make(map[string]*salesOutboundRecord)
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
				if item.Material.Unit != nil {
					materialBaseUnitName = item.Material.Unit.Name
				}
			}

			recordMap[key] = &salesOutboundRecord{
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
	var records []*salesOutboundRecord
	for _, record := range recordMap {
		records = append(records, record)
	}

	return records, nil
}

// saveOutboundSummaryRecords 保存出库汇总记录到ttpos_warehouse_in_out_log表（从 DailySalesOutboundSummaryTask 提取）
func (s *salesOutboundSummarySrv) saveOutboundSummaryRecords(
	companyUuid uint64,
	records []*salesOutboundRecord,
	openingHours string,
) error {
	db := s.dbm.GetDB(companyUuid)

	// 生成出库单号
	orderNo, err := s.generateOrderNo(companyUuid, openingHours)
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
				logger.Logger.Error("保存出库记录失败", zap.Uint64("warehouse_uuid", record.WarehouseUuid), zap.Uint64("material_uuid", record.MaterialUuid), zap.Error(err))
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

// generateOrderNo 生成出库单号，格式：SSCK + YYYYMMDD + 4位序号（从 DailySalesOutboundSummaryTask 提取并修改）
func (s *salesOutboundSummarySrv) generateOrderNo(companyUuid uint64, openingHours string) (string, error) {
	db := s.dbm.GetDB(companyUuid)

	// 从 opening_hours 中提取日期（格式：YYYYMMDD HH:mm-HH:mm）
	dateStr := openingHours[:8] // 取前8位作为日期

	// 使用 repository 方法查询该日期已有的SSCK开头的出库单号
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)

	// 解析日期，获取时间范围
	targetDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("解析日期失败: %w", err)
	}

	startOfDay := targetDate.Truncate(24 * time.Hour).Unix()
	endOfDay := targetDate.Truncate(24 * time.Hour).Add(24 * time.Hour).Add(-time.Second).Unix()

	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(constant.WarehouseInOutLogLogTypeOut), // 出库
		warehouseLogRepo.WhereScene(constant.WarehouseInOutLogSceneSale),    // 销售出库
		warehouseLogRepo.WhereCreateTimeBetween(int(startOfDay), int(endOfDay)),
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return "", err
	}

	// 解析最大序号
	sequence := 1
	for _, log := range existingLogs {
		if len(log.OrderNo) >= 16 && log.OrderNo[:12] == "SSCK"+dateStr {
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

// RegenerateOrderMaterial 重新生成指定订单的材料用料记录
func (s *salesOutboundSummarySrv) RegenerateOrderMaterial(
	ctx *gin.Context,
	companyUuid uint64,
	saleOrderUuid uint64,
) (*resp.RegenerateOrderMaterialResp, error) {
	startTime := time.Now()
	db := s.dbm.GetDB(companyUuid)

	// 1. 获取订单信息
	orderRepo := repository.NewOrderRepo(db)

	// 先通过 sale_order_uuid 获取 sale_bill_uuid
	saleOrder, err := orderRepo.GetSaleBillSaleOrderRecord(saleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单信息失败")
	}
	if saleOrder == nil {
		return nil, errors.New("订单不存在")
	}

	saleBillUuid := saleOrder.SaleBillUuid

	// 获取订单完整信息（包含商品、BOM、材料关联等）
	saleBill, err := orderRepo.GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单完整信息失败")
	}
	if saleBill == nil {
		return nil, errors.New("订单账单不存在")
	}

	// 获取指定的订单
	saleOrderFromBill := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrderFromBill == nil {
		return nil, errors.New("订单不存在于账单中")
	}

	// 计算材料用量
	materialStocks := saleOrderFromBill.GetValidSaleOrderProductMaterialList()

	// 构建材料记录列表
	saleOrderMaterials := make([]*model.SaleOrderMaterial, 0)
	for _, materialStock := range materialStocks {
		saleOrderMaterials = append(saleOrderMaterials, &model.SaleOrderMaterial{
			BaseModel: model.BaseModel{
				CreateTime: saleOrderFromBill.FinishTime, // 原料使用时间=销售订单完成时间
			},
			SaleOrderUuid:     saleOrderFromBill.Uuid,
			SaleBillUuid:      saleBillUuid,
			MaterialUuid:      materialStock.MaterialUuid,
			WarehouseUuid:     materialStock.WarehouseUuid,
			Num:               decimal.NewFromFloat(materialStock.StockNum).Round(4).InexactFloat64(), // 保留4位小数
			StaffShiftLogUuid: saleOrderFromBill.StaffShiftLogUuid,
		})
	}

	// 使用事务执行删除和插入操作
	var deletedCount int64
	var insertedCount int64

	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 创建 repository 实例（使用事务的 tx）
		saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(tx)

		// 查询旧记录数量
		oldMaterials, err := saleOrderMaterialRepo.GetSaleOrderMaterialBySaleOrderUuid(saleOrderUuid)
		if err != nil {
			return errors.WithMessage(err, "查询旧记录失败")
		}
		deletedCount = int64(len(oldMaterials))

		// 软删除旧记录（按 sale_order_uuid 删除）
		if deletedCount > 0 {
			if err := saleOrderMaterialRepo.DeleteSaleOrderMaterialBySaleOrderUuid(saleOrderUuid); err != nil {
				return errors.WithMessage(err, "删除旧记录失败")
			}
		}

		// 插入新记录
		if len(saleOrderMaterials) > 0 {
			if err := saleOrderMaterialRepo.BatchInsertSaleOrderMaterial(saleOrderMaterials); err != nil {
				return errors.WithMessage(err, "插入新记录失败")
			}
			insertedCount = int64(len(saleOrderMaterials))
		}

		return nil
	})

	if err != nil {
		return nil, errors.WithMessage(err, "重新生成订单材料记录失败")
	}

	durationMs := time.Since(startTime).Milliseconds()

	return &resp.RegenerateOrderMaterialResp{
		DeletedCount:  int(deletedCount),
		InsertedCount: int(insertedCount),
		DurationMs:    durationMs,
	}, nil
}
