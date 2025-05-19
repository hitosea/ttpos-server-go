package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// SendSMSRequest 发送短信请求结构体
type SendSMSRequest struct {
	TemplateID string                 `json:"template_id"`
	Phone      string                 `json:"phone"`
	Language   string                 `json:"language"`
	Params     map[string]interface{} `json:"params"`
}

// SMSResponse 短信响应结构体
type SMSResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// smsClient 短信客户端实现
type smsClient struct {
	apiKey      string
	projectName string
	baseURL     string
	httpClient  *http.Client
}

var (
	instance SMSClient
	once     sync.Once
)

// InitClient 初始化短信客户端
func InitClient(apiKey, baseURL, projectName string) {
	once.Do(func() {
		instance = &smsClient{
			apiKey:      apiKey,
			projectName: projectName,
			baseURL:     baseURL,
			httpClient:  &http.Client{},
		}
	})
}

// GetSMSClient 获取短信客户端实例
func GetSMSClient() SMSClient {
	if instance == nil {
		panic("sms client not initialized, please call InitClient first")
	}
	return instance
}

// SendMemberConsumptionSMS 发送会员消费短信
func (c *smsClient) SendMemberConsumptionSMS(phone, language string, params *MemberConsumptionRequest) (*SMSResponse, error) {
	req := &SendSMSRequest{
		TemplateID: TemplateMemberConsumption,
		Phone:      phone,
		Language:   language,
		Params: map[string]interface{}{
			"company":         params.Company,
			"consumption":     params.Consumption,
			"member_pay":      params.MemberPay,
			"increase_points": params.IncreasePoints,
			"balance":         params.Balance,
			"points_balance":  params.PointsBalance,
		},
	}
	if logger.Logger != nil {
		logger.Logger.Info("发送会员消费短信", zap.Any("req", req))
	}
	return c.SendSMS(req)
}

// SendMemberRechargeSMS 发送会员充值短信
func (c *smsClient) SendMemberRechargeSMS(phone, language string, params *MemberRechargeRequest) (*SMSResponse, error) {
	req := &SendSMSRequest{
		TemplateID: TemplateMemberRecharge,
		Phone:      phone,
		Language:   language,
		Params: map[string]interface{}{
			"company":        params.Company,
			"recharge":       params.Recharge,
			"bonus_money":    params.BonusMoney,
			"bonus_points":   params.BonusPoints,
			"balance":        params.Balance,
			"points_balance": params.PointsBalance,
		},
	}
	if logger.Logger != nil {
		logger.Logger.Info("发送会员充值短信", zap.Any("req", req))
	}
	return c.SendSMS(req)
}

// SendMemberRechargeRefundSMS 发送会员充值退款短信
func (c *smsClient) SendMemberRechargeRefundSMS(phone, language string, params *MemberRechargeRefundRequest) (*SMSResponse, error) {
	req := &SendSMSRequest{
		TemplateID: TemplateMemberRechargeRefund,
		Phone:      phone,
		Language:   language,
		Params: map[string]interface{}{
			"company":         params.Company,
			"recharge_refund": params.RechargeRefund,
			"balance":         params.Balance,
			"points_balance":  params.PointsBalance,
		},
	}
	if logger.Logger != nil {
		logger.Logger.Info("发送会员充值退款短信", zap.Any("req", req))
	}
	return c.SendSMS(req)
}

// SendMemberOrderRefundSMS 发送会员用餐订单退款短信
func (c *smsClient) SendMemberOrderRefundSMS(phone, language string, params *MemberOrderRefundRequest) (*SMSResponse, error) {
	req := &SendSMSRequest{
		TemplateID: TemplateMemberOrderRefund,
		Phone:      phone,
		Language:   language,
		Params: map[string]interface{}{
			"company":        params.Company,
			"order_refund":   params.OrderRefund,
			"balance":        params.Balance,
			"points_balance": params.PointsBalance,
		},
	}
	if logger.Logger != nil {
		logger.Logger.Info("发送会员用餐订单退款短信", zap.Any("req", req))
	}
	return c.SendSMS(req)
}

// SendSMS 发送短信
func (c *smsClient) SendSMS(req *SendSMSRequest) (*SMSResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+APIPathSend, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	httpReq.Header.Set("api-key", c.apiKey)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	return &smsResp, nil
}

// CheckConfig 检查客户端配置是否正确
func (c *smsClient) CheckConfig() error {
	if c.apiKey == "" {
		return fmt.Errorf("api key is not configured")
	}
	if c.baseURL == "" {
		return fmt.Errorf("base URL is not configured")
	}
	if c.projectName == "" {
		return fmt.Errorf("project name is not configured")
	}
	if c.httpClient == nil {
		return fmt.Errorf("http client is not configured")
	}

	// 检查服务健康状态
	healthReq, err := http.NewRequest("GET", c.baseURL+APIPathHealth, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	healthResp, err := c.httpClient.Do(healthReq)
	if err != nil {
		return fmt.Errorf("failed to connect to SMS service: %v", err)
	}
	defer healthResp.Body.Close()

	if healthResp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS service is not healthy, status code: %d", healthResp.StatusCode)
	}

	// 验证 API key
	checkKeyURL := fmt.Sprintf("%s%s?project_name=%s&api_key=%s", c.baseURL, APIPathCheckKey, c.projectName, c.apiKey)
	keyReq, err := http.NewRequest("GET", checkKeyURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create API key check request: %v", err)
	}

	keyResp, err := c.httpClient.Do(keyReq)
	if err != nil {
		return fmt.Errorf("failed to check API key: %v", err)
	}
	defer keyResp.Body.Close()

	var smsResp SMSResponse
	if err := json.NewDecoder(keyResp.Body).Decode(&smsResp); err != nil {
		return fmt.Errorf("failed to decode API key check response: %v", err)
	}

	if smsResp.Code != ResponseCodeSuccess {
		return fmt.Errorf("invalid API key, code: %d, msg: %s. project name: %s, API key: %s", smsResp.Code, smsResp.Msg, c.projectName, c.apiKey)
	}

	return nil
}
