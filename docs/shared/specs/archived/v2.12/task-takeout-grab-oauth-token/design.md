# 技术设计: Grab OAuth2 Token 接口实现

> **关联 Spec**: [requirements.md](./requirements.md)
> **状态**: Draft
> **作者**: rikugun
> **日期**: 2025-12-05

## 1. 架构设计 (Architecture)

### 1.1 模块位置
- **模块**: `ttpos-takeout`
- **层级**: Logic Layer (`internal/logic/grab`)

### 1.2 依赖关系
- `g.Cfg()`: 读取配置文件
- `g.Redis()`: 缓存 Token
- `g.Client()`: 发起 OAuth2 HTTP 请求

### 1.3 核心流程
1. `APIClient.getAccessToken` 被调用。
2. 构造 Redis Key: `ttpos:takeout:grab:token:{client_id}`。
3. **READ**: `g.Redis().Get(ctx, key)`。
4. **HIT**: 返回 Token。
5. **MISS**:
   - 调用 `fetchTokenFromGrab()`。
   - 成功后，计算 TTL = `expires_in - 60s`。
   - **WRITE**: `g.Redis().SetEx(ctx, key, token, ttl)`。
   - 返回 Token。

## 2. 数据结构设计 (Data Structures)

### 2.1 Redis Key
| Key Pattern | Type | TTL | Description |
|---|---|---|---|
| `ttpos:takeout:grab:token:{client_id}` | String | Dynamic | 存储 Access Token，TTL 为 Token 有效期减去 60秒 |

### 2.2 配置文件 (`manifest/config/config.yaml`)
新增结构:
```yaml
takeout:
  grab:
    client_id: "${GRAB_CLIENT_ID}"
    client_secret: "${GRAB_CLIENT_SECRET}"
    env: "staging" # or production
    timeout: 30s
```

## 3. 接口设计 (API Design)

无需新增外部 HTTP/gRPC 接口，仅重构内部 Logic 方法。

### 3.1 `internal/logic/grab/client.go`
修改 `ClientConfig` 结构体：
```go
type ClientConfig struct {
    ClientID     string
    ClientSecret string
    Environment  string
    Timeout      time.Duration
}
```

修改 `getAccessToken` 方法签名保持不变，内部逻辑替换为 Redis 实现。

## 4. 详细设计 (Detailed Design)

### 4.1 Token Fetch 逻辑
使用 `g.Client()` 发起 POST 请求:
- URL: 
  - Staging: `https://api.stg-myteksi.com/grabid/v1/oauth2/token`
  - Production: `https://api.grab.com/grabid/v1/oauth2/token`
- Content-Type: `application/x-www-form-urlencoded`
- Body: `client_id=...&client_secret=...&grant_type=client_credentials&scope=grab_food.partner`

### 4.2 错误处理
- Redis 故障: 降级为仅内存获取（虽然无法共享，但能保证服务可用），或者直接返回错误（取决于一致性要求，建议 Log Error 后尝试远程获取但不缓存）。本设计暂定为：**Redis 错误视为 Miss，尝试远程获取，写入失败仅 Log**。
- Grab API 故障: 返回 error，由上层重试。

## 5. 安全考虑 (Security)
- `client_secret` 必须通过环境变量注入，严禁提交到代码仓库。
- Redis 中仅存储 Token 值，不存储敏感配置。

## 6. 测试策略 (Testing Strategy)
- **单元测试**: Mock Redis 和 HTTP Server，测试 Cache Hit/Miss 和 Token 过期场景。
- **集成测试**: 配置真实 Staging 账号，验证能成功获取 Token 并写入 Redis。

