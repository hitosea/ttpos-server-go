# Grab JWT 认证中间件使用说明

## 概述

`MiddlewareGrabJWTAuth` 是用于验证 Grab Partner API 请求中 JWT Token 的中间件。

## 功能

1. **Token 提取**: 从 `Authorization` 请求头中提取 `Bearer Token`
2. **Token 验证**: 调用 `PartnerTokenService.ParsePartnerToken` 验证 Token 的有效性
3. **Claims 存储**: 将验证通过的 Claims 存入请求 Context，供后续处理器使用
4. **错误处理**: 验证失败时返回标准的 OAuth 2.0 错误响应

## 路由配置

在 `internal/cmd/cmd.go` 中配置：

```go
// 注册 Grab 路由
s.Group("/api/v1/grab", func(group *ghttp.RouterGroup) {
    grabController := grab.NewV1()
    
    // 公开路由：OAuth Token 获取（不需要 JWT 认证）
    group.POST("/oauth/partner/webhook", grabController.OAuthPartnerWebhook)
    
    // 受保护路由：需要 JWT Token 认证
    group.Group("/", func(authGroup *ghttp.RouterGroup) {
        authGroup.Middleware(middleware.MiddlewareGrabJWTAuth)
        authGroup.Bind(grabController)
    })
})
```

## 请求示例

### 1. 获取 Token（公开接口）

```bash
curl -X POST http://localhost:8080/api/v1/grab/oauth/partner/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "scope": "food.partner_api"
  }'
```

**响应:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### 2. 调用受保护的 API

```bash
curl -X GET http://localhost:8080/api/v1/grab/menu/get \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Grab-Signature: ..." \
  -H "X-Grab-Timestamp: ..."
```

## 在 Controller 中使用 Claims

```go
func (c *ControllerV1) SomeProtectedEndpoint(ctx context.Context, req *v1.SomeReq) (res *v1.SomeRes, err error) {
    // 从 Context 中获取 JWT Claims
    claims := r.GetCtxVar(middleware.ContextKeyGrabPartnerClaims).Val()
    if partnerClaims, ok := claims.(*grab.PartnerTokenClaims); ok {
        g.Log().Infof(ctx, "Request from client_id=%s, scope=%s", 
            partnerClaims.ClientID, partnerClaims.Scope)
    }
    
    // 继续业务逻辑...
}
```

## 错误响应

验证失败时，中间件返回标准的 OAuth 2.0 错误格式：

```json
{
  "error": "unauthorized",
  "error_description": "Invalid or expired token"
}
```

常见错误：
- `Authorization header is required` - 缺少 Authorization 头
- `Invalid Authorization header format` - Authorization 格式错误
- `Token is required` - Token 为空
- `Invalid or expired token` - Token 无效或已过期

## 安全注意事项

1. **HTTPS**: 生产环境必须使用 HTTPS 传输 Token
2. **Token 过期**: Token 默认有效期 900 秒（15 分钟）
3. **Secret Key**: 确保 `secretKey` 配置安全，不要硬编码
4. **Scope 验证**: 根据需要在业务层额外验证 Scope 权限

## 配置

Token 签名密钥配置在 `manifest/config/config.yaml`:

```yaml
app:
  provider:
    grab:
      secretKey: "$GRAB_SECRET_KEY"  # Webhook 签名密钥，用于 JWT Token 签发和验证
      partner:
        default:
          clientId: "$GRAB_PARTNER_DEFAULT_CLIENT_ID"
          clientSecret: "$GRAB_PARTNER_DEFAULT_CLIENT_SECRET"
          scope: "food.partner_api"
          environment: "staging"
          timeout: "60s"
```

## 相关文件

- 中间件实现: `internal/middleware/grab_jwt_auth.go`
- Token 服务: `internal/logic/grab/partner_token_service.go`
- Claims 定义: `internal/model/dto/grab/partner_token.go`
- 路由配置: `internal/cmd/cmd.go`

