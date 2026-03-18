package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// GenerateMerchantOrderNo 生成订单号
func GenerateMerchantOrderNo(prefix string) string {
	datePart := time.Now().Format("20060102150405")
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	return prefix + datePart + randomPart
}

// FormatAmount 格式化金额
func FormatAmount(amount float64) string {
	if amount == 0 {
		return "0"
	}
	// 先格式化为2位小数
	formattedAmount := strconv.FormatFloat(amount, 'f', 2, 64)

	// 分割整数和小数部分
	parts := strings.Split(formattedAmount, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// 处理整数部分，添加千分位
	var result strings.Builder
	length := len(integerPart)
	for i := length - 1; i >= 0; i-- {
		result.WriteByte(integerPart[i])
		if i > 0 && (length-i)%3 == 0 {
			result.WriteByte(',')
		}
	}

	// 反转字符串
	reversed := result.String()
	runes := []rune(reversed)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	integerPart = string(runes)

	// 处理小数部分，去除尾部的0
	decimalPart = strings.TrimRight(decimalPart, "0")

	// 组合结果
	if decimalPart != "" {
		return integerPart + "." + decimalPart
	}
	return integerPart
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GetRandomString(length int) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// string → []byte，【注意】不能修改返回的[]byte
func UnsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// []byte → string
func UnsafeBytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// IsValidInternalCode 验证内部编码是否有效
func IsValidInternalCode(internalCode string) bool {
	if len(internalCode) < 1 || len(internalCode) > 13 {
		return false
	}
	for _, char := range internalCode {
		if '0' <= char && char <= '9' {
			continue
		}
		if 'a' <= char && char <= 'z' {
			continue
		}
		if 'A' <= char && char <= 'Z' {
			continue
		}
		return false
	}
	return true
}

// FullReductionResult 满减字符串解析结果
type FullReductionResult struct {
	Type            string  // 类型："满" 或 "每满"
	FullAmount      float64 // 满减的金额阈值（n）
	ReductionAmount float64 // 减免的金额（m）
}

// ParseFullReductionString 解析"满n减m"或"每满n减m"格式的字符串，提取类型和两个浮点数
// 例如：
//   - "满100.00减10.00" -> FullReductionResult{Type: "满", FullAmount: 100.00, ReductionAmount: 10.00}
//   - "每满100.00减10.00" -> FullReductionResult{Type: "每满", FullAmount: 100.00, ReductionAmount: 10.00}
//
// 参数:
//   - s: 待解析的字符串
//
// 返回:
//   - result: 解析结果，包含类型和数值
//   - error: 解析错误，如果格式不匹配或数字解析失败则返回错误
func ParseFullReductionString(s string) (*FullReductionResult, error) {
	if s == "" {
		return nil, fmt.Errorf("输入字符串为空")
	}

	var result FullReductionResult
	var matches []string

	// 优先匹配"每满n减m"格式（循环满减）
	patternEveryFull := `每满([\d.]+)减([\d.]+)`
	reEveryFull := regexp.MustCompile(patternEveryFull)
	matches = reEveryFull.FindStringSubmatch(s)

	if len(matches) == 3 {
		result.Type = "每满"
	} else {
		// 匹配"满n减m"格式（阶梯满减）
		patternFull := `满([\d.]+)减([\d.]+)`
		reFull := regexp.MustCompile(patternFull)
		matches = reFull.FindStringSubmatch(s)

		if len(matches) == 3 {
			result.Type = "满"
		} else {
			return nil, fmt.Errorf("字符串格式不匹配，期望格式：满n减m 或 每满n减m，实际：%s", s)
		}
	}

	// 解析第一个数字（满减阈值）
	fullAmount, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("解析满减阈值失败：%w", err)
	}

	// 解析第二个数字（减免金额）
	reductionAmount, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("解析减免金额失败：%w", err)
	}

	result.FullAmount = fullAmount
	result.ReductionAmount = reductionAmount

	return &result, nil
}

// ProcessStoreCodeForSort 处理 StoreCode 用于排序
// 去掉 "No." 前缀（不区分大小写），返回处理后的字符串
func ProcessStoreCodeForSort(storeCode string) string {
	codeLower := strings.ToLower(storeCode)
	if strings.HasPrefix(codeLower, "no.") {
		return storeCode[3:] // 去掉前3个字符 "No."
	}
	return storeCode
}

// IsOnlyDigits 判断字符串是否只包含数字
func IsOnlyDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ParseNumberWithoutLeadingZeros 去掉前缀0并转换为数字
func ParseNumberWithoutLeadingZeros(s string) (int, error) {
	// 去掉前缀0
	trimmed := strings.TrimLeft(s, "0")
	if trimmed == "" {
		return 0, nil
	}
	return strconv.Atoi(trimmed)
}

// StoreCodeSortable 门店排序所需的接口
type StoreCodeSortable interface {
	GetStoreCode() string
	GetUuid() uint64
}

// SortByStoreCode 按门店编码通用排序
// 规则：空编码排最前（按UUID升序）；非空编码去 "No." 前缀后，纯数字按数值排序，数字优先于字母，字母不区分大小写排序
func SortByStoreCode[T StoreCodeSortable](list []T) {
	sort.Slice(list, func(i, j int) bool {
		code1 := list[i].GetStoreCode()
		code2 := list[j].GetStoreCode()
		isEmpty1 := code1 == ""
		isEmpty2 := code2 == ""

		if isEmpty1 != isEmpty2 {
			return isEmpty1
		}
		if isEmpty1 && isEmpty2 {
			return list[i].GetUuid() < list[j].GetUuid()
		}

		processed1 := ProcessStoreCodeForSort(code1)
		processed2 := ProcessStoreCodeForSort(code2)
		isDigits1 := IsOnlyDigits(processed1)
		isDigits2 := IsOnlyDigits(processed2)

		if isDigits1 && isDigits2 {
			num1, _ := ParseNumberWithoutLeadingZeros(processed1)
			num2, _ := ParseNumberWithoutLeadingZeros(processed2)
			return num1 < num2
		}
		if isDigits1 != isDigits2 {
			return isDigits1
		}
		return strings.ToLower(processed1) < strings.ToLower(processed2)
	})
}
