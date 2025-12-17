// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	// ErrInvalidSignature 签名无效
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrExpiredTimestamp 时间戳过期
	ErrExpiredTimestamp = errors.New("timestamp expired")
	// ErrMissingHeader 缺少必要的请求头
	ErrMissingHeader = errors.New("missing required header")
)

const (
	// SignatureValidityWindow 签名有效窗口时间 (5分钟)
	SignatureValidityWindow = 5 * time.Minute
	// HeaderXGrabSignature Grab 签名请求头
	HeaderXGrabSignature = "X-Grab-Signature"
	// HeaderXGrabTimestamp Grab 时间戳请求头
	HeaderXGrabTimestamp = "X-Grab-Timestamp"
)

// SignatureVerifier 签名验证器
// 用于验证 Grab Webhook 请求的 HMAC-SHA256 签名
type SignatureVerifier struct {
	secretKey string
}

// newSignatureVerifier 创建签名验证器 (内部使用)
// 生产环境通过 service.Grab().getVerifier() 获取
func newSignatureVerifier(secretKey string) *SignatureVerifier {
	return &SignatureVerifier{
		secretKey: secretKey,
	}
}

// VerifySignature 验证 Grab Webhook 签名
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (v *SignatureVerifier) VerifySignature(signature, timestamp string, body []byte) error {
	// 1. 验证必要参数
	if signature == "" {
		return fmt.Errorf("%w: %s", ErrMissingHeader, HeaderXGrabSignature)
	}
	if timestamp == "" {
		return fmt.Errorf("%w: %s", ErrMissingHeader, HeaderXGrabTimestamp)
	}

	// 2. 验证时间戳有效性
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}
	requestTime := time.Unix(ts, 0)
	if time.Since(requestTime) > SignatureValidityWindow {
		return ErrExpiredTimestamp
	}

	// 3. 计算预期签名
	// Grab 签名格式: HMAC-SHA256(timestamp + "." + body)
	message := fmt.Sprintf("%s.%s", timestamp, string(body))
	expectedSignature := v.computeHMAC(message)

	// 4. 安全比较签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return ErrInvalidSignature
	}

	return nil
}

// computeHMAC 计算 HMAC-SHA256 签名
func (v *SignatureVerifier) computeHMAC(message string) string {
	mac := hmac.New(sha256.New, []byte(v.secretKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateSignature 生成签名 (用于测试或向 Grab 发请求)
func (v *SignatureVerifier) GenerateSignature(timestamp string, body []byte) string {
	message := fmt.Sprintf("%s.%s", timestamp, string(body))
	return v.computeHMAC(message)
}
