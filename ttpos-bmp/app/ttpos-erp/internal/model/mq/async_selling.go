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

// SalesInvoiceDLQMsg Sales Invoice DLQ 消息（应用层重试耗尽后投入）
type SalesInvoiceDLQMsg struct {
	OriginalMsg  AsyncSalesInvoiceMsg `json:"original_msg"`            // 原始消息
	RetryCount   int                  `json:"retry_count"`             // 已重试次数
	LastError    string               `json:"last_error"`              // 最后一次错误信息
	LastMqMsgId  string               `json:"last_mq_msg_id"`          // 最后一次 MQ 消息 ID
	FailedAt     int64                `json:"failed_at"`               // 失败时间戳（Unix秒）
	CompanyUuid  string               `json:"company_uuid,omitempty"`  // 商户UUID（用于快速筛选）
	OrderType    string               `json:"order_type,omitempty"`    // 订单类型（sale_order/recharge/takeout）
	OrderUuid    string               `json:"order_uuid,omitempty"`    // 订单UUID（用于追溯）
}
