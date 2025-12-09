// Package v1 定义 Grab 获取 Partner Token 的 Webhook 接口
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OAuthPartnerWebhookReq Grab 获取 Partner Token 的 Webhook 请求
type OAuthPartnerWebhookReq struct {
	g.Meta       `path:"/oauth/partner" tags:"Grab OAuth" method:"post" summary:"Grab 获取 Partner Token"`
	ClientId     string `json:"client_id" v:"required" dc:"Grab 分配的 client_id"`
	ClientSecret string `json:"client_secret" v:"required" dc:"Grab 分配的 client_secret"`
	Scope        string `json:"scope" dc:"Grab 分配的 scope"`
}

// OAuthPartnerWebhookRes Grab 获取 Partner Token 的 Webhook 响应
type OAuthPartnerWebhookRes struct {
	g.Meta      `mime:"application/json"`
	AccessToken string `json:"access_token" dc:"访问令牌"`
	ExpiresIn   int    `json:"expires_in" dc:"过期时间"`
	TokenType   string `json:"token_type" dc:"令牌类型" default:"Bearer"`
}
