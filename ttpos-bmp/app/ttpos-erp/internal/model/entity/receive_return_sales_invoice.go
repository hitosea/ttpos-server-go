package entity

// ReceiveReturnSalesInvoice is the golang structure for table erp_receive_return_sales_invoice.
type ReceiveReturnSalesInvoice struct {
	Id                int64  `json:"id"                orm:"id"                  description:"ID"`
	OrderNo           string `json:"orderNo"           orm:"order_no"            description:"TTPOS订单号"`
	SaleOrderUuid     string `json:"saleOrderUuid"     orm:"sale_order_uuid"     description:"TTPOS订单UUID"`
	PosProfile        string `json:"posProfile"        orm:"pos_profile"         description:"POS Profile名称"`
	PostingDatetime   int64  `json:"postingDatetime"   orm:"posting_datetime"    description:"过账时间戳"`
	Docstatus         string `json:"docstatus"         orm:"docstatus"           description:"文档状态: 0=Draft 1=Submitted 2=Cancelled"`
	SalesInvoiceName  string `json:"salesInvoiceName"  orm:"sales_invoice_name"  description:"Credit Note名称"`
	PaymentEntryNames string `json:"paymentEntryNames" orm:"payment_entry_names" description:"Refund Payment Entry名称(JSON)"`
	SiteCode          string `json:"siteCode"          orm:"site_code"           description:"ERP site code"`
	ReqMessage        string `json:"reqMessage"        orm:"req_message"         description:"请求数据(base64)"`
	RespMessage       string `json:"respMessage"       orm:"resp_message"        description:"响应数据(base64)"`
	ReqBody           string `json:"reqBody"           orm:"req_body"            description:"请求文本"`
	RespBody          string `json:"respBody"          orm:"resp_body"           description:"响应文本"`
	RetryCount        int    `json:"retryCount"        orm:"retry_count"         description:"重试次数"`
	CreatedAt         int    `json:"createdAt"         orm:"created_at"          description:"创建时间"`
	UpdatedAt         int    `json:"updatedAt"         orm:"updated_at"          description:"更新时间"`
}
