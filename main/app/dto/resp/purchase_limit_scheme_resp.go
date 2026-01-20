package resp

import "ttpos-server-go/app/dto"

// PurchaseLimitSchemeResp 限购方案响应
type PurchaseLimitSchemeResp struct {
	Uuid            uint64                        `json:"uuid"`
	NameLocale      dto.LocaleResponse            `json:"name_locale"`
	Status          int8                          `json:"status"`
	ApplyToAllShops int8                          `json:"apply_to_all_shops"`
	DailyLimit      int                           `json:"daily_limit"`
	Weekdays        []int8                        `json:"weekdays"`
	Items           []PurchaseLimitSchemeItemResp `json:"items"`
	Shops           []uint64                      `json:"shops"`
	CreateTime      int64                         `json:"create_time"`
	UpdateTime      int64                         `json:"update_time"`
}

// PurchaseLimitSchemeItemResp 限购方案物品配置响应
type PurchaseLimitSchemeItemResp struct {
	MaterialUuid uint64  `json:"material_uuid"`
	QuotaLimit   float64 `json:"quota_limit"`
}

// PurchaseLimitSchemeListResp 限购方案列表响应
type PurchaseLimitSchemeListResp struct {
	List []PurchaseLimitSchemeSummaryResp `json:"list"`
	Meta PageMeta                         `json:"meta"`
}

// PurchaseLimitSchemeSummaryResp 限购方案摘要响应（列表用）
type PurchaseLimitSchemeSummaryResp struct {
	Uuid       uint64             `json:"uuid"`
	NameLocale dto.LocaleResponse `json:"name_locale"`
	Status     int8               `json:"status"`
	WeekdayStr string             `json:"weekday_str"` // 例如: "周一、周三、周五"
	ShopCount  int                `json:"shop_count"`  // 门店数量
	ItemCount  int                `json:"item_count"`  // 物品数量
	DailyLimit int                `json:"daily_limit"`
	CreateTime int64              `json:"create_time"`
	UpdateTime int64              `json:"update_time"`
}
