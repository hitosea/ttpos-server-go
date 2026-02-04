package req

import "ttpos-server-go/app/dto"

// PurchaseLimitSchemeCreateReq 创建限购方案请求
type PurchaseLimitSchemeCreateReq struct {
	LocaleName      dto.LocaleResponse           `json:"locale_name" binding:"required"` // 多语言名称
	Status          int8                         `json:"status"`                         // 0-关闭 1-开启
	ApplyToAllShops int8                         `json:"apply_to_all_shops"`             // 0-否 1-是
	DailyLimit      int                          `json:"daily_limit"`                    // -1=不限制 其他=限制
	Weekdays        []int8                       `json:"weekdays"`                       // 星期配置 1-7 表示周一到周日，逗号分隔
	Items           []PurchaseLimitSchemeItemReq `json:"items"`                          // 物品配置
	Shops           []uint64                     `json:"shops"`                          // 门店UUID列表
}

// PurchaseLimitSchemeItemReq 限购方案物品配置请求
type PurchaseLimitSchemeItemReq struct {
	MaterialUuid    uint64  `json:"material_uuid" binding:"required"` // 物品UUID
	QuotaLimit      float64 `json:"quota_limit"`                      // 限购数量
	IsAllowPurchase string  `json:"is_allow_purchase"`                // 是否允许采购 yes/no，默认 yes
}

// PurchaseLimitSchemeUpdateReq 更新限购方案请求
type PurchaseLimitSchemeUpdateReq struct {
	Uuid            uint64                       `json:"uuid" binding:"required"`        // 方案UUID
	LocaleName      dto.LocaleResponse           `json:"locale_name" binding:"required"` // 多语言名称
	Status          int8                         `json:"status"`                         // 0-关闭 1-开启
	ApplyToAllShops int8                         `json:"apply_to_all_shops"`             // 0-否 1-是
	DailyLimit      int                          `json:"daily_limit"`                    // -1=不限制 其他=限制
	Weekdays        []int8                       `json:"weekdays"`                       // 星期配置 1-7 表示周一到周日，逗号分隔
	Items           []PurchaseLimitSchemeItemReq `json:"items"`                          // 物品配置
	Shops           []uint64                     `json:"shops"`                          // 门店UUID列表
}

// PurchaseLimitSchemeDetailReq 限购方案详情请求
type PurchaseLimitSchemeDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid"` // 方案UUID
}

// PurchaseLimitSchemeListReq 限购方案列表请求
type PurchaseLimitSchemeListReq struct {
	PageNo   int    `json:"page_no" form:"page_no"`     // 页码
	PageSize int    `json:"page_size" form:"page_size"` // 每页数量
	Status   *int8  `json:"status" form:"status"`       // 0-关闭 1-开启
	Name     string `json:"name" form:"name"`           // 方案名称
}

// PurchaseLimitSchemeDeleteReq 删除限购方案请求
type PurchaseLimitSchemeDeleteReq struct {
	Uuid uint64 `json:"uuid" form:"uuid"` // 方案UUID
}
