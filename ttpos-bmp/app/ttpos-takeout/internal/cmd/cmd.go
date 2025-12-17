package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"ttpos-bmp/app/ttpos-takeout/internal/controller/callback"
	"ttpos-bmp/app/ttpos-takeout/internal/controller/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/controller/hello"
	"ttpos-bmp/app/ttpos-takeout/internal/middleware"
	pkgMiddleware "ttpos-bmp/internal/pkg/middleware"
	"ttpos-bmp/internal/pkg/otlp"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {

			//rpc.InitRpc(ctx)
			shutdownOtlp := otlp.InitOtlp()
			defer shutdownOtlp(ctx)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse,
					pkgMiddleware.MiddlewareHandlerVersion)
				group.Bind(
					hello.NewV1(),
					callback.NewV1(),
				)
			})

			// 注册 Grab 路由（OAuth 无需认证，Webhook 使用签名验证）
			s.Group("/api/v1/grab", func(group *ghttp.RouterGroup) {
				group.Middleware(middleware.MiddlewareDirectResponse, middleware.MiddlewareGrabJWTAuth)
				group.Bind(
					grab.NewV1(),
				)
			})

			s.Run()
			return nil
		},
	}
)
