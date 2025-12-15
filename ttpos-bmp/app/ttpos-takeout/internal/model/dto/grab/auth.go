package grab

// ============================================================================
// Authentication
// ============================================================================

// OAuthTokenRequest OAuth 令牌请求
type OAuthTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"` // client_credentials
	Scope        string `json:"scope"`      // grab_food.partner
}

// OAuthTokenResponse OAuth 令牌响应
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"` // Bearer
	ExpiresIn   int    `json:"expires_in"` // 秒
}

// WebhookHeaders Webhook 请求头
type WebhookHeaders struct {
	ContentType    string `json:"Content-Type"`     // application/json
	Authorization  string `json:"Authorization"`    // Bearer {token}
	XGrabSignature string `json:"X-Grab-Signature"` // HMAC-SHA256 签名
	XGrabTimestamp string `json:"X-Grab-Timestamp"` // 时间戳
}
