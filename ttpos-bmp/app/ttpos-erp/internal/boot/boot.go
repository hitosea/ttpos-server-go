package boot

import (
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/os/gctx"
)

var (
	ctx = gctx.GetInitCtx()
)

func init() {
	InitRpc(ctx)
	InitConsumer(ctx)
	uuid.InitIdGenerator(ctx)
}
