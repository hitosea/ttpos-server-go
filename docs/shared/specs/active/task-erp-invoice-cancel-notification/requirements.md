# ERP 发票取消通知 需求文档

## 📋 基本信息

| 项目              | 内容                                                                               |
| ----------------- | ---------------------------------------------------------------------------------- |
| **Spec ID**       | task-erp-invoice-cancel-notification                                               |
| **来源 Proposal** | [erp-invoice-cancel-notification](../../../team/proposals/2026-01/erp-invoice-cancel-notification.md) |
| **创建日期**      | 2026-01-29                                                                         |
| **负责人**        | rikugun                                                                            |
| **目标版本**      | v2.16                                                                              |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已通过     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-29 |

---

## 📝 用户故事

**作为** 外部 ERP 系统 / 内部调用方
**我想** 在调用 ReturnPosInvoiceReq 退票接口时传入附注信息，并在退票处理完成后收到 MQ 通知
**以便于** 及时同步发票取消状态、追溯退票原因、减少人工核对成本

---

## 功能需求

### Requirement 1: 扩展 ReturnPosInvoiceReq 结构

**用户故事**: 作为调用方，我想在退票请求中传入附注信息，以便于记录退票原因

#### 验收标准

1. **WHEN** 调用 ReturnPosInvoiceReq 接口并传入 remark 字段 **THEN** 系统 **SHALL** 正确接收并处理该字段
2. **WHEN** 未传入 remark 字段 **THEN** 系统 **SHALL** 正常处理请求（remark 为可选字段）
3. **IF** remark 字段为空字符串 **THEN** 系统 **SHALL** 正常处理且消息中 remark 为空

---

### Requirement 2: 发送 MQ 消息通知

**用户故事**: 作为外部 ERP 系统，我想订阅发票取消事件，以便于及时同步发票状态

#### 验收标准

1. **WHEN** 退票消费逻辑处理完成 **THEN** 系统 **SHALL** 发送 MQ 消息到 `erp-invoice-cancel` topic
2. **WHEN** MQ 消息发送 **THEN** 消息体 **SHALL** 包含以下字段：
   - `remark`: 附注信息（可为空）
   - `order_no`: 订单号
   - `invoice_name`: 发票名称
3. **IF** MQ 发送失败 **THEN** 系统 **SHALL** 记录错误日志，但不阻塞主流程

---

### Requirement 3: 消息格式规范

**用户故事**: 作为消费方，我想获取结构化的消息，以便于快速解析和处理

#### 消息格式

```json
{
  "order_no": "ORD20260129001",
  "invoice_name": "发票名称",
  "remark": "用户请求退票：商品质量问题"
}
```

#### 验收标准

1. **WHEN** 消息发送 **THEN** 消息格式 **SHALL** 符合上述 JSON 结构
2. **WHEN** remark 为空 **THEN** 消息中 remark 字段值 **SHALL** 为空字符串

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 包含 MQ 消息发送的集成测试

### 平台兼容性

- [x] BMP 后端服务（GoFrame v2.x）
- [x] RocketMQ 消息队列

### 日志要求

- [ ] MQ 发送成功/失败均需记录日志
- [ ] 日志包含 company_uuid 字段

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 分层架构: Controller → Logic → DAO
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/go-rules.mdc 规范
- 使用 `gerror` 处理错误（不用标准库 errors）
- MQ 发送采用异步方式，不阻塞主流程

### 资源约束

- Story Point: 2-3（待技术评审确认）

---

## 风险和缓解

### 风险 1: MQ 消息发送失败

**影响**: 中
**缓解措施**: MQ 发送采用异步方式，失败时记录日志便于排查，不阻塞退票主流程

### 风险 2: 消费者未正确订阅 topic

**影响**: 中
**缓解措施**: 提供清晰的接入文档，说明 topic 名称和消息格式

---

**版本**: v1.0.0
**创建日期**: 2026-01-29
