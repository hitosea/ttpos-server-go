package purchase_order

import (
	"fmt"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/file"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
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
	dbm            *database.DBManager
	validator      *purchaseOrderValidator
	helper         *purchaseOrderHelper
	receiptFileSrv service.IPurchaseReceiptFileSrv
}

// newPurchaseReceiptOrderSrv 创建收货单服务实例
func newPurchaseReceiptOrderSrv(dbm *database.DBManager) *purchaseReceiptOrderSrv {
	return &purchaseReceiptOrderSrv{
		dbm:            dbm,
		validator:      &purchaseOrderValidator{},
		helper:         &purchaseOrderHelper{},
		receiptFileSrv: service.NewPurchaseReceiptFileSrv(dbm),
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

		// 查询采购申请（包含物品和单位信息，用于校验）
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(
			req.PurchaseOrderUuid,
			purchaseOrderRepo.WithItems(),
		)
		if err != nil {
			logger.Logger.Error("CreatePurchaseReceiptOrder-GetByUuid", zap.Any("purchaseOrderUuid", req.PurchaseOrderUuid), zap.Any("err", err))
			return errors.WithMessage(errors.New("采购申请不存在"), err.Error())
		}
		if !purchaseOrder.CanReceive() {
			return errors.New("采购单状态不允许收货")
		}

		// 版本与采购单类型校验
		hasErpSaleOrderNo := purchaseOrder.ErpSaleOrderNo != ""

		// 版本 < 2.16.0 且有 ErpSaleOrderNo：提示去新版处理
		if ctx.Version(context.LT, constant.ClientVersionV2160) && hasErpSaleOrderNo {
			return errors.New("请更新软件版本再尝试")
		}

		// v2.16.0+ 新采购单（有 ErpSaleOrderNo）收货来源校验
		if ctx.Version(context.GTE, constant.ClientVersionV2160) && hasErpSaleOrderNo {
			hasSourceParam := req.DeliveryNoteNo != "" || req.SourceSupplierCode != ""

			// 集采订单必须指定收货来源（DN单号或供应商编码）
			if !hasSourceParam {
				return errors.New("集采订单必须指定收货来源（DN单号或供应商编码）")
			}

			// DN或供应商类型的详细校验
			// 构建采购单物品UUID到物品的映射
			purchaseOrderItemsMap := make(map[uint64]*model.PurchaseOrderItem)
			for i := range purchaseOrder.Items {
				item := &purchaseOrder.Items[i]
				purchaseOrderItemsMap[item.Uuid] = item
			}

			if req.DeliveryNoteNo != "" {
				// DN类型校验
				_, err := s.validator.validateDNReceipt(ctx, s.dbm, purchaseOrder, req.DeliveryNoteNo, req.Items, purchaseOrderItemsMap)
				if err != nil {
					return err
				}
			} else if req.SourceSupplierCode != "" {
				// 供应商类型校验
				err := s.validator.validateSupplierReceipt(ctx, tx, purchaseOrder, req.SourceSupplierCode, req.Items, purchaseOrderItemsMap)
				if err != nil {
					return err
				}
			}
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

		// 获取 saas 数据库连接
		saasDB := s.dbm.GetDB(constant.DefaultDB)
		if saasDB == nil {
			return errors.New("saas 数据库连接失败")
		}

		// 获取公司 UUID（使用总部 UUID 或当前公司 UUID）
		companyUuid := ctx.GetCompanySetting().HeadquarterUuid
		if companyUuid == 0 {
			companyUuid = ctx.GetCompanyUuid()
		}

		// 确定前缀和编号类型
		var prefix, numberType string
		if purchaseOrder.PurchaseType == 2 {
			// 品采收货（内部）
			prefix = "TPHY"
			numberType = constant.NumberTypeBrandReceipt
		} else {
			// 采购收货（外部）
			prefix = "PRC"
			numberType = constant.NumberTypePurchaseReceipt
		}

		// 生成收货单编号
		receiptNo, err := s.helper.generateReceiptNo(
			saasDB,
			companyUuid,
			prefix,
			numberType,
			ctx.GetCompanySetting().Timezone,
		)
		if err != nil {
			return errors.WithMessage(err, "生成收货单编号失败")
		}

		// 创建收货单
		receiptOrderUuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成雪花ID失败", zap.Error(err))
			return errors.WithMessage(err)
		}
		// 处理供应商信息（如果指定了供应商编码）
		supplierName := purchaseOrder.SupplierName
		supplierErpCode := purchaseOrder.SupplierErpCode
		if req.SourceSupplierCode != "" {
			supplier, err := repository.NewSupplierRepo(tx).GetByErpCode(req.SourceSupplierCode)
			if err == nil && supplier != nil {
				supplierErpCode = supplier.ErpCode
				supplierName = supplier.Name
			}
		}

		receiptOrder := &model.PurchaseReceiptOrder{
			BaseModel: model.BaseModel{
				Uuid: receiptOrderUuid,
			},
			OrderNo:                receiptNo,
			Status:                 utils.IfInt(req.IsConfirm, constant.ReceiptOrderStatusReceived, constant.ReceiptOrderStatusPending),
			PurchaseOrderUuid:      req.PurchaseOrderUuid,
			PurchaseOrderNo:        purchaseOrder.OrderNo,
			PurchaseTime:           purchaseOrder.OrderTime,
			Num:                    float64(len(req.Items)),
			ExpectArrivalTime:      purchaseOrder.ExpectArrivalTime,
			SupplierName:           supplierName,
			SupplierErpCode:        supplierErpCode,
			ReceiveTime:            req.ReceiveTime,
			PurchaseOrder:          *purchaseOrder,
			SourceWarehouseErpCode: purchaseOrder.WarehouseErpCode,
			SourceWarehouseName:    purchaseOrder.WarehouseName,
			TargetWarehouseErpCode: purchaseOrder.DefaultWarehouseErpCode,
			TargetWarehouseName:    purchaseOrder.DefaultWarehouseName,
			DeliveryNoteNo:         req.DeliveryNoteNo,
			IsFromDeliveryNote:     utils.IfInt(req.DeliveryNoteNo != "", 1, 0),
			IsAutoReceipt:          utils.IfInt(req.IsAutoReceipt, 1, 0),
			ReceiptType: func() int {
				if purchaseOrder.PurchaseType == 2 {
					return 2
				}
				return 1
			}(),
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
				if err == gorm.ErrRecordNotFound {
					return errors.New(fmt.Sprintf("采购申请明细不存在，物品UUID: %d", itemReq.PurchaseOrderItemUuid))
				}
				return errors.New(fmt.Sprintf("查询采购申请明细失败: %s", err.Error()))
			}

			// 计算收货数量
			reqNum := 0.0
			if err, num := s.validator.validateReceiptQuantityNew(ctx, orderItem, itemReq.UnitList); err != nil {
				return err
			} else {
				reqNum = num
			}
			if reqNum == 0 {
				continue
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
							if unit.Num == 0 {
								continue
							}
							units = append(units, model.PurchaseReceiptOrderItemUnit{
								ItemUuid:                 orderItem.Uuid,
								PurchaseReceiptOrderUuid: receiptOrder.Uuid,
								Num:                      unit.Num,
								UnitUuid:                 unit.Uuid,
								UnitName:                 orderItemUnit.UnitName,
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

				// 更新采购申请明细的到货数量
				orderItem.ArrivalNum = newArrivalNum
				orderItem.SetNil()
				err = purchaseOrderItemRepo.Update(orderItem)
				if err != nil {
					logger.Logger.Error("CreatePurchaseReceiptOrder-Update", zap.Any("orderItem", orderItem), zap.Any("err", err))
					return errors.WithMessage(errors.New("更新采购申请明细失败"), err.Error())
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
		if len(receiptItems) == 0 {
			return errors.New("收货数量不能为0")
		}

		// 创建收货单
		receiptOrder.Num = float64(len(receiptItems))
		err = receiptOrderRepo.Create(receiptOrder)
		if err != nil {
			logger.Logger.Error("CreatePurchaseReceiptOrder-Create", zap.Any("receiptOrder", receiptOrder), zap.Any("err", err))
			return errors.WithMessage(errors.New("创建收货单失败"), err.Error())
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
			err = s.updateMaterialStock(ctx, tx, receiptOrder, req.DeliveryNoteNo)
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

	// 保存附件关联（在事务外进行，避免影响主流程）
	if len(req.FileUuids) > 0 {
		if len(req.FileUuids) > 10 {
			return resp.PurchaseReceiptOrderCreateResp{}, errors.New("最多支持10个附件")
		}

		err = s.receiptFileSrv.SaveReceiptFiles(ctx, result.Uuid, req.FileUuids)
		if err != nil {
			logger.Logger.Warn("保存收货单附件失败", zap.Error(err), zap.Uint64("receiptOrderUuid", result.Uuid))
			// 不影响收货单创建，只记录日志
		}

		s.asyncUploadFilesToErp(ctx, db, result.Uuid)
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

	var status int
	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)
		purchaseOrderItemUnitRepo := repository.NewPurchaseOrderItemUnitRepo(tx)

		// 查询收货单
		receiptOrder, err := receiptOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("收货单不存在"), err.Error())
		}

		status = receiptOrder.Status

		// 查询采购申请
		purchaseOrder, err := repository.NewPurchaseOrderRepo(tx).GetByUuid(receiptOrder.PurchaseOrderUuid)
		if err != nil {
			return errors.WithMessage(err, "采购申请不存在")
		}

		// 版本与采购单类型校验
		hasErpSaleOrderNo := purchaseOrder.ErpSaleOrderNo != ""

		// 版本 < 2.16.0 且有 ErpSaleOrderNo：提示去新版处理
		if ctx.Version(context.LT, constant.ClientVersionV2160) && hasErpSaleOrderNo {
			return errors.New("请更新软件版本再尝试")
		}

		// 确认收货时的额外校验
		if req.IsConfirm {
			if !purchaseOrder.CanReceive() {
				return errors.New("采购单状态不允许收货")
			}
			receiptOrder.PurchaseOrder = *purchaseOrder
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

		// 重新创建收货明细并更新采购申请明细的到货数量
		var receiptItems []model.PurchaseReceiptOrderItem
		for _, itemReq := range req.Items {
			// 查询收货单明细
			receiptOrderItem, err := receiptOrderItemRepo.GetByUuid(itemReq.Uuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return errors.New(fmt.Sprintf("收货单明细不存在，明细UUID: %d", itemReq.Uuid))
				}
				return errors.New(fmt.Sprintf("查询收货单明细失败: %s", err.Error()))
			}

			// 查询采购申请明细
			purchaseOrderItem, err := purchaseOrderItemRepo.GetByUuid(receiptOrderItem.PurchaseOrderItemUuid, purchaseOrderItemRepo.WithPreloadUnits())
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return errors.New(fmt.Sprintf("采购申请明细不存在，明细UUID: %d", receiptOrderItem.PurchaseOrderItemUuid))
				}
				return errors.New(fmt.Sprintf("查询采购申请明细失败: %s", err.Error()))
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
				// 更新采购申请明细单位到货数量
				for _, unit := range itemReq.UnitList {
					for _, orderItemUnit := range purchaseOrderItem.Units {
						if orderItemUnit.UnitUuid == unit.Uuid {
							orderItemUnit.ArrivalNum = orderItemUnit.ArrivalNum + unit.Num
							err = purchaseOrderItemUnitRepo.Update(orderItemUnit)
							if err != nil {
								return errors.WithMessage(errors.New("更新采购申请明细单位失败"), err.Error())
							}
						}
					}
				}

				// 更新采购申请明细的到货数量
				purchaseOrderItem.ArrivalNum = newArrivalNum
				purchaseOrderItem.SetNil()
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
		receiptOrder.Num = float64(len(receiptItems))

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
			err = s.updateMaterialStock(ctx, tx, receiptOrder, receiptOrder.DeliveryNoteNo)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 处理附件（仅草稿状态可修改）
	if status == constant.ReceiptOrderStatusPending {
		// 验证附件数量
		if len(req.FileUuids) > 10 {
			return errors.New("最多支持10个附件")
		}

		// 删除旧的附件关联
		err = s.receiptFileSrv.DeleteAllReceiptFiles(ctx, req.Uuid)
		if err != nil {
			logger.Logger.Warn("删除收货单附件失败", zap.Error(err), zap.Uint64("receiptOrderUuid", req.Uuid))
			// 不影响收货单更新，只记录日志
		}

		// 保存新的附件关联
		if len(req.FileUuids) > 0 {
			err = s.receiptFileSrv.SaveReceiptFiles(ctx, req.Uuid, req.FileUuids)
			if err != nil {
				logger.Logger.Warn("保存收货单附件失败", zap.Error(err), zap.Uint64("receiptOrderUuid", req.Uuid))
				// 不影响收货单更新，只记录日志
			}
		}
	}

	s.asyncUploadFilesToErp(ctx, db, req.Uuid)

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
		return resp.PurchaseReceiptOrderListResp{}, errors.New(fmt.Sprintf("查询收货单列表失败: %s", err.Error()))
	}

	// 转换响应数据
	listResp := make([]*resp.PurchaseReceiptOrderInfo, 0, len(receipts))
	for _, receipt := range receipts {
		receiptInfo := &resp.PurchaseReceiptOrderInfo{}
		if err := copier.Copy(receiptInfo, &receipt); err != nil {
			continue
		}
		receiptInfo.IsAutoReceipt = receipt.IsAutoReceipt == 1
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
			return resp.PurchaseReceiptOrderDetailResp{}, errors.New(fmt.Sprintf("收货单不存在，收货单UUID: %d", req.Uuid))
		}
		return resp.PurchaseReceiptOrderDetailResp{}, errors.New(fmt.Sprintf("查询收货单详情失败: %s", err.Error()))
	}

	// 转换响应数据
	var detailResp resp.PurchaseReceiptOrderDetailResp
	err = copier.Copy(&detailResp, receipt)
	if err != nil {
		return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 补充收货单额外字段
	detailResp.SupplierName = receipt.SupplierName
	detailResp.LocaleWarehouseName = *language.JsonToLocaleResponse(receipt.SourceWarehouseName)
	detailResp.IsFromDeliveryNote = receipt.IsFromDeliveryNote == 1
	detailResp.IsAutoReceipt = receipt.IsAutoReceipt == 1

	// 如果是DN收货单，预先获取DN数据和同DN的已到货数据
	// key: "material_code:erpnext_uom", value: {dnQty, arrivedQty}
	type dnUnitData struct {
		DnQty      float64 // DN中的采购数量
		ArrivedQty float64 // 同DN已确认收货单的到货数量
	}
	dnUnitDataMap := make(map[string]dnUnitData)

	if receipt.IsFromDeliveryNote == 1 && receipt.DeliveryNoteNo != "" {
		// 获取DN详情
		erpSrv := erp.NewIErpSrv(s.dbm)
		targetDN, err := erpSrv.GetDeliveryNote(ctx, receipt.DeliveryNoteNo)
		if err == nil && targetDN != nil {
			// 构建DN物品单位数量映射
			for _, dnItem := range targetDN.Items {
				key := dnItem.ItemCode + ":" + dnItem.Uom
				data := dnUnitDataMap[key]
				data.DnQty += dnItem.Qty
				dnUnitDataMap[key] = data
			}

			// 获取同DN的所有已确认收货单，计算已到货数量
			sameReceiptOrders, err := receiptOrderRepo.GetList(
				receiptOrderRepo.WherePurchaseOrderUuid(receipt.PurchaseOrderUuid),
				receiptOrderRepo.WhereDeliveryNoteNo(receipt.DeliveryNoteNo),
				receiptOrderRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusReceived}),
				receiptOrderRepo.WithItems(),
			)
			if err == nil {
				for _, ro := range sameReceiptOrders {
					for _, item := range ro.Items {
						if len(item.Units) > 0 {
							for _, unit := range item.Units {
								key := item.MaterialCode + ":" + unit.ErpnextUom
								data := dnUnitDataMap[key]
								data.ArrivedQty += unit.Num
								dnUnitDataMap[key] = data
							}
						} else {
							key := item.MaterialCode + ":" + item.ErpnextUom
							data := dnUnitDataMap[key]
							data.ArrivedQty += item.Num
							dnUnitDataMap[key] = data
						}
					}
				}
			}
		}
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
		itemInfo.Specification = func(item model.PurchaseReceiptOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.Specification
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
		itemInfo.Units = func() []resp.PurchaseOrderItemUnit {
			unitList := []resp.PurchaseOrderItemUnit{}
			if len(item.Units) == 0 {
				// DN收货单：从DN数据获取采购数量和已到货数量
				purchaseNum := item.PurchaseOrderItem.Num
				arrivalNum := item.PurchaseOrderItem.ArrivalNum
				if receipt.IsFromDeliveryNote == 1 {
					key := item.MaterialCode + ":" + item.ErpnextUom
					if data, exists := dnUnitDataMap[key]; exists {
						purchaseNum = data.DnQty
						arrivalNum = data.ArrivedQty
					}
				}
				unitList = append(unitList, resp.PurchaseOrderItemUnit{
					Num:         item.Num,
					ArrivalNum:  arrivalNum,
					PurchaseNum: purchaseNum,
					UnitUuid:    item.UnitUuid,
					LocaleName:  *language.JsonToLocaleResponse(item.UnitName),
				})
			} else {
				for _, unit := range item.Units {
					purchaseNum := 0.0
					arrivalNum := 0.0
					if receipt.IsFromDeliveryNote == 1 {
						// DN收货单：从DN数据获取采购数量和已到货数量
						key := item.MaterialCode + ":" + unit.ErpnextUom
						if data, exists := dnUnitDataMap[key]; exists {
							purchaseNum = data.DnQty
							arrivalNum = data.ArrivedQty
						}
					} else {
						// 非DN收货单：从采购单物品单位获取
						for _, purchaseOrderItemUnit := range item.PurchaseOrderItem.Units {
							if purchaseOrderItemUnit.UnitUuid == unit.UnitUuid {
								purchaseNum += purchaseOrderItemUnit.Num
								arrivalNum += purchaseOrderItemUnit.ArrivalNum
							}
						}
					}
					unitList = append(unitList, resp.PurchaseOrderItemUnit{
						Num:         unit.Num,    // 当前收货单对应物品单位的数量
						ArrivalNum:  arrivalNum,  // 同DN已确认收货单的到货数量
						PurchaseNum: purchaseNum, // DN中的采购数量
						UnitUuid:    unit.UnitUuid,
						LocaleName: func() dto.LocaleResponse {
							if unit.UnitName == "" && unit.MaterialUnit != nil {
								return *language.JsonToLocaleResponse(unit.MaterialUnit.Name)
							}
							return *language.JsonToLocaleResponse(unit.UnitName)
						}(),
					})
				}
			}
			return unitList
		}()

		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	// 查询附件列表
	files, err := s.receiptFileSrv.GetReceiptFiles(ctx, req.Uuid)
	if err != nil {
		logger.Logger.Warn("查询收货单附件失败", zap.Error(err), zap.Uint64("receiptOrderUuid", req.Uuid))
		// 不影响收货单详情查询，只记录日志
		detailResp.Files = make([]resp.ReceiptFileInfo, 0)
	} else {
		detailResp.Files = files
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
	dn string,
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
			PurchaseOrderName:     receiptOrder.PurchaseOrder.ErpOrderNo,
			InterCompanyReference: dn,
			Items:                 make([]*buying.PurchaseOrderItem, 0, len(receiptOrder.Items)),
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
						Uom:      unit.ErpnextUom,
						Qty:      unit.Num,
					})
				}
			} else if item.Num > 0 {
				erpReq.Items = append(erpReq.Items, &buying.PurchaseOrderItem{
					ItemCode: item.MaterialCode,
					ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
					Uom:      item.ErpnextUom,
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

// asyncUploadFilesToErp 异步上传收货单附件到 ERP
// 在同步上下文中提取必要数据后启动异步协程
func (s *purchaseReceiptOrderSrv) asyncUploadFilesToErp(ctx context.Context, db *gorm.DB, receiptOrderUuid uint64) {
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)
	receipt, err := receiptOrderRepo.GetByUuid(receiptOrderUuid)
	if err != nil {
		logger.Logger.Error("查询收货单失败", zap.Uint64("receiptOrderUuid", receiptOrderUuid), zap.Error(err))
		return
	}
	if receipt == nil || receipt.ErpOrderNo == "" {
		return
	}

	// 在同步上下文中提取所有需要的数据，避免异步协程访问已回收的 gin.Context
	baseURL := utils.GetBaseURL(ctx.GetGin().Request)
	companyUuid := ctx.GetCompanyUuid()

	ctx2 := ctx.Copy()
	ctx2.SetDB(s.dbm.GetDB(companyUuid))

	utils.Go(func() {
		s.uploadFilesToErp(ctx2, receipt, baseURL, companyUuid)
	})
}

// uploadFilesToErp 上传收货单附件到 ERP（内部方法，由 asyncUploadFilesToErp 调用）
func (s *purchaseReceiptOrderSrv) uploadFilesToErp(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder, baseURL string, companyUuid uint64) {
	// 直接通过 Repository 查询附件，避免在异步协程中通过 Service 层访问可能已回收的 gin.Context
	db := ctx.GetDB()
	fileRepo := repository.NewPurchaseReceiptFileRepo(db)
	orderFiles, err := fileRepo.GetByReceiptOrderUuidWithFiles(receiptOrder.Uuid)
	if err != nil {
		logger.Logger.Warn("获取收货单附件列表失败",
			zap.Uint64("receiptOrderUuid", receiptOrder.Uuid),
			zap.Error(err))
		return
	}
	if len(orderFiles) == 0 {
		return
	}

	// 逐个上传文件到 ERP
	for _, f := range orderFiles {
		if f.File == nil {
			continue
		}

		fileUrl := fmt.Sprintf("%sapi/v1/passport/file_redirect?uuid=%d&company_uuid=%d",
			baseURL, f.FileUuid, companyUuid)

		uploadReq := &file.UploadFileUrlReq{
			FileUrl:  fileUrl,
			FileName: f.File.RealName,
			DocType:  "Purchase Receipt",
			DocName:  receiptOrder.ErpOrderNo,
		}

		_, err := erp.NewIErpSrv(s.dbm).UploadFileUrl(ctx, uploadReq)
		if err != nil {
			logger.Logger.Warn("上传收货单附件到ERP失败",
				zap.Uint64("receiptOrderUuid", receiptOrder.Uuid),
				zap.Uint64("fileUuid", f.FileUuid),
				zap.String("fileName", f.File.RealName),
				zap.Error(err))
			// 不影响主流程，继续上传其他文件
		}
	}
}
