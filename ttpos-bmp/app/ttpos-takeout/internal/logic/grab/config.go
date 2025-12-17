// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

// MustConfig 获取 Grab 配置 (平台维度)
// 读取 app.provider.grab.platform 节点，Partner 配置由 PartnerConfigLoader 独立读取
func MustConfig() *conf.Grab {
	ctx := gctx.New()

	var grab conf.Grab
	if err := g.Cfg().MustGet(ctx, "app.provider.grab.platform").Scan(&grab); err != nil {
		g.Log().Fatal(ctx, err)
		panic(gerror.Newf("获取 Grab 平台配置失败: %v", err))
	}

	return &grab
}
