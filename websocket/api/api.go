package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"websocket/constant"
	"websocket/pkg/cache"
	"websocket/utils"
)

func PushClient(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	var params struct {
		CompanyUuid  uint        `json:"company_uuid"`
		SourceClient string      `json:"source_client"`
		DeviceId     string      `json:"device_id"`
		MessageType  string      `json:"message_type"`
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

	// 推送
	err := cache.GlobalRedis.Client.Publish(context.Background(), "websocket_msg_push", utils.StructToJson(params)).Err()
	if err != nil {
		err := cache.GlobalRedis.Client.Publish(context.Background(), "websocket_msg_push", utils.StructToJson(params)).Err()
		if err != nil {
			fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
				"code":    constant.CodeFail,
				"error":   err,
				"message": "failed",
			}))
		} else {
			fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
				"code":    constant.CodeSuccess,
				"message": "success",
			}))
		}
	} else {
		fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
			"code":    constant.CodeSuccess,
			"message": "success",
		}))
	}
}
