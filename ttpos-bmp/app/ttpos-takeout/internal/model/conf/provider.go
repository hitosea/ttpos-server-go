package conf

import "time"

// Skootar Skootar配置
type Skootar struct {
	Endpoint string `json:"endpoint"`
	ApiKey   string `json:"apiKey"`
	UserName string `json:"userName"`
	Channel  string `json:"channel"`
}

// GrabPlatform Grab 平台配置（platform 节点）
type Grab struct {
	SecretKey    string        `json:"secretKey"`    // Webhook 签名密钥
	ClientID     string        `json:"clientId"`     // OAuth Client ID
	ClientSecret string        `json:"clientSecret"` // OAuth Client Secret
	Environment  string        `json:"environment"`  // production 或 staging
	Timeout      time.Duration `json:"timeout"`      // API 超时时间
}

// Lineman LINE MAN 平台配置（platform 节点）
type Lineman struct {
	PartnerId    string        `json:"partnerId"`    // Partner ID（默认值）
	ClientID     string        `json:"clientId"`     // OAuth Client ID
	ClientSecret string        `json:"clientSecret"` // OAuth Client Secret
	SecretKey    string        `json:"secretKey"`    // JWT 签名密钥
	Endpoint     string        `json:"endpoint"`     // API 端点地址
	Environment  string        `json:"environment"`  // production 或 staging
	Timeout      time.Duration `json:"timeout"`      // API 超时时间
}

// LinemanPartner LINE MAN Partner 配置
type LinemanPartner struct {
	ClientID     string        `json:"clientId"`     // OAuth Client ID
	ClientSecret string        `json:"clientSecret"` // OAuth Client Secret
	Environment  string        `json:"environment"`  // production 或 staging
	Timeout      time.Duration `json:"timeout"`      // API 超时时间
}

// LinemanPartnerMap LINE MAN Partner 配置映射（partner_code -> config）
type LinemanPartnerMap map[string]*LinemanPartner
