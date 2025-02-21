package model

// ProductCategory 商品类别 ttpos_product_category
type ProductCategory struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:名称"`
	ParentUuid            uint64 `gorm:"default:0;column:parent_uuid;comment:父级ID"`
	IsSpecial             uint   `gorm:"default:0;column:is_special;comment:是否特殊, 0-否、1-是"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:多语言名称ID"`
	Status                uint   `gorm:"default:1;column:status;comment:状态, 1-开启、0-关闭"`
	Sort                  uint   `gorm:"default:0;column:sort;comment:排序"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// 获取一级分类uuid
func (model *ProductCategory) GetFirstCategoryUuid() uint64 {
	// 如果没有父级uuid，则该分类时一级
	if model.ParentUuid == 0 {
		return model.Uuid
	}
	return model.ParentUuid
}
