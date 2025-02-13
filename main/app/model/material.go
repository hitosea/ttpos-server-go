package model

// Material 原料信息表 ttpos_material
type Material struct {
	BaseModel
	Name                  string  `gorm:"default:'';column:name;comment:'原料名称'"`
	MultiLanguageNameUuid uint    `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CategoryUuid          uint64  `gorm:"default:0;column:category_uuid;comment:'类别ID'"`
	SupplierUuid          uint    `gorm:"default:0;column:supplier_uuid;comment:'供应商ID'"`
	ImageUuid             uint64  `gorm:"default:0;column:image_uuid;comment:'图片ID'"`
	ImageName             string  `gorm:"default:'';column:image_name;comment:'图片名称'"`
	UnitUuid              uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"`
	Price                 float64 `gorm:"default:0;column:price;comment:'采购单价'"`
	StockNum              float64 `gorm:"default:0;column:stock_num;comment:'库存数量'"`
	BarcodeValue          string  `gorm:"default:'';column:barcode_value;comment:'条形码值'"`
	Status                bool    `gorm:"default:false;column:status;comment:'状态,true上架 false下架'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}
