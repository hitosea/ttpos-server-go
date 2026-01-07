# OAuth 认证 API

## API 信息

**Resource:** `POST /v1/lmwn/oauth2/token`

**方向:** LINE MAN ← Partner (TTPOS)

## Header 参数

| Name | Type | Required | Value |
| --- | --- | --- | --- |
| Content-Type | String | M | application/x-www-form-urlencoded |

## Request Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| grant_type | String | M | OAuth 授权类型，固定值：`client_credentials` |
| client_id | String | M | LINE MAN 分配的客户端 ID |
| client_secret | String | M | LINE MAN 分配的客户端密钥 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| access_token | String | M | 访问令牌，用于后续 API 调用 |
| token_type | String | M | 令牌类型，固定值：`Bearer` |
| expires_in | int | M | 令牌有效期（秒），通常为 3600 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 认证成功，返回访问令牌 |
| 400 | 请求格式错误或缺少必需参数 |
| 401 | 客户端 ID 或密钥无效 |
| 500 | 服务器内部错误 |

## 说明

- 访问令牌用于所有后续的 LINE MAN API 调用
- 令牌过期后需要重新申请
- 请妥善保管 client_id 和 client_secret		