package websocket

import (
	"context"
	"testing"

	v1 "ttpos-bmp/app/ttpos-websocket/api/websocket"
	"ttpos-bmp/app/ttpos-websocket/internal/logic/websocket"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
)

func init() {
	// 注册服务实现
	service.RegisterWebsocket(websocket.New())
}

// TestNew 测试创建控制器
func TestNew(t *testing.T) {
	ctrl := New()
	if ctrl == nil {
		t.Fatal("创建控制器失败")
	}
}

// TestValidatePushMessageReq 测试推送消息参数验证
func TestValidatePushMessageReq(t *testing.T) {
	ctrl := New()

	tests := []struct {
		name        string
		req         *v1.PushMessageReq
		expectError bool
	}{
		{
			name: "有效的请求",
			req: &v1.PushMessageReq{
				CompanyUuid:  1,
				MessageType:  "update_order",
				SourceClient: "cashier",
				DeviceId:     "*",
				Data:         `{"update_time": 1234567890}`,
			},
			expectError: false,
		},
		{
			name: "缺少公司UUID",
			req: &v1.PushMessageReq{
				CompanyUuid:  0,
				MessageType:  "update_order",
				SourceClient: "cashier",
				DeviceId:     "*",
				Data:         `{"update_time": 1234567890}`,
			},
			expectError: true,
		},
		{
			name: "缺少消息类型",
			req: &v1.PushMessageReq{
				CompanyUuid:  1,
				MessageType:  "",
				SourceClient: "cashier",
				DeviceId:     "*",
				Data:         `{"update_time": 1234567890}`,
			},
			expectError: true,
		},
		{
			name: "缺少消息数据",
			req: &v1.PushMessageReq{
				CompanyUuid:  1,
				MessageType:  "update_order",
				SourceClient: "cashier",
				DeviceId:     "*",
				Data:         "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctrl.validatePushMessageReq(tt.req)

			if tt.expectError && err == nil {
				t.Error("期望返回错误，但没有错误")
			}

			if !tt.expectError && err != nil {
				t.Errorf("不期望返回错误，但返回了: %v", err)
			}
		})
	}
}

// TestValidateCheckDeviceOnlineReq 测试检查设备在线参数验证
func TestValidateCheckDeviceOnlineReq(t *testing.T) {
	ctrl := New()

	tests := []struct {
		name        string
		req         *v1.CheckDeviceOnlineReq
		expectError bool
	}{
		{
			name: "有效的请求",
			req: &v1.CheckDeviceOnlineReq{
				CompanyUuid:  1,
				SourceClient: "cashier",
				DeviceId:     "device001",
			},
			expectError: false,
		},
		{
			name: "缺少公司UUID",
			req: &v1.CheckDeviceOnlineReq{
				CompanyUuid:  0,
				SourceClient: "cashier",
				DeviceId:     "device001",
			},
			expectError: true,
		},
		{
			name: "缺少设备ID",
			req: &v1.CheckDeviceOnlineReq{
				CompanyUuid:  1,
				SourceClient: "cashier",
				DeviceId:     "",
			},
			expectError: true,
		},
		{
			name: "可以不提供来源客户端",
			req: &v1.CheckDeviceOnlineReq{
				CompanyUuid:  1,
				SourceClient: "",
				DeviceId:     "device001",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctrl.validateCheckDeviceOnlineReq(tt.req)

			if tt.expectError && err == nil {
				t.Error("期望返回错误，但没有错误")
			}

			if !tt.expectError && err != nil {
				t.Errorf("不期望返回错误，但返回了: %v", err)
			}
		})
	}
}

// TestGetConnectionStats 测试获取连接统计接口
func TestGetConnectionStats(t *testing.T) {
	ctrl := New()
	ctx := context.Background()

	req := &v1.GetConnectionStatsReq{
		CompanyUuid: 0, // 获取所有公司的统计
	}

	resp, err := ctrl.GetConnectionStats(ctx, req)
	if err != nil {
		t.Fatalf("获取连接统计失败: %v", err)
	}

	if resp == nil {
		t.Fatal("响应为空")
	}

	if !resp.Success {
		t.Errorf("响应失败: %s", resp.Message)
	}

	if resp.Stats == nil {
		t.Error("统计信息为空")
	}

	t.Logf("连接统计: 总数=%d, 按公司=%v, 按来源=%v",
		resp.Stats.TotalConnections,
		resp.Stats.ByCompany,
		resp.Stats.BySource)
}

// TestCloseConnection 测试关闭连接接口
func TestCloseConnection(t *testing.T) {
	ctrl := New()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *v1.CloseConnectionReq
		expectError bool
	}{
		{
			name: "关闭指定公司的所有连接",
			req: &v1.CloseConnectionReq{
				CompanyUuid:  999, // 使用不存在的公司ID避免影响其他测试
				SourceClient: "",
				DeviceId:     "",
			},
			expectError: false,
		},
		{
			name: "关闭指定设备的连接",
			req: &v1.CloseConnectionReq{
				CompanyUuid:  999,
				SourceClient: "cashier",
				DeviceId:     "test_device",
			},
			expectError: false,
		},
		{
			name: "缺少公司UUID",
			req: &v1.CloseConnectionReq{
				CompanyUuid:  0,
				SourceClient: "cashier",
				DeviceId:     "test_device",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ctrl.CloseConnection(ctx, tt.req)

			if err != nil {
				t.Fatalf("调用失败: %v", err)
			}

			if tt.expectError && resp.Success {
				t.Error("期望返回错误，但成功了")
			}

			if !tt.expectError && !resp.Success {
				t.Errorf("期望成功，但失败了: %s", resp.Message)
			}

			t.Logf("关闭连接数: %d", resp.ClosedCount)
		})
	}
}
