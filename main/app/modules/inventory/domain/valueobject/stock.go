package valueobject

import (
	"math"
)

// Stock 库存数量值对象（不可变）
type Stock struct {
	value float64
}

// NewStock 创建库存数量
// 负数会被转换为0
func NewStock(value float64) Stock {
	if value < 0 {
		return Stock{value: 0}
	}
	// 保留两位小数
	return Stock{value: math.Round(value*100) / 100}
}

// Value 获取库存数量值
func (s Stock) Value() float64 {
	return s.value
}

// IsZero 是否为零库存
func (s Stock) IsZero() bool {
	return s.value == 0
}

// IsPositive 是否为正库存
func (s Stock) IsPositive() bool {
	return s.value > 0
}

// Add 增加库存（返回新的Stock）
func (s Stock) Add(quantity float64) Stock {
	return NewStock(s.value + quantity)
}

// Subtract 减少库存（返回新的Stock）
func (s Stock) Subtract(quantity float64) Stock {
	return NewStock(s.value - quantity)
}

// Equals 比较两个库存是否相等
func (s Stock) Equals(other Stock) bool {
	return s.value == other.value
}

// GreaterThan 是否大于另一个库存
func (s Stock) GreaterThan(other Stock) bool {
	return s.value > other.value
}

// LessThan 是否小于另一个库存
func (s Stock) LessThan(other Stock) bool {
	return s.value < other.value
}

// GreaterThanOrEqual 是否大于等于另一个库存
func (s Stock) GreaterThanOrEqual(other Stock) bool {
	return s.value >= other.value
}

// LessThanOrEqual 是否小于等于另一个库存
func (s Stock) LessThanOrEqual(other Stock) bool {
	return s.value <= other.value
}

// String 返回字符串表示
func (s Stock) String() string {
	return formatFloat(s.value)
}

// formatFloat 格式化浮点数
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune(int64(f)))
	}
	// 使用 %g 格式化，去除尾部的0
	return string([]byte{})
}

