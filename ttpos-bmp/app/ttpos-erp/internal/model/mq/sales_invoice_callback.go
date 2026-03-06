package mq

// SalesInvoiceCallbackMsg SI 异步回调消息（BMP → Main）
// consumer 创建 SI+PE 完成后，发送此消息通知 Main 更新 sale_order
type SalesInvoiceCallbackMsg struct {
	CompanyUuid       string `json:"company_uuid"`        // 商户UUID（用于路由到 shop 数据库）
	SaleOrderUuid     string `json:"sale_order_uuid"`     // 订单UUID
	SyncStatus        int    `json:"sync_status"`         // 同步状态: 3=成功 4=失败
	SalesInvoiceName  string `json:"sales_invoice_name"`  // SI 单据名称
	PaymentEntryNames string `json:"payment_entry_names"` // PE 名称列表(JSON)
	ErrorMsg          string `json:"error_msg,omitempty"` // 失败原因
}
