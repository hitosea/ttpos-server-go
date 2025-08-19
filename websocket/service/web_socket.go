package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
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

// WsClients 存储所有的连接
var WsClients []ConnectionInfo

// 初始化函数，启动定时检查心跳超时的连接
func init() {
	go checkHeartbeatTimeout()
	// go checkPrinterHeartbeatTimeout()
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

	// 发送ping消息
	go func() {
		for range ticker.C {
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("[%s] Error sending ping message: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				return
			}
		}
	}()

	// 主循环
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("[%s] Error reading message: %v DeviceId: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err, newConn.DeviceId)
			break
		}
		// 处理消息
		handleMessage(ws, msg, newConn)
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
			Msg:   "No binding device-id: " + claims.DeviceId,
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
		fmt.Printf("[%s] Error sending ping message: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
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
					fmt.Printf("[%s] 断开超时2分钟没有心跳的连接: %s, 最后心跳时间: %s\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), conn.DeviceId, conn.LastHeartbeat)
					// 关闭连接
					conn.ws.Close()
					// 从列表中移除
					WsClients = slices.Delete(WsClients, i, i+1)
				}
			}
		}
	}
}

// checkPrinterHeartbeatTimeout 检查打印机心跳超时并更新状态为离线
func checkPrinterHeartbeatTimeout() {
	// 每5秒检查一次
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 创建打印机仓库
		repo := repository.NewPrinterRepository(database.Instance)
		// 获取所有公司ID - 从现有连接中提取
		companyUuids := make(map[uint64]ConnectionInfo)
		for _, conn := range WsClients {
			companyUuids[conn.CompanyUuid] = conn
		}
		// 遍历每个公司
		for companyUuid, conn := range companyUuids {
			// 获取该公司的所有USB打印机
			printers := repo.GetUsbListByStatus(companyUuid, 1)
			// 当前时间戳
			currentTime := uint(time.Now().Unix())
			// 遍历所有打印机
			for _, printer := range printers {
				// 如果打印机状态为在线，且最后心跳时间超过8秒，则更新为离线
				if printer.Status == 1 && (currentTime-printer.LastHeartbeatTime) > 8 {
					if err := repo.UpdateBySourceDeviceSn(companyUuid, printer.ID, conn.DeviceId, map[string]interface{}{
						"status": 0,
					}); err != nil {
						fmt.Printf("[%s] Error updating printer status to offline: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
					} else {
						logger.Logger.Info(fmt.Sprintf("更新打印机状态为离线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", companyUuid, conn.DeviceId, printer.Uuid, printer.Name, printer.LastHeartbeatTime))
						fmt.Printf("[%s] 更新打印机状态为离线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), companyUuid, conn.DeviceId, printer.Uuid, printer.Name, printer.LastHeartbeatTime)
					}
				}
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
				fmt.Printf("[%s] Error deleting old messages: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
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

			err = conn.ws.WriteMessage(websocket.TextMessage, getMsgData(message))
			if err != nil {
				fmt.Printf("[%s] Error sending message to client: %v\n", time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05"), err)
				logger.Logger.Error(fmt.Sprintf("Error sending message to client: %v\n", err))
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
}
