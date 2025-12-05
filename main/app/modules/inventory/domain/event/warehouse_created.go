package event

// WarehouseCreatedEvent 仓库创建事件
type WarehouseCreatedEvent struct {
	BaseDomainEvent
	WarehouseUuid uint64 // 仓库UUID
	Code          string // 仓库编码
	Name          string // 仓库名称
	Type          string // 仓库类型
}

// EventName 事件名称
func (e WarehouseCreatedEvent) EventName() string {
	return "warehouse.created"
}

// NewWarehouseCreatedEvent 创建仓库创建事件
func NewWarehouseCreatedEvent(warehouseUuid uint64, code, name, warehouseType string) WarehouseCreatedEvent {
	return WarehouseCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(warehouseUuid),
		WarehouseUuid:   warehouseUuid,
		Code:            code,
		Name:            name,
		Type:            warehouseType,
	}
}
