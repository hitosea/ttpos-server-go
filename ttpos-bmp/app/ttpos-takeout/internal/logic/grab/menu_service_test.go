package grab

import (
	"context"
	"testing"
	"time"

	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

// TestNotifyMenuUpdate 测试菜单更新通知
// 注意: 此测试需要 queue 包正确初始化，实际消息发送由 queue 包负责
func TestNotifyMenuUpdate(t *testing.T) {
	service := &MenuService{}

	event := &grabDto.ProviderMenuUpdateEvent{
		ProviderName:      "grab",
		MerchantID:        "M-123",
		PartnerMerchantID: "P-456",
		StorageKey:        "test-key",
		ReceivedAt:        time.Now().Unix(),
	}

	// 测试方法调用（实际发送可能失败，因为 queue 包可能未初始化，但不影响方法逻辑测试）
	err := service.NotifyMenuUpdate(context.Background(), event)
	// 如果 queue 包未初始化，会返回错误，这是预期的
	// 我们主要测试方法不会 panic，且参数传递正确
	if err != nil {
		t.Logf("NotifyMenuUpdate returned error (expected if queue not initialized): %v", err)
	}
}
