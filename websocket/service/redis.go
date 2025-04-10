package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"websocket/pkg/cache"
)

type MessageData struct {
	CompanyUuid  uint64      `json:"company_uuid"`
	SourceClient string      `json:"source_client"`
	DeviceId     string      `json:"device_id"`
	NotDeviceId  string      `json:"not_device_id"`
	StaffUuid    uint64      `json:"staff_uuid"`
	NotStaffUuid uint64      `json:"not_staff_uuid"`
	MessageType  string      `json:"message_type"`
	Data         interface{} `json:"data"`
}

func RedisSubscribe() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsub := cache.GlobalRedis.Client.Subscribe(ctx, "websocket_msg_push")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			fmt.Println("Redis Error receiving message:", err)
			time.Sleep(100 * time.Microsecond) // 重试间隔
			continue
		}

		// 处理消息
		var message MessageData
		err = json.Unmarshal([]byte(msg.Payload), &message)
		if err != nil {
			fmt.Println("Redis Error parsing message:", msg.Payload)
			continue
		}

		// 发送消息
		PushClient(message)
	}
}
