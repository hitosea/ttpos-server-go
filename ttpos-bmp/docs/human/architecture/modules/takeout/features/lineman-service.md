# LINE MAN 对接服务

## 概览

LINE MAN 对接服务是 TTPOS 外卖系统中的一个重要服务，负责与 LINE MAN 平台对接，实现外卖订单的创建、取消、状态更新等功能。

## 主要功能

1. 接收 LINE MAN 平台的订单创建请求
2. 处理订单状态更新通知
3. 响应菜单同步请求
4. 提供 OAuth 认证服务
5. 处理订单更新通知

## 服务端点

### 端点列表 提供给Lineman 调用

| Key | Description | Endpoint (Beta) |
| --- | --- | --- |
| Host | Partner's host | `https://ttpos-test1.ttpos.com/` |
| Placing Order Webhook URL | 订单创建端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/partners/{partnerId}/stores/{storeId}/orders` |
| Order Status Webhook URL | 订单状态更新端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/partners/{partnerId}/stores/{storeId}/order/status` |
| Sync menu notification Webhook URL | 菜单同步通知端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/partners/{partnerId}/stores/{storeId}/menus/notification` |
| Sync menu Webhook URL | 菜单同步触发端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/partners/{partnerId}/stores/{storeId}/menus/trigger-sync` |
| Order Update Notification Webhook URL | 订单更新通知端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/partners/{partnerId}/stores/{storeId}/orders/notification` |
| OAuth URL | OAuth 令牌端点 | `https://ttpos-test1.ttpos.com/api/v1/lmwn/oauth2/token` |

### 路径参数说明

- `{partnerId}`: 合作伙伴 ID
- `{storeId}`: 门店 ID