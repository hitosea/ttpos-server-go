package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// IntegrationStatusReq Grab 门店集成状态回调请求
// 接收 Grab 平台推送的门店集成状态变更通知
// 参考: https://developer.grab.com/docs/grabfood/api/#tag/push-integration-status-webhook
type IntegrationStatusReq struct {
	g.Meta `path:"/partner/integration/status" tags:"Grab Webhook" method:"post" summary:"接收Grab门店集成状态"`
	// 嵌入 SDK 请求结构（包含 PartnerMerchantID, GrabMerchantID, IntegrationStatus）
	grabfood.PushIntegrationStatusWebhookRequest
}

// IntegrationStatusRes Grab 门店集成状态回调响应
type IntegrationStatusRes struct {
	g.Meta `mime:"application/json"`
}
