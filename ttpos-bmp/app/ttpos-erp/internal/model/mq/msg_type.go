package mq

type MsgTyp string

const (
	MsgTypeSavePosInvoice   MsgTyp = "save-pos-invoice"
	MsgTypeReturnPosInvoice MsgTyp = "return-pos-invoice"
	MsgTypeCancelPosInvoice MsgTyp = "cancel-pos-invoice"
	MsgTypeClosePosEntry    MsgTyp = "close-pos-entry"
)
