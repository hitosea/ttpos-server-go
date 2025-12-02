package erp

// PrintFormatListReq Print Format 列表查询请求
type PrintFormatListReq struct {
	DocType    string `json:"doc_type,omitempty"`    // DocType 过滤
	Limit      int    `json:"limit,omitempty"`       // 分页大小
	LimitStart int    `json:"limit_start,omitempty"` // 分页起始位置
}

// PrintFormatCreateUpdateReq Print Format 创建/更新请求
type PrintFormatCreateUpdateReq struct {
	Name            string `json:"name"`                        // Print Format 名称（必填）
	DocType         string `json:"doc_type"`                    // 关联的 DocType（必填）
	Standard        int    `json:"standard,omitempty"`          // 是否标准格式：0-否，1-是
	Module          string `json:"module,omitempty"`            // 模块名称
	HTML            string `json:"html,omitempty"`              // HTML 模板内容
	CSS             string `json:"css,omitempty"`               // CSS 样式
	PrintFormat     string `json:"print_format,omitempty"`      // 打印格式类型：Standard, Jinja, JS
	PrintFormatType string `json:"print_format_type,omitempty"` // 格式类型
	PrintFormatFor  string `json:"print_format_for,omitempty"`  // Print Format For
	CustomFormat    int    `json:"custom_format,omitempty"`     // 是否自定义格式：0-否，1-是
}

// PrintFormatDetailResp Print Format 详情响应
type PrintFormatDetailResp struct {
	Name            string `json:"name,omitempty"`              // Print Format 名称
	DocType         string `json:"doc_type,omitempty"`          // 关联的 DocType
	Standard        int    `json:"standard,omitempty"`          // 是否标准格式：0-否，1-是
	Module          string `json:"module,omitempty"`            // 模块名称
	HTML            string `json:"html,omitempty"`              // HTML 模板内容
	CSS             string `json:"css,omitempty"`               // CSS 样式
	PrintFormat     string `json:"print_format,omitempty"`      // 打印格式类型：Standard, Jinja, JS
	PrintFormatType string `json:"print_format_type,omitempty"` // 格式类型
	PrintFormatFor  string `json:"print_format_for,omitempty"`  // Print Format For
	CustomFormat    int    `json:"custom_format,omitempty"`     // 是否自定义格式：0-否，1-是
	// ERPNext 标准字段
	Owner      string `json:"owner,omitempty"`       // 拥有者
	Creation   string `json:"creation,omitempty"`    // 创建时间
	Modified   string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"` // 修改者
	Docstatus  int    `json:"docstatus,omitempty"`   // 文档状态
	Doctype    string `json:"doctype,omitempty"`     // 单据类型
}

// PrintFormatListResp Print Format 列表响应
type PrintFormatListResp struct {
	List  []*PrintFormatDetailResp `json:"list,omitempty"`  // Print Format 列表
	Total int                      `json:"total,omitempty"` // 总数
}
