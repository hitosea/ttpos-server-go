package mq

type AsyncSellingMsg struct {
	RecordId         int64  `json:"record_id,omitempty"`           //记录ID
	MsgType          MsgTyp `json:"msg_type"`                      //消息类型
	PosOpenEntryName string `json:"pos_open_entry_name,omitempty"` //开单名称
	SiteCode         string `json:"site_code,omitempty"`           //站点编码
}

// AsyncSalesInvoiceMsg Sales Invoice 异步消息
type AsyncSalesInvoiceMsg struct {
	RecordId int64  `json:"record_id,omitempty"` // receive_sales_invoice 记录ID
	MsgType  MsgTyp `json:"msg_type"`            // 消息类型
	SiteCode string `json:"site_code,omitempty"` // 站点编码
}
