我来帮你梳理 TTPOS 与 Grab 的集成逻辑。首先让我检查 Serena 的状态并搜索相关代码。

[2 tools called]

搜索结果太多，让我更精确地分模块搜索 Grab 相关代码：

[3 tools called]

发现了关键代码。让我进一步了解整体架构：

[4 tools called]

让我继续探索 BMP 侧的 Grab 实现和整体架构：

[3 tools called]

让我继续查看 BMP 侧的核心实现和 API：

[5 tools called]

让我继续查看关键服务和数据模型：

[3 tools called]

让我继续查看消息队列处理和 webhook 流程：

[4 tools called]

让我继续查看 BMP 侧的 Grab 相关实现：

[2 tools called]

让我查看更多 BMP 侧的关键实现：

[3 tools called]

现在我已经收集了足够的信息，让我为你梳理 TTPOS 与 Grab 的完整集成逻辑。

---

# TTPOS 与 Grab 集成架构文档

## 一、总体架构概览

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│                                  Grab Platform                                      │
│                           (GrabFood API + Webhook)                                 │
└───────────────────────────────────┬────────────────────────────────────────────────┘
                                    │
           ┌────────────────────────┼────────────────────────┐
           │                        │                        │
           ▼                        ▼                        ▼
    ┌─────────────┐        ┌─────────────┐         ┌─────────────┐
    │  获取菜单    │        │  订单推送    │         │  菜单推送    │
    │ GET /menu   │        │ submitOrder │         │ pushGrabMenu│
    └──────┬──────┘        └──────┬──────┘         └──────┬──────┘
           │                      │                       │
           └──────────────────────┼───────────────────────┘
                                  │
                                  ▼
┌────────────────────────────────────────────────────────────────────────────────────┐
│                          ttpos-bmp (中间层/网关)                                    │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    internal/controller/grab/         (HTTP Controller - Webhook 入口)        │  │
│  │    - grab_v1_get_menu.go            ← Grab 获取菜单                          │  │
│  │    - grab_v1_submit_order.go        ← Grab 提交订单                          │  │
│  │    - grab_v1_push_grab_menu_webhook.go ← Grab 推送菜单                       │  │
│  │    - grab_v1_push_order_state.go    ← Grab 推送订单状态                      │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    internal/logic/grab/              (业务逻辑层)                            │  │
│  │    - grab.go                        ← 核心服务入口                           │  │
│  │    - grab_menu.go                   ← 菜单处理逻辑                           │  │
│  │    - grab_order.go                  ← 订单处理逻辑                           │  │
│  │    - grab_webhook.go                ← Webhook 验签+处理                      │  │
│  │    - grab_store.go                  ← 门店配置管理                           │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
│                                  │                                                 │
│                    RocketMQ 消息 │                                                 │
│                    (异步解耦)    ▼                                                 │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    Topic: takeout_grab_order        ← 订单事件                               │  │
│  │    Topic: takeout_provider_menu_update ← 菜单更新事件                         │  │
│  │    Topic: takeout_store_integration_state ← 门店集成状态                      │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────┬────────────────────────────────────────────────┘
                                    │
                              gRPC  │  RocketMQ
                                    ▼
┌────────────────────────────────────────────────────────────────────────────────────┐
│                           ttpos-server-go (Main 模块)                              │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    app/queue/takeout/                (MQ 消费者)                             │  │
│  │    - takeout_provider_order_update.go  ← 订单消息处理                        │  │
│  │    - takeout_provider_menu_update.go   ← 菜单消息处理                        │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    app/modules/takeout/application/   (应用层)                               │  │
│  │    - takeout_order_service.go        ← 订单应用服务                          │  │
│  │    - takeout_app_service.go          ← 外卖应用服务                          │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    app/modules/takeout/infrastructure/adapter/grab/                          │  │
│  │    - grab_menu_converter.go          ← Grab 菜单格式转换                     │  │
│  │    - grab_order_converter.go         ← Grab 订单格式转换                     │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────────┐  │
│  │    app/modules/takeout/infrastructure/adapter/rpc/                           │  │
│  │    - bmp_client.go                   ← BMP gRPC 客户端                       │  │
│  └──────────────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 二、核心数据流

### 2.1 订单流程（Grab → TTPOS）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   Grab   │     │ ttpos-bmp│     │ RocketMQ │     │ Main     │     │ 数据库   │
│ Platform │     │  (网关)   │     │  消息队列 │     │ 模块     │     │ MySQL   │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. POST /submitOrder            │                │                │
     │ (SubmitOrderRequest)            │                │                │
     ├───────────────>│                │                │                │
     │                │                │                │                │
     │                │ 2. 验证签名 (Middleware)         │                │
     │                │    (VerifyWebhookSignature)    │                │
     │                │                │                │                │
     │                │ 3. HandleSubmitOrder           │                │
     │                │    保存订单到 BMP 数据库        │                │
     │                ├────────────────────────────────────────────────>│
     │                │                │                │                │
     │                │ 4. 发送 MQ 消息                 │                │
     │                │    Topic: takeout_grab_order   │                │
     │                ├───────────────>│                │                │
     │                │                │                │                │
     │  5. 返回 200 OK│                │                │                │
     │<───────────────┤                │                │                │
     │                │                │                │                │
     │                │                │ 6. 消费消息    │                │
     │                │                ├───────────────>│                │
     │                │                │                │                │
     │                │                │  7. HandlePushOrderState       │
     │                │                │     调用 RPC 获取订单详情       │
     │                │<───────────────────────────────┤ (gRPC)         │
     │                │                │                │                │
     │                │ 8. GetOrderInfo│                │                │
     │                ├────────────────────────────────>│                │
     │                │                │                │                │
     │                │                │  9. SyncNewOrder               │
     │                │                │     转换订单格式                │
     │                │                │     (GrabConverter)            │
     │                │                │                │                │
     │                │                │  10. 保存订单到 Main 数据库    │
     │                │                │                ├───────────────>│
     │                │                │                │                │
     │                │                │  11. 触发领域事件               │
     │                │                │      - 库存检查                 │
     │                │                │      - 送厨单生成               │
     │                │                │      - 打印通知                 │
     │                │                │                │                │
```

### 2.2 菜单同步流程（TTPOS → Grab）

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ 商家操作  │     │ Main     │     │ ttpos-bmp│     │   Grab   │     │ RocketMQ │
│ (Shop端) │     │ 模块     │     │  (网关)   │     │ Platform │     │  消息队列 │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │                │
     │ 1. 启用 Grab 外卖              │                │                │
     ├───────────────>│                │                │                │
     │                │                │                │                │
     │                │ 2. ToggleTakeoutStatus         │                │
     │                │    创建 Grab 支付方式           │                │
     │                │                │                │                │
     │                │ 3. PushMenuToPlatform          │                │
     │                │    加载 TTPOS 商品数据          │                │
     │                │    (LoadMenuFromDatabase)      │                │
     │                │                │                │                │
     │                │ 4. 转换为 Grab 格式             │                │
     │                │    (GrabConverter)             │                │
     │                │                │                │                │
     │                │ 5. 调用 BMP RPC 保存菜单快照    │                │
     │                ├───────────────>│                │                │
     │                │                │                │                │
     │                │                │ 6. SaveMenuSnapshot            │
     │                │                │    保存到 channel_menu 表      │
     │                │                │                │                │
     │                │                │ 7. 调用 Grab API               │
     │                │                │    SyncMenu (UpdateMenuV2)     │
     │                │                ├───────────────>│                │
     │                │                │                │                │
     │                │                │  8. Grab 异步处理菜单          │
     │                │                │<───────────────┤                │
     │                │                │                │                │
     │                │                │ 9. Webhook: MenuSyncState      │
     │                │<───────────────┤                │                │
     │                │                │                │                │
     │                │ 10. 更新同步状态                │                │
     │                │                │                │                │
```

### 2.3 Grab 获取菜单流程

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   Grab   │     │ ttpos-bmp│     │ Main     │     │ 数据库   │
│ Platform │     │  (网关)   │     │ 模块     │     │ MySQL   │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. GET /menu   │                │                │
     │    ?partnerMerchantID=xxx       │                │
     ├───────────────>│                │                │
     │                │                │                │
     │                │ 2. HandleGetMenuWrapper         │
     │                │    验证签名                     │
     │                │                │                │
     │                │ 3. 查询本地菜单快照              │
     │                ├────────────────────────────────>│
     │                │                │                │
     │                │ 4. if 本地快照为空:              │
     │                │    回退调用 Main /menu/export   │
     │                ├───────────────>│                │
     │                │                │                │
     │                │                │ 5. 从数据库加载 │
     │                │                │    TTPOS 商品   │
     │                │                ├───────────────>│
     │                │                │                │
     │                │                │ 6. 转换格式    │
     │                │<───────────────┤                │
     │                │                │                │
     │ 7. 返回菜单数据│                │                │
     │<───────────────┤                │                │
     │                │                │                │
```

---

## 三、关键组件说明

### 3.1 BMP 侧关键组件

| 文件路径                                                     | 职责                   |
| ------------------------------------------------------------ | ---------------------- |
| `internal/controller/grab/grab_v1_submit_order.go`           | Grab 订单 Webhook 入口 |
| `internal/controller/grab/grab_v1_get_menu.go`               | Grab 获取菜单入口      |
| `internal/controller/grab/grab_v1_push_grab_menu_webhook.go` | Grab 推送菜单 Webhook  |
| `internal/logic/grab/grab.go`                                | Grab 服务核心入口      |
| `internal/logic/grab/grab_order.go`                          | 订单处理逻辑           |
| `internal/logic/grab/grab_menu.go`                           | 菜单处理逻辑           |
| `internal/logic/grab/grab_webhook.go`                        | Webhook 签名验证       |
| `internal/middleware/grab_signature_auth.go`                 | 签名验证中间件         |

### 3.2 Main 侧关键组件

| 文件路径                                                                  | 职责                     |
| ------------------------------------------------------------------------- | ------------------------ |
| `app/queue/takeout/takeout_provider_order_update.go`                      | 订单 MQ 消费者           |
| `app/modules/takeout/application/takeout_order_service.go`                | 订单应用服务             |
| `app/modules/takeout/application/takeout_app_service.go`                  | 外卖应用服务（菜单管理） |
| `app/modules/takeout/infrastructure/adapter/grab/grab_order_converter.go` | Grab 订单格式转换        |
| `app/modules/takeout/infrastructure/adapter/grab/grab_menu_converter.go`  | Grab 菜单格式转换        |
| `app/modules/takeout/infrastructure/adapter/rpc/bmp_client.go`            | BMP gRPC 客户端          |
| `app/modules/takeout/domain/service/takeout_order_service.go`             | 订单领域服务             |

---

## 四、详细调用链

### 4.1 订单创建调用链

```go
// 1. BMP 接收 Webhook
grab_v1_submit_order.go::SubmitOrder()
    │
    ▼
// 2. 业务逻辑处理
grab_order.go::HandleSubmitOrder(ctx, *grabfood.SubmitOrderRequest)
    │
    ├── saveOrderFromSDK()  // 保存订单到 BMP 数据库
    │       │
    │       ├── dao.Order.Insert()       // 主订单表
    │       ├── dao.OrderItem.Insert()   // 订单商品表
    │       └── dao.OrderItemModifier.Insert()  // 商品修饰符表
    │
    └── queue.Push(TopicGrabOrder, OrderEvent)  // 发送 MQ 消息
            │
            ▼
// 3. Main 消费 MQ 消息
takeout_provider_order_update.go::TakeoutProviderOrderUpdateHandler()
    │
    ▼
// 4. 订单应用服务
takeout_order_service.go::HandlePushOrderState()
    │
    ├── case "create": SyncNewOrder()
    │       │
    │       ├── converters["grab"].ParseOrderWebhook()  // 解析 Webhook 数据
    │       ├── converters["grab"].ConvertOrderToTakeoutOrder()  // 转换订单
    │       ├── converters["grab"].ConvertReceiverInfo()  // 转换收货人
    │       ├── converters["grab"].ConvertCampaigns()  // 转换活动
    │       └── orderService.CreateOrder()  // 创建订单
    │               │
    │               ├── persistence.TakeoutOrderRepo.Create()
    │               ├── persistence.TakeoutOrderItemRepo.CreateBatch()
    │               ├── persistence.TakeoutOrderItemModifierRepo.CreateBatch()
    │               ├── persistence.TakeoutOrderReceiverRepo.Create()
    │               └── 触发领域事件 (库存检查、送厨单)
    │
    └── case "status_update": UpdateOrderStatus()
```

### 4.2 菜单推送调用链

```go
// 1. Shop 端启用外卖
api/v1/shop/shop_takeout.go::ToggleTakeoutStatus()
    │
    ▼
// 2. 应用服务
takeout_app_service.go::ToggleTakeoutStatus()
    │
    ├── SaveGrabPaymentMethod()  // 创建 Grab 支付方式
    └── PushMenuToPlatform()
            │
            ▼
// 3. 菜单导出
takeout_app_service.go::ExportMenu()
    │
    ├── grabConverter.LoadMenuFromDatabase()  // 加载 TTPOS 商品
    │       │
    │       ├── 查询分类 (ttpos_product_category)
    │       ├── 查询外卖商品 (ttpos_product_package_takeout)
    │       ├── 转换为 grabfood.MenuCategory
    │       ├── 转换为 grabfood.MenuItem
    │       └── 转换为 grabfood.ModifierGroup
    │
    └── rpcService.SaveMenuSnapshot()  // 调用 BMP
            │
            ▼
// 4. BMP 保存并推送
grab_menu.go::SaveMenuSnapshot()
    │
    ├── dao.ChannelMenu.Upsert()  // 保存菜单快照
    └── SyncMenu()
            │
            └── grabclient.Default().UpdateMenuV2()  // 调用 Grab API
```

### 4.3 Grab 获取菜单调用链

```go
// 1. Grab 调用获取菜单
grab_v1_get_menu.go::GetMenu()
    │
    ▼
// 2. 业务逻辑
grab_webhook.go::HandleGetMenuWrapper()
    │
    ├── HandleGetMenu()
    │       │
    │       ├── service.ChannelMenu().GetTtposMenu()  // 查询本地快照
    │       │
    │       └── if 为空: fetchMenuFromTTpos()  // 回退调用 Main
    │               │
    │               └── HTTP POST /api/v1/takeout/menu/export
    │
    └── 设置 MerchantID / PartnerMerchantID
```

---

## 五、数据模型映射

### 5.1 订单数据映射

| Grab 字段                | TTPOS 字段             | 说明          |
| ------------------------ | ---------------------- | ------------- |
| `orderID`                | `platform_order_id`    | 平台订单号    |
| `shortOrderNumber`       | `short_order_number`   | 短订单号      |
| `partnerMerchantID`      | `company_uuid`         | 商户标识      |
| `orderState`             | `order_state`          | 订单状态      |
| `price.subtotal`         | `subtotal`             | 小计          |
| `price.eaterPayment`     | `total_amount`         | 总金额        |
| `items[].id`             | `platform_item_id`     | 平台商品 ID   |
| `items[].modifiers[].id` | `platform_modifier_id` | 平台修饰符 ID |

### 5.2 菜单数据映射

| TTPOS 字段                       | Grab 字段     | 说明         |
| -------------------------------- | ------------- | ------------ |
| `product_category.uuid`          | `category.id` | 分类 ID      |
| `product_package.uuid`           | `item.id`     | 商品 ID      |
| `product_bom.uuid`               | `modifier.id` | 规格/加料 ID |
| `product_package_attribute.uuid` | `modifier.id` | 属性 ID      |

---

## 六、关键配置

### 6.1 RocketMQ Topics

```go
// main/app/queue/queue.go
const (
    TopicProviderMenuUpdate    = "takeout_provider_menu_update"   // 菜单更新
    TopicProviderOrderUpdate   = "takeout_grab_order"             // 订单更新
    TopicStoreIntegrationState = "takeout_store_integration_state" // 门店状态
)
```

### 6.2 平台标识

```go
// 平台常量
const (
    TakeoutPlatformGrab = "grab"
    TakeoutPlatformLineman = "lineman"
)
```

---

## 七、错误处理与重试

### 7.1 订单创建失败处理

1. **BMP 侧**：订单保存失败返回错误，Grab 会重试
2. **MQ 消费失败**：消息保留在队列，自动重试
3. **Main 侧转换失败**：记录日志，不影响 BMP 已保存的订单

### 7.2 菜单同步失败处理

1. **导出失败**：返回错误给前端，用户可重试
2. **BMP 推送失败**：记录日志，更新同步状态为失败
3. **Grab API 失败**：通过 MenuSyncState Webhook 回调通知

---

## 八、扩展说明

### 8.1 新增平台适配

1. 在 `infrastructure/adapter/` 下创建新平台目录
2. 实现 `IPlatformConverter` 接口
3. 在 `application/` 层注册转换器
4. 配置新的 MQ Topic（如需要）

### 8.2 SDK 依赖

```go
// Grab SDK
import grabfood "github.com/grab/grabfood-api-sdk-go"
```

---

**文档版本**: v1.0  
**最后更新**: 2026-01-09
