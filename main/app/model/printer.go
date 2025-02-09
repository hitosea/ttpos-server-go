package model

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
