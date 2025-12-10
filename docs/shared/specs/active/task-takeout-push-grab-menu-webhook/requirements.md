# Push Grab Menu Webhook 需求规格说明书

> 来源 Proposal: [push-grab-menu-webhook](../../../../team/proposals/2025-12/push-grab-menu-webhook.md)

## 1. 基础信息

| 项目 | 内容 |
| --- | --- |
| **功能名称** | Push Grab Menu Webhook |
| **功能模块** | takeout (外卖集成) |
| **优先级** | P1 |
| **目标版本** | v2.11.0 |
| **责任人** | rikugun |
| **状态** | 待审核 |

---

## 2. 需求背景

在 GrabFood Self-Serve Activation 流程中，当商户选择"导出当前 Grab 门店菜单到 POS"时，GrabFood 会调用 Partner 的 `Push Grab Menu Webhook` 端点，将现有菜单数据推送给 POS 系统。

当前 `ttpos-takeout` 项目中已定义了 `PushGrabMenuWebhookReq` 和 `PushGrabMenuWebhookRes` 的 API 结构，但缺少完整的请求体定义和业务逻辑实现。

**业务价值**:
- 支持 GrabFood Self-Serve Activation 流程的完整实现
- 允许商户在激活集成时导出现有 Grab 菜单到 POS 系统
- 减少商户手动录入菜单的工作量
- 支持 Grab 菜单结构的标准化 JSON 格式

---

## 3. 功能需求 (Functional Requirements)

### 3.1 接收 Grab 菜单推送

**ID**: FR-001
**描述**: 实现 `Push Grab Menu Webhook` API 端点，接收 GrabFood 推送的 JSON 格式菜单数据。
**验收标准**:
1. API 路径与 Grab Partner Portal 配置一致（通常为 `/grab/v1/push_grab_menu` 或类似）。
2. 请求体必须与 Grab SDK `GetMenuNewResponse` 结构完全兼容（包含 `merchantID`, `partnerMerchantID`, `currency`, `sellingTimes`, `categories` 等）。
3. 成功接收数据后，返回 HTTP 204 No Content。
4. 鉴权机制复用现有的 Partner OAuth Token 验证逻辑（Authorization Header）。

### 3.2 存储菜单数据

**ID**: FR-002
**描述**: 接收到菜单数据后，将其暂存以便后续处理。
**验收标准**:
1. 将接收到的完整 JSON 菜单数据存储到 Redis 或数据库中。
2. 存储 Key 或关联 ID 需包含 `partnerMerchantID` 或 `merchantID`，确保数据隔离。
3. 记录接收时间。

### 3.3 触发菜单更新通知

**ID**: FR-003
**描述**: 菜单数据保存成功后，发送消息通知下游系统。
**验收标准**:
1. 发送 RocketMQ 消息（Topic: `provider_menu_update`）。
2. 消息内容需包含 `ProviderName` (grab), `MerchantID`, `PartnerMerchantID` 以及数据存储的引用（或数据本身，视大小而定）。
3. 确保消息发送的可靠性。

---

## 4. 非功能需求 (Non-Functional Requirements)

### 4.1 兼容性
- 必须使用 Grab 官方 SDK (`github.com/grab/grabfood-api-sdk-go`) 的类型定义，确保 100% 字段对齐。

### 4.2 性能
- Webhook 响应时间需在 3秒内，建议异步处理（接收 -> 存 MQ/Redis -> 返回 204）。

### 4.3 安全性
- 验证 `Authorization` Header 中的 Token 有效性。
- 日志中不得打印完整的菜单 JSON 如果包含敏感信息（菜单通常不含敏感信息，但需注意日志量）。

---

## 5. UI/UX 需求

本需求为后端 API 实现，无前端 UI 界面。

---

## 6. 数据需求

- **菜单数据模型**: 复用 Grab SDK 定义的数据结构。
- **存储**: 临时存储（Redis TTL 1小时）或持久化存储（DB）视业务后续处理流程而定（建议先存 Redis + 触发事件）。

---

## 7. 影响范围

**模块**: `ttpos-bmp/app/ttpos-takeout`
**技术栈**: Go (GoFrame)

---

## 8. 待确认事项

- [ ] 菜单数据是否需要立即持久化到 MySQL `channel_menu_snapshot` 表？（建议：是，作为快照保存）
- [ ] `provider_menu_update` 消息的具体 Payload 格式定义。
