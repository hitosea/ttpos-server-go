package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// GetMenuReq Grab 拉取菜单请求
// Grab 平台主动拉取商户菜单数据
type GetMenuReq struct {
	g.Meta `path:"/partner/menu" tags:"Grab Webhook" method:"get" summary:"响应Grab菜单拉取"`
	// 请求头
	XGrabSignature    string `header:"X-Grab-Signature" dc:"Grab HMAC-SHA256 签名"`
	XGrabTimestamp    string `header:"X-Grab-Timestamp" dc:"Grab 时间戳"`
	PartnerMerchantID string `query:"partnerMerchantID" dc:"合作商户ID"`
}

// GetMenuRes Grab 菜单响应
type GetMenuRes struct {
	g.Meta `mime:"application/json"`
	*grabfood.GetMenuNewResponse
}
