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

// lanPrinterReportHandler 处理 LAN 打印机上报
func lanPrinterReportHandler(ctx context.Context, msg *primitive.MessageExt) error {
	if config.Server.Mode == gin.DebugMode {
		logger.Logger.Info("收到 LAN 打印机上报消息",
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

	// 更新为离线：检查数据库中的打印机是否在新上报列表中
	if len(dbLanList) > 0 {
		// 优化：创建上报打印机映射以快速查找，避免嵌套循环
		// 使用 ip:port 作为key，因为同一个IP可能有不同端口
		printerReportMap := make(map[string]bool)
		for _, printer := range reportMsg.Printers {
			if printer.IP != "" && printer.Port > 0 {
				printerKey := getPrinterKey(printer.IP, printer.Port)
				printerReportMap[printerKey] = true
			}
		}

		// 检查数据库中的打印机是否在新上报列表中
		for _, lanPrinter := range dbLanList {
			printerKey := getPrinterKey(lanPrinter.Ip, lanPrinter.Port)
			// 如果不在新列表中且状态为在线，则更新为离线
			if !printerReportMap[printerKey] && lanPrinter.Status == 1 {
				err := lanPrinterScanRepo.Update(lanPrinter.ID, map[string]interface{}{
					"status": 0,
				})
				if err != nil {
					logger.Logger.Error("更新打印机为离线状态失败",
						zap.Error(err),
						zap.String("ip", lanPrinter.Ip),
						zap.Int("port", lanPrinter.Port),
						zap.Uint("id", lanPrinter.ID),
					)
				} else {
					if config.Server.Mode == gin.DebugMode {
						logger.Logger.Debug("更新打印机为离线状态",
							zap.String("ip", lanPrinter.Ip),
							zap.Int("port", lanPrinter.Port),
						)
					}
				}
			}
		}
	}

	// 处理上报的打印机：创建或更新为在线
	if len(reportMsg.Printers) > 0 {
		// 优化：创建已有打印机映射，避免多次循环查询
		dbLanMap := make(map[string]model.LanPrinterScan)
		for _, lanPrinter := range dbLanList {
			printerKey := getPrinterKey(lanPrinter.Ip, lanPrinter.Port)
			dbLanMap[printerKey] = lanPrinter
		}

		// 处理每个上报的打印机
		for _, printer := range reportMsg.Printers {
			if printer.IP == "" || printer.Port == 0 {
				if config.Server.Mode == gin.DebugMode {
					logger.Logger.Warn("跳过无效的打印机信息",
						zap.String("ip", printer.IP),
						zap.Int("port", printer.Port),
					)
				}
				continue
			}

			printerKey := getPrinterKey(printer.IP, printer.Port)
			if dbLan, exists := dbLanMap[printerKey]; exists {
				// 已存在的打印机，更新状态为在线
				err := lanPrinterScanRepo.Update(dbLan.ID, map[string]interface{}{
					"status":      1,
					"remark":      printer.Remark,
					"update_time": time.Now().Unix(),
				})
				if err != nil {
					logger.Logger.Error("更新打印机状态失败",
						zap.Error(err),
						zap.String("ip", printer.IP),
						zap.Int("port", printer.Port),
					)
				} else {
					if config.Server.Mode == gin.DebugMode {
						logger.Logger.Debug("更新打印机为在线状态",
							zap.String("ip", printer.IP),
							zap.Int("port", printer.Port),
						)
					}
				}
			} else {
				// 不存在的打印机，创建新记录
				newScan := model.LanPrinterScan{
					Ip:             printer.IP,
					Port:           printer.Port,
					Status:         1, // 新上报的打印机默认为在线
					Remark:         printer.Remark,
					SourceDeviceSn: reportMsg.DeviceID,
				}
				now := time.Now().Unix()
				newScan.CreateTime = now
				newScan.UpdateTime = now

				err := lanPrinterScanRepo.Create(newScan)
				if err != nil {
					logger.Logger.Error("创建打印机记录失败",
						zap.Error(err),
						zap.String("ip", printer.IP),
						zap.Int("port", printer.Port),
					)
				} else {
					if config.Server.Mode == gin.DebugMode {
						logger.Logger.Debug("创建打印机记录",
							zap.String("ip", printer.IP),
							zap.Int("port", printer.Port),
							zap.String("device_sn", reportMsg.DeviceID),
						)
					}
				}
			}
		}
	}

	if config.Server.Mode == gin.DebugMode {
		logger.Logger.Info("LAN 打印机上报处理完成",
			zap.String("msg_id", msg.MsgId),
			zap.Uint64("company_uuid", reportMsg.CompanyUUID),
			zap.String("device_id", reportMsg.DeviceID),
			zap.Int("printer_count", len(reportMsg.Printers)),
		)
	}

	return nil
}

// getPrinterKey 生成打印机唯一标识（ip:port）
func getPrinterKey(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}
