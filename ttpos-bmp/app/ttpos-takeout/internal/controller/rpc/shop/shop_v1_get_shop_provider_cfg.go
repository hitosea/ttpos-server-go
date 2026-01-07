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

// GetShopProviderCfg 查询门店第三方配置
func (c *Controller) GetShopProviderCfg(ctx context.Context, req *api.GetShopProviderCfgReq) (res *takeout.ApiResponse, err error) {
	// 参数校验
	if req.ShopUuid == 0 {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 不能为 0",
		}, nil
	}

	// 调用 Logic 层
	resp, err := service.ShopProviderCfg().GetShopProviderCfgForRPC(ctx, req.ShopUuid, req.ProviderName)
	if err != nil {
		g.Log().Errorf(ctx, "[Shop] GetShopProviderCfg 失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "查询门店配置失败: " + err.Error(),
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

	g.Log().Debugf(ctx, "[Shop] GetShopProviderCfg 成功: shop_uuid=%d, provider_count=%d", resp.ShopUuid, len(resp.Providers))
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}
