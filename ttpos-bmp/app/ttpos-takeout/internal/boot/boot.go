package boot

import (
	"ttpos-bmp/app/ttpos-takeout/internal/global"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/os/gctx"
)

var ctx = gctx.GetInitCtx()

func init() {
	// 指定应用类型为 Takeout，避免跨应用 ID 冲突
	uuid.InitIdGenerator(ctx, uuid.AppTypeTakeout)
	DbCheck()

	InitRpc(ctx)

	global.Init(ctx)
}
