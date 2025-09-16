package utils

import (
	"testing"
)

func TestIsValidInternalCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// 有效的情况
		{
			name:     "单个数字",
			input:    "1",
			expected: true,
		},
		{
			name:     "单个小写字母",
			input:    "a",
			expected: true,
		},
		{
			name:     "单个大写字母",
			input:    "A",
			expected: true,
		},
		{
			name:     "纯数字",
			input:    "123456789",
			expected: true,
		},
		{
			name:     "纯小写字母",
			input:    "abcdefgh",
			expected: true,
		},
		{
			name:     "纯大写字母",
			input:    "ABCDEFGH",
			expected: true,
		},
		{
			name:     "数字和小写字母混合",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "数字和大写字母混合",
			input:    "ABC123",
			expected: true,
		},
		{
			name:     "大小写字母混合",
			input:    "AbC",
			expected: true,
		},
		{
			name:     "数字和大小写字母混合",
			input:    "A1b2C3",
			expected: true,
		},
		{
			name:     "最大长度13位纯数字",
			input:    "1234567890123",
			expected: true,
		},
		{
			name:     "最大长度13位纯字母",
			input:    "abcdefghijklm",
			expected: true,
		},
		{
			name:     "最大长度13位混合",
			input:    "A1b2C3d4E5f6G",
			expected: true,
		},
		{
			name:     "边界长度1位",
			input:    "Z",
			expected: true,
		},
		{
			name:     "边界长度13位",
			input:    "1234567890ABC",
			expected: true,
		},

		// 无效的情况
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
		{
			name:     "长度超过13位",
			input:    "12345678901234",
			expected: false,
		},
		{
			name:     "包含特殊字符",
			input:    "abc@123",
			expected: false,
		},
		{
			name:     "包含空格",
			input:    "abc 123",
			expected: false,
		},
		{
			name:     "包含下划线",
			input:    "abc_123",
			expected: false,
		},
		{
			name:     "包含连字符",
			input:    "abc-123",
			expected: false,
		},
		{
			name:     "包含点号",
			input:    "abc.123",
			expected: false,
		},
		{
			name:     "包含中文字符",
			input:    "abc中文123",
			expected: false,
		},
		{
			name:     "包含表情符号",
			input:    "abc😀123",
			expected: false,
		},
		{
			name:     "包含制表符",
			input:    "abc\t123",
			expected: false,
		},
		{
			name:     "包含换行符",
			input:    "abc\n123",
			expected: false,
		},
		{
			name:     "包含回车符",
			input:    "abc\r123",
			expected: false,
		},
		{
			name:     "包含其他特殊字符",
			input:    "abc!123",
			expected: false,
		},
		{
			name:     "包含井号",
			input:    "abc#123",
			expected: false,
		},
		{
			name:     "包含美元符号",
			input:    "abc$123",
			expected: false,
		},
		{
			name:     "包含百分号",
			input:    "abc%123",
			expected: false,
		},
		{
			name:     "包含与号",
			input:    "abc&123",
			expected: false,
		},
		{
			name:     "包含星号",
			input:    "abc*123",
			expected: false,
		},
		{
			name:     "包含加号",
			input:    "abc+123",
			expected: false,
		},
		{
			name:     "包含等号",
			input:    "abc=123",
			expected: false,
		},
		{
			name:     "包含问号",
			input:    "abc?123",
			expected: false,
		},
		{
			name:     "包含斜杠",
			input:    "abc/123",
			expected: false,
		},
		{
			name:     "包含反斜杠",
			input:    "abc\\123",
			expected: false,
		},
		{
			name:     "包含竖线",
			input:    "abc|123",
			expected: false,
		},
		{
			name:     "包含波浪号",
			input:    "abc~123",
			expected: false,
		},
		{
			name:     "包含反引号",
			input:    "abc`123",
			expected: false,
		},
		{
			name:     "包含方括号",
			input:    "abc[123]",
			expected: false,
		},
		{
			name:     "包含花括号",
			input:    "abc{123}",
			expected: false,
		},
		{
			name:     "包含圆括号",
			input:    "abc(123)",
			expected: false,
		},
		{
			name:     "包含尖括号",
			input:    "abc<123>",
			expected: false,
		},
		{
			name:     "包含冒号",
			input:    "abc:123",
			expected: false,
		},
		{
			name:     "包含分号",
			input:    "abc;123",
			expected: false,
		},
		{
			name:     "包含单引号",
			input:    "abc'123",
			expected: false,
		},
		{
			name:     "包含双引号",
			input:    "abc\"123",
			expected: false,
		},
		{
			name:     "包含逗号",
			input:    "abc,123",
			expected: false,
		},
		{
			name:     "包含句号",
			input:    "abc.123",
			expected: false,
		},
		{
			name:     "包含小于号",
			input:    "abc<123",
			expected: false,
		},
		{
			name:     "包含大于号",
			input:    "abc>123",
			expected: false,
		},
		{
			name:     "包含数字0",
			input:    "0",
			expected: true,
		},
		{
			name:     "包含数字9",
			input:    "9",
			expected: true,
		},
		{
			name:     "包含字母a",
			input:    "a",
			expected: true,
		},
		{
			name:     "包含字母z",
			input:    "z",
			expected: true,
		},
		{
			name:     "包含字母A",
			input:    "A",
			expected: true,
		},
		{
			name:     "包含字母Z",
			input:    "Z",
			expected: true,
		},
		{
			name:     "边界测试：14位数字",
			input:    "12345678901234",
			expected: false,
		},
		{
			name:     "边界测试：14位字母",
			input:    "abcdefghijklmn",
			expected: false,
		},
		{
			name:     "边界测试：14位混合",
			input:    "A1b2C3d4E5f6G7",
			expected: false,
		},
		{
			name:     "实际业务场景：商品编码",
			input:    "SP001",
			expected: true,
		},
		{
			name:     "实际业务场景：原料编码",
			input:    "YL2024001",
			expected: true,
		},
		{
			name:     "实际业务场景：内部编码",
			input:    "INT001ABC",
			expected: true,
		},
		{
			name:     "实际业务场景：批次编码",
			input:    "BATCH2024A",
			expected: true,
		},
		{
			name:     "实际业务场景：序列号",
			input:    "SN123456789",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidInternalCode(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidInternalCode(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsValidInternalCode_EdgeCases 测试边界情况
func TestIsValidInternalCode_EdgeCases(t *testing.T) {
	// 测试各种边界长度
	lengthTests := []struct {
		length   int
		expected bool
	}{
		{0, false},  // 空字符串
		{1, true},   // 最小长度
		{13, true},  // 最大长度
		{14, false}, // 超过最大长度
	}

	for _, lt := range lengthTests {
		var input string
		if lt.length == 0 {
			input = ""
		} else {
			// 生成指定长度的有效字符串
			input = ""
			for i := 0; i < lt.length; i++ {
				if i%3 == 0 {
					input += "1" // 数字
				} else if i%3 == 1 {
					input += "a" // 小写字母
				} else {
					input += "A" // 大写字母
				}
			}
		}

		result := IsValidInternalCode(input)
		if result != lt.expected {
			t.Errorf("IsValidInternalCode(length=%d) = %v, expected %v, input=%q", lt.length, result, lt.expected, input)
		}
	}
}

// TestIsValidInternalCode_CharacterRanges 测试字符范围
func TestIsValidInternalCode_CharacterRanges(t *testing.T) {
	// 测试数字范围 0-9
	for i := 0; i <= 9; i++ {
		input := string(rune('0' + i))
		result := IsValidInternalCode(input)
		if !result {
			t.Errorf("IsValidInternalCode(%q) = false, expected true (数字 %d)", input, i)
		}
	}

	// 测试小写字母范围 a-z
	for i := 0; i < 26; i++ {
		input := string(rune('a' + i))
		result := IsValidInternalCode(input)
		if !result {
			t.Errorf("IsValidInternalCode(%q) = false, expected true (小写字母 %c)", input, 'a'+i)
		}
	}

	// 测试大写字母范围 A-Z
	for i := 0; i < 26; i++ {
		input := string(rune('A' + i))
		result := IsValidInternalCode(input)
		if !result {
			t.Errorf("IsValidInternalCode(%q) = false, expected true (大写字母 %c)", input, 'A'+i)
		}
	}

	// 测试边界外的字符
	invalidChars := []rune{'0' - 1, '9' + 1, 'a' - 1, 'z' + 1, 'A' - 1, 'Z' + 1}
	for _, char := range invalidChars {
		input := string(char)
		result := IsValidInternalCode(input)
		if result {
			t.Errorf("IsValidInternalCode(%q) = true, expected false (无效字符 %c)", input, char)
		}
	}
}

// TestIsValidInternalCode_Performance 性能测试
func TestIsValidInternalCode_Performance(t *testing.T) {
	// 测试长字符串的性能
	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "A"
	}

	// 这个应该很快返回false，因为长度超过13
	result := IsValidInternalCode(longString)
	if result {
		t.Errorf("IsValidInternalCode(longString) = true, expected false")
	}
}

// BenchmarkIsValidInternalCode 基准测试
func BenchmarkIsValidInternalCode(b *testing.B) {
	testCases := []string{
		"",               // 空字符串
		"A",              // 单个字符
		"ABC123",         // 混合字符
		"1234567890123",  // 最大长度
		"12345678901234", // 超长
		"abc@123",        // 包含特殊字符
		"abcdefghijklm",  // 13位纯字母
		"1234567890ABC",  // 13位混合
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsValidInternalCode(tc)
			}
		})
	}
}
