# 订单来源追踪 需求文档

> 本文档定义 订单来源追踪 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11/sale-bill-source-tracking.md](../../../../team/proposals/2025-11/sale-bill-source-tracking.md) |
| **创建日期**      | 2025-11-27                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-11-27             |
| **审核意见** | 需求明确，可以进入设计阶段         |

---

## 📋 概述

在创建订单时，系统需要根据请求来源（通过 JWT token 中的 source 信息）自动设置 `SaleBill` 表的 `source` 字段值，以准确记录订单来自哪个客户端。这将支持后续的数据分析、业务决策和问题排查。

**业务价值**：
- 数据统计准确性：准确记录订单来源，支持按来源维度进行数据分析
- 业务决策支持：了解各客户端的订单占比，优化产品策略
- 问题排查能力：当订单出现问题时，可快速定位来源客户端
- 运营分析：支持按来源分析订单转化率、客单价等关键指标

## 🎯 产品对齐

该功能支持产品愿景中的"数据驱动决策"目标，通过准确记录订单来源，为商户提供更精细化的运营数据分析能力，帮助商户了解各客户端的使用情况，优化产品策略和资源配置。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在订单数据中看到订单来自哪个客户端  
**以便于** 分析各客户端的使用情况和订单转化率，优化产品策略

---

## 功能需求

### Requirement 1: 定义 Source 映射常量

**用户故事**: 作为 开发人员，我想 定义 JWT Source 到 SaleBill.source 字段值的映射关系，以便于 统一管理订单来源标识

#### 验收标准

1. **WHEN** 系统需要将 JWT Source 转换为 SaleBill.source 字段值 **THEN** 系统 **SHALL** 使用预定义的映射常量
2. **IF** JWT Source 为 "cashier" **THEN** 系统 **SHALL** 映射为 source = 1
3. **IF** JWT Source 为 "assistant" **THEN** 系统 **SHALL** 映射为 source = 2
4. **IF** JWT Source 为 "tablet" **THEN** 系统 **SHALL** 映射为 source = 3
5. **IF** JWT Source 为 "h5" **THEN** 系统 **SHALL** 映射为 source = 4
6. **IF** JWT Source 为 "member" **THEN** 系统 **SHALL** 映射为 source = 5
7. **IF** JWT Source 无法识别或为空 **THEN** 系统 **SHALL** 映射为 source = 0（默认值）

#### 具体要求

- [ ] 1.1 在 `main/app/constant/` 包中定义 Source 映射常量
- [ ] 1.2 定义映射函数或映射表，将 JWT Source 字符串映射到 uint 类型的 source 值
- [ ] 1.3 支持以下映射关系：
  - `jwt.SourceCashier` ("cashier") → 1
  - `jwt.SourceAssistant` ("assistant") → 2
  - `jwt.SourceTablet` ("tablet") → 3
  - `jwt.SourceH5` ("h5") → 4
  - `jwt.SourceMember` ("member") → 5
  - 其他或未知 → 0
- [ ] 1.4 添加单元测试验证映射正确性

---

### Requirement 2: 在创建即时订单时设置 source

**用户故事**: 作为 系统，我想 在创建即时订单时自动设置 source 字段，以便于 记录订单来源

#### 验收标准

1. **WHEN** 通过收银机创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 1
2. **WHEN** 通过点餐助手创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 2
3. **WHEN** 通过平板创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 3
4. **WHEN** 通过 H5 创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 4

#### 具体要求

- [ ] 2.1 在 `CreateInstantOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 2.2 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 2.3 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 2.4 添加单元测试验证不同来源的订单 source 字段设置正确

---

### Requirement 3: 在创建桌台订单时设置 source

**用户故事**: 作为 系统，我想 在创建桌台订单时自动设置 source 字段，以便于 记录订单来源

#### 验收标准

1. **WHEN** 通过收银机创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 1
2. **WHEN** 通过点餐助手创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 2
3. **WHEN** 通过平板创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 3
4. **WHEN** 通过 H5 创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 4

#### 具体要求

- [ ] 3.1 在 `CreateDeskOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 3.2 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 3.3 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 3.4 添加单元测试验证不同来源的订单 source 字段设置正确

---

### Requirement 4: 在创建会员端订单时设置 source

**用户故事**: 作为 系统，我想 在创建会员端订单时自动设置 source 字段，以便于 记录订单来源

#### 验收标准

1. **WHEN** 通过会员端创建订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 5
2. **WHEN** 通过其他来源创建会员端订单 **THEN** 系统 **SHALL** 根据实际来源设置 source 字段

#### 具体要求

- [ ] 4.1 在 `createMemberOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 4.2 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 4.3 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 4.4 添加单元测试验证会员端订单 source 字段设置正确

---

### Requirement 5: 确保所有创建 SaleBill 的路径都设置 source

**用户故事**: 作为 开发人员，我想 确保所有创建 SaleBill 的代码路径都正确设置 source 字段，以便于 保证数据一致性

#### 验收标准

1. **WHEN** 系统创建 SaleBill 记录 **THEN** 系统 **SHALL** 在所有创建路径中都设置 source 字段
2. **IF** 存在未设置 source 的创建路径 **THEN** 系统 **SHALL** 在代码审查中被识别并修复

#### 具体要求

- [ ] 5.1 全面搜索代码库中所有 `CreateSaleBill` 调用
- [ ] 5.2 确保所有创建 SaleBill 的地方都根据 `ctx.GetSource()` 设置 source 字段
- [ ] 5.3 对于无法获取来源的场景（如导入订单），使用默认值 0
- [ ] 5.4 添加代码审查检查清单，确保不遗漏任何创建路径

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

**说明**: 本功能不涉及新增 API 接口，仅修改现有订单创建逻辑。

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**说明**: `source` 字段已存在于 `ttpos_sale_bill` 表中，无需新增数据库字段。

### 性能要求

- [ ] 本地响应时间 < 200ms（不影响现有订单创建性能）
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

**说明**: 本功能仅是在创建订单时设置字段值，不增加额外的数据库查询，对性能影响可忽略。

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

**说明**: 本功能不涉及前端变更。

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [x] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

**具体要求**：
- [ ] 为 Source 映射函数添加单元测试
- [ ] 为每个订单创建方法添加 source 字段设置的测试用例
- [ ] 测试不同 JWT Source 值映射到正确的 source 字段值
- [ ] 测试未知来源时使用默认值 0

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

**说明**: 本功能不涉及用户可见文案，无需国际化处理。

### 安全要求

- [x] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

**说明**: 本功能使用现有的身份验证机制，不引入新的安全风险。

### 可靠性要求

- [ ] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

**具体要求**：
- [ ] 如果无法获取 source，使用默认值 0 并记录警告日志
- [ ] 确保 source 字段设置失败不影响订单创建（使用默认值）

---

## 验收标准

### 功能验收

1. **Source 映射正确性**: 所有 JWT Source 值都能正确映射到对应的 source 字段值（包括会员端映射到 5）
2. **即时订单 source 设置**: 通过不同客户端创建即时订单时，source 字段正确设置
3. **桌台订单 source 设置**: 通过不同客户端创建桌台订单时，source 字段正确设置
4. **会员端订单 source 设置**: 通过会员端创建订单时，source 字段设置为 5
5. **默认值处理**: 无法识别来源时，source 字段设置为 0
6. **数据一致性**: 所有创建 SaleBill 的路径都正确设置 source 字段

### 测试验收

1. **单元测试**: 覆盖率达标，特别是 Source 映射函数和订单创建方法
2. **API 测试**: 所有订单创建接口测试通过，验证 source 字段设置正确
3. **集成测试**: 端到端流程测试通过，验证不同来源的订单 source 字段正确
4. **手动测试**: 通过不同客户端创建订单，验证数据库中的 source 字段值

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 无需新增 API 文档
3. **数据库文档**: 迁移脚本和表结构文档完整（source 字段已存在）
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

### 业务约束

- **向后兼容**: 历史订单的 source 字段为 0，无法追溯真实来源，这是可接受的
- **默认值处理**: 对于无法识别来源的请求，必须使用默认值 0，不能导致订单创建失败
- **数据完整性**: 所有新创建的订单必须设置 source 字段，不能为 NULL

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/constant/jwt` - JWT Source 常量定义
- `main/pkg/context` - Context 接口，提供 `GetSource()` 方法
- `main/app/model` - SaleBill 模型定义
- `main/app/service` - 订单创建服务

### 服务依赖

- **Main → BMP**: 无依赖
- **Admin → Main**: 无依赖
- **Frontend → Admin**: 无依赖

### 业务依赖

- **JWT Token**: 依赖 JWT token 中正确设置 source 信息
- **订单创建流程**: 依赖现有的订单创建流程正常工作

---

## 风险和缓解

### 风险 1: 遗漏创建入口

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 全面搜索代码库中所有 `CreateSaleBill` 调用
- 代码审查时重点检查所有创建 SaleBill 的路径
- 添加单元测试覆盖所有创建路径

### 风险 2: JWT Source 值不完整

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 对于无法识别的来源，使用默认值 0
- 记录警告日志，便于后续排查和修复
- 不影响订单创建流程的正常执行

### 风险 3: 历史数据无法追溯

**影响**: 低  
**概率**: 高（已发生）  
**缓解措施**:

- 接受历史数据无法追溯的现实
- 从新版本开始记录，确保未来数据完整
- 在文档中明确说明历史数据的局限性

---

## 时间表

- **Phase 1 - 定义常量映射**: 0.5 天
- **Phase 2 - 修改订单创建逻辑**: 1-1.5 天
- **Phase 3 - 单元测试和验证**: 0.5-1 天
- **总计**: 2-3 天（SP = 3）

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
- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关代码文件

- `main/app/constant/jwt/jwt.go` - JWT Source 常量定义
- `main/app/model/sale_bill.go` - SaleBill 模型定义
- `main/app/service/order.go` - 订单创建服务
- `main/app/service/order_base.go` - 订单基础服务
- `main/pkg/context/context.go` - Context 接口实现
- `admin/database/migrations/20251126201148_add_source_to_sale_bill.php` - 数据库迁移文件

### 外部参考

- [提案文档](../../../../team/proposals/2025-11/sale-bill-source-tracking.md)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: xiezhihuan  
**审核者**: {审核者}

