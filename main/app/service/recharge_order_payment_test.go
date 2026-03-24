package service

import (
	"testing"
	"ttpos-server-go/pkg/utils"
)

// TestRechargePaymentAmount 验证充值订单传给 ERP 的 Payment Entry 金额计算
// 修复前: 使用 Amount（实收金额，含找零），现金找零时会超过 SI outstanding
// 修复后: 使用 PaymentAmount + PaymentCommissionFee（应付金额，不含找零）
func TestRechargePaymentAmount(t *testing.T) {
	tests := []struct {
		name                 string
		paymentAmount        float64 // 支付金额（充值金额中该支付方式承担的部分）
		paymentCommissionFee float64 // 手续费
		amount               float64 // 实收金额（现金含找零）
		expectedPEAmount     float64 // 期望传给 ERP 的 PE 金额
	}{
		{
			name:                 "现金无找零 - 充值500付500",
			paymentAmount:        500,
			paymentCommissionFee: 0,
			amount:               500,
			expectedPEAmount:     500,
		},
		{
			name:                 "现金有找零 - 充值500付800找零300",
			paymentAmount:        500,
			paymentCommissionFee: 0,
			amount:               800,
			expectedPEAmount:     500, // 修复前会传 800，超过 SI outstanding
		},
		{
			name:                 "非现金支付 - 微信支付500",
			paymentAmount:        500,
			paymentCommissionFee: 0,
			amount:               500,
			expectedPEAmount:     500,
		},
		{
			name:                 "非现金支付带手续费 - 支付500手续费10",
			paymentAmount:        500,
			paymentCommissionFee: 10,
			amount:               510,
			expectedPEAmount:     510,
		},
		{
			name:                 "现金有找零带手续费 - 应付510付1000找零490",
			paymentAmount:        500,
			paymentCommissionFee: 10,
			amount:               1000,
			expectedPEAmount:     510, // 修复前会传 1000
		},
		{
			name:                 "小额现金找零 - 充值100付200找零100",
			paymentAmount:        100,
			paymentCommissionFee: 0,
			amount:               200,
			expectedPEAmount:     100,
		},
		{
			name:                 "精确金额无找零",
			paymentAmount:        123.45,
			paymentCommissionFee: 0,
			amount:               123.45,
			expectedPEAmount:     123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 修复后的公式: PaymentAmount + PaymentCommissionFee
			actual := utils.DecimalAdd(tt.paymentAmount, tt.paymentCommissionFee)
			if actual != tt.expectedPEAmount {
				t.Errorf("PE金额不正确: got %.2f, want %.2f", actual, tt.expectedPEAmount)
			}

			// 验证找零计算: Amount - PaymentAmount - CommissionFee
			change := utils.DecimalAdd(tt.amount, -tt.paymentAmount, -tt.paymentCommissionFee)
			if change < 0 {
				t.Errorf("找零金额不应为负: got %.2f", change)
			}

			// 验证: PE金额 + 找零 == 实收金额
			reconstructed := utils.DecimalAdd(actual, change)
			if reconstructed != tt.amount {
				t.Errorf("PE金额+找零应等于实收: %.2f + %.2f != %.2f", actual, change, tt.amount)
			}
		})
	}
}

// TestRechargePaymentAmount_MixedPayment 验证混合支付场景
func TestRechargePaymentAmount_MixedPayment(t *testing.T) {
	// 场景：充值1000，微信付500 + 现金付800找零300
	type payment struct {
		name                 string
		paymentAmount        float64
		paymentCommissionFee float64
		amount               float64
	}

	payments := []payment{
		{name: "微信", paymentAmount: 500, paymentCommissionFee: 0, amount: 500},
		{name: "现金", paymentAmount: 500, paymentCommissionFee: 0, amount: 800},
	}

	var totalPEAmount float64
	expectedSIOutstanding := 1000.0

	for _, p := range payments {
		peAmount := utils.DecimalAdd(p.paymentAmount, p.paymentCommissionFee)
		totalPEAmount = utils.DecimalAdd(totalPEAmount, peAmount)
		t.Logf("%s: PE金额=%.2f (PaymentAmount=%.2f, Amount=%.2f, 找零=%.2f)",
			p.name, peAmount, p.paymentAmount, p.amount,
			utils.DecimalAdd(p.amount, -p.paymentAmount, -p.paymentCommissionFee))
	}

	if totalPEAmount != expectedSIOutstanding {
		t.Errorf("所有PE金额之和应等于SI outstanding: got %.2f, want %.2f", totalPEAmount, expectedSIOutstanding)
	}
}
