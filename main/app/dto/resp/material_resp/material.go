package material_resp

import (
	"ttpos-server-go/app/dto"
)

// Material 物品信息
type Material struct {
	Uuid             uint64         `json:"uuid"`               // 物品UUID
	Name             string         `json:"name"`               // 物品名称
	ErpCode          string         `json:"erp_code"`           // erp编码
	CategoryUuid     uint64         `json:"category_uuid"`      // 分类UUID
	Image            string         `json:"image"`              // 图片
	Status           int            `json:"status"`             // 状态 1-启用 0-停用
	UnitName         string         `json:"unit_name"`          // 基准单位名称
	UnitUuid         uint64         `json:"unit_uuid"`          // 基准单位UUID
	PurchaseUnitName string         `json:"purchase_unit_name"` // 采购单位名称
	PurchaseUnitUuid uint64         `json:"purchase_unit_uuid"` // 采购单位UUID
	CostUnitName     string         `json:"cost_unit_name"`     // 成本单位名称
	CostUnitUuid     uint64         `json:"cost_unit_uuid"`     // 成本单位UUID
	UnitList         []MaterialUnit `json:"unit_list"`          // 单位列表
}

// MaterialListWithPaginationResp 物品列表响应
type MaterialListWithPaginationResp struct {
	List []Material       `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// MaterialDetailResp 物品详情响应
type MaterialDetailResp struct {
	Uuid             uint64               `json:"uuid"`               // 物品UUID
	LocaleName       dto.LocaleResponse   `json:"locale_name"`        // 物品名称
	Code             string               `json:"code"`               // 物品编码
	CategoryUuid     uint64               `json:"category_uuid"`      // 分类UUID
	CategoryName     string               `json:"category_name"`      // 分类名称
	Status           int                  `json:"status"`             // 状态 1-启用 0-停用
	Valuation        float64              `json:"valuation"`          // 估值率
	BarcodeValue     string               `json:"barcode_value"`      // 条形码值
	UnitName         string               `json:"unit_name"`          // 单位名称
	UnitList         MaterialUnitListResp `json:"unit_list"`          // 单位列表
	PurchaseUnitName string               `json:"purchase_unit_name"` // 采购单位名称
	PurchaseUnitUuid uint64               `json:"purchase_unit_uuid"` // 采购单位UUID
	CostUnitName     string               `json:"cost_unit_name"`     // 成本单位名称
	CostUnitUuid     uint64               `json:"cost_unit_uuid"`     // 成本单位UUID
}

// MaterialUnitListResp 物品单位列表响应
type MaterialUnitListResp struct {
	List []MaterialUnit `json:"list"`
}

// MaterialUnit 物品单位
type MaterialUnit struct {
	Uuid           uint64  `json:"uuid"`            // 单位UUID
	Name           string  `json:"name"`            // 单位名称
	ConversionRate float64 `json:"conversion_rate"` // 转换率
}

// MaterialSearchResp 物品搜索响应
type MaterialSearchResp struct {
	List []Material `json:"list"`
}

// MaterialCategory 物品分类
type MaterialCategory struct {
	Uuid uint64 `json:"uuid"` // 分类UUID
	Name string `json:"name"` // 分类名称
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
	Uuid      uint64                   `json:"uuid"`      // 成本卡UUID
	Name      string                   `json:"name"`      // 成本卡名称
	Num       float64                  `json:"num"`       // 加工份数
	Materials []ProductBomCardMaterial `json:"materials"` // 材料列表
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
