# LINE MAN OAuth API 文档

> TTPOS 系统实现的 LINE MAN OAuth 2.0 认证接口

---

## 概述

本文档描述 TTPOS 系统为 LINE MAN 平台提供的 OAuth 2.0 认证接口。LINE MAN 通过此接口获取访问令牌，用于后续的 Webhook 回调认证。

### 认证流程

```mermaid
sequenceDiagram
    participant LM as LINE MAN Platform
    participant TTPOS as TTPOS BMP

    LM->>TTPOS: POST /oauth2/token (client_credentials)
    TTPOS->>TTPOS: 验证 client_id 和 client_secret
    TTPOS-->>LM: 返回 access_token

    Note over LM,TTPOS: 后续 Webhook 请求
    LM->>TTPOS: POST /orders (Bearer Token)
    TTPOS->>TTPOS: 验证 Token 有效性
    TTPOS-->>LM: 处理结果
```

---

## 接口定义

### 获取访问令牌

LINE MAN 通过 Client Credentials 授权模式获取访问令牌。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `POST /oauth2/token` |
| **方法** | POST |
| **Content-Type** | application/x-www-form-urlencoded |
| **认证** | 无（此接口用于获取认证） |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `grant_type` | string | Y | OAuth 授权类型，固定值 | `client_credentials` |
| `client_id` | string | Y | LINE MAN 分配的客户端 ID | `lineman_client_123` |
| `client_secret` | string | Y | LINE MAN 分配的客户端密钥 | `secret_abc123` |

#### 请求示例

```http
POST /oauth2/token HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=lineman_client_123&client_secret=secret_abc123
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJsaW5lbWFuIiwiZXhwIjoxNzM2NzU0ODM0fQ.xxx",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**响应字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `access_token` | string | 访问令牌，用于后续 API 调用的 Authorization 头 |
| `token_type` | string | 令牌类型，固定值 `Bearer` |
| `expires_in` | int | 令牌有效期（秒），通常为 3600（1小时） |

##### 失败响应

认证失败时返回错误信息：

```json
{
  "status": "fail",
  "code": "401",
  "message": "Invalid client credentials"
}
```

**错误码说明**：

| HTTP 状态码 | 错误原因 |
|------------|---------|
| 400 | 请求参数错误（缺少必填参数或 grant_type 不正确） |
| 401 | client_id 或 client_secret 无效 |
| 500 | 服务器内部错误 |

---

## Token 使用

### 在 Webhook 请求中使用

获取 Token 后，LINE MAN 在所有 Webhook 请求的 Header 中携带：

```http
Authorization: Bearer {access_token}
```

**示例**：

```http
POST /v1/partners/partner-123/stores/store-456/orders HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "orderId": "LMF-260113-338798091",
  ...
}
```

### Token 刷新

- Token 过期后，LINE MAN 需重新调用 `/oauth2/token` 获取新 Token
- 建议在 Token 过期前提前刷新，避免请求失败
- TTPOS 不提供 refresh_token 机制，每次需完整认证

---

## 安全注意事项

1. **客户端凭证保护**：client_secret 必须安全存储，不得泄露
2. **HTTPS 传输**：所有请求必须通过 HTTPS 传输
3. **Token 存储**：access_token 应安全存储，不得记录到日志
4. **有效期管理**：注意 Token 有效期，及时刷新

---

## 代码参考

- **API 定义**: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/oauth.go`
- **Controller**: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/`

---

## 版本历史

| 版本 | 日期 | 作者 | 变更内容 |
|------|------|------|----------|
| v1.0.0 | 2026-01-14 | Claude | 初始版本 |

---

**维护者**: TTPOS 后端开发组
**最后更新**: 2026-01-14
