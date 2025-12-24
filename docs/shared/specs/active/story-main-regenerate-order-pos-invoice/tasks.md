# 重新生成订单POS发票 任务分解

> 本文档定义重新生成订单POS发票功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 命令行工具框架

- [ ] 1.1 创建命令文件框架

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 创建命令行工具的基础框架，包括命令定义、参数解析和初始化逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有命令: `main/command/regenerate_sale_bill_material_outbound.go`
  - Prompt: Role: Go Developer specializing in CLI tools | Task: 创建 regenerate-order-pos-invoice 命令的基础框架，参考 regenerate_sale_bill_material_outbound.go 的结构 | Context: 使用 Cobra 框架，定义命令名称、描述、参数（company-uuid, sale-order-uuid, dry-run），实现 PreRun 初始化逻辑（配置、日志、数据库、缓存等） | Restrictions: 遵循 .cursor/rules/go-main.mdc，命令文件放在 main/command/ 目录 | Success: 命令框架创建成功，参数解析正确，PreRun 初始化完整

---

## Phase 2: 订单信息读取和验证

- [ ] 2.1 实现订单信息读取逻辑

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 从数据库读取订单和账单完整信息
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有 Repository: `main/app/repository/sale_order.go` (`GetSaleOrderByUuid`), `main/app/repository/order.go` (`GetSaleBillAllInfo`)
  - Prompt: Role: Go Developer | Task: 实现订单信息读取逻辑，读取 saleOrder 和 saleBill | Context: 使用 SaleOrderRepo.GetSaleOrderByUuid() 读取订单，通过 saleOrder.SaleBillUuid 获取账单UUID，使用 OrderRepo.GetSaleBillAllInfo() 读取账单完整信息 | Restrictions: 验证订单存在性（saleOrder.Uuid != 0），验证订单状态（saleOrder.FinishTime > 0），处理错误情况 | Success: 订单信息读取正确，包含所有必要的关联数据

- [ ] 2.2 实现订单状态验证

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 验证订单是否已完成结账
  - Requirements: 2.3
  - Leverage: 现有 Model: `main/app/model/sale_order.go`
  - Prompt: Role: Go Developer | Task: 实现订单状态验证逻辑，确保订单已完成结账 | Context: 检查 saleOrder.FinishTime > 0，如果未结账则返回明确的错误信息 | Restrictions: 仅支持已完成结账的订单 | Success: 订单状态验证正确，错误提示明确

---

## Phase 3: ERP配置验证

- [ ] 3.1 实现ERP配置验证逻辑

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 验证ERP Phase3是否启用，SiteCode是否配置
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 Service: `main/app/service/setting.go` (获取公司设置)
  - Prompt: Role: Go Developer | Task: 实现ERP配置验证逻辑，检查ERP Phase3和SiteCode配置 | Context: 获取公司信息和设置（ctx.GetCompany(), ctx.GetCompanySetting()），检查 company.IsOpenErpPhase3() 是否为 true，检查 companySetting.ErpnextSiteCode 是否不为空 | Restrictions: 提供明确的错误提示信息 | Success: ERP配置验证正确，错误提示明确

---

## Phase 4: 发票生成和更新

- [ ] 4.1 实现SavePosInvoice方法调用

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 调用 SavePosInvoice 方法生成发票
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: 现有 Service: `main/app/service/order.go` (`SavePosInvoice` 方法)
  - Prompt: Role: Go Developer | Task: 实现 SavePosInvoice 方法调用逻辑 | Context: 创建 gin.Context 对象（命令行环境），初始化 OrderSrv 服务实例，调用 orderSrv.SavePosInvoice(ctx, saleOrder, saleBill, db) 方法，处理返回结果 | Restrictions: 处理 SavePosInvoice 返回的错误（包括班次已交班的情况），记录详细的日志信息 | Success: SavePosInvoice 调用成功，发票已保存到ERP系统

- [ ] 4.2 实现订单发票信息更新

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 更新订单的发票名称字段
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有 Repository: `main/app/repository/sale_order.go` (`UpdateSaleOrderErpInvoice` 方法)
  - Prompt: Role: Go Developer | Task: 实现订单发票信息更新逻辑 | Context: 使用 SaleOrderRepo.UpdateSaleOrderErpInvoice() 方法更新发票名称，更新字段：ErpProductsInvoiceName、ErpMaterialInvoiceName | Restrictions: 发票保存成功后才更新订单信息，如果更新失败需要记录错误日志 | Success: 订单发票信息更新成功，字段值正确

---

## Phase 5: 错误处理和日志记录

- [ ] 5.1 实现错误处理和日志记录

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 完善的错误处理和日志记录
  - Requirements: 6.1, 6.2, 6.3, 6.4, 6.5
  - Leverage: 现有工具: `logger.Logger`, 参考 `regenerate_sale_bill_material_outbound.go` 的日志输出格式
  - Prompt: Role: Go Developer | Task: 实现错误处理和日志记录逻辑 | Context: 使用 logger.Logger 记录操作日志，使用颜色输出（成功绿色、错误红色、警告黄色、信息蓝色），显示操作耗时和结果统计，提供用户确认机制（非 --dry-run 模式） | Restrictions: 所有错误都要返回明确的错误信息，记录详细的日志便于排查 | Success: 错误处理完善，日志记录详细，输出格式清晰

---

## Phase 6: 预览模式实现

- [ ] 6.1 实现 dry-run 预览模式

  - File: `main/command/regenerate_order_pos_invoice.go`
  - Purpose: 实现预览模式，不实际执行发票保存
  - Requirements: 1.4, 2.5, 3.5, 4.5, 5.4
  - Leverage: 参考 `regenerate_sale_bill_material_outbound.go` 的 dry-run 实现
  - Prompt: Role: Go Developer | Task: 实现 dry-run 预览模式，显示操作预览信息 | Context: 当 --dry-run 参数为 true 时，读取订单信息和ERP配置，显示订单基本信息（订单号、金额等）、ERP配置信息（SiteCode、CompanyAbbr等）、预计操作（将调用 SavePosInvoice 方法生成发票），但不实际执行发票保存和订单更新 | Restrictions: 预览模式要显示足够的信息，让用户了解将要执行的操作 | Success: 预览模式工作正常，显示信息完整

---

## Phase 7: 测试和文档

- [ ] 7.1 编写命令行工具测试

  - File: `main/command/regenerate_order_pos_invoice_test.go` (可选)
  - Purpose: 测试命令行工具的各种场景
  - Requirements: 6.1, 6.2, 6.3, 6.4
  - Leverage: 现有测试: `main/command/` 目录下的测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 regenerate-order-pos-invoice 命令编写测试，覆盖所有参数组合和错误场景 | Context: 测试参数验证、订单读取、ERP配置验证、发票生成、订单更新、预览模式等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 7.2 编写集成测试

  - File: 集成测试文件（可选）
  - Purpose: 测试完整的发票生成流程
  - Requirements: 6.1, 6.2, 6.3, 6.4
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Engineer | Task: 编写集成测试，测试完整的发票生成流程 | Context: 创建测试数据（订单、账单），执行重新生成发票操作，验证发票已保存到ERP系统，验证订单发票名称已更新 | Restrictions: 测试环境需要配置ERP系统 | Success: 集成测试通过，发票生成流程正确

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: xiezhihuan

