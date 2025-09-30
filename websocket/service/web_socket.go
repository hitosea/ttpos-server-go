package service

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
	"websocket/config"
	"websocket/constant"
	"websocket/model"
	"websocket/pkg/database"
	"websocket/pkg/logger"
	"websocket/repository"
	"websocket/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsClients 存储所有的连接 - 使用sync.Map提高并发安全性
var WsClients sync.Map

// 全局context用于控制协程生命周期
var globalCtx context.Context
var globalCancel context.CancelFunc

// 初始化函数，启动定时检查心跳超时的连接
func init() {
	globalCtx, globalCancel = context.WithCancel(context.Background())
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
	Type  string      `json:"type"`
	MsgId any         `json:"msg_id,omitempty"`
	Data  interface{} `json:"data"`
}

// ConnectionInfo represents a generic message structure
type ConnectionInfo struct {
	CompanyUuid   uint64 `json:"company_id"`
	SourceClient  string `json:"source_client"`
	DeviceId      string `json:"device_id"`
	StaffUuid     uint64 `json:"staff_uuid"`
	LastHeartbeat string `json:"last_heartbeat"`
	ws            *websocket.Conn
	writeMutex    *sync.Mutex // 使用指针避免复制锁
}

type UsbPrinter struct {
	M_name string `json:"m_name"`
	Name   string `json:"name"`
	Pid    any    `json:"pid"`
	Sn     string `json:"sn"`
	Vid    any    `json:"vid"`
}

type LanPrinter struct {
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
	Status int    `json:"status"`
	Remark string `json:"remark"`
}

// getMsgData 获取消息数据
func getMsgData(msg PushMessage) []byte {
	jsonData, _ := json.Marshal(msg)
	return jsonData
}

// safeWriteMessage 安全写入消息，防止并发冲突
func safeWriteMessage(conn *ConnectionInfo, messageType int, data []byte, timeout time.Duration) error {
	conn.writeMutex.Lock()
	defer conn.writeMutex.Unlock()

	// 设置写入超时时间
	conn.ws.SetWriteDeadline(time.Now().Add(timeout))
	return conn.ws.WriteMessage(messageType, data)
}

// fixJSONFormat 修复常见的JSON格式问题
func fixJSONFormat(data []byte) []byte {
	str := string(data)
	// 去除前后空白字符
	str = strings.TrimSpace(str)
	// 如果不是以 { 开头，可能不是JSON，直接返回
	if !strings.HasPrefix(str, "{") {
		return data
	}
	// 修复常见问题：键没有双引号
	str = strings.ReplaceAll(str, "{type:", `{"type":`)
	str = strings.ReplaceAll(str, ", type:", `, "type":`)
	str = strings.ReplaceAll(str, ",type:", `, "type":`)
	// 修复msg_id没有双引号
	str = strings.ReplaceAll(str, "{msg_id:", `{"msg_id":`)
	str = strings.ReplaceAll(str, ", msg_id:", `, "msg_id":`)
	str = strings.ReplaceAll(str, ",msg_id:", `, "msg_id":`)
	// 修复data没有双引号
	str = strings.ReplaceAll(str, "{data:", `{"data":`)
	str = strings.ReplaceAll(str, ", data:", `, "data":`)
	str = strings.ReplaceAll(str, ",data:", `, "data":`)
	// 修复值没有双引号的问题（仅对字符串值）
	// 例如：{"type": reply} -> {"type": "reply"}
	str = strings.ReplaceAll(str, `: reply`, `: "reply"`)
	str = strings.ReplaceAll(str, `: heartbeat`, `: "heartbeat"`)
	str = strings.ReplaceAll(str, `: usb_print_report`, `: "usb_print_report"`)
	str = strings.ReplaceAll(str, `: lan_print_report`, `: "lan_print_report"`)
	return []byte(str)
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

	// 创建连接专用的context
	connCtx, connCancel := context.WithCancel(globalCtx)
	defer connCancel()

	// 发送ping消息 - 使用context控制协程生命周期
	go func() {
		for {
			select {
			case <-connCtx.Done():
				// 连接关闭或服务停止时退出
				return
			case <-ticker.C:
				// 使用安全的写入函数发送ping消息
				if err := safeWriteMessage(newConn, websocket.PingMessage, nil, 10*time.Second); err != nil {
					fmt.Printf("[%s] Error sending ping message: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, newConn.DeviceId)
					connCancel() // 发送失败时取消context
					return
				}
			}
		}
	}()

	// 添加资源清理
	defer func() {
		// 清理连接资源
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

		// 停止心跳协程
		connCancel()

		fmt.Printf("[%s] 连接已清理 DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId)
	}()

	// 主循环
	for {
		// 检查context是否已取消
		select {
		case <-connCtx.Done():
			fmt.Printf("[%s] 连接被取消，退出读取循环 DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId)
			return
		default:
			// 继续读取消息
		}

		// 检查连接是否仍然有效
		if ws == nil {
			fmt.Printf("[%s] WebSocket连接已断开 DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId)
			break
		}

		// 设置读取超时
		ws.SetReadDeadline(time.Now().Add(30 * time.Second))

		// 读取消息
		_, msg, err := ws.ReadMessage()
		if err != nil {
			// 检查是否是网络临时错误，可以重试
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				fmt.Printf("[%s] 网络临时错误，重试中 DeviceId: %s: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId, err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 检查是否是超时或连接关闭错误
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("[%s] WebSocket连接异常关闭: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, newConn.DeviceId)
			} else {
				fmt.Printf("[%s] 读取消息错误: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, newConn.DeviceId)
			}
			break
		}

		// 检查消息大小，防止内存溢出
		if len(msg) > 1024*1024 { // 1MB限制
			fmt.Printf("[%s] 消息过大，忽略 DeviceId: %s, Size: %d\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId, len(msg))
			continue
		}

		// 使用recover保护消息处理
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[%s] 消息处理异常 DeviceId: %s: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId, r)
					logger.Logger.Error(fmt.Sprintf("消息处理异常 DeviceId: %s: %v", newConn.DeviceId, r))
				}
			}()
			handleMessage(ws, msg, newConn)
		}()
	}
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
		// 创建临时连接信息用于安全写入
		tempConn := &ConnectionInfo{ws: ws, writeMutex: &sync.Mutex{}}
		safeWriteMessage(tempConn, websocket.TextMessage, getMsgData(PushMessage{
			Event: "connect",
			State: constant.CodeFail,
			Msg:   "Token Error",
		}), 5*time.Second)
		ws.Close()
		return nil
	}

	// 验证设备是否绑定
	deviceRepo := repository.NewDeviceRepository(database.Instance)
	existsDevice := deviceRepo.GetRecordBySourceAndDeviceId(claims.CompanyUuid, claims.Source, claims.DeviceId)
	if existsDevice.ID == 0 {
		// 创建临时连接信息用于安全写入
		tempConn := &ConnectionInfo{ws: ws, writeMutex: &sync.Mutex{}}
		safeWriteMessage(tempConn, websocket.TextMessage, getMsgData(PushMessage{
			Event: "connect",
			State: constant.CodeFail,
			Msg:   "No binding device-id: " + claims.DeviceId,
		}), 5*time.Second)
		ws.Close()
		return nil
	}

	// 允许一个设备最多存在三个连接，超过则清空之前的连接
	// 计算当前设备的连接数量
	connCount := 0

	// 遍历现有连接，统计同设备的连接数
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == claims.CompanyUuid &&
				conn.SourceClient == existsDevice.Source &&
				conn.DeviceId == existsDevice.DeviceId &&
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
				if conn.CompanyUuid == claims.CompanyUuid &&
					conn.SourceClient == existsDevice.Source &&
					conn.DeviceId == existsDevice.DeviceId &&
					conn.ws != ws {
					conn.ws.Close()
					WsClients.Delete(key)
				}
			}
			return true
		})
	}

	// 添加新连接
	newConn := ConnectionInfo{
		CompanyUuid:   claims.CompanyUuid,
		StaffUuid:     claims.StaffUuid,
		SourceClient:  existsDevice.Source,
		DeviceId:      existsDevice.DeviceId,
		LastHeartbeat: time.Now().Format(time.RFC3339),
		ws:            ws,
		writeMutex:    &sync.Mutex{}, // 初始化写入锁
	}

	// 生成唯一的连接key
	connKey := fmt.Sprintf("%d_%s_%s_%d", claims.CompanyUuid, existsDevice.Source, existsDevice.DeviceId, time.Now().UnixNano())
	WsClients.Store(connKey, newConn)

	// 监听关闭事件以删除连接详细信息
	ws.SetCloseHandler(func(code int, text string) error {
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.CompanyUuid == claims.CompanyUuid &&
					conn.SourceClient == existsDevice.Source &&
					conn.DeviceId == existsDevice.DeviceId &&
					conn.ws == ws {
					WsClients.Delete(key)
					return false // 找到后停止遍历
				}
			}
			return true
		})
		return nil
	})

	// 发送连接成功消息
	if err := safeWriteMessage(&newConn, websocket.TextMessage, getMsgData(PushMessage{
		Event: "connect",
		State: constant.CodeSuccess,
		Msg:   "Connected successfully",
		Data: map[string]interface{}{
			"source":       claims.Source,
			"company_uuid": claims.CompanyUuid,
			"staff_uuid":   claims.StaffUuid,
			"device_id":    claims.DeviceId,
		},
	}), 10*time.Second); err != nil {
		fmt.Printf("[%s] Error sending connect success message: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, claims.DeviceId)
		return nil
	}
	return &newConn
}

// handleMessage 处理消息
func handleMessage(ws *websocket.Conn, msg []byte, newConn *ConnectionInfo) {
	// 解析消息
	clientMessage := ClientMessage{}
	err := json.Unmarshal(fixJSONFormat(msg), &clientMessage)
	if err != nil {
		fmt.Printf("[%s] WebSocket JSON解析错误 [DeviceId: %s]: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId, err)
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
	if clientMessage.Type == "heartbeat" {
		// 更新心跳时间
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.CompanyUuid == newConn.CompanyUuid &&
					conn.SourceClient == newConn.SourceClient &&
					conn.DeviceId == newConn.DeviceId &&
					conn.ws == newConn.ws {
					// 更新心跳时间
					updatedConn := conn
					updatedConn.LastHeartbeat = time.Now().Format(time.RFC3339)
					WsClients.Store(key, updatedConn)
					return false // 找到后停止遍历
				}
			}
			return true
		})

		isOnline := false
		WsClients.Range(func(key, value interface{}) bool {
			if conn, ok := value.(ConnectionInfo); ok {
				if conn.DeviceId == newConn.DeviceId {
					isOnline = true
					return false // 找到后停止遍历
				}
			}
			return true
		})

		if !isOnline {
			fmt.Printf("[%s] Heartbeat message DeviceId - 离线: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId)
			logger.Logger.Info(fmt.Sprintf("Heartbeat message DeviceId - 离线: %s", newConn.DeviceId))
		} else {
			fmt.Printf("[%s] Heartbeat message DeviceId - 在线: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.DeviceId)
			logger.Logger.Info(fmt.Sprintf("Heartbeat message DeviceId - 在线: %s", newConn.DeviceId))
		}
		return
	}

	// 处理已读删除
	if clientMessage.Type == "reply" {
		repo := repository.NewWebSocketMsgRepository(database.Instance)
		err := repo.DeleteByTypeAndId(msgId)
		if err != nil {
			fmt.Printf("[%s] Error updating message status: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
		}
		return
	}

	// 上报Usb打印机数据
	if clientMessage.Type == "usb_print_report" && newConn.SourceClient == "cashier" {
		reportUsbPrinter(newConn, clientMessage)
	}

	// 上报Lan打印机数据
	if clientMessage.Type == "lan_print_report" && newConn.SourceClient == "cashier" {
		reportLanPrinter(newConn, clientMessage)
	}
}

// checkHeartbeatTimeout 检查心跳超时的连接并断开
func checkHeartbeatTimeout() {
	// 每30秒检查一次
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-globalCtx.Done():
			// 服务停止时退出
			return
		case <-ticker.C:
			// 遍历所有连接
			var expiredKeys []interface{}
			WsClients.Range(func(key, value interface{}) bool {
				if conn, ok := value.(ConnectionInfo); ok {
					// 解析最后心跳时间
					lastHeartbeat, err := time.Parse(time.RFC3339, conn.LastHeartbeat)
					if err != nil {
						// 如果解析出错，可能是初始值或无效值，设置为当前时间
						lastHeartbeat = time.Now()
					}
					// 检查是否超过2分钟没有心跳
					if time.Since(lastHeartbeat) > 2*time.Minute {
						fmt.Printf("[%s] 断开超时2分钟没有心跳的连接: %s, 最后心跳时间: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), conn.DeviceId, conn.LastHeartbeat)
						// 关闭连接
						conn.ws.Close()
						// 记录需要删除的key
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

// 上报USB打印机数据
func reportUsbPrinter(newConn *ConnectionInfo, clientMessage ClientMessage) {
	// 根据Data的类型进行不同的处理
	dataJson, err := json.Marshal(clientMessage.Data)
	if err != nil {
		fmt.Printf("[%s] Error marshaling array data: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
		return
	}

	// 解析Data为UsbPrinter数组
	usbPrinters := []UsbPrinter{}
	err = json.Unmarshal(dataJson, &usbPrinters)
	if err != nil {
		fmt.Printf("[%s] Error unmarshaling array data: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
		return
	}

	// 已有记录
	repo := repository.NewPrinterRepository(database.Instance)
	dbUsbList := repo.GetUsbList(newConn.CompanyUuid)

	// 更新为离线
	if len(dbUsbList) > 0 {
		// 优化：创建配置JSON映射以快速查找打印机，避免嵌套循环
		printerConfigMap := make(map[string]bool)
		for _, printer := range usbPrinters {
			printerConfigMap[utils.JsonToStr(printer)] = true
		}
		// 检查数据库中的打印机是否在新上报列表中
		for _, usb := range dbUsbList {
			// 如果不在新列表中且状态为在线，则更新为离线
			if !printerConfigMap[usb.ConfigJson] && usb.Status == 1 {
				if err := repo.UpdateBySourceDeviceSn(newConn.CompanyUuid, usb.ID, newConn.DeviceId, map[string]interface{}{
					"status": 0,
				}); err != nil {
					fmt.Printf("[%s] Error updating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				}
				// 打印日志
				logger.Logger.Info(fmt.Sprintf("更新打印机状态为离线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", newConn.CompanyUuid, newConn.DeviceId, usb.Uuid, usb.Name, usb.LastHeartbeatTime))
				fmt.Printf("[%s] 更新打印机状态为离线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.CompanyUuid, newConn.DeviceId, usb.Uuid, usb.Name, usb.LastHeartbeatTime)
			}
		}
	}

	// 有新的打印机数据，遍历处理
	if len(usbPrinters) > 0 {
		// 优化：创建已有打印机映射，避免多次循环查询
		dbUsbMap := make(map[string]model.Printer)
		for _, usb := range dbUsbList {
			dbUsbMap[usb.ConfigJson] = usb
		}

		// 处理每个上报的USB打印机
		for _, usbPrinter := range usbPrinters {
			printerJson := utils.JsonToStr(usbPrinter)
			if dbUsb, exists := dbUsbMap[printerJson]; exists {
				// 已存在的打印机，更新状态
				if err := repo.Update(newConn.CompanyUuid, dbUsb.ID, map[string]interface{}{
					"status":              1,
					"last_heartbeat_time": uint(time.Now().Unix()),
					"source_device_sn":    newConn.DeviceId,
				}); err != nil {
					fmt.Printf("[%s] Error updating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				}
				// 打印日志
				if dbUsb.Status == 0 {
					logger.Logger.Info(fmt.Sprintf("更新打印机状态为在线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", newConn.CompanyUuid, newConn.DeviceId, dbUsb.Uuid, dbUsb.Name, uint(time.Now().Unix())))
					fmt.Printf("[%s] 更新打印机状态为在线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.CompanyUuid, newConn.DeviceId, dbUsb.Uuid, dbUsb.Name, uint(time.Now().Unix()))
				}
			} else {
				uuid := utils.GetID()
				printerTypeKey := constant.PRINTER_TYPE_XPRINTER_LAN
				// 区分类型
				if usbPrinter.Vid.(float64) == 1137 && usbPrinter.Pid.(float64) == 85 {
					if usbPrinter.M_name == "Zhuhai Howbest Label Printer Co.,Ltd." {
						printerTypeKey = constant.PRINTER_TYPE_GP_C200IV
					}
					if usbPrinter.M_name == "ZHU HAI HOWBEST Receipt Printer Co.,Ltd." {
						printerTypeKey = constant.PRINTER_TYPE_GP_D300I
					}
				}
				// 获取打印机类型(只查询一次)
				printerType := repository.NewPrinterTypeRepository(database.Instance).GetRecordByKey(newConn.CompanyUuid, printerTypeKey)
				if printerType.ID == 0 {
					fmt.Printf("[%s] Error: printer type XPRINTER_LAN not found\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"))
					logger.Logger.Error(fmt.Sprintf("Error: printer type XPRINTER_LAN not found\n"))
					return
				}
				// 新打印机，创建记录
				if err := repo.Create(newConn.CompanyUuid, model.Printer{
					Uuid:              uuid,
					Name:              usbPrinter.Name,
					PrinterTypeUuid:   printerType.Uuid,
					ConfigJson:        printerJson,
					Copies:            1,
					Sort:              0,
					IsUsb:             1,
					SourceDeviceSn:    newConn.DeviceId,
					CreateTime:        uint(time.Now().Unix()),
					UpdateTime:        uint(time.Now().Unix()),
					Status:            1,
					LastHeartbeatTime: uint(time.Now().Unix()),
				}); err != nil {
					fmt.Printf("[%s] Error creating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				}
				// 打印日志
				logger.Logger.Info(fmt.Sprintf("新增打印机: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", newConn.CompanyUuid, newConn.DeviceId, uuid, usbPrinter.Name, uint(time.Now().Unix())))
				fmt.Printf("[%s] 新增打印机: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), newConn.CompanyUuid, newConn.DeviceId, uuid, usbPrinter.Name, uint(time.Now().Unix()))
			}
		}
	}
}

// 上报Lan打印机数据
func reportLanPrinter(newConn *ConnectionInfo, clientMessage ClientMessage) {
	// 根据Data的类型进行不同的处理
	dataJson, err := json.Marshal(clientMessage.Data)
	if err != nil {
		fmt.Printf("[%s] Error marshaling array data: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
		return
	}

	// 解析Data为LanPrinter数组
	lanPrinters := []LanPrinter{}
	err = json.Unmarshal(dataJson, &lanPrinters)
	if err != nil {
		fmt.Printf("[%s] Error unmarshaling array data: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
		return
	}

	// 已有记录
	repo := repository.NewLanPrinterScanRepository(database.Instance)
	dbLanList := repo.GetList(newConn.CompanyUuid)

	// 更新为离线
	if len(dbLanList) > 0 {
		// 优化：创建配置JSON映射以快速查找打印机，避免嵌套循环
		printerConfigMap := make(map[string]bool)
		for _, printer := range lanPrinters {
			printerConfigMap[printer.Ip] = true
		}
		// 检查数据库中的打印机是否在新上报列表中
		for _, lanPrinter := range dbLanList {
			// 如果不在新列表中且状态为在线，则更新为离线
			if !printerConfigMap[lanPrinter.Ip] && lanPrinter.Status == 1 {
				if err := repo.Update(newConn.CompanyUuid, lanPrinter.ID, map[string]interface{}{
					"status": 0,
				}); err != nil {
					fmt.Printf("[%s] Error updating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				}
			}
		}
	}

	// 有新的打印机数据，遍历处理
	if len(lanPrinters) > 0 {
		// 优化：创建已有打印机映射，避免多次循环查询
		dbLanMap := make(map[string]model.LanPrinterScan)
		for _, lanPrinter := range dbLanList {
			dbLanMap[lanPrinter.Ip] = lanPrinter
		}
		// 处理每个上报的Lan打印机
		for _, lanPrinter := range lanPrinters {
			if dbLan, exists := dbLanMap[lanPrinter.Ip]; exists {
				// 已存在的打印机，更新状态
				if err := repo.Update(newConn.CompanyUuid, dbLan.ID, map[string]interface{}{
					"status":           1,
					"source_device_sn": newConn.DeviceId,
				}); err != nil {
					fmt.Printf("[%s] Error updating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
					logger.Logger.Error(fmt.Sprintf("Error updating usb print: %v\n", err))
				}
			} else if lanPrinter.Ip != "" && lanPrinter.Port != 0 {
				if err := repo.Create(newConn.CompanyUuid, model.LanPrinterScan{
					Uuid:           utils.GetID(),
					Ip:             lanPrinter.Ip,
					Port:           lanPrinter.Port,
					SourceDeviceSn: newConn.DeviceId,
					CreateTime:     uint(time.Now().Unix()),
					UpdateTime:     uint(time.Now().Unix()),
					Status:         1,
				}); err != nil {
					fmt.Printf("[%s] Error creating usb print: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
					logger.Logger.Error(fmt.Sprintf("Error creating usb print: %v\n", err))
				}
			}
		}
	}
}

// PushClient 推送消息
func PushClient(messageData MessageData) {
	// 复用数据库连接，避免重复创建
	repo := repository.NewWebSocketMsgRepository(database.Instance)

	// 1. 先删后加 - 同一个类型，只保留最新的
	err := repo.DeleteByTypeAndCompanyId(messageData.MessageType, messageData.CompanyUuid)
	if err != nil {
		fmt.Printf("[%s] Error deleting old messages: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
	}

	// 收集匹配的连接
	var matchedConnections []ConnectionInfo
	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			if conn.CompanyUuid == messageData.CompanyUuid &&
				(conn.StaffUuid == messageData.StaffUuid || messageData.StaffUuid == 0) &&
				(conn.StaffUuid != messageData.NotStaffUuid || messageData.NotStaffUuid == 0) &&
				(conn.SourceClient == messageData.SourceClient || messageData.SourceClient == "*") &&
				(conn.DeviceId == messageData.DeviceId || messageData.DeviceId == "*") &&
				(conn.DeviceId != messageData.NotDeviceId || messageData.NotDeviceId == "*" || messageData.NotDeviceId == "") {
				matchedConnections = append(matchedConnections, conn)
			}
		}
		return true
	})

	// 批量处理匹配的连接
	for _, conn := range matchedConnections {
		// 2. 创建消息记录
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
			fmt.Printf("[%s] Error creating WebSocket message: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
			logger.Logger.Error(fmt.Sprintf("Error creating WebSocket message: %v\n", err))
			continue
		}

		// 发送消息
		message := PushMessage{
			Event: messageData.MessageType,
			State: constant.CodeSuccess,
			Data:  messageData.Data,
			MsgId: int(id),
		}

		// 使用安全的写入函数发送消息
		err = safeWriteMessage(&conn, websocket.TextMessage, getMsgData(message), 10*time.Second)
		if err != nil {
			fmt.Printf("[%s] Error sending message to client: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, conn.DeviceId)
			logger.Logger.Error(fmt.Sprintf("Error sending message to client: %v DeviceId: %s\n", err, conn.DeviceId))
		} else {
			fmt.Printf("[%s] 推送消息: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), utils.StructToJson(map[string]interface{}{
				"SourceClient": conn.SourceClient,
				"CompanyUuid":  messageData.CompanyUuid,
				"DeviceId":     conn.DeviceId,
				"Event":        messageData.MessageType,
				"State":        constant.CodeSuccess,
				"Data":         messageData.Data,
				"MsgId":        int(id),
			}))
			logger.Logger.Info(fmt.Sprintf("推送消息: SourceClient=%s, CompanyUuid=%d, DeviceId=%s, Event=%s, State=%d, Data=%s, MsgId=%d", conn.SourceClient, messageData.CompanyUuid, conn.DeviceId, messageData.MessageType, constant.CodeSuccess, utils.StructToJson(messageData.Data), id))
		}
	}
}

// GetConnectionStats 获取连接统计信息
func GetConnectionStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_connections": 0,
		"by_company":        make(map[uint64]int),
		"by_source":         make(map[string]int),
		"by_device":         make(map[string]int),
	}

	WsClients.Range(func(key, value interface{}) bool {
		if conn, ok := value.(ConnectionInfo); ok {
			stats["total_connections"] = stats["total_connections"].(int) + 1

			// 按公司统计
			if byCompany, ok := stats["by_company"].(map[uint64]int); ok {
				byCompany[conn.CompanyUuid]++
			}

			// 按来源统计
			if bySource, ok := stats["by_source"].(map[string]int); ok {
				bySource[conn.SourceClient]++
			}

			// 按设备统计
			if byDevice, ok := stats["by_device"].(map[string]int); ok {
				byDevice[conn.DeviceId]++
			}
		}
		return true
	})

	return stats
}
