package service

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IStockReconciliationSrv 盘点单服务接口
type IStockReconciliationSrv interface {
	GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error)       // 获取盘点单列表
	GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error) // 获取盘点单详情
	CreateStockReconciliation(ctx context.Context, req req.StockReconciliationCreateReq) error                                          // 创建盘点单
	SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) error                                              // 更新盘点单
	DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error                                          // 删除盘点单
	ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) error                                        // 审核盘点单
	RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error                                          // 驳回盘点单
}

// stockReconciliationSrv 盘点单服务实现
type stockReconciliationSrv struct {
	dbm  *database.DBManager
	lock lock.Lock
}

// NewStockReconciliationSrv 创建盘点单服务
func NewStockReconciliationSrv(dbm *database.DBManager) IStockReconciliationSrv {
	return NewStockReconciliationSrvImpl(dbm)
}

// NewStockReconciliationSrvImpl 创建盘点单服务实现
func NewStockReconciliationSrvImpl(dbm *database.DBManager) IStockReconciliationSrv {
	return &stockReconciliationSrv{
		dbm:  dbm,
		lock: lock.NewSystemLock(),
	}
}

// GetStockReconciliationList 获取盘点单列表
func (s *stockReconciliationSrv) GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 构建查询选项
	var opts []repository.DBOption

	// 多仓库筛选
	if len(req.WarehouseUuids) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereWarehouseUuids(req.WarehouseUuids))
	}

	// 关键字搜索（单据编号和ERP盘点单号）
	if req.Keyword != "" {
		opts = append(opts, stockReconciliationRepo.WhereKeyword(req.Keyword))
	}

	// 创建时间范围筛选
	if req.StartCreateTime > 0 || req.EndCreateTime > 0 {
		opts = append(opts, stockReconciliationRepo.WhereCreateTimeRange(req.StartCreateTime, req.EndCreateTime))
	}

	// 状态列表筛选
	if len(req.StatusIn) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereStatusIn(req.StatusIn))
	}

	// 查询数据
	list, total, err := stockReconciliationRepo.GetStockReconciliationListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.StockReconciliationInfo, 0, len(list))
	for _, item := range list {
		info := &resp.StockReconciliationInfo{}
		if err := copier.Copy(info, item); err != nil {
			continue
		}
		listResp = append(listResp, info)
	}

	return resp.StockReconciliationListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetStockReconciliationDetail 获取盘点单详情
func (s *stockReconciliationSrv) GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return resp.StockReconciliationDetailResp{}, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return resp.StockReconciliationDetailResp{}, errors.New("盘点单不存在")
	}

	// 查询盘点单物品明细
	items, err := stockReconciliationRepo.GetStockReconciliationItemListByReconciliationUuid(req.Uuid)
	if err != nil {
		return resp.StockReconciliationDetailResp{}, errors.WithMessage(err, "查询盘点单物品明细失败")
	}

	// 转换响应数据
	var detailResp resp.StockReconciliationDetailResp
	if err := copier.Copy(&detailResp, stockReconciliation); err != nil {
		return resp.StockReconciliationDetailResp{}, errors.WithMessage(err, "转换盘点单数据失败")
	}

	// 查询物品单位明细
	materialRepo := repository.NewMaterialRepo(db)
	itemsResp := make([]*resp.StockReconciliationItemInfo, 0, len(items))
	for _, item := range items {
		itemInfo := &resp.StockReconciliationItemInfo{}
		if err := copier.Copy(itemInfo, item); err != nil {
			continue
		}

		// 查询物品信息
		material, _ := materialRepo.GetMaterialByUuid(item.MaterialUuid)
		if material.Uuid > 0 {
			itemInfo.LocaleName = *language.JsonToLocaleResponse(material.Name)
			itemInfo.MaterialCode = material.Code
		}

		// 查询单位明细
		units, err := stockReconciliationRepo.GetStockReconciliationItemUnitListByItemUuid(item.Uuid)
		if err == nil && len(units) > 0 {
			unitsResp := make([]*resp.StockReconciliationItemUnitInfo, 0, len(units))
			for _, unit := range units {
				unitInfo := &resp.StockReconciliationItemUnitInfo{}
				if err := copier.Copy(unitInfo, unit); err != nil {
					continue
				}
				unitInfo.LocaleName = unit.MaterialUnit.Unit.MultiLanguageName.GetNames()
				unitsResp = append(unitsResp, unitInfo)
			}
			itemInfo.Units = unitsResp
		}

		// 盘盈盘亏状态
		if item.CountedQuantity.GreaterThan(item.BookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusProfit
		} else if item.CountedQuantity.LessThan(item.BookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusLoss
		} else {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusNormal
		}
		// 是否盘盈盘亏异常（账面和实盘数量差值的绝对值大于20%）
		itemInfo.IsInventoryStatusException = item.CountedQuantity.Sub(item.BookedQuantity).Abs().Div(item.BookedQuantity).GreaterThan(decimal.NewFromFloat(0.2))

		itemsResp = append(itemsResp, itemInfo)
	}
	detailResp.Items = itemsResp

	return detailResp, nil
}

// CreateStockReconciliation 创建盘点单
func (s *stockReconciliationSrv) CreateStockReconciliation(ctx context.Context, req req.StockReconciliationCreateReq) error {
	// 加锁保证单号唯一性（基于公司UUID和日期）
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
	lockKey := fmt.Sprintf("stock_reconciliation_%d_%s", ctx.GetCompanyUuid(), dateStr)
	s.lock.LockUuidString(lockKey)
	defer s.lock.UnlockUuidString(lockKey)

	db := ctx.GetDB()
	warehouseItems, materials, err := s.validateWarehouseAndItems(db, req.WarehouseUuid, req.Items)
	if err != nil {
		return err
	}

	materialUuidStockMap := map[uint64]float64{}
	for _, warehouseItem := range warehouseItems {
		materialUuidStockMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
	}

	materialUnitMap := make(map[uint64]map[uint64]float64)
	for _, material := range materials {
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
		}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 在事务内部生成单据编号
		orderNo := s.generateOrderNo(tx, timezone)

		// 创建盘点单
		stockReconciliation := &model.StockReconciliation{
			OrderNo:       orderNo,
			Type:          req.Type,
			WarehouseUuid: req.WarehouseUuid,
			Purpose:       req.Purpose,
			Status:        constant.StockReconciliationStatusSaved, // 0-已保存
		}

		if err := stockReconciliationRepo.CreateStockReconciliation(stockReconciliation); err != nil {
			return errors.WithMessage(errors.New("创建盘点单失败"), err.Error())
		}

		// 创建盘点单物品明细

		var stockReconciliationItemUnits []*model.StockReconciliationItemUnit
		for _, reqItem := range req.Items {
			// 计算实盘数量（基准单位）
			countedQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitItem := range reqItem.Units {
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitItem.MaterialUnitUuid]
					unitQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					countedQuantity = countedQuantity.Add(unitQuantity)
				}
			}
			countedQuantity = countedQuantity.Truncate(3)

			item := &model.StockReconciliationItem{
				StockReconciliationUuid: stockReconciliation.Uuid,
				MaterialUuid:            reqItem.MaterialUuid,
				BookedQuantity:          decimal.NewFromFloat(materialUuidStockMap[reqItem.MaterialUuid]),
				CountedQuantity:         countedQuantity,
			}

			// 先创建item以获取自动生成的uuid
			if err := stockReconciliationRepo.CreateStockReconciliationItem(item); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品明细失败"), err.Error())
			}

			// 创建单位明细
			for _, unitItem := range reqItem.Units {
				stockReconciliationItemUnits = append(stockReconciliationItemUnits, &model.StockReconciliationItemUnit{
					StockReconciliationItemUuid: item.Uuid,
					MaterialUnitUuid:            unitItem.MaterialUnitUuid,
					Quantity:                    &unitItem.Quantity,
				})

			}
		}

		if err := stockReconciliationRepo.CreateStockReconciliationItemUnitBatch(stockReconciliationItemUnits); err != nil {
			return errors.WithMessage(errors.New("创建盘点单物品单位明细失败"), err.Error())
		}

		return nil
	})

	return err
}

// SaveStockReconciliation 保存盘点单
func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已保存状态的盘点单才能修改
	if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
		return errors.New("当前状态不允许修改")
	}

	// 验证仓库和物品明细
	warehouseItems, materials, err := s.validateWarehouseAndItems(db, req.WarehouseUuid, req.Items)
	if err != nil {
		return err
	}

	materialUuidStockMap := map[uint64]float64{}
	for _, warehouseItem := range warehouseItems {
		materialUuidStockMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
	}

	materialUnitMap := make(map[uint64]map[uint64]float64)
	for _, material := range materials {
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
		}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 更新盘点单
		stockReconciliation.WarehouseUuid = req.WarehouseUuid
		stockReconciliation.Purpose = req.Purpose
		stockReconciliation.Type = req.Type

		if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
			return errors.WithMessage(err, "更新盘点单失败")
		}
		// 删除原有的物品明细
		if err := stockReconciliationRepo.DeleteStockReconciliationItemByReconciliationUuid(req.Uuid); err != nil {
			return errors.WithMessage(err, "删除盘点单物品明细失败")
		}
		// 删除原有物品单位明细
		if err := stockReconciliationRepo.DeleteStockReconciliationItemUnitByReconciliationUuid(req.Uuid); err != nil {
			return errors.WithMessage(err, "删除盘点单物品单位明细失败")
		}

		var stockReconciliationItemUnits []*model.StockReconciliationItemUnit
		for _, reqItem := range req.Items {
			// 计算实盘数量（基准单位）
			countedQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitItem := range reqItem.Units {
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitItem.MaterialUnitUuid]
					unitQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					countedQuantity = countedQuantity.Add(unitQuantity)
				}
			}
			countedQuantity = countedQuantity.Truncate(3)

			item := &model.StockReconciliationItem{
				StockReconciliationUuid: stockReconciliation.Uuid,
				MaterialUuid:            reqItem.MaterialUuid,
				BookedQuantity:          decimal.NewFromFloat(materialUuidStockMap[reqItem.MaterialUuid]),
				CountedQuantity:         countedQuantity,
			}

			// 先创建item以获取自动生成的uuid
			if err := stockReconciliationRepo.CreateStockReconciliationItem(item); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品明细失败"), err.Error())
			}

			// 创建单位明细
			for _, unitItem := range reqItem.Units {
				stockReconciliationItemUnits = append(stockReconciliationItemUnits, &model.StockReconciliationItemUnit{
					StockReconciliationItemUuid: item.Uuid,
					MaterialUnitUuid:            unitItem.MaterialUnitUuid,
					Quantity:                    &unitItem.Quantity,
				})
			}
		}

		if err := stockReconciliationRepo.CreateStockReconciliationItemUnitBatch(stockReconciliationItemUnits); err != nil {
			return errors.WithMessage(errors.New("创建盘点单物品单位明细失败"), err.Error())
		}

		return nil
	})

	if err != nil {
		errMsg := "保存盘点单失败"
		if req.IsSubmit {
			errMsg = "提交盘点单失败"
		}
		return errors.WithMessage(errors.New(errMsg), err.Error())
	}

	// 提交盘点单
	if req.IsSubmit && ctx.GetCompany().IsOpenErp() {
		// TODO 组合数据提交erp盘点单
		// return s.ApproveStockReconciliation(ctx, req.Uuid)

		// 更新盘点单erp_code
	}

	return nil
}

// DeleteStockReconciliation 删除盘点单
func (s *stockReconciliationSrv) DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已保存和已驳回状态的盘点单才能删除
	if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
		return errors.New("当前状态不允许删除")
	}

	// 删除盘点单以及物品明细和单位明细
	if err := stockReconciliationRepo.DeleteStockReconciliation(req.Uuid); err != nil {
		return errors.WithMessage(err, "删除盘点单失败")
	}

	// 清理UUID锁资源
	s.lock.ClearUuidLock(req.Uuid)

	return nil
}

// ApproveStockReconciliation 审核盘点单
func (s *stockReconciliationSrv) ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)
	// 查询盘点单
	stockReconciliation, err := repository.NewStockReconciliationRepo(db).GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能审核
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return errors.New("当前状态不允许审核")
	}

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 更新盘点单状态为已审核
		updateData := map[string]any{
			"status":      constant.StockReconciliationStatusApproved, // 2-已审核
			"update_time": int(time.Now().Unix()),
		}
		if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
			return errors.WithMessage(err, "审核盘点单失败")
		}

		// TODO: 审核通过后，需要更新库存
		// 需要调用库存服务更新库存数据
		// 增加盘盈盘亏出入库记录

		return nil
	})
}

// RejectStockReconciliation 驳回盘点单
func (s *stockReconciliationSrv) RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能驳回
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return errors.New("盘点单状态不允许驳回")
	}

	if ctx.GetCompany().IsOpenErp() {
		// 调用erp接口拒绝
		// return errors.New("当前公司未开通ERP，无法驳回盘点单")
	}

	// 更新盘点单状态为已驳回
	updateData := map[string]any{
		"status":      constant.StockReconciliationStatusRejected, // 3-已驳回
		"update_time": int(time.Now().Unix()),
	}
	if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
		return errors.WithMessage(err, "驳回盘点单失败")
	}

	return nil
}

// generateOrderNo 生成单据编号（必须在事务内部调用）
func (s *stockReconciliationSrv) generateOrderNo(db *gorm.DB, timezone string) string {
	// 生成格式：ST + 年月日 + 0000（4位序列号）
	// 例如：ST202510160001
	// 序列号从0001开始递增，每天重置
	repo := repository.NewStockReconciliationRepo(db)

	// 使用商家时区格式化日期
	dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
	prefix := fmt.Sprintf("ST%s", dateStr)

	// 查询当天最大的订单号
	maxOrderNo, err := repo.GetMaxOrderNoByPrefix(prefix)
	if err != nil {
		logger.Logger.Error("查询最大单据编号失败", zap.Error(err))
		// 如果查询失败，返回第一个序列号
		return fmt.Sprintf("%s0001", prefix)
	}

	// 如果没有找到当天的订单号，从0001开始
	if maxOrderNo == "" {
		return fmt.Sprintf("%s0001", prefix)
	}

	// 从订单号中提取序列号（最后4位）
	if len(maxOrderNo) < 4 {
		return fmt.Sprintf("%s0001", prefix)
	}

	// 获取序列号部分
	seqStr := maxOrderNo[len(maxOrderNo)-4:]
	seq := 0
	fmt.Sscanf(seqStr, "%d", &seq)

	// 序列号+1
	seq++

	// 生成新的订单号
	return fmt.Sprintf("%s%04d", prefix, seq)
}

// validateWarehouseAndItems 验证仓库和物品明细
// 验证内容：
// 1. 仓库是否存在且类型为 normal
// 2. 物品是否属于该仓库
// 3. 物品单位是否正确
func (s *stockReconciliationSrv) validateWarehouseAndItems(db *gorm.DB, warehouseUuid uint64, items []*req.StockReconciliationItemReq) ([]model.WarehouseItem, []*model.Material, error) {
	// 判断仓库是否存在，且类型为normal
	warehouseRepo := repository.NewWarehouseRepo(db)
	warehouse, err := warehouseRepo.GetByUuid(warehouseUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(errors.New("查询仓库失败"), err.Error())
	}
	if warehouse == nil || warehouse.Type != constant.WarehouseTypeNormal {
		return nil, nil, errors.New("仓库参数错误")
	}

	// 判断盘点物品明细列表是否正确，要求所有物品均为仓库内的物品，且单位均为仓库内物品的单位
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	// 获取仓库Uuid获取仓库物品信息列表
	warehouseItems, err := warehouseItemRepo.GetByWarehouseUuid(warehouseUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}

	if len(warehouseItems) == 0 {
		return nil, nil, errors.New("仓库无物品")
	}

	// 获取所有物品UUID
	materialUuids := make([]uint64, 0, len(warehouseItems))
	for _, item := range warehouseItems {
		materialUuids = append(materialUuids, item.MaterialUuid)
	}

	// 批量查询物品详情
	materialRepo := repository.NewMaterialRepo(db)
	var materials []*model.Material
	materialMap := make(map[uint64]*model.Material)
	if len(materialUuids) > 0 {
		materials, err = materialRepo.GetMaterialDetailByUuids(materialUuids)
		if err != nil {
			return nil, nil, errors.WithMessage(errors.New("查询物品详情失败"), err.Error())
		}
		for _, material := range materials {
			materialMap[material.Uuid] = material
		}
	}

	// 验证请求中的物品和单位
	for _, item := range items {
		material, exists := materialMap[item.MaterialUuid]
		if !exists {
			return nil, nil, errors.New("物品参数错误")
		}

		// 验证物品单位
		for _, unit := range item.Units {
			unitExists := false
			// 检查基准单位
			if material.Unit != nil && unit.MaterialUnitUuid == material.Unit.Uuid {
				unitExists = true
			} else {
				// 检查非基准单位列表
				for _, materialUnit := range material.NotBaseUnitList {
					if unit.MaterialUnitUuid == materialUnit.Uuid {
						unitExists = true
						break
					}
				}
			}
			if !unitExists {
				return nil, nil, errors.New("物品单位参数错误")
			}
		}
	}

	return warehouseItems, materials, nil
}
