package grab

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"
)

// CreateSelfServeJourney 创建自助激活链接
func (c *Controller) CreateSelfServeJourney(ctx context.Context, req *api.CreateSelfServeJourneyReq) (res *takeout.ApiResponse, err error) {
	// 参数校验
	if req.ProviderName == "" {
		return &takeout.ApiResponse{
			Code:    "4001",
			Message: "provider_name 不能为空",
		}, nil
	}
	if req.ShopUuid == "" {
		return &takeout.ApiResponse{
			Code:    "4001",
			Message: "shop_uuid 不能为空",
		}, nil
	}

	// 调用 Logic 层
	activationURL, requestID, err := service.Grab().CreateSelfServeJourney(ctx, req.ShopUuid)
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] CreateSelfServeJourney failed: %v", err)
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: "创建自助激活链接失败: " + err.Error(),
		}, nil
	}

	// 构造响应对象
	resp := &api.CreateSelfServeJourneyResp{
		ProviderName: req.ProviderName,
		SelfServeUrl: activationURL,
		RequestId:    requestID,
	}

	// 如果请求中有 request_id，使用请求的 request_id
	if req.RequestId != "" {
		resp.RequestId = req.RequestId
	}

	// 将 resp 转换为 anypb.Any
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: "序列化响应数据失败",
		}, nil
	}

	g.Log().Debugf(ctx, "[Grab] CreateSelfServeJourney success: %+v", resp)
	return &takeout.ApiResponse{
		Code:    "0",
		Message: "success",
		Data:    dataAny,
	}, nil
}
