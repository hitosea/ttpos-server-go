package service

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
)

// hasPaidPaymentOrder 提取自 SyncMemberOrderToErp 的支付状态检查逻辑
// 检查 SaleOrder 是否有已支付的 PaymentOrder（未删除且 Status == PaymentOrderStatusPaid）
func hasPaidPaymentOrder(saleOrder *model.SaleOrder) bool {
	for _, po := range saleOrder.PaymentOrders {
		if !po.IsDelete() && po.Status == constant.PaymentOrderStatusPaid {
			return true
		}
	}
	return false
}

// TestHasPaidPaymentOrder_NoPO 没有付款单时应返回 false
func TestHasPaidPaymentOrder_NoPO(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{},
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when no payment orders")
	}
}

// TestHasPaidPaymentOrder_NilPO PaymentOrders 为 nil 时应返回 false
func TestHasPaidPaymentOrder_NilPO(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: nil,
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when payment orders is nil")
	}
}

// TestHasPaidPaymentOrder_UnpaidOnly 只有未支付的付款单时应返回 false
func TestHasPaidPaymentOrder_UnpaidOnly(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				Status: constant.PaymentOrderStatusUnPay, // 0 = 未支付
			},
		},
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when only unpaid payment orders")
	}
}

// TestHasPaidPaymentOrder_PaidExists 有已支付的付款单时应返回 true
func TestHasPaidPaymentOrder_PaidExists(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				Status: constant.PaymentOrderStatusPaid, // 1 = 已支付
			},
		},
	}
	if !hasPaidPaymentOrder(so) {
		t.Error("expected true when paid payment order exists")
	}
}

// TestHasPaidPaymentOrder_DeletedPaid 已删除的已支付付款单不应计入
func TestHasPaidPaymentOrder_DeletedPaid(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				BaseModel: model.BaseModel{DeleteTime: 1000}, // 已删除
				Status:    constant.PaymentOrderStatusPaid,
			},
		},
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when paid payment order is deleted")
	}
}

// TestHasPaidPaymentOrder_MixedStatus 混合状态：一个未支付 + 一个已删除已支付 + 一个已支付
func TestHasPaidPaymentOrder_MixedStatus(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				Status: constant.PaymentOrderStatusUnPay,
			},
			{
				BaseModel: model.BaseModel{DeleteTime: 1000},
				Status:    constant.PaymentOrderStatusPaid,
			},
			{
				Status: constant.PaymentOrderStatusPaid,
			},
		},
	}
	if !hasPaidPaymentOrder(so) {
		t.Error("expected true when at least one valid paid payment order exists")
	}
}

// TestHasPaidPaymentOrder_AllDeletedPaid 所有已支付的付款单都被删除
func TestHasPaidPaymentOrder_AllDeletedPaid(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				BaseModel: model.BaseModel{DeleteTime: 1000},
				Status:    constant.PaymentOrderStatusPaid,
			},
			{
				BaseModel: model.BaseModel{DeleteTime: 2000},
				Status:    constant.PaymentOrderStatusPaid,
			},
		},
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when all paid payment orders are deleted")
	}
}

// TestHasPaidPaymentOrder_RefundedStatus 退款状态不应视为已支付
func TestHasPaidPaymentOrder_RefundedStatus(t *testing.T) {
	so := &model.SaleOrder{
		PaymentOrders: []*model.PaymentOrder{
			{
				Status: constant.PaymentOrderStatusRefund, // 2 = 已退款
			},
		},
	}
	if hasPaidPaymentOrder(so) {
		t.Error("expected false when payment order is refunded")
	}
}
