# 订单状态更新通知 API (Webhook)

## API 信息

**Name:** Order Status Update Notification API (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/order/status`

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
| orderId | String(20) | M | 订单唯一 ID，格式：LMF-yyMMdd-{generated number}，如：LMF-221031-338798091 |
| orderStatus | String | M | 订单状态，可选值：<br>- FINISH：已完成<br>- CANCELED：已取消 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示失败 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 订单状态更新成功 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 409 | 相同 Order ID 的订单已成功提交 |
| 500 | 服务器内部错误 |		