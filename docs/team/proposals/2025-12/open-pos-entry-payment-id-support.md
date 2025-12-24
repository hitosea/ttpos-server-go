# 开账接口支持 PaymentID 提案

## 📋 基本信息

| 项目         | 内容                                                               |
| ------------ | ------------------------------------------------------------------ |
| **提案人**   | rikugun                                                            |
| **创建日期** | 2025-12-24                                                         |
| **状态**     | 进行中                                                             |
| **优先级**   | 中                                                                 |
| **Story Point** | 2                                                                |
| **涉及模块** | ttpos-erp (Go BMP)                                                 |
| **关联 Spec** | [task-erp-open-pos-entry-payment-id](../../../shared/specs/active/task-erp-open-pos-entry-payment-id/) |
| **关联任务** | -                                                                 |

---

## 📝 问题描述

### 当前现状

在 `OpenPosEntry` 接口中，`OpenPosEntryDetail` 消息只支持通过 `mode_of_payment`（支付方式名称）来指定支付方式。这与 ERPNext 内部的命名规则强耦合，调用方需要：

1. 先通过 `GetModeOfPayment` 接口查询支付方式的 `name` 字段
2. 再将查询到的 `name` 传递给 `OpenPosEntry` 接口

### 存在问题

1. **额外的查询开销**：调用方必须先调用 `GetModeOfPayment` 获取 `mode_of_payment` 名称
2. **接口设计不一致**：`SavePosInvoice`、`ClosePosEntry` 等接口已支持 `payment_id`，但 `OpenPosEntry` 仍只支持 `mode_of_payment`
3. **命名耦合度高**：调用方需要关心 ERP 内部的支付方式命名规则

### 业务价值

- **简化调用流程**：调用方可以直接使用 `payment_id`，无需额外查询
- **统一接口设计**：与 `SavePosInvoice`、`ClosePosEntry` 等接口保持一致
- **降低耦合度**：减少调用方对 ERP 内部实现细节的依赖
- **提升开发效率**：减少接口调用次数，简化业务逻辑

---

## 💡 解决方案概述

### 核心思路

在 `OpenPosEntryDetail` 消息中增加 `payment_id` 可选字段，并将 `mode_of_payment` 改为可选。当调用方提供 `payment_id` 时，系统自动查询对应的 `mode_of_payment` 完成开账操作。

### 技术实现

**Protobuf 定义调整**：
```protobuf
message OpenPosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一（必填其中之一）
  double opening_amount = 2; // 开帐金额,必填
  optional string payment_id = 3; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
  // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
}
```

**Logic 层处理逻辑**（伪代码）：
```go
// 处理 OpenPosEntryDetail
for _, detail := range req.OpenPosEntryDetail {
    // 参数校验
    if detail.PaymentId == "" && detail.ModeOfPayment == "" {
        return gerror.New("payment_id 和 mode_of_payment 不能同时为空")
    }

    modeOfPayment := detail.ModeOfPayment

    // 如果提供了 payment_id，自动查询 mode_of_payment
    if detail.PaymentId != "" {
        resp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
            PaymentId: detail.PaymentId,
        })
        if err != nil {
            return gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", detail.PaymentId)
        }
        modeOfPayment = resp.Name
    }

    // 使用 modeOfPayment 进行后续处理
    // ...
}
```

---

## 🎯 核心功能点

### 1. Protobuf 定义调整

- [x] 将 `OpenPosEntryDetail.mode_of_payment` 改为 `optional string`
- [x] 新增 `OpenPosEntryDetail.payment_id` 字段（`optional string`，字段编号 3）
- [x] 添加字段注释说明两个字段的关系

### 2. 参数校验

- [x] `payment_id` 和 `mode_of_payment` 不能同时为空
- [x] 返回明确的错误信息

### 3. 自动查询逻辑

- [x] `payment_id` 不为空时，自动调用 `GetModeOfPayment` 查询 `mode_of_payment`
- [x] 查询失败时返回详细错误信息（包含 `payment_id`）
- [x] 支付方式未启用时返回明确错误

### 4. 向后兼容

- [x] 保持原有 `mode_of_payment` 字段可用
- [x] 仅提供 `mode_of_payment` 的调用方式继续有效

---

## 📊 影响评估

### 影响范围

| 层级          | 影响组件                                                      | 影响类型 |
| ------------- | ------------------------------------------------------------- | -------- |
| **Protobuf**  | `selling.proto` - `OpenPosEntryDetail` 消息定义                | 修改     |
| **Controller** | `selling.go` - 参数校验逻辑                                    | 新增     |
| **Logic**     | `selling.go` - `OpenPosEntry` 方法中的自动查询逻辑             | 新增     |
| **Service**   | 复用现有 `GetModeOfPayment` 服务                               | 无变更   |
| **调用方**    | Main 模块、其他需要开账的模块                                  | 可选升级 |

### 风险评估

| 风险               | 影响 | 概率 | 缓解措施                                    |
| ------------------ | ---- | ---- | ------------------------------------------- |
| 向后兼容性破坏     | 高   | 低   | 保持 `mode_of_payment` 字段可用，充分测试   |
| 查询性能影响       | 中   | 低   | `GetModeOfPayment` 内部已有缓存机制         |
| 参数混用导致混乱   | 中   | 中   | 明确优先级（优先使用 `payment_id`），记录日志 |

---

## 🔍 技术细节

### 数据流

```
调用方
  ↓ (提供 payment_id 或 mode_of_payment)
OpenPosEntry Controller (参数校验)
  ↓
OpenPosEntry Logic
  ↓ (如果有 payment_id)
GetModeOfPayment Service (查询 mode_of_payment)
  ↓
继续开账流程 (使用 mode_of_payment)
  ↓
ERPNext API
```

### 错误处理

| 错误场景                           | 错误信息示例                                       |
| ---------------------------------- | -------------------------------------------------- |
| 两个参数都为空                     | `payment_id 和 mode_of_payment 不能同时为空`       |
| `payment_id` 查询失败              | `查询支付方式失败，payment_id: PID123456`          |
| 支付方式不存在或未启用             | `支付方式不存在或未启用，payment_id: PID123456`    |

### 日志记录

```go
// 成功查询
g.Log().Info(ctx, "开账详情: 通过 payment_id 查询到 mode_of_payment",
    g.Map{"index": i, "payment_id": detail.PaymentId, "mode_of_payment": modeOfPayment})

// 查询失败
g.Log().Error(ctx, "查询支付方式失败",
    g.Map{"payment_id": detail.PaymentId, "error": err.Error()})
```

---

## ✅ 验收标准

### 功能验收

1. **Protobuf 定义**
   - `OpenPosEntryDetail` 包含 `optional string payment_id` 字段
   - `mode_of_payment` 改为 `optional string`
   - 字段注释完整

2. **参数校验**
   - 两个字段同时为空时返回错误
   - 错误信息包含 detail 的索引

3. **自动查询**
   - 提供 `payment_id` 时能成功查询 `mode_of_payment`
   - 查询失败时返回明确错误
   - 日志记录完整

4. **向后兼容**
   - 仅提供 `mode_of_payment` 的调用方式继续有效
   - 不影响现有业务

### 测试验收

1. **单元测试**
   - Logic 层测试覆盖率 ≥ 80%
   - 覆盖参数校验、自动查询、错误处理等场景

2. **集成测试**
   - 使用 `payment_id` 开账成功
   - 使用 `mode_of_payment` 开账成功（向后兼容）
   - 参数都为空时返回错误
   - `payment_id` 无效时返回错误

3. **手动测试**
   - 使用 grpcurl 或 BloomRPC 测试各种场景

---

## 📅 初步排期

| 阶段                  | 预估时间 | 说明                                          |
| --------------------- | -------- | --------------------------------------------- |
| Protobuf 定义调整     | 0.5 天   | 修改 proto 文件，重新生成代码                 |
| Logic 层实现          | 0.5 天   | 参数校验、自动查询逻辑、错误处理              |
| 测试与文档            | 0.5 天   | 单元测试、集成测试、API 文档                  |
| **总计**              | **1.5 天** | **SP = 2**                                  |

---

## 🔗 参考资料

### 相关 Proposal

- `docs/team/proposals/2025-12/close-pos-entry-payment-id-support.md` - 关账接口 PaymentID 支持（已完成）
- `docs/team/proposals/2025-12/payment-id-update-logic.md` - 支付方式 PaymentID 更新逻辑

### 相关 Spec

- `docs/shared/specs/active/task-erp-close-pos-entry-payment-id/` - 关账接口 PaymentID 支持 Spec

### 相关代码

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - Logic 实现
- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go` - Controller 实现

### 规范文档

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
- `.cursor/rules/api.mdc` - API 设计规范

---

## 💬 Scrum 评审清单

### 产品价值（PM 必填）

- [ ] 该功能是否解决了明确的业务痛点？
  - ✅ 是，简化了开账接口的调用流程
- [ ] 是否对核心业务流程有正向影响？
  - ✅ 是，提升开发效率，统一接口设计
- [ ] 用户故事是否清晰？
  - ✅ 是，作为后端开发者，直接使用 `payment_id` 开账

### 技术可行性（Tech Lead 必填）

- [x] 技术方案是否合理？
  - ✅ 是，与已有的 `ClosePosEntry` 实现保持一致
- [x] 是否有技术风险？
  - ⚠️ 低风险，需确保向后兼容性
- [x] 是否需要额外的基础设施支持？
  - ✅ 否，复用现有服务

### 资源评估（Scrum Master 必填）

- [ ] Story Point 评估是否合理（≤ 5）？
  - ✅ SP = 2，合理
- [ ] 是否在当前 Sprint 资源范围内？
  - ⏳ 待确认
- [ ] 是否有依赖阻塞？
  - ✅ 无

### 质量保证（QA 必填）

- [ ] 是否有明确的验收标准？
  - ✅ 是
- [ ] 测试策略是否完整？
  - ✅ 是，包含单元测试、集成测试、手动测试
- [ ] 是否需要额外的测试环境？
  - ✅ 否

---

## ✍️ 评审结果

| 项目         | 内容      |
| ------------ | --------- |
| **评审状态** | 已通过    |
| **评审人**   | rikugun   |
| **评审日期** | 2025-12-24 |
| **评审意见** | 技术方案与 ClosePosEntry 一致，可行性高 |
| **下一步**   | ✅ Spec 已创建：[task-erp-open-pos-entry-payment-id](../../../shared/specs/active/task-erp-open-pos-entry-payment-id/)<br>⏳ 等待产品审核 requirements.md |

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-24  
**最后更新**: 2025-12-24  
**作者**: rikugun


