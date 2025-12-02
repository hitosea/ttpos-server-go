package websocket

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/grpc"

	v1 "ttpos-bmp/app/ttpos-websocket/api/websocket"
	"ttpos-server-go/app/cloud"
)

// GrpcClient WebSocket gRPC 客户端
type GrpcClient struct {
	conn   *grpc.ClientConn
	client v1.WebSocketClient
}

var defaultClient *GrpcClient

// NewWebSocketClient 创建 WebSocket gRPC 客户端（使用 Nacos 服务发现）
// 返回：WebSocket客户端、gRPC连接、错误信息
func NewWebSocketClient() (v1.WebSocketClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.WebSocketServiceName)
	if err != nil {
		return nil, nil, err
	}
	return v1.NewWebSocketClient(conn), conn, nil
}

// PushClient 推送消息的便捷函数（保持向后兼容）
// 参数：
//   - companyUuid: 公司UUID
//   - sourceClient: 来源客户端
//   - deviceId: 设备ID
//   - messageType: 消息类型
//   - data: 消息数据
func PushClient(companyUuid uint64, sourceClient, deviceId, messageType string, data map[string]interface{}) error {
	ctx := context.Background()

	// 如果全局客户端已初始化，使用全局客户端
	if defaultClient != nil {
		return defaultClient.PushMessage(ctx, companyUuid, sourceClient, deviceId, messageType, data)
	}

	// 否则每次创建新的客户端连接（推荐在启动时初始化全局客户端）
	client, conn, err := NewWebSocketClient()
	if err != nil {
		return fmt.Errorf("创建 WebSocket gRPC 客户端失败: %w", err)
	}
	defer conn.Close()

	// 将 data 转为 JSON 字符串
	messageKey, dataJSON := getMessageKeyAndDataJSON(companyUuid, sourceClient, deviceId, messageType, data)

	// 调用 gRPC 接口
	resp, err := client.PushMessage(ctx, &v1.PushMessageReq{
		CompanyUuid:  companyUuid,
		StaffUuid:    0,
		NotStaffUuid: 0,
		SourceClient: sourceClient,
		DeviceId:     deviceId,
		NotDeviceId:  "",
		MessageType:  messageType,
		MessageKey:   messageKey,
		Data:         string(dataJSON),
	})

	if err != nil {
		return fmt.Errorf("调用 PushMessage gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("推送消息失败: %s", resp.Message)
	}

	return nil
}

// PushMessage 推送消息到 WebSocket 客户端
// 参数：
//   - ctx: 上下文对象
//   - companyUuid: 公司UUID
//   - sourceClient: 来源客户端（shop/cashier/tablet/kitchen/assistant/H5/*）
//   - deviceId: 设备ID，*表示所有设备
//   - messageType: 消息类型
//   - data: 消息数据（map类型，会自动转为JSON）
func (c *GrpcClient) PushMessage(ctx context.Context, companyUuid uint64, sourceClient, deviceId, messageType string, data map[string]interface{}) error {
	// 将 data 转为 JSON 字符串
	messageKey, dataJSON := getMessageKeyAndDataJSON(companyUuid, sourceClient, deviceId, messageType, data)

	// 调用 gRPC 接口
	resp, err := c.client.PushMessage(ctx, &v1.PushMessageReq{
		CompanyUuid:  companyUuid,
		StaffUuid:    0, // 0表示推送给所有员工
		NotStaffUuid: 0,
		SourceClient: sourceClient,
		DeviceId:     deviceId,
		NotDeviceId:  "",
		MessageKey:   messageKey,
		MessageType:  messageType,
		Data:         string(dataJSON),
	})

	if err != nil {
		return fmt.Errorf("调用 PushMessage gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("推送消息失败: %s", resp.Message)
	}

	return nil
}

// PushOptions 推送选项
type PushOptions struct {
	CompanyUuid  uint64                 // 公司UUID，必填
	StaffUuid    uint64                 // 员工UUID，0表示推送给所有员工
	NotStaffUuid uint64                 // 排除的员工UUID
	SourceClient string                 // 来源客户端
	DeviceId     string                 // 设备ID，*表示所有设备
	NotDeviceId  string                 // 排除的设备ID
	MessageType  string                 // 消息类型
	Data         map[string]interface{} // 消息数据
}

// PushMessageWithOptions 推送消息到 WebSocket 客户端（高级选项）
// 参数：
//   - ctx: 上下文对象
//   - options: 推送选项
func (c *GrpcClient) PushMessageWithOptions(ctx context.Context, options *PushOptions) error {
	// 将 data 转为 JSON 字符串
	dataJSON, err := json.Marshal(options.Data)
	if err != nil {
		return fmt.Errorf("序列化消息数据失败: %w", err)
	}

	// 调用 gRPC 接口
	resp, err := c.client.PushMessage(ctx, &v1.PushMessageReq{
		CompanyUuid:  options.CompanyUuid,
		StaffUuid:    options.StaffUuid,
		NotStaffUuid: options.NotStaffUuid,
		SourceClient: options.SourceClient,
		DeviceId:     options.DeviceId,
		NotDeviceId:  options.NotDeviceId,
		MessageType:  options.MessageType,
		Data:         string(dataJSON),
	})

	if err != nil {
		return fmt.Errorf("调用 PushMessage gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("推送消息失败: %s", resp.Message)
	}

	return nil
}

// GetConnectionStats 获取连接统计信息
func (c *GrpcClient) GetConnectionStats(ctx context.Context, companyUuid uint64) (*v1.ConnectionStats, error) {
	resp, err := c.client.GetConnectionStats(ctx, &v1.GetConnectionStatsReq{
		CompanyUuid: companyUuid,
	})

	if err != nil {
		return nil, fmt.Errorf("调用 GetConnectionStats gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("获取连接统计失败: %s", resp.Message)
	}

	return resp.Stats, nil
}

// CheckDeviceOnline 检查设备是否在线
func (c *GrpcClient) CheckDeviceOnline(ctx context.Context, companyUuid uint64, sourceClient, deviceId string) (bool, error) {
	resp, err := c.client.CheckDeviceOnline(ctx, &v1.CheckDeviceOnlineReq{
		CompanyUuid:  companyUuid,
		SourceClient: sourceClient,
		DeviceId:     deviceId,
	})

	if err != nil {
		return false, fmt.Errorf("调用 CheckDeviceOnline gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return false, fmt.Errorf("检查设备在线状态失败: %s", resp.Message)
	}

	return resp.IsOnline, nil
}

// CloseConnection 关闭指定连接
func (c *GrpcClient) CloseConnection(ctx context.Context, companyUuid uint64, sourceClient, deviceId string) (int32, error) {
	resp, err := c.client.CloseConnection(ctx, &v1.CloseConnectionReq{
		CompanyUuid:  companyUuid,
		SourceClient: sourceClient,
		DeviceId:     deviceId,
	})

	if err != nil {
		return 0, fmt.Errorf("调用 CloseConnection gRPC 接口失败: %w", err)
	}

	if !resp.Success {
		return 0, fmt.Errorf("关闭连接失败: %s", resp.Message)
	}

	return resp.ClosedCount, nil
}

// 生成消息键和数据JSON
func getMessageKeyAndDataJSON(
	companyUuid uint64,
	sourceClient string,
	deviceId string,
	messageType string,
	data map[string]interface{},
) (string, []byte) {
	// 将 data 转为 JSON 字符串
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", nil
	}

	var jsonData []byte
	// 计算包含关键参数的MD5值
	if messageType == UPDATE_ORDER {
		jsonData = fmt.Appendf(nil, "%d", data["sale_bill_uuid"])
	}

	// 创建缓存键
	var cacheKey string
	// 更新桌台的时候
	if messageType == UPDATE_DESK {
		key := fmt.Sprintf("%d:%s:%s:%s", companyUuid, sourceClient, deviceId, messageType)
		md5Sum := fmt.Sprintf("%x", md5.Sum([]byte(key)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	} else if messageType == PRINT_DATA {
		key := fmt.Sprintf("%d:%s:%s:%s", companyUuid, sourceClient, deviceId, messageType)
		var updateTime string
		// 安全地获取更新时间，支持多种可能的键名和类型
		if updateTimeVal, exists := data["update_times"]; exists {
			switch v := updateTimeVal.(type) {
			case int64:
				updateTime = strconv.FormatInt(v, 10)
			case int:
				updateTime = strconv.FormatInt(int64(v), 10)
			case string:
				updateTime = v
			default:
				updateTime = fmt.Sprintf("%v", v)
			}
		}
		md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), []byte(updateTime)...)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	} else {
		key := fmt.Sprintf("%d:%s:%s:%s", companyUuid, sourceClient, deviceId, messageType)
		md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), jsonData...)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	}

	return cacheKey, dataJSON
}
