package grab_order

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
)

// TestCancelOrder_参数验证失败_订单实体为空 测试参数验证：订单实体为空
func TestCancelOrder_参数验证失败_订单实体为空(t *testing.T) {
	s := New()
	ctx := context.Background()

	res, err := s.CancelOrder(ctx, nil, 1)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "订单实体不能为空")
}

// TestCancelOrder_参数验证失败_订单渠道错误 测试参数验证：订单渠道错误
func TestCancelOrder_参数验证失败_订单渠道错误(t *testing.T) {
	s := New()
	ctx := context.Background()

	orderEntity := &entity.Order{
		Uuid:         "test-uuid",
		ProviderName: "unsupported_platform", // 非grab平台
		RawData:      `{"orderID":"G-123","merchantID":"M-001"}`,
	}

	res, err := s.CancelOrder(ctx, orderEntity, 1)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "订单渠道错误，期望 grab")
}

// TestParseOrderData_正常解析 测试parseOrderData正常解析
func TestParseOrderData_正常解析(t *testing.T) {
	s := New()

	rawData := `{"orderID":"G-123456","merchantID":"M-001"}`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.NoError(t, err)
	assert.Equal(t, "G-123456", orderID)
	assert.Equal(t, "M-001", merchantID)
}

// TestParseOrderData_RawData为空 测试parseOrderData：RawData为空
func TestParseOrderData_RawData为空(t *testing.T) {
	s := New()

	orderID, merchantID, err := s.parseOrderData("")

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "raw_data 为空")
}

// TestParseOrderData_JSON解析失败 测试parseOrderData：JSON解析失败
func TestParseOrderData_JSON解析失败(t *testing.T) {
	s := New()

	rawData := `invalid json`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "解析 JSON 失败")
}

// TestParseOrderData_缺少orderID字段 测试parseOrderData：缺少orderID字段
func TestParseOrderData_缺少orderID字段(t *testing.T) {
	s := New()

	rawData := `{"merchantID":"M-001"}`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "缺少 orderID 字段")
}

// TestParseOrderData_缺少merchantID字段 测试parseOrderData：缺少merchantID字段
func TestParseOrderData_缺少merchantID字段(t *testing.T) {
	s := New()

	rawData := `{"orderID":"G-123456"}`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "缺少 merchantID 字段")
}

// TestParseOrderData_orderID类型错误 测试parseOrderData：orderID类型错误
func TestParseOrderData_orderID类型错误(t *testing.T) {
	s := New()

	rawData := `{"orderID":12345,"merchantID":"M-001"}`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "orderID 字段类型错误")
}

// TestParseOrderData_merchantID类型错误 测试parseOrderData：merchantID类型错误
func TestParseOrderData_merchantID类型错误(t *testing.T) {
	s := New()

	rawData := `{"orderID":"G-123456","merchantID":12345}`
	orderID, merchantID, err := s.parseOrderData(rawData)

	assert.Error(t, err)
	assert.Empty(t, orderID)
	assert.Empty(t, merchantID)
	assert.Contains(t, err.Error(), "merchantID 字段类型错误")
}

// TestCancelOrder_解析订单数据失败 测试CancelOrder：解析订单数据失败
func TestCancelOrder_解析订单数据失败(t *testing.T) {
	s := New()
	ctx := context.Background()

	orderEntity := &entity.Order{
		Uuid:         "test-uuid",
		ProviderName: "grab",
		RawData:      `invalid json`, // 无效JSON
	}

	res, err := s.CancelOrder(ctx, orderEntity, 1)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "解析订单数据失败")
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
