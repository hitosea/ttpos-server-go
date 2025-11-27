# ERP 品牌采购订单税率传值任务分解

> 本文档定义 ERP PO/SO 税率传值功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 14  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: DTO 扩展

- [ ] 1.0a 新增税费模板 DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/tax_template.go`（新建）
  - Purpose: 定义税费模板结构体，用于查询结果解析
  - Requirements: 0.1
  - Leverage: 现有 DTO 结构（如 `Company`）
  - Prompt: Role: Go Developer | Task: 创建 `tax_template.go` 文件，定义 `PurchaseTaxesAndChargesTemplate` 和 `SalesTaxesAndChargesTemplate` 结构体 | Context: 包含字段 `Name`, `Title`, `Company`, `IsDefault`, `Disabled`，均为可选字段 | Restrictions: 遵循现有代码风格 | Success: DTO 创建成功，可用于解析 ERPNext 查询结果

- [ ] 1.0b 新增 DocType 常量

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/consts.go`
  - Purpose: 添加税费模板 DocType 常量
  - Requirements: 0.2
  - Leverage: 现有 DocType 常量定义
  - Prompt: Role: Go Developer | Task: 在 `consts.go` 中添加 `DocTypePurchaseTaxesTemplate = "Purchase Taxes and Charges Template"` 和 `DocTypeSalesTaxesTemplate = "Sales Taxes and Charges Template"` | Context: 参考现有 DocType 常量风格 | Restrictions: 遵循现有命名规范 | Success: 常量添加成功

- [ ] 1.1 扩展 PurchaseOrder DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - Purpose: 添加 `TaxesAndCharges` 字段支持采购税费模板
  - Requirements: 1.1
  - Leverage: 现有 `POSInvoice` 结构体中的 `TaxesAndCharges` 字段定义
  - Prompt: Role: Go Developer | Task: 在 `PurchaseOrder` 结构体中添加 `TaxesAndCharges string` 字段，JSON tag 为 `taxes_and_charges,omitempty`，中文注释为"采购税费模板名称 (Purchase Taxes and Charges Template)" | Context: 参考 `selling.go` 中 `POSInvoice` 的 `TaxesAndCharges` 字段 | Restrictions: 遵循现有代码风格 | Success: 字段添加成功，JSON 序列化正确

- [ ] 1.2 确认/扩展 SaleOrder DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - Purpose: 确保 `SaleOrder` 结构体包含 `TaxesAndCharges` 字段
  - Requirements: 2.1
  - Leverage: 现有 `SaleOrder` 结构体
  - Prompt: Role: Go Developer | Task: 检查 `SaleOrder` 结构体是否包含 `TaxesAndCharges` 字段，如果没有则添加 | Context: 字段类型为 string，JSON tag 为 `taxes_and_charges,omitempty` | Restrictions: 遵循现有代码风格 | Success: 字段存在或添加成功

- [ ] 1.3 扩展 CreatePurchaseFromMqReq DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - Purpose: 添加税费参数到请求 DTO
  - Requirements: 1.2
  - Leverage: 现有 `CreatePurchaseFromMqReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 `CreatePurchaseFromMqReq` 结构体中添加 `TaxesAndCharges string` 和 `TaxCategory string` 字段 | Context: 两个字段都是可选的，使用 `omitempty` 标签 | Restrictions: 字段为可选，保持向后兼容 | Success: 字段添加成功

- [ ] 1.4 扩展 CreateInnerSaleOrderFromPurchaseOrderReq DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - Purpose: 添加税费参数到请求 DTO
  - Requirements: 2.2
  - Leverage: 现有 `CreateInnerSaleOrderFromPurchaseOrderReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 `CreateInnerSaleOrderFromPurchaseOrderReq` 结构体中添加 `TaxesAndCharges string` 和 `TaxCategory string` 字段 | Context: 两个字段都是可选的，使用 `omitempty` 标签 | Restrictions: 字段为可选，保持向后兼容 | Success: 字段添加成功

---

## Phase 2: Accounts 模块创建

- [ ] 2.0a 创建 Accounts 服务入口

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/accounts/accounts.go`（新建）
  - Purpose: 创建 Accounts 服务入口和注册
  - Requirements: 0.3, 0.4
  - Leverage: 现有 `logic/company/company.go` 服务注册模式
  - Prompt: Role: Go Developer | Task: 创建 `accounts/accounts.go` 文件，定义 `sAccounts` 结构体并注册到 service | Context: 参考 `company.go` 的服务注册方式 | Restrictions: 遵循现有服务注册规范 | Success: 服务注册成功

- [ ] 2.0b 实现税费模板查询方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/accounts/tax_template.go`（新建）
  - Purpose: 查询公司默认税费模板和获取模板详情
  - Requirements: 0.3, 0.4
  - Leverage: `service.Document().List` 和 `service.Document().Get` 方法
  - Prompt: Role: Go Developer | Task: 创建 `tax_template.go` 文件，实现四个方法：1) `GetDefaultPurchaseTaxTemplate` 查询默认采购模板名称 2) `GetDefaultSalesTaxTemplate` 查询默认销售模板名称 3) `GetPurchaseTaxTemplateDetails` 获取采购模板税费明细 4) `GetSalesTaxTemplateDetails` 获取销售模板税费明细 | Context: 使用 filters `[company, is_default=1, disabled=0]` 查询默认模板 | Restrictions: 查询失败时返回空值，不影响主流程 | Success: 能正确查询并返回税费模板信息

- [ ] 2.0c 生成 Service 接口

  - Command: `cd ttpos-bmp && gf gen service`
  - Purpose: 自动生成 `internal/service/accounts.go` 接口文件
  - Requirements: 0.3, 0.4
  - Leverage: GoFrame `gf gen service` 工具
  - Prompt: Role: Go Developer | Task: 在 `ttpos-bmp` 目录下执行 `gf gen service` 命令 | Context: 该命令会扫描 `internal/logic/accounts/` 目录，自动生成对应的 service 接口 | Restrictions: 不要手动修改 `internal/service/` 目录下的文件 | Success: `internal/service/accounts.go` 文件自动生成，包含 `IAccounts` 接口

---

## Phase 3: Buying 模块修改

- [ ] 3.0 实现税费明细转换方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 将模板中的税费明细转换为 PO/SO 可用的格式
  - Requirements: 0.3, 0.4
  - Leverage: 税费明细 DTO 结构
  - Prompt: Role: Go Developer | Task: 实现 `convertPurchaseTaxesToInterface` 和 `convertSalesTaxesToInterface` 方法 | Context: 复制必要字段（ChargeType, AccountHead, Rate 等），清除计算字段（TaxAmount, Total 等由 ERPNext 自动计算） | Restrictions: 只复制配置字段，不复制计算结果 | Success: 税费明细能正确复制到 PO/SO

- [ ] 3.1 修改 CreatePurchaseFromMq 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 将税费参数传递给 PurchaseOrder，支持自动获取公司默认配置并复制税费明细
  - Requirements: 0.5, 1.3, 3.1-3.3
  - Leverage: 现有 `CreatePurchaseFromMq` 方法实现, `service.Accounts()` 服务
  - Prompt: Role: Go Developer | Task: 在 `CreatePurchaseFromMq` 方法中，设置 `purchaseOrder.TaxesAndCharges` 和 `purchaseOrder.Taxes` | Context: 1) 如果 `req.TaxesAndCharges != ""` 则使用传入的值，否则调用 `service.Accounts().GetDefaultPurchaseTaxTemplate()` 2) 获取模板名称后调用 `service.Accounts().GetPurchaseTaxTemplateDetails()` 获取税费明细 3) 复制到 PO | Restrictions: 优先使用传入参数，只复制配置字段不复制计算字段 | Success: PO 创建时自动应用税费模板和税费明细

- [ ] 3.2 修改 CreateInnerSaleOrderFromPurchaseOrder 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 将税费参数传递给 SaleOrder，支持自动获取公司默认配置并复制税费明细
  - Requirements: 0.6, 2.3, 3.1-3.3
  - Leverage: 现有 `CreateInnerSaleOrderFromPurchaseOrder` 方法实现, `service.Accounts()` 服务
  - Prompt: Role: Go Developer | Task: 在 `CreateInnerSaleOrderFromPurchaseOrder` 方法中，设置 `salesOrder.TaxesAndCharges` 和 `salesOrder.Taxes` | Context: 1) 如果 `req.TaxesAndCharges != ""` 则使用传入的值，否则调用 `service.Accounts().GetDefaultSalesTaxTemplate()` 2) 获取模板名称后调用 `service.Accounts().GetSalesTaxTemplateDetails()` 获取税费明细 3) 复制到 SO | Restrictions: 优先使用传入参数，只复制配置字段不复制计算字段 | Success: SO 创建时自动应用税费模板和税费明细

---

## Phase 4: 测试

- [ ] 4.1 编写 CreatePurchaseFromMq 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go`
  - Purpose: 测试税费参数处理逻辑和自动获取公司默认配置
  - Requirements: 0.1-0.3, 1.1-1.3, 3.1-3.3
  - Leverage: 现有测试文件（如果存在）
  - Prompt: Role: QA Engineer | Task: 为 `CreatePurchaseFromMq` 编写单元测试，覆盖：1) 传入 TaxesAndCharges 参数 2) 不传入但公司有默认配置 3) 不传入且公司无默认配置 4) 传入参数优先于公司默认配置 | Context: 使用 mock 模拟 ERPNext API 和 Accounts 服务 | Restrictions: 测试所有场景和优先级 | Success: 测试覆盖所有场景

- [ ] 4.2 编写 CreateInnerSaleOrderFromPurchaseOrder 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go`
  - Purpose: 测试税费参数处理逻辑和自动获取公司默认配置
  - Requirements: 0.1-0.2, 0.4, 2.1-2.3, 3.1-3.3
  - Leverage: 现有测试文件（如果存在）
  - Prompt: Role: QA Engineer | Task: 为 `CreateInnerSaleOrderFromPurchaseOrder` 编写单元测试，覆盖：1) 传入 TaxesAndCharges 参数 2) 不传入但公司有默认配置 3) 不传入且公司无默认配置 4) 传入参数优先于公司默认配置 | Context: 使用 mock 模拟 ERPNext API 和 Accounts 服务 | Restrictions: 测试所有场景和优先级 | Success: 测试覆盖所有场景

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/task-erp-po-so-tax-rate/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/task-erp-po-so-tax-rate/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/task-erp-po-so-tax-rate/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

**模板版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: rikugun

