# 采购限制方案-是否允许采购控制 需求文档

## 📋 基本信息

| 项目              | 内容                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------ |
| **来源 Proposal** | [shop-purchase-allow-control](../../../team/proposals/2026-02/shop-purchase-allow-control.md) |
| **DooTask**       | #39414                                                                               |
| **创建日期**      | 2026-02-03                                                                           |
| **负责人**        | weifashi                                                                             |
| **目标 Sprint**   | 待定                                                                                 |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 开发中 |
| **审核人**   | -      |
| **审核日期** | -      |

---

## 📝 用户故事

**作为** 采购员
**我想** 在采购限制方案中设置是否允许采购某些商品
**以便于** 更灵活地管理采购策略，提升采购管理的精细化控制能力

---

## 功能需求

### Requirement 1: 数据库字段扩展

**用户故事**: 作为采购员，我想在采购限制方案明细中控制是否允许采购，以便于禁止采购特定商品

#### 验收标准

1. **WHEN** 系统初始化 **THEN** `ttpos_purchase_limit_scheme_item` 表 **SHALL** 包含 `is_allow_purchase` 字段（string 类型，默认值 "yes"）
2. **WHEN** 现有数据迁移 **THEN** 系统 **SHALL** 将所有现有记录的 `is_allow_purchase` 设置为 "yes"

### Requirement 2: 创建接口适配

**用户故事**: 作为采购员，我想在创建采购限制方案时设置是否允许采购，以便于在配置阶段就控制采购权限

#### 验收标准

1. **WHEN** 调用 `/api/v1/shop/purchase/limit/scheme/create` 接口 **THEN** 系统 **SHALL** 支持 `is_allow_purchase` 参数
2. **IF** 未传入 `is_allow_purchase` 参数 **THEN** 系统 **SHALL** 使用默认值 "yes"
3. **WHEN** 传入 `is_allow_purchase` 参数 **THEN** 系统 **SHALL** 仅接受 "yes" 或 "no" 值

### Requirement 3: 更新接口适配

**用户故事**: 作为采购员，我想在更新采购限制方案时修改是否允许采购，以便于动态调整采购策略

#### 验收标准

1. **WHEN** 调用 `/api/v1/shop/purchase/limit/scheme/update` 接口 **THEN** 系统 **SHALL** 支持 `is_allow_purchase` 参数
2. **WHEN** 传入 `is_allow_purchase` 参数 **THEN** 系统 **SHALL** 更新对应记录的值
3. **IF** 传入无效的 `is_allow_purchase` 值 **THEN** 系统 **SHALL** 返回参数校验错误

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] API 接口测试覆盖

### 平台兼容性

- [x] Shop 商家管理端

### 数据兼容性

- [ ] 现有数据迁移后功能正常
- [ ] 向后兼容（未传参数时使用默认值）

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范

### 数据库约束

- 字段类型: VARCHAR (string)
- 允许值: "yes" | "no"
- 默认值: "yes"
- 迁移文件需同步更新 `admin/database/seeds/shop_01.sql`

### 资源约束

- Story Point: 2 (待技术评审确认)

---

## 风险和缓解

### 风险 1: 现有数据迁移

**影响**: 低
**缓解措施**: 迁移脚本设置默认值为 "yes"，不影响现有业务逻辑

### 风险 2: 前端 UI 适配

**影响**: 中
**缓解措施**: 与前端团队同步需求，协调开发进度

---

**版本**: v1.0.0
**创建日期**: 2026-02-03
