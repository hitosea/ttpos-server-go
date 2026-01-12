package req

// PurchaseQuotaConfigCreateReq 创建限购配置请求
type PurchaseQuotaConfigCreateReq struct {
	MaterialUuid    uint64   `json:"material_uuid" binding:"required" label:"物品UUID"`
	UnitUuid        uint64   `json:"unit_uuid" binding:"required" label:"单位UUID"`
	QuotaLimit      float64  `json:"quota_limit" binding:"required,gt=0" label:"限购数量"`
	ApplyToAllShops uint8    `json:"apply_to_all_shops" binding:"oneof=0 1" label:"是否应用到全部店铺"`
	ShopUuids       []uint64 `json:"shop_uuids" label:"店铺UUID列表"`
	PeriodType      uint8    `json:"period_type" binding:"oneof=0 1" label:"周期类型"`
	StrictMode      uint8    `json:"strict_mode" binding:"oneof=1" label:"超限策略"`
	ConfigSource    uint8    `json:"config_source" binding:"oneof=1 2" label:"配置来源"`
	Status          uint8    `json:"status" binding:"oneof=0 1" label:"状态"`
}

// PurchaseQuotaConfigUpdateReq 更新限购配置请求
type PurchaseQuotaConfigUpdateReq struct {
	Uuid            uint64   `json:"uuid" binding:"required" label:"配置UUID"`
	MaterialUuid    uint64   `json:"material_uuid" binding:"required" label:"物品UUID"`
	UnitUuid        uint64   `json:"unit_uuid" binding:"required" label:"单位UUID"`
	QuotaLimit      float64  `json:"quota_limit" binding:"required,gt=0" label:"限购数量"`
	ApplyToAllShops uint8    `json:"apply_to_all_shops" binding:"oneof=0 1" label:"是否应用到全部店铺"`
	ShopUuids       []uint64 `json:"shop_uuids" label:"店铺UUID列表"`
	PeriodType      uint8    `json:"period_type" binding:"oneof=0 1" label:"周期类型"`
	StrictMode      uint8    `json:"strict_mode" binding:"oneof=1" label:"超限策略"`
	ConfigSource    uint8    `json:"config_source" binding:"oneof=1 2" label:"配置来源"`
	Status          uint8    `json:"status" binding:"oneof=0 1" label:"状态"`
}

// PurchaseQuotaConfigListReq 限购配置列表请求
type PurchaseQuotaConfigListReq struct {
	MaterialUuid uint64 `form:"material_uuid" label:"物品UUID"`
	ShopUuid     uint64 `form:"shop_uuid" label:"店铺UUID"`
	Status       *uint8 `form:"status" label:"状态"`
	Page         int    `form:"page" binding:"required,min=1" label:"页码"`
	PageSize     int    `form:"page_size" binding:"required,min=1,max=100" label:"每页数量"`
}

// PurchaseQuotaConfigDetailReq 限购配置详情请求
type PurchaseQuotaConfigDetailReq struct {
	Uuid uint64 `form:"uuid" binding:"required" label:"配置UUID"`
}

// PurchaseQuotaConfigDeleteReq 删除限购配置请求
type PurchaseQuotaConfigDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required" label:"配置UUID"`
}

// PurchaseQuotaUsageReq 限购使用情况请求
type PurchaseQuotaUsageReq struct {
	MaterialUuid uint64 `form:"material_uuid" binding:"required" label:"物品UUID"`
	UnitUuid     uint64 `form:"unit_uuid" binding:"required" label:"单位UUID"`
	ShopUuid     uint64 `form:"shop_uuid" binding:"required" label:"店铺UUID"`
}
