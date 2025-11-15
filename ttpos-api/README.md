# TTPOS API - 全局消息结构体包

## 📋 项目简介

`ttpos-api` 是 TTPOS 系统的全局消息结构体定义包，用于统一管理各个服务之间的消息队列（MQ）消息结构体，确保消息格式的一致性和类型安全。

## 🎯 设计目标

1. **统一管理**：集中管理所有服务间的消息结构体定义
2. **类型安全**：通过 Go 的类型系统确保消息格式正确
3. **易于维护**：单一职责，便于版本管理和升级
4. **跨服务共享**：可被 main、ttpos-bmp、websocket 等多个服务引用

## 📦 包结构

```
ttpos-api/
├── go.mod                          # Go 模块定义
├── README.md                       # 项目文档
├── docs/                            # 📚 文档目录
│   ├── README.md                  # 文档导航中心
│   ├── CHANGELOG.md               # 版本更新记录
│   ├── guide/                     # 📘 使用指南
│   │   ├── USAGE.md              # 详细使用指南
│   │   └── INTEGRATION.md        # 集成指南
│   └── reference/                 # 📚 参考文档
│       ├── MODULES.md            # 模块结构说明
│       ├── ERRORS.md             # 错误定义规范
│       └── WEBSOCKET.md          # WebSocket 详细文档
├── common/                         # 通用组件（所有模块共享）
│   ├── message/                    # 基础消息定义
│   │   ├── base.go                # 基础消息结构体
│   │   └── errors.go              # 公共错误定义
│   ├── constant/                   # 通用常量
│   │   ├── status.go              # 状态常量
│   │   └── topic.go               # Topic 常量汇总
│   └── util/                       # 工具函数
│       ├── json.go                # JSON 序列化工具
│       ├── validator.go           # 消息验证工具
│       └── helper.go              # 辅助工具函数
├── ttpos-message/                  # ttpos-message 服务模块
│   ├── message/                    # 消息结构体
│   │   └── message.go             # 消息发送、重试、状态变更
│   ├── constant/                   # 常量和错误定义
│   │   ├── constant.go            # Topic、消息类型、消息状态
│   │   └── errors.go              # 模块特定错误
│   └── examples/                   # 使用示例
│       └── message_send_example.go # 消息发送示例
└── ttpos-websocket/                # ttpos-websocket 服务模块
    ├── message/                    # 消息结构体
    │   └── websocket.go           # WebSocket 消息类型（订单、桌台、打印机等）
    ├── constant/                   # 常量和错误定义
    │   ├── constant.go            # Topic、动作类型、业务常量
    │   └── errors.go              # 模块特定错误
    └── examples/                   # 使用示例
        ├── README.md              # 示例说明文档
        ├── websocket_example.go   # WebSocket 消息示例
        └── lan_printer_example.go # LAN 打印机上报示例
```

## 🚀 快速开始

### 安装

在您的服务中引入 ttpos-api 包：

```bash
# 在 main、ttpos-bmp 或其他服务中
go get ttpos-api@latest
```

或在 go.mod 中使用本地路径：

```go
replace ttpos-api => ../ttpos-api
```

### 基本使用

#### 1. 使用 ttpos-message 模块

```go
package main

import (
    "ttpos-api/ttpos-message/message"
    "ttpos-api/ttpos-message/constant"
)

func publishMessage() {
    // 创建消息
    msg := message.NewMessageSendMessage("msg-uuid-123", constant.MessageTypeEmail)
    msg.WithCompanyUUID("company-uuid-456")
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return
    }
    
    // 序列化为 JSON
    data, _ := msg.ToJSON()
    
    // 发送到 MQ
    mq.Publish(constant.TopicMessageSend, data)
}
```

#### 2. 使用 ttpos-websocket 模块

**示例 1：订单更新消息**

```go
package main

import (
    "ttpos-api/ttpos-websocket/message"
    "ttpos-api/ttpos-websocket/constant"
)

func broadcastOrderUpdate() {
    // 创建订单更新消息
    msg := message.NewOrderUpdateMessage(
        "client-123",
        "order-uuid-456",
        1, // 订单状态：已支付
    )
    msg.OrderAmount = 99.99
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return
    }
    
    // 发送到 WebSocket
    data, _ := msg.ToJSON()
    ws.Broadcast(constant.TopicWebSocketOrderUpdate, data)
}
```

**示例 2：LAN 打印机上报消息**

```go
package main

import (
    "ttpos-api/ttpos-websocket/message"
    "ttpos-api/ttpos-websocket/constant"
)

func reportLanPrinters() {
    // 创建打印机列表
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
    }
    
    // 创建 LAN 打印机上报消息
    msg := message.NewLanPrinterReportMessage(
        8609817471094784, // companyUUID
        1234567890,       // staffUUID
        "cashier_001",    // deviceID
        "cashier",        // sourceClient
        printers,
    )
    msg.ReportTime = "2025-11-15 16:45:30"
    msg.Timestamp = 1731689130
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return
    }
    
    // 发布到消息队列
    data, _ := msg.ToJSON()
    mq.Publish(constant.TopicLanPrinterReport, data)
}
```

## 📝 消息结构体规范

### 基础消息结构

所有消息必须包含 `BaseMessage` 字段：

```go
type BaseMessage struct {
    MessageID   string `json:"message_id"`   // 消息唯一标识
    Topic       string `json:"topic"`        // 消息主题
    Timestamp   int64  `json:"timestamp"`    // 消息时间戳
    Version     string `json:"version"`      // 消息版本
    CompanyUUID string `json:"company_uuid"` // 公司UUID（可选）
}
```

### 消息命名规范

- 消息结构体命名：`{业务模块}{动作}Message`
- 例如：`MessageSendMessage`、`OrderCreatedMessage`、`MemberUpdatedMessage`

### 字段命名规范

- 使用大驼峰命名法（PascalCase）
- JSON 标签使用蛇形命名法（snake_case）
- 添加清晰的中文注释

## 🔧 开发指南

### 添加新的消息类型

1. 在 `message/` 目录下创建对应的文件
2. 定义消息结构体，继承 `BaseMessage`
3. 在 `constant/` 目录下添加相关常量
4. 更新文档说明

示例：

```go
// message/order.go
package message

// OrderCreatedMessage 订单创建消息
type OrderCreatedMessage struct {
    BaseMessage
    OrderUUID   string  `json:"order_uuid"`   // 订单UUID
    OrderAmount float64 `json:"order_amount"` // 订单金额
    CustomerID  string  `json:"customer_id"`  // 客户ID
}
```

### 版本管理

- 使用语义化版本号（Semantic Versioning）
- 主版本号变更：不兼容的 API 修改
- 次版本号变更：向下兼容的功能性新增
- 修订号变更：向下兼容的问题修正

## 🔒 最佳实践

1. **消息不可变性**：一旦发布，消息结构体不应修改，应创建新版本
2. **向后兼容**：新增字段使用指针类型，保持可选
3. **文档完整**：每个消息结构体都要有清晰的注释
4. **类型安全**：使用常量而不是字符串字面量
5. **验证机制**：提供消息验证方法，确保数据完整性

## 📚 文档导航

### 📘 使用指南 (docs/guide/)

实用的操作指南和教程：

- **[使用指南](docs/guide/USAGE.md)** - 详细的使用说明和最佳实践
- **[集成指南](docs/guide/INTEGRATION.md)** - 各服务集成步骤和注意事项

### 📚 参考文档 (docs/reference/)

详细的技术参考资料：

- **[模块说明](docs/reference/MODULES.md)** - 模块结构和职责划分（推荐先看）
- **[错误规范](docs/reference/ERRORS.md)** - 错误定义和使用规范
- **[WebSocket 文档](docs/reference/WEBSOCKET.md)** - WebSocket 消息类型和使用详解

### 📝 其他文档

- **[变更日志](docs/CHANGELOG.md)** - 版本更新记录
- **[文档中心](docs/README.md)** - 完整的文档导航

### 💡 示例代码

- **[ttpos-websocket 示例](ttpos-websocket/examples/)** - WebSocket 消息示例
  - [LAN 打印机上报](ttpos-websocket/examples/lan_printer_example.go)
  - [WebSocket 消息](ttpos-websocket/examples/websocket_example.go)
- **[ttpos-message 示例](ttpos-message/examples/)** - 消息服务示例
  - [消息发送](ttpos-message/examples/message_send_example.go)

## 🤝 贡献指南

1. 创建新分支进行开发
2. 遵循代码规范和命名约定
3. 添加必要的测试用例
4. 更新相关文档
5. 提交 Pull Request

## 📄 许可证

Copyright © 2025 TTPOS Team. All rights reserved.

