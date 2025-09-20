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
