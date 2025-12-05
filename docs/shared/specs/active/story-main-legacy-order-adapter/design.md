# Legacy Order Adapter 设计文档

> 本文档定义 Legacy Order Adapter 的技术设计和实现方案。

## 📋 概述

本设计旨在通过适配器模式 (Adapter Pattern) 将旧的 TTPOS 订单服务 (SaleOrderService) 迁移到新的 Order Core 模块。核心思想是保持上层 API 调用不变，底层通过 `LegacyOrderAdapter` 将请求转发给 `OrderCoreService`，并利用领域事件解耦积分、库存、打印等周边业务。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- **分层设计**: Adapter 属于 Service 层的一种特殊形式，位于 `modules/order_core/adapter`，依赖 `modules/order_core/service`。
- **接口隔离**: 定义 `ILegacyOrderAdapter` 接口，确保依赖倒置。
- **错误处理**: 适配器负责将 Core 层的错误转换为旧业务层兼容的错误码。

### API 设计规范 (api.mdc)

- 本次设计主要涉及内部 Service 适配，暂不涉及新增对外 API，但现有 API 的响应结构保持不变。

### 数据库规范 (database.mdc)

- **数据共存**: 新旧模块共用 `ttpos_sale_order` 等表，通过字段映射保证兼容性。
- **Model 复用**: Adapter 内部可能需要同时操作旧 Model (为了兼容返回值) 和新 Model (为了调用 Core)。

---

## 🔄 代码复用分析

### 可复用的现有组件

- **OrderCoreService**: `main/app/modules/order_core/service/core_order_service.go` - 核心订单能力提供者。
- **EventBus**: `main/pkg/eventbus/` - 用于发布和订阅领域事件。
- **SaleOrderRepo**: `main/app/repository/sale_order_repo.go` - 旧的订单仓储，Adapter 可能需要用它来组装旧格式的返回值。

### 集成点

- **Service 层**: 旧的 `SaleOrderService` 将被 `LegacyOrderAdapter` 替换或在其内部调用 Adapter。
- **Event Listeners**: 现有的业务逻辑（积分、打印）将被重构为 Event Listener，监听 `OrderCore` 发出的事件。

---

## 🏗️ 架构设计

### 架构图

```mermaid
graph TD
    A[API / Old Service Caller] --> B[LegacyOrderAdapter]
    B --> C[OrderCoreService]
    C --> D[CoreOrderRepo]
    D --> E[Database (Shared Tables)]
    
    C -- Publishes --> F[Event Bus]
    F -- Triggers --> G[InventoryListener]
    F -- Triggers --> H[MemberListener]
    F -- Triggers --> I[PrinterListener]
```

### 模块划分

#### Go Main 模块

- **Adapter 层**: `main/app/modules/order_core/adapter/`
  - `legacy_order_service.go`: 实现旧订单服务的接口适配。
- **Event Listeners**: `main/app/modules/order_core/listener/`
  - `inventory_listener.go`: 监听 OrderCreated/Paid，处理库存。
  - `member_listener.go`: 监听 OrderPaid，处理积分。
  - `printer_listener.go`: 监听 OrderPaid，处理打印。

---

## 🗄️ 数据库设计

### 数据表设计

本方案**不新增数据表**，而是基于现有的 `ttpos_sale_order` 和 `ttpos_sale_order_product` 表。
新模块 `order_core` 的 Model (`CoreSaleOrder`) 已经映射到了这些表。

### 数据兼容性

- **CoreSaleOrder** 包含核心字段：`id`, `uuid`, `total_amount`, `pay_amount`, `status` 等。
- **SaleOrder** (旧) 包含所有字段。
- **策略**: Adapter 在写入时，通过 `OrderCoreService` 写入核心字段；如果旧业务需要写入额外字段（如 `device_id`），需要在 Core 层的 DTO 中预留或通过扩展字段处理。目前假设 Core 层的 DTO 已经足够覆盖核心交易所需的字段。

---

## 📊 数据模型

### DTO 转换

需要实现 `SaleOrder` (旧 Model) 与 `CoreSaleOrder` (新 Model) 及其 DTO 之间的转换函数。

```go
// main/app/modules/order_core/adapter/converter.go

func ToCoreCreateReq(oldReq *dto.CreateOrderReq) *dto_core.CreateOrderReq {
    // ... 映射逻辑
}

func ToOldOrderResp(coreOrder *model_core.CoreSaleOrder) *dto.OrderResp {
    // ... 映射逻辑
}
```

---

## 🧩 组件和接口

### Adapter 接口

```go
// main/app/modules/order_core/adapter/i_legacy_order_service.go
type ILegacyOrderAdapter interface {
    CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.OrderResp, error)
    PayOrder(ctx context.Context, req *dto.PayOrderReq) error
    CancelOrder(ctx context.Context, req *dto.CancelOrderReq) error
}
```

### Adapter 实现

```go
// main/app/modules/order_core/adapter/legacy_order_service.go
type LegacyOrderAdapter struct {
    coreSrv service.ICoreOrderService
    // 可能还需要旧的 Repo 来补充一些非核心数据的读取
}

func (a *LegacyOrderAdapter) CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.OrderResp, error) {
    // 1. DTO 转换 Old -> Core
    coreReq := ToCoreCreateReq(req)
    
    // 2. 调用 Core Service
    coreResp, err := a.coreSrv.CreateOrder(ctx, coreReq)
    if err != nil {
        return nil, ConvertError(err)
    }
    
    // 3. DTO 转换 Core -> Old (可能需要查库补全信息)
    return ToOldOrderResp(coreResp), nil
}
```

---

## ⚡ 缓存设计

- 继续沿用 `OrderCore` 的缓存策略。
- Adapter 层本身不处理缓存，直接依赖 Core Service 的能力。

---

## 🚨 错误处理

- **错误码映射**: Core Service 返回的标准 Error 需要转换为旧业务层使用的 `constant.Code*` 或自定义错误结构。

---

## 🧪 测试策略

### 单元测试

- **Adapter 测试**: Mock `ICoreOrderService`，验证 Adapter 的参数转换和错误处理逻辑。
- **Listener 测试**: 验证 Listener 能正确处理 Event 并调用下游服务（如库存服务）。

### 集成测试

- 构造一个完整的下单请求，通过 `LegacyOrderAdapter` 入口，验证：
  1. 数据库是否正确写入。
  2. 是否触发了 `OrderCreated` 事件。
  3. 库存是否被预占（通过 Listener）。

---

## 📚 实现清单

### Phase 1: Adapter 基础框架

- [ ] 定义 `ILegacyOrderAdapter` 接口
- [ ] 实现 DTO 转换工具函数
- [ ] 实现 `LegacyOrderAdapter` 的 Create/Pay/Cancel 方法

### Phase 2: 事件监听器迁移

- [ ] 实现 `InventoryListener` (库存)
- [ ] 实现 `MemberListener` (积分)
- [ ] 实现 `PrinterListener` (打印)
- [ ] 在 SystemBus 中注册这些 Listener

### Phase 3: 集成验证

- [ ] 编写集成测试验证全流程
- [ ] 替换旧 API 的入口为 Adapter (灰度或开关控制)

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0
**创建日期**: 2025-12-05
**作者**: xiezhihuan
**审核者**: 待定

