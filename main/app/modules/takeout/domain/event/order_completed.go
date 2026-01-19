package event

// OrderCompletedEvent 订单完成事件
type OrderCompletedEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderCompletedEvent) EventName() string {
	return "takeout.order.completed"
}

// NewOrderCompletedEvent 创建订单完成事件
func NewOrderCompletedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
) OrderCompletedEvent {
	return OrderCompletedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
	}
}
