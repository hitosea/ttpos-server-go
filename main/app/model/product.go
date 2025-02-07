package model

// 商品规格表，定义商品的规格信息
type ProductFlavor struct {
	Id                    uint   `gorm:"primaryKey;autoIncrement;comment:'记录唯一标识符'"`
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
	Id                    uint   `gorm:"primaryKey;autoIncrement;comment:'记录唯一标识符'"`
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

// 产品包表，定义产品包的相关信息
type ProductPackage struct {
	ID                         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                       uint64 `gorm:"column:uuid;not null;default:0;comment:'UUID'"`
	Name                       string `gorm:"column:name;not null;default:'';comment:'产品包名称'"`
	MultiLanguageNameUuid      uint64 `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称ID'"`
	ImageName                  string `gorm:"column:image_name;not null;default:'';comment:'图片名称'"`
	ImageUrl                   string `gorm:"column:image_url;not null;default:'';comment:'图片URL'"`
	InventoryCalculationMethod uint   `gorm:"column:inventory_calculation_method;not null;default:0;comment:'库存计算方法, 0-下单减库存 1-付款减库存'"`
	UnitUuid                   uint64 `gorm:"column:unit_id;not null;default:0;comment:'单位ID'"`
	DineTaxUuid                uint64 `gorm:"column:dine_tax_id;not null;default:0;comment:'堂食税ID'"`
	CategoryKey                string `gorm:"column:category_key;not null;default:'';comment:'类别关键字'"`
	CategoryUuid               uint64 `gorm:"column:category_id;not null;default:0;comment:'类别ID'"`
	TakeoutTaxUuid             uint64 `gorm:"column:takeout_tax_id;not null;default:0;comment:'外卖税ID'"`
	SpecialCategoryUuid        uint64 `gorm:"column:special_category_id;not null;default:0;comment:'特殊类别ID'"`
	PrinterTagUuid             uint64 `gorm:"column:printer_tag_id;not null;default:0;comment:'打印机标签ID'"`
	Status                     uint   `gorm:"column:status;not null;default:0;comment:'状态, 0-上架 1-下架'"`
	DeviceCashier              uint   `gorm:"column:device_cashier;not null;default:false;comment:'是否在收银设备显示, 0-否 1-是'"`
	DeviceTablet               uint   `gorm:"column:device_tablet;not null;default:false;comment:'是否在平板设备显示, 0-否 1-是'"`
	DeviceKitchen              uint   `gorm:"column:device_kitchen;not null;default:false;comment:'是否在厨房设备显示, 0-否 1-是'"`
	DeviceAssistant            uint   `gorm:"column:device_assistant;not null;default:false;comment:'是否在助手设备显示, 0-否 1-是'"`
	DeviceH5                   uint   `gorm:"column:device_h5;not null;default:false;comment:'是否在H5设备显示, 0-否 1-是'"`
	OrderBy                    uint   `gorm:"column:order_by;not null;default:0;comment:'排序'"`
	LimitedPurchaseQuantity    uint   `gorm:"column:limited_purchase_quantity;not null;default:0;comment:'限购数量'"`
	Description                string `gorm:"column:description;not null;default:'';comment:'卖点描述'"`
	IsMust                     uint   `gorm:"column:is_must;not null;default:false;comment:'是否必选, 0-否 1-是'"`
	MaxSelection               uint   `gorm:"column:max_selection;not null;default:0;comment:'最大选择数量'"`
	OpenDiscount               uint   `gorm:"column:open_discount;not null;default:false;comment:'是否开启会员折扣, 0-否 1-是'"`
	CreateTime                 int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime                 int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime                 int64  `gorm:"column:delete_time;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	ProductUnit       ProductUnit       `gorm:"foreignKey:unit_uuid;references:uuid"`                // 单位
	ProductBoms       []ProductBom      `gorm:"foreignKey:product_package_uuid;references:uuid"`     // BOM
}

// 产品BOM表，定义产品BOM的相关信息
type ProductBom struct {
	ID                    uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid                  uint64  `gorm:"column:uuid;not null;default:0;comment:'UUID'"`
	Name                  string  `gorm:"column:name;not null;default:'';comment:'名称'"`
	MultiLanguageNameUuid uint64  `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称ID'"`
	Price                 float64 `gorm:"column:price;not null;default:0;comment:'价格'"`
	FlavorUuid            uint64  `gorm:"column:flavor_uuid;not null;default:0;comment:'规格ID'"`
	ProductPackageUuid    uint64  `gorm:"column:product_package_uuid;not null;default:0;comment:'产品包ID'"`
	RefCount              uint    `gorm:"column:ref_count;not null;default:0;comment:'引用计数'"`
	IsDefaultSelect       uint    `gorm:"column:is_default_select;not null;default:false;comment:'是否默认选择, 0-否 1-是'"`
	CreateTime            int64   `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime            int64   `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime            int64   `gorm:"column:delete_time;comment:'删除时间（时间戳）'"`
}
