// Package util 提供辅助工具函数
package util

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

// GetCurrentTimestamp 获取当前时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTimestampMillis 获取当前时间戳（毫秒）
func GetCurrentTimestampMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TimestampToTime 时间戳转换为 time.Time
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FormatTimestamp 格式化时间戳为字符串
func FormatTimestamp(timestamp int64, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return time.Unix(timestamp, 0).Format(layout)
}

// MD5Hash 计算字符串的 MD5 哈希值
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GenerateMessageID 生成消息ID
// 使用时间戳和随机字符串组合
func GenerateMessageID(prefix string) string {
	if prefix == "" {
		prefix = "msg"
	}
	now := time.Now()
	timestampStr := now.Format("20060102150405")
	hashInput := timestampStr + "-" + now.Format("000000000")
	return prefix + "-" + timestampStr + "-" + MD5Hash(hashInput)[:8]
}

// IsEmptyString 检查字符串是否为空
func IsEmptyString(s string) bool {
	return len(s) == 0
}

// StringInSlice 检查字符串是否在切片中
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
