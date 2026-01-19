package event

// WarehouseDeletedEvent 仓库删除事件
type WarehouseDeletedEvent struct {
	BaseDomainEvent
	WarehouseUuid uint64 // 仓库UUID
	Code          string // 仓库编码
}

// EventName 事件名称
func (e WarehouseDeletedEvent) EventName() string {
	return "warehouse.deleted"
}

// NewWarehouseDeletedEvent 创建仓库删除事件
func NewWarehouseDeletedEvent(warehouseUuid uint64, code string) WarehouseDeletedEvent {
	return WarehouseDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(warehouseUuid),
		WarehouseUuid:   warehouseUuid,
		Code:            code,
	}
}
