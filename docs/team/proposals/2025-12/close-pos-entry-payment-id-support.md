# 关账接口支持 PaymentID 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-24   |
| **目标版本** | v2.10.0 |
| **状态**   | 进行中   |
| **关联任务** | - |
| **关联 Spec** | [task-erp-close-pos-entry-payment-id](../../shared/specs/active/task-erp-close-pos-entry-payment-id/) |

---

## 🎯 背景和动机

### 问题描述

当前 `ClosePosEntry` 接口的 `ClosePosEntryDetail` 消息中，支付方式仅支持通过 `mode_of_payment` 字段传入。随着支付方式管理的升级，系统引入了 `payment_id` 作为支付方式的唯一标识（PaymentID）。

现有问题：
1. 关账接口无法直接使用 `payment_id` 进行支付方式识别
2. 调用方需要先查询 `mode_of_payment` 再传入，增加了额外的接口调用
3. 与其他接口（如 `SavePosInvoice`）的设计不一致，后者已支持 `payment_id`

### 业务价值

- **简化调用流程**：调用方可直接使用 `payment_id`，无需额外查询
- **提升一致性**：与 `SavePosInvoice` 等接口保持一致的设计模式
- **增强灵活性**：支持两种方式传入支付方式，兼容旧有调用方式
- **降低耦合度**：调用方无需关心 ERP 内部的 `mode_of_payment` 命名规则

### 目标用户

- [x] 后端开发者（调用 ttpos-erp gRPC 服务）
- [x] 前端开发者（通过 Main 模块调用）
- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

在 `ClosePosEntryDetail` 消息中新增 `payment_id` 字段（可选），并调整 `mode_of_payment` 为可选字段。两个字段至少需要提供一个：

- **当 `payment_id` 不为空时**：Logic 服务自动调用 `GetModeOfPayment` 通过 `payment_id` 查询对应的 `mode_of_payment`
- **当 `payment_id` 为空时**：直接使用 `mode_of_payment` 字段（保持向后兼容）
- **参数校验**：`payment_id` 和 `mode_of_payment` 不能同时为空

### 核心功能点

1. **Protobuf 定义调整**
   - `ClosePosEntryDetail.mode_of_payment` 改为 `optional string`
   - 新增 `ClosePosEntryDetail.payment_id` 字段（`optional string`）

2. **Logic 层逻辑增强**
   - 参数校验：确保 `payment_id` 和 `mode_of_payment` 至少提供一个
   - 自动查询：当 `payment_id` 不为空时，调用 `GetModeOfPayment` 获取 `mode_of_payment`
   - 错误处理：查询失败时返回明确的错误信息

3. **向后兼容**
   - 保持原有 `mode_of_payment` 字段可用
   - 不影响现有调用方

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（Protobuf 定义）
- [ ] 数据模型
- [x] 业务逻辑（Logic 层）
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1 天
- **预估 SP**: 2（待技术评审确认）

### 风险识别

**潜在风险**：
1. **向后兼容性**：需确保现有调用方不受影响
2. **查询失败处理**：`payment_id` 查询不到对应的 `mode_of_payment` 时的错误处理

**缓解措施**：
1. 保持 `mode_of_payment` 字段可用，仅将其改为可选
2. 在 Logic 层添加明确的错误提示，指导调用方正确传参
3. 添加单元测试覆盖各种参数组合场景

---

## 🔗 相关资源

### 参考需求

- 类似功能: `SavePosInvoice` 接口已支持 `payment_id` 和 `mode_of_payment` 二选一
- 相关提案: `docs/team/proposals/2025-12/payment-id-update-logic.md`

### 相关文档

- Protobuf 文件: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- Logic 服务: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
- 支付方式管理: `GetModeOfPayment` 服务

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | -      |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-erp-close-pos-entry-payment-id` (已完成 2025-12-24)
- [ ] 分配负责人：-
- [ ] 目标 Sprint：Sprint -
- [ ] 产品审核：requirements.md
- [ ] 技术设计：执行 `/spec-design task-erp-close-pos-entry-payment-id`

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发者  
**我想** 在调用 `ClosePosEntry` 接口时直接使用 `payment_id`  
**以便于** 简化调用流程，无需额外查询 `mode_of_payment`

### AC 验收标准（初稿）

1. **WHEN** 调用 `ClosePosEntry` 接口且 `payment_id` 不为空 **THEN** 系统 **SHALL** 自动查询对应的 `mode_of_payment` 并完成关账
2. **WHEN** 调用 `ClosePosEntry` 接口且仅提供 `mode_of_payment` **THEN** 系统 **SHALL** 直接使用该值完成关账（向后兼容）
3. **IF** `payment_id` 和 `mode_of_payment` 同时为空 **THEN** 系统 **SHALL** 返回参数错误
4. **IF** `payment_id` 查询不到对应的支付方式 **THEN** 系统 **SHALL** 返回明确的错误信息

### 技术实现要点

#### Protobuf 调整

```protobuf
message ClosePosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一
  double opening_amount = 2; // 开帐金额,必填
  double closing_amount = 3; // 关帐金额,必填
  optional string payment_id = 4; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
}
```

#### Logic 层伪代码

```go
// 处理 ClosePosEntryDetail
for _, detail := range req.ClosePosEntryDetail {
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
        modeOfPayment = resp.ModeOfPayment.Name
    }
    
    // 使用 modeOfPayment 进行后续处理
    // ...
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/go-bmp.mdc`, `.cursor/rules/proto-rules.mdc`

