// Package grab 提供 Grab SDK 客户端封装
package grab

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

var (
	// ErrInvalidSignature 签名无效
	ErrInvalidSignature = gerror.New("签名无效")
	// ErrExpiredTimestamp 时间戳过期
	ErrExpiredTimestamp = gerror.New("时间戳已过期")
	// ErrMissingHeader 缺少必要的请求头
	ErrMissingHeader = gerror.New("缺少必需的请求头")
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
// 生产环境通过 client.Default().GetVerifier() 获取
func newSignatureVerifier(secretKey string) *SignatureVerifier {
	return &SignatureVerifier{
		secretKey: secretKey,
	}
}

// NewSignatureVerifier 创建签名验证器 (公开方法，用于测试)
func NewSignatureVerifier(secretKey string) *SignatureVerifier {
	return newSignatureVerifier(secretKey)
}

// VerifySignature 验证 Grab Webhook 签名
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (v *SignatureVerifier) VerifySignature(signature, timestamp string, body []byte) error {
	// 1. 验证必要参数
	if signature == "" {
		return gerror.Wrapf(ErrMissingHeader, "%s", HeaderXGrabSignature)
	}
	if timestamp == "" {
		return gerror.Wrapf(ErrMissingHeader, "%s", HeaderXGrabTimestamp)
	}

	// 2. 验证时间戳有效性
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return gerror.Wrapf(err, "时间戳格式无效")
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
