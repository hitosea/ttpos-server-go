package websocket

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ttpos-server-go/pkg/cache"
)

const (
	UPDATE_ORDER   = "update_order"   // 更新订单 data = {"update_time": 1742971471,"sale_bill_uuid": 3655262269341697,"sale_order_uuid": 3655262269341699}
	UPDATE_PRODUCT = "update_product" // 更新商品
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(company_uuid uint64, source_client, not_device_id, message_type string, data map[string]interface{}) error {
	// 检查是否需要跳过
	{
		// 计算包含关键参数的MD5值
		key := fmt.Sprintf("%d:%s:%s:%s", company_uuid, source_client, not_device_id, message_type)
		jsonData, _ := json.Marshal(data)
		md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), jsonData...)))
		// 创建缓存键
		cacheKey := fmt.Sprintf("ws_msg:%s", md5Sum)
		// 检查缓存中是否存在相同的消息 -  如果存在相同的消息，跳过发送
		if val, exists := cache.Global.Get(cacheKey); exists && val != nil {
			return nil
		}
		// 设置缓存，3秒过期
		cache.Global.Set(cacheKey, 1, 3*time.Second)
	}

	// 判断当前是否在容器内执行
	url := fmt.Sprintf("http://127.0.0.1:%s/ws/push", os.Getenv("NGINX_PORT"))
	if _, err := os.Stat("/.dockerenv"); err == nil {
		url = "http://nginx/ws/push"
	}
	// 构建请求体
	payload := map[string]interface{}{
		"company_uuid":  company_uuid,
		"source_client": source_client,
		"not_device_id": not_device_id,
		"message_type":  message_type,
		"data":          data, // 直接传递 data 对象，不进行额外的 JSON 编码
	}
	//
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal payload: %v", err)
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}
	//
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Failed to create request: %v", err)
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request: %v", err)
		log.Printf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK response: %v", resp.Status)
		log.Printf("Received non-OK response: %v", resp.Status)
		return fmt.Errorf("received non-OK response: %v", resp.Status)
	}

	return nil
}
