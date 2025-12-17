# Takeout Grab OAuth Partner Webhook 简单 Token 实现 需求文档

> 本文档定义 Grab 调用 `ttpos-takeout` 获取 Partner Token（Get partner OAuth access token webhook）的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                          |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-grab-oauth-partner-webhook-simple.md](../../../../team/proposals/2025-12/takeout-grab-oauth-partner-webhook-simple.md) |
| **创建日期**      | 2025-12-05                                                                                                                                   |
| **负责人**        | 待定                                                                                                                                        |
| **目标 Sprint**   | Sprint 待定                                                                                                                                  |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                                    |

## 📋 审核状态

| 项目         | 内容           |
| ------------ | -------------- |
| **审核状态** | 已通过         |
| **审核人**   | rikugun        |
| **审核日期** | 2025-12-08     |
| **审核意见** | 需求明确，可进入设计阶段 |

---

## 📋 概述

GrabFood 官方推荐通过 **Get partner OAuth access token webhook** 方式，由 Grab 主动调用合作方暴露的 Webhook 来获取访问后续 Partner API 所需的 Token。  
本需求在 `ttpos-bmp/app/ttpos-takeout` 模块中提供一个符合 GrabFood 文档规范的简易实现：  
当 Grab 调用我们的 Webhook 时，我们基于 `config.yaml` 中的 `grab.partner.xx` 配置校验 Partner 信息并生成短期有效的访问 Token，返回给 Grab 使用。

## 🎯 产品对齐

- 支持现有 `story-takeout-grab-integration` 的整体对接目标，补齐 Token 获取链路。
- 为后续更复杂的 OAuth / 多 Partner 扩展（如不同品牌、不同环境）打下配置和接口基础。
- 降低初期联调成本，先提供一个**简单、稳定、易排查**的 Token 提供方实现。

## 📝 用户故事

**作为** Grab 对接工程师 / 后端集成工程师  
**我想** 通过 Grab 官方的 `post-oauth-partner-webhook` 接口，由 Grab 主动向 `ttpos-takeout` 请求 Partner Token  
**以便于** Grab 能安全地带着该 Token 访问我们为 Grab 暴露的订单、菜单等接口，不需要在多处手动配置和管理访问凭据。

---

## 功能需求

### Requirement 1: 提供符合 Grab 文档的 Partner Token Webhook 接口

**用户故事**: 作为 Grab 平台，我想通过固定的 Webhook URL 获取 Partner Token，以便后续所有调用都复用这一授权链路。

#### 验收标准

1. **WHEN** Grab 按文档向指定 URL（如 `/grab/oauth/partner/webhook`，最终路径由设计阶段确定）发起 `POST` 请求 **THEN** 系统能够成功接收请求并记录必要日志（不包含敏感字段）。
2. **IF** 请求结构符合 Grab `post-oauth-partner-webhook` 文档要求 **THEN** 系统能够正确解析请求体中的 Partner / Client 信息。
3. **WHEN** Webhook 处理成功 **THEN** 响应结构中的字段名、类型和语义需与 Grab 官方文档保持一致（如 `accessToken`、`expiresIn` 等）。

#### 具体要求

- [ ] 1.1 按 Grab 文档定义 HTTP 方法为 `POST`，Content-Type 支持 `application/json`（如官方要求有差异，以文档为准）。
- [ ] 1.2 接口路径前缀建议统一归类在 Grab 集成下（如 `/grab/oauth/partner/webhook` 或 `/api/grab/oauth/partner/token`），避免与其他第三方混淆。
- [ ] 1.3 请求和响应示例需在设计阶段补充到接口文档中，便于与 Grab 联调。

---

### Requirement 2: 支持从配置中心读取 grab.partner.xx 凭据（含 scope）

**用户故事**: 作为运维/后端工程师，我想在 `config.yaml` 中集中配置 Grab Partner 的 `client_id` / `client_secret` 等信息，以便统一管理和按环境切换。

#### 验收标准

1. **WHEN** 系统启动 **THEN** 可以从 `manifest/config/config.yaml`（以及 `config.tpl.yaml` 模板）中读取 `grab.partner.xx` 配置且结构清晰可读，其中 `scope` 字段如未显式配置则默认使用 `food.partner_api`。
2. **IF** Webhook 请求中包含 Partner 标识（如 `client_id` 或自定义 partner code） **THEN** 系统能够根据该标识映射到对应的 `grab.partner.xx` 配置。
3. **IF** 对应 Partner 配置不存在或字段不完整 **THEN** Webhook 应返回明确的错误信息，且日志中包含足够的排查信息（但不泄露密钥）。

#### 具体要求

- [ ] 2.1 在 `config.yaml` 中新增 `grab.partner.{code}` 段，其中至少包括 `client_id`、`client_secret`、`scope`、`environment`、`timeout`，字段命名遵循项目现有配置风格；`scope` 未配置时，逻辑层默认使用 `food.partner_api`。
- [ ] 2.2 所有敏感字段（如 `client_secret`）必须支持通过环境变量注入，不允许硬编码。
- [ ] 2.3 在 `internal/model/conf` 下补充或扩展配置 struct，使其能承载 `grab.partner.xx` 的结构，并由 `g.Cfg()` 正确解析。

---

### Requirement 3: 简单 Token 生成与生命周期管理

**用户故事**: 作为 Grab 平台，我想拿到一个在短期内有效的 Partner Token，用于访问后续接口；作为后端工程师，我希望 Token 逻辑足够简单、易于扩展。

#### 验收标准

1. **WHEN** Webhook 收到合法请求并成功匹配到 Partner 配置 **THEN** 系统生成一个访问 Token 并返回给 Grab，`expiresIn` 字段为秒级整型。
2. **IF** 同一 Partner 在短时间内多次请求 Token **THEN** 当前阶段允许每次重新生成 Token（无需实现复杂缓存），但需保证生成算法是确定性的且易于后续替换为更安全的实现。
3. **WHEN** Token 生成失败（如签名算法错误、依赖异常） **THEN** Webhook 返回合适的错误码和描述，日志记录详细错误栈。

#### 具体要求

- [ ] 3.1 当前阶段允许采用“简单签名 Token”方案（例如基于 `client_id` + 时间戳 + 服务端 secret 生成 HMAC 字符串），不强制引入完整 JWT 机制，但实现需封装在独立 Service 中。
- [ ] 3.2 `expiresIn` 默认建议 300–900 秒（具体值在设计阶段确定），并在代码中集中配置，避免散落魔法数字。
- [ ] 3.3 为后续引入 Redis 缓存（避免频繁生成 Token）预留扩展点，例如统一的 `TokenService` 接口。

---

### Requirement 4: 错误处理与日志

**用户故事**: 作为运维人员，我希望在联调或线上故障时能快速定位问题；作为 Grab 平台，我希望在出错时得到明确的错误信息。

#### 验收标准

1. **WHEN** 请求参数缺失或格式错误 **THEN** Webhook 应返回合理的 HTTP 状态码（如 400 系列）以及结构化错误信息。
2. **WHEN** 内部处理异常（如配置读取失败、Token 生成异常） **THEN** 应返回 5xx 或约定的错误状态，并在日志中记录完整错误上下文。
3. **IF** 涉及敏感数据（client_secret / 内部密钥） **THEN** 在日志和错误响应中不得明文输出。

#### 具体要求

- [ ] 4.1 使用统一的日志组件（GoFrame `g.Log()`），打上模块标识（如 `grab-oauth-webhook`）便于筛选。
- [ ] 4.2 对常见错误场景（配置缺失、Partner 未匹配、Token 生成失败）设计清晰的错误码和错误信息文本。
- [ ] 4.3 在联调阶段提供简要的故障排查步骤文档（可在 `design.md` 或 troubleshooting 文档中补充）。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository / Config 分层，Webhook 控制器不直接操作配置和加密细节。
- **单一职责原则**: Webhook Controller 只负责参数解析和调用 Service；Token 生成逻辑封装在独立 Service。
- **模块化设计**: Grab 相关逻辑继续集中在 `internal/logic/grab` 与 `internal/controller/grab` 下，避免跨模块耦合。
- **依赖管理**: Service 之间通过接口依赖，遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`。

### API 设计要求

- [ ] URL 使用 snake_case 或项目既有风格命名（例如 `/grab/oauth/partner_webhook` 或 `/grab/oauth/partner/webhook`，以设计阶段决定为准）。
- [ ] 如果该接口需要纳入统一 API 网关层，应保持响应格式可与网关兼容。
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范。

### 数据库设计要求

- 本需求不强制引入新表，如后续设计需要持久化 Partner Token 或审计日志，应遵循 `.cursor/rules/database.mdc` 规范单独评审。

### 性能要求

- [ ] 在正常调用频率下，Webhook 端到端响应时间（不依赖外部慢服务的前提下）应控制在 200ms 内。
- [ ] Token 生成算法应为 O(1) 复杂度，不引入重型加密库导致明显性能下降。

### 测试要求

- [ ] 对 Token Service 实现单元测试覆盖 Token 生成和异常场景。
- [ ] 针对 Webhook Controller 的集成测试覆盖：成功返回、配置缺失、Partner 未匹配、内部异常等主要分支。

### 安全要求

- [ ] 如需对 Webhook 做来源验证（例如 IP 白名单或额外 Header 校验），在设计阶段明确并实现。
- [ ] 敏感信息只保存在配置和内存中，不写入普通业务日志。
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范。

### 可靠性要求

- [ ] 内部依赖（配置中心、时间服务等）异常时应优雅失败，不导致服务整体崩溃。
- [ ] 对于频繁失败的请求场景，日志应支持快速聚合分析（统一错误码）。

---

## 验收标准

### 功能验收

1. **Webhook 联调通过**: 按 Grab 官方文档配置后，Grab 能通过 `post-oauth-partner-webhook` 成功获取 Token，并能用该 Token 访问至少一个已实现的 Grab Partner API 接口。
2. **配置生效验证**: 修改 `grab.partner.xx` 配置后，重启或按规范刷新配置后，Webhook 行为与配置一致（例如更换 client_id 后旧值不再可用）。
3. **错误路径验证**: 针对配置缺失、Partner 未匹配、内部异常等场景，均能返回预期错误响应并在日志中留痕。

### 测试验收

1. **单元测试**: Token 相关逻辑单元测试通过，关键分支均覆盖。
2. **集成测试**: 通过模拟 Grab 请求（本地或测试环境）完成端到端验证。

### 文档验收

1. **技术文档**: 在 `design.md` 中补充接口设计、配置结构说明和主要时序图。
2. **联调说明**: 提供给 Grab 的对接文档中包含 Webhook URL、请求/响应示例及错误码说明。

---

## 约束条件

### 技术约束（Go BMP 模块）

- 必须使用 GoFrame 2.x。
- 禁止修改 `dao/entity/do/` 目录（自动生成）。
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`。

### 业务约束

- 本需求仅解决“Grab 调用我们获取 Partner Token”的最小实现，不包含对 Token 的长期存储、跨服务共享和完整 OAuth 授权管理，这些将由后续任务（如 `task-takeout-grab-oauth-token`）扩展。

### 资源约束

- 开发时间: 预估 1–2 天（与 `story-takeout-grab-integration` 同期排期）。

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab` 现有 Grab 业务逻辑。
- `ttpos-bmp/app/ttpos-takeout/manifest/config/config.yaml` 配置管理。

### 业务依赖

- 上游需求：`story-takeout-grab-integration`。
- 相关任务：`task-takeout-grab-oauth-token`（通用 OAuth Token 获取方案）。

---

## 风险和缓解

### 风险 1: 对 Grab 文档理解偏差

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在设计阶段对照官方示例请求/响应逐字段校对。
- 优先在 SandBox / Staging 环境联调，记录差异。

### 风险 2: 后续 Token 策略调整导致接口行为变化

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 将 Token 生成、TTL 策略集中封装在 Service 中，对外接口保持尽量稳定。
- 在文档中明确当前版本的行为和后续演进方向。

---

## 时间表

- **Phase 1 - 需求确认**: 0.5 天
- **Phase 2 - 设计与开发**: 1–1.5 天
- **Phase 3 - 联调与验收**: 0.5 天
- **总计**: 2–2.5 天（SP ≈ 2–3）

---

## 参考资料

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范  
- `.cursor/rules/api.mdc` - API 设计规范  
- `.cursor/rules/security.mdc` - 安全开发规范  
- `docs/team/proposals/2025-12/takeout-grab-oauth-partner-webhook-simple.md`  
- `docs/shared/specs/active/story-takeout-grab-integration/requirements.md`  
- Grab 官方文档: [GrabFood API - Get partner OAuth access token webhook](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-oauth-partner-webhook/operation/post-oauth-partner-webhook)


