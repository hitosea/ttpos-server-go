package model

// Printer 打印机信息表 ttpos_printer
type Printer struct {
	BaseModel
	Name            string `gorm:"column:name;type:varchar(255);comment:打印机名称;NOT NULL" json:"name"`
	PrinterTypeUuid uint64 `gorm:"column:printer_type_uuid;type:bigint(20) unsigned;default:0;comment:打印机类型ID;NOT NULL" json:"printer_type_uuid"`
	ConfigJson      string `gorm:"column:config_json;type:text;comment:打印机json配置" json:"config_json"`
	Copies          uint   `gorm:"column:copies;type:int(11) unsigned;default:0;comment:打印份数;NOT NULL" json:"copies"`
	Sort            uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序;NOT NULL" json:"sort"`

	PrinterType *PrinterType `gorm:"foreignKey:PrinterTypeUuid;references:Uuid"` // 关联 printer_type
}

// PrinterType 打印机类型信息表 ttpos_printer_type
type PrinterType struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(255);comment:打印机类型名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	Key                   string `gorm:"column:key;type:varchar(255);comment:打印机类型key;NOT NULL" json:"key"`
	ConfigJson            string `gorm:"column:config_json;type:text;comment:打印机类型json配置,描述需要填写的字段" json:"config_json"`

	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// PrinterLog 打印日志表 ttpos_printer_log
type PrinterLog struct {
	BaseModel
	PrinterUuid     uint64 `gorm:"column:printer_uuid;type:bigint(20) unsigned;default:0;comment:打印机id;NOT NULL" json:"printer_uuid"`
	CashierDeviceId string `gorm:"column:cashier_device_id;type:varchar(255);comment:收银机绑定的id;NOT NULL" json:"cashier_device_id"`
	RelatedType     int    `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_type"`
	RelatedUuid     uint64 `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:销售账单、充值订单id;NOT NULL" json:"related_uuid"`
	Data            string `gorm:"column:data;type:varchar(255);comment:打印数据" json:"data"`
	Type            int    `gorm:"column:type;type:int(11);default:0;comment:类型:0系统默认队列,1云上服务下放;NOT NULL" json:"type"`
	DataType        int    `gorm:"column:data_type;type:tinyint(2);default:1;comment:数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单;NOT NULL" json:"data_type"`
	PrintMethod     int    `gorm:"column:print_method;type:tinyint(2);default:1;comment:打印方式 1文本打印, 2图片打印;NOT NULL" json:"print_method"`
	Num             int    `gorm:"column:num;type:int(11);default:0;comment:打印次数;NOT NULL" json:"num"`
	Status          int    `gorm:"column:status;type:tinyint(2);default:1;comment:状态(0结束,1进行中,2成功);NOT NULL" json:"status"`
	Reason          string `gorm:"column:reason;type:varchar(255);comment:原因" json:"reason"`
	PrinterTime     int64  `gorm:"column:printer_time;type:int(11);default:0;comment:打印时间;NOT NULL" json:"printer_time"`
	FirstExecution  int    `gorm:"column:first_execution;type:tinyint(2);default:0;comment:是否首次执行打印 1-是 0-否;NOT NULL" json:"first_execution"`

	Printer  *Printer  `gorm:"foreignKey:PrinterUuid;references:Uuid"` // 关联 printer
	SaleBill *SaleBill `gorm:"foreignKey:RelatedUuid;references:Uuid"` // 关联 sale_order
}

// PrinterReadLog 打印读取日志表 ttpos_printer_read_log
type PrinterReadLog struct {
	BaseModel
	LogUuid  int    `gorm:"column:log_uuid;type:int(11);default:0;comment:打印uuid" json:"log_uuid"`
	DeviceId string `gorm:"column:device_id;type:varchar(255);comment:设备id" json:"device_id"`
}

// ProductPrinter 产品打印机信息表 ttpos_product_printer
type ProductPrinter struct {
	BaseModel
	Name               string `gorm:"column:name;type:varchar(100);comment:名称.厨显上叫档口;NOT NULL" json:"name"`
	Status             int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,1-open开启 1、0-close关闭;NOT NULL" json:"status"`
	PrintMode          int    `gorm:"column:print_mode;type:tinyint(2);default:0;comment:打印模式,0-payment付款打印 1-kitchen送厨打印;NOT NULL" json:"print_mode"`
	PrintMethod        int    `gorm:"column:print_method;type:tinyint(2);default:0;comment:打印方式,0-order整单打印 1-item按一菜一单打印;NOT NULL" json:"print_method"`
	PrintProductSelect int    `gorm:"column:print_product_select;type:tinyint(2);default:0;comment:打印商品选择,0-category按商品分类 1-tag按打印标签;NOT NULL" json:"print_product_select"`
	PrintModeScene     int    `gorm:"column:print_mode_scene;type:tinyint(2);default:0;comment:打印模式场景,0-merge合并 1-separate分开;NOT NULL" json:"print_mode_scene"`
}

// ProductPrinterRegion 产品打印机区域信息表 ttpos_product_printer_region
type ProductPrinterRegion struct {
	BaseModel
	ProductPrinterUuid uint `gorm:"default:0;column:product_printer_uuid;comment:'产品打印机ID'"`
	DeskRegionUuid     uint `gorm:"default:0;column:desk_region_uuid;comment:'桌台区域ID'"`
}

// ProductPrinterItem 商品打印（档口）打印机信息表 ttpos_product_printer_item
type ProductPrinterItem struct {
	BaseModel
	ProductPrinterUuid uint `gorm:"default:0;column:product_printer_uuid;comment:'商品打印（档口）ID'"`
	PrinterUuid        uint `gorm:"default:0;column:printer_uuid;comment:'打印机ID'"`
}

// ProductPrinterProductItem 产品打印机产品明细信息表 ttpos_product_printer_product_item
type ProductPrinterProductItem struct {
	BaseModel
	ProductPrinterUuid uint `gorm:"default:0;column:product_printer_uuid;comment:'产品打印机ID'"`
	ProductPackageUuid uint `gorm:"default:0;column:product_package_uuid;comment:'产品包ID'"`
}
