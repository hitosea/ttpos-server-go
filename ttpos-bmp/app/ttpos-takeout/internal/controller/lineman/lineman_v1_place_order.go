package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
)

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) (res *v1.PlaceOrderRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
