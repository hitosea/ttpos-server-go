package material_resp

import (
	"ttpos-server-go/app/dto"
)

// Material 物品信息
type Material struct {
	Uuid             uint64             `json:"uuid"`               // 物品UUID
	Name             string             `json:"name"`               // 物品名称
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 物品名称
	ErpCode          string             `json:"erp_code"`           // erp编码
	InternalCode     string             `json:"internal_code"`      // 内部编码
	BarcodeValue     string             `json:"barcode_value"`      // 条形码值
	Num              float64            `json:"num"`                // 库存数量
	AvailableNum     float64            `json:"available_num"`      // 可用库存数量
	TransitNum       float64            `json:"transit_num"`        // 在途库存数量
	CategoryUuid     uint64             `json:"category_uuid"`      // 分类UUID
	Image            string             `json:"image"`              // 图片
	Status           int                `json:"status"`             // 状态 1-启用 0-停用
	UnitName         string             `json:"unit_name"`          // 基准单位名称
	UnitUuid         uint64             `json:"unit_uuid"`          // 基准单位UUID
	PurchaseUnitName string             `json:"purchase_unit_name"` // 采购单位名称
	PurchaseUnitUuid uint64             `json:"purchase_unit_uuid"` // 采购单位UUID
	CostUnitName     string             `json:"cost_unit_name"`     // 成本单位名称
	CostUnitUuid     uint64             `json:"cost_unit_uuid"`     // 成本单位UUID
	UnitList         []MaterialUnit     `json:"unit_list"`          // 单位列表

	UnitLocaleName         dto.LocaleResponse `json:"unit_locale_name"`          // 基准单位名称
	CostUnitLocaleName     dto.LocaleResponse `json:"cost_unit_locale_name"`     // 成本单位名称
	PurchaseUnitLocaleName dto.LocaleResponse `json:"purchase_unit_locale_name"` // 采购单位名称
}

// MaterialListWithPaginationResp 物品列表响应
type MaterialListWithPaginationResp struct {
	List []Material       `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// MaterialDetailResp 物品详情响应
type MaterialDetailResp struct {
	Uuid                   uint64               `json:"uuid"`                      // 物品UUID
	LocaleName             dto.LocaleResponse   `json:"locale_name"`               // 物品名称
	Code                   string               `json:"code"`                      // 物品编码
	CategoryUuid           uint64               `json:"category_uuid"`             // 分类UUID
	CategoryName           string               `json:"category_name"`             // 分类名称
	Status                 int                  `json:"status"`                    // 状态 1-启用 0-停用
	Valuation              float64              `json:"valuation"`                 // 估值率
	BarcodeValue           string               `json:"barcode_value"`             // 条形码值
	InternalCode           string               `json:"internal_code"`             // 内部编码
	UnitName               string               `json:"unit_name"`                 // 单位名称
	UnitUuid               uint64               `json:"unit_uuid"`                 // 单位UUID
	FromUnitUuid           uint64               `json:"from_unit_uuid"`            // 来源单位UUID
	UnitLocaleName         dto.LocaleResponse   `json:"unit_locale_name"`          // 单位名称
	UnitList               MaterialUnitListResp `json:"unit_list"`                 // 单位列表
	PurchaseUnitName       string               `json:"purchase_unit_name"`        // 采购单位名称
	PurchaseUnitLocaleName dto.LocaleResponse   `json:"purchase_unit_locale_name"` // 采购单位名称
	PurchaseUnitUuid       uint64               `json:"purchase_unit_uuid"`        // 采购单位UUID
	FromPurchaseUnitUuid   uint64               `json:"from_purchase_unit_uuid"`   // 来源采购单位UUID
	CostUnitName           string               `json:"cost_unit_name"`            // 成本单位名称
	CostUnitLocaleName     dto.LocaleResponse   `json:"cost_unit_locale_name"`     // 成本单位名称
	CostUnitUuid           uint64               `json:"cost_unit_uuid"`            // 成本单位UUID
	FromCostUnitUuid       uint64               `json:"from_cost_unit_uuid"`       // 来源成本单位UUID
	IsEditable             bool                 `json:"is_editable"`               // 是否可编辑
}

// MaterialStockDetailResp 物品库存详情响应
type MaterialStockDetailResp struct {
	Uuid       uint64             `json:"uuid"`        // 物品UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 物品名称
	Code       string             `json:"code"`        // 物品编码
	Warehouses WarehouseList      `json:"warehouses"`  // 库存列表
}

// WarehouseList 仓库列表
type WarehouseList struct {
	Amount float64     `json:"amount"` // 合计库存数
	List   []Warehouse `json:"list"`
}

// Warehouse 仓库
type Warehouse struct {
	Uuid       uint64             `json:"uuid"`        // 仓库UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 仓库名称
	Num        float64            `json:"num"`         // 物品库存数量
}

// MaterialUnitListResp 物品单位列表响应
type MaterialUnitListResp struct {
	List []MaterialUnit `json:"list"`
}

// MaterialUnit 物品单位
type MaterialUnit struct {
	Uuid           uint64             `json:"uuid"`            // 单位UUID
	FromUnitUuid   uint64             `json:"from_unit_uuid"`  // 来源单位UUID
	Name           string             `json:"name"`            // 单位名称
	LocaleName     dto.LocaleResponse `json:"locale_name"`     // 单位名称
	ConversionRate float64            `json:"conversion_rate"` // 转换率
}

// MaterialSearchResp 物品搜索响应
type MaterialSearchResp struct {
	List []Material `json:"list"`
}

// MaterialCategory 物品分类
type MaterialCategory struct {
	Uuid       uint64             `json:"uuid"`        // 分类UUID
	Name       string             `json:"name"`        // 分类名称
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分类名称
	Code       string             `json:"code"`        // 分类编码
	Sort       int                `json:"sort"`        // 排序
	IsRelated  bool               `json:"is_related"`  // 是否关联了物品
	IsEditable bool               `json:"is_editable"` // 是否可编辑
}

// MaterialImportUnitListItem 导入物品单位列表项
type MaterialImportUnitListItem struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 单位名称
	Uuid       uint64             `json:"uuid"`        // 单位UUID
}

// MaterialCategoryListResp 物品类别列表响应
type MaterialCategoryListResp struct {
	List []MaterialCategory `json:"list"`
}

// ProductBomCardDetailResp 规格商品成本卡详情响应
type ProductBomCardDetailResp struct {
	Uuid       uint64                   `json:"uuid"`        // 成本卡UUID
	Name       string                   `json:"name"`        // 成本卡名称
	LocaleName dto.LocaleResponse       `json:"locale_name"` // 成本卡名称
	Num        float64                  `json:"num"`         // 加工份数
	Materials  []ProductBomCardMaterial `json:"materials"`   // 材料列表
	IsEditable bool                     `json:"is_editable"` // 是否可编辑
}

type ProductBomCardMaterial struct {
	Material MaterialInfo   `json:"material"`  // 材料信息
	Num      float64        `json:"num"`       // 净耗量
	Unit     MaterialUnit   `json:"unit"`      // 成本单位
	UnitList []MaterialUnit `json:"unit_list"` // 单位列表, 用于成本卡编辑选择单位
}

type MaterialInfo struct {
	Uuid uint64 `json:"uuid"` // 材料UUID
	Name string `json:"name"` // 材料名称
	Code string `json:"code"` // 材料编码
}

// MaterialImportListItem 导入物品列表项
type MaterialImportListItem struct {
	LocaleName        dto.LocaleResponse `json:"locale_name" binding:"required"`      // 物品名称
	LocaleNameIsExist dto.LocaleResponse `json:"locale_name_is_exist"`                // 物品名称是否存在，对应的key不为空则表示存在
	CategoryName      string             `json:"category_name" binding:"required"`    // 分类名称
	CategoryUuid      uint64             `json:"category_uuid"`                       // 分类UUID
	UnitUuid          uint64             `json:"unit_uuid"`                           // 单位UUID
	UnitName          string             `json:"unit_name"`                           // 单位名称
	Status            int                `json:"status"`                              // 状态 1-启用 0-停用
	BarcodeValue      string             `json:"barcode_value"`                       // 条形码值
	BarcodeIsExist    bool               `json:"barcode_is_exist"`                    // 条形码是否存在，存在则不保存
	Valuation         float64            `json:"valuation" binding:"required,min=0"`  // 估值率
	InitStock         float64            `json:"init_stock" binding:"required,min=0"` // 期初库存
	Row               int                `json:"row" binding:"required"`              // 行号
}

// MaterialImportResp 导入物品响应
type MaterialImportResp struct {
	List         []MaterialImportListItem     `json:"list" binding:"required,dive"`      // 商品列表
	CategoryList []MaterialCategory           `json:"category_list" binding:"required"`  // 分类列表
	UnitList     []MaterialImportUnitListItem `json:"unit_list" binding:"required,dive"` // 单位列表
}
