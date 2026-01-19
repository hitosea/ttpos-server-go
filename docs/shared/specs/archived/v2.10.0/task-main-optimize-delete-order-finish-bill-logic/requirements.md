> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 优化删除拆单时的账单完成判断逻辑 需求文档

> 本文档定义优化删除拆单时账单完成判断逻辑的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                      |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/optimize-delete-order-finish-bill-logic.md](../../../../team/proposals/2025-11/optimize-delete-order-finish-bill-logic.md) |
| **创建日期**      | 2025-11-26                                                                                                                                |
| **负责人**        | xiezhihuan                                                                                                                                |
| **目标 Sprint**   | Sprint 下个迭代                                                                                                                            |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                               |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | ✅ 已通过 - 已实施     |
| **审核人**   | xiezhihuan          |
| **审核日期** | 2025-11-26          |
| **审核意见** | 核心功能已实现，测试用例完整，待实际环境验证          |

---

## 📋 概述

优化 `InstantOrderSaleOrderDelete` 方法在删除拆单时的账单完成判断逻辑。当前实现只在剩余订单数为 2 时检查是否完成账单，导致在多订单场景（如 3 个订单且部分已结账）下无法自动完成账单。

本优化将判断逻辑从"订单数量检查"改为"剩余订单结账状态检查"，通过在 `SaleBill` model 中新增方法来实现更通用的状态判断，提升系统健壮性和用户体验。

## 🎯 产品对齐

该功能支持系统用户体验和代码质量优化目标：

- **完整的业务逻辑**：覆盖所有多订单部分结账的场景
- **提升用户体验**：自动完成账单，无需人工干预
- **减少操作步骤**：避免额外的完成账单操作
- **提高代码质量**：遵循分层架构和设计原则
- **增强可维护性**：状态判断逻辑集中在 Model 层

## 📝 用户故事

**作为** 收银员  
**我想** 在删除空的拆单后，系统能自动完成已全部结账的账单  
**以便于** 减少手动完成账单的操作，提高工作效率

---

## 功能需求

### Requirement 1: Model 层新增账单完成判断方法

**用户故事**: 作为开发者，我想在 `SaleBill` model 中添加状态判断方法，以便于在不同场景下复用该逻辑

#### 验收标准

1. **WHEN** 调用 `SaleBill.ShouldFinishBillAfterDelete(deleteOrderUuid)` **THEN** 系统 **SHALL** 返回删除指定订单后剩余订单是否全部已结账的布尔值
2. **IF** 删除指定订单后剩余订单全部已结账 **THEN** 方法 **SHALL** 返回 `true`
3. **IF** 删除指定订单后存在未结账的订单 **THEN** 方法 **SHALL** 返回 `false`
4. **WHEN** 订单列表为空 **THEN** 方法 **SHALL** 返回 `true`

#### 具体要求

- [x] 1.1 在 `main/app/model/sale_bill.go` 中新增 `ShouldFinishBillAfterDelete` 方法
- [x] 1.2 方法签名: `func (sb *SaleBill) ShouldFinishBillAfterDelete(deleteOrderUuid uint64) bool`
- [x] 1.3 方法实现遍历 `SaleOrders`，跳过要删除的订单，检查剩余订单的结账状态
- [x] 1.4 添加详细的方法注释，说明参数和返回值
- [x] 1.5 方法遵循 Go 语言命名规范和单一职责原则

---

### Requirement 2: Service 层优化判断逻辑

**用户故事**: 作为开发者，我想重构 service 层的判断逻辑，以便于支持更多业务场景

#### 验收标准

1. **WHEN** 删除空订单且剩余订单全部已结账（无论订单数量） **THEN** 系统 **SHALL** 自动完成销售账单
2. **WHEN** 删除空订单但存在未结账订单 **THEN** 系统 **SHALL NOT** 完成销售账单
3. **WHEN** 原有的删除拆单场景 **THEN** 系统 **SHALL** 保持原有行为不变（向后兼容）

#### 具体要求

- [x] 2.1 在 `main/app/service/order_base.go` 的 `InstantOrderSaleOrderDelete` 方法中替换判断逻辑
- [x] 2.2 将 `len(saleBill.SaleOrders) == 2` 的硬编码检查改为调用 `saleBill.ShouldFinishBillAfterDelete(saleOrderFrom.Uuid)`
- [x] 2.3 保持其他业务逻辑不变（删除订单、事件发布等）
- [x] 2.4 确保在获取 `businessSetting` 和调用 `FinishSaleBill` 的逻辑保持不变
- [x] 2.5 保持错误处理和事务管理的一致性

---

### Requirement 3: 单元测试覆盖

**用户故事**: 作为开发者，我想编写完整的单元测试，以便于验证功能正确性和防止回归

#### 验收标准

1. **WHEN** 执行 Model 层单元测试 **THEN** 系统 **SHALL** 覆盖所有订单状态组合
2. **WHEN** 执行 Service 层集成测试 **THEN** 系统 **SHALL** 覆盖所有删除拆单场景
3. **WHEN** 运行测试套件 **THEN** 所有测试用例 **SHALL** 通过

#### 具体要求

- [x] 3.1 在 `sale_bill_test.go` 中编写 Model 层单元测试（至少 4 个用例）
- [x] 3.2 在 `order_base_test.go` 中编写 Service 层集成测试（至少 6 个用例）
- [x] 3.3 测试覆盖率达到 ≥ 80%
- [x] 3.4 包含边界条件测试（空订单列表、单个订单等）
- [x] 3.5 包含正向和负向测试用例

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 状态判断放在 Model 层，业务流程放在 Service 层
- **单一职责原则**: `ShouldFinishBillAfterDelete` 方法只负责状态判断
- **高内聚低耦合**: 相关逻辑聚合在 Model 层
- **可复用性**: Model 层方法可被其他 Service 调用
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### 设计原则

- [x] **单一职责原则（SRP）**: 状态判断是 Model 的职责
- [x] **高内聚**: 订单状态判断与订单数据紧密相关
- [x] **可测试性**: Model 层纯逻辑更容易测试
- [x] **开闭原则（OCP）**: 对扩展开放，对修改关闭

### 性能要求

- [x] Model 层方法时间复杂度 O(n)，n 为订单数量
- [x] Service 层逻辑不增加额外的数据库查询
- [x] 不影响现有的性能指标

### 测试要求

- [x] Model 层测试覆盖率 ≥ 80%
- [x] Service 层测试覆盖率 ≥ 80%
- [x] **Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖所有业务场景
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [x] 保持现有的并发安全机制（分布式锁）
- [x] 不引入新的安全风险
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 保持现有的事务管理机制
- [x] 保持现有的错误处理逻辑
- [x] 向后兼容，不影响现有功能

---

## 验收标准

### 功能验收

1. **场景1: 2个订单删除空订单**: 订单1（已结账）+ 订单2（空订单），删除订单2后自动完成账单 ✅
2. **场景2: 3个订单全部已结账**: 订单1（已结账）+ 订单2（空订单）+ 订单3（已结账），删除订单2后自动完成账单 ✅
3. **场景3: 3个订单存在未结账**: 订单1（已结账）+ 订单2（空订单）+ 订单3（未结账），删除订单2后不完成账单 ✅
4. **场景4: 4个订单全部已结账**: 删除空订单后自动完成账单 ✅
5. **向后兼容**: 原有的所有删除拆单场景保持原有行为 ✅

### 测试验收

1. **Model 层单元测试**: 
   - `TestShouldFinishBillAfterDelete_AllSettled`: 全部已结账场景
   - `TestShouldFinishBillAfterDelete_HasUnSettled`: 存在未结账场景
   - `TestShouldFinishBillAfterDelete_OnlyOneLeft`: 只剩一个订单场景
   - `TestShouldFinishBillAfterDelete_EmptyOrders`: 空订单列表场景

2. **Service 层集成测试**:
   - IT1-IT6: 覆盖所有业务场景

3. **测试覆盖率**: Model 层和 Service 层测试覆盖率均 ≥ 80%

### 文档验收

1. **分析文档**: `order-instant-order-sale-order-delete-analysis.md` 已更新
2. **提案文档**: `optimize-delete-order-finish-bill-logic.md` 完整
3. **需求文档**: 本文档（requirements.md）完整
4. **测试文档**: 测试用例清单完整

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- Model 层方法不能依赖外部服务
- Service 层可以调用 Model 层方法
- 保持现有的分层架构
- 不使用 panic，返回 error

### 业务约束

- 不能删除第一个销售订单（订单1）
- 删除有商品的订单时必须先移动商品
- 已送厨的商品必须先退菜
- 保持现有的业务规则不变

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 SP（必须 ≤ 5）
- 测试时间: 1 小时
- 代码审查: 1 小时

---

## 依赖关系

### 技术依赖

无新增外部依赖

### 内部模块依赖

- `main/app/model/sale_bill.go`: SaleBill model
- `main/app/model/sale_order.go`: SaleOrder model（使用 `IsSettled()` 方法）
- `main/app/service/order_base.go`: orderSrv service

### 业务依赖

- 依赖现有的拆单功能（`story-main-table-multi-order-lock`）
- 依赖 `SaleOrder.IsSettled()` 方法正确返回订单结账状态
- 依赖 `FinishSaleBill` 方法正确完成账单

---

## 风险和缓解

### 风险 1: 逻辑覆盖不全

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 编写完整的单元测试覆盖所有订单状态组合
- 代码审查确保逻辑正确
- 在测试环境充分验证

### 风险 2: 影响现有业务流程

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 保持向后兼容，不修改其他业务逻辑
- 编写集成测试覆盖所有删除拆单场景
- 灰度发布，先在测试环境验证

### 风险 3: 并发安全问题

**影响**: 高  
**概率**: 极低  
**缓解措施**:

- 保持现有的分布式锁机制不变
- Model 层方法是纯逻辑，不涉及并发操作
- Service 层已有完善的并发控制

---

## 时间表

- **Phase 1 - Model 层开发**: 0.5 小时
  - 新增 `ShouldFinishBillAfterDelete` 方法
  - 编写方法注释

- **Phase 2 - Service 层重构**: 0.5 小时
  - 替换判断逻辑
  - 保持其他逻辑不变

- **Phase 3 - 单元测试**: 1 小时
  - Model 层单元测试（4 个用例）
  - Service 层集成测试（6 个用例）

- **Phase 4 - 代码审查和优化**: 1 小时
  - 团队 Review
  - 性能和安全检查
  - 测试验证

- **总计**: 3 小时（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 分析文档

- `docs/human/guides/order-instant-order-sale-order-delete-analysis.md` - InstantOrderSaleOrderDelete 方法分析

### 提案文档

- `docs/team/proposals/2025-11/optimize-delete-order-finish-bill-logic.md` - 优化提案

### 关联 Spec

- `docs/shared/specs/active/story-main-table-multi-order-lock/` - 原始拆单功能 Spec

### 代码位置

- `main/app/model/sale_bill.go` - SaleBill model（新增方法）
- `main/app/service/order_base.go:924-1085` - InstantOrderSaleOrderDelete 方法（优化逻辑）

---

## Graphiti & 活动日志

- Related Episode: `优化删除拆单时的账单完成判断逻辑`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-11/2025-11-26.md`
- 提醒：优化完成后应记录到 Graphiti，包含设计决策和架构优化经验

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: xiezhihuan  
**审核者**: 待审核

