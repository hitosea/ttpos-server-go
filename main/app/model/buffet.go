package model

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

	MultiLanguageName       MultiLanguageName       `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	BuffetCustomerTypePrice BuffetCustomerTypePrice `gorm:"foreignKey:buffet_package_uuid;references:uuid"`
}

type BuffetCustomerType struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"default:0;column:uuid;comment:'自助餐客户类型ID'"`
	Name       string `gorm:"default:'';column:name;comment:'自助餐客户类型名称'"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

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
