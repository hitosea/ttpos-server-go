package sms

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendSMS(t *testing.T) {
	// 初始化测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		if r.Header.Get("api-key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(SMSResponse{
				Code: ResponseCodeInvalidAPIKey,
				Msg:  "Invalid API Key",
			})
			return
		}

		// 验证请求体
		var req SendSMSRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SMSResponse{
				Code: 400,
				Msg:  "Invalid request body",
			})
			return
		}

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SMSResponse{
			Code: ResponseCodeSuccess,
			Msg:  "success",
		})
	}))
	defer server.Close()

	// 初始化客户端
	InitClient("test-api-key", server.URL)

	// 测试会员充值短信
	t.Run("TestSendMemberRechargeSMS", func(t *testing.T) {
		params := &MemberRechargeRequest{
			Company:       "测试餐厅",
			Recharge:      100,
			BonusMoney:    50,
			BonusPoints:   100,
			Balance:       100,
			PointsBalance: 100,
		}

		resp, err := GetSMSClient().SendMemberRechargeSMS("+8617777777777", LanguageChinese, params)
		if err != nil {
			t.Errorf("SendMemberRechargeSMS failed: %v", err)
		}
		if resp.Code != ResponseCodeSuccess {
			t.Errorf("Expected success response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})

	// 测试会员消费短信
	t.Run("TestSendMemberConsumptionSMS", func(t *testing.T) {
		params := &MemberConsumptionRequest{
			Company:        "测试餐厅",
			Consumption:    200,
			MemberPay:      150,
			IncreasePoints: 50,
			Balance:        100,
			PointsBalance:  100,
		}

		resp, err := GetSMSClient().SendMemberConsumptionSMS("+8617777777777", LanguageChinese, params)
		if err != nil {
			t.Errorf("SendMemberConsumptionSMS failed: %v", err)
		}
		if resp.Code != ResponseCodeSuccess {
			t.Errorf("Expected success response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})

	// 测试会员充值退款短信
	t.Run("TestSendMemberRechargeRefundSMS", func(t *testing.T) {
		params := &MemberRechargeRefundRequest{
			Company:        "测试餐厅",
			RechargeRefund: 100,
			Balance:        100,
			PointsBalance:  100,
		}

		resp, err := GetSMSClient().SendMemberRechargeRefundSMS("+8617777777777", LanguageChinese, params)
		if err != nil {
			t.Errorf("SendMemberRechargeRefundSMS failed: %v", err)
		}
		if resp.Code != ResponseCodeSuccess {
			t.Errorf("Expected success response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})

	// 测试会员用餐订单退款短信
	t.Run("TestSendMemberOrderRefundSMS", func(t *testing.T) {
		params := &MemberOrderRefundRequest{
			Company:       "测试餐厅",
			OrderRefund:   100,
			Balance:       100,
			PointsBalance: 100,
		}

		resp, err := GetSMSClient().SendMemberOrderRefundSMS("+8617777777777", LanguageChinese, params)
		if err != nil {
			t.Errorf("SendMemberOrderRefundSMS failed: %v", err)
		}
		if resp.Code != ResponseCodeSuccess {
			t.Errorf("Expected success response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})

	// 测试查询短信状态
	t.Run("TestQuerySMSStatus", func(t *testing.T) {
		// 创建新的测试服务器专门处理查询请求
		queryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求方法
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(SMSResponse{
					Code: 405,
					Msg:  "Method not allowed",
				})
				return
			}

			// 验证请求头
			if r.Header.Get("api-key") != "test-api-key" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(SMSResponse{
					Code: ResponseCodeInvalidAPIKey,
					Msg:  "Invalid API Key",
				})
				return
			}

			// 验证URL参数
			messageID := r.URL.Query().Get("message_id")
			if messageID == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(SMSResponse{
					Code: 400,
					Msg:  "Missing message_id parameter",
				})
				return
			}

			// 返回成功响应
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SMSResponse{
				Code: ResponseCodeSuccess,
				Msg:  "success",
			})
		}))
		defer queryServer.Close()

		// 直接创建 smsClient 实例
		client := &smsClient{
			apiKey:     "test-api-key",
			baseURL:    queryServer.URL,
			httpClient: &http.Client{},
		}

		// 测试正常查询
		resp, err := client.QuerySMSStatus("test-message-id")
		if err != nil {
			t.Errorf("QuerySMSStatus failed: %v", err)
		}
		if resp.Code != ResponseCodeSuccess {
			t.Errorf("Expected success response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}

		// 测试缺少message_id参数
		resp, err = client.QuerySMSStatus("")
		if err != nil {
			t.Errorf("QuerySMSStatus failed: %v", err)
		}
		if resp.Code != 400 {
			t.Errorf("Expected bad request response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}

		// 测试无效的API密钥
		invalidClient := &smsClient{
			apiKey:     "invalid-api-key",
			baseURL:    queryServer.URL,
			httpClient: &http.Client{},
		}
		resp, err = invalidClient.QuerySMSStatus("test-message-id")
		if err != nil {
			t.Errorf("QuerySMSStatus failed: %v", err)
		}
		if resp.Code != ResponseCodeInvalidAPIKey {
			t.Errorf("Expected invalid API key response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})

	// 测试无效的API密钥
	t.Run("TestInvalidAPIKey", func(t *testing.T) {
		// 创建新的短信客户端实例，使用错误的API密钥
		client := &smsClient{
			apiKey:     "invalid-api-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
		}

		params := &MemberRechargeRequest{
			Company:       "测试餐厅",
			Recharge:      100,
			BonusMoney:    50,
			BonusPoints:   100,
			Balance:       100,
			PointsBalance: 100,
		}

		resp, err := client.SendMemberRechargeSMS("+8617777777777", LanguageChinese, params)
		if err != nil {
			t.Errorf("SendMemberRechargeSMS failed: %v", err)
		}
		if resp.Code != ResponseCodeInvalidAPIKey {
			t.Errorf("Expected invalid API key response, got code: %d, msg: %s", resp.Code, resp.Msg)
		}
	})
}
