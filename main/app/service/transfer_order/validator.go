package transfer_order

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// transferOrderValidator 调拨单验证器
type transferOrderValidator struct{}

// 验证单位数量是否为0
func (v *transferOrderValidator) validateOrderItemUnitNumZero(
	ctx context.Context,
	items []*model.TransferOrderItem,
) error {
	var itemNames []string
	for _, item := range items {
		if item.GetUnitsTotalConversionRateNum() == 0 {
			itemNames = append(itemNames, item.MaterialName)
		}
	}
	if len(itemNames) > 0 {
		return errors.NewWithCode(
			constant.CodeErrorConfirmClose,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "单据存在调拨数量为0的物品%d个，提交后将移除该物品，是否继续提交？"), len(itemNames)),
		)
	}
	return nil
}

// 验证物品列表
func (v *transferOrderValidator) ValidateMaterialsByCodes(
	ctx ttposContext.Context,
	db *gorm.DB,
	items []*model.TransferOrderItem,
) (err error) {
	if len(items) == 0 {
		return nil
	}
	materialRepo := repository.NewMaterialRepo(db)

	// 收集所有的 MaterialUuid（去重）
	materialCodeSet := make(map[string]bool)
	materialCodes := make([]string, 0, len(items))
	for _, item := range items {
		if !materialCodeSet[item.MaterialCode] {
			materialCodeSet[item.MaterialCode] = true
			materialCodes = append(materialCodes, item.MaterialCode)
		}
	}

	// 批量查询所有材料
	dbMaterials, err := materialRepo.GetMaterialContainsDeletedByCodes(materialCodes, materialRepo.WithNotBaseUnitList())
	if err != nil {
		return errors.WithMessage(errors.New("批量查询物品失败"), err.Error())
	}

	// 构建 map 用于快速查找
	materialMap := make(map[string]*model.Material, len(dbMaterials))
	unitMap := make(map[string]*model.MaterialUnit, len(dbMaterials))
	for i := range dbMaterials {
		materialMap[dbMaterials[i].Code] = dbMaterials[i]
		for _, unit := range dbMaterials[i].NotBaseUnitList {
			unitMap[unit.Unit.ErpnextUom] = unit
		}
	}

	notFoundMaterialNames := make([]string, 0)
	disabledMaterialNames := make([]string, 0)
	notFoundUnitNames := make([]string, 0)
	// 按照原始 items 的顺序返回结果，并验证所有材料都存在
	for _, item := range items {
		material, exists := materialMap[item.MaterialCode]
		if !exists {
			notFoundMaterialNames = append(notFoundMaterialNames, func() string {
				if material != nil && material.Name != "" {
					return language.JsonToLocaleResponse(material.Name).GetLocale(ctx.GetLanguage())
				}
				return fmt.Sprintf("%s", item.MaterialCode)
			}())
		} else {
			if material.DeleteTime > 0 {
				notFoundMaterialNames = append(notFoundMaterialNames, language.JsonToLocaleResponse(material.Name).GetLocale(ctx.GetLanguage()))
			} else if !material.Status {
				notFoundMaterialNames = append(notFoundMaterialNames, language.JsonToLocaleResponse(material.Name).GetLocale(ctx.GetLanguage()))
			} else {
				for _, unit := range item.Units {
					unit, exists := unitMap[unit.ErpnextUom]
					if !exists {
						notFoundUnitNames = append(notFoundUnitNames, unit.Name)
					}
				}
			}
		}

	}
	if len(disabledMaterialNames) > 0 {
		return errors.NewWithCodeAndData(
			constant.CodeErrorConfirmClose,
			disabledMaterialNames,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 的状态已关闭。\n\n请修改物品状态"), v.joinNames(disabledMaterialNames)),
		)
	}
	if len(notFoundMaterialNames) > 0 {
		return errors.NewWithCodeAndData(
			constant.CodeErrorConfirmClose,
			notFoundMaterialNames,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 未找到。"), v.joinNames(notFoundMaterialNames)),
		)
	}
	if len(notFoundUnitNames) > 0 {
		return errors.NewWithCodeAndData(
			constant.CodeErrorConfirmClose,
			notFoundUnitNames,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "单位 %s 未找到。"), v.joinNames(notFoundUnitNames)),
		)
	}
	return nil
}

// joinNames 拼接物料名称
func (v *transferOrderValidator) joinNames(names []string) string {
	result := ""
	for i, name := range names {
		if i > 0 {
			result += "、"
		}
		result += name
	}
	return result
}

// validateOrderItemStockNotEnough 验证物品库存不足
func (v *transferOrderValidator) validateOrderItemStockNotEnough(
	ctx context.Context,
	dbm *database.DBManager,
	transferOrder *model.TransferOrder,
) ([]string, error) {
	otherDb := dbm.GetDB(transferOrder.SenderCompanyUuid)
	if otherDb == nil {
		return nil, errors.WithMessage(errors.New("获取其他店数据失败"), "其他店数据库不存在")
	}

	// 当前库存
	erpCodes := []string{}
	for _, item := range transferOrder.Items {
		erpCodes = append(erpCodes, item.MaterialCode)
	}
	warehouseItemList, err := repository.NewWarehouseItemRepo(otherDb).GetNormalByMaterialCodes(erpCodes, transferOrder.OutWarehouseErpCode)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取物品库存失败"), err.Error())
	}

	// 计算可用库存
	var notEnoughItemNames []string
	for _, item := range transferOrder.Items {
		availableNum := decimal.NewFromFloat(0)
		for _, warehouseItem := range warehouseItemList {
			if item.MaterialCode == warehouseItem.MaterialCode {
				availableNum = availableNum.Add(availableNum)
			}
		}
		if availableNum.LessThan(decimal.NewFromFloat(item.GetUnitsTotalConversionRateNum())) {
			notEnoughItemNames = append(notEnoughItemNames, item.MaterialName)
		}
	}
	if len(notEnoughItemNames) > 0 {
		return notEnoughItemNames, errors.NewWithCode(
			constant.CodeErrorConfirmClose,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 的可调拨数量不足。\n\n请更换出库仓库"), v.joinNames(notEnoughItemNames)),
		)
	}
	return notEnoughItemNames, nil
}
