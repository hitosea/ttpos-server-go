package model

// TakeoutOrderMaterial 外卖订单原料 `ttpos_takeout_order_material`
type TakeoutOrderMaterial struct {
	BaseModel
	// 基础标识字段
	TakeoutOrderUuid             uint64  `gorm:"column:takeout_order_uuid;type:bigint(20);default:0;comment:外卖订单ID" json:"takeout_order_uuid"`
	TakeoutOrderItemUuid         uint64  `gorm:"column:takeout_order_item_uuid;type:bigint(20);default:0;comment:外卖订单商品UUID(关联ttpos_takeout_order_item.uuid)" json:"takeout_order_item_uuid"`
	TakeoutOrderItemModifierUuid uint64  `gorm:"column:takeout_order_item_modifier_uuid;type:bigint(20);default:0;comment:外卖订单商品修饰符UUID(关联ttpos_takeout_order_item_modifier.uuid)" json:"takeout_order_item_modifier_uuid"`
	ProductBomUuid               uint64  `gorm:"column:product_bom_uuid;type:bigint(20);default:0;comment:BOM UUID(关联ttpos_product_bom.uuid)" json:"product_bom_uuid"`
	MaterialUuid                 uint64  `gorm:"column:material_uuid;type:bigint(20);default:0;comment:原料ID" json:"material_uuid"`
	MaterialName                 string  `gorm:"column:material_name;type:text;comment:原料名称(来自Material.Name)" json:"material_name"`
	ErpCode                      string  `gorm:"column:erp_code;type:varchar(50);default:'';comment:ERP编码(来自Material.Code)" json:"erp_code"`
	BaseUnitUom                  string  `gorm:"column:base_unit_uom;type:varchar(255);default:'';comment:基准单位ERPNext UOM(来自RelatedMaterial.BaseUnitUom)" json:"base_unit_uom"`
	WarehouseUuid                uint64  `gorm:"column:warehouse_uuid;type:bigint(20);default:0;comment:仓库ID" json:"warehouse_uuid"`
	Num                          float64 `gorm:"column:num;type:decimal(12,2);default:0;comment:出库数量" json:"num"`
	IsSummarized                 int     `gorm:"column:is_summarized;type:int(11);default:0;comment:是否已经统计,0-未统计 1-已统计" json:"is_summarized"`

	Material *Material `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
}

func (m *TakeoutOrderMaterial) TableName() string {
	return "ttpos_takeout_order_material"
}

func (m *TakeoutOrderMaterial) SetNil() {
	m.Material = nil
}

// Material 原料信息表 `ttpos_material`
type Material struct {
	BaseModel
	Name                  string   `gorm:"type:text;default:'';column:name;comment:'原料名称'"`
	Code                  string   `gorm:"default:'';column:code;comment:'原料编码'"`
	InitStock             float64  `gorm:"type:decimal(14,4);default:0.0000;column:init_stock;comment:'期初库存'"`
	MultiLanguageNameUuid uint64   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	CategoryUuid          uint64   `gorm:"default:0;column:category_uuid;comment:'类别ID'"`
	SupplierUuid          uint64   `gorm:"default:0;column:supplier_uuid;comment:'供应商ID'"`
	ImageUuid             uint64   `gorm:"default:0;column:image_uuid;comment:'图片ID'"`
	ImageName             string   `gorm:"default:'';column:image_name;comment:'图片名称'"`
	UnitUuid              uint64   `gorm:"default:0;column:unit_uuid;comment:'单位ID'"`
	PurchaseUnitUuid      uint64   `gorm:"default:0;column:purchase_unit_uuid;comment:'采购单位ID'"`
	CostUnitUuid          uint64   `gorm:"default:0;column:cost_unit_uuid;comment:'成本单位ID'"`
	Price                 float64  `gorm:"default:0;column:price;comment:'采购单价'"`
	StockNum              float64  `gorm:"default:0;column:stock_num;comment:'库存数量'"`
	SafetyStock           *float64 `gorm:"column:safety_stock;comment:'安全库存数量'"`
	ActualSaleNum         float64  `gorm:"default:0;column:actual_sale_num;comment:'实际销量'"`
	BarcodeValue          string   `gorm:"default:'';column:barcode_value;comment:'条形码值'"`
	InternalCode          string   `gorm:"default:'';column:internal_code;comment:'内部编码'"`
	Status                bool     `gorm:"default:false;column:status;comment:'状态,true上架 false下架'"`
	HeadquarterUuid       uint64   `gorm:"default:0;column:headquarter_uuid;comment:'总部ID'"`
	WarehouseUuid         uint64   `gorm:"default:0;column:warehouse_uuid;comment:'默认仓库Uuid，表示该原料的来自哪个仓库'"`
	OriginCountryCode     string   `gorm:"type:varchar(10);default:'';column:origin_country_code;comment:'原产地国家编码（ISO 3166-1 alpha-2）'"`
	AllowNegativeStock    int      `gorm:"column:allow_negative_stock;default:0;comment:'是否允许负库存：1-允许，0-不允许'"`

	Unit *MaterialUnit `gorm:"foreignKey:uuid;references:unit_uuid"` // 基准单位
}

type MaterialUnit struct {
	BaseModel
	Name           string  `gorm:"type:text;default:'';column:name;comment:'原料单位名称'"`
	UnitUuid       uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"` // 商品单位ID
	ConversionRate float64 `gorm:"type:decimal(12,4);default:1;column:conversion_rate;comment:'转换率'"`
	FromUnitUuid   uint64  `gorm:"default:0;column:from_unit_uuid;comment:'来源单位ID. 来源单位为克，则转换率为1000，该原料单位为千克'"`
	IsDefault      int     `gorm:"default:0;column:is_default;comment:'是否为基准单位, 0-否 1-是'"`
	MaterialUuid   uint64  `gorm:"default:0;column:material_uuid;comment:'原料ID'"`
}
