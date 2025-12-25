package event

// OrderStatusUpdatedEvent 订单状态更新事件（通用状态更新）
type OrderStatusUpdatedEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	OldOrderState    int    // 旧订单状态
	NewOrderState    int    // 新订单状态
	PlatformState    string // 平台原始状态
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderStatusUpdatedEvent) EventName() string {
	return "takeout.order.status_updated"
}

// NewOrderStatusUpdatedEvent 创建订单状态更新事件
func NewOrderStatusUpdatedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	oldOrderState int,
	newOrderState int,
	platformState string,
	companyUuid uint64,
) OrderStatusUpdatedEvent {
	return OrderStatusUpdatedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		OldOrderState:    oldOrderState,
		NewOrderState:    newOrderState,
		PlatformState:    platformState,
		CompanyUuid:      companyUuid,
	}
}
