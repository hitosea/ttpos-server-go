# 采购单列表支持多 Name 查询 需求文档

## 📋 基本信息

| 项目              | 内容                                                                                      |
| ----------------- | ----------------------------------------------------------------------------------------- |
| **Spec ID**       | task-all-purchase-order-multi-name-query                                                  |
| **来源 Proposal** | [shop-purchase-order-multi-name-query](../../../team/proposals/2026-01/shop-purchase-order-multi-name-query.md) |
| **创建日期**      | 2026-01-29                                                                                |
| **负责人**        | rikugun                                                                                   |
| **目标版本**      | v2.16                                                                                     |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 已完成 |
| **审核人**   | -      |
| **审核日期** | -      |

---

## 📝 用户故事

**作为** 前端开发者
**我想** 在调用采购单列表接口时支持传入多个 name 进行筛选
**以便于** 减少 API 调用次数，匹配前端多选筛选能力，提升用户查询体验

---

## 功能需求

### Requirement 1: name 字段支持 IN 查询

**用户故事**: 作为前端开发者，我想传入多个 name 值进行筛选，以便于一次请求获取多个条件的结果

#### 验收标准

1. **WHEN** `GetPurchaseOrderListReq.name` 包含逗号分隔的多个值（如 `"供应商A,供应商B,供应商C"`）**THEN** 系统 **SHALL** 使用 IN 查询返回 name 匹配任一值的所有记录
2. **WHEN** `GetPurchaseOrderListReq.name` 为单个值（不含逗号）**THEN** 系统 **SHALL** 保持原有等值查询行为，确保向后兼容
3. **WHEN** `GetPurchaseOrderListReq.name` 为空字符串或未传入 **THEN** 系统 **SHALL** 不应用 name 筛选条件

### Requirement 2: 实现方式

**技术实现**: 在 `buildPurchaseOrderListFilters` 方法中修改 name 字段的 filter 构建逻辑

#### 验收标准

1. **WHEN** name 存在且包含逗号 **THEN** 系统 **SHALL** 生成 filter `{"name", "in", "value1,value2,value3"}`
2. **WHEN** name 存在且不含逗号 **THEN** 系统 **SHALL** 生成 filter `{"name", "=", "value"}`（或保持原有逻辑）

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 包含多 name 查询的集成测试用例

### 性能要求

- [ ] IN 查询支持最多 50 个值
- [ ] 查询响应时间不应因多值查询显著增加

### 平台兼容性

- [x] 所有调用 `buying.GetPurchaseOrderListReq` 接口的终端

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM（Main 模块）或 GoFrame 2.x（BMP 模块）
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范

### 资源约束

- Story Point: 1（改动范围小，风险可控）

---

## 风险和缓解

### 风险 1: IN 查询值过多影响性能

**影响**: 低
**缓解措施**: 限制 IN 查询的最大值数量（建议不超过 50 个）

### 风险 2: 前后端参数格式不一致

**影响**: 中
**缓解措施**: 与前端确认参数传递格式，确保使用逗号分隔的字符串

---

**版本**: v1.0.0
**创建日期**: 2026-01-29
