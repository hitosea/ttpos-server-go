# LAN 打印机上报功能

## 📋 功能概述

LAN 打印机上报功能允许收银机客户端通过 WebSocket 连接上报局域网内的打印机信息，服务端接收后将数据发布到消息队列（MQ），供其他服务消费和处理。

## 🔄 工作流程

```
收银机客户端 → WebSocket 连接 → ttpos-websocket 服务 → 消息队列 → 其他服务消费
```

### 详细流程

1. **客户端上报**
   - 收银机扫描局域网内的打印机
   - 通过 WebSocket 发送打印机信息
   - 消息类型：`lan_print_report`

2. **服务端接收**
   - WebSocket 服务接收客户端消息
   - 解析打印机数据
   - 验证数据格式

3. **发布到 MQ**
   - 构建标准消息格式
   - 发布到 `lan-printer-report` 主题
   - 记录发布日志

4. **其他服务消费**
   - 订阅 `lan-printer-report` 主题
   - 处理打印机信息
   - 更新数据库或执行其他业务逻辑

## 📨 消息格式

### 客户端上报格式

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

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | string | 消息类型，固定为 `lan_print_report` |
| `data` | array | 打印机列表 |
| `data[].ip` | string | 打印机 IP 地址 |
| `data[].port` | int | 打印机端口号 |
| `data[].status` | int | 打印机状态（1: 在线, 0: 离线） |
| `data[].remark` | string | 打印机备注信息 |

### MQ 消息格式

服务端发布到消息队列的消息格式：

```json
{
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
    }
  ],
  "report_time": "2025-11-15 16:45:30",
  "timestamp": 1731689130
}
```

### MQ 消息字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `company_uuid` | uint64 | 公司 UUID |
| `staff_uuid` | uint64 | 员工 UUID |
| `device_id` | string | 设备 ID |
| `source_client` | string | 来源客户端（cashier/kitchen/waiter） |
| `printers` | array | 打印机列表 |
| `report_time` | string | 上报时间（格式化字符串） |
| `timestamp` | int64 | 上报时间戳（Unix 时间戳） |

## 🔧 实现细节

### 服务端代码结构

```go
// reportLanPrinter 处理 LAN 打印机上报
func (s *sWebSocket) reportLanPrinter(ctx context.Context, newConn *ConnectionInfo, clientMessage ClientMessage) {
    // 1. 解析打印机数据
    // 2. 记录日志
    // 3. 发布到 MQ
}

// publishLanPrinterReport 发布 LAN 打印机上报消息到 MQ
func (s *sWebSocket) publishLanPrinterReport(ctx context.Context, conn *ConnectionInfo, printers []LanPrinter) {
    // 1. 转换为 ttpos-api 定义的结构体
    // 2. 验证消息
    // 3. 发布到消息队列
    // 4. 记录日志
}
```

### 数据结构

**内部结构体**（WebSocket 服务内部使用）：

```go
// LanPrinter LAN 打印机信息
type LanPrinter struct {
    Ip     string `json:"ip"`     // IP地址
    Port   int    `json:"port"`   // 端口
    Status int    `json:"status"` // 状态
    Remark string `json:"remark"` // 备注
}
```

**MQ 消息结构体**（使用 `ttpos-api` 定义）：

```go
// 引入 ttpos-api
import ttposWebsocketMsg "ttpos-api/ttpos-websocket/message"

// LanPrinterInfo LAN 打印机信息
type LanPrinterInfo struct {
    IP     string `json:"ip"`     // IP地址
    Port   int    `json:"port"`   // 端口号
    Status int    `json:"status"` // 状态（1: 在线, 0: 离线）
    Remark string `json:"remark"` // 备注信息
}

// LanPrinterReportMessage LAN 打印机上报消息
type LanPrinterReportMessage struct {
    message.BaseMessage
    CompanyUUID  uint64           `json:"company_uuid"`  // 公司UUID
    StaffUUID    uint64           `json:"staff_uuid"`    // 员工UUID
    DeviceID     string           `json:"device_id"`     // 设备ID
    SourceClient string           `json:"source_client"` // 来源客户端
    Printers     []LanPrinterInfo `json:"printers"`      // 打印机列表
    ReportTime   string           `json:"report_time"`   // 上报时间
    Timestamp    int64            `json:"timestamp"`     // 上报时间戳
}
```

### 使用 ttpos-api 的优势

1. **类型安全**：统一的消息结构体定义，避免格式不一致
2. **自动验证**：内置 `Validate()` 方法，自动验证消息字段
3. **易于维护**：集中管理消息结构，便于版本升级
4. **跨服务共享**：可被多个服务引用，确保消息格式一致

## 📊 消息队列配置

### Topic 信息

- **Topic 名称**: `lan-printer-report`
- **消息类型**: JSON
- **持久化**: 是
- **消费模式**: 集群消费

### 消费者示例

```go
// 订阅 LAN 打印机上报消息
func subscribeLanPrinterReport(ctx context.Context) error {
    return queue.Subscribe(ctx, "lan-printer-report", func(ctx context.Context, mqMsg queue.MqMsg) error {
        // 解析消息
        var message map[string]interface{}
        if err := json.Unmarshal(mqMsg.Body, &message); err != nil {
            return err
        }
        
        // 处理打印机信息
        companyUuid := message["company_uuid"].(uint64)
        printers := message["printers"].([]interface{})
        
        // 业务逻辑处理
        // ...
        
        return nil
    })
}
```

## 🔍 日志记录

### 接收日志

```
[INFO] 收到LAN打印机上报 device_id=cashier_001 count=2
```

### 发布成功日志

```
[INFO] LAN 打印机上报消息已发布到 MQ 
       company_uuid=8609817471094784 
       device_id=cashier_001 
       printer_count=2 
       topic=lan-printer-report
```

### 发布失败日志

```
[ERROR] 发布 LAN 打印机上报消息失败 
        company_uuid=8609817471094784 
        device_id=cashier_001 
        printer_count=2 
        error=连接队列失败
```

## 🎯 使用场景

### 场景 1：打印机自动发现

收银机启动后自动扫描局域网内的打印机，并上报到服务端：

```javascript
// 客户端代码示例
const printers = await scanLanPrinters();
ws.send(JSON.stringify({
  type: 'lan_print_report',
  data: printers
}));
```

### 场景 2：打印机状态监控

定期上报打印机状态，用于监控打印机是否在线：

```javascript
// 每 5 分钟上报一次
setInterval(async () => {
  const printers = await scanLanPrinters();
  ws.send(JSON.stringify({
    type: 'lan_print_report',
    data: printers
  }));
}, 5 * 60 * 1000);
```

### 场景 3：打印机配置管理

后台服务消费 MQ 消息，自动更新打印机配置：

```go
func handlePrinterReport(ctx context.Context, message map[string]interface{}) error {
    companyUuid := message["company_uuid"].(uint64)
    deviceId := message["device_id"].(string)
    printers := message["printers"].([]interface{})
    
    // 更新数据库中的打印机配置
    for _, printer := range printers {
        p := printer.(map[string]interface{})
        err := updatePrinterConfig(ctx, companyUuid, deviceId, p)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## ⚠️ 注意事项

### 1. 数据验证

- 客户端上报的数据必须符合格式要求
- IP 地址格式验证
- 端口号范围验证（1-65535）
- 状态值验证（0 或 1）

### 2. 性能考虑

- 避免频繁上报，建议间隔至少 1 分钟
- 单次上报的打印机数量建议不超过 50 台
- 大量打印机可分批上报

### 3. 错误处理

- MQ 发布失败不影响 WebSocket 连接
- 发布失败会记录错误日志
- 可以通过日志监控发布失败率

### 4. 安全性

- 只有认证通过的收银机才能上报
- 验证 `company_uuid` 和 `device_id` 的合法性
- 防止恶意上报大量数据

## 📈 监控指标

### 关键指标

1. **上报频率**
   - 每分钟上报次数
   - 每小时上报次数

2. **打印机数量**
   - 平均每次上报的打印机数量
   - 单个公司的打印机总数

3. **发布成功率**
   - MQ 发布成功次数 / 总上报次数
   - 发布失败次数和原因

4. **消息延迟**
   - 从接收到发布的时间
   - MQ 消息堆积情况

### 监控告警

- 发布失败率 > 5% 时告警
- 单次上报打印机数量 > 100 时告警
- 消息堆积 > 1000 时告警

## 🔗 相关文档

- [WebSocket 连接管理](./connection.md)
- [消息推送机制](./message.md)
- [消息队列配置](../../ttpos-message/features/queue.md)

## 📝 更新日志

### v1.0.0 (2025-11-15)
- ✅ 实现 LAN 打印机上报功能
- ✅ 集成消息队列发布
- ✅ 添加完整日志记录

