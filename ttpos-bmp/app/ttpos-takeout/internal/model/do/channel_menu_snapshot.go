// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ChannelMenuSnapshot is the golang structure of table takeout_channel_menu_snapshot for DAO operations like Where/Data.
type ChannelMenuSnapshot struct {
	g.Meta         `orm:"table:takeout_channel_menu_snapshot, do:true"`
	Id             any // 主键ID
	Uuid           any // 唯一标识
	ShopUuid       any // 商户UUID
	ProviderName   any // 渠道名称 (grab, lineman)
	MenuData       any // 菜单数据快照 (JSON)
	CreatedAt      any // 创建时间
	UpdatedAt      any // 更新时间
	DeletedAt      any // 删除时间
	SyncState      any // 同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED
	TtposMenuData  any // TTPOS 侧菜单原始数据 (JSON)
	TtposUpdatedAt any // TTPOS 侧菜单数据更新时间
}
