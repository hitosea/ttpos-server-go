package lineman

// MenuStatusUpdateReq Lineman 菜单状态更新请求
type MenuStatusUpdateReq struct {
	MenuItems []MenuItemStatus `json:"menuItems"`
}

// MenuItemStatus 菜单项状态（用于状态更新 API）
type MenuItemStatus struct {
	ID         string `json:"id"`         // Partner Item ID
	MenuStatus string `json:"menuStatus"` // AVAILABLE, SUSPENDED, SOLD_OUT_TODAY
}

// MenuStatusUpdateResp Lineman 菜单状态更新响应
type MenuStatusUpdateResp struct {
	Status  string `json:"status"`  // ok, fail
	Code    string `json:"code"`    // SUCCESS, ERROR
	Message string `json:"message"` // 响应消息
}
