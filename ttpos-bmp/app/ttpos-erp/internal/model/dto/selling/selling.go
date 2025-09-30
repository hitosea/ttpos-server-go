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
