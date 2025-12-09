package grab

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"ttpos-bmp/app/ttpos-takeout/api/grab/v1"
)

func (c *ControllerV1) PushOrderState(ctx context.Context, req *v1.PushOrderStateReq) (res *v1.PushOrderStateRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
