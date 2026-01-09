// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	// Grab Grab服务实例
	Grab = new(sGrab)
)

type sGrab struct{}

func init() {
	service.RegisterGrab(Grab)
}

// MustConfig 获取 Grab 配置 (平台维度)
// 读取 app.provider.grab.platform 节点
func MustConfig(ctx context.Context) *conf.Grab {
	if ctx == nil {
		ctx = gctx.New()
	}

	var grabCfg conf.Grab
	if err := g.Cfg().MustGet(ctx, "app.provider.grab.platform").Scan(&grabCfg); err != nil {
		g.Log().Fatal(ctx, err)
		panic(gerror.Newf("获取 Grab 平台配置失败: %v", err))
	}

	return &grabCfg
}
