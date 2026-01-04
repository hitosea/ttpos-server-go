> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 重新生成每日销售出库汇总记录 需求文档

> 本文档定义重新生成每日销售出库汇总记录功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/regenerate-daily-sales-outbound-summary.md](../../../../team/proposals/2025-12/regenerate-daily-sales-outbound-summary.md) |
| **创建日期**      | 2025-12-15                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

提供一个管理功能，支持重新生成指定日期的销售出库汇总记录。当发现某一天的销售出库汇总记录有误时（如数据统计错误、营业时段配置错误等），管理员可以通过管理后台、命令行工具或 API 接口快速删除旧记录并重新生成正确的汇总记录。

**核心价值**：
- **提升数据准确性**：快速修正历史销售出库汇总记录，确保报表数据准确
- **降低运维成本**：减少需要开发人员介入的数据修正场景
- **提高操作效率**：管理员可以自主处理数据修正，无需等待定时任务或开发支持
- **增强系统可靠性**：提供数据修复能力，提升系统容错性

**功能范围**：
- ✅ 按门店和日期删除旧的销售出库汇总记录（软删除）
- ✅ 重新计算并生成指定日期的销售出库汇总记录
- ✅ 支持管理后台界面操作（Shop 商家管理端）
- ✅ 支持命令行工具批量操作
- ✅ 支持 API 接口调用
- ✅ 提供安全控制和操作日志

## 🎯 产品对齐

本功能支持以下产品目标：
- **数据准确性**：确保销售出库汇总记录与订单数据一致，为报表和 ERP 系统提供准确的数据源
- **运维效率**：减少数据修正的复杂度和时间成本，提升系统可维护性
- **用户体验**：管理员可以快速自主处理数据问题，无需等待开发支持

## 📝 用户故事

**作为** 商户管理员  
**我想** 重新生成指定日期的销售出库汇总记录  
**以便于** 修正历史数据错误，确保报表数据准确

**作为** 系统管理员  
**我想** 通过命令行工具批量重新生成多个门店的销售出库汇总记录  
**以便于** 快速修复批量数据问题

---

## 功能需求

### Requirement 1: 删除指定日期的旧销售出库汇总记录

**用户故事**: 作为商户管理员，我想删除指定日期的旧销售出库汇总记录，以便于重新生成正确的记录

#### 验收标准

1. **WHEN** 管理员指定门店 UUID 和日期 **THEN** 系统 **SHALL** 查询该日期所有营业时段的销售出库汇总记录（`log_type=1, scene=1`）
2. **IF** 找到旧记录 **THEN** 系统 **SHALL** 软删除这些记录（更新 `delete_time` 字段）
3. **WHEN** 删除操作完成 **THEN** 系统 **SHALL** 返回删除的记录数量
4. **IF** 删除过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性

#### 具体要求

- [ ] 1.1 根据 `opening_hours` 字段匹配指定日期的所有营业时段记录（格式：`YYYYMMDD HH:mm-HH:mm`）
- [ ] 1.2 仅删除 `log_type=1`（出库）且 `scene=1`（销售出库）的记录
- [ ] 1.3 使用软删除方式（更新 `delete_time` 字段），不物理删除数据
- [ ] 1.4 记录操作日志，包括操作人、时间、删除的记录数
- [ ] 1.5 支持事务回滚，确保删除操作的原子性

---

### Requirement 2: 重新生成指定日期的销售出库汇总记录

**用户故事**: 作为商户管理员，我想重新生成指定日期的销售出库汇总记录，以便于获得正确的统计数据

#### 验收标准

1. **WHEN** 管理员指定门店 UUID 和日期 **THEN** 系统 **SHALL** 获取该日期营业时段内的所有销售订单材料数据
2. **IF** 存在销售出库数据 **THEN** 系统 **SHALL** 按仓库和物料分组汇总数量
3. **WHEN** 汇总计算完成 **THEN** 系统 **SHALL** 生成新的销售出库汇总记录并保存到 `ttpos_warehouse_in_out_log` 表
4. **IF** 生成过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性

#### 具体要求

- [ ] 2.1 复用 `DailySalesOutboundSummaryTask.getDailySalesOutboundRecords` 方法获取销售出库数据
- [ ] 2.2 根据门店的营业时段配置，计算该日期的开始和结束时间戳
- [ ] 2.3 查询 `ttpos_sale_order_material` 表中该时间范围内的所有记录
- [ ] 2.4 按 `warehouse_uuid` 和 `material_uuid` 分组汇总 `num` 字段
- [ ] 2.5 复用 `DailySalesOutboundSummaryTask.saveOutboundSummaryRecords` 方法保存新记录
- [ ] 2.6 生成出库单号（格式：`SSCK + YYYYMMDD + 4位序号`）
- [ ] 2.7 设置 `opening_hours` 字段为 `YYYYMMDD HH:mm-HH:mm` 格式
- [ ] 2.8 更新 `ttpos_sale_order_material.is_summarized` 状态
- [ ] 2.9 使用事务保证数据一致性

---

### Requirement 3: 管理后台界面操作

**用户故事**: 作为商户管理员，我想在管理后台界面重新生成销售出库汇总记录，以便于便捷地修正数据

#### 验收标准

1. **WHEN** 管理员进入出入库记录列表页面 **THEN** 系统 **SHALL** 显示"重新生成"按钮
2. **IF** 管理员点击"重新生成"按钮 **THEN** 系统 **SHALL** 弹出日期选择对话框
3. **WHEN** 管理员选择日期并确认 **THEN** 系统 **SHALL** 显示确认对话框，显示将删除的记录数和预计生成的记录数
4. **IF** 管理员确认操作 **THEN** 系统 **SHALL** 调用后端 API 执行重新生成操作
5. **WHEN** 操作完成后 **THEN** 系统 **SHALL** 显示操作结果，包括删除记录数、生成记录数、操作耗时

#### 具体要求

- [ ] 3.1 在 Shop 商家管理端的出入库记录列表页面添加"重新生成"按钮
- [ ] 3.2 按钮仅对管理员角色可见（权限控制）
- [ ] 3.3 点击按钮后弹出日期选择器，默认选择当前日期
- [ ] 3.4 确认对话框显示操作预览信息（删除记录数、预计生成记录数）
- [ ] 3.5 操作过程中显示加载状态，防止重复点击
- [ ] 3.6 操作完成后显示成功提示，并刷新列表数据
- [ ] 3.7 操作失败时显示错误提示信息

---

### Requirement 4: 命令行工具操作

**用户故事**: 作为系统管理员，我想通过命令行工具批量重新生成销售出库汇总记录，以便于快速修复批量数据问题

#### 验收标准

1. **WHEN** 管理员执行命令行工具 **THEN** 系统 **SHALL** 支持 `--company-uuid` 和 `--date` 参数
2. **IF** 提供 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作，不实际执行
3. **WHEN** 执行重新生成操作 **THEN** 系统 **SHALL** 输出操作进度和结果
4. **IF** 操作失败 **THEN** 系统 **SHALL** 输出错误信息并返回非零退出码

#### 具体要求

- [ ] 4.1 新增 `regenerate-sales-outbound` 子命令到 `main/command/`
- [ ] 4.2 支持 `--company-uuid` 参数（必填，门店 UUID）
- [ ] 4.3 支持 `--date` 参数（必填，格式：YYYY-MM-DD）
- [ ] 4.4 支持 `--dry-run` 参数（可选，预览模式）
- [ ] 4.5 输出操作日志，包括删除记录数、生成记录数、操作耗时
- [ ] 4.6 支持批量操作（通过脚本循环调用）

---

### Requirement 5: API 接口调用

**用户故事**: 作为系统集成方，我想通过 API 接口调用重新生成功能，以便于集成到其他系统或自动化脚本中

#### 验收标准

1. **WHEN** 调用 API 接口 **THEN** 系统 **SHALL** 验证请求参数（门店 UUID、日期）
2. **IF** 参数验证通过 **THEN** 系统 **SHALL** 执行重新生成操作
3. **WHEN** 操作完成 **THEN** 系统 **SHALL** 返回操作结果，包括删除记录数、生成记录数
4. **IF** 操作失败 **THEN** 系统 **SHALL** 返回错误信息和错误码

#### 具体要求

- [ ] 5.1 新增 API 接口：`POST /api/shop/inventory/regenerate-sales-outbound-summary`
- [ ] 5.2 请求参数：`company_uuid`（必填，uint64）、`date`（必填，string，格式：YYYY-MM-DD）
- [ ] 5.3 响应格式：`{code, message, data: {deleted_count, generated_count, duration_ms}}`
- [ ] 5.4 需要身份验证和权限校验（仅管理员可操作）
- [ ] 5.5 返回标准的错误码和错误信息

---

### Requirement 6: 安全控制和并发控制

**用户故事**: 作为系统管理员，我想确保重新生成操作的安全性和数据一致性，以便于避免数据冲突和误操作

#### 验收标准

1. **IF** 非管理员用户尝试操作 **THEN** 系统 **SHALL** 拒绝请求并返回权限错误
2. **IF** 同一门店同一日期正在执行重新生成操作 **THEN** 系统 **SHALL** 拒绝新的操作请求
3. **WHEN** 操作执行中 **THEN** 系统 **SHALL** 使用分布式锁防止并发操作
4. **IF** 操作超时或失败 **THEN** 系统 **SHALL** 自动释放分布式锁

#### 具体要求

- [ ] 6.1 权限校验：仅管理员角色可操作（Shop 商家管理端）
- [ ] 6.2 使用 Redis 分布式锁，锁的 key 为：`regenerate_sales_outbound_summary:{company_uuid}:{date}`
- [ ] 6.3 锁的超时时间设置为 5 分钟
- [ ] 6.4 操作完成后自动释放锁
- [ ] 6.5 操作失败时也要释放锁（使用 defer 确保）
- [ ] 6.6 记录详细的操作日志，包括操作人、时间、参数、结果

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/shop/inventory/regenerate-sales-outbound-summary`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 复用现有表结构 `ttpos_warehouse_in_out_log`
- [x] 使用软删除（`delete_time` 字段）
- [x] 事务保证数据一致性
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 单次操作响应时间 < 5 秒（对于正常数据量）
- [ ] 数据库查询优化（使用索引，特别是 `opening_hours`、`log_type`、`scene` 字段）
- [ ] 对于大量数据，考虑异步处理或分批处理
- [ ] 并发控制（使用分布式锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程（删除旧记录 + 重新生成）
- [ ] 并发测试（分布式锁有效性）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有前端文案使用多语言实现
- [ ] API 错误信息支持多语言
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 权限校验（仅管理员可操作）
- [x] SQL 注入防护（使用参数化查询）
- [x] 操作日志记录（审计追踪）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 分布式锁防止并发操作
- [x] 操作失败时自动回滚

---

## 验收标准

### 功能验收

1. **删除旧记录功能**: 能够正确删除指定日期的所有营业时段的销售出库汇总记录
2. **重新生成功能**: 能够正确重新计算并生成指定日期的销售出库汇总记录
3. **管理后台界面**: 管理员可以在界面中便捷地执行重新生成操作
4. **命令行工具**: 系统管理员可以通过命令行工具批量操作
5. **API 接口**: 其他系统可以通过 API 接口调用功能
6. **安全控制**: 权限校验和并发控制正常工作

### 测试验收

1. **单元测试**: Service 和 Repository 层测试覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（删除 + 重新生成）
4. **并发测试**: 分布式锁有效性测试通过
5. **手动测试**: 管理后台界面操作测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整
3. **操作文档**: 命令行工具使用说明完整
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 复用现有 `DailySalesOutboundSummaryTask` 中的逻辑

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 仅支持 ERP 商品的门店（`company.IsOpenErp() == true`）
- 仅处理销售出库场景（`log_type=1, scene=1`）
- 操作需要管理员权限
- 同一门店同一日期不能并发操作

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/tasks/daily_sales_outbound_summary.go` - 复用定时任务的统计逻辑
- `main/app/repository/warehouse_in_out_log.go` - 仓库出入库记录 Repository
- `main/app/repository/sale_order_material.go` - 销售订单材料 Repository
- `main/pkg/lock` - 分布式锁工具

### 服务依赖

- **Frontend → Main**: HTTP API 调用
- **Command → Main**: 内部 Service 调用

### 业务依赖

- 依赖 `DailySalesOutboundSummaryTask` 中的统计逻辑
- 依赖门店营业时段配置（`ttpos_company_setting`）
- 依赖销售订单材料数据（`ttpos_sale_order_material`）

---

## 风险和缓解

### 风险 1: 数据一致性风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用事务保证删除和重新生成的原子性
- 使用分布式锁防止并发操作
- 操作失败时自动回滚

### 风险 2: 性能风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 对于大量数据，考虑异步处理或分批处理
- 优化数据库查询（使用索引）
- 设置操作超时时间（5 分钟）

### 风险 3: 误操作风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 前端提供二次确认对话框
- 命令行提供 `--dry-run` 预览模式
- 详细的操作日志记录
- 使用软删除，可通过数据库 binlog 恢复

---

## 时间表

- **Phase 1 - 后端 Service 开发**: 1-2 天
  - 提取公共方法
  - 实现删除旧记录功能
  - 实现重新生成功能
- **Phase 2 - API 和命令行工具**: 1 天
  - 实现 API 接口
  - 实现命令行工具
- **Phase 3 - 前端界面开发**: 1-2 天
  - 实现管理后台界面
  - 集成 API 调用
- **Phase 4 - 测试和联调**: 1 天
  - 单元测试
  - 集成测试
  - 手动测试
- **总计**: 3-5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
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

### 相关代码

- `main/app/tasks/daily_sales_outbound_summary.go` - 定时任务代码（复用逻辑）
- `main/app/service/cost_card_correction_service.go` - 成本卡修正服务（参考删除旧记录逻辑）
- `main/app/model/warehouse_in_out_log.go` - 数据模型
- `main/app/repository/warehouse_in_out_log.go` - Repository 实现

### 外部参考

- [提案文档](../../../../team/proposals/2025-12/regenerate-daily-sales-outbound-summary.md)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-15  
**作者**: xiezhihuan  
**审核者**: {审核者}

