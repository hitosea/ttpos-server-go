package model

// 商品类别
type ProductCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符"`
	Uuid                  uint64 `gorm:"default:0;comment:UUID"`
	Name                  string `gorm:"default:'';comment:名称"`
	ParentUuid            uint64 `gorm:"default:NULL;comment:父级ID"`
	IsSpecial             uint   `gorm:"default:0;comment:是否特殊, 0-否、1-是"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:多语言名称ID"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭"`
	OrderBy               uint   `gorm:"default:0;comment:排序"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// 原料类别
type MaterialCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符"`
	Uuid                  uint64 `gorm:"default:0;comment:UUID"`
	Name                  string `gorm:"default:'';comment:名称"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;comment:多语言名称ID"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭"`
	Level                 uint   `gorm:"default:0;comment:层级"`
	ParentUuid            uint64 `gorm:"default:NULL;comment:父级ID"`
	OrderBy               uint   `gorm:"default:0;comment:排序"`
	RefCount              uint   `gorm:"default:0;comment:关联数量"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// 商品特殊类别
type ProductSpecialCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符"`
	Uuid                  uint64 `gorm:"default:0;comment:UUID"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭"`
	Name                  string `gorm:"default:'';comment:名称"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;comment:多语言名称ID"`
	OrderBy               uint   `gorm:"default:0;comment:排序"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
