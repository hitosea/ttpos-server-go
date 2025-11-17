package boot

import (
	"ttpos-bmp/internal/pkg/cache"
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
	//设置缓存
	cache.SetAdapter(ctx)
}
