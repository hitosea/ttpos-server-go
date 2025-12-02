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

在创建订单时，系统需要根据请求来源（通过 JWT token 中的 source 信息）自动设置 `SaleBill` 表的 `source` 字段值，同时根据请求头中的 `Client-Version` 自动设置 `client_version` 字段值，以准确记录订单来自哪个客户端和客户端版本。这将支持后续的数据分析、业务决策和问题排查。

**业务价值**：
- 数据统计准确性：准确记录订单来源和客户端版本，支持按来源和版本维度进行数据分析
- 业务决策支持：了解各客户端的订单占比和版本分布，优化产品策略和版本发布策略
- 问题排查能力：当订单出现问题时，可快速定位来源客户端和版本，提高问题排查效率
- 运营分析：支持按来源和版本分析订单转化率、客单价等关键指标
- 版本兼容性分析：了解各版本的使用情况，评估版本升级进度和兼容性问题

## 🎯 产品对齐

该功能支持产品愿景中的"数据驱动决策"目标，通过准确记录订单来源，为商户提供更精细化的运营数据分析能力，帮助商户了解各客户端的使用情况，优化产品策略和资源配置。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在订单数据中看到订单来自哪个客户端和客户端版本  
**以便于** 分析各客户端的使用情况、订单转化率，以及排查特定版本的问题，优化产品策略

**作为** 开发人员  
**我想** 在订单数据中看到订单来自哪个客户端和客户端版本  
**以便于** 快速定位问题订单的来源和版本，提高问题排查效率

---

## 功能需求

### Requirement 0: 数据库迁移 - 新增 client_version 字段

**用户故事**: 作为 系统，我想 在 SaleBill 表中新增 client_version 字段，以便于 记录客户端版本信息

#### 验收标准

1. **WHEN** 数据库迁移执行完成 **THEN** 系统 **SHALL** 在 `ttpos_sale_bill` 表中新增 `client_version` 字段
2. **IF** `client_version` 字段已存在 **THEN** 系统 **SHALL** 跳过创建，不报错
3. **WHEN** 查询 SaleBill 记录 **THEN** 系统 **SHALL** 能够正确读取 `client_version` 字段值

#### 具体要求

- [ ] 0.1 创建数据库迁移文件，在 `ttpos_sale_bill` 表中新增 `client_version` 字段
- [ ] 0.2 字段类型：`varchar(20)`，默认值：空字符串 `''`
- [ ] 0.3 字段注释：`客户端版本号（如 2.10.0、2.9.0）`
- [ ] 0.4 字段允许 NULL：否（NOT NULL）
- [ ] 0.5 更新 SaleBill 模型，添加 `ClientVersion` 字段定义
- [ ] 0.6 验证迁移脚本可以正常执行和回滚

---

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

- [x] 1.1 在 `main/app/constant/` 包中定义 Source 映射常量
- [x] 1.2 定义映射函数或映射表，将 JWT Source 字符串映射到 uint 类型的 source 值
- [x] 1.3 支持以下映射关系：
  - `jwt.SourceCashier` ("cashier") → 1
  - `jwt.SourceAssistant` ("assistant") → 2
  - `jwt.SourceTablet` ("tablet") → 3
  - `jwt.SourceH5` ("h5") → 4
  - `jwt.SourceMember` ("member") → 5
  - 其他或未知 → 0
- [x] 1.4 添加单元测试验证映射正确性

---

### Requirement 2: 在创建即时订单时设置 source 和 client_version

**用户故事**: 作为 系统，我想 在创建即时订单时自动设置 source 和 client_version 字段，以便于 记录订单来源和客户端版本

#### 验收标准

1. **WHEN** 通过收银机创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 1
2. **WHEN** 通过点餐助手创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 2
3. **WHEN** 通过平板创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 3
4. **WHEN** 通过 H5 创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 4
5. **WHEN** 客户端在请求头中发送 `Client-Version: 2.10.0` 创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为 `2.10.0`
6. **WHEN** 客户端未发送 `Client-Version` 请求头创建即时订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为空字符串
7. **WHEN** 客户端发送的版本号格式为 `v2.10.0` **THEN** 系统 **SHALL** 自动去除前缀 `v`，记录为 `2.10.0`

#### 具体要求

- [x] 2.1 在 `CreateInstantOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 2.2 在 `CreateInstantOrder` 方法中，创建 SaleBill 时根据 `ctx.GetVersion()` 设置 client_version 字段
- [x] 2.3 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 2.4 对 client_version 进行格式标准化处理（去除前缀 `v`，统一为 x.y.z 格式）
- [x] 2.5 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 2.6 确保 client_version 字段在 SaleBill 创建时正确设置
- [ ] 2.7 添加单元测试验证不同来源和版本的订单字段设置正确

---

### Requirement 3: 在创建桌台订单时设置 source 和 client_version

**用户故事**: 作为 系统，我想 在创建桌台订单时自动设置 source 和 client_version 字段，以便于 记录订单来源和客户端版本

#### 验收标准

1. **WHEN** 通过收银机创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 1
2. **WHEN** 通过点餐助手创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 2
3. **WHEN** 通过平板创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 3
4. **WHEN** 通过 H5 创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 4
5. **WHEN** 客户端在请求头中发送 `Client-Version: 2.10.0` 创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为 `2.10.0`
6. **WHEN** 客户端未发送 `Client-Version` 请求头创建桌台订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为空字符串

#### 具体要求

- [x] 3.1 在 `CreateDeskOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 3.2 在 `CreateDeskOrder` 方法中，创建 SaleBill 时根据 `ctx.GetVersion()` 设置 client_version 字段
- [x] 3.3 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 3.4 对 client_version 进行格式标准化处理
- [x] 3.5 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 3.6 确保 client_version 字段在 SaleBill 创建时正确设置
- [ ] 3.7 添加单元测试验证不同来源和版本的订单字段设置正确

---

### Requirement 4: 在创建会员端订单时设置 source 和 client_version

**用户故事**: 作为 系统，我想 在创建会员端订单时自动设置 source 和 client_version 字段，以便于 记录订单来源和客户端版本

#### 验收标准

1. **WHEN** 通过会员端创建订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 source 字段记录为 5
2. **WHEN** 通过其他来源创建会员端订单 **THEN** 系统 **SHALL** 根据实际来源设置 source 字段
3. **WHEN** 客户端在请求头中发送 `Client-Version: 2.10.0` 创建会员端订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为 `2.10.0`
4. **WHEN** 客户端未发送 `Client-Version` 请求头创建会员端订单 **THEN** 系统 **SHALL** 在 SaleBill 表的 client_version 字段记录为空字符串

#### 具体要求

- [x] 4.1 在 `createMemberOrder` 方法中，创建 SaleBill 时根据 `ctx.GetSource()` 设置 source 字段
- [ ] 4.2 在 `createMemberOrder` 方法中，创建 SaleBill 时根据 `ctx.GetVersion()` 设置 client_version 字段
- [x] 4.3 使用 Requirement 1 中定义的映射函数获取 source 值
- [ ] 4.4 对 client_version 进行格式标准化处理
- [x] 4.5 确保 source 字段在 SaleBill 创建时正确设置
- [ ] 4.6 确保 client_version 字段在 SaleBill 创建时正确设置
- [ ] 4.7 添加单元测试验证会员端订单字段设置正确

---

### Requirement 5: 确保所有创建 SaleBill 的路径都设置 source 和 client_version

**用户故事**: 作为 开发人员，我想 确保所有创建 SaleBill 的代码路径都正确设置 source 和 client_version 字段，以便于 保证数据一致性

#### 验收标准

1. **WHEN** 系统创建 SaleBill 记录 **THEN** 系统 **SHALL** 在所有创建路径中都设置 source 和 client_version 字段
2. **IF** 存在未设置 source 或 client_version 的创建路径 **THEN** 系统 **SHALL** 在代码审查中被识别并修复

#### 具体要求

- [x] 5.1 全面搜索代码库中所有 `CreateSaleBill` 调用
- [x] 5.2 确保所有创建 SaleBill 的地方都根据 `ctx.GetSource()` 设置 source 字段
- [ ] 5.3 确保所有创建 SaleBill 的地方都根据 `ctx.GetVersion()` 设置 client_version 字段
- [x] 5.4 对于无法获取来源的场景（如导入订单），使用默认值 0
- [ ] 5.5 对于无法获取版本的场景，使用空字符串作为默认值
- [x] 5.6 添加代码审查检查清单，确保不遗漏任何创建路径

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

**说明**: 
- `source` 字段已存在于 `ttpos_sale_bill` 表中，无需新增
- `client_version` 字段需要新增（Requirement 0）

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
- [x] 为 Source 映射函数添加单元测试
- [ ] 为每个订单创建方法添加 source 和 client_version 字段设置的测试用例
- [x] 测试不同 JWT Source 值映射到正确的 source 字段值（包括会员端 → 5）
- [x] 测试未知来源时使用默认值 0
- [ ] 测试客户端版本号格式标准化（去除前缀 `v`）
- [ ] 测试未发送 `Client-Version` 请求头时使用空字符串
- [ ] 测试不同版本的客户端创建订单时 client_version 字段设置正确

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
- [ ] 如果无法获取 client_version，使用空字符串并记录警告日志（可选，不强制）
- [ ] 确保 source 字段设置失败不影响订单创建（使用默认值）
- [ ] 确保 client_version 字段设置失败不影响订单创建（使用空字符串）

---

## 验收标准

### 功能验收

**订单来源追踪**：
1. **Source 映射正确性**: 所有 JWT Source 值都能正确映射到对应的 source 字段值（包括会员端映射到 5）
2. **即时订单 source 设置**: 通过不同客户端创建即时订单时，source 字段正确设置
3. **桌台订单 source 设置**: 通过不同客户端创建桌台订单时，source 字段正确设置
4. **会员端订单 source 设置**: 通过会员端创建订单时，source 字段设置为 5
5. **默认值处理**: 无法识别来源时，source 字段设置为 0
6. **数据一致性**: 所有创建 SaleBill 的路径都正确设置 source 字段

**客户端版本追踪**：
7. **数据库字段**: `client_version` 字段已成功添加到 `ttpos_sale_bill` 表
8. **版本号记录**: 客户端发送 `Client-Version` 请求头时，client_version 字段正确记录
9. **版本号格式标准化**: 版本号格式为 `v2.10.0` 时，自动去除前缀 `v`，记录为 `2.10.0`
10. **默认值处理**: 客户端未发送 `Client-Version` 请求头时，client_version 字段设置为空字符串
11. **数据一致性**: 所有创建 SaleBill 的路径都正确设置 client_version 字段
12. **版本查询**: 能够按客户端版本筛选和查询订单

### 测试验收

1. **单元测试**: 覆盖率达标，特别是 Source 映射函数、版本号格式标准化函数和订单创建方法
2. **API 测试**: 所有订单创建接口测试通过，验证 source 和 client_version 字段设置正确
3. **集成测试**: 端到端流程测试通过，验证不同来源和版本的订单字段正确
4. **手动测试**: 通过不同客户端和版本创建订单，验证数据库中的 source 和 client_version 字段值
5. **版本格式测试**: 测试不同格式的版本号（如 `2.10.0`、`v2.10.0`、`2.10`）都能正确处理

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 无需新增 API 文档
3. **数据库文档**: 迁移脚本和表结构文档完整（source 字段已存在，client_version 字段需新增）
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

- **向后兼容**: 历史订单的 source 字段为 0，client_version 字段为空，无法追溯真实来源和版本，这是可接受的
- **默认值处理**: 对于无法识别来源的请求，必须使用默认值 0；对于无法获取版本的请求，使用空字符串，不能导致订单创建失败
- **数据完整性**: 所有新创建的订单必须设置 source 和 client_version 字段，不能为 NULL
- **版本格式**: 客户端版本号统一处理为标准格式（x.y.z），去除前缀 `v` 等非标准字符

### 资源约束

- 开发时间: 3-4 天
- Story Point: 5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/constant/jwt` - JWT Source 常量定义
- `main/pkg/context` - Context 接口，提供 `GetSource()` 和 `GetVersion()` 方法
- `main/app/model` - SaleBill 模型定义
- `main/app/service` - 订单创建服务
- `admin/database/migrations` - 数据库迁移脚本

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

### 风险 4: 客户端版本格式不一致

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在设置 client_version 时，统一处理版本号格式（去除前缀 `v`，统一为 x.y.z 格式）
- 添加版本号格式标准化函数，确保数据一致性
- 对于无法识别的版本格式，记录原始值，并在日志中记录警告

### 风险 5: 请求头缺失

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 对于未发送 `Client-Version` 请求头的客户端，使用空字符串作为默认值
- 不影响订单创建流程的正常执行
- 可选：记录警告日志，便于后续排查和修复

---

## 时间表

- **Phase 0 - 数据库迁移**: 0.5 天（新增 client_version 字段）
- **Phase 1 - 定义常量映射**: 0.5 天
- **Phase 2 - 修改订单创建逻辑**: 1.5-2 天（设置 source 和 client_version）
- **Phase 3 - 单元测试和验证**: 0.5-1 天
- **总计**: 3-4 天（SP = 5）

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
- `main/pkg/context/context.go` - Context 接口实现（GetSource 和 GetVersion 方法）
- `admin/database/migrations/20251126201148_add_source_to_sale_bill.php` - 数据库迁移文件（source 字段）
- `admin/database/migrations/{待创建}_add_client_version_to_sale_bill.php` - 数据库迁移文件（client_version 字段）

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

