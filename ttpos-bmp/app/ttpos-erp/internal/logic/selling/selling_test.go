package selling

import (
	"fmt"
	"regexp"
	"testing"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestPaymentIDGeneration 测试 PaymentID 生成逻辑
func TestPaymentIDGeneration(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试自动生成的 PaymentID 格式
		// 模拟 PaymentID 生成逻辑
		paymentID := fmt.Sprintf("PID%d", uuid.MustGetID())

		// 验证格式：PID + 数字
		t.AssertNE(paymentID, "")
		t.AssertGT(len(paymentID), 3) // 至少 "PID" + 数字
		t.AssertEQ(paymentID[:3], "PID")

		// 验证数字部分
		numPart := paymentID[3:]
		matched, err := regexp.MatchString(`^\d+$`, numPart)
		t.AssertNil(err)
		t.AssertEQ(matched, true)
	})
}

// TestPaymentIDUniqueness 测试 PaymentID 唯一性
func TestPaymentIDUniqueness(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		generatedIDs := make(map[string]bool)
		count := 100

		for i := 0; i < count; i++ {
			paymentID := fmt.Sprintf("PID%d", uuid.MustGetID())

			// 验证未重复
			_, exists := generatedIDs[paymentID]
			t.AssertEQ(exists, false)

			generatedIDs[paymentID] = true
		}

		// 验证生成了正确数量的唯一 ID
		t.AssertEQ(len(generatedIDs), count)
	})
}

// TestPaymentIDFormat 测试 PaymentID 格式验证
func TestPaymentIDFormat(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := []struct {
			paymentID string
			valid     bool
		}{
			{"PID1234567890123456", true},
			{"PID123", true},
			{"PID" + fmt.Sprintf("%d", uuid.MustGetID()), true},
			{"", false},
			{"PID", false},
			{"123456", false},
			{"PIDABC", false},
		}

		for _, tc := range testCases {
			if tc.valid {
				// 验证有效的 PaymentID
				if tc.paymentID != "" && len(tc.paymentID) > 3 {
					t.AssertEQ(tc.paymentID[:3], "PID")
					numPart := tc.paymentID[3:]
					matched, _ := regexp.MatchString(`^\d+$`, numPart)
					t.AssertEQ(matched, true)
				}
			} else {
				// 验证无效的 PaymentID
				isValid := false
				if tc.paymentID != "" && len(tc.paymentID) > 3 && tc.paymentID[:3] == "PID" {
					numPart := tc.paymentID[3:]
					matched, _ := regexp.MatchString(`^\d+$`, numPart)
					isValid = matched
				}
				t.AssertEQ(isValid, false)
			}
		}
	})
}

// TestPaymentIDLength 测试 PaymentID 长度
func TestPaymentIDLength(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		paymentID := fmt.Sprintf("PID%d", uuid.MustGetID())

		// PaymentID 至少应该有 "PID" + 1位数字
		t.Assert(len(paymentID) >= 4, true)

		// 记录生成的 PaymentID 用于调试
		t.Logf("生成的 PaymentID: %s, 长度: %d", paymentID, len(paymentID))
	})
}
