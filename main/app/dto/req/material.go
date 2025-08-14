package req

import (
	"ttpos-server-go/app/dto"
)

// MaterialCategoryAddReq 创建物品类别请求
type MaterialCategoryAddReq struct {
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 物品类别名称
}

// MaterialCategoryListReq 获取物品类别列表请求
type MaterialCategoryListReq struct {
}

// MaterialListReq 物品列表查询
type MaterialListReq struct {
	dto.PageReq         // 分页参数
	Keyword      string `form:"keyword" json:"keyword"`             // 关键字
	Status       int    `form:"status" json:"status"`               // 状态，0-全部 1-启用 2-停用
	CategoryUuid uint64 `form:"category_uuid" json:"category_uuid"` // 分类UUID
}

// MaterialDetailReq 物品详情查询
type MaterialDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 物品UUID
}

// MaterialAddReq 添加物品请求
type MaterialAddReq struct {
	LocaleName       dto.LocaleResponse `json:"locale_name" binding:"required"`        // 物品名称
	CategoryUuid     uint64             `json:"category_uuid" binding:"required"`      // 分类UUID
	Status           int                `json:"status" binding:"required"`             // 状态，1-启用 2-停用
	Valuation        float64            `json:"valuation" binding:"required,min=0"`    // 估值率
	InitStock        float64            `json:"init_stock" binding:"required,min=0"`   // 期初库存
	BarcodeValue     string             `json:"barcode_value" binding:"required"`      // 条形码值
	UnitUuid         uint64             `json:"unit_uuid" binding:"required"`          // 基准单位UUID
	UnitList         []MaterialUnitReq  `json:"unit_list" binding:"required,dive"`     // 单位列表
	PurchaseUnitUuid uint64             `json:"purchase_unit_uuid" binding:"required"` // 采购单位UUID
	CostUnitUuid     uint64             `json:"cost_unit_uuid" binding:"required"`     // 成本单位UUID
}

// MaterialUnitReq 物品单位请求
type MaterialUnitReq struct {
	UnitUuid       uint64  `json:"unit_uuid" binding:"required"`             // 单位UUID
	ConversionRate float64 `json:"conversion_rate" binding:"required,min=0"` // 转换率
}

// MaterialEditReq 编辑物品请求
type MaterialEditReq struct {
	Uuid             uint64             `json:"uuid" binding:"required"`               // 物品UUID
	LocaleName       dto.LocaleResponse `json:"locale_name" binding:"required"`        // 物品名称
	CategoryUuid     uint64             `json:"category_uuid" binding:"required"`      // 分类UUID
	Status           int                `json:"status"`                                // 状态，1-启用 0-停用
	Valuation        float64            `json:"valuation" binding:"required,min=0"`    // 估值率
	BarcodeValue     string             `json:"barcode_value" binding:"required"`      // 条形码值
	UnitList         []MaterialUnitReq  `json:"unit_list" binding:"required,dive"`     // 单位列表,新增的非基准单位
	PurchaseUnitUuid uint64             `json:"purchase_unit_uuid" binding:"required"` // 采购单位UUID
	CostUnitUuid     uint64             `json:"cost_unit_uuid" binding:"required"`     // 成本单位UUID
}

// MaterialDeleteReq 删除物品请求
type MaterialDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 物品UUID
}

// MaterialStatusReq 修改物品状态请求
type MaterialStatusReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"`   // 物品UUID
	Status bool   `json:"status" binding:"required"` // 状态，true-上架 false-下架
}

// MaterialSearchReq 物品搜索查询
type MaterialSearchReq struct {
	Keyword string `form:"keyword" json:"keyword" binding:"required"` // 搜索关键词
}

// MaterialImportReq 导入物品请求
type MaterialImportReq struct {
	List []MaterialImportItemReq `json:"list" binding:"required,dive"` // 物品列表
}

// MaterialImportItemReq 导入物品项请求
type MaterialImportItemReq struct {
	MaterialName string  `json:"material_name" binding:"required"` // 物品名称
	CategoryName string  `json:"category_name" binding:"required"` // 分类名称
	UnitName     string  `json:"unit_name" binding:"required"`     // 单位名称
	Price        float64 `json:"price" binding:"required,min=0"`   // 采购单价
	StockNum     float64 `json:"stock_num" binding:"min=0"`        // 库存数量
	BarcodeValue string  `json:"barcode_value"`                    // 条形码值
	Status       bool    `json:"status"`                           // 状态，true-上架 false-下架
	Row          int     `json:"row" binding:"required"`           // excel表的行编号
}
