package callback

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"ttpos-bmp/app/ttpos-takeout/api/callback/v1"
)

func (c *ControllerV1) SkootarStatus(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error) {
	g.Log().Infof(ctx, "获取skootar状态回调:%+v", req)
	res, err = service.Skootar().JobStatusChange(ctx, req)
	return
}
