package selling_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"ttpos-bmp/utility/uuid"
)

func init() {
	// 在测试包初始化时初始化 UUID 生成器
	// 使用 AppTypeERP 作为应用类型
	ctx := context.Background()
	uuid.InitIdGenerator(ctx, uuid.AppTypeERP)
}

// TestPaymentIDGeneration 测试 PaymentID 生成逻辑
// 这个测试不依赖 GoFrame 框架和配置文件，可以独立运行
func TestPaymentIDGeneration(t *testing.T) {
	// 测试自动生成的 PaymentID 格式
	paymentID := fmt.Sprintf("PID%d", uuid.MustGetID())

	// 验证格式：PID + 数字
	if paymentID == "" {
		t.Error("PaymentID 不应为空")
	}

	if len(paymentID) <= 3 {
		t.Errorf("PaymentID 长度应大于3，实际: %d", len(paymentID))
	}

	if paymentID[:3] != "PID" {
		t.Errorf("PaymentID 前缀应为 PID，实际: %s", paymentID[:3])
	}

	// 验证数字部分
	numPart := paymentID[3:]
	matched, err := regexp.MatchString(`^\d+$`, numPart)
	if err != nil {
		t.Errorf("正则表达式匹配失败: %v", err)
	}
	if !matched {
		t.Errorf("PaymentID 数字部分格式不正确: %s", numPart)
	}

	t.Logf("✓ 生成的 PaymentID: %s (长度: %d)", paymentID, len(paymentID))
}

// TestPaymentIDUniqueness 测试 PaymentID 唯一性
func TestPaymentIDUniqueness(t *testing.T) {
	generatedIDs := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		paymentID := fmt.Sprintf("PID%d", uuid.MustGetID())

		// 验证未重复
		if _, exists := generatedIDs[paymentID]; exists {
			t.Errorf("PaymentID %s 重复生成（第 %d 次）", paymentID, i+1)
		}

		generatedIDs[paymentID] = true
	}

	// 验证生成了正确数量的唯一 ID
	if len(generatedIDs) != count {
		t.Errorf("期望生成 %d 个唯一 ID，实际: %d", count, len(generatedIDs))
	}

	t.Logf("✓ 成功生成 %d 个唯一的 PaymentID", count)
}

// TestPaymentIDFormat 测试 PaymentID 格式验证
func TestPaymentIDFormat(t *testing.T) {
	testCases := []struct {
		name      string
		paymentID string
		valid     bool
	}{
		{"有效-16位数字", "PID1234567890123456", true},
		{"有效-3位数字", "PID123", true},
		{"有效-动态生成", "PID" + fmt.Sprintf("%d", uuid.MustGetID()), true},
		{"无效-空字符串", "", false},
		{"无效-仅前缀", "PID", false},
		{"无效-无前缀", "123456", false},
		{"无效-包含字母", "PIDABC", false},
		{"无效-包含特殊字符", "PID123-456", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := isValidPaymentID(tc.paymentID)
			if isValid != tc.valid {
				t.Errorf("PaymentID %q 期望有效性: %v，实际: %v", tc.paymentID, tc.valid, isValid)
			}
		})
	}
}

// isValidPaymentID 验证 PaymentID 格式是否有效
func isValidPaymentID(paymentID string) bool {
	if paymentID == "" || len(paymentID) <= 3 {
		return false
	}

	if paymentID[:3] != "PID" {
		return false
	}

	numPart := paymentID[3:]
	matched, _ := regexp.MatchString(`^\d+$`, numPart)
	return matched
}

// BenchmarkPaymentIDGeneration 基准测试 PaymentID 生成性能
func BenchmarkPaymentIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("PID%d", uuid.MustGetID())
	}
}
