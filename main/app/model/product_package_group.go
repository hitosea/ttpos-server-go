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
