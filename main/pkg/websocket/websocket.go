package websocket

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	// 更新订单，刷新购物车和桌台列表都用它（desk_uuid不等于0代表是桌台订单，desk_uuid等于0代表是点餐订单），data = {"update_time": 1742971471,"sale_bill_uuid": 3655262269341697,"desk_uuid": 3655262269341699}
	UPDATE_ORDER = "update_order"
	// 更新商品
	UPDATE_PRODUCT = "update_product"
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(company_uuid uint64, source_client, not_device_id, message_type string, data map[string]interface{}) error {
	var jsonData []byte
	var err error
	// 计算包含关键参数的MD5值
	if message_type == UPDATE_ORDER {
		jsonData, err = json.Marshal(map[string]interface{}{
			"sale_bill_uuid": data["sale_bill_uuid"],
		})
	} else if message_type == UPDATE_PRODUCT {
		jsonData, err = json.Marshal(map[string]interface{}{
			"sale_bill_uuid": data["sale_bill_uuid"],
		})
	}
	if err != nil {
		fmt.Println("Failed to marshal JSON: %v", err)
		log.Printf("Failed to marshal JSON: %v", err)
		return err
	}

	// 创建缓存键
	key := fmt.Sprintf("%d:%s:%s:%s", company_uuid, source_client, not_device_id, message_type)
	md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), jsonData...)))
	cacheKey := fmt.Sprintf("ws_msg:%s", md5Sum)

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
		"message_key":   cacheKey,
		"data":          data,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal payload: %v", err)
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

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
