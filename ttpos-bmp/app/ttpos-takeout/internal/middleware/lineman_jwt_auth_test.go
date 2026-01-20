package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// 测试用密钥
const testLinemanSecretKey = "test-lineman-secret-key-for-unit-test"

// generateTestLinemanToken 生成测试用 JWT Token
func generateTestLinemanToken(claims *lineman.PartnerTokenClaims, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// setupTestLinemanSecretKey 设置测试用密钥
func setupTestLinemanSecretKey() {
	linemanMu.Lock()
	defer linemanMu.Unlock()
	linemanSecretKey = testLinemanSecretKey
}

// resetLinemanSecretKey 重置密钥
func resetLinemanSecretKey() {
	linemanMu.Lock()
	defer linemanMu.Unlock()
	linemanSecretKey = ""
}

func TestParseLinemanToken_EmptyToken(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	_, err := parseLinemanToken("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 不能为空")
}

func TestParseLinemanToken_InvalidFormat(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	_, err := parseLinemanToken("invalid-token-without-dots")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 格式错误")
}

func TestParseLinemanToken_InvalidFormat_TwoParts(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	_, err := parseLinemanToken("part1.part2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 格式错误")
}

func TestParseLinemanToken_InvalidSignature(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	// 使用不同的密钥签名
	claims := &lineman.PartnerTokenClaims{
		ClientID:    "test-client",
		Scope:       "read write",
		PartnerCode: "TEST001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := generateTestLinemanToken(claims, "wrong-secret-key")
	assert.NoError(t, err)

	_, err = parseLinemanToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 解析失败")
}

func TestParseLinemanToken_ExpiredToken(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	// 生成过期的 token
	claims := &lineman.PartnerTokenClaims{
		ClientID:    "test-client",
		Scope:       "read write",
		PartnerCode: "TEST001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 已过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token, err := generateTestLinemanToken(claims, testLinemanSecretKey)
	assert.NoError(t, err)

	_, err = parseLinemanToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 解析失败")
}

func TestParseLinemanToken_ValidToken(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	// 生成有效的 token
	expectedClaims := &lineman.PartnerTokenClaims{
		ClientID:    "test-client-id",
		Scope:       "read write",
		PartnerCode: "PARTNER001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := generateTestLinemanToken(expectedClaims, testLinemanSecretKey)
	assert.NoError(t, err)

	claims, err := parseLinemanToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, expectedClaims.ClientID, claims.ClientID)
	assert.Equal(t, expectedClaims.Scope, claims.Scope)
	assert.Equal(t, expectedClaims.PartnerCode, claims.PartnerCode)
}

func TestParseLinemanToken_WrongSigningMethod(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	// 使用 HS384 签名方法（不支持，只支持 HS256）
	claims := &lineman.PartnerTokenClaims{
		ClientID:    "test-client",
		Scope:       "read",
		PartnerCode: "TEST001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenStr, err := token.SignedString([]byte(testLinemanSecretKey))
	assert.NoError(t, err)

	_, err = parseLinemanToken(tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的签名方法")
}

func TestParseLinemanToken_WhitespaceToken(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	_, err := parseLinemanToken("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token 不能为空")
}

func TestParseLinemanToken_TokenWithWhitespace(t *testing.T) {
	setupTestLinemanSecretKey()
	defer resetLinemanSecretKey()

	// 生成有效的 token
	claims := &lineman.PartnerTokenClaims{
		ClientID:    "test-client",
		Scope:       "read",
		PartnerCode: "TEST001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := generateTestLinemanToken(claims, testLinemanSecretKey)
	assert.NoError(t, err)

	// 添加前后空白
	parsedClaims, err := parseLinemanToken("  " + token + "  ")
	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
	assert.Equal(t, "test-client", parsedClaims.ClientID)
}

func TestGetTokenPreview(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"empty", "", ""},
		{"short", "abc", "abc"},
		{"exact 20", "12345678901234567890", "12345678901234567890"},
		{"long", "12345678901234567890abcdef", "12345678901234567890"},
		{"19 chars", "1234567890123456789", "1234567890123456789"},
		{"21 chars", "123456789012345678901", "12345678901234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTokenPreview(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextKeyLinemanPartnerClaims(t *testing.T) {
	// 验证常量值
	assert.Equal(t, "LinemanPartnerClaims", ContextKeyLinemanPartnerClaims)
}
