package event

// OrderReadyEvent 订单准备完成事件（呼叫骑手）
type OrderReadyEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderReadyEvent) EventName() string {
	return "takeout.order.ready"
}

// NewOrderReadyEvent 创建订单准备完成事件
func NewOrderReadyEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
) OrderReadyEvent {
	return OrderReadyEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
	}
}
