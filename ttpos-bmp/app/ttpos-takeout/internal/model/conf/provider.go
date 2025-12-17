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
