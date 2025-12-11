package resp

// TakeoutMenuExportResp 外卖菜单导出响应
type TakeoutMenuExportResp struct {
	Platform string      `json:"platform"` // 平台名称
	MenuData interface{} `json:"menuData"` // 平台格式的菜单数据
}

// GrabProductImportFailure 失败明细
type GrabProductImportFailure struct {
	GrabProductId string `json:"grabProductId"`
	Message       string `json:"message"`
}

// GrabProductImportResp Grab 商品导入响应
type GrabProductImportResp struct {
	SuccessCount int                        `json:"successCount"`
	FailureCount int                        `json:"failureCount"`
	CreatedItems int                        `json:"createdItems"`
	UpdatedItems int                        `json:"updatedItems"`
	Failures     []GrabProductImportFailure `json:"failures"`
}

// GrabMenuImportResp Grab 菜单导入响应（兼容菜单导入返回结构）
type GrabMenuImportResp = GrabProductImportResp

// GrabBindingLinkResp 获取 Grab 绑定链接响应
type GrabBindingLinkResp struct {
	BindingLink string `json:"bindingLink"` // 绑定链接 URL
	ExpiresAt   int64  `json:"expiresAt"`   // 过期时间（Unix 时间戳）
}

// GrabBindingStatusResp 检查 Grab 绑定状态响应
type GrabBindingStatusResp struct {
	IsBound      bool   `json:"isBound"`      // 是否已绑定
	BoundAt      int64  `json:"boundAt"`      // 绑定时间（Unix 时间戳）
	MerchantID   string `json:"merchantId"`   // Grab 商户 ID
	MerchantName string `json:"merchantName"` // Grab 商户名称
}

// GrabMenuResp 获取 Grab 菜单响应
type GrabMenuResp struct {
	Menu interface{} `json:"menu"` // Grab 菜单数据
}
