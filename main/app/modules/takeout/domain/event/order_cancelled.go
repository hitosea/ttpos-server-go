package event

// OrderCancelEvent 订单取消事件
type OrderCancelEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	CompanyUuid      uint64 // 公司UUID
	CancelReason     string // 取消原因
}

// EventName 事件名称
func (e OrderCancelEvent) EventName() string {
	return "takeout.order.cancelled"
}

// NewOrderCancelEvent 创建订单取消事件
func NewOrderCancelEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
	cancelReason string,
) OrderCancelEvent {
	return OrderCancelEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
		CancelReason:     cancelReason,
	}
}
