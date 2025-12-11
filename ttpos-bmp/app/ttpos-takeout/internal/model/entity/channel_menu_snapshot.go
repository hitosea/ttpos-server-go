// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ChannelMenuSnapshot is the golang structure for table channel_menu_snapshot.
type ChannelMenuSnapshot struct {
	Id             uint64 `json:"id"             orm:"id"               description:"主键ID"`                                        // 主键ID
	Uuid           uint64 `json:"uuid"           orm:"uuid"             description:"唯一标识"`                                        // 唯一标识
	ShopUuid       uint64 `json:"shopUuid"       orm:"shop_uuid"        description:"商户UUID"`                                      // 商户UUID
	ProviderName   string `json:"providerName"   orm:"provider_name"    description:"渠道名称 (grab, lineman)"`                        // 渠道名称 (grab, lineman)
	MenuData       string `json:"menuData"       orm:"menu_data"        description:"菜单数据快照 (JSON)"`                               // 菜单数据快照 (JSON)
	CreatedAt      int    `json:"createdAt"      orm:"created_at"       description:"创建时间"`                                        // 创建时间
	UpdatedAt      int    `json:"updatedAt"      orm:"updated_at"       description:"更新时间"`                                        // 更新时间
	DeletedAt      int    `json:"deletedAt"      orm:"deleted_at"       description:"删除时间"`                                        // 删除时间
	SyncState      string `json:"syncState"      orm:"sync_state"       description:"同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED"` // 同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED
	TtposMenuData  string `json:"ttposMenuData"  orm:"ttpos_menu_data"  description:"TTPOS 侧菜单原始数据 (JSON)"`                        // TTPOS 侧菜单原始数据 (JSON)
	TtposUpdatedAt int    `json:"ttposUpdatedAt" orm:"ttpos_updated_at" description:"TTPOS 侧菜单数据更新时间"`                             // TTPOS 侧菜单数据更新时间
}
