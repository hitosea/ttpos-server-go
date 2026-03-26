package req

import "ttpos-server-go/app/dto"

// CreateAutoReceiptRuleReq 创建自动收货规则
type CreateAutoReceiptRuleReq struct {
	LocaleName       dto.LocaleResponse `json:"locale_name" binding:"required"` // 规则名称（多语言）
	WarehouseErpCode string             `json:"warehouse_erp_code" binding:"required"`
	ShopUuids        []uint64           `json:"shop_uuids" binding:"required,min=1"`
	DelayDays        int                `json:"delay_days" binding:"gte=0"`
	Status           *int               `json:"status" binding:"required,oneof=0 1"`
}

// UpdateAutoReceiptRuleReq 更新自动收货规则（全量更新，所有字段必填）
type UpdateAutoReceiptRuleReq struct {
	Uuid             uint64             `json:"uuid" binding:"required"`
	LocaleName       dto.LocaleResponse `json:"locale_name" binding:"required"`        // 规则名称（多语言）
	WarehouseErpCode string             `json:"warehouse_erp_code" binding:"required"` // 发货仓库ERP编码
	DelayDays        int                `json:"delay_days" binding:"gte=0"`
	Status           *int               `json:"status" binding:"required,oneof=0 1"`
	ShopUuids        []uint64           `json:"shop_uuids" binding:"required,min=1"`
}

// DeleteAutoReceiptRuleReq 删除自动收货规则（整条卡片）
type DeleteAutoReceiptRuleReq struct {
	Uuids []uint64 `json:"uuids" binding:"required,min=1"`
}

// AutoReceiptRuleListReq 规则列表查询（无筛选条件，返回当前租户全部规则）
type AutoReceiptRuleListReq struct{}

// AutoReceiptShopListReq 可选门店列表查询
type AutoReceiptShopListReq struct {
	WarehouseErpCode string `form:"warehouse_erp_code" binding:"required"`
}

// AutoReceiptLogListReq 自动收货记录列表查询
type AutoReceiptLogListReq struct {
	dto.PageReq
	ShopCompanyUuid uint64 `form:"shop_company_uuid"`                                           // 门店筛选
	StartTime       string `form:"start_time" binding:"omitempty,datetime=2006-01-02 15:04:05"` // 收货时间范围-开始（格式: 2006-01-02 15:04:05）
	EndTime         string `form:"end_time" binding:"omitempty,datetime=2006-01-02 15:04:05"`   // 收货时间范围-结束（格式: 2006-01-02 15:04:05）
}

// AutoReceiptLogDetailReq 自动收货记录详情
type AutoReceiptLogDetailReq struct {
	Uuid uint64 `form:"uuid" binding:"required"`
}
