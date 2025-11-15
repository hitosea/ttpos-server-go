# WebSocket 消息结构体使用指南

本文档详细说明如何使用 `ttpos-api` 中的 WebSocket 相关消息结构体。

## 📋 目录

- [消息类型概览](#消息类型概览)
- [基础使用](#基础使用)
- [消息类型详解](#消息类型详解)
- [常量定义](#常量定义)
- [使用场景](#使用场景)
- [最佳实践](#最佳实践)

## 消息类型概览

ttpos-api 提供了 8 种 WebSocket 消息类型：

| 消息类型 | 结构体 | Topic | 用途 |
|---------|--------|-------|------|
| 订单更新 | `OrderUpdateMessage` | `websocket.order.update` | 实时推送订单状态变更 |
| 桌台状态 | `DeskStatusMessage` | `websocket.desk.status` | 实时推送桌台状态变更 |
| 打印机通知 | `PrinterNotifyMessage` | `websocket.printer.notify` | 实时推送打印任务状态 |
| 厨房订单 | `KitchenOrderMessage` | `websocket.kitchen.order` | 实时推送厨房订单 |
| 呼叫服务员 | `CallWaiterMessage` | `websocket.call.waiter` | 实时推送呼叫服务员请求 |
| 系统通知 | `SystemNotifyMessage` | `websocket.system.notify` | 实时推送系统通知 |
| 在线状态 | `OnlineStatusMessage` | `websocket.online.status` | 实时推送用户在线状态 |
| 基础消息 | `WebSocketMessage` | 自定义 | WebSocket 基础消息（可扩展） |

## 基础使用

### 导入包

```go
import (
    "ttpos-api/message"
    "ttpos-api/constant"
)
```

### 创建和发送消息

```go
// 创建订单更新消息
msg := message.NewOrderUpdateMessage(
    "client-123",           // 客户端ID
    "order-uuid-456",       // 订单UUID
    constant.OrderStatusPaid, // 订单状态
)

// 设置可选字段
msg.OrderAmount = 99.99
msg.UpdateTime = time.Now().Unix()

// 验证消息
if err := msg.Validate(); err != nil {
    return err
}

// 序列化为 JSON
data, _ := msg.ToJSON()

// 发送到 WebSocket
ws.Send(msg.ClientID, data)
```

## 消息类型详解

### 1. 订单更新消息 (OrderUpdateMessage)

**用途**: 实时推送订单状态变更

**字段说明**:
- `OrderUUID`: 订单UUID（必填）
- `OrderStatus`: 订单状态（必填）
- `OrderAmount`: 订单金额
- `UpdateTime`: 更新时间

**使用示例**:

```go
msg := message.NewOrderUpdateMessage(
    "client-123",
    "order-uuid-456",
    constant.OrderStatusPaid,
)
msg.OrderAmount = 99.99
msg.UpdateTime = util.GetCurrentTimestamp()
```

**推送场景**:
- 订单支付成功
- 订单状态变更
- 订单取消
- 订单完成

### 2. 桌台状态消息 (DeskStatusMessage)

**用途**: 实时推送桌台状态变更

**字段说明**:
- `DeskUUID`: 桌台UUID（必填）
- `DeskNumber`: 桌台号（必填）
- `DeskStatus`: 桌台状态（必填）
- `UpdateTime`: 更新时间

**使用示例**:

```go
msg := message.NewDeskStatusMessage(
    "client-123",
    "desk-uuid-456",
    "A01",
    constant.DeskStatusOccupied,
)
msg.UpdateTime = util.GetCurrentTimestamp()
msg.RoomID = "floor-1" // 推送到指定楼层
```

**推送场景**:
- 桌台开台
- 桌台关台
- 桌台预订
- 桌台清理

### 3. 打印机通知消息 (PrinterNotifyMessage)

**用途**: 实时推送打印任务状态

**字段说明**:
- `PrinterUUID`: 打印机UUID（必填）
- `PrinterName`: 打印机名称
- `TaskID`: 打印任务ID（必填）
- `TaskStatus`: 任务状态（必填）
- `ErrorMsg`: 错误信息

**使用示例**:

```go
msg := message.NewPrinterNotifyMessage(
    "client-123",
    "printer-uuid-456",
    "task-789",
    constant.PrintTaskStatusSuccess,
)
msg.PrinterName = "厨房打印机-1"
```

**推送场景**:
- 打印任务开始
- 打印任务完成
- 打印任务失败

### 4. 厨房订单消息 (KitchenOrderMessage)

**用途**: 实时推送厨房订单

**字段说明**:
- `OrderUUID`: 订单UUID（必填）
- `DeskNumber`: 桌台号
- `OrderItems`: 订单项（JSON 数组）
- `OrderTime`: 下单时间
- `KitchenType`: 厨房类型（必填）
- `Priority`: 优先级（1-5）
- `SpecialNotes`: 特殊备注

**使用示例**:

```go
msg := message.NewKitchenOrderMessage(
    "client-123",
    "order-uuid-456",
    "A01",
    constant.KitchenTypeHot,
)
msg.OrderItems = []string{
    `{"name":"宫保鸡丁","quantity":1}`,
    `{"name":"麻婆豆腐","quantity":2}`,
}
msg.Priority = 5 // 高优先级
msg.SpecialNotes = "不要辣"
msg.RoomID = "kitchen-hot" // 推送到热厨房间
```

**推送场景**:
- 新订单下单
- 订单催单
- 订单修改

### 5. 呼叫服务员消息 (CallWaiterMessage)

**用途**: 实时推送呼叫服务员请求

**字段说明**:
- `DeskUUID`: 桌台UUID（必填）
- `DeskNumber`: 桌台号
- `CallType`: 呼叫类型（必填）
- `CallTime`: 呼叫时间
- `IsUrgent`: 是否紧急
- `Remark`: 备注

**使用示例**:

```go
msg := message.NewCallWaiterMessage(
    "client-123",
    "desk-uuid-456",
    "A01",
    constant.CallTypeWaiter,
)
msg.IsUrgent = true
msg.Remark = "需要加水"
msg.RoomID = "waiters" // 推送到服务员房间
```

**推送场景**:
- 顾客呼叫服务员
- 顾客呼叫经理
- 顾客请求结账
- 顾客请求清理

### 6. 系统通知消息 (SystemNotifyMessage)

**用途**: 实时推送系统通知

**字段说明**:
- `NotifyType`: 通知类型（必填）
- `Title`: 通知标题（必填）
- `Content`: 通知内容
- `NotifyTime`: 通知时间
- `AutoClose`: 是否自动关闭
- `CloseDelay`: 关闭延迟（秒）
- `ActionURL`: 操作链接
- `ActionLabel`: 操作按钮文本

**使用示例**:

```go
msg := message.NewSystemNotifyMessage(
    "client-123",
    constant.NotifyTypeSuccess,
    "订单支付成功",
    "订单 #12345 已成功支付，金额：¥99.99",
)
msg.AutoClose = true
msg.CloseDelay = 3
msg.ActionURL = "/orders/12345"
msg.ActionLabel = "查看订单"
```

**推送场景**:
- 操作成功提示
- 错误警告
- 系统维护通知
- 权限变更通知

### 7. 在线状态消息 (OnlineStatusMessage)

**用途**: 实时推送用户在线状态

**字段说明**:
- `UserUUID`: 用户UUID（必填）
- `UserName`: 用户名
- `UserType`: 用户类型
- `IsOnline`: 是否在线
- `UpdateTime`: 更新时间

**使用示例**:

```go
msg := message.NewOnlineStatusMessage(
    "client-123",
    "user-uuid-456",
    "张三",
    constant.UserTypeStaff,
    true,
)
msg.UpdateTime = util.GetCurrentTimestamp()
msg.RoomID = "staff-room" // 广播到员工房间
```

**推送场景**:
- 用户上线
- 用户下线
- 用户状态变更

## 常量定义

### 动作类型 (Action)

```go
constant.ActionSubscribe      // 订阅
constant.ActionUnsubscribe    // 取消订阅
constant.ActionNotify         // 通知
constant.ActionOrderUpdate    // 订单更新
constant.ActionDeskStatus     // 桌台状态
constant.ActionPrinterNotify  // 打印机通知
constant.ActionKitchenOrder   // 厨房订单
constant.ActionCallWaiter     // 呼叫服务员
constant.ActionSystemNotify   // 系统通知
constant.ActionOnlineStatus   // 在线状态
constant.ActionHeartbeat      // 心跳
constant.ActionPing           // Ping
constant.ActionPong           // Pong
```

### 呼叫类型 (CallType)

```go
constant.CallTypeWaiter    // 呼叫服务员
constant.CallTypeManager   // 呼叫经理
constant.CallTypeCheckout  // 呼叫结账
constant.CallTypeClean     // 呼叫清理
```

### 通知类型 (NotifyType)

```go
constant.NotifyTypeInfo     // 信息通知
constant.NotifyTypeWarning  // 警告通知
constant.NotifyTypeError    // 错误通知
constant.NotifyTypeSuccess  // 成功通知
```

### 厨房类型 (KitchenType)

```go
constant.KitchenTypeHot      // 热厨
constant.KitchenTypeCold     // 冷厨
constant.KitchenTypeDrink    // 饮品
constant.KitchenTypeDessert  // 甜品
```

### 打印任务状态 (PrintTaskStatus)

```go
constant.PrintTaskStatusPending   // 待打印
constant.PrintTaskStatusPrinting  // 打印中
constant.PrintTaskStatusSuccess   // 打印成功
constant.PrintTaskStatusFailed    // 打印失败
```

### 用户类型 (UserType)

```go
constant.UserTypeStaff     // 员工
constant.UserTypeCustomer  // 顾客
constant.UserTypeManager   // 管理员
constant.UserTypeKitchen   // 厨房
```

## 使用场景

### 场景 1: 订单支付成功后推送

```go
// 订单支付成功后
func OnOrderPaid(orderUUID string, amount float64) {
    // 创建订单更新消息
    msg := message.NewOrderUpdateMessage(
        "", // 不指定客户端ID，广播给所有人
        orderUUID,
        constant.OrderStatusPaid,
    )
    msg.OrderAmount = amount
    msg.UpdateTime = util.GetCurrentTimestamp()
    
    // 广播到所有连接的客户端
    ws.BroadcastToAll(msg)
}
```

### 场景 2: 桌台开台后推送

```go
// 桌台开台后
func OnDeskOpened(deskUUID, deskNumber string) {
    // 创建桌台状态消息
    msg := message.NewDeskStatusMessage(
        "",
        deskUUID,
        deskNumber,
        constant.DeskStatusOccupied,
    )
    msg.UpdateTime = util.GetCurrentTimestamp()
    msg.RoomID = "floor-1" // 推送到指定楼层
    
    // 广播到指定房间
    ws.BroadcastToRoom(msg.RoomID, msg)
}
```

### 场景 3: 厨房收到新订单

```go
// 新订单下单后
func OnNewOrder(orderUUID, deskNumber string, items []OrderItem) {
    // 创建厨房订单消息
    msg := message.NewKitchenOrderMessage(
        "",
        orderUUID,
        deskNumber,
        constant.KitchenTypeHot,
    )
    
    // 转换订单项为 JSON
    itemsJSON := make([]string, len(items))
    for i, item := range items {
        data, _ := json.Marshal(item)
        itemsJSON[i] = string(data)
    }
    msg.OrderItems = itemsJSON
    msg.OrderTime = util.GetCurrentTimestamp()
    msg.Priority = 3
    msg.RoomID = "kitchen-hot"
    
    // 推送到热厨房间
    ws.BroadcastToRoom(msg.RoomID, msg)
}
```

### 场景 4: 顾客呼叫服务员

```go
// 顾客呼叫服务员
func OnCallWaiter(deskUUID, deskNumber string, isUrgent bool) {
    // 创建呼叫服务员消息
    msg := message.NewCallWaiterMessage(
        "",
        deskUUID,
        deskNumber,
        constant.CallTypeWaiter,
    )
    msg.CallTime = util.GetCurrentTimestamp()
    msg.IsUrgent = isUrgent
    msg.RoomID = "waiters"
    
    // 推送到服务员房间
    ws.BroadcastToRoom(msg.RoomID, msg)
}
```

## 最佳实践

### 1. 使用房间 (RoomID) 进行分组推送

```go
// ✅ 推荐：使用房间进行分组推送
msg.RoomID = "kitchen-hot"
ws.BroadcastToRoom(msg.RoomID, msg)

// ❌ 不推荐：广播给所有人
ws.BroadcastToAll(msg)
```

### 2. 使用目标用户列表进行定向推送

```go
// ✅ 推荐：定向推送给指定用户
msg.TargetUsers = []string{"user-1", "user-2", "user-3"}
ws.SendToUsers(msg.TargetUsers, msg)
```

### 3. 设置消息优先级

```go
// ✅ 推荐：为紧急消息设置高优先级
msg.Priority = 5 // 高优先级
msg.IsUrgent = true
```

### 4. 添加链路追踪ID

```go
// ✅ 推荐：添加链路追踪ID
msg.WithTraceID(ctx.GetTraceID())
```

### 5. 验证消息

```go
// ✅ 推荐：发送前验证消息
if err := msg.Validate(); err != nil {
    log.Error("消息验证失败", err)
    return err
}
```

### 6. 错误处理

```go
// ✅ 推荐：完整的错误处理
data, err := msg.ToJSON()
if err != nil {
    log.Error("消息序列化失败", err)
    return err
}

if err := ws.Send(clientID, data); err != nil {
    log.Error("消息发送失败", err)
    return err
}
```

## 相关文档

- [README.md](README.md) - 项目概述
- [USAGE.md](USAGE.md) - 详细使用指南
- [INTEGRATION.md](INTEGRATION.md) - 集成指南
- [examples/websocket_example.go](examples/websocket_example.go) - 完整示例代码


