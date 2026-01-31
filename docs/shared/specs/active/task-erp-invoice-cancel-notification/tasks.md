# task-erp-invoice-cancel-notification 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 5 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: Protobuf + 消费者逻辑

### 1.1 扩展 ReturnPosInvoiceReq protobuf

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` |
| Purpose | 增加 remark 可选字段 |
| Requirements | Req1 - 扩展 ReturnPosInvoiceReq 结构 |
| Leverage | 参考现有 protobuf 字段定义模式 |

**变更内容**:
```protobuf
message ReturnPosInvoiceReq {
  // ... 现有字段 ...
  string remark = 10;  // [NEW] 附注/备注，可选
}
```

**执行命令**:
```bash
cd ttpos-bmp/app/ttpos-erp && make pb
```

- [ ] 完成

---

### 1.2 扩展 Consumer 处理逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go` |
| Purpose | 退票处理完成后发送通知消息 |
| Requirements | Req2 - 发送 MQ 消息通知 |
| Leverage | `queue.PushWithContext()` MQ 发送函数 |

**修改点**:
- 在 `ReturnPosInvoiceConsumer` 处理成功后调用 `sendInvoiceCancelNotify()`
- 通知发送失败仅记录日志，不阻塞主流程

- [ ] 完成

---

## Phase 2: Topic + 消息结构

### 2.1 新增 Topic 常量

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/consts/topic.go` |
| Purpose | 定义 erp-invoice-cancel topic 常量 |
| Requirements | Req2 - 发送到指定 topic |
| Leverage | 参考现有 Topic 常量定义模式 |

**新增内容**:
```go
const (
    // ... 现有常量 ...
    TopicErpInvoiceCancel = Topic("erp-invoice-cancel")
)
```

- [ ] 完成

---

### 2.2 新建通知消息结构

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/mq/invoice_cancel_notify.go` |
| Purpose | 定义发票取消通知消息结构 |
| Requirements | Req3 - 消息格式规范 |
| Leverage | 参考现有 mq 消息结构定义 |

**新建内容**:
```go
package mq

// InvoiceCancelNotifyMsg 发票取消通知消息
type InvoiceCancelNotifyMsg struct {
    OrderNo     string `json:"order_no"`      // 订单号
    InvoiceName string `json:"invoice_name"`  // 发票名称
    Remark      string `json:"remark"`        // 附注信息
}
```

- [ ] 完成

---

### 2.3 新增通知发送方法

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go` |
| Purpose | 实现 sendInvoiceCancelNotify 方法 |
| Requirements | Req2 - 发送 MQ 消息通知 |
| Leverage | `queue.PushWithContext()` + `g.Log()` |

**新增方法**:
```go
// sendInvoiceCancelNotify 发送发票取消通知
func (c *SellingConsumer) sendInvoiceCancelNotify(ctx context.Context, req *selling.ReturnPosInvoiceReq) {
    notifyMsg := &mq.InvoiceCancelNotifyMsg{
        OrderNo:     req.OrderNo,
        InvoiceName: req.InvoiceName,
        Remark:      req.Remark,
    }

    if err := queue.PushWithContext(ctx, string(consts.TopicErpInvoiceCancel), notifyMsg); err != nil {
        g.Log().Errorf(ctx, "发送发票取消通知失败: %v", err)
        return  // 仅记录日志，不阻塞主流程
    }

    g.Log().Infof(ctx, "发票取消通知已发送: order_no=%s, invoice_name=%s",
        req.OrderNo, req.InvoiceName)
}
```

- [ ] 完成

---

## 提交清单

### 代码质量

- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] `make pb` 执行（生成 protobuf 代码）

### 功能完整性

- [ ] remark 字段可选且正确处理
- [ ] MQ 消息正确发送到 erp-invoice-cancel topic
- [ ] 消息格式符合规范（order_no, invoice_name, remark）
- [ ] remark 为空时正常处理

### 日志要求

- [ ] MQ 发送成功/失败均有日志记录
- [ ] 日志包含 order_no、invoice_name 等关键信息

---

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `manifest/protobuf/selling/selling.proto` | 修改 | 增加 remark 字段 |
| `internal/consts/topic.go` | 修改 | 新增 TopicErpInvoiceCancel |
| `internal/model/mq/invoice_cancel_notify.go` | 新建 | 通知消息结构 |
| `internal/consumer/selling/selling_consumer.go` | 修改 | 增加通知发送逻辑 |

---

**版本**: v1.0.0
**创建日期**: 2026-01-29
