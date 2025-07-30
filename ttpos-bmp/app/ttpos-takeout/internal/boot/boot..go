package boot

import (
	"github.com/gogf/gf/v2/os/gctx"
	"ttpos-bmp/app/ttpos-takeout/internal/global"
)

var ctx = gctx.GetInitCtx()

func init() {
	DbCheck()

	InitRpc(ctx)

	global.Init(ctx)
}
