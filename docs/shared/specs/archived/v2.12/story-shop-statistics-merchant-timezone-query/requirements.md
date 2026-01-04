> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 统计报表按商户时区查询 需求文档

> 本文档定义统计报表按商户时区查询的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/merchant-timezone-based-statistics-query.md](../../../../team/proposals/2025-12/merchant-timezone-based-statistics-query.md) |
| **创建日期**      | 2025-12-30                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-30             |
| **审核意见** | 同意进入技术设计阶段         |

---

## 📋 概述

所有统计报表、数据查询、导出功能统一使用商户设置的时区进行计算和查询，而非设备/浏览器时区。确保跨时区商户（如日本、泰国、土耳其等）能够准确查看基于自己时区的营业数据。

**核心价值**：
- ✅ **数据准确性**：确保商户查看的是基于自己时区的真实营业数据
- ✅ **用户体验**：无论管理员在哪个时区，都能正确查看商户数据
- ✅ **业务合规**：符合跨时区业务场景的实际需求

## 🎯 产品对齐

支持跨时区商户的准确数据查询，提升国际化业务场景下的数据准确性，减少因时区问题导致的业务决策错误。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在查看统计报表时，系统使用我设置的商户时区来计算时间范围  
**以便于** 无论我在哪个时区，都能准确查看基于商户时区的营业数据

---

## 功能需求

### Requirement 1: 新管理后台统计接口统一使用商户时区

**用户故事**: 作为商户管理员，我想查看统计报表时使用商户时区，以便于准确查看营业数据

#### 验收标准

1. **WHEN** 商户设置时区为 `Asia/Bangkok` (UTC+7) **AND** 管理员在北京时区 (UTC+8) 12:00 查询"今日"数据 **THEN** 系统 **SHALL** 返回商户时区今日（00:00:00 - 23:59:59）的营业数据

2. **WHEN** 商户设置时区为 `Asia/Shanghai` (UTC+8) **AND** 管理员在泰国时区 (UTC+7) 查询"今日"数据 **THEN** 系统 **SHALL** 返回商户时区今日的营业数据

3. **WHEN** 查询"本周"、"本月"等时间范围 **THEN** 系统 **SHALL** 使用商户时区计算时间范围

#### 具体要求

**需要调整的接口列表（Go Main - shop_statistics.go）**：

- [x] 1.1 `/shop/statistics/business` - 统计营业数据（移动管理端首页-店内概况）
- [x] 1.2 `/shop/statistics/payment_method` - 统计支付方式
- [x] 1.3 `/shop/statistics/product_category` - 统计商品分类
- [x] 1.4 `/shop/statistics/product` - 统计商品
- [x] 1.5 `/shop/statistics/area` - 统计区域（移动管理端首页-区域数据）
- [x] 1.6 `/shop/statistics/product_rank` - 统计商品排行（移动管理端首页-销量、销售额排行）
- [x] 1.7 `/shop/statistics/product_sales` - 统计商品销售
- [x] 1.8 `/shop/statistics/product_sales/export` - 导出商品销售统计
- [x] 1.9 `/shop/statistics/7days` - 统计7天
- [x] 1.10 `/shop/statistics/export` - 统计导出
- [x] 1.11 `/shop/statistics/shift_refund_amount` - 统计班次退款金额（无需调整）
- [x] 1.12 `/shop/statistics/home` - 统计首页
- [x] 1.13 `/shop/statistics/kitchen/efficiency_analysis` - 统计后厨效率分析
- [x] 1.14 `/shop/statistics/kitchen/efficiency_analysis/avg` - 统计后厨效率分析平均时长
- [x] 1.15 `/shop/statistics/kitchen/production_detail` - 后厨菜品出品明细
- [x] 1.16 `/shop/statistics/kitchen/production_detail/export` - 导出后厨菜品出品明细
- [x] 1.17 `/shop/statistics/kitchen/efficiency_analysis/export` - 导出后厨效率分析
- [x] 1.18 `/shop/statistics/business/time_period` - 统计营业时段数据（移动端-报表-营业报表-时段营业统计）
- [x] 1.19 `/shop/statistics/business/summary` - 综合运营统计（移动端-报表-营业报表-综合运营统计）
- [x] 1.20 `/shop/statistics/business/payment_method` - 统计支付方式（移动端-报表-营业报表-支付方式统计）
- [x] 1.21 `/shop/statistics/business/time_period/export` - 导出营业时段数据
- [x] 1.22 `/shop/statistics/business/summary/export` - 导出综合运营统计
- [x] 1.23 `/shop/statistics/business/payment_method/export` - 导出营业收款统计
- [x] 1.24 `/shop/statistics/channel_sales` - 渠道营业统计查询
- [x] 1.25 `/shop/statistics/channel_sales/export` - 导出渠道营业统计
- [x] 1.26 `/shop/statistics/user_analysis` - 用户分析统计查询
- [x] 1.27 `/shop/statistics/user_analysis/export` - 导出用户分析
- [x] 1.28 `/shop/member_order/list` - 获取会员订单列表
- [x] 1.29 `/shop/recharge_order/list` - 获取充值订单列表（已支持）
- [x] 1.30 `/shop/order/list` - 获取订单列表（已支持）

**技术实现要点**：
- 所有接口统一从 `ctx.GetCompanySetting().Timezone` 获取商户时区
- 使用 `req.BusinessDataCountReq.GetParam(timezone, openingHours)` 等方法转换时间范围
- Service 层确保时区信息传递到 Repository 层
- SQL 查询中使用商户时区（参考 `bug-251226-001` 修复方案）

---

### Requirement 2: 收银机统计接口统一使用商户时区

**用户故事**: 作为收银员，我想在收银机查看统计数据时使用商户时区，以便于准确查看营业数据

#### 验收标准

1. **WHEN** 商户设置时区为 `Asia/Tokyo` (UTC+9) **AND** 收银机查询"今日"数据 **THEN** 系统 **SHALL** 返回商户时区今日的营业数据

2. **WHEN** 打印统计数据 **THEN** 打印的数据 **SHALL** 基于商户时区的时间范围

#### 具体要求

**需要调整的接口列表（Go Main - cashier_statistics.go）**：

- [x] 2.1 `/cashier/statistics/printer` - 打印统计数据
- [x] 2.2 `/cashier/statistics/business` - 统计营业数据
- [x] 2.3 `/cashier/statistics/payment_method` - 统计支付方式
- [x] 2.4 `/cashier/statistics/product_category` - 统计商品分类
- [x] 2.5 `/cashier/statistics/product` - 统计商品
- [x] 2.6 `/cashier/order/list` - 获取订单列表
- [x] 2.7 `/cashier/recharge_order/list` - 获取充值订单列表
- [x] 2.8 `/cashier/printer/list` - 获取打印列表

**技术实现要点**：
- 与 Requirement 1 相同的时区获取和转换机制
- 打印功能使用 `req.BusinessDataPrinterReq.GetParam(timezone, openingHours)` 转换时间范围

---

### Requirement 3: 旧管理后台统计接口统一使用商户时区

**用户故事**: 作为商户管理员，我想在旧管理后台查看统计数据时使用商户时区，以便于准确查看营业数据

#### 验收标准

1. **WHEN** 商户设置时区为 `Asia/Bangkok` (UTC+7) **AND** 在旧管理后台查询统计数据 **THEN** 系统 **SHALL** 返回商户时区的时间范围数据

2. **WHEN** 查询"最近7天"等时间范围 **THEN** 系统 **SHALL** 使用商户时区计算时间范围

#### 具体要求

**需要调整的接口列表（PHP Admin）**：

- [x] 3.1 `/shop/statistics/sales/index` - 销售数据统计（`admin/app/shop/controller/statistics/Sales.php`）- 无需调整
- [x] 3.2 `/shop/statistics/sales/order` - 订单数据查询（`admin/app/shop/controller/statistics/Sales.php`）- 无需调整
- [x] 3.3 `/shop/statistics/access/index` - 访问数据统计（`admin/app/shop/controller/statistics/Access.php`）- 无需调整
- [x] 3.4 `/shop/statistics/access/data` - 访问数据查询（`admin/app/shop/controller/statistics/Access.php`）- 无需调整
- [x] 3.5 `/shop/statistics/user/index` - 会员数据统计（`admin/app/shop/controller/statistics/User.php`）- 无需调整
- [x] 3.6 `/shop/statistics/user/scale` - 成交会员占比（`admin/app/shop/controller/statistics/User.php`）- 无需调整
- [x] 3.7 `/shop/statistics/user/new_user` - 新增会员（`admin/app/shop/controller/statistics/User.php`）- 无需调整
- [x] 3.8 `/shop/statistics/user/pay_user` - 成交会员数（`admin/app/shop/controller/statistics/User.php`）- 无需调整
- [x] 3.9 `/shop/statistics/supplier/index` - 供应商数据统计（`admin/app/shop/controller/statistics/Supplier.php`）- 无需调整
- [x] 3.10 `/shop/statistics/supplier/data` - 供应商数据查询（`admin/app/shop/controller/statistics/Supplier.php`）- 无需调整
- [x] 3.11 `/shop/statistics/order/index` - 订单数据统计（`admin/app/shop/controller/statistics/Order.php`）- 无需调整
- [x] 3.12 `/shop/statistics/product/index` - 商品数据统计（`admin/app/shop/controller/statistics/Product.php`）- 无需调整

**技术实现要点**：
- PHP 层从商户设置获取时区（`store_setting.time_zone` 或 `company_setting.timezone`）
- 使用 PHP 的 `DateTime` 和 `DateTimeZone` 进行时区转换
- 修复 `getDays()` 等方法，使用商户时区而非服务器时区
- Service 层统一处理时区转换逻辑

---

### Requirement 4: SQL 查询时区修复

**用户故事**: 作为系统，我想在 SQL 查询中使用商户时区，以便于准确统计和分组数据

#### 验收标准

1. **WHEN** SQL 查询使用 `FROM_UNIXTIME` 函数进行日期格式化 **THEN** 系统 **SHALL** 使用商户时区而非 MySQL 服务器时区

2. **WHEN** 按日期分组统计（如按日、按月） **THEN** 系统 **SHALL** 使用商户时区进行分组

#### 具体要求

- [ ] 4.1 修复所有使用 `FROM_UNIXTIME` 的 SQL 查询，统一使用商户时区
- [ ] 4.2 参考 `bug-251226-001-statistics-from-unixtime-timezone-statistics-error` 修复方案
- [ ] 4.3 使用 `CONVERT_TZ` 函数或应用层时区转换
- [ ] 4.4 确保时区信息从 Service 层传递到 Repository 层

**涉及的主要查询**：
- 综合运营统计 (`CountBusinessSummary`)
- 支付方式统计 (`CountBusinessPaymentMethod`)
- 其他使用 `FROM_UNIXTIME` 的统计查询（约 14 处）

---

### Requirement 5: 数据导出功能使用商户时区

**用户故事**: 作为商户管理员，我想导出统计数据时使用商户时区，以便于准确分析营业数据

#### 验收标准

1. **WHEN** 导出统计报表 **THEN** 导出的数据 **SHALL** 基于商户时区的时间范围

2. **WHEN** 导出文件名包含日期 **THEN** 文件名中的日期 **SHALL** 使用商户时区

#### 具体要求

**需要调整的导出接口**：

- [x] 5.1 `/shop/statistics/product_sales/export` - 导出商品销售统计
- [x] 5.2 `/shop/statistics/kitchen/production_detail/export` - 导出后厨菜品出品明细
- [x] 5.3 `/shop/statistics/kitchen/efficiency_analysis/export` - 导出后厨效率分析
- [x] 5.4 `/shop/statistics/business/time_period/export` - 导出营业时段数据
- [x] 5.5 `/shop/statistics/business/summary/export` - 导出综合运营统计
- [x] 5.6 `/shop/statistics/business/payment_method/export` - 导出营业收款统计
- [x] 5.7 `/shop/statistics/channel_sales/export` - 导出渠道营业统计
- [x] 5.8 `/shop/statistics/user_analysis/export` - 导出用户分析

**技术实现要点**：
- 导出数据的时间范围使用商户时区
- 导出文件名中的日期使用商户时区格式
- Excel/CSV 文件中的时间列显示商户时区时间

---

### Requirement 6: 前端时区显示优化

**用户故事**: 作为商户管理员，我想在前端看到明确的时区信息，以便于理解数据的时间范围

#### 验收标准

1. **WHEN** 前端显示时间信息 **THEN** 系统 **SHALL** 明确标识使用的时区（商户时区）

2. **WHEN** 前端显示"今日"、"本周"等时间范围 **THEN** 系统 **SHALL** 显示商户时区的具体时间范围

#### 具体要求

- [ ] 6.1 前端页面显示商户时区信息（如"商户时区：Asia/Bangkok (UTC+7)"）
- [ ] 6.2 时间选择器默认使用商户时区
- [ ] 6.3 统计图表中的时间轴使用商户时区
- [ ] 6.4 可选：同时显示本地时区时间作为参考

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0（UTC 时间戳）
- [x] 时区转换在应用层完成，数据库存储 UTC 时间戳
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] SQL 查询时区转换性能优化（使用索引，避免全表扫描）
- [ ] 缓存策略（如适用）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程（多时区场景）
- [ ] API 测试覆盖所有接口
- [ ] 时区测试用例：覆盖 UTC+7, UTC+8, UTC+9 等主要时区
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 时区显示文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] SQL 注入防护（使用参数化查询）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 时区数据异常时优雅降级（使用默认时区）
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 时区转换失败时的错误处理

---

## 验收标准

### 功能验收

1. **时区获取**: 所有统计接口正确获取商户时区
2. **时间范围转换**: "今日"、"本周"、"本月"等时间类型正确转换为商户时区的具体时间戳范围
3. **SQL 查询**: 所有 SQL 查询使用商户时区进行日期格式化和分组
4. **数据导出**: 导出功能使用商户时区，文件名和数据内容都基于商户时区
5. **前端显示**: 前端明确显示商户时区信息

### 测试验收

1. **单元测试**: 时区转换工具类测试覆盖率达标
2. **API 测试**: 所有统计接口测试通过（多时区场景）
3. **集成测试**: 端到端流程测试通过（跨时区场景）
4. **手动测试**: 多时区场景测试通过（UTC+7, UTC+8, UTC+9）

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档更新（如有变更）
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 时区获取：`ctx.GetCompanySetting().Timezone`
- 时区转换：`utils.SetTimezone(timezone)`

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 时区转换：使用 PHP `DateTime` 和 `DateTimeZone`

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 时区显示：明确标识商户时区

### 业务约束

- 历史数据兼容性：数据库存储 UTC 时间戳，无需迁移
- 时区数据来源：优先使用 `company_setting.timezone`，其次使用 `store_setting.time_zone`

### 资源约束

- 开发时间: 5-7 天
- Story Point: 8-13 SP（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `main/pkg/utils/time.go` - 时区转换工具类
- `main/app/dto/req/statistics.go` - 统计请求 DTO（已有 `GetParam` 方法）

### 服务依赖

- **Main → Setting Service**: 获取商户时区设置
- **Admin → Main**: HTTP API 调用（如适用）

### 业务依赖

- 商户时区设置功能（已存在）
- 时区工具类（已存在，需要统一使用）

---

## 风险和缓解

### 风险 1: SQL 查询性能下降

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用 MySQL `CONVERT_TZ` 函数（数据库层转换，性能较好）
- 确保相关时间字段有索引
- 参考 `bug-251226-001` 修复方案中的性能优化建议

### 风险 2: 历史数据兼容性

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 确认历史数据存储方式（UTC 时间戳），无需迁移
- 时区转换在查询时完成，不影响存储

### 风险 3: 前端时区显示混淆

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 前端明确显示商户时区信息
- 提供时区说明文档
- 可选：同时显示本地时区时间作为参考

### 风险 4: 测试覆盖不足

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 创建多时区测试用例（UTC+7, UTC+8, UTC+9 等）
- 覆盖跨时区场景（管理员在不同时区查询商户数据）
- 集成测试覆盖核心流程

---

## 时间表

- **Phase 1 - 后端 API 统一时区处理**: 2-3 天
- **Phase 2 - SQL 查询时区修复**: 2-3 天
- **Phase 3 - 前端时区显示适配**: 1-2 天
- **Phase 4 - 测试验证**: 1 天
- **总计**: 5-7 天（SP = 8-13）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 相关文档

- Bug 修复方案: `docs/shared/bugs/active/bug-251226-001-statistics-from-unixtime-timezone-statistics-error/solution.md`
- 时区工具类: `main/pkg/utils/time.go`
- 统计请求 DTO: `main/app/dto/req/statistics.go`
- 商户设置: `main/app/dto/resp/setting/store_setting.go`

### 相关代码

- 新管理后台统计接口: `main/app/api/v1/shop/shop_statistics.go`
- 收银机统计接口: `main/app/api/v1/cashier/cashier_statistics.go`
- 旧管理后台统计接口: `admin/app/shop/controller/statistics/`
- 时区转换工具: `utils.SetTimezone(timezone).TodayStartEndUnix()`
- 商户时区获取: `ctx.GetCompanySetting().Timezone`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-30  
**作者**: 王昱  
**审核者**: {审核者}

