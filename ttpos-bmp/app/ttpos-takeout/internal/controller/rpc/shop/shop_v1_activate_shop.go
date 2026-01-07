package shop

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/shop"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"
)

// ActivateShop 激活门店外卖渠道
func (c *Controller) ActivateShop(ctx context.Context, req *api.ActivateShopReq) (res *takeout.ApiResponse, err error) {
	// 参数校验
	if req.ProviderName == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "provider_name 不能为空",
		}, nil
	}
	if req.ShopUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 不能为空",
		}, nil
	}

	// 调用 Logic 层
	resp, err := service.ShopActivate().ActivateShop(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "[Shop] ActivateShop 失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "激活门店外卖渠道失败: " + err.Error(),
		}, nil
	}

	// 将 resp 转换为 anypb.Any
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Debugf(ctx, "[Shop] ActivateShop 成功: shop_uuid=%d, provider=%s", resp.ShopUuid, resp.ProviderName)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}
