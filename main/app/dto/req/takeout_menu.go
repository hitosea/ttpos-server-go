package req

// TakeoutMenuExportReq 外卖菜单导出请求
type TakeoutMenuExportReq struct {
	Platform    string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	CompanyUuid uint64 `json:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
}

// TakeoutMenuImportReq 外卖菜单导入请求
type TakeoutMenuImportReq struct {
	Platform string      `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	MenuData interface{} `json:"menuData" binding:"required"` // 平台菜单 JSON 数据
}
