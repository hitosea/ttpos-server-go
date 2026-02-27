package takeout

import (
	"testing"

	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	valueObject "ttpos-server-go/app/modules/takeout/domain/value_object"
)

// TestDetermineRecordType_Completed 测试完成订单（有接单人）返回 "inc"
func TestDetermineRecordType_Completed(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:    valueObject.TakeoutOrderStateCompleted, // 40 - 已完成
		AcceptedBy:    1001,
		AcceptedTime:  1700000000,
		CompletedTime: 1700003600,
	}

	result := determineRecordType(order)
	if result != "inc" {
		t.Errorf("Expected 'inc' for completed order with acceptor, got '%s'", result)
	}
}

// TestDetermineRecordType_CompletedNoAcceptor 测试完成订单（无接单人）返回 ""
func TestDetermineRecordType_CompletedNoAcceptor(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:    valueObject.TakeoutOrderStateCompleted,
		AcceptedBy:    0, // 无接单人
		AcceptedTime:  0,
		CompletedTime: 1700003600,
	}

	result := determineRecordType(order)
	if result != "" {
		t.Errorf("Expected '' for completed order without acceptor, got '%s'", result)
	}
}

// TestDetermineRecordType_CompletedNoCompletedTime 测试完成订单（无完成时间）返回 ""
func TestDetermineRecordType_CompletedNoCompletedTime(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:    valueObject.TakeoutOrderStateCompleted,
		AcceptedBy:    1001,
		AcceptedTime:  1700000000,
		CompletedTime: 0, // 无完成时间
	}

	result := determineRecordType(order)
	if result != "" {
		t.Errorf("Expected '' for completed order without completed time, got '%s'", result)
	}
}

// TestDetermineRecordType_Accepted 测试接单订单返回 ""
func TestDetermineRecordType_Accepted(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:   valueObject.TakeoutOrderStateAccepted, // 10 - 已接单
		AcceptedBy:   1001,
		AcceptedTime: 1700000000,
	}

	result := determineRecordType(order)
	if result != "" {
		t.Errorf("Expected '' for accepted order (not completed), got '%s'", result)
	}
}

// TestDetermineRecordType_Canceled 测试取消订单返回 ""
func TestDetermineRecordType_Canceled(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:   valueObject.TakeoutOrderStateCanceled, // 60 - 已取消
		AcceptedBy:   1001,
		AcceptedTime: 1700000000,
		RejectedTime: 1700001800,
	}

	result := determineRecordType(order)
	if result != "" {
		t.Errorf("Expected '' for canceled order, got '%s'", result)
	}
}

// TestDetermineRecordType_Pending 测试待接单订单返回 ""
func TestDetermineRecordType_Pending(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState: valueObject.TakeoutOrderStatePending, // 0 - 待接单
	}

	result := determineRecordType(order)
	if result != "" {
		t.Errorf("Expected '' for pending order, got '%s'", result)
	}
}

// TestBuildSaleBillFromTakeoutOrder_UsesCompletedTime 测试使用完成时间
func TestBuildSaleBillFromTakeoutOrder_UsesCompletedTime(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:    valueObject.TakeoutOrderStateCompleted,
		AcceptedBy:    1001,
		AcceptedTime:  1700000000,
		CompletedTime: 1700003600, // 完成时间
		PlatformTotal: 99.50,
	}

	saleBill := buildSaleBillFromTakeoutOrder(order)

	if saleBill == nil {
		t.Fatal("Expected saleBill, got nil")
	}

	// 验证使用完成时间
	if saleBill.FinishTime != order.CompletedTime {
		t.Errorf("Expected FinishTime %d, got %d", order.CompletedTime, saleBill.FinishTime)
	}

	// 验证使用接单人
	if saleBill.CashierUuid != order.AcceptedBy {
		t.Errorf("Expected CashierUuid %d, got %d", order.AcceptedBy, saleBill.CashierUuid)
	}

	// 验证金额
	if saleBill.PaymentAmount != order.PlatformTotal {
		t.Errorf("Expected PaymentAmount %f, got %f", order.PlatformTotal, saleBill.PaymentAmount)
	}
}

// TestBuildSaleBillFromTakeoutOrder_NoCompletedTime 测试无完成时间返回 nil
func TestBuildSaleBillFromTakeoutOrder_NoCompletedTime(t *testing.T) {
	order := &takeoutModel.TakeoutOrder{
		OrderState:    valueObject.TakeoutOrderStateCompleted,
		AcceptedBy:    1001,
		AcceptedTime:  1700000000,
		CompletedTime: 0, // 无完成时间
		PlatformTotal: 99.50,
	}

	saleBill := buildSaleBillFromTakeoutOrder(order)

	if saleBill != nil {
		t.Errorf("Expected nil for order without completed time, got saleBill")
	}
}
