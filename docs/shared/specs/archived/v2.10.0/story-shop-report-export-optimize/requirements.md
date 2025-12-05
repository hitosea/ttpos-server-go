> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 优化新管理端导出报表名称 需求文档

> 本文档定义优化新管理端导出报表名称功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/optimize-new-admin-report-export-names.md](../../../../team/proposals/2025-12/optimize-new-admin-report-export-names.md) |
| **创建日期**      | 2025-12-01                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
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

优化新管理端报表导出功能的文件命名和子表命名规则，提升用户体验和文件管理效率。

**核心改进**：
1. 文件名从时间戳格式改为日期格式（`报表名YYYY-MM-DD.xlsx`）
2. 同一天多次导出同名报表时，自动添加序号避免冲突
3. 统一子表名称为 `Sheet1`（用户分析报表除外）

**业务价值**：
- 提升用户体验：文件名包含日期，用户一眼就能识别导出时间
- 便于文件管理：同一天导出的报表文件名统一，便于按日期分类整理
- 标准化命名：统一的命名规则，提升系统专业性

## 🎯 产品对齐

该功能优化了新管理端的报表导出体验，符合产品"提升商户管理效率"的愿景。通过优化文件命名规则，帮助商户管理员更高效地管理和查找历史报表数据，支持数据分析和决策。

## 📝 用户故事

**作为** 商户管理员  
**我想** 导出报表时文件名包含日期，且同一天导出的同名报表自动编号  
**以便于** 快速识别导出时间，便于文件管理和整理

---

## 功能需求

### Requirement 1: 优化报表文件名格式

**用户故事**: 作为商户管理员，我想导出报表时文件名包含日期，以便于快速识别导出时间

#### 验收标准

1. **WHEN** 用户导出报表 **THEN** 文件名格式为 `报表名YYYY-MM-DD.xlsx` **SHALL** 日期使用商户时区
2. **WHEN** 导出时段营业统计报表 **THEN** 文件名 **SHALL** 为 `时段营业统计2025-10-10.xlsx`（示例）
3. **WHEN** 导出综合运营统计报表 **THEN** 文件名 **SHALL** 为 `综合运营统计2025-10-10.xlsx`（示例）
4. **WHEN** 导出营业收款统计报表 **THEN** 文件名 **SHALL** 为 `营业收款统计2025-10-10.xlsx`（示例）
5. **WHEN** 导出渠道营业统计报表 **THEN** 文件名 **SHALL** 为 `渠道营业统计2025-10-10.xlsx`（示例）
6. **WHEN** 导出商品销售统计报表 **THEN** 文件名 **SHALL** 为 `商品销售统计2025-10-10.xlsx`（示例）
7. **WHEN** 导出用户分析报表 **THEN** 文件名 **SHALL** 为 `用户分析2025-10-10.xlsx`（示例）
8. **WHEN** 导出后厨菜品出品明细报表 **THEN** 文件名 **SHALL** 为 `菜品出品明细2025-10-10.xlsx`（示例）
9. **WHEN** 导出后厨效率分析报表 **THEN** 文件名 **SHALL** 为 `菜品出品详情2025-10-10.xlsx`（示例）

#### 具体要求

- [ ] 1.1 文件名中的日期必须使用商户时区（`ctx.GetCompanySetting().Timezone`），而非服务器时区
- [ ] 1.2 日期格式统一为 `YYYY-MM-DD`（如：`2025-10-10`）
- [ ] 1.3 文件名格式：`{报表名}{日期}.xlsx`，中间无分隔符
- [ ] 1.4 报表名称使用多语言支持，根据用户语言设置显示对应语言
- [ ] 1.5 涉及的报表类型：
  - 时段营业统计（`ExportTypeBusinessData`）
  - 综合运营统计（`ExportTypeBusinessDataSummary`）
  - 营业收款统计（`ExportTypeBusinessDataPaymentMethod`）
  - 渠道营业统计（`ExportTypeChannelSales`）
  - 商品销售统计（`ExportTypeProductSales`）
  - 用户分析（`ExportTypeUserAnalysis`）
  - 后厨菜品出品明细（`ExportTypeKitchenProductionDetail`）
  - 后厨效率分析（`ExportTypeKitchenEfficiencyAnalysis`）

---

### Requirement 2: 同名报表自动编号

**用户故事**: 作为商户管理员，我想同一天多次导出同名报表时自动编号，以便于区分不同时间导出的报表

#### 验收标准

1. **WHEN** 同一天多次导出同名报表 **THEN** 文件名自动添加序号 `报表名YYYY-MM-DD（1）.xlsx`、`报表名YYYY-MM-DD（2）.xlsx` **SHALL** 序号从1开始递增
2. **WHEN** 第一次导出报表 **THEN** 文件名 **SHALL** 为 `报表名YYYY-MM-DD.xlsx`（无序号）
3. **WHEN** 第二次导出同名报表（同一天） **THEN** 文件名 **SHALL** 为 `报表名YYYY-MM-DD（1）.xlsx`
4. **WHEN** 第三次导出同名报表（同一天） **THEN** 文件名 **SHALL** 为 `报表名YYYY-MM-DD（2）.xlsx`
5. **IF** 文件名已存在（同一天、同报表类型、同商户） **THEN** 系统 **SHALL** 自动检测并添加序号，避免覆盖

#### 具体要求

- [ ] 2.1 需要查询同一天、同报表类型的已导出记录，计算序号（数据库连接已包含商户隔离）
- [ ] 2.2 序号格式：中文括号 `（数字）`，从1开始递增
- [ ] 2.3 查询范围：仅查询同一天（商户时区）的导出记录
- [ ] 2.4 并发控制：使用数据库事务确保文件名唯一性，避免并发导出时文件名冲突
- [ ] 2.5 查询条件：`export_type` + `日期（商户时区）`（数据库连接已包含商户隔离）

---

### Requirement 3: 统一子表名称为 Sheet1

**用户故事**: 作为商户管理员，我想导出的 Excel 文件子表名称统一为 Sheet1，以便于标准化处理

#### 验收标准

1. **WHEN** 导出报表（除用户分析外） **THEN** Excel 子表名称 **SHALL** 为 `Sheet1`
2. **WHEN** 导出时段营业统计报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
3. **WHEN** 导出综合运营统计报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
4. **WHEN** 导出营业收款统计报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
5. **WHEN** 导出渠道营业统计报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
6. **WHEN** 导出商品销售统计报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
7. **WHEN** 导出后厨菜品出品明细报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
8. **WHEN** 导出后厨效率分析报表 **THEN** 子表名称 **SHALL** 为 `Sheet1`
9. **WHEN** 导出用户分析报表 **THEN** Excel 子表名称 **SHALL** 保持原有逻辑不变（可能有多个子表）

#### 具体要求

- [ ] 3.1 将子表名称从多语言的"报表"/"Report"等改为标准的 `Sheet1`
- [ ] 3.2 用户分析报表（`ExportTypeUserAnalysis`）保持原有子表名称逻辑不变
- [ ] 3.3 涉及的报表类型：
  - 时段营业统计（`ExportTypeBusinessData`）
  - 综合运营统计（`ExportTypeBusinessDataSummary`）
  - 营业收款统计（`ExportTypeBusinessDataPaymentMethod`）
  - 渠道营业统计（`ExportTypeChannelSales`）
  - 商品销售统计（`ExportTypeProductSales`）
  - 后厨菜品出品明细（`ExportTypeKitchenProductionDetail`）
  - 后厨效率分析（`ExportTypeKitchenEfficiencyAnalysis`）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
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

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

**说明**：本功能不涉及新增 API 接口，仅优化现有导出功能的文件命名逻辑。

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**说明**：本功能不涉及数据库表结构变更，仅优化查询逻辑。

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（Redis）
- [x] 并发处理（使用 UUID 锁）

**具体要求**：
- [ ] 查询同一天导出记录时，需要添加索引优化（`export_type`, `create_time`）
- [ ] 并发导出时使用数据库事务确保文件名唯一性
- [ ] 数据库连接已包含商户隔离，无需额外的 `company_uuid` 字段过滤

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

**说明**：本功能为后端优化，不涉及前端变更。

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

**具体要求**：
- [ ] 测试文件名生成逻辑（包含日期格式、序号计算）
- [ ] 测试时区处理正确性
- [ ] 测试并发导出场景
- [ ] 测试子表名称修改

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

**说明**：报表名称本身已支持多语言，文件名中的报表名称需要根据用户语言设置显示对应语言。

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

**具体要求**：
- [ ] 文件名生成失败时，记录错误日志并返回友好错误提示
- [ ] 并发导出时使用数据库事务确保数据一致性

---

## 验收标准

### 功能验收

1. **文件名格式验收**: 所有报表导出时，文件名格式为 `报表名YYYY-MM-DD.xlsx`，日期使用商户时区
2. **自动编号验收**: 同一天多次导出同名报表时，文件名自动添加序号，从1开始递增
3. **子表名称验收**: 除用户分析外，所有报表的子表名称为 `Sheet1`
4. **用户分析例外验收**: 用户分析报表的子表名称保持原有逻辑不变
5. **时区处理验收**: 文件名中的日期正确使用商户时区，而非服务器时区
6. **并发导出验收**: 并发导出时，文件名不冲突，自动添加序号

### 测试验收

1. **单元测试**: 文件名生成逻辑测试覆盖率 ≥ 70%
2. **API 测试**: 所有导出接口测试通过
3. **集成测试**: 端到端导出流程测试通过
4. **手动测试**: 验证文件名格式、序号计算、子表名称

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整（如需要）
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

**涉及文件**：
- `main/app/service/business.go` - 导出服务实现
- `main/app/model/export_record.go` - 导出记录模型
- `main/app/repository/export_record.go` - 导出记录 Repository（如需要新增查询方法）

### 业务约束

- 历史导出记录不受影响，仅影响新导出的文件
- 文件名生成逻辑需要向后兼容，不影响现有导出功能
- 用户分析报表的子表名称逻辑保持不变

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

**分解**：
- 文件名生成逻辑优化：1 天
- 同名文件检测和编号逻辑：0.5 天
- 子表名称修改：0.5 天
- 测试和验证：1 天

---

## 依赖关系

### 技术依赖

- `github.com/xuri/excelize/v2` - Excel 文件生成库（已存在）
- `main/app/utils` - 时区工具类（已存在）
- `main/app/model` - 导出记录模型（已存在）
- `main/app/repository` - 导出记录 Repository（已存在）

### 服务依赖

- **Main → BMP**: 无
- **Admin → Main**: 无
- **Frontend → Admin**: 无

**说明**：本功能为后端内部优化，不涉及跨服务调用。

### 业务依赖

- 导出记录表（`ttpos_export_record`）已存在
- 导出功能已实现，仅优化文件命名逻辑

---

## 风险和缓解

### 风险 1: 时区处理错误

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用商户时区工具类 `utils.SetTimezone()`，确保日期计算正确
- 添加单元测试验证时区处理逻辑
- 在测试环境验证不同时区的商户导出文件名

### 风险 2: 并发导出冲突

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在创建导出记录时使用数据库事务和唯一索引，确保文件名唯一性
- 查询同一天导出记录时，使用 `FOR UPDATE` 锁定记录
- 添加并发测试验证文件名不冲突

### 风险 3: 历史数据兼容

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 仅影响新导出的文件，历史记录不受影响
- 文件名生成逻辑向后兼容，不影响现有导出功能
- 添加回归测试验证现有导出功能不受影响

---

## 时间表

- **Phase 1 - 文件名生成逻辑优化**: 1 天
- **Phase 2 - 同名文件检测和编号逻辑**: 0.5 天
- **Phase 3 - 子表名称修改**: 0.5 天
- **Phase 4 - 测试和验证**: 1 天
- **总计**: 3 天（SP = 3-5）

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

### 代码位置

**文件名生成相关代码**：
- `main/app/service/business.go:2263` - 时段营业统计导出
- `main/app/service/business.go:2467` - 综合运营统计导出
- `main/app/service/business.go:2676` - 营业收款统计导出
- `main/app/service/business.go:2953` - 渠道营业统计导出
- `main/app/service/business.go:789` - 商品销售统计导出
- `main/app/service/business.go:3554` - 用户分析导出
- `main/app/service/business.go:1522` - 后厨效率分析导出
- `main/app/service/business.go:1652` - 后厨菜品出品明细导出

**子表名称相关代码**：
- `main/app/service/business.go:908-920` - 商品销售统计子表名称
- `main/app/service/business.go:2017-2030` - 后厨菜品出品明细子表名称
- `main/app/service/business.go:2150-2165` - 后厨效率分析子表名称
- 其他报表的子表名称设置位置待确认

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**作者**: 王昱  
**审核者**: {审核者}

