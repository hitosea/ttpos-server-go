# Grab 店铺集成状态落库与旅程关联需求文档

> 本文档定义 Grab 店铺集成状态落库与旅程关联的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                      |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-shop-provider-cfg-integration-status.md](../../../../team/proposals/2025-12/grab-shop-provider-cfg-integration-status.md) |
| **创建日期**      | 2025-12-11                                                                                                                |
| **负责人**        | rikugun                                                                                                                   |
| **目标 Sprint**   | Sprint TBD                                                                                                                |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | - |
| **审核日期** | - |
| **审核意见** | - |

---

## 📋 概述

为 Grab 外卖集成补齐门店级别的配置与状态落库，统一存储在 `shop_provider_cfg`，确保 webhook 推送与创建自助旅程后的状态一致可查，并为后续监控/排障提供数据基础。

## 🎯 产品对齐

支撑商家管理端对第三方外卖（Grab）集成的可观测性与稳定性，减少人工校验，加速上线与问题恢复。

## 📝 用户故事

**作为** 商户管理员/运营  
**我想** 在 Grab 集成创建或状态变化时有可靠的门店状态记录  
**以便于** 及时了解集成健康度并快速排查问题

---

## 功能需求

### Requirement 1: 新增 shop_provider_cfg 表

**用户故事**: 作为运维/研发，我想持久化每个门店的第三方集成状态，以便排查与统计。

#### 验收标准

1. **WHEN** 执行迁移 **THEN** 系统创建 `shop_provider_cfg`，包含字段：`id`(PK), `uuid`, `shop_uuid`, `provider_name`, `provider_merchant_id`, `provider_shop_status`(enum: INACTIVE/ACTIVE/SYNCING/FAILED), `created_at`, `updated_at`, `deleted_at`。
2. **IF** 未指定默认值 **THEN** 时间字段默认为 0，枚举默认 INACTIVE。
3. **WHEN** 查询表结构 **THEN** 索引覆盖 `shop_uuid` + `provider_name`。

#### 具体要求

- [ ] 1.1 迁移脚本位于 `ttpos-bmp/app/ttpos-takeout/manifest/sql/`，含 up/down。
- [ ] 1.2 表名与字段使用 snake_case，时间为 int。
- [ ] 1.3 provider_name 预设值包含 grab

---

### Requirement 2: Grab integrationStatus 回调落库

**用户故事**: 作为运营，我想在 Grab 推送 integrationStatus 时，系统自动更新门店状态，以便实时查看。

#### 验收标准

1. **WHEN** 收到 Grab integrationStatus webhook **THEN** 系统按 `shop_uuid + provider_name=grab` 定位记录并更新 `provider_shop_status`，无记录则创建。
2. **WHEN** 状态为 ACTIVE/INACTIVE/SYNCING/FAILED **THEN** 与表枚举一致且更新时间写入。
3. **IF** 重复推送同一状态 **THEN** 不产生重复错误，更新时间可覆盖或保持最近一次。
4. **WHEN** 接口处理失败 **THEN** 记录错误日志便于排查。
5. **WHEN** 状态更新完成 **THEN** 发布门店集成状态消息到 `takeout_store_integration_state`，携带 shop_uuid、provider_name、provider_shop_status、provider_merchant_id、更新时间。

#### 具体要求

- [ ] 2.1 integrationStatus 与内部枚举映射表落在代码常量。
- [ ] 2.2 更新逻辑幂等（upsert / 先查再更新），避免并发脏写。
- [ ] 2.3 保留 provider_merchant_id（若推送携带或已有存量）。
- [ ] 2.4 状态更新后推送消息到 `takeout_store_integration_state`，与内部消息格式保持一致（含状态、门店、提供商、merchant_id、时间戳）。

---

### Requirement 3: CreateSelfServeJourney 成功后写入配置

**用户故事**: 作为运营，我想在创建 Grab 自助旅程时自动建立门店配置记录，确保后续状态闭环。

#### 验收标准

1. **WHEN** CreateSelfServeJourney 成功 **THEN** 系统写入/更新 `shop_provider_cfg`，填充 `shop_uuid`、`provider_name=grab`、`provider_merchant_id`，并将 `provider_shop_status` 置为 SYNCING 或 ACTIVE（按业务定义）。
2. **IF** 记录已存在 **THEN** 更新 `provider_merchant_id` 与状态，不新增重复行。
3. **WHEN** 上游失败 **THEN** 不写入成功状态，允许落 FAILED 以便追踪。

#### 具体要求

- [ ] 3.1 逻辑放在 `self_serve_journey` 流程，使用同一枚举常量。
- [ ] 3.2 日志记录旅程创建结果与落库结果，便于排查。
- [ ] 3.3 未来可复用到其他旅程创建路径（预留接口/Service 方法）。

---

### Requirement 4: gRPC 查询 shop_provider_cfg（按 shop_uuid）

**用户故事**: 作为内部/管理端服务调用方，我想通过 gRPC 以 shop_uuid 查询门店的 Grab 集成状态，以便界面或任务快速获取最新状态。

#### 验收标准

1. **WHEN** 传入合法 `shop_uuid` **THEN** gRPC 接口返回对应 `shop_provider_cfg` 记录（至少包含 provider_name、provider_shop_status、provider_merchant_id、updated_at）。
2. **IF** 无记录或已软删 **THEN** 返回 not found/空结果，且 code/message 明确。
3. **WHEN** provider_name 多渠道 **THEN** 支持按 provider_name=grab 过滤，默认只返回 grab。
4. **WHEN** 调用异常 **THEN** 记录错误日志，保持稳定的错误码。

#### 具体要求

- [ ] 4.1 gRPC 定义放在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`（新增 message/request/response 及 service 方法）。
- [ ] 4.2 Service 层查询使用 `shop_uuid + provider_name=grab` 唯一索引，遵循幂等与只读。
- [ ] 4.3 接口注册到 Nacos/服务发现，遵循现有 grab gRPC 命名规范与鉴权策略。
- [ ] 4.4 返回字段对齐落库字段命名（snake_case / protobuf style），时间使用 int。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Controller → Logic/Service → DAO，遵循 GoFrame 与项目规范。
- **单一职责**: 表迁移、Webhook 处理、旅程创建各自独立。
- **依赖管理**: Logic 依赖 Service 接口，不直接耦合 DAO。
- **遵循规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`、`.cursor/rules/database.mdc`、`.cursor/rules/api.mdc`。

### API 设计要求

- [ ] 使用 snake_case URL；data 为对象非 null。
- [ ] 统一响应 `{code, message, data{}}`。

### 数据库设计要求

- [ ] 必含 `id, uuid, created_at, updated_at, deleted_at`，int 时间。
- [ ] 针对 `shop_uuid + provider_name` 建唯一或联合索引。

### 性能要求

- [ ] Webhook/旅程处理本地响应 < 200ms（不含外部调用）。

### 测试要求

- [ ] Service/Logic 单元测试覆盖 webhook 映射与 upsert。
- [ ] 迁移脚本通过本地执行验证。

### 安全要求

- [ ] Webhook 需验证签名/来源（沿用现有 Grab 验证方式）。
- [ ] 防止 SQL 注入，使用参数化。

### 可靠性要求

- [ ] 幂等更新，避免重复通知导致数据漂移。
- [ ] 记录错误日志，必要时告警。

---

## 验收标准

### 功能验收

1. **表结构可见且符合字段/索引要求**: up/down 迁移均成功。
2. **Webhook 推送 ACTIVE/FAILED/SYNCING/INACTIVE**: 对应记录状态正确更新，可查询到最新时间。
3. **CreateSelfServeJourney 成功**: 表中存在/更新对应记录，状态与 merchant_id 正确。

### 测试验收

1. **单元测试**: 覆盖 webhook 状态映射与 upsert 流程。
2. **集成测试**: 模拟 webhook + 旅程创建串联验证状态闭环。

### 文档验收

1. **技术文档**: design.md 补充枚举映射与落库流程。
2. **数据库文档**: 迁移脚本与表结构说明完备。

---

## 约束条件

### 技术约束

#### Go BMP 模块
- 使用 GoFrame 2.x；遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`。
- 不直接修改自动生成的 dao/entity/do。
- gRPC 服务注册到 Nacos（如新增接口）。

### 业务约束

- 仅针对 Grab 外卖集成；其他渠道后续复用同表结构。

### 资源约束

- 开发时间: 3 天
- Story Point: 5 (≤5)

---

## 依赖关系

### 技术依赖

- Grab webhook/integrationStatus 现有验证逻辑。
- 现有旅程创建 gRPC/SDK 能力。

### 业务依赖

- 门店已具备有效的 shop_uuid 与授权凭据。

---

## 风险和缓解

### 风险 1: 状态映射不一致

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 明确映射表，编写单元测试覆盖四种状态。
- 设计幂等 upsert，避免重复推送异常。

### 风险 2: 并发/重复通知导致脏写

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 以 `shop_uuid + provider_name` 唯一约束，使用事务或乐观更新。
- 记录上次状态与时间，避免回退。

---

## 时间表

- **Phase 1 - 迁移与枚举梳理**: 1 天
- **Phase 2 - Webhook 落库与测试**: 1 天
- **Phase 3 - 自助旅程落库与联调**: 1 天
- **总计**: 3 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc`
- `.cursor/rules/go-bmp.mdc`
- `.cursor/rules/api.mdc`
- `.cursor/rules/database.mdc`
- `.cursor/rules/security.mdc`

### 外部参考

- Grab integrationStatus webhook 文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: TTPOS Team  
**审核者**: -  
