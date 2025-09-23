package erp

// PosPermissionRule POS权限规则结构体
type PosPermissionRule struct {
	Name        string                  `json:"name,omitempty"`         // 规则名称标识
	Owner       string                  `json:"owner,omitempty"`        // 所有者
	Creation    string                  `json:"creation,omitempty"`     // 创建时间
	Modified    string                  `json:"modified,omitempty"`     // 修改时间
	ModifiedBy  string                  `json:"modified_by,omitempty"`  // 修改者
	Docstatus   int                     `json:"docstatus,omitempty"`    // 文档状态：0-已保存，1-已提交，2-已取消
	Idx         int                     `json:"idx,omitempty"`          // 索引位置
	RuleCode    string                  `json:"rule_code,omitempty"`    // 规则代码
	RuleName    string                  `json:"rule_name,omitempty"`    // 规则名称
	RuleType    string                  `json:"rule_type,omitempty"`    // 规则类型：White-白名单，Black-黑名单
	Doctype     string                  `json:"doctype,omitempty"`      // 文档类型
	CompanyList []PermissionCompanyList `json:"company_list,omitempty"` // 公司列表
}

// PermissionCompanyList 权限公司列表结构体
type PermissionCompanyList struct {
	Name        string `json:"name,omitempty"`        // 记录名称标识
	Owner       string `json:"owner,omitempty"`       // 所有者
	Creation    string `json:"creation,omitempty"`    // 创建时间
	Modified    string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"` // 修改者
	Docstatus   int    `json:"docstatus,omitempty"`   // 文档状态：0-已保存，1-已提交，2-已取消
	Idx         int    `json:"idx,omitempty"`         // 索引位置
	Company     string `json:"company,omitempty"`     // 公司名称
	Parent      string `json:"parent,omitempty"`      // 父级文档
	Parentfield string `json:"parentfield,omitempty"` // 父级字段名
	Parenttype  string `json:"parenttype,omitempty"`  // 父级文档类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}
