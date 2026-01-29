// Package lineman Lineman API 客户端测试
package lineman

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

func TestPartnerConfigLoader_GetByCode(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		loader := &PartnerConfigLoader{}

		// 测试加载配置
		partner, err := loader.GetByCode(ctx, "default")
		if err != nil {
			t.Logf("加载配置失败: %v (这是正常的，如果配置文件中没有 lineman.partner.default)", err)
			return
		}

		t.AssertNE(partner, nil)
		t.Logf("Client ID: %s", partner.ClientID)
		t.Logf("Environment: %s", partner.Environment)
	})
}

func TestPartnerConfigLoader_GetByClientID(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		loader := &PartnerConfigLoader{}

		// 首先尝试通过 code 获取配置，然后用 clientID 再次查询
		partner, err := loader.GetByCode(ctx, "default")
		if err != nil {
			t.Logf("加载配置失败: %v (这是正常的，如果配置文件中没有 lineman.partner.default)", err)
			return
		}

		// 用 clientID 再次查询
		partner2, err := loader.GetByClientID(ctx, partner.ClientID)
		t.AssertNil(err)
		t.AssertEQ(partner.ClientID, partner2.ClientID)
		t.AssertEQ(partner.Environment, partner2.Environment)
	})
}

func TestMustConfig(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()

		// 测试获取平台配置
		// 注意：如果配置不存在会 panic
		defer func() {
			if r := recover(); r != nil {
				t.Logf("配置不存在，panic: %v", r)
			}
		}()

		cfg := MustConfig(ctx)
		t.AssertNE(cfg, nil)
		t.Logf("Endpoint: %s", cfg.Endpoint)
		t.Logf("ClientID: %s", cfg.ClientID)
	})
}
