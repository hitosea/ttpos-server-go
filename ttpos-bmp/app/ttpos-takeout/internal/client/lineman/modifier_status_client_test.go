package lineman_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// TestModifierStatusClient_UpdateModifierStatus_Success 测试成功场景
func TestModifierStatusClient_UpdateModifierStatus_Success(t *testing.T) {
	t.Skip("需要 Lineman API 配置和 Token，跳过集成测试")

	client := lineman.NewModifierStatusClient()
	ctx := context.Background()

	req := &dto.ModifierStatusUpdateReq{
		PropertyValues: []dto.ModifierPropertyValue{
			{
				ID:     "test-modifier-id",
				Status: 1, // AVAILABLE
			},
		},
	}

	resp, err := client.UpdateModifierStatus(ctx, "test-store-id", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ok", resp.Status)
}

// TestModifierStatusClient_UpdateModifierStatus_ApiError 测试 API 错误场景
func TestModifierStatusClient_UpdateModifierStatus_ApiError(t *testing.T) {
	t.Skip("需要 Mock HTTP Server，暂时跳过")
	// TODO: 实现 Mock HTTP Server 测试 API 错误场景
}

// TestModifierStatusClient_UpdateModifierStatus_HttpError 测试 HTTP 错误场景
func TestModifierStatusClient_UpdateModifierStatus_HttpError(t *testing.T) {
	t.Skip("需要 Mock HTTP Server，暂时跳过")
	// TODO: 实现 Mock HTTP Server 测试 HTTP 错误（非 200）场景
}

// TestModifierStatusClient_UpdateModifierStatusWithRetry_Success 测试重试机制
func TestModifierStatusClient_UpdateModifierStatusWithRetry_Success(t *testing.T) {
	t.Skip("需要 Lineman API 配置和 Token，跳过集成测试")

	client := lineman.NewModifierStatusClient()
	ctx := context.Background()

	req := &dto.ModifierStatusUpdateReq{
		PropertyValues: []dto.ModifierPropertyValue{
			{
				ID:     "test-modifier-id",
				Status: 3, // SUSPENDED
			},
		},
	}

	resp, err := client.UpdateModifierStatusWithRetry(ctx, "test-store-id", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ok", resp.Status)
}
