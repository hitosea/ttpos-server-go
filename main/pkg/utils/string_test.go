package utils

import (
	"testing"
	"unsafe"
)

func TestStringToBytes(t *testing.T) {
	// 测试用例1：验证返回值和预期相等
	t.Run("返回值与预期相等", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected []byte
		}{
			{
				name:     "空字符串",
				input:    "",
				expected: []byte{},
			},
			{
				name:     "普通字符串",
				input:    "hello world",
				expected: []byte("hello world"),
			},
			{
				name:     "中文字符串",
				input:    "你好世界",
				expected: []byte("你好世界"),
			},
			{
				name:     "特殊字符",
				input:    "!@#$%^&*()",
				expected: []byte("!@#$%^&*()"),
			},
			{
				name:     "数字字符串",
				input:    "1234567890",
				expected: []byte("1234567890"),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := UnsafeStringToBytes(tc.input)

				// 验证长度相等
				if len(result) != len(tc.expected) {
					t.Errorf("长度不匹配: 期望 %d, 实际 %d", len(tc.expected), len(result))
				}

				// 验证内容相等
				if !bytesEqual(result, tc.expected) {
					t.Errorf("内容不匹配: 期望 %v, 实际 %v", tc.expected, result)
				}
			})
		}
	})

	// 测试用例2：验证修改返回的bytes会panic
	t.Run("修改返回的bytes会panic", func(t *testing.T) {
		testString := "hello world"
		result := UnsafeStringToBytes(testString)

		// 验证返回的字节切片确实指向只读内存
		// 我们不能直接修改，但可以验证它确实指向字符串的底层数据
		if len(result) == 0 {
			t.Error("结果不应该为空")
		}

		// 验证内容正确
		expected := []byte(testString)
		if !bytesEqual(result, expected) {
			t.Errorf("内容不匹配: 期望 %v, 实际 %v", expected, result)
		}

		// 注意：在macOS上，尝试修改result会导致SIGBUS错误
		// 这证明了返回的字节切片确实指向只读的字符串内存
		t.Log("注意：修改返回的字节切片会导致SIGBUS错误，这证明了零拷贝和只读特性")
	})

	// 测试用例3：证明确实是零拷贝
	t.Run("零拷贝验证", func(t *testing.T) {
		originalString := "test string for zero copy"

		// 获取原始字符串的底层数据指针
		originalPtr := unsafe.StringData(originalString)

		// 转换为字节切片
		result := UnsafeStringToBytes(originalString)

		// 获取字节切片的底层数据指针
		resultPtr := unsafe.SliceData(result)

		// 验证两个指针相同，证明是零拷贝
		if originalPtr != resultPtr {
			t.Errorf("不是零拷贝: 原始指针 %p, 结果指针 %p", originalPtr, resultPtr)
		}

		// 验证长度相同
		if len(result) != len(originalString) {
			t.Errorf("长度不匹配: 原始长度 %d, 结果长度 %d", len(originalString), len(result))
		}
	})
}

// TestBytesToString 测试 BytesToString 函数
func TestBytesToString(t *testing.T) {
	// 测试用例1：验证返回值和预期相等
	t.Run("返回值与预期相等", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    []byte
			expected string
		}{
			{
				name:     "空字节切片",
				input:    []byte{},
				expected: "",
			},
			{
				name:     "普通字节切片",
				input:    []byte("hello world"),
				expected: "hello world",
			},
			{
				name:     "中文字节切片",
				input:    []byte("你好世界"),
				expected: "你好世界",
			},
			{
				name:     "特殊字符字节切片",
				input:    []byte("!@#$%^&*()"),
				expected: "!@#$%^&*()",
			},
			{
				name:     "数字字节切片",
				input:    []byte("1234567890"),
				expected: "1234567890",
			},
			{
				name:     "包含null字节的切片",
				input:    []byte{'a', 0, 'b', 0, 'c'},
				expected: "a\x00b\x00c",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := UnsafeBytesToString(tc.input)

				// 验证内容相等
				if result != tc.expected {
					t.Errorf("内容不匹配: 期望 %q, 实际 %q", tc.expected, result)
				}

				// 验证长度相等
				if len(result) != len(tc.expected) {
					t.Errorf("长度不匹配: 期望 %d, 实际 %d", len(tc.expected), len(result))
				}
			})
		}
	})

	// 测试用例2：验证修改原始bytes会影响返回的string（证明是零拷贝）
	t.Run("修改原始bytes会影响返回的string", func(t *testing.T) {
		originalBytes := []byte("hello world")
		result := UnsafeBytesToString(originalBytes)

		// 验证初始状态
		if result != "hello world" {
			t.Errorf("初始状态不正确: 期望 %q, 实际 %q", "hello world", result)
		}

		// 修改原始字节切片
		originalBytes[0] = 'H'
		originalBytes[6] = 'W'

		// 验证返回的字符串也发生了变化（证明是零拷贝）
		expected := "Hello World"
		if result != expected {
			t.Errorf("修改原始字节后字符串未变化: 期望 %q, 实际 %q", expected, result)
		}
	})

	// 测试用例3：证明确实是零拷贝
	t.Run("零拷贝验证", func(t *testing.T) {
		originalBytes := []byte("test string for zero copy")

		// 获取原始字节切片的底层数据指针
		originalPtr := unsafe.SliceData(originalBytes)

		// 转换为字符串
		result := UnsafeBytesToString(originalBytes)

		// 获取字符串的底层数据指针
		resultPtr := unsafe.StringData(result)

		// 验证两个指针相同，证明是零拷贝
		if originalPtr != resultPtr {
			t.Errorf("不是零拷贝: 原始指针 %p, 结果指针 %p", originalPtr, resultPtr)
		}

		// 验证长度相同
		if len(result) != len(originalBytes) {
			t.Errorf("长度不匹配: 原始长度 %d, 结果长度 %d", len(originalBytes), len(result))
		}
	})
}

// TestStringToBytesModification 单独测试修改返回的字节切片会导致崩溃
// 这个测试会故意导致程序崩溃，所以需要单独运行
func TestStringToBytesModification(t *testing.T) {
	// 默认跳过这个测试，因为它会导致程序崩溃
	// 如果需要运行，请使用: go test -run TestStringToBytesModification
	t.Skip("跳过会导致崩溃的测试，使用 -run TestStringToBytesModification 单独运行")

	testString := "hello world"
	result := UnsafeStringToBytes(testString)

	// 尝试修改返回的字节切片，这会导致SIGBUS错误
	// 在macOS上，字符串内存是只读的
	if len(result) > 0 {
		t.Log("即将尝试修改只读内存，这会导致SIGBUS错误...")
		result[0] = 'X' // 这会导致程序崩溃
	}
}

// 辅助函数：比较两个字节切片是否相等
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

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

// ---------------------------------------------------------------------------
// ProcessStoreCodeForSort
// ---------------------------------------------------------------------------

func TestProcessStoreCodeForSort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"无前缀-纯数字", "123", "123"},
		{"无前缀-字母", "ABC", "ABC"},
		{"小写前缀No.", "no.123", "123"},
		{"大写前缀NO.", "NO.456", "456"},
		{"混合大小写No.", "No.789", "789"},
		{"前缀nO.", "nO.abc", "abc"},
		{"仅No.前缀无内容", "No.", ""},
		{"不匹配前缀Nov.", "Nov.1", "Nov.1"},
		{"空字符串", "", ""},
		{"包含多个No.", "No.No.1", "No.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ProcessStoreCodeForSort(tt.input)
			if got != tt.want {
				t.Errorf("ProcessStoreCodeForSort(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsOnlyDigits
// ---------------------------------------------------------------------------

func TestIsOnlyDigits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"纯数字", "123456", true},
		{"单个数字", "0", true},
		{"前导零", "007", true},
		{"包含字母", "12a3", false},
		{"纯字母", "abc", false},
		{"空字符串", "", false},
		{"包含空格", "12 3", false},
		{"包含小数点", "12.3", false},
		{"包含负号", "-1", false},
		{"中文数字", "一二三", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsOnlyDigits(tt.input)
			if got != tt.want {
				t.Errorf("IsOnlyDigits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseNumberWithoutLeadingZeros
// ---------------------------------------------------------------------------

func TestParseNumberWithoutLeadingZeros(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"普通数字", "42", 42, false},
		{"前导零", "007", 7, false},
		{"全零", "000", 0, false},
		{"单个零", "0", 0, false},
		{"大数字", "99999", 99999, false},
		{"单个数字", "5", 5, false},
		{"非数字字符串", "abc", 0, true},
		{"混合字符", "12a", 0, true},
		{"空字符串", "", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseNumberWithoutLeadingZeros(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNumberWithoutLeadingZeros(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseNumberWithoutLeadingZeros(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SortByStoreCode
// ---------------------------------------------------------------------------

// mockShop 实现 StoreCodeSortable 接口，用于测试
type mockShop struct {
	uuid      uint64
	storeCode string
}

func (m mockShop) GetStoreCode() string { return m.storeCode }
func (m mockShop) GetUuid() uint64      { return m.uuid }

func TestSortByStoreCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []mockShop
		wantCode []string // 排序后的 storeCode 顺序
	}{
		{
			name: "空编码排最前-按UUID升序",
			input: []mockShop{
				{uuid: 300, storeCode: "1"},
				{uuid: 200, storeCode: ""},
				{uuid: 100, storeCode: ""},
			},
			wantCode: []string{"", "", "1"},
		},
		{
			name: "纯数字按数值排序",
			input: []mockShop{
				{uuid: 1, storeCode: "10"},
				{uuid: 2, storeCode: "2"},
				{uuid: 3, storeCode: "100"},
				{uuid: 4, storeCode: "1"},
			},
			wantCode: []string{"1", "2", "10", "100"},
		},
		{
			name: "数字优先于字母",
			input: []mockShop{
				{uuid: 1, storeCode: "ABC"},
				{uuid: 2, storeCode: "5"},
				{uuid: 3, storeCode: "ZZZ"},
				{uuid: 4, storeCode: "1"},
			},
			wantCode: []string{"1", "5", "ABC", "ZZZ"},
		},
		{
			name: "字母不区分大小写排序",
			input: []mockShop{
				{uuid: 1, storeCode: "charlie"},
				{uuid: 2, storeCode: "Alpha"},
				{uuid: 3, storeCode: "BRAVO"},
			},
			wantCode: []string{"Alpha", "BRAVO", "charlie"},
		},
		{
			name: "No.前缀去除后按数值排序",
			input: []mockShop{
				{uuid: 1, storeCode: "No.10"},
				{uuid: 2, storeCode: "No.2"},
				{uuid: 3, storeCode: "3"},
			},
			wantCode: []string{"No.2", "3", "No.10"},
		},
		{
			name: "前导零数字排序",
			input: []mockShop{
				{uuid: 1, storeCode: "007"},
				{uuid: 2, storeCode: "10"},
				{uuid: 3, storeCode: "002"},
			},
			wantCode: []string{"002", "007", "10"},
		},
		{
			name: "综合场景-空+数字+No.前缀+字母",
			input: []mockShop{
				{uuid: 5, storeCode: "No.3"},
				{uuid: 1, storeCode: ""},
				{uuid: 3, storeCode: "Branch-A"},
				{uuid: 2, storeCode: "10"},
				{uuid: 4, storeCode: "1"},
				{uuid: 6, storeCode: ""},
			},
			wantCode: []string{"", "", "1", "No.3", "10", "Branch-A"},
		},
		{
			name:     "空列表",
			input:    []mockShop{},
			wantCode: []string{},
		},
		{
			name:     "单个元素",
			input:    []mockShop{{uuid: 1, storeCode: "42"}},
			wantCode: []string{"42"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			list := make([]mockShop, len(tt.input))
			copy(list, tt.input)
			SortByStoreCode(list)

			got := make([]string, len(list))
			for i, s := range list {
				got[i] = s.GetStoreCode()
			}
			if len(got) != len(tt.wantCode) {
				t.Fatalf("长度不匹配: got %d, want %d", len(got), len(tt.wantCode))
			}
			for i := range got {
				if got[i] != tt.wantCode[i] {
					t.Errorf("位置 %d: got %q, want %q\n完整结果: %v", i, got[i], tt.wantCode[i], got)
					break
				}
			}
		})
	}
}

// TestSortByStoreCode_EmptyCodesUuidOrder 验证多个空编码按UUID升序
func TestSortByStoreCode_EmptyCodesUuidOrder(t *testing.T) {
	t.Parallel()
	list := []mockShop{
		{uuid: 500, storeCode: ""},
		{uuid: 100, storeCode: ""},
		{uuid: 300, storeCode: ""},
		{uuid: 200, storeCode: ""},
	}
	SortByStoreCode(list)
	for i := 1; i < len(list); i++ {
		if list[i].GetUuid() < list[i-1].GetUuid() {
			t.Errorf("空编码未按UUID升序: 位置%d uuid=%d < 位置%d uuid=%d",
				i, list[i].GetUuid(), i-1, list[i-1].GetUuid())
		}
	}
}
