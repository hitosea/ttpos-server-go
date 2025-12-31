package event

// OrderRejectedEvent 订单拒绝事件
type OrderRejectedEvent struct {
	BaseDomainEvent
	OrderUuid        uint64 // 订单UUID
	Platform         string // 平台
	PlatformOrderId  string // 平台订单ID
	ShortOrderNumber string // 短订单号
	TakeoutOrderUuid string // 外卖订单UUID
	RejectedBy       uint64 // 拒单人UUID
	RejectReasonCode string // 拒单原因代码
	RejectReason     string // 拒单原因
	CompanyUuid      uint64 // 公司UUID
}

// EventName 事件名称
func (e OrderRejectedEvent) EventName() string {
	return "takeout.order.rejected"
}

// NewOrderRejectedEvent 创建订单拒绝事件
func NewOrderRejectedEvent(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	rejectedBy uint64,
	rejectReasonCode string,
	rejectReason string,
	companyUuid uint64,
) OrderRejectedEvent {
	return OrderRejectedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		RejectedBy:       rejectedBy,
		RejectReasonCode: rejectReasonCode,
		RejectReason:     rejectReason,
		CompanyUuid:      companyUuid,
	}
}
