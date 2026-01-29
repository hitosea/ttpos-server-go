# Lineman 订单金额单位转换 需求文档

## 📋 基本信息

| 项目              | 内容                                                                         |
| ----------------- | ---------------------------------------------------------------------------- |
| **Spec ID**       | story-bmp-lineman-currency-conversion                                        |
| **Level**         | task (技术任务)                                                              |
| **来源 Proposal** | [bmp-lineman-currency-conversion](../../../team/proposals/2026-01/bmp-lineman-currency-conversion.md) |
| **创建日期**      | 2026-01-22                                                                   |
| **负责人**        | -                                                                            |
| **目标 Sprint**   | -                                                                            |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已完成     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-22 |

---

## 📝 用户故事

**作为** 开发/运维团队
**我想** 在 BMP 外送模块的 Lineman 订单处理中，将金额从泰铢转换为分
**以便于** 确保数据准确性、系统一致性，以及下游系统（Main 模块、POS 端）能正确处理金额

---

## 功能需求

### Requirement 1: placeOrder 金额转换

**用户故事**: 作为开发团队，我想在 Lineman placeOrder Webhook 处理时自动转换金额单位，以便于新订单的金额存储正确

#### 验收标准

1. **WHEN** Lineman placeOrder Webhook 推送订单 **THEN** 系统 **SHALL** 将订单主表金额（`TotalAmount`、`Subtotal`）乘以 100 转换为分后存入数据库
2. **WHEN** Lineman placeOrder Webhook 推送订单 **THEN** 系统 **SHALL** 将订单明细金额（`Price`、`TotalPrice`）乘以 100 转换为分后存入数据库
3. **WHEN** 金额转换完成 **THEN** 系统 **SHALL** 在日志中记录原始金额和转换后金额，便于问题排查

---

### Requirement 2: orderUpdate 金额转换

**用户故事**: 作为开发团队，我想在 Lineman orderUpdate Webhook 处理时自动转换金额单位，以便于订单更新后的金额存储正确

#### 验收标准

1. **WHEN** Lineman orderUpdate Webhook 更新订单 **THEN** 系统 **SHALL** 将订单主表金额（`TotalAmount`、`Subtotal`）乘以 100 转换为分后更新数据库
2. **WHEN** Lineman orderUpdate Webhook 更新订单 **THEN** 系统 **SHALL** 将订单明细金额（`Price`、`TotalPrice`）乘以 100 转换为分后存入数据库
3. **WHEN** 金额转换完成 **THEN** 系统 **SHALL** 保持与 placeOrder 相同的转换逻辑，确保一致性

---

## 非功能需求

### 测试要求

- [x] 单元测试覆盖率 ≥ 80%
- [x] 编写针对金额转换逻辑的单元测试
- [x] 使用模拟 Webhook 数据进行集成测试

### 平台兼容性

- [x] ttpos-bmp/ttpos-takeout 模块
- [x] GoFrame v2.x 框架

### 代码位置

涉及修改的文件：
- `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - `saveOrder` 方法（第 70-173 行）
  - `updateOrder` 方法（第 236-291 行）

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/go-rules.mdc 规范
- 使用 `gerror` 处理错误（不用标准库 errors）
- 业务逻辑写在 `internal/logic/` 目录

### 转换规则

| 来源 | 单位 | 目标 | 单位 | 转换公式 |
|------|------|------|------|----------|
| Lineman API | 泰铢（元） | TTPOS 数据库 | 分 | 金额 × 100 |

### 涉及字段

| 表 | 字段 | 说明 |
|---|---|---|
| ttpos_takeout_order | total_amount | 订单总金额 |
| ttpos_takeout_order | subtotal | 订单小计 |
| ttpos_takeout_order_item | price | 商品单价 |
| ttpos_takeout_order_item | total_price | 商品总价 |

### 资源约束

- Story Point: 2

---

## 风险和缓解

### 风险 1: 历史数据金额错误

**影响**: 中
**缓解措施**: 提供数据修复 SQL 脚本，在部署后执行修复历史订单金额

### 风险 2: 精度丢失

**影响**: 低
**缓解措施**: 使用整数运算（乘以 100），避免浮点数精度问题；数据库字段使用 DECIMAL 类型

---

**版本**: v1.0.0
**创建日期**: 2026-01-22
