package model

// BuffetPackage 自助餐套餐信息表 ttpos_buffet_package
type BuffetPackage struct {
	ID                    uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid                  uint64 `gorm:"default:0;column:uuid;comment:'自助餐ID'"`
	Name                  string `gorm:"default:'';column:name;comment:'自助餐名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:多语言名称ID"`
	Sort                  uint   `gorm:"default:0;column:sort;comment:排序顺序"`
	TaxUuid               uint64 `gorm:"default:0;column:tax_uuid;comment:税率ID"`
	IsLimitTime           uint   `gorm:"default:0;column:is_limit_time;comment:是否限时, 0-否、1-是"`
	LimitTime             uint   `gorm:"default:0;column:limit_time;comment:限时时间"`
	CanCombined           uint   `gorm:"default:0;comment:是否可组合, 0-否、1-是"`
	NonOrderingTime       uint   `gorm:"default:0;comment:不可下单时间（分钟）"`
	ReminderOrderTime     uint   `gorm:"default:0;column:reminder_order_time;comment:提醒下单时间（分钟）"`
	CreateTime            int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime            int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime            int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	MultiLanguageName        MultiLanguageName         `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	BuffetCustomerTypePrices []BuffetCustomerTypePrice `gorm:"foreignKey:buffet_package_uuid;references:uuid"`
}

// BuffetCustomerType 自助餐客户类型信息表 ttpos_buffet_customer_type
type BuffetCustomerType struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"default:0;column:uuid;comment:'自助餐客户类型ID'"`
	Name       string `gorm:"default:'';column:name;comment:'自助餐客户类型名称'"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

// BuffetCustomerTypePrice 自助餐客户类型价格信息表 ttpos_buffet_customer_type_price
type BuffetCustomerTypePrice struct {
	ID                uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid              uint64  `gorm:"default:0;column:uuid;comment:'自助餐客户类型ID'"`
	BuffetPackageUuid uint64  `gorm:"default:0;column:buffet_package_uuid;comment:自助餐套餐信息表ID"`
	CustomerTypeUuid  uint64  `gorm:"default:0;column:customer_type_uuid;comment:自助餐客户类型信息表ID"`
	Price             float64 `gorm:"column:price;default:0;comment:'价格'"`
	CreateTime        int64   `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime        int64   `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime        int64   `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	BuffetCustomerType BuffetCustomerType `gorm:"foreignKey:customer_type_uuid;references:uuid"`
}

// BuffetProduct 自助餐产品信息表 ttpos_buffet_product
type BuffetProduct struct {
	ID                 uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid               uint64 `gorm:"default:0;column:uuid;comment:'自助餐产品ID'"`
	ProductPackageUuid uint64 `gorm:"default:0;column:product_package_uuid;comment:产品包ID"`
	BuffetPackageUuid  uint64 `gorm:"default:0;column:buffet_package_uuid;comment:自助餐套餐信息表ID"`
	IsShowCashier      uint   `gorm:"default:0;column:is_show_cashier;comment:是否在收银台显示, 0-否、1-是"`
	IsShowTablet       uint   `gorm:"default:0;column:is_show_tablet;comment:是否在平板显示, 0-否、1-是"`
	IsShowKitchen      uint   `gorm:"default:0;column:is_show_kitchen;comment:是否在厨房显示, 0-否、1-是"`
	IsShowAssistant    uint   `gorm:"default:0;column:is_show_assistant;comment:是否在助手显示, 0-否、1-是"`
	Limit              uint   `gorm:"default:0;column:limit;comment:限购数量"`
	CreateTime         int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime         int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime         int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

// BuffetDelay 自助餐加钟价格表 `ttpos_buffet_delay`
type BuffetDelay struct {
	ID         uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64  `gorm:"default:0;column:uuid;comment:'自助餐加钟价格ID'"`
	Name       string  `gorm:"default:'';column:name;comment:'自助餐加钟价格名称'"`
	DelayTime  uint    `gorm:"default:0;column:delay_time;comment:'加钟时间(分钟)'"`
	Price      float64 `gorm:"default:0;column:price;comment:'价格'"`
	Status     uint    `gorm:"default:0;column:status;comment:'状态 0-禁用 1-启用'"`
	CreateTime int64   `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime int64   `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime int64   `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

// SaleOrderBuffetDelayProduct 销售订单加钟价格商品表 `ttpos_sale_order_buffet_delay_product`
type SaleOrderBuffetDelayProduct struct {
	ID              uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid            uint64 `gorm:"default:0;column:uuid;comment:'自助餐加钟价格ID'"`
	SaleOrderUuid   uint64 `gorm:"default:0;column:sale_order_uuid;comment:'销售订单ID'"`
	BuffetDelayUuid uint64 `gorm:"default:0;column:buffet_delay_uuid;comment:'自助餐加钟价格ID'"`
	Num             uint   `gorm:"default:0;column:num;comment:'数量'"`
	CreateTime      int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime      int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime      int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}
