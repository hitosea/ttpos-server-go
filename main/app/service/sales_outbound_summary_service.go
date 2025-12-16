package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/selling"
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

	// RegenerateSaleBillMaterialOutbound 重新生成指定销售账单的材料出库记录
	// ctx: gin.Context（可为 nil，用于命令行环境）
	// companyUuid: 门店UUID
	// saleBillUuid: 销售账单UUID
	RegenerateSaleBillMaterialOutbound(ctx *gin.Context, companyUuid uint64, saleBillUuid uint64) (*resp.RegenerateSaleBillMaterialOutboundResp, error)

	// RegenerateOrderPosInvoice 重新生成指定订单的POS发票
	// ctx: gin.Context（可为 nil，用于命令行环境）
	// companyUuid: 门店UUID
	// saleOrderUuid: 销售订单UUID
	RegenerateOrderPosInvoice(ctx *gin.Context, companyUuid uint64, saleOrderUuid uint64) (*resp.RegenerateOrderPosInvoiceResp, error)
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

// RegenerateSaleBillMaterialOutbound 重新生成指定销售账单的材料出库记录
func (s *salesOutboundSummarySrv) RegenerateSaleBillMaterialOutbound(
	ctx *gin.Context,
	companyUuid uint64,
	saleBillUuid uint64,
) (*resp.RegenerateSaleBillMaterialOutboundResp, error) {
	startTime := time.Now()
	db := s.dbm.GetDB(companyUuid)

	// 1. 获取分布式锁
	lockKey := fmt.Sprintf("regenerate_sale_bill_material_outbound:%d:%d", companyUuid, saleBillUuid)
	systemLock := lock.NewSystemLock()
	if !systemLock.TryLockUuidString(lockKey) {
		return nil, errors.New("操作进行中，请稍后再试")
	}
	defer systemLock.UnlockUuidString(lockKey)

	// 2. 查询销售账单的所有材料出库记录（scene = 0 AND revoke_time = 0 AND material_uuid != 0）
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItem(
		repository.CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.WhereByNotRevoked(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("scene = ? AND material_uuid != ?", constant.WarehouseOutFormSceneSales, 0)
		},
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询材料出库记录失败")
	}

	materialItems := warehouseOutFormItems

	// 按 warehouse_out_form_uuid 分组
	formItemMap := make(map[uint64][]*model.WarehouseOutFormItem)
	for _, item := range materialItems {
		formItemMap[item.WarehouseOutFormUuid] = append(formItemMap[item.WarehouseOutFormUuid], item)
	}

	// 3. 获取订单信息并计算材料消耗
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单完整信息失败")
	}
	if saleBill == nil {
		return nil, errors.New("销售账单不存在")
	}

	// 计算所有订单的材料消耗
	materialStocksMap := make(map[uint64]*model.MaterialStock) // key: material_uuid
	for _, saleOrder := range saleBill.SaleOrders {
		materialStocks := saleOrder.GetValidSaleOrderProductMaterialList()
		for _, materialStock := range materialStocks {
			if existing, ok := materialStocksMap[materialStock.MaterialUuid]; ok {
				// 累加相同材料的数量
				existing.StockNum = decimal.NewFromFloat(existing.StockNum).
					Add(decimal.NewFromFloat(materialStock.StockNum)).
					Round(4).
					InexactFloat64()
			} else {
				materialStocksMap[materialStock.MaterialUuid] = &model.MaterialStock{
					MaterialUuid:  materialStock.MaterialUuid,
					WarehouseUuid: materialStock.WarehouseUuid,
					StockNum:      materialStock.StockNum,
				}
			}
		}
	}

	// 转换为列表
	materialStocksList := make([]*model.MaterialStock, 0)
	for _, materialStock := range materialStocksMap {
		materialStocksList = append(materialStocksList, materialStock)
	}

	// 4. 使用事务执行退库、软删除、创建和扣库操作
	var deletedCount int64
	var insertedCount int64

	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 4.1 退回库存
		if len(materialItems) > 0 {
			// 统计要退回的材料
			returnStockMap, err := s.calculateReturnStockMap(tx, materialItems)
			if err != nil {
				return errors.WithMessage(err, "统计要退回的材料失败")
			}

			// 执行退回库存操作
			if err := s.returnStock(tx, returnStockMap); err != nil {
				return errors.WithMessage(err, "退回库存失败")
			}
		}

		// 4.2 软删除原记录
		if len(materialItems) > 0 {
			result := tx.Model(&model.WarehouseOutFormItem{}).
				Where("sale_bill_uuid = ? AND scene = ? AND revoke_time = ? AND material_uuid != ? AND delete_time = ?",
					saleBillUuid, constant.WarehouseOutFormSceneSales, 0, 0, constant.NotDeleted).
				Update("delete_time", time.Now().Unix())
			if result.Error != nil {
				return errors.WithMessage(result.Error, "软删除原记录失败")
			}
			deletedCount = result.RowsAffected
		}

		// 4.3 创建新记录并关联原出库单UUID
		warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
		newItems := make([]*model.WarehouseOutFormItem, 0)
		for warehouseOutFormUuid, originalItems := range formItemMap {
			// 验证出库单是否存在
			warehouseOutForm, err := warehouseFormRepo.GetWarehouseForm(
				repository.CommonRepo.WhereByUuid(warehouseOutFormUuid),
				repository.CommonRepo.WhereBySoftDelete(),
			)
			if err != nil || warehouseOutForm == nil {
				logger.Logger.Warn("出库单不存在，使用原UUID",
					zap.Uint64("warehouseOutFormUuid", warehouseOutFormUuid),
					zap.Uint64("saleBillUuid", saleBillUuid),
					zap.Error(err),
				)
			}

			// 为每个材料创建新记录
			for _, materialStock := range materialStocksList {
				// 查找对应的原记录（按 material_uuid 匹配）
				var originalItem *model.WarehouseOutFormItem
				for _, item := range originalItems {
					if item.MaterialUuid == materialStock.MaterialUuid {
						originalItem = item
						break
					}
				}

				// 如果没有对应的原记录，跳过（可能该材料是新添加的，或者不在这个出库单中）
				if originalItem == nil {
					continue
				}

				// 创建新记录
				uuid, _ := utils.GetID()
				newItem := &model.WarehouseOutFormItem{
					BaseModel: model.BaseModel{
						Uuid:       uuid,
						CreateTime: time.Now().Unix(),
					},
					WarehouseOutFormUuid: warehouseOutFormUuid, // 关联原出库单UUID
					WarehouseUuid:        materialStock.WarehouseUuid,
					MaterialUuid:         materialStock.MaterialUuid,
					SaleBillUuid:         saleBillUuid,
					SaleOrderUuid:        originalItem.SaleOrderUuid,
					StaffShiftLogUuid:    originalItem.StaffShiftLogUuid,
					Num:                  decimal.NewFromFloat(materialStock.StockNum).Round(4).InexactFloat64(), // 保留4位小数
					Scene:                constant.WarehouseOutFormSceneSales,
					Status:               constant.WarehouseOutFormItemStatusSuccess,
					ReduceStock:          constant.WarehouseOutFormItemReduceStockNotProcessed,
				}
				newItems = append(newItems, newItem)
			}
		}

		// 批量创建新记录
		if len(newItems) > 0 {
			if err := warehouseFormRepo.CreateWarehouseOutFormItemRecords(newItems); err != nil {
				return errors.WithMessage(err, "创建新记录失败")
			}
			insertedCount = int64(len(newItems))

			// 4.4 扣减库存
			// 统计要扣减的材料
			reduceStockMap, err := s.calculateReduceStockMap(tx, newItems)
			if err != nil {
				return errors.WithMessage(err, "统计要扣减的材料失败")
			}

			// 执行扣减库存操作
			if err := s.reduceStock(tx, reduceStockMap); err != nil {
				return errors.WithMessage(err, "扣减库存失败")
			}

			// 更新出库单明细的 reduce_stock = 1
			if err := warehouseFormRepo.UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid); err != nil {
				return errors.WithMessage(err, "更新 reduce_stock 失败")
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.WithMessage(err, "重新生成销售账单材料出库记录失败")
	}

	durationMs := time.Since(startTime).Milliseconds()

	return &resp.RegenerateSaleBillMaterialOutboundResp{
		DeletedCount:  int(deletedCount),
		InsertedCount: int(insertedCount),
		DurationMs:    durationMs,
	}, nil
}

// RegenerateOrderPosInvoice 重新生成指定订单的POS发票
func (s *salesOutboundSummarySrv) RegenerateOrderPosInvoice(
	ctx *gin.Context,
	companyUuid uint64,
	saleOrderUuid uint64,
) (*resp.RegenerateOrderPosInvoiceResp, error) {
	startTime := time.Now()
	db := s.dbm.GetDB(companyUuid)

	// 1. 读取订单信息
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	saleOrder, err := saleOrderRepo.GetSaleOrderByUuid(saleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单信息失败")
	}
	if saleOrder == nil || saleOrder.Uuid == 0 {
		return nil, errors.New("订单不存在")
	}

	// 2. 验证订单状态（必须已完成结账）
	if saleOrder.FinishTime == 0 {
		return nil, errors.New("订单未完成结账，无法生成发票")
	}

	// 3. 读取账单信息
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(saleOrder.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取账单信息失败")
	}
	if saleBill == nil {
		return nil, errors.New("账单不存在")
	}

	saleOrder = saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("SaleOrder订单不存在")
	}

	// 4. 获取公司信息和设置
	companyRepo := repository.NewCompanyRepo(db)
	company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取公司信息失败")
	}

	// 5. 验证ERP配置
	if !company.IsOpenErpPhase3() {
		return nil, errors.New("ERP Phase3未启用，无法生成发票")
	}
	if company.CompanySetting.ErpnextSiteCode == "" {
		return nil, errors.New("ERP SiteCode未配置，无法生成发票")
	}

	// 6. 创建 Context（命令行环境）
	ttposCtx := context.NewContext(
		context.WithCompanyUuid(companyUuid),
		context.WithCompany(*company),
		context.WithCompanySetting(*company.CompanySetting),
		context.WithLogger(logger.Logger),
	)
	ttposCtx.SetDB(db)

	// 7. 设置员工信息（从订单中获取，SavePosInvoice 需要）
	var staff *model.Staff
	if saleOrder.CashierUuid > 0 {
		staffRepo := repository.NewStaffRepo(db)
		staffFromDB, err := staffRepo.GetStaff(repository.CommonRepo.WhereByUuid(saleOrder.CashierUuid))
		if err == nil {
			staff = &staffFromDB
			// 重新创建 Context 包含员工信息
			ttposCtx = context.NewContext(
				context.WithCompanyUuid(companyUuid),
				context.WithCompany(*company),
				context.WithCompanySetting(*company.CompanySetting),
				context.WithStaff(*staff),
				context.WithStaffUuid(staff.Uuid),
				context.WithLogger(logger.Logger),
			)
			ttposCtx.SetDB(db)
		}
	}

	// 8. 获取 shiftLog（用于 SavePosInvoice 的 WithShiftLog 选项）
	shiftLog := s.getShiftLogForSavePosInvoice(db, staff)
	if shiftLog == nil {
		return nil, errors.New("获取 shiftLog 失败,当前门店没有当班记录")
	}

	// 9. 创建 OrderSrv 实例并调用 SavePosInvoice 方法
	localeSrv := NewLocaleSrv()
	mustPlanSrv := NewMustPlanSrv(s.dbm)
	paymentMethodSrv := NewPaymentMethodSrv(s.dbm, s.settingSrv)
	memberSrv := NewMemberSrv(s.dbm, s.cache)
	cashBoxSrv := NewCashBoxSrv(s.dbm)
	orderSrv := NewOrderSrv(s.dbm, localeSrv, s.settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv)

	// 调用 SavePosInvoice 方法（通过接口调用，如果提供了 shiftLog 则通过选项传入）
	var savePosInvoiceResp *selling.SavePosInvoiceResp
	savePosInvoiceResp, err = orderSrv.SavePosInvoice(ttposCtx, saleOrder, saleBill, db, WithShiftLog(shiftLog))
	if err != nil {
		return nil, errors.WithMessage(err, "保存发票失败")
	}

	// 9. 更新订单发票信息
	err = saleOrderRepo.UpdateSaleOrderErpInvoice(
		saleOrder.Uuid,
		savePosInvoiceResp.ProductsInvoiceName,
		savePosInvoiceResp.MaterialInvoiceName,
	)
	if err != nil {
		logger.Logger.Warn("更新订单发票信息失败", zap.Uint64("saleOrderUuid", saleOrder.Uuid), zap.Error(err))
		// 发票已保存到ERP，但更新订单信息失败，返回警告但不影响整体流程
	}

	durationMs := time.Since(startTime).Milliseconds()

	return &resp.RegenerateOrderPosInvoiceResp{
		ProductsInvoiceName: savePosInvoiceResp.ProductsInvoiceName,
		MaterialInvoiceName: savePosInvoiceResp.MaterialInvoiceName,
		DurationMs:          durationMs,
	}, nil
}

// returnStockInfo 退回库存信息结构
type returnStockInfo struct {
	WarehouseUuid uint64
	MaterialUuid  uint64
	ReturnNum     float64
	Material      *model.Material
}

// calculateReturnStockMap 统计要退回的材料（按 warehouse_uuid 和 material_uuid 分组汇总）
func (s *salesOutboundSummarySrv) calculateReturnStockMap(tx *gorm.DB, materialItems []*model.WarehouseOutFormItem) (map[string]*returnStockInfo, error) {
	materialRepo := repository.NewMaterialRepo(tx)

	// 按 warehouse_uuid 和 material_uuid 分组汇总需要退回的数量
	returnStockMap := make(map[string]*returnStockInfo)

	for _, item := range materialItems {
		key := fmt.Sprintf("%d_%d", item.WarehouseUuid, item.MaterialUuid)
		if returnStock, ok := returnStockMap[key]; ok {
			returnStock.ReturnNum += item.Num
		} else {
			// 获取材料信息（预加载关联材料和基准单位）
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid,
				materialRepo.WithRelatedMaterialList(),
				repository.CommonRepo.Preload(
					repository.WithPreload{
						Query: "NotBaseUnitList.Unit.MultiLanguageName",
					},
				),
			)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("获取材料信息失败: %d", item.MaterialUuid))
			}

			returnStockMap[key] = &returnStockInfo{
				WarehouseUuid: item.WarehouseUuid,
				MaterialUuid:  item.MaterialUuid,
				ReturnNum:     item.Num,
				Material:      &material,
			}
		}
	}

	return returnStockMap, nil
}

// returnStock 执行退回库存操作（增加库存、更新关联库存）
func (s *salesOutboundSummarySrv) returnStock(tx *gorm.DB, returnStockMap map[string]*returnStockInfo) error {
	warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
	materialRepo := repository.NewMaterialRepo(tx)

	// 收集需要更新关联库存的材料UUID
	relatedMaterialUuids := make([]uint64, 0)
	relatedMaterialUuidSet := make(map[uint64]bool)

	// 退回库存
	for _, returnInfo := range returnStockMap {
		if returnInfo.ReturnNum <= 0 {
			continue
		}

		// 获取或创建仓库物品库存记录
		warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterialOrCreate(
			returnInfo.WarehouseUuid,
			returnInfo.MaterialUuid,
			returnInfo.Material.Code,
			returnInfo.Material.Valuation,
		)
		if err != nil {
			return errors.WithMessage(err, "获取仓库物品库存失败")
		}

		// 增加库存
		if err := warehouseItemRepo.AddStock(warehouseItem.Uuid, returnInfo.ReturnNum); err != nil {
			return errors.WithMessage(err, "增加材料库存失败")
		}

		// 收集需要更新关联库存的材料UUID
		relatedUuids := returnInfo.Material.GetRelatedMaterialUuids()
		for _, uuid := range relatedUuids {
			if !relatedMaterialUuidSet[uuid] {
				relatedMaterialUuids = append(relatedMaterialUuids, uuid)
				relatedMaterialUuidSet[uuid] = true
			}
		}
	}

	// 更新关联材料库存
	if len(relatedMaterialUuids) > 0 {
		if err := materialRepo.UpdateRelatedMaterialStock(relatedMaterialUuids); err != nil {
			return errors.WithMessage(err, "更新关联材料库存失败")
		}
	}

	return nil
}

// reduceStockInfo 扣减库存信息结构
type reduceStockInfo struct {
	WarehouseUuid uint64
	MaterialUuid  uint64
	ReduceNum     float64
	Material      *model.Material
}

// calculateReduceStockMap 统计要扣减的材料（按 warehouse_uuid 和 material_uuid 分组汇总）
func (s *salesOutboundSummarySrv) calculateReduceStockMap(tx *gorm.DB, newItems []*model.WarehouseOutFormItem) (map[string]*reduceStockInfo, error) {
	materialRepo := repository.NewMaterialRepo(tx)

	// 按 warehouse_uuid 和 material_uuid 分组汇总需要扣减的数量
	reduceStockMap := make(map[string]*reduceStockInfo)

	for _, item := range newItems {
		key := fmt.Sprintf("%d_%d", item.WarehouseUuid, item.MaterialUuid)
		if reduceStock, ok := reduceStockMap[key]; ok {
			reduceStock.ReduceNum += item.Num
		} else {
			// 获取材料信息（预加载关联材料和基准单位）
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid,
				materialRepo.WithRelatedMaterialList(),
				repository.CommonRepo.Preload(
					repository.WithPreload{
						Query: "NotBaseUnitList.Unit.MultiLanguageName",
					},
				),
			)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("获取材料信息失败: %d", item.MaterialUuid))
			}

			reduceStockMap[key] = &reduceStockInfo{
				WarehouseUuid: item.WarehouseUuid,
				MaterialUuid:  item.MaterialUuid,
				ReduceNum:     item.Num,
				Material:      &material,
			}
		}
	}

	return reduceStockMap, nil
}

// reduceStock 执行扣减库存操作（扣减库存、更新关联库存）
func (s *salesOutboundSummarySrv) reduceStock(tx *gorm.DB, reduceStockMap map[string]*reduceStockInfo) error {
	warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
	materialRepo := repository.NewMaterialRepo(tx)

	// 收集需要更新关联库存的材料UUID
	relatedMaterialUuids := make([]uint64, 0)
	relatedMaterialUuidSet := make(map[uint64]bool)

	// 扣减库存
	for _, reduceInfo := range reduceStockMap {
		if reduceInfo.ReduceNum <= 0 {
			continue
		}

		// 获取仓库物品库存记录
		warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(reduceInfo.WarehouseUuid, reduceInfo.MaterialUuid)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("获取仓库物品库存失败: material_uuid=%d, warehouse_uuid=%d", reduceInfo.MaterialUuid, reduceInfo.WarehouseUuid))
		}

		// 检查库存是否充足
		if warehouseItem.Stock < reduceInfo.ReduceNum {
			return errors.New(fmt.Sprintf("材料库存不足: material_uuid=%d, warehouse_uuid=%d, 需要=%f, 当前=%f",
				reduceInfo.MaterialUuid, reduceInfo.WarehouseUuid, reduceInfo.ReduceNum, warehouseItem.Stock))
		}

		// 扣减库存
		if err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, reduceInfo.ReduceNum); err != nil {
			return errors.WithMessage(err, "扣减库存失败")
		}

		// 收集需要更新关联库存的材料UUID
		relatedUuids := reduceInfo.Material.GetRelatedMaterialUuids()
		for _, uuid := range relatedUuids {
			if !relatedMaterialUuidSet[uuid] {
				relatedMaterialUuids = append(relatedMaterialUuids, uuid)
				relatedMaterialUuidSet[uuid] = true
			}
		}
	}

	// 更新关联材料库存
	if len(relatedMaterialUuids) > 0 {
		if err := materialRepo.UpdateRelatedMaterialStock(relatedMaterialUuids); err != nil {
			return errors.WithMessage(err, "更新关联材料库存失败")
		}
	}

	return nil
}

// getShiftLogForSavePosInvoice 获取 shiftLog，用于为 SavePosInvoice 的 WithShiftLog 提供参数
// 获取规则：
// 1. 如果订单的收银员还在当班，就取该收银员的当班记录
// 2. 如果收银员不当班了，则选择最新的一个正在当班的 shiftLog
func (s *salesOutboundSummarySrv) getShiftLogForSavePosInvoice(db *gorm.DB, staff *model.Staff) *model.StaffShiftLog {
	shiftLogRepo := repository.NewShiftLogRepo(db)

	// 如果订单有收银员且收银员还在当班，尝试获取该收银员的当班记录
	if staff != nil && staff.CashierOnline == 1 && staff.DutyNo != "" {
		shiftLog, err := shiftLogRepo.GetShiftLog(
			repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
			repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
			func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", constant.StaffNotHandedOver) // 未交班（正在当班）
			},
		)
		if err == nil && shiftLog.Uuid > 0 {
			return &shiftLog
		}
	}

	// 如果收银员不当班了或没有当班记录，则选择最新的一个正在当班的 shiftLog
	shiftLogList, err := shiftLogRepo.GetShiftLogList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", constant.StaffNotHandedOver). // 未交班（正在当班）
											Order("create_time DESC"). // 按创建时间倒序
											Limit(1)                   // 只取最新的一条
		},
	)
	if err == nil && len(shiftLogList) > 0 {
		return &shiftLogList[0]
	}

	// 如果都没有找到，返回 nil
	return nil
}
