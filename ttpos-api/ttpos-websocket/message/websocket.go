// Package message 定义 ttpos-websocket 服务相关的消息结构体
package message

import (
	"time"
	"ttpos-api/common/message"
	"ttpos-api/ttpos-websocket/constant"
)

// LanPrinterInfo LAN 打印机信息
// 用于描述局域网内的打印机设备
type LanPrinterInfo struct {
	IP     string `json:"ip"`     // IP地址
	Port   int    `json:"port"`   // 端口号
	Status int    `json:"status"` // 状态（1: 在线, 0: 离线）
	Remark string `json:"remark"` // 备注信息
}

// LanPrinterReportMessage LAN 打印机上报消息
// 用于收银机上报局域网内的打印机信息到消息队列
type LanPrinterReportMessage struct {
	message.BaseMessage
	CompanyUUID  uint64           `json:"company_uuid"`  // 公司UUID
	StaffUUID    uint64           `json:"staff_uuid"`    // 员工UUID
	DeviceID     string           `json:"device_id"`     // 设备ID
	SourceClient string           `json:"source_client"` // 来源客户端（cashier/kitchen/waiter）
	Printers     []LanPrinterInfo `json:"printers"`      // 打印机列表
	ReportTime   string           `json:"report_time"`   // 上报时间（格式化字符串）
}

// NewLanPrinterReportMessage 创建 LAN 打印机上报消息
func NewLanPrinterReportMessage(companyUUID, staffUUID uint64, deviceID, sourceClient string, printers []LanPrinterInfo) *LanPrinterReportMessage {
	return &LanPrinterReportMessage{
		BaseMessage:  message.NewBaseMessage(constant.TopicLanPrinterReport),
		CompanyUUID:  companyUUID,
		StaffUUID:    staffUUID,
		DeviceID:     deviceID,
		SourceClient: sourceClient,
		Printers:     printers,
		ReportTime:   time.Now().Format(time.DateTime),
	}
}

// Validate 验证消息字段
func (m *LanPrinterReportMessage) Validate() error {
	// 验证基础字段
	if err := m.BaseMessage.Validate(); err != nil {
		return err
	}

	// 验证业务字段
	if m.CompanyUUID == 0 {
		return constant.ErrCompanyUUIDRequired
	}
	if m.DeviceID == "" {
		return constant.ErrDeviceIDRequired
	}
	return nil
}
