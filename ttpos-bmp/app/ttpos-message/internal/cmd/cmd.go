package cmd

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"

	v1 "ttpos-bmp/app/ttpos-message/api/message/v1"
	"ttpos-bmp/app/ttpos-message/internal/boot"
	"ttpos-bmp/app/ttpos-message/internal/controller/rpc/message"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start ttpos-message service",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化服务
			boot.Init(ctx)

			// 启动 HTTP 服务（用于健康检查等）
			httpServer := g.Server()
			boot.RegisterHTTPRoutes(httpServer)

			// 启动 gRPC 服务
			grpcServer := grpcx.Server.New()
			v1.RegisterMessageServer(grpcServer.Server, message.New())

			g.Log().Info(ctx, "TTPOS 消息中心服务启动成功")
			g.Log().Info(ctx, "HTTP 服务地址:", httpServer.GetListenedAddress())
			g.Log().Info(ctx, "gRPC 服务地址:", grpcServer.GetListenedAddress())

			// 启动服务
			grpcServer.Run()

			return nil
		},
	}
)
