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
	UPDATE_PRODUCT = "update_product" // 更新商品
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(company_id uint, source_client, device_id, message_type string, data any) error {
	// 判断当前是否在容器内执行
	url := fmt.Sprintf("http://localhost:%s/ws/push", os.Getenv("NGINX_PORT"))
	if _, err := os.Stat("/.dockerenv"); err == nil {
		url = "http://nginx/ws/push"
	}
	//
	payload := map[string]interface{}{
		"company_id":    company_id,
		"source_client": source_client,
		"device_id":     device_id,
		"message_type":  message_type,
		"data":          data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK response: %v", resp.Status)
		return fmt.Errorf("received non-OK response: %v", resp.Status)
	}

	return nil
}
