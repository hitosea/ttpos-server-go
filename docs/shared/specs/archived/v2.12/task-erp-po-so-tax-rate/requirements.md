> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# ERP 品牌采购订单税率传值需求文档

> 本文档定义 ERP PO/SO 税率传值功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-27-erp-po-so-tax-rate.md](../../../team/proposals/2025-11-27-erp-po-so-tax-rate.md)         |
| **任务编号**      | 37110                                                                                                                    |
| **创建日期**      | 2025-11-27                                                                                                               |
| **负责人**        | rikugun                                                                                                                  |
| **目标 Sprint**   | TBD                                                                                                                      |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                                                                                      |

---

## 📋 概述

在品牌采购流程中，从物料请求（Material Request）创建采购订单（PO）和内部销售订单（SO）时，需要支持传递税率相关参数，以便 ERPNext 能够正确计算订单的含税金额。

**核心需求变更**：当未传入税费模板参数时，系统应自动从 ERPNext 查询对应公司的默认税费模板配置。

## 🎯 产品对齐

该功能支持品牌采购业务的财务合规需求，确保订单金额计算准确，税务信息完整记录。

## 📝 用户故事

**作为** 品牌采购管理人员  
**我想** 在创建采购订单和内部销售订单时能够传递税率信息  
**以便于** 订单金额计算准确，满足财务合规要求

---

## 功能需求

### Requirement 0: 自动获取公司默认税费模板

**用户故事**: 作为品牌采购人员，我希望系统能自动使用公司配置的默认税费模板，无需每次手动指定

#### 验收标准

1. **WHEN** 调用 `CreatePurchaseFromMq` 未传入 `TaxesAndCharges` **THEN** 系统 **SHALL** 查询公司默认 `Purchase Taxes and Charges Template` 并自动应用
2. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 未传入 `TaxesAndCharges` **THEN** 系统 **SHALL** 查询公司默认 `Sales Taxes and Charges Template` 并自动应用
3. **IF** 税费模板中 `taxes` 子表不为空 **THEN** 系统 **SHALL** 将税费明细复制到对应的 PO/SO 中
4. **IF** 公司未配置默认税费模板 **THEN** 系统 **SHALL** 不设置税费模板字段（保持原有行为）

#### 具体要求

- [ ] 0.1 新增税费模板 DTO（`PurchaseTaxesAndChargesTemplate` 和 `SalesTaxesAndChargesTemplate`）
- [ ] 0.2 新增 DocType 常量（`DocTypePurchaseTaxesTemplate` 和 `DocTypeSalesTaxesTemplate`）
- [ ] 0.3 实现 `getDefaultPurchaseTaxTemplate` 方法，查询公司默认采购税费模板
- [ ] 0.4 实现 `getDefaultSalesTaxTemplate` 方法，查询公司默认销售税费模板
- [ ] 0.5 修改 `CreatePurchaseFromMq` 方法，未传入参数时自动查询公司默认采购税费模板
- [ ] 0.6 修改 `CreateInnerSaleOrderFromPurchaseOrder` 方法，未传入参数时自动查询公司默认销售税费模板

#### ERPNext 税费模板 DocType

| DocType 名称 | 用途 | 查询条件 |
|-------------|------|----------|
| `Purchase Taxes and Charges Template` | 采购税费模板 | `company = {公司}` AND `is_default = 1` |
| `Sales Taxes and Charges Template` | 销售税费模板 | `company = {公司}` AND `is_default = 1` |

---

### Requirement 1: 采购订单 (PO) 税率传值

**用户故事**: 作为品牌采购人员，我想在创建采购订单时指定税费模板，以便于订单包含正确的税费信息

#### 验收标准

1. **WHEN** 调用 `CreatePurchaseFromMq` 时传入 `TaxesAndCharges` (Purchase Taxes and Charges Template 名称) **THEN** 系统 **SHALL** 查询该模板的 `taxes` 明细并复制到 PO 中
2. **WHEN** 调用 `CreatePurchaseFromMq` 时传入 `TaxCategory` **THEN** 系统 **SHALL** 在创建的 PO 中设置对应的税类别

#### 具体要求

- [ ] 1.1 在 `PurchaseOrder` 结构体中添加 `TaxesAndCharges` 字段
- [ ] 1.2 在 `CreatePurchaseFromMqReq` 中添加 `TaxesAndCharges` 和 `TaxCategory` 参数
- [ ] 1.3 修改 `CreatePurchaseFromMq` 方法，将税率参数传递给 `PurchaseOrder`

---

### Requirement 2: 销售订单 (SO) 税率传值

**用户故事**: 作为品牌采购人员，我想在创建内部销售订单时指定税费模板，以便于 SO 包含正确的税费信息

#### 验收标准

1. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 时传入 `TaxesAndCharges` (Sales Taxes and Charges Template 名称) **THEN** 系统 **SHALL** 查询该模板的 `taxes` 明细并复制到 SO 中
2. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 时传入 `TaxCategory` **THEN** 系统 **SHALL** 在创建的 SO 中设置对应的税类别

#### 具体要求

- [ ] 2.1 确保 `SaleOrder` 结构体中包含 `TaxesAndCharges` 字段
- [ ] 2.2 在 `CreateInnerSaleOrderFromPurchaseOrderReq` 中添加 `TaxesAndCharges` 和 `TaxCategory` 参数
- [ ] 2.3 修改 `CreateInnerSaleOrderFromPurchaseOrder` 方法，将税率参数传递给 `SaleOrder`

---

### Requirement 3: 向后兼容与默认行为

**用户故事**: 作为系统集成方，我希望现有的调用方式不受影响，同时能够享受自动税费配置的便利

#### 验收标准

1. **IF** 未传入税率相关参数 **AND** 公司配置了默认税费模板 **THEN** 系统 **SHALL** 自动使用公司默认配置
2. **IF** 未传入税率相关参数 **AND** 公司未配置默认税费模板 **THEN** 系统 **SHALL** 保持原有行为（不设置税费字段）
3. **IF** 传入了 `TaxesAndCharges` 参数 **THEN** 系统 **SHALL** 优先使用传入的值，忽略公司默认配置

#### 具体要求

- [ ] 3.1 所有税率参数均为可选字段
- [ ] 3.2 传入的参数优先于公司默认配置
- [ ] 3.3 公司未配置默认模板时，保持原有行为

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 遵循 ttpos-bmp Logic 层规范
- **单一职责原则**: 税率处理逻辑封装在现有方法内
- **模块化设计**: 不新增独立模块，扩展现有功能

### 遵循规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范

### 测试要求

- [ ] 修改的方法需有单元测试覆盖
- [ ] 测试含税率参数和不含税率参数两种场景

---

## 验收标准

### 功能验收

1. **自动税费模板**: 不传入税费参数时，系统自动查询公司默认配置并应用
2. **PO 税费模板**: 调用 API 时传入 `TaxesAndCharges`，创建的 PO 在 ERPNext 中正确显示税费信息
3. **SO 税费模板**: 调用 API 时传入 `TaxesAndCharges`，创建的 SO 在 ERPNext 中正确显示税费信息
4. **优先级正确**: 传入参数优先于公司默认配置
5. **向后兼容**: 公司未配置默认模板时，功能正常

### 测试验收

1. **单元测试**: 覆盖新增参数处理逻辑
2. **集成测试**: 调用 ERPNext API 验证文档创建正确

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### ERPNext 约束

- `taxes_and_charges` 字段值必须是 ERPNext 中已存在的模板名称
- 采购订单使用 **Purchase Taxes and Charges Template**
- 销售订单使用 **Sales Taxes and Charges Template**

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3

---

## 依赖关系

### 技术依赖

- ERPNext API 文档创建接口
- ERPNext Company 文档查询接口
- 现有 `CreatePurchaseFromMq` 方法
- 现有 `CreateInnerSaleOrderFromPurchaseOrder` 方法
- 现有 `Company.GetCompany` 方法

### 服务依赖

- **ttpos-erp → ERPNext**: HTTP API 调用

---

## 风险和缓解

### 风险 1: 税费模板名称不存在

**影响**: 中  
**概率**: 中  
**缓解措施**:

- ERPNext API 会返回错误，需正确处理并返回给调用方
- 建议调用方先查询可用的税费模板列表

### 风险 2: 公司未配置默认税费模板

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 系统检测到公司未配置默认模板时，保持原有行为
- 日志记录警告信息，提示管理员配置默认模板
- 支持调用方显式传入模板名称覆盖

### 风险 3: 查询公司配置增加 API 调用开销

**影响**: 低  
**概率**: 高  
**缓解措施**:

- 在创建 PO/SO 时一并查询公司信息，复用现有查询
- 考虑缓存公司默认税费配置（后续优化）

---

## 时间表

- **Phase 1 - DTO 扩展**: 0.5 天（Company DTO + 请求 DTO）
- **Phase 2 - 公司默认配置查询**: 0.5 天（新增获取公司默认税费模板方法）
- **Phase 3 - 业务逻辑修改**: 1 天（PO/SO 创建逻辑）
- **Phase 4 - 测试和验证**: 1 天
- **总计**: 3 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范

### 相关代码

- `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` - 采购业务逻辑
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go` - 采购 DTO
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/selling.go` - 销售 DTO

### ERPNext 文档

- Purchase Taxes and Charges Template
- Sales Taxes and Charges Template

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: TBD

