package repository

import (
	"strings"
	"testing"

	"ttpos-server-go/app/constant"
)

// ==================== buildBillTypeList ====================

func TestBuildBillTypeList(t *testing.T) {
	tests := []struct {
		name        string
		req         CountBusinessTimePeriodReq
		expectedLen int
		contains    []uint
	}{
		{
			name:        "all false returns empty",
			req:         CountBusinessTimePeriodReq{},
			expectedLen: 0,
		},
		{
			name:        "only desk",
			req:         CountBusinessTimePeriodReq{IsDesk: true},
			expectedLen: 1,
			contains:    []uint{constant.SaleBillTypeDesk},
		},
		{
			name:        "only instant",
			req:         CountBusinessTimePeriodReq{IsInstant: true},
			expectedLen: 1,
			contains:    []uint{constant.SaleBillTypeInstant},
		},
		{
			name:        "only takeout",
			req:         CountBusinessTimePeriodReq{IsTakeout: true},
			expectedLen: 1,
			contains:    []uint{constant.SaleBillTypeTakeout},
		},
		{
			name:        "desk and instant",
			req:         CountBusinessTimePeriodReq{IsDesk: true, IsInstant: true},
			expectedLen: 2,
			contains:    []uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant},
		},
		{
			name:        "all types",
			req:         CountBusinessTimePeriodReq{IsDesk: true, IsInstant: true, IsTakeout: true},
			expectedLen: 3,
			contains:    []uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant, constant.SaleBillTypeTakeout},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildBillTypeList(tt.req)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(result))
			}
			for _, expected := range tt.contains {
				found := false
				for _, got := range result {
					if got == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to contain %d, got %v", expected, result)
				}
			}
		})
	}
}

// ==================== timePeriodAggregateWrapper ====================

func TestTimePeriodAggregateWrapper(t *testing.T) {
	subQuery := "SELECT * FROM orders"
	result := timePeriodAggregateWrapper(subQuery)

	if !strings.Contains(result, "SELECT * FROM orders") {
		t.Error("should contain the subQuery")
	}
	if !strings.Contains(result, "SUM(order_amount)") {
		t.Error("should contain SUM(order_amount)")
	}
	if !strings.Contains(result, "SUM(pay_amount)") {
		t.Error("should contain SUM(pay_amount)")
	}
	if !strings.Contains(result, "SUM(refund_amount)") {
		t.Error("should contain SUM(refund_amount)")
	}
	if !strings.Contains(result, "COUNT(DISTINCT sale_bill_uuid)") {
		t.Error("should contain COUNT(DISTINCT sale_bill_uuid)")
	}
	if !strings.Contains(result, "GROUP BY period_start_time") {
		t.Error("should contain GROUP BY period_start_time")
	}
	if !strings.Contains(result, "LIMIT ?") {
		t.Error("should contain LIMIT placeholder")
	}
	if !strings.Contains(result, "OFFSET ?") {
		t.Error("should contain OFFSET placeholder")
	}
}

// ==================== containsOrderType ====================

func TestContainsOrderType(t *testing.T) {
	types := []uint{0, 1, 2}

	if !containsOrderType(types, 0) {
		t.Error("should contain 0")
	}
	if !containsOrderType(types, 1) {
		t.Error("should contain 1")
	}
	if !containsOrderType(types, 2) {
		t.Error("should contain 2")
	}
	if containsOrderType(types, 3) {
		t.Error("should not contain 3")
	}
	if containsOrderType(nil, 0) {
		t.Error("nil slice should not contain anything")
	}
}

// ==================== buildStateInCondition ====================

func TestBuildStateInCondition(t *testing.T) {
	tests := []struct {
		name     string
		states   []int
		expected string
	}{
		{"empty", []int{}, ""},
		{"single", []int{40}, "(40)"},
		{"multiple", []int{10, 20, 30, 40}, "(10,20,30,40)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildStateInCondition(tt.states)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
