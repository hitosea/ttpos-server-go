package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"
	"websocket/config"
	"websocket/constant"
	"websocket/model"
	"websocket/pkg/database"
	"websocket/repository"
	"websocket/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsClients 存储所有的连接
var WsClients []ConnectionInfo

// 初始化函数，启动定时检查心跳超时的连接
func init() {
	go checkHeartbeatTimeout()
}

// Message represents a generic message structure
type PushMessage struct {
	Event string      `json:"event"`
	State int         `json:"state"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	MsgId int         `json:"msg_id,omitempty"`
}

// ClientMessage represents a generic message structure
type ClientMessage struct {
	Type  string `json:"type"`
	Event string `json:"event"`
	MsgId any    `json:"msg_id,omitempty"`
}

// ConnectionInfo represents a generic message structure
type ConnectionInfo struct {
	CompanyUuid   uint64 `json:"company_id"`
	SourceClient  string `json:"source_client"`
	DeviceId      string `json:"device_id"`
	StaffUuid     uint64 `json:"staff_uuid"`
	LastHeartbeat string `json:"last_heartbeat"`
	ws            *websocket.Conn
}

// HandleConnections 处理连接
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(w, "Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	// 处理连接
	newConn := handleConnectionSuccess(ws, r)
	if newConn == nil {
		return
	}

	// 设置一个自动收报机来发送ping消息
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 发送ping消息
	go func() {
		for range ticker.C {
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("Error sending ping message:", err)
				return
			}
		}
	}()

	// 主循环
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err, " DeviceId:", newConn.DeviceId)
			break
		}
		// 处理消息
		handleMessage(ws, msg, newConn)
	}
}

// getMsgData 获取消息数据
func getMsgData(msg PushMessage) []byte {
	jsonData, _ := json.Marshal(msg)
	return jsonData
}

// handleConnectionSuccess 处理连接成功
func handleConnectionSuccess(ws *websocket.Conn, r *http.Request) *ConnectionInfo {
	// 验证参数
	client := r.URL.Query().Get("client")
	token := r.URL.Query().Get("token")
	if client == "" || token == "" {
		ws.Close()
		return nil
	}

	// 验证token
	claims, err := utils.ParseToken(token, config.JWT.Secret)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, getMsgData(PushMessage{
			Event: "connect",
			State: constant.CodeFail,
			Msg:   "Token Error",
		}))
		ws.Close()
		return nil
	}

	// 验证设备是否绑定
	deviceRepo := repository.NewDeviceRepository(database.Instance)
	existsDevice := deviceRepo.GetRecordBySourceAndDeviceId(claims.CompanyUuid, claims.Source, claims.DeviceId)
	if existsDevice.ID == 0 {
		ws.WriteMessage(websocket.TextMessage, getMsgData(PushMessage{
			Event: "connect",
			State: constant.CodeFail,
			Msg:   "No binding",
		}))
		ws.Close()
		return nil
	}

	// 允许一个设备最多存在两个连接，超过则清空之前的连接
	// 计算当前设备的连接数量
	connCount := 0
	connIndexes := []int{}
	for i, conn := range WsClients {
		if conn.CompanyUuid == claims.CompanyUuid &&
			conn.SourceClient == existsDevice.Source &&
			conn.DeviceId == existsDevice.DeviceId &&
			conn.ws != ws {
			connCount++
			connIndexes = append(connIndexes, i)
		}
	}

	// 如果连接数量大于等于2，则清空之前的连接
	if connCount >= 2 {
		// 从后向前遍历，避免删除元素影响索引
		for i := len(connIndexes) - 1; i >= 0; i-- {
			index := connIndexes[i]
			// 确保索引有效
			if index >= 0 && index < len(WsClients) {
				WsClients[index].ws.Close()
				WsClients = slices.Delete(WsClients, index, index+1)
			}
		}
	}

	// 添加新连接
	newConn := ConnectionInfo{
		CompanyUuid:   claims.CompanyUuid,
		StaffUuid:     claims.StaffUuid,
		SourceClient:  existsDevice.Source,
		DeviceId:      existsDevice.DeviceId,
		LastHeartbeat: time.Now().Format(time.RFC3339),
		ws:            ws,
	}
	WsClients = append(WsClients, newConn)

	// 监听关闭事件以删除连接详细信息
	ws.SetCloseHandler(func(code int, text string) error {
		for i, conn := range WsClients {
			if conn.CompanyUuid == claims.CompanyUuid &&
				conn.SourceClient == existsDevice.Source &&
				conn.DeviceId == existsDevice.DeviceId &&
				conn.ws == ws {
				WsClients = slices.Delete(WsClients, i, i+1)
				break
			}
		}
		return nil
	})

	// 发送连接成功消息
	if err := ws.WriteMessage(websocket.TextMessage, getMsgData(PushMessage{
		Event: "connect",
		State: constant.CodeSuccess,
		Msg:   "Connected successfully",
		Data: map[string]interface{}{
			"source":       claims.Source,
			"company_uuid": claims.CompanyUuid,
			"staff_uuid":   claims.StaffUuid,
			"device_id":    claims.DeviceId,
		},
	})); err != nil {
		fmt.Println("Error sending ping message:", err)
		return nil
	}
	return &newConn
}

// handleMessage 处理消息
func handleMessage(ws *websocket.Conn, msg []byte, newConn *ConnectionInfo) {
	// 解析消息
	msgId := uint(0)
	clientMessage := ClientMessage{}
	err := json.Unmarshal(msg, &clientMessage)
	if err != nil {
		fmt.Println("WebSocket Error parsing message:", err)
		return
	}

	// 将字符串转换为uint
	switch v := clientMessage.MsgId.(type) {
	case float64:
		msgId = uint(v)
	case string:
		if msgIdInt, err := strconv.ParseUint(v, 10, 64); err == nil {
			msgId = uint(msgIdInt)
		}
	}

	// 处理心跳消息
	if clientMessage.Type == "heartbeat" {
		// 更新心跳时间
		for i, conn := range WsClients {
			if conn.CompanyUuid == newConn.CompanyUuid &&
				conn.SourceClient == newConn.SourceClient &&
				conn.DeviceId == newConn.DeviceId &&
				conn.ws == newConn.ws {
				WsClients[i].LastHeartbeat = time.Now().Format(time.RFC3339)
				break
			}
		}

		isOnline := false
		for _, conn := range WsClients {
			if conn.DeviceId == newConn.DeviceId {
				isOnline = true
			}
		}
		if !isOnline {
			fmt.Println("Heartbeat message DeviceId - 离线: ", newConn.DeviceId)
		} else {
			fmt.Println("Heartbeat message DeviceId - 在线: ", newConn.DeviceId)
		}
		// todo 暂时不回复心跳消息
		// 发送回复消息
		// message := PushMessage{
		// 	Event: "reply_heartbeat",
		// 	State: constant.CodeSuccess,
		// 	Msg:   "Reply successfully",
		// }
		// err := ws.WriteMessage(websocket.TextMessage, getMsgData(message))
		// if err != nil {
		// 	fmt.Println("Error writing message:", err)
		// }
		return
	}

	// 处理已读删除
	if clientMessage.Type == "reply" {
		repo := repository.NewWebSocketMsgRepository(database.Instance)
		err := repo.DeleteByTypeAndId(msgId)
		if err != nil {
			fmt.Printf("Error updating message status: %v\n", err)
		}
		// todo 暂时不回复心跳消息
		// message := PushMessage{
		// 	Event: "reply",
		// 	State: constant.CodeSuccess,
		// 	Msg:   "Reply successfully",
		// }
		// err = ws.WriteMessage(websocket.TextMessage, getMsgData(message))
		// if err != nil {
		// 	fmt.Printf("Error sending message to client: %v\n", err)
		// }
		return
	}
}

// checkHeartbeatTimeout 检查心跳超时的连接并断开
func checkHeartbeatTimeout() {
	// 每30秒检查一次
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 遍历所有连接
		for i := len(WsClients) - 1; i >= 0; i-- {
			// 确保索引有效
			if i < len(WsClients) {
				conn := WsClients[i]
				// 解析最后心跳时间
				lastHeartbeat, err := time.Parse(time.RFC3339, conn.LastHeartbeat)
				if err != nil {
					// 如果解析出错，可能是初始值或无效值，设置为当前时间
					lastHeartbeat = time.Now()
				}
				// 检查是否超过2分钟没有心跳
				if time.Since(lastHeartbeat) > 2*time.Minute {
					fmt.Printf("断开超时2分钟没有心跳的连接: %s, 最后心跳时间: %s\n", conn.DeviceId, conn.LastHeartbeat)
					// 关闭连接
					conn.ws.Close()
					// 从列表中移除
					WsClients = slices.Delete(WsClients, i, i+1)
				}
			}
		}
	}
}

// PushClient 推送消息
func PushClient(messageData MessageData) {
	for _, conn := range WsClients {
		if conn.CompanyUuid == messageData.CompanyUuid &&
			(conn.StaffUuid == messageData.StaffUuid || messageData.StaffUuid == 0) &&
			(conn.StaffUuid != messageData.NotStaffUuid || messageData.NotStaffUuid == 0) &&
			(conn.SourceClient == messageData.SourceClient || messageData.SourceClient == "*") &&
			(conn.DeviceId == messageData.DeviceId || messageData.DeviceId == "*") &&
			(conn.DeviceId != messageData.NotDeviceId || messageData.NotDeviceId == "*" || messageData.NotDeviceId == "") {
			// 创建一个 WebSocketMsgRepository 实例
			repo := repository.NewWebSocketMsgRepository(database.Instance)

			// 1. 先删后加 - 同一个类型，只保留最新的
			err := repo.DeleteByTypeAndCompanyId(messageData.MessageType, messageData.CompanyUuid)
			if err != nil {
				fmt.Printf("Error deleting old messages: %v\n", err)
			}

			// 2. 创建
			id, err := repo.Create(model.WebSocketMsg{
				CompanyUuid:  messageData.CompanyUuid,
				Uid:          conn.DeviceId,
				Msg:          utils.StructToJson(messageData.Data),
				Type:         messageData.MessageType,
				SourceClient: conn.SourceClient,
				Status:       0,
				IsOffline:    0,
				CreateTime:   uint(time.Now().Unix()),
				UpdateTime:   uint(time.Now().Unix()),
			})
			if err != nil {
				fmt.Printf("Error creating WebSocket message: %v\n", err)
				continue
			}

			// 发送消息
			message := PushMessage{
				Event: messageData.MessageType,
				State: constant.CodeSuccess,
				Data:  messageData.Data,
				MsgId: int(id),
			}

			err = conn.ws.WriteMessage(websocket.TextMessage, getMsgData(message))
			if err != nil {
				fmt.Printf("Error sending message to client: %v\n", err)
			} else {
				fmt.Println("推送消息:", utils.StructToJson(map[string]interface{}{
					"SourceClient": conn.SourceClient,
					"CompanyUuid":  messageData.CompanyUuid,
					"DeviceId":     conn.DeviceId,
					"Event":        messageData.MessageType,
					"State":        constant.CodeSuccess,
					"Data":         messageData.Data,
					"MsgId":        int(id),
				}))
			}
		}
	}
}
