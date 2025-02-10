package model

type ReturnFoodReason struct {
	ID                    uint   `gorm:"primaryKey;autoIncrement;comment:'退菜原因唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

type GiftOrFreeOrderReason struct {
	ID                    uint   `gorm:"primaryKey;autoIncrement;comment:'赠品或免费订单原因唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
