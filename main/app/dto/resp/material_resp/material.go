package material_resp

import (
	"ttpos-server-go/app/dto"
)

// Material 物品信息
type Material struct {
	Uuid         uint64 `json:"uuid"`          // 物品UUID
	Name         string `json:"name"`          // 物品名称
	CategoryUuid uint64 `json:"category_uuid"` // 分类UUID
	Image        string `json:"image"`         // 图片
	Status       int    `json:"status"`        // 状态 1-启用 2-停用
	UnitName     string `json:"unit_name"`     // 单位名称
}

// MaterialListResp 物品列表响应
type MaterialListResp struct {
	List []Material `json:"list"`
}

// MaterialListWithPaginationResp 物品列表响应
type MaterialListWithPaginationResp struct {
	List MaterialListResp `json:"list"`
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
	Name           string  `json:"name"`            // 单位名称
	ConversionRate float64 `json:"conversion_rate"` // 转换率
}

// MaterialSearchResp 物品搜索响应
type MaterialSearchResp struct {
	List []Material `json:"list"`
}

// MaterialImportResp 导入物品响应
type MaterialImportResp struct {
	List         []MaterialImportListItem     `json:"list"`          // 物品列表
	CategoryList []MaterialCategory           `json:"category_list"` // 分类列表
	UnitList     []MaterialImportUnitListItem `json:"unit_list"`     // 单位列表
}

// MaterialImportListItem 导入物品列表项
type MaterialImportListItem struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`            // 物品名称
	CategoryName        string             `json:"category_name"`          // 分类名称
	UnitName            string             `json:"unit_name"`              // 单位名称
	Price               float64            `json:"price"`                  // 采购单价
	StockNum            float64            `json:"stock_num"`              // 库存数量
	BarcodeValue        string             `json:"barcode_value"`          // 条形码值
	Status              bool               `json:"status"`                 // 状态
	Row                 int                `json:"row"`                    // excel表的行编号
	MaterialNameIsExist dto.LocaleResponse `json:"material_name_is_exist"` // 物品名称是否存在
	BarcodeIsExist      bool               `json:"barcode_is_exist"`       // 条形码是否存在
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
