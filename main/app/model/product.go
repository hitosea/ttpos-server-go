package model

import (
	"slices"
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

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

	MultiLanguageName MultiLanguageName  `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	SauceMaterials    []*RelatedMaterial `gorm:"foreignKey:product_sauce_uuid;references:uuid"`       // 小料的组成材料
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

func (model *ProductPackageAttributeGroup) IsMustBool() bool {
	return model.IsMust == constant.ProductAttributeGroupRequiredOn
}

// ProductPackageAttribute 产品包属性表,定义产品包的属性信息 ttpos_product_package_attribute
type ProductPackageAttribute struct {
	BaseModel
	ProductPackageAttributeGroupUuid uint64 `gorm:"default:0;column:product_package_attribute_group_uuid;comment:'产品包属性组UUID'"`
	AttributeUuid                    uint64 `gorm:"default:0;column:attribute_uuid;comment:'产品属性UUID'"`
	IsDefaultSelected                uint   `gorm:"default:0;column:is_default_selected;comment:'是否默认选中, 0-否 1-是'"`

	Attribute ProductAttribute `gorm:"foreignKey:attribute_uuid;references:uuid"` // 产品属性
}

func (model *ProductPackageAttribute) IsDefaultSelectedBool() bool {
	return model.IsDefaultSelected == constant.ProductAttributeDefaultSelectionOn
}

// ProductPackage 产品包表,定义产品包的相关信息 `ttpos_product_package`
type ProductPackage struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'产品包名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`
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

	SauceRequired     uint8 `gorm:"default:0;column:sauce_required;comment:'是否必选小料, 0-否 1-是'"`
	SauceMaxSelection uint  `gorm:"default:0;column:sauce_max_selection;comment:'小料最大选择数量'"`
	OpenDiscount      uint  `gorm:"default:0;column:open_discount;comment:'是否开启会员折扣, 0-否 1-是'"`

	MultiLanguageName             MultiLanguageName              `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	ProductUnit                   ProductUnit                    `gorm:"foreignKey:unit_uuid;references:uuid"`                // 单位
	ProductBoms                   []ProductBom                   `gorm:"foreignKey:product_package_uuid;references:uuid"`     // BOM
	ProductPackageAttributeGroups []ProductPackageAttributeGroup `gorm:"foreignKey:product_package_uuid;references:uuid"`     // 产品包属性组
	DineTax                       Tax                            `gorm:"foreignKey:dine_tax_uuid;references:uuid"`            // 堂食税
	TakeoutTax                    Tax                            `gorm:"foreignKey:takeout_tax_uuid;references:uuid"`         // 外卖税
	ProductCategory               ProductCategory                `gorm:"foreignKey:category_uuid;references:uuid"`            // 类别
	ImageFile                     File                           `gorm:"foreignKey:image_file_uuid;references:uuid"`          // 图片
}

// IsCookingDeductStock 判断商品包是否是下单减库存
func (model *ProductPackage) IsCookingDeductStock() bool {
	return model.DeductStockType == constant.ProductPackageDeductStockTypeCooking
}

// IsPayDeductStock 判断商品包是否是结账减库存
func (model *ProductPackage) IsPayDeductStock() bool {
	return model.DeductStockType == constant.ProductPackageDeductStockTypePay
}

func (model *ProductPackage) GetSauceRequired() bool {
	return model.SauceRequired == constant.ProductPackageSauceRequiredOn
}

func (model *ProductPackage) TaxRate(dineType uint) float64 {
	if dineType == constant.SaleBillDiningMethodDineIn {
		return model.DineTax.TaxRate
	}
	return model.TakeoutTax.TaxRate
}

// IsUp 判断商品是否是上架状态。排除下架、删除状态
func (model *ProductPackage) IsUp() bool {
	if model.Status == constant.ProductStatusOffSale || model.DeleteTime != constant.NotDeleted {
		return false
	}
	if model.Status == constant.ProductStatusOnSale {
		return true
	}
	return false
}

// 是否是无选择的商品。即可以直接加购不需要弹出弹框再选择的商品
// 这类商品只有一个商品规格、没有商品属性、没有商品加料
func (model *ProductPackage) IsNoSelectProduct() bool {
	// 只有一个商品规格、没有加料
	if len(model.ProductBoms) > 1 {
		return false
	}
	// 没有商品属性
	if len(model.ProductPackageAttributeGroups) > 0 {
		return false
	}
	return true
}

func (model *ProductPackage) GetMinPrice() float64 {
	minPrice := float64(0)
	for _, productBom := range model.ProductBoms {
		if !productBom.IsFlavor() {
			continue
		}
		if minPrice == 0 {
			minPrice = productBom.Price
			continue
		}
		if productBom.Price < minPrice {
			minPrice = productBom.Price
		}
	}
	return minPrice
}

// ProductBom 产品BOM表,定义产品的规格、小料相关信息 `ttpos_product_bom`
type ProductBom struct {
	BaseModel
	PurchasePrice   float64 `gorm:"column:purchase_price;type:decimal(12,2);default:0;comment:采购单价;NOT NULL" json:"purchase_price"`
	Price           float64 `gorm:"column:price;type:decimal(12,2);default:0;comment:价格;NOT NULL" json:"price"`
	Name            string  `gorm:"column:name;type:text;comment:商品名称或小料名称(不用于业务显示)" json:"name"`
	StockNum        float64 `gorm:"column:stock_num;type:decimal(12,4);default:0.0000;comment:库存数量;NOT NULL" json:"stock_num"`
	BarcodeValue    string  `gorm:"column:barcode_value;type:varchar(255);comment:条形码值;NOT NULL" json:"barcode_value"`
	IsDefaultSelect int     `gorm:"column:is_default_select;type:tinyint(1);default:0;comment:是否默认选择, 0-否 1-是;NOT NULL" json:"is_default_select"`
	Status          int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-下架 1-上架. 同步商品包的状态;NOT NULL" json:"status"`
	IsSoldOut       int     `gorm:"column:is_sold_out;type:tinyint(1);default:0;comment:是否沽清, 0-否 1-是;NOT NULL" json:"is_sold_out"`

	// 关联ID
	ProductFlavorUuid  uint64 `gorm:"column:product_flavor_uuid;type:bigint(20) unsigned;default:0;comment:商品规格ID(仅商品使用);NOT NULL" json:"product_flavor_uuid"`
	ProductSauceUuid   uint64 `gorm:"column:product_sauce_uuid;type:bigint(20) unsigned;default:0;comment:商品小料ID(仅小料使用);NOT NULL" json:"product_sauce_uuid"`
	ProductPackageUuid uint64 `gorm:"column:product_package_uuid;type:bigint(20) unsigned;default:0;comment:商品包ID;NOT NULL" json:"product_package_uuid"`

	ProductPackage  ProductPackage     `gorm:"foreignKey:ProductPackageUuid;references:uuid"`  // 商品
	ProductFlavor   ProductFlavor      `gorm:"foreignKey:product_flavor_uuid;references:uuid"` // 商品规格
	ProductSauce    ProductSauce       `gorm:"foreignKey:product_sauce_uuid;references:uuid"`  // 商品小料
	FlavorMaterials []*RelatedMaterial `gorm:"foreignKey:related_uuid;references:uuid"`        // 规格商品的组成材料
}

// RelatedMaterial 关联材料表,定义关联材料的相关信息 ttpos_related_material
type RelatedMaterial struct {
	BaseModel
	MaterialUUID     uint64  `gorm:"column:material_uuid;type:bigint(20) unsigned;default:0;comment:'原料ID'"`
	RelatedUuid      uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:'物料清单BOM的ID'"`
	ProductBomUuid   uint64  `gorm:"column:product_bom_uuid;type:bigint(20) unsigned;default:0;comment:'物料清单BOM的ID'"`
	ProductSauceUuid uint64  `gorm:"column:product_sauce_uuid;type:bigint(20) unsigned;default:0;comment:'商品小料ID'"`
	Num              float64 `gorm:"column:num;type:decimal(12,4);default:0;comment:'材料用量,可小数'"`
}

// GetDecreaseNum 获取减少的库存数量. 减少的库存数量 = 材料用量 * 商品数量
func (model *RelatedMaterial) GetDecreaseNum(productNum uint) float64 {
	return decimal.NewFromFloat(model.Num).Mul(decimal.NewFromInt(int64(productNum))).InexactFloat64()
}

// IsStockShortage 判断库存是否不足
func (model *ProductBom) IsStockShortage(productNum uint) bool {
	return model.StockNum < float64(productNum)
}

// IsPriceChanged 判断商品价格是否变动
func (model *ProductBom) IsPriceChanged(price float64) bool {
	return model.Price != price
}

// IsSoldOutStatus 判断是否标记沽清、或售罄无库存
func (model *ProductBom) IsSoldOutStatus() bool {
	return model.IsSoldOut == constant.ProductStatusSaleOut || model.StockNum <= 0
}

func (model *ProductBom) IsDefaultSelectBool() bool {
	return model.IsDefaultSelect == constant.ProductPackageSauceDefaultSelectionOn
}

// IsNotSoldOutStatus 判断bom是否还可以销售
func (model *ProductBom) IsNotSoldOutStatus() bool {
	return !model.IsSoldOutStatus()
}

func (model *ProductBom) IsUp() bool {
	if model.IsSoldOut == constant.ProductStatusSaleOut ||
		model.Status == constant.ProductStatusOffSale ||
		model.DeleteTime != constant.NotDeleted {
		return false
	}
	if model.Status == constant.ProductStatusOnSale {
		return true
	}
	return false
}

func (model *ProductBom) IsDown() bool {
	return !model.IsUp()
}

// IsProductPackageDown 判断商品包是否下架
func (model *ProductBom) IsProductPackageDown() bool {
	return !model.ProductPackage.IsUp()
}

// IsDown 判断商品包是否下架
func (model *ProductPackage) IsDown() bool {
	return !model.IsUp()
}

// IsDelete 判断商品包是否删除
func (model *ProductPackage) IsDelete() bool {
	return model.DeleteTime != 0
}

// GenerateOriginalAmount 生成商品原始价格
func (model *ProductPackage) GenerateOriginalAmount(sauceUuids []uint64) (float64, float64, float64, float64) {
	var (
		flavorPrice  float64
		saucePrice   float64
		productPrice float64
		salePrice    float64
	)
	for _, bom := range model.ProductBoms {
		if bom.IsFlavor() {
			flavorPrice = bom.Price
		}
		if bom.IsSauce() && slices.Contains(sauceUuids, bom.ProductSauceUuid) {
			saucePrice = decimal.NewFromFloat(bom.Price).Add(decimal.NewFromFloat(saucePrice)).InexactFloat64()
		}
	}
	productPrice = decimal.NewFromFloat(flavorPrice).Add(decimal.NewFromFloat(saucePrice)).InexactFloat64()
	return flavorPrice, saucePrice, productPrice, salePrice
}

// GetFlavor 获取商品规格
func (model *ProductPackage) GetFlavor() ProductFlavor {
	for _, bom := range model.ProductBoms {
		if bom.IsFlavor() {
			return bom.ProductFlavor
		}
	}

	return ProductFlavor{}
}

// GetSauces 获取商品小料
func (model *ProductPackage) GetSauces() []ProductSauce {
	sauces := make([]ProductSauce, 0)
	for _, bom := range model.ProductBoms {
		if bom.IsSauce() {
			sauces = append(sauces, bom.ProductSauce)
		}
	}

	return sauces
}

// IsFlavor 判断是否为商品规格
func (model *ProductBom) IsFlavor() bool {
	return model.ProductFlavorUuid != 0
}

// IsSauceProduct 判断是否为商品小料
func (model *ProductBom) IsSauce() bool {
	return model.ProductSauceUuid != 0
}
