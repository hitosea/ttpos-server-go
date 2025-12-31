# Grab 订单状态推送 Webhook 实现 任务分解

> 本文档定义 Grab 订单状态推送 Webhook 控制器的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 7  
**已完成**: 7  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 接口修改

### 任务 1.1: 修改 OrderEvent 结构体

- [x] 1.1 增加 ShopUUID 字段

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - **Purpose**: 扩展 MQ 消息结构，包含门店 UUID
  - **Requirements**: 3.1
  - **Leverage**: 现有 `OrderEvent` 结构体
  - **Changes**:
    ```go
    type OrderEvent struct {
        Action       string `json:"action"`
        ProviderName string `json:"providerName"`
        ShopUUID     string `json:"shopUuid"`     // 新增
        OrderUUID    string `json:"orderUuid"`
        OrderID      string `json:"orderId"`
        MerchantID   string `json:"merchantId"`
        Status       string `json:"status"`
        Timestamp    int64  `json:"timestamp"`
    }
    ```

---

### 任务 1.2: 修改 HandlePushOrderState 方法签名

- [x] 1.2 修改 grab_order.go 中的方法

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - **Purpose**: 统一接口风格，接收类型化请求
  - **Requirements**: 2.1
  - **Leverage**: `HandleSubmitOrder` 方法模式
  - **Changes**:
    - 修改方法签名：`func (s *sGrabOrder) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error`
    - 移除 JSON 解析逻辑（GoFrame 已自动解析）
    - 使用 `gjson.EncodeString(req)` 生成 rawData
    - MQ 事件增加 `ShopUUID: order.ShopUuid`
    - 使用 `req.GetMerchantID()` 替代原来的 `req.GetPartnerMerchantID()`

---

### 任务 1.3: 修改 grab.go 代理方法

- [x] 1.3 修改代理方法签名

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go`
  - **Purpose**: 同步更新代理方法
  - **Requirements**: 2.2
  - **Leverage**: 现有代理方法
  - **Changes**:
    ```go
    func (s *sGrab) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error {
        return service.GrabOrder().HandlePushOrderState(ctx, req)
    }
    ```

---

### 任务 1.4: 重新生成 Service 接口

- [x] 1.4 执行 gf gen service

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/service/`
  - **Purpose**: 自动更新 Service 接口定义
  - **Requirements**: 2.3
  - **Command**:
    ```bash
    cd ttpos-bmp/app/ttpos-takeout && gf gen service
    ```
  - **Success**: `internal/service/grab.go` 和 `internal/service/grab_order.go` 接口签名已更新

---

## Phase 2: 控制器实现

### 任务 2.1: 实现 PushOrderState 控制器

- [x] 2.1 实现控制器方法

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_push_order_state.go`
  - **Purpose**: 接收 Grab Webhook 并调用 Service
  - **Requirements**: 1.1, 1.2, 1.3
  - **Leverage**: `grab_v1_submit_order.go` 实现模式
  - **Code**:
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

## Phase 3: 同步更新（可选）

### 任务 3.1: 更新 HandleSubmitOrder 中的 MQ 消息

- [x] 3.1 同步增加 ShopUUID 字段

  - **File**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - **Purpose**: 保持 `HandleSubmitOrder` 与 `HandlePushOrderState` 的 MQ 消息格式一致
  - **Requirements**: 3.3
  - **Leverage**: 任务 1.1 的 OrderEvent 结构体
  - **Changes**: 在 `HandleSubmitOrder` 中的 OrderEvent 增加 `ShopUUID` 字段
  - **Note**: 需要从 `saveOrderFromSDK` 返回 `shopUuid` 或重新查询

---

## Phase 4: 验证测试

### 任务 4.1: 本地验证

- [x] 4.1 编译验证和本地测试

  - **File**: -
  - **Purpose**: 确保代码编译通过，功能正常
  - **Requirements**: 所有
  - **Commands**:
    ```bash
    cd ttpos-bmp/app/ttpos-takeout
    go build ./...
    go vet ./...
    ```
  - **Success**: 编译无错误，vet 无警告

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go build` 和 `go vet`
- [ ] Service 接口已通过 `gf gen service` 重新生成

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/logic/grab_order/grab_order.go` | 修改 | OrderEvent 增加 ShopUUID，HandlePushOrderState 签名变更 |
| `internal/logic/grab/grab.go` | 修改 | 代理方法签名变更 |
| `internal/service/grab.go` | 自动生成 | 接口签名更新 |
| `internal/service/grab_order.go` | 自动生成 | 接口签名更新 |
| `internal/controller/grab/grab_v1_push_order_state.go` | 修改 | 实现控制器 |

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-takeout-grab-push-order-state/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-takeout-grab-push-order-state/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^\- \[x\]" docs/shared/specs/active/task-takeout-grab-push-order-state/tasks.md) * 100 / $(grep -c "^\- \[" docs/shared/specs/active/task-takeout-grab-push-order-state/tasks.md)" | bc
```

### 执行顺序

1. **任务 1.1** → 修改 OrderEvent 结构体
2. **任务 1.2** → 修改 HandlePushOrderState 方法签名
3. **任务 1.3** → 修改代理方法
4. **任务 1.4** → 执行 gf gen service
5. **任务 2.1** → 实现控制器
6. **任务 3.1** → （可选）同步更新 HandleSubmitOrder
7. **任务 4.1** → 本地验证

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-19  
**维护者**: AI Assistant

