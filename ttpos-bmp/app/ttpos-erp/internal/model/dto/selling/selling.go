package selling

import "time"

type GetPosInvoiceListReq struct {
	StartDate  time.Time `json:"start"`
	EndDate    time.Time `json:"end"`
	PosProfile string    `json:"pos_profile"`
	User       string    `json:"user,omitempty"`
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
