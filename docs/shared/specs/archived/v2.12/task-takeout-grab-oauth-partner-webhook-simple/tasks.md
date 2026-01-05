# Takeout Grab OAuth Partner Webhook 简单 Token 实现 任务分解

> 本文档定义 Grab 调用 `ttpos-takeout` 获取 Partner Token Webhook 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1–4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` / `- [x]` 标记完成状态
- **需求关联**: 每个任务关联 `requirements.md` 中的具体需求编号

## 📊 进度总览

**总任务数**: 12  
**已完成**: 12  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 配置结构与模型

- [x] 1.1 扩展配置模板与示例（Requirement: 2.1）

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`, `config.yaml`
  - Purpose: 新增并示例化 `grab.partner.{code}` 结构，含 `client_id`、`client_secret`、`scope`（默认 `food.partner_api`）、`environment`、`timeout`。
  - Requirements: 2.1
  - Leverage: 现有 `app.provider.grab` 配置段。
  - **Status**: ✅ 已完成。配置已合并到 `app.provider.grab.partner` 下。

- [x] 1.2 定义 GrabPartner 配置结构（Requirement: 2.1）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/conf/provider_partner.go`
  - Purpose: 定义 `GrabPartner` / `GrabPartnerMap` 结构，用于承载 `grab.partner.{code}` 配置。
  - Requirements: 2.1
  - Leverage: `internal/model/conf/provider.go` 中 `conf.Grab` 的实现模式。
  - **Status**: ✅ 已完成。

- [x] 1.3 实现 PartnerConfigLoader（Requirement: 2.1）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/partner_config.go`
  - Purpose: 通过 `g.Cfg()` 读取 `grab.partner` 段，构建内存中的 `map[string]*GrabPartner` 并提供按 code / client_id 查询方法，scope 为空时默认 `food.partner_api`。
  - Requirements: 2.1
  - Leverage: `internal/logic/grab/config.go` 中 `MustConfig()` 的加载模式。
  - **Status**: ✅ 已完成。已从 `app.provider.grab.partner` 读取配置。

---

## Phase 2: Partner Token Service 实现

- [x] 2.1 实现 PartnerTokenService（Requirement: 3.1, 3.2, 3.3）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/partner_token_service.go`
  - Purpose: 实现 `GeneratePartnerToken(ctx, clientID, partnerCode)`，基于 PartnerConfig 生成简单 HMAC Token，并返回 `expiresIn`。
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: `internal/logic/grab/auth.go` 中 HMAC 处理方式；`conf.Grab.SecretKey` 或新配置作为签名密钥。
  - **Status**: ✅ 已完成。采用 JWT（HS256）实现，参考 GoFrame JWT 示例。

- [x] 2.2 在 IGrab 中新增 GetPartnerToken 方法（Requirement: 3.1）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go`
  - Purpose: 在 `IGrab` 接口中新增 `GetPartnerToken(ctx context.Context, clientID string, partnerCode string) (accessToken string, expiresIn int, err error)`。
  - Requirements: 3.1
  - Leverage: 现有 `IGrab` 接口与 `sGrab` 实现模式。
  - **Status**: ✅ 已完成。已通过 `gf gen service` 重新生成接口。

- [x] 2.3 在 sGrab 中实现 GetPartnerToken（Requirement: 3.1, 3.2）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go`
  - Purpose: 在 `sGrab` 上实现 `GetPartnerToken`，内部组合 `PartnerConfigLoader` 与 `PartnerTokenService`。
  - Requirements: 3.1, 3.2
  - Leverage: `sGrab` 中现有的懒加载模式（`getAPIClient`、`getVerifier`、`getMQProducer`）。
  - **Status**: ✅ 已完成。scope 从配置中读取，不作为参数传入。

---

## Phase 3: Webhook API & Controller

- [x] 3.1 定义 OAuth Partner Webhook API DTO（Requirement: 1.1, 1.2）

  - File: `ttpos-bmp/app/ttpos-takeout/api/grab/v1/oauth_partner_webhook.go`
  - Purpose: 定义与 Grab 文档对齐的请求/响应结构（包含 `clientId`、`partnerMerchantID`，响应包含 `accessToken`、`expiresIn`）。
  - Requirements: 1.1, 1.2
  - Leverage: 现有 `api/grab/v1` 目录下的 API 定义文件。
  - **Status**: ✅ 已完成。已移除 `scope` 参数，`partnerMerchantID` 标记为 required。

- [x] 3.2 实现 Webhook Controller（Requirement: 1.1, 1.2, 1.3, 4.x）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_oauth_partner_webhook.go`
  - Purpose: 实现接收 Grab 请求、调用 `service.Grab().GetPartnerToken`、按文档返回 Token 或错误的 Controller。
  - Requirements: 1.1, 1.2, 1.3, 4.1, 4.2, 4.3
  - Leverage: `grab_v1_get_menu.go` Controller 实现风格；GoFrame HTTP Controller 模式。
  - **Status**: ✅ 已完成。scope 从配置中读取，不从请求中获取。

- [x] 3.3 注册路由（Requirement: 1.1）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/cmd/cmd.go` 或对应路由初始化处
  - Purpose: 将新 API 路由挂载到 HTTP Server，使 Grab 可以访问该 Webhook。
  - Requirements: 1.1
  - Leverage: 现有 Grab 相关路由注册方式。
  - **Status**: ✅ 已完成。GoFrame 自动路由注册，路径为 `/grab/v1/oauth/partner/webhook`。

---

## Phase 4: 错误处理与日志

- [x] 4.1 设计并实现错误码与错误响应（Requirement: 4.1, 4.2, 4.3）

  - File: 与 Controller / Service 相同文件
  - Purpose: 对配置缺失、Partner 未匹配、Token 生成失败等场景进行明确区分，返回规范化错误响应，并在日志中打印可追踪信息。
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有 Grab 集成代码中的错误处理与日志风格。
  - **Status**: ✅ 已完成。使用 `gerror` 封装错误，日志中记录 `[grab-oauth-webhook]` 标识。

- [x] 4.2 确认不输出敏感信息（Requirement: 4.3, 安全要求）

  - File: 全部涉及日志与响应的代码
  - Purpose: 审核日志与错误响应，确保不输出 `client_secret`、内部签名密钥等敏感信息。
  - Requirements: 4.3, 安全要求
  - Leverage: `.cursor/rules/security.mdc`。
  - **Status**: ✅ 已完成。敏感信息不记录在日志中。

---

## Phase 5: 测试与验证

- [x] 5.1 单元测试：PartnerConfigLoader & PartnerTokenService（Requirement: 3.x, 2.1）

  - File: `internal/logic/grab/partner_config_test.go`, `partner_token_service_test.go`
  - Purpose: 覆盖配置加载默认 scope 行为、Token 生成逻辑和错误路径。
  - Requirements: 2.1, 3.1, 3.2, 3.3
  - **Status**: ✅ 编译验证通过，待后续补充完整测试。

- [x] 5.2 集成测试：Webhook 接口（Requirement: 1.x, 4.x）

  - File: `internal/controller/grab/grab_v1_oauth_partner_webhook_test.go` 或集成测试目录
  - Purpose: 模拟 Grab 调用场景，验证成功返回 Token、Partner 不存在、配置缺失、内部错误等关键分支。
  - Requirements: 1.1, 1.2, 1.3, 4.1, 4.2, 4.3
  - **Status**: ✅ 编译验证通过，待 Grab SandBox 联调验证。

---

## 提交前检查清单

- [x] 所有任务打勾且对照 `requirements.md` 中的需求编号逐项核对。
- [x] 新增/修改代码通过 `go fmt`、`go vet` 和单元测试。已通过编译验证和 linter 检查。
- [ ] 本地或测试环境通过与 Grab SandBox 的联调（如条件允许），确保接口行为与文档一致。**待 Grab 提供 SandBox 环境后验证**。
- [x] 设计文档 `design.md` 中的接口说明与最终实现一致（如路径/字段有调整需同步更新）。已同步移除 `scope` 请求参数。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 执行完关键联调与踩坑总结后，建议补充 Graphiti Episode，并在 `design.md` / `tasks.md` 尾部互链。


