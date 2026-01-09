# 优化 RedoPosConsumer 增加 SiteCode 过滤 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-01   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [task-bmp-redo-pos-consumer-site-code-filter](../../../shared/specs/archived/v2.12/task-bmp-redo-pos-consumer-site-code-filter/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前 `RedoPosConsumer` 在处理重做消息时，查询未处理的订单时只使用了 `OpenPosEntryName` 和 `Docstatus` 作为过滤条件，缺少 `SiteCode` 过滤。

**问题场景**：
> 在多站点（多商户）环境下，当某个站点的 POS 开单需要重做时，`RedoPosConsumer` 可能会误查询到其他站点的未处理订单，导致数据混乱和业务错误。

**代码现状**：
```go
// 当前代码只使用 OpenPosEntryName 和 Docstatus 过滤
posInvoiceDao.Where(do.ReceivePosInvoice{
    OpenPosEntryName: msg.PosOpenEntryName,
    Docstatus:        erp.DocstatusDraft,
}).Scan(&posInvoiceList)
```

**对比其他 Consumer**：
- `SavePosInvoiceConsumer`、`ReturnPosInvoiceConsumer`、`CancelPosInvoice`、`ClosePosEntryConsumer` 等都在查询时使用了 `SiteCode` 过滤
- `AsyncSellingMsg` 结构体中已经包含 `SiteCode` 字段

### 业务价值

- **数据隔离**：确保多站点环境下数据查询的准确性，避免跨站点数据污染
- **业务安全**：防止误操作其他站点的订单，降低业务风险
- **代码一致性**：与其他 Consumer 保持一致的过滤逻辑，提高代码可维护性
- **系统稳定性**：减少因数据查询错误导致的系统异常

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [x] 系统运维人员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: 开发人员

---

## 💡 解决方案概述

### 方案描述

在 `RedoPosConsumer` 的 `Handle` 方法中，为所有查询操作增加 `SiteCode` 过滤条件，确保只查询当前站点的未处理订单。

**实现要点**：
1. 从消息中获取 `SiteCode`（`msg.SiteCode`）
2. 在所有查询条件中添加 `SiteCode` 过滤
3. 处理 `SiteCode` 为空的情况（向后兼容）

### 核心功能点

1. **增加 SiteCode 过滤**：在 `RedoPosConsumer.Handle` 方法的所有查询操作中添加 `SiteCode` 过滤条件
2. **消息验证**：验证消息中的 `SiteCode` 字段是否有效
3. **向后兼容**：当 `SiteCode` 为空时，保持原有查询逻辑（可选）

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: 消息队列 Consumer

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯业务逻辑调整，无架构变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 0.5 天
- **预估 SP**: 1 SP（待技术评审确认）

### 风险识别

**潜在风险**：
1. **向后兼容性**：如果历史消息中没有 `SiteCode` 字段，可能导致查询失败
2. **消息发送方**：需要确保发送重做消息时包含 `SiteCode` 字段

**缓解措施**：
1. **兼容处理**：当 `SiteCode` 为空时，可以跳过该过滤条件（保持原有行为）或记录警告日志
2. **消息验证**：在消息处理前验证 `SiteCode` 是否存在，如果不存在则记录警告日志
3. **测试覆盖**：增加单元测试和集成测试，覆盖多站点场景

---

## 🔗 相关资源

### 参考需求

- 类似功能: `SavePosInvoiceConsumer`、`ReturnPosInvoiceConsumer`、`CancelPosInvoice`、`ClosePosEntryConsumer` 的 SiteCode 过滤实现
- 竞品分析: -

### 相关文档

- 代码位置: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
- 消息结构: `ttpos-bmp/app/ttpos-erp/internal/model/mq/async_selling.go`
- 相关规范: `.cursor/rules/go-bmp.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-bmp-redo-pos-consumer-site-code-filter`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 系统运维人员  
**我想** 在重做 POS 订单时只处理当前站点的订单  
**以便于** 避免跨站点数据污染和业务错误

### AC 验收标准（初稿）

1. **WHEN** 收到重做消息且消息包含 `SiteCode` **THEN** 系统 **SHALL** 只查询该站点的未处理订单
2. **IF** 消息中的 `SiteCode` 为空 **THEN** 系统 **SHALL** 记录警告日志并跳过 SiteCode 过滤（向后兼容）
3. **WHEN** 查询未处理的商品发票 **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
4. **WHEN** 查询未处理的取消发票 **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
5. **WHEN** 查询未处理的退货发票 **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
6. **WHEN** 查询未处理的关账记录 **THEN** 系统 **SHALL** 使用 `PosOpenEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件

### 线框图/原型（可选）

[代码修改示例]

```go
// 修改前
posInvoiceDao.Where(do.ReceivePosInvoice{
    OpenPosEntryName: msg.PosOpenEntryName,
    Docstatus:        erp.DocstatusDraft,
}).Scan(&posInvoiceList)

// 修改后
whereCondition := do.ReceivePosInvoice{
    OpenPosEntryName: msg.PosOpenEntryName,
    Docstatus:        erp.DocstatusDraft,
}
if msg.SiteCode != "" {
    whereCondition.SiteCode = msg.SiteCode
}
posInvoiceDao.Where(whereCondition).Scan(&posInvoiceList)
```

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/go-bmp.mdc`

