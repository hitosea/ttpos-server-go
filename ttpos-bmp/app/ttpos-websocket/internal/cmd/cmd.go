// Package cmd 命令行工具
package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"

	"ttpos-bmp/app/ttpos-websocket/internal/boot"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start ttpos-websocket service",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化服务
			boot.Init(ctx)

			// 启动 HTTP 服务器（包含 WebSocket）
			g.Server().Run()

			return nil
		},
	}
)
