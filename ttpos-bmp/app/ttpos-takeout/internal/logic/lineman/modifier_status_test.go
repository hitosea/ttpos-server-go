package lineman_test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ttpos-bmp/app/ttpos-takeout/internal/logic/lineman"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// ==================== 状态映射函数测试 ====================

// TestMapStatusToLinemanModifier_Available 测试 AVAILABLE 映射
func TestMapStatusToLinemanModifier_Available(t *testing.T) {
	status, err := lineman.MapStatusToLinemanModifier("AVAILABLE")
	assert.NoError(t, err)
	assert.Equal(t, 1, status)
}

// TestMapStatusToLinemanModifier_Unavailable 测试 UNAVAILABLE 映射
func TestMapStatusToLinemanModifier_Unavailable(t *testing.T) {
	status, err := lineman.MapStatusToLinemanModifier("UNAVAILABLE")
	assert.NoError(t, err)
	assert.Equal(t, 3, status)
}

// TestMapStatusToLinemanModifier_SoldOutToday 测试 SOLD_OUT_TODAY 映射
func TestMapStatusToLinemanModifier_SoldOutToday(t *testing.T) {
	status, err := lineman.MapStatusToLinemanModifier("SOLD_OUT_TODAY")
	assert.NoError(t, err)
	assert.Equal(t, 2, status)
}

// TestMapStatusToLinemanModifier_Empty 测试空字符串返回错误
func TestMapStatusToLinemanModifier_Empty(t *testing.T) {
	status, err := lineman.MapStatusToLinemanModifier("")
	assert.Error(t, err)
	assert.Equal(t, 0, status)
	assert.Contains(t, err.Error(), "available_status 不能为空")
}

// TestMapStatusToLinemanModifier_Invalid 测试不支持的状态返回错误
func TestMapStatusToLinemanModifier_Invalid(t *testing.T) {
	status, err := lineman.MapStatusToLinemanModifier("INVALID_STATUS")
	assert.Error(t, err)
	assert.Equal(t, 0, status)
	assert.Contains(t, err.Error(), "不支持的状态")
}

// ==================== ModifierStatusLogic 测试 ====================

// MockModifierStatusClient Mock 修饰符状态客户端
type MockModifierStatusClient struct {
	mock.Mock
}

func (m *MockModifierStatusClient) UpdateModifierStatusWithRetry(
	ctx context.Context,
	storeId string,
	req *lineman_dto.ModifierStatusUpdateReq,
) (*lineman_dto.ModifierStatusUpdateResp, error) {
	args := m.Called(ctx, storeId, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lineman_dto.ModifierStatusUpdateResp), args.Error(1)
}

// TestModifierStatusLogic_UpdateModifierStatus_Success 测试成功场景
func TestModifierStatusLogic_UpdateModifierStatus_Success(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	// Mock 返回成功响应
	mockClient.On("UpdateModifierStatusWithRetry",
		ctx,
		"test-store-id",
		mock.MatchedBy(func(req *lineman_dto.ModifierStatusUpdateReq) bool {
			return len(req.PropertyValues) == 1 &&
				req.PropertyValues[0].ID == "test-modifier-id" &&
				req.PropertyValues[0].Status == 1
		}),
	).Return(&lineman_dto.ModifierStatusUpdateResp{
		Status:  "ok",
		Code:    "SUCCESS",
		Message: "Property values status updated",
	}, nil)

	err := logic.UpdateModifierStatus(ctx, "test-store-id", "test-modifier-id", 1)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestModifierStatusLogic_UpdateModifierStatus_EmptyStoreId 测试 storeId 为空
func TestModifierStatusLogic_UpdateModifierStatus_EmptyStoreId(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	err := logic.UpdateModifierStatus(ctx, "", "test-modifier-id", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storeId 不能为空")
}

// TestModifierStatusLogic_UpdateModifierStatus_EmptyModifierId 测试 modifierId 为空
func TestModifierStatusLogic_UpdateModifierStatus_EmptyModifierId(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	err := logic.UpdateModifierStatus(ctx, "test-store-id", "", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modifierId 不能为空")
}

// TestModifierStatusLogic_UpdateModifierStatus_InvalidStatus 测试无效的 status 值
func TestModifierStatusLogic_UpdateModifierStatus_InvalidStatus(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	// 测试 status = 0（无效）
	err := logic.UpdateModifierStatus(ctx, "test-store-id", "test-modifier-id", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的 status")

	// 测试 status = 4（无效）
	err = logic.UpdateModifierStatus(ctx, "test-store-id", "test-modifier-id", 4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的 status")
}

// TestModifierStatusLogic_UpdateModifierStatus_ApiError 测试 API 错误
func TestModifierStatusLogic_UpdateModifierStatus_ApiError(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	// Mock 返回 API 错误
	mockClient.On("UpdateModifierStatusWithRetry",
		ctx,
		"test-store-id",
		mock.Anything,
	).Return(nil, gerror.New("network error"))

	err := logic.UpdateModifierStatus(ctx, "test-store-id", "test-modifier-id", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "调用 Lineman API 失败")
	mockClient.AssertExpectations(t)
}

// TestModifierStatusLogic_UpdateModifierStatus_BusinessError 测试业务错误
func TestModifierStatusLogic_UpdateModifierStatus_BusinessError(t *testing.T) {
	mockClient := new(MockModifierStatusClient)
	logic := lineman.NewModifierStatusLogic(mockClient)
	ctx := context.Background()

	// Mock 返回业务错误（status=fail）
	mockClient.On("UpdateModifierStatusWithRetry",
		ctx,
		"test-store-id",
		mock.Anything,
	).Return(&lineman_dto.ModifierStatusUpdateResp{
		Status:  "fail",
		Code:    "ERROR_INVALID_STATUS",
		Message: "Invalid status value",
	}, nil)

	err := logic.UpdateModifierStatus(ctx, "test-store-id", "test-modifier-id", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Lineman API 返回错误")
	assert.Contains(t, err.Error(), "ERROR_INVALID_STATUS")
	mockClient.AssertExpectations(t)
}
