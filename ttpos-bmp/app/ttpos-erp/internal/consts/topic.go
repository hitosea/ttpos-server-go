package consts

type Topic string

const (
	TopicItemSync      Topic = "item-sync"
	TopicItemSyncDelay Topic = "item-sync-delay"
	//TopicDocChange Erpnext 系统文档变更
	TopicDocChange Topic = "erp-doc-change"
	//TopicItemChange 物品变更
	TopicItemChange Topic = "erp-item-change"

	TopicSavePosInvoice   = Topic("save-pos-invoice")
	TopicReturnPosInvoice = Topic("return-pos-invoice")
	TopicCancelPosInvoice = Topic("cancel-pos-invoice")
	TopicClosePosEntry    = Topic("close-pos-entry")
	//TopicRedoPos 重做发票订单
	TopicRedoPos = Topic("redo-pos")
	//TopicErpInvoiceCancel 发票取消通知（退票处理完成后发送）
	TopicErpInvoiceCancel = Topic("erp-invoice-cancel")

	// Sales Invoice (SI) 相关
	TopicSaveSalesInvoice   = Topic("save-sales-invoice")
	TopicCancelSalesInvoice = Topic("cancel-sales-invoice")
	TopicReturnSalesInvoice = Topic("return-sales-invoice")

	// SI 异步回调（BMP → Main）
	TopicErpSalesInvoiceCallback = Topic("erp-sales-invoice-callback")

	// SI 自定义 DLQ（应用层重试耗尽后投入）
	TopicSaveSalesInvoiceDLQ   = Topic("save-sales-invoice-dlq")
	TopicCancelSalesInvoiceDLQ = Topic("cancel-sales-invoice-dlq")
	TopicReturnSalesInvoiceDLQ = Topic("return-sales-invoice-dlq")
)
