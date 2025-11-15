# 错误定义规范

本文档说明 `ttpos-api` 项目中错误的组织和使用方式。

## 📋 错误分层

### 1. 公共错误 (`common/message/errors.go`)

定义所有模块共享的基础错误：

```go
package message

var (
    // 基础消息错误（所有模块共享）
    
    // ErrMessageIDRequired 消息ID不能为空
    ErrMessageIDRequired = errors.New("消息ID不能为空")
    
    // ErrTopicRequired 消息主题不能为空
    ErrTopicRequired = errors.New("消息主题不能为空")
    
    // ErrTimestampRequired 消息时间戳不能为空
    ErrTimestampRequired = errors.New("消息时间戳不能为空")
    
    // ErrInvalidJSON JSON格式无效
    ErrInvalidJSON = errors.New("JSON格式无效")
)

// NewValidationError 创建验证错误
func NewValidationError(field, message string, args ...interface{}) error {
    // ...
}
```

### 2. 模块特定错误

每个模块在其 `constant/errors.go` 中定义自己的错误。

#### ttpos-websocket 模块 (`ttpos-websocket/constant/errors.go`)

```go
package constant

var (
    // WebSocket 基础错误
    ErrActionRequired = errors.New("动作类型不能为空")
    ErrClientIDRequired = errors.New("客户端ID不能为空")
    
    // 订单相关错误
    ErrOrderUUIDRequired = errors.New("订单UUID不能为空")
    
    // 桌台相关错误
    ErrDeskUUIDRequired = errors.New("桌台UUID不能为空")
    
    // 打印机相关错误
    ErrPrinterUUIDRequired = errors.New("打印机UUID不能为空")
    ErrTaskIDRequired = errors.New("任务ID不能为空")
    
    // LAN 打印机上报相关错误
    ErrCompanyUUIDRequired = errors.New("公司UUID不能为空")
    ErrDeviceIDRequired = errors.New("设备ID不能为空")
    ErrSourceClientRequired = errors.New("来源客户端不能为空")
    ErrPrintersRequired = errors.New("打印机列表不能为空")
    
    // ... 其他错误
)
```

#### ttpos-message 模块 (`ttpos-message/constant/errors.go`)

```go
package constant

var (
    // 消息发送相关错误
    ErrMessageUUIDRequired = errors.New("消息UUID不能为空")
    ErrMessageTypeRequired = errors.New("消息类型不能为空")
    ErrInvalidMessageType = errors.New("无效的消息类型")
    ErrRecipientRequired = errors.New("接收人不能为空")
)
```

## 🎯 使用方式

### 1. 在消息结构体中使用模块错误

```go
// ttpos-websocket/message/websocket.go
package message

import (
    commonMsg "ttpos-api/common/message"
    "ttpos-api/ttpos-websocket/constant"
)

func (m *LanPrinterReportMessage) Validate() error {
    // 使用公共错误验证基础字段
    if err := m.BaseMessage.Validate(); err != nil {
        return err
    }
    
    // 使用模块特定错误验证业务字段
    if m.CompanyUUID == 0 {
        return constant.ErrCompanyUUIDRequired
    }
    if m.DeviceID == "" {
        return constant.ErrDeviceIDRequired
    }
    
    // 使用公共验证错误函数
    for i, printer := range m.Printers {
        if printer.IP == "" {
            return commonMsg.NewValidationError("printers[%d].ip", "打印机IP地址不能为空", i)
        }
    }
    
    return nil
}
```

### 2. 在服务中使用错误

```go
package main

import (
    "ttpos-api/ttpos-websocket/message"
    "ttpos-api/ttpos-websocket/constant"
)

func processMessage(msg *message.LanPrinterReportMessage) error {
    if err := msg.Validate(); err != nil {
        // 判断具体的错误类型
        if err == constant.ErrCompanyUUIDRequired {
            // 处理公司UUID缺失错误
            return fmt.Errorf("无法处理消息: %w", err)
        }
        return err
    }
    
    // 处理消息
    return nil
}
```

## 📝 命名规范

### 1. 错误变量命名

- 使用 `Err` 前缀
- 使用大驼峰命名法（PascalCase）
- 使用 `Required` 后缀表示必填字段错误
- 使用 `Invalid` 前缀表示格式错误

**示例**：
```go
ErrMessageIDRequired    // 必填字段错误
ErrInvalidJSON          // 格式错误
ErrActionRequired       // 必填字段错误
```

### 2. 错误消息

- 使用中文
- 简洁明了
- 描述问题而不是解决方案

**示例**：
```go
✅ 正确：errors.New("消息ID不能为空")
❌ 错误：errors.New("请提供消息ID")

✅ 正确：errors.New("无效的消息类型")
❌ 错误：errors.New("消息类型格式不正确，请检查")
```

## 🔍 错误分类

### 基础错误（common）

适用于所有模块的通用错误：
- 消息ID相关
- Topic相关
- 时间戳相关
- JSON格式相关
- 通用验证函数

### 模块错误（constant）

特定于某个模块的业务错误：
- **ttpos-websocket**：WebSocket连接、订单、桌台、打印机、厨房、呼叫服务员、系统通知等
- **ttpos-message**：消息发送、消息类型、接收人等

## 📊 错误使用统计

| 模块 | 错误数量 | 文件路径 |
|------|---------|----------|
| common | 4 | `common/message/errors.go` |
| ttpos-websocket | 16 | `ttpos-websocket/constant/errors.go` |
| ttpos-message | 4 | `ttpos-message/constant/errors.go` |

## ✅ 最佳实践

### 1. 错误定义位置

- ✅ 基础通用错误 → `common/message/errors.go`
- ✅ 模块特定错误 → `{module}/constant/errors.go`
- ❌ 不要在 `message/` 目录下定义错误

### 2. 错误引用

```go
// ✅ 正确：使用模块的 constant 包
import "ttpos-api/ttpos-websocket/constant"

if err == constant.ErrCompanyUUIDRequired {
    // ...
}

// ❌ 错误：不要在其他模块中定义相同的错误
var ErrCompanyUUIDRequired = errors.New("...")
```

### 3. 错误包装

```go
// ✅ 使用 fmt.Errorf 包装错误
if err := msg.Validate(); err != nil {
    return fmt.Errorf("验证消息失败: %w", err)
}

// ✅ 使用公共验证错误函数
return commonMsg.NewValidationError("field", "message", args...)
```

### 4. 错误处理

```go
// ✅ 判断具体的错误类型
if err == constant.ErrCompanyUUIDRequired {
    // 特定处理
}

// ✅ 使用 errors.Is 判断包装的错误
if errors.Is(err, constant.ErrCompanyUUIDRequired) {
    // 特定处理
}
```

## 🔄 迁移指南

如果需要添加新的错误：

1. **判断错误类型**
   - 是所有模块都会用到的？→ 添加到 `common/message/errors.go`
   - 只有特定模块使用？→ 添加到 `{module}/constant/errors.go`

2. **定义错误**
   ```go
   // 在对应的文件中添加
   // ErrXXXRequired XXX不能为空
   ErrXXXRequired = errors.New("XXX不能为空")
   ```

3. **使用错误**
   ```go
   // 导入对应的包
   import "ttpos-api/{module}/constant"
   
   // 使用错误
   if condition {
       return constant.ErrXXXRequired
   }
   ```

## 📚 相关文档

- [项目结构说明](./README.md)
- [ttpos-websocket 常量定义](./ttpos-websocket/constant/constant.go)
- [ttpos-message 常量定义](./ttpos-message/constant/constant.go)
- [公共错误定义](./common/message/errors.go)

