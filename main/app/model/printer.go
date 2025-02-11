package model

// Printer 打印机信息表 ttpos_printer
type Printer struct {
	ID              uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID            uint   `gorm:"default:0;unique;comment:'打印机ID'"`
	Name            string `gorm:"default:'';comment:'打印机名称'"`
	PrinterTypeUuid uint   `gorm:"default:0;comment:'打印机类型ID'"`
	ConfigJson      string `gorm:"default:'';comment:'打印机json配置'"`
	Copies          uint   `gorm:"default:0;comment:'打印份数'"`
	Sort            uint   `gorm:"default:0;comment:'排序'"`
	CreateTime      int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime      int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime      int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// PrinterType 打印机类型信息表 ttpos_printer_type
type PrinterType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint   `gorm:"default:0;unique;comment:'打印机类型ID'"`
	Name       string `gorm:"default:'';comment:'打印机类型名称'"`
	Key        string `gorm:"default:'';comment:'打印机类型key'"`
	ConfigJson string `gorm:"default:'';comment:'打印机类型json配置,描述需要填写的字段'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// ProductPrinter 产品打印机信息表 ttpos_product_printer
type ProductPrinter struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID               uint   `gorm:"default:0;comment:'产品打印机ID'"`
	Name               string `gorm:"default:'';comment:'名称.厨显上叫档口'"`
	Status             uint   `gorm:"default:0;comment:'状态,1-开启 1、0-关闭'"`
	PrintMode          uint   `gorm:"default:0;comment:'打印模式,0-付款打印 1-下单（送厨）打印'"`
	PrintMethod        uint   `gorm:"default:0;comment:'打印方式,0-整单打印 1-按一菜一单打印'"`
	PrintProductSelect uint   `gorm:"default:0;comment:'打印商品选择,0-按商品分类 1-按打印标签'"`
	PrintModeScene     uint   `gorm:"default:0;comment:'打印模式场景,0-合并 1-分开'"`
	CreateTime         int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime         int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime         int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// ProductPrinterRegion 产品打印机区域信息表 ttpos_product_printer_region
type ProductPrinterRegion struct {
	ID                 uint  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID               uint  `gorm:"default:0;comment:'产品打印机区域ID'"`
	ProductPrinterUuid uint  `gorm:"default:0;comment:'产品打印机ID'"`
	DeskRegionUuid     uint  `gorm:"default:0;comment:'桌台区域ID'"`
	CreateTime         int64 `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime         int64 `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime         int64 `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// ProductPrinterItem 商品打印（档口）打印机信息表 ttpos_product_printer_item
type ProductPrinterItem struct {
	ID                 uint  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID               uint  `gorm:"default:0;comment:'商品打印（档口）打印机ID'"`
	ProductPrinterUuid uint  `gorm:"default:0;comment:'商品打印（档口）ID'"`
	PrinterUuid        uint  `gorm:"default:0;comment:'打印机ID'"`
	CreateTime         int64 `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime         int64 `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime         int64 `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// ProductPrinterProductItem 产品打印机产品明细信息表 ttpos_product_printer_product_item
type ProductPrinterProductItem struct {
	ID                 uint  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID               uint  `gorm:"default:0;comment:'产品打印机产品明细ID'"`
	ProductPrinterUuid uint  `gorm:"default:0;comment:'产品打印机ID'"`
	ProductPackageUuid uint  `gorm:"default:0;comment:'产品包ID'"`
	CreateTime         int64 `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime         int64 `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime         int64 `gorm:"default:0;comment:'删除时间（时间戳）'"`
}
