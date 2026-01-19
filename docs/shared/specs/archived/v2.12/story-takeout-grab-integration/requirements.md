> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# GrabFood 外卖平台对接 (v1.1.3) 需求文档

> 本文档定义 GrabFood 对接功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-grab-integration.md](../../../../team/proposals/2025-12/takeout-grab-integration.md) |
| **创建日期**      | 2025-12-04                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | TBD                      |
| **审核日期** | -                        |

---

## 📋 概述

本功能旨在通过 `ttpos-takeout` 模块实现与 GrabFood 外卖平台 (API v1.1.3) 的对接。通过实现菜单同步、订单自动接收（Webhook + Persistence + MQ）和门店状态管理，帮助商家拓展线上销售渠道。
**注意**：本期重点在于 `ttpos-takeout` 模块的完整性（包括数据持久化），暂不包含 `ttpos-erp` (Main 模块) 的消费逻辑。

## 🎯 产品对齐

该功能直接支持公司“餐饮一体化”战略，补齐了在东南亚市场主流外卖平台对接的缺口。

## 📝 用户故事

**作为** 收银员
**我想** 直接在 POS 收到 GrabFood 的订单（最终效果）
**以便于** 不需要单独操作外卖平板

**作为** 系统管理员
**我想** 确保 Grab 订单数据在 Takeout 模块有独立备份
**以便于** 在 ERP 系统故障时也能查询原始订单记录

---

## 功能需求

### Requirement 1: 菜单同步 (Menu Sync)

**用户故事**: 作为 店长，我想 将 POS 的菜单同步到 GrabFood，以便于 保持线上线下菜单一致。

#### 验收标准

1.  **WHEN** 店长在后台点击“同步菜单” **THEN** 系统 **SHALL** 将 POS 菜品结构转换为 GrabFood 格式并调用 API 推送。
2.  **IF** 推送成功 **THEN** 系统 **SHALL** 提示“同步已提交，等待平台处理”。
3.  **WHEN** 收到 GrabFood 的菜单审核回调 **THEN** 系统 **SHALL** 更新同步状态并在日志中记录。

#### 具体要求

- [ ] 1.1 实现 POS Category/Product/Modifier 到 GrabFood Menu 结构的映射。
- [ ] 1.2 处理 GrabFood 的 `PUT /partner/v1/menu` 接口调用。
- [ ] 1.3 接收并处理 `POST /partner/v1/menu/notify` Webhook 回调。

---

### Requirement 2: 订单接收与持久化 (Order Processing & Persistence)

**用户故事**: 作为 系统，我想 持久化存储 GrabFood 订单数据，以便于 审计和异步处理。

#### 验收标准

1.  **WHEN** GrabFood 推送新订单 (Webhook) **THEN** `ttpos-takeout` **SHALL** 接收并返回 200 OK。
2.  **THEN** `ttpos-takeout` **SHALL** 将订单详情保存到 `takeout_order` 及相关子表中。
3.  **THEN** `ttpos-takeout` **SHALL** 将标准化的 `ThirdPartyOrderEvent` 发送到 RocketMQ。

#### 具体要求

- [ ] 2.1 实现 `POST /partner/v1/orders` Webhook 接口，进行签名验证 (HMAC)。
- [ ] 2.2 设计并创建 `takeout_order`, `takeout_order_item`, `takeout_order_modifier` 等数据库表。
- [ ] 2.3 将 Grab 订单数据完整保存到上述数据库表中。
- [ ] 2.4 定义 RocketMQ Topic `takeout_order` 和消息结构，发送 MQ 消息。
- [ ] 2.5 (Out of Scope) `ttpos-erp` 消费逻辑暂不实现。

---

### Requirement 3: 门店状态管理 (Store Status)

**用户故事**: 作为 店长，我想 在忙碌时暂停 GrabFood 接单，以便于 缓解后厨压力。

#### 验收标准

1.  **WHEN** 店长在 POS 设置“暂停外卖” **THEN** 系统 **SHALL** 调用 Grab API 暂停门店营业。
2.  **WHEN** 店长恢复营业 **THEN** 系统 **SHALL** 调用 Grab API 恢复门店营业。

#### 具体要求

- [ ] 3.1 实现 `PUT /partner/v1/merchants/{merchantID}/store/status` 接口调用。

---

## 非功能需求

### 代码架构和模块化

- **模块划分**: `ttpos-takeout` 独立负责外卖平台对接和数据存储。
- **数据一致性**: 确保 Webhook 接收后先落库再发 MQ。

### 数据库设计要求

- [ ] 新增 `takeout_order` 表及其子表，存储 Grab 完整订单信息。
- [ ] 字段需覆盖 Grab API 返回的关键信息（价格、税费、折扣、备注等）。

### 性能要求

- [ ] Webhook 响应必须在 3s 内完成。
- [ ] 数据库写入应高效。

### 安全要求

- [ ] 必须验证 Grab Webhook 签名。

---

## 验收标准

### 功能验收

1.  **数据落库**: 模拟 Grab 下单，数据库中能查到完整的订单和菜品信息。
2.  **MQ 发送**: RocketMQ Console 能看到符合格式的订单消息。
3.  **API 响应**: Grab 侧收到 200 OK。

### 测试验收

1.  **单元测试**: 覆盖签名验证、数据落库逻辑。
2.  **集成测试**: 模拟完整 Webhook 流程，验证 DB 和 MQ。

---

## 约束条件

### 技术约束

#### Go BMP 模块 (ttpos-takeout)
- 使用 GoFrame 2.x。
- 使用 RocketMQ。
- 使用 MySQL 存储订单数据。

### 业务约束
- 暂不处理与 ERP 的交互细节。

---

## 时间表

- **Phase 1 - 基础框架 & 菜单同步**: 3 天
- **Phase 2 - 订单接收与持久化**: 3 天
- **Phase 3 - 订单确认与状态管理**: 2 天
- **总计**: 8 天

---

## 参考资料

- [GrabFood API Docs](https://developer.grab.com/docs/grabfood/api/v1-1-3/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.1.0
**创建日期**: 2025-12-04
**作者**: rikugun
