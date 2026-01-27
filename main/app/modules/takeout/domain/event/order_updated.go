package event

import (
	"ttpos-server-go/app/modules/takeout/domain/value_object"
)

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

	// 订单变动信息（用于打印和库存处理）
	ChangeResult *value_object.OrderChangeResult // 变动结果，nil 表示无菜品变动
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
		ChangeResult:     nil,
	}
}

// NewOrderUpdatedEventWithChange 创建带变动信息的订单更新事件
func NewOrderUpdatedEventWithChange(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	shortOrderNumber string,
	takeoutOrderUuid string,
	companyUuid uint64,
	changeResult *value_object.OrderChangeResult,
) OrderUpdatedEvent {
	return OrderUpdatedEvent{
		BaseDomainEvent:  NewBaseDomainEvent(orderUuid),
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		CompanyUuid:      companyUuid,
		ChangeResult:     changeResult,
	}
}

// HasItemChange 是否有菜品变动
func (e OrderUpdatedEvent) HasItemChange() bool {
	return e.ChangeResult != nil && e.ChangeResult.HasChange
}
