package model

// ProductCategory 商品类别 ttpos_product_category
type ProductCategory struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
	Source                string `gorm:"column:source;type:varchar(50);default:'';comment:来源标记(grab/manual等);NOT NULL" json:"source"`
	SourceId              string `gorm:"column:source_id;type:varchar(191);default:'';comment:来源平台的分类ID;NOT NULL;index:idx_source_id" json:"source_id"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	Status                int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 1-开启 0-关闭;NOT NULL" json:"status"`
	IsDisplayInStore      int    `gorm:"column:is_display_in_store;type:tinyint(1);default:1;comment:是否在店内显示;NOT NULL" json:"is_display_in_store"`
	IsDisplayInTakeout    int    `gorm:"column:is_display_in_takeout;type:tinyint(1);default:0;comment:是否在外卖平台显示;NOT NULL" json:"is_display_in_takeout"`
	ParentUuid            uint64 `gorm:"column:parent_uuid;type:bigint(20) unsigned;comment:父级ID" json:"parent_uuid"`
	IsSpecial             int    `gorm:"column:is_special;type:tinyint(1);default:0;comment:特殊分类, 1-是 0-否;NOT NULL" json:"is_special"`
	CategoryKey           string `gorm:"column:category_key;type:varchar(255);comment:关键字;NOT NULL" json:"category_key"`
	Sort                  uint   `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
	Code                  string `gorm:"column:code;type:varchar(255);comment:分类编码;NOT NULL" json:"code"`
	HeadquarterUuid       uint64 `gorm:"column:headquarter_uuid;type:bigint(20) unsigned;default:0;comment:总部ID;NOT NULL" json:"headquarter_uuid"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// GetFirstCategoryUuid 获取一级分类uuid
func (model *ProductCategory) GetFirstCategoryUuid() uint64 {
	// 如果没有父级uuid，则该分类时一级
	if model.ParentUuid == 0 {
		return model.Uuid
	}
	return model.ParentUuid
}

func (model *ProductCategory) IsEditable() bool {
	return model.HeadquarterUuid == 0
}
