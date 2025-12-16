package event

// DefaultWarehouseChangedEvent 默认仓库变更事件
type DefaultWarehouseChangedEvent struct {
	BaseDomainEvent
	PreviousDefaultUuid uint64 // 之前的默认仓库UUID（0表示之前没有默认仓库）
	NewDefaultUuid      uint64 // 新的默认仓库UUID
}

// EventName 事件名称
func (e DefaultWarehouseChangedEvent) EventName() string {
	return "warehouse.default_changed"
}

// NewDefaultWarehouseChangedEvent 创建默认仓库变更事件
func NewDefaultWarehouseChangedEvent(previousDefaultUuid, newDefaultUuid uint64) DefaultWarehouseChangedEvent {
	return DefaultWarehouseChangedEvent{
		BaseDomainEvent:     NewBaseDomainEvent(newDefaultUuid),
		PreviousDefaultUuid: previousDefaultUuid,
		NewDefaultUuid:      newDefaultUuid,
	}
}
