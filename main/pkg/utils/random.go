package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	// 字符集
	Numbers      = "0123456789"
	LowerLetters = "abcdefghijklmnopqrstuvwxyz"
	UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Symbols      = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

// RandomString 生成指定长度的随机字符串
// chars参数指定字符集，如果为空则使用数字+字母
func RandomString(length int, chars ...string) string {
	charset := strings.Join(chars, "")
	if charset == "" {
		charset = Numbers + LowerLetters + UpperLetters
	}

	b := make([]byte, length)
	maxNum := big.NewInt(int64(len(charset)))

	for i := range b {
		n, err := rand.Int(rand.Reader, maxNum)
		if err != nil {
			// 发生错误时使用索引作为后备方案
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

// RandomNumber 生成指定长度的随机数字字符串
func RandomNumber(length int) string {
	return RandomString(length, Numbers)
}

// RandomLetter 生成指定长度的随机字母字符串
func RandomLetter(length int) string {
	return RandomString(length, LowerLetters, UpperLetters)
}

// GenerateRandomNumber 生成指定长度的随机数字
func GenerateRandomNumber(length int) string {
	return RandomString(length, Numbers)
}

// RandomLowerLetter 生成指定长度的随机小写字母字符串
func RandomLowerLetter(length int) string {
	return RandomString(length, LowerLetters)
}

// RandomUpperLetter 生成指定长度的随机大写字母字符串
func RandomUpperLetter(length int) string {
	return RandomString(length, UpperLetters)
}

// RandomNumberLetter 生成指定长度的随机数字和字母组合
func RandomNumberLetter(length int) string {
	return RandomString(length, Numbers, LowerLetters, UpperLetters)
}

// RandomSafeString 生成指定长度的安全随机字符串（包含特殊字符）
func RandomSafeString(length int) string {
	return RandomString(length, Numbers, LowerLetters, UpperLetters, Symbols)
}

// RandomUUID 生成UUID格式的随机字符串
func RandomUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return RandomString(32)
	}

	// 设置版本 (4) 和变体位
	uuid[6] = (uuid[6] & 0x0F) | 0x40 // 版本 4
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // RFC 4122 变体

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	)
}
