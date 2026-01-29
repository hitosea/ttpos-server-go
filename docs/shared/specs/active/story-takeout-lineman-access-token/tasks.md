# LINE MAN OAuth Access Token 缓存功能 任务分解

> 本文档定义 LINE MAN OAuth Access Token 缓存功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 10  
**进行中**: 0  
**完成率**: 91%

---

## Phase 1: 配置和基础结构 ✅ 已完成

- [x] 1.1 添加 endpoint 配置到 config.tpl.yaml
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
  - Purpose: 配置 LINE MAN API 端点地址
  - Requirements: 5.1, 5.2, 5.4
  - Status: ✅ 已完成（v2.13.1）

- [x] 1.2 更新 Lineman 配置结构体
  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/conf/provider.go`
  - Purpose: 添加 Endpoint 字段
  - Requirements: 5.3
  - Status: ✅ 已完成（v2.13.1）

- [x] 1.3 创建配置文档
  - Files: `lineman-env-example.md`, `CHANGELOG-v2.13.1.md`
  - Purpose: 提供配置示例和变更说明
  - Requirements: 5.8
  - Status: ✅ 已完成（v2.13.1）

---

## Phase 2: 核心实现（Go BMP Logic 层）

### OAuth Token 获取

- [x] 2.1 创建 OAuth 响应 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/oauth.go` (新建)
  - Purpose: 定义 LINE MAN OAuth API 响应数据结构
  - Requirements: 1.5
  - Leverage: 参考 `internal/model/dto/grab/` 目录下的 DTO 定义
  - Prompt: Role: Go Developer | Task: 创建 LinemanOAuthTokenResp 结构体，映射 LINE MAN OAuth API 响应 | Context: 字段包含 access_token, expires_in, token_type，使用 json 标签 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，包含中文注释 | Success: DTO 定义完整，字段映射正确

- [x] 2.2 实现 FetchTokenFromAPI() 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token.go`
  - Purpose: 从 LINE MAN OAuth 服务器获取 Access Token
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8
  - Leverage: 参考 `internal/logic/grab/grab.go` 的 `fetchTokenFromSDK()` 方法（L119-L149）
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 实现 FetchTokenFromAPI() 方法，调用 LINE MAN OAuth API | Context: 使用 g.Client().Post() 发送请求，endpoint 从配置读取，请求体包含 grant_type, client_id, client_secret, scope，解析响应的 access_token 和 expires_in | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 处理错误，使用 g.Log() 记录日志，敏感信息不记录 | Success: 方法实现完整，OAuth API 调用成功，错误处理完善

### Redis Token 缓存

- [x] 2.3 在 sLinemanToken 结构体中添加 tokenLock 字段

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token.go`
  - Purpose: 添加互斥锁，用于双重检查锁机制
  - Requirements: 3.2
  - Leverage: 参考 `internal/logic/grab/grab.go` 的结构体定义
  - Code:
    ```go
    type sLinemanToken struct {
        cfgLoader *PartnerConfigLoader
        secretKey string
        expiresIn int
        tokenLock sync.Mutex  // ✨ 新增：互斥锁
    }
    ```
  - Success: 字段添加成功

- [x] 2.4 实现 GetAccessToken() 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token.go`
  - Purpose: 实现 Redis 缓存 + 双重检查锁机制
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7
  - Leverage: **直接复制** `internal/logic/grab/grab.go` 的 `getAccessToken()` 方法（L151-L193），调整为 LINE MAN 版本
  - Prompt: Role: Go Developer with Redis and concurrency expertise | Task: 实现 GetAccessToken() 方法，使用 Redis 缓存和双重检查锁（DCL）| Context: Redis Key 格式为 `lineman:oauth:token:{client_id}`，TTL = expires_in - 60 秒，第一次检查 Redis（无锁），获取互斥锁，第二次检查 Redis（持锁），调用 FetchTokenFromAPI()，写入 Redis | Restrictions: Redis 故障不影响业务，仅记录警告日志 | Success: 方法实现完整，双重检查锁正确，Redis 缓存正确，并发安全

### Authorization Header 生成

- [x] 2.5 实现 GetAuthorizationHeader() 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token.go`
  - Purpose: 生成格式化的 Authorization 请求头
  - Requirements: 4.1, 4.2, 4.3, 4.4
  - Leverage: **直接复制** `internal/logic/grab/grab.go` 的 `getAuthorizationHeader()` 方法（L195-L202）
  - Prompt: Role: Go Developer | Task: 实现 GetAuthorizationHeader() 方法，返回 Bearer Token 格式 | Context: 调用 GetAccessToken() 获取 Token，返回 "Bearer " + token | Restrictions: 错误传递，不额外处理 | Success: 方法实现完整，返回格式正确

### Service 接口更新

- [x] 2.6 更新 Service 接口定义

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman_token.go`
  - Purpose: 将新方法添加到 Service 接口
  - Requirements: 所有功能需求
  - Leverage: 现有 Service 接口定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen service`
  - Prompt: Role: Go Developer | Task: 在 lineman_token.go 中添加新方法签名 | Context: 添加 FetchTokenFromAPI(), GetAccessToken(), GetAuthorizationHeader() 三个方法 | Restrictions: 遵循 GoFrame 规范，方法签名与实现一致 | Success: Service 接口更新成功

---

## Phase 3: 测试和文档

### 单元测试

- [x] 3.1 编写 FetchTokenFromAPI() 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token_test.go` (新建或追加)
  - Purpose: 测试 OAuth API 调用逻辑
  - Requirements: 1.1-1.8
  - Leverage: 参考 `internal/logic/grab_token/grab_token_test.go` 测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 FetchTokenFromAPI() 编写单元测试 | Context: 测试场景包括：1) OAuth API 调用成功，2) 网络错误，3) 响应解析错误，4) 缺少必需字段 | Restrictions: 使用 httptest 模拟 HTTP 服务器，覆盖率 ≥ 80% | Success: 测试覆盖率达标，所有场景测试通过

- [x] 3.2 编写 GetAccessToken() 单元测试（集成测试待补充）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token_test.go`
  - Purpose: 测试 Redis 缓存和双重检查锁逻辑
  - Requirements: 2.1-2.5, 3.1-3.7
  - Leverage: 参考 Grab 相关测试
  - Prompt: Role: QA Engineer specializing in concurrency testing | Task: 为 GetAccessToken() 编写单元测试 | Context: 测试场景包括：1) Redis 缓存命中，2) Redis 缓存未命中，3) Redis 故障降级，4) 并发安全（100 个并发请求只调用 1 次 OAuth API）| Restrictions: 使用 mock Redis，使用 sync.WaitGroup 测试并发 | Success: 测试覆盖率达标，并发测试通过

- [x] 3.3 编写 GetAuthorizationHeader() 单元测试（集成测试待补充）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token_test.go`
  - Purpose: 测试 Authorization Header 生成逻辑
  - Requirements: 4.1-4.4
  - Leverage: 简单测试，验证返回格式
  - Prompt: Role: QA Engineer | Task: 为 GetAuthorizationHeader() 编写单元测试 | Context: 测试场景：1) 正常返回 Bearer Token，2) Token 获取失败时错误传递 | Restrictions: Mock GetAccessToken() 方法 | Success: 测试通过，返回格式正确

### 集成测试（可选）

- [ ] 3.4 编写集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/lineman_oauth_test.go` (新建，可选)
  - Purpose: 测试与真实 LINE MAN OAuth API 的集成
  - Requirements: 所有功能需求
  - Leverage: 使用 staging 环境配置
  - Prompt: Role: QA Integration Test Engineer | Task: 实现 LINE MAN OAuth 集成测试 | Context: 使用 staging 环境的 client_id 和 client_secret，调用真实 OAuth API，验证 Token 有效性，验证缓存过期和刷新 | Restrictions: 仅在 staging 环境运行，不影响生产 | Success: 集成测试通过，Token 获取成功

### 文档更新

- [x] 3.5 更新代码注释

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman_token/lineman_token.go`
  - Purpose: 确保所有导出方法包含中文注释
  - Requirements: 所有功能需求
  - Leverage: 现有代码注释风格
  - Prompt: Role: Technical Writer | Task: 为新增的三个方法添加完整的中文注释 | Context: 注释格式参考 Grab 实现，包括方法说明、参数说明、返回值说明 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 所有方法包含完整中文注释

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 所有测试通过 `go test ./...`
- [ ] 无 Linter 错误

### 功能完整性

- [ ] requirements.md 中的所有需求已满足（Requirement 1-5）
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成：
  - ✅ OAuth Token 获取成功
  - ✅ Redis 缓存正常工作
  - ✅ 双重检查锁避免并发重复请求
  - ✅ Authorization Header 格式正确
  - ✅ 配置支持完整

### 文档同步

- [ ] 代码注释完整（中文）
- [ ] design.md 已更新（标记 Phase 2-3 完成）
- [ ] tasks.md 已更新（所有任务标记为 [x]）
- [ ] CHANGELOG.md 已更新（如有版本发布）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/structs.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`（敏感信息不记录）
- [ ] 遵循 `.cursor/rules/version.mdc`（Git 提交规范）

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-access-token/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-access-token/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-lineman-access-token/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-access-token/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-access-token/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（按 Phase 顺序）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **更新进度**: 更新"进度总览"中的完成率
10. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：核心 Prompt 模板

### Go BMP Logic 层开发

```
Role: Go Developer specializing in GoFrame and concurrency control

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径（直接复制或参考）}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 使用 g.Redis() 访问 Redis
- 使用 g.Client() 发送 HTTP 请求
- 使用 g.Log() 记录日志（中文）
- 使用 gerror 包处理错误
- 所有导出方法包含中文注释
- 敏感信息（client_secret）不记录到日志

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 遵循 GoFrame 规范
```

### Go BMP 单元测试

```
Role: QA Engineer with Go and GoFrame testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80%

Test Cases Required:
- 正常场景测试
- 异常场景测试（网络错误、解析错误等）
- 边界条件测试（空值、特殊字符等）
- 并发场景测试（如适用）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 使用 httptest 模拟 HTTP 服务器
- 使用 mock 模拟 Redis
- 必须包含并发安全测试

Success Criteria:
- 测试覆盖率 ≥ 80%
- 所有测试通过
- 边界情况已覆盖
- 并发测试通过
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充 - 实现完成后记录经验]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`
- 在执行任务过程中，若遇到关键技术决策或踩坑经验，请记录 Episode：
  - 双重检查锁的实现细节
  - Redis 缓存失败降级策略
  - 并发场景下的性能优化

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-07  
**最后更新**: 2026-01-07  
**维护者**: rikugun

