// Package boot 服务启动初始化
package boot

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	_ "ttpos-bmp/app/ttpos-websocket/internal/logic"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
)

// Init 初始化服务
func Init(ctx context.Context) {
	g.Log().Info(ctx, "服务启动初始化开始")

	// 初始化 RPC 服务
	InitRpc(ctx)

	// 初始化 HTTP 服务（WebSocket 升级）
	InitHttp(ctx)

	g.Log().Info(ctx, "服务启动初始化完成")
}

// InitHttp 初始化 HTTP 服务
// 用于提供 WebSocket 升级接口
func InitHttp(ctx context.Context) {
	s := g.Server()

	// 设置 WebSocket 升级路由
	s.BindHandler("/ws", service.Websocket().HandleConnections)

	g.Log().Info(ctx, "HTTP 服务初始化完成")
}

// Shutdown 服务关闭
func Shutdown(ctx context.Context) {
	g.Log().Info(ctx, "服务正在关闭...")
	// TODO: 清理资源，关闭所有 WebSocket 连接
	g.Log().Info(ctx, "服务已关闭")
}
