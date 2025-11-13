package cmd

import (
    "context"
    "ttpos-bmp/app/ttpos-takeout/internal/controller/callback"
    "ttpos-bmp/internal/pkg/middleware"

    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
    "github.com/gogf/gf/v2/os/gcmd"

    "ttpos-bmp/app/ttpos-takeout/internal/controller/hello"
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
                    middleware.MiddlewareHandlerVersion)
				group.Bind(
					hello.NewV1(),
					callback.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
