package purchase_order

import (
	"time"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// purchaseReceiptOrderSrv 收货单服务
type purchaseReceiptOrderSrv struct {
	dbm       *database.DBManager
	validator *purchaseOrderValidator
	helper    *purchaseOrderHelper
}

// newPurchaseReceiptOrderSrv 创建收货单服务实例
func newPurchaseReceiptOrderSrv(dbm *database.DBManager) *purchaseReceiptOrderSrv {
	return &purchaseReceiptOrderSrv{
		dbm:       dbm,
		validator: &purchaseOrderValidator{},
		helper:    &purchaseOrderHelper{},
	}
}

// CreatePurchaseReceiptOrder 创建收货单
func (s *purchaseReceiptOrderSrv) CreatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptCreateReq,
) (resp.PurchaseReceiptOrderCreateResp, error) {
	db := ctx.GetDB()

	// 判断物品明细是否已经停用
	if req.IsConfirm {
		itemUuids := make([]uint64, 0, len(req.Items))
		for _, item := range req.Items {
			itemUuids = append(itemUuids, item.PurchaseOrderItemUuid)
		}
		if err := s.validator.validateReceiptMaterialStatus(ctx, db, itemUuids); err != nil {
			return resp.PurchaseReceiptOrderCreateResp{}, err
		}
	}

	var result resp.PurchaseReceiptOrderCreateResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		purchaseOrderItemUnitRepo := repository.NewPurchaseOrderItemUnitRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.PurchaseOrderUuid)
		if err != nil {
			logger.Logger.Error("CreatePurchaseReceiptOrder-GetByUuid", zap.Any("purchaseOrderUuid", req.PurchaseOrderUuid), zap.Any("err", err))
			return errors.WithMessage(errors.New("采购申请不存在"), err.Error())
		}
		if !purchaseOrder.CanReceive() {
			return errors.New("采购单状态不允许收货")
		}

		// 总部相关信息预处理
		var headquarterInfo *HeadquarterUpdateInfo
		if req.IsConfirm && purchaseOrder.IsHeadquarterPurchase() {
			hqInfo, err := s.helper.initHeadquarterInfo(ctx, s.dbm, purchaseOrder)
			if err != nil {
				return err
			}
			headquarterInfo = hqInfo
		}

		// 创建收货单
		receiptOrder := &model.PurchaseReceiptOrder{
			OrderNo:                s.helper.generateReceiptNo(tx, ctx.GetCompanySetting().Timezone),
			Status:                 utils.IfInt(req.IsConfirm, constant.ReceiptOrderStatusReceived, constant.ReceiptOrderStatusPending),
			PurchaseOrderUuid:      req.PurchaseOrderUuid,
			PurchaseOrderNo:        purchaseOrder.OrderNo,
			PurchaseTime:           purchaseOrder.OrderTime,
			Num:                    float64(len(req.Items)),
			ExpectArrivalTime:      purchaseOrder.ExpectArrivalTime,
			SupplierName:           purchaseOrder.SupplierName,
			SupplierErpCode:        purchaseOrder.SupplierErpCode,
			ReceiveTime:            req.ReceiveTime,
			PurchaseOrder:          *purchaseOrder,
			SourceWarehouseErpCode: purchaseOrder.WarehouseErpCode,
			SourceWarehouseName:    purchaseOrder.WarehouseName,
			TargetWarehouseErpCode: purchaseOrder.DefaultWarehouseErpCode,
			TargetWarehouseName:    purchaseOrder.DefaultWarehouseName,
			ReceiptType: func() int {
				if purchaseOrder.PurchaseType == 2 {
					return 2
				}
				return 1
			}(),
		}

		err = receiptOrderRepo.Create(receiptOrder)
		if err != nil {
			logger.Logger.Error("CreatePurchaseReceiptOrder-Create", zap.Any("receiptOrder", receiptOrder), zap.Any("err", err))
			return errors.WithMessage(errors.New("创建收货单失败"))
		}

		// 创建收货明细并更新采购申请明细的到货数量
		var receiptItems []model.PurchaseReceiptOrderItem
		for _, itemReq := range req.Items {
			// 查询采购申请明细
			orderItem, err := purchaseOrderItemRepo.GetByUuid(
				itemReq.PurchaseOrderItemUuid,
				purchaseOrderItemRepo.WithPreloadUnits(),
			)
			if err != nil {
				logger.Logger.Error("CreatePurchaseReceiptOrder-GetByUuid", zap.Any("purchaseOrderItemUuid", itemReq.PurchaseOrderItemUuid), zap.Any("err", err))
				return errors.WithMessage(errors.New("查询采购申请明细失败"), err.Error())
			}

			// 计算收货数量
			reqNum := 0.0
			if err, num := s.validator.validateReceiptQuantityNew(ctx, orderItem, itemReq.UnitList); err != nil {
				return err
			} else {
				reqNum = num
			}

			// 更新采购申请明细的到货数量
			newArrivalNum := orderItem.ArrivalNum + reqNum

			// 创建收货明细
			units := func() []model.PurchaseReceiptOrderItemUnit {
				if len(orderItem.Units) == 0 && orderItem.BaseUnitUuid != 0 {
					orderItem.Units = make([]model.PurchaseOrderItemUnit, 0, len(itemReq.UnitList))
					orderItem.Units = append(orderItem.Units, model.PurchaseOrderItemUnit{
						UnitUuid:           orderItem.UnitUuid,
						UnitName:           orderItem.UnitName,
						UnitConversionRate: orderItem.UnitConversionRate,
						BaseUnitUuid:       orderItem.BaseUnitUuid,
						BaseUnitName:       orderItem.BaseUnitName,
						ErpnextUom:         orderItem.ErpnextUom,
					})
				}
				units := make([]model.PurchaseReceiptOrderItemUnit, 0, len(itemReq.UnitList))
				for _, unit := range itemReq.UnitList {
					for _, orderItemUnit := range orderItem.Units {
						if orderItemUnit.UnitUuid == unit.Uuid {
							units = append(units, model.PurchaseReceiptOrderItemUnit{
								ItemUuid:                 orderItem.Uuid,
								PurchaseReceiptOrderUuid: receiptOrder.Uuid,
								Num:                      unit.Num,
								UnitUuid:                 unit.Uuid,
								UnitName:                 orderItem.UnitName,
								UnitConversionRate:       orderItemUnit.UnitConversionRate,
								BaseUnitUuid:             orderItemUnit.BaseUnitUuid,
								BaseUnitName:             orderItemUnit.BaseUnitName,
								ErpnextUom:               orderItemUnit.ErpnextUom,
							})
						}
					}
				}
				return units
			}()
			receiptItems = append(receiptItems, model.PurchaseReceiptOrderItem{
				ReceiptOrderUuid:      receiptOrder.Uuid,
				PurchaseOrderItemUuid: orderItem.Uuid,
				MaterialCode:          orderItem.MaterialCode,
				MaterialName:          orderItem.MaterialName,
				MaterialUuid:          orderItem.MaterialUuid,
				Num:                   reqNum,
				UnitUuid:              orderItem.UnitUuid,
				UnitName:              orderItem.UnitName,
				BaseUnitUuid:          orderItem.BaseUnitUuid,
				BaseUnitName:          orderItem.BaseUnitName,
				UnitConversionRate:    orderItem.UnitConversionRate,
				ErpnextUom:            orderItem.ErpnextUom,
				BaseErpnextUom:        orderItem.BaseErpnextUom,
				Valuation:             orderItem.Valuation,
				TotalPrice:            orderItem.TotalPrice,
				Units:                 units,
			})

			// 确认收货时，更新采购申请明细的到货数量
			if req.IsConfirm {
				orderItem.ArrivalNum = newArrivalNum
				err = purchaseOrderItemRepo.Update(orderItem)
				if err != nil {
					logger.Logger.Error("CreatePurchaseReceiptOrder-Update", zap.Any("orderItem", orderItem), zap.Any("err", err))
					return errors.WithMessage(errors.New("更新采购申请明细失败"), err.Error())
				}

				// 更新采购申请明细单位到货数量
				for _, unit := range itemReq.UnitList {
					for _, orderItemUnit := range orderItem.Units {
						if orderItemUnit.UnitUuid == unit.Uuid {
							orderItemUnit.ArrivalNum = orderItemUnit.ArrivalNum + unit.Num
							err = purchaseOrderItemUnitRepo.Update(orderItemUnit)
							if err != nil {
								return errors.WithMessage(errors.New("更新采购申请明细单位失败"), err.Error())
							}
						}
					}
				}

				// 收集需要更新的总部采购明细信息
				if headquarterInfo != nil {
					headquarterInfo.ItemsToUpdate = append(headquarterInfo.ItemsToUpdate, HeadquarterItemUpdate{
						MaterialCode:  orderItem.MaterialCode,
						NewArrivalNum: newArrivalNum,
						UnitList:      itemReq.UnitList,
					})
				}
			}
		}

		// 批量创建收货明细
		err = receiptOrderItemRepo.CreateBatch(receiptItems)
		if err != nil {
			logger.Logger.Error("CreatePurchaseReceiptOrder-CreateBatch", zap.Any("receiptItems", receiptItems), zap.Any("err", err))
			return errors.WithMessage(errors.New("创建收货明细失败"), err.Error())
		}

		// 更新收货单明细
		receiptOrder.Items = receiptItems

		// 批量更新总部采购申请明细
		if headquarterInfo != nil && len(headquarterInfo.ItemsToUpdate) > 0 {
			err = s.helper.batchUpdateHeadquarterItems(ctx, headquarterInfo)
			if err != nil {
				logger.Logger.Error("CreatePurchaseReceiptOrder-batchUpdateHeadquarterItems", zap.Any("headquarterInfo", headquarterInfo), zap.Any("err", err))
				return errors.WithMessage(errors.New("批量更新总部采购申请明细失败"), err.Error())
			}
		}

		// 检查收货单是否完成
		if receiptOrder.Status == constant.ReceiptOrderStatusReceived {
			// 更新采购单状态
			err = s.helper.checkAndUpdatePurchaseOrderStatus(ctx, tx, req.PurchaseOrderUuid)
			if err != nil {
				return err
			}
			// 更新总部采购单状态
			if purchaseOrder.IsHeadquarterPurchase() {
				err = s.helper.checkAndUpdatePurchaseOrderStatus(ctx, headquarterInfo.DB, headquarterInfo.PurchaseOrder.Uuid)
				if err != nil {
					return err
				}
			}
			// 添加物料库存
			err = s.updateMaterialStock(ctx, tx, receiptOrder)
			if err != nil {
				return err
			}
		}

		result.Uuid = receiptOrder.Uuid
		result.OrderNo = receiptOrder.OrderNo

		return nil
	})

	if err != nil {
		return resp.PurchaseReceiptOrderCreateResp{}, err
	}

	return result, nil
}

// UpdatePurchaseReceiptOrder 更新收货单
func (s *purchaseReceiptOrderSrv) UpdatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderUpdateReq,
) error {
	db := ctx.GetDB()

	// 判断物品明细是否已经停用
	if req.IsConfirm {
		itemUuids := make([]uint64, 0, len(req.Items))
		for _, item := range req.Items {
			itemUuids = append(itemUuids, item.PurchaseOrderItemUuid)
		}
		if err := s.validator.validateReceiptMaterialStatus(ctx, db, itemUuids); err != nil {
			return err
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)

		// 查询收货单
		receiptOrder, err := receiptOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("收货单不存在"), err.Error())
		}

		// 查询采购申请
		var purchaseOrder *model.PurchaseOrder
		if req.IsConfirm {
			purchaseOrder, err = repository.NewPurchaseOrderRepo(tx).GetByUuid(receiptOrder.PurchaseOrderUuid)
			if err != nil {
				return errors.WithMessage(err, "采购申请不存在")
			}
			if !purchaseOrder.CanReceive() {
				return errors.New("采购单状态不允许收货")
			}
			receiptOrder.PurchaseOrder = *purchaseOrder
		}

		// 总部相关信息预处理
		var headquarterInfo *HeadquarterUpdateInfo
		if req.IsConfirm && purchaseOrder != nil && purchaseOrder.IsHeadquarterPurchase() {
			hqInfo, err := s.helper.initHeadquarterInfo(ctx, s.dbm, purchaseOrder)
			if err != nil {
				return err
			}
			headquarterInfo = hqInfo
		}

		// 重新创建收货明细并更新采购申请明细的到货数量
		var receiptItems []model.PurchaseReceiptOrderItem
		for _, itemReq := range req.Items {
			// 查询收货单明细
			receiptOrderItem, err := receiptOrderItemRepo.GetByUuid(itemReq.Uuid)
			if err != nil {
				return errors.WithMessage(errors.New("查询收货单明细失败"), err.Error())
			}

			// 查询采购申请明细
			purchaseOrderItem, err := purchaseOrderItemRepo.GetByUuid(receiptOrderItem.PurchaseOrderItemUuid, purchaseOrderItemRepo.WithPreloadUnits())
			if err != nil {
				return errors.WithMessage(errors.New("查询采购申请明细失败"), err.Error())
			}

			// 验证收货数量
			reqNum := 0.0
			if err, num := s.validator.validateReceiptQuantityNew(ctx, purchaseOrderItem, itemReq.UnitList); err != nil {
				return err
			} else {
				reqNum = num
			}

			// 计算新的到货数量
			newArrivalNum := purchaseOrderItem.ArrivalNum + reqNum

			// 创建收货明细
			units := func() []model.PurchaseReceiptOrderItemUnit {
				if len(purchaseOrderItem.Units) == 0 && purchaseOrderItem.BaseUnitUuid != 0 {
					purchaseOrderItem.Units = make([]model.PurchaseOrderItemUnit, 0, len(itemReq.UnitList))
					purchaseOrderItem.Units = append(purchaseOrderItem.Units, model.PurchaseOrderItemUnit{
						UnitUuid:           purchaseOrderItem.UnitUuid,
						UnitName:           purchaseOrderItem.UnitName,
						UnitConversionRate: purchaseOrderItem.UnitConversionRate,
						BaseUnitUuid:       purchaseOrderItem.BaseUnitUuid,
						BaseUnitName:       purchaseOrderItem.BaseUnitName,
						ErpnextUom:         purchaseOrderItem.ErpnextUom,
					})
				}
				units := make([]model.PurchaseReceiptOrderItemUnit, 0, len(itemReq.UnitList))
				for _, unit := range itemReq.UnitList {
					for _, orderItemUnit := range purchaseOrderItem.Units {
						if orderItemUnit.UnitUuid == unit.Uuid {
							units = append(units, model.PurchaseReceiptOrderItemUnit{
								ItemUuid:                 receiptOrderItem.Uuid,
								PurchaseReceiptOrderUuid: receiptOrder.Uuid,
								Num:                      unit.Num,
								UnitUuid:                 unit.Uuid,
								UnitName:                 purchaseOrderItem.UnitName,
								UnitConversionRate:       orderItemUnit.UnitConversionRate,
								BaseUnitUuid:             orderItemUnit.BaseUnitUuid,
								BaseUnitName:             orderItemUnit.BaseUnitName,
								ErpnextUom:               orderItemUnit.ErpnextUom,
							})
						}
					}
				}
				return units
			}()
			receiptItems = append(receiptItems, model.PurchaseReceiptOrderItem{
				ReceiptOrderUuid:      receiptOrder.Uuid,
				PurchaseOrderItemUuid: purchaseOrderItem.Uuid,
				MaterialCode:          purchaseOrderItem.MaterialCode,
				MaterialName:          purchaseOrderItem.MaterialName,
				MaterialUuid:          purchaseOrderItem.MaterialUuid,
				Num:                   reqNum,
				UnitUuid:              purchaseOrderItem.UnitUuid,
				UnitName:              purchaseOrderItem.UnitName,
				BaseUnitUuid:          purchaseOrderItem.BaseUnitUuid,
				BaseUnitName:          purchaseOrderItem.BaseUnitName,
				UnitConversionRate:    purchaseOrderItem.UnitConversionRate,
				ErpnextUom:            purchaseOrderItem.ErpnextUom,
				BaseErpnextUom:        purchaseOrderItem.BaseErpnextUom,
				Valuation:             purchaseOrderItem.Valuation,
				TotalPrice:            purchaseOrderItem.TotalPrice,
				Units:                 units,
			})

			// 更新采购申请明细的到货数量
			if req.IsConfirm {
				purchaseOrderItem.ArrivalNum = newArrivalNum
				err = purchaseOrderItemRepo.Update(purchaseOrderItem)
				if err != nil {
					return errors.WithMessage(errors.New("更新采购申请明细失败"), err.Error())
				}

				// 收集需要更新的总部采购明细信息
				if headquarterInfo != nil {
					headquarterInfo.ItemsToUpdate = append(headquarterInfo.ItemsToUpdate, HeadquarterItemUpdate{
						MaterialCode:  purchaseOrderItem.MaterialCode,
						NewArrivalNum: newArrivalNum,
						UnitList:      itemReq.UnitList,
					})
				}
			}
		}

		// 删除所有现有收货明细
		err = receiptOrderItemRepo.DeleteByReceiptOrderUuid(receiptOrder.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("删除收货明细失败"), err.Error())
		}

		// 批量创建收货明细
		err = receiptOrderItemRepo.CreateBatch(receiptItems)
		if err != nil {
			return errors.WithMessage(errors.New("创建收货明细失败"), err.Error())
		}

		// 更新收货单状态
		receiptOrder.Status = utils.IfInt(req.IsConfirm, constant.ReceiptOrderStatusReceived, constant.ReceiptOrderStatusPending)
		receiptOrder.ReceiveTime = req.ReceiveTime
		err = receiptOrderRepo.Update(receiptOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新收货单状态失败"), err.Error())
		}
		receiptOrder.Items = receiptItems

		// 批量更新总部采购申请明细
		if headquarterInfo != nil && len(headquarterInfo.ItemsToUpdate) > 0 {
			err = s.helper.batchUpdateHeadquarterItems(ctx, headquarterInfo)
			if err != nil {
				return err
			}
		}

		// 检查收货单是否完成
		if receiptOrder.Status == constant.ReceiptOrderStatusReceived {
			// 更新采购单状态
			err = s.helper.checkAndUpdatePurchaseOrderStatus(ctx, tx, receiptOrder.PurchaseOrderUuid)
			if err != nil {
				return err
			}
			// 更新总部采购单状态
			if purchaseOrder.IsHeadquarterPurchase() {
				err = s.helper.checkAndUpdatePurchaseOrderStatus(ctx, headquarterInfo.DB, headquarterInfo.PurchaseOrder.Uuid)
				if err != nil {
					return err
				}
			}
			// 添加物料库存
			err = s.updateMaterialStock(ctx, tx, receiptOrder)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetPurchaseReceiptOrderList 获取收货单列表
func (s *purchaseReceiptOrderSrv) GetPurchaseReceiptOrderList(
	ctx context.Context,
	reqs req.PurchaseReceiptOrderListReq,
) (resp.PurchaseReceiptOrderListResp, error) {
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(ctx.GetDB())

	// 构建查询选项
	var opts []repository.DBOption
	if reqs.OrderNo != "" {
		opts = append(opts, receiptOrderRepo.WhereReceiptNo(reqs.OrderNo))
	}
	if len(reqs.StatusIn) > 0 {
		opts = append(opts, receiptOrderRepo.WhereStatusIn(reqs.StatusIn))
	}
	if reqs.ReceiptTimeStart > 0 || reqs.ReceiptTimeEnd > 0 {
		opts = append(opts, receiptOrderRepo.WhereReceiptTimeRange(int(reqs.ReceiptTimeStart), int(reqs.ReceiptTimeEnd)))
	}
	if reqs.ReceiptType > 0 {
		opts = append(opts, receiptOrderRepo.WhereReceiptType(reqs.ReceiptType))
	}
	if reqs.CreateTimeStart > 0 || reqs.CreateTimeEnd > 0 {
		opts = append(opts, receiptOrderRepo.WhereCreateTimeRange(int(reqs.CreateTimeStart), int(reqs.CreateTimeEnd)))
	}
	if len(reqs.UuidIn) > 0 {
		opts = append(opts, receiptOrderRepo.WhereUuidIn(reqs.UuidIn))
	}
	// 排序
	opts = append(opts, receiptOrderRepo.OrderByCreateTime(true))

	// 查询数据
	receipts, total, err := receiptOrderRepo.GetListWithPagination(reqs.PageNo, reqs.PageSize, opts...)
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询收货单列表失败"), err.Error())
	}

	// 转换响应数据
	listResp := make([]*resp.PurchaseReceiptOrderInfo, 0, len(receipts))
	for _, receipt := range receipts {
		receiptInfo := &resp.PurchaseReceiptOrderInfo{}
		if err := copier.Copy(receiptInfo, &receipt); err != nil {
			continue
		}
		listResp = append(listResp, receiptInfo)
	}

	return resp.PurchaseReceiptOrderListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   reqs.PageNo,
			PageSize: reqs.PageSize,
			Total:    total,
		},
	}, nil
}

// GetPurchaseReceiptOrderDetail 获取收货单详情
func (s *purchaseReceiptOrderSrv) GetPurchaseReceiptOrderDetail(
	ctx context.Context,
	req req.PurchaseReceiptOrderDetailReq,
) (resp.PurchaseReceiptOrderDetailResp, error) {
	db := ctx.GetDB()
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 查询收货单详情
	receipt, err := receiptOrderRepo.GetByUuid(req.Uuid, receiptOrderRepo.WithItems())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseReceiptOrderDetailResp{}, errors.New("收货单不存在")
		}
		return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(errors.New("查询收货单详情失败"), err.Error())
	}

	// 转换响应数据
	var detailResp resp.PurchaseReceiptOrderDetailResp
	err = copier.Copy(&detailResp, receipt)
	if err != nil {
		return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 转换收货明细数据
	detailResp.Items = make([]resp.PurchaseReceiptItemInfo, 0, len(receipt.Items))
	for _, item := range receipt.Items {
		itemInfo := resp.PurchaseReceiptItemInfo{}
		if err = copier.Copy(&itemInfo, &item); err != nil {
			return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
		}
		itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
		itemInfo.PurchaseNum = item.PurchaseOrderItem.Num
		itemInfo.ArrivalNum = item.Num
		itemInfo.LocaleUnitName = *language.JsonToLocaleResponse(item.UnitName)
		itemInfo.LocaleBaseUnitName = *language.JsonToLocaleResponse(item.BaseUnitName)
		itemInfo.InternalCode = func(item model.PurchaseReceiptOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.InternalCode
		}(item)
		itemInfo.BarcodeValue = func(item model.PurchaseReceiptOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.BarcodeValue
		}(item)
		// 采购单位列表
		itemInfo.UnitList = func(item model.PurchaseReceiptOrderItem) []resp.PurchaseOrderItemMaterialUnit {
			unitList := []resp.PurchaseOrderItemMaterialUnit{}
			for _, unit := range item.Material.NotBaseUnitList {
				unitList = append(unitList, resp.PurchaseOrderItemMaterialUnit{
					Uuid: unit.Uuid,
					LocaleName: func() dto.LocaleResponse {
						if unit.Unit == nil {
							return dto.LocaleResponse{}
						}
						return unit.Unit.MultiLanguageName.GetNames()
					}(),
				})
			}
			return unitList
		}(item)
		// 单位列表
		itemInfo.Units = func(item model.PurchaseReceiptOrderItem) []resp.PurchaseOrderItemUnit {
			unitList := []resp.PurchaseOrderItemUnit{}
			if len(item.Units) == 0 {
				unitList = append(unitList, resp.PurchaseOrderItemUnit{
					Num:        item.Num,
					ArrivalNum: item.Num,
					UnitUuid:   item.UnitUuid,
					LocaleName: *language.JsonToLocaleResponse(item.UnitName),
				})
			} else {
				for _, unit := range item.Units {
					unitList = append(unitList, resp.PurchaseOrderItemUnit{
						Num:        unit.Num,
						ArrivalNum: unit.Num,
						UnitUuid:   unit.UnitUuid,
						LocaleName: func() dto.LocaleResponse {
							if item.Material == nil {
								return *language.JsonToLocaleResponse(unit.UnitName)
							}
							if len(item.Material.NotBaseUnitList) == 0 {
								return *language.JsonToLocaleResponse(unit.UnitName)
							}
							for _, materialUnit := range item.Material.NotBaseUnitList {
								if materialUnit.Uuid == unit.UnitUuid {
									if materialUnit.Unit == nil {
										return *language.JsonToLocaleResponse(materialUnit.Name)
									}
									return materialUnit.Unit.MultiLanguageName.GetNames()
								}
							}
							return *language.JsonToLocaleResponse(unit.UnitName)
						}(),
					})
				}
			}
			return unitList
		}(item)

		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// CancelPurchaseReceiptOrder 取消收货单
func (s *purchaseReceiptOrderSrv) CancelPurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderCancelReq,
) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 查询收货单
	receiptOrder, err := receiptOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("收货单不存在"), err.Error())
	}

	// 检查收货单状态
	if receiptOrder.Status != constant.ReceiptOrderStatusPending {
		return errors.New("收货单状态不允许取消")
	}

	// 取消收货单
	receiptOrder.Status = constant.ReceiptOrderStatusRejected
	receiptOrder.CancelTime = time.Now().Unix()
	err = receiptOrderRepo.Update(receiptOrder)
	if err != nil {
		return errors.WithMessage(errors.New("取消收货单失败"), err.Error())
	}

	return nil
}

// updateMaterialStock 更新物料库存
func (s *purchaseReceiptOrderSrv) updateMaterialStock(
	ctx context.Context,
	db *gorm.DB,
	receiptOrder *model.PurchaseReceiptOrder,
) error {
	if receiptOrder.Status != constant.ReceiptOrderStatusReceived {
		return nil
	}

	// 记录erp的入库记录
	err := s.helper.recordErpStockInLog(db, receiptOrder)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 获取供应商ID
	supplierUuid := func() uint64 {
		supplier, err := repository.NewSupplierRepo(db).GetByErpCode(receiptOrder.GetSupplierErpCode())
		if err != nil {
			return 0
		}
		return supplier.Uuid
	}()

	// 更新在途仓库库存
	transitWarehouse, _ := repository.NewWarehouseRepo(db).GetTransitWarehouse()
	if transitWarehouse != nil {
		warehouseItemRepo := repository.NewWarehouseItemRepo(db)
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)
		for _, item := range receiptOrder.Items {
			actualNum := item.GetUnitsTotalConversionRateNum()
			if actualNum <= 0 {
				continue
			}
			// 获取物料信息
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
			if err != nil || warehouseItem == nil {
				continue
			}
			if warehouseItem.Stock < actualNum {
				actualNum = warehouseItem.Stock
			}
			// 减少在途仓库库存
			err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
			if err != nil {
				return errors.WithMessage(errors.New("减少在途仓库库存失败"), err.Error())
			}
			// 记录在途仓出库日志
			warehouseLog := &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeOut,      // 出库
				Scene:                constant.WarehouseInOutLogSceneTransitOut, // 在途出库
				WarehouseUuid:        transitWarehouse.Uuid,
				MaterialUuid:         item.MaterialUuid,
				MaterialName:         item.MaterialName,
				MaterialBaseUnitUuid: item.BaseUnitUuid,
				MaterialBaseUnitName: item.BaseUnitName,
				Num:                  actualNum,
				Price:                item.Valuation,
				Amount: decimal.NewFromFloat(item.Valuation).
					Mul(decimal.NewFromFloat(actualNum)).
					InexactFloat64(),
				SupplierUuid:    supplierUuid,
				SupplierErpCode: receiptOrder.GetSupplierErpCode(),
				SupplierName:    receiptOrder.SupplierName,
				OrderNo:         receiptOrder.OrderNo,
				OtherOrgUuid:    supplierUuid,
				OtherOrgType:    0,
				OtherOrgName:    receiptOrder.SupplierName,
			}
			err = warehouseLogRepo.Create(warehouseLog)
			if err != nil {
				return errors.WithMessage(errors.New("记录在途仓出库日志失败"), err.Error())
			}
		}
	}

	// 调用erp接口
	if ctx.GetCompany().IsOpenErp() {
		// 调用erp接口
		erpReq := buying.SavePurchaseReceiptReq{
			PurchaseOrderName: receiptOrder.PurchaseOrder.ErpOrderNo,
			Items:             make([]*buying.PurchaseOrderItem, 0, len(receiptOrder.Items)),
		}
		for _, item := range receiptOrder.Items {
			if item.GetUnitsTotalConversionRateNum() <= 0 {
				continue
			}
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					erpReq.Items = append(erpReq.Items, &buying.PurchaseOrderItem{
						ItemCode: item.MaterialCode,
						ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
						StockUom: language.JsonToLocaleResponse(unit.UnitName).EN,
						Qty:      unit.Num,
					})
				}
			} else if item.Num > 0 {
				erpReq.Items = append(erpReq.Items, &buying.PurchaseOrderItem{
					ItemCode: item.MaterialCode,
					ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
					StockUom: language.JsonToLocaleResponse(item.UnitName).EN,
					Qty:      item.Num,
				})
			}
		}
		resp, err := erp.NewIErpSrv(s.dbm).SavePurchaseReceipt(ctx, &erpReq)
		if err != nil {
			return err
		}
		receiptOrder.ErpOrderNo = resp.PurchaseReceipt.PurchaseReceiptName
		err = repository.NewPurchaseReceiptOrderRepo(db).Update(receiptOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新收货单号失败"), err.Error())
		}
	}

	return nil
}
