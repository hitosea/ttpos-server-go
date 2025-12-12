# ttpos-takeout MQ Topic

## 总览

- **模块**：`app/ttpos-takeout`
- **队列驱动**：RocketMQ（见 `app/ttpos-takeout/manifest/config/config.tpl.yaml`）
- **特征**：当前代码侧主要是 **生产事件**（未发现该模块内的消费者实现）。

## Topic 清单

### 1) `takeout_provider_menu_update`

- **用途**：Grab 菜单更新事件（通知下游执行菜单同步/刷新）。
- **生产者**：`app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` → `(*sGrabMenu).NotifyMenuUpdate()`
  - 发送：`queue.PushWithContext(ctx, "takeout_provider_menu_update", event)`
- **消费者**：未在 `ttpos-takeout` 内发现（可能由其他模块/服务订阅）。
- **消息体**：`app/ttpos-takeout/internal/model/dto/grab/menu.go` → `grab.ProviderMenuUpdateEvent`

```json
{
  "provider_name": "grab",
  "merchant_id": "...",
  "partner_merchant_id": "...",
  "shop_uuid": "...",
  "uuid": 123,
  "received_at": 1730000000
}
```

### 2) `takeout_store_integration_state`

- **用途**：门店第三方集成状态变更事件（如 Grab 门店绑定/解绑/状态变更）。
- **生产者**：`app/ttpos-takeout/internal/logic/shop_provider_cfg/shop_provider_cfg.go` → `(*sShopProviderCfg).NotifyStoreIntegrationState()`
  - 发送：`queue.PushWithContext(ctx, "takeout_store_integration_state", event)`
- **消费者**：未在 `ttpos-takeout` 内发现。
- **消息体**：`app/ttpos-takeout/internal/model/dto/grab/event.go` → `grab.ShopIntegrationStatusEvent`

> 注意：该结构体当前没有 `json` tag；通过 `gjson.EncodeString` 序列化时，字段名可能表现为 `ShopUuid/ProviderName/...`（首字母大写）。若有跨语言订阅者，建议补齐 `json` tag 并保持字段名稳定。

### 3) `takeout_grab_order`

- **用途**：Grab 订单事件（新订单、状态变更等），用于异步处理/通知下游。
- **生产者**：`app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - `(*sGrabOrder).HandleSubmitOrder()`：发送 `Action=create`
  - `(*sGrabOrder).HandlePushOrderState()`：发送 `Action=status_update`
- **消费者**：未在 `ttpos-takeout` 内发现。
- **消息体**：`app/ttpos-takeout/internal/logic/grab_order/grab_order.go` → `OrderEvent`

```json
{
  "action": "create|status_update|cancel",
  "providerName": "grab",
  "orderUuid": "...",
  "orderId": "...",
  "merchantId": "...",
  "status": "...",
  "timestamp": 1730000000
}
```

## 备注

- 若下游需要可靠消费（幂等、重试、死信等），建议在消费者侧按 `orderUuid`/`uuid` 做幂等。
