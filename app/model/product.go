package model

type ProductFlavor struct {
	Id                  uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Name                string `gorm:"column:name;not null;default:'';comment:'名称'"`
	MultiLanguageNameId uint   `gorm:"column:multi_language_name_id;not null;default:0;comment:'多语言名称ID'"`
	CreateTime          int    `gorm:"column:create_time;not null;default:0;comment:'创建时间（时间戳）'"`
	UpdateTime          int    `gorm:"column:update_time;not null;default:0;comment:'更新时间（时间戳）'"`
	DeleteTime          int    `gorm:"column:delete_time;not null;default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_id;references:id"`
}

type ProductUnit struct {
	Id                  uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Name                string `gorm:"column:name;not null;default:'';comment:'单位名称'"`
	MultiLanguageNameId uint   `gorm:"column:multi_language_name_id;not null;default:0;comment:'多语言名称ID'"`
	CreateTime          int    `gorm:"column:create_time;not null;default:0;comment:'创建时间（时间戳）'"`
	UpdateTime          int    `gorm:"column:update_time;not null;default:0;comment:'更新时间（时间戳）'"`
	DeleteTime          int    `gorm:"column:delete_time;not null;default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_id;references:id"`
}

type PrinterTag struct {
	Id         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Name       string `gorm:"column:name;not null;default:'';comment:'名称'"`
	RefCount   uint   `gorm:"column:ref_count;not null;default:0;comment:'引用计数'"`
	CreateTime int    `gorm:"column:create_time;not null;default:0;comment:'创建时间（时间戳）'"`
	UpdateTime int    `gorm:"column:update_time;not null;default:0;comment:'更新时间（时间戳）'"`
	DeleteTime int    `gorm:"column:delete_time;not null;default:0;comment:'删除时间（时间戳）'"`
}
