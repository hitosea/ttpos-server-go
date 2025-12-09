package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// IntegrationStatusReq Grab 门店集成状态回调请求
// 接收 Grab 平台推送的门店集成状态变更通知
type IntegrationStatusReq struct {
	g.Meta `path:"/partner/integration/status" tags:"Grab Webhook" method:"post" summary:"接收Grab门店集成状态"`
	// 请求头
	XGrabSignature string `header:"X-Grab-Signature" dc:"Grab HMAC-SHA256 签名"`
	XGrabTimestamp string `header:"X-Grab-Timestamp" dc:"Grab 时间戳"`
	// 请求体 (原始 JSON，由 Logic 层解析)
}

// IntegrationStatusRes Grab 门店集成状态回调响应
type IntegrationStatusRes struct {
	g.Meta `mime:"application/json"`
}
