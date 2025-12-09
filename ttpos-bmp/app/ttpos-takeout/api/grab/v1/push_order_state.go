package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PushOrderStateReq Grab 订单状态变更 Webhook 请求
// 接收 Grab 平台推送的订单状态变更通知
type PushOrderStateReq struct {
	g.Meta `path:"/partner/orders/state" tags:"Grab Webhook" method:"put" summary:"接收Grab订单状态变更"`
	// 请求头
	XGrabSignature string `header:"X-Grab-Signature" dc:"Grab HMAC-SHA256 签名"`
	XGrabTimestamp string `header:"X-Grab-Timestamp" dc:"Grab 时间戳"`
	// 请求体 (原始 JSON，由 Logic 层解析)
}

// PushOrderStateRes Grab 订单状态变更 Webhook 响应
type PushOrderStateRes struct {
	g.Meta `mime:"application/json"`
}
