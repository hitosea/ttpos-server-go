// Package websocket 实现 WebSocket 服务的业务逻辑
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"

	"ttpos-bmp/app/ttpos-websocket/internal/consts"
	"ttpos-bmp/app/ttpos-websocket/internal/dao"
	"ttpos-bmp/app/ttpos-websocket/internal/model/do"
	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
	"ttpos-bmp/app/ttpos-websocket/internal/model/entity"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
	"ttpos-bmp/app/ttpos-websocket/utility"
	"ttpos-bmp/internal/pkg/queue"

	"ttpos-api/ttpos-websocket/constant"
	ttposWebsocketMsg "ttpos-api/ttpos-websocket/message"
)

// sWebSocket WebSocket 业务逻辑实现
type sWebSocket struct{}

func init() {
	service.RegisterWebsocket(New())
}

// New 创建 WebSocket 服务实例
func New() *sWebSocket {
	return &sWebSocket{}
}

// WebSocket 升级器配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ConnectionInfo WebSocket 连接信息
type ConnectionInfo struct {
	CompanyUuid   uint64          `json:"company_uuid"`   // 公司UUID
	SourceClient  string          `json:"source_client"`  // 来源客户端
	DeviceId      string          `json:"device_id"`      // 设备ID
	StaffUuid     uint64          `json:"staff_uuid"`     // 员工UUID
	LastHeartbeat string          `json:"last_heartbeat"` // 最后心跳时间
	ws            *websocket.Conn // WebSocket 连接
	writeMutex    *sync.Mutex     // 写入锁
}

// PushMessage 推送消息结构
type PushMessage struct {
	Event string      `json:"event"`            // 事件名称
	State int         `json:"state"`            // 状态码
	Msg   string      `json:"msg"`              // 消息描述
	Data  interface{} `json:"data,omitempty"`   // 消息数据
	MsgId int         `json:"msg_id,omitempty"` // 消息ID
}

// ClientMessage 客户端消息结构
type ClientMessage struct {
	Type  string      `json:"type"`             // 消息类型
	MsgId interface{} `json:"msg_id,omitempty"` // 消息ID
	Data  interface{} `json:"data"`             // 消息数据
}

// UsbPrinter USB打印机信息
type UsbPrinter struct {
	MName string      `json:"m_name"` // 制造商名称
	Name  string      `json:"name"`   // 打印机名称
	Pid   interface{} `json:"pid"`    // 产品ID
	Sn    string      `json:"sn"`     // 序列号
	Vid   interface{} `json:"vid"`    // 供应商ID
}

// LanPrinter LAN打印机信息
type LanPrinter struct {
	Ip     string `json:"ip"`     // IP地址
	Port   int    `json:"port"`   // 端口
	Status int    `json:"status"` // 状态
	Remark string `json:"remark"` // 备注
}

// WsClients 存储所有的连接 - 使用sync.Map提高并发安全性
var WsClients sync.Map

// 全局context用于控制协程生命周期
var globalCtx context.Context
var globalCancel context.CancelFunc

func init() {
	globalCtx, globalCancel = context.WithCancel(context.Background())
	// 启动定时检查心跳超时的连接
	go checkHeartbeatTimeout()
}

// 防抖相关的全局变量
var (
	// 基于MessageKey的细粒度锁映射
	messageKeyMutexes = sync.Map{} // map[string]*sync.Mutex
	// 保护messageKeyMutexes的锁
	mutexMapLock sync.Mutex
	// 取消等待的 channel 映射（用于提前取消定时器）
	cancelChannels = sync.Map{} // map[string]chan struct{}
)

// PushMessage 推送消息到客户端
// 根据条件筛选连接并推送消息
// 参数：
//   - ctx: 上下文对象
//   - in: 推送消息输入参数
//
// 返回：
//   - out: 推送消息输出参数
//   - err: 错误信息
func (s *sWebSocket) PushMessage(ctx context.Context, in *dto.PushMessageInput) (out *dto.PushMessageOutput, err error) {
	out = &dto.PushMessageOutput{}

	// 参数验证
	if in.CompanyUuid == 0 {
		return out, gerror.New("公司UUID不能为空")
	}
	if in.MessageType == "" {
		return out, gerror.New("消息类型不能为空")
	}

	g.Log().Info(ctx, "推送消息",
		"company_uuid", in.CompanyUuid,
		"message_type", in.MessageType,
		"source_client", in.SourceClient,
		"device_id", in.DeviceId,
		"message_key", in.MessageKey,
	)

	// 如果设置了 MessageKey，启用防抖机制
	if in.MessageKey != "" {
		// 异步处理防抖逻辑
		go s.handleDebouncedPush(ctx, in)

		out.Success = true
		out.Message = "消息已加入防抖队列"
		out.PushCount = 0
		return out, nil
	}

	// 没有 MessageKey，直接推送
	return s.directPushMessage(ctx, in)
}

// directPushMessage 直接推送消息（无防抖）
// 注意：在集群环境下，通过 Redis 发布消息到所有节点
func (s *sWebSocket) directPushMessage(ctx context.Context, in *dto.PushMessageInput) (out *dto.PushMessageOutput, err error) {
	out = &dto.PushMessageOutput{}

	// 将消息序列化为 JSON
	dataBytes, err := json.Marshal(in)
	if err != nil {
		return out, gerror.Wrap(err, "序列化消息失败")
	}

	// 通过 Redis 发布消息到所有节点
	_, err = g.Redis().Publish(ctx, consts.ChannelWebsocketMsgPush, string(dataBytes))
	if err != nil {
		g.Log().Error(ctx, "发布消息到Redis失败", err)
		return out, gerror.Wrap(err, "发布消息失败")
	}

	g.Log().Info(ctx, "消息已发布到Redis", "message_type", in.MessageType)

	out.Success = true
	out.Message = "消息已发布"
	out.PushCount = 0 // 由订阅者处理，这里无法统计

	return out, nil
}

// processPushMessage 处理推送消息（由 Redis 订阅者调用）
// 这个方法在每个节点上执行，推送到本节点的 WebSocket 连接
func (s *sWebSocket) processPushMessage(ctx context.Context, in *dto.PushMessageInput) error {
	// 1. 先删后加 - 同一个类型，只保留最新的
	if err := s.deleteOldMessages(ctx, in); err != nil {
		g.Log().Warning(ctx, "删除旧消息失败", err)
	}

	// 2. 收集匹配的连接（只在本节点）
	matchedConnections := s.collectMatchedConnections(in)

	// 3. 批量处理匹配的连接
	pushCount := 0
	for _, conn := range matchedConnections {
		if err := s.pushToConnection(ctx, conn, in); err != nil {
			g.Log().Warning(ctx, "推送到连接失败", err, "device_id", conn.DeviceId)
		} else {
			pushCount++
		}
	}

	return nil
}

// handleDebouncedPush 处理防抖推送
// 实现防抖机制：在900ms内如果有相同MessageKey的新消息，则取消旧消息的推送
// 使用 Timer + Channel 替代 Sleep，支持提前取消等待
func (s *sWebSocket) handleDebouncedPush(ctx context.Context, in *dto.PushMessageInput) {
	// 创建独立的context用于防抖逻辑，避免因原始context取消导致Redis操作失败
	// 保留原始context的trace信息用于日志关联
	debounceCtx := context.Background()

	// 生成唯一的UUID标识本次推送
	pushUUID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 获取MessageKey专用锁
	keyMutex := getMessageKeyMutex(in.MessageKey)
	keyMutex.Lock()

	// 如果存在旧的等待 channel，关闭它以取消旧的等待
	if oldChan, exists := cancelChannels.Load(in.MessageKey); exists {
		close(oldChan.(chan struct{}))
		g.Log().Debug(debounceCtx, "取消旧的防抖等待", "message_key", in.MessageKey)
	}

	// 创建新的取消 channel
	cancelChan := make(chan struct{})
	cancelChannels.Store(in.MessageKey, cancelChan)

	// 设置当前推送的UUID到Redis，过期时间2秒
	_, err := g.Redis().Set(debounceCtx, in.MessageKey, pushUUID)
	if err != nil {
		g.Log().Error(debounceCtx, "设置Redis防抖键失败", err, "message_key", in.MessageKey)
		keyMutex.Unlock()
		return
	}
	_, err = g.Redis().Expire(debounceCtx, in.MessageKey, 2)
	if err != nil {
		g.Log().Error(debounceCtx, "设置Redis过期时间失败", err, "message_key", in.MessageKey)
	}

	// 使用 Redis INCR 原子操作增加计数器（避免竞态条件）
	countKey := fmt.Sprintf("%s_count", in.MessageKey)
	_, err = g.Redis().Incr(debounceCtx, countKey)
	if err != nil {
		g.Log().Error(debounceCtx, "增加计数器失败", err, "count_key", countKey)
		keyMutex.Unlock()
		return
	}

	// 设置计数器过期时间
	_, err = g.Redis().Expire(debounceCtx, countKey, 2)
	if err != nil {
		g.Log().Error(debounceCtx, "设置计数器过期时间失败", err, "count_key", countKey)
	}

	keyMutex.Unlock()

	// 使用 Timer 替代 Sleep，支持提前取消
	timer := time.NewTimer(900 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		// 定时器到期，继续执行推送逻辑
		g.Log().Debug(debounceCtx, "防抖等待完成", "message_key", in.MessageKey)
	case <-cancelChan:
		// 被新的请求取消
		g.Log().Info(debounceCtx, "防抖等待被取消", "message_key", in.MessageKey, "reason", "有新的推送请求")
		return
	}

	// 检查是否应该推送
	shouldPush := true

	// 重新获取锁并检查状态
	keyMutex.Lock()
	defer keyMutex.Unlock()

	// 重新读取最新的计数值（避免使用过时数据）
	currentCount, _ := g.Redis().Get(debounceCtx, countKey)
	currentCountInt := currentCount.Int64()

	// 如果计数小于等于10次，检查UUID是否被更新
	if currentCountInt <= 10 {
		currentUUID, _ := g.Redis().Get(debounceCtx, in.MessageKey)
		if currentUUID.String() != pushUUID {
			// UUID已被更新，说明有新的推送请求，取消本次推送
			shouldPush = false
			g.Log().Info(debounceCtx, "防抖取消推送", "message_key", in.MessageKey, "reason", "UUID已变更", "count", currentCountInt)
		}
	}

	// 清理 cancel channel
	cancelChannels.Delete(in.MessageKey)

	// 如果不应该推送，直接返回
	if !shouldPush {
		return
	}

	// 清理计数器
	_, err = g.Redis().Del(debounceCtx, countKey)
	if err != nil {
		g.Log().Warning(debounceCtx, "删除计数器失败", err, "count_key", countKey)
	}

	// 执行实际推送前释放锁（避免长时间持有锁）
	keyMutex.Unlock()

	// 执行实际推送（使用独立的context）
	g.Log().Info(debounceCtx, "防抖推送消息", "message_key", in.MessageKey, "final_count", currentCountInt)
	_, err = s.directPushMessage(debounceCtx, in)
	if err != nil {
		g.Log().Error(ctx, "防抖推送失败", err, "message_key", in.MessageKey)
	}

	// 推送后重新获取锁清理 UUID
	keyMutex.Lock()
	_, _ = g.Redis().Del(debounceCtx, in.MessageKey)
}

// GetConnectionStats 获取连接统计信息
// 返回当前所有 WebSocket 连接的统计数据
// 参数：
//   - ctx: 上下文对象
//   - in: 获取连接统计输入参数
//
// 返回：
//   - out: 获取连接统计输出参数
//   - err: 错误信息
func (s *sWebSocket) GetConnectionStats(ctx context.Context, in *dto.GetConnectionStatsInput) (out *dto.GetConnectionStatsOutput, err error) {
	out = &dto.GetConnectionStatsOutput{
		Success:          true,
		Message:          "获取统计信息成功",
		TotalConnections: 0,
		ByCompany:        make(map[uint64]int32),
		BySource:         make(map[string]int32),
		ByDevice:         make(map[string]int32),
	}

	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			// 如果指定了公司UUID，则只统计该公司的连接
			if in.CompanyUuid != 0 && conn.CompanyUuid != in.CompanyUuid {
				return true
			}

			out.TotalConnections++
			out.ByCompany[conn.CompanyUuid]++
			out.BySource[conn.SourceClient]++
			out.ByDevice[conn.DeviceId]++
		}
		return true
	})

	return out, nil
}

// CheckDeviceOnline 检查设备是否在线
// 根据公司UUID和设备ID检查设备是否在线
// 参数：
//   - ctx: 上下文对象
//   - in: 检查设备在线输入参数
//
// 返回：
//   - out: 检查设备在线输出参数
//   - err: 错误信息
func (s *sWebSocket) CheckDeviceOnline(ctx context.Context, in *dto.CheckDeviceOnlineInput) (out *dto.CheckDeviceOnlineOutput, err error) {
	out = &dto.CheckDeviceOnlineOutput{
		Success:  true,
		Message:  "查询成功",
		IsOnline: false,
	}

	// 参数验证
	if in.CompanyUuid == 0 {
		return out, gerror.New("公司UUID不能为空")
	}
	if in.DeviceId == "" {
		return out, gerror.New("设备ID不能为空")
	}

	// 遍历查找匹配的连接
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == in.CompanyUuid &&
				conn.DeviceId == in.DeviceId &&
				(in.SourceClient == "" || conn.SourceClient == in.SourceClient) {
				out.IsOnline = true
				out.LastHeartbeat = conn.LastHeartbeat
				return false // 找到后停止遍历
			}
		}
		return true
	})

	return out, nil
}

// CloseConnection 关闭指定连接
// 根据条件关闭 WebSocket 连接
// 参数：
//   - ctx: 上下文对象
//   - in: 关闭连接输入参数
//
// 返回：
//   - out: 关闭连接输出参数
//   - err: 错误信息
func (s *sWebSocket) CloseConnection(ctx context.Context, in *dto.CloseConnectionInput) (out *dto.CloseConnectionOutput, err error) {
	out = &dto.CloseConnectionOutput{
		Success:     true,
		Message:     "关闭成功",
		ClosedCount: 0,
	}

	// 参数验证
	if in.CompanyUuid == 0 {
		return out, gerror.New("公司UUID不能为空")
	}

	// 收集需要关闭的连接
	var keysToClose []interface{}
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == in.CompanyUuid &&
				(in.SourceClient == "" || conn.SourceClient == in.SourceClient) &&
				(in.DeviceId == "" || conn.DeviceId == in.DeviceId) {
				keysToClose = append(keysToClose, key)
			}
		}
		return true
	})

	// 关闭连接
	for _, key := range keysToClose {
		if value, ok := WsClients.Load(key); ok {
			if conn, ok := value.(ConnectionInfo); ok {
				conn.ws.Close()
				WsClients.Delete(key)
				// 更新设备离线状态
				s.updateDeviceOffline(ctx, key.(string))
				out.ClosedCount++
			}
		}
	}

	g.Log().Info(ctx, "关闭连接完成", "closed_count", out.ClosedCount)

	return out, nil
}

// CloseAllConnections 关闭所有 WebSocket 连接
// 用于服务优雅关闭时清理所有连接
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - closedCount: 关闭的连接数量
func CloseAllConnections(ctx context.Context) int {
	g.Log().Info(ctx, "开始关闭所有 WebSocket 连接")

	closedCount := 0
	var keysToClose []interface{}

	// 1. 收集所有连接
	WsClients.Range(func(key, value interface{}) bool {
		keysToClose = append(keysToClose, key)
		return true
	})

	// 2. 发送关闭消息并关闭连接
	for _, key := range keysToClose {
		if value, ok := WsClients.Load(key); ok {
			if conn, ok := value.(ConnectionInfo); ok {
				// 尝试发送关闭消息给客户端
				closeMessage := ClientMessage{
					Type: "server_shutdown",
					Data: map[string]interface{}{
						"message": "服务器正在关闭，连接即将断开",
						"time":    time.Now().Format("2006-01-02 15:04:05"),
					},
				}

				// 使用写锁保护
				conn.writeMutex.Lock()
				// 设置写入超时
				conn.ws.SetWriteDeadline(time.Now().Add(2 * time.Second))
				// 尝试发送关闭通知
				if err := conn.ws.WriteJSON(closeMessage); err != nil {
					g.Log().Warning(ctx, "发送关闭通知失败", err, "device_id", conn.DeviceId)
				}
				conn.writeMutex.Unlock()

				// 发送 WebSocket 关闭帧
				conn.ws.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "服务器关闭"),
					time.Now().Add(time.Second),
				)

				// 关闭连接
				conn.ws.Close()

				// 从连接池中删除
				WsClients.Delete(key)

				// 更新设备离线状态
				New().updateDeviceOffline(ctx, key.(string))

				closedCount++

				g.Log().Debug(ctx, "已关闭连接",
					"company_uuid", conn.CompanyUuid,
					"source_client", conn.SourceClient,
					"device_id", conn.DeviceId,
				)
			}
		}
	}

	g.Log().Info(ctx, "所有 WebSocket 连接已关闭", "total_closed", closedCount)
	return closedCount
}

// HandleConnections GoFrame HTTP处理器（适配器）
// 这是 HTTP 升级到 WebSocket 的处理函数
func (s *sWebSocket) HandleConnections(r *ghttp.Request) {
	// 获取标准库的ResponseWriter和Request
	w := r.Response.Writer
	httpReq := r.Request

	// 调用实际的WebSocket处理逻辑
	s.handleWebSocketUpgrade(w, httpReq)
}

// handleWebSocketUpgrade 处理 WebSocket 升级和连接（内部实现）
func (s *sWebSocket) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(w, "Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	ctx := context.Background()

	// 处理连接
	newConn := s.handleConnectionSuccess(ctx, ws, r)
	if newConn == nil {
		return
	}

	// 设置一个定时器来发送ping消息 - 调整为25秒，确保在读取超时前有活动
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	// 创建连接专用的context
	connCtx, connCancel := context.WithCancel(globalCtx)
	defer connCancel()

	// 发送ping消息 - 使用context控制协程生命周期
	go func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case <-ticker.C:
				if err := safeWriteMessage(newConn, websocket.PingMessage, nil, 10*time.Second); err != nil {
					g.Log().Warning(ctx, "发送ping消息失败", err, "device_id", newConn.DeviceId)
					connCancel()
					return
				}
			}
		}
	}()

	// 添加资源清理
	defer func() {
		if ws != nil {
			ws.Close()
		}

		// 从连接池中移除并更新离线状态
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.ws == ws {
					WsClients.Delete(key)
					// 更新设备离线状态
					s.updateDeviceOffline(context.Background(), key.(string))
					return false
				}
			}
			return true
		})

		connCancel()
		g.Log().Info(ctx, "连接已清理", "device_id", newConn.DeviceId)
	}()

	// 主循环
	for {
		select {
		case <-connCtx.Done():
			g.Log().Info(ctx, "连接被取消，退出读取循环", "device_id", newConn.DeviceId)
			return
		default:
			// 继续读取消息
		}

		if ws == nil {
			g.Log().Warning(ctx, "WebSocket连接已断开", "device_id", newConn.DeviceId)
			break
		}

		// 检查连接状态 - 通过尝试设置读取截止时间来验证连接是否有效
		// 设置为2分钟，给客户端更充足的时间（ping间隔25秒 + 缓冲时间1分钟30秒）
		if err := ws.SetReadDeadline(time.Now().Add(2 * time.Minute)); err != nil {
			g.Log().Warning(ctx, "WebSocket连接已失效", "device_id", newConn.DeviceId, "error", err)
			break
		}

		// 读取消息
		_, msg, err := ws.ReadMessage()
		if err != nil {
			// 检查是否是网络临时错误
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				g.Log().Warning(ctx, "网络临时错误，重试中", "device_id", newConn.DeviceId, "error", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 检查连接状态，如果连接已关闭则退出
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				g.Log().Info(ctx, "WebSocket连接已关闭", "device_id", newConn.DeviceId, "close_code", err)
				break
			}

			// 检查是否是意外关闭
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				g.Log().Warning(ctx, "WebSocket连接异常关闭", err, "device_id", newConn.DeviceId)
			} else {
				g.Log().Warning(ctx, "读取消息错误", err, "device_id", newConn.DeviceId)
			}
			break
		}

		// 检查消息大小
		if len(msg) > 1024*1024 { // 1MB限制
			g.Log().Warning(ctx, "消息过大，忽略", "device_id", newConn.DeviceId, "size", len(msg))
			continue
		}

		// 使用recover保护消息处理
		func() {
			defer func() {
				if r := recover(); r != nil {
					g.Log().Error(ctx, "消息处理异常", r, "device_id", newConn.DeviceId)
				}
			}()
			s.handleMessage(ctx, ws, msg, newConn)
		}()
	}
}

// handleConnectionSuccess 处理连接成功
func (s *sWebSocket) handleConnectionSuccess(ctx context.Context, ws *websocket.Conn, r *http.Request) *ConnectionInfo {
	// 验证参数
	client := r.URL.Query().Get("client")
	token := r.URL.Query().Get("token")
	if client == "" || token == "" {
		ws.Close()
		return nil
	}

	// 验证token
	// 使用 GoFrame 框架的配置管理获取 JWT 密钥
	jwtSecret := g.Cfg().MustGet(ctx, "jwt.secret").String()
	if jwtSecret == "" {
		g.Log().Error(ctx, "JWT密钥未配置")
		ws.Close()
		return nil
	}

	claims, err := utility.ParseToken(token, jwtSecret)
	if err != nil {
		// 创建临时连接信息用于安全写入
		tempConn := &ConnectionInfo{ws: ws, writeMutex: &sync.Mutex{}}
		safeWriteMessage(tempConn, websocket.TextMessage, getMsgData(PushMessage{
			Event: "connect",
			State: consts.CodeFail,
			Msg:   "Token Error",
		}), 5*time.Second)
		ws.Close()
		return nil
	}

	companyUuid := claims.CompanyUuid
	staffUuid := claims.StaffUuid
	deviceId := claims.DeviceId

	// 允许一个设备最多存在三个连接
	connCount := 0
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == companyUuid &&
				conn.SourceClient == client &&
				conn.DeviceId == deviceId &&
				conn.ws != ws {
				connCount++
			}
		}
		return true
	})

	// 如果连接数量大于等于3，则清空之前的连接
	if connCount >= 3 {
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.CompanyUuid == companyUuid &&
					conn.SourceClient == client &&
					conn.DeviceId == deviceId &&
					conn.ws != ws {
					conn.ws.Close()
					WsClients.Delete(key)
				}
			}
			return true
		})
	}

	// 添加新连接
	newConn := &ConnectionInfo{
		CompanyUuid:   companyUuid,
		StaffUuid:     staffUuid,
		SourceClient:  client,
		DeviceId:      deviceId,
		LastHeartbeat: time.Now().Format(time.RFC3339),
		ws:            ws,
		writeMutex:    &sync.Mutex{},
	}

	// 生成唯一的连接key
	connKey := fmt.Sprintf("%d_%s_%s_%d", companyUuid, client, deviceId, time.Now().UnixNano())
	WsClients.Store(connKey, *newConn)

	// 获取客户端IP地址
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// 创建或更新设备在线记录（同一设备只保留一条记录）
	nowTime := int(time.Now().Unix())

	// 先查询是否已存在该设备的记录
	var existingRecord *entity.DeviceOnline
	err = dao.DeviceOnline.Ctx(ctx).
		Where(dao.DeviceOnline.Columns().CompanyUuid, companyUuid).
		Where(dao.DeviceOnline.Columns().DeviceId, deviceId).
		Where(dao.DeviceOnline.Columns().SourceClient, client).
		Scan(&existingRecord)

	if err == nil && existingRecord != nil {
		// 记录已存在，更新为在线状态
		_, err = dao.DeviceOnline.Ctx(ctx).
			Where(dao.DeviceOnline.Columns().Id, existingRecord.Id).
			Update(do.DeviceOnline{
				StaffUuid:         int64(staffUuid),
				ConnectionKey:     connKey,
				Status:            1, // 在线
				ConnectTime:       nowTime,
				LastHeartbeatTime: nowTime,
				IpAddress:         ipAddress,
				UserAgent:         userAgent,
				UpdateTime:        nowTime,
			})
		if err != nil {
			g.Log().Warning(ctx, "更新设备在线记录失败", err, "device_id", deviceId)
		}
	} else {
		// 记录不存在，创建新记录
		deviceOnlineRecord := &do.DeviceOnline{
			CompanyUuid:       int64(companyUuid),
			StaffUuid:         int64(staffUuid),
			DeviceId:          deviceId,
			SourceClient:      client,
			ConnectionKey:     connKey,
			Status:            1, // 在线
			ConnectTime:       nowTime,
			LastHeartbeatTime: nowTime,
			IpAddress:         ipAddress,
			UserAgent:         userAgent,
			CreateTime:        nowTime,
			UpdateTime:        nowTime,
		}

		_, err = dao.DeviceOnline.Ctx(ctx).Data(deviceOnlineRecord).Insert()
		if err != nil {
			g.Log().Warning(ctx, "创建设备在线记录失败", err, "device_id", deviceId)
		}
	}

	// 监听关闭事件
	ws.SetCloseHandler(func(code int, text string) error {
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.CompanyUuid == companyUuid &&
					conn.SourceClient == client &&
					conn.DeviceId == deviceId &&
					conn.ws == ws {
					WsClients.Delete(key)
					// 更新设备在线记录为离线状态
					s.updateDeviceOffline(context.Background(), connKey)
					return false
				}
			}
			return true
		})
		return nil
	})

	// 发送连接成功消息
	if err := safeWriteMessage(newConn, websocket.TextMessage, getMsgData(PushMessage{
		Event: "connect",
		State: consts.CodeSuccess,
		Msg:   "Connected successfully",
		Data: map[string]interface{}{
			"source":       client,
			"company_uuid": companyUuid,
			"staff_uuid":   staffUuid,
			"device_id":    deviceId,
		},
	}), 10*time.Second); err != nil {
		g.Log().Warning(ctx, "发送连接成功消息失败", err, "device_id", deviceId)
		return nil
	}

	return newConn
}

// handleMessage 处理客户端消息
func (s *sWebSocket) handleMessage(ctx context.Context, ws *websocket.Conn, msg []byte, newConn *ConnectionInfo) {
	// 解析消息
	clientMessage := ClientMessage{}
	err := json.Unmarshal(fixJSONFormat(msg), &clientMessage)
	if err != nil {
		g.Log().Warning(ctx, "JSON解析错误", err, "device_id", newConn.DeviceId)
		return
	}

	// 将字符串转换为uint
	msgId := uint(0)
	switch v := clientMessage.MsgId.(type) {
	case float64:
		msgId = uint(v)
	case string:
		if msgIdInt, err := strconv.ParseUint(v, 10, 64); err == nil {
			msgId = uint(msgIdInt)
		}
	}

	// 处理心跳消息
	if clientMessage.Type == consts.ClientMessageTypeHeartbeat {
		s.handleHeartbeat(ctx, newConn)
		return
	}

	// 处理已读删除
	if clientMessage.Type == consts.ClientMessageTypeReply {
		s.handleReply(ctx, msgId)
		return
	}

	// 上报LAN打印机数据
	if clientMessage.Type == consts.ClientMessageTypeLanPrintReport && newConn.SourceClient == consts.SourceCashier {
		s.reportLanPrinter(ctx, newConn, clientMessage)
	}
}

// handleHeartbeat 处理心跳消息
func (s *sWebSocket) handleHeartbeat(ctx context.Context, newConn *ConnectionInfo) {
	// 更新心跳时间
	var connKey string
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == newConn.CompanyUuid &&
				conn.SourceClient == newConn.SourceClient &&
				conn.DeviceId == newConn.DeviceId &&
				conn.ws == newConn.ws {
				updatedConn := conn
				updatedConn.LastHeartbeat = time.Now().Format(time.RFC3339)
				WsClients.Store(key, updatedConn)
				connKey = key.(string)
				return false
			}
		}
		return true
	})

	// 更新数据库中的心跳时间
	if connKey != "" {
		s.updateDeviceHeartbeat(ctx, connKey)
	}

	g.Log().Debug(ctx, "收到心跳消息", "company_uuid", newConn.CompanyUuid, "device_id", newConn.DeviceId, "source_client", newConn.SourceClient)
}

// handleReply 处理已读回复
func (s *sWebSocket) handleReply(ctx context.Context, msgId uint) {
	if msgId == 0 {
		return
	}

	// 删除消息记录
	_, err := dao.WebsocketMsg.Ctx(ctx).Where(dao.WebsocketMsg.Columns().Id, msgId).Delete()
	if err != nil {
		g.Log().Warning(ctx, "删除消息记录失败", err, "msg_id", msgId)
	}
}

// reportLanPrinter 上报LAN打印机数据
func (s *sWebSocket) reportLanPrinter(ctx context.Context, newConn *ConnectionInfo, clientMessage ClientMessage) {
	// 解析打印机数据
	dataJson, err := json.Marshal(clientMessage.Data)
	if err != nil {
		g.Log().Warning(ctx, "解析打印机数据失败", err)
		return
	}

	lanPrinters := []LanPrinter{}
	err = json.Unmarshal(dataJson, &lanPrinters)
	if err != nil {
		g.Log().Warning(ctx, "解析LAN打印机数据失败", err)
		return
	}

	g.Log().Info(ctx, "收到LAN打印机上报",
		"device_id", newConn.DeviceId,
		"count", len(lanPrinters),
	)

	// 发布 LAN 打印机上报消息到 MQ
	s.publishLanPrinterReport(ctx, newConn, lanPrinters)
}

// publishLanPrinterReport 发布 LAN 打印机上报消息到 MQ
// 将收银机上报的局域网打印机信息发布到消息队列，供其他服务消费
// 参数：
//   - ctx: 上下文对象
//   - conn: WebSocket 连接信息
//   - printers: LAN 打印机列表
func (s *sWebSocket) publishLanPrinterReport(ctx context.Context, conn *ConnectionInfo, printers []LanPrinter) {
	// 转换打印机列表为 ttpos-api 定义的结构体
	apiPrinters := make([]ttposWebsocketMsg.LanPrinterInfo, 0, len(printers))
	for _, p := range printers {
		apiPrinters = append(apiPrinters, ttposWebsocketMsg.LanPrinterInfo{
			IP:     p.Ip,
			Port:   p.Port,
			Status: p.Status,
			Remark: p.Remark,
		})
	}

	// 创建消息体（使用 ttpos-api 定义的结构体）
	message := ttposWebsocketMsg.NewLanPrinterReportMessage(
		conn.CompanyUuid,
		conn.StaffUuid,
		conn.DeviceId,
		conn.SourceClient,
		apiPrinters,
	)

	// 验证消息
	if err := message.Validate(); err != nil {
		g.Log().Error(ctx, "LAN 打印机上报消息验证失败", err,
			"company_uuid", conn.CompanyUuid,
			"device_id", conn.DeviceId,
			"printer_count", len(printers),
		)
		return
	}

	// 发布到消息队列
	// Topic: constant.TopicLanPrinterReport
	if err := queue.PushWithContext(ctx, constant.TopicLanPrinterReport, message); err != nil {
		g.Log().Error(ctx, "发布 LAN 打印机上报消息失败", err,
			"company_uuid", conn.CompanyUuid,
			"device_id", conn.DeviceId,
			"printer_count", len(printers),
		)
		return
	}

	g.Log().Info(ctx, "LAN 打印机上报消息已发布到 MQ",
		"company_uuid", conn.CompanyUuid,
		"device_id", conn.DeviceId,
		"printer_count", len(printers),
		"topic", constant.TopicLanPrinterReport,
	)
}

// checkHeartbeatTimeout 检查心跳超时的连接并断开
func checkHeartbeatTimeout() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		select {
		case <-globalCtx.Done():
			return
		case <-ticker.C:
			var expiredKeys []interface{}
			WsClients.Range(func(key, value interface{}) bool {
				if conn, ok := value.(ConnectionInfo); ok {
					lastHeartbeat, err := time.Parse(time.RFC3339, conn.LastHeartbeat)
					if err != nil {
						lastHeartbeat = time.Now()
					}
					// 检查是否超过2分钟没有心跳
					if time.Since(lastHeartbeat) > 2*time.Minute {
						g.Log().Warning(ctx, "断开超时连接",
							"device_id", conn.DeviceId,
							"last_heartbeat", conn.LastHeartbeat,
						)
						conn.ws.Close()
						expiredKeys = append(expiredKeys, key)
					}
				}
				return true
			})

			// 删除过期的连接
			for _, key := range expiredKeys {
				WsClients.Delete(key)
			}
		}
	}
}

// deleteOldMessages 删除旧消息
func (s *sWebSocket) deleteOldMessages(ctx context.Context, in *dto.PushMessageInput) error {
	_, err := dao.WebsocketMsg.Ctx(ctx).
		Where(dao.WebsocketMsg.Columns().Type, in.MessageType).
		Where(dao.WebsocketMsg.Columns().CompanyUuid, in.CompanyUuid).
		Delete()
	return err
}

// collectMatchedConnections 收集匹配的连接
func (s *sWebSocket) collectMatchedConnections(in *dto.PushMessageInput) []ConnectionInfo {
	var matchedConnections []ConnectionInfo
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == in.CompanyUuid &&
				(conn.StaffUuid == in.StaffUuid || in.StaffUuid == 0) &&
				(conn.StaffUuid != in.NotStaffUuid || in.NotStaffUuid == 0) &&
				(conn.SourceClient == in.SourceClient || in.SourceClient == consts.SourceAll) &&
				(conn.DeviceId == in.DeviceId || in.DeviceId == "*") &&
				(conn.DeviceId != in.NotDeviceId || in.NotDeviceId == "*" || in.NotDeviceId == "") {
				matchedConnections = append(matchedConnections, conn)
			}
		}
		return true
	})
	return matchedConnections
}

// pushToConnection 推送消息到指定连接
func (s *sWebSocket) pushToConnection(ctx context.Context, conn ConnectionInfo, in *dto.PushMessageInput) error {
	// 创建消息记录
	msgRecord := &do.WebsocketMsg{
		CompanyUuid:  in.CompanyUuid,
		Uid:          conn.DeviceId,
		Msg:          in.Data,
		Type:         in.MessageType,
		SourceClient: conn.SourceClient,
		Status:       0,
		IsOffline:    0,
		CreateTime:   uint(time.Now().Unix()),
		UpdateTime:   uint(time.Now().Unix()),
	}

	// 插入数据库
	result, err := dao.WebsocketMsg.Ctx(ctx).Data(msgRecord).Insert()
	if err != nil {
		return gerror.Wrap(err, "创建消息记录失败")
	}

	msgId, _ := result.LastInsertId()

	// 解析消息数据
	var dataMap interface{}
	if err := json.Unmarshal([]byte(in.Data), &dataMap); err != nil {
		dataMap = in.Data
	}

	// 发送消息
	message := PushMessage{
		Event: in.MessageType,
		State: consts.CodeSuccess,
		Data:  dataMap,
		MsgId: int(msgId),
	}

	err = safeWriteMessage(&conn, websocket.TextMessage, getMsgData(message), 10*time.Second)
	if err != nil {
		g.Log().Warning(ctx, "推送消息失败", err, "device_id", conn.DeviceId)
		return err
	}

	g.Log().Info(ctx, "推送消息成功",
		"device_id", conn.DeviceId,
		"message_type", in.MessageType,
		"msg_id", msgId,
	)

	return nil
}

// safeWriteMessage 安全写入消息，防止并发冲突
func safeWriteMessage(conn *ConnectionInfo, messageType int, data []byte, timeout time.Duration) error {
	conn.writeMutex.Lock()
	defer conn.writeMutex.Unlock()

	conn.ws.SetWriteDeadline(time.Now().Add(timeout))
	return conn.ws.WriteMessage(messageType, data)
}

// getMsgData 获取消息数据
func getMsgData(msg PushMessage) []byte {
	jsonData, _ := json.Marshal(msg)
	return jsonData
}

// fixJSONFormat 修复常见的JSON格式问题
func fixJSONFormat(data []byte) []byte {
	str := string(data)
	str = strings.TrimSpace(str)
	if !strings.HasPrefix(str, "{") {
		return data
	}

	// 修复常见问题：键没有双引号
	str = strings.ReplaceAll(str, "{type:", `{"type":`)
	str = strings.ReplaceAll(str, ", type:", `, "type":`)
	str = strings.ReplaceAll(str, ",type:", `, "type":`)
	str = strings.ReplaceAll(str, "{msg_id:", `{"msg_id":`)
	str = strings.ReplaceAll(str, ", msg_id:", `, "msg_id":`)
	str = strings.ReplaceAll(str, ",msg_id:", `, "msg_id":`)
	str = strings.ReplaceAll(str, "{data:", `{"data":`)
	str = strings.ReplaceAll(str, ", data:", `, "data":`)
	str = strings.ReplaceAll(str, ",data:", `, "data":`)

	// 修复值没有双引号的问题
	str = strings.ReplaceAll(str, `: reply`, `: "reply"`)
	str = strings.ReplaceAll(str, `: heartbeat`, `: "heartbeat"`)
	str = strings.ReplaceAll(str, `: usb_print_report`, `: "usb_print_report"`)
	str = strings.ReplaceAll(str, `: lan_print_report`, `: "lan_print_report"`)

	return []byte(str)
}

// updateDeviceHeartbeat 更新设备心跳时间
// 在数据库中更新设备的最后心跳时间
func (s *sWebSocket) updateDeviceHeartbeat(ctx context.Context, connectionKey string) {
	nowTime := int(time.Now().Unix())
	_, err := dao.DeviceOnline.Ctx(ctx).
		Where(dao.DeviceOnline.Columns().ConnectionKey, connectionKey).
		Where(dao.DeviceOnline.Columns().Status, 1).
		Update(do.DeviceOnline{
			LastHeartbeatTime: nowTime,
			UpdateTime:        nowTime,
		})
	if err != nil {
		g.Log().Warning(ctx, "更新设备心跳时间失败", err, "connection_key", connectionKey)
	}
}

// updateDeviceOffline 更新设备为离线状态
// 在数据库中标记设备为离线状态
func (s *sWebSocket) updateDeviceOffline(ctx context.Context, connectionKey string) {
	nowTime := int(time.Now().Unix())
	_, err := dao.DeviceOnline.Ctx(ctx).
		Where(dao.DeviceOnline.Columns().ConnectionKey, connectionKey).
		Update(do.DeviceOnline{
			Status:         0, // 离线
			DisconnectTime: nowTime,
			UpdateTime:     nowTime,
		})
	if err != nil {
		g.Log().Warning(ctx, "更新设备离线状态失败", err, "connection_key", connectionKey)
	}
}

// updateDeviceOfflineByInfo 通过设备信息更新设备为离线状态
// 用于无法获取 connection_key 的场景
func (s *sWebSocket) updateDeviceOfflineByInfo(ctx context.Context, companyUuid uint64, deviceId, sourceClient string) {
	nowTime := int(time.Now().Unix())
	_, err := dao.DeviceOnline.Ctx(ctx).
		Where(dao.DeviceOnline.Columns().CompanyUuid, companyUuid).
		Where(dao.DeviceOnline.Columns().DeviceId, deviceId).
		Where(dao.DeviceOnline.Columns().SourceClient, sourceClient).
		Update(do.DeviceOnline{
			Status:         0, // 离线
			DisconnectTime: nowTime,
			UpdateTime:     nowTime,
		})
	if err != nil {
		g.Log().Warning(ctx, "更新设备离线状态失败", err,
			"company_uuid", companyUuid,
			"device_id", deviceId,
			"source_client", sourceClient,
		)
	}
}

// getClientIP 获取客户端真实IP地址
// 优先从代理头中获取，如果没有则从RemoteAddr获取
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头获取
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-IP 头获取
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 从 RemoteAddr 获取
	ip = r.RemoteAddr
	// RemoteAddr 格式为 "IP:Port"，需要去掉端口
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// getMessageKeyMutex 获取指定MessageKey的专用锁
func getMessageKeyMutex(messageKey string) *sync.Mutex {
	// 先尝试获取已存在的锁
	if mutex, exists := messageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 如果不存在，需要创建新锁
	mutexMapLock.Lock()
	defer mutexMapLock.Unlock()

	// 双重检查，防止并发创建
	if mutex, exists := messageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 创建新锁并存储
	newMutex := &sync.Mutex{}
	messageKeyMutexes.Store(messageKey, newMutex)

	return newMutex
}
