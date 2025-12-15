# 提案: Grab OAuth2 Token 接口实现

> **状态**: 已采纳
> **关联 Spec**: [task-takeout-grab-oauth-token](../../../shared/specs/active/task-takeout-grab-oauth-token/requirements.md)

## 1. 背景与动机 (Background and Motivation)
目前 `ttpos-takeout` 模块集成了 GrabFood 业务。为了调用 GrabFood API，需要获取 OAuth2 Access Token。
现有的 `internal/logic/grab/client.go` 中虽然包含获取 Token 的逻辑，但存在以下改进空间：
1. **配置管理**: `client_id` 和 `client_secret` 需要从标准配置文件 (`manifest/config/config.yaml`) 中读取，而不是硬编码或手动传递。
2. **缓存机制**: 当前使用内存缓存 (`c.accessToken`)，在多实例/集群部署下效率较低且无法共享。建议使用 Redis 缓存。
3. **接口标准化**: 参考 Grab 官方文档 `get-oauth-partner-webhook` (即 Client Credentials Flow)，规范化 Token 获取流程。

## 2. 目标 (Goals)
1. **配置驱动**: 从 `g.Cfg()` 读取 Grab 相关配置。
2. **Redis 缓存**: 使用 Redis 存储 Access Token，支持分布式共享，TTL 设置为 `expires_in - 60s`。
3. **逻辑优化**: 封装独立的 `AuthService` 或优化现有的 `APIClient`，提供稳定的 `GetToken()` 接口供其他业务调用。

## 3. 解决方案概述 (Solution Overview)

### 3.1 配置设计
在 `manifest/config/config.yaml` 中添加：
```yaml
takeout:
  grab:
    client_id: "${GRAB_CLIENT_ID}"         # 从环境变量获取
    client_secret: "${GRAB_CLIENT_SECRET}" # 从环境变量获取
    env: "staging"                         # production 或 staging
```

### 3.2 核心逻辑 (`internal/logic/grab`)
修改或增强 `client.go` 中的 `getAccessToken` 方法：
1. **读取缓存**: 尝试从 Redis 读取 Key (e.g., `ttpos:takeout:grab:token`).
2. **请求 Grab**: 若缓存不存在，调用 Grab OAuth2 Endpoint (`/grabid/v1/oauth2/token`).
   - Grant Type: `client_credentials`
   - Scope: `grab_food.partner`
3. **写入缓存**: 获取成功后写入 Redis，设置过期时间。

### 3.3 接口定义
如果需要暴露给外部调试，可以在 `internal/controller/grab` 添加简单的 HTTP 接口（仅限管理员或内部调用）。

## 4. 参考资料 (References)
- Grab OAuth Partner Webhook: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-oauth-partner-webhook
- Client Credentials Flow

## 5. 验收标准 (Acceptance Criteria)
- [ ] 能够从配置文件读取 `client_id` 和 `client_secret`。
- [ ] Token 获取成功后存储在 Redis 中。
- [ ] Token 过期后自动刷新。
- [ ] 现有 Grab API 调用（如 UpdateStoreStatus）能正常使用新的 Token 逻辑。

## 6. 计划 (Plan)
1. 修改 `config.tpl.yaml` 添加配置项。
2. 修改 `internal/logic/grab/client.go` 引入 Redis 依赖和 Config 读取。
3. 测试 Token 获取与刷新流程。

