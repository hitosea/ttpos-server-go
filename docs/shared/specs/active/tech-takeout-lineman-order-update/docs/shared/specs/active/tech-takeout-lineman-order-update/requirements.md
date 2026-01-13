# LINE MAN 订单更新 Webhook 需求文档

> 本文档定义 LINE MAN 订单更新 Webhook 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [v2.14.0-lineman-order-update-webhook.md](../../../../team/proposals/2026-01/v2.14.0-lineman-order-update-webhook.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标版本**      | v2.14.0                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {待指定}             |

---

## 📋 概述

实现 LINE MAN 订单更新 Webhook 接口（`PUT /v1/partners/{partnerId}/stores/{storeId}/orders`）。

**核心功能**：
1. 接收 LINE MAN Webhook 请求
2. 查询现有订单并检查幂等性
3. 更新订单数据（事务）
4. 发送 RocketMQ 事件到 Main 模块
5. 返回统一格式响应

**详细需求**: 请参考来源 Proposal 文档。

---

## 验收标准

1. **Webhook 接收**: LINE MAN 发送订单更新 Webhook，TTPOS 成功接收并验证签名
2. **订单更新**: 订单数据成功更新到数据库
3. **幂等性**: 相同 `orderUpdatedTime` 的请求只处理一次
4. **RocketMQ 事件**: 订单更新事件成功发送到 RocketMQ
5. **响应格式**: 返回 LINE MAN 期望的响应格式

---

## 参考资料

- **Proposal**: [v2.14.0-lineman-order-update-webhook.md](../../../../team/proposals/2026-01/v2.14.0-lineman-order-update-webhook.md)
- **LINE MAN API**: [Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=586287212)
- **参考实现**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`

---

**创建日期**: 2026-01-12  
**作者**: rikugun
