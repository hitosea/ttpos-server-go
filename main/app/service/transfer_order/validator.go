package transfer_order

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"

	"gorm.io/gorm"
)

// transferOrderValidator 调拨单验证器
type transferOrderValidator struct{}

// validateMaterialStatus 验证物料状态
// allowDisabled: true-允许禁用物料并返回警告, false-直接返回错误
func (v *transferOrderValidator) validateMaterialStatus(
	ctx context.Context,
	db *gorm.DB,
	items []req.TransferOrderItemCreateReq,
	allowDisabled bool,
) (material []model.Material, disabledMaterials []string, err error) {
	materialRepo := repository.NewMaterialRepo(db)
	var materials []model.Material

	// 获取物品列表
	for _, item := range items {
		material, err := materialRepo.GetMaterialByUuid(
			item.MaterialUuid,
			materialRepo.WithNotBaseUnitList(),
		)
		if allowDisabled && err != nil {
			return nil, nil, errors.WithMessage(errors.New("物品不存在"), err.Error())
		} else if !allowDisabled && (err != nil || !material.Status) {
			materialName := language.JsonToLocaleResponse(material.Name).GetLocale(ctx.GetLanguage())
			disabledMaterials = append(disabledMaterials, materialName)
		}
		materials = append(materials, material)
	}

	// 如果有禁用的物料
	if len(disabledMaterials) > 0 {
		if allowDisabled {
			return materials, disabledMaterials, nil
		}
		// 不允许禁用物料，返回错误
		materialNames := v.joinMaterialNames(disabledMaterials)
		return materials, disabledMaterials, errors.NewWithCodeAndData(
			constant.CodeMaterialDisabled,
			disabledMaterials,
			fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 的状态已关闭。\n\n请修改物品状态"), materialNames),
		)
	}

	return materials, disabledMaterials, nil
}

// joinMaterialNames 拼接物料名称
func (v *transferOrderValidator) joinMaterialNames(names []string) string {
	result := ""
	for i, name := range names {
		if i > 0 {
			result += "、"
		}
		result += name
	}
	return result
}
