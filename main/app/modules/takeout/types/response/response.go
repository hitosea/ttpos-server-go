package response

// BindingLinkResponse 绑定链接响应
type BindingLinkResponse struct {
	BindingLink string `json:"binding_link"` // 绑定链接 URL
}

// BindingStatusResponse 绑定状态响应
type BindingStatusResponse struct {
	IsBound bool `json:"is_bound"` // 是否已绑定
}

// GrabMenuResponse Grab 菜单响应
type GrabMenuResponse struct {
	Platform string      `json:"platform"` // 平台名称
	Menu     interface{} `json:"menu"`     // Grab 菜单数据
}

// TakeoutStatusResponse 外卖状态响应
type TakeoutStatusResponse struct {
	Platform  string `json:"platform"`   // 外卖平台
	Enabled   bool   `json:"enabled"`    // 是否开启
	IsBound   bool   `json:"is_bound"`   // 是否已绑定
	Skip      bool   `json:"skip"`       // 是否跳过绑定
	UpdatedAt int64  `json:"updated_at"` // 更新时间
}

// TakeoutStatusListResponse 外卖状态列表响应
type TakeoutStatusListResponse struct {
	List []*TakeoutStatusResponse `json:"list"` // 平台状态列表
}
