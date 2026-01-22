package event

// OrderPeakTimeRecordEvent 订单高峰期记录事件（批量）
type OrderPeakTimeRecordEvent struct {
	BaseDomainEvent
	OrderUuids  []uint64 // 订单UUID列表
	CompanyUuid uint64   // 公司UUID
}

// EventName 事件名称
func (e OrderPeakTimeRecordEvent) EventName() string {
	return "takeout.order.peak_time.record"
}

// NewOrderPeakTimeRecordEvent 创建订单高峰期记录事件（批量）
func NewOrderPeakTimeRecordEvent(
	orderUuids []uint64,
	companyUuid uint64,
) OrderPeakTimeRecordEvent {
	// 使用第一个订单UUID作为聚合根ID（如果列表不为空）
	aggregateID := uint64(0)
	if len(orderUuids) > 0 {
		aggregateID = orderUuids[0]
	}
	return OrderPeakTimeRecordEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		OrderUuids:      orderUuids,
		CompanyUuid:     companyUuid,
	}
}
