// Package grab_token 提供 Grab Partner Token 生成与验证服务
package grab_token

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

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
