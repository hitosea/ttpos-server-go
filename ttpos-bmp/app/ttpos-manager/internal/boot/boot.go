package boot

import "github.com/gogf/gf/v2/os/gctx"

var (
	ctx = gctx.GetInitCtx()
)

func init() {
	InitRpc(ctx)
}
