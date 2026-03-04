# Grab/LINE MAN 支付方式保存幂等性优化 需求文档

## 📋 基本信息

| 项目              | 内容                                                                               |
| ----------------- | ---------------------------------------------------------------------------------- |
| **Spec ID**       | task-backend-grab-lineman-payment-idempotent                                       |
| **Level**         | task（技术任务）                                                                   |
| **来源 Proposal** | [all-grab-lineman-payment-sync](../../../team/proposals/2026-02/all-grab-lineman-payment-sync.md) |
| **创建日期**      | 2026-02-27                                                                         |
| **负责人**        | 王昱                                                                               |
| **目标 Sprint**   | 待定                                                                               |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发完成   |
| **审核人**   | -          |
| **审核日期** | 2026-02-28 |

---

## 📝 用户故事

**作为** 运维人员/开发人员/商户管理员
**我想** 在保存 Grab/LINE MAN 支付方式时能够自动处理 ERP 已存在的情况
**以便于** 避免 ERP 和 TTPOS 数据不一致导致业务流程阻塞，减少人工介入修复的工作量

---

## 功能需求

### Requirement 1: ERP 支付方式存在性检查

**用户故事**: 作为开发人员，我想在创建支付方式前先检查 ERP 是否已存在，以便于避免重复创建导致数据不一致

#### 验收标准

1. **WHEN** 调用 `SaveGrabPaymentMethod` 或 `SaveLineManPaymentMethod` **THEN** 系统 **SHALL** 先查询 ERP 是否已存在对应的支付方式
2. **IF** ERP 已存在对应支付方式 **THEN** 系统 **SHALL** 直接获取该支付方式数据，不再调用 ERP 创建接口

---

### Requirement 2: ERP 已存在时复用数据

**用户故事**: 作为开发人员，我想在 ERP 已存在支付方式时直接复用，以便于保证数据一致性

#### 验收标准

1. **WHEN** ERP 已存在 Grab/LINE MAN 支付方式 **THEN** 系统 **SHALL** 获取 ERP 返回的支付方式数据并同步到 TTPOS
2. **WHEN** TTPOS 中不存在该支付方式 **THEN** 系统 **SHALL** 使用 ERP 数据在 TTPOS 中创建
3. **WHEN** TTPOS 中已存在该支付方式 **THEN** 系统 **SHALL** 更新 TTPOS 数据与 ERP 保持一致

---

### Requirement 3: ERP 创建失败时重新确认

**用户故事**: 作为开发人员，我想在 ERP 返回错误时能够重新确认实际状态，以便于处理网络超时等异常场景

#### 验收标准

1. **WHEN** ERP 创建接口返回错误 **THEN** 系统 **SHALL** 重新查询 ERP 确认支付方式是否实际已创建
2. **IF** 查询确认 ERP 已创建 **THEN** 系统 **SHALL** 按「已存在」流程处理，同步到 TTPOS
3. **IF** 查询确认 ERP 未创建 **THEN** 系统 **SHALL** 返回原始错误给调用方

---

### Requirement 4: TTPOS 侧幂等性保证

**用户故事**: 作为开发人员，我想保证重复调用保存方法时结果一致，以便于提高系统健壮性

#### 验收标准

1. **WHEN** 重复调用 `SaveGrabPaymentMethod` 或 `SaveLineManPaymentMethod` **THEN** 系统 **SHALL** 返回一致的结果
2. **WHEN** TTPOS 中已存在该支付方式 **THEN** 系统 **SHALL** 返回已存在的数据，不创建重复记录

---

### Requirement 5: createPaymentFromERP 幂等性优化

**用户故事**: 作为开发人员，我想优化从 ERP 同步支付方式到 TTPOS 的逻辑，以便于正确处理 ERP 已创建但 TTPOS 未创建的场景

#### 验收标准

1. **WHEN** 调用 `createPaymentFromERP` **AND** ERP 中已存在支付方式 **AND** TTPOS 中不存在 **THEN** 系统 **SHALL** 从 ERP 获取数据并在 TTPOS 创建
2. **WHEN** 调用 `createPaymentFromERP` **AND** 两侧数据不一致 **THEN** 系统 **SHALL** 以 ERP 数据为准更新 TTPOS

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 测试场景包括：
  - ERP 已存在、TTPOS 不存在
  - ERP 已存在、TTPOS 已存在
  - ERP 不存在、TTPOS 不存在
  - ERP 创建返回错误但实际已创建
  - 并发调用场景

### 日志要求

- [ ] 所有日志必须包含 `company_uuid` 字段
- [ ] 记录 ERP 查询/创建的请求和响应
- [ ] 记录幂等性判断的关键决策点

### 平台兼容性

- [x] Go 后端服务（main 模块）

---

## 技术约束

### Go 后端约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- 通过 `ctx.GetDB()` 获取数据库连接
- 数据库操作必须通过 Repository 层
- 多表操作使用事务包裹

### 涉及代码

- `main/app/service/payment_method.go`
  - `SaveGrabPaymentMethod`
  - `SaveLineManPaymentMethod`
  - `createPaymentFromERP`

### 资源约束

- Story Point: 3（待确认）

---

## 风险和缓解

### 风险 1: ERP 查询接口不存在或不支持

**影响**: 高
**缓解措施**: 确认 ERP 是否提供按条件查询已有支付方式的接口；若不支持需协调 ERP 团队添加

### 风险 2: 并发场景下的竞态条件

**影响**: 中
**缓解措施**: 使用分布式锁或数据库唯一约束防止并发创建

### 风险 3: ERP 和 TTPOS 数据模型不一致

**影响**: 低
**缓解措施**: 明确数据映射关系，编写转换逻辑

---

**版本**: v1.1.0
**创建日期**: 2026-02-27
**更新日期**: 2026-02-28
