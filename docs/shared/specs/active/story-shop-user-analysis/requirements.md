# 用户分析统计 需求文档

> 本文档定义用户分析统计功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11/user-analysis.md](../../../../team/proposals/2025-11/user-analysis.md) |
| **创建日期**      | 2025-11-26                                                                                                   |
| **负责人**        | 待指派                                                                                                       |
| **目标 Sprint**   | Sprint 待定                                                                                                  |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | 产品组                   |
| **审核日期** | 2025-11-26               |
| **审核意见** | 需求明确，可进入设计阶段 |

---

## 📋 概述

在商家管理端报表中心新增"用户分析"统计页面，提供多维度订单统计分析功能，帮助商户了解不同国籍、点餐方式来源、桌台方式来源、用餐方式的订单分布情况，为精准营销、优化服务流程、调整运营策略提供数据支持。

## 🎯 产品对齐

该功能支撑商户进行用户画像分析，通过多维度订单统计，帮助商户：
- 了解不同国籍顾客的消费习惯，制定针对性营销策略
- 分析点餐方式来源分布，优化各渠道资源配置
- 了解桌台方式和用餐方式偏好，优化服务流程
- 为运营决策提供数据支持

## 📝 用户故事

**作为** 商户管理员/运营人员  
**我想** 查看按国籍、点餐方式来源、桌台方式来源、用餐方式统计的订单数和占比  
**以便于** 了解用户分布情况，制定精准营销策略和优化服务流程

---

## 功能需求

### Requirement 1: 用户分析统计页面基础功能

**用户故事**: 作为商户管理员，我想在报表中心查看用户分析统计页面，以便了解用户订单分布情况。

#### 验收标准

1. **WHEN** 用户进入用户分析页面 **THEN** 系统 **SHALL** 默认显示今天的统计数据（不显示时间选择器，不可选择未来日期）。
2. **WHEN** 页面加载 **THEN** 系统 **SHALL** 显示四个统计维度：国籍、点餐方式来源、桌台方式来源、用餐方式。
3. **WHEN** 统计查询 **THEN** 系统 **SHALL** 排除已被数据管理的订单（`ttpos_data_manage` 表中 `type = DataManageTypeOrder` 的记录）。

#### 具体要求

- [ ] 1.1 新增用户分析统计页面路由和入口（商家管理端报表中心）。
- [ ] 1.2 页面默认查询今天的数据（使用系统时区，00:00:00-23:59:59）。
- [ ] 1.3 页面不显示时间选择器，固定查询今天数据。
- [ ] 1.4 统计查询时排除数据管理订单：使用 `WhereNotInDataManageSubQuery("sale_bill_uuid")` 过滤条件。
- [ ] 1.5 仅统计已完成订单（`sale_bill.status = SaleBillStatusComplete`）。

---

### Requirement 2: 按国籍统计订单数和占比

**用户故事**: 作为商户管理员，我想查看不同国籍的订单数和占比，以便了解顾客国籍分布。

#### 验收标准

1. **WHEN** 系统统计国籍维度 **THEN** 系统 **SHALL** 按订单的国籍字段（`statistics_sale.nationality_uuid`）分组统计。
2. **IF** 查询范围内所有订单的 `nationality_uuid = 0` **THEN** 系统 **SHALL** 不统计此维度（返回空数组）。
3. **WHEN** 存在 `nationality_uuid > 0` 的订单 **THEN** 系统 **SHALL** 仅统计这些订单，关联 `ttpos_nationality` 表获取国籍名称。
4. **WHEN** 系统计算占比 **THEN** 系统 **SHALL** 保留2位小数（如：25.67%）。
5. **WHEN** 系统展示统计结果 **THEN** 系统 **SHALL** 按订单数升序排序。
6. **WHEN** 统计查询 **THEN** 系统 **SHALL** 排除已被数据管理的订单。

#### 具体要求

- [ ] 2.1 统计前先检查：如果查询范围内所有订单的 `nationality_uuid = 0`，则返回空数组，不进行统计。
- [ ] 2.2 仅统计 `nationality_uuid > 0` 的订单，统计字段：国籍名称（从 `nationality` 表关联获取）、订单数、占比（%）。
- [ ] 2.3 占比计算公式：`(该国籍订单数 / 总订单数) * 100`，使用 `decimal` 类型进行精确计算，保留2位小数。
- [ ] 2.4 排序规则：按订单数升序（`ORDER BY order_count ASC`）。
- [ ] 2.5 如果国籍已被删除，仍显示原国籍名称（历史数据保护）。
- [ ] 2.6 响应格式：`{nationality_name, order_count, percentage}`。

---

### Requirement 3: 按点餐方式来源统计订单数和占比

**用户故事**: 作为商户管理员，我想查看不同点餐方式来源的订单数和占比，以便了解店内和外卖订单分布。

#### 验收标准

1. **WHEN** 系统统计点餐方式来源维度 **THEN** 系统 **SHALL** 按订单来源字段（`sale_bill.order_source_uuid`）分组统计。
2. **WHEN** `order_source_uuid = 0` **THEN** 系统 **SHALL** 归类为"店内"。
3. **WHEN** `order_source_uuid > 0` **THEN** 系统 **SHALL** 从 `ttpos_order_source` 表关联获取来源名称（通过 `multi_language_name_uuid` 关联多语言名称表）。
4. **IF** `order_source` 已被删除或不存在 **THEN** 系统 **SHALL** 显示原名称或"未知来源"。
5. **WHEN** 系统计算占比 **THEN** 系统 **SHALL** 保留2位小数（如：45.23%）。
6. **WHEN** 系统展示统计结果 **THEN** 系统 **SHALL** 按订单数升序排序。
7. **WHEN** 统计查询 **THEN** 系统 **SHALL** 排除已被数据管理的订单。

#### 具体要求

- [ ] 3.1 统计字段：来源名称（"店内"或从 `order_source` 表获取的具体来源名称）、订单数、占比（%）。
- [ ] 3.2 占比计算公式：`(该来源订单数 / 总订单数) * 100`，使用 `decimal` 类型进行精确计算，保留2位小数。
- [ ] 3.3 排序规则：按订单数升序（`ORDER BY order_count ASC`）。
- [ ] 3.4 仅统计点餐订单（`sale_bill.bill_type = SaleBillTypeInstant`）。
- [ ] 3.5 响应格式：`{source_name, order_count, percentage}`。

---

### Requirement 4: 按桌台方式来源统计订单数和占比

**用户故事**: 作为商户管理员，我想查看不同桌台方式来源的订单数和占比，以便了解哪个端开的桌台订单最多。

#### 验收标准

1. **WHEN** 系统统计桌台方式来源维度 **THEN** 系统 **SHALL** 按订单的来源字段（`sale_bill.source`）分组统计。
2. **WHEN** `source = "cashier"` **THEN** 系统 **SHALL** 归类为"收银机"。
3. **WHEN** `source = "assistant"` **THEN** 系统 **SHALL** 归类为"点餐助手"。
4. **WHEN** `source = "tablet"` **THEN** 系统 **SHALL** 归类为"平板"。
5. **WHEN** `source = "h5"` **THEN** 系统 **SHALL** 归类为"H5"。
6. **WHEN** `source` 为空或其他值 **THEN** 系统 **SHALL** 归类为"未记录"。
7. **WHEN** 系统计算占比 **THEN** 系统 **SHALL** 保留2位小数（如：30.45%）。
8. **WHEN** 系统展示统计结果 **THEN** 系统 **SHALL** 按订单数升序排序。
9. **WHEN** 统计查询 **THEN** 系统 **SHALL** 排除已被数据管理的订单。

#### 具体要求

- [ ] 4.1 统计字段：来源名称（"收银机"、"点餐助手"、"平板"、"H5"、"未记录"）、订单数、占比（%）。
- [ ] 4.2 占比计算公式：`(该来源订单数 / 总订单数) * 100`，使用 `decimal` 类型进行精确计算，保留2位小数。
- [ ] 4.3 排序规则：按订单数升序（`ORDER BY order_count ASC`）。
- [ ] 4.4 仅统计桌台订单（`sale_bill.bill_type = SaleBillTypeDesk`）。
- [ ] 4.5 响应格式：`{source_name, order_count, percentage}`。

---

### Requirement 5: 按用餐方式统计订单数和占比

**用户故事**: 作为商户管理员，我想查看不同用餐方式的订单数和占比，以便了解打包和店内用餐的分布。

#### 验收标准

1. **WHEN** 系统统计用餐方式维度 **THEN** 系统 **SHALL** 从 `sale_bill.dining_method` 字段获取并分组统计。
2. **WHEN** `dining_method = 0` **THEN** 系统 **SHALL** 归类为"店内用餐"。
3. **WHEN** `dining_method = 1` **THEN** 系统 **SHALL** 归类为"打包"。
4. **WHEN** 统计桌台订单（`bill_type = SaleBillTypeDesk`） **THEN** 系统 **SHALL** 统一归类为"店内用餐"（桌台订单的 `dining_method` 按 0 处理）。
5. **WHEN** 统计点餐订单（`bill_type = SaleBillTypeInstant`） **THEN** 系统 **SHALL** 按 `dining_method` 字段统计。
6. **WHEN** 系统计算占比 **THEN** 系统 **SHALL** 保留2位小数（如：60.12%）。
7. **WHEN** 系统展示统计结果 **THEN** 系统 **SHALL** 按订单数升序排序。
8. **WHEN** 统计查询 **THEN** 系统 **SHALL** 排除已被数据管理的订单。

#### 具体要求

- [ ] 5.1 统计字段：用餐方式名称（"打包"、"店内用餐"）、订单数、占比（%）。
- [ ] 5.2 占比计算公式：`(该用餐方式订单数 / 总订单数) * 100`，使用 `decimal` 类型进行精确计算，保留2位小数。
- [ ] 5.3 排序规则：按订单数升序（`ORDER BY order_count ASC`）。
- [ ] 5.4 统计范围：包括点餐订单和桌台订单。
- [ ] 5.5 桌台订单统一归类为"店内用餐"（业务规则：桌台订单即使有单品打包，整体仍算店内用餐）。
- [ ] 5.6 响应格式：`{dining_method_name, order_count, percentage}`。

---

### Requirement 6: 用户分析统计 API

**用户故事**: 作为前端开发者，我想调用用户分析统计 API，以便在页面展示统计数据。

#### 验收标准

1. **WHEN** 调用用户分析统计 API **THEN** 系统 **SHALL** 返回四个维度的统计数据。
2. **WHEN** 请求参数为空 **THEN** 系统 **SHALL** 默认查询今天的数据。
3. **WHEN** API 调用成功 **THEN** 响应结构 **SHALL** 符合 `{code, message, data{nationality, order_source, desk_source, dining_method}}` 格式。

#### 具体要求

- [ ] 6.1 新增 `GET /api/v1/shop/statistics/user_analysis` API（最终路径以 `main/app/api/v1/shop` 命名规范为准）。
- [ ] 6.2 请求参数：无（固定查询今天数据）。
- [ ] 6.3 Controller 调用 Service；Service 调用 Repository 新增方法（例如 `CountUserAnalysis`）。
- [ ] 6.4 返回字段：
  - `nationality`: `[{nationality_name, order_count, percentage}]`
  - `order_source`: `[{source_name, order_count, percentage}]`
  - `desk_source`: `[{source_name, order_count, percentage}]`
  - `dining_method`: `[{dining_method_name, order_count, percentage}]`
- [ ] 6.5 对请求做权限校验（必须为当前店铺管理员会话）。
- [ ] 6.6 所有百分比字段使用 `decimal` 类型进行精确计算，保留2位小数，最后转换为 `float64` 返回。

---

### Requirement 7: 用户分析统计导出 API

**用户故事**: 作为商户管理员，我想导出用户分析统计报表，以便向老板或财务分享。

#### 验收标准

1. **WHEN** 调用导出接口 **THEN** 系统 **SHALL** 生成与查询接口相同口径的数据并以表格形式输出。
2. **WHEN** 导出成功 **THEN** 返回导出任务记录，文件名格式 `user_analysis_{shop_id}_{YYYYMMDDHHMM}.xlsx`。
3. **WHEN** 没有数据需要导出 **THEN** 接口 **SHALL** 返回错误提示"没有数据需要导出"。
4. **WHEN** 已有正在导出的任务 **THEN** 接口 **SHALL** 返回错误提示"正在导出,请稍后再操作"。

#### 具体要求

- [ ] 7.1 新增 `GET /api/v1/shop/statistics/user_analysis/export` API。
- [ ] 7.2 请求参数：无（固定导出今天数据，与查询接口一致）。
- [ ] 7.3 导出文件包含四个统计维度：
  - 国籍统计：列名（国籍、订单数、占比%）
  - 点餐方式来源统计：列名（来源、订单数、占比%）
  - 桌台方式来源统计：列名（来源、订单数、占比%）
  - 用餐方式统计：列名（用餐方式、订单数、占比%）
- [ ] 7.4 Service 端调用同一统计方法（`CountUserAnalysis`），避免重复计算。
- [ ] 7.5 导出使用现有通用 Excel 工具，支持多语言文件名。
- [ ] 7.6 导出结果应写入操作日志（`export_record` 表），包含操作者、门店、导出类型。
- [ ] 7.7 导出任务异步处理，使用 `ExportRecord` 模型记录导出状态。
- [ ] 7.8 文件生成后上传到文件服务，返回 `file_uuid` 供前端下载。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/shop/statistics/user_analysis`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 复用现有表结构，不新增表
- [ ] 查询需使用索引优化（`sale_bill` 表的 `finish_time`、`status`、`bill_type` 等字段）
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 查询响应时间 < 500ms（默认今天数据）
- [ ] 数据库查询优化（使用索引、避免全表扫描）
- [ ] 考虑使用聚合查询，减少数据库往返次数
- [ ] 导出功能异步处理，避免阻塞用户请求
- [ ] 导出文件大小需控制（≤5MB），必要时分页生成

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现（如："店内"、"外卖"、"收银机"等）
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 错误日志记录（使用 Logger）
- [ ] 数据统计异常时返回空数组，不抛错

---

## 验收标准

### 功能验收

1. **页面展示**: 用户分析页面默认显示今天数据，展示四个统计维度。
2. **国籍统计**: 按国籍统计订单数和占比，保留2位小数，按订单数升序排序。
3. **点餐方式来源统计**: 按店内/外卖统计订单数和占比，保留2位小数，按订单数升序排序。
4. **桌台方式来源统计**: 按收银机/点餐助手/平板/H5统计订单数和占比，保留2位小数，按订单数升序排序。
5. **用餐方式统计**: 按打包/店内用餐统计订单数和占比，桌台订单统一为店内用餐，保留2位小数，按订单数升序排序。
6. **数据管理排除**: 所有统计均排除已被数据管理的订单。
7. **导出功能**: 导出接口可生成包含四个统计维度的 Excel 文件，数据与查询接口一致。

### 测试验收

1. **单元测试**: Repository 新增方法 ≥80% 覆盖。
2. **API 测试**: 接口测试用例覆盖默认查询、无数据、数据管理排除等场景。
3. **集成测试**: 验证统计查询与数据管理过滤逻辑。

### 文档验收

1. **技术文档**: design.md 需描述接口定义、数据口径、查询逻辑。
2. **API 文档**: API 接口文档完整。
3. **tasks.md**: 包含对应开发与测试任务。

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 仅针对商家管理端的统计数据，暂不支持自营多门店聚合。
- 固定查询今天数据，不支持历史数据查询。
- 桌台订单统一归类为"店内用餐"（业务规则）。

### 资源约束

- 开发时间：5 天
- Story Point: 8（需在设计阶段确认）

---

## 依赖关系

### 技术依赖

- `main/app/repository/statistics.go`：复用统计查询逻辑。
- `main/app/repository/common.go`：使用 `WhereNotInDataManageSubQuery` 排除数据管理订单。
- `main/app/model/sale_bill.go`：订单模型定义。
- `main/app/model/export_record.go`：导出记录模型。
- `main/app/service/business.go`：参考 `ExportChannelSales` 实现导出逻辑。

### 服务依赖

- **Frontend → Main**: HTTP API 调用

### 业务依赖

- 订单来源和国籍功能（`story-order-source-nationality`）：依赖订单中的 `order_source_uuid` 和 `nationality_uuid` 字段。

---

## 风险和缓解

### 风险 1: 历史订单缺少部分字段

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 对于缺失字段的订单，归类为"未记录"
- 在统计结果中明确标注"未记录"的数量

### 风险 2: 多维度统计查询性能问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用索引优化查询（`finish_time`、`status`、`bill_type` 等字段）
- 使用聚合查询，减少数据库往返次数
- 考虑使用缓存（Redis）缓存今天的数据

### 风险 3: 字段定义不明确导致统计口径不一致

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 在技术设计阶段明确各字段的定义和统计口径
- 编写详细的测试用例验证统计逻辑
- 与产品确认业务规则（如桌台订单统一为店内用餐）

---

## 时间表

- **Phase 1 - API 开发**: 3 天
- **Phase 2 - 前端页面开发**: 2 天
- **总计**: 5 天（SP = 8）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南

### 外部参考

- 类似功能: `docs/shared/specs/active/story-shop-channel-sales/` - 渠道营业统计

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 产品组  
**审核者**: {审核者}

