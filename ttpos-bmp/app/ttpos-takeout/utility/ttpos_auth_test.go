package utility

import (
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/test/gtest"
)

// setupTestConfig 设置测试配置
func setupTestConfig(content string) {
	adapter, _ := gcfg.NewAdapterContent(content)
	g.Cfg().SetAdapter(adapter)
}

func TestGenerateTtposAuth_Success(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 设置测试配置
		testSecret := "test-secret-123"

		setupTestConfig(`
app:
  callbackSecret: "` + testSecret + `"
`)

		// 测试
		identifier := "12345"
		auth, err := GenerateTtposAuth(identifier)

		// 验证
		t.AssertNil(err)
		t.Assert(len(auth) > 0, true)

		// 验证 MD5 计算正确性
		expected, _ := gmd5.EncryptString(identifier + testSecret)
		t.Assert(auth, expected)
	})
}

func TestGenerateTtposAuth_DifferentIdentifier(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 设置测试配置
		testSecret := "test-secret-456"

		setupTestConfig(`
app:
  callbackSecret: "` + testSecret + `"
`)

		// 不同的 identifier 应该生成不同的 auth
		auth1, err1 := GenerateTtposAuth("12345")
		auth2, err2 := GenerateTtposAuth("67890")

		t.AssertNil(err1)
		t.AssertNil(err2)
		t.AssertNE(auth1, auth2)
	})
}

func TestGenerateTtposAuth_WithShopUUID(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 设置测试配置
		testSecret := "test-secret-789"

		setupTestConfig(`
app:
  callbackSecret: "` + testSecret + `"
`)

		// 测试使用 shopUUID 转换为字符串
		shopUUID := uint64(12345)
		identifier := fmt.Sprintf("%d", shopUUID)
		auth, err := GenerateTtposAuth(identifier)

		// 验证
		t.AssertNil(err)
		t.Assert(len(auth) > 0, true)

		// 验证 MD5 计算正确性
		expected, _ := gmd5.EncryptString(identifier + testSecret)
		t.Assert(auth, expected)
	})
}
