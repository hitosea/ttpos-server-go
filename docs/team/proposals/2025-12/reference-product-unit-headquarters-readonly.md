# 参考商品单位实现，来源总部的数据不可编辑 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 曾振华   |
| **日期**   | 2025-12-08   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | DooTask #37479 |
| **关联 Spec** | [story-shop-headquarters-data-readonly](../../../shared/specs/active/story-shop-headquarters-data-readonly/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

在总部-分店颗粒化同步场景中，分店同步了总部的数据后，需要确保这些来源总部的数据不可编辑，以保持数据一致性和统一管理。

**参考实现**：商品单位（ProductUnit）模块已经实现了总部来源数据不可编辑的功能，包括：
- ✅ 列表/详情接口返回 `is_editable` 字段
- ✅ 编辑/删除接口增加总部来源数据校验
- ✅ 使用 `isEditable(ctx, headquarterUuid)` 函数判断

**需要实现**：参考商品单位的实现方式，为以下模块实现相同的功能：
- 新管理端-菜品标签（ProductLabel）：总部来源不可编辑
- 新管理端-满额减（FullReductionActivity）：总部来源不可编辑
- 新管理端-商品（ProductPackage）：总部来源不可编辑（只能修改外卖的价格、上下架）

### 业务价值

- **数据一致性**：确保总部数据在分店中保持统一，避免分店误操作导致数据不一致
- **统一管理**：总部可以统一管理物品单位，分店只能使用，不能修改
- **降低风险**：减少因分店误编辑导致的数据错误和业务风险
- **用户体验**：明确区分可编辑和不可编辑的数据，提升用户操作体验

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: 店长、运营人员

---

## 💡 解决方案概述

### 方案描述

参考商品单位（ProductUnit）的实现方式，为以下模块实现总部来源数据不可编辑的功能：
- 菜品标签（ProductLabel）
- 满额减（FullReductionActivity）
- 商品（ProductPackage）

**核心实现**（参考 ProductUnit）：
1. 在后端 API 中判断数据是否来源总部（通过 `headquarter_uuid` 字段）
2. 在响应数据中添加 `is_editable` 字段，标识是否可编辑
3. 前端根据 `is_editable` 字段控制编辑按钮的显示和表单字段的禁用状态
4. 后端在更新接口中增加校验，拒绝编辑来源总部的数据

### 核心功能点

1. **后端 API 增强**（每个模块）
   - 在列表/详情接口中返回 `is_editable` 字段
   - 在更新/删除接口中增加总部来源数据校验，拒绝编辑请求
   - 使用 `isEditable(ctx, headquarterUuid)` 函数判断（参考 ProductUnit）

2. **前端 UI 控制**（每个模块）
   - 列表页：来源总部的数据显示"不可编辑"标识
   - 详情页：来源总部的数据禁用编辑按钮和表单字段
   - 提供明确的视觉反馈，告知用户为什么不可编辑

3. **特殊处理**
   - 商品（ProductPackage）：只能修改外卖的价格、上下架
   - 支付方式（PaymentMethod）：依照支付方式模块的特殊规则

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（新管理端）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [x] UI 组件（新管理端各模块管理页面）
- [x] API 接口（菜品标签、满额减、特价菜、商品、支付方式相关接口）
- [x] 数据模型（ProductLabel、FullReductionActivity、MarketingActivity、ProductPackage、PaymentMethod）
- [x] 业务逻辑（总部来源数据校验）
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：此功能已有多个模块的实现参考，技术实现相对成熟，主要是参考现有实现模式进行适配。

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3 SP（待技术评审确认）

**分解**：
- 后端 API 修改：1 天
- 前端 UI 实现：1 天
- 联调测试：0.5-1 天

### 风险识别

**潜在风险**：
1. **数据一致性风险**：如果已有分店数据误标记了 `headquarter_uuid`，可能导致正常数据被误判为不可编辑
2. **用户体验风险**：如果提示信息不明确，用户可能不理解为什么不能编辑
3. **同步逻辑风险**：需要确保同步后的数据正确标记总部来源

**缓解措施**：
1. **数据校验**：在实现前检查现有数据，确保 `headquarter_uuid` 字段的正确性
2. **明确提示**：在 UI 中明确显示"来源总部，不可编辑"的提示信息
3. **参考实现**：参考已有模块（菜品标签、商品等）的实现方式，确保逻辑一致性

---

## 🔗 相关资源

### 参考需求

- **类似功能实现**：
  - 菜品标签总部来源不可编辑：`docs/shared/specs/active/shop-headquarters-branch-granular-sync-backend/`
  - 商品总部来源不可编辑：同上
  - 支付方式总部来源不可编辑：同上

### 相关文档

- **总部-分店颗粒化同步需求文档**：`docs/shared/specs/active/shop-headquarters-branch-granular-sync-backend/requirements.md`
- **相关数据指南**：`docs/shared/specs/active/shop-headquarters-branch-granular-sync-backend/RELATED_DATA_GUIDE.md`
- **前端仓库**：shop-headquarters-branch-granular-sync

### 代码参考

- **参考实现（商品单位）**：
  - 模型：`main/app/model/product.go` - `ProductUnit` 结构体
  - 服务：`main/app/service/product.go` - `GetProductUnitList()`, `GetProductUnit()`, `EditProductUnit()`, `DeleteProductUnit()`
  - 响应结构：`main/app/dto/resp/product_resp/product.go` - `ProductUnitItem`, `ProductUnitDetail`
  - 判断函数：`main/app/service/product.go` - `isEditable(ctx, headquarterUuid)`

- **需要实现的模块**：
  - 菜品标签：`main/app/service/product_label.go`, `main/app/dto/resp/product_label.go`
  - 满额减：`main/app/service/full_reduction_activity_srv.go`, `main/app/dto/resp/full_reduction_activity_resp.go`
  - 特价菜：`main/app/service/marketing_activity_srv.go`, `main/app/dto/resp/member_resp/marketing_activity_list.go`
  - 商品：`main/app/service/product.go`, `main/app/dto/resp/product_resp/product.go`
  - 支付方式：`main/app/service/payment_method_srv.go`, `main/app/dto/resp/payment_method.go`

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

- [ ] 创建 Spec：`story-shop-headquarters-data-readonly`（涵盖所有模块）
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 店长/运营人员  
**我想** 在分店中查看和使用总部的数据（菜品标签、满额减、商品），但不能编辑来源总部的数据  
**以便于** 保持数据一致性，避免误操作导致的数据错误

### AC 验收标准（初稿）

**通用验收标准**（适用于所有模块）：
1. **WHEN** 分店同步了总部的数据 **THEN** 系统 **SHALL** 在列表中显示"来源总部，不可编辑"标识
2. **WHEN** 用户尝试编辑来源总部的数据 **THEN** 系统 **SHALL** 禁用编辑按钮和表单字段，并显示明确的提示信息
3. **WHEN** 用户通过 API 尝试更新来源总部的数据 **THEN** 系统 **SHALL** 返回错误提示，拒绝更新请求
4. **IF** 数据的 `headquarter_uuid` 字段为 0 或空 **THEN** 系统 **SHALL** 允许正常编辑
5. **WHEN** 分店同步总部的数据后 **THEN** 系统 **SHALL** 正确标记 `headquarter_uuid`，并自动设置为不可编辑状态

**特殊验收标准**：
- 商品（ProductPackage）：总部来源的商品只能修改外卖的价格、上下架，其他字段不可编辑

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

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
**创建日期**: 2025-12-08  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
