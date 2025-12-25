package event

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	EaterPayment     int64  // 顾客实付(分)
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderCreatedEvent) EventName() string {
	return "takeout.order.created"
}

// NewOrderCreatedEvent 创建订单创建事件
func NewOrderCreatedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	eaterPayment int64,
	companyUuid uint64,
) OrderCreatedEvent {
	return OrderCreatedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		EaterPayment:     eaterPayment,
		CompanyUuid:      companyUuid,
	}
}
