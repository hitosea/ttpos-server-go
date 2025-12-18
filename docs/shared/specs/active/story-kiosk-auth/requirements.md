# Kiosk 自助点餐机登录认证功能 需求文档

> 本文档定义 Kiosk 自助点餐机登录认证功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-kiosk-auth.md](../../../../team/proposals/2025-12/v2.12.0-kiosk-auth.md) |
| **创建日期**      | 2025-12-17                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | -                        |
| **审核意见** | -                        |

---

## 📋 概述

实现自助点餐机终端的登录认证功能，使用商家员工账号登录（与 POS、Assistant、Tablet 等终端一致），支持邮箱/手机号登录、图形验证码验证、记住密码功能，确保设备安全启动。

**实现范围**：实现后端 API 接口，参考收银机（Cashier）的登录认证接口实现。

## 🎯 产品对齐

- **设备安全**：确保只有授权员工才能启动自助点餐机设备
- **统一认证**：与 POS、Assistant、Tablet 等终端使用相同的登录认证机制，降低维护成本
- **用户体验**：支持记住密码功能，减少重复登录操作
- **安全防护**：图形验证码验证，防止暴力破解攻击

## 📝 用户故事

**作为** 门店员工  
**我想** 使用我的员工账号登录自助点餐机  
**以便于** 启动设备，让顾客使用自助点餐功能

**作为** 门店员工  
**我想** 在自助点餐机上使用记住密码功能  
**以便于** 减少重复登录操作，提升使用效率

---

## 功能需求

### Requirement 1: 登录功能

**用户故事**: 作为门店员工，我想使用我的员工账号登录自助点餐机，以便于启动设备，让顾客使用自助点餐功能

#### 验收标准

1. **WHEN** 员工打开自助点餐机 **THEN** 系统 **SHALL** 显示登录界面，包含账号输入框、密码输入框、验证码输入框

2. **WHEN** 员工输入账号（邮箱/手机号）和密码 **THEN** 系统 **SHALL** 显示图形验证码，要求用户输入验证码

3. **WHEN** 员工输入正确的账号、密码和验证码 **THEN** 系统 **SHALL** 调用统一认证接口（SaasLogin），验证员工身份，登录成功返回 Token 和 RefreshToken

4. **WHEN** 员工输入错误的账号、密码或验证码 **THEN** 系统 **SHALL** 显示错误提示，要求重新输入

5. **IF** 员工账号未授权或已禁用 **THEN** 系统 **SHALL** 返回错误提示，拒绝登录

6. **WHEN** 员工登录成功 **THEN** 系统 **SHALL** 保存 Token，进入首页界面

#### 具体要求

- [ ] 1.1 实现登录接口 (`/kiosk/login` POST)，支持邮箱/手机号登录，验证码验证
- [ ] 1.2 使用统一认证服务（SaasLogin），与 POS、Assistant、Tablet 等终端一致
- [ ] 1.3 登录接口返回 Token 和 RefreshToken
- [ ] 1.4 登录接口需要验证码 sign（X-SIGN header）
- [ ] 1.5 登录接口设置 Source 为 `constant.SourceKiosk`
- [ ] 1.6 支持版本判断（版本号 >= 2.11.0 使用统一认证登录）

**参考实现**：
- `/cashier/login` - 收银机登录接口
- `/tablet/login` - 平板端登录接口
- `/assistant/login` - 助手端登录接口

---

### Requirement 2: Token 刷新功能

**用户故事**: 作为门店员工，我想在 Token 过期后自动刷新，以便于无需重复登录即可继续使用设备

#### 验收标准

1. **WHEN** Token 过期 **THEN** 系统 **SHALL** 自动使用 RefreshToken 刷新 Token
2. **WHEN** RefreshToken 也过期 **THEN** 系统 **SHALL** 要求用户重新登录
3. **WHEN** 调用刷新 Token 接口 **THEN** 系统 **SHALL** 返回新的 Token 和 RefreshToken

#### 具体要求

- [ ] 2.1 实现刷新 Token 接口 (`/kiosk/refresh_token` GET)
- [ ] 2.2 刷新 Token 接口需要 JWT Token 认证
- [ ] 2.3 刷新 Token 接口返回新的 Token 和 RefreshToken
- [ ] 2.4 前端实现 Token 自动刷新机制

**参考实现**：
- `/cashier/refresh_token` - 收银机刷新 Token 接口
- `/tablet/refresh_token` - 平板端刷新 Token 接口

---

### Requirement 3: 退出登录功能

**用户故事**: 作为门店员工，我想退出登录，以便于清除登录状态，保护设备安全

#### 验收标准

1. **WHEN** 员工点击退出登录 **THEN** 系统 **SHALL** 清除登录状态，返回登录界面
2. **WHEN** 退出登录成功 **THEN** 系统 **SHALL** 清除本地存储的 Token 和密码

#### 具体要求

- [ ] 3.1 实现退出登录接口 (`/kiosk/logout` POST)
- [ ] 3.2 退出登录接口需要 JWT Token 认证
- [ ] 3.3 退出登录接口清除服务端的登录状态
- [ ] 3.4 前端清除本地存储的 Token 和密码

**参考实现**：
- `/cashier/logout` - 收银机退出登录接口
- `/tablet/logout` - 平板端退出登录接口

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

- [x] URL 使用 snake_case 命名（如：`/kiosk/login`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 无需新增数据库表（复用现有员工账号和认证表）

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 登录接口响应时间 < 500ms
- [ ] Token 刷新接口响应时间 < 200ms

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖登录流程
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有错误提示使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（除登录接口外）
- [x] 登录接口需要验证码 sign（X-SIGN header）
- [x] 密码加密传输（HTTPS）
- [x] Token 安全存储
- [x] 图形验证码防止暴力破解
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 错误日志记录（使用 Logger）
- [ ] Token 刷新失败时提示重新登录

---

## 验收标准

### 功能验收

1. **登录功能**: 员工可以使用邮箱/手机号登录，验证码验证，登录成功返回 Token
2. **Token 刷新**: Token 过期后自动刷新，无需重新登录
3. **退出登录**: 退出登录清除登录状态，返回登录界面
4. **统一认证**: 与 POS、Assistant、Tablet 等终端使用相同的登录认证机制

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过（登录、刷新Token、退出登录）
3. **集成测试**: 端到端登录流程测试通过
4. **安全测试**: 验证码验证、Token 安全测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整
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
- 复用现有的统一认证服务（SaasLogin）

### 业务约束

- 登录认证方式与 POS、Assistant、Tablet 等终端保持一致
- 每个员工账号都可以登录自助点餐机
- 支持版本判断（版本号 >= 2.11.0 使用统一认证登录）
- 记住密码功能由前端实现（加密存储）

### 资源约束

- 开发时间: 3-4 天
- Story Point: 5-8 (必须 ≤ 5，可能需要拆分)

---

## 依赖关系

### 技术依赖

- `现有统一认证服务` - SaasLogin 服务
- `现有验证码服务` - CaptchaSrv 服务
- `现有认证中间件` - Auth 中间件

### 服务依赖

- **Kiosk → Main**: HTTP API 调用（登录认证接口）

### 业务依赖

- 统一认证服务（已存在）
- 员工账号管理（已存在）
- 验证码服务（已存在）

---

## 风险和缓解

### 风险 1: 统一认证兼容性风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 参考收银机、平板端等终端的登录认证实现，确保兼容性
- 使用相同的统一认证服务（SaasLogin）
- 设置正确的 Source 常量（`constant.SourceKiosk`）

### 风险 2: 验证码显示风险

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 验证码显示适配大屏设备，确保清晰可见
- 参考其他终端的验证码实现

### 风险 3: Token 管理风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 参考收银机、平板端等终端的 Token 管理实现
- 实现 Token 自动刷新机制
- Token 安全存储（前端加密存储）

---

## 时间表

- **Phase 1 - 登录接口实现**: 1 天
- **Phase 2 - Token 刷新接口**: 0.5 天
- **Phase 3 - 退出登录接口**: 0.5 天
- **Phase 4 - 测试与联调**: 1-2 天
- **总计**: 3-4 天（SP = 5-8，可能需要拆分）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `main/app/api/v1/cashier/cashier_auth.go` - 收银机登录认证接口
- `main/app/api/v1/tablet/tablet_auth.go` - 平板端登录认证接口
- `main/app/api/v1/assistant/assistant_auth.go` - 助手端登录认证接口
- `main/app/service/auth.go` - 统一认证服务

### 架构文档

- `docs/human/architecture/features/auth.md` - 统一认证架构文档

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

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

