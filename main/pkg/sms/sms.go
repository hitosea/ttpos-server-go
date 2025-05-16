package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

var (
	instance SMSClient
	once     sync.Once
)

// InitClient 初始化短信客户端
func InitClient(apiKey, baseURL string) {
	once.Do(func() {
		instance = &smsClient{
			apiKey:     apiKey,
			baseURL:    baseURL,
			httpClient: &http.Client{},
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

// QuerySMSStatus 查询短信状态
func (c *smsClient) QuerySMSStatus(messageIDs string) (*SMSResponse, error) {
	url := fmt.Sprintf("%s%s?message_id=%s", c.baseURL, APIPathQuery, messageIDs)

	httpReq, err := http.NewRequest("GET", url, nil)
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
