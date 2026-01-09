package grab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
)

// TestCancelOrder_参数验证失败_订单实体为空 测试参数验证：订单实体为空
func TestCancelOrder_参数验证失败_订单实体为空(t *testing.T) {
	s := &sGrab{}
	ctx := context.Background()

	res, err := s.CancelOrderEntity(ctx, nil, "1")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "订单实体不能为空")
}

// TestCancelOrder_参数验证失败_订单渠道错误 测试参数验证：订单渠道错误
func TestCancelOrder_参数验证失败_订单渠道错误(t *testing.T) {
	s := &sGrab{}
	ctx := context.Background()

	orderEntity := &entity.Order{
		Uuid:               "test-uuid",
		ProviderName:       "unsupported_platform", // 非grab平台
		ProviderOrderId:    "G-123",
		ProviderMerchantId: "M-001",
		RawData:            `{"orderID":"G-123","merchantID":"M-001"}`,
	}

	res, err := s.CancelOrderEntity(ctx, orderEntity, "1")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "订单渠道错误，期望 grab")
}

// TestCancelOrder_集成测试说明 集成测试说明
func TestCancelOrder_集成测试说明(t *testing.T) {
	/*
		注意：由于 CancelOrder 方法依赖 service.Grab() 的调用，
		这里只测试参数验证和数据解析逻辑。

		完整的业务逻辑测试需要通过集成测试完成，包括：
		1. 预检查API调用（CheckOrderCancelable）
		2. 取消API调用（CancelOrder）
		3. 不可取消场景的返回
		4. 可取消场景的成功取消

		这些测试需要：
		- Mock service.Grab() 接口
		- 设置测试环境和数据库
		- 模拟外部API调用
	*/
	t.Skip("这是一个说明性测试，实际的集成测试需要Mock service依赖")
}
