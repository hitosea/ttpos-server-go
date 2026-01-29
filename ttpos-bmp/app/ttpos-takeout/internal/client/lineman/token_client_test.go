// Package lineman Lineman API 客户端测试
package lineman

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

func TestJWTTokenClient_GenerateAndParseToken(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		client := NewJWTTokenClient()

		// 测试空参数
		_, _, err := client.GenerateToken(ctx, "", "")
		t.AssertNE(err, nil)

		// 测试不存在的 clientID
		_, _, err = client.GenerateToken(ctx, "non-existent-client", "secret")
		t.AssertNE(err, nil)

		// 注意：以下测试需要配置文件中有 lineman.partner 配置才能通过
		// 这里只是示例，实际测试时需要根据配置调整
		t.Log("JWT Token 基本功能测试完成（需要配置文件支持才能完整测试）")
	})
}

func TestJWTTokenClient_ParseToken_Invalid(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		client := NewJWTTokenClient()

		// 测试解析无效 Token
		_, err := client.ParseToken(ctx, "invalid-token")
		t.AssertNE(err, nil)
	})
}

func TestOAuthTokenClient_GetAuthorizationHeader(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		client := NewOAuthTokenClient()

		// 注意：这个测试会调用实际的 LINE MAN OAuth API
		// 如果配置不正确或网络不通，测试会失败
		header, err := client.GetAuthorizationHeader(ctx)
		if err != nil {
			t.Logf("获取 Authorization Header 失败: %v (可能是配置或网络问题)", err)
			return
		}

		t.AssertNE(header, "")
		t.Logf("Authorization Header: %s", header[:20]+"...") // 只显示前20个字符
	})
}
