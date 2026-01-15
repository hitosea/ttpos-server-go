package resp

// PurchaseQuotaConfigResp 限购配置响应
type PurchaseQuotaConfigResp struct {
	Uuid            uint64   `json:"uuid"`
	MaterialUuid    uint64   `json:"material_uuid"`
	MaterialCode    string   `json:"material_code"`
	MaterialName    string   `json:"material_name"`
	UnitUuid        uint64   `json:"unit_uuid"`
	UnitName        string   `json:"unit_name"`
	QuotaLimit      float64  `json:"quota_limit"`
	ApplyToAllShops uint8    `json:"apply_to_all_shops"`
	ShopUuids       []uint64 `json:"shop_uuids,omitempty"`
	PeriodType      uint8    `json:"period_type"`
	StrictMode      uint8    `json:"strict_mode"`
	ConfigSource    uint8    `json:"config_source"`
	Status          uint8    `json:"status"`
	CreateTime      int64    `json:"create_time"`
	UpdateTime      int64    `json:"update_time"`
}

// PurchaseQuotaUsageResp 限购使用情况响应
type PurchaseQuotaUsageResp struct {
	MaterialUuid uint64  `json:"material_uuid"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	UnitUuid     uint64  `json:"unit_uuid"`
	UnitName     string  `json:"unit_name"`
	QuotaLimit   float64 `json:"quota_limit"`
	UsedQty      float64 `json:"used_qty"`
	RemainQty    float64 `json:"remain_qty"`
	PeriodType   uint8   `json:"period_type"`
}
