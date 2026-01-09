# LINE MAN 集成流程映射文档

> 基于 Grab 流程架构，标注差异点和跳过逻辑

---

## 一、总体架构对比

### Grab 架构（现有）

```
Grab Platform ←→ ttpos-bmp ←→ RocketMQ ←→ ttpos-server-go (Main)
     │                │
     │  双向交互       │  双向 gRPC + MQ
     │  (推送+拉取)    │
```

### LINE MAN 架构（映射）

```
LINE MAN Platform ←→ ttpos-bmp ←→ RocketMQ ←→ ttpos-server-go (Main)
     │                    │
     │  单向交互           │  复用现有通道
     │  (仅推送为主)       │
     │                    │
     │  ⚠️ 差异：无 GetMenu 拉取
     │  ⚠️ 差异：无订单操作 API
     │  🆕 新增：Trigger Sync Menu
     │  🆕 新增：门店开关控制
```

---

## 二、订单流程映射

### 2.1 Grab 订单流程（原版）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   Grab   │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │     │ 数据库   │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │     │ MySQL   │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. POST /submitOrder            │                │                │
     ├───────────────>│                │                │                │
     │                │ 2. 验证签名     │                │                │
     │                │ 3. 保存订单到 BMP               │                │
     │                ├────────────────────────────────────────────────>│
     │                │ 4. 发送 MQ      │                │                │
     │                ├───────────────>│                │                │
     │  5. 返回 200   │                │                │                │
     │<───────────────┤                │ 6. 消费消息    │                │
     │                │                ├───────────────>│                │
     │                │                │ 7. 调用 RPC    │                │
     │                │<───────────────────────────────┤                │
     │                │                │ 8. 转换+保存   │                │
     │                │                │                ├───────────────>│
```

### 2.2 LINE MAN 订单流程（映射版）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ LINE MAN │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │     │ 数据库   │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │     │ MySQL   │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. POST /place_order_notification               │                │
     │    ⚠️ 差异：接口名和数据格式不同                  │                │
     ├───────────────>│                │                │                │
     │                │                │                │                │
     │                │ 2. 验证 Token (OAuth2)          │                │
     │                │    ✅ 复用：认证逻辑相同         │                │
     │                │                │                │                │
     │                │ 3. HandlePlaceOrder             │                │
     │                │    🆕 新增：Lineman 订单解析     │                │
     │                │    保存订单到 BMP 数据库        │                │
     │                ├────────────────────────────────────────────────>│
     │                │                │                │                │
     │                │ 4. 发送 MQ 消息                 │                │
     │                │    Topic: takeout_lineman_order │                │
     │                │    ⚠️ 差异：新 Topic 或复用现有  │                │
     │                ├───────────────>│                │                │
     │                │                │                │                │
     │  5. 返回 200 OK│                │                │                │
     │<───────────────┤                │                │                │
     │                │                │                │                │
     │                │                │ 6. 消费消息    │                │
     │                │                ├───────────────>│                │
     │                │                │                │                │
     │                │                │ 7. HandlePushOrderState         │
     │                │                │    ✅ 复用：调用链相同           │
     │                │                │                │                │
     │                │                │ 8. 调用 RPC 获取订单详情         │
     │                │<───────────────────────────────┤ (gRPC)         │
     │                │                │                │                │
     │                │ 9. GetOrderInfo│                │                │
     │                │    ✅ 复用：现有 RPC            │                │
     │                ├────────────────────────────────>│                │
     │                │                │                │                │
     │                │                │ 10. SyncNewOrder                │
     │                │                │     🆕 使用 LinemanConverter    │
     │                │                │     转换订单格式                 │
     │                │                │                │                │
     │                │                │ 11. 保存订单到 Main 数据库      │
     │                │                │                ├───────────────>│
     │                │                │                │                │
     │                │                │ 12. 触发领域事件                 │
     │                │                │     ✅ 复用：库存检查            │
     │                │                │     ✅ 复用：送厨单生成          │
     │                │                │     ✅ 复用：打印通知            │
     │                │                │                │                │
     │                │                │                │                │
     │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                │
     │  ❌ 跳过：Grab 的 PrepareOrder (接单/拒单)       │                │
     │  ❌ 跳过：Grab 的 MarkOrderReady (准备完成)      │                │
     │  ❌ 跳过：Grab 的 CancelOrder (取消订单)        │                │
     │                │                │                │                │
     │  📝 跳过方案：                  │                │                │
     │  - 订单自动设为"已接单"状态      │                │                │
     │  - 或保持"待处理"由平板操作      │                │                │
     │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                │
```

### 2.3 订单状态更新流程（可选）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ LINE MAN │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. POST /status_update_notification             │
     │    ⚠️ 差异：Grab 用 PushOrderState              │
     ├───────────────>│                │                │
     │                │                │                │
     │                │ 2. HandleStatusUpdate           │
     │                │    🆕 新增：状态映射            │
     │                │    Lineman 状态 → TTPOS 状态   │
     │                │                │                │
     │                │ 3. 发送 MQ 消息                 │
     │                │    action: "status_update"     │
     │                ├───────────────>│                │
     │                │                │                │
     │  4. 返回 200   │                │ 5. 消费消息    │
     │<───────────────┤                ├───────────────>│
     │                │                │                │
     │                │                │ 6. UpdateOrderStatus
     │                │                │    ✅ 复用：现有逻辑            │
     │                │                │                │
```

### 2.4 订单编辑流程（🆕 Lineman 特有）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ LINE MAN │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. POST /order_update_notification              │
     │    🆕 Grab 不支持此功能                          │
     ├───────────────>│                │                │
     │                │                │                │
     │                │ 2. HandleOrderUpdate            │
     │                │    🆕 新增：订单编辑处理        │
     │                │    解析变更内容                 │
     │                │                │                │
     │                │ 3. 发送 MQ 消息                 │
     │                │    action: "order_update"      │
     │                ├───────────────>│                │
     │                │                │                │
     │  4. 返回 200   │                │ 5. 消费消息    │
     │<───────────────┤                ├───────────────>│
     │                │                │                │
     │                │                │ 6. UpdateOrderItems
     │                │                │    🆕 新增：更新订单商品        │
     │                │                │    重新计算金额                 │
     │                │                │    通知厨房变更                 │
     │                │                │                │
```

---

## 三、菜单流程映射

### 3.1 Grab 菜单推送流程（原版）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ 商家操作  │     │ Main     │     │ ttpos-bmp│     │   Grab   │     │ RocketMQ │
│ (Shop端) │     │ 模块     │     │  (网关)   │     │ Platform │     │  消息队列 │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. 启用外卖    │                │                │                │
     ├───────────────>│                │                │                │
     │                │ 2. ToggleTakeoutStatus         │                │
     │                │ 3. PushMenuToPlatform          │                │
     │                │    LoadMenuFromDatabase        │                │
     │                │    GrabConverter 转换          │                │
     │                ├───────────────>│                │                │
     │                │                │ 4. SaveMenuSnapshot            │
     │                │                │ 5. 调用 Grab API               │
     │                │                │    UpdateMenuV2                │
     │                │                ├───────────────>│                │
     │                │                │  6. Grab 异步处理              │
     │                │                │<───────────────┤                │
     │                │                │ 7. MenuSyncState Webhook       │
     │                │<───────────────┤                │                │
```

### 3.2 LINE MAN 菜单推送流程（映射版）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ 商家操作  │     │ Main     │     │ ttpos-bmp│     │ LINE MAN │     │ RocketMQ │
│ (Shop端) │     │ 模块     │     │  (网关)   │     │ Platform │     │  消息队列 │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. 启用 Lineman 外卖            │                │                │
     ├───────────────>│                │                │                │
     │                │                │                │                │
     │                │ 2. ToggleTakeoutStatus         │                │
     │                │    ✅ 复用：现有逻辑            │                │
     │                │    🆕 创建 Lineman 支付方式     │                │
     │                │                │                │                │
     │                │ 3. PushMenuToPlatform          │                │
     │                │    ✅ 复用：LoadMenuFromDatabase│                │
     │                │    🆕 使用 LinemanConverter    │                │
     │                │    转换为 Lineman 菜单格式     │                │
     │                │                │                │                │
     │                │ 4. 调用 BMP RPC                │                │
     │                ├───────────────>│                │                │
     │                │                │                │                │
     │                │                │ 5. SaveMenuSnapshot            │
     │                │                │    ✅ 复用：保存菜单快照        │
     │                │                │                │                │
     │                │                │ 6. SyncMenu (调用 Lineman API) │
     │                │                │    ⚠️ 差异：API 端点不同        │
     │                │                │    POST /menu_sync             │
     │                │                ├───────────────>│                │
     │                │                │                │                │
     │                │                │  7. Lineman 处理菜单           │
     │                │                │<───────────────┤                │
     │                │                │                │                │
     │                │                │ 8. Menu sync notification      │
     │                │                │    (可选) 同步结果回调          │
     │                │<───────────────┤                │                │
     │                │                │                │                │
```

### 3.3 触发同步流程（🆕 Lineman 特有）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ LINE MAN │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. POST /trigger_sync_menu                      │
     │    🆕 Grab 无此接口（Grab 用 GetMenu 拉取）      │
     ├───────────────>│                │                │
     │                │                │                │
     │                │ 2. HandleTriggerSync            │
     │                │    🆕 新增：接收触发请求        │
     │                │                │                │
     │                │ 3. 发送 MQ 消息                 │
     │                │    Topic: takeout_provider_menu_update          │
     │                │    ✅ 复用：现有 Topic          │
     │                ├───────────────>│                │
     │                │                │                │
     │  4. 返回 200   │                │ 5. 消费消息    │
     │<───────────────┤                ├───────────────>│
     │                │                │                │
     │                │                │ 6. PushMenuToPlatform          │
     │                │                │    ✅ 复用：菜单推送逻辑        │
     │                │<───────────────┤                │
     │                │                │                │
     │                │ 7. 执行菜单推送流程（同 3.2）   │
     │                │    ...                         │
     │                │                │                │
```

### 3.4 Grab 菜单拉取流程（❌ Lineman 不支持）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   Grab   │     │ ttpos-bmp│     │ Main     │     │ 数据库   │
│ Platform │     │  (网关)   │     │ 模块     │     │ MySQL   │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. GET /menu   │                │                │
     ├───────────────>│                │                │
     │                │ 2. HandleGetMenu               │
     │                │    查询本地快照                 │
     │                ├────────────────────────────────>│
     │                │                │                │
     │                │ 3. if 为空: 回退调用 Main      │
     │                ├───────────────>│                │
     │                │                │                │
     │ 4. 返回菜单    │                │                │
     │<───────────────┤                │                │
     │                │                │                │

─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
❌ Lineman 不支持：平台不会主动拉取菜单

📝 跳过方案：
- 无需实现此接口
- Lineman 只接收 TTPOS 推送的菜单
- 当 Lineman 需要菜单时，会调用 Trigger Sync Menu
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

---

## 四、门店流程映射

### 4.1 Grab 门店绑定流程（❌ Lineman 不支持）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ 商家操作  │     │ Main     │     │ ttpos-bmp│     │   Grab   │
│ (Shop端) │     │ 模块     │     │  (网关)   │     │ Platform │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. 获取绑定链接│                │                │
     ├───────────────>│                │                │
     │                │ 2. GetGrabBindingLink          │
     │                ├───────────────>│                │
     │                │                │ 3. CreateSelfServeJourney
     │                │                ├───────────────>│
     │                │                │<───────────────┤
     │                │<───────────────┤                │
     │<───────────────┤                │                │
     │                │                │                │
     │ 4. 商家点击链接完成绑定         │                │
     │────────────────────────────────────────────────>│
     │                │                │                │
     │                │                │ 5. IntegrationStatus Webhook
     │                │<───────────────┤<───────────────┤
     │                │                │                │

─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
❌ Lineman 不支持：无自助激活链接机制

📝 跳过方案：
- 提供手动配置界面
- 商家在 Lineman 后台获取 merchant_id
- 在 TTPOS Shop 端手动输入配置
- 或通过 Lineman 运营人员后台配置
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

### 4.2 LINE MAN 门店开关流程（🆕 Lineman 特有）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ POS 操作  │     │ Main     │     │ ttpos-bmp│     │ LINE MAN │
│ (收银端)  │     │ 模块     │     │  (网关)   │     │ Platform │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. 关闭/开启门店               │                │
     │    🆕 Grab 不支持此功能         │                │
     ├───────────────>│                │                │
     │                │                │                │
     │                │ 2. ForceCloseOpenRestaurant    │
     │                │    🆕 新增：门店控制逻辑        │
     │                ├───────────────>│                │
     │                │                │                │
     │                │                │ 3. 调用 Lineman API
     │                │                │    POST /force_close 或
     │                │                │    POST /force_open
     │                │                ├───────────────>│
     │                │                │                │
     │                │                │<───────────────┤
     │                │<───────────────┤                │
     │<───────────────┤                │                │
     │                │                │                │
```

---

## 五、关键差异汇总表

### 5.1 接口差异对照

| 流程          | Grab 接口                | Lineman 接口                 | 差异类型      | 处理方案                   |
| ------------- | ------------------------ | ---------------------------- | ------------- | -------------------------- |
| **订单推送**  | `SubmitOrder`            | `place_order_notification`   | ⚠️ 格式不同   | 新增 LinemanOrderConverter |
| **订单状态**  | `PushOrderState`         | `status_update_notification` | ⚠️ 状态值不同 | 状态映射表                 |
| **订单编辑**  | ❌ 不支持                | `order_update_notification`  | 🆕 新增       | 新增处理逻辑               |
| **菜单推送**  | `UpdateMenuV2`           | `menu_sync`                  | ⚠️ 格式不同   | 新增 LinemanMenuConverter  |
| **菜单拉取**  | `GetMenu`                | ❌ 不支持                    | ❌ 跳过       | 不实现                     |
| **触发同步**  | ❌ 不支持                | `trigger_sync_menu`          | 🆕 新增       | 新增 Webhook 端点          |
| **同步结果**  | `MenuSyncWebhook`        | `menu_sync_notification`     | ✅ 相似       | 复用逻辑                   |
| **单品状态**  | `UpdateMenuItem`         | `update_menu_item_status`    | ⚠️ 参数不同   | 适配参数                   |
| **接单/拒单** | `PrepareOrder`           | ❌ 不支持                    | ❌ 跳过       | 自动接单或平板操作         |
| **准备完成**  | `MarkOrderReady`         | ❌ 不支持                    | ❌ 跳过       | 不通知平台                 |
| **取消订单**  | `CancelOrder`            | ❌ 不支持                    | ❌ 跳过       | 仅平板操作                 |
| **门店绑定**  | `CreateSelfServeJourney` | ❌ 不支持                    | ❌ 跳过       | 手动配置                   |
| **门店开关**  | ❌ 不支持                | `force_close/open`           | 🆕 新增       | 新增功能                   |

### 5.2 不支持功能的跳过逻辑

```go
// main/app/modules/takeout/domain/service/takeout_order_service.go

// 订单操作跳过逻辑
func (s *TakeoutOrderService) PrepareOrder(ctx context.Context, orderUuid string, toState string) error {
    order, _ := s.orderRepo.GetByUuid(orderUuid)

    // ❌ Lineman 不支持接单操作，跳过 RPC 调用
    if order.Platform == "lineman" {
        logger.Logger.Info("Lineman 订单跳过 PrepareOrder，由平板处理",
            zap.String("order_uuid", orderUuid))
        // 方案1: 直接更新本地状态为已接单
        return s.orderRepo.UpdateStatus(orderUuid, OrderStateAccepted)
        // 方案2: 保持待处理状态，由平板操作
        // return nil
    }

    // Grab 正常调用 RPC
    return s.rpcService.PrepareOrder(ctx, orderUuid, toState)
}

func (s *TakeoutOrderService) MarkOrderReady(ctx context.Context, orderUuid string) error {
    order, _ := s.orderRepo.GetByUuid(orderUuid)

    // ❌ Lineman 不支持准备完成通知，跳过
    if order.Platform == "lineman" {
        logger.Logger.Info("Lineman 订单跳过 MarkOrderReady",
            zap.String("order_uuid", orderUuid))
        // 仅更新本地状态
        return s.orderRepo.UpdateStatus(orderUuid, OrderStateReady)
    }

    return s.rpcService.MarkOrderReady(ctx, orderUuid)
}

func (s *TakeoutOrderService) CancelOrder(ctx context.Context, orderUuid string, reason string) error {
    order, _ := s.orderRepo.GetByUuid(orderUuid)

    // ❌ Lineman 不支持 POS 取消订单
    if order.Platform == "lineman" {
        logger.Logger.Warn("Lineman 订单不支持 POS 取消，请使用 Lineman 平板",
            zap.String("order_uuid", orderUuid))
        return errors.New("Lineman 订单请使用平板取消")
    }

    return s.rpcService.CancelOrder(ctx, orderUuid, reason)
}
```

```go
// main/app/modules/takeout/application/takeout_app_service.go

// 菜单操作跳过逻辑
func (s *takeoutAppService) GetBindingLink(ctx context.Context, platform string, companyUuid uint64) (string, error) {
    // ❌ Lineman 不支持自助绑定链接
    if platform == "lineman" {
        logger.Logger.Info("Lineman 不支持自助绑定，请手动配置")
        return "", errors.New("Lineman 需要手动配置商户信息，请联系运营")
    }

    return s.rpcService.GetGrabBindingLink(ctx, companyUuid)
}
```

---

## 六、代码结构映射

### 6.1 BMP 层结构

```
ttpos-bmp/app/ttpos-takeout/
├── internal/
│   ├── controller/
│   │   ├── grab/                         # 现有 Grab 控制器
│   │   │   ├── grab_v1_submit_order.go
│   │   │   ├── grab_v1_get_menu.go
│   │   │   └── ...
│   │   └── lineman/                      # 🆕 新增 Lineman 控制器
│   │       ├── lineman_v1_place_order.go       # ⚠️ 对应 Grab SubmitOrder
│   │       ├── lineman_v1_status_update.go     # ⚠️ 对应 Grab PushOrderState
│   │       ├── lineman_v1_order_update.go      # 🆕 Lineman 特有
│   │       ├── lineman_v1_trigger_sync.go      # 🆕 Lineman 特有
│   │       └── lineman_v1_menu_sync_notify.go  # ✅ 对应 Grab MenuSyncState
│   │
│   ├── logic/
│   │   ├── grab/                         # 现有 Grab 逻辑
│   │   │   ├── grab_order.go
│   │   │   ├── grab_menu.go
│   │   │   └── ...
│   │   └── lineman/                      # 🆕 新增 Lineman 逻辑
│   │       ├── lineman.go                      # 核心服务
│   │       ├── lineman_order.go                # 订单处理
│   │       ├── lineman_menu.go                 # 菜单处理
│   │       └── lineman_auth.go                 # OAuth 认证
│   │
│   └── client/
│       ├── grab/                         # 现有 Grab 客户端
│       └── lineman/                      # 🆕 新增 Lineman 客户端
│           ├── client.go                       # HTTP 客户端
│           └── auth.go                         # Token 管理
```

### 6.2 Main 层结构

```
main/app/modules/takeout/
├── infrastructure/
│   └── adapter/
│       ├── grab/                         # 现有 Grab 适配器
│       │   ├── grab_menu_converter.go
│       │   └── grab_order_converter.go
│       └── lineman/                      # 🆕 新增 Lineman 适配器
│           ├── lineman_menu_converter.go       # 菜单格式转换
│           └── lineman_order_converter.go      # 订单格式转换
│
├── domain/
│   └── service/
│       └── takeout_order_service.go      # ⚠️ 修改：添加平台判断逻辑
│
└── application/
    ├── takeout_app_service.go            # ⚠️ 修改：注册 Lineman 转换器
    └── takeout_order_service.go          # ⚠️ 修改：添加平台判断逻辑
```

---

## 七、MQ Topic 设计

### 方案 A：复用现有 Topic（推荐）

```go
// 复用现有 Topic，通过 provider_name 区分
const (
    TopicProviderMenuUpdate  = "takeout_provider_menu_update"   // Grab + Lineman 共用
    TopicProviderOrderUpdate = "takeout_grab_order"             // 改名为通用名或保持不变
)

// Event 结构中用 provider_name 区分
type OrderEvent struct {
    Action       string `json:"action"`        // create, status_update, cancel, order_update
    ProviderName string `json:"providerName"`  // "grab" 或 "lineman"
    // ...
}
```

### 方案 B：独立 Topic

```go
const (
    // Grab Topic
    TopicGrabMenuUpdate  = "takeout_grab_menu_update"
    TopicGrabOrder       = "takeout_grab_order"

    // Lineman Topic
    TopicLinemanMenuUpdate = "takeout_lineman_menu_update"
    TopicLinemanOrder      = "takeout_lineman_order"
)
```

**建议使用方案 A**，减少代码变更和维护成本。

---

## 八、总结：Lineman 接入核心改动点

| 改动类型 | 模块 | 文件/目录                  | 工作量 |
| -------- | ---- | -------------------------- | ------ |
| 🆕 新增  | BMP  | `controller/lineman/`      | 2d     |
| 🆕 新增  | BMP  | `logic/lineman/`           | 3d     |
| 🆕 新增  | BMP  | `client/lineman/`          | 1d     |
| 🆕 新增  | Main | `adapter/lineman/`         | 3d     |
| ⚠️ 修改  | Main | `takeout_order_service.go` | 1d     |
| ⚠️ 修改  | Main | `takeout_app_service.go`   | 0.5d   |
| ⚠️ 修改  | Main | `queue/takeout/`           | 0.5d   |
| ✅ 复用  | 全局 | RocketMQ Topic             | -      |
| ✅ 复用  | 全局 | 领域事件、送厨单、打印     | -      |

**总预估：11 人天（MVP）**
