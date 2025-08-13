package model

// Material 原料信息表 ttpos_material
type Material struct {
	BaseModel
	Name                  string  `gorm:"default:'';column:name;comment:'原料名称'"`
	Code                  string  `gorm:"default:'';column:code;comment:'原料编码'"`
	Valuation             float64 `gorm:"type:decimal(12,2);default:0;column:valuation;comment:'估值率'"`
	InitStock             float64 `gorm:"type:decimal(14,4);default:0.0000;column:init_stock;comment:'期初库存'"`
	MultiLanguageNameUuid uint64  `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CategoryUuid          uint64  `gorm:"default:0;column:category_uuid;comment:'类别ID'"`
	SupplierUuid          uint64  `gorm:"default:0;column:supplier_uuid;comment:'供应商ID'"`
	ImageUuid             uint64  `gorm:"default:0;column:image_uuid;comment:'图片ID'"`
	ImageName             string  `gorm:"default:'';column:image_name;comment:'图片名称'"`
	UnitUuid              uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"`
	PurchaseUnitUuid      uint64  `gorm:"default:0;column:purchase_unit_uuid;comment:'采购单位ID'"`
	CostUnitUuid          uint64  `gorm:"default:0;column:cost_unit_uuid;comment:'成本单位ID'"`
	Price                 float64 `gorm:"default:0;column:price;comment:'采购单价'"`
	StockNum              float64 `gorm:"default:0;column:stock_num;comment:'库存数量'"`
	ActualSaleNum         float64 `gorm:"default:0;column:actual_sale_num;comment:'实际销量'"`
	BarcodeValue          string  `gorm:"default:'';column:barcode_value;comment:'条形码值'"`
	Status                bool    `gorm:"default:false;column:status;comment:'状态,true上架 false下架'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	Unit              MaterialUnit      `gorm:"foreignKey:unit_uuid;references:uuid"`                // 单位
	PurchaseUnit      MaterialUnit      `gorm:"foreignKey:purchase_unit_uuid;references:uuid"`       // 采购单位
	CostUnit          MaterialUnit      `gorm:"foreignKey:cost_unit_uuid;references:uuid"`           // 成本单位
}

func (model *Material) SetNil() {
	model.MultiLanguageName = MultiLanguageName{}
	model.Unit = MaterialUnit{}
	model.PurchaseUnit = MaterialUnit{}
	model.CostUnit = MaterialUnit{}
}

// MaterialUnit 原料单位表 ttpos_material_unit
type MaterialUnit struct {
	BaseModel
	Name           string  `gorm:"default:'';column:name;comment:'原料单位名称'"`
	UnitUuid       uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"`
	ConversionRate float64 `gorm:"type:decimal(12,4);default:1;column:conversion_rate;comment:'转换率'"`
	FromUnitUuid   uint64  `gorm:"default:0;column:from_unit_uuid;comment:'来源单位ID. 来源单位为克，则转换率为1000，该原料单位为千克'"`
	IsDefault      int     `gorm:"default:0;column:is_default;comment:'是否为基准单位, 0-否 1-是'"`
}
