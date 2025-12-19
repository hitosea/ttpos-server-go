// Package v1 定义 Grab Webhook API 接口结构
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// SubmitOrderReq Grab 提交订单 Webhook 请求
// 接收 Grab 平台推送的新订单数据
type SubmitOrderReq struct {
	g.Meta `path:"/partner/orders" tags:"Grab Webhook" method:"post" summary:"接收Grab订单"`
	*grabfood.SubmitOrderRequest
}

// SubmitOrderRes Grab 提交订单 Webhook 响应
type SubmitOrderRes struct {
	g.Meta `mime:"application/json"`
}
