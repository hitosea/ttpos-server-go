package model

// ReturnFoodReason 退菜原因表 ttpos_return_food_reason
type ReturnFoodReason struct {
	ID                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'退菜原因唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;column:uuid;comment:'UUID'"`
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// FreeReason 赠品或免费订单原因表 ttpos_free_reason
type FreeReason struct {
	ID                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'赠品或免费订单原因唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;column:uuid;comment:'UUID'"`
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
