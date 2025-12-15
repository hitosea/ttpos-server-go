# 整合 Skootar 订单逻辑到现有订单模型 需求文档

> 本文档定义 整合 Skootar 订单逻辑到现有订单模型 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-skootar-integration.md](../../../../team/proposals/2025-12/takeout-skootar-integration.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | User                                                                                                       |
| **目标 Sprint**   | Sprint 24                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | TBD             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

目前 `ttpos-takeout` 模块中的 Skootar 配送逻辑实现较早，直接依赖 `takeout_job` 表中的特定字段，与新引入的 Grab 对接模型（`takeout_order`）不一致。本需求旨在重构 Skootar 订单逻辑，采用 "主表 (`takeout_order`) + 扩展表 (`takeout_order_skootar`)" 的模式，统一订单模型，提升系统可扩展性和维护性。

## 🎯 产品对齐

该重构支持产品向多渠道外卖平台扩展的战略目标，通过统一的数据模型降低后续接入新配送商（如 Lalamove, Foodpanda）的成本，并减少维护技术债务。

## 📝 用户故事

**作为** 开发人员  
**我想** 将 Skootar 订单数据结构迁移到通用的订单模型中  
**以便于** 统一管理所有外卖渠道的订单，并支持快速接入新的配送服务商

---

## 功能需求

### Requirement 1: 数据库模型重构与数据迁移

**用户故事**: 作为 运维人员，我想 确保历史 Skootar 订单数据完整迁移到新表结构，以便于 系统平滑升级且数据不丢失。

#### 验收标准

1. **WHEN** 执行迁移脚本 **THEN** 原 `takeout_job` 表中的通用字段数据（如金额、状态、客户信息）应准确迁移至 `takeout_order` 表。
2. **WHEN** 执行迁移脚本 **THEN** 原 `takeout_job` 表中的 Skootar 特有字段（如 `skootar_id`）应保留在扩展表（原表瘦身或新表）中，并通过 `order_uuid` 与主表关联。
3. **IF** 迁移过程中发生错误 **THEN** 脚本应支持回滚，保证数据一致性。

#### 具体要求

- [ ] 1.1 创建 `takeout_order_skootar` 表（或重构 `takeout_job`），仅保留 `uuid`, `order_uuid`, `skootar_id`, `skootar_name`, `skootar_phone`, `skootar_rating` 等特有字段。
- [ ] 1.2 编写 SQL 迁移脚本，将历史数据的通用部分写入 `takeout_order`，特有部分保留/写入扩展表。
- [ ] 1.3 验证数据迁移的准确性（记录数、关键字段值一致）。

---

### Requirement 2: 业务逻辑适配 (CreateOrder)

**用户故事**: 作为 系统，我想 在创建 Skootar 订单时使用新的数据模型，以便于 保持数据结构的一致性。

#### 验收标准

1. **WHEN** 调用 `CreateOrder` 接口下单 **THEN** 系统应在 `takeout_order` 表创建通用订单记录。
2. **WHEN** 调用 `CreateOrder` 接口下单 **THEN** 系统应在扩展表创建 Skootar 特有信息记录。
3. **WHEN** 下单成功 **THEN** 返回的响应结构应保持不变，兼容现有前端/API 调用。

#### 具体要求

- [ ] 2.1 修改 `internal/logic/skootar` 的 `CreateOrder` 方法，适配 "主表+扩展表" 的写入逻辑。
- [ ] 2.2 确保事务一致性：主表和扩展表的写入应在同一事务中完成。

---

### Requirement 3: 业务逻辑适配 (Query/GetDriverInfo)

**用户故事**: 作为 前端应用，我想 获取 Skootar 订单详情（含司机信息），以便于 展示给用户。

#### 验收标准

1. **WHEN** 调用 `GetDriverInfo` 或查询订单详情 **THEN** 系统应同时查询主表和扩展表，聚合数据后返回。
2. **IF** 订单为 Skootar 渠道 **THEN** 返回的司机信息应准确来源于扩展表。
3. **IF** 订单为其他渠道 **THEN** 不应查询 Skootar 扩展表。

#### 具体要求

- [ ] 3.1 修改 `internal/controller` 及 `logic` 层的数据获取方法，通过 Join 或多次查询聚合数据。
- [ ] 3.2 保持对外 gRPC/HTTP 接口契约不变。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 扩展表必须包含 `order_uuid` 作为外键（逻辑外键）关联主表。
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 测试要求

- [ ] Skootar 下单流程集成测试通过。
- [ ] 历史数据迁移测试通过。
- [ ] API 兼容性测试通过。

### 可靠性要求

- [ ] 数据库迁移需在低峰期进行，并有备份方案。

---

## 验收标准

### 功能验收

1. **迁移验证**: 历史 Skootar 订单在重构后的系统中能正常查询和显示详情。
2. **新单验证**: 新创建的 Skootar 订单能正确落库到主表和扩展表。
3. **流程验证**: 接单、取消、状态回调等全流程功能正常。

### 测试验收

1. **API 测试**: `CreateOrder`, `GetDriverInfo`, `Callback` 等接口响应结构与重构前一致。

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5

---

## 依赖关系

### 技术依赖

- `takeout_order` 表结构（需先由 Grab 集成需求创建或同步创建）。

### 业务依赖

- 需确认 `takeout_order` 表字段定义能覆盖 Skootar 的通用需求。

---

## 风险和缓解

### 风险 1: 数据迁移失败

**影响**: 高  
**概率**: 低  
**缓解措施**:
- 编写可逆的迁移脚本（Up/Down）。
- 在 Staging 环境全量演练。

### 风险 2: API 兼容性破坏

**影响**: 高  
**概率**: 中  
**缓解措施**:
- 严格对比重构前后的 API 响应 JSON。
- 保持 Controller 层 DTO 定义不变，仅修改内部映射逻辑。

---

## 时间表

- **Phase 1 - 数据库重构与迁移脚本**: 1.5 天
- **Phase 2 - 业务逻辑适配**: 2 天
- **Phase 3 - 测试与验证**: 1.5 天
- **总计**: 5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: User  
**审核者**: TBD

