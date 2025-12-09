# 需求规格: Grab OAuth2 Token 接口实现

- **状态**: 已通过
- **模块**: ttpos-takeout (BMP)
- **类型**: Technical Task
- **优先级**: P1
- **负责人**: rikugun
- **来源 Proposal**: [Grab OAuth2 Token 接口实现](../../../../team/proposals/2025-12/grab-oauth-token.md)
- **创建日期**: 2025-12-05

## 1. 概述 (Overview)

本需求旨在规范化 `ttpos-takeout` 模块中 GrabFood API 的认证流程。目前 Token 获取逻辑存在硬编码配置、内存缓存（不支持集群）等问题。
本任务将实现标准的 OAuth2 Client Credentials Flow，使用配置文件管理凭证，并引入 Redis 进行分布式缓存，确保多实例部署下的稳定性和安全性。

## 2. 核心需求 (Core Requirements)

### 2.1 配置管理
- **功能**: 从标准配置文件读取 OAuth2 凭证。
- **配置项**:
  - `client_id`: 必填，从环境变量注入。
  - `client_secret`: 必填，从环境变量注入。
  - `environment`: `production` 或 `staging`，影响 API Base URL。
  - `scopes`: 默认为 `grab_food.partner`。

### 2.2 Token 获取与刷新
- **功能**: 实现 `GetToken()` 方法，保证调用时能获取有效 Token。
- **逻辑**:
  1. **检查缓存**: 优先从 Redis 读取有效 Token。
  2. **远程获取**: 缓存未命中或过期时，调用 Grab `/grabid/v1/oauth2/token` 接口。
  3. **写入缓存**: 获取成功后写入 Redis，TTL 设置为 `expires_in - 60s` (预留缓冲时间)。
  4. **并发控制**: (可选) 防止缓存击穿，考虑简单的锁或单飞模式（Singleflight）。

### 2.3 Redis 缓存集成
- **Key 格式**: `ttpos:takeout:grab:token:{client_id}` (建议包含 client_id 以支持多账号，若单账号可简化)。
- **存储内容**: Access Token 字符串。
- **过期策略**: 自动过期，无需手动清理。

## 3. 技术约束 (Technical Constraints)

- **框架**: GoFrame v2.x
- **依赖**:
  - 配置读取: `g.Cfg()`
  - Redis 操作: `g.Redis()`
  - HTTP 客户端: `g.Client()` 或现有 `http.Client`
- **代码位置**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/`
- **兼容性**: 必须保持 `APIClient` 现有方法的签名不变，或进行低侵入式重构，避免破坏现有业务调用。

## 4. 验收标准 (Acceptance Criteria)

1. **配置读取**: 修改配置文件后，应用能正确读取新的 `client_id` 和 `client_secret`。
2. **Token 缓存**:
   - 首次启动或缓存过期时，能看到对 Grab OAuth 接口的请求。
   - 再次调用时，直接从 Redis 获取，无 Grab 接口请求。
   - Redis 中能查看到对应的 Key 和正确的 TTL。
3. **接口可用性**: 使用获取的 Token 调用 `GetStoreStatus` 等接口返回 HTTP 200。
4. **异常处理**: 当 Grab OAuth 接口返回错误（如 401 Invalid Client）时，系统应记录错误日志并返回明确错误，而不是 panic 或死循环。

## 5. 影响范围 (Impact Analysis)

- **受影响模块**: `ttpos-takeout`
- **受影响代码**: `internal/logic/grab/client.go`
- **数据库变更**: 无
- **配置变更**: `manifest/config/config.tpl.yaml`

## 6. 待确认事项 (Questions)
- 是否需要支持多店铺（多 Client ID）？目前假设单店铺/单 Client ID 模式。

