> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Grab 订单状态推送 Webhook 实现 需求文档

> 本文档定义 Grab 订单状态推送 Webhook 控制器实现的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                   |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/push-order-state-implementation.md](../../../../team/proposals/2025-12/push-order-state-implementation.md) |
| **创建日期**      | 2025-12-19                                                                                                                             |
| **负责人**        | -                                                                                                                                      |
| **目标 Sprint**   | -                                                                                                                                      |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                                                                                                    |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已通过     |
| **审核人**   | -          |
| **审核日期** | 2025-12-19 |
| **审核意见** | -          |

---

## 📋 概述

实现 `grab_v1_push_order_state.go` 控制器，接收 Grab 平台推送的订单状态变更通知。当前控制器返回 `CodeNotImplemented` 错误，需要通过调用现有的 `HandlePushOrderState` 方法完成实现。

同时需要：
1. 统一接口风格，使 `HandlePushOrderState` 接收类型化请求（与 `HandleSubmitOrder` 一致）
2. 扩展 `OrderEvent` 结构体，增加 `ShopUUID` 字段用于 MQ 消息

## 🎯 产品对齐

- 完成 Grab 外卖集成的订单状态同步功能
- 支持订单生命周期管理（接受/拒绝/取消/配送等状态）
- 为门店运营提供实时订单状态更新

## 📝 用户故事

**作为** 外卖订单系统  
**我想** 接收 Grab 平台推送的订单状态变更  
**以便于** 实时更新本地订单状态并通知相关系统

---

## 功能需求

### Requirement 1: 实现 PushOrderState 控制器

**用户故事**: 作为系统，我想接收 Grab 订单状态 Webhook，以便于更新订单状态

#### 验收标准

1. **WHEN** Grab 推送订单状态变更 **THEN** 系统 **SHALL** 调用 `HandlePushOrderState` 处理
2. **IF** 处理成功 **THEN** 系统 **SHALL** 返回空响应体（HTTP 200）
3. **IF** 处理失败 **THEN** 系统 **SHALL** 返回错误信息

#### 具体要求

- [ ] 1.1 控制器调用 `service.Grab().HandlePushOrderState(ctx, req.OrderStateRequest)`
- [ ] 1.2 签名验证由中间件完成，控制器只处理业务逻辑
- [ ] 1.3 成功处理返回 `&v1.PushOrderStateRes{}`

---

### Requirement 2: 统一接口为类型化请求

**用户故事**: 作为开发者，我想保持接口一致性，以便于维护和理解

#### 验收标准

1. **WHEN** 修改接口签名 **THEN** `HandlePushOrderState` **SHALL** 接收 `*grabfood.OrderStateRequest`
2. **WHEN** 接口变更 **THEN** 系统 **SHALL** 执行 `gf gen service` 重新生成接口

#### 具体要求

- [ ] 2.1 修改 `internal/logic/grab_order/grab_order.go` 中 `HandlePushOrderState` 方法签名
- [ ] 2.2 修改 `internal/logic/grab/grab.go` 中代理方法
- [ ] 2.3 执行 `gf gen service` 重新生成 `internal/service/` 接口文件

---

### Requirement 3: 扩展 OrderEvent 结构体

**用户故事**: 作为 MQ 消费者，我想获取订单所属门店信息，以便于路由消息

#### 验收标准

1. **WHEN** 发送 MQ 消息 **THEN** `OrderEvent` **SHALL** 包含 `ShopUUID` 字段
2. **WHEN** 更新订单状态 **THEN** `ShopUUID` **SHALL** 从订单记录中获取

#### 具体要求

- [ ] 3.1 `OrderEvent` 增加 `ShopUUID string` 字段，JSON 标签为 `shopUuid`
- [ ] 3.2 状态更新时从 `order.ShopUuid` 获取值
- [ ] 3.3 同步更新 `HandleSubmitOrder` 中的 MQ 消息（如需要）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Controller → Service → Logic 分层
- **单一职责原则**: 控制器只做请求转发，业务逻辑在 Logic 层
- **遵循规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`

### API 设计要求

- [x] Webhook 端点: `PUT /partner/orders/state`
- [x] 响应格式: 空响应体（符合 Grab API 规范）
- [x] 签名验证: 由 `grab_signature_auth` 中间件处理

### 性能要求

- [ ] Webhook 响应时间 < 500ms
- [ ] 数据库操作使用事务

### 测试要求

- [ ] 单元测试覆盖 `HandlePushOrderState` 方法
- [ ] 测试订单状态流转（ACCEPTED → READY → COLLECTED 等）

---

## 验收标准

### 功能验收

1. **控制器实现**: Grab 订单状态 Webhook 能正确接收并处理
2. **状态更新**: 订单状态正确更新到数据库
3. **MQ 消息**: `OrderEvent` 包含正确的 `ShopUUID` 字段
4. **日志记录**: 状态变更日志正确写入 `order_status_log` 表

### 测试验收

1. **单元测试**: `HandlePushOrderState` 测试通过
2. **集成测试**: Webhook 端到端测试通过

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 `dao/entity/do/` 目录（自动生成）
- 修改 Logic 后执行 `gf gen service` 重新生成 Service 接口
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - Grab SDK 类型定义
- `github.com/gogf/gf/v2` - GoFrame 框架

### 服务依赖

- `grab_signature_auth` 中间件 - Webhook 签名验证
- `order` 表 - 订单数据
- `order_status_log` 表 - 状态日志

---

## 风险和缓解

### 风险 1: 接口变更影响其他调用方

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 通过 grep 搜索确认 `HandlePushOrderState` 无其他调用方
- Service 接口由框架自动生成，执行 `gf gen service` 即可更新

---

## 时间表

- **Phase 1 - 接口修改**: 0.25 天
- **Phase 2 - 控制器实现**: 0.25 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范

### 参考文件

- API 定义: `api/grab/v1/push_order_state.go`
- 服务实现: `internal/logic/grab_order/grab_order.go`
- 类似控制器: `internal/controller/grab/grab_v1_submit_order.go`

### 外部参考

- [GrabFood Order State Webhook](https://developer.grab.com/docs/grabfood/api/#tag/push-order-state-webhook)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-19  
**作者**: AI Assistant  
**审核者**: -

