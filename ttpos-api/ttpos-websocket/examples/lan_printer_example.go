package examples

import (
	"context"
	"fmt"
	"ttpos-api/ttpos-websocket/constant"
	"ttpos-api/ttpos-websocket/message"
)

// LanPrinterReportExample LAN 打印机上报消息示例
func LanPrinterReportExample() {
	// 1. 创建打印机列表
	printers := []message.LanPrinterInfo{
		{
			IP:     "192.168.1.100",
			Port:   9100,
			Status: 1,
			Remark: "前台打印机",
		},
		{
			IP:     "192.168.1.101",
			Port:   9100,
			Status: 1,
			Remark: "厨房打印机",
		},
		{
			IP:     "192.168.1.102",
			Port:   9100,
			Status: 0,
			Remark: "备用打印机（离线）",
		},
	}

	// 2. 创建 LAN 打印机上报消息
	msg := message.NewLanPrinterReportMessage(
		8609817471094784, // companyUUID
		1234567890,       // staffUUID
		"cashier_001",    // deviceID
		"cashier",        // sourceClient
		printers,
	)

	// 3. 设置上报时间
	msg.ReportTime = "2025-11-15 16:45:30"
	msg.Timestamp = 1731689130

	// 4. 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 5. 序列化为 JSON
	jsonData, err := msg.ToJSON()
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	fmt.Printf("消息序列化成功:\n%s\n", string(jsonData))

	// 6. 发布到消息队列（示例）
	// 实际使用时需要替换为真实的队列发布代码
	publishToQueue(context.Background(), constant.TopicLanPrinterReport, msg)
}

// publishToQueue 发布消息到队列（示例函数）
func publishToQueue(ctx context.Context, topic string, msg interface{}) {
	fmt.Printf("\n发布消息到队列:\n")
	fmt.Printf("Topic: %s\n", topic)
	fmt.Printf("Message: %+v\n", msg)

	// 实际使用时，这里应该调用真实的队列发布代码
	// 例如：queue.PushWithContext(ctx, topic, msg)
}

// LanPrinterReportConsumerExample LAN 打印机上报消息消费示例
func LanPrinterReportConsumerExample() {
	// 模拟从队列接收到的 JSON 数据
	jsonData := []byte(`{
		"message_id": "550e8400-e29b-41d4-a716-446655440000",
		"topic": "lan-printer-report",
		"timestamp": 1731689130,
		"version": "1.0",
		"company_uuid": "8609817471094784",
		"company_uuid": 8609817471094784,
		"staff_uuid": 1234567890,
		"device_id": "cashier_001",
		"source_client": "cashier",
		"printers": [
			{
				"ip": "192.168.1.100",
				"port": 9100,
				"status": 1,
				"remark": "前台打印机"
			},
			{
				"ip": "192.168.1.101",
				"port": 9100,
				"status": 1,
				"remark": "厨房打印机"
			}
		],
		"report_time": "2025-11-15 16:45:30",
		"timestamp": 1731689130
	}`)

	// 1. 反序列化消息
	var msg message.LanPrinterReportMessage
	if err := msg.FromJSON(jsonData); err != nil {
		fmt.Printf("反序列化失败: %v\n", err)
		return
	}

	// 2. 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 3. 处理消息
	fmt.Printf("\n接收到 LAN 打印机上报消息:\n")
	fmt.Printf("公司UUID: %d\n", msg.CompanyUUID)
	fmt.Printf("员工UUID: %d\n", msg.StaffUUID)
	fmt.Printf("设备ID: %s\n", msg.DeviceID)
	fmt.Printf("来源客户端: %s\n", msg.SourceClient)
	fmt.Printf("上报时间: %s\n", msg.ReportTime)
	fmt.Printf("打印机数量: %d\n", len(msg.Printers))

	// 4. 处理每个打印机信息
	for i, printer := range msg.Printers {
		fmt.Printf("\n打印机 %d:\n", i+1)
		fmt.Printf("  IP地址: %s\n", printer.IP)
		fmt.Printf("  端口: %d\n", printer.Port)
		fmt.Printf("  状态: %s\n", getPrinterStatusText(printer.Status))
		fmt.Printf("  备注: %s\n", printer.Remark)

		// 实际业务逻辑
		// 例如：更新数据库中的打印机配置
		// updatePrinterConfig(msg.CompanyUUID, msg.DeviceID, printer)
	}
}

// getPrinterStatusText 获取打印机状态文本
func getPrinterStatusText(status int) string {
	if status == 1 {
		return "在线"
	}
	return "离线"
}

// ValidateLanPrinterExample 验证 LAN 打印机消息示例
func ValidateLanPrinterExample() {
	// 测试用例 1: 正常消息
	fmt.Println("测试用例 1: 正常消息")
	validMsg := message.NewLanPrinterReportMessage(
		8609817471094784,
		1234567890,
		"cashier_001",
		"cashier",
		[]message.LanPrinterInfo{
			{IP: "192.168.1.100", Port: 9100, Status: 1, Remark: "前台打印机"},
		},
	)
	if err := validMsg.Validate(); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 验证通过")
	}

	// 测试用例 2: 缺少公司UUID
	fmt.Println("\n测试用例 2: 缺少公司UUID")
	invalidMsg1 := message.NewLanPrinterReportMessage(
		0, // 公司UUID为0
		1234567890,
		"cashier_001",
		"cashier",
		[]message.LanPrinterInfo{
			{IP: "192.168.1.100", Port: 9100, Status: 1, Remark: "前台打印机"},
		},
	)
	if err := invalidMsg1.Validate(); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 验证通过")
	}

	// 测试用例 3: 打印机列表为空
	fmt.Println("\n测试用例 3: 打印机列表为空")
	invalidMsg2 := message.NewLanPrinterReportMessage(
		8609817471094784,
		1234567890,
		"cashier_001",
		"cashier",
		[]message.LanPrinterInfo{}, // 空列表
	)
	if err := invalidMsg2.Validate(); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 验证通过")
	}

	// 测试用例 4: 打印机IP为空
	fmt.Println("\n测试用例 4: 打印机IP为空")
	invalidMsg3 := message.NewLanPrinterReportMessage(
		8609817471094784,
		1234567890,
		"cashier_001",
		"cashier",
		[]message.LanPrinterInfo{
			{IP: "", Port: 9100, Status: 1, Remark: "前台打印机"}, // IP为空
		},
	)
	if err := invalidMsg3.Validate(); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 验证通过")
	}

	// 测试用例 5: 打印机端口无效
	fmt.Println("\n测试用例 5: 打印机端口无效")
	invalidMsg4 := message.NewLanPrinterReportMessage(
		8609817471094784,
		1234567890,
		"cashier_001",
		"cashier",
		[]message.LanPrinterInfo{
			{IP: "192.168.1.100", Port: 99999, Status: 1, Remark: "前台打印机"}, // 端口超出范围
		},
	)
	if err := invalidMsg4.Validate(); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 验证通过")
	}
}
