package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder_ChainedBuild(t *testing.T) {
	// 链式创建订单
	order, err := NewOrder(1001).
		WithOrderNo("ORD20241204001").
		WithCustomer(2001).
		WithDesk(3001).
		WithRemark("不要辣").
		AddItem(100, "红烧肉", 1, 68.00).
		AddItem(101, "米饭", 2, 3.00).
		ApplyPercentDiscount(10, "新客优惠").
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "ORD20241204001", order.OrderNo())
	assert.Equal(t, uint64(2001), order.CustomerUuid())
	assert.Equal(t, uint64(3001), order.DeskUuid())
	assert.Equal(t, "不要辣", order.Remark())
	assert.Equal(t, 2, order.ItemCount())
}

func TestNewOrder_EmptyOrderNo(t *testing.T) {
	// 订单号为空应该报错
	order, err := NewOrder(1001).
		WithOrderNo("").
		Build()

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, "订单号不能为空", err.Error())
}

func TestOrder_AddItem(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 68.00).
		AddItem(101, "米饭", 2, 3.00).
		AddItemWithRemark(102, "可乐", 1, 8.00, "加冰").
		Build()

	assert.NoError(t, err)
	assert.Equal(t, 3, order.ItemCount())

	// 验证小计
	// 68 + 6 + 8 = 82
	assert.Equal(t, 82.0, order.SubTotal())
}

func TestOrder_AddItem_InvalidQuantity(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 0, 68.00). // 数量为0
		Build()

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, "数量必须大于0", err.Error())
}

func TestOrder_ApplyPercentDiscount(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 100.00).
		ApplyPercentDiscount(10, "新客优惠"). // 10%折扣
		Build()

	assert.NoError(t, err)
	assert.Equal(t, 100.0, order.SubTotal())
	assert.Equal(t, 10.0, order.TotalDiscount()) // 100 * 10% = 10
	assert.Equal(t, 90.0, order.Total())         // 100 - 10 = 90
}

func TestOrder_ApplyFixedDiscount(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 100.00).
		ApplyFixedDiscount(15, "满100减15").
		Build()

	assert.NoError(t, err)
	assert.Equal(t, 100.0, order.SubTotal())
	assert.Equal(t, 15.0, order.TotalDiscount())
	assert.Equal(t, 85.0, order.Total()) // 100 - 15 = 85
}

func TestOrder_MultipleDiscounts(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 100.00).
		ApplyPercentDiscount(10, "新客优惠"). // 10
		ApplyFixedDiscount(5, "优惠券").     // 5
		Build()

	assert.NoError(t, err)
	assert.Equal(t, 100.0, order.SubTotal())
	assert.Equal(t, 15.0, order.TotalDiscount()) // 10 + 5 = 15
	assert.Equal(t, 85.0, order.Total())         // 100 - 15 = 85
}

func TestOrder_Confirm(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 68.00).
		Build()

	assert.NoError(t, err)
	assert.True(t, order.Status().IsPending())

	// 确认订单
	err = order.Confirm()
	assert.NoError(t, err)
	assert.True(t, order.Status().IsConfirmed())
}

func TestOrder_Confirm_NoItems(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		Build()

	assert.NoError(t, err)

	// 没有商品不能确认
	err = order.Confirm()
	assert.Error(t, err)
	assert.Equal(t, "订单必须包含至少一个商品", err.Error())
}

func TestOrder_Cancel(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 68.00).
		Build()

	assert.NoError(t, err)

	// 取消订单
	err = order.Cancel()
	assert.NoError(t, err)
	assert.True(t, order.Status().IsCancelled())
}

func TestOrder_Cancel_Completed(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 68.00).
		Build()

	assert.NoError(t, err)

	// 模拟订单流程
	order.Confirm()
	order.StartPreparing()
	order.Complete()

	// 已完成的订单不能取消
	err = order.Cancel()
	assert.Error(t, err)
	assert.Equal(t, "当前订单状态不允许取消", err.Error())
}

func TestOrder_CannotAddItemAfterConfirmed(t *testing.T) {
	order, err := NewOrder(1001).
		WithOrderNo("ORD001").
		AddItem(100, "红烧肉", 1, 68.00).
		Build()

	assert.NoError(t, err)
	order.Confirm()
	order.StartPreparing() // 进入制作中状态

	// 制作中不能添加商品
	order.AddItem(101, "米饭", 1, 3.00)
	_, err = order.Build()
	assert.Error(t, err)
	assert.Equal(t, "当前订单状态不允许添加商品", err.Error())
}

func TestOrder_FullWorkflow(t *testing.T) {
	// 完整的订单流程
	order, err := NewOrder(1001).
		WithOrderNo("ORD20241204001").
		WithCustomer(2001).
		WithDesk(3001).
		AddItem(100, "红烧肉", 1, 68.00).
		AddItem(101, "米饭", 2, 3.00).
		ApplyPercentDiscount(10, "新客优惠").
		Build()

	assert.NoError(t, err)

	// 待处理 -> 已确认
	assert.True(t, order.Status().IsPending())
	err = order.Confirm()
	assert.NoError(t, err)
	assert.True(t, order.Status().IsConfirmed())

	// 已确认 -> 制作中
	err = order.StartPreparing()
	assert.NoError(t, err)
	assert.True(t, order.Status().IsPreparing())

	// 制作中 -> 已完成
	err = order.Complete()
	assert.NoError(t, err)
	assert.True(t, order.Status().IsCompleted())

	// 验证最终金额
	// 小计: 68 + 6 = 74
	// 优惠: 74 * 10% = 7.4
	// 总计: 74 - 7.4 = 66.6
	assert.Equal(t, 74.0, order.SubTotal())
	assert.Equal(t, 7.4, order.TotalDiscount())
	assert.Equal(t, 66.6, order.Total())
}
