# Kiosk 自助点餐机登录认证功能 任务分解

> 本文档定义 Kiosk 自助点餐机登录认证功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 8  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: 常量定义

- [x] 1.1 添加 SourceKiosk 常量到 JWT 包

  - File: `main/app/constant/jwt/jwt.go`
  - Purpose: 定义 Kiosk 终端的 Source 常量，用于 JWT Token 中标识来源
  - Requirements: 1.5
  - Leverage: 现有常量定义: `main/app/constant/jwt/jwt.go`，参考 `SourceCashier`, `SourceTablet`, `SourceAssistant`
  - Prompt: Role: Go Developer | Task: 在 jwt 包中添加 SourceKiosk 常量，值为 "kiosk" | Context: 参考现有 Source 常量的定义方式，添加到常量组中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: SourceKiosk 常量添加成功，值为 "kiosk"

- [x] 1.2 添加 SourceKiosk 常量到 device 包

  - File: `main/app/constant/device.go`
  - Purpose: 定义 Kiosk 终端的 Source 常量，用于设备来源标识
  - Requirements: 1.5
  - Leverage: 现有常量定义: `main/app/constant/device.go`，参考 `SourceCashier`, `SourceTablet`, `SourceAssistant`
  - Prompt: Role: Go Developer | Task: 在 device 包中添加 SourceKiosk 常量，值为 "kiosk"，并添加到 SourceTextMap 中，文本为 "自助点餐机" | Context: 参考现有 Source 常量的定义方式，添加到常量组和文本映射中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: SourceKiosk 常量添加成功，文本映射正确

---

## Phase 2: API 实现

### API 层

- [x] 2.1 创建 Kiosk API 目录和认证文件

  - File: `main/app/api/v1/kiosk/kiosk_auth.go`
  - Purpose: 创建 Kiosk 终端的认证 API 文件
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_auth.go`, `main/app/api/v1/tablet/tablet_auth.go`, `main/app/api/v1/assistant/assistant_auth.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 kiosk_auth.go 文件，定义 AuthHandler 结构体和 RegisterAuthHandlers 函数 | Context: 参考 cashier_auth.go 的实现，使用相同的结构和初始化方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc | Success: 文件创建成功，结构体定义正确

- [x] 2.2 实现登录接口

  - File: `main/app/api/v1/kiosk/kiosk_auth.go`
  - Purpose: 实现登录接口，支持邮箱/手机号登录，验证码验证
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有实现: `main/app/api/v1/cashier/cashier_auth.go` 的 `Login()` 方法，统一认证服务: `main/app/service/auth.go` 的 `SaasLogin()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 Login() 方法，设置 Source 为 constant.SourceKiosk，使用统一认证服务 SaasLogin()，支持版本判断（版本号 >= 2.11.0） | Context: 参考 cashier_auth.go 的 Login() 实现，使用相同的逻辑和错误处理方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc，不使用 panic，返回 error | Success: 登录接口实现成功，支持统一认证，版本判断正确

- [x] 2.3 实现刷新 Token 接口

  - File: `main/app/api/v1/kiosk/kiosk_auth.go`
  - Purpose: 实现刷新 Token 接口，支持 Token 自动刷新
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有实现: `main/app/api/v1/cashier/cashier_auth.go` 的 `RefreshToken()` 方法，统一认证服务: `main/app/service/auth.go` 的 `RefreshToken()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 RefreshToken() 方法，调用统一认证服务的 RefreshToken() 方法 | Context: 参考 cashier_auth.go 的 RefreshToken() 实现，使用相同的逻辑和错误处理方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc | Success: 刷新 Token 接口实现成功，错误处理正确

- [x] 2.4 实现退出登录接口

  - File: `main/app/api/v1/kiosk/kiosk_auth.go`
  - Purpose: 实现退出登录接口，清除登录状态
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有实现: `main/app/api/v1/cashier/cashier_auth.go` 的 `Logout()` 方法，统一认证服务: `main/app/service/auth.go` 的 `Logout()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 Logout() 方法，调用统一认证服务的 Logout() 方法 | Context: 参考 cashier_auth.go 的 Logout() 实现，使用相同的逻辑和错误处理方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc | Success: 退出登录接口实现成功，错误处理正确

- [x] 2.5 实现路由注册函数

  - File: `main/app/api/v1/kiosk/kiosk_auth.go`
  - Purpose: 实现 RegisterAuthHandlers() 函数，注册认证路由
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有实现: `main/app/api/v1/cashier/cashier_auth.go` 的 `RegisterAuthHandlers()` 函数
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 RegisterAuthHandlers() 函数，初始化服务，注册登录、刷新Token、退出登录路由 | Context: 参考 cashier_auth.go 的 RegisterAuthHandlers() 实现，使用相同的服务初始化方式，登录接口放在 publicApi，其他接口放在 privateApi（需要认证） | Restrictions: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc | Success: 路由注册函数实现成功，路由分组正确

---

## Phase 3: 路由注册

- [x] 3.1 在路由文件中注册 Kiosk 路由组

  - File: `main/router/router.go`
  - Purpose: 在路由文件中注册 Kiosk 路由组，调用 RegisterAuthHandlers() 注册认证路由
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有路由注册: `main/router/router.go`，参考 cashier、tablet、assistant 的路由注册方式
  - Prompt: Role: Go Developer | Task: 在 router.go 中添加 kiosk 包的导入，创建 kioskGroup 路由组，调用 kiosk.RegisterAuthHandlers() 注册认证路由 | Context: 参考 cashier、tablet、assistant 的路由注册方式，使用相同的模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Kiosk 路由组注册成功，认证路由正确

---

## Phase 4: 测试

- [ ] 4.1 编写登录接口单元测试

  - File: `main/app/api/v1/kiosk/kiosk_auth_test.go`
  - Purpose: 测试登录接口的功能和错误处理
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有测试: `main/app/api/v1/cashier/cashier_auth_test.go`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Login() 方法编写单元测试，测试正常登录、错误账号密码、错误验证码、版本判断等场景 | Context: 使用 gin 的测试工具，模拟请求和响应 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 登录接口测试通过，覆盖率 ≥ 70%

- [ ] 4.2 编写刷新 Token 接口单元测试

  - File: `main/app/api/v1/kiosk/kiosk_auth_test.go`
  - Purpose: 测试刷新 Token 接口的功能和错误处理
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有测试: `main/app/api/v1/cashier/cashier_auth_test.go`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RefreshToken() 方法编写单元测试，测试正常刷新、Token 过期、RefreshToken 过期等场景 | Context: 使用 gin 的测试工具，模拟请求和响应 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 刷新 Token 接口测试通过，覆盖率 ≥ 70%

- [ ] 4.3 编写退出登录接口单元测试

  - File: `main/app/api/v1/kiosk/kiosk_auth_test.go`
  - Purpose: 测试退出登录接口的功能和错误处理
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有测试: `main/app/api/v1/cashier/cashier_auth_test.go`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Logout() 方法编写单元测试，测试正常退出、Token 无效等场景 | Context: 使用 gin 的测试工具，模拟请求和响应 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 退出登录接口测试通过，覆盖率 ≥ 70%

- [ ] 4.4 编写集成测试

  - File: `test/integration/kiosk_auth_test.go`（或相应测试目录）
  - Purpose: 测试端到端登录流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，测试登录、Token 刷新、退出登录的完整流程 | Context: 测试真实用户场景，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，端到端流程正常

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - API: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-kiosk-auth/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-auth/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-kiosk-auth/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-auth/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-kiosk-auth/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-17  
**维护者**: 后端开发组

