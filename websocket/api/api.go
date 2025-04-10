package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"websocket/constant"
	"websocket/pkg/cache"
	"websocket/utils"

	"github.com/google/uuid"
)

func PushClient(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	var params struct {
		CompanyUuid  uint        `json:"company_uuid"`
		SourceClient string      `json:"source_client"`
		DeviceId     string      `json:"device_id"`
		NotDeviceId  string      `json:"not_device_id"`
		StaffUuid    uint64      `json:"staff_uuid"`
		NotStaffUuid uint64      `json:"not_staff_uuid"`
		MessageType  string      `json:"message_type"`
		MessageKey   string      `json:"message_key"`
		Data         interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		fmt.Println("Failed to decode JSON", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if params.DeviceId == "" {
		params.DeviceId = "*"
	}
	if params.SourceClient == "" {
		params.SourceClient = "*"
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 设置缓存
	if params.MessageKey != "" {
		cache.GlobalRedis.Set(params.MessageKey, uuidStr, 2*time.Second)
	}

	// 启动延时发送的goroutine
	go func(_uuid string) {
		// 检查Redis缓存是否被更新
		if params.MessageKey != "" {
			time.Sleep(900 * time.Millisecond)
			if cachedUUID, exists := cache.GlobalRedis.Get(params.MessageKey); exists {
				if _uuid != cachedUUID.(string) {
					return
				}
			}
		}
		// 推送
		err := cache.GlobalRedis.Publish("websocket_msg_push", utils.StructToJson(params))
		if err != nil {
			fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
				"code":    constant.CodeFail,
				"error":   err,
				"message": "failed",
			}))
		}
	}(uuidStr)

	// 返回成功
	fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
		"code":    constant.CodeSuccess,
		"message": "success",
	}))
}
