package event

// OrderAcceptedEvent 订单接受事件
type OrderAcceptedEvent struct {
	BaseDomainEvent
	OrderUuid         uint64 // 订单UUID
	Platform          string // 平台
	PlatformOrderId   string // 平台订单ID
	ShortOrderNumber  string // 短订单号
	TakeoutOrderUuid  string // 外卖订单UUID
	AcceptedBy        uint64 // 接单人UUID
	OrderAcceptedType string // 接单类型（手动/自动）
	CompanyUuid       uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderAcceptedEvent) EventName() string {
	return "takeout.order.accepted"
}

// NewOrderAcceptedEvent 创建订单接受事件
func NewOrderAcceptedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	acceptedBy uint64,
	orderAcceptedType string,
	companyUuid uint64,
) OrderAcceptedEvent {
	return OrderAcceptedEvent{
		BaseDomainEvent:   NewBaseDomainEvent(orderUuid),
		OrderUuid:         orderUuid,
		Platform:          platform,
		PlatformOrderId:   platformOrderId,
		ShortOrderNumber:  shortOrderNumber,
		TakeoutOrderUuid:  takeoutOrderUuid,
		AcceptedBy:        acceptedBy,
		OrderAcceptedType: orderAcceptedType,
		CompanyUuid:       companyUuid,
	}
}
