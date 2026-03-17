package resp

import "ttpos-server-go/app/dto"

// AutoReceiptRuleListResp 规则列表响应
type AutoReceiptRuleListResp struct {
	List               []AutoReceiptRuleGroup `json:"list"`
	ConfiguredCount    int                    `json:"configured_count"`
	UnconfiguredCount  int                    `json:"unconfigured_count"`
	UnconfiguredShops  []UnconfiguredShopItem `json:"unconfigured_shops"`
}

// UnconfiguredShopItem 未配置规则的门店
type UnconfiguredShopItem struct {
	Uuid uint64 `json:"uuid"`
	Name string `json:"name"`
}

// AutoReceiptRuleGroup 规则卡片（一条主表记录 = 一个卡片）
type AutoReceiptRuleGroup struct {
	Uuid                uint64                `json:"uuid"`
	LocaleName          dto.LocaleResponse    `json:"locale_name"`
	WarehouseErpCode    string                `json:"warehouse_erp_code"`
	WarehouseLocaleName dto.LocaleResponse    `json:"warehouse_locale_name"`
	DelayDays           int                   `json:"delay_days"`
	Status              int                   `json:"status"`
	ShopCount           int                   `json:"shop_count"`
	Shops               []AutoReceiptRuleShop `json:"shops"`
}

// AutoReceiptRuleShop 规则门店信息
type AutoReceiptRuleShop struct {
	Uuid     uint64 `json:"uuid"`
	ShopUuid uint64 `json:"shop_uuid"`
	ShopCode string `json:"shop_code"`
	ShopName string `json:"shop_name"`
}

// AutoReceiptShopListResp 可选门店列表响应
type AutoReceiptShopListResp struct {
	List []AutoReceiptShopItem `json:"list"`
}

// AutoReceiptShopItem 可选门店项
type AutoReceiptShopItem struct {
	Uuid           uint64 `json:"uuid"`
	Name           string `json:"name"`
	StoreCode      string `json:"store_code"`
	Status         int    `json:"status"`
	Disabled       bool   `json:"disabled"`
	DisabledReason string `json:"disabled_reason"`
}

func (s AutoReceiptShopItem) GetStoreCode() string { return s.StoreCode }
func (s AutoReceiptShopItem) GetUuid() uint64      { return s.Uuid }

// AutoReceiptWarehouseListResp 发货仓库列表响应
type AutoReceiptWarehouseListResp struct {
	List []AutoReceiptWarehouseItem `json:"list"`
}

// AutoReceiptWarehouseItem 发货仓库项
type AutoReceiptWarehouseItem struct {
	ErpCode    string             `json:"erp_code"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
	Type       string             `json:"type"`
}

// AutoReceiptLogListResp 自动收货记录列表响应
type AutoReceiptLogListResp struct {
	List []AutoReceiptLogItem `json:"list"`
	Meta dto.PageResponse     `json:"meta"`
}

// AutoReceiptLogItem 自动收货记录项
type AutoReceiptLogItem struct {
	Uuid              uint64 `json:"uuid"`
	ShopCompanyUuid   uint64 `json:"shop_company_uuid"`
	ShopName          string `json:"shop_name"`
	ReceiptOrderUuid  uint64 `json:"receipt_order_uuid"`
	ReceiptOrderNo    string `json:"receipt_order_no"`
	ReceiptErpOrderNo string `json:"receipt_erp_order_no"`
	ReceiptTime       int64  `json:"receipt_time"`
}
