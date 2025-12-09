// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ChannelMenuSnapshot is the golang structure for table channel_menu_snapshot.
type ChannelMenuSnapshot struct {
	Id           uint64 `json:"id"           orm:"id"            description:"主键ID"`                 // 主键ID
	Uuid         uint64 `json:"uuid"         orm:"uuid"          description:"唯一标识"`                 // 唯一标识
	ShopUuid     uint64 `json:"shopUuid"     orm:"shop_uuid"     description:"商户UUID"`               // 商户UUID
	ProviderName string `json:"providerName" orm:"provider_name" description:"渠道名称 (grab, lineman)"` // 渠道名称 (grab, lineman)
	MenuData     string `json:"menuData"     orm:"menu_data"     description:"菜单数据快照 (JSON)"`        // 菜单数据快照 (JSON)
	CreateTime   int    `json:"createTime"   orm:"create_time"   description:"创建时间"`                 // 创建时间
	UpdateTime   int    `json:"updateTime"   orm:"update_time"   description:"更新时间"`                 // 更新时间
}
