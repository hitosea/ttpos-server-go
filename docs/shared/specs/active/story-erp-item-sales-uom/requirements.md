# ERP 商品销售单位字段 需求文档

## 📋 基本信息

| 项目              | 内容                                                      |
| ----------------- | --------------------------------------------------------- |
| **Spec ID**       | story-erp-item-sales-uom                                  |
| **Level**         | story                                                     |
| **来源 Proposal** | [erp-item-sales-uom](../../../team/proposals/2026-01/erp-item-sales-uom.md) |
| **任务 ID**       | 38954                                                     |
| **目标版本**      | v2.15                                                     |
| **创建日期**      | 2026-01-16                                                |
| **负责人**        | rikugun                                                   |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 开发中 |
| **审核人**   | -      |
| **审核日期** | -      |

---

## 📝 用户故事

**作为** 系统集成服务
**我想** 在商品信息中包含销售单位（sales_uom）字段
**以便于** 支持库存单位与销售单位不同的业务场景，保持与 ERP 系统数据完整一致

---

## 功能需求

### Requirement 1: 新增 sales_uom 字段

**用户故事**: 作为系统集成服务，我想在 ItemInfo 消息中有 sales_uom 字段，以便于传递商品的销售单位信息

#### 验收标准

1. **WHEN** 在 `item.proto` 的 `ItemInfo` 消息中 **THEN** 系统 **SHALL** 包含 `optional string sales_uom` 字段
2. **WHEN** 调用 `GetItemList` 接口 **THEN** 系统 **SHALL** 在返回的 `ItemInfo` 中包含 `sales_uom` 字段（如果有值）
3. **WHEN** 调用 `GetItem` 接口 **THEN** 系统 **SHALL** 在返回的 `ItemInfo` 中包含 `sales_uom` 字段（如果有值）
4. **WHEN** 调用 `SaveItem` 接口并传入 `sales_uom` 字段 **THEN** 系统 **SHALL** 正确保存该字段值到 ERP
5. **WHEN** 调用 `SaveItem` 接口未传入 `sales_uom` 字段 **THEN** 系统 **SHALL** 保持原有逻辑，不影响其他字段

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖新增字段的查询逻辑
- [ ] 单元测试覆盖新增字段的保存逻辑

### 兼容性要求

- [x] 向后兼容：使用 `optional` 关键字，不破坏现有调用方

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/ 规范
- 使用 `optional` 关键字确保字段可选

### 实现范围

- 仅实现 ttpos-erp 模块部分
- 涉及文件：
  - `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto`
  - `ttpos-bmp/app/ttpos-erp/internal/logic/` 相关 logic 文件

### 资源约束

- Story Point: 1（极简单，0.5-1 天）

---

## 风险和缓解

### 风险 1: 现有调用方需要适配

**影响**: 低
**缓解措施**: 使用 `optional` 关键字，新字段不会影响现有调用方的兼容性

---

**版本**: v1.0.0
**创建日期**: 2026-01-16
