// Package boot 实现服务启动初始化逻辑
package boot

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	_ "ttpos-bmp/app/ttpos-message/internal/logic/message"
	_ "ttpos-bmp/app/ttpos-message/internal/logic/queue"
	_ "ttpos-bmp/app/ttpos-message/internal/logic/sender"
	"ttpos-bmp/app/ttpos-message/internal/service"
)

// Init 初始化服务
// 在服务启动时执行，初始化各个组件
func Init(ctx context.Context) {
	g.Log().Info(ctx, "开始初始化 TTPOS 消息中心服务...")

	// 初始化 Mailgun 邮件服务
	if err := service.Mailgun().Init(ctx); err != nil {
		g.Log().Error(ctx, "初始化 Mailgun 服务失败", err)
	} else {
		g.Log().Info(ctx, "Mailgun 服务初始化成功")
	}

	// 初始化 RocketMQ 队列服务
	if err := service.Queue().Init(ctx); err != nil {
		g.Log().Error(ctx, "初始化队列服务失败", err)
	} else {
		g.Log().Info(ctx, "队列服务初始化成功")
	}

	g.Log().Info(ctx, "TTPOS 消息中心服务初始化完成")
}

// Shutdown 服务关闭时的清理工作
func Shutdown(ctx context.Context) {
	g.Log().Info(ctx, "开始关闭 TTPOS 消息中心服务...")
	// TODO: 添加资源清理逻辑（如关闭队列连接等）
	g.Log().Info(ctx, "TTPOS 消息中心服务已关闭")
}

// RegisterHTTPRoutes 注册 HTTP 路由（如果需要）
func RegisterHTTPRoutes(s *ghttp.Server) {
	// 健康检查接口
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"service": "ttpos-message",
		})
	})

	// 配置查询接口（仅用于调试）
	s.BindHandler("/debug/config", func(r *ghttp.Request) {
		mailgunConfig := service.Mailgun().GetConfig()
		queueEnabled := service.Queue().IsEnabled()

		r.Response.WriteJson(g.Map{
			"mailgun": mailgunConfig,
			"queue": g.Map{
				"enabled": queueEnabled,
			},
		})
	})
}
