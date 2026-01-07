# 触发菜单同步 API (Webhook)

## API 信息

**Name:** Trigger Sync Menu API (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/menus/trigger-sync`

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

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示失败 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 菜单同步触发成功 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 500 | 服务器内部错误 |		