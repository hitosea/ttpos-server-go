package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
)

func (c *ControllerV1) OrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) (res *v1.OrderUpdateRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
