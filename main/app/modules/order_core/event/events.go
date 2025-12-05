package event

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	BillUuid  uint64 `json:"bill_uuid"`
	OrderUuid uint64 `json:"order_uuid"`
}

// OrderPaidEvent 订单支付事件
type OrderPaidEvent struct {
	BillUuid uint64 `json:"bill_uuid"`
	PayTime  int64  `json:"pay_time"`
}
