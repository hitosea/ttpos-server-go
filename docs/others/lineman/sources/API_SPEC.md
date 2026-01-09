# LINE MAN Partner Integration API Specification

> 转换自 `sources/API_SPEC.xlsx`

---

## 目录

1. [Authentication](#1-authentication)
2. [Menu Sync (v1)](#2-menu-sync-v1)
3. [Menu Sync (v2)](#3-menu-sync-v2)
4. [Place Order (Webhook)](#4-place-order-webhook)
5. [Update Menu Item Status](#5-update-menu-item-status)
6. [Update Menu Property Values Status](#6-update-menu-property-values-status)
7. [Force Close/Open Restaurant](#7-force-closeopen-restaurant)
8. [Trigger Sync Menu (Webhook)](#8-trigger-sync-menu-webhook)
9. [Order Update Notification (Webhook)](#9-order-update-notification-webhook)
10. [Order Status Update Notification (Webhook)](#10-order-status-update-notification-webhook)
11. [Menu Sync Notification (Webhook)](#11-menu-sync-notification-webhook)

---

## 1. Authentication

**Resource:** `POST /v1/oauth/token`  
**Direction:** Partner ↔ LM

### Header Parameters

| Name         | Type   | Required | Value                             |
| ------------ | ------ | -------- | --------------------------------- |
| Content-Type | String | M        | application/x-www-form-urlencoded |

### Request Body

| Name          | Type   | Required | Description                                               |
| ------------- | ------ | -------- | --------------------------------------------------------- |
| grant_type    | String | M        | The only supported grant type is the "client_credentials" |
| client_id     | String | M        | The client ID for partner                                 |
| client_secret | String | M        | The client secret for partner                             |

### Successful Response Body

| Name         | Type   | Required | Description                                                                                                                                  |
| ------------ | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| access_token | String | M        | Any string value that is going to be used as token for the communication                                                                     |
| token_type   | String | M        | The token type to be used in the Authorization header before the access token. Note: The only supported token type is the Bearer token type. |
| expires_in   | int    | M        | The lifetime of the access token in seconds (measured from the creation time).                                                               |

### Response Codes

| HTTP Status | Description                                         |
| ----------- | --------------------------------------------------- |
| 200         | The authorization request succeeded                 |
| 401         | Invalid authorization information                   |
| 500         | The application has experienced an internal problem |

---

## 2. Menu Sync (v1)

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/menus`  
**Direction:** Partner → LM

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type       | Required | Description                              |
| --------- | ---------- | -------- | ---------------------------------------- |
| partnerId | String(50) | M        | Unique ID of the partner                 |
| storeId   | String(50) | M        | Unique ID of the partner store or branch |

### Request Body

| Name                            | Type        | Required | Description                                                                                                                                                                                                          |
| ------------------------------- | ----------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| menuGroups                      | Array       | M        | Contains the list of menu groups in a store                                                                                                                                                                          |
| menuGroups[].id                 | String(30)  | M        | Unique ID to which the menu group to be updated belongs                                                                                                                                                              |
| menuGroups[].name               | Object      | M        | Contains the menu group names                                                                                                                                                                                        |
| ...name.thai                    | String(255) | M        | Menu group Thai name                                                                                                                                                                                                 |
| ...name.english                 | String(255) | O        | Menu group English name                                                                                                                                                                                              |
| menuGroups[].useSellingTime     | boolean     | M        | If true, the menu group will be available by the sellingTime parameters                                                                                                                                              |
| menuGroups[].startSellingTime   | int         | C        | Available from this time (UTC/GMT+7) on each day. Required if useSellingTime is true. Note: The value is the number of minutes that start from 0 at midnight. For example, if the value is 120, that equals 2:00 AM. |
| menuGroups[].endSellingTime     | int         | C        | Available to this time (UTC/GMT+7) on each day. Required if useSellingTime is true. Note: The value is the number of minutes that start from 0 at midnight. For example, if the value is 240, that equals 4:00 AM.   |
| menuGroups[].menuItems          | Array       | M        | Contains the list of menu items in a menu group                                                                                                                                                                      |
| ...menuItems[].id               | String(30)  | M        | Unique ID to which the menu item to be updated belongs and is used in order placing.                                                                                                                                 |
| ...menuItems[].name             | Object      | M        | Contains the menu item names                                                                                                                                                                                         |
| ...name.thai                    | String(125) | M        | Menu item Thai name                                                                                                                                                                                                  |
| ...name.english                 | String(125) | O        | Menu item English name                                                                                                                                                                                               |
| ...menuItems[].description      | Object      | M        | Contains the menu item descriptions                                                                                                                                                                                  |
| ...description.thai             | String      | O        | Menu item Thai description                                                                                                                                                                                           |
| ...description.english          | String      | O        | Menu item English description                                                                                                                                                                                        |
| ...menuItems[].price            | double      | M        | Price of the menu item                                                                                                                                                                                               |
| ...menuItems[].photoUrl         | String      | O        | The image url of the menu item                                                                                                                                                                                       |
| ...menuItems[].menuStatus       | String      | O        | The status of the menu item: AVAILABLE (default), SUSPENDED, SOLD_OUT_TODAY (automatically AVAILABLE tomorrow)                                                                                                       |
| ...menuItems[].startSellingTime | int         | O        | Available from this time (UTC/GMT+7) on each day. Note: The value is the number of minutes that start from 0 at midnight.                                                                                            |
| ...menuItems[].endSellingTime   | int         | C        | Available to this time (UTC/GMT+7) on each day. Required if the startSellingTime has value.                                                                                                                          |
| ...menuItems[].properties       | Array       | O        | Contains the list of properties in a menu item                                                                                                                                                                       |
| ...properties[].id              | String(160) | M        | Unique ID to which the option to be updated belongs                                                                                                                                                                  |
| ...properties[].name            | Object      | M        | Contains the property names                                                                                                                                                                                          |
| ...name.thai                    | String(512) | M        | Property Thai name                                                                                                                                                                                                   |
| ...name.english                 | String(512) | O        | Property English name                                                                                                                                                                                                |
| ...properties[].min             | int         | M        | Minimum number of selected property value                                                                                                                                                                            |
| ...properties[].max             | int         | O        | Maximum number of selected property value                                                                                                                                                                            |
| ...properties[].type            | String      | M        | Input type: 1 = RADIO, 2 = CHECKBOX                                                                                                                                                                                  |
| ...properties[].values          | Array       | M        | Contains the list of property values in a property                                                                                                                                                                   |
| ...values[].id                  | String(160) | M        | Unique ID to which the selected option to be updated belongs                                                                                                                                                         |
| ...values[].name                | Object      | M        | Contains the property value names                                                                                                                                                                                    |
| ...name.thai                    | String(512) | M        | Property value Thai name                                                                                                                                                                                             |
| ...name.english                 | String(512) | O        | Property value English name                                                                                                                                                                                          |
| ...values[].price               | double      | O        | Price of the property value if selected (default is 0.00)                                                                                                                                                            |
| ...values[].status              | int         | M        | Status: AVAILABLE (1), SOLD_OUT_TODAY (2), SUSPENDED (3)                                                                                                                                                             |

### Response Body

| Name              | Type   | Required | Description                                          |
| ----------------- | ------ | -------- | ---------------------------------------------------- |
| status            | String | M        | Result: "ok" for success, "fail" for error           |
| code              | String | M        | Contains success or error result code                |
| message           | String | O        | Result description                                   |
| menuSyncRequestId | String | O        | The unique ID which the successful menu sync request |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | The menu has been successfully accepted                   |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 3. Menu Sync (v2)

**Resource:** `PUT /v2/partners/{partnerId}/stores/{storeId}/menus`  
**Direction:** Partner → LM

> v2 相比 v1 新增了 `salesChannelsAvailability` 和 `salesChannelsPrice` 字段，支持 Delivery 和 Pickup 渠道差异化配置。

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type       | Required | Description                              |
| --------- | ---------- | -------- | ---------------------------------------- |
| partnerId | String(50) | M        | Unique ID of the partner                 |
| storeId   | String(50) | M        | Unique ID of the partner store or branch |

### Request Body (v2 新增字段)

| Name                                     | Type    | Required | Description                                                                  |
| ---------------------------------------- | ------- | -------- | ---------------------------------------------------------------------------- |
| ...menuItems[].salesChannelsAvailability | Object  | M        | Sales channels of the menu item. Required at least one sales channel enabled |
| ...salesChannelsAvailability.delivery    | boolean | C        | If true, the menu item will be available in delivery sales channels          |
| ...salesChannelsAvailability.pickup      | boolean | C        | If true, the menu item will be available in pick & go sales channels         |
| ...menuItems[].salesChannelsPrice        | Object  | O        | Sales channels price of the menu item                                        |
| ...salesChannelsPrice.pickup             | double  | O        | Pickup price of the menu item                                                |
| ...values[].salesChannelsPrice           | Object  | O        | Sales channels price of the property value                                   |
| ...salesChannelsPrice.pickup             | double  | O        | Pickup price of the property value                                           |

### Response Body

| Name              | Type   | Required | Description                                          |
| ----------------- | ------ | -------- | ---------------------------------------------------- |
| status            | String | M        | Result: "ok" for success, "fail" for error           |
| code              | String | M        | Contains success or error result code                |
| message           | String | O        | Result description                                   |
| menuSyncRequestId | String | O        | The unique ID which the successful menu sync request |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | The menu has been successfully accepted                   |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 4. Place Order (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/orders`  
**Direction:** LM → Partner

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type   | Required | Description                              |
| --------- | ------ | -------- | ---------------------------------------- |
| partnerId | String | M        | Unique ID of the partner                 |
| storeId   | String | M        | Unique ID of the partner store or branch |

### Request Body

| Name                   | Type         | Required | Description                                                                                      |
| ---------------------- | ------------ | -------- | ------------------------------------------------------------------------------------------------ |
| orderId                | String(20)   | M        | Order Unique ID in format `LMF-yyMMdd-{generated number}` such as `LMF-221031-338798091`         |
| orderShortCode         | String(4)    | M        | Short order ID, which is the last four digits of orderId                                         |
| restaurantRevenue      | double       | M        | The total restaurant revenue of the order already subtracts partner subsidiary discounts         |
| orderAcceptedTime      | String       | M        | The order accepted time in ISO 8601 Format such as `2022-11-01T13:08:06+07:00`                   |
| items                  | Array        | M        | List of the ordered menu items (duplicated items may occur due to different property or memo)    |
| items[].id             | String(255)  | M        | Menu item id                                                                                     |
| items[].quantity       | int          | M        | Menu Item quantity                                                                               |
| items[].unitPrice      | double       | M        | Final Menu Item Price (THB). Includes extra charge for options and promotional discount applied. |
| items[].memo           | String       | O        | Menu Item memo requesting from customer                                                          |
| items[].promotionId    | String(255)  | O        | If this item has a promotion, contains a unique promotion id                                     |
| items[].discount       | double       | O        | If this item has a promotion, contains a discount (restaurant subsidize) amount                  |
| items[].properties     | Array        | O        | Options for the item                                                                             |
| ...properties[].id     | String(255)  | M        | Option id                                                                                        |
| ...properties[].values | Array        | M        | Selected options for the item                                                                    |
| ...values[].id         | String(255)  | M        | Selected option id                                                                               |
| ...values[].price      | double       | M        | Selected option price (THB)                                                                      |
| additionalItems        | Array        | M        | List of the additional items of order                                                            |
| additionalItems[].name | String(1024) | M        | Name to contain an additional message such as "ไม่รับช้อนส้อมพลาสติก"                            |
| memberId               | String(255)  | O        | The user member ID of the partner's system binds with the LINE MAN account                       |
| customerType           | String(32)   | M        | The received order method: DELIVERY or PICKUP                                                    |

### Response Body

| Name    | Type   | Required | Description                                    |
| ------- | ------ | -------- | ---------------------------------------------- |
| status  | String | M        | Result: "ok" for success, "fail" for unsuccess |
| code    | String | M        | Contains success or error result code          |
| message | String | O        | Result description                             |

### Response Codes

| HTTP Status | Description                                                                          |
| ----------- | ------------------------------------------------------------------------------------ |
| 200         | The order is successfully sent to the store                                          |
| 400         | The request is malformed or missing mandatory information                            |
| 401         | Invalid authorization information                                                    |
| 404         | Invalid partner ID and/or store ID                                                   |
| 409         | An order with the same Order ID has already been successfully submitted to the store |
| 500         | The application has experienced an internal problem                                  |

---

## 5. Update Menu Item Status

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/menu/items/status`  
**Direction:** Partner → LM

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type       | Required | Description                              |
| --------- | ---------- | -------- | ---------------------------------------- |
| partnerId | String(50) | M        | Unique ID of the partner                 |
| storeId   | String(50) | M        | Unique ID of the partner store or branch |

### Request Body

| Name                      | Type        | Required | Description                                                                               |
| ------------------------- | ----------- | -------- | ----------------------------------------------------------------------------------------- |
| menuItems                 | Array       | M        | Contains the list of menu item status in a store. **Max array size: 100**                 |
| ...menuItems[].id         | String(255) | M        | The menu Item's ID on the partner system (same as menuItem.id in Menu Sync)               |
| ...menuItems[].menuStatus | String      | M        | Status: AVAILABLE (default), SUSPENDED, SOLD_OUT_TODAY (automatically AVAILABLE tomorrow) |

### Response Body

| Name    | Type   | Required | Description                                |
| ------- | ------ | -------- | ------------------------------------------ |
| status  | String | M        | Result: "ok" for success, "fail" for error |
| code    | String | M        | Contains success or error result code      |
| message | String | O        | Result description                         |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | The menu has been successfully accepted                   |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 6. Update Menu Property Values Status

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/menu/property/values/status`  
**Direction:** Partner → LM

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type       | Required | Description                              |
| --------- | ---------- | -------- | ---------------------------------------- |
| partnerId | String(50) | M        | Unique ID of the partner                 |
| storeId   | String(50) | M        | Unique ID of the partner store or branch |

### Request Body

| Name                       | Type        | Required | Description                                                                     |
| -------------------------- | ----------- | -------- | ------------------------------------------------------------------------------- |
| propertyValues             | Array       | M        | Contains the list of property values status in a store. **Max array size: 100** |
| ...propertyValues[].id     | String(255) | M        | The property value's ID on the partner system (same as values.id in Menu Sync)  |
| ...propertyValues[].status | int         | M        | Status: AVAILABLE (1), SOLD_OUT_TODAY (2), SUSPENDED (3)                        |

### Response Body

| Name    | Type   | Required | Description                                |
| ------- | ------ | -------- | ------------------------------------------ |
| status  | String | M        | Result: "ok" for success, "fail" for error |
| code    | String | M        | Contains success or error result code      |
| message | String | O        | Result description                         |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | The menu has been successfully accepted                   |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 7. Force Close/Open Restaurant

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/restaurant/availability`  
**Direction:** Partner → LM

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type       | Required | Description                              |
| --------- | ---------- | -------- | ---------------------------------------- |
| partnerId | String(50) | M        | Unique ID of the partner                 |
| storeId   | String(50) | M        | Unique ID of the partner store or branch |

### Request Body

| Name     | Type | Required | Description                                                                |
| -------- | ---- | -------- | -------------------------------------------------------------------------- |
| status   | int  | M        | 1 = force open restaurants, 2 = force close the restaurant                 |
| duration | int  | O        | Duration Time (seconds). Default (0) = ends before the next delivery time. |

**Force Close/Open 逻辑说明：**

- **Force close (status=2)**: 餐厅将关闭直到第二天的营业时间

  - 例如：营业时间 9:00-15:00，当前 10:00 执行 force close，餐厅将关闭直到明天 9:00

- **Force open (status=1) 在营业时间前**: 餐厅将在当天营业时间结束时关闭

  - 例如：营业时间 9:00-15:00，当前 8:00 执行 force open，餐厅将开放直到今天 15:00

- **Force open (status=1) 在营业时间后**: 餐厅将在第二天营业时间结束时关闭

  - 例如：营业时间 9:00-15:00，当前 16:00 执行 force open，餐厅将开放直到明天 15:00

- **duration 参数**: 设置持续时间后，到期后恢复正常营业时间
  - 例如：营业时间 9:00-15:00，当前 6:00 执行 force open 且 duration=3600，餐厅将开放到 7:00，然后在 9:00 正常营业

### Response Body

| Name    | Type   | Required | Description                                |
| ------- | ------ | -------- | ------------------------------------------ |
| status  | String | M        | Result: "ok" for success, "fail" for error |
| code    | String | M        | Contains success or error result code      |
| message | String | O        | Result description                         |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | The menu has been successfully accepted                   |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 8. Trigger Sync Menu (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/menus/trigger-sync`  
**Direction:** LM → Partner

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type   | Required | Description                              |
| --------- | ------ | -------- | ---------------------------------------- |
| partnerId | String | M        | Unique ID of the partner                 |
| storeId   | String | M        | Unique ID of the partner store or branch |

### Response Body

| Name    | Type   | Required | Description                                    |
| ------- | ------ | -------- | ---------------------------------------------- |
| status  | String | M        | Result: "ok" for success, "fail" for unsuccess |
| code    | String | M        | Contains success or error result code          |
| message | String | O        | Result description                             |

### Response Codes

| HTTP Status | Description                                               |
| ----------- | --------------------------------------------------------- |
| 200         | Trigger sync is successfully trigger                      |
| 400         | The request is malformed or missing mandatory information |
| 401         | Invalid authorization information                         |
| 404         | Invalid partner ID and/or store ID                        |
| 500         | The application has experienced an internal problem       |

---

## 9. Order Update Notification (Webhook)

**Resource:** `PUT /v1/partners/{partnerId}/stores/{storeId}/orders`  
**Direction:** LM → Partner

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type   | Required | Description                              |
| --------- | ------ | -------- | ---------------------------------------- |
| partnerId | String | M        | Unique ID of the partner                 |
| storeId   | String | M        | Unique ID of the partner store or branch |

### Request Body

| Name                   | Type        | Required | Description                                                                              |
| ---------------------- | ----------- | -------- | ---------------------------------------------------------------------------------------- |
| orderId                | String(20)  | M        | Order Unique ID in format `LMF-yyMMdd-{generated number}`                                |
| orderShortCode         | String(4)   | M        | Short order ID, which is the last four digits of orderId                                 |
| restaurantRevenue      | double      | M        | The total restaurant revenue of the order already subtracts partner subsidiary discounts |
| orderAcceptedTime      | String      | M        | The order accepted time in ISO 8601 Format                                               |
| orderUpdatedTime       | String      | M        | The order updated time in ISO 8601 Format                                                |
| items                  | Array       | M        | List of the ordered menu items                                                           |
| items[].id             | String(255) | M        | Menu item id                                                                             |
| items[].quantity       | int         | M        | Menu Item quantity                                                                       |
| items[].unitPrice      | double      | M        | Final Menu Item Price (THB)                                                              |
| items[].memo           | String      | O        | Menu Item memo requesting from customer                                                  |
| items[].promotionId    | String(255) | O        | Unique promotion id if this item has a promotion                                         |
| items[].discount       | double      | O        | Discount amount from a promotion                                                         |
| items[].properties     | Array       | O        | Options for the item                                                                     |
| ...properties[].id     | String(255) | M        | Option id                                                                                |
| ...properties[].values | Array       | M        | Selected options for the item                                                            |
| ...values[].id         | String(255) | M        | Selected option id                                                                       |
| ...values[].price      | double      | M        | Selected option price (THB)                                                              |
| additionalItems[].name | String      | M        | Additional message for order                                                             |
| memberId               | String(255) | O        | The user member ID of the partner's system                                               |
| customerType           | String(32)  | M        | Order method: DELIVERY or PICKUP                                                         |

### Response Body

| Name    | Type   | Required | Description                                    |
| ------- | ------ | -------- | ---------------------------------------------- |
| status  | String | M        | Result: "ok" for success, "fail" for unsuccess |
| code    | String | M        | Contains success or error result code          |
| message | String | O        | Result description                             |

### Response Codes

| HTTP Status | Description                                                                          |
| ----------- | ------------------------------------------------------------------------------------ |
| 200         | The order is successfully sent to the store                                          |
| 400         | The request is malformed or missing mandatory information                            |
| 401         | Invalid authorization information                                                    |
| 404         | Invalid partner ID and/or store ID                                                   |
| 409         | An order with the same Order ID has already been successfully submitted to the store |
| 500         | The application has experienced an internal problem                                  |

---

## 10. Order Status Update Notification (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/order/status`  
**Direction:** LM → Partner

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type   | Required | Description                              |
| --------- | ------ | -------- | ---------------------------------------- |
| partnerId | String | M        | Unique ID of the partner                 |
| storeId   | String | M        | Unique ID of the partner store or branch |

### Request Body

| Name        | Type       | Required | Description                                               |
| ----------- | ---------- | -------- | --------------------------------------------------------- |
| orderId     | String(20) | M        | Order Unique ID in format `LMF-yyMMdd-{generated number}` |
| orderStatus | String     | M        | Order status: FINISH or CANCELED                          |

### Response Body

| Name    | Type   | Required | Description                                    |
| ------- | ------ | -------- | ---------------------------------------------- |
| status  | String | M        | Result: "ok" for success, "fail" for unsuccess |
| code    | String | M        | Contains success or error result code          |
| message | String | O        | Result description                             |

### Response Codes

| HTTP Status | Description                                                                          |
| ----------- | ------------------------------------------------------------------------------------ |
| 200         | The order is successfully sent to the store                                          |
| 400         | The request is malformed or missing mandatory information                            |
| 401         | Invalid authorization information                                                    |
| 404         | Invalid partner ID and/or store ID                                                   |
| 409         | An order with the same Order ID has already been successfully submitted to the store |
| 500         | The application has experienced an internal problem                                  |

---

## 11. Menu Sync Notification (Webhook)

**Resource:** `POST /v1/partners/{partnerId}/stores/{storeId}/menus/notification`  
**Direction:** LM → Partner

### Header Parameters

| Name          | Type   | Required | Value                 |
| ------------- | ------ | -------- | --------------------- |
| Content-Type  | String | M        | application/json      |
| Authorization | String | M        | Bearer {access_token} |

### Path Parameters

| Name      | Type   | Required | Description                              |
| --------- | ------ | -------- | ---------------------------------------- |
| partnerId | String | M        | Unique ID of the partner                 |
| storeId   | String | M        | Unique ID of the partner store or branch |

### Request Body

| Name              | Type        | Required | Description                                              |
| ----------------- | ----------- | -------- | -------------------------------------------------------- |
| menuSyncRequestId | String(255) | M        | Unique ID to which the menu sync request                 |
| updatedAt         | String      | M        | The update time in ISO 8601 Format                       |
| status            | String      | M        | Result of the menu sync job: SUCCESS or FAILED           |
| error             | String      | O        | Error description if status is FAILED (empty if SUCCESS) |

### Response Body

| Name    | Type   | Required | Description                                    |
| ------- | ------ | -------- | ---------------------------------------------- |
| status  | String | M        | Result: "ok" for success, "fail" for unsuccess |
| code    | String | M        | Contains success or error result code          |
| message | String | O        | Result description                             |

### Response Codes

| HTTP Status | Description                                                                          |
| ----------- | ------------------------------------------------------------------------------------ |
| 200         | The order is successfully sent to the store                                          |
| 400         | The request is malformed or missing mandatory information                            |
| 401         | Invalid authorization information                                                    |
| 404         | Invalid partner ID and/or store ID                                                   |
| 409         | An order with the same Order ID has already been successfully submitted to the store |
| 500         | The application has experienced an internal problem                                  |

---

## API 概览

### Partner → LM (主动调用)

| API                           | Method | Endpoint                                                                |
| ----------------------------- | ------ | ----------------------------------------------------------------------- |
| Authentication                | POST   | `/v1/oauth/token`                                                       |
| Menu Sync v1                  | PUT    | `/v1/partners/{partnerId}/stores/{storeId}/menus`                       |
| Menu Sync v2                  | PUT    | `/v2/partners/{partnerId}/stores/{storeId}/menus`                       |
| Update Menu Item Status       | PUT    | `/v1/partners/{partnerId}/stores/{storeId}/menu/items/status`           |
| Update Property Values Status | PUT    | `/v1/partners/{partnerId}/stores/{storeId}/menu/property/values/status` |
| Force Close/Open Restaurant   | PUT    | `/v1/partners/{partnerId}/stores/{storeId}/restaurant/availability`     |

### LM → Partner (Webhook 回调)

| API                       | Method | Endpoint                                                       |
| ------------------------- | ------ | -------------------------------------------------------------- |
| Place Order               | POST   | `/v1/partners/{partnerId}/stores/{storeId}/orders`             |
| Order Update Notification | PUT    | `/v1/partners/{partnerId}/stores/{storeId}/orders`             |
| Order Status Update       | POST   | `/v1/partners/{partnerId}/stores/{storeId}/order/status`       |
| Trigger Sync Menu         | POST   | `/v1/partners/{partnerId}/stores/{storeId}/menus/trigger-sync` |
| Menu Sync Notification    | POST   | `/v1/partners/{partnerId}/stores/{storeId}/menus/notification` |

---

## 字段必填标识说明

| 标识 | 含义                   |
| ---- | ---------------------- |
| M    | Mandatory (必填)       |
| O    | Optional (可选)        |
| C    | Conditional (条件必填) |
