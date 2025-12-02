# ttpos-websocket 使用示例

本目录包含 `ttpos-websocket` 模块的使用示例代码。

## 📋 示例列表

### 1. websocket_example.go

WebSocket 消息的基本使用示例，包含以下消息类型：

- **订单更新消息** (`OrderUpdateMessage`)
- **桌台状态消息** (`DeskStatusMessage`)
- **打印机通知消息** (`PrinterNotifyMessage`)
- **厨房订单消息** (`KitchenOrderMessage`)
- **呼叫服务员消息** (`CallWaiterMessage`)
- **系统通知消息** (`SystemNotifyMessage`)
- **在线状态消息** (`OnlineStatusMessage`)

### 2. lan_printer_example.go

LAN 打印机上报消息的使用示例，包含：

- **发布消息示例** (`LanPrinterReportExample`)
  - 创建打印机列表
  - 构建上报消息
  - 验证消息
  - 发布到消息队列

- **消费消息示例** (`LanPrinterReportConsumerExample`)
  - 从队列接收消息
  - 反序列化消息
  - 验证消息
  - 处理业务逻辑

- **验证示例** (`ValidateLanPrinterExample`)
  - 正常消息验证
  - 缺少必填字段验证
  - 字段格式验证
  - 字段范围验证

## 🚀 运行示例

### 方式 1：直接运行

```bash
cd /home/coder/workspaces/ttpos-server-go/ttpos-api/ttpos-websocket/examples

# 运行 WebSocket 示例
go run websocket_example.go

# 运行 LAN 打印机示例
go run lan_printer_example.go
```

### 方式 2：在您的项目中引用

```go
package main

import (
    "ttpos-api/ttpos-websocket/message"
    "ttpos-api/ttpos-websocket/constant"
)

func main() {
    // 创建 LAN 打印机上报消息
    printers := []message.LanPrinterInfo{
        {
            IP:     "192.168.1.100",
            Port:   9100,
            Status: 1,
            Remark: "前台打印机",
        },
    }
    
    msg := message.NewLanPrinterReportMessage(
        8609817471094784, // companyUUID
        1234567890,       // staffUUID
        "cashier_001",    // deviceID
        "cashier",        // sourceClient
        printers,
    )
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        panic(err)
    }
    
    // 发布到消息队列
    // queue.Publish(constant.TopicLanPrinterReport, msg)
}
```

## 📚 相关文档

- [ttpos-websocket 消息定义](../message/websocket.go)
- [ttpos-websocket 常量定义](../constant/constant.go)
- [ttpos-api 主文档](../../README.md)
- [LAN 打印机上报功能文档](../../../../ttpos-bmp/docs/app/ttpos-websocket/features/lan-printer-report.md)

## 💡 最佳实践

### 1. 消息验证

始终在发布消息前调用 `Validate()` 方法：

```go
msg := message.NewLanPrinterReportMessage(...)
if err := msg.Validate(); err != nil {
    // 处理验证错误
    return err
}
```

### 2. 错误处理

使用 `ttpos-api/common/message` 包中定义的错误：

```go
import "ttpos-api/common/message"

if err := msg.Validate(); err != nil {
    if err == message.ErrCompanyUUIDRequired {
        // 处理公司UUID缺失错误
    }
}
```

### 3. 消息序列化

使用内置的 `ToJSON()` 和 `FromJSON()` 方法：

```go
// 序列化
jsonData, err := msg.ToJSON()

// 反序列化
var msg message.LanPrinterReportMessage
err := msg.FromJSON(jsonData)
```

### 4. Topic 常量

使用 `constant` 包中定义的 Topic 常量：

```go
import "ttpos-api/ttpos-websocket/constant"

// 使用常量而不是硬编码字符串
queue.Publish(constant.TopicLanPrinterReport, msg)
```

## ⚠️ 注意事项

1. **字段验证**：所有必填字段必须提供，否则 `Validate()` 会返回错误
2. **IP 格式**：打印机 IP 地址必须是有效的 IPv4 地址
3. **端口范围**：打印机端口必须在 1-65535 之间
4. **状态值**：打印机状态只能是 0（离线）或 1（在线）
5. **消息大小**：避免单次上报过多打印机（建议不超过 50 台）

## 🔍 调试技巧

### 1. 启用详细日志

```go
import "log"

msg := message.NewLanPrinterReportMessage(...)
jsonData, _ := msg.ToJSON()
log.Printf("消息内容: %s", string(jsonData))
```

### 2. 验证消息格式

```go
if err := msg.Validate(); err != nil {
    log.Printf("验证失败: %v", err)
    // 检查具体的错误类型
}
```

### 3. 测试消息序列化

```go
// 序列化
jsonData, _ := msg.ToJSON()

// 反序列化
var newMsg message.LanPrinterReportMessage
newMsg.FromJSON(jsonData)

// 比较两个消息是否一致
```

## 📞 获取帮助

如果您在使用过程中遇到问题，请：

1. 查看 [ttpos-api 主文档](../../README.md)
2. 查看 [消息定义源码](../message/websocket.go)
3. 运行示例代码进行参考
4. 联系项目维护者

