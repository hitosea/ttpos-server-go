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
	"ttpos-bmp/app/ttpos-websocket/internal/service"
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
	)

	// 1. 先删后加 - 同一个类型，只保留最新的
	if err := s.deleteOldMessages(ctx, in); err != nil {
		g.Log().Warning(ctx, "删除旧消息失败", err)
	}

	// 2. 收集匹配的连接
	matchedConnections := s.collectMatchedConnections(in)

	g.Log().Info(ctx, "找到匹配连接",
		"count", len(matchedConnections),
		"message_type", in.MessageType,
	)

	// 3. 批量处理匹配的连接
	pushCount := int32(0)
	for _, conn := range matchedConnections {
		if err := s.pushToConnection(ctx, conn, in); err != nil {
			g.Log().Warning(ctx, "推送到连接失败", err, "device_id", conn.DeviceId)
		} else {
			pushCount++
		}
	}

	out.Success = true
	out.Message = "推送成功"
	out.PushCount = pushCount

	return out, nil
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
				out.ClosedCount++
			}
		}
	}

	g.Log().Info(ctx, "关闭连接完成", "closed_count", out.ClosedCount)

	return out, nil
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

		// 从连接池中移除
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.ws == ws {
					WsClients.Delete(key)
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

	// TODO: 验证token和设备绑定
	// 这里需要从数据库查询设备信息和验证token
	// 暂时使用简化逻辑

	companyUuid := uint64(1) // TODO: 从token解析
	staffUuid := uint64(1)   // TODO: 从token解析
	deviceId := "device_1"   // TODO: 从token解析

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

	// 监听关闭事件
	ws.SetCloseHandler(func(code int, text string) error {
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.CompanyUuid == companyUuid &&
					conn.SourceClient == client &&
					conn.DeviceId == deviceId &&
					conn.ws == ws {
					WsClients.Delete(key)
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
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == newConn.CompanyUuid &&
				conn.SourceClient == newConn.SourceClient &&
				conn.DeviceId == newConn.DeviceId &&
				conn.ws == newConn.ws {
				updatedConn := conn
				updatedConn.LastHeartbeat = time.Now().Format(time.RFC3339)
				WsClients.Store(key, updatedConn)
				return false
			}
		}
		return true
	})

	g.Log().Debug(ctx, "收到心跳消息", "device_id", newConn.DeviceId)
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

	// TODO: 实现LAN打印机数据处理逻辑
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
