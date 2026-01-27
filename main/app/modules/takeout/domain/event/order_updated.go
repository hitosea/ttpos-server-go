package event

// OrderUpdatedEvent 订单更新事件
type OrderUpdatedEvent struct {
	BaseDomainEvent
	OrderUuid        uint64  // 订单UUID
	Platform         string  // 平台
	PlatformOrderId  string  // 平台订单ID
	ShortOrderNumber string  // 短订单号
	TakeoutOrderUuid string  // 外卖订单UUID
	EaterPayment     float64 // 顾客实付(分)
	CompanyUuid      uint64  // 公司UUID
}

// EventName 事件名称
func (e OrderUpdatedEvent) EventName() string {
	return "takeout.order.updated"
}

// NewOrderUpdatedEvent 创建订单更新事件
func NewOrderUpdatedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
) OrderUpdatedEvent {
	return OrderUpdatedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
	}
}
