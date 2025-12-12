# ttpos-erp MQ Topic

## 总览

- **模块**：`app/ttpos-erp`
- **队列驱动**：RocketMQ（见 `app/ttpos-erp/manifest/config/config.tpl.yaml`）
- **消费者注册入口**：`app/ttpos-erp/internal/boot/consumer.go`（`queue.RegisterConsumer(...)` + `queue.StartConsumersListener(ctx)`）

## Topic 清单

### 1) `item-sync`

- **用途**：触发商品同步任务（当前消费者仅打印日志，业务逻辑待补齐）。
- **生产者**：`app/ttpos-erp/internal/logic/stock/item.go` → `(*sItem).SyncDelay()`
  - 发送：`queue.Push("item-sync", &erp.Item{})`
- **消费者**：`app/ttpos-erp/internal/consumer/item_sync.go` → `ItemSyncConsumer`
  - 处理：仅 `g.Log().Info(...)` 打印消息体
- **消息体**：当前生产者发送 `erp.Item{}`（空对象）；消费者未解析。

### 2) `item-sync-delay`

- **用途**：延迟触发商品同步任务（10s）。
- **生产者**：`app/ttpos-erp/internal/logic/stock/item.go` → `(*sItem).SyncDelay()`
  - 发送：`queue.DelayPush("item-sync-delay", &erp.Item{}, 10s)`
- **消费者**：**未在代码中注册**（当前只看到 `item-sync` 有消费者）。
- **消息体**：同 `item-sync`。

### 3) `erp-doc-change`

- **用途**：ERPNext 文档变更回调异步化（回调入口快速返回，后续由消费者处理）。
- **生产者**：`app/ttpos-erp/internal/controller/callback/callback_v1_doc_change.go` → `(*ControllerV1).DocChange()`
  - 发送：`queue.PushWithContext(ctx, "erp-doc-change", req)`
- **消费者**：**未在代码中找到 GetTopic/Handle 实现**（仅发现生产侧）。
- **消息体**：`app/ttpos-erp/api/callback/v1/doc.go` → `v1.DocChangeReq`

```json
{
  "event": "...",
  "data": {},
  "docType": "...",
  "docName": "...",
  "siteCode": "..."
}
```

### 4) `erp-item-change`

- **用途**：ERPNext 商品变更回调异步化。
- **生产者**：`app/ttpos-erp/internal/controller/callback/callback_v1_item_change.go` → `(*ControllerV1).ItemChange()`
  - 发送：`queue.PushWithContext(ctx, "erp-item-change", req)`
- **消费者**：**未在代码中找到 GetTopic/Handle 实现**（仅发现生产侧）。
- **消息体**：`app/ttpos-erp/api/callback/v1/item.go` → `v1.ItemChangeReq`

```json
{
  "event": "...",
  "data": {},
  "docType": "...",
  "docName": "...",
  "siteCode": "..."
}
```

### 5) `save-pos-invoice`

- **用途**：POS 下单发票保存异步化（先落库，再通过 MQ 驱动后台处理）。
- **生产者**：`app/ttpos-erp/internal/logic/selling/async_selling.go` → `(*sAsyncSelling).SavePosInvoice()`
  - 行为：插入 `receive_pos_invoice` 记录，发送 `AsyncSellingMsg{record_id, msg_type}`
- **消费者**：`app/ttpos-erp/internal/consumer/selling/selling_consumer.go` → `SavePosInvoiceConsumer`
  - 行为：
    1) 解析消息 `AsyncSellingMsg`
    2) 按 `record_id` 读取暂存表
    3) base64 解码并 proto 反序列化原始请求
    4) 调用 `service.Selling().SavePosInvoice(...)`
    5) 回写响应与状态
- **消息体**：`app/ttpos-erp/internal/model/mq/async_selling.go` → `mq.AsyncSellingMsg`

```json
{
  "record_id": 123,
  "msg_type": "save-pos-invoice"
}
```

### 6) `return-pos-invoice`

- **用途**：POS 退款发票异步化。
- **生产者**：`app/ttpos-erp/internal/logic/selling/async_selling.go` → `(*sAsyncSelling).ReturnPosInvoice()`
- **消费者**：`app/ttpos-erp/internal/consumer/selling/selling_consumer.go` → `ReturnPosInvoiceConsumer`
- **消息体**：同 `AsyncSellingMsg`

```json
{ "record_id": 123, "msg_type": "return-pos-invoice" }
```

### 7) `cancel-pos-invoice`

- **用途**：POS 取消发票异步化。
- **生产者**：`app/ttpos-erp/internal/logic/selling/async_selling.go` → `(*sAsyncSelling).CancelPosInvoice()`
- **消费者**：`app/ttpos-erp/internal/consumer/selling/selling_consumer.go` → `CancelPosInvoice`
- **消息体**：同 `AsyncSellingMsg`

```json
{ "record_id": 123, "msg_type": "cancel-pos-invoice" }
```

### 8) `close-pos-entry`

- **用途**：POS 关账异步化。
- **生产者**：`app/ttpos-erp/internal/logic/selling/async_selling.go` → `(*sAsyncSelling).ClosePosEntry()`
- **消费者**：`app/ttpos-erp/internal/consumer/selling/selling_consumer.go` → `ClosePosEntryConsumer`
- **消息体**：同 `AsyncSellingMsg`

```json
{ "record_id": 123, "msg_type": "close-pos-entry" }
```

### 9) `redo-pos`

- **用途**：重做未处理的 POS 异步任务（按开单号重发 Draft 状态记录）。
- **生产者**：未在 `ttpos-erp` 代码中直接发现（可能由运维/脚本/其他服务触发）。
- **消费者**：`app/ttpos-erp/internal/consumer/selling/selling_consumer.go` → `RedoPosConsumer`
- **消息体**：`mq.AsyncSellingMsg`，会使用 `pos_open_entry_name`，`site_code` 可选。

```json
{
  "msg_type": "save-pos-invoice",
  "pos_open_entry_name": "POS-OPE-2025-00238",
  "site_code": "SITE001"
}
```

## 运维/排查提示

- `ttpos-bmp/internal/pkg/queue/rocketmq.go` 在发送前会尝试自动创建 topic（配置了 `queue.rocketmq.brokerAddr` 时）。
- `redo-pos` 建议在消息体中带上 `site_code`，避免跨站点误重发。
