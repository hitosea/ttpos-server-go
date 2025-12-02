# LAN 打印机上报功能实现

## 📋 功能概述

在 main 项目中实现了订阅 LAN 打印机上报消息的功能。当收银机客户端通过 WebSocket 上报局域网打印机信息后，ttpos-websocket 服务会将消息发布到 RocketMQ 的 `lan-printer-report` 主题，main 服务订阅该主题并处理打印机信息。

## 🔄 工作流程

```
收银机客户端 → WebSocket → ttpos-websocket → RocketMQ → main 服务订阅处理
```

## 📁 文件说明

### 1. `/main/app/queue/lan_printer_report.go`

LAN 打印机上报消息处理器，负责：
- 解析 RocketMQ 消息
- 验证消息字段
- 保存或更新打印机扫描记录到数据库

**核心函数：**

- `lanPrinterReportHandler`: 处理 LAN 打印机上报消息的主函数
- `saveLanPrinterScan`: 保存或更新打印机扫描记录

**消息体结构：**

```go
type LanPrinterReportMessage struct {
    CompanyUUID  uint64           `json:"company_uuid"`  // 公司UUID
    StaffUUID    uint64           `json:"staff_uuid"`    // 员工UUID
    DeviceID     string           `json:"device_id"`     // 设备ID
    SourceClient string           `json:"source_client"` // 来源客户端（cashier/kitchen/waiter）
    Printers     []LanPrinterInfo `json:"printers"`      // 打印机列表
    ReportTime   string           `json:"report_time"`   // 上报时间（格式化字符串）
    Timestamp    int64            `json:"timestamp"`     // 上报时间戳（Unix时间戳）
}

type LanPrinterInfo struct {
    IP     string `json:"ip"`     // IP地址
    Port   int    `json:"port"`   // 端口号
    Status int    `json:"status"` // 状态（1: 在线, 0: 离线）
    Remark string `json:"remark"` // 备注信息
}
```

### 2. `/main/app/queue/queue.go`

队列初始化文件，添加了：
- `TopicLanPrinterReport` 常量定义
- 在 `Init()` 函数中订阅 `lan-printer-report` 主题

## 🗄️ 数据库操作

### 表：`ttpos_lan_printer_scan`

存储 LAN 打印机扫描记录，字段包括：
- `ip`: 打印机 IP 地址
- `port`: 打印机端口号
- `status`: 打印机状态（1: 在线, 0: 离线）
- `remark`: 备注信息
- `source_device_sn`: 来源设备 SN

### 数据处理逻辑

1. 根据 `ip`、`port`、`source_device_sn` 查询记录是否存在
2. 如果不存在，创建新记录
3. 如果存在，更新 `status` 和 `remark` 字段

## 📝 日志记录

系统会记录以下日志：

- **Info 级别**：
  - 收到 LAN 打印机上报消息
  - LAN 打印机上报详情（公司、设备、打印机数量等）
  - LAN 打印机上报处理完成

- **Debug 级别**：
  - 处理每个打印机信息
  - 创建/更新打印机扫描记录

- **Error 级别**：
  - 解析消息失败
  - 保存/更新记录失败

- **Warn 级别**：
  - 公司UUID或设备ID为空

## 🚀 使用方式

### 1. 启动服务

服务启动时会自动初始化队列订阅：

```go
// main/app/queue/queue.go
func Init() {
    // ...
    
    // 订阅 LAN 打印机上报消息
    err = manager.Subscribe(config.Rocketmq.GroupName, TopicLanPrinterReport, lanPrinterReportHandler)
    if err != nil {
        logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicLanPrinterReport))
    }
}
```

### 2. 消息示例

WebSocket 客户端发送：

```json
{
  "type": "lan_print_report",
  "data": [
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
  ]
}
```

RocketMQ 消息体：

```json
{
  "company_uuid": 8609817471094784,
  "staff_uuid": 1234567890,
  "device_id": "DEVICE-001",
  "source_client": "cashier",
  "printers": [
    {
      "ip": "192.168.1.100",
      "port": 9100,
      "status": 1,
      "remark": "前台打印机"
    }
  ],
  "report_time": "2025-11-16 10:30:00",
  "timestamp": 1700123400
}
```

## 🔍 监控和调试

### 查看处理日志

```bash
# 查看 main 服务日志
tail -f logs/main.log | grep "LAN 打印机"
```

### 查看数据库记录

```sql
-- 查看某个设备上报的打印机
SELECT * FROM ttpos_lan_printer_scan 
WHERE source_device_sn = 'DEVICE-001' 
AND delete_time = 0;

-- 查看在线的打印机
SELECT * FROM ttpos_lan_printer_scan 
WHERE status = 1 
AND delete_time = 0;
```

## ⚠️ 注意事项

1. **消息验证**：会验证 `company_uuid` 和 `device_id` 是否为空，为空则跳过处理
2. **错误处理**：单个打印机记录保存失败不会影响其他打印机的处理
3. **幂等性**：相同的打印机信息多次上报会更新而非重复创建
4. **数据库连接**：根据 `company_uuid` 获取对应公司的数据库连接

## 🎯 后续扩展

可根据业务需求扩展以下功能：

1. 打印机状态变更通知（通过 WebSocket 推送）
2. 打印机上下线日志记录
3. 打印机健康检查和告警
4. 打印机使用统计和分析
5. 自动发现和配置新打印机

