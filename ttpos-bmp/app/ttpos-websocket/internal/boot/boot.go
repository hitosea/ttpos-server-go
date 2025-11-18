// Package boot 服务启动初始化
package boot

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	httpController "ttpos-bmp/app/ttpos-websocket/internal/controller/http"
	_ "ttpos-bmp/app/ttpos-websocket/internal/logic"
	websocketLogic "ttpos-bmp/app/ttpos-websocket/internal/logic/websocket"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
)

// Init 初始化服务
func Init(ctx context.Context) {
	g.Log().Info(ctx, "服务启动初始化开始")

	// 初始化 RPC 服务
	InitRpc(ctx)

	// 初始化 HTTP 服务（WebSocket 升级）
	InitHttp(ctx)

	// 启动 Redis 订阅者（集群环境下的消息分发）
	InitRedisSubscriber(ctx)

	g.Log().Info(ctx, "服务启动初始化完成")
}

// InitRedisSubscriber 初始化 Redis 订阅者
func InitRedisSubscriber(ctx context.Context) {
	// 在 goroutine 中启动订阅者
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Error(ctx, "Redis订阅者异常", r)
			}
		}()

		// 调用 logic 层的订阅函数
		websocketLogic.StartRedisSubscriber(ctx)
	}()

	g.Log().Info(ctx, "Redis订阅者已启动")
}

// InitHttp 初始化 HTTP 服务
// 用于提供 WebSocket 升级接口和 HTTP API
func InitHttp(ctx context.Context) {
	s := g.Server()

	// 设置 WebSocket 升级路由
	s.BindHandler("/ws", service.Websocket().HandleConnections)

	// 设置 HTTP 推送接口（兼容旧API）
	s.Group("/ws", func(group *ghttp.RouterGroup) {
		group.POST("/push", httpController.PushMessage)
	})

	g.Log().Info(ctx, "HTTP 服务初始化完成")
}

// Shutdown 服务关闭
// 优雅关闭服务，清理所有资源
func Shutdown(ctx context.Context) {
	g.Log().Info(ctx, "服务正在关闭...")

	// 1. 关闭所有 WebSocket 连接
	closedCount := websocketLogic.CloseAllConnections(ctx)
	g.Log().Info(ctx, "已关闭所有 WebSocket 连接", "count", closedCount)

	// 2. 停止 Redis 订阅者
	// Redis 订阅者会通过 context.Done() 自动停止

	// 3. 等待一小段时间，确保所有连接都已关闭
	time.Sleep(500 * time.Millisecond)

	g.Log().Info(ctx, "服务已关闭")
}
