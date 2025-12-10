package grab

import (
	"context"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) GetMenu(ctx context.Context, req *v1.GetMenuReq) (res *v1.GetMenuRes, err error) {
	// 调用 Service 层处理请求
	menuResp, err := service.Grab().HandleGetMenu(ctx, req.XGrabSignature, req.XGrabTimestamp, req.PartnerMerchantID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	return &v1.GetMenuRes{
		GetMenuResponse: menuResp,
	}, nil
}
