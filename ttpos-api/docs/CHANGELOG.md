# 更新日志

## [v2.0.0] - 2025-11-15

### 重大变更 🎉
- 🔄 **重构项目结构**：按模块划分，提升代码可读性和可维护性
  - 创建 `common/` 目录存放所有模块共享的基础组件
  - 创建 `ttpos-message/` 目录存放 ttpos-message 服务相关代码
  - 创建 `ttpos-websocket/` 目录存放 ttpos-websocket 服务相关代码

### 新增
- ✨ 模块化目录结构
  - `common/message/` - 基础消息结构体和错误定义
  - `common/constant/` - 通用常量定义
  - `common/util/` - 工具函数
  - `ttpos-message/message/` - 消息服务消息结构体
  - `ttpos-message/constant/` - 消息服务常量
  - `ttpos-websocket/message/` - WebSocket 消息结构体
  - `ttpos-websocket/constant/` - WebSocket 常量

### 改进
- 📦 更清晰的模块划分，便于理解和维护
- 📖 更新所有文档以反映新的结构
- 🎯 每个模块有独立的 constant 包，避免混淆

### 迁移指南
旧的导入方式：
```go
import "ttpos-api/message"
import "ttpos-api/constant"
```

新的导入方式：
```go
// ttpos-message 模块
import "ttpos-api/ttpos-message/message"
import "ttpos-api/ttpos-message/constant"

// ttpos-websocket 模块
import "ttpos-api/ttpos-websocket/message"
import "ttpos-api/ttpos-websocket/constant"

// 通用工具
import "ttpos-api/common/message"  // 基础消息
import "ttpos-api/util"             // 工具函数
```

## [v1.1.0] - 2025-11-14

### 新增
- ✨ WebSocket 相关消息结构体
  - `WebSocketMessage` - WebSocket 基础消息结构体
  - `OrderUpdateMessage` - 订单更新 WebSocket 消息
  - `DeskStatusMessage` - 桌台状态 WebSocket 消息
  - `PrinterNotifyMessage` - 打印机通知 WebSocket 消息
  - `KitchenOrderMessage` - 厨房订单 WebSocket 消息
  - `CallWaiterMessage` - 呼叫服务员 WebSocket 消息
  - `SystemNotifyMessage` - 系统通知 WebSocket 消息
  - `OnlineStatusMessage` - 在线状态 WebSocket 消息
- ✨ WebSocket 相关常量定义
  - 动作类型常量（subscribe/unsubscribe/notify 等）
  - 呼叫类型常量（waiter/manager/checkout 等）
  - 通知类型常量（info/warning/error/success）
  - 厨房类型常量（hot/cold/drink/dessert）
  - 打印任务状态常量
  - 用户类型常量
- ✨ WebSocket 相关 Topic（7个）
- ✨ WebSocket 使用示例代码

## [v1.0.0] - 2025-11-14

### 新增
- ✨ 初始版本发布
- ✨ 基础消息结构体 `BaseMessage`
- ✨ ttpos-message 服务相关消息结构体
  - `MessageSendMessage` - 消息发送队列消息
  - `MessageRetryMessage` - 消息重试队列消息
  - `MessageStatusChangeMessage` - 消息状态变更通知消息
- ✨ 常量定义
  - Topic 常量（消息队列主题）
  - 消息类型常量（email/sms）
  - 状态常量（消息状态、订单状态、会员状态、桌台状态）
- ✨ 工具函数
  - JSON 序列化/反序列化工具
  - 消息验证工具
  - 辅助工具函数
- ✨ 完整的使用文档和示例代码
- ✨ Makefile 构建工具

### 文档
- 📝 README.md - 项目概述和快速开始
- 📝 USAGE.md - 详细使用指南
- 📝 examples/ - 示例代码
- 📝 CHANGELOG.md - 更新日志

### 开发工具
- 🔧 go.mod - Go 模块定义
- 🔧 Makefile - 构建工具
- 🔧 .gitignore - Git 忽略文件配置

---

## 版本说明

版本号格式：`vMAJOR.MINOR.PATCH`

- **MAJOR**: 不兼容的 API 修改
- **MINOR**: 向下兼容的功能性新增
- **PATCH**: 向下兼容的问题修正

## 贡献者

- TTPOS Team

## 许可证

Copyright © 2025 TTPOS Team. All rights reserved.

