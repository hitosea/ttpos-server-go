package model

// ProductFlavor 商品规格表,定义商品的规格信息 ttpos_product_flavor
type ProductFlavor struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// ProductSauce 商品小料表,定义商品小料的相关信息 ttpos_product_sauce
type ProductSauce struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// ProductUnit 商品单位表,定义商品的单位信息 ttpos_product_unit
type ProductUnit struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'单位名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// PrinterTag 打印标签表,定义打印标签的相关信息 ttpos_printer_tag
type PrinterTag struct {
	BaseModel
	Name string `gorm:"default:'';column:name;comment:'名称'"`
}

// ProductAttributeGroup 产品属性组表,定义产品的属性分组信息 ttpos_product_attribute_group
type ProductAttributeGroup struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// ProductAttribute 商品属性表,定义商品的属性信息 ttpos_product_attribute
type ProductAttribute struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`
	AttributeGroupUuid    uint64 `gorm:"default:0;column:attribute_group_uuid;comment:'属性组UUID'"`

	MultiLanguageName MultiLanguageName     `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	AttributeGroup    ProductAttributeGroup `gorm:"foreignKey:attribute_group_uuid;references:uuid"`     // 属性组
}

// ProductPackageAttributeGroup 产品包属性组表,定义产品包的属性分组信息 ttpos_product_package_attribute_group
type ProductPackageAttributeGroup struct {
	BaseModel
	IsMust                    uint   `gorm:"default:0;column:is_must;comment:'是否必选, 0-否 1-是'"`
	MaxSelection              uint   `gorm:"default:0;column:max_selection;comment:'最大选择数量'"`
	ProductPackageUuid        uint64 `gorm:"default:0;column:product_package_uuid;comment:'产品包UUID'"`
	ProductAttributeGroupUuid uint64 `gorm:"default:0;column:product_attribute_group_uuid;comment:'商品属性组UUID'"`

	ProductAttributeGroup    ProductAttributeGroup     `gorm:"foreignKey:product_attribute_group_uuid;references:uuid"`         // 商品属性组
	ProductPackageAttributes []ProductPackageAttribute `gorm:"foreignKey:product_package_attribute_group_uuid;references:uuid"` // 产品包属性
}

// ProductPackageAttribute 产品包属性表,定义产品包的属性信息 ttpos_product_package_attribute
type ProductPackageAttribute struct {
	BaseModel
	ProductPackageAttributeGroupUuid uint64 `gorm:"default:0;column:product_package_attribute_group_uuid;comment:'产品包属性组UUID'"`
	AttributeUuid                    uint64 `gorm:"default:0;column:attribute_uuid;comment:'产品属性UUID'"`
	IsDefaultSelected                uint   `gorm:"default:0;column:is_default_selected;comment:'是否默认选中, 0-否 1-是'"`

	Attribute ProductAttribute `gorm:"foreignKey:attribute_uuid;references:uuid"` // 产品属性
}

// ProductPackage 产品包表,定义产品包的相关信息 `ttpos_product_package`
type ProductPackage struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'产品包名称'"`
	MultiLanguageNameUuid uint   `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`
	ImageName             string `gorm:"default:'';column:image_name;comment:'图片名称'"`
	ImageFileUuid         uint64 `gorm:"default:0;column:image_file_uuid;comment:'图片UUID'"`
	DeductStockType       uint   `gorm:"default:0;column:deduct_stock_type;comment:'库存计算方法, 0-下单减库存 1-付款减库存'"`
	UnitUuid              uint64 `gorm:"default:0;column:unit_uuid;comment:'单位UUID'"`
	DineTaxUuid           uint64 `gorm:"default:0;column:dine_tax_uuid;comment:'堂食税UUID'"`
	CategoryUuid          uint64 `gorm:"default:0;column:category_uuid;comment:'类别UUID'"`
	TakeoutTaxUuid        uint64 `gorm:"default:0;column:takeout_tax_uuid;comment:'外卖税UUID'"`
	SpecialCategoryUuid   uint64 `gorm:"default:0;column:special_category_uuid;comment:'特殊类别UUID'"`
	PrinterTagUuid        uint64 `gorm:"default:0;column:printer_tag_uuid;comment:'打印机标签UUID'"`
	SupplierUuid          uint64 `gorm:"default:0;column:supplier_uuid;comment:'供应商UUID'"`
	Status                uint   `gorm:"default:0;column:status;comment:'状态, 0-上架 1-下架'"`
	IsShowCashier         uint   `gorm:"default:0;column:is_show_cashier;comment:'是否在收银设备显示, 0-否 1-是'"`
	IsShowTablet          uint   `gorm:"default:0;column:is_show_tablet;comment:'是否在平板设备显示, 0-否 1-是'"`
	IsShowKitchen         uint   `gorm:"default:0;column:is_show_kitchen;comment:'是否在厨房设备显示, 0-否 1-是'"`
	IsShowAssistant       uint   `gorm:"default:0;column:is_show_assistant;comment:'是否在助手设备显示, 0-否 1-是'"`
	IsShowH5              uint   `gorm:"default:0;column:is_show_h5;comment:'是否在H5设备显示, 0-否 1-是'"`
	Sort                  uint   `gorm:"default:0;column:sort;comment:'排序'"`
	LimitNum              uint   `gorm:"default:0;column:limit_num;comment:'限购数量'"`
	Describe              string `gorm:"default:'';column:describe;comment:'卖点描述'"`
	SauceRequired         uint8  `gorm:"default:0;column:sauce_required;comment:'是否必选小料, 0-否 1-是'"`
	SauceMaxSelection     uint   `gorm:"default:0;column:sauce_max_selection;comment:'小料最大选择数量'"`
	OpenDiscount          uint   `gorm:"default:0;column:open_discount;comment:'是否开启会员折扣, 0-否 1-是'"`

	MultiLanguageName             MultiLanguageName              `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	ProductUnit                   ProductUnit                    `gorm:"foreignKey:unit_uuid;references:uuid"`                // 单位
	ProductBoms                   []ProductBom                   `gorm:"foreignKey:product_package_uuid;references:uuid"`     // BOM
	ProductPackageAttributeGroups []ProductPackageAttributeGroup `gorm:"foreignKey:product_package_uuid;references:uuid"`     // 产品包属性组
}

func (model *ProductPackage) IsDown() bool {
	if model.Status == 0 {
		return true
	}
	return false
}

func (model *ProductPackage) IsDelete() bool {
	if model.DeleteTime != 0 {
		return true
	}
	return false
}

// ProductBom 产品BOM表,定义产品BOM的相关信息 ttpos_product_bom
type ProductBom struct {
	BaseModel
	Price              float64 `gorm:"column:price;not null;default:0;comment:'价格'"`
	ProductFlavorUuid  uint64  `gorm:"column:product_flavor_uuid;not null;default:0;comment:'商品规格UUID（仅商品使用）'"`
	ProductSauceUuid   uint64  `gorm:"column:product_sauce_uuid;not null;default:0;comment:'商品小料UUID（仅小料使用）'"`
	ProductPackageUuid uint64  `gorm:"column:product_package_uuid;not null;default:0;comment:'产品包UUID'"`
	IsDefaultSelect    uint    `gorm:"column:is_default_select;not null;default:0;comment:'是否默认选择, 0-否 1-是'"`
	IsSoldOut          uint    `gorm:"column:is_sold_out;not null;default:0;comment:'是否售罄：0-否 1-是'"`

	ProductPackage ProductPackage `gorm:"foreignKey:ProductPackageUuid;references:uuid"`  // 商品
	ProductFlavor  ProductFlavor  `gorm:"foreignKey:product_flavor_uuid;references:uuid"` // 商品规格
	ProductSauce   ProductSauce   `gorm:"foreignKey:product_sauce_uuid;references:uuid"`  // 商品小料
}

// IsFlavorProduct 判断是否为商品规格
func (model *ProductBom) IsFlavorProduct() bool {
	return model.ProductFlavorUuid != 0
}

// IsSauceProduct 判断是否为商品小料
func (model *ProductBom) IsSauce() bool {
	return model.ProductSauceUuid != 0
}
