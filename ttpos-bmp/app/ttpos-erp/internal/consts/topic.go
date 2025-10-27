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
)
