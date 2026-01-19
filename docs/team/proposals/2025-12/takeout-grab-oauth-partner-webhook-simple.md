# Takeout Grab OAuth Partner Webhook 简单 Token 实现 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **提案人**   | rikugun                  |
| **日期**     | 2025-12-05              |
| **目标版本** | v2.10.x                 |
| **状态**     | 待评审                   |
| **关联任务** | -                        |
| **关联 Spec** | [task-takeout-grab-oauth-partner-webhook-simple](../../shared/specs/archived/v2.12/task-takeout-grab-oauth-partner-webhook-simple/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

目前 `ttpos-takeout` 模块已经在推进 GrabFood 对接工作（参考 `takeout-grab-integration` 提案），并补充了通用的 OAuth Token 获取方案（参考 `grab-oauth-token` 提案）。  
但在实际对接 GrabFood Partner API 时，Grab 官方推荐使用 **Get partner OAuth access token webhook** 方式，由 Grab 主动调用合作方的 Webhook 来获取访问下游接口所需的 Token（见官方文档 `post-oauth-partner-webhook` 接口说明）。  

当前项目中尚未提供一个符合 GrabFood 官方规范的、**专用于 `get-oauth-partner-webhook` 的简化实现**，同时也没有约定统一的 `grab.partner.xx` 配置结构，导致：

- Grab 无法通过标准化 Webhook 向本系统“索取”访问令牌；
- `client_id`/`client_secret` 等凭据分散在现有配置结构中，不便于后续多 Partner 扩展与环境切换；
- 对于只需要一个“简单 Token 提供方实现”的场景（例如前期联调 / PoC），成本相对偏高。

### 业务价值

1. **快速满足 Grab 联调要求**：提供最小可用实现，支持 Grab 通过 Webhook 拉取访问 Token，打通授权链路。
2. **统一配置入口**：通过 `config.yaml` 中的 `grab.partner.xx` 结构，统一管理不同 Partner / 环境下的凭据，便于后续扩展。
3. **降低实现复杂度**：将该能力封装为独立、易于复用的 Auth 能力，为后续完整 OAuth 方案（Redis 缓存、失败重试等）打基础。

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 后端集成工程师 / 运维

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-takeout` 模块中，围绕 Grab 官方文档中的 **Get partner OAuth access token webhook** 接口，提供一个“轻量级”的实现：

1. 在 `manifest/config/config.yaml`（以及 `config.tpl.yaml`）中补充 `grab.partner.xx` 配置段，集中管理 `client_id`、`client_secret`、环境等信息。
2. 在 `internal/logic/grab` 下新增简单的 Token Service，从配置中读取对应 Partner 的凭据，并向 Grab OAuth 端点发起请求获取 `access_token`（可先不引入复杂缓存，仅预留扩展点）。
3. 在 `internal/controller/grab` 下实现 Grab 调用的 `post-oauth-partner-webhook` HTTP 接口，对请求进行基础校验后，调用上述 Token Service 并按 Grab 要求的响应结构返回 Token。

该实现聚焦于“让 Grab 能够通过文档规范的 Webhook 成功拿到 Token”，优先保证可用性与清晰的配置结构，后续更复杂的缓存与容错能力可以通过已有 `grab-oauth-token` Spec 进一步演进。

> 参考文档：[GrabFood API - Get partner OAuth access token webhook](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-oauth-partner-webhook/operation/post-oauth-partner-webhook)

### 核心功能点

1. **配置结构设计 (`grab.partner.xx`)**
   - 在 `config.yaml` 中增加 `grab.partner.{code}` 段（如 `grab.partner.default`），包含 `client_id`、`client_secret`、环境、超时等字段。
   - 支持通过环境变量注入敏感信息，与现有配置管理规范保持一致。
2. **Token Service 简单实现**
   - 在 `internal/logic/grab` 中封装 `GetPartnerToken(ctx, partnerCode string)` 方法，从配置读取凭据，调用 Grab OAuth 接口获取 `access_token` 和 `expires_in`。
   - 当前版本可以只做“每次请求即实时向 Grab 取 Token”的简单实现，并在代码中预留后续引入 Redis 缓存的扩展点。
3. **Webhook 控制器实现**
   - 在 `internal/controller/grab` 下实现 `post-oauth-partner-webhook` 接口，解析请求体中的 Partner 标识 / Client 信息，映射到对应的 `grab.partner.xx` 配置。
   - 基于 Token Service 返回结果，按 Grab 文档要求组装响应 JSON（包含 `accessToken`、`expiresIn` 等字段），并处理基础错误响应。

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [x] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要后端实现 Webhook 接口与外部 OAuth 调用，逻辑相对清晰
- [ ] **高**：涉及架构调整、复杂算法

### 工作量预估

- **预计天数**: 1-2 天
- **预估 SP**: 2-3（待技术评审确认）

### 风险识别

**潜在风险**：
1. Grab 文档对 `post-oauth-partner-webhook` 请求 / 响应字段的理解偏差，导致联调失败。
2. `grab.partner.xx` 配置设计与后续完整 OAuth/多平台方案不一致，需要二次调整。

**缓解措施**：
1. 在开发前对照 Grab 官方示例请求/响应进行详细梳理，并与 `grab-oauth-token` 相关 Spec 对齐字段含义。
2. 配置结构在设计时预留冗余字段（如 `scope`、`env`），并在文档中标记为“可扩展”，降低未来变更成本。

---

## 🔗 相关资源

### 参考需求

- `docs/team/proposals/2025-12/takeout-grab-integration.md`
- `docs/team/proposals/2025-12/grab-oauth-token.md`
- [GrabFood API - Get partner OAuth access token webhook](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-oauth-partner-webhook/operation/post-oauth-partner-webhook)

### 相关文档

- `ttpos-bmp/app/ttpos-takeout/manifest/config/config.yaml`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/auth.go`
- `.cursor/rules/go-bmp.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |
| 测试代表     |        |           |
| UI/UX 设计师 |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`task-takeout-grab-oauth-partner-webhook-simple`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** Grab 对接工程师  
**我想** 有一个符合 GrabFood 官方 `get-oauth-partner-webhook` 规范的简单 Token 提供方实现  
**以便于** 在最小改动的前提下完成与 Grab 的 Token 授权链路联调。

### AC 验收标准（初稿）

1. **WHEN** Grab 调用 `post-oauth-partner-webhook` Webhook **THEN** 系统能够根据 `grab.partner.xx` 配置成功返回 `accessToken` 与 `expiresIn` 字段。
2. **IF** 配置缺失或凭据错误 **THEN** Webhook 接口返回符合 Grab 要求的错误码与错误信息，便于排查。

### 线框图/原型（可选）

[当前为纯后端能力，无需 UI 原型]

---

## 📄 模板使用说明

> 本提案基于 `docs/agent/templates/proposal-template.md` 模板生成，后续如有结构调整请同步更新模板与已有提案。


