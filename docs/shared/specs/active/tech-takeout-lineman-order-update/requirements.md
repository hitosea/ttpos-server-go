# LINE MAN 订单更新 Webhook 需求文档

> 本文档定义 LINE MAN 订单更新 Webhook 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14.0-lineman-order-update-webhook.md](../../../../team/proposals/2026-01/v2.14.0-lineman-order-update-webhook.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标版本**      | v2.14.0                                                                                                   |
| **目标 Sprint**   | Sprint {待确定}                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {待指定}             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

实现 LINE MAN 订单更新 Webhook 接口（\`PUT /v1/partners/{partnerId}/stores/{storeId}/orders\`），支持 LINE MAN 平台向 TTPOS 推送订单内容修改通知。

**核心价值**：
- 确保 TTPOS 系统中的订单信息与 LINE MAN 平台保持实时同步
- 完善 LINE MAN 订单生命周期管理，支持顾客修改订单场景
- 通过 RocketMQ 可靠传递订单更新事件到 Main 模块

**应用场景**：
- 顾客在配送前修改订单内容（增加/减少商品、修改备注）
- LINE MAN 平台端调整促销折扣，导致订单金额变化
- 发现订单商品信息有误需要修正

## 🎯 产品对齐

该功能属于 **LINE MAN 外卖集成** 的核心能力建设，支持以下产品目标：

1. **外卖平台完整对接**：完善 LINE MAN 订单生命周期管理（创建 → 更新 → 状态变更）
2. **数据一致性保障**：确保多平台订单数据实时同步，减少客诉和退款
3. **商家服务质量**：商家能及时收到订单变更通知，避免备餐错误

## 📝 用户故事

**作为** LINE MAN 外卖平台  
**我想** 通过 Webhook 通知 TTPOS 订单信息已更新  
**以便于** TTPOS 系统能够同步最新的订单数据并通知商家

**作为** 商家  
**我想** 在 POS/KDS 端看到顾客修改后的订单内容  
**以便于** 按照最新订单内容备餐，避免出错


---

## 功能需求

### Requirement 1: 接收订单更新 Webhook

**用户故事**: 作为 TTPOS BMP 系统，我想接收 LINE MAN 的订单更新 Webhook 请求，以便于更新本地订单数据

#### 验收标准

1. **WHEN** LINE MAN 发送订单更新 Webhook（\`PUT /v1/partners/{partnerId}/stores/{storeId}/orders\`）**THEN** BMP **SHALL** 接收并验证请求签名
2. **WHEN** 请求签名验证通过 **THEN** BMP **SHALL** 校验请求参数格式（使用 GoFrame 验证器）
3. **IF** 请求参数缺失或格式错误 **THEN** BMP **SHALL** 返回 \`400 Bad Request\` 和错误描述
4. **IF** 请求签名验证失败 **THEN** BMP **SHALL** 返回 \`401 Unauthorized\`
5. **WHEN** 参数验证通过 **THEN** BMP **SHALL** 调用 Logic 层处理订单更新

#### 具体要求

- [x] 1.1 复用 LINE MAN 认证中间件验证请求签名
- [x] 1.2 使用 \`OrderUpdateReq\` 结构体定义（已在 \`api/lineman/v1/order.go\` 定义）
- [x] 1.3 GoFrame 自动参数验证（基于 \`v\` 标签）
- [x] 1.4 Controller 层只负责接收请求和返回响应，不包含业务逻辑
- [ ] 1.5 记录请求日志（包含 orderId、partnerId、storeId）

---

### Requirement 2: 查询和幂等性检查

**用户故事**: 作为 TTPOS BMP Logic 层，我想查询现有订单并检查幂等性，以便于防止重复处理和旧数据覆盖

#### 验收标准

1. **WHEN** 收到订单更新请求 **THEN** BMP **SHALL** 根据 \`orderId\` 和 \`providerName="lineman"\` 查询现有订单
2. **IF** 订单不存在 **THEN** BMP **SHALL** 返回错误（订单不存在）
3. **IF** 订单存在 **AND** \`orderUpdatedTime\` 晚于现有记录 **THEN** BMP **SHALL** 执行更新
4. **IF** \`orderUpdatedTime\` 早于或等于现有记录 **THEN** BMP **SHALL** 跳过更新（幂等性保护）
5. **WHEN** 跳过更新 **THEN** BMP **SHALL** 记录日志并返回成功（200）

#### 具体要求

- [ ] 2.1 使用 \`dao.Order.Where("provider_order_id = ? AND provider_name = ?")\` 查询订单
- [ ] 2.2 比较 \`order_updated_time\` 字段判断是否需要更新
- [ ] 2.3 幂等性检查逻辑放在 Logic 层，不在 Controller 层
- [ ] 2.4 记录幂等性跳过日志（包含 orderId、现有时间、新时间）

---

### Requirement 3: 更新订单数据

**用户故事**: 作为 TTPOS BMP Logic 层，我想使用事务更新订单数据，以便于确保数据一致性

#### 验收标准

1. **WHEN** 通过幂等性检查 **THEN** BMP **SHALL** 开启数据库事务更新订单
2. **WHEN** 更新订单 **THEN** BMP **SHALL** 更新订单主表（\`order\`）的以下字段：
   - \`total_amount\` = \`restaurantRevenue\`
   - \`subtotal\` = \`restaurantRevenue\`
   - \`order_time\` = \`orderAcceptedTime\`（解析 ISO 8601 格式）
   - \`order_updated_time\` = \`orderUpdatedTime\`（解析 ISO 8601 格式）
   - \`updated_at\` = 当前时间
3. **WHEN** 更新订单主表成功 **THEN** BMP **SHALL** 删除旧的订单明细（\`order_item\` 表）
4. **WHEN** 删除旧明细成功 **THEN** BMP **SHALL** 插入新的订单明细
5. **IF** 任一步骤失败 **THEN** BMP **SHALL** 回滚事务并返回错误
6. **WHEN** 事务提交成功 **THEN** BMP **SHALL** 记录更新成功日志

#### 具体要求

- [ ] 3.1 使用 \`dao.Order.Transaction()\` 开启事务
- [ ] 3.2 订单主表更新使用 \`Where("uuid", orderUUID).Update()\`
- [ ] 3.3 订单明细删除使用 \`Where("order_uuid", orderUUID).Delete()\`
- [ ] 3.4 订单明细插入循环遍历 \`req.Items\`，序列化 \`properties\` 为 JSON
- [ ] 3.5 错误处理：事务失败返回 \`gerror.Wrap(err, "更新订单失败")\`
- [ ] 3.6 记录详细的更新日志（包含 orderUUID、orderId、更新字段）


---

### Requirement 4: 发送 RocketMQ 事件

**用户故事**: 作为 TTPOS BMP Logic 层，我想发送 RocketMQ 事件，以便于通知 Main 模块处理订单更新

#### 验收标准

1. **WHEN** 订单更新成功 **THEN** BMP **SHALL** 构造 \`grab.OrderEvent\` 消息
2. **WHEN** 构造消息 **THEN** BMP **SHALL** 设置以下字段：
   - \`Action\` = \`consts.OrderActionUpdate\`（"update"）
   - \`ProviderName\` = \`"lineman"\`
   - \`ShopUUID\` = \`req.StoreId\`
   - \`OrderUUID\` = 订单 UUID
   - \`OrderID\` = \`req.OrderId\`
   - \`Status\` = \`consts.OrderStatusAccepted\`
   - \`Timestamp\` = 当前时间戳
3. **WHEN** 消息构造完成 **THEN** BMP **SHALL** 使用 \`queue.PushWithContext()\` 发送到 Topic \`takeout_grab_order\`
4. **IF** RocketMQ 发送失败 **THEN** BMP **SHALL** 记录警告日志但不影响主流程（订单已入库）
5. **WHEN** RocketMQ 发送成功 **THEN** BMP **SHALL** 记录成功日志

#### 具体要求

- [ ] 4.1 复用 \`grab.OrderEvent\` 结构体和 Topic（\`takeout_grab_order\`）
- [ ] 4.2 需要先添加常量 \`consts.OrderActionUpdate = "update"\`
- [ ] 4.3 RocketMQ 发送失败只记录日志，不返回错误（依赖 RocketMQ 重试机制）
- [ ] 4.4 记录 RocketMQ 发送日志（包含 Topic、OrderUUID、Action）

---

### Requirement 5: 返回响应

**用户故事**: 作为 TTPOS BMP Controller 层，我想返回统一格式的响应，以便于 LINE MAN 平台识别处理结果

#### 验收标准

1. **WHEN** Logic 层处理成功 **THEN** Controller **SHALL** 返回 \`200 OK\` 和 \`{"status": "ok", "code": "200", "message": "Order updated successfully"}\`
2. **IF** Logic 层返回错误 **THEN** Controller **SHALL** 返回 \`200 OK\` 和 \`{"status": "fail", "code": "500", "message": "{错误信息}"}\`（注意：仍返回 HTTP 200）
3. **WHEN** 返回响应 **THEN** Controller **SHALL** 使用 \`LinemanCommonResData\` 结构体
4. **IF** 订单不存在 **THEN** Controller **SHALL** 返回 \`code: "404"\`
5. **IF** 幂等性跳过 **THEN** Controller **SHALL** 返回成功响应（\`status: "ok"\`）

#### 具体要求

- [ ] 5.1 Controller 层不抛出 HTTP 错误，统一返回 200 和 JSON 响应
- [ ] 5.2 使用 \`OrderUpdateRes\` 结构体（已在 \`api/lineman/v1/order.go\` 定义）
- [ ] 5.3 错误信息使用 \`err.Error()\` 获取
- [ ] 5.4 记录响应日志（包含 orderId、status、code）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层
  - **Controller 层**：接收请求、参数验证、返回响应
  - **Service 层**：接口定义（\`ILinemanOrder.HandleOrderUpdate\`）
  - **Logic 层**：业务逻辑实现（查询、幂等性、更新、RocketMQ）
- **单一职责原则**: 
  - Controller 只负责 HTTP 交互
  - Logic 层封装所有业务逻辑
  - DAO 层只负责数据库操作
- **遵循规范**:
  - \`ttpos-bmp/.cursor/rules/go-rules.mdc\` - Go BMP 开发规范
  - \`.cursor/rules/api.mdc\` - API 设计规范
  - \`.cursor/rules/database.mdc\` - 数据库开发规范

### 数据库设计要求

- [x] 1.1 **新增字段**: \`order\` 表添加 \`order_updated_time\` 字段（TIMESTAMP NULL）
- [x] 1.2 字段定义：\`order_updated_time\` TIMESTAMP NULL DEFAULT NULL COMMENT '订单更新时间（LINE MAN）'
- [x] 1.3 **迁移脚本**: 创建数据库迁移脚本
- [x] 1.4 **Entity 更新**: 运行 \`gf gen dao\` 重新生成实体文件

### 性能要求

- [ ] 订单查询响应时间 < 50ms
- [ ] 订单更新事务完成时间 < 200ms
- [ ] RocketMQ 发送耗时 < 100ms
- [ ] 总体 Webhook 响应时间 < 500ms

### 测试要求

- [ ] **单元测试**: Logic 层测试覆盖率 ≥ 80%
- [ ] **集成测试**: 端到端流程测试通过
- [ ] **手动测试**: Postman + RocketMQ 控制台验证

---

## 验收标准

### 功能验收

1. **Webhook 接收**: LINE MAN 发送订单更新 Webhook，TTPOS 成功接收并验证签名
2. **订单更新**: 订单数据（商品、金额、时间）成功更新到数据库
3. **幂等性**: 相同 \`orderUpdatedTime\` 的请求只处理一次，旧数据被拒绝
4. **RocketMQ 事件**: 订单更新事件成功发送到 RocketMQ（Main 模块可接收）
5. **响应格式**: 返回 LINE MAN 期望的响应格式
6. **错误处理**: 订单不存在、参数错误、签名失败等场景正确处理

---

## 时间表

- **Phase 1 - 数据库迁移和常量定义**: 0.4 天
- **Phase 2 - Service 和 Logic 实现**: 1.4 天
- **Phase 3 - Controller 和测试**: 1.2 天
- **Phase 4 - 文档和发布**: 0.5 天
- **总计**: 3.5 天（SP = 5）

---

## 参考资料

### 核心规范

- \`ttpos-bmp/.cursor/rules/go-rules.mdc\` - Go BMP 开发规范
- \`.cursor/rules/api.mdc\` - API 设计规范
- \`.cursor/rules/database.mdc\` - 数据库开发规范

### 参考实现

- **LINE MAN 订单创建**: \`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go\` - \`HandlePlaceOrder()\`
- **Grab 订单处理**: \`ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go\`
- **OrderEvent 定义**: \`ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/event.go\`

### API 定义

- **OrderUpdateReq/Res**: \`ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go\`（已定义）
- **LINE MAN API 规范**: [Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=586287212#gid=586287212)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**审核者**: {待指定}
