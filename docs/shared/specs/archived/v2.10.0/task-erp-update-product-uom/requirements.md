> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# UpdateProduct 增加 UOM 字段支持 需求文档

> 本文档定义 ERP UpdateProduct 接口增加 UOM 字段更新支持的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                     |
| ----------------- | ---------------------------------------------------------------------------------------- |
| **来源 Proposal** | [erp-update-product-uom.md](../../../../team/proposals/2025-11/erp-update-product-uom.md) |
| **创建日期**      | 2025-11-27                                                                               |
| **负责人**        | rikugun                                                                                  |
| **目标 Sprint**   | 待定                                                                                     |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)               |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 待审核 |
| **审核人**   | -      |
| **审核日期** | -      |
| **审核意见** | -      |

---

## 📋 概述

扩展 `ttpos-erp` 模块的 `UpdateProduct` gRPC 接口，增加对 `stock_uom`（库存计量单位）字段的更新支持，使商户能够通过 API 修改商品的计量单位。

## 🎯 产品对齐

- 完善 ERP 商品管理 API，支持更多商品属性的修改
- 减少后台运维操作，提升系统自动化程度
- 与 ERPNext Item 数据模型保持一致

## 📝 用户故事

**作为** 商户管理员  
**我想** 通过 API 更新商品的计量单位（UOM）  
**以便于** 灵活管理商品属性，无需登录 ERPNext 后台操作

---

## 功能需求

### Requirement 1: Protobuf 接口扩展

**用户故事**: 作为 API 调用方，我想在 UpdateProductReq 中传入 stock_uom 字段，以便于更新商品的计量单位

#### 验收标准

1. **WHEN** 调用 `UpdateProduct` 接口并传入有效的 `stock_uom` 值 **THEN** 系统 **SHALL** 成功更新商品的计量单位
2. **WHEN** 调用 `UpdateProduct` 接口不传入 `stock_uom` 字段 **THEN** 系统 **SHALL** 不修改商品的计量单位（保持现有行为）
3. **WHEN** 更新成功 **THEN** 响应 **SHALL** 在 `UpdateProductResp` 中返回更新后的 `stock_uom` 值

#### 具体要求

- [ ] 1.1 在 `UpdateProductReq` 消息中增加 `string stock_uom = 6;` 字段
- [ ] 1.2 在 `UpdateProductResp` 消息中增加 `string stock_uom = 6;` 字段
- [ ] 1.3 执行 `gf gen pb` 重新生成 API 代码

---

### Requirement 2: 业务逻辑实现

**用户故事**: 作为系统，我需要在更新商品时处理 stock_uom 字段，以便于将其同步到 ERPNext

#### 验收标准

1. **WHEN** 请求中包含非空的 `stock_uom` **THEN** 系统 **SHALL** 将其映射到 ERPNext Item 文档的 `stock_uom` 字段
2. **IF** 请求中 `stock_uom` 为空字符串 **THEN** 系统 **SHALL** 不修改该字段（保持 ERPNext 中的现有值）

#### 具体要求

- [ ] 2.1 修改 `sProduct.UpdateProduct` 方法，处理 `req.StockUom` 字段
- [ ] 2.2 当 `req.StockUom` 非空时，设置 `itemInfo.StockUom = req.StockUom`
- [ ] 2.3 在响应中返回更新后的 `stock_uom` 值

---

## 非功能需求

### 代码架构和模块化

- **遵循规范**: `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
- **Protobuf 规范**: `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### API 设计要求

- [x] 字段命名使用 snake_case（`stock_uom`）
- [x] 字段编号递增（使用 6）
- [x] 响应格式保持一致

### 性能要求

- [x] 无额外性能开销（仅增加一个字段的映射）

### 测试要求

- [ ] 测试传入有效 `stock_uom` 的更新场景
- [ ] 测试不传入 `stock_uom` 的兼容性场景
- [ ] 测试传入空字符串的场景

---

## 验收标准

### 功能验收

1. **接口定义**: `UpdateProductReq` 和 `UpdateProductResp` 包含 `stock_uom` 字段
2. **更新功能**: 传入有效 `stock_uom` 时能成功更新 ERPNext 中的商品计量单位
3. **向后兼容**: 不传入 `stock_uom` 时行为与原有逻辑一致

### 测试验收

1. **单元测试**: 覆盖新增字段的处理逻辑
2. **API 测试**: gRPC 接口测试通过

### 文档验收

1. **Protobuf 注释**: 字段有清晰的中文注释

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 修改 protobuf 后执行 `gf gen pb` 重新生成

### 业务约束

- ERPNext 对 `stock_uom` 更新可能有业务限制（如已有库存记录时的约束）
- UOM 值必须是 ERPNext 系统中已定义的有效单位

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go` - 业务逻辑实现
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go` - Item DTO（已包含 StockUom 字段）

### 服务依赖

- **BMP → ERPNext**: Document Update API

---

## 风险和缓解

### 风险 1: ERPNext 更新约束

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 查阅 ERPNext 文档确认 `stock_uom` 更新规则
- 测试时验证各种场景下的更新行为

---

## 时间表

- **Phase 1 - Protobuf 定义**: 0.25 天
- **Phase 2 - Logic 实现**: 0.25 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go 代码规范

### 现有代码参考

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto` - 现有接口定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go` - 现有实现
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go` - Item DTO

### 外部参考

- ERPNext Item DocType: https://github.com/frappe/erpnext/blob/develop/erpnext/stock/doctype/item/item.json

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: 待定

