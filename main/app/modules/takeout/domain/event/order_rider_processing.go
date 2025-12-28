package event

// OrderRiderProcessingEvent 骑手配送中事件
type OrderRiderProcessingEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderRiderProcessingEvent) EventName() string {
	return "takeout.order.rider.processing"
}

// NewOrderRiderProcessingEvent 创建骑手配送中事件
func NewOrderRiderProcessingEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
) OrderRiderProcessingEvent {
	return OrderRiderProcessingEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
	}
}
