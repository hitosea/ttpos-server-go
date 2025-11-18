package websocket

import (
	"context"
	"encoding/json"
	"time"
	"ttpos-bmp/app/ttpos-websocket/internal/consts"
	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
	"ttpos-bmp/app/ttpos-websocket/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// StartRedisSubscriber 启动 Redis 订阅者
// 在服务启动时调用，订阅 websocket_msg_push 频道
func StartRedisSubscriber(ctx context.Context) {
	g.Log().Info(ctx, "启动Redis订阅者")

	// 使用 GoFrame Redis 订阅
	conn, _, err := g.Redis().Subscribe(ctx, consts.ChannelWebsocketMsgPush)
	if err != nil {
		g.Log().Error(ctx, "订阅Redis频道失败", err)
		return
	}
	defer conn.Close(ctx)

	// 循环接收消息
	for {
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "Redis订阅者已停止")
			return
		default:
			// 接收订阅消息
			msg, err := conn.ReceiveMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					g.Log().Info(ctx, "Redis订阅者已停止")
					return
				}
				g.Log().Error(ctx, "接收Redis消息失败", err)
				time.Sleep(100 * time.Millisecond) // 错误时等待一下再重试
				continue
			}

			if msg == nil {
				continue
			}

			g.Log().Debug(ctx, "收到Redis消息", "payload", msg.Payload)

			// 解析消息
			var pushInput dto.PushMessageInput
			if err := json.Unmarshal([]byte(msg.Payload), &pushInput); err != nil {
				g.Log().Error(ctx, "解析Redis消息失败", err, "payload", msg.Payload)
				continue
			}

			// 处理推送消息（在本节点推送）
			if err := service.Websocket().(*sWebSocket).processPushMessage(ctx, &pushInput); err != nil {
				g.Log().Error(ctx, "处理推送消息失败", err)
			}
		}
	}
}
