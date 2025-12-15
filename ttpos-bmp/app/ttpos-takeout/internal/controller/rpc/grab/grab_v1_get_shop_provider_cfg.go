package grab

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"google.golang.org/protobuf/types/known/anypb"
)

// GetShopProviderCfg 查询门店第三方配置
func (c *Controller) GetShopProviderCfg(ctx context.Context, req *api.GetShopProviderCfgReq) (res *takeout.ApiResponse, err error) {
	resp, err := service.Grab().GetShopProviderCfg(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: "获取门店配置失败: " + err.Error(),
		}, nil
	}

	// 将 resp 转换为 anypb.Any
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: "序列化响应数据失败",
		}, nil
	}

	return &takeout.ApiResponse{
		Code:    "0",
		Message: "success",
		Data:    dataAny,
	}, nil
}
