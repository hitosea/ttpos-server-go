package request

// ImportMenuRequest 导入 Grab 菜单请求
type ImportMenuRequest struct {
	Platform string      `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	MenuData interface{} `json:"menuData" binding:"required"` // 平台菜单 JSON 数据
}

// ExportMenuRequest 导出菜单请求
type ExportMenuRequest struct {
	Platform       string   // 平台名称：grab, lineman 等
	CompanyUuid    uint64   // 公司 UUID
	CurrencyUnit   string   // 货币单位
	CategoryIDs    []uint64 // 分类 ID 列表（可选）
	SellingTimeIDs []uint64 // 售卖时段 ID 列表（可选）
}
