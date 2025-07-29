package boot

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func DbCheck() {
	if err := g.DB().Ctx(gctx.GetInitCtx()).PingMaster(); err != nil {
		g.Log().Error(gctx.GetInitCtx(), "数据库链接失败", err)
		panic(err)
	}
}
