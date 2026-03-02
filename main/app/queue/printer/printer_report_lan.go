package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	ttposWebsocketMsg "ttpos-api/ttpos-websocket/message"
)

// LanPrinterReportHandler 处理 LAN 打印机上报
func LanPrinterReportHandler(ctx context.Context, msg *primitive.MessageExt) error {
	if config.Server.Mode == gin.DebugMode {
		logger.Logger.Debug("收到 LAN 打印机上报消息",
			zap.String("msg_id", msg.MsgId),
			zap.String("topic", msg.Topic),
		)
	}

	// 解析消息体
	var reportMsg ttposWebsocketMsg.LanPrinterReportMessage
	err := json.Unmarshal(msg.Body, &reportMsg)
	if err != nil {
		logger.Logger.Error("解析 LAN 打印机上报消息失败",
			zap.Error(err),
			zap.String("msg_id", msg.MsgId),
			zap.ByteString("body", msg.Body),
		)
		return err
	}

	// 验证消息
	if err := reportMsg.Validate(); err != nil {
		logger.Logger.Error("验证 LAN 打印机上报消息失败",
			zap.Error(err),
			zap.String("msg_id", msg.MsgId),
			zap.ByteString("body", msg.Body),
		)
		return reportMsg.Validate()
	}

	// 获取数据库连接
	dbm := database.GetDBManager(config.Database)
	db := dbm.GetDB(reportMsg.CompanyUUID)
	if db == nil {
		logger.Logger.Error("获取数据库连接失败",
			zap.Error(err),
			zap.String("msg_id", msg.MsgId),
			zap.ByteString("body", msg.Body),
		)
		return err
	}

	// 获取 Repository
	lanPrinterScanRepo := repository.NewLanPrinterScanRepository(db)

	// 获取该设备已有的打印机记录
	dbLanList := lanPrinterScanRepo.GetListByDeviceSn(reportMsg.DeviceID)

	// 构建上报打印机映射 (ip:port -> true)
	printerReportMap := make(map[string]bool)
	for _, printer := range reportMsg.Printers {
		if printer.IP != "" && printer.Port > 0 {
			printerReportMap[getPrinterKey(printer.IP, printer.Port)] = true
		}
	}

	// 构建已有打印机映射 (ip:port -> record)
	dbLanMap := make(map[string]model.LanPrinterScan)
	for _, lanPrinter := range dbLanList {
		dbLanMap[getPrinterKey(lanPrinter.Ip, lanPrinter.Port)] = lanPrinter
	}

	// 收集需要批量更新为离线的 ID
	var offlineIDs []uint
	for _, lanPrinter := range dbLanList {
		printerKey := getPrinterKey(lanPrinter.Ip, lanPrinter.Port)
		if !printerReportMap[printerKey] && lanPrinter.Status == 1 {
			offlineIDs = append(offlineIDs, lanPrinter.ID)
		}
	}

	// 批量更新为离线
	if len(offlineIDs) > 0 {
		if err := lanPrinterScanRepo.BatchUpdateByIDs(offlineIDs, map[string]any{
			"status": 0,
		}); err != nil {
			logger.Logger.Error("批量更新打印机为离线状态失败",
				zap.Error(err),
				zap.Uints("ids", offlineIDs),
			)
		} else if config.Server.Mode == gin.DebugMode {
			logger.Logger.Debug("批量更新打印机为离线状态",
				zap.Uints("ids", offlineIDs),
			)
		}
	}

	// 收集需要批量更新为在线的 ID，以及需要新增的记录
	var onlineIDs []uint
	var newScans []model.LanPrinterScan
	now := time.Now().Unix()
	for _, printer := range reportMsg.Printers {
		if printer.IP == "" || printer.Port == 0 {
			continue
		}
		printerKey := getPrinterKey(printer.IP, printer.Port)
		if dbLan, exists := dbLanMap[printerKey]; exists {
			onlineIDs = append(onlineIDs, dbLan.ID)
		} else {
			newScans = append(newScans, model.LanPrinterScan{
				Ip:             printer.IP,
				Port:           printer.Port,
				Status:         1,
				Remark:         printer.Remark,
				SourceDeviceSn: reportMsg.DeviceID,
				BaseModel:      model.BaseModel{CreateTime: now, UpdateTime: now},
			})
		}
	}

	// 批量更新为在线
	if len(onlineIDs) > 0 {
		if err := lanPrinterScanRepo.BatchUpdateByIDs(onlineIDs, map[string]any{
			"status":      1,
			"update_time": now,
		}); err != nil {
			logger.Logger.Error("批量更新打印机为在线状态失败",
				zap.Error(err),
				zap.Uints("ids", onlineIDs),
			)
		}
	}

	// 批量创建新记录
	for _, newScan := range newScans {
		if err := lanPrinterScanRepo.Create(newScan); err != nil {
			logger.Logger.Error("创建打印机记录失败",
				zap.Error(err),
				zap.String("ip", newScan.Ip),
				zap.Int("port", newScan.Port),
			)
		} else if config.Server.Mode == gin.DebugMode {
			logger.Logger.Debug("创建打印机记录",
				zap.String("ip", newScan.Ip),
				zap.Int("port", newScan.Port),
				zap.String("device_sn", reportMsg.DeviceID),
			)
		}
	}

	return nil
}

// getPrinterKey 生成打印机唯一标识（ip:port）
func getPrinterKey(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}
