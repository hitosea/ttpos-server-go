# LINE MAN PlaceOrder 订单接收功能 需求文档

> 本文档定义 LINE MAN PlaceOrder 订单接收功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14.0-lineman-place-order.md](../../../../team/proposals/2026-01/v2.14.0-lineman-place-order.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **目标版本**      | v2.14.0                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-12             |
| **审核意见** | 需求明确，字段映射参考文档完整，可以进入技术设计阶段         |

---

## 📋 概述

LINE MAN 是泰国主流的外卖配送平台之一，目前 ttpos-takeout 模块已完成 LINE MAN 的菜单同步、OAuth 认证等基础功能，但订单接收功能尚未实现。本需求旨在实现 LINE MAN 的 PlaceOrder Webhook 接口，接收并处理 LINE MAN 平台推送的订单数据，将订单保存到 TTPOS 系统并通知 Main 模块进行后续处理。

**业务价值**：
- 完善 LINE MAN 平台的完整订单流程，从菜单同步到订单接收
- 订单自动同步，减少人工操作，降低错误率
- 支持泰国主流外卖平台，增强产品竞争力
- 复用现有 Grab 订单处理架构，降低开发和维护成本

## 🎯 产品对齐

本功能支持 TTPOS 在泰国市场的外卖业务扩展战略，通过集成 LINE MAN 平台，帮助商户：
- 在一个系统中统一管理多个外卖渠道的订单
- 提升订单处理效率，缩短顾客等待时间
- 减少人工录入订单的错误和遗漏
- 提供完整的外卖业务闭环（菜单→订单→配送→结算）

## 📝 用户故事

**作为** 商户管理员  
**我想** LINE MAN 平台的订单能够自动同步到 TTPOS 系统  
**以便于** 厨房和收银人员能够及时处理订单，提升运营效率

**作为** LINE MAN 平台  
**我想** 通过 PlaceOrder Webhook 将新订单推送到 TTPOS  
**以便于** 商户能够在 TTPOS 系统中统一管理所有渠道的订单

---

## 功能需求

### Requirement 1: 订单数据接收

**用户故事**: 作为 LINE MAN 平台，我想通过 Webhook 将订单推送到 TTPOS，以便于商户能够接收订单

#### 验收标准

1. **WHEN** LINE MAN 平台发送 PlaceOrder 请求（包含完整的订单数据）**THEN** 系统 **SHALL** 成功接收请求并验证参数完整性
2. **IF** 请求参数缺失或格式错误 **THEN** 系统 **SHALL** 返回参数错误（HTTP 400）并记录详细日志
3. **WHEN** 请求验证通过 **THEN** 系统 **SHALL** 提取核心字段（partnerId、storeId、orderId、items 等）

#### 具体要求

- [ ] 1.1 实现 PlaceOrder Webhook 接口（POST `/partners/:partnerId/stores/:storeId/orders`）
- [ ] 1.2 验证路径参数（partnerId、storeId）和请求体参数（orderId、items 等）
- [ ] 1.3 参数验证失败时返回标准错误响应，包含错误码和详细说明
- [ ] 1.4 记录完整的请求日志，包含请求体、IP 地址、时间戳等

---

### Requirement 2: 数据模型转换

**用户故事**: 作为系统开发者，我想将 LINE MAN 订单数据转换为 TTPOS 统一订单模型，以便于系统能够统一处理多渠道订单

> **⚠️ 重要参考文档**  
> 字段映射设计必须参考官方映射文档：[Lineman API定义及TTPOS 映射 - Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)  
> 该文档包含 LINE MAN 字段与 Grab 字段的完整对照关系，以及 TTPOS 字段的映射说明。

#### 验收标准

1. **WHEN** 接收到 LINE MAN 订单数据 **THEN** 系统 **SHALL** 将订单字段正确映射到 TTPOS 订单模型
2. **WHEN** 转换订单金额字段 **THEN** 系统 **SHALL** 正确计算和存储 total_amount、subtotal 等金额
3. **WHEN** 转换商品明细 **THEN** 系统 **SHALL** 遍历 items 数组并转换为 order_items 记录
4. **WHEN** 处理商品选项（properties）**THEN** 系统 **SHALL** 序列化为 JSON 格式存储

#### 具体要求

- [ ] 2.1 实现订单主表字段映射（参考下表）
- [ ] 2.2 实现订单明细字段映射（包含商品 ID、名称、数量、单价、选项等）
- [ ] 2.3 将 additionalItems 序列化为文本存储到 note 字段
- [ ] 2.4 将 properties 结构序列化为 JSON 存储到 modifiers 字段
- [ ] 2.5 设置 provider_name 字段为 "lineman"，用于区分订单来源

**字段映射表**：

> **参考**: [Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)

| LINE MAN 字段 | TTPOS 字段 | 对应 Grab 字段 | 数据类型 | 说明 |
|--------------|-----------|--------------|---------|------|
| orderId | provider_order_id | orderID | string | LINE MAN 订单 ID |
| orderShortCode | short_order_number | shortOrderNumber | string | 短订单号（后4位） |
| storeId | provider_merchant_id | partnerMerchantID | string | 门店 ID |
| storeId | shop_uuid | partnerMerchantID | string | TTPOS 门店 UUID（需查询映射） |
| restaurantRevenue | total_amount | price.eaterPayment | decimal | 商户收入总额（用户实付金额） |
| orderAcceptedTime | order_time | orderTime | timestamp | 订单接受时间（ISO 8601 → Unix） |
| customerType | order_type | - | string | DELIVERY → DeliveryByProvider, PICKUP → SelfPickup |
| items | - | items | - | 需遍历转换为 order_items |
| items[].id | provider_item_id | items[].id | string | 商品 ID |
| items[].quantity | quantity | items[].quantity | int | 商品数量 |
| items[].unitPrice | price | items[].price | decimal | 商品单价（已含选项费用和折扣） |
| items[].memo | note | - | string | 单品备注 |
| items[].promotionId | - | - | string | 促销活动 ID（可选） |
| items[].discount | discount_amount | - | decimal | 促销折扣金额（可选） |
| items[].properties | modifiers | items[].modifiers | json | 商品选项（序列化） |
| additionalItems | note | - | json | 附加项列表（序列化为文本） |
| memberId | - | - | string | 会员 ID（可选） |

---

### Requirement 3: 订单数据持久化

**用户故事**: 作为系统开发者，我想将订单数据保存到数据库，以便于后续查询和处理

#### 验收标准

1. **WHEN** 订单数据转换完成 **THEN** 系统 **SHALL** 在事务中保存订单主表和明细表
2. **IF** 数据库保存失败 **THEN** 系统 **SHALL** 回滚事务并返回服务器错误（HTTP 500）
3. **WHEN** 订单保存成功 **THEN** 系统 **SHALL** 生成订单 UUID 并返回给调用方
4. **WHEN** 保存订单 **THEN** 系统 **SHALL** 将原始请求数据存储到 raw_data 字段

#### 具体要求

- [ ] 3.1 使用事务保证订单主表和明细表的数据一致性
- [ ] 3.2 生成订单 UUID（使用 guid.S() 或类似方法）
- [ ] 3.3 将原始请求 JSON 序列化后存储到 raw_data 字段
- [ ] 3.4 设置订单状态为 "ACCEPTED"（LINE MAN 订单已被平台接受）
- [ ] 3.5 事务失败时记录详细错误日志并回滚
- [ ] 3.6 查询 shop_uuid（通过 storeId 查询门店配置表）

---

### Requirement 4: 消息队列通知

**用户故事**: 作为系统架构师，我想通过消息队列通知 Main 模块，以便于触发后续的订单处理流程

#### 验收标准

1. **WHEN** 订单保存成功 **THEN** 系统 **SHALL** 发送消息队列事件到 takeout_grab_order topic
2. **WHEN** MQ 发送失败 **THEN** 系统 **SHALL** 记录警告日志但不影响订单保存（订单已入库）
3. **WHEN** 发送 MQ 事件 **THEN** 系统 **SHALL** 包含订单 UUID、订单 ID、门店 UUID、订单来源等信息

#### 具体要求

- [ ] 4.1 复用现有的 `takeout_grab_order` topic（支持多渠道订单）
- [ ] 4.2 发送 OrderEvent 事件，包含以下字段：
  - Action: "create"
  - ProviderName: "lineman"
  - ShopUUID: storeId（或查询到的 shop_uuid）
  - OrderUUID: 订单 UUID
  - OrderID: orderId
  - MerchantID: storeId
  - Status: "ACCEPTED"
  - Timestamp: 当前时间戳
- [ ] 4.3 MQ 发送失败时记录警告日志但不抛出错误（订单已入库）
- [ ] 4.4 使用 `queue.PushWithContext` 发送消息

---

### Requirement 5: 错误处理和日志

**用户故事**: 作为系统运维人员，我想查看详细的日志记录，以便于快速定位和解决问题

#### 验收标准

1. **WHEN** 发生任何错误 **THEN** 系统 **SHALL** 记录详细的错误日志（包含上下文信息）
2. **WHEN** 参数验证失败 **THEN** 系统 **SHALL** 返回标准错误响应（HTTP 400）
3. **WHEN** 数据库操作失败 **THEN** 系统 **SHALL** 返回服务器错误（HTTP 500）
4. **WHEN** 订单保存成功 **THEN** 系统 **SHALL** 记录成功日志（包含订单 UUID 和订单 ID）

#### 具体要求

- [ ] 5.1 使用 `g.Log()` 记录日志，遵循 GoFrame 日志规范
- [ ] 5.2 错误日志包含上下文信息（订单 ID、门店 ID、错误堆栈等）
- [ ] 5.3 成功日志包含关键信息（订单 UUID、订单 ID、处理时间等）
- [ ] 5.4 参数验证失败时返回标准错误格式：`{code, message, data: null}`
- [ ] 5.5 数据库错误不暴露敏感信息（SQL 语句、表名等）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Controller → Logic → DAO，遵循 GoFrame 架构
- **单一职责原则**: Logic 层专注订单处理逻辑，不包含 HTTP 响应处理
- **模块化设计**: 订单转换逻辑独立封装，便于测试和复用
- **依赖管理**: Logic 层依赖 Service 接口（如 ShopProviderCfg 服务）
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范
  - `.cursor/rules/database.mdc` - 数据库规范

### API 设计要求

- [x] Webhook 路径：`POST /partners/:partnerId/stores/:storeId/orders`
- [x] 请求体格式：JSON（参考 LINE MAN API 定义）
- [x] 响应格式：`{code, message, data: {}}` 或 LINE MAN 标准响应
- [x] 成功响应：HTTP 200
- [x] 错误响应：HTTP 400（参数错误）、HTTP 500（服务器错误）

### 数据库设计要求

- [x] 使用现有的 `ttpos_takeout_order` 和 `ttpos_takeout_order_item` 表
- [x] 必须包含: `uuid`, `provider_name`, `provider_order_id`, `raw_data`
- [x] 金额字段使用 decimal(20,8)
- [x] 时间字段使用 timestamp 或 int（Unix 时间戳）
- [x] `raw_data` 字段存储完整的原始请求 JSON

### 性能要求

- [ ] Webhook 响应时间 < 500ms（不含 MQ 发送时间）
- [ ] 数据库事务执行时间 < 200ms
- [ ] 支持并发处理多个订单（无资源竞争）
- [ ] MQ 发送失败不阻塞主流程

### 测试要求

- [ ] Logic 层单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖完整流程（Webhook → 数据库 → MQ）
- [ ] 测试用例包含：
  - 正常订单接收和保存
  - 参数验证失败
  - 数据库保存失败（事务回滚）
  - MQ 发送失败（容错处理）
  - 商品选项（properties）序列化
  - 订单来源区分（lineman vs grab）

### 国际化要求

- [ ] 错误消息支持国际化（使用 i18n）
- [ ] 日志消息使用英文（便于国际化团队协作）

### 安全要求

- [ ] Webhook 请求验证（如需要，参考 LINE MAN 签名验证）
- [ ] SQL 注入防护（使用 GoFrame ORM 参数化查询）
- [ ] 敏感数据不记录到日志（如客户电话、地址等）
- [ ] 原始请求数据存储到 raw_data 时考虑数据脱敏（如有必要）

### 可靠性要求

- [ ] 事务管理保证数据一致性（订单主表和明细表）
- [ ] MQ 发送失败时优雅降级（记录日志但不影响订单入库）
- [ ] 网络异常时返回适当的错误响应
- [ ] 错误日志记录完整（便于问题排查）

---

## 验收标准

### 功能验收

1. **订单接收**: LINE MAN 发送 PlaceOrder 请求，系统成功接收并验证参数
2. **数据转换**: 订单数据正确转换为 TTPOS 订单模型，字段映射准确
3. **订单保存**: 订单主表和明细表正确保存，事务保证一致性
4. **消息队列**: MQ 事件成功发送，Main 模块能够接收并处理
5. **错误处理**: 参数错误、数据库错误、MQ 错误均有正确的处理和日志
6. **订单区分**: 能够通过 provider_name 字段区分 LINE MAN 订单和 Grab 订单

### 测试验收

1. **单元测试**: Logic 层测试覆盖率 ≥ 80%
2. **集成测试**: Webhook → 数据库 → MQ 完整流程测试通过
3. **边界测试**: 参数验证、金额计算、选项序列化等边界情况测试通过
4. **性能测试**: Webhook 响应时间 < 500ms

### 文档验收

1. **技术文档**: design.md 完整且准确（待 `/spec-design` 创建）
2. **API 文档**: LINE MAN PlaceOrder Webhook 接口文档完整
3. **数据库文档**: 字段映射表和 raw_data 格式说明完整
4. **测试文档**: tasks.md 中的测试任务完成（待 `/spec-design` 创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录（自动生成）
- Controller 层代码已生成，只需实现业务逻辑调用
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc` 开发规范
- Logic 层使用 `internal/logic/lineman/` 包（新建）
- Service 接口使用 `internal/service/lineman.go`（需注册）

#### 数据库约束

- 使用现有的 `ttpos_takeout_order` 和 `ttpos_takeout_order_item` 表
- 不修改表结构（除非经过评审批准）
- 使用 GoFrame DAO 进行数据库操作
- 事务中使用 `dao.Order.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {...})`

#### 消息队列约束

- 复用现有的 `takeout_grab_order` topic
- 使用 `queue.PushWithContext(ctx, TopicGrabOrder, event)` 发送消息
- MQ 发送失败不影响订单入库（容错设计）

### 业务约束

- LINE MAN 订单状态固定为 "ACCEPTED"（平台已接受）
- 订单类型映射：DELIVERY → DeliveryByProvider，PICKUP → SelfPickup
- 商品选项（properties）序列化为 JSON 存储
- 原始请求数据完整存储到 raw_data 字段（用于问题排查）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 5（待技术评审确认）
- 必须 ≤ 5 SP，否则需要拆分为多个 Spec

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/internal/dao` - 数据库操作
- `ttpos-bmp/internal/service` - 服务接口
- `ttpos-bmp/internal/pkg/queue` - 消息队列客户端
- `ttpos-bmp/app/ttpos-takeout/api/lineman/v1` - LINE MAN API 定义
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab` - OrderEvent 事件定义

### 服务依赖

- **BMP → Main**: 通过消息队列（takeout_grab_order topic）通知订单创建
- **BMP → ShopProviderCfg 服务**: 查询门店配置（storeId → shop_uuid 映射）

### 业务依赖

- LINE MAN OAuth 认证功能已实现（v2.13.1）
- LINE MAN 菜单同步功能已实现（v2.13.1）
- 门店配置已关联 LINE MAN storeId 和 TTPOS shop_uuid
- Main 模块已支持多渠道订单处理（通过 provider_name 区分）

---

## 风险和缓解

### 风险 1: 数据库 schema 不兼容

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 先对比 LINE MAN 和 Grab 的订单字段，评估是否需要数据库调整
- 如果 properties 结构与 modifiers 差异较大，考虑新增字段或调整 JSON 格式
- 提前与 DBA 确认 raw_data 字段容量是否足够

### 风险 2: 金额计算逻辑不明确

**影响**: 高  
**概率**: 低（已有参考文档）  
**缓解措施**:

- **参考官方映射文档确认字段关系**: [Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
  - `restaurantRevenue` → `total_amount` (对应 Grab 的 `price.eaterPayment`)
  - `items[].unitPrice` → `price` (对应 Grab 的 `items[].price`)
- 与业务方确认金额映射关系
- 编写单元测试覆盖金额计算场景

### 风险 3: MQ Topic 复用导致 Main 模块兼容性问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在 OrderEvent 中增加 `ProviderName` 字段区分订单来源（已支持）
- 与 Main 模块开发者确认 MQ 事件格式兼容性
- 编写集成测试验证 Main 模块能够正确处理 LINE MAN 订单

### 风险 4: 商品选项（properties）序列化失败

**影响**: 中  
**概率**: 低（已有参考实现）  
**缓解措施**:

- **参考官方映射文档**: `items[].properties` → `modifiers` (对应 Grab 的 `items[].modifiers`)
- 参考 Grab modifiers 的 JSON 序列化处理方式（`grab_order.go` 中的实现）
- 编写单元测试覆盖 properties 序列化场景（嵌套数组结构）
- 序列化失败时记录详细日志并返回错误

---

## 时间表

- **Phase 1 - Logic 层实现**: 0.5 天（4h）
  - 创建 lineman_order.go 文件
  - 实现订单转换逻辑（参考 grab_order.go）
  - 实现订单保存逻辑（事务处理）
  
- **Phase 2 - Controller 和 Service 实现**: 0.3 天（2h）
  - 完成 Controller 业务逻辑调用
  - 注册 Lineman Service 接口
  - 实现 MQ 事件发送
  
- **Phase 3 - 测试和文档**: 0.9 天（7h）
  - 单元测试（订单转换、入库）
  - 集成测试（Webhook + MQ + Main 模块）
  - 更新 API 文档
  
- **Phase 4 - 联调和优化**: 0.3 天（3h）
  - 与 LINE MAN 平台联调
  - 与 Main 模块联调
  - 性能优化和错误处理完善

- **总计**: 2-3 天（SP = 5）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go` - Grab 订单处理参考实现
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/event.go` - OrderEvent 事件定义

### 开发指南

- `docs/others/lineman/sources/PartnerIntegrationWorkflow.md` - LINE MAN API 文档
- `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go` - LINE MAN 订单 API 定义

### 外部参考

- **LINE MAN & Grab 字段映射文档**: [Lineman API定义及TTPOS 映射 - Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165) ⭐ **核心参考文档**
- LINE MAN Developer Portal: [待补充链接]
- GoFrame 文档: https://goframe.org

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-12.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**审核者**: 待审核
