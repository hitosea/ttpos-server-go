package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ICostCardCorrectionSrv 成本卡材料消耗修正服务接口
type ICostCardCorrectionSrv interface {
	// PreviewCorrection 预览修正影响
	// 根据订单UUID列表，预览修正操作的影响范围，包括材料退回数量、新消耗量、受影响的日期等
	PreviewCorrection(ctx context.Context, req req.CostCardCorrectionPreviewReq) (*resp.CostCardCorrectionPreviewResp, error)

	// ExecuteCorrection 执行修正操作
	// 执行完整的修正流程：材料退回 → 重新计算材料消耗 → 重新生成出库记录 → ERP同步 → 每日销售出库修正
	ExecuteCorrection(ctx context.Context, req req.CostCardCorrectionReq) (*resp.CostCardCorrectionResp, error)

	// GetCorrectionLogs 查询修正日志
	// 支持按修正UUID、订单UUID查询，支持分页
	GetCorrectionLogs(ctx context.Context, req req.CostCardCorrectionLogsReq) (*resp.CostCardCorrectionLogsResp, error)
}

// costCardCorrectionSrv 成本卡材料消耗修正服务实现
type costCardCorrectionSrv struct {
	dbm                  *database.DBManager
	saleOrderRepo        repository.ISaleOrderQueryRepo
	saleOrderProductRepo repository.ISaleOrderProductQueryRepo
	warehouseFormRepo    repository.IWarehouseFormQueryRepo
	warehouseItemRepo    repository.IWarehouseItemRepo
	warehouseLogRepo     repository.IWarehouseInOutLogRepo
	materialRepo         repository.IMaterialRepo
	orderProductBomRepo  repository.IOrderProductBomRepo
	correctionLogRepo    repository.ICostCardCorrectionLogRepo
}

// NewCostCardCorrectionSrv 创建成本卡修正服务实例
func NewCostCardCorrectionSrv(dbm *database.DBManager) ICostCardCorrectionSrv {
	return &costCardCorrectionSrv{
		dbm: dbm,
	}
}

// PreviewCorrection 预览修正影响
func (s *costCardCorrectionSrv) PreviewCorrection(ctx context.Context, req req.CostCardCorrectionPreviewReq) (*resp.CostCardCorrectionPreviewResp, error) {
	companyUuid := ctx.GetCompanyUuid()

	// 查询订单列表
	orders, err := s.getOrdersWithBomCards(ctx, companyUuid, req.OrderUuids)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 构建预览响应
	previewResp := &resp.CostCardCorrectionPreviewResp{
		Orders:        make([]resp.OrderCorrectionInfo, 0),
		TotalOrders:   len(orders),
		AffectedDates: make([]string, 0),
	}

	// 用于去重日期
	dateSet := make(map[string]bool)

	// 遍历订单，识别使用成本卡的商品
	for _, order := range orders {
		orderInfo := resp.OrderCorrectionInfo{
			OrderUuid:  order.Uuid,
			OrderNo:    order.OrderNo,
			CreateTime: order.CreateTime,
			Products:   make([]resp.ProductCorrectionInfo, 0),
		}

		// 识别使用成本卡的商品
		productsWithBomCard := s.identifyProductsWithBomCard(order)
		for _, product := range productsWithBomCard {
			// 获取商品的第一个 BOM（规格或加料）
			var firstBomUuid uint64
			for _, bom := range product.SaleOrderProductBoms {
				if !bom.IsDelete() {
					firstBomUuid = bom.ProductBomUuid
					break
				}
			}

			productInfo := resp.ProductCorrectionInfo{
				ProductBomUuid: firstBomUuid,
				ProductName:    product.Name,
				BomCardUuid:    0, // 将在下面设置
				Materials:      make([]resp.MaterialCorrectionInfo, 0),
			}

			// 获取成本卡信息
			bomCard, materials := s.getBomCardAndMaterials(product)
			if bomCard != nil {
				productInfo.BomCardUuid = bomCard.Uuid
				// 计算材料修正信息（这里先占位，实际计算在后续任务中实现）
				productInfo.Materials = s.calculateMaterialCorrectionInfo(product, materials)
			}

			if len(productInfo.Materials) > 0 {
				orderInfo.Products = append(orderInfo.Products, productInfo)
			}
		}

		if len(orderInfo.Products) > 0 {
			previewResp.Orders = append(previewResp.Orders, orderInfo)

			// 收集受影响的日期（从订单完成时间计算营业日期）
			businessDate := s.getBusinessDateFromOrder(ctx, companyUuid, order)
			if businessDate != "" {
				if !dateSet[businessDate] {
					dateSet[businessDate] = true
					previewResp.AffectedDates = append(previewResp.AffectedDates, businessDate)
				}
			}
		}
	}

	return previewResp, nil
}

// ExecuteCorrection 执行修正操作
func (s *costCardCorrectionSrv) ExecuteCorrection(ctx context.Context, req req.CostCardCorrectionReq) (*resp.CostCardCorrectionResp, error) {
	companyUuid := ctx.GetCompanyUuid()
	staff := ctx.GetStaff()

	// 生成修正操作UUID
	correctionUuid, err := utils.GetID()
	if err != nil {
		return nil, fmt.Errorf("生成修正操作UUID失败: %w", err)
	}

	// 查询订单列表
	orders, err := s.getOrdersWithBomCards(ctx, companyUuid, req.OrderUuids)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("未找到需要修正的订单，请确保选择的订单都是已完成的订单（订单状态为已结账且完成时间 > 0）")
	}

	// 识别受影响的日期范围
	affectedDates, err := s.identifyAffectedDates(ctx, companyUuid, orders)
	if err != nil {
		return nil, fmt.Errorf("识别受影响日期失败: %w", err)
	}

	// 执行结果统计
	successCount := 0
	failCount := 0
	failedOrders := make([]resp.FailedOrderInfo, 0)

	// 遍历每个订单，执行修正操作
	for _, order := range orders {
		orderErr := s.correctSingleOrder(ctx, companyUuid, correctionUuid, order, staff.Uuid, staff.GetUserName())
		if orderErr != nil {
			failCount++
			failedOrders = append(failedOrders, resp.FailedOrderInfo{
				OrderUuid:    order.Uuid,
				OrderNo:      order.OrderNo,
				ErrorMessage: orderErr.Error(),
			})

			// 记录失败日志
			logData := &model.CostCardCorrectionLog{
				CorrectionUuid: correctionUuid,
				OrderUuid:      order.Uuid,
				OrderNo:        order.OrderNo,
				OperatorUuid:   staff.Uuid,
				OperatorName:   staff.GetUserName(),
				OperationType:  "execute",
				Status:         "failed",
				Message:        orderErr.Error(),
			}
			_ = s.recordCorrectionLog(ctx, companyUuid, logData)
		} else {
			successCount++

			// 记录成功日志
			logData := &model.CostCardCorrectionLog{
				CorrectionUuid: correctionUuid,
				OrderUuid:      order.Uuid,
				OrderNo:        order.OrderNo,
				OperatorUuid:   staff.Uuid,
				OperatorName:   staff.GetUserName(),
				OperationType:  "execute",
				Status:         "success",
				Message:        "修正成功",
			}
			_ = s.recordCorrectionLog(ctx, companyUuid, logData)
		}
	}

	// 重新统计每日销售出库记录（对所有受影响的日期）
	for _, date := range affectedDates {
		if err := s.recalculateDailySalesOutbound(ctx, companyUuid, date); err != nil {
			logger.Logger.Error("重新统计每日销售出库记录失败",
				zap.String("date", date),
				zap.Uint64("correction_uuid", correctionUuid),
				zap.Error(err))
			// 不中断流程，只记录错误
		}
	}

	return &resp.CostCardCorrectionResp{
		CorrectionUuid: correctionUuid,
		SuccessCount:   successCount,
		FailCount:      failCount,
		FailedOrders:   failedOrders,
	}, nil
}

// correctSingleOrder 修正单个订单
func (s *costCardCorrectionSrv) correctSingleOrder(ctx context.Context, companyUuid uint64, correctionUuid uint64, order model.SaleOrder, staffUuid uint64, staffName string) error {
	db := s.dbm.GetDB(companyUuid)

	// 在事务中执行所有修正操作
	return repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 1. 退回错误扣减的材料
		if err := s.returnMaterials(ctx, companyUuid, order.Uuid, order.OrderNo); err != nil {
			return fmt.Errorf("退回材料失败: %w", err)
		}

		// 2. 删除旧的材料消耗记录
		if err := s.deleteOldMaterialConsumptionRecords(ctx, companyUuid, order.Uuid); err != nil {
			return fmt.Errorf("删除旧材料消耗记录失败: %w", err)
		}

		// 3. 重新计算材料消耗（基于当前正确的成本卡）
		if err := s.recalculateMaterialConsumption(ctx, companyUuid, order); err != nil {
			return fmt.Errorf("重新计算材料消耗失败: %w", err)
		}

		// 4. 重新生成出库记录
		if err := s.regenerateOutboundForm(ctx, companyUuid, order, order.SaleBillUuid, staffUuid); err != nil {
			return fmt.Errorf("重新生成出库记录失败: %w", err)
		}

		// 5. 重新同步 ERP 数据
		if err := s.resyncErpData(ctx, companyUuid, order, order.SaleBillUuid); err != nil {
			// ERP 同步失败不中断流程，只记录错误
			logger.Logger.Error("ERP 数据同步失败",
				zap.Uint64("order_uuid", order.Uuid),
				zap.String("order_no", order.OrderNo),
				zap.Error(err))
			// 不返回错误，允许继续执行
		}

		return nil
	})
}

// GetCorrectionLogs 查询修正日志
func (s *costCardCorrectionSrv) GetCorrectionLogs(ctx context.Context, req req.CostCardCorrectionLogsReq) (*resp.CostCardCorrectionLogsResp, error) {
	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)
	logRepo := repository.NewCostCardCorrectionLogRepo(db)

	// 构建查询选项
	opts := make([]repository.DBOption, 0)
	if req.CorrectionUuid > 0 {
		opts = append(opts, logRepo.WhereCorrectionUuid(req.CorrectionUuid))
	}
	if req.OrderUuid > 0 {
		opts = append(opts, logRepo.WhereOrderUuid(req.OrderUuid))
	}

	// 设置分页参数
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询日志列表
	logs, total, err := logRepo.GetListWithPagination(pageNo, pageSize, opts...)
	if err != nil {
		return nil, fmt.Errorf("查询修正日志失败: %w", err)
	}

	// 转换为响应 DTO
	logList := make([]resp.CostCardCorrectionLog, 0, len(logs))
	for _, log := range logs {
		logList = append(logList, resp.CostCardCorrectionLog{
			Uuid:           log.Uuid,
			CorrectionUuid: log.CorrectionUuid,
			OrderUuid:      log.OrderUuid,
			OrderNo:        log.OrderNo,
			OperatorUuid:   log.OperatorUuid,
			OperatorName:   log.OperatorName,
			OperationType:  log.OperationType,
			Status:         log.Status,
			Message:        log.Message,
			Details:        log.Details,
			CreateTime:     log.CreateTime,
		})
	}

	return &resp.CostCardCorrectionLogsResp{
		List:     logList,
		Total:    total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// recordCorrectionLog 记录修正日志
func (s *costCardCorrectionSrv) recordCorrectionLog(ctx context.Context, companyUuid uint64, logData *model.CostCardCorrectionLog) error {
	db := s.dbm.GetDB(companyUuid)
	logRepo := repository.NewCostCardCorrectionLogRepo(db)

	// 如果没有 UUID，生成一个
	if logData.Uuid == 0 {
		uuid, err := utils.GetID()
		if err != nil {
			return fmt.Errorf("生成日志UUID失败: %w", err)
		}
		logData.Uuid = uuid
	}

	// 设置创建时间
	if logData.CreateTime == 0 {
		logData.CreateTime = time.Now().Unix()
	}

	// 保存日志
	if err := logRepo.Create(logData); err != nil {
		logger.Logger.Error("记录修正日志失败",
			zap.Uint64("correction_uuid", logData.CorrectionUuid),
			zap.Uint64("order_uuid", logData.OrderUuid),
			zap.Error(err))
		return fmt.Errorf("记录修正日志失败: %w", err)
	}

	return nil
}

// getOrdersWithBomCards 查询订单列表，并预加载商品、BOM、成本卡等信息
// 只能查询已完成的订单（Status = SaleOrderStatusFinish 且 FinishTime > 0）
func (s *costCardCorrectionSrv) getOrdersWithBomCards(ctx context.Context, companyUuid uint64, orderUuids []uint64) ([]model.SaleOrder, error) {
	if len(orderUuids) == 0 {
		return nil, fmt.Errorf("订单UUID列表不能为空")
	}

	var orders []model.SaleOrder
	db := s.dbm.GetDB(companyUuid)

	// 查询订单，预加载商品、BOM、成本卡等信息
	// 只能查询已完成的订单（Status = SaleOrderStatusFinish 且 FinishTime > 0）
	err := db.Where("uuid IN ?", orderUuids).
		Where("delete_time = ?", constant.NotDeleted).
		Where("status = ?", constant.SaleOrderStatusFinish).
		Where("finish_time > ?", 0).
		Preload("SaleOrderProducts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", constant.NotDeleted)
		}).
		Preload("SaleOrderProducts.SaleOrderProductBoms", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", constant.NotDeleted)
		}).
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard.RelatedMaterials").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard.RelatedMaterials.Material").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard.RelatedMaterials.Material.WarehouseItem").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.ProductBomCard").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.ProductBomCard.RelatedMaterials").
		Preload("SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.ProductBomCard.RelatedMaterials.Material").
		Find(&orders).Error

	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Any("order_uuids", orderUuids))
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return orders, nil
}

// identifyProductsWithBomCard 识别订单中使用成本卡的商品
func (s *costCardCorrectionSrv) identifyProductsWithBomCard(order model.SaleOrder) []*model.SaleOrderProduct {
	productsWithBomCard := make([]*model.SaleOrderProduct, 0)

	for _, product := range order.SaleOrderProducts {
		// 跳过已删除、已取消、未送厨的商品
		if product.IsDelete() || product.IsCancelProduct() || !product.IsSendKitchen() {
			continue
		}

		// 检查商品是否使用成本卡
		hasBomCard := false
		for _, bom := range product.SaleOrderProductBoms {
			if bom.IsDelete() {
				continue
			}
			// 检查规格商品或加料是否使用成本卡
			// ProductBom 是值类型，需要检查是否为零值
			if bom.ProductBom.Uuid == 0 {
				continue
			}
			if bom.IsFlavor() && bom.ProductBom.HasProductBomCard() {
				hasBomCard = true
				break
			} else if bom.IsSauce() && bom.ProductBom.ProductSauceUuid != 0 && bom.ProductBom.ProductSauce.HasProductBomCard() {
				hasBomCard = true
				break
			}
		}

		if hasBomCard {
			productsWithBomCard = append(productsWithBomCard, product)
		}
	}

	return productsWithBomCard
}

// getBomCardAndMaterials 获取商品的成本卡和材料信息
func (s *costCardCorrectionSrv) getBomCardAndMaterials(product *model.SaleOrderProduct) (*model.ProductBomCard, []*model.RelatedMaterial) {
	for _, bom := range product.SaleOrderProductBoms {
		if bom.IsDelete() {
			continue
		}
		// ProductBom 是值类型，需要检查是否为零值
		if bom.ProductBom.Uuid == 0 {
			continue
		}
		if bom.IsFlavor() && bom.ProductBom.HasProductBomCard() {
			card := bom.ProductBom.ProductBomCard
			if card != nil && card.Uuid != 0 && len(card.RelatedMaterials) > 0 {
				return card, card.RelatedMaterials
			}
		} else if bom.IsSauce() && bom.ProductBom.ProductSauceUuid != 0 && bom.ProductBom.ProductSauce.HasProductBomCard() {
			card := bom.ProductBom.ProductSauce.ProductBomCard
			if card != nil && card.Uuid != 0 && len(card.RelatedMaterials) > 0 {
				return card, card.RelatedMaterials
			}
		}
	}
	return nil, nil
}

// getBusinessDateFromOrder 从订单获取营业日期
// 只能选择已经完成的订单（FinishTime > 0）
func (s *costCardCorrectionSrv) getBusinessDateFromOrder(ctx context.Context, companyUuid uint64, order model.SaleOrder) string {
	// 只能使用已完成订单的完成时间（FinishTime）计算营业日期
	if order.FinishTime <= 0 {
		// 订单未完成，返回空字符串
		return ""
	}

	// 使用完成时间计算营业日期
	return s.getBusinessDateFromOrderTime(ctx, companyUuid, order.FinishTime)
}

// getBusinessDateFromOrderTime 根据订单时间计算营业日期（内部方法，需要时区和营业时间）
func (s *costCardCorrectionSrv) getBusinessDateFromOrderTime(ctx context.Context, companyUuid uint64, orderTime int64) string {
	if orderTime <= 0 {
		return ""
	}

	db := s.dbm.GetDB(companyUuid)

	// 获取公司设置（时区和营业时间）
	companyRepo := repository.NewCompanyRepo(db)
	company, err := companyRepo.GetCompany()
	if err != nil {
		logger.Logger.Warn("获取公司信息失败，使用默认时区", zap.Error(err))
		// 使用默认时区和营业时间
		timeUtil := utils.Timezone("Asia/Shanghai")
		return timeUtil.FormatUnixTime(orderTime, "2006-01-02")
	}

	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}
	timeUtil := utils.Timezone(timezone)

	// 获取营业时间
	settingSrv := setting.NewSrvImpl(s.dbm, cache.Global)
	settingCtx := context.NewContext()
	settingCtx.SetCompanyUuid(companyUuid)
	businessSetting, err := settingSrv.GetBusinessSetting(settingCtx)
	if err != nil {
		logger.Logger.Warn("获取营业时间设置失败，使用默认营业时间", zap.Error(err))
		// 使用默认营业时间
		return timeUtil.FormatUnixTime(orderTime, "2006-01-02")
	}
	openingHours := businessSetting.OpeningHours
	if openingHours == "" {
		openingHours = "00:00-23:59" // 默认营业时间
	}

	// 将订单时间转换为日期字符串
	orderDate := timeUtil.FormatUnixTime(orderTime, "2006-01-02")

	// 计算该日期的营业开始时间
	// 先获取当天的开始时间戳（00:00:00）
	dayStartTime, err := timeUtil.FormatTimeToUnix(orderDate + " 00:00:00")
	if err != nil {
		// 如果解析失败，直接返回订单日期
		return orderDate
	}

	// 计算营业开始时间相对于当天的偏移
	startOffset, _ := timeUtil.OpeningHoursStartEndUnix(openingHours, utils.WithOpeningHoursType(1))
	// 营业开始时间 = 当天开始时间 + 营业开始时间相对于当天的偏移（秒）
	dateStartTime := dayStartTime + (startOffset % (24 * 3600))

	// 如果订单时间在营业开始时间之前，则属于前一天的营业日
	if orderTime < dateStartTime {
		// 获取前一天的日期
		prevTime := time.Unix(orderTime, 0).AddDate(0, 0, -1)
		prevDate := timeUtil.FormatUnixTime(prevTime.Unix(), "2006-01-02")
		return prevDate
	}

	return orderDate
}

// calculateMaterialCorrectionInfo 计算材料修正信息（占位实现，实际计算在后续任务中）
func (s *costCardCorrectionSrv) calculateMaterialCorrectionInfo(product *model.SaleOrderProduct, materials []*model.RelatedMaterial) []resp.MaterialCorrectionInfo {
	materialInfos := make([]resp.MaterialCorrectionInfo, 0)

	for _, material := range materials {
		if material.Material == nil {
			continue
		}

		materialInfo := resp.MaterialCorrectionInfo{
			MaterialUuid:   material.MaterialUuid,
			MaterialName:   material.Material.Name,
			OldConsumption: 0, // TODO: 将在后续任务中计算
			NewConsumption: 0, // TODO: 将在后续任务中计算
			ReturnQuantity: 0, // TODO: 将在后续任务中计算
		}

		materialInfos = append(materialInfos, materialInfo)
	}

	return materialInfos
}

// getHistoricalOutboundRecords 查询订单的历史出库记录
func (s *costCardCorrectionSrv) getHistoricalOutboundRecords(ctx context.Context, companyUuid uint64, orderUuid uint64) ([]*model.WarehouseOutFormItem, error) {
	// 查询该订单的所有出库单记录
	outFormItems, err := s.warehouseFormRepo.GetWarehouseOutFormItemBySaleOrderUuid(orderUuid)
	if err != nil {
		return nil, fmt.Errorf("查询订单出库记录失败: %w", err)
	}

	// 过滤：只返回已出库且已减库存的记录
	filteredItems := make([]*model.WarehouseOutFormItem, 0)
	for _, item := range outFormItems {
		// 只查询已出库的记录（status=1, reduce_stock=1）
		if item.Status == constant.WarehouseOutFormItemStatusSuccess &&
			item.ReduceStock == constant.WarehouseOutFormItemReduceStockSuccess &&
			item.RevokeTime == 0 { // 未撤销的记录
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

// returnMaterials 退回错误扣减的材料
func (s *costCardCorrectionSrv) returnMaterials(ctx context.Context, companyUuid uint64, orderUuid uint64, orderNo string) error {
	db := s.dbm.GetDB(companyUuid)

	// 查询历史出库记录
	outFormItems, err := s.getHistoricalOutboundRecords(ctx, companyUuid, orderUuid)
	if err != nil {
		return err
	}

	// 按材料UUID和仓库UUID汇总需要退回的数量
	materialReturnMap := make(map[string]*struct {
		MaterialUuid  uint64
		WarehouseUuid uint64
		ReturnNum     float64
		Material      *model.Material
		WarehouseItem *model.WarehouseItem
	})

	for _, item := range outFormItems {
		if item.MaterialUuid == 0 {
			continue
		}

		key := fmt.Sprintf("%d_%d", item.MaterialUuid, item.WarehouseUuid)
		if _, exists := materialReturnMap[key]; !exists {
			materialReturnMap[key] = &struct {
				MaterialUuid  uint64
				WarehouseUuid uint64
				ReturnNum     float64
				Material      *model.Material
				WarehouseItem *model.WarehouseItem
			}{
				MaterialUuid:  item.MaterialUuid,
				WarehouseUuid: item.WarehouseUuid,
				ReturnNum:     0,
			}
		}
		materialReturnMap[key].ReturnNum += item.Num
	}

	// 使用事务确保数据一致性
	return db.Transaction(func(tx *gorm.DB) error {
		warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 收集需要更新关联库存的材料UUID
		relatedMaterialUuids := make([]uint64, 0)
		relatedMaterialUuidSet := make(map[uint64]bool)

		// 遍历每个材料，退回库存
		for _, returnInfo := range materialReturnMap {
			if returnInfo.ReturnNum <= 0 {
				continue
			}

			// 获取或创建仓库物品库存记录
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterialOrCreate(
				returnInfo.WarehouseUuid,
				returnInfo.MaterialUuid,
				"", // materialCode 可以从 Material 中获取，但这里先传空
				0,  // valuation 可以从 Material 中获取，但这里先传0
			)
			if err != nil {
				return fmt.Errorf("获取仓库物品库存失败: %w", err)
			}

			// 增加库存
			err = warehouseItemRepo.AddStock(warehouseItem.Uuid, returnInfo.ReturnNum)
			if err != nil {
				return fmt.Errorf("增加材料库存失败: %w", err)
			}

			// 获取材料信息（用于记录日志）
			material, err := materialRepo.GetMaterialByUuid(returnInfo.MaterialUuid, materialRepo.WithRelatedMaterialList())
			if err != nil {
				logger.Logger.Warn("获取材料信息失败", zap.Error(err), zap.Uint64("material_uuid", returnInfo.MaterialUuid))
				// 继续执行，不中断流程
			} else {
				returnInfo.Material = &material

				// 收集需要更新关联库存的材料UUID
				relatedUuids := material.GetRelatedMaterialUuids()
				for _, uuid := range relatedUuids {
					if !relatedMaterialUuidSet[uuid] {
						relatedMaterialUuids = append(relatedMaterialUuids, uuid)
						relatedMaterialUuidSet[uuid] = true
					}
				}
			}

			// 记录入库日志（退回）
			materialName := ""
			materialBaseUnitUuid := uint64(0)
			materialBaseUnitName := ""
			if returnInfo.Material != nil {
				materialName = returnInfo.Material.Name
				// Material 模型中没有 BaseUnitUuid 和 BaseUnitName 字段
				// 可以从 RelatedMaterial 或其他地方获取，这里先设为0
				materialBaseUnitUuid = 0
				materialBaseUnitName = ""
			}

			warehouseLog := &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeIn,     // 入库（退回）
				Scene:                constant.WarehouseInOutLogSceneProfitIn, // 修正退回（使用盘盈入库场景）
				WarehouseUuid:        returnInfo.WarehouseUuid,
				MaterialUuid:         returnInfo.MaterialUuid,
				MaterialName:         materialName,
				MaterialBaseUnitUuid: materialBaseUnitUuid,
				MaterialBaseUnitName: materialBaseUnitName,
				Num:                  returnInfo.ReturnNum,
				Price:                0, // TODO: 可以从历史记录中获取价格
				Amount:               0, // TODO: 计算金额
				OrderNo:              orderNo,
			}
			err = warehouseLogRepo.Create(warehouseLog)
			if err != nil {
				return fmt.Errorf("记录入库日志失败: %w", err)
			}
		}

		// 更新规格/加料关联材料库存
		if len(relatedMaterialUuids) > 0 {
			err = materialRepo.UpdateRelatedMaterialStock(relatedMaterialUuids)
			if err != nil {
				return fmt.Errorf("更新关联材料库存失败: %w", err)
			}
		}

		return nil
	})
}

// recalculateProductInventory 重新计算商品库存
func (s *costCardCorrectionSrv) recalculateProductInventory(ctx context.Context, companyUuid uint64, materialUuids []uint64) error {
	// TODO: 实现商品库存重新计算逻辑
	// 根据成本卡计算：材料库存/材料用量（取最小值）
	// 更新所有使用该材料的成本卡关联的商品库存
	// 参考：main/app/modules/inventory/domain/service/bom_card_product_inventory_strategy.go
	// 参考：main/app/model/product.go (CalculateExpectedProductionNum)

	// 占位实现，将在后续任务中完善
	return nil
}

// deleteOldMaterialConsumptionRecords 删除订单的旧材料消耗记录
func (s *costCardCorrectionSrv) deleteOldMaterialConsumptionRecords(ctx context.Context, companyUuid uint64, orderUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 软删除：更新 delete_time
	now := time.Now().Unix()
	err := db.Model(&model.SaleOrderProductBom{}).
		Where("sale_order_uuid = ?", orderUuid).
		Where("delete_time = ?", constant.NotDeleted).
		Update("delete_time", now).Error

	if err != nil {
		return fmt.Errorf("删除旧材料消耗记录失败: %w", err)
	}

	return nil
}

// recalculateMaterialConsumption 根据正确的成本卡重新计算材料消耗量
func (s *costCardCorrectionSrv) recalculateMaterialConsumption(ctx context.Context, companyUuid uint64, order model.SaleOrder) error {
	// 获取订单中有效售出的商品
	validProducts := order.GetValidSaleOrderProductList()

	// 重新生成材料消耗记录（SaleOrderProductBom）
	// 注意：SaleOrderProductBom 记录的是规格/加料，不是材料消耗
	// 但根据 requirements.md，我们需要重新生成这些记录，确保它们关联的是正确的成本卡
	newBoms := make([]*model.SaleOrderProductBom, 0)

	for _, product := range validProducts {
		// 遍历商品的规格和加料
		for _, bom := range product.SaleOrderProductBoms {
			if bom.IsDelete() {
				continue
			}

			// 检查是否使用成本卡
			var bomCard *model.ProductBomCard
			if bom.IsFlavor() && bom.ProductBom.Uuid != 0 && bom.ProductBom.HasProductBomCard() {
				bomCard = bom.ProductBom.ProductBomCard
			} else if bom.IsSauce() && bom.ProductBom.ProductSauceUuid != 0 && bom.ProductBom.ProductSauce.HasProductBomCard() {
				bomCard = bom.ProductBom.ProductSauce.ProductBomCard
			}

			if bomCard == nil || bomCard.Uuid == 0 {
				continue
			}

			// 创建新的 SaleOrderProductBom 记录
			// 确保关联的是当前正确的 ProductBom（包含正确的成本卡）
			uuid, _ := utils.GetID()
			newBom := &model.SaleOrderProductBom{
				BaseModel: model.BaseModel{
					Uuid:       uuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				Name:                 bom.Name,
				Price:                bom.Price,
				IsFlavorBom:          bom.IsFlavorBom,
				SaleOrderUuid:        order.Uuid,
				SaleOrderProductUuid: product.Uuid,
				ProductBomUuid:       bom.ProductBomUuid,
			}
			newBoms = append(newBoms, newBom)
		}
	}

	// 批量创建新的 SaleOrderProductBom 记录
	if len(newBoms) > 0 {
		err := s.orderProductBomRepo.CreateBatch(newBoms)
		if err != nil {
			return fmt.Errorf("创建材料消耗记录失败: %w", err)
		}
	}

	return nil
}

// regenerateOutboundForm 重新生成出库记录
func (s *costCardCorrectionSrv) regenerateOutboundForm(ctx context.Context, companyUuid uint64, order model.SaleOrder, saleBillUuid uint64, staffUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 1. 获取订单中有效售出的商品（送厨的商品）
	validProducts := order.GetValidSaleOrderProductList()
	if len(validProducts) == 0 {
		return nil // 没有需要出库的商品
	}

	// 2. 根据当前正确的成本卡重新计算材料消耗，生成减库存清单
	decreaseStockList, err := s.calculateDecreaseStockList(ctx, companyUuid, validProducts)
	if err != nil {
		return fmt.Errorf("计算减库存清单失败: %w", err)
	}

	if len(decreaseStockList) == 0 {
		return nil // 没有需要出库的材料
	}

	// 3. 获取员工班次信息
	staffShiftLogUuid := uint64(0)
	staffShiftLog, err := GetCurrentStaffShiftLog(db, staffUuid)
	if err != nil {
		logger.Logger.Warn("获取员工班次信息失败，使用默认值", zap.Uint64("staffUuid", staffUuid), zap.Error(err))
	} else {
		staffShiftLogUuid = staffShiftLog.Uuid
	}

	// 4. 构建出库单（isCheckout=true 表示已结账，状态为已出库）
	warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, true, saleBillUuid, staffUuid, staffShiftLogUuid)

	// 5. 在事务中创建出库单、扣减库存、记录日志
	return repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
		warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 创建出库单和出库单明细
		for _, warehouseOutForm := range warehouseOutForms {
			if len(warehouseOutForm.WarehouseOutFormItems) == 0 {
				continue
			}

			// 创建出库单
			if err := warehouseFormRepo.CreateWarehouseOutFormRecord(*warehouseOutForm); err != nil {
				return fmt.Errorf("创建出库单失败: %w", err)
			}

			// 创建出库单明细
			if err := warehouseFormRepo.CreateWarehouseOutFormItemRecords(warehouseOutForm.WarehouseOutFormItems); err != nil {
				return fmt.Errorf("创建出库单明细失败: %w", err)
			}

			// 处理每个出库单明细：扣减库存、记录日志、更新关联材料库存
			for _, item := range warehouseOutForm.WarehouseOutFormItems {
				// 只处理材料出库（MaterialUuid != 0）
				if item.MaterialUuid == 0 {
					continue
				}

				// 获取材料信息
				material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
				if err != nil {
					return fmt.Errorf("获取材料信息失败: %w", err)
				}
				if material.Uuid == 0 {
					continue
				}

				// 获取材料的基准单位
				baseUnitUuid := uint64(0)
				baseUnitName := ""
				baseUnit := material.GetBaseUnit()
				if baseUnit != nil {
					baseUnitUuid = baseUnit.Uuid
					if baseUnit.Unit != nil {
						baseUnitName = baseUnit.Unit.MultiLanguageName.ToJson()
					}
				}

				// 获取或创建 WarehouseItem
				warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterialOrCreate(
					item.WarehouseUuid,
					item.MaterialUuid,
					material.Code,
					material.Valuation,
				)
				if err != nil {
					return fmt.Errorf("获取仓库物品失败: %w", err)
				}

				// 扣减库存
				if err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, item.Num); err != nil {
					return fmt.Errorf("扣减库存失败: %w", err)
				}

				// 记录出库日志（使用销售出库场景）
				warehouseLog := &model.WarehouseInOutLog{
					LogType:              constant.WarehouseInOutLogLogTypeOut,
					Scene:                constant.WarehouseInOutLogSceneSale, // 销售出库
					WarehouseUuid:        item.WarehouseUuid,
					MaterialUuid:         item.MaterialUuid,
					MaterialName:         material.Name,
					MaterialBaseUnitUuid: baseUnitUuid,
					MaterialBaseUnitName: baseUnitName,
					Num:                  item.Num,
					OrderNo:              order.OrderNo,
				}
				if err := warehouseLogRepo.Create(warehouseLog); err != nil {
					return fmt.Errorf("记录出库日志失败: %w", err)
				}

				// 更新关联材料库存
				relatedMaterialUuids := material.GetRelatedMaterialUuids()
				if len(relatedMaterialUuids) > 0 {
					if err := materialRepo.UpdateRelatedMaterialStock(relatedMaterialUuids); err != nil {
						return fmt.Errorf("更新关联材料库存失败: %w", err)
					}
				}
			}
		}

		// 6. 重新计算商品库存（出库后）
		materialUuids := make([]uint64, 0)
		for _, product := range decreaseStockList {
			for _, material := range product.ProductBomMaterials {
				materialUuids = append(materialUuids, material.MaterialUuid)
			}
		}
		if len(materialUuids) > 0 {
			if err := s.recalculateProductInventory(ctx, companyUuid, materialUuids); err != nil {
				logger.Logger.Warn("重新计算商品库存失败", zap.Error(err))
				// 不返回错误，因为这不是关键步骤
			}
		}

		return nil
	})
}

// calculateDecreaseStockList 根据当前正确的成本卡计算减库存清单
func (s *costCardCorrectionSrv) calculateDecreaseStockList(ctx context.Context, companyUuid uint64, products []*model.SaleOrderProduct) (model.ProductList, error) {
	list := make(model.ProductList, 0)

	for _, product := range products {
		for _, saleOrderProductBom := range product.SaleOrderProductBoms {
			if saleOrderProductBom.IsDelete() {
				continue
			}

			// 获取原材料的出库数量
			productBomMaterials := make([]*model.ProductBomMaterials, 0)

			// 如果是规格商品
			if saleOrderProductBom.IsFlavor() {
				var flavorMaterials []*model.RelatedMaterial
				// 如果有成本卡，则使用成本卡的原材料
				if saleOrderProductBom.ProductBom.HasProductBomCard() {
					flavorMaterials = saleOrderProductBom.ProductBom.ProductBomCard.RelatedMaterials
				} else {
					// 如果没有成本卡，则使用规格商品的原材料
					flavorMaterials = saleOrderProductBom.ProductBom.FlavorMaterials
				}

				// 遍历原材料
				for _, productBomMaterial := range flavorMaterials {
					if productBomMaterial.IsDelete() {
						continue
					}
					// 防止 Material 为空
					if productBomMaterial.Material == nil {
						continue
					}
					// 如果材料被禁用，则跳过，不扣减库存
					if productBomMaterial.Material.Status == false {
						continue
					}
					if num := productBomMaterial.GetDecreaseNum(product.Num); num > 0 {
						productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
							MaterialUuid:  productBomMaterial.MaterialUuid,
							WarehouseUuid: productBomMaterial.Material.WarehouseUuid,
							Num:           num,
							SaleOrderUuid: product.SaleOrderUuid,
						})
					}
				}
			}

			// 如果是小料
			if saleOrderProductBom.IsSauce() {
				var sauceMaterials []*model.RelatedMaterial
				// 如果有成本卡，则使用成本卡的原材料
				if saleOrderProductBom.ProductBom.ProductSauce.HasProductBomCard() {
					sauceMaterials = saleOrderProductBom.ProductBom.ProductSauce.ProductBomCard.RelatedMaterials
				} else {
					// 如果没有成本卡，则使用小料的原材料
					sauceMaterials = saleOrderProductBom.ProductBom.ProductSauce.SauceMaterials
				}

				// 遍历原材料
				for _, material := range sauceMaterials {
					// 防止 Material 为空
					if material.Material == nil {
						continue
					}
					// 如果材料被禁用，则跳过，不扣减库存
					if material.Material.Status == false {
						continue
					}
					if num := material.GetDecreaseNum(product.Num); num > 0 {
						productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
							MaterialUuid:  material.MaterialUuid,
							WarehouseUuid: material.Material.WarehouseUuid,
							Num:           num,
							SaleOrderUuid: product.SaleOrderUuid,
						})
					}
				}
			}

			// 获取规格商品的出库数量
			if product.Num > 0 {
				list = append(list, &model.Product{
					ProductBomUuid:       saleOrderProductBom.ProductBomUuid,
					PackageUuid:          product.PackageUuid,
					SaleOrderProductUuid: product.Uuid,
					SaleOrderUuid:        product.SaleOrderUuid,
					Num:                  product.Num,
					ProductBomMaterials:  productBomMaterials,
				})
			}
		}
	}

	return list, nil
}

// resyncErpData 重新同步 ERP 数据
func (s *costCardCorrectionSrv) resyncErpData(ctx context.Context, companyUuid uint64, order model.SaleOrder, saleBillUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 检查是否开启 ERP
	companyRepo := repository.NewCompanyRepo(db)
	company, err := companyRepo.GetCompany()
	if err != nil {
		return fmt.Errorf("获取公司信息失败: %w", err)
	}
	if !company.IsOpenErpPhase3() {
		logger.Logger.Info("公司未开启 ERP，跳过 ERP 数据同步", zap.Uint64("companyUuid", companyUuid))
		return nil // 未开启 ERP，跳过
	}

	// 查询订单关联的账单
	saleBillRepo := repository.NewSaleBillRepo(db)
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return fmt.Errorf("查询账单失败: %w", err)
	}

	// 创建临时的 orderSrv 实例来调用 SavePosInvoice
	// 注意：这里需要创建必要的依赖服务
	localeSrv := NewLocaleSrv()
	settingSrv := setting.NewSrvImpl(s.dbm, cache.Global)
	mustPlanSrv := NewMustPlanSrv(s.dbm)
	paymentMethodSrv := NewPaymentMethodSrv(s.dbm, settingSrv)
	memberSrv := NewMemberSrv(s.dbm, cache.Global)
	cashBoxSrv := NewCashBoxSrv(s.dbm)
	orderSrvInterface := NewOrderSrv(s.dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv)

	// 类型断言为 orderSrv 以访问 SavePosInvoice 方法
	// 注意：SavePosInvoice 不是接口方法，需要直接调用实现
	orderSrvImpl, ok := orderSrvInterface.(*orderSrv)
	if !ok {
		return fmt.Errorf("无法获取 orderSrv 实现")
	}

	// 重新生成 POS Invoice 数据并同步到 ERP
	// 注意：SavePosInvoice 会使用订单的 GetErpProductBomMaterials 方法获取材料消耗
	// 该方法会使用当前订单的 SaleOrderProductBoms，这些记录已经通过 recalculateMaterialConsumption 更新
	res, err := orderSrvImpl.SavePosInvoice(ctx, &order, saleBill, db)
	if err != nil {
		// ERP 同步失败，记录错误但不中断整个修正流程
		logger.Logger.Error("ERP 数据同步失败",
			zap.Uint64("orderUuid", order.Uuid),
			zap.String("orderNo", order.OrderNo),
			zap.Error(err))
		return fmt.Errorf("ERP 数据同步失败: %w", err)
	}

	// 更新订单的 ERP invoice 名称
	order.ErpProductsInvoiceName = res.ProductsInvoiceName
	order.ErpMaterialInvoiceName = res.MaterialInvoiceName
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	if err := saleOrderRepo.UpdateSaleOrderRecord(order); err != nil {
		logger.Logger.Error("更新订单ERP信息失败",
			zap.Uint64("orderUuid", order.Uuid),
			zap.Error(err))
		return fmt.Errorf("更新订单ERP信息失败: %w", err)
	}

	logger.Logger.Info("ERP 数据同步成功",
		zap.Uint64("orderUuid", order.Uuid),
		zap.String("orderNo", order.OrderNo),
		zap.String("productsInvoiceName", res.ProductsInvoiceName),
		zap.String("materialInvoiceName", res.MaterialInvoiceName))

	return nil
}

// identifyAffectedDates 识别受影响的日期范围
// 只能使用已完成订单的完成时间（FinishTime > 0）
func (s *costCardCorrectionSrv) identifyAffectedDates(ctx context.Context, companyUuid uint64, orders []model.SaleOrder) ([]string, error) {
	// 收集所有订单的营业日期
	dateSet := make(map[string]bool)
	for _, order := range orders {
		// 只处理已完成订单的完成时间
		if order.FinishTime > 0 {
			businessDate := s.getBusinessDateFromOrderTime(ctx, companyUuid, order.FinishTime)
			if businessDate != "" {
				dateSet[businessDate] = true
			}
		}
	}

	// 转换为排序后的日期列表
	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	// 使用 utils 的排序功能
	for i := 0; i < len(dates)-1; i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	return dates, nil
}

// recalculateDailySalesOutbound 重新统计每日销售出库记录
func (s *costCardCorrectionSrv) recalculateDailySalesOutbound(ctx context.Context, companyUuid uint64, businessDate string) error {
	db := s.dbm.GetDB(companyUuid)

	// 获取公司设置（时区和营业时间）
	companyRepo := repository.NewCompanyRepo(db)
	company, err := companyRepo.GetCompany()
	if err != nil {
		return fmt.Errorf("获取公司信息失败: %w", err)
	}

	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}
	timeUtil := utils.Timezone(timezone)

	// 获取营业时间
	settingSrv := setting.NewSrvImpl(s.dbm, cache.Global)
	settingCtx := context.NewContext()
	settingCtx.SetCompanyUuid(companyUuid)
	businessSetting, err := settingSrv.GetBusinessSetting(settingCtx)
	if err != nil {
		return fmt.Errorf("获取营业时间设置失败: %w", err)
	}
	openingHours := businessSetting.OpeningHours
	if openingHours == "" {
		openingHours = "00:00-23:59" // 默认营业时间
	}

	// 计算该营业日的开始和结束时间
	// 先获取当天的开始时间戳（00:00:00）
	dayStart, err := timeUtil.FormatTimeToUnix(businessDate + " 00:00:00")
	if err != nil {
		return fmt.Errorf("解析业务日期失败: %w", err)
	}
	dayEnd, err := timeUtil.FormatTimeToUnix(businessDate + " 23:59:59")
	if err != nil {
		return fmt.Errorf("解析业务日期结束时间失败: %w", err)
	}
	// 计算营业时间的开始和结束时间戳（相对于当天的偏移）
	startOffset, endOffset := timeUtil.OpeningHoursStartEndUnix(openingHours, utils.WithOpeningHoursType(1))
	// 营业开始时间 = 当天开始时间 + 营业开始时间相对于当天的偏移（秒）
	startTime := dayStart + (startOffset % (24 * 3600))
	// 营业结束时间 = 当天开始时间 + 营业结束时间相对于当天的偏移（秒）
	endTime := dayStart + (endOffset % (24 * 3600))
	// 如果营业结束时间小于开始时间，说明跨天了，结束时间应该是第二天的结束时间
	if endTime < startTime {
		endTime = dayEnd
	}

	// 生成营业时段标识（格式：YYYYMMDD HH:MM-HH:MM）
	openingYearHours := fmt.Sprintf("%s %s", strings.ReplaceAll(businessDate, "-", ""), openingHours)

	// 查询该营业日期的所有 SaleOrderMaterial 记录（包括已统计和未统计的）
	// 注意：需要修改查询条件，不限制 is_summarized，因为我们要重新统计
	// 直接查询数据库，不限制 is_summarized
	var saleOrderMaterials []*model.SaleOrderMaterial
	err = db.Model(&model.SaleOrderMaterial{}).
		Preload("Material.Unit").
		Where("create_time BETWEEN ? AND ? AND delete_time = 0", startTime, endTime).
		Find(&saleOrderMaterials).Error
	if err != nil {
		return fmt.Errorf("查询销售订单原料记录失败: %w", err)
	}

	// 按仓库和物料分组汇总数量
	recordMap := make(map[string]*OutboundRecord)
	for _, item := range saleOrderMaterials {
		key := fmt.Sprintf("%d_%d", item.WarehouseUuid, item.MaterialUuid)
		if record, exists := recordMap[key]; exists {
			record.TotalNum += item.Num
			// 收集所有相关的 SaleOrderMaterial UUID（用于后续更新 is_summarized 状态）
			record.SaleOrderMaterialUuids = append(record.SaleOrderMaterialUuids, item.Uuid)
		} else {
			materialName := ""
			materialBaseUnitUuid := uint64(0)
			materialBaseUnitName := ""
			valuation := 0.0
			supplierUuid := uint64(0)

			// 从关联的物料信息中获取数据
			if item.Material != nil {
				materialName = item.Material.Name
				materialBaseUnitUuid = item.Material.UnitUuid
				valuation = item.Material.Valuation
				supplierUuid = item.Material.SupplierUuid

				// 获取单位名称
				if item.Material.Unit != nil && item.Material.Unit.Unit != nil {
					materialBaseUnitName = item.Material.Unit.Unit.MultiLanguageName.ToJson()
				}
			}

			recordMap[key] = &OutboundRecord{
				Uuid:                   item.Uuid,
				WarehouseUuid:          item.WarehouseUuid,
				MaterialUuid:           item.MaterialUuid,
				TotalNum:               item.Num,
				Valuation:              valuation,
				SupplierUuid:           supplierUuid,
				MaterialName:           materialName,
				MaterialBaseUnitUuid:   materialBaseUnitUuid,
				MaterialBaseUnitName:   materialBaseUnitName,
				SaleOrderMaterialUuids: []uint64{item.Uuid},
			}
		}
	}

	if len(recordMap) == 0 {
		logger.Logger.Info("该营业日期无销售出库记录", zap.String("businessDate", businessDate))
		return nil
	}

	// 转换为切片
	records := make([]*OutboundRecord, 0, len(recordMap))
	for _, record := range recordMap {
		records = append(records, record)
	}

	// 删除该营业日期的旧汇总记录，并重新生成
	return s.saveOutboundSummaryRecords(ctx, companyUuid, records, openingYearHours, businessDate)
}

// OutboundRecord 出库记录汇总（与 daily_sales_outbound_summary.go 中的结构一致）
type OutboundRecord struct {
	Uuid                   uint64   `json:"uuid"`
	WarehouseUuid          uint64   `json:"warehouse_uuid"`
	MaterialUuid           uint64   `json:"material_uuid"`
	TotalNum               float64  `json:"total_num"`
	Valuation              float64  `json:"valuation"` // 估值率
	SupplierUuid           uint64   `json:"supplier_uuid"`
	MaterialName           string   `json:"material_name"`
	MaterialBaseUnitUuid   uint64   `json:"material_base_unit_uuid"`
	MaterialBaseUnitName   string   `json:"material_base_unit_name"`
	SaleOrderMaterialUuids []uint64 `json:"sale_order_material_uuids"` // 关联的 SaleOrderMaterial UUID 列表
}

// saveOutboundSummaryRecords 保存出库汇总记录到 ttpos_warehouse_in_out_log 表
func (s *costCardCorrectionSrv) saveOutboundSummaryRecords(ctx context.Context, companyUuid uint64, records []*OutboundRecord, openingYearHours string, businessDate string) error {
	db := s.dbm.GetDB(companyUuid)

	// 生成出库单号
	orderNo, err := s.generateOutboundOrderNo(companyUuid, businessDate)
	if err != nil {
		return fmt.Errorf("生成出库单号失败: %w", err)
	}

	// 收集所有需要更新 is_summarized 状态的 SaleOrderMaterial UUID
	allSaleOrderMaterialUuids := make([]uint64, 0)
	for _, record := range records {
		allSaleOrderMaterialUuids = append(allSaleOrderMaterialUuids, record.SaleOrderMaterialUuids...)
	}

	return repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)

		// 删除该营业日期的旧汇总记录
		opts := []repository.DBOption{
			warehouseLogRepo.WhereLogType(constant.WarehouseInOutLogLogTypeOut), // 出库
			warehouseLogRepo.WhereScene(constant.WarehouseInOutLogSceneSale),    // 销售出库
			func(db *gorm.DB) *gorm.DB {
				return db.Where("opening_hours = ?", openingYearHours)
			},
		}
		oldLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
		if err != nil {
			logger.Logger.Warn("查询旧汇总记录失败，继续执行", zap.Error(err))
		} else {
			// 软删除旧记录
			for _, log := range oldLogs {
				if err := warehouseLogRepo.Delete(log.Uuid); err != nil {
					logger.Logger.Warn("删除旧汇总记录失败", zap.Uint64("uuid", log.Uuid), zap.Error(err))
				}
			}
		}

		// 创建新的汇总记录
		for _, record := range records {
			amount := record.TotalNum * record.Valuation
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
				Amount:               amount,
				SupplierUuid:         record.SupplierUuid,
				OrderNo:              orderNo,
				OpeningHours:         openingYearHours,
			}

			if err := warehouseLogRepo.Create(logRecord); err != nil {
				logger.Logger.Error("保存出库汇总记录失败",
					zap.Uint64("warehouse_uuid", record.WarehouseUuid),
					zap.Uint64("material_uuid", record.MaterialUuid),
					zap.Error(err))
				return fmt.Errorf("保存出库汇总记录失败: %w", err)
			}
		}

		// 更新销售订单原料的统计状态
		if len(allSaleOrderMaterialUuids) > 0 {
			saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(tx)
			if err := saleOrderMaterialRepo.UpdateSaleOrderMaterialIsSummarized(allSaleOrderMaterialUuids); err != nil {
				return fmt.Errorf("更新销售订单原料统计状态失败: %w", err)
			}
		}

		return nil
	})
}

// generateOutboundOrderNo 生成出库单号，格式：SSCK + YYYYMMDD + 4位序号
func (s *costCardCorrectionSrv) generateOutboundOrderNo(companyUuid uint64, businessDate string) (string, error) {
	db := s.dbm.GetDB(companyUuid)

	// 将业务日期转换为 YYYYMMDD 格式
	dateStr := strings.ReplaceAll(businessDate, "-", "")
	if len(dateStr) != 8 {
		return "", fmt.Errorf("业务日期格式错误: %s", businessDate)
	}

	// 使用 repository 方法查询该日期已有的 SSCK 开头的出库单号
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)

	// 计算该日期的开始和结束时间戳
	timeUtil := utils.Timezone("Asia/Shanghai")
	startTime, err := timeUtil.FormatTimeToUnix(businessDate + " 00:00:00")
	if err != nil {
		return "", fmt.Errorf("解析业务日期开始时间失败: %w", err)
	}
	endTime, err := timeUtil.FormatTimeToUnix(businessDate + " 23:59:59")
	if err != nil {
		return "", fmt.Errorf("解析业务日期结束时间失败: %w", err)
	}

	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(constant.WarehouseInOutLogLogTypeOut), // 出库
		warehouseLogRepo.WhereScene(constant.WarehouseInOutLogSceneSale),    // 销售出库
		warehouseLogRepo.WhereCreateTimeBetween(int(startTime), int(endTime)),
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return "", fmt.Errorf("查询现有出库单号失败: %w", err)
	}

	// 解析最大序号
	sequence := 1
	prefix := "SSCK" + dateStr
	for _, log := range existingLogs {
		if len(log.OrderNo) >= 16 && strings.HasPrefix(log.OrderNo, prefix) {
			seqStr := log.OrderNo[12:16] // 取最后4位作为序号
			if seq, err := strconv.Atoi(seqStr); err == nil && seq >= sequence {
				sequence = seq + 1
			}
		}
	}

	// 生成4位序号，不足补0
	sequenceStr := fmt.Sprintf("%04d", sequence)

	return prefix + sequenceStr, nil
}
