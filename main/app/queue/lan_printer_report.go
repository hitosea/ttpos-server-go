package queue

import (
	"context"
	"encoding/json"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LanPrinterInfo LAN 打印机信息
type LanPrinterInfo struct {
	IP     string `json:"ip"`     // IP地址
	Port   int    `json:"port"`   // 端口号
	Status int    `json:"status"` // 状态（1: 在线, 0: 离线）
	Remark string `json:"remark"` // 备注信息
}

// LanPrinterReportMessage LAN 打印机上报消息体
type LanPrinterReportMessage struct {
	CompanyUUID  uint64           `json:"company_uuid"`  // 公司UUID
	StaffUUID    uint64           `json:"staff_uuid"`    // 员工UUID
	DeviceID     string           `json:"device_id"`     // 设备ID
	SourceClient string           `json:"source_client"` // 来源客户端（cashier/kitchen/waiter）
	Printers     []LanPrinterInfo `json:"printers"`      // 打印机列表
	ReportTime   string           `json:"report_time"`   // 上报时间（格式化字符串）
	Timestamp    int64            `json:"timestamp"`     // 上报时间戳（Unix时间戳）
}

// lanPrinterReportHandler 处理 LAN 打印机上报
func lanPrinterReportHandler(ctx context.Context, msg *primitive.MessageExt) error {
	logger.Logger.Info("收到 LAN 打印机上报消息",
		zap.String("msg_id", msg.MsgId),
		zap.String("topic", msg.Topic),
	)

	// 解析消息体
	var reportMsg LanPrinterReportMessage
	err := json.Unmarshal(msg.Body, &reportMsg)
	if err != nil {
		logger.Logger.Error("解析 LAN 打印机上报消息失败",
			zap.Error(err),
			zap.String("msg_id", msg.MsgId),
			zap.ByteString("body", msg.Body),
		)
		return err
	}

	// 验证必需字段
	if reportMsg.CompanyUUID == 0 {
		logger.Logger.Warn("公司UUID为空，跳过处理", zap.String("msg_id", msg.MsgId))
		return nil
	}

	if reportMsg.DeviceID == "" {
		logger.Logger.Warn("设备ID为空，跳过处理", zap.String("msg_id", msg.MsgId))
		return nil
	}

	// 记录上报详情
	logger.Logger.Info("LAN 打印机上报详情",
		zap.Uint64("company_uuid", reportMsg.CompanyUUID),
		zap.Uint64("staff_uuid", reportMsg.StaffUUID),
		zap.String("device_id", reportMsg.DeviceID),
		zap.String("source_client", reportMsg.SourceClient),
		zap.Int("printer_count", len(reportMsg.Printers)),
		zap.String("report_time", reportMsg.ReportTime),
		zap.Int64("timestamp", reportMsg.Timestamp),
	)

	// 获取数据库连接
	dbm := database.GetDBManager(config.Database)
	db := dbm.GetDB(reportMsg.CompanyUUID)

	// 获取 Repository
	lanPrinterScanRepo := repository.NewLanPrinterScanRepository(db)

	// 处理每个打印机信息
	for i, printer := range reportMsg.Printers {
		logger.Logger.Debug("处理打印机信息",
			zap.Int("index", i),
			zap.String("ip", printer.IP),
			zap.Int("port", printer.Port),
			zap.Int("status", printer.Status),
			zap.String("remark", printer.Remark),
		)

		// 保存或更新 LAN 打印机扫描记录
		err := saveLanPrinterScan(db, lanPrinterScanRepo, reportMsg.DeviceID, printer)
		if err != nil {
			logger.Logger.Error("保存 LAN 打印机扫描记录失败",
				zap.Error(err),
				zap.String("ip", printer.IP),
				zap.Int("port", printer.Port),
			)
			continue
		}
	}

	logger.Logger.Info("LAN 打印机上报处理完成",
		zap.String("msg_id", msg.MsgId),
		zap.Uint64("company_uuid", reportMsg.CompanyUUID),
		zap.String("device_id", reportMsg.DeviceID),
		zap.Int("printer_count", len(reportMsg.Printers)),
	)

	return nil
}

// saveLanPrinterScan 保存或更新 LAN 打印机扫描记录
func saveLanPrinterScan(db *gorm.DB, repo *repository.LanPrinterScanRepository, deviceSn string, printer LanPrinterInfo) error {
	// 查询是否已存在该打印机记录
	var existingScan model.LanPrinterScan
	err := db.Where("ip = ? AND port = ? AND source_device_sn = ? AND delete_time = ?",
		printer.IP, printer.Port, deviceSn, 0).
		First(&existingScan).Error

	now := time.Now().Unix()

	if err != nil {
		// 记录不存在，创建新记录
		newScan := model.LanPrinterScan{
			Ip:             printer.IP,
			Port:           printer.Port,
			Status:         printer.Status,
			Remark:         printer.Remark,
			SourceDeviceSn: deviceSn,
		}
		newScan.CreateTime = now
		newScan.UpdateTime = now

		err = db.Create(&newScan).Error
		if err != nil {
			return err
		}

		logger.Logger.Debug("创建 LAN 打印机扫描记录",
			zap.String("ip", printer.IP),
			zap.Int("port", printer.Port),
			zap.String("device_sn", deviceSn),
		)
	} else {
		// 记录已存在，更新状态和备注
		updates := map[string]interface{}{
			"status":      printer.Status,
			"remark":      printer.Remark,
			"update_time": now,
		}

		err = db.Model(&model.LanPrinterScan{}).
			Where("id = ?", existingScan.ID).
			Updates(updates).Error

		if err != nil {
			return err
		}

		logger.Logger.Debug("更新 LAN 打印机扫描记录",
			zap.String("ip", printer.IP),
			zap.Int("port", printer.Port),
			zap.String("device_sn", deviceSn),
			zap.Int("status", printer.Status),
		)
	}

	return nil
}
