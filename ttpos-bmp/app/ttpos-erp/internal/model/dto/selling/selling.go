package selling

type GetPosInvoiceListReq struct {
	StartDate  string `json:"start"`
	EndDate    string `json:"end"`
	PosProfile string `json:"pos_profile"`
	User       string `json:"user,omitempty"`
	Docstatus  string `json:"docstatus,omitempty"` //文档状态
	IsReturn   string `json:"is_return,omitempty"` //是否退款

	CustomPosOpeningEntry string `json:"custom_pos_opening_entry,omitempty"` // 自定义POS开帐分录
}

// SimplePosInvoice 结构体定义
// 用于表示简化的POS发票信息，包含核心字段
type SimplePosInvoice struct {
	Name          string      `json:"name,omitempty"`           // 发票名称
	PostingDate   string      `json:"posting_date,omitempty"`   // 过账日期
	Customer      string      `json:"customer,omitempty"`       // 客户
	GrandTotal    float64     `json:"grand_total,omitempty"`    // 总金额
	IsReturn      int         `json:"is_return,omitempty"`      // 是否为退货
	ReturnAgainst interface{} `json:"return_against,omitempty"` // 退货关联
}

// ListCustomersReq 获取客户列表请求
type ListCustomersReq struct {
	Name               string `json:"name,omitempty"`                 // 客户主键
	CustomerName       string `json:"customer_name,omitempty"`        // 客户名称（模糊查询）
	CustomerType       string `json:"customer_type,omitempty"`        // 客户类型
	CustomerGroup      string `json:"customer_group,omitempty"`       // 客户组
	RepresentsCompany  string `json:"represents_company,omitempty"`   // 代表公司
	CompanyAbbr        string `json:"company_abbr,omitempty"`         // 公司缩写
	IsInternalCustomer int    `json:"is_internal_customer,omitempty"` // 是否内部客户
	Language           string `json:"language,omitempty"`             // 语言
	IsFrozen           int    `json:"is_frozen,omitempty"`            // 是否冻结
	IncludeDisabled    bool   `json:"include_disabled,omitempty"`     // 是否包含禁用的客户
	PageNo             int32  `json:"page_no,omitempty"`              // 页码
	PageSize           int32  `json:"page_size,omitempty"`            // 每页数量
}
