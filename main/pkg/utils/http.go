package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HttpPost
func HttpPost(url string, payload map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
	// 构建请求体
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Failed to create request: %v", err)
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	// 设置默认 headers
	req.Header.Set("Content-Type", "application/json")
	// 设置自定义 headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	//
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request: %v", err)
		log.Printf("Failed to send request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK response: %s %s", resp.Status, url)
		log.Printf("Received non-OK response: %s %s", resp.Status, url)
		return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
	}
	// 解析响应
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Failed to decode response: %v", err)
		log.Printf("Failed to decode response: %v", err)
		return nil, err
	}
	return result, nil
}

// HttpGet
func HttpGet(url string, payload map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
	// 构建查询参数
	query := ""
	if len(payload) > 0 {
		values := make([]string, 0)
		for key, value := range payload {
			values = append(values, fmt.Sprintf("%v=%v", key, value))
		}
		query = "?" + strings.Join(values, "&")
	}

	// 创建请求
	req, err := http.NewRequest("GET", url+query, nil)
	if err != nil {
		fmt.Println("Failed to create request: %v", err)
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	// 设置自定义 headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	//
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request: %v", err)
		log.Printf("Failed to send request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK response: %s %s", resp.Status, url)
		log.Printf("Received non-OK response: %s %s", resp.Status, url)
		return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
	}
	// 解析响应
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Failed to decode response: %v", err)
		log.Printf("Failed to decode response: %v", err)
		return nil, err
	}
	return result, nil
}
