package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-websocket/internal/consts"
	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
	"ttpos-bmp/app/ttpos-websocket/internal/service"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

// StartRedisSubscriber 启动 Redis 订阅者
// 在服务启动时调用，订阅 websocket_msg_push 频道
func StartRedisSubscriber(ctx context.Context) {
	g.Log().Info(ctx, "启动Redis订阅者")

	// 使用 GoFrame Redis 订阅
	conn, _, err := g.Redis().Subscribe(ctx, consts.ChannelWebsocketMsgPush)
	if err != nil {
		g.Log().Error(ctx, "订阅Redis频道失败 重试中...", err)
		time.Sleep(2 * time.Second)
		StartRedisSubscriber(ctx)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			g.Log().Error(ctx, "Redis订阅者异常-02 重试中...", r)
		}
		conn.Close(ctx)
		// 延迟2秒重连
		time.Sleep(2 * time.Second)
		go StartRedisSubscriber(ctx)
	}()

	// 循环接收消息
	for {
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "Redis订阅者已停止")
			return
		default:
			// 接收订阅消息（Receive可以正确处理Subscription和Message类型）
			msgVar, err := conn.Receive(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					g.Log().Info(ctx, "Redis订阅者已停止")
					return
				} else if strings.Contains(err.Error(), "connect: connection refused") {
					g.Log().Error(ctx, "Redis连接失败", err)
					return
				} else if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "connection") {
					// 连接断开，返回以触发重连
					g.Log().Warning(ctx, "Redis连接断开，准备重连", err)
					return
				}
				g.Log().Error(ctx, "接收Redis消息失败", err)
				time.Sleep(100 * time.Millisecond) // 错误时等待一下再重试
				continue
			}

			if msgVar == nil {
				continue
			}

			// 获取实际的消息对象
			msg := msgVar.Val()
			if msg == nil {
				continue
			}

			// 类型判断：只处理消息类型，忽略订阅确认等其他类型
			switch v := msg.(type) {
			case *gredis.Message:
				g.Log().Info(ctx, "收到Redis消息", "payload", v.Payload)

				// 解析消息
				var pushInput dto.PushMessageInput
				if err := json.Unmarshal([]byte(v.Payload), &pushInput); err != nil {
					g.Log().Error(ctx, "解析Redis消息失败", err, "payload", v.Payload)
					continue
				}

				// 处理推送消息（在本节点推送）
				if err := service.Websocket().(*sWebSocket).processPushMessage(ctx, &pushInput); err != nil {
					g.Log().Error(ctx, "处理推送消息失败", err)
				}
			case *gredis.Subscription:
				// 订阅确认消息，记录日志即可
				g.Log().Debug(ctx, "Redis订阅确认", "channel", v.Channel, "count", v.Count)
			default:
				// 其他未知类型
				g.Log().Debug(ctx, "收到未知类型的Redis消息", "type", fmt.Sprintf("%T", v))
			}
		}
	}
}
