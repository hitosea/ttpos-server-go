# 强制开启/关闭餐厅 API

## API 信息

**Name:** Force Close/Open Restaurant API

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/restaurant/availability`

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
| status | int | M | 餐厅状态，可选值：<br>- 1：强制开启餐厅<br>- 2：强制关闭餐厅<br><br>**强制关闭说明：**<br>餐厅将在次日营业时间自动开启。<br>例如：餐厅营业时间为 9:00-15:00，现在是 10:00 强制关闭，餐厅将关闭至明天 9:00。<br><br>**强制开启说明：**<br>- 营业前强制开启：餐厅当天营业结束时间关闭。<br>例如：餐厅营业时间为 9:00-15:00，现在是 8:00 强制开启，餐厅将开启至今天 15:00。<br>- 营业后强制开启：餐厅在次日营业结束时间关闭。<br>例如：餐厅营业时间为 9:00-15:00，现在是 16:00 强制开启，餐厅将开启至明天 15:00。 |
| duration | int | O | 持续时间（秒）<br><br>默认值为 0，持续时间将在下一个营业时间前结束（见 status 参数说明）。<br><br>设置持续时间后，餐厅将在持续时间结束后按营业时间自动开启/关闭。<br>**注意：** 以秒为单位，从 1 开始计数。<br>例如：餐厅营业时间为 9:00-15:00，现在是 6:00 强制开启并设置持续时间为 3600 秒，餐厅将开启至 7:00，然后在 9:00 按营业时间自动开启。 |

## Response Body

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | String | M | 结果状态："ok" 表示成功，"fail" 表示错误 |
| code | String | M | 结果代码 |
| message | String | O | 结果描述 |

## Response 状态码

| HTTP Status Code | Description |
| --- | --- |
| 200 | 餐厅状态更新成功 |
| 400 | 请求格式错误或缺少必需信息 |
| 401 | 无效的授权信息 |
| 404 | 无效的 partner ID 或 store ID |
| 500 | 服务器内部错误 |		