# Kiosk 自助点餐机首页功能模块 任务分解

> 本文档定义 Kiosk 自助点餐机首页功能模块的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 6  
**进行中**: -  
**完成率**: 40%

---

## Phase 1: DTO 和 Service 层

### DTO 层

- [x] 1.1 创建 KioskBase 响应 DTO

  - File: `main/app/dto/resp/base.go`
  - Purpose: 定义自助点餐机首页基本信息响应结构体
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 DTO: `main/app/dto/resp/base.go`（参考 `CashierBase`, `TabletBase`, `AssistantBase` 结构）
  - Prompt: Role: Go Developer | Task: 在 base.go 中添加 KioskBase 结构体，包含 username, device_id, device_remark, company, currency, business, kiosk, update_time 字段 | Context: kiosk 字段类型为 setting.KioskResp，包含 language_list 和 carousel | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段命名使用 snake_case | Success: DTO 创建成功，字段定义正确

### Service 层

- [x] 1.2 在 Auth Service 中添加 KioskBase 方法

  - File: `main/app/service/auth.go`
  - Purpose: 实现获取自助点餐机基本信息业务逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有方法: `main/app/service/auth.go` - `CashierBase()`, `TabletBase()`, `AssistantBase()` 方法
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 auth.go 中实现 KioskBase() 方法，参考 CashierBase() 和 TabletBase() 的实现方式 | Context: 获取商家信息、设备信息、货币设置、业务设置、自助点餐机设置（包含语言列表和轮播广告），使用 settingSrv.GetKioskSetting() 获取配置 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口，不使用 panic，返回 error | Success: KioskBase() 方法实现完整，业务逻辑正确，错误处理正确

- [ ] 1.3 实现配置缓存机制

  - File: `main/app/service/auth.go`
  - Purpose: 实现自助点餐机配置的 Redis 缓存，提升性能
  - Requirements: 1.6
  - Leverage: 现有缓存实现: `main/app/service/` 中的缓存使用方式
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 KioskBase() 方法中实现配置缓存，使用 Cache-Aside Pattern | Context: Key 格式为 `ttpos:kiosk:base:{company_uuid}:{device_id}`，过期时间 5 分钟，缓存未命中时查询数据库并写入缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存机制实现完成，缓存命中率 > 80%

- [ ] 1.4 编写 Service 单元测试

  - File: `main/app/service/auth_test.go`
  - Purpose: 确保 KioskBase() 方法业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/auth_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 KioskBase() 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试正常场景、配置获取失败场景、缓存场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 2: API 层

### Base API

- [x] 2.1 创建 Base Handler

  - File: `main/app/api/v1/kiosk/kiosk_base.go`
  - Purpose: 实现获取首页基本信息 API 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_base.go`, `main/app/api/v1/tablet/tablet_base.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 kiosk_base.go，实现 GetBase() 方法和 RegisterBaseHandlers() 函数 | Context: 使用 BaseHandler 结构体，依赖 authSrv, settingSrv, deviceSrv，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 2.2 注册 Base API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/base` 路由
  - Requirements: 1.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Prompt: Role: Go Developer | Task: 在 router.go 的 kioskGroup 中注册 base handlers | Context: 调用 kiosk.RegisterBaseHandlers(kioskGroup, dbm, cache) | Restrictions: 遵循现有路由注册模式 | Success: 路由注册成功，接口可访问

### Call API

- [x] 2.3 创建 Call Handler

  - File: `main/app/api/v1/kiosk/kiosk_call.go`
  - Purpose: 实现呼叫服务员 API 接口
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: 现有 API: `main/app/api/v1/tablet/tablet_call.go`, `main/app/api/v1/h5/h5_handler.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 kiosk_call.go，实现 Call() 方法和 RegisterCallHandlers() 函数 | Context: 使用 CallHandler 结构体，依赖 callSrv，调用 callSrv.Call() 方法，返回成功提示 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 2.4 注册 Call API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/call` 路由
  - Requirements: 5.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Prompt: Role: Go Developer | Task: 在 router.go 的 kioskGroup 中注册 call handlers | Context: 调用 kiosk.RegisterCallHandlers(kioskGroup, dbm, cache) | Restrictions: 遵循现有路由注册模式 | Success: 路由注册成功，接口可访问

- [ ] 2.5 编写 API 集成测试

  - File: `main/app/api/v1/kiosk/kiosk_base_test.go`, `main/app/api/v1/kiosk/kiosk_call_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/tablet/tablet_base_test.go`, `main/app/api/v1/tablet/tablet_call_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 kiosk base 和 call API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 错误处理和优化

- [ ] 3.1 实现配置获取失败降级方案

  - File: `main/app/service/auth.go`
  - Purpose: 配置获取失败时使用默认配置，保证系统可用性
  - Requirements: 1.7, 可靠性要求
  - Leverage: 现有错误处理: `main/app/service/auth.go` 中的错误处理方式
  - Prompt: Role: Go Developer | Task: 在 KioskBase() 方法中实现配置获取失败的降级方案 | Context: 当 GetKioskSetting() 失败时，使用默认配置（默认语言、空轮播广告），记录警告日志 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 降级方案实现完成，系统可用性保证

- [ ] 3.2 实现呼叫服务错误处理

  - File: `main/app/api/v1/kiosk/kiosk_call.go`
  - Purpose: 呼叫服务员失败时返回友好错误提示
  - Requirements: 5.4, 可靠性要求
  - Leverage: 现有错误处理: `main/app/api/v1/tablet/tablet_call.go` 中的错误处理方式
  - Prompt: Role: Go Developer | Task: 在 Call() 方法中实现错误处理，返回友好错误提示 | Context: 捕获 callSrv.Call() 的错误，使用 errors.WithMessage() 包装错误，返回友好的错误信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 错误处理实现完成，错误提示友好

- [ ] 3.3 增加呼叫日志记录

  - File: `main/app/service/call.go`
  - Purpose: 记录呼叫日志，便于问题排查
  - Requirements: 5.5
  - Leverage: 现有日志: `main/app/service/call.go` 中的日志记录方式
  - Prompt: Role: Go Developer | Task: 在 Call() 方法中增加日志记录 | Context: 记录呼叫请求信息（设备ID、呼叫类型、桌台信息），记录呼叫成功/失败日志 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 logger.Logger | Success: 日志记录实现完成，日志信息完整

- [ ] 3.4 性能优化和缓存优化

  - File: `main/app/service/auth.go`
  - Purpose: 优化数据库查询和缓存策略
  - Requirements: 性能要求
  - Leverage: 现有优化: `main/app/service/auth.go` 中的查询优化方式
  - Prompt: Role: Performance Engineer | Task: 优化 KioskBase() 方法的性能，减少数据库查询次数，提升缓存命中率 | Context: 批量获取配置，使用索引查询，优化缓存 Key 设计 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 性能优化完成，响应时间 < 200ms，缓存命中率 > 80%

---

## Phase 4: 集成测试和文档

- [ ] 4.1 端到端集成测试

  - File: `test/integration/kiosk_home_page_test.go`
  - Purpose: 测试端到端功能流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户登录后获取首页信息，测试呼叫服务员流程，测试配置缓存，测试错误场景 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 更新 API 文档

  - File: `docs/shared/api/kiosk_api.md`
  - Purpose: 更新 API 接口文档
  - Requirements: 文档要求
  - Leverage: 现有 API 文档: `docs/shared/api/`
  - Prompt: Role: Technical Writer | Task: 更新 Kiosk API 文档，添加 base 和 call 接口说明 | Context: 包含接口路径、请求参数、响应格式、错误码说明 | Restrictions: 文档准确完整 | Success: API 文档已更新

- [ ] 4.3 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG: `CHANGELOG.md`
  - Prompt: Role: Technical Writer | Task: 在 CHANGELOG.md 中记录 Kiosk 首页功能模块的变更 | Context: 记录新增接口、功能说明、版本号 | Restrictions: 遵循 CHANGELOG 格式 | Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - API: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 设计文档已更新（如有变更）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-kiosk-home-page/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-home-page/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-kiosk-home-page/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-home-page/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-kiosk-home-page/tasks.md)" | bc
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
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-18  
**维护者**: 后端开发组

