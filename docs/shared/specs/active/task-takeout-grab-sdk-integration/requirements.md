# Takeout 模块集成 GrabFood 官方 SDK 需求文档

> 本文档定义 Grab 官方 SDK 集成的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-grab-sdk-integration.md](../../../../team/proposals/2025-12/takeout-grab-sdk-integration.md) |
| **创建日期**      | 2025-12-08                                                                                                               |
| **负责人**        | rikugun                                                                                                                  |
| **目标 Sprint**   | Sprint TBD                                                                                                               |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                               |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已通过     |
| **审核人**   | rikugun    |
| **审核日期** | 2025-12-08 |

---

## 📋 概述

本任务旨在将 `ttpos-bmp/app/ttpos-takeout` 模块中现有的 **Grab API 自实现方案**迁移到 **Grab 官方 SDK** (`github.com/grab/grabfood-api-sdk-go`)，以降低维护成本、提升开发效率、增强功能覆盖。

迁移范围涵盖两个方向的 API：
- **Outbound（调用 Grab API）**：使用 SDK Client 替换 `client.go` 中的自实现
- **Inbound（被 Grab 调用的 Webhook）**：使用 SDK Model 替换 `dto/grab/` 下的自定义结构

## 🎯 产品对齐

该任务支持 `story-takeout-grab-integration` 的技术优化，为后续接入 Campaign、DineIn 等高级功能奠定基础。

## 📝 用户故事

**作为** 后端开发工程师  
**我想** 使用 Grab 官方 SDK 进行 API 集成  
**以便于** 减少维护成本，快速接入新功能，提高代码可靠性

---

## 功能需求

### Requirement 1: SDK 引入与 Wrapper 层搭建 (Phase 1)

**用户故事**: 作为 开发者，我想 引入官方 SDK 并搭建统一的 Wrapper 层，以便于 后续迁移有统一入口。

#### 验收标准

1. **WHEN** 执行 `go get` **THEN** 系统 **SHALL** 成功引入 `github.com/grab/grabfood-api-sdk-go` 依赖。
2. **IF** SDK 版本与 Go 1.23 兼容 **THEN** 项目 **SHALL** 正常编译无冲突。
3. **WHEN** 调用 Grab API **THEN** 开发者 **SHALL** 通过 Wrapper 层统一访问。

#### 具体要求

- [ ] 1.1 添加 `github.com/grab/grabfood-api-sdk-go` 依赖到 `go.mod`。
- [ ] 1.2 验证 SDK 与 Go 1.23、GoFrame 2.x 的兼容性。
- [ ] 1.3 创建 SDK Wrapper 层 (`internal/logic/grab/sdk_client.go`)，封装配置管理和环境切换。
- [ ] 1.4 保持现有 Redis Token 缓存策略，集成到 Wrapper 层。

---

### Requirement 2: Outbound 功能迁移 (Phase 2)

**用户故事**: 作为 系统，我想 通过官方 SDK 调用 Grab API，以便于 获得官方维护的稳定性和新功能支持。

#### 验收标准

1. **WHEN** 需要获取 OAuth Token **THEN** 系统 **SHALL** 通过 SDK 的 `GetOauthGrabAPI` 获取。
2. **WHEN** 需要操作订单（接受/拒绝/取消/标记就绪）**THEN** 系统 **SHALL** 通过 SDK 对应 API 执行。
3. **WHEN** 需要管理门店（暂停/恢复/获取状态）**THEN** 系统 **SHALL** 通过 SDK 对应 API 执行。
4. **IF** 迁移完成 **THEN** 现有业务功能 **SHALL** 保持一致性。

#### 具体要求

- [ ] 2.1 迁移 OAuth Token 获取到 SDK 实现，保留 Redis 缓存。
- [ ] 2.2 迁移订单操作 API (`AcceptRejectOrderAPI`, `CancelOrderAPI`, `MarkOrderReadyAPI`, `UpdateDeliveryStateAPI`)。
- [ ] 2.3 迁移门店管理 API (`PauseStoreAPI`, `GetStoreStatusAPI`, `GetStoreHourAPI`)。
- [ ] 2.4 迁移菜单通知 API (`UpdateMenuNotificationAPI`)。
- [ ] 2.5 删除或废弃 `client.go` 中的旧实现代码。

---

### Requirement 3: Inbound DTO 替换 (Phase 3)

**用户故事**: 作为 开发者，我想 使用 SDK 提供的 Webhook Model，以便于 减少 DTO 同步风险和重复定义。

#### 验收标准

1. **WHEN** 收到订单提交 Webhook **THEN** 系统 **SHALL** 使用 SDK 的 `SubmitOrderRequest` 解析请求。
2. **WHEN** 收到订单状态变更 Webhook **THEN** 系统 **SHALL** 使用 SDK 的 `OrderStateRequest` 解析请求。
3. **WHEN** 收到菜单拉取请求 **THEN** 系统 **SHALL** 使用 SDK 的 `GetMenuNewResponse` 构造响应。
4. **IF** 迁移完成 **THEN** Webhook 处理逻辑 **SHALL** 保持功能一致。

#### 具体要求

- [ ] 3.1 使用 SDK `SubmitOrderRequest` 替换 `dto/grab/order.go` 中的自定义结构。
- [ ] 3.2 使用 SDK `OrderStateRequest` 替换订单状态变更 DTO。
- [ ] 3.3 使用 SDK `GetMenuNewResponse` 替换菜单响应 DTO。
- [ ] 3.4 使用 SDK `MenuSyncWebhookRequest` 替换菜单同步 DTO。
- [ ] 3.5 使用 SDK `PartnerOauthRequest/Response` 替换 Partner OAuth DTO。
- [ ] 3.6 **保留** `auth.go` 中的 HMAC-SHA256 签名验证逻辑（SDK 不提供）。
- [ ] 3.7 清理或归档 `dto/grab/` 下的废弃文件。

---

### Requirement 4: 新功能启用 (Phase 4 - 可选)

**用户故事**: 作为 产品经理，我想 快速接入 SDK 提供的新功能，以便于 扩展业务能力。

#### 验收标准

1. **WHEN** 需要接入 Campaign（营销活动）**THEN** 开发者 **SHALL** 直接使用 SDK 的 `CreateCampaignAPI` 等。
2. **WHEN** 需要菜单批量更新 **THEN** 开发者 **SHALL** 直接使用 SDK 的 `BatchUpdateMenu`。

#### 具体要求

- [ ] 4.1 评估 Campaign API 接入的业务价值和工作量。
- [ ] 4.2 评估 DineIn API 接入的业务价值和工作量。
- [ ] 4.3 评估 Menu 批量更新 API 的接入需求。
- [ ] 4.4 根据评估结果决定是否在本期实现。

---

## 非功能需求

### 代码架构和模块化

- **Wrapper 层隔离**：SDK 调用统一通过 Wrapper 层，业务代码不直接依赖 SDK。
- **渐进式迁移**：新旧实现可并存，逐步切换，降低风险。
- **配置统一**：环境切换（Staging/Production）通过 Wrapper 层统一管理。

### 数据库设计要求

- **无新增表**：本任务为纯代码重构，不涉及数据库变更。

### 性能要求

- [ ] SDK 调用性能不低于原自实现方案。
- [ ] Token 缓存策略保持不变，避免频繁请求 OAuth。

### 安全要求

- [ ] 必须保留 Webhook 签名验证（SDK 不提供此功能）。
- [ ] Client Secret 等敏感配置通过环境变量注入。

---

## 验收标准

### 功能验收

1. **Outbound 功能**：所有现有 Grab API 调用通过 SDK 执行，功能保持一致。
2. **Inbound 功能**：所有 Webhook 处理使用 SDK Model，功能保持一致。
3. **签名验证**：Webhook 签名验证功能正常。

### 测试验收

1. **单元测试**：覆盖 SDK Wrapper 层的核心逻辑。
2. **集成测试**：模拟完整 API 调用流程，验证与 Grab 的交互。
3. **回归测试**：确保现有功能无退化。

---

## 约束条件

### 技术约束

#### Go BMP 模块 (ttpos-takeout)
- 使用 GoFrame 2.x。
- 使用 `github.com/grab/grabfood-api-sdk-go` SDK。
- 保持现有 RocketMQ 消息结构不变。

### 业务约束
- 本期仅做 SDK 迁移，不新增业务功能。
- Phase 4 新功能为可选项，视评估结果决定。

---

## 时间表

| 阶段    | 工作内容                           | 预估时间 |
| ------- | ---------------------------------- | -------- |
| Phase 1 | SDK 引入、Wrapper 层搭建           | 0.5 天   |
| Phase 2 | Outbound 功能迁移（OAuth、订单、门店） | 1-2 天   |
| Phase 3 | Inbound DTO 替换、清理废弃代码     | 1 天     |
| Phase 4 | 新功能评估与接入（可选）           | 1-2 天   |
| **总计** |                                    | 3-5 天   |

---

## 风险识别

| 风险                 | 影响 | 缓解措施                                     |
| -------------------- | ---- | -------------------------------------------- |
| SDK 版本兼容性       | 中   | 本地验证 SDK 与 Go 1.23 的兼容性             |
| Token 管理差异       | 中   | 封装 Wrapper 层统一 Token 获取，保持 Redis 缓存 |
| 响应结构变化         | 低   | 分阶段迁移，每个 API 单独测试验证            |
| 业务功能退化         | 高   | 完整的回归测试覆盖                           |

---

## 参考资料

- [Grab 官方 SDK](https://github.com/grab/grabfood-api-sdk-go)
- [GrabFood API Docs](https://developer.grab.com)
- 现有实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
- 现有 DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/`
- 关联 Spec: `docs/shared/specs/active/story-takeout-grab-integration/`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: rikugun
