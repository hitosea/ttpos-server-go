# 菜单同步通知 API (Webhook)

## API 信息

**Name:** Menu Sync Notification API (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/menus/notification`

**方向:** LINE MAN → Partner (TTPOS)

## Header 参数

| Name | Type | Required | Value |
| --- | --- | --- | --- |
| Content-Type | String | M | application/json |
| Authorization | String | M | Bearer {access_token} |

## Path 参数

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| partnerId | String | M | 合作伙伴唯一 ID |
| storeId | String | M | 门店唯一 ID |

## Request Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| menuSyncRequestId | String(255) | M | 菜单同步请求唯一 ID |
| updatedAt | String | M | 更新时间，ISO 8601 格式，如：2022-11-01T13:08:06+07:00 |
| status | String | M | 菜单同步结果状态，可选值：<br>- SUCCESS：成功<br>- FAILED：失败 |
| error | String | O | 菜单同步过程中的错误信息，状态为 SUCCESS 时为空字符串 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示失败 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 菜单同步通知处理成功 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 409 | 相同请求已成功处理 |
| 500 | 服务器内部错误 |		