package response

// BindingLinkResponse 绑定链接响应
type BindingLinkResponse struct {
	BindingLink string `json:"bindingLink"` // 绑定链接 URL
	ExpiresAt   int64  `json:"expiresAt"`   // 过期时间（Unix 时间戳）
}

// BindingStatusResponse 绑定状态响应
type BindingStatusResponse struct {
	IsBound      bool   `json:"isBound"`      // 是否已绑定
	BoundAt      int64  `json:"boundAt"`      // 绑定时间（Unix 时间戳）
	MerchantID   string `json:"merchantId"`   // Grab 商户 ID
	MerchantName string `json:"merchantName"` // Grab 商户名称
}

// GrabMenuResponse Grab 菜单响应
type GrabMenuResponse struct {
	Menu interface{} `json:"menu"` // Grab 菜单数据
}
