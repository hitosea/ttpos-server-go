// Package lineman_token 提供 LINE MAN OAuth Token 生成与验证服务
//
// ⚠️ 临时方案: 后续将迁移到统一权限中心 SSO
package lineman_token

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

// MustConfig 获取 LINE MAN 平台配置
// 读取 app.provider.lineman.platform 节点
func MustConfig(ctx context.Context) *conf.Lineman {
	if ctx == nil {
		ctx = gctx.New()
	}

	var linemanCfg conf.Lineman
	if err := g.Cfg().MustGet(ctx, "app.provider.lineman.platform").Scan(&linemanCfg); err != nil {
		g.Log().Fatal(ctx, err)
		panic(gerror.Newf("获取 LINE MAN 平台配置失败: %v", err))
	}

	return &linemanCfg
}
