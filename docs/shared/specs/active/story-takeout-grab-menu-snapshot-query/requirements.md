# Grab 菜单快照数据查询 gRPC 服务 需求文档

> 本文档定义 Grab 菜单快照数据查询 gRPC 服务 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-menu-snapshot-query.md](../../../../team/proposals/2025-12/grab-menu-snapshot-query.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

本功能旨在 `ttpos-takeout` 服务中提供一个 gRPC 接口，允许开发和运维人员通过 `provider_name`, `shop_uuid` 和 `request_id` 查询特定时刻的 Grab 菜单快照原始数据。这将极大地简化菜单同步问题的排查过程，提供数据的可追溯性。

## 🎯 产品对齐

该功能支持系统可观测性和运维效率的提升，确保在出现菜单问题时能够快速定位根因，保障商户业务的连续性。

## 📝 用户故事

**作为** 开发人员或运维人员
**我想** 通过 API 查询历史菜单快照的原始数据
**以便于** 对比当前数据与原始数据，排查同步错误或验证数据一致性

---

## 功能需求

### Requirement 1: 提供菜单快照查询 gRPC 接口

**用户故事**: 作为开发人员，我想调用 gRPC 接口获取菜单快照，以便于进行问题排查。

#### 验收标准

1. **WHEN** 客户端调用 `GetMenuSnapshot` 方法并提供有效的 `provider_name`, `shop_uuid`, `request_id` **THEN** 服务端应返回对应的菜单快照 JSON 字符串。
2. **IF** 指定的 `request_id` 在数据库中不存在 **THEN** 服务端应返回 NotFound 错误。
3. **WHEN** `provider_name` 无效或不支持 **THEN** 服务端应返回 InvalidArgument 错误。

#### 具体要求

- [ ] 1.1 在 `ttpos-takeout` 的 proto 文件中定义 `GetMenuSnapshot` RPC 方法。
- [ ] 1.2Request 消息应包含 `string provider_name`, `string shop_uuid`, `string request_id`。
- [ ] 1.3 Response 消息应包含：
  - `string content` (Provider 侧原始菜单 JSON 数据)
  - `int64 updated_at` (快照更新时间)
  - `string sync_state` (菜单同步状态)
- [ ] 1.4 实现服务层逻辑，根据参数查询数据库存储的快照。

### 数据库变更要求

- [ ] 表 `channel_menu_snapshot` 字段重命名：
  - `create_time` -> `created_at`
  - `update_time` -> `updated_at`
- [ ] 表 `channel_menu_snapshot` 新增字段：
  - `deleted_at` int(11) DEFAULT NULL COMMENT '删除时间'
  - `sync_state` varchar(32) NOT NULL DEFAULT 'QUEUEING' COMMENT '同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED'
  - `ttpos_menu_data` longtext COMMENT 'TTPOS 侧菜单原始数据 (JSON)'
  - `ttpos_updated_at` int(11) NOT NULL DEFAULT 0 COMMENT 'TTPOS 侧菜单数据更新时间'
- [ ] 创建对应的数据库迁移脚本 (migration)
- [ ] 重新生成 GoFrame Entity 代码

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [ ] 遵循 gRPC 接口定义规范
- [ ] 错误码使用 gRPC 标准状态码 (如 NotFound, InvalidArgument)

### 数据库设计要求

- [ ] 复用现有的菜单快照存储表（需确认是否存在，若不存在需新增）

### 性能要求

- [ ] 接口响应时间 < 200ms (在数据量合理范围内)
- [ ] 避免查询超大 Blob 数据导致内存溢出

### 安全要求

- [ ] 接口应受权限控制（如果对外暴露），或者限制为内部调用

---

## 验收标准

### 功能验收

1. **正常查询**: 输入存在的 ID 能查到数据。
2. **异常处理**: 输入不存在的 ID 能正确报错。

### 测试验收

1. **单元测试**: Service 层逻辑覆盖率 ≥ 70%。
2. **API 测试**: gRPC 接口测试通过。

### 文档验收

1. **技术文档**: 更新 `ttpos-takeout` 的 API 文档。

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-takeout` 模块

### 业务依赖

- 菜单同步流程中必须已经包含“保存快照”的逻辑，否则无法查询。本需求假设快照存储已存在或作为本需求的一部分补全（需确认）。

---

## 风险和缓解

### 风险 1: 数据未存储或已过期

**影响**: 高
**概率**: 中
**缓解措施**:
- 确认现有的菜单同步逻辑中是否持久化了原始 JSON。如果没有，需要在此次开发中补充“保存快照”的逻辑，或者明确本功能仅对未来数据生效。

---

## 时间表

- **Phase 1 - 接口定义与实现**: 0.5 天
- **Phase 2 - 测试与联调**: 0.5 天
- **总计**: 1 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0
**创建日期**: 2025-12-11
**作者**: rikugun
**审核者**: {审核者}
