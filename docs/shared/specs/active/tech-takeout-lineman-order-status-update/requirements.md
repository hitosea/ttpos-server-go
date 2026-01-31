# Lineman 订单状态更新 Webhook 需求文档

> 本文档定义 LINE MAN 订单状态更新 Webhook 功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14.0-lineman-order-status-update-webhook.md](../../../../team/proposals/2026-01/v2.14.0-lineman-order-status-update-webhook.md) |
| **创建日期**      | 2026-01-13                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | 待指定             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

实现 LINE MAN 订单状态更新 Webhook 接口，接收 LINE MAN 平台发送的订单完成（FINISH）或取消（CANCELED）通知，将状态同步到 TTPOS 系统，并通过 RocketMQ 消息队列通知下游系统（Main 模块）处理订单状态变更。

**核心价值**：
- 确保 TTPOS 系统订单状态与 LINE MAN 平台保持实时同步
- 完善 LINE MAN 订单生命周期管理（创建 → 内容更新 → 状态更新）
- 自动化订单状态流转，减少人工干预和对账工作
- 支持订单报表和数据分析的准确性

## 🎯 产品对齐

该功能支持以下产品目标：
1. **外卖平台集成完整性**：完成 LINE MAN 订单生命周期的最后一环（状态更新）
2. **运营效率提升**：自动化订单状态同步，减少商家手动查询和对账工作量
3. **数据准确性**：保证订单状态数据的实时性和准确性，支持经营分析
4. **系统可靠性**：通过消息队列确保状态变更事件可靠传递到各业务模块

## 📝 用户故事

**作为** LINE MAN 外卖平台  
**我想** 通过 Webhook 通知 TTPOS 订单状态已变更（完成或取消）  
**以便于** TTPOS 系统能够同步最新的订单状态并更新商家端显示

**作为** 商家  
**我想** 在 POS/Shop 端自动看到订单完成或取消状态  
**以便于** 无需手动查询 LINE MAN 平台即可了解订单最终状态

---

## 功能需求

### Requirement 1: 接收订单状态更新 Webhook

**用户故事**: 作为 TTPOS 系统，我想接收 LINE MAN 的订单状态更新通知，以便于及时更新本地订单状态。

#### 验收标准

1. **WHEN** LINE MAN 发送 `POST /v1/partners/{partnerId}/stores/{storeId}/order/status` 请求 **THEN** TTPOS **SHALL** 成功接收并验证请求参数
2. **WHEN** 请求参数完整且有效 **THEN** TTPOS **SHALL** 解析 `orderId` 和 `orderStatus` 字段
3. **WHEN** 请求参数缺失或格式错误 **THEN** TTPOS **SHALL** 返回 `400` 错误和错误描述
4. **WHEN** 请求签名无效 **THEN** TTPOS **SHALL** 返回 `401` 错误

#### 具体要求

- [x] 1.1 实现 Controller 层接收 Webhook 请求（`lineman_v1_order_status_update.go`）
- [x] 1.2 使用 GoFrame 验证器自动验证请求参数（`OrderStatusUpdateReq`）
- [x] 1.3 调用 Service 层接口 `HandleOrderStatusUpdate(ctx, req)`
- [x] 1.4 返回统一格式响应（`LinemanCommonResData`）

---

### Requirement 2: 状态映射转换

**用户故事**: 作为 TTPOS 系统，我想将 LINE MAN 的状态映射为内部状态，以便于保持状态命名的一致性。

#### 验收标准

1. **WHEN** 收到状态为 `FINISH` **THEN** TTPOS **SHALL** 映射为内部状态 `COMPLETED`
2. **WHEN** 收到状态为 `CANCELED` **THEN** TTPOS **SHALL** 保持状态为 `CANCELED`
3. **WHEN** 收到未知状态 **THEN** TTPOS **SHALL** 保持原状态不变并记录警告日志

#### 具体要求

- [x] 2.1 实现状态映射函数 `mapLinemanStatusToTTPOS(linemanStatus string) string`
- [x] 2.2 定义 LINE MAN 状态常量（`LinemanStatusFinish` / `LinemanStatusCanceled`）
- [x] 2.3 在 Logic 层中调用状态映射函数
- [x] 2.4 记录状态映射日志（调试用）

---

### Requirement 3: 更新订单状态到数据库

**用户故事**: 作为 TTPOS 系统，我想将订单状态更新到数据库，以便于保持数据的准确性和一致性。

#### 验收标准

1. **WHEN** 订单存在 **THEN** TTPOS **SHALL** 更新订单状态为映射后的 TTPOS 状态
2. **WHEN** 订单不存在 **THEN** TTPOS **SHALL** 返回 `404` 错误和 `"订单不存在"` 消息
3. **WHEN** 订单当前状态已经是目标状态 **THEN** TTPOS **SHALL** 跳过更新并返回成功（幂等性）
4. **WHEN** 订单状态更新成功 **THEN** TTPOS **SHALL** 同时更新 `updated_at` 字段为当前时间戳
5. **IF** 数据库更新失败 **THEN** TTPOS **SHALL** 返回 `500` 错误和错误描述

#### 具体要求

- [x] 3.1 根据 `provider_order_id` 和 `provider_name = "lineman"` 查询订单
- [x] 3.2 检查订单是否存在，不存在则返回错误
- [x] 3.3 检查当前状态是否与目标状态相同（幂等性检查）
- [x] 3.4 更新 `status` 字段为映射后的状态
- [x] 3.5 更新 `updated_at` 字段为当前时间戳（`gtime.Now().Unix()`）
- [x] 3.6 记录订单状态更新日志（Info 级别）

---

### Requirement 4: 发送 RocketMQ 事件通知

**用户故事**: 作为 TTPOS BMP 模块，我想通过 RocketMQ 发送订单状态变更事件，以便于 Main 模块能够接收并处理订单状态变更。

#### 验收标准

1. **WHEN** 订单状态更新成功 **THEN** TTPOS **SHALL** 构造 `OrderEvent` 消息
2. **WHEN** 构造 `OrderEvent` **THEN** TTPOS **SHALL** 设置以下字段：
   - `action`: `"status_update"`
   - `providerName`: `"lineman"`
   - `status`: 映射后的 TTPOS 状态（`COMPLETED` 或 `CANCELED`）
   - `orderUUID`: 订单 UUID
   - `orderID`: LINE MAN 订单 ID
   - `shopUUID`: 商店 UUID
   - `timestamp`: 当前时间戳
3. **WHEN** 发送 RocketMQ 消息 **THEN** TTPOS **SHALL** 使用 Topic `takeout_grab_order`
4. **IF** RocketMQ 发送失败 **THEN** TTPOS **SHALL** 记录警告日志但仍返回成功（订单状态已更新）

#### 具体要求

- [x] 4.1 复用 `grab.OrderEvent` 结构体
- [x] 4.2 设置 `action` 为 `"status_update"` 或 `consts.OrderActionStatusUpdate`
- [x] 4.3 设置 `providerName` 为 `consts.ProviderLineman`
- [x] 4.4 设置 `status` 为映射后的 TTPOS 状态
- [x] 4.5 使用 `queue.PushWithContext(ctx, "takeout_grab_order", event)` 发送消息
- [x] 4.6 捕获 RocketMQ 发送错误并记录日志（不抛出异常）

---

### Requirement 5: 返回统一格式响应

**用户故事**: 作为 LINE MAN 平台，我想收到 TTPOS 的标准格式响应，以便于判断请求是否成功处理。

#### 验收标准

1. **WHEN** 订单状态更新成功 **THEN** TTPOS **SHALL** 返回 HTTP 200 和 `{"status": "ok", "code": "200", "message": "Order status updated successfully"}`
2. **WHEN** 订单不存在 **THEN** TTPOS **SHALL** 返回 HTTP 200 和 `{"status": "fail", "code": "404", "message": "订单不存在"}`
3. **WHEN** 数据库更新失败 **THEN** TTPOS **SHALL** 返回 HTTP 200 和 `{"status": "fail", "code": "500", "message": "{错误描述}"}`
4. **WHEN** 请求参数错误 **THEN** TTPOS **SHALL** 返回 HTTP 400 和错误描述

#### 具体要求

- [x] 5.1 成功时返回 `{"status": "ok", "code": "200"}`
- [x] 5.2 失败时返回 `{"status": "fail", "code": "{错误码}", "message": "{错误描述}"}`
- [x] 5.3 Controller 层捕获 Logic 层错误并转换为响应格式
- [x] 5.4 符合 LINE MAN API 规范的响应格式

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic → DAO 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 层独立实现业务逻辑，可复用
- **依赖管理**: Controller 只依赖 Service 接口，Logic 层只依赖 DAO
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - ttpos-bmp 项目专用规范
  - `.cursor/rules/api.mdc` - API 设计规范

### API 设计要求

- [x] 端点：`POST /v1/partners/{partnerId}/stores/{storeId}/order/status`
- [x] 请求参数：`orderId` (String), `orderStatus` (String: FINISH/CANCELED)
- [x] 响应格式：`{status: "ok/fail", code: "200/404/500", message: "..."}`
- [x] 符合 LINE MAN API 规范
- [x] 参考: [LINE MAN API 定义](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=102046225#gid=102046225)

### 数据库设计要求

- [x] 无需修改表结构（复用现有 `takeout_order` 表的 `status` 字段）
- [x] 更新 `status` 字段为映射后的状态
- [x] 更新 `updated_at` 字段为当前时间戳（int 类型）
- [x] 查询使用 `provider_order_id` 和 `provider_name` 组合索引

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询使用索引优化
- [x] 幂等性检查在内存中完成（不额外查询数据库）
- [x] RocketMQ 发送异步处理（不阻塞主流程）

### 测试要求

- [x] Logic 层测试覆盖率 ≥ 70%
- [x] 单元测试覆盖所有状态映射场景（FINISH → COMPLETED, CANCELED → CANCELED）
- [x] 单元测试覆盖幂等性场景（重复请求）
- [x] 单元测试覆盖错误场景（订单不存在、数据库错误）
- [x] 集成测试覆盖完整 Webhook 流程（接收 → 更新 → MQ 发送）

### 国际化要求

- [x] 错误消息使用中文（日志和响应）
- [ ] 暂不需要支持多语言（内部接口）

### 安全要求

- [x] 复用 Lineman 认证中间件验证请求签名
- [x] 使用 GoFrame 验证器校验请求参数
- [x] 记录所有请求日志（Info 级别）
- [x] 记录所有错误日志（Error 级别）

### 可靠性要求

- [x] 幂等性设计（重复请求不重复更新）
- [x] RocketMQ 发送失败不影响主流程（订单状态已更新）
- [x] 事务管理（使用 DAO 层事务，保证数据一致性）
- [x] 错误日志记录（使用 `g.Log()`）

---

## 验收标准

### 功能验收

1. **状态映射正确**: LINE MAN `FINISH` 成功映射为 TTPOS `COMPLETED`，`CANCELED` 保持不变
2. **订单状态更新**: 订单状态在数据库中正确更新，`updated_at` 字段同步更新
3. **RocketMQ 事件发送**: Main 模块能够接收到 `OrderEvent` 消息，字段正确
4. **幂等性处理**: 重复请求不重复更新订单状态，返回成功
5. **错误处理**: 订单不存在返回 404，数据库错误返回 500

### 测试验收

1. **单元测试**: Logic 层测试覆盖率 ≥ 70%，所有状态映射和错误场景覆盖
2. **集成测试**: 使用 Postman 模拟 LINE MAN Webhook 请求，验证完整流程
3. **手动测试**: 
   - 发送 `FINISH` 状态，验证数据库状态为 `COMPLETED`
   - 发送 `CANCELED` 状态，验证数据库状态为 `CANCELED`
   - 重复发送相同状态，验证幂等性（不重复更新）
   - 发送不存在的 `orderId`，验证返回 404

### 文档验收

1. **技术文档**: design.md 完整且准确（待 `/spec-design` 创建）
2. **API 文档**: 更新 LINE MAN API 集成文档
3. **测试文档**: tasks.md 中的测试任务完成（待创建）
4. **故障排查**: 创建 troubleshooting 文档（如有必要）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- Service 接口定义在 `service/` 目录
- Logic 实现在 `logic/` 目录
- Controller 在 `controller/` 目录
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### 代码规范

- 所有代码使用中文注释
- 不使用 panic，返回 error
- 使用 `g.Log()` 记录日志
- 错误使用 `gerror.Wrap()` 包装
- 上下文使用 `context.Context` 传递

### 业务约束

- 只处理 LINE MAN 订单（`provider_name = "lineman"`）
- 只支持 `FINISH` 和 `CANCELED` 两种状态
- 订单状态更新后不可回退（终态）

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2-3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `GoFrame 2.x` - Web 框架
- `RocketMQ` - 消息队列
- `MySQL 8.0+` - 数据库
- `ttpos-bmp/internal/dao` - 数据库访问层
- `ttpos-bmp/internal/model/dto/grab` - OrderEvent 结构体
- `ttpos-bmp/api/lineman/v1` - API 请求/响应结构体

### 服务依赖

- **BMP → Main**: RocketMQ 消息（订单状态变更事件）
- **LINE MAN → BMP**: Webhook 回调（订单状态通知）

### 业务依赖

- LINE MAN 订单创建功能（PlaceOrder）已实现
- LINE MAN OAuth 认证功能已实现
- Main 模块订单状态监听器已实现（监听 RocketMQ）

---

## 风险和缓解

### 风险 1: 状态映射错误导致订单状态不一致

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 编写完整的单元测试覆盖所有状态映射场景
- 使用常量定义状态，避免硬编码
- Code Review 严格检查状态映射逻辑

### 风险 2: RocketMQ 消息丢失导致 Main 模块无法感知状态变更

**影响**: 中  
**概率**: 低  
**缓解措施**:

- RocketMQ 自带重试机制，确保消息最终送达
- 发送失败只记录日志，不影响主流程（订单状态已更新）
- Main 模块可以通过定时任务主动查询订单状态（兜底机制）

### 风险 3: LINE MAN 重复发送状态通知导致重复处理

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 实现幂等性检查（检查当前状态是否与目标状态相同）
- 如果状态未变化，跳过更新并返回成功
- 记录日志便于排查

### 风险 4: 订单不存在导致状态更新失败

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 返回明确的 404 错误码和错误描述
- 记录错误日志便于排查
- LINE MAN 会根据 404 响应进行相应处理

---

## 时间表

- **Phase 1 - 常量和接口定义**: 0.2 天
  - 添加 `OrderActionStatusUpdate` 常量
  - 添加 `HandleOrderStatusUpdate` 接口方法
- **Phase 2 - Logic 层实现**: 0.8 天
  - 实现状态映射函数
  - 实现订单查询和更新逻辑
  - 实现 RocketMQ 事件发送
  - 实现幂等性检查
- **Phase 3 - Controller 层实现**: 0.1 天
  - 实现 `OrderStatusUpdate()` 方法
  - 调用 Service 层
  - 封装响应格式
- **Phase 4 - 测试和文档**: 0.9 天
  - 单元测试（0.3 天）
  - 集成测试（0.3 天）
  - 更新文档（0.3 天）
- **总计**: 2.0 天（SP = 2-3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - ttpos-bmp 项目专用规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- [LINE MAN API 规范 - Order Status Update Notification API](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=102046225#gid=102046225)

### 相关实现

- LINE MAN 订单创建：`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` - `HandlePlaceOrder()`
- LINE MAN 订单更新：`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` - `HandleOrderUpdate()`
- Grab 订单状态处理：`ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go`
- OrderEvent 定义：`ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/event.go`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-13.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-13  
**作者**: rikugun  
**审核者**: 待指定
