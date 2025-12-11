package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// MenuSyncStateReq Grab 菜单同步状态回调请求
// 接收 Grab 平台推送的菜单同步结果通知
type MenuSyncStateReq struct {
	g.Meta `path:"/partner/menu/sync/state" tags:"Grab Webhook" method:"post" summary:"接收Grab菜单同步状态"`
	*grabfood.MenuSyncWebhookRequest
}

// MenuSyncStateRes Grab 菜单同步状态回调响应
type MenuSyncStateRes struct {
	g.Meta `mime:"application/json"`
}
