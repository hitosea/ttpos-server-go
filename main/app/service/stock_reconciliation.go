package service

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
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
	GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error)             // 获取盘点单列表
	GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error)       // 获取盘点单详情
	SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) (uint64, error)                                          // 更新盘点单
	DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error                                                // 删除盘点单
	ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error)                      // 审核盘点单
	RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error                                                // 驳回盘点单
	CheckMaterials(ctx context.Context, req req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error) // 检查物品
}

// stockReconciliationSrv 盘点单服务实现
type stockReconciliationSrv struct {
	productSrv IProductSrv
	dbm        *database.DBManager
	lock       lock.Lock
}

// NewStockReconciliationSrv 创建盘点单服务
func NewStockReconciliationSrv(dbm *database.DBManager, productSrv IProductSrv) IStockReconciliationSrv {
	return NewStockReconciliationSrvImpl(dbm, productSrv)
}

// NewStockReconciliationSrvImpl 创建盘点单服务实现
func NewStockReconciliationSrvImpl(dbm *database.DBManager, productSrv IProductSrv) IStockReconciliationSrv {
	return &stockReconciliationSrv{
		dbm:        dbm,
		productSrv: productSrv,
		lock:       lock.NewSystemLock(),
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

	opts = append(opts, stockReconciliationRepo.WithWarehouseMultiLanguageName())

	// 查询数据
	list, total, err := stockReconciliationRepo.GetStockReconciliationListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.StockReconciliationInfo, 0, len(list))

	stockReconciliationUuidList := make([]uint64, 0, len(list))
	for _, item := range list {
		stockReconciliationUuidList = append(stockReconciliationUuidList, item.Uuid)
	}

	// 根据盘点单获取每个盘点单的物品数量，返回map[盘点单UUID]物品数量
	itemsCountMap, err := stockReconciliationRepo.GetStockReconciliationItemCountListByReconciliationUuidList(stockReconciliationUuidList)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单物品明细失败")
	}

	for _, item := range list {
		info := &resp.StockReconciliationInfo{}
		if err := copier.Copy(info, item); err != nil {
			continue
		}
		info.WarehouseLocaleName = item.Warehouse.MultiLanguageName.GetNames()
		info.ItemsCount = itemsCountMap[item.Uuid]
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

// getBookedStockMap 获取仓库物品的账面库存数量
func (s *stockReconciliationSrv) getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
	bookedStockMap := make(map[uint64]decimal.Decimal)
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(warehouseItemRepo.WhereWarehouseUuid(warehouseUuid))
	if err != nil {
		return bookedStockMap, errors.WithMessage(err, "查询仓库物品列表失败")
	}
	for _, warehouseItem := range warehouseItems {
		bookedStockMap[warehouseItem.MaterialUuid] = decimal.NewFromFloat(warehouseItem.Stock)
	}
	return bookedStockMap, nil
}

// GetStockReconciliationDetail 获取盘点单详情
func (s *stockReconciliationSrv) GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
	var detailResp resp.StockReconciliationDetailResp

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(req.Uuid),
		stockReconciliationRepo.WithWarehouseMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemMaterialUnits(),
	}
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
	if err != nil {
		return detailResp, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return detailResp, errors.New("盘点单不存在")
	}

	// 转换响应数据
	if err := copier.Copy(&detailResp, stockReconciliation); err != nil {
		return detailResp, errors.WithMessage(err, "转换盘点单数据失败")
	}
	detailResp.WarehouseName = stockReconciliation.Warehouse.MultiLanguageName.GetNames()

	bookedQuantityMap, err := s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
	if err != nil {
		return detailResp, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}

	// 物品单位明细
	itemsResp := make([]*resp.StockReconciliationItemInfo, 0, len(stockReconciliation.StockReconciliationItems))
	for _, item := range stockReconciliation.StockReconciliationItems {
		// 明细中的物品已删除，跳过
		if item.DeleteTime > 0 {
			continue
		}
		itemInfo := &resp.StockReconciliationItemInfo{}
		if err := copier.Copy(itemInfo, item); err != nil {
			continue
		}

		itemInfo.BookedQuantity = item.BookedQuantity.InexactFloat64()
		itemInfo.CountedQuantity = item.CountedQuantity.InexactFloat64()

		// 查询物品信息
		if item.Material != nil {
			itemInfo.LocaleName = *language.JsonToLocaleResponse(item.Material.Name)
			itemInfo.MaterialCode = item.Material.Code
			itemInfo.InternalCode = item.Material.InternalCode
			itemInfo.MaterialBarcode = item.Material.BarcodeValue
		}

		itemInfo.Units = make([]resp.MaterialUnitInfo, 0)
		for _, unit := range item.Material.NotBaseUnitList {
			if unit.Unit != nil {
				itemInfo.Units = append(itemInfo.Units, resp.MaterialUnitInfo{
					MaterialUnitUuid: unit.Uuid,
					UnitUuid:         unit.UnitUuid,
					UnitName:         unit.Unit.MultiLanguageName.GetNames(),
					ConversionRate:   unit.ConversionRate,
					IsDefault:        unit.IsDefault,
				})
			}
		}

		// 查询单位明细
		itemUnits, err := stockReconciliationRepo.GetStockReconciliationItemUnitListByItemUuid(item.Uuid)
		if err == nil && len(itemUnits) > 0 {
			unitsResp := make([]*resp.StockReconciliationItemUnitInfo, 0, len(itemUnits))
			for _, itemUnit := range itemUnits {
				if itemUnit.MaterialUnit == nil || itemUnit.MaterialUnit.Unit == nil || itemUnit.MaterialUnit.Unit.MultiLanguageName.Uuid == 0 ||
					(stockReconciliation.Status == constant.StockReconciliationStatusSubmitted && itemUnit.Quantity == nil) {
					continue
				}
				unitInfo := &resp.StockReconciliationItemUnitInfo{}
				if err := copier.Copy(unitInfo, itemUnit); err != nil {
					continue
				}
				unitInfo.LocaleName = itemUnit.MaterialUnit.Unit.MultiLanguageName.GetNames()
				for _, unit := range itemInfo.Units {
					if unit.MaterialUnitUuid == itemUnit.MaterialUnitUuid {
						unitInfo.ConversionRate = unit.ConversionRate
						break
					}
				}
				unitsResp = append(unitsResp, unitInfo)
			}
			itemInfo.ItemUnits = unitsResp
		}

		bookedQuantity := item.BookedQuantity
		// 已保存状态，账面库存数量要实时读取；其他状态，账面库存数量为盘点单中的数量
		if stockReconciliation.Status == constant.StockReconciliationStatusSaved {
			bookedQuantity = bookedQuantityMap[item.MaterialUuid]
		}
		// 盘盈盘亏状态
		if item.CountedQuantity.GreaterThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusProfit
		} else if item.CountedQuantity.LessThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusLoss
		} else {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusNormal
		}
		// 是否盘盈盘亏异常（账面和实盘数量差值的绝对值大于20%）
		itemInfo.IsInventoryStatusException = s.getIsInventoryStatusException(bookedQuantity, item.CountedQuantity)
		itemInfo.DiffQuantity = item.CountedQuantity.Sub(bookedQuantity).Truncate(3).InexactFloat64()
		itemsResp = append(itemsResp, itemInfo)
	}
	detailResp.Items = itemsResp

	return detailResp, nil
}

// SaveStockReconciliation 保存盘点单
func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, saveReq req.StockReconciliationSaveReq) (uint64, error) {

	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	stockReconciliationUuid := saveReq.Uuid
	var stockReconciliation *model.StockReconciliation
	var err error

	if saveReq.Uuid == 0 { // 新建
		// 加锁保证单号唯一性（基于公司UUID和日期）
		dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
		lockKey := fmt.Sprintf("stock_reconciliation_%d_%s", ctx.GetCompanyUuid(), dateStr)
		s.lock.LockUuidString(lockKey)
		defer s.lock.UnlockUuidString(lockKey)
	} else { // 修改
		// 加锁
		s.lock.LockUuid(saveReq.Uuid)
		defer s.lock.UnlockUuid(saveReq.Uuid)

		stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
		// 查询盘点单
		stockReconciliation, err = stockReconciliationRepo.GetStockReconciliationByUuid(saveReq.Uuid)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "查询盘点单失败")
		}
		if stockReconciliation == nil {
			return stockReconciliationUuid, errors.New("盘点单不存在")
		}

		// 只有已保存状态的盘点单才能修改
		if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
			if saveReq.IsSubmit {
				return stockReconciliationUuid, errors.New("当前状态不允许提交")
			} else {
				return stockReconciliationUuid, errors.New("当前状态不允许修改")
			}
		}
	}

	// A、在列表上直接提交
	if saveReq.IsSubmit && !saveReq.SubmitAfterSave && saveReq.Uuid > 0 {
		// 直接提交
		return stockReconciliationUuid, s.submitStockReconciliation(ctx, saveReq.Uuid, true)
	}

	// B、在详情中保存或者提交

	// 验证仓库和物品明细
	warehouseItems, materials, err := s.validateWarehouseAndItems(db, saveReq)
	if err != nil {
		return stockReconciliationUuid, err
	}

	bookedQuantityMap := map[uint64]float64{}
	for _, warehouseItem := range warehouseItems {
		bookedQuantityMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
	}

	materialUnitMap := make(map[uint64]map[uint64]float64)
	for _, material := range materials {
		materialUnitMap[material.Uuid] = make(map[uint64]float64)
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
		}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		if saveReq.Uuid == 0 { // 新建
			// 在事务内部生成单据编号
			orderNo := s.generateOrderNo(tx, timezone)
			// 创建盘点单
			stockReconciliation = &model.StockReconciliation{
				OrderNo:       orderNo,
				Type:          saveReq.Type,
				WarehouseUuid: saveReq.WarehouseUuid,
				Purpose:       saveReq.Purpose,
				Status:        constant.StockReconciliationStatusSaved, // 0-已保存
			}
			if err := stockReconciliationRepo.CreateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(errors.New("创建盘点单失败"), err.Error())
			}

			stockReconciliationUuid = stockReconciliation.Uuid
		} else { // 更新
			// 更新盘点单
			stockReconciliation.WarehouseUuid = saveReq.WarehouseUuid
			stockReconciliation.Purpose = saveReq.Purpose
			stockReconciliation.Type = saveReq.Type

			if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(err, "更新盘点单失败")
			}
			// 删除原有的物品明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品明细失败")
			}
			// 删除原有物品单位明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemUnitByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品单位明细失败")
			}
		}

		var stockReconciliationItemUnits []*model.StockReconciliationItemUnit
		for _, reqItem := range saveReq.Items {
			// 计算实盘数量（基准单位）
			countedQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitItem := range reqItem.Units {
					if unitItem.Quantity == nil {
						continue
					}
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitItem.MaterialUnitUuid]
					unitQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					countedQuantity = countedQuantity.Add(unitQuantity)
				}
			}
			countedQuantity = countedQuantity.Truncate(3)

			item := &model.StockReconciliationItem{
				StockReconciliationUuid: stockReconciliation.Uuid,
				MaterialUuid:            reqItem.MaterialUuid,
				BookedQuantity:          decimal.NewFromFloat(bookedQuantityMap[reqItem.MaterialUuid]), // 每次保存都实时读取账面库存数量
				CountedQuantity:         countedQuantity,
			}

			// 先创建item以获取自动生成的uuid
			if err := stockReconciliationRepo.CreateStockReconciliationItem(item); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品明细失败"), err.Error())
			}

			// 创建单位明细
			for _, unitItem := range reqItem.Units {
				var quantity *float64
				if unitItem.Quantity != nil {
					quantityDecimal := unitItem.Quantity.InexactFloat64()
					quantity = &quantityDecimal
				}
				stockReconciliationItemUnits = append(stockReconciliationItemUnits, &model.StockReconciliationItemUnit{
					StockReconciliationItemUuid: item.Uuid,
					MaterialUnitUuid:            unitItem.MaterialUnitUuid,
					Quantity:                    quantity,
				})
			}
		}

		if len(stockReconciliationItemUnits) > 0 {
			if err := stockReconciliationRepo.CreateStockReconciliationItemUnitBatch(stockReconciliationItemUnits); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品单位明细失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		errMsg := "保存失败"
		if saveReq.IsSubmit {
			errMsg = "提交失败"
		}
		return stockReconciliationUuid, errors.WithMessage(errors.New(errMsg), err.Error())
	}

	// 提交盘点单
	if saveReq.IsSubmit && ctx.GetCompany().IsOpenErp() {
		err = s.submitStockReconciliation(ctx, stockReconciliation.Uuid, false)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "提交盘点单失败")
		}
	}

	return stockReconciliationUuid, nil
}

// 提交盘点单
// stockReconciliationUuid: 盘点单UUID
// isDirectSubmit: 是否列表上直接提交，true表示在列表上点击提交，false表示保存后提交
func (s *stockReconciliationSrv) submitStockReconciliation(ctx context.Context, stockReconciliationUuid uint64, isDirectSubmit bool) error {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(stockReconciliationUuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsUnits(),
		stockReconciliationRepo.WithWarehouse(),
	}
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)

	if err != nil {
		return errors.WithMessage(errors.New("查询盘点单失败"), err.Error())
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	if ctx.Version(context.GTE, constant.ClientVersionV2100) && stockReconciliation.Warehouse != nil && stockReconciliation.Warehouse.IsDisabled() {
		return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
	}

	bookedQuantityMap := make(map[uint64]decimal.Decimal)
	if isDirectSubmit {
		var err error
		bookedQuantityMap, err = s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
		if err != nil {
			return errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
		}
	}

	companySetting := ctx.GetCompanySetting()

	// 根据时区获取过账日期和时间
	now := utils.SetTimezone(companySetting.GetTimezone()).Now()

	var erpItems []*stock.StockReconciliationItem
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)
		for _, item := range stockReconciliation.StockReconciliationItems {
			// 物品已禁用，标记item的delete_time(删除)
			if !item.Material.Status {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除已关闭物品失败"), err.Error())
				}
				continue
			}
			// 不往erp传递已删除的物品明细
			if item.DeleteTime > 0 {
				continue
			}

			var unitExists bool
			for _, unit := range item.StockReconciliationItemUnits {
				if unit.DeleteTime == 0 && unit.Quantity != nil {
					unitExists = true
					break
				}
			}
			if !unitExists {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除待盘点物品失败"), err.Error())
				}
				continue
			}

			if isDirectSubmit {
				stockReconciliationItem := *item
				stockReconciliationItem.BookedQuantity = bookedQuantityMap[item.MaterialUuid]
				if err := stockReconciliationRepo.UpdateStockReconciliationItem(&stockReconciliationItem); err != nil {
					return errors.WithMessage(errors.New("更新盘点单物品明细失败"), err.Error())
				}
			}

			erpItems = append(erpItems, &stock.StockReconciliationItem{
				ItemCode: item.Material.Code,
				Qty:      item.CountedQuantity.InexactFloat64(),
			})
		}

		if len(erpItems) == 0 {
			return errors.New("物品列表为空，请先添加物品后再操作")
		}
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpReq, err := erpSrv.SubmitStockReconciliation(ctx, companySetting, &stock.SaveStockReconciliationReq{
			CompanyAbbr: companySetting.ErpnextCompanyAbbr,
			Branch:      companySetting.ErpnextBranchName,
			PostingDate: now.Format("2006-01-02"),
			PostingTime: now.Format("15:04:05"),
			Warehouse:   stockReconciliation.Warehouse.ErpCode,
			Items:       erpItems,
		})
		if err != nil {
			logger.Logger.Error("提交盘点单失败", zap.Error(err))
			// 检查是否是仓库禁用错误
			if ctx.Version(context.GTE, constant.ClientVersionV2100) && strings.Contains(err.Error(), "Disabled Warehouse") {
				return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
			}
			// 提取物品名称
			itemName := s.extractName("Item", "is disabled", err.Error())
			for _, item := range stockReconciliation.StockReconciliationItems {
				if item.Material.Code == itemName {
					materialName := item.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					if ctx.Version(context.GTE, constant.ClientVersionV2100) {
						return errors.NewWithCode(constant.CodeItemDisabled, materialName)
					}
					message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", materialName)
					return errors.New(message)
				}
			}
			if itemName != "" {
				if ctx.Version(context.GTE, constant.ClientVersionV2100) {
					return errors.NewWithCode(constant.CodeItemDisabled, itemName)
				}
				message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", itemName)
				return errors.New(message)
			}
			return errors.WithMessage(errors.New("提交盘点单失败"), err.Error())
		}
		// 更新盘点单erp_code和提交时间
		stockReconciliation.ErpCode = erpReq.StockReconciliationName
		stockReconciliation.SubmitTime = int(time.Now().Unix())
		stockReconciliation.Status = constant.StockReconciliationStatusSubmitted
		if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
			return errors.WithMessage(errors.New("更新盘点单状态失败"), err.Error())
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
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
func (s *stockReconciliationSrv) ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error) {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(req.Uuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsMaterialBaseUnit(),
		stockReconciliationRepo.WithWarehouse(),
	}
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return nil, errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能审核
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return nil, errors.New("当前状态不允许审核")
	}

	// 检查仓库是否被禁用
	if ctx.Version(context.GTE, constant.ClientVersionV2100) && stockReconciliation.Warehouse != nil && stockReconciliation.Warehouse.IsDisabled() {
		return nil, errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
	}

	disabledMaterials := make([]dto.LocaleResponse, 0)
	for _, item := range stockReconciliation.StockReconciliationItems {
		// 跳过已删除物品明细
		if item.DeleteTime > 0 {
			continue
		}
		// 如果从提交到审核期间，物品已禁用，则提示请求该物品为开启
		if !item.Material.Status {
			disabledMaterials = append(disabledMaterials, item.Material.MultiLanguageName.GetNames())
		}
	}

	if len(disabledMaterials) > 0 {
		return disabledMaterials, errors.New("请修改物品状态")
	}

	// 获取仓库Uuid获取仓库物品信息列表
	warehouseItems, err := repository.NewWarehouseItemRepo(db).GetByWarehouseUuid(stockReconciliation.WarehouseUuid)
	if err != nil {
		return disabledMaterials, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}
	warehouseMaterialUuids := make(map[uint64]struct{})
	for _, item := range warehouseItems {
		warehouseMaterialUuids[item.MaterialUuid] = struct{}{}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 更新盘点单状态为已审核
		updateData := map[string]any{
			"status":      constant.StockReconciliationStatusApproved, // 2-已审核
			"update_time": int(time.Now().Unix()),
		}
		if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
			return errors.WithMessage(errors.New("审核盘点单失败"), err.Error())
		}
		// 遍历所有物品
		var warehouseLogs []*model.WarehouseInOutLog
		for _, item := range stockReconciliation.StockReconciliationItems {
			// 跳过已删除物品明细
			if item.DeleteTime > 0 {
				continue
			}
			material := item.Material
			// 审核时，物品被删除，则删除物品明细
			if item.Material.DeleteTime > 0 {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("审核盘点单时移除已删除物品失败"), err.Error())
				}
				continue
			}
			stockQuantity := item.CountedQuantity.Truncate(3).InexactFloat64()
			_, exists := warehouseMaterialUuids[item.MaterialUuid]
			if !exists { // 仓库不存在该物品，则新增仓库物品，库存数量为实盘数量
				if err := tx.Create(&model.WarehouseItem{
					WarehouseUuid: stockReconciliation.WarehouseUuid,
					MaterialUuid:  item.MaterialUuid,
					MaterialCode:  item.Material.Code,
					Stock:         stockQuantity,
					Valuation:     1.0, // 和同步仓库物品库存一致
				}).Error; err != nil {
					return errors.WithMessage(errors.New("创建仓库物品失败"), err.Error())
				}
			} else { // 仓库存在该物品，则更新仓库物品库存为实盘数量
				if err := tx.Model(&model.WarehouseItem{}).
					Where("warehouse_uuid = ?", stockReconciliation.WarehouseUuid).
					Where("material_uuid = ?", item.MaterialUuid).Update("stock", stockQuantity).Error; err != nil {
					return errors.WithMessage(errors.New("更新仓库物品库存失败"), err.Error())
				}
			}

			// 有盈亏才记录日志
			if !item.CountedQuantity.Equal(item.BookedQuantity) {
				scene := constant.WarehouseInOutLogSceneProfitIn        // 盘盈
				logType := constant.WarehouseInOutLogLogTypeIn          // 入库
				if item.CountedQuantity.LessThan(item.BookedQuantity) { // 盘亏出库
					logType = constant.WarehouseInOutLogLogTypeOut
					scene = constant.WarehouseInOutLogSceneLossOut
				}
				diff := item.CountedQuantity.Sub(item.BookedQuantity).Abs()
				warehouseLogs = append(warehouseLogs, &model.WarehouseInOutLog{
					LogType:              logType,
					Scene:                scene,
					WarehouseUuid:        stockReconciliation.WarehouseUuid,
					MaterialUuid:         item.MaterialUuid,
					MaterialName:         item.MaterialName,
					MaterialBaseUnitUuid: material.Unit.Uuid, // 基准单位
					MaterialBaseUnitName: material.Unit.Name, // 基准单位名称
					Num:                  diff.Truncate(3).InexactFloat64(),
					Price:                material.Valuation,
					Amount:               decimal.NewFromFloat(material.Valuation).Mul(diff).Truncate(3).InexactFloat64(),
					OrderNo:              stockReconciliation.OrderNo,
				})
			}
		}
		if len(warehouseLogs) > 0 {
			if err := tx.Create(&warehouseLogs).Error; err != nil {
				return errors.WithMessage(errors.New("创建盘盈盘亏出入库记录失败"), err.Error())
			}
		}

		// 调用erp接口审核盘点单
		if ctx.GetCompany().IsOpenErp() && stockReconciliation.ErpCode != "" {
			companySetting := ctx.GetCompanySetting()
			erpSrv := erp.NewIErpSrv(s.dbm)
			_, err := erpSrv.ApproveStockReconciliation(ctx, companySetting, &stock.SubmitStockReconciliationReq{
				StockReconciliationName: stockReconciliation.ErpCode,
			})
			if err != nil {
				logger.Logger.Error("审核盘点单失败", zap.Error(err))
				// 检查是否是仓库禁用错误
				if ctx.Version(context.GTE, constant.ClientVersionV2100) && strings.Contains(err.Error(), "Disabled Warehouse") {
					return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
				}
				// 提取物品名称
				itemName := s.extractName("Item", "is disabled", err.Error())
				for _, item := range stockReconciliation.StockReconciliationItems {
					if item.Material.Code == itemName {
						materialName := item.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
						if ctx.Version(context.GTE, constant.ClientVersionV2100) {
							return errors.NewWithCode(constant.CodeItemDisabled, materialName)
						}
						message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", materialName)
						return errors.New(message)
					}
				}
				if itemName != "" {
					if ctx.Version(context.GTE, constant.ClientVersionV2100) {
						return errors.NewWithCode(constant.CodeItemDisabled, itemName)
					}
					message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", itemName)
					return errors.New(message)
				}
				return errors.WithMessage(errors.New("审核盘点单失败"), err.Error())
			}
		}

		return nil
	})
	if err != nil {
		return disabledMaterials, err
	}
	utils.Go(func() {
		// 计算所有关联成本卡的商品的库存
		err = s.productSrv.SyncProductStockByBomCard(ctx)
		if err != nil {
			logger.Logger.Error("审核通过盘点单-计算商品库存失败", zap.Error(err))
		}
	})
	return disabledMaterials, nil
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
func (s *stockReconciliationSrv) validateWarehouseAndItems(db *gorm.DB, req req.StockReconciliationSaveReq) ([]model.WarehouseItem, []*model.Material, error) {
	warehouseUuid := req.WarehouseUuid
	if warehouseUuid == 0 {
		return nil, nil, errors.New("仓库参数错误")
	}
	if req.Purpose != 1 && req.Purpose != 2 {
		return nil, nil, errors.New("盘点目的参数错误")
	}
	// 判断仓库是否存在，且类型为normal，且未被禁用
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

	// 获取所有物品UUID
	materialUuids := make([]uint64, 0, len(req.Items))
	for _, item := range req.Items {
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
	for _, item := range req.Items {
		material, exists := materialMap[item.MaterialUuid]
		if !exists {
			return nil, nil, errors.New("物品参数错误")
		}

		unitExists := false
		// 验证物品单位
		for _, unit := range item.Units {
			// 检查单位列表
			for _, materialUnit := range material.NotBaseUnitList {
				if unit.MaterialUnitUuid == materialUnit.Uuid && unit.Quantity != nil {
					unitExists = true
					break
				}
			}
		}
		if !unitExists && req.IsSubmit {
			return nil, nil, errors.New("物品单位参数错误")
		}
	}

	return warehouseItems, materials, nil
}

// getIsInventoryStatusException 获取是否盘盈盘亏异常
func (s *stockReconciliationSrv) getIsInventoryStatusException(bookedQuantity decimal.Decimal, countedQuantity decimal.Decimal) bool {
	if bookedQuantity.IsZero() {
		if countedQuantity.IsZero() {
			return false
		}
		return true
	}
	return countedQuantity.Sub(bookedQuantity).Abs().Div(bookedQuantity).GreaterThan(decimal.NewFromFloat(0.2))
}

func (s *stockReconciliationSrv) CheckMaterials(ctx context.Context, checkReq req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error) {

	var listResp resp.StockReconciliationCheckMaterialsListResp
	itemResp := make([]resp.StockReconciliationCheckMaterialsResp, 0)

	db := ctx.GetDB()

	var materialUuids []uint64

	bookedQuantityMap := make(map[uint64]decimal.Decimal)

	if checkReq.Uuid != 0 {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
		opts := []repository.DBOption{
			stockReconciliationRepo.WhereUuid(checkReq.Uuid),
			stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
			stockReconciliationRepo.WithStockReconciliationItemsUnits(),
		}
		// 查询盘点单
		stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
		if err != nil {
			return listResp, errors.WithMessage(err, "查询盘点单失败")
		}
		if stockReconciliation == nil {
			return listResp, errors.New("盘点单不存在")
		}

		bookedQuantityMap, err = s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
		}

		limitedMaterialUuids := make([]uint64, 0)
		for _, item := range checkReq.Items {
			limitedMaterialUuids = append(limitedMaterialUuids, item.MaterialUuid)
		}
		for _, item := range stockReconciliation.StockReconciliationItems {
			if item.DeleteTime > 0 || (len(limitedMaterialUuids) > 0 && !slices.Contains(limitedMaterialUuids, item.MaterialUuid)) {
				continue
			}
			materialUuids = append(materialUuids, item.MaterialUuid)
			bookedQuantity := item.BookedQuantity
			// 已保存状态，账面库存数量要实时读取；其他状态，账面库存数量为盘点单中的数量
			if stockReconciliation.Status == constant.StockReconciliationStatusSaved {
				bookedQuantity = bookedQuantityMap[item.MaterialUuid]
			}

			var unitCount uint
			for _, unit := range item.StockReconciliationItemUnits {
				if unit.Quantity != nil {
					unitCount++
				}
			}

			itemResp = append(itemResp, resp.StockReconciliationCheckMaterialsResp{
				LocaleName:                 item.Material.MultiLanguageName.GetNames(),
				IsInventoryStatusException: unitCount > 0 && s.getIsInventoryStatusException(bookedQuantity, item.CountedQuantity),
				Status:                     item.Material.Status,
				IsDeleted:                  item.Material.DeleteTime > 0,
				UnitCount:                  unitCount,
			})
		}
	}

	// 没传递盘点单UUID，则根据仓库UUID查询仓库物品
	if checkReq.WarehouseUuid != 0 && checkReq.Uuid == 0 {
		var err error
		bookedQuantityMap, err = s.getBookedQuantityMap(db, checkReq.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
		}
	}
	var warehouseDisabled bool
	if checkReq.WarehouseUuid != 0 && ctx.Version(context.GTE, constant.ClientVersionV2100) {
		warehouseRepo := repository.NewWarehouseRepo(db)
		warehouse, err := warehouseRepo.GetByUuid(checkReq.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New("查询仓库失败"), err.Error())
		}
		warehouseDisabled = warehouse != nil && warehouse.IsDisabled()
	}

	if len(checkReq.Items) > 0 {
		var newMaterialUuids []uint64
		itemMap := make(map[uint64]req.StockReconciliationCheckMaterialsItem)
		// 过滤掉在materialUuids中的物品
		for _, item := range checkReq.Items {
			if !slices.Contains(materialUuids, item.MaterialUuid) {
				newMaterialUuids = append(newMaterialUuids, item.MaterialUuid)
			}
			itemMap[item.MaterialUuid] = item
		}

		var materials []model.Material
		db.Model(&model.Material{}).Preload("MultiLanguageName").Where("uuid IN (?)", newMaterialUuids).Find(&materials)

		for _, material := range materials {
			countedQuantity := itemMap[material.Uuid].CountedQuantity
			itemResp = append(itemResp, resp.StockReconciliationCheckMaterialsResp{
				LocaleName:                 material.MultiLanguageName.GetNames(),
				Status:                     material.Status,
				IsDeleted:                  material.DeleteTime > 0,
				IsInventoryStatusException: s.getIsInventoryStatusException(bookedQuantityMap[material.Uuid], countedQuantity),
			})
		}
	}

	return resp.StockReconciliationCheckMaterialsListResp{
		List:              itemResp,
		WarehouseDisabled: warehouseDisabled,
	}, nil
}

// extractName 从错误信息中提取物品名称
func (s *stockReconciliationSrv) extractName(name, after, errorMsg string) string {
	// 转义正则表达式中的特殊字符
	escapedName := regexp.QuoteMeta(name)
	escapedAfter := regexp.QuoteMeta(after)
	// 使用正则表达式匹配，支持HTML标签和普通文本
	var re *regexp.Regexp
	if strings.Contains(name, "<") || strings.Contains(after, "<") {
		// HTML标签模式：不要求空格分隔
		re = regexp.MustCompile(escapedName + `(.+?)` + escapedAfter)
	} else {
		// 普通文本模式：要求空格分隔
		re = regexp.MustCompile(escapedName + `\s+(.+?)\s+` + escapedAfter)
	}
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		supplierInfo := strings.TrimSpace(matches[1])
		// 如果包含编码#名称格式，提取名称部分
		if strings.Contains(supplierInfo, "#") {
			parts := strings.SplitN(supplierInfo, "#", 2)
			if len(parts) == 2 {
				return parts[1] // 返回物品erp_code
			}
		}
		return supplierInfo
	}
	return ""
}
