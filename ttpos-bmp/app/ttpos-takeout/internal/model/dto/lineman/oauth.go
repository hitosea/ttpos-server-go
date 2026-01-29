package lineman

// ============================================================================
// OAuth Authentication
// ============================================================================

// LinemanOAuthTokenRequest LINE MAN OAuth 令牌请求
type LinemanOAuthTokenRequest struct {
	ClientID     string `json:"client_id"`     // LINE MAN 分配的 client_id
	ClientSecret string `json:"client_secret"` // LINE MAN 分配的 client_secret
	GrantType    string `json:"grant_type"`    // 固定值: client_credentials
}

// LinemanOAuthTokenResponse LINE MAN OAuth 令牌响应
type LinemanOAuthTokenResponse struct {
	AccessToken string `json:"access_token"` // OAuth Access Token
	TokenType   string `json:"token_type"`   // Token 类型，通常为 Bearer
	ExpiresIn   int    `json:"expires_in"`   // Token 有效期（秒）
}
