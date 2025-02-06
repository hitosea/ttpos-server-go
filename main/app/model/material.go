package model

type Material struct {
	Id                    uint    `gorm:"primaryKey;AUTO_INCREMENT;comment:'原料唯一标识符'"`
	Uuid                  uint    `gorm:"default:0;comment:'UUID'"`
	Name                  string  `gorm:"default:'';comment:'原料名称'"`
	MultiLanguageNameUuid uint    `gorm:"default:0;comment:'多语言名称ID'"`
	CategoryKey           string  `gorm:"default:'';comment:'类别关键字'"`
	CategoryUuid          uint    `gorm:"default:0;comment:'类别ID'"`
	SupplierUuid          uint    `gorm:"default:0;comment:'供应商ID'"`
	ImageUrl              string  `gorm:"default:'';comment:'图片URL'"`
	ImageName             string  `gorm:"default:'';comment:'图片名称'"`
	UnitUuid              uint    `gorm:"default:0;comment:'单位ID'"`
	Price                 float64 `gorm:"default:0;comment:'采购单价'"`
	Num                   uint    `gorm:"default:0;comment:'库存数量'"`
	BarcodeValue          string  `gorm:"default:'';comment:'条形码值'"`
	Status                string  `gorm:"default:'';comment:'状态,up上架、down下架'"`
	CreateTime            int     `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime            int     `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime            int     `gorm:"default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}
