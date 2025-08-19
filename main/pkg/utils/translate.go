package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TranslateItem 翻译项目结构
type TranslateItem struct {
	Lang    string `json:"lang"`    // 语言代码，如：zh, en, ja等
	Content string `json:"content"` // 要翻译的内容
}

// TranslateRequest 翻译请求结构
type TranslateRequest struct {
	Data []TranslateItem `json:"data"`
}

// TranslateResponseItem 翻译响应项目结构
type TranslateResponseItem struct {
	Key  string `json:"key"`   // 原始内容
	Zh   string `json:"zh"`    // 翻译后的中文
	ZhTw string `json:"zh-TW"` // 翻译后的中文繁体
	Tr   string `json:"tr"`    // 翻译后的土耳其语
	Th   string `json:"th"`    // 翻译后的泰语
	Ja   string `json:"ja"`    // 翻译后的日语
	Ko   string `json:"ko"`    // 翻译后的韩语
	En   string `json:"en"`    // 翻译后的英语
	My   string `json:"my"`    // 翻译后的缅甸语
	De   string `json:"de"`    // 翻译后的德语
	Sv   string `json:"sv"`    // 翻译后的瑞典语
}

// TranslateResponse 翻译响应结构
type TranslateResponse struct {
	Code    int                     `json:"code"`    // 响应码
	Message string                  `json:"message"` // 响应消息
	Data    []TranslateResponseItem `json:"data"`    // 翻译结果数据
}

// TranslateClient 翻译客户端
type TranslateClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewTranslateClient 创建新的翻译客户端
func NewTranslateClient() *TranslateClient {
	return &TranslateClient{
		BaseURL: "https://aitrans.ttpos.com",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 忽略SSL证书验证
				},
			},
		},
		Timeout: 30 * time.Second,
	}
}

// NewTranslateClientWithTimeout 创建带自定义超时的翻译客户端
func NewTranslateClientWithTimeout(timeout time.Duration) *TranslateClient {
	client := NewTranslateClient()
	client.Timeout = timeout
	client.HTTPClient.Timeout = timeout
	return client
}

// Translate 执行翻译请求
func (tc *TranslateClient) Translate(ctx context.Context, items []TranslateItem) (*TranslateResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("翻译项目不能为空")
	}

	// 构建请求数据
	reqData := TranslateRequest{
		Data: items,
	}

	// JSON序列化
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/translate", tc.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TTPOS-TranslateClient/1.0")

	// 发起HTTP请求
	resp, err := tc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发起HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("翻译请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应JSON
	var translateResp TranslateResponse
	if err := json.Unmarshal(respBody, &translateResp); err != nil {
		return nil, fmt.Errorf("解析响应JSON失败: %v", err)
	}

	return &translateResp, nil
}

// TranslateText 翻译单个文本（便捷方法）
func (tc *TranslateClient) TranslateText(ctx context.Context, lang, content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("翻译内容不能为空")
	}

	items := []TranslateItem{
		{
			Lang:    lang,
			Content: content,
		},
	}

	resp, err := tc.Translate(ctx, items)
	if err != nil {
		return "", err
	}

	// 检查响应状态
	if resp.Code != 1 && resp.Code != 200 && resp.Code != 0 { // 常见的成功状态码
		return "", fmt.Errorf("翻译失败: %s (code: %d)", resp.Message, resp.Code)
	}

	// 检查是否有翻译结果
	if len(resp.Data) == 0 {
		return "", fmt.Errorf("没有翻译结果")
	}

	return resp.Data[0].Zh, nil
}

// TranslateMultiple 翻译多个文本项目
func (tc *TranslateClient) TranslateMultiple(ctx context.Context, items []TranslateItem) (map[string]string, error) {
	resp, err := tc.Translate(ctx, items)
	if err != nil {
		return nil, err
	}

	// 检查响应状态
	if resp.Code != 1 && resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("翻译失败: %s (code: %d)", resp.Message, resp.Code)
	}

	// 构建结果映射 lang -> translated_content
	result := make(map[string]string)
	for _, item := range resp.Data {
		result[item.Key] = item.Zh
	}

	return result, nil
}

// 全局默认翻译客户端
var defaultTranslateClient = NewTranslateClient()

// Translate 使用默认客户端进行翻译（便捷函数）
func Translate(ctx context.Context, items []TranslateItem) (*TranslateResponse, error) {
	return defaultTranslateClient.Translate(ctx, items)
}

// TranslateText 使用默认客户端翻译单个文本（便捷函数）
func TranslateText(ctx context.Context, lang, content string) (string, error) {
	return defaultTranslateClient.TranslateText(ctx, lang, content)
}

// TranslateMultiple 使用默认客户端翻译多个文本（便捷函数）
func TranslateMultiple(ctx context.Context, items []TranslateItem) (map[string]string, error) {
	return defaultTranslateClient.TranslateMultiple(ctx, items)
}

// TranslateToMultipleLanguages 将一段内容翻译为多种语言（便捷函数）
func TranslateToMultipleLanguages(ctx context.Context, sourceContent string, targetLangs []string) (map[string]string, error) {
	if sourceContent == "" {
		return nil, fmt.Errorf("源内容不能为空")
	}
	if len(targetLangs) == 0 {
		return nil, fmt.Errorf("目标语言列表不能为空")
	}

	// 构建翻译项目
	items := make([]TranslateItem, len(targetLangs))
	for i, lang := range targetLangs {
		items[i] = TranslateItem{
			Lang:    lang,
			Content: sourceContent,
		}
	}

	return TranslateMultiple(ctx, items)
}
