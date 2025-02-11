package model

// Material 原料信息表 ttpos_material
type Material struct {
	ID                    uint    `gorm:"column:id;primaryKey;AUTO_INCREMENT;comment:'原料唯一标识符'"`
	Uuid                  uint64  `gorm:"default:0;column:uuid;comment:'UUID'"`
	Name                  string  `gorm:"default:'';column:name;comment:'原料名称'"`
	MultiLanguageNameUuid uint    `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CategoryUuid          uint64  `gorm:"default:0;column:category_uuid;comment:'类别ID'"`
	SupplierUuid          uint    `gorm:"default:0;column:supplier_uuid;comment:'供应商ID'"`
	ImageUuid             uint64  `gorm:"default:0;column:image_uuid;comment:'图片ID'"`
	ImageName             string  `gorm:"default:'';column:image_name;comment:'图片名称'"`
	UnitUuid              uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"`
	Price                 float64 `gorm:"default:0;column:price;comment:'采购单价'"`
	Num                   float64 `gorm:"default:0;column:num;comment:'库存数量'"`
	BarcodeValue          string  `gorm:"default:'';column:barcode_value;comment:'条形码值'"`
	Status                bool    `gorm:"default:false;column:status;comment:'状态,true上架 false下架'"`
	CreateTime            int     `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime            int     `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime            int     `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}
