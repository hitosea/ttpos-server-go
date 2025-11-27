# 优化删除拆单时的账单完成判断逻辑 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-11-26   |
| **目标版本** | v2.x |
| **状态**   | ✅ 已批准 - 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [task-main-optimize-delete-order-finish-bill-logic](../../shared/specs/active/task-main-optimize-delete-order-finish-bill-logic/requirements.md)      |
| **原始 Spec** | story-main-table-multi-order-lock      |
| **关联文档** | [InstantOrderSaleOrderDelete 方法分析](../../human/guides/order-instant-order-sale-order-delete-analysis.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `InstantOrderSaleOrderDelete` 方法在删除拆单时，只在剩余订单数量为 2 时检查是否完成销售账单。这导致在多订单场景下存在业务逻辑缺陷：

**问题场景**：
- 有 3 个订单：订单1（已结账）+ 订单2（空订单）+ 订单3（已结账）
- 删除空订单2后，剩余订单1和订单3都已结账
- **预期行为**：整个销售账单应该自动完成
- **实际行为**：只是删除了订单2，账单未自动完成

**代码限制**：
```go
// 当前代码：只检查订单数量是否为 2
if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) == 2 {
    // 完成账单
    if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
        return errors.WithMessage(err)
    }
}
```

### 业务价值

**优化收益**：
1. **完整的业务逻辑**：覆盖所有多订单部分结账的场景
2. **提升用户体验**：自动完成账单，无需人工干预
3. **减少操作步骤**：避免额外的完成账单操作
4. **系统健壮性**：处理更复杂的业务场景

**业务影响**：
- 适用于桌台订单的部分结账场景
- 减少收银员的操作复杂度
- 提高订单状态管理的准确性

### 目标用户

- [x] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 店长、运营人员

---

## 💡 解决方案概述

### 方案描述

将账单完成的判断逻辑从"订单数量检查"改为"剩余订单结账状态检查"：

**核心改进**：
- 不再仅检查 `len(saleBill.SaleOrders) == 2`
- 检查删除指定订单后，剩余订单是否全部已结账
- 如果剩余订单全部已结账，则自动触发账单完成流程

**优化后行为**：
```
场景1: 订单1已结账 + 订单2空订单 → 删除订单2 → 完成账单 ✅
场景2: 订单1已结账 + 订单2空订单 + 订单3已结账 → 删除订单2 → 完成账单 ✅
场景3: 订单1已结账 + 订单2空订单 + 订单3未结账 → 删除订单2 → 不完成账单 ✅
```

### 核心功能点

1. **在 SaleBill model 中新增方法** `ShouldFinishBillAfterDelete`
   - 判断删除指定订单后，剩余订单是否全部已结账
   - 返回 true/false
   - 放在 model 层符合单一职责原则

2. **重构 service 层判断逻辑**
   - 替换硬编码的订单数量检查
   - 调用 model 层的状态检查方法

3. **保持现有行为**
   - 不影响其他业务场景
   - 保持向后兼容

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [x] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口（无接口变更）
- [x] 数据模型（SaleBill model 新增方法）
- [x] 业务逻辑（service 层逻辑优化）
- [ ] 第三方集成

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯业务逻辑优化，无架构变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**复杂度说明**：
- 仅涉及一个方法的逻辑优化
- 新增一个辅助判断方法
- 无数据库结构变更
- 无 API 接口变更

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1 SP（待技术评审确认）

**工作内容**：
1. 在 `sale_bill.go` model 中新增 `ShouldFinishBillAfterDelete` 方法（30分钟）
2. 在 `order_base.go` service 中重构判断逻辑（30分钟）
3. 编写单元测试（model 层 + service 层）（1小时）
4. 代码审查和优化（1小时）

### 风险识别

**潜在风险**：
1. **逻辑覆盖不全**：可能存在其他未考虑的边缘场景
2. **测试不充分**：需要覆盖多种订单组合场景
3. **兼容性问题**：可能影响现有的业务流程

**缓解措施**：
1. **充分的单元测试**：覆盖所有订单状态组合
2. **代码审查**：团队 Review 确保逻辑正确
3. **灰度发布**：先在测试环境验证，再逐步发布
4. **监控告警**：监控账单完成的异常情况

### 设计亮点

**架构优化**：将判断逻辑从 Service 层移到 Model 层

**优势分析**：

| 维度 | Service 层实现 | Model 层实现 ✅ |
|-----|--------------|---------------|
| **职责划分** | 混合业务流程和状态判断 | 状态判断属于 Model 职责 |
| **可测试性** | 依赖数据库和其他服务 | 纯逻辑，易于单元测试 |
| **可复用性** | 仅限当前 Service 使用 | 任何地方都可调用 |
| **可维护性** | 修改影响范围大 | 修改仅影响 Model 层 |
| **代码质量** | 违反单一职责原则 | 符合分层架构设计 |

**符合的设计原则**：
- ✅ **单一职责原则（SRP）**：状态判断是 Model 的职责
- ✅ **高内聚低耦合**：相关逻辑聚合在一起
- ✅ **开闭原则（OCP）**：扩展性更好
- ✅ **依赖倒置原则（DIP）**：Service 依赖 Model 的抽象

---

## 🔗 相关资源

### 参考需求

- 原始功能: [桌台多订单操作并发控制](./table-operations-multi-order-lock.md)
- 分析文档: [InstantOrderSaleOrderDelete 方法分析](../../human/guides/order-instant-order-sale-order-delete-analysis.md)

### 相关文档

- 需求设计: `docs/shared/specs/active/story-main-table-multi-order-lock/design.md`
- 任务清单: `docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md`
- 代码位置: `main/app/service/order_base.go:924-1085`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | -      |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | N/A    | N/A       |

### 评审结论

- [x] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
提案评审通过：
- 设计合理，将判断逻辑放在 Model 层符合分层架构原则
- 测试用例完整，覆盖所有业务场景
- 工作量估算合理（1 SP）
- 优先级中等，建议下个迭代 Sprint 实现
```

**下一步行动**：

- [x] 创建 Spec：`task-main-optimize-delete-order-finish-bill-logic`
- [x] 分配负责人：xiezhihuan
- [ ] 目标 Sprint：下个迭代
- [ ] 技术设计：待 `/spec-design` 创建

---

## 📝 附录

### User Story（初稿）

**作为** 收银员  
**我想** 在删除空的拆单后，系统能自动完成已全部结账的账单  
**以便于** 减少手动完成账单的操作，提高工作效率

### AC 验收标准（初稿）

#### AC1: 删除空订单后自动完成账单（2个订单）
**GIVEN** 存在销售账单，包含订单1（已结账）和订单2（空订单）  
**WHEN** 删除订单2  
**THEN** 系统 **SHALL** 自动完成该销售账单

#### AC2: 删除空订单后自动完成账单（3个订单）
**GIVEN** 存在销售账单，包含订单1（已结账）、订单2（空订单）、订单3（已结账）  
**WHEN** 删除订单2  
**THEN** 系统 **SHALL** 自动完成该销售账单

#### AC3: 删除空订单后不完成账单（存在未结账订单）
**GIVEN** 存在销售账单，包含订单1（已结账）、订单2（空订单）、订单3（未结账）  
**WHEN** 删除订单2  
**THEN** 系统 **SHALL NOT** 完成该销售账单

#### AC4: 向后兼容
**GIVEN** 原有的所有删除拆单场景  
**WHEN** 执行删除操作  
**THEN** 系统 **SHALL** 保持原有行为不变

### 技术方案概要

**设计原则**：
将判断逻辑放在 `SaleBill` model 层而非 service 层，遵循以下设计原则：
- ✅ **单一职责**：状态判断属于 model 的职责
- ✅ **高内聚**：订单状态判断与订单数据紧密相关
- ✅ **可复用**：其他 service 也可能需要此判断逻辑
- ✅ **易测试**：model 层方法更容易进行单元测试

**1. 在 SaleBill model 中新增方法**（`main/app/model/sale_bill.go`）：
```go
// ShouldFinishBillAfterDelete 判断删除指定订单后，剩余订单是否全部已结账
// 参数：deleteOrderUuid - 要删除的订单UUID
// 返回：true - 剩余订单全部已结账，应该完成账单；false - 仍有未结账订单
func (sb *SaleBill) ShouldFinishBillAfterDelete(deleteOrderUuid uint64) bool {
    for _, order := range sb.SaleOrders {
        if order.Uuid == deleteOrderUuid {
            continue // 跳过要删除的订单
        }
        if !order.IsSettled() {
            return false // 存在未结账订单
        }
    }
    return true // 所有剩余订单都已结账
}
```

**2. 优化 service 层判断逻辑**（`main/app/service/order_base.go`）：
```go
// 改进前：硬编码订单数量检查
if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) == 2 {
    // 完成账单
    if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
        return errors.WithMessage(err)
    }
}

// 改进后：调用 model 层方法检查订单状态
if len(moveProductList) == 0 && saleBill.ShouldFinishBillAfterDelete(saleOrderFrom.Uuid) {
    // 获取业务设置
    businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
    if err != nil {
        return errors.WithMessage(err)
    }
    // 完成销售账单
    if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
        return errors.WithMessage(err)
    }
}
```

### 测试用例清单

#### Model 层单元测试（`sale_bill_test.go`）

| 用例编号 | 测试方法 | 订单配置 | deleteOrderUuid | 预期返回 |
|---------|---------|---------|----------------|---------|
| UT1 | TestShouldFinishBillAfterDelete_AllSettled | 订单1(已结账) + 订单2(已结账) + 订单3(已结账) | 订单2 | true ✅ |
| UT2 | TestShouldFinishBillAfterDelete_HasUnSettled | 订单1(已结账) + 订单2(已结账) + 订单3(未结账) | 订单2 | false ✅ |
| UT3 | TestShouldFinishBillAfterDelete_OnlyOneLeft | 订单1(已结账) + 订单2(已结账) | 订单2 | true ✅ |
| UT4 | TestShouldFinishBillAfterDelete_EmptyOrders | 空订单列表 | 任意 | true ✅ |

#### Service 层集成测试（`order_base_test.go`）

| 用例编号 | 场景描述 | 订单配置 | 预期结果 |
|---------|---------|---------|---------|
| IT1 | 2个订单，删除空订单 | 订单1(已结账) + 订单2(空) | 完成账单 ✅ |
| IT2 | 3个订单，删除空订单，全部已结账 | 订单1(已结账) + 订单2(空) + 订单3(已结账) | 完成账单 ✅ |
| IT3 | 3个订单，删除空订单，存在未结账 | 订单1(已结账) + 订单2(空) + 订单3(未结账) | 不完成账单 ✅ |
| IT4 | 4个订单，删除空订单，全部已结账 | 订单1(已结账) + 订单2(空) + 订单3(已结账) + 订单4(已结账) | 完成账单 ✅ |
| IT5 | 删除有商品的订单 | 订单1(已结账) + 订单2(有商品) | 返回错误 ✅ |
| IT6 | 删除订单1 | 订单1(任意) + 订单2(任意) | 返回错误 ✅ |

---

## 📄 模板使用说明

### 与原需求的关系

本提案是对现有功能 `story-main-table-multi-order-lock` 的优化改进：
- **原需求**：实现桌台多订单操作的并发控制和拆单功能
- **本提案**：优化删除拆单时的账单完成判断逻辑
- **状态**：原需求已实现，本提案为增量优化

### 优先级评估

- **紧急程度**: 中（非阻塞性缺陷）
- **重要程度**: 中（影响用户体验）
- **实现成本**: 低（1 SP）
- **建议排期**: 下个迭代 Sprint

---

## 📝 版本更新记录

| 版本 | 日期 | 修改内容 | 修改人 |
|-----|------|---------|--------|
| v1.0 | 2025-11-26 | 初始提案创建 | xiezhihuan |
| v1.1 | 2025-11-26 | 将判断逻辑方法调整到 SaleBill model 层，增强设计合理性 | xiezhihuan |

---

**版本**: v1.1  
**创建日期**: 2025-11-26  
**最后更新**: 2025-11-26  
**维护者**: TTPOS 开发团队  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

