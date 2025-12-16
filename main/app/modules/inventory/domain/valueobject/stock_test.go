package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewStock 测试创建库存
func TestNewStock(t *testing.T) {
	stock := NewStock(100.5)
	assert.Equal(t, 100.5, stock.Value())
}

// TestNewStock_Negative 测试负数转换为0
func TestNewStock_Negative(t *testing.T) {
	stock := NewStock(-50.0)
	assert.Equal(t, 0.0, stock.Value())
}

// TestNewStock_Precision 测试精度（保留两位小数）
func TestNewStock_Precision(t *testing.T) {
	stock := NewStock(100.555)
	assert.Equal(t, 100.56, stock.Value()) // 四舍五入

	stock = NewStock(100.554)
	assert.Equal(t, 100.55, stock.Value())

	stock = NewStock(100.123456)
	assert.Equal(t, 100.12, stock.Value())
}

// TestStock_IsZero 测试是否为零
func TestStock_IsZero(t *testing.T) {
	stock := NewStock(0)
	assert.True(t, stock.IsZero())

	stock = NewStock(0.01)
	assert.False(t, stock.IsZero())
}

// TestStock_IsPositive 测试是否为正数
func TestStock_IsPositive(t *testing.T) {
	stock := NewStock(10.0)
	assert.True(t, stock.IsPositive())

	stock = NewStock(0)
	assert.False(t, stock.IsPositive())
}

// TestStock_Add 测试加法
func TestStock_Add(t *testing.T) {
	stock := NewStock(100.0)
	newStock := stock.Add(50.5)

	assert.Equal(t, 100.0, stock.Value()) // 原值不变（不可变）
	assert.Equal(t, 150.5, newStock.Value())
}

// TestStock_Subtract 测试减法
func TestStock_Subtract(t *testing.T) {
	stock := NewStock(100.0)
	newStock := stock.Subtract(30.5)

	assert.Equal(t, 100.0, stock.Value()) // 原值不变（不可变）
	assert.Equal(t, 69.5, newStock.Value())
}

// TestStock_Subtract_Negative 测试减法结果为负数
func TestStock_Subtract_Negative(t *testing.T) {
	stock := NewStock(50.0)
	newStock := stock.Subtract(60.0)

	assert.Equal(t, 0.0, newStock.Value()) // 负数转为0
}

// TestStock_Equals 测试相等比较
func TestStock_Equals(t *testing.T) {
	stock1 := NewStock(100.0)
	stock2 := NewStock(100.0)
	stock3 := NewStock(99.99)

	assert.True(t, stock1.Equals(stock2))
	assert.False(t, stock1.Equals(stock3))
}

// TestStock_GreaterThan 测试大于比较
func TestStock_GreaterThan(t *testing.T) {
	stock1 := NewStock(100.0)
	stock2 := NewStock(50.0)

	assert.True(t, stock1.GreaterThan(stock2))
	assert.False(t, stock2.GreaterThan(stock1))
	assert.False(t, stock1.GreaterThan(stock1))
}

// TestStock_LessThan 测试小于比较
func TestStock_LessThan(t *testing.T) {
	stock1 := NewStock(50.0)
	stock2 := NewStock(100.0)

	assert.True(t, stock1.LessThan(stock2))
	assert.False(t, stock2.LessThan(stock1))
	assert.False(t, stock1.LessThan(stock1))
}

// TestStock_GreaterThanOrEqual 测试大于等于比较
func TestStock_GreaterThanOrEqual(t *testing.T) {
	stock1 := NewStock(100.0)
	stock2 := NewStock(100.0)
	stock3 := NewStock(50.0)

	assert.True(t, stock1.GreaterThanOrEqual(stock2))
	assert.True(t, stock1.GreaterThanOrEqual(stock3))
	assert.False(t, stock3.GreaterThanOrEqual(stock1))
}

// TestStock_LessThanOrEqual 测试小于等于比较
func TestStock_LessThanOrEqual(t *testing.T) {
	stock1 := NewStock(100.0)
	stock2 := NewStock(100.0)
	stock3 := NewStock(150.0)

	assert.True(t, stock1.LessThanOrEqual(stock2))
	assert.True(t, stock1.LessThanOrEqual(stock3))
	assert.False(t, stock3.LessThanOrEqual(stock1))
}

// TestStock_Immutability 测试不可变性
func TestStock_Immutability(t *testing.T) {
	original := NewStock(100.0)

	// 所有操作都应该返回新对象，不修改原对象
	_ = original.Add(50.0)
	assert.Equal(t, 100.0, original.Value())

	_ = original.Subtract(30.0)
	assert.Equal(t, 100.0, original.Value())
}

// TestStock_ChainOperations 测试链式操作
func TestStock_ChainOperations(t *testing.T) {
	stock := NewStock(100.0)
	result := stock.Add(50.0).Subtract(30.0).Add(10.0)

	assert.Equal(t, 100.0, stock.Value())  // 原值不变
	assert.Equal(t, 130.0, result.Value()) // (100+50-30+10)
}
