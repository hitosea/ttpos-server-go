package model

// 商品规格表，定义商品的规格信息
type ProductFlavor struct {
	Id                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// 商品单位表，定义商品的单位信息
type ProductUnit struct {
	Id                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'单位名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// 打印标签表，定义打印标签的相关信息
type PrinterTag struct {
	Id         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid       uint   `gorm:"default:0;comment:'UUID'"`
	Name       string `gorm:"default:'';comment:'名称'"`
	RefCount   uint   `gorm:"default:0;comment:'引用计数'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// 产品属性组表，定义产品的属性分组信息
type ProductAttributeGroup struct {
	Id                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// 商品属性表，定义商品的属性信息
type ProductAttribute struct {
	Id                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                  uint   `gorm:"default:0;comment:'UUID'"`
	Name                  string `gorm:"default:'';comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:'多语言名称ID'"`
	AttributeGroupUuid    uint   `gorm:"default:0;comment:'属性组ID'"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// 产品包属性组表，定义产品包的属性分组信息
type ProductPackageAttributeGroup struct {
	Id                 uint  `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid               uint  `gorm:"default:0;comment:'UUID'"`
	IsMust             bool  `gorm:"default:false;comment:'是否必选, 0-否 1-是'"`
	MaxSelection       uint  `gorm:"default:0;comment:'最大选择数量'"`
	ProductPackageUuid uint  `gorm:"default:0;comment:'产品包ID'"`
	CreateTime         int64 `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime         int64 `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime         int64 `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// 产品包属性表，定义产品包的属性信息
type ProductPackageAttribute struct {
	Id                               uint  `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                             uint  `gorm:"default:0;comment:'UUID'"`
	ProductPackageAttributeGroupUuid uint  `gorm:"default:0;comment:'产品包属性组ID'"`
	AttributeUuid                    uint  `gorm:"default:0;comment:'产品属性ID'"`
	IsDefaultSelected                bool  `gorm:"default:false;comment:'是否默认选中, 0-否 1-是'"`
	CreateTime                       int64 `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime                       int64 `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime                       int64 `gorm:"default:0;comment:'删除时间（时间戳）'"`

	Attribute ProductAttribute `gorm:"foreignKey:attribute_uuid;references:uuid"` // 产品属性
}
