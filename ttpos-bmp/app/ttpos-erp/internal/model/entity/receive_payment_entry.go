package entity

// ReceivePaymentEntry is the golang structure for table receive_payment_entry.
type ReceivePaymentEntry struct {
	Id               int64   `json:"id"               orm:"id"                description:"ID"`
	SaleOrderUuid    string  `json:"saleOrderUuid"    orm:"sale_order_uuid"   description:"TTPOS订单UUID"`
	ModeOfPayment    string  `json:"modeOfPayment"    orm:"mode_of_payment"   description:"支付方式"`
	PaymentEntryName string  `json:"paymentEntryName" orm:"payment_entry_name" description:"ERP Payment Entry名称"`
	Docstatus        string  `json:"docstatus"        orm:"docstatus"         description:"文档状态"`
	PaidAmount       float64 `json:"paidAmount"       orm:"paid_amount"       description:"支付金额"`
	SiteCode         string  `json:"siteCode"         orm:"site_code"         description:"ERP site code"`
	ReqBody          string  `json:"reqBody"          orm:"req_body"          description:"请求文本"`
	RespBody         string  `json:"respBody"         orm:"resp_body"         description:"响应文本"`
	RetryCount       int     `json:"retryCount"       orm:"retry_count"       description:"重试次数"`
	CreatedAt        int     `json:"createdAt"        orm:"created_at"        description:"创建时间"`
	UpdatedAt        int     `json:"updatedAt"        orm:"updated_at"        description:"更新时间"`
}
