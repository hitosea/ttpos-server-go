package grab

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// CreateSelfServeJourney 创建自助激活链接
func (c *Controller) CreateSelfServeJourney(ctx context.Context, req *api.CreateSelfServeJourneyReq) (res *api.CreateSelfServeJourneyResp, err error) {
	// 参数校验
	if req.ProviderName == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "shop_uuid 不能为空")
	}

	// 调用 Logic 层
	resp, err := service.GrabSelfServe().CreateSelfServeJourney(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] CreateSelfServeJourney failed: %v", err)
		return nil, gerror.Wrap(err, "创建自助激活链接失败")
	}

	g.Log().Debugf(ctx, "[Grab] CreateSelfServeJourney success: %+v", resp)
	return resp, nil
}
