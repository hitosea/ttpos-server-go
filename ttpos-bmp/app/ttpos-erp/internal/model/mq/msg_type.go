package mq

type MsgTyp string

const (
	MsgTypeSavePosInvoice   MsgTyp = "save-pos-invoice"
	MsgTypeReturnPosInvoice MsgTyp = "return-pos-invoice"
	MsgTypeCancelPosInvoice MsgTyp = "cancel-pos-invoice"
	MsgTypeClosePosEntry    MsgTyp = "close-pos-entry"

	// Sales Invoice (SI) 相关
	MsgTypeSaveSalesInvoice   MsgTyp = "save-sales-invoice"
	MsgTypeCancelSalesInvoice MsgTyp = "cancel-sales-invoice"
	MsgTypeReturnSalesInvoice MsgTyp = "return-sales-invoice"
)
