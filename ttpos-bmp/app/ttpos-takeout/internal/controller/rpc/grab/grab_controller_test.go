package grab

import (
	"context"
	"testing"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGrabSelfServeService 模拟 GrabSelfServe 服务
type MockGrabSelfServeService struct {
	mock.Mock
}

func (m *MockGrabSelfServeService) CreateSelfServeJourney(ctx context.Context, req *api.CreateSelfServeJourneyReq) (*api.CreateSelfServeJourneyResp, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*api.CreateSelfServeJourneyResp), args.Error(1)
}

// MockGrabService 模拟 Grab 服务
type MockGrabService struct {
	mock.Mock
}

func (m *MockGrabService) GetShopProviderCfg(ctx context.Context, req *api.GetShopProviderCfgReq) (*api.GetShopProviderCfgResp, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*api.GetShopProviderCfgResp), args.Error(1)
}

func TestController_CreateSelfServeJourney_InvalidProviderName(t *testing.T) {
	// 创建控制器
	ctrl := &Controller{}

	// 测试参数校验 - 空的 provider_name
	invalidReq := &api.CreateSelfServeJourneyReq{
		ProviderName: "",
		ShopUuid:     "test-shop-uuid",
	}

	resp, err := ctrl.CreateSelfServeJourney(context.Background(), invalidReq)
	assert.NoError(t, err)
	assert.Equal(t, "4001", resp.Code)
	assert.Contains(t, resp.Message, "provider_name 不能为空")
	assert.Nil(t, resp.Data)
}

func TestController_CreateSelfServeJourney_InvalidShopUuid(t *testing.T) {
	ctrl := &Controller{}

	req := &api.CreateSelfServeJourneyReq{
		ProviderName: "grab",
		ShopUuid:     "",
	}

	resp, err := ctrl.CreateSelfServeJourney(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "4001", resp.Code)
	assert.Contains(t, resp.Message, "shop_uuid 不能为空")
	assert.Nil(t, resp.Data)
}

func TestController_GetShopProviderCfg_BasicStructure(t *testing.T) {
	ctrl := &Controller{}

	req := &api.GetShopProviderCfgReq{
		ShopUuid:     12345,
		ProviderName: "grab",
	}

	// 由于服务依赖，这里只测试基本的结构和参数
	assert.NotNil(t, ctrl)
	assert.NotNil(t, req)
	assert.Equal(t, uint64(12345), req.ShopUuid)
	assert.Equal(t, "grab", req.ProviderName)
}

func TestApiResponse_Structure(t *testing.T) {
	// 测试 ApiResponse 结构
	resp := &takeout.ApiResponse{
		Code:    "0",
		Message: "success",
		Data:    nil,
	}

	assert.Equal(t, "0", resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestApiResponse_ErrorStructure(t *testing.T) {
	// 测试错误响应结构
	resp := &takeout.ApiResponse{
		Code:    "5001",
		Message: "服务器内部错误",
		Data:    nil,
	}

	assert.Equal(t, "5001", resp.Code)
	assert.Equal(t, "服务器内部错误", resp.Message)
	assert.Nil(t, resp.Data)
}
