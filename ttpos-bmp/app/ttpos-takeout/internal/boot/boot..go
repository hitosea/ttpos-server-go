package boot

import (
	"ttpos-bmp/app/ttpos-takeout/internal/global"

	"github.com/gogf/gf/v2/os/gctx"
)

var ctx = gctx.GetInitCtx()

func init() {
	DbCheck()

	InitRpc(ctx)

	global.Init(ctx)
}
