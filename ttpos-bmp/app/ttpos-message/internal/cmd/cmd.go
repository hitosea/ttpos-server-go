package cmd

import (
	"context"
	"ttpos-bmp/internal/pkg/otlp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"ttpos-bmp/app/ttpos-message/internal/boot"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start ttpos-message service",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化服务
			boot.Init(ctx)
			// 初始化OTLP
			shutdownOtlp := otlp.InitOtlp()
			defer shutdownOtlp(ctx)
			// 启动 HTTP 服务（用于健康检查等）
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				boot.RegisterHTTPRoutes(s)
			})

			g.Log().Info(ctx, "TTPOS 消息中心服务启动成功")

			// 启动 HTTP 服务
			s.Run()

			return nil
		},
	}
)
