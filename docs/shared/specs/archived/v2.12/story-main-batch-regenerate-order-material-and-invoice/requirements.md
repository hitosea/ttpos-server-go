> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 批量重新生成订单材料消耗和POS发票 需求文档

> 本文档定义批量重新生成订单材料消耗和POS发票功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/batch-regenerate-order-material-and-invoice.md](../../../../team/proposals/2025-12/batch-regenerate-order-material-and-invoice.md) |
| **创建日期**      | 2025-12-17                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

提供一个命令行工具，用于批量重新生成指定公司和日期范围内的所有订单的材料消耗和POS发票。当需要批量修复某个公司从某个日期开始的所有订单数据时（如成本卡配置变更、材料统计逻辑变更等），可以通过此工具自动执行批量处理，无需手动逐个订单执行多个命令。

**核心价值**：
- **提升运维效率**：批量处理订单，减少手动操作时间
- **降低操作风险**：自动化流程减少人为错误
- **支持断点续传**：程序中断后可从断点继续，避免重复操作
- **进度可视化**：实时查看处理进度和结果统计（可选功能）
- **错误追踪**：记录详细的错误日志，便于问题排查和修复

**功能范围**：
- ✅ 根据公司列表和起始日期查询所有符合条件的订单
- ✅ 按日期分组订单，生成四层结构的JSON任务清单（公司 → 日期 → 订单 → 步骤）
- ✅ 按日期分组执行，每个日期先执行所有订单的步骤，再执行日期级别的汇总步骤
- ✅ 支持断点续传，自动跳过已完成的步骤
- ✅ 支持进度显示（可选功能），实时显示公司/日期/订单/步骤级别的进度
- ✅ 记录详细的执行日志和错误信息

## 🎯 产品对齐

本功能支持以下产品目标：
- **数据准确性**：确保订单材料消耗和POS发票数据与最新的成本卡配置一致，为成本核算和出库汇总提供准确的数据源
- **运维效率**：减少批量数据修正的复杂度和时间成本，提升系统可维护性
- **问题修复**：支持快速批量修复历史订单数据，无需手动逐个订单执行多个命令

## 📝 用户故事

**作为** 运维人员/技术支持人员  
**我想** 通过命令行工具批量重新生成指定公司和日期范围内的所有订单的材料消耗和POS发票  
**以便于** 快速修复批量订单数据，无需手动逐个订单执行多个命令

---

## 功能需求

### Requirement 1: 生成任务清单

**用户故事**: 作为运维人员，我想根据公司列表和起始日期生成JSON格式的任务清单，以便于批量处理订单

#### 验收标准

1. **WHEN** 提供 `--company-uuids` 和 `--start-date` 参数 **THEN** 系统 **SHALL** 查询所有符合条件的订单（`company_uuid IN (?) AND created_at >= ? AND status = ?`）
2. **WHEN** 查询到订单后 **THEN** 系统 **SHALL** 按日期分组订单
3. **WHEN** 按日期分组后 **THEN** 系统 **SHALL** 为每个日期生成日期级别的步骤（regenerate-sales-outbound）
4. **WHEN** 为每个订单生成任务时 **THEN** 系统 **SHALL** 为每个订单生成3个步骤（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice）
5. **WHEN** 任务清单生成完成 **THEN** 系统 **SHALL** 将任务清单保存为JSON文件（默认路径：`./batch-regenerate-task-{timestamp}.json`）
6. **IF** 使用 `--task-file` 参数 **THEN** 系统 **SHALL** 使用指定的文件路径保存任务清单
7. **IF** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅生成任务清单，不实际执行任务

#### 具体要求

- [ ] 1.1 查询订单时使用 `repository.NewSaleOrderRepo(db).GetSaleOrdersByCompanyAndDateRange()` 方法
- [ ] 1.2 按订单的 `created_at` 字段的日期部分进行分组
- [ ] 1.3 任务清单采用四层JSON结构：公司 → 日期 → 订单 → 步骤
- [ ] 1.4 每个步骤包含：step（序号）、name（步骤名称）、status（状态）、start_time、end_time、error
- [ ] 1.5 任务清单包含 summary 统计信息：总公司数、总日期数、总订单数、总步骤数、完成数、失败数、待执行数
- [ ] 1.6 任务清单包含 created_at 和 updated_at 时间戳

---

### Requirement 2: 执行任务清单

**用户故事**: 作为运维人员，我想按照任务清单自动执行批量处理，以便于无需手动操作

#### 验收标准

1. **WHEN** 任务清单生成完成 **THEN** 系统 **SHALL** 自动开始执行任务，按公司→日期→订单→步骤的顺序处理
2. **WHEN** 执行每个日期时 **THEN** 系统 **SHALL** 先执行该日期下所有订单的步骤，等所有订单执行完成后，再执行日期级别的步骤（regenerate-sales-outbound）
3. **WHEN** 执行每个订单步骤时 **THEN** 系统 **SHALL** 调用对应的Service方法执行步骤
4. **WHEN** 该日期下所有订单的所有步骤都完成后（completed或failed） **THEN** 系统 **SHALL** 自动执行日期级别的步骤（regenerate-sales-outbound）
5. **WHEN** 步骤执行前 **THEN** 系统 **SHALL** 检查步骤状态，跳过状态为 `completed` 的步骤
6. **WHEN** 步骤执行完成后 **THEN** 系统 **SHALL** 更新任务清单中对应步骤的状态为 `completed`，并记录开始时间和结束时间
7. **WHEN** 步骤执行失败 **THEN** 系统 **SHALL** 更新任务清单中对应步骤的状态为 `failed`，并记录错误信息，然后继续处理下一个订单或日期
8. **WHEN** 每个步骤执行完成后 **THEN** 系统 **SHALL** 更新任务清单文件的状态，保持文件与执行状态同步

#### 具体要求

- [ ] 2.1 按公司顺序遍历，按日期顺序遍历，按订单顺序遍历，按步骤顺序执行
- [ ] 2.2 日期级别步骤执行时机：检查该日期下所有订单的所有步骤是否都已完成（completed或failed）
- [ ] 2.3 步骤执行映射：
  - 日期级别：`regenerate-sales-outbound` → `salesOutboundSummarySrv.RegenerateSalesOutboundSummary()`
  - 订单步骤1：`regenerate-order-material` → `salesOutboundSummarySrv.RegenerateOrderMaterial()`
  - 订单步骤2：`regenerate-sale-order-material-outbound` → `salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound()`
  - 订单步骤3：`regenerate-order-pos-invoice` → `salesOutboundSummarySrv.RegenerateOrderPosInvoice()`
- [ ] 2.4 使用文件锁机制防止多个实例同时操作同一任务清单文件
- [ ] 2.5 每个步骤执行完成后立即保存任务清单状态，避免程序崩溃导致进度丢失

---

### Requirement 3: 断点续传

**用户故事**: 作为运维人员，我想从现有任务清单继续执行，以便于程序中断后可以从断点继续

#### 验收标准

1. **IF** 使用 `--resume` 参数 **THEN** 系统 **SHALL** 从现有任务清单文件继续执行
2. **WHEN** 读取现有任务清单时 **THEN** 系统 **SHALL** 自动跳过状态为 `completed` 的步骤
3. **WHEN** 遇到状态为 `failed` 的步骤 **THEN** 系统 **SHALL** 重新执行该步骤
4. **WHEN** 程序重新执行时 **THEN** 系统 **SHALL** 自动跳过已完成的步骤，只执行未完成或失败的步骤
5. **IF** 任务清单文件不存在或格式错误 **THEN** 系统 **SHALL** 提示错误信息并退出

#### 具体要求

- [ ] 3.1 使用 `--task-file` 参数指定任务清单文件路径
- [ ] 3.2 读取JSON文件并解析任务清单结构
- [ ] 3.3 验证任务清单文件的完整性和一致性
- [ ] 3.4 跳过状态为 `completed` 的步骤，继续执行 `pending` 或 `failed` 的步骤

---

### Requirement 4: 进度显示（可选功能）

**用户故事**: 作为运维人员，我想查看详细的进度信息，以便于了解处理进度和预计完成时间

#### 验收标准

1. **IF** 使用 `--show-progress` 参数 **THEN** 系统 **SHALL** 显示详细的进度信息
2. **WHEN** 显示进度信息时 **THEN** 系统 **SHALL** 显示公司级别进度：总共有多少个公司、还有多少个公司未开始、哪些公司已经完成、当前正在处理哪个公司
3. **WHEN** 显示进度信息时 **THEN** 系统 **SHALL** 显示日期级别进度：当前公司下总共有多少个日期、还有多少个日期未开始、哪些日期已经完成、当前正在处理哪个日期
4. **WHEN** 显示进度信息时 **THEN** 系统 **SHALL** 显示订单级别进度：当前日期下总共有多少个订单、还有多少个订单未开始、哪些订单已经完成、当前正在处理哪个订单
5. **WHEN** 显示进度信息时 **THEN** 系统 **SHALL** 显示步骤级别进度：当前订单的步骤执行情况、当前步骤的执行状态
6. **WHEN** 显示进度信息时 **THEN** 系统 **SHALL** 显示总体统计：总体完成百分比、已完成的步骤数、失败的步骤数、待执行的步骤数、预计剩余时间
7. **IF** 使用 `--progress-interval` 参数 **THEN** 系统 **SHALL** 按照指定的间隔（秒数）刷新进度信息显示（默认：5秒）

#### 具体要求

- [ ] 4.1 使用定时器定期更新进度信息（默认每5秒刷新一次）
- [ ] 4.2 计算各级别的完成百分比和剩余数量
- [ ] 4.3 使用颜色和格式化输出，清晰展示进度信息
- [ ] 4.4 支持实时更新，不阻塞主执行流程
- [ ] 4.5 预计剩余时间计算：基于已完成步骤的平均耗时和剩余步骤数

---

### Requirement 5: 日志记录和错误处理

**用户故事**: 作为运维人员，我想查看详细的执行日志和错误信息，以便于问题排查和修复

#### 验收标准

1. **WHEN** 执行过程中 **THEN** 系统 **SHALL** 在控制台实时显示当前处理的公司、日期、订单和步骤
2. **WHEN** 执行过程中 **THEN** 系统 **SHALL** 记录详细的执行日志到文件（`logs/batch-regenerate-{timestamp}.log`）
3. **WHEN** 步骤执行失败时 **THEN** 系统 **SHALL** 记录错误信息到任务清单JSON文件的 `error` 字段
4. **WHEN** 步骤执行失败时 **THEN** 系统 **SHALL** 记录详细的错误信息（错误类型、错误消息、堆栈信息）
5. **WHEN** 单个步骤失败时 **THEN** 系统 **SHALL** 不影响其他步骤和订单的执行
6. **WHEN** 所有任务执行完成 **THEN** 系统 **SHALL** 输出统计信息（总公司数、总日期数、总订单数、完成数、失败数、耗时等）

#### 具体要求

- [ ] 5.1 控制台输出使用颜色区分不同状态（成功/失败/进行中）
- [ ] 5.2 文件日志记录到 `logs/batch-regenerate-{timestamp}.log`
- [ ] 5.3 错误日志记录到任务清单JSON文件的 `error` 字段
- [ ] 5.4 使用 `logger.Logger` 记录详细日志
- [ ] 5.5 错误信息包含：错误类型、错误消息、堆栈信息、时间戳

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Command → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 不涉及 API 接口（命令行工具）

### 数据库设计要求

- [ ] 不涉及数据库表结构变更（复用现有表结构）
- [ ] 使用现有表：`ttpos_sale_order`、`ttpos_sale_order_material`、`ttpos_warehouse_out_form_item` 等

### 性能要求

- [ ] 支持大批量订单处理（1000+ 订单）
- [ ] 每个步骤执行使用数据库事务，确保原子性
- [ ] 支持分批处理订单，避免一次性加载过多数据到内存
- [ ] 任务清单文件读写操作要快速，不阻塞主执行流程

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 命令行工具集成测试覆盖核心流程
- [ ] 断点续传功能测试覆盖各种场景
- [ ] 进度显示功能测试（可选）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 命令行输出支持中文和英文（可选）
- [ ] 错误信息使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 文件锁机制防止并发操作
- [ ] 任务清单文件状态验证，防止文件被意外修改
- [ ] 敏感信息（如错误堆栈）记录到日志文件，不输出到控制台
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制（支持断点续传）
- [ ] 定期保存任务清单状态，避免程序崩溃导致进度丢失

---

## 验收标准

### 功能验收

1. **任务清单生成**: 能够根据公司列表和起始日期生成正确的JSON任务清单
2. **任务执行**: 能够按照任务清单正确执行所有步骤
3. **执行顺序**: 每个日期先执行所有订单步骤，再执行日期级别步骤
4. **断点续传**: 能够从现有任务清单继续执行，跳过已完成的步骤
5. **进度显示**: 能够显示详细的进度信息（如果启用）
6. **日志记录**: 能够记录详细的执行日志和错误信息
7. **错误处理**: 单个步骤失败不影响其他步骤和订单的执行

### 测试验收

1. **单元测试**: 覆盖率达标
2. **集成测试**: 端到端流程测试通过
3. **断点续传测试**: 各种场景测试通过
4. **进度显示测试**: 功能测试通过（如果实现）

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **命令行文档**: 使用说明和参数说明完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架（命令行工具使用 Cobra）
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 每个步骤的执行必须使用数据库事务，确保原子性
- 日期级别步骤必须在该日期下所有订单步骤完成后才能执行
- 任务清单文件必须定期保存，避免进度丢失

### 资源约束

- 开发时间: 4-6 天（包含进度显示功能）
- Story Point: 6-9 SP (必须 ≤ 5，需要拆分)
- **说明**：进度显示功能为可选功能，如果不需要此功能，可以减少1天工作量

---

## 依赖关系

### 技术依赖

- `github.com/spf13/cobra` - 命令行工具框架
- `main/app/service` - 销售出库汇总服务（复用现有Service方法）

### 服务依赖

- **Main → Main**: 复用现有的四个命令的Service方法
  - `salesOutboundSummarySrv.RegenerateOrderMaterial()`
  - `salesOutboundSummarySrv.RegenerateSalesOutboundSummary()`
  - `salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound()`
  - `salesOutboundSummarySrv.RegenerateOrderPosInvoice()`

### 业务依赖

- **依赖功能**:
  - `regenerate-order-material` 命令
  - `regenerate-sales-outbound` 命令
  - `regenerate-sale-order-material-outbound` 命令
  - `regenerate-order-pos-invoice` 命令
- **前置条件**: 以上四个命令的功能必须已实现并可用

---

## 风险和缓解

### 风险 1: 数据一致性风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 每个步骤的执行使用数据库事务，确保原子性
- 使用文件锁机制防止多个实例同时操作同一任务清单文件
- 定期保存任务清单状态，避免程序崩溃导致进度丢失

### 风险 2: 性能风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 支持分批处理订单，避免一次性加载过多数据到内存
- 优化查询逻辑，使用索引加速查询
- 进度显示功能使用异步更新，不阻塞主执行流程

### 风险 3: 任务清单文件风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用文件锁机制防止多个实例同时操作
- 执行前验证任务清单文件的完整性和一致性
- 定期保存任务清单状态，避免文件损坏导致进度丢失

### 风险 4: 错误恢复风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 记录详细的错误日志，便于问题排查和手动修复
- 支持断点续传，可以重新执行失败的步骤
- 单个步骤失败不影响其他步骤和订单的执行

---

## 时间表

- **Phase 1 - 任务清单生成**: 1 天
- **Phase 2 - 任务执行引擎**: 1.5 天
- **Phase 3 - 日期级别步骤执行逻辑**: 0.5 天
- **Phase 4 - 订单步骤完成状态检查**: 0.5 天
- **Phase 5 - 断点续传功能**: 0.5 天
- **Phase 6 - 日志记录系统**: 0.5 天
- **Phase 7 - 错误处理和统计信息**: 0.5 天
- **Phase 8 - 进度显示功能（可选）**: 1 天
- **Phase 9 - 单元测试和集成测试**: 1 天
- **总计**: 4-6 天（SP = 6-9，需要拆分）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 关联命令文档

- `regenerate-order-material` - [提案文档](../../../../team/proposals/2025-12/regenerate-order-material.md) | [Spec文档](../story-main-regenerate-order-material/requirements.md)
- `regenerate-sales-outbound` - [提案文档](../../../../team/proposals/2025-12/regenerate-daily-sales-outbound-summary.md) | [Spec文档](../story-main-regenerate-sales-outbound-summary/requirements.md)
- `regenerate-sale-order-material-outbound` - [提案文档](../../../../team/proposals/2025-12/regenerate-sale-bill-material-outbound.md) | [Spec文档](../story-main-regenerate-sale-bill-material-outbound/requirements.md)
- `regenerate-order-pos-invoice` - [提案文档](../../../../team/proposals/2025-12/regenerate-order-pos-invoice.md) | [Spec文档](../story-main-regenerate-order-pos-invoice/requirements.md)

### 外部参考

- Cobra 命令行工具框架: https://github.com/spf13/cobra

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-17  
**作者**: xiezhihuan  
**审核者**: {审核者}

