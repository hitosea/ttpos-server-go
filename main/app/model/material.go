package model

type Material struct {
	Id                  uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'原料唯一标识符'"`
	Name                string  `gorm:"column:name;not null;default:'';comment:'原料名称'"`
	MultiLanguageNameId uint    `gorm:"column:multi_language_name_id;not null;default:0;comment:'多语言名称ID'"`
	CategoryKey         string  `gorm:"column:category_key;not null;default:'';comment:'类别关键字'"`
	CategoryId          uint    `gorm:"column:category_id;not null;default:0;comment:'类别ID'"`
	SupplierId          uint    `gorm:"column:supplier_id;not null;default:0;comment:'供应商ID'"`
	ImageUrl            string  `gorm:"column:image_url;not null;default:'';comment:'图片URL'"`
	ImageName           string  `gorm:"column:image_name;not null;default:'';comment:'图片名称'"`
	UnitId              uint    `gorm:"column:unit_id;not null;default:0;comment:'单位ID'"`
	Price               float64 `gorm:"column:price;not null;default:0;comment:'采购单价'"`
	Num                 uint    `gorm:"column:num;not null;default:0;comment:'库存数量'"`
	BarcodeValue        string  `gorm:"column:barcode_value;not null;default:'';comment:'条形码值'"`
	Status              string  `gorm:"column:status;not null;default:'';comment:'状态,up上架、down下架'"`
	CreateTime          int     `gorm:"column:create_time;not null;default:0;comment:'创建时间（时间戳）'"`
	UpdateTime          int     `gorm:"column:update_time;not null;default:0;comment:'更新时间（时间戳）'"`
	DeleteTime          int     `gorm:"column:delete_time;not null;default:0;comment:'删除时间（时间戳）'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_id;references:id"` // 多语言名称
}
