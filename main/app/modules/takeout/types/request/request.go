package request

// ImportMenuRequest 导入 Grab 菜单请求
type ImportMenuRequest struct {
	Platform string      `json:"platform" binding:"required,oneof=grab lineman hamburger"` // 平台名称：grab, lineman, hamburger 等
	MenuData interface{} `json:"menuData" binding:"required"`                              // 平台菜单 JSON 数据
}

// ExportMenuRequest 导出菜单请求
type ExportMenuRequest struct {
	Platform     string `json:"platform" binding:"required"`     // 平台名称：grab, lineman 等
	CompanyUuid  uint64 `json:"company_uuid" binding:"required"` // 公司 UUID
	CurrencyUnit string `json:"currency_unit"`                   // 货币单位
}

// ToggleTakeoutStatusRequest 切换外卖状态请求
type ToggleTakeoutStatusRequest struct {
	Platform string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	Enabled  bool   `json:"enabled" binding:"required"`  // 是否开启外卖
}

// UpdateBindingStatusRequest 更新绑定状态请求
type UpdateBindingStatusRequest struct {
	Platform string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
}

// GetImportLogsRequest 获取导入日志列表请求
type GetImportLogsRequest struct {
	Platform   string // 外卖平台筛选
	ImportType *int8  // 导入类型筛选（1-TTPOS推送到平台 2-平台推送到TTPOS）
	Status     *int8  // 状态筛选（0-进行中 1-成功 2-失败）
	PageNo     int    // 页码
	PageSize   int    // 每页数量
}
