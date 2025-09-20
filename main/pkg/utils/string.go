package utils

import (
	"fmt"
	"math/rand"
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
