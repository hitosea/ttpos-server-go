package model

import (
	"ttpos-server-go/app/errors"
)

// Material 原料信息表 `ttpos_material`
type Material struct {
	BaseModel
	Name                  string  `gorm:"type:text;default:'';column:name;comment:'原料名称'"`
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
	InternalCode          string  `gorm:"default:'';column:internal_code;comment:'内部编码'"`
	Status                bool    `gorm:"default:false;column:status;comment:'状态,true上架 false下架'"`
	HeadquarterUuid       uint64  `gorm:"default:0;column:headquarter_uuid;comment:'总部ID'"`
	WarehouseUuid         uint64  `gorm:"default:0;column:warehouse_uuid;comment:'默认仓库Uuid，表示该原料的来自哪个仓库'"`

	MultiLanguageName   MultiLanguageName  `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
	Unit                *MaterialUnit      `gorm:"foreignKey:uuid;references:unit_uuid"`                // 基准单位
	PurchaseUnit        *MaterialUnit      `gorm:"foreignKey:purchase_unit_uuid;references:uuid"`       // 采购单位
	CostUnit            *MaterialUnit      `gorm:"foreignKey:cost_unit_uuid;references:uuid"`           // 成本单位
	Category            MaterialCategory   `gorm:"foreignKey:category_uuid;references:uuid"`            // 分类
	NotBaseUnitList     []*MaterialUnit    `gorm:"foreignKey:material_uuid;references:uuid"`            // 非基准单位列表
	ImageFile           *File              `gorm:"foreignKey:image_uuid;references:uuid"`               // 图片
	RelatedMaterialList []*RelatedMaterial `gorm:"foreignKey:material_uuid;references:uuid"`            // 规格/加料关联材料
	WarehouseItems      []*WarehouseItem   `gorm:"foreignKey:material_uuid;references:uuid"`            // 仓库物品库存
}

func (model *Material) SetNil() {
	model.MultiLanguageName = MultiLanguageName{}
	model.Unit = nil
	model.PurchaseUnit = nil
	model.CostUnit = nil
	model.Category = MaterialCategory{}
	model.NotBaseUnitList = nil
	model.ImageFile = nil
}

// GetStockNum 获取库存数量。获取物品在默认仓库中的库存数量
func (model *Material) GetStockNum() float64 {
	for _, warehouseItem := range model.WarehouseItems {
		if warehouseItem.WarehouseUuid == model.WarehouseUuid {
			return warehouseItem.Stock
		}
	}
	return 0
}

// GetValuation
func (model *Material) GetValuation() float64 {
	if model.Valuation != 0 {
		return model.Valuation
	}
	return 1
}

// 通过uom名获取单位uuid
func (model *Material) GetUnitUuidByUom(uom string) (uint64, error) {
	for _, unit := range model.NotBaseUnitList {
		if unit.Unit != nil && unit.Unit.ErpnextUom == uom {
			return unit.Uuid, nil
		}
	}
	return 0, errors.New("单位不存在")
}

// 通过uom名获取单位
func (model *Material) GetUnitByUom(uom string) (*MaterialUnit, error) {
	for _, unit := range model.NotBaseUnitList {
		if unit.Unit.ErpnextUom == uom {
			return unit, nil
		}
	}
	return nil, errors.New("单位不存在")
}

// 判断该单位是不是该物品的基准单位或非基准单位
func (model *Material) IsUnit(unitUuid uint64) bool {
	for _, unit := range model.NotBaseUnitList {
		if unit.Uuid == unitUuid {
			return true
		}
	}
	return false
}

// 判断该物品是不是总部物品
func (model *Material) IsHeadquarter() bool {
	return model.HeadquarterUuid != 0
}

// 获取物品的某个单位信息
func (model *Material) GetUnit(unitUuid uint64) *MaterialUnit {
	for _, unit := range model.NotBaseUnitList {
		if unit.Uuid == unitUuid {
			return unit
		}
	}
	return nil
}

// 获取物品的基准单位信息
func (model *Material) GetBaseUnit() *MaterialUnit {
	for _, unit := range model.NotBaseUnitList {
		if unit.IsDefault == 1 {
			return unit
		}
	}
	return nil
}

func (model *Material) GetImage(baseUrl string) string {
	if model.ImageFile != nil {
		return model.ImageFile.GetUrl(baseUrl)
	}
	return ""
}

// MaterialUnit 原料单位表 ttpos_material_unit
type MaterialUnit struct {
	BaseModel
	Name           string  `gorm:"type:text;default:'';column:name;comment:'原料单位名称'"`
	UnitUuid       uint64  `gorm:"default:0;column:unit_uuid;comment:'单位ID'"` // 商品单位ID
	ConversionRate float64 `gorm:"type:decimal(12,4);default:1;column:conversion_rate;comment:'转换率'"`
	FromUnitUuid   uint64  `gorm:"default:0;column:from_unit_uuid;comment:'来源单位ID. 来源单位为克，则转换率为1000，该原料单位为千克'"`
	IsDefault      int     `gorm:"default:0;column:is_default;comment:'是否为基准单位, 0-否 1-是'"`
	MaterialUuid   uint64  `gorm:"default:0;column:material_uuid;comment:'原料ID'"`

	Unit *ProductUnit `gorm:"foreignKey:unit_uuid;references:uuid"`
}

func (model *MaterialUnit) SetNil() {
	model.Unit = nil
}

// MaterialCategory 原料分类表 ttpos_material_category
type MaterialCategory struct {
	BaseModel
	Name                  string `gorm:"type:text;default:'';column:name;comment:'原料分类名称'"`
	Code                  string `gorm:"default:'';column:code;comment:'原料分类编码'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	Sort                  int    `gorm:"default:0;column:sort;comment:'排序'"`
	HeadquarterUuid       uint64 `gorm:"default:0;column:headquarter_uuid;comment:'总部Uuid'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

func (model *MaterialCategory) IsHeadquarter() bool {
	return model.HeadquarterUuid != 0
}

func (model *MaterialCategory) SetNil() {
	model.MultiLanguageName = MultiLanguageName{}
}

// 用总部分类信息更新子公司分类
func (model *MaterialCategory) UpdateFromHeadquarter(headquarterCategory MaterialCategory) {
	model.Code = headquarterCategory.Code
	model.Name = headquarterCategory.Name
	model.MultiLanguageName.InitByLocaleResponse(headquarterCategory.MultiLanguageName.GetNames())
}
