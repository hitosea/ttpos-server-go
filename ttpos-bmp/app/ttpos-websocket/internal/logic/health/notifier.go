// Package health 健康检查逻辑层
package health

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Notifier 健康状态通知器接口（预留扩展）
type Notifier interface {
	// NotifyDown 通知服务降级
	NotifyDown(ctx context.Context, response *HealthResponse) error
	// NotifyUp 通知服务恢复
	NotifyUp(ctx context.Context, response *HealthResponse) error
}

// LogNotifier 日志通知器（默认实现）
type LogNotifier struct {
	lastDownTime time.Time
}

// NewLogNotifier 创建日志通知器
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// NotifyDown 通知服务降级
func (n *LogNotifier) NotifyDown(ctx context.Context, response *HealthResponse) error {
	n.lastDownTime = time.Now()

	g.Log().Errorf(ctx, "⚠️ 健康检查失败，服务降级: status=%s, components=%+v",
		response.Status, response.Components)

	// TODO: 实际环境中可以：
	// 1. 调用 webhook 通知运维群
	// 2. 触发告警系统
	// 3. 暂停消息推送（如果配置了 degrade.pausePush）

	return nil
}

// NotifyUp 通知服务恢复
func (n *LogNotifier) NotifyUp(ctx context.Context, response *HealthResponse) error {
	downDuration := time.Since(n.lastDownTime)

	g.Log().Infof(ctx, "✅ 健康检查恢复: status=%s, down_duration=%s",
		response.Status, downDuration)

	// TODO: 实际环境中可以：
	// 1. 调用 webhook 通知运维群
	// 2. 恢复消息推送

	return nil
}

// WebhookNotifier Webhook 通知器（预留）
type WebhookNotifier struct {
	webhookURL string
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(webhookURL string) *WebhookNotifier {
	return &WebhookNotifier{
		webhookURL: webhookURL,
	}
}

// NotifyDown 通知服务降级
func (n *WebhookNotifier) NotifyDown(ctx context.Context, response *HealthResponse) error {
	// TODO: 实现 HTTP POST 到 webhook URL
	g.Log().Infof(ctx, "Webhook 通知服务降级: url=%s", n.webhookURL)
	return nil
}

// NotifyUp 通知服务恢复
func (n *WebhookNotifier) NotifyUp(ctx context.Context, response *HealthResponse) error {
	// TODO: 实现 HTTP POST 到 webhook URL
	g.Log().Infof(ctx, "Webhook 通知服务恢复: url=%s", n.webhookURL)
	return nil
}
