package grab

import (
	"strconv"
	"testing"
	"time"
)

// TestNewSignatureVerifier 测试签名验证器创建
func TestNewSignatureVerifier(t *testing.T) {
	secretKey := "test-secret-key"
	verifier := newSignatureVerifier(secretKey)

	if verifier == nil {
		t.Fatal("newSignatureVerifier returned nil")
	}
	if verifier.secretKey != secretKey {
		t.Errorf("secretKey mismatch: got %s, want %s", verifier.secretKey, secretKey)
	}
}

// TestVerifySignature_Success 测试签名验证成功场景
func TestVerifySignature_Success(t *testing.T) {
	secretKey := "my-webhook-secret"
	verifier := newSignatureVerifier(secretKey)

	// 构造测试数据
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"orderID":"G-123456","merchantID":"M-001"}`)

	// 生成正确签名
	signature := verifier.GenerateSignature(timestamp, body)

	// 验证签名
	err := verifier.VerifySignature(signature, timestamp, body)
	if err != nil {
		t.Errorf("VerifySignature failed: %v", err)
	}
}

// TestVerifySignature_MissingSignature 测试缺少签名的场景
func TestVerifySignature_MissingSignature(t *testing.T) {
	verifier := newSignatureVerifier("secret")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"test":"data"}`)

	err := verifier.VerifySignature("", timestamp, body)
	if err == nil {
		t.Error("Expected error for missing signature, got nil")
	}
	if !containsError(err, ErrMissingHeader) {
		t.Errorf("Expected ErrMissingHeader, got: %v", err)
	}
}

// TestVerifySignature_MissingTimestamp 测试缺少时间戳的场景
func TestVerifySignature_MissingTimestamp(t *testing.T) {
	verifier := newSignatureVerifier("secret")
	body := []byte(`{"test":"data"}`)

	err := verifier.VerifySignature("some-signature", "", body)
	if err == nil {
		t.Error("Expected error for missing timestamp, got nil")
	}
	if !containsError(err, ErrMissingHeader) {
		t.Errorf("Expected ErrMissingHeader, got: %v", err)
	}
}

// TestVerifySignature_InvalidTimestamp 测试无效时间戳格式
func TestVerifySignature_InvalidTimestamp(t *testing.T) {
	verifier := newSignatureVerifier("secret")
	body := []byte(`{"test":"data"}`)

	err := verifier.VerifySignature("some-signature", "not-a-number", body)
	if err == nil {
		t.Error("Expected error for invalid timestamp, got nil")
	}
}

// TestVerifySignature_ExpiredTimestamp 测试过期时间戳
func TestVerifySignature_ExpiredTimestamp(t *testing.T) {
	verifier := newSignatureVerifier("secret")

	// 使用 10 分钟前的时间戳 (超过 5 分钟有效窗口)
	expiredTime := time.Now().Add(-10 * time.Minute).Unix()
	timestamp := strconv.FormatInt(expiredTime, 10)
	body := []byte(`{"test":"data"}`)
	signature := verifier.GenerateSignature(timestamp, body)

	err := verifier.VerifySignature(signature, timestamp, body)
	if err == nil {
		t.Error("Expected error for expired timestamp, got nil")
	}
	if err != ErrExpiredTimestamp {
		t.Errorf("Expected ErrExpiredTimestamp, got: %v", err)
	}
}

// TestVerifySignature_InvalidSignature 测试无效签名
func TestVerifySignature_InvalidSignature(t *testing.T) {
	verifier := newSignatureVerifier("secret")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"test":"data"}`)

	// 使用错误的签名
	err := verifier.VerifySignature("wrong-signature", timestamp, body)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
	if err != ErrInvalidSignature {
		t.Errorf("Expected ErrInvalidSignature, got: %v", err)
	}
}

// TestVerifySignature_TamperedBody 测试被篡改的请求体
func TestVerifySignature_TamperedBody(t *testing.T) {
	verifier := newSignatureVerifier("secret")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	originalBody := []byte(`{"orderID":"G-123"}`)
	tamperedBody := []byte(`{"orderID":"G-456"}`)

	// 使用原始数据生成签名
	signature := verifier.GenerateSignature(timestamp, originalBody)

	// 用篡改后的数据验证
	err := verifier.VerifySignature(signature, timestamp, tamperedBody)
	if err == nil {
		t.Error("Expected error for tampered body, got nil")
	}
	if err != ErrInvalidSignature {
		t.Errorf("Expected ErrInvalidSignature, got: %v", err)
	}
}

// TestGenerateSignature_Consistency 测试签名生成的一致性
func TestGenerateSignature_Consistency(t *testing.T) {
	verifier := newSignatureVerifier("consistent-secret")
	timestamp := "1234567890"
	body := []byte(`{"key":"value"}`)

	sig1 := verifier.GenerateSignature(timestamp, body)
	sig2 := verifier.GenerateSignature(timestamp, body)

	if sig1 != sig2 {
		t.Errorf("Signature not consistent: %s != %s", sig1, sig2)
	}
}

// TestGenerateSignature_DifferentSecrets 测试不同密钥生成不同签名
func TestGenerateSignature_DifferentSecrets(t *testing.T) {
	verifier1 := newSignatureVerifier("secret-1")
	verifier2 := newSignatureVerifier("secret-2")

	timestamp := "1234567890"
	body := []byte(`{"key":"value"}`)

	sig1 := verifier1.GenerateSignature(timestamp, body)
	sig2 := verifier2.GenerateSignature(timestamp, body)

	if sig1 == sig2 {
		t.Error("Different secrets should produce different signatures")
	}
}

// TestComputeHMAC_Format 测试 HMAC 计算结果格式
func TestComputeHMAC_Format(t *testing.T) {
	verifier := newSignatureVerifier("test-key")
	result := verifier.computeHMAC("test-message")

	// HMAC-SHA256 输出为 64 个十六进制字符
	if len(result) != 64 {
		t.Errorf("HMAC result length should be 64, got %d", len(result))
	}

	// 检查是否为有效的十六进制字符串
	for _, c := range result {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Invalid hex character in HMAC result: %c", c)
		}
	}
}

// containsError 检查错误是否包含目标错误
func containsError(err, target error) bool {
	if err == nil {
		return target == nil
	}
	return err.Error() == target.Error() || (len(err.Error()) > len(target.Error()) && err.Error()[:len(target.Error())] == target.Error())
}

// BenchmarkVerifySignature 性能测试
func BenchmarkVerifySignature(b *testing.B) {
	verifier := newSignatureVerifier("benchmark-secret")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"orderID":"G-123456","merchantID":"M-001","items":[{"id":"item1","quantity":2}]}`)
	signature := verifier.GenerateSignature(timestamp, body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.VerifySignature(signature, timestamp, body)
	}
}

// BenchmarkGenerateSignature 签名生成性能测试
func BenchmarkGenerateSignature(b *testing.B) {
	verifier := newSignatureVerifier("benchmark-secret")
	timestamp := "1234567890"
	body := []byte(`{"orderID":"G-123456","merchantID":"M-001","items":[{"id":"item1","quantity":2}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.GenerateSignature(timestamp, body)
	}
}
