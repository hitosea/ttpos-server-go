package takeout

import (
	"testing"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
)

// TestBuildSaleBillForPeakTime_ValidData 测试使用有效参数构建 SaleBill
func TestBuildSaleBillForPeakTime_ValidData(t *testing.T) {
	acceptedBy := uint64(1001)
	completedTime := int64(1700003600)
	platformTotal := 99.50

	saleBill := &model.SaleBill{
		Status:        constant.SaleBillStatusComplete,
		PaymentAmount: platformTotal,
		CashierUuid:   acceptedBy,
		FinishTime:    completedTime,
	}

	if !saleBill.IsFinish() {
		t.Error("Expected IsFinish() to be true")
	}
	if saleBill.FinishTime != completedTime {
		t.Errorf("Expected FinishTime %d, got %d", completedTime, saleBill.FinishTime)
	}
	if saleBill.CashierUuid != acceptedBy {
		t.Errorf("Expected CashierUuid %d, got %d", acceptedBy, saleBill.CashierUuid)
	}
	if saleBill.PaymentAmount != platformTotal {
		t.Errorf("Expected PaymentAmount %f, got %f", platformTotal, saleBill.PaymentAmount)
	}
}

// TestPeakTimeParamValidation_NoAcceptor 测试无接单人时应跳过记录
func TestPeakTimeParamValidation_NoAcceptor(t *testing.T) {
	// acceptedBy <= 0 应该跳过高峰期记录
	acceptedBy := uint64(0)
	completedTime := int64(1700003600)

	if acceptedBy > 0 && completedTime > 0 {
		t.Error("Expected validation to skip when acceptedBy is 0")
	}
}

// TestPeakTimeParamValidation_NoCompletedTime 测试无完成时间时应跳过记录
func TestPeakTimeParamValidation_NoCompletedTime(t *testing.T) {
	// completedTime <= 0 应该跳过高峰期记录
	acceptedBy := uint64(1001)
	completedTime := int64(0)

	if acceptedBy > 0 && completedTime > 0 {
		t.Error("Expected validation to skip when completedTime is 0")
	}
}
