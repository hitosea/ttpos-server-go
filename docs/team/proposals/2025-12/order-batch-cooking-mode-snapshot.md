# 订单分批送厨模式快照 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-02   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-main-order-batch-cooking-mode-snapshot](../../shared/specs/archived/v2.12/story-main-order-batch-cooking-mode-snapshot/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前分批送厨模式（`BatchCookingMode`）存储在门店业务设置（`business_setting`）中，是一个全局配置。当 shop 端修改分批送厨模式后，所有订单（包括已创建的订单）都会受到影响，导致：

1. **已创建订单行为不一致**：订单创建时使用的是"前置"模式，但 shop 端修改为"后置"模式后，该订单的送厨行为可能发生变化
2. **历史订单追溯困难**：无法准确知道某个订单创建时使用的分批送厨模式
3. **业务逻辑混乱**：订单创建后的送厨行为应该保持创建时的模式，而不应受后续配置变更影响

**示例场景**：
> 门店在上午 10:00 创建订单时，分批送厨模式为"前置"（pre），订单按前置模式送厨。下午 14:00 shop 端将模式修改为"后置"（post），此时已创建的订单不应该受到影响，仍应保持"前置"模式。

### 业务价值

- **保证订单一致性**：订单创建时的配置快照，确保订单在整个生命周期内行为一致
- **提升可追溯性**：可以准确查询历史订单创建时的分批送厨模式
- **避免业务风险**：防止因配置变更导致已创建订单的行为异常
- **符合业务逻辑**：订单配置应该在创建时确定，后续不应受全局配置影响

### 目标用户

- [x] 收银员
- [x] 商户管理员
- [x] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

在 `sale_bill_setting`（销售账单设置）表中新增 `batch_cooking_mode` 字段，用于记录订单创建时的分批送厨模式。订单创建时，从 `business_setting` 中读取当前的分批送厨模式并保存到 `sale_bill_setting` 中，后续订单的送厨逻辑使用 `sale_bill_setting` 中保存的值，不再读取全局配置。

**核心原则**：
- 订单创建时快照：创建订单时，将当前 `business_setting.BatchCookingMode` 的值保存到 `sale_bill_setting.batch_cooking_mode`
- 订单使用快照值：订单的送厨逻辑使用 `sale_bill_setting.batch_cooking_mode`，不再读取全局配置
- 新订单才生效：只有新创建的订单才会使用新的分批送厨模式，已创建的订单保持原有模式

### 核心功能点

1. **数据库字段扩展**
   - 在 `ttpos_sale_bill_setting` 表中新增 `batch_cooking_mode` 字段（VARCHAR，存储 "pre" 或 "post"）
   - 更新 `SaleBillSetting` 模型，添加 `BatchCookingMode` 字段

2. **订单创建逻辑修改**
   - 修改 `NewSaleBillSetting` 函数，从 `businessSetting.BatchCookingMode` 读取值并保存到 `sale_bill_setting.batch_cooking_mode`
   - 确保默认值为 "post"（与 `business_setting` 保持一致）

3. **送厨逻辑修改**
   - 修改所有使用分批送厨模式的地方，从 `sale_bill_setting.batch_cooking_mode` 读取，而不是从 `business_setting` 读取
   - 确保已创建订单的行为不受全局配置变更影响

4. **数据迁移**
   - 为历史订单的 `batch_cooking_mode` 字段设置默认值（建议使用 "post"）

### 影响范围

**涉及终端**：
- [x] POS 收银端（订单创建、送厨）
- [x] Shop 商家管理端（配置修改不影响已创建订单）
- [x] KDS 厨显端（送厨逻辑）
- [ ] QDS 排号叫号端
- [x] Assistant 助手端（送厨）
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型（`SaleBillSetting`）
- [x] 业务逻辑（订单创建、送厨逻辑）
- [ ] 第三方集成
- [x] 数据库迁移

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**工作项分解**：
1. 数据库迁移脚本编写和测试（0.5 天）
2. 模型和 DTO 更新（0.5 天）
3. 订单创建逻辑修改（0.5 天）
4. 送厨逻辑修改和测试（1 天）
5. 联调和回归测试（0.5-1 天）

### 风险识别

**潜在风险**：
1. **历史数据兼容性**：历史订单的 `batch_cooking_mode` 字段为空，需要设置默认值
2. **送厨逻辑分散**：可能存在多处读取 `business_setting.BatchCookingMode` 的地方，需要全面排查
3. **测试覆盖不足**：送厨逻辑涉及多个终端和场景，需要充分测试

**缓解措施**：
1. **数据迁移策略**：为历史订单设置默认值 "post"，与当前业务逻辑保持一致
2. **代码审查**：使用代码搜索工具全面查找所有使用 `BatchCookingMode` 的地方
3. **测试计划**：制定详细的测试用例，覆盖前置/后置模式、订单创建、送厨等场景

---

## 🔗 相关资源

### 参考需求

- 类似功能: `assistant-batch-cooking-pre-mode.md`、`cashier-batch-cooking-pre-mode.md`
- 相关字段: `business_setting.go` 第 33 行 `BatchCookingMode` 字段

### 相关文档

- 数据库规范: `.cursor/rules/database.mdc`
- Go Main 开发规范: `.cursor/rules/go-main.mdc`
- 订单服务代码: `main/app/service/order.go` (NewSaleBillSetting 函数)

### 相关代码位置

- 模型定义: `main/app/model/order.go` (SaleBillSetting)
- 订单创建: `main/app/service/order.go` (NewSaleBillSetting)
- 业务设置: `main/app/dto/resp/setting/business_setting.go` (BatchCookingMode)
- 数据库表: `ttpos_sale_bill_setting`

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

- [x] 创建 Spec：`story-main-order-batch-cooking-mode-snapshot` ✅
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 在 shop 端修改分批送厨模式时，已创建的订单保持创建时的模式不变  
**以便于** 确保订单行为一致，避免因配置变更导致已创建订单的行为异常

**作为** 收银员/厨房人员  
**我想** 订单的送厨行为在创建后保持稳定  
**以便于** 按照预期的送厨模式进行工作，不会因配置变更而混乱

### AC 验收标准（初稿）

1. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 将当前 `business_setting.BatchCookingMode` 的值保存到 `sale_bill_setting.batch_cooking_mode`
2. **IF** shop 端修改了 `business_setting.BatchCookingMode` **THEN** 已创建的订单 **SHALL** 仍使用创建时保存的 `batch_cooking_mode` 值
3. **WHEN** 订单送厨时 **THEN** 系统 **SHALL** 使用 `sale_bill_setting.batch_cooking_mode` 的值，而不是 `business_setting.BatchCookingMode`
4. **IF** 历史订单的 `batch_cooking_mode` 为空 **THEN** 系统 **SHALL** 使用默认值 "post"
5. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 返回该订单创建时的 `batch_cooking_mode` 值

### 技术实现要点

1. **数据库字段**
   ```sql
   ALTER TABLE `ttpos_sale_bill_setting` 
   ADD COLUMN `batch_cooking_mode` VARCHAR(10) NOT NULL DEFAULT 'post' 
   COMMENT '分批送厨模式: pre-前置 / post-后置，默认 post';
   ```

2. **模型字段**
   ```go
   BatchCookingMode string `gorm:"column:batch_cooking_mode;type:varchar(10);default:'post';comment:分批送厨模式: pre-前置 / post-后置，默认 post" json:"batch_cooking_mode"`
   ```

3. **订单创建逻辑**
   - 在 `NewSaleBillSetting` 函数中，从 `businessSetting.BatchCookingMode` 读取值
   - 如果为空，使用默认值 "post"
   - 保存到 `saleBillSetting.BatchCookingMode`

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
**创建日期**: 2025-12-02  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/database.mdc`

