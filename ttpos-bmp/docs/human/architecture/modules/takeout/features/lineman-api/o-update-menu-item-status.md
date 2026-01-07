# 更新菜单商品状态 API

## API 信息

**Name:** Update Menu Item Status API

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/menu/items/status`

**方向:** Partner (TTPOS) → LINE MAN

## Header 参数

| Name | Type | Required | Value |
| --- | --- | --- | --- |
| Content-Type | String | M | application/json |
| Authorization | String | M | Bearer {access_token} |

## Path 参数

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| partnerId | String(50) | M | 合作伙伴唯一 ID |
| storeId | String(50) | M | 门店唯一 ID |

## Request Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| menuItems | Array | M | 门店商品状态列表<br>**备注：** 最大数组长度 100 |
| ...menuItems[].id | String(255) | M | 商品 ID（与菜单同步中的 menuItem.id 相同） |
| ...menuItems[].menuStatus | String | M | 商品状态，可选值：<br>- AVAILABLE（默认，可售）<br>- SUSPENDED（暂停销售）<br>- SOLD_OUT_TODAY（今日售罄，明天自动恢复） |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示错误 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 菜单状态更新成功 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 500 | 服务器内部错误 |		