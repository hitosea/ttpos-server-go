package grab

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"
)

// NotifyMenuUpdate 通知菜单更新
func (c *Controller) NotifyMenuUpdate(ctx context.Context, req *api.NotifyMenuUpdateReq) (res *takeout.ApiResponse, err error) {
	// 参数校验
	if req.MerchantId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "merchant_id 不能为空",
		}, nil
	}

	// 调用 Logic 层
	requestID, err := service.Grab().NotifyMenuUpdate(ctx, req.MerchantId)
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 通知菜单更新失败: merchant_id=%s, err=%v", req.MerchantId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "通知菜单更新失败: " + err.Error(),
		}, nil
	}

	// 构建响应
	resp := &api.NotifyMenuUpdateResp{
		RequestId: requestID,
	}

	// 将 resp 转换为 anypb.Any
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: "序列化响应数据失败",
		}, nil
	}

	g.Log().Infof(ctx, "[Grab] 通知菜单更新成功: merchant_id=%s, request_id=%s", req.MerchantId, requestID)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: "成功",
		Data:    dataAny,
	}, nil
}
