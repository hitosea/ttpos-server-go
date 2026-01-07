# 下单 API (Webhook)

## API 信息

**Name:** Place Order API (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/orders`

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
| orderShortCode | String(4) | M | 短订单 ID，为 orderId 的后四位 |
| restaurantRevenue | double | M | 商户收入总额（已扣除合作伙伴补贴折扣） |
| orderAcceptedTime | String | M | 订单接受时间，ISO 8601 格式，如：2022-11-01T13:08:06+07:00 |
| items | Array | M | 订单商品列表（可能包含重复商品，因属性或备注不同） |
| items[].id | String(255) | M | 菜单项 ID |
| items[].quantity | int | M | 商品数量 |
| items[].unitPrice | double | M | 商品单价（THB），包含额外选项费用，已应用促销折扣（商户补贴） |
| items[].memo | String | O | 顾客备注 |
| items[].promotionId | String(255) | O | 促销活动 ID（如有） |
| items[].discount | double | O | 促销折扣金额（商户补贴） |
| items[].properties | Array | O | 商品选项 |
| ...properties[].id | String(255) | M | 选项 ID |
| ...properties[].values | Array | M | 已选择的选项值 |
| ...values[].id | String(255) | M | 选项值 ID |
| ...values[].price | double | M | 选项值价格（THB） |
| additionalItems | Array | M | 订单附加项列表 |
| additionalItems[].name | String(1024) | M | 附加信息，如："ไม่รับช้อนส้อมพลาสติก" |
| memberId | String(255) | O | 绑定 LINE MAN 账号的会员 ID |
| customerType | String(32) | M | 订单类型，可选值：<br>- DELIVERY：外送<br>- PICKUP：自取 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示失败 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 订单成功发送到门店 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 409 | 相同 Order ID 的订单已成功提交 |
| 500 | 服务器内部错误 |		