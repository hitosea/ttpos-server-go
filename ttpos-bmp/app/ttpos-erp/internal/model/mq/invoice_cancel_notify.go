package mq

// InvoiceCancelNotifyMsg 发票取消通知消息
// 用于退票处理完成后通知外部 ERP 系统
type InvoiceCancelNotifyMsg struct {
	OrderNo     string `json:"order_no"`     // 订单号
	InvoiceName string `json:"invoice_name"` // 发票名称
	Remark      string `json:"remark"`       // 附注信息
}
