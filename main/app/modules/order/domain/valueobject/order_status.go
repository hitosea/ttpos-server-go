package valueobject

// OrderStatus 订单状态值对象
type OrderStatus int

const (
	// StatusPending 待处理
	StatusPending OrderStatus = 0
	// StatusConfirmed 已确认
	StatusConfirmed OrderStatus = 1
	// StatusPreparing 制作中
	StatusPreparing OrderStatus = 2
	// StatusCompleted 已完成
	StatusCompleted OrderStatus = 3
	// StatusCancelled 已取消
	StatusCancelled OrderStatus = 4
)

// IsPending 是否待处理
func (s OrderStatus) IsPending() bool {
	return s == StatusPending
}

// IsConfirmed 是否已确认
func (s OrderStatus) IsConfirmed() bool {
	return s == StatusConfirmed
}

// IsPreparing 是否制作中
func (s OrderStatus) IsPreparing() bool {
	return s == StatusPreparing
}

// IsCompleted 是否已完成
func (s OrderStatus) IsCompleted() bool {
	return s == StatusCompleted
}

// IsCancelled 是否已取消
func (s OrderStatus) IsCancelled() bool {
	return s == StatusCancelled
}

// CanAddItem 是否可以添加商品
func (s OrderStatus) CanAddItem() bool {
	return s == StatusPending || s == StatusConfirmed
}

// CanApplyDiscount 是否可以应用优惠
func (s OrderStatus) CanApplyDiscount() bool {
	return s == StatusPending || s == StatusConfirmed
}

// CanCancel 是否可以取消
func (s OrderStatus) CanCancel() bool {
	return s != StatusCompleted && s != StatusCancelled
}

// String 转换为字符串
func (s OrderStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusConfirmed:
		return "confirmed"
	case StatusPreparing:
		return "preparing"
	case StatusCompleted:
		return "completed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// ToInt 转换为整数
func (s OrderStatus) ToInt() int {
	return int(s)
}
