package conf

import "time"

// GrabPartner Grab Partner Webhook 配置
type GrabPartner struct {
	ClientID     string        `json:"clientId"`     // Partner client_id
	ClientSecret string        `json:"clientSecret"` // Partner client_secret
	Scope        string        `json:"scope"`        // OAuth scope，默认 food.partner_api
	Environment  string        `json:"environment"`  // production 或 staging
	Timeout      time.Duration `json:"timeout"`      // HTTP 超时时间
}

// GrabPartnerMap Grab Partner 配置集合
type GrabPartnerMap map[string]*GrabPartner
