# 自助餐顾客类型名称快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-09 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-buffet-customer-type-name-snapshot-fix](../../shared/specs/active/story-main-buffet-customer-type-name-snapshot-fix/) |
| **父提案** | `order-attribute-snapshot-fix.md` |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，`ttpos_sale_order_buffet_customer_type` 表中的 `name` 字段（顾客类型名称）会随后台数据变更而改变，导致订单历史信息不准确。这是订单商品信息快照修复需求（`order-attribute-snapshot-fix.md`）的子任务。

**具体场景**：

1. **顾客类型被删除**：
   - 订单中 `SaleOrderBuffetCustomerType` 记录的 `name` 字段为"老人"（下单时选择了"老人"顾客类型）
   - 后台删除了"老人"顾客类型配置
   - 查询订单时，虽然可以通过 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 关联查询，但如果关联数据被删除，名称信息可能丢失或显示错误

2. **顾客类型被改名**：
   - 订单中 `SaleOrderBuffetCustomerType` 记录的 `name` 字段为"老人"（下单时选择了"老人"顾客类型）
   - 后台将"老人"顾客类型改名为"长者"
   - 查询订单时显示："长者"（显示的是新名称，而非下单时的名称）

3. **数据一致性问题**：
   - `SaleOrderBuffetCustomerType` 表中有 `name` 字段（VARCHAR(255)），但这是单语言字段，不支持多语言
   - 查询时依赖 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 关联查询，可能获取到错误或已删除的数据
   - 如果顾客类型被删除，关联查询失败，名称信息丢失

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、统计报表等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 顾客类型被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能
- ❌ 当前 `name` 字段是单语言（VARCHAR(255)），不支持多语言显示

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保 `SaleOrderBuffetCustomerType` 记录的顾客类型名称准确反映下单时的状态
- ✅ **多语言支持**：支持多语言快照，满足国际化需求
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **业务可靠性**：避免因数据变更导致的业务逻辑错误

### 目标用户

- [x] 收银员（查看历史订单）
- [x] 商户管理员（对账、报表）
- [ ] 厨房人员
- [ ] 顾客
- [x] 财务人员（对账、审计）

---

## 💡 解决方案概述

### 方案描述

**核心思路**：将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段从 `VARCHAR(255)` 修改为 `TEXT` 类型，保存顾客类型名称的多语言 JSON 快照。

**现状分析**：

1. **数据库设计问题**：
   - `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段是 `VARCHAR(255)`，只能存储单语言文本
   - 查询时依赖 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 关联查询，可能获取到错误或已删除的数据

2. **代码实现问题**：
   - 目前查询时通过 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取顾客类型名称
   - 如果关联数据被删除或改名，订单显示信息变化
   - 不支持多语言显示

**解决方案**：

1. **数据库结构变更**：
   - 将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段从 `VARCHAR(255)` 修改为 `TEXT` 类型
   - 用于保存顾客类型名称的多语言 JSON 快照

2. **Go Model 修改**：
   - 修改 `SaleOrderBuffetCustomerType` 结构体中的 `Name` 字段类型为 `string`，GORM 标签改为 `type:text`
   - 实现 `GetLocaleName()` 方法：优先使用快照字段（JSON），降级使用关联表数据
   - 实现 `SetNameSnapshot()` 方法：从 `BuffetCustomerTypePrice.BuffetCustomerType.MultiLanguageName` 序列化为 JSON 保存

3. **查询逻辑修改**：
   - 替换所有使用 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 的地方
   - 统一使用 `orderBuffetCustomer.GetLocaleName()` 方法
   - 确保查询时优先使用快照字段，降级使用关联表数据

4. **下单逻辑修改**：
   - 在创建 `SaleOrderBuffetCustomerType` 时，自动保存顾客类型名称快照
   - 从 `BuffetCustomerTypePrice.BuffetCustomerType.MultiLanguageName` 获取多语言数据并序列化为 JSON

### 核心功能点

1. **数据库迁移**：修改 `name` 字段类型为 `TEXT`
2. **快照保存**：下单时自动保存顾客类型名称的多语言 JSON 快照
3. **快照查询**：查询时优先使用快照字段，降级使用关联表数据
4. **多语言支持**：快照字段保存完整的多语言 JSON（`dto.LocaleResponse`）

### 影响范围

**涉及终端**：
- [x] POS 收银端（订单查询）
- [x] Shop 商家管理端（订单管理、报表）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（订单查询相关接口）
- [x] 数据模型（`SaleOrderBuffetCustomerType`）
- [x] 业务逻辑（订单查询、下单逻辑）
- [ ] 第三方集成
- [x] 数据库（表结构变更）

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

### 风险识别

**潜在风险**：
1. **数据库结构变更风险**：修改字段类型可能影响现有数据
2. **历史数据不完整**：历史订单的 `name` 字段是单语言文本，需要降级处理
3. **多语言支持问题**：需要确保快照字段保存完整的多语言 JSON

**缓解措施**：
1. 使用 `ALTER TABLE MODIFY COLUMN` 修改字段类型，不影响现有数据（VARCHAR 可以转换为 TEXT）
2. 实现降级逻辑，确保历史订单正常显示（如果快照字段为空或无效，降级使用关联表数据）
3. 采用 JSON 格式保存多语言数据（`dto.LocaleResponse`），与自助餐套餐名称快照方案保持一致

---

## 🔗 相关资源

### 参考需求

- 父提案: `order-attribute-snapshot-fix.md`
- 类似功能: `buffet-customer-type-package-name-snapshot-fix.md`（自助餐套餐名称快照修复）
- 类似功能: `product-attribute-snapshot-fix.md`（商品属性信息快照修复）

### 相关文档

- 订单商品信息快照修复 Spec: `docs/shared/specs/active/story-main-product-attribute-snapshot-fix/`
- 自助餐顾客类型套餐名称快照修复 Spec: `docs/shared/specs/active/story-main-buffet-customer-type-package-name-snapshot-fix/`

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

- [ ] 创建 Spec：`story-main-buffet-customer-type-name-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实顾客类型名称  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的顾客类型名称  
**以便于** 准确处理退款和客户咨询

### AC 验收标准（初稿）

1. **WHEN** 执行数据库迁移脚本 **THEN** 系统 **SHALL** 将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段类型修改为 `TEXT`
2. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 正确保存顾客类型名称快照（JSON 格式）
3. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 优先使用快照字段，降级使用关联表数据
4. **IF** 顾客类型被删除或改名 **THEN** 历史订单 **SHALL** 仍显示下单时的名称
5. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

