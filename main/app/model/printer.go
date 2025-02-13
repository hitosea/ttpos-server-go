package model

// Printer 打印机信息表 ttpos_printer
type Printer struct {
	BaseModel
	Name            string `gorm:"default:'';column:name;comment:'打印机名称'"`
	PrinterTypeUuid uint   `gorm:"default:0;column:printer_type_uuid;comment:'打印机类型ID'"`
	ConfigJson      string `gorm:"default:'';column:config_json;comment:'打印机json配置'"`
	Copies          uint   `gorm:"default:0;column:copies;comment:'打印份数'"`
	Sort            uint   `gorm:"default:0;column:sort;comment:'排序'"`
}

// PrinterType 打印机类型信息表 ttpos_printer_type
type PrinterType struct {
	BaseModel
	Name       string `gorm:"default:'';column:name;comment:'打印机类型名称'"`
	Key        string `gorm:"default:'';column:key;comment:'打印机类型key'"`
	ConfigJson string `gorm:"default:'';column:config_json;comment:'打印机类型json配置,描述需要填写的字段'"`
}

// ProductPrinter 产品打印机信息表 ttpos_product_printer
type ProductPrinter struct {
	BaseModel
	Name               string `gorm:"default:'';column:name;comment:'名称.厨显上叫档口'"`
	Status             uint   `gorm:"default:0;column:status;comment:'状态,1-开启 1、0-关闭'"`
	PrintMode          uint   `gorm:"default:0;column:print_mode;comment:'打印模式,0-付款打印 1-下单（送厨）打印'"`
	PrintMethod        uint   `gorm:"default:0;column:print_method;comment:'打印方式,0-整单打印 1-按一菜一单打印'"`
	PrintProductSelect uint   `gorm:"default:0;column:print_product_select;comment:'打印商品选择,0-按商品分类 1-按打印标签'"`
	PrintModeScene     uint   `gorm:"default:0;column:print_mode_scene;comment:'打印模式场景,0-合并 1-分开'"`
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
