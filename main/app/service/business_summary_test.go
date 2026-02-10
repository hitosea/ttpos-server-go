package service

import (
	"sort"
	"strings"
	"testing"

	"ttpos-server-go/app/dto/resp"
)

// TestStoreCodeSorting 测试店铺编号排序逻辑
// 排序规则：
// 1. 无编号优先，按创建时间顺序排序
// 2. 有编号：数字(0-9)优先，字母(a-z)其次
// 3. 编号相同时按创建时间顺序排序
func TestStoreCodeSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    []*resp.CompanySummaryItem
		expected []uint64 // 排序后的 CompanyUuid 顺序（用于验证完整排序结果）
	}{
		{
			name: "无编号优先，按创建时间排序",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "001", CreateTime: 100},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "", CreateTime: 300},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "", CreateTime: 200},
			},
			expected: []uint64{3, 2, 1}, // 无编号按创建时间排序：200 < 300，然后是有编号的
		},
		{
			name: "数字优先于字母",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "ABC", CreateTime: 100},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "123", CreateTime: 200},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "XYZ", CreateTime: 300},
			},
			expected: []uint64{2, 1, 3}, // 数字优先：123, ABC, XYZ
		},
		{
			name: "综合排序：无编号按创建时间 > 数字 > 字母",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "B01", CreateTime: 100},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "", CreateTime: 500},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "001", CreateTime: 200},
				{CompanyUuid: 4, CompanyName: "Store D", StoreCode: "A02", CreateTime: 300},
				{CompanyUuid: 5, CompanyName: "Store E", StoreCode: "", CreateTime: 400},
				{CompanyUuid: 6, CompanyName: "Store F", StoreCode: "002", CreateTime: 600},
			},
			expected: []uint64{5, 2, 3, 6, 4, 1}, // 无编号按创建时间(400,500)，数字(001,002)，字母(A02,B01)
		},
		{
			name: "编号相同时按创建时间排序",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "001", CreateTime: 300},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "001", CreateTime: 100},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "001", CreateTime: 200},
			},
			expected: []uint64{2, 3, 1}, // 编号相同，按创建时间排序：100 < 200 < 300
		},
		{
			name: "大小写不敏感排序，相同时按创建时间",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "abc", CreateTime: 300},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "ABC", CreateTime: 100},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "Abc", CreateTime: 200},
			},
			expected: []uint64{2, 3, 1}, // 大小写相同，按创建时间排序
		},
		{
			name: "数字排序（字符串排序）",
			input: []*resp.CompanySummaryItem{
				{CompanyUuid: 1, CompanyName: "Store A", StoreCode: "10", CreateTime: 100},
				{CompanyUuid: 2, CompanyName: "Store B", StoreCode: "2", CreateTime: 200},
				{CompanyUuid: 3, CompanyName: "Store C", StoreCode: "1", CreateTime: 300},
			},
			expected: []uint64{3, 1, 2}, // 字符串排序：1 < 10 < 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 复制输入以避免修改原始数据
			companyList := make([]*resp.CompanySummaryItem, len(tt.input))
			for i, item := range tt.input {
				companyList[i] = &resp.CompanySummaryItem{
					CompanyUuid: item.CompanyUuid,
					CompanyName: item.CompanyName,
					StoreCode:   item.StoreCode,
					CreateTime:  item.CreateTime,
				}
			}

			// 执行排序（与 GetCompanyList 中的排序逻辑相同）
			sort.Slice(companyList, func(i, j int) bool {
				codeI := companyList[i].StoreCode
				codeJ := companyList[j].StoreCode

				// 无编号优先
				if codeI == "" && codeJ != "" {
					return true
				}
				if codeI != "" && codeJ == "" {
					return false
				}
				if codeI == "" && codeJ == "" {
					// 都无编号时按创建时间顺序排序
					return companyList[i].CreateTime < companyList[j].CreateTime
				}

				// 获取首字符进行类型判断
				firstI := rune(codeI[0])
				firstJ := rune(codeJ[0])

				// 数字优先于字母
				isDigitI := firstI >= '0' && firstI <= '9'
				isDigitJ := firstJ >= '0' && firstJ <= '9'

				if isDigitI && !isDigitJ {
					return true
				}
				if !isDigitI && isDigitJ {
					return false
				}

				// 同类型按字符串排序（大小写不敏感）
				lowerI := strings.ToLower(codeI)
				lowerJ := strings.ToLower(codeJ)
				if lowerI != lowerJ {
					return lowerI < lowerJ
				}
				// 编号相同时按创建时间顺序排序
				return companyList[i].CreateTime < companyList[j].CreateTime
			})

			// 验证结果
			for i, expectedUuid := range tt.expected {
				if companyList[i].CompanyUuid != expectedUuid {
					t.Errorf("位置 %d: 期望 CompanyUuid=%d, 实际 CompanyUuid=%d (StoreCode=%q, CreateTime=%d)",
						i, expectedUuid, companyList[i].CompanyUuid, companyList[i].StoreCode, companyList[i].CreateTime)
				}
			}
		})
	}
}

// TestStoreNameFormatting 测试店铺名称格式化逻辑
func TestStoreNameFormatting(t *testing.T) {
	tests := []struct {
		name         string
		storeCode    string
		companyName  string
		expectedName string
	}{
		{
			name:         "有编号时格式化为 {编号} {名称}",
			storeCode:    "001",
			companyName:  "华莱士xxx门店",
			expectedName: "001 华莱士xxx门店",
		},
		{
			name:         "无编号时仅显示名称",
			storeCode:    "",
			companyName:  "华莱士xxx门店",
			expectedName: "华莱士xxx门店",
		},
		{
			name:         "字母编号格式化",
			storeCode:    "ABC",
			companyName:  "测试门店",
			expectedName: "ABC 测试门店",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟格式化逻辑
			var formattedName string
			if tt.storeCode != "" {
				formattedName = tt.storeCode + " " + tt.companyName
			} else {
				formattedName = tt.companyName
			}

			if formattedName != tt.expectedName {
				t.Errorf("期望名称=%q, 实际名称=%q", tt.expectedName, formattedName)
			}
		})
	}
}

// TestCashACCalculation 测试现金AC计算逻辑
func TestCashACCalculation(t *testing.T) {
	tests := []struct {
		name       string
		cashTC     int64
		cashAmount float64
		expectedAC float64
	}{
		{
			name:       "正常计算",
			cashTC:     10,
			cashAmount: 1000.00,
			expectedAC: 100.00,
		},
		{
			name:       "除零场景返回0",
			cashTC:     0,
			cashAmount: 0.00,
			expectedAC: 0.00,
		},
		{
			name:       "保留2位小数",
			cashTC:     3,
			cashAmount: 100.00,
			expectedAC: 33.33,
		},
		{
			name:       "有现金金额但TC为0",
			cashTC:     0,
			cashAmount: 100.00,
			expectedAC: 0.00, // 除零保护
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cashAC float64
			if tt.cashTC > 0 {
				// 使用与实际代码相同的计算方式
				cashAC = float64(int(tt.cashAmount/float64(tt.cashTC)*100+0.5)) / 100
			}

			if cashAC != tt.expectedAC {
				t.Errorf("期望 CashAC=%.2f, 实际 CashAC=%.2f", tt.expectedAC, cashAC)
			}
		})
	}
}

// TestCalculateBusinessAverage_Normal 测试营业数据平均值计算 - 正常数据
func TestCalculateBusinessAverage_Normal(t *testing.T) {
	items := []resp.CompanyBusinessSummaryItem{
		{
			OrderAmount:      1000.00,
			PayAmount:        900.00,
			OrderNum:         10,
			CashTC:           5,
			CashAmount:       500.00,
			CashAC:           100.00,
			InStoreAmount:    600.00,
			InStoreOrderNum:  6,
			DeliveryAmount:   400.00,
			DeliveryOrderNum: 4,
			MealNum:          20,
			DeskNum:          5,
			AvgCustomerPrice: 45.00,
			OrderAmountMealAvg: 50.00,
			OrderAmountAvg:   100.00,
			PayAmountAvg:     90.00,
			InstantOrderAmount: 300.00,
			DeskOrderAmount:    300.00,
			TakeoutOrderAmount: 200.00,
		},
		{
			OrderAmount:      2000.00,
			PayAmount:        1800.00,
			OrderNum:         20,
			CashTC:           10,
			CashAmount:       1000.00,
			CashAC:           100.00,
			InStoreAmount:    1200.00,
			InStoreOrderNum:  12,
			DeliveryAmount:   800.00,
			DeliveryOrderNum: 8,
			MealNum:          40,
			DeskNum:          10,
			AvgCustomerPrice: 45.00,
			OrderAmountMealAvg: 50.00,
			OrderAmountAvg:   100.00,
			PayAmountAvg:     90.00,
			InstantOrderAmount: 600.00,
			DeskOrderAmount:    600.00,
			TakeoutOrderAmount: 400.00,
		},
	}

	avg := calculateBusinessAverage(items)

	// 验证平均值计算（2条数据）
	if avg.OrderAmount != 1500.00 {
		t.Errorf("OrderAmount: 期望 1500.00, 实际 %.2f", avg.OrderAmount)
	}
	if avg.PayAmount != 1350.00 {
		t.Errorf("PayAmount: 期望 1350.00, 实际 %.2f", avg.PayAmount)
	}
	if avg.OrderNum != 15.00 {
		t.Errorf("OrderNum: 期望 15.00, 实际 %.2f", avg.OrderNum)
	}
	if avg.InStoreAmount != 900.00 {
		t.Errorf("InStoreAmount: 期望 900.00, 实际 %.2f", avg.InStoreAmount)
	}
	if avg.InStoreOrderNum != 9.00 {
		t.Errorf("InStoreOrderNum: 期望 9.00, 实际 %.2f", avg.InStoreOrderNum)
	}
	if avg.DeliveryAmount != 600.00 {
		t.Errorf("DeliveryAmount: 期望 600.00, 实际 %.2f", avg.DeliveryAmount)
	}
	if avg.DeliveryOrderNum != 6.00 {
		t.Errorf("DeliveryOrderNum: 期望 6.00, 实际 %.2f", avg.DeliveryOrderNum)
	}
}

// TestCalculateBusinessAverage_Empty 测试营业数据平均值计算 - 空数据返回零值
func TestCalculateBusinessAverage_Empty(t *testing.T) {
	items := []resp.CompanyBusinessSummaryItem{}

	avg := calculateBusinessAverage(items)

	// 验证空数据返回零值
	if avg.OrderAmount != 0 {
		t.Errorf("OrderAmount: 期望 0, 实际 %.2f", avg.OrderAmount)
	}
	if avg.PayAmount != 0 {
		t.Errorf("PayAmount: 期望 0, 实际 %.2f", avg.PayAmount)
	}
	if avg.InStoreAmount != 0 {
		t.Errorf("InStoreAmount: 期望 0, 实际 %.2f", avg.InStoreAmount)
	}
	if avg.DeliveryAmount != 0 {
		t.Errorf("DeliveryAmount: 期望 0, 实际 %.2f", avg.DeliveryAmount)
	}
}

// TestCalculateRefundAverage_Normal 测试退款平均值计算 - 正常数据
func TestCalculateRefundAverage_Normal(t *testing.T) {
	items := []resp.CompanyRefundSummaryItem{
		{
			RefundAmount:        100.00,
			RefundNum:           5,
			RefundRate:          10.00,
			AvgRefundAmount:     20.00,
			PartialRefundAmount: 60.00,
			PartialRefundNum:    3,
			FullRefundAmount:    40.00,
			FullRefundNum:       2,
		},
		{
			RefundAmount:        200.00,
			RefundNum:           10,
			RefundRate:          20.00,
			AvgRefundAmount:     20.00,
			PartialRefundAmount: 120.00,
			PartialRefundNum:    6,
			FullRefundAmount:    80.00,
			FullRefundNum:       4,
		},
	}

	avg := calculateRefundAverage(items)

	// 验证平均值计算
	if avg.RefundAmount != 150.00 {
		t.Errorf("RefundAmount: 期望 150.00, 实际 %.2f", avg.RefundAmount)
	}
	if avg.RefundNum != 7.50 {
		t.Errorf("RefundNum: 期望 7.50, 实际 %.2f", avg.RefundNum)
	}
	if avg.AvgRefundAmount != 20.00 {
		t.Errorf("AvgRefundAmount: 期望 20.00, 实际 %.2f", avg.AvgRefundAmount)
	}
}

// TestCalculateRefundAverage_Empty 测试退款平均值计算 - 空数据返回零值
func TestCalculateRefundAverage_Empty(t *testing.T) {
	items := []resp.CompanyRefundSummaryItem{}

	avg := calculateRefundAverage(items)

	// 验证空数据返回零值
	if avg.RefundAmount != 0 {
		t.Errorf("RefundAmount: 期望 0, 实际 %.2f", avg.RefundAmount)
	}
	if avg.RefundNum != 0 {
		t.Errorf("RefundNum: 期望 0, 实际 %.2f", avg.RefundNum)
	}
	if avg.AvgRefundAmount != 0 {
		t.Errorf("AvgRefundAmount: 期望 0, 实际 %.2f", avg.AvgRefundAmount)
	}
}

// TestAvgRefundAmountCalculation 测试平均退款额计算
func TestAvgRefundAmountCalculation(t *testing.T) {
	tests := []struct {
		name               string
		refundAmount       float64
		refundNum          int64
		expectedAvgRefund  float64
	}{
		{
			name:              "正常计算",
			refundAmount:      100.00,
			refundNum:         5,
			expectedAvgRefund: 20.00,
		},
		{
			name:              "除数为0返回0",
			refundAmount:      100.00,
			refundNum:         0,
			expectedAvgRefund: 0.00,
		},
		{
			name:              "金额为0",
			refundAmount:      0.00,
			refundNum:         5,
			expectedAvgRefund: 0.00,
		},
		{
			name:              "保留2位小数",
			refundAmount:      100.00,
			refundNum:         3,
			expectedAvgRefund: 33.33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var avgRefundAmount float64
			if tt.refundNum > 0 {
				avgRefundAmount = float64(int(tt.refundAmount/float64(tt.refundNum)*100+0.5)) / 100
			}

			if avgRefundAmount != tt.expectedAvgRefund {
				t.Errorf("期望 AvgRefundAmount=%.2f, 实际 AvgRefundAmount=%.2f", tt.expectedAvgRefund, avgRefundAmount)
			}
		})
	}
}

// TestDeliveryAmountCalculation 测试外卖业绩计算（包含外卖平台订单）
func TestDeliveryAmountCalculation(t *testing.T) {
	tests := []struct {
		name                       string
		takeoutOrderAmount         float64 // 会员外送金额
		instantOrderTakeawayAmount float64 // 第三方外卖金额（系统内）
		externalTakeoutAmount      float64 // 外卖平台金额（Grab/LINE MAN）
		takeoutOrderNum            int64   // 会员外送订单数
		instantOrderTakeawayNum    int64   // 第三方外卖订单数（系统内）
		externalTakeoutNum         int64   // 外卖平台订单数（Grab/LINE MAN）
		expectedDeliveryAmount     float64
		expectedDeliveryOrderNum   int64
	}{
		{
			name:                       "仅会员外送",
			takeoutOrderAmount:         1000.00,
			instantOrderTakeawayAmount: 0.00,
			externalTakeoutAmount:      0.00,
			takeoutOrderNum:            10,
			instantOrderTakeawayNum:    0,
			externalTakeoutNum:         0,
			expectedDeliveryAmount:     1000.00,
			expectedDeliveryOrderNum:   10,
		},
		{
			name:                       "仅第三方外卖（系统内）",
			takeoutOrderAmount:         0.00,
			instantOrderTakeawayAmount: 500.00,
			externalTakeoutAmount:      0.00,
			takeoutOrderNum:            0,
			instantOrderTakeawayNum:    5,
			externalTakeoutNum:         0,
			expectedDeliveryAmount:     500.00,
			expectedDeliveryOrderNum:   5,
		},
		{
			name:                       "仅外卖平台（Grab/LINE MAN）",
			takeoutOrderAmount:         0.00,
			instantOrderTakeawayAmount: 0.00,
			externalTakeoutAmount:      800.00,
			takeoutOrderNum:            0,
			instantOrderTakeawayNum:    0,
			externalTakeoutNum:         8,
			expectedDeliveryAmount:     800.00,
			expectedDeliveryOrderNum:   8,
		},
		{
			name:                       "综合外卖业绩",
			takeoutOrderAmount:         1000.00,
			instantOrderTakeawayAmount: 500.00,
			externalTakeoutAmount:      800.00,
			takeoutOrderNum:            10,
			instantOrderTakeawayNum:    5,
			externalTakeoutNum:         8,
			expectedDeliveryAmount:     2300.00, // 1000 + 500 + 800
			expectedDeliveryOrderNum:   23,      // 10 + 5 + 8
		},
		{
			name:                       "精度测试",
			takeoutOrderAmount:         100.33,
			instantOrderTakeawayAmount: 200.44,
			externalTakeoutAmount:      300.55,
			takeoutOrderNum:            1,
			instantOrderTakeawayNum:    2,
			externalTakeoutNum:         3,
			expectedDeliveryAmount:     601.32, // 保留2位小数
			expectedDeliveryOrderNum:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 business.go 中的计算逻辑
			deliveryAmount := float64(int((tt.takeoutOrderAmount+tt.instantOrderTakeawayAmount+tt.externalTakeoutAmount)*100+0.5)) / 100
			deliveryOrderNum := tt.takeoutOrderNum + tt.instantOrderTakeawayNum + tt.externalTakeoutNum

			if deliveryAmount != tt.expectedDeliveryAmount {
				t.Errorf("DeliveryAmount: 期望 %.2f, 实际 %.2f", tt.expectedDeliveryAmount, deliveryAmount)
			}
			if deliveryOrderNum != tt.expectedDeliveryOrderNum {
				t.Errorf("DeliveryOrderNum: 期望 %d, 实际 %d", tt.expectedDeliveryOrderNum, deliveryOrderNum)
			}
		})
	}
}

// TestInStoreAmountCalculation 测试到店业绩计算
func TestInStoreAmountCalculation(t *testing.T) {
	tests := []struct {
		name                    string
		deskOrderAmount         float64 // 桌台订单金额
		instantOrderAmount      float64 // 店内点餐金额
		deskNum                 int64   // 桌台数
		instantOrderNum         int64   // 店内点餐订单数
		expectedInStoreAmount   float64
		expectedInStoreOrderNum int64
	}{
		{
			name:                    "仅桌台订单",
			deskOrderAmount:         1000.00,
			instantOrderAmount:      0.00,
			deskNum:                 10,
			instantOrderNum:         0,
			expectedInStoreAmount:   1000.00,
			expectedInStoreOrderNum: 10,
		},
		{
			name:                    "仅店内点餐",
			deskOrderAmount:         0.00,
			instantOrderAmount:      500.00,
			deskNum:                 0,
			instantOrderNum:         5,
			expectedInStoreAmount:   500.00,
			expectedInStoreOrderNum: 5,
		},
		{
			name:                    "综合到店业绩",
			deskOrderAmount:         1000.00,
			instantOrderAmount:      500.00,
			deskNum:                 10,
			instantOrderNum:         5,
			expectedInStoreAmount:   1500.00,
			expectedInStoreOrderNum: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 business.go 中的计算逻辑
			inStoreAmount := float64(int((tt.deskOrderAmount+tt.instantOrderAmount)*100+0.5)) / 100
			inStoreOrderNum := tt.deskNum + tt.instantOrderNum

			if inStoreAmount != tt.expectedInStoreAmount {
				t.Errorf("InStoreAmount: 期望 %.2f, 实际 %.2f", tt.expectedInStoreAmount, inStoreAmount)
			}
			if inStoreOrderNum != tt.expectedInStoreOrderNum {
				t.Errorf("InStoreOrderNum: 期望 %d, 实际 %d", tt.expectedInStoreOrderNum, inStoreOrderNum)
			}
		})
	}
}

// TestCalculatePaymentMethodAverage_Normal 测试支付方式平均值计算 - 正常数据
func TestCalculatePaymentMethodAverage_Normal(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{
		{
			PaymentAmount: 1000.00,
			PaymentNum:    10,
			PaymentRatio:  50.00,
		},
		{
			PaymentAmount: 2000.00,
			PaymentNum:    20,
			PaymentRatio:  50.00,
		},
	}

	avg := calculatePaymentMethodAverage(items)

	if avg.PaymentAmount != 1500.00 {
		t.Errorf("PaymentAmount: 期望 1500.00, 实际 %.2f", avg.PaymentAmount)
	}
	if avg.PaymentNum != 15.00 {
		t.Errorf("PaymentNum: 期望 15.00, 实际 %.2f", avg.PaymentNum)
	}
}

// TestCalculatePaymentMethodAverage_Empty 测试支付方式平均值计算 - 空数据返回零值
func TestCalculatePaymentMethodAverage_Empty(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{}

	avg := calculatePaymentMethodAverage(items)

	if avg.PaymentAmount != 0 {
		t.Errorf("PaymentAmount: 期望 0, 实际 %.2f", avg.PaymentAmount)
	}
	if avg.PaymentNum != 0 {
		t.Errorf("PaymentNum: 期望 0, 实际 %.2f", avg.PaymentNum)
	}
}
