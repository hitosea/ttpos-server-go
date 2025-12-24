# 重新生成订单POS发票 需求文档

> 本文档定义重新生成订单POS发票功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/regenerate-order-pos-invoice.md](../../../../team/proposals/2025-12/regenerate-order-pos-invoice.md) |
| **创建日期**      | 2025-12-16                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | xiezhihuan             |
| **审核日期** | 2025-12-16             |
| **审核意见** | 命令行工具功能，技术实现明确，可直接进入设计阶段         |

---

## 📋 概述

提供一个命令行工具命令 `regenerate-order-pos-invoice`，用于重新生成指定销售订单的POS发票。当订单的POS发票因ERP系统异常、网络问题等原因未能正确生成或保存时，可以通过此工具快速重新生成发票，无需重新走完整结账流程。

**核心价值**：
- **提升运维效率**：快速修复发票问题，无需重新走完整结账流程
- **降低数据风险**：避免重新结账带来的数据重复处理风险
- **减少操作时间**：通过命令行工具快速完成发票重新生成
- **提高数据准确性**：确保订单发票数据与ERP系统保持一致

**功能范围**：
- ✅ 读取订单信息（`saleOrder`、`saleBill`）
- ✅ 验证订单状态和ERP配置
- ✅ 调用 `SavePosInvoice` 方法生成发票
- ✅ 更新订单的发票信息（`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`）
- ✅ 支持 `--dry-run` 预览模式
- ✅ 提供安全确认机制和详细日志输出

## 🎯 产品对齐

本功能支持以下产品目标：
- **数据准确性**：确保订单发票数据与ERP系统保持一致，为财务核算提供准确的数据源
- **运维效率**：减少发票修复的复杂度和时间成本，提升系统可维护性
- **问题修复**：支持快速修复历史订单的发票问题，无需重新结账

## 📝 用户故事

**作为** 技术支持人员/运维人员  
**我想** 通过命令行工具重新生成订单的POS发票  
**以便于** 快速修复发票问题，无需重新走完整结账流程

---

## 功能需求

### Requirement 1: 命令行工具参数和验证

**用户故事**: 作为技术支持人员，我想通过命令行工具指定订单信息，以便于重新生成发票

#### 验收标准

1. **WHEN** 执行 `regenerate-order-pos-invoice --company-uuid {uuid} --sale-order-uuid {uuid} --open-pos-entry-name {name}` **THEN** 系统 **SHALL** 解析参数并验证有效性
2. **IF** `--company-uuid` 参数缺失或无效 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **IF** `--sale-order-uuid` 参数缺失或无效 **THEN** 系统 **SHALL** 提示错误信息并退出
4. **IF** `--open-pos-entry-name` 参数缺失或为空 **THEN** 系统 **SHALL** 提示错误信息并退出
5. **WHEN** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 进入预览模式，不实际执行发票保存
6. **WHEN** 参数验证通过 **THEN** 系统 **SHALL** 继续执行后续步骤

#### 具体要求

- [ ] 1.1 使用 Cobra 框架创建命令行工具
- [ ] 1.2 定义 `--company-uuid` 参数（必填，uint64类型）
- [ ] 1.3 定义 `--sale-order-uuid` 参数（必填，uint64类型）
- [ ] 1.4 定义 `--open-pos-entry-name` 参数（必填，string类型）
- [ ] 1.5 定义 `--dry-run` 参数（可选，bool类型，默认false）
- [ ] 1.6 在 `PreRun` 中初始化配置、数据库、日志等基础设施
- [ ] 1.6 参考 `regenerate-sale-bill-material-outbound` 命令的实现方式

---

### Requirement 2: 订单信息读取和验证

**用户故事**: 作为技术支持人员，我想读取订单完整信息，以便于生成发票

#### 验收标准

1. **WHEN** 提供有效的 `sale-order-uuid` **THEN** 系统 **SHALL** 从数据库读取订单完整信息
2. **IF** 订单不存在或已删除 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **WHEN** 订单读取成功 **THEN** 系统 **SHALL** 获取 `saleOrder` 和 `saleBill` 信息
4. **IF** 订单未完成结账 **THEN** 系统 **SHALL** 提示错误信息并退出（仅支持已结账订单）
5. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 显示订单基本信息（订单号、金额等）

#### 具体要求

- [ ] 2.1 使用 `OrderRepo.GetSaleBillAllInfo()` 或类似方法获取订单完整信息
- [ ] 2.2 验证订单存在性：检查 `saleOrder.Uuid != 0`
- [ ] 2.3 验证订单状态：确保订单已完成结账（`saleOrder.FinishTime > 0`）
- [ ] 2.4 获取关联的 `saleBill` 信息（通过 `saleOrder.SaleBillUuid`）
- [ ] 2.5 预加载必要的关联数据：商品、支付方式、会员信息等

---

### Requirement 3: ERP配置验证

**用户故事**: 作为技术支持人员，我想验证ERP配置是否启用，以便于确保发票可以正常生成

#### 验收标准

1. **WHEN** 订单信息读取成功 **THEN** 系统 **SHALL** 检查ERP配置是否启用
2. **IF** ERP Phase3 未启用（`company.IsOpenErpPhase3() == false`） **THEN** 系统 **SHALL** 提示错误信息并退出
3. **IF** ERP SiteCode 未配置（`companySetting.ErpnextSiteCode == ""`） **THEN** 系统 **SHALL** 提示错误信息并退出
4. **WHEN** ERP配置验证通过 **THEN** 系统 **SHALL** 继续执行发票生成流程
5. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 显示ERP配置信息（SiteCode、CompanyAbbr等）

#### 具体要求

- [ ] 3.1 获取公司信息：`ctx.GetCompany()`
- [ ] 3.2 获取公司设置：`ctx.GetCompanySetting()`
- [ ] 3.3 检查 `company.IsOpenErpPhase3()` 是否为 true
- [ ] 3.4 检查 `companySetting.ErpnextSiteCode` 是否不为空
- [ ] 3.5 提供明确的错误提示信息

---

### Requirement 4: 调用SavePosInvoice方法生成发票

**用户故事**: 作为技术支持人员，我想调用SavePosInvoice方法生成发票，以便于保存发票到ERP系统

#### 验收标准

1. **WHEN** 订单信息和ERP配置验证通过 **THEN** 系统 **SHALL** 调用 `orderSrv.SavePosInvoice()` 方法
2. **IF** `SavePosInvoice` 方法返回错误 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **WHEN** 发票保存成功 **THEN** 系统 **SHALL** 获取返回的发票名称（`ProductsInvoiceName`、`MaterialInvoiceName`）
4. **IF** 班次已交班（`SavePosInvoice` 内部检查） **THEN** 系统 **SHALL** 提示错误信息（可能需要特殊处理）
5. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 仅预览操作，不实际调用 `SavePosInvoice` 方法

#### 具体要求

- [ ] 4.1 创建 `gin.Context` 对象（命令行环境）
- [ ] 4.2 初始化 `orderSrv` 服务实例
- [ ] 4.3 调用 `orderSrv.SavePosInvoice(ctx, saleOrder, saleBill, db)` 方法
- [ ] 4.4 处理返回结果：`(*selling.SavePosInvoiceResp, error)`
- [ ] 4.5 记录详细的日志信息（成功/失败）
- [ ] 4.6 参考 `order_pay.go:929-939` 的实现逻辑

---

### Requirement 5: 更新订单发票信息

**用户故事**: 作为技术支持人员，我想更新订单的发票信息，以便于记录发票名称

#### 验收标准

1. **WHEN** 发票保存成功 **THEN** 系统 **SHALL** 更新订单的 `ErpProductsInvoiceName` 和 `ErpMaterialInvoiceName` 字段
2. **IF** 更新操作失败 **THEN** 系统 **SHALL** 提示错误信息（但发票已保存到ERP）
3. **WHEN** 更新成功 **THEN** 系统 **SHALL** 显示更新结果
4. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 仅预览更新操作，不实际执行

#### 具体要求

- [ ] 5.1 使用 `SaleOrderRepo.UpdateSaleOrderErpInvoice()` 方法更新发票名称
- [ ] 5.2 更新字段：`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`
- [ ] 5.3 使用事务确保数据一致性（如需要）
- [ ] 5.4 参考 `order_pay.go:936` 的实现逻辑

---

### Requirement 6: 错误处理和日志记录

**用户故事**: 作为技术支持人员，我想看到详细的错误信息和日志，以便于排查问题

#### 验收标准

1. **WHEN** 任何步骤发生错误 **THEN** 系统 **SHALL** 返回明确的错误信息
2. **WHEN** 操作成功完成 **THEN** 系统 **SHALL** 显示成功信息和发票名称
3. **WHEN** 执行过程中 **THEN** 系统 **SHALL** 记录详细的日志信息
4. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 显示预览信息（订单号、ERP配置、预计操作等）

#### 具体要求

- [ ] 6.1 使用 `logger.Logger` 记录操作日志
- [ ] 6.2 使用颜色输出（成功绿色、错误红色、警告黄色、信息蓝色）
- [ ] 6.3 显示操作耗时和结果统计
- [ ] 6.4 提供用户确认机制（非 `--dry-run` 模式）
- [ ] 6.5 参考 `regenerate-sale-bill-material-outbound` 命令的日志输出格式

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Command → Service → Repository 分层
- **单一职责原则**: 命令行工具只负责参数解析和调用服务方法
- **模块化设计**: 复用现有的 `SavePosInvoice` 方法，不重复实现
- **依赖管理**: Command 调用 Service 接口，Service 调用 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 不涉及API接口（命令行工具）

### 数据库设计要求

- [ ] 不涉及数据库表结构变更
- [ ] 使用现有的订单表和发票字段

### 性能要求

- [ ] 命令执行时间 < 5秒（正常情况下）
- [ ] 数据库查询优化（使用索引）
- [ ] 支持并发执行（使用分布式锁，如需要）

### 浏览器兼容性（管理后台）

- [ ] 不涉及前端界面

### 测试要求

- [ ] 命令行工具测试：覆盖所有参数组合和错误场景
- [ ] 集成测试：测试完整的发票生成流程
- [ ] 手动测试：验证发票在ERP系统中的正确性
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 命令行输出使用中文（符合项目规范）
- [ ] 错误信息使用中文

### 安全要求

- [ ] 参数验证：防止SQL注入（使用参数化查询）
- [ ] 权限控制：命令行工具需要适当的执行权限
- [ ] 日志记录：记录操作日志，便于审计
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级（ERP系统不可用时提示错误）
- [ ] 事务管理（更新订单发票信息时保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制（支持重试，如需要）

---

## 验收标准

### 功能验收

1. **命令行工具执行**: 能够成功执行 `regenerate-order-pos-invoice` 命令并生成发票
2. **参数验证**: 缺失或无效参数时能够正确提示错误
3. **订单验证**: 订单不存在或未结账时能够正确提示错误
4. **ERP配置验证**: ERP未启用或配置缺失时能够正确提示错误
5. **发票生成**: 能够成功调用 `SavePosInvoice` 方法并保存发票到ERP系统
6. **订单更新**: 能够成功更新订单的发票名称字段
7. **预览模式**: `--dry-run` 模式能够正确预览操作，不实际执行

### 测试验收

1. **单元测试**: 命令行工具参数解析和验证测试通过
2. **集成测试**: 端到端流程测试通过（读取订单→生成发票→更新订单）
3. **错误场景测试**: 各种错误场景（订单不存在、ERP未启用、发票保存失败等）测试通过
4. **手动测试**: 验证发票在ERP系统中的正确性

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **命令行文档**: 命令使用说明完整
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Cobra 框架创建命令行工具
- 必须使用 Gin Context（命令行环境）
- 必须复用现有的 `SavePosInvoice` 方法
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc` 规范

### 业务约束

- **订单状态**: 仅支持已完成结账的订单
- **ERP配置**: 必须启用ERP Phase3且配置SiteCode
- **班次检查**: `SavePosInvoice` 方法会检查班次是否已交班，可能需要特殊处理
- **发票覆盖**: 如果订单已有发票，需要确认是否覆盖或报错（待明确）

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/spf13/cobra` - 命令行框架
- `main/app/service/order.go` - `SavePosInvoice` 方法
- `main/app/repository/order.go` - 订单信息读取
- `main/app/repository/sale_order.go` - 订单发票信息更新

### 服务依赖

- **Main → BMP**: gRPC 调用（`SavePosInvoice` 内部调用ERP服务）

### 业务依赖

- **ERP系统**: 必须可用且配置正确
- **订单数据**: 订单必须已完成结账
- **班次信息**: 班次必须未交班（或提供 `--force` 参数跳过检查）

---

## 风险和缓解

### 风险 1: ERP系统不可用

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 在保存发票前检查ERP系统连接状态（如可能）
- 提供明确的错误提示信息
- 支持重试机制（可选）

### 风险 2: 班次已交班

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 提供 `--force` 参数跳过班次检查（需要评估业务影响）
- 或提供明确的错误提示，指导用户处理
- 记录详细的错误日志

### 风险 3: 订单已有发票

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 检查订单是否已有发票名称
- 提供 `--force` 参数强制覆盖（需要评估业务影响）
- 或提供明确的提示信息，询问是否覆盖

### 风险 4: 发票保存成功但订单更新失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用事务确保数据一致性（如需要）
- 记录详细的错误日志
- 提供手动修复的指导信息

---

## 时间表

- **Phase 1 - 命令行工具框架**: 0.5天
- **Phase 2 - 订单信息读取和验证**: 0.5天
- **Phase 3 - 发票生成和更新**: 0.5天
- **Phase 4 - 错误处理和测试**: 0.5天
- **总计**: 2天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考代码

- `main/app/service/order_pay.go:929-939` - 订单支付后保存发票逻辑
- `main/app/service/order.go:4182` - `SavePosInvoice` 方法实现
- `main/command/regenerate_sale_bill_material_outbound.go` - 命令行工具参考实现

### 相关文档

- ERP集成文档: `docs/human/architecture/features/recharge_order.md`
- 提案文档: `docs/team/proposals/2025-12/regenerate-order-pos-invoice.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: xiezhihuan  
**审核者**: {审核者}

