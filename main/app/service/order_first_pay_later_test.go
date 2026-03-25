package service

import (
	"testing"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
)

// TestGetMemberDineInOrderStatusInfo_OrderFirstPayLater tests status determination
// for "order-first-pay-later" mode orders (SaleBillStatusPending + H5Order present).
func TestGetMemberDineInOrderStatusInfo_OrderFirstPayLater(t *testing.T) {
	t.Parallel()

	srv := &orderSrv{}

	tests := []struct {
		name                 string
		saleBillStatus       uint
		isOrderFirstPayLater uint
		h5Order              *model.H5Order
		isProductionFinished bool
		wantStatus           string
	}{
		// 先下单后付模式 (OrderFirst)
		{
			name:                 "OrderFirst: Pending + H5 waiting acceptance → pending",
			saleBillStatus:       constant.SaleBillStatusPending,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusOrder},
			wantStatus:           constant.MemberDineInDetailStatusPending,
		},
		{
			name:                 "OrderFirst: Pending + H5 accepted, production not finished → preparing",
			saleBillStatus:       constant.SaleBillStatusPending,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusAccepted},
			wantStatus:           constant.MemberDineInDetailStatusPreparing,
		},
		{
			name:                 "OrderFirst: Pending + H5 accepted, production finished → unpaid",
			saleBillStatus:       constant.SaleBillStatusPending,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusAccepted},
			isProductionFinished: true,
			wantStatus:           constant.MemberDineInDetailStatusUnpaid,
		},
		{
			name:                 "OrderFirst: Pending + H5 rejected → rejected",
			saleBillStatus:       constant.SaleBillStatusPending,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusRejected},
			wantStatus:           constant.MemberDineInDetailStatusRejected,
		},
		{
			name:                 "OrderFirst: Complete → completed (final state)",
			saleBillStatus:       constant.SaleBillStatusComplete,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusAccepted},
			wantStatus:           constant.MemberDineInDetailStatusCompleted,
		},
		{
			name:                 "OrderFirst: Canceled + H5 rejected → rejected",
			saleBillStatus:       constant.SaleBillStatusCanceled,
			isOrderFirstPayLater: constant.OrderFirstPayLaterYes,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusRejected},
			wantStatus:           constant.MemberDineInDetailStatusRejected,
		},
		// 先付后下单模式 (PayFirst)
		{
			name:           "PayFirst: Pending + no H5 → unpaid",
			saleBillStatus: constant.SaleBillStatusPending,
			h5Order:        nil,
			wantStatus:     constant.MemberDineInDetailStatusUnpaid,
		},
		{
			name:           "PayFirst: Complete + H5 accepted, production not finished → preparing",
			saleBillStatus: constant.SaleBillStatusComplete,
			h5Order:        &model.H5Order{Status: constant.H5OrderStatusAccepted},
			wantStatus:     constant.MemberDineInDetailStatusPreparing,
		},
		{
			name:                 "PayFirst: Complete + H5 accepted, production finished → completed",
			saleBillStatus:       constant.SaleBillStatusComplete,
			h5Order:              &model.H5Order{Status: constant.H5OrderStatusAccepted},
			isProductionFinished: true,
			wantStatus:           constant.MemberDineInDetailStatusCompleted,
		},
		{
			name:           "PayFirst: Canceled + no H5 → cancelled",
			saleBillStatus: constant.SaleBillStatusCanceled,
			h5Order:        nil,
			wantStatus:     constant.MemberDineInDetailStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			saleBill := &model.SaleBill{
				Status:               tt.saleBillStatus,
				IsOrderFirstPayLater: tt.isOrderFirstPayLater,
			}
			result := srv.getMemberDineInOrderStatusInfo(saleBill, tt.h5Order, tt.isProductionFinished, "en")

			if result.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", result.Status, tt.wantStatus)
			}
		})
	}
}

// TestGetMemberDineInOrderStatusInfo_StatusTextNotEmpty ensures all status paths return non-empty text.
func TestGetMemberDineInOrderStatusInfo_StatusTextNotEmpty(t *testing.T) {
	t.Parallel()

	srv := &orderSrv{}

	cases := []struct {
		status  uint
		h5Order *model.H5Order
	}{
		{constant.SaleBillStatusPending, nil},
		{constant.SaleBillStatusPending, &model.H5Order{Status: constant.H5OrderStatusOrder}},
		{constant.SaleBillStatusPending, &model.H5Order{Status: constant.H5OrderStatusAccepted}},
		{constant.SaleBillStatusComplete, &model.H5Order{Status: constant.H5OrderStatusAccepted}},
		{constant.SaleBillStatusCanceled, nil},
		{constant.SaleBillStatusCanceled, &model.H5Order{Status: constant.H5OrderStatusRejected}},
	}

	for _, c := range cases {
		bill := &model.SaleBill{Status: c.status}
		result := srv.getMemberDineInOrderStatusInfo(bill, c.h5Order, false, "en")
		if result == (resp.MemberDineInOrderStatusInfo{}) {
			t.Errorf("empty result for status=%d h5=%v", c.status, c.h5Order)
		}
		if result.StatusText == "" {
			t.Errorf("empty StatusText for status=%d h5=%v", c.status, c.h5Order)
		}
	}
}
