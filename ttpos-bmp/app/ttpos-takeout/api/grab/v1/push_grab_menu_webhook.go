package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// PushGrabMenuWebhookReq Grab 菜单推送请求
// GrabFood 在 Self-Serve Activation 流程中调用此 webhook 推送现有门店菜单
// 参考: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/push-grab-menu-webhook
type PushGrabMenuWebhookReq struct {
	g.Meta `path:"/partner/pushGrabMenu" tags:"Grab Webhook" method:"post" summary:"响应Grab菜单推送"`

	// Grab 商户 ID
	MerchantID string `json:"merchantID,omitempty" dc:"Grab 商户 ID"`
	// Partner 商户 ID (POS 系统中的商户标识)
	PartnerMerchantID string `json:"partnerMerchantID,omitempty" dc:"Partner 商户 ID"`
	// 货币信息
	Currency grabfood.Currency `json:"currency" dc:"货币信息"`
	// 销售时间数组 (最多 20 个)
	SellingTimes []grabfood.SellingTime `json:"sellingTimes" dc:"销售时间数组"`
	// 菜单分类数组 (最多 100 个)
	Categories []grabfood.MenuCategory `json:"categories" dc:"菜单分类数组"`
}

// PushGrabMenuWebhookRes Grab 菜单推送响应
// 成功时返回 HTTP 204 No Content
type PushGrabMenuWebhookRes struct {
	g.Meta `mime:"application/json"`
}
