package event

import (
	"time"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
)

// sendTakeoutOrderWebSocketNotification 发送外卖订单 WebSocket 通知的辅助函数
func sendTakeoutOrderWebSocketNotification(
	companyUuid uint64,
	takeoutOrderUuid string,
	platform string,
	shortOrderNumber string,
	eventType string,
	extraData map[string]any,
) {
	// 构建 WebSocket 消息数据
	data := map[string]any{
		"update_time":        time.Now().Unix(),
		"takeout_order_uuid": takeoutOrderUuid,
		"platform":           platform,
		"short_order_number": shortOrderNumber,
		"type":               eventType,
	}

	// 合并额外数据
	for k, v := range extraData {
		data[k] = v
	}

	// 推送 WebSocket 消息到所有客户端
	if err := websocket.PushClient(
		companyUuid,
		websocket.SourceCashier, // 推送给收银机客户端
		"*",                     // 所有设备
		websocket.UPDATE_TAKEOUT_ORDER,
		data,
	); err != nil {
		logger.Logger.Error("推送外卖订单通知失败",
			zap.Uint64("companyUuid", companyUuid),
			zap.String("takeoutOrderUuid", takeoutOrderUuid),
			zap.String("eventType", eventType),
			zap.Error(err))
	}
}
