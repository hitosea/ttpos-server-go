// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ShopProviderCfg is the golang structure for table shop_provider_cfg.
type ShopProviderCfg struct {
	Id                 uint64 `json:"id"                 orm:"id"                   description:"主键ID"`         // 主键ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"唯一标识"`         // 唯一标识
	ShopUuid           uint64 `json:"shopUuid"           orm:"shop_uuid"            description:"门店UUID"`       // 门店UUID
	ProviderName       string `json:"providerName"       orm:"provider_name"        description:"第三方名称，如 grab"` // 第三方名称，如 grab
	ProviderMerchantId string `json:"providerMerchantId" orm:"provider_merchant_id" description:"第三方商户ID"`      // 第三方商户ID
	ProviderShopStatus string `json:"providerShopStatus" orm:"provider_shop_status" description:"门店集成状态"`       // 门店集成状态
	CreatedAt          int    `json:"createdAt"          orm:"created_at"           description:"创建时间"`         // 创建时间
	UpdatedAt          int    `json:"updatedAt"          orm:"updated_at"           description:"更新时间"`         // 更新时间
	DeletedAt          int    `json:"deletedAt"          orm:"deleted_at"           description:"删除时间"`         // 删除时间
}
