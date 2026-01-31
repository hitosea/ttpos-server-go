package erp

import (
	"testing"

	"ttpos-server-go/app/constant"
)

// TestGetChannelBySource 测试 GetChannelBySource 方法
func TestGetChannelBySource(t *testing.T) {
	tests := []struct {
		name     string
		source   int
		expected string
	}{
		{
			name:     "LianLianPay source",
			source:   constant.PaymentMethodSourceLianLianPay,
			expected: "LianLianPay",
		},
		{
			name:     "Default source",
			source:   constant.PaymentMethodSourceDefault,
			expected: "",
		},
		{
			name:     "System source",
			source:   constant.PaymentMethodSourceSystem,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChannelBySource(tt.source)
			if result != tt.expected {
				t.Errorf("GetChannelBySource(%d) = %s, expected %s", tt.source, result, tt.expected)
			}
		})
	}
}
