# 根据邮箱或手机号查询员工接口 需求文档

> 本文档定义根据邮箱或手机号查询员工接口的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/query-employee-by-contact.md](../../../../team/proposals/2025-12/query-employee-by-contact.md) |
| **创建日期**      | 2025-12-23                                                                                                 |
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

当前收银机/点餐助手在进行敏感操作（如退菜、打折、取消订单等）时，需要手动输入授权员工的邮箱或手机号，操作时间较长，影响收银效率。

本功能通过新增一个简化的员工查询接口，支持根据邮箱或手机号（支持模糊搜索）查询员工信息，返回员工基本信息（UUID、姓名、邮箱、手机号），用于前端下拉列表展示和快速选择，从而提升收银效率和用户体验。

## 🎯 产品对齐

- **提升收银效率**：减少授权操作时间，提高高峰期处理能力
- **改善用户体验**：简化操作流程，降低输入错误率
- **提高准确性**：通过下拉选择避免输入错误
- **支持快速授权**：常用授权员工可快速选择，无需完整输入

---

## 📝 用户故事

**作为** 收银员/点餐助手  
**我想** 通过下拉列表快速选择授权员工（而不需要手动输入完整邮箱/手机号）  
**以便于** 提高敏感操作的授权效率，减少操作时间

---

## 功能需求

### Requirement 1: 根据邮箱或手机号查询员工接口

**用户故事**: 作为收银员/点餐助手，我想通过下拉列表快速选择授权员工，以便于提高敏感操作的授权效率

#### 验收标准

1. **WHEN** 收银员点击授权账号输入框或下拉箭头按钮 **THEN** 系统 **SHALL** 显示下拉列表，下拉列表的员工为 Shop 管理端所配置的授权员工
2. **WHEN** 收银员在输入框中输入关键词（邮箱或手机号） **THEN** 系统 **SHALL** 实时过滤下拉列表，仅显示匹配的员工（模糊搜索，不区分大小写）
3. **WHEN** 收银员选择下拉列表中的员工 **THEN** 系统 **SHALL** 自动填充该员工的账号到输入框，并关闭下拉列表
4. **WHEN** 收银员手动输入账号 **THEN** 系统 **SHALL** 保持原有验证逻辑不变，下拉列表根据输入内容实时过滤
5. **WHEN** 收银员点击下拉列表外部区域 **THEN** 系统 **SHALL** 关闭下拉列表
6. **WHEN** Shop 管理端更新授权员工配置 **THEN** 系统 **SHALL** 在授权时获取最新列表
7. **WHEN** 查询结果超过 20 条 **THEN** 系统 **SHALL** 仅返回前 20 条匹配结果

#### 具体要求

- [x] 1.1 新增 API 接口 `GET /cashier/order/query-staff-by-contact` 和 `GET /assistant/order/query-staff-by-contact`，支持根据关键词查询员工
- [ ] 1.2 接口支持邮箱和手机号的模糊搜索（不区分大小写）
- [ ] 1.3 接口自动识别输入的是邮箱还是手机号格式
- [ ] 1.4 接口返回员工基本信息：UUID、姓名、邮箱、手机号
- [ ] 1.5 接口返回结果限制在 20 条以内
- [ ] 1.6 接口仅返回当前门店可见范围内的授权员工
- [ ] 1.7 接口基于 Shop 管理端配置的授权员工列表进行过滤
- [ ] 1.8 接口需要身份验证（JWT Token）
- [ ] 1.9 接口需要权限验证（确保用户有查询权限）

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

- [x] URL 使用 snake_case 命名（如：`/cashier/order/query-staff-by-contact`）
- [x] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中（本接口不需要分页）
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

**接口设计**：

**请求参数**：
- `keyword` (string, 可选): 搜索关键词（邮箱或手机号，支持模糊匹配）
- `limit` (int, 可选): 返回结果数量限制（默认 20，最大 20）

**响应数据**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123456,
        "real_name": "张三",
        "email": "zhangsan@example.com",
        "phone": "13800138000"
      }
    ]
  }
}
```

### 数据库设计要求

- [x] 复用现有 `saas.ttpos_staff` 表，无需新增表
- [x] 使用现有索引优化查询性能（`idx_phone`, `uk_email`）
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用现有索引）
- [ ] 缓存策略（Redis）- 可选，根据实际性能决定
- [ ] 并发处理（使用 UUID 锁）- 本接口为查询接口，无需锁

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+（不适用，本功能用于 POS 收银端和助手端）
- [ ] Safari 14+（不适用）
- [ ] Firefox 88+（不适用）
- [ ] Edge 90+（不适用）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（JWT Token）
- [x] 权限控制（仅返回当前门店可见的授权员工）
- [x] SQL 注入防护（使用参数化查询）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 错误日志记录（使用 Logger）
- [x] 查询失败时返回空列表，不抛出异常

---

## 验收标准

### 功能验收

1. **接口可用性**: 接口能够正常响应请求，返回正确的员工信息
2. **搜索功能**: 支持邮箱和手机号的模糊搜索，不区分大小写
3. **结果限制**: 返回结果不超过 20 条
4. **权限控制**: 仅返回当前门店可见范围内的授权员工
5. **兼容性**: 保持现有手动输入方式不变，新接口作为增强功能

### 测试验收

1. **单元测试**: Service 和 Repository 层测试覆盖率达标
2. **API 测试**: 接口测试通过，包括正常场景和异常场景
3. **集成测试**: 端到端流程测试通过
4. **性能测试**: 响应时间 < 200ms

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整（Swagger 注释）
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

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

- 仅返回 Shop 管理端配置的授权员工
- 仅返回当前门店可见范围内的员工
- 保持现有授权验证逻辑不变

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/repository/saas_staff.go` - 员工数据访问层
- `main/app/service/saas_staff.go` - 员工服务层
- `main/app/api/v1/shop/shop_staff.go` - 员工管理接口

### 服务依赖

- **Main → BMP**: 无依赖
- **Admin → Main**: 无依赖
- **Frontend → Admin**: 无依赖

### 业务依赖

- Shop 管理端的授权员工配置功能（已存在）
- 门店可见性逻辑（已存在）
- 现有员工查询逻辑（已存在，可复用）

---

## 风险和缓解

### 风险 1: 性能风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 限制返回结果数量（最多 20 条）
- 使用数据库索引优化查询（现有 `idx_phone`, `uk_email`）
- 如果性能不达标，考虑添加 Redis 缓存

### 风险 2: 权限风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 复用现有的门店可见性逻辑
- 确保只返回当前门店可见的授权员工
- 添加单元测试验证权限控制逻辑

### 风险 3: 兼容性风险

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 保持现有手动输入方式不变
- 新接口作为增强功能，不影响现有流程
- 前端向后兼容，支持两种方式

---

## 时间表

- **Phase 1 - 后端 API 开发**: 1 天
- **Phase 2 - 前端组件开发**: 1 天
- **Phase 3 - 联调测试**: 0.5-1 天
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
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 相关代码

- `main/app/api/v1/shop/shop_staff.go` - 现有员工管理接口
- `main/app/service/saas_staff.go` - 员工服务层
- `main/app/repository/saas_staff.go` - 员工数据访问层
- `main/app/dto/req/staff.go` - 员工请求 DTO
- `main/app/dto/resp/saas_staff.go` - 员工响应 DTO

### 外部参考

- DooTask #37897 - 产品需求任务

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: xiezhihuan  
**审核者**: {审核者}

