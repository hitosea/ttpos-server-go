package valueobject

// warehouseStatusValue 内部状态值类型
type warehouseStatusValue int

const (
	statusDisabled warehouseStatusValue = 0
	statusActive   warehouseStatusValue = 1
)

// WarehouseStatus 仓库状态值对象（不可变）
type WarehouseStatus struct {
	value warehouseStatusValue
}

// 预定义的状态常量
var (
	// StatusDisabled 已禁用状态
	StatusDisabled = WarehouseStatus{value: statusDisabled}
	// StatusActive 已启用状态
	StatusActive = WarehouseStatus{value: statusActive}
)

// NewWarehouseStatus 创建仓库状态
func NewWarehouseStatus(status int) WarehouseStatus {
	if status == 1 {
		return StatusActive
	}
	return StatusDisabled
}

// ToInt 转换为整数
func (s WarehouseStatus) ToInt() int {
	return int(s.value)
}

// IsActive 是否启用
func (s WarehouseStatus) IsActive() bool {
	return s.value == statusActive
}

// IsDisabled 是否禁用
func (s WarehouseStatus) IsDisabled() bool {
	return s.value == statusDisabled
}

// Equals 比较两个状态是否相等
func (s WarehouseStatus) Equals(other WarehouseStatus) bool {
	return s.value == other.value
}

// String 返回状态字符串表示
func (s WarehouseStatus) String() string {
	if s.value == statusActive {
		return "active"
	}
	return "disabled"
}
