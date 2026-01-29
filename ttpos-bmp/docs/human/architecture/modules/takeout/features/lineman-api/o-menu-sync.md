# 菜单同步 API

## API 信息

**Name:** Menu Sync API

**Resource:** `PUT /v2/partners/{partnerId}/stores/{storeId}/menus`

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
| menuGroups | Array | M | 门店菜单分组列表 |
| menuGroups[].id | String(30) | M | 菜单分组唯一 ID |
| menuGroups[].name | Object | M | 菜单分组名称对象 |
| ...name.thai | String(255) | M | 菜单分组泰文名称 |
| ...name.english | String(255) | O | 菜单分组英文名称 |
| menuGroups[].useSellingTime | boolean | M | 是否启用销售时间限制 |
| menuGroups[].startSellingTime | int | C | 每日销售开始时间（UTC/GMT+7），当 useSellingTime 为 true 时必填<br>**注意：** 以分钟为单位，从午夜 0 开始计数，例如 120 = 2:00 AM |
| menuGroups[].endSellingTime | int | C | 每日销售结束时间（UTC/GMT+7），当 useSellingTime 为 true 时必填<br>**注意：** 以分钟为单位，从午夜 0 开始计数，例如 240 = 4:00 AM |
| menuGroups[].menuItems | Array | M | 菜单分组中的商品列表 |
| ...menuItems[].id | String(30) | M | 商品唯一 ID，用于下单 |
| ...menuItems[].name | Object | M | 商品名称对象 |
| ...name.thai | String(125) | M | 商品泰文名称 |
| ...name.english | String(125) | O | 商品英文名称 |
| ...menuItems[].description | Object | M | 商品描述对象 |
| ...description.thai | String | O | 商品泰文描述 |
| ...description.english | String | O | 商品英文描述 |
| ...menuItems[].price | double | M | 商品价格 |
| ...menuItems[].photoUrl | String | O | 商品图片 URL |
| ...menuItems[].menuStatus | String | O | 商品状态，可选值：<br>- AVAILABLE（默认，可售）<br>- SUSPENDED（暂停销售）<br>- SOLD_OUT_TODAY（今日售罄，明天自动恢复） |
| ...menuItems[].startSellingTime | int | O | 每日销售开始时间（UTC/GMT+7）<br>**注意：** 以分钟为单位，从午夜 0 开始计数，例如 120 = 2:00 AM |
| ...menuItems[].endSellingTime | int | C | 每日销售结束时间（UTC/GMT+7），当 startSellingTime 有值时必填<br>**注意：** 以分钟为单位，从午夜 0 开始计数，例如 240 = 4:00 AM |
| ...menuItems[].salesChannelsAvailability | Object | M | 销售渠道配置<br>**注意：** 至少启用一个销售渠道 |
| ...salesChannelsAvailability.delivery | boolean | C | 是否在外送渠道可售 |
| ...salesChannelsAvailability.pickup | boolean | C | 是否在自取渠道可售 |
| ...menuItems[].salesChannelsPrice | Object | O | 销售渠道价格 |
| ...salesChannelsPrice.pickup | double | O | 自取价格 |
| ...menuItems[].properties | Array | O | 商品属性列表 |
| ...properties[].id | String(160) | M | 属性唯一 ID |
| ...properties[].name | Object | M | 属性名称对象 |
| ...name.thai | String(512) | M | 属性泰文名称 |
| ...name.english | String(512) | O | 属性英文名称 |
| ...properties[].min | int | M | 最少选择数量 |
| ...properties[].max | int | O | 最多选择数量 |
| ...properties[].type | String | M | 属性类型，可选值：<br>- 1：单选（RADIO）<br>- 2：多选（CHECKBOX） |
| ...properties[].values | Array | M | 属性值列表 |
| ...values[].id | String(160) | M | 属性值唯一 ID |
| ...values[].name | Object | M | 属性值名称对象 |
| ...name.thai | String(512) | M | 属性值泰文名称 |
| ...name.english | String(512) | O | 属性值英文名称 |
| ...values[].price | double | O | 属性值加价（默认 0.00） |
| ...values[].status | int | M | 属性值状态，可选值：<br>- 1：AVAILABLE（可用）<br>- 2：SOLD_OUT_TODAY（今日售罄，明天自动恢复）<br>- 3：SUSPENDED（暂停） |
| ...values[].salesChannelsPrice | Object | O | 销售渠道价格 |
| ...salesChannelsPrice.pickup | double | O | 自取价格 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示错误 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |
| menuSyncRequestId | String | O | 菜单同步请求唯一 ID |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 菜单同步请求成功接受 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 500 | 服务器内部错误 |		