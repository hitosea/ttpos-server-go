package model

// 商品类别
type ProductCategory struct {
	ID                    uint   `gorm:"column:id;primaryKey;comment:记录唯一标识符"`
	Uuid                  uint64 `gorm:"default:0;column:uuid;comment:UUID"`
	Name                  string `gorm:"default:'';column:name;comment:名称"`
	ParentUuid            uint64 `gorm:"default:0;column:parent_uuid;comment:父级ID"`
	IsSpecial             uint   `gorm:"default:0;column:is_special;comment:是否特殊, 0-否、1-是"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:多语言名称ID"`
	Status                uint   `gorm:"default:false;column:status;comment:状态, 1-开启、0-关闭"`
	Sort                  uint   `gorm:"default:0;column:sort;comment:排序"`
	CreateTime            int64  `gorm:"autoCreateTime;column:create_time;comment:创建时间（时间戳）"`
	UpdateTime            int64  `gorm:"autoUpdateTime;column:update_time;comment:更新时间（时间戳）"`
	DeleteTime            int64  `gorm:"default:0;column:delete_time;comment:删除时间（时间戳）"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
