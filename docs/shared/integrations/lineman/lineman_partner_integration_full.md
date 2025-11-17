# LINE MAN Partner Integration Workflow V2 逐字稿

> 来源：`LINE MAN - Partner Integration Workflow - V2.pdf`

## 1. 议程（Agenda）

- 架构概览（Architecture Overview）
- 认证（Authentication）
- 菜单（Menu）
  - 同步菜单流程（Sync Menu Workflow）
  - 触发菜单同步流程（Trigger Sync Menu Workflow）
  - 菜单同步状态流程（可选，Sync Menu Status Workflow）
- 订单（Order）
  - 同步订单流程（Sync Order Workflow）
  - 同步取消订单流程（可选，Sync Cancel Order Workflow）
  - 订单更新通知流程（可选，Order Update Notification Workflow）
- 门店（Restaurant）
  - 强制关店/开店流程（可选，Force Close/Open Restaurant Workflow）
- API 总结（API Summary）

## 2. API 范围与要求

| 范围           | API                              | 是否必选 | 说明                                  |
| -------------- | -------------------------------- | -------- | ------------------------------------- |
| Authentication | Authentication                   | 必选     | 合作方系统访问 LINE MAN 系统          |
| Menu           | Menu sync                        | 必选     | 合作方将菜单同步至 LINE MAN           |
| Menu           | Trigger sync menu                | 必选     | LINE MAN 请求合作方重新同步菜单       |
| Menu           | Menu sync notification           | 可选     | LINE MAN 通知合作方菜单同步结果       |
| Menu           | Update menu item status          | 可选     | 合作方同步部分菜单项的状态            |
| Menu           | Update menu propertyValue status | 可选     | 合作方同步部分菜单选项状态            |
| Order          | Place order notification         | 必选     | LINE MAN 推送新订单详情               |
| Order          | Status update notification       | 可选     | LINE MAN 推送订单完成/取消状态        |
| Order          | Order update notification        | 可选     | LINE MAN 推送订单编辑后的详情         |
| Restaurant     | Force close/open restaurant      | 可选     | 合作方通过 POS 控制 LINE MAN 门店开关 |

## 3. 架构概览（Architecture Overview）

原文仅展示架构图，强调合作伙伴与 LINE MAN 间的互联互通。

## 4. 认证（Authentication）

- OAuth 2.0 Client Credentials 授权模式。
- 可选：IP 白名单。
- 双方系统均需对彼此进行认证。

## 5. 菜单（Menu）

### 5.1 同步菜单流程（Sync Menu Workflow）

- 菜单同步 API：合作方向 LINE MAN 推送完整菜单快照。
  - 任意菜单或状态变更时触发。
  - 当前仅支持 “推送” 模式。
  - 合作方可随时更新菜单。
  - 门店工作人员仍可在平板上修改（功能保留）。
- 菜单同步通知 API（可选）：LINE MAN 告知合作方同步结果。
  - SUCCESS
  - FAILED

### 5.2 触发菜单同步流程（Trigger Sync Menu Workflow）

- 触发同步菜单 API：LINE MAN 请求合作方重新同步菜单。
- 菜单同步 API：合作方再次推送完整菜单快照。
- 菜单同步通知 API（可选）：返回同步结果。

### 5.3 菜单状态同步流程（可选，Sync Menu Status Workflow）

- 更新菜单项状态 API：合作方同步部分菜单项状态。
  - AVAILABLE（默认）
  - SOLD_OUT_TODAY
  - SUSPENDED
- 更新菜单选项状态 API：合作方同步选项项状态。
  - AVAILABLE（1）
  - SOLD_OUT_TODAY（2）
  - SUSPENDED（3）

## 6. 订单（Order）

### 6.1 同步订单流程（Sync Order Workflow）

- 下单通知 API：LINE MAN 推送新订单详情（价格、商品、订单号等）。
- 状态更新通知 API（可选）：订单完成或取消时推送状态。
- 仍建议保留平板作为兜底接单设备。

### 6.2 取消订单流程（可选，Sync Cancel Order Workflow）

- 状态更新通知 API（可选）：LINE MAN 推送订单完成或取消状态。
- 平板仍作为兜底设备。

### 6.3 订单更新通知流程（可选，Order Update Notification Workflow）

- 订单更新通知 API（可选）：订单编辑后，LINE MAN 推送最新详情。

## 7. 门店（Restaurant）

### 7.1 强制关店/开店流程（可选，Force Close/Open Restaurant Workflow）

- 强制关店/开店 API（可选）：合作方 POS 直接控制 LINE MAN 门店营业状态。

## 8. API 总结（API Summary）

同 “API 范围与要求” 表格，强调 4 个必选、6 个可选接口构成完整对接范围。
