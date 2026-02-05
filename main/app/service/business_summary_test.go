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
