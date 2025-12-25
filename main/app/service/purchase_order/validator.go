package purchase_order

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// purchaseOrderValidator 采购订单验证器
type purchaseOrderValidator struct{}

// validateMaterialStatus 验证物料状态
// allowDisabled: true-允许禁用物料并返回警告, false-直接返回错误
func (v *purchaseOrderValidator) validateMaterialStatus(
	ctx context.Context,
	db *gorm.DB,
	items []model.PurchaseOrderItem,
	allowDisabled bool,
) ([]string, error) {
	materialRepo := repository.NewMaterialRepo(db)
	var disabledMaterials []string

	for _, item := range items {
		if item.Num <= 0 {
			continue
		}
		material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
		if err != nil || !material.Status {
			materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(ctx.GetLanguage())
			disabledMaterials = append(disabledMaterials, materialName)
		}
	}

	// 如果有禁用的物料
	if len(disabledMaterials) > 0 {
		if allowDisabled {
			return disabledMaterials, nil
		}
		// 不允许禁用物料，返回错误
		materialNames := v.joinMaterialNames(disabledMaterials)
		return disabledMaterials, errors.NewWithCodeAndData(
			constant.CodeMaterialDisabled,
			disabledMaterials,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 的状态已关闭。\n\n请修改物品状态"), materialNames),
		)
	}

	return nil, nil
}

// validateReceiptMaterialStatus 验证收货单物料状态
func (v *purchaseOrderValidator) validateReceiptMaterialStatus(
	ctx context.Context,
	db *gorm.DB,
	itemUuids []uint64,
) error {
	purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)
	materialRepo := repository.NewMaterialRepo(db)
	var disabledMaterials []string

	for _, itemUuid := range itemUuids {
		purchaseOrderItem, err := purchaseOrderItemRepo.GetByUuid(itemUuid)
		if err != nil {
			return errors.WithMessage(errors.New("查询采购申请明细失败"), err.Error())
		}

		material, err := materialRepo.GetMaterialByUuid(purchaseOrderItem.MaterialUuid)
		if err != nil {
			return errors.WithMessage(errors.New("查询物品明细失败"), err.Error())
		}

		// 判断物品是否停用
		if !material.Status {
			materialName := language.JsonToLocaleResponse(purchaseOrderItem.MaterialName).GetLocale(ctx.GetLanguage())
			disabledMaterials = append(disabledMaterials, materialName)
		}
	}

	// 如果有停用的物品，返回相应的错误消息
	if len(disabledMaterials) > 0 {
		return errors.NewWithCode(
			constant.CodeMaterialDisabled,
			fmt.Sprintf(
				i18n.Translate(ctx.GetLanguage(), "有%d项物品已停用，您可启用物品后再进行收货"),
				len(disabledMaterials),
			),
		)
	}

	return nil
}

// validateSupplierStatus 验证供应商状态
func (v *purchaseOrderValidator) validateSupplierStatus(
	db *gorm.DB,
	supplierErpCode string,
) error {
	// 跳过空供应商或总部供应商
	if supplierErpCode == "" || supplierErpCode == constant.ErpHeadquartersSupplierCode {
		return nil
	}

	supplier, err := repository.NewSupplierRepo(db).GetByErpCode(supplierErpCode)
	if err != nil {
		return errors.WithMessage(errors.New("供应商不存在,请先同步供应商数据"), err.Error())
	}

	if supplier.Status == 0 {
		return errors.NewWithCode(
			constant.CodePurchaseOrderSupplierDisabled,
			"供应商已禁用，请修改供应商状态",
		)
	}

	return nil
}

// // validateReceiptQuantity 验证收货数量
// func (v *purchaseOrderValidator) validateReceiptQuantity(
// 	ctx context.Context,
// 	orderItem *model.PurchaseOrderItem,
// 	receiptNum float64,
// ) error {
// 	newArrivalNum := orderItem.ArrivalNum + receiptNum
// 	if newArrivalNum > orderItem.Num {
// 		return errors.New(
// 			fmt.Sprintf(
// 				i18n.Translate(ctx.GetLanguage(), "物品 %s 的收货数量不能超过申请数量（申请数量：%.0f，已到货：%.0f，本次收货：%.0f）"),
// 				language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(ctx.GetLanguage()),
// 				orderItem.Num,
// 				orderItem.ArrivalNum,
// 				receiptNum,
// 			),
// 		)
// 	}
// 	return nil
// }

// validateReceiptQuantity 验证收货数量
func (v *purchaseOrderValidator) validateReceiptQuantityNew(
	ctx context.Context,
	orderItem *model.PurchaseOrderItem,
	unitList []req.PurchaseReceiptItemMaterialUnitReq,
) (error, float64) {
	reqNum := 0.0
	if len(orderItem.Units) == 0 && orderItem.BaseUnitUuid != 0 {
		for _, unit := range unitList {
			reqNum += unit.Num
		}
		if reqNum > orderItem.Num {
			return errors.New(fmt.Sprintf("物品 %s 的收货数量不能超过申请数量（申请数量：%.0f，已到货：%.0f，本次收货：%.0f）", language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(ctx.GetLanguage()), orderItem.Num, orderItem.ArrivalNum, reqNum)), reqNum
		}
	} else {
		for _, unit := range unitList {
			for _, orderItemUnit := range orderItem.Units {
				if orderItemUnit.UnitUuid == unit.Uuid {
					newArrivalNum := orderItemUnit.ArrivalNum + unit.Num
					if newArrivalNum > orderItemUnit.Num {
						return errors.New(
							fmt.Sprintf(
								i18n.Translate(ctx.GetLanguage(), "物品 %s 的收货数量不能超过申请数量（申请数量：%.0f，已到货：%.0f，本次收货：%.0f）"),
								language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(ctx.GetLanguage()),
								orderItemUnit.Num,
								orderItemUnit.ArrivalNum,
								unit.Num,
							),
						), reqNum
					}
					reqNum += unit.Num
				}
			}
		}
	}
	return nil, reqNum
}

// joinMaterialNames 拼接物料名称
func (v *purchaseOrderValidator) joinMaterialNames(names []string) string {
	result := ""
	for i, name := range names {
		if i > 0 {
			result += "、"
		}
		result += name
	}
	return result
}

// PurchaseOrderItemReq 采购订单明细请求（临时定义，实际应该在dto/req包中）
type PurchaseOrderItemReq struct {
	MaterialUuid uint64                                 `json:"material_uuid"`
	UnitList     []req.PurchaseOrderItemMaterialUnitReq `json:"unit_list"`
}

// buildPurchaseOrderItems 构建采购订单明细
func (v *purchaseOrderValidator) buildPurchaseOrderItems(
	db *gorm.DB,
	purchaseOrderUuid uint64,
	itemReqs []PurchaseOrderItemReq,
) ([]model.PurchaseOrderItem, []model.PurchaseOrderItemUnit, error) {
	// 批量查询物料
	materialUuids := make([]uint64, 0, len(itemReqs))
	for _, itemReq := range itemReqs {
		materialUuids = append(materialUuids, itemReq.MaterialUuid)
	}

	materialRepo := base.NewMaterialRepo(db)
	materialList, err := materialRepo.GetMaterialByUuids(
		materialUuids,
		materialRepo.WithPreload("Unit.Unit"),
		materialRepo.WithPreload("NotBaseUnitList.Unit"),
	)
	if err != nil {
		return nil, nil, errors.WithMessage(errors.New("查询物品失败"), err.Error())
	}

	// 转换为map方便查询
	materials := make(map[uint64]*model.Material, len(materialList))
	for _, material := range materialList {
		materials[material.Uuid] = material
	}

	// 验证所有物料存在
	for _, itemReq := range itemReqs {
		if _, exists := materials[itemReq.MaterialUuid]; !exists {
			return nil, nil, errors.New(fmt.Sprintf("物品UUID %d 不存在", itemReq.MaterialUuid))
		}
	}

	// 构建明细
	items := make([]model.PurchaseOrderItem, 0, len(itemReqs))
	unitList := make([]model.PurchaseOrderItemUnit, 0, len(itemReqs))
	for _, itemReq := range itemReqs {
		material := materials[itemReq.MaterialUuid]
		item := v.buildPurchaseOrderItem(purchaseOrderUuid, material, itemReq.UnitList)
		items = append(items, item)
		unitList = append(unitList, item.Units...)
	}

	return items, unitList, nil
}

// buildPurchaseOrderItem 构建单个采购订单明细
func (v *purchaseOrderValidator) buildPurchaseOrderItem(
	purchaseOrderUuid uint64,
	material *model.Material,
	unitList []req.PurchaseOrderItemMaterialUnitReq,
) model.PurchaseOrderItem {
	itemUuid, err := utils.GetID()
	if err != nil {
		logger.Logger.Error("生成雪花ID失败", zap.Error(err))
		return model.PurchaseOrderItem{}
	}
	totalNum := 0.0
	purchaseOrderItemUnits := make([]model.PurchaseOrderItemUnit, 0)
	for _, reqUnit := range unitList {
		totalNum += reqUnit.Num
		for _, materialUnit := range material.NotBaseUnitList {
			if materialUnit.Uuid == reqUnit.Uuid {
				purchaseOrderItemUnits = append(purchaseOrderItemUnits, model.PurchaseOrderItemUnit{
					ItemUuid:           itemUuid,
					PurchaseOrderUuid:  purchaseOrderUuid,
					UnitUuid:           reqUnit.Uuid,
					Num:                reqUnit.Num,
					UnitName:           materialUnit.Name,
					UnitConversionRate: materialUnit.ConversionRate,
					ErpnextUom:         materialUnit.Unit.ErpnextUom,
					BaseUnitUuid:       material.UnitUuid,
					BaseUnitName:       material.Unit.Name,
				})
				break
			}
		}
	}

	item := model.PurchaseOrderItem{
		BaseModel: model.BaseModel{
			Uuid: itemUuid,
		},
		PurchaseOrderUuid: purchaseOrderUuid,
		MaterialUuid:      material.Uuid,
		MaterialCode:      material.Code,
		MaterialName:      material.Name,
		Num:               totalNum,
		UnitUuid:          material.PurchaseUnitUuid,
		Valuation:         0, // TODO v2.12.0: ttpos测没有估值率的值,若需要请调用erp接口获取
		Units:             purchaseOrderItemUnits,
	}

	// 计算总价
	item.TotalPrice = decimal.NewFromFloat(item.Valuation).
		Mul(decimal.NewFromFloat(totalNum)).
		InexactFloat64()

	return item
}
