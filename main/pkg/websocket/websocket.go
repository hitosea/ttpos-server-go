package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	UPDATE_ORDER   = "update_order"   // 更新订单 data = {"timestamp": 1742971471,"sale_bill_uuid": 3655262269341697,"sale_order_uuid": 3655262269341699}
	UPDATE_PRODUCT = "update_product" // 更新商品
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(company_uuid uint64, source_client, not_device_id, message_type string, data map[string]interface{}) error {
	// 判断当前是否在容器内执行
	url := fmt.Sprintf("http://127.0.0.1:%s/ws/push", "8099")
	if _, err := os.Stat("/.dockerenv"); err == nil {
		url = "http://nginx/ws/push"
	}
	//
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
