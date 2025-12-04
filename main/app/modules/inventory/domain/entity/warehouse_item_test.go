package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewWarehouseItem 测试创建新的库存物品
func TestNewWarehouseItem(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)

	assert.NotNil(t, item)
	assert.Equal(t, uint64(1), item.WarehouseUuid())
	assert.Equal(t, uint64(100), item.MaterialUuid())
	assert.Equal(t, "MAT001", item.MaterialCode())
	assert.Equal(t, 10.5, item.Valuation())
	assert.Equal(t, 0.0, item.Stock().Value())
	assert.Equal(t, 0.0, item.ReservedStock().Value())
	assert.True(t, item.AvailableStock().IsZero())
}

// TestWarehouseItem_AddStock 测试增加库存
func TestWarehouseItem_AddStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)

	err := item.AddStock(100.0)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, item.Stock().Value())

	err = item.AddStock(50.5)
	assert.NoError(t, err)
	assert.Equal(t, 150.5, item.Stock().Value())
}

// TestWarehouseItem_AddStock_Negative 测试增加负数库存
func TestWarehouseItem_AddStock_Negative(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)

	err := item.AddStock(-10.0)
	assert.Error(t, err)
	assert.Equal(t, "增加的库存数量不能为负数", err.Error())
}

// TestWarehouseItem_ReduceStock 测试减少库存
func TestWarehouseItem_ReduceStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)

	err := item.ReduceStock(30.0)
	assert.NoError(t, err)
	assert.Equal(t, 70.0, item.Stock().Value())
}

// TestWarehouseItem_ReduceStock_Insufficient 测试库存不足
func TestWarehouseItem_ReduceStock_Insufficient(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(50.0)

	err := item.ReduceStock(60.0)
	assert.Error(t, err)
	assert.Equal(t, "库存不足", err.Error())
	assert.Equal(t, 50.0, item.Stock().Value()) // 库存不变
}

// TestWarehouseItem_ReserveStock 测试预留库存
func TestWarehouseItem_ReserveStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)

	err := item.ReserveStock(30.0)
	assert.NoError(t, err)
	assert.Equal(t, 30.0, item.ReservedStock().Value())
	assert.Equal(t, 70.0, item.AvailableStock().Value())
}

// TestWarehouseItem_ReserveStock_Insufficient 测试可用库存不足
func TestWarehouseItem_ReserveStock_Insufficient(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(80.0)

	err := item.ReserveStock(30.0) // 可用库存只有20
	assert.Error(t, err)
	assert.Equal(t, "可用库存不足，无法预留", err.Error())
	assert.Equal(t, 80.0, item.ReservedStock().Value()) // 预留库存不变
}

// TestWarehouseItem_ReleaseReservedStock 测试释放预留库存
func TestWarehouseItem_ReleaseReservedStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(50.0)

	err := item.ReleaseReservedStock(20.0)
	assert.NoError(t, err)
	assert.Equal(t, 30.0, item.ReservedStock().Value())
	assert.Equal(t, 70.0, item.AvailableStock().Value())
}

// TestWarehouseItem_ReleaseReservedStock_Exceed 测试释放超过预留量
func TestWarehouseItem_ReleaseReservedStock_Exceed(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(30.0)

	err := item.ReleaseReservedStock(40.0)
	assert.Error(t, err)
	assert.Equal(t, "释放的预留库存不能超过已预留的库存", err.Error())
	assert.Equal(t, 30.0, item.ReservedStock().Value()) // 不变
}

// TestWarehouseItem_ConsumeReservedStock 测试消耗预留库存
func TestWarehouseItem_ConsumeReservedStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(50.0)

	err := item.ConsumeReservedStock(30.0)
	assert.NoError(t, err)
	assert.Equal(t, 70.0, item.Stock().Value())          // 库存减少
	assert.Equal(t, 20.0, item.ReservedStock().Value())  // 预留减少
	assert.Equal(t, 50.0, item.AvailableStock().Value()) // 可用库存 = 70 - 20
}

// TestWarehouseItem_ConsumeReservedStock_ExceedReserved 测试消耗超过预留量
func TestWarehouseItem_ConsumeReservedStock_ExceedReserved(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(30.0)

	err := item.ConsumeReservedStock(40.0)
	assert.Error(t, err)
	assert.Equal(t, "消耗的库存不能超过已预留的库存", err.Error())
}

// TestWarehouseItem_ConsumeReservedStock_InsufficientStock 测试消耗时库存不足
func TestWarehouseItem_ConsumeReservedStock_InsufficientStock(t *testing.T) {
	item := ReconstructWarehouseItem(1, 1, 100, "MAT001", 20.0, 50.0, 10.5, 0, 0)

	err := item.ConsumeReservedStock(30.0)
	assert.Error(t, err)
	assert.Equal(t, "库存不足", err.Error())
}

// TestWarehouseItem_UpdateValuation 测试更新估值单价
func TestWarehouseItem_UpdateValuation(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)

	err := item.UpdateValuation(12.8)
	assert.NoError(t, err)
	assert.Equal(t, 12.8, item.Valuation())

	// 负数应该返回错误
	err = item.UpdateValuation(-5.0)
	assert.Error(t, err)
	assert.Equal(t, "估值单价不能为负数", err.Error())
	assert.Equal(t, 12.8, item.Valuation()) // 不变
}

// TestWarehouseItem_HasStock 测试是否有库存
func TestWarehouseItem_HasStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	assert.False(t, item.HasStock())

	_ = item.AddStock(10.0)
	assert.True(t, item.HasStock())
}

// TestWarehouseItem_HasAvailableStock 测试是否有可用库存
func TestWarehouseItem_HasAvailableStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	assert.True(t, item.HasAvailableStock())

	_ = item.ReserveStock(100.0) // 全部预留
	assert.False(t, item.HasAvailableStock())
}

// TestWarehouseItem_TotalValue 测试库存总价值
func TestWarehouseItem_TotalValue(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)

	totalValue := item.TotalValue()
	assert.Equal(t, 1050.0, totalValue) // 100 * 10.5
}

// TestWarehouseItem_SetStock 测试设置库存（同步场景）
func TestWarehouseItem_SetStock(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(50.0)

	item.SetStock(200.0)
	assert.Equal(t, 200.0, item.Stock().Value())
}

// TestWarehouseItem_AvailableStock_Complex 测试复杂场景的可用库存计算
func TestWarehouseItem_AvailableStock_Complex(t *testing.T) {
	item := NewWarehouseItem(1, 100, "MAT001", 10.5)
	_ = item.AddStock(100.0)
	_ = item.ReserveStock(30.0)
	_ = item.ReduceStock(20.0)

	assert.Equal(t, 80.0, item.Stock().Value())
	assert.Equal(t, 30.0, item.ReservedStock().Value())
	assert.Equal(t, 50.0, item.AvailableStock().Value())
}

// TestReconstructWarehouseItem 测试从持久化数据重建
func TestReconstructWarehouseItem(t *testing.T) {
	item := ReconstructWarehouseItem(
		12345,
		1,
		100,
		"MAT001",
		150.0,
		30.0,
		10.5,
		1638360000,
		1638360000,
	)

	assert.Equal(t, uint64(12345), item.Uuid())
	assert.Equal(t, uint64(1), item.WarehouseUuid())
	assert.Equal(t, uint64(100), item.MaterialUuid())
	assert.Equal(t, "MAT001", item.MaterialCode())
	assert.Equal(t, 150.0, item.Stock().Value())
	assert.Equal(t, 30.0, item.ReservedStock().Value())
	assert.Equal(t, 120.0, item.AvailableStock().Value())
	assert.Equal(t, 10.5, item.Valuation())
}
