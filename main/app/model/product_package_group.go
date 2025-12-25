package model

// ProductPackageGroup 商品套餐组模型 `ttpos_product_package_group`
type ProductPackageGroup struct {
	BaseModel
	Name                  string `json:"name" gorm:"type:text;comment:名称"`
	MultiLanguageNameUuid uint64 `json:"multi_language_name_uuid" gorm:"index:idx_multi_language_name_uuid;not null;default:0;comment:多语言名称ID"`
	ProductPackageUuid    uint64 `json:"product_package_uuid" gorm:"index:idx_product_package_uuid;not null;default:0;comment:商品套餐UUID"`
	GroupType             int    `json:"group_type" gorm:"type:tinyint;not null;default:0;comment:分组类型 0-固定 1-可选"`
	OptionalMinCount      int    `json:"optional_min_count" gorm:"type:int;not null;default:0;comment:最小可选数量"`
	OptionalCount         int    `json:"optional_count" gorm:"type:int;not null;default:0;comment:最大可选数量，表示本组商品中最多可以选择多少个商品"`
	Sort                  int    `json:"sort" gorm:"type:int;not null;default:0;comment:排序字段，数值越小越靠前"` // 排序字段

	ProductPackageGroupItems []ProductPackageGroupItem `gorm:"foreignKey:product_package_group_uuid;references:uuid"` // 商品套餐组商品
	MultiLanguageName        MultiLanguageName         `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`   // 多语言名称
	ProductPackage           *ProductPackage           `gorm:"foreignKey:product_package_uuid;references:uuid"`       // 商品套餐
}

// 获取最小可选数量、最大可选数量
// 返回合理的范围，确保 min <= max，兼容 min 或 max 为 0 的情况
func (model *ProductPackageGroup) GetOptionalRange() (int, int) {
	min := model.OptionalMinCount
	max := model.OptionalCount

	// 确保非负数
	if min < 0 {
		min = 0
	}
	if max < 0 {
		max = 0
	}

	// 如果 max 为 0，表示不能选任何，min 也应该为 0
	if max == 0 {
		min = 0
		return min, max
	}

	// 确保 min <= max
	if min > max {
		min = max
	}

	return min, max
}
