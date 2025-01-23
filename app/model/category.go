package model

// 商品类别
type ProductCategory struct {
	Id                  uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Name                string `gorm:"not null;default:'';comment:名称" json:"name"`
	ParentId            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_id"`
	MultiLanguageNameId uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_id"`
	Status              string `gorm:"not null;default:'';comment:状态, open开启、close关闭" json:"status"`
	Level               uint   `gorm:"not null;default:0;comment:层级" json:"level"`
	CategoryKey         string `gorm:"not null;default:'';comment:关键字" json:"category_key"`
	OrderBy             uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	CreateTime          uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime          uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime          uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`
}

// 原料类别
type MaterialCategory struct {
	Id                  uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Name                string `gorm:"not null;default:'';comment:名称" json:"name"`
	MultiLanguageNameId uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_id"`
	Status              string `gorm:"not null;default:'';comment:状态, open开启、close关闭" json:"status"`
	Level               uint   `gorm:"not null;default:0;comment:层级" json:"level"`
	ParentId            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_id"`
	CategoryKey         string `gorm:"not null;default:'';comment:关键字" json:"category_key"`
	OrderBy             uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	RefCount            uint   `gorm:"not null;default:0;comment:关联数量" json:"ref_count"`
	CreateTime          uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime          uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime          uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`
}

// 商品特殊类别
type ProductSpecialCategory struct {
	Id                  uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Status              string `gorm:"not null;default:'';comment:状态, open开启、close关闭" json:"status"`
	Name                string `gorm:"not null;default:'';comment:名称" json:"name"`
	MultiLanguageNameId uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_id"`
	OrderBy             uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	RefCount            uint   `gorm:"not null;default:0;comment:关联数量" json:"ref_count"`
	CreateTime          uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime          uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime          uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`
}
