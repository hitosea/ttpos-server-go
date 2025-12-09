// Package v1 定义 Grab Webhook API 接口结构
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SubmitOrderReq Grab 提交订单 Webhook 请求
// 接收 Grab 平台推送的新订单数据
type SubmitOrderReq struct {
	g.Meta `path:"/partner/orders" tags:"Grab Webhook" method:"post" summary:"接收Grab订单"`
	// 请求头
	XGrabSignature string `header:"X-Grab-Signature" dc:"Grab HMAC-SHA256 签名"`
	XGrabTimestamp string `header:"X-Grab-Timestamp" dc:"Grab 时间戳"`
	// 请求体 (原始 JSON，由 Logic 层解析)
}

// SubmitOrderRes Grab 提交订单 Webhook 响应
type SubmitOrderRes struct {
	g.Meta `mime:"application/json"`
}
