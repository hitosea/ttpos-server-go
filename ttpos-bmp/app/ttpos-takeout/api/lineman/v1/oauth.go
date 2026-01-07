package v1

import "github.com/gogf/gf/v2/frame/g"

// OAuthTokenReq OAuth 令牌请求
// LINE MAN OAuth 2.0 Client Credentials 认证请求
type OAuthTokenReq struct {
	g.Meta       `path:"/oauth2/token" method:"post" tags:"LINE MAN OAuth" summary:"OAuth 认证接口" mime:"application/x-www-form-urlencoded"`
	GrantType    string `json:"grant_type" v:"required|in:client_credentials#授权类型不能为空|授权类型必须为client_credentials" dc:"OAuth 授权类型，固定值：client_credentials"`
	ClientId     string `json:"client_id" v:"required#客户端ID不能为空" dc:"LINE MAN 分配的客户端 ID"`
	ClientSecret string `json:"client_secret" v:"required#客户端密钥不能为空" dc:"LINE MAN 分配的客户端密钥"`
}

// OAuthTokenRes OAuth 令牌响应
// LINE MAN OAuth 2.0 认证成功后的响应
type OAuthTokenRes struct {
	g.Meta      `mime:"application/json"`
	AccessToken string `json:"access_token" dc:"访问令牌，用于后续 API 调用"`
	TokenType   string `json:"token_type" dc:"令牌类型，固定值：Bearer"`
	ExpiresIn   int    `json:"expires_in" dc:"令牌有效期（秒），通常为 3600"`
}

// OAuthTokenErrorRes OAuth 认证错误响应
// LINE MAN OAuth 2.0 认证失败时的响应
type OAuthTokenErrorRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}
