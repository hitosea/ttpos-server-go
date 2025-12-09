package channel_menu

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("New() returned nil")
	}
}

// Test_sChannelMenu_Methods 验证方法签名和基本调用
// 注意：由于缺少真实的数据库连接，预期这些调用会失败或返回连接错误
func Test_sChannelMenu_Methods(t *testing.T) {
	s := New()
	ctx := context.Background()

	// 1. 测试 SaveChannelMenu
	// 预期：因为没有数据库连接，应该返回错误，但 panic 说明代码有问题
	err := s.SaveChannelMenu(ctx, 123, "grab", "{}")
	if err == nil {
		// 如果在有 mock DB 的环境下，可能成功
	} else {
		t.Logf("SaveChannelMenu returned error as expected (no DB): %v", err)
	}

	// 2. 测试 GetChannelMenu
	_, err = s.GetChannelMenu(ctx, 123, "grab")
	if err == nil {
		// 同上
	} else {
		t.Logf("GetChannelMenu returned error as expected (no DB): %v", err)
	}
}
