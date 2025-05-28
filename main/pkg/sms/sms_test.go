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
	InitClient("test-api-key", server.URL, "ttpos")

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

	// 测试检查配置
	t.Run("TestCheckConfig", func(t *testing.T) {
		// 创建测试服务器
		configServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 健康检查接口
			if r.URL.Path == APIPathHealth {
				w.WriteHeader(http.StatusOK)
				return
			}

			// API key 检查接口
			if r.URL.Path == APIPathCheckKey {
				projectName := r.URL.Query().Get("project_name")
				apiKey := r.URL.Query().Get("api_key")

				if projectName != "ttpos" || apiKey != "test-api-key" {
					json.NewEncoder(w).Encode(SMSResponse{
						Code: ResponseCodeInvalidAPIKey,
						Msg:  "Invalid API Key",
					})
					return
				}

				json.NewEncoder(w).Encode(SMSResponse{
					Code: ResponseCodeSuccess,
					Msg:  "Exists",
				})
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
		defer configServer.Close()

		// 测试正常配置
		client := &smsClient{
			apiKey:      "test-api-key",
			baseURL:     configServer.URL,
			projectName: "ttpos",
			httpClient:  &http.Client{},
		}
		if err := client.CheckConfig(); err != nil {
			t.Errorf("CheckConfig failed with valid config: %v", err)
		}

		// 测试缺少API密钥
		client.apiKey = ""
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail with missing API key")
		} else if err.Error() != "api key is not configured" {
			t.Errorf("Expected error message 'api key is not configured', got: %v", err)
		}

		// 测试无效的API密钥
		client.apiKey = "invalid-api-key"
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail with invalid API key")
		} else if err.Error() != "invalid API key, code: 401, msg: Invalid API Key. project name: ttpos, API key: invalid-api-key" {
			t.Errorf("Expected error message about invalid API key, got: %v", err)
		}

		// 测试缺少baseURL
		client.apiKey = "test-api-key"
		client.baseURL = ""
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail with missing base URL")
		} else if err.Error() != "base URL is not configured" {
			t.Errorf("Expected error message 'base URL is not configured', got: %v", err)
		}

		// 测试缺少projectName
		client.baseURL = configServer.URL
		client.projectName = ""
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail with missing project name")
		} else if err.Error() != "project name is not configured" {
			t.Errorf("Expected error message 'project name is not configured', got: %v", err)
		}

		// 测试缺少httpClient
		client.projectName = "ttpos"
		client.httpClient = nil
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail with missing http client")
		} else if err.Error() != "http client is not configured" {
			t.Errorf("Expected error message 'http client is not configured', got: %v", err)
		}

		// 测试健康检查失败
		unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer unhealthyServer.Close()

		client.httpClient = &http.Client{}
		client.baseURL = unhealthyServer.URL
		if err := client.CheckConfig(); err == nil {
			t.Error("CheckConfig should fail when health check fails")
		} else if err.Error() != "SMS service is not healthy, status code: 500" {
			t.Errorf("Expected error message about unhealthy service, got: %v", err)
		}
	})
}
