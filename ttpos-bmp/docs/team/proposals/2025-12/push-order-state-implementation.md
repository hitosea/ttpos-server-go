# PushOrderState 控制器实现提案

> 实现 Grab 订单状态推送 Webhook 控制器

---

## 📋 提案信息

| 项目         | 内容                               |
| ------------ | ---------------------------------- |
| **提案人**   | AI Assistant                       |
| **日期**     | 2025-12-19                         |
| **目标版本** | -                                  |
| **状态**     | ✅ 已批准                           |
| **关联任务** | -                                  |
| **关联 Spec** | [task-takeout-grab-push-order-state](../../../../docs/shared/specs/active/task-takeout-grab-push-order-state/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

`grab_v1_push_order_state.go` 控制器当前未实现，返回 `CodeNotImplemented` 错误。需要通过调用现有的 `HandlePushOrderState` 方法完成实现，以接收 Grab 平台推送的订单状态变更通知。

### 现状分析

1. **API 定义** (`api/grab/v1/push_order_state.go`):
   - `PushOrderStateReq` 嵌入 `*grabfood.OrderStateRequest`
   - GoFrame 会自动将 JSON 解析到类型化结构体

2. **服务接口** (`service/grab.go`):
   ```go
   HandlePushOrderState(ctx context.Context, body []byte) error
   ```
   - 接收 `[]byte` 原始字节

3. **接口不一致问题**:
   - 控制器收到的是类型化请求 `*grabfood.OrderStateRequest`
   - 服务层期望的是 `[]byte`
   - 这与 `HandleSubmitOrder` 直接接收类型化请求的模式不一致

### 业务价值

- 完成 Grab 订单状态推送功能
- 保持代码架构一致性
- 支持订单状态实时同步（接受/拒绝/取消/配送等）

---

## 💡 解决方案概述

### 方案 A：修改接口为类型化请求（推荐）

修改 `HandlePushOrderState` 接口，使其接收 `*grabfood.OrderStateRequest` 而非 `[]byte`，与 `HandleSubmitOrder` 保持一致。

**优点**：
- 接口一致性好
- 与 `HandleSubmitOrder` 模式相同
- 类型安全

**需要修改的文件**：

1. `internal/logic/grab_order/grab_order.go` - 增加 `OrderEvent.ShopUUID` 字段
2. `internal/logic/grab_order/grab_order.go` - 修改 `HandlePushOrderState` 方法签名
3. `internal/logic/grab/grab.go` - 修改代理方法
4. 执行 `gf gen service` 重新生成 service 接口

### 方案 B：控制器中序列化（备选）

在控制器中将类型化请求序列化为 JSON 字节，传递给现有接口。

**优点**：
- 不需要修改服务层接口

**缺点**：
- 额外的序列化开销
- 可能丢失原始请求中未映射的字段
- 与其他控制器模式不一致

---

## 🔧 推荐实现方案（方案 A）

### 1. 修改 `OrderEvent` 结构体

```go
// OrderEvent 订单事件
type OrderEvent struct {
    Action       string `json:"action"`       // create, status_update, cancel
    ProviderName string `json:"providerName"` // grab
    ShopUUID     string `json:"shopUuid"`     // TTPOS 店铺 UUID
    OrderUUID    string `json:"orderUuid"`    // 订单 UUID
    OrderID      string `json:"orderId"`      // 平台订单 ID
    MerchantID   string `json:"merchantId"`   // 商户 ID
    Status       string `json:"status"`       // 当前状态
    Timestamp    int64  `json:"timestamp"`    // 事件时间戳
}
```

### 2. 修改 `grab_order.go` - HandlePushOrderState

```go
// HandlePushOrderState 处理订单状态变更 Webhook
// 签名验证已由中间件完成，此处只处理业务逻辑
// 使用 SDK grabfood.OrderStateRequest
func (s *sGrabOrder) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error {
    // 1. 查询订单
    var order entity.Order
    err := dao.Order.Ctx(ctx).
        Where(dao.Order.Columns().ProviderName, string(consts.ProviderGrab)).
        Where(dao.Order.Columns().ProviderOrderId, req.GetOrderID()).
        Scan(&order)
    if err != nil {
        g.Log().Errorf(ctx, "订单不存在: %s", req.GetOrderID())
        return gerror.Newf("订单不存在: %s", req.GetOrderID())
    }

    // 2. 序列化用于保存原始数据
    rawData, _ := gjson.EncodeString(req)

    // 3. 记录状态变更日志
    logUUID := guid.S()
    var driverEta int
    if req.HasDriverETA() {
        driverEta = int(req.GetDriverETA())
    }

    logDo := &do.OrderStatusLog{
        Uuid:         logUUID,
        OrderUuid:    order.Uuid,
        ProviderName: string(consts.ProviderGrab),
        StatusBefore: order.OrderStatus,
        StatusAfter:  req.GetState(),
        ChangeSource: "WEBHOOK",
        DriverEta:    driverEta,
        Remark:       req.GetMessage(),
        RawData:      rawData,
    }

    _, err = dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert()
    if err != nil {
        g.Log().Errorf(ctx, "插入状态日志失败: %v", err)
        return gerror.Wrap(err, "插入状态日志失败")
    }

    // 4. 更新订单状态
    _, err = dao.Order.Ctx(ctx).
        Where(dao.Order.Columns().Uuid, order.Uuid).
        Data(g.Map{
            dao.Order.Columns().OrderStatus: req.GetState(),
            dao.Order.Columns().UpdatedAt:   gtime.Now(),
        }).Update()
    if err != nil {
        g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
        return gerror.Wrap(err, "更新订单状态失败")
    }

    // 5. 发送 MQ 消息
    event := &OrderEvent{
        Action:       "status_update",
        ProviderName: string(consts.ProviderGrab),
        ShopUUID:     order.ShopUuid,  // 从订单记录获取
        OrderUUID:    order.Uuid,
        OrderID:      req.GetOrderID(),
        MerchantID:   req.GetMerchantID(),
        Status:       req.GetState(),
        Timestamp:    gtime.Now().Unix(),
    }
    if err := queue.PushWithContext(ctx, TopicGrabOrder, event); err != nil {
        g.Log().Warningf(ctx, "发送订单状态更新 MQ 事件失败 %s: %v", order.Uuid, err)
    }

    g.Log().Infof(ctx, "订单状态已更新: %s -> %s (订单ID: %s)", order.OrderStatus, req.GetState(), req.GetOrderID())
    return nil
}
```

### 3. 修改 `grab.go` 代理方法

```go
// HandlePushOrderState 处理订单状态变更 Webhook
// 签名验证已由中间件完成
func (s *sGrab) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error {
    return service.GrabOrder().HandlePushOrderState(ctx, req)
}
```

### 4. 重新生成 service 接口

```bash
cd app/ttpos-takeout && gf gen service
```

### 5. 实现控制器

```go
package grab

import (
    "context"

    v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
    "ttpos-bmp/app/ttpos-takeout/internal/service"
)

// PushOrderState 处理 Grab 订单状态变更 Webhook
// GrabFood 在订单状态变更时调用此端点推送状态通知
func (c *ControllerV1) PushOrderState(ctx context.Context, req *v1.PushOrderStateReq) (res *v1.PushOrderStateRes, err error) {
    // 调用 Service 层处理订单状态变更
    // 签名验证已由中间件完成
    err = service.Grab().HandlePushOrderState(ctx, req.OrderStateRequest)
    if err != nil {
        return nil, err
    }

    // Webhook 成功处理，返回空的响应体
    return &v1.PushOrderStateRes{}, nil
}
```

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：接口调整，无复杂业务逻辑变更

### 工作量预估

- **预计时间**: 0.5 天
- **预估 SP**: 1

### 风险识别

**潜在风险**：
1. 接口变更可能影响其他调用方

**缓解措施**：
1. 通过搜索确认无其他调用方使用旧接口
2. `service/` 目录代码由 GoFrame CLI 自动生成，执行 `gf gen service` 即可更新

---

## 🔗 相关资源

### 参考文件

- API 定义: `api/grab/v1/push_order_state.go`
- 服务实现: `internal/logic/grab_order/grab_order.go`
- 类似控制器: `internal/controller/grab/grab_v1_submit_order.go`

### GrabFood API 文档

- [Order State Webhook](https://developer.grab.com/docs/grabfood/api/#tag/push-order-state-webhook)

---

## 🤝 评审结论

- [x] ✅ **批准**：进入实现阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合需求或优先级

**下一步行动**：

- [ ] 批准后按方案 A 实现
- [ ] 修改 logic 层代码
- [ ] 执行 `gf gen service` 重新生成接口
- [ ] 实现控制器

---

**版本**: v1.0.0  
**创建日期**: 2025-12-19

