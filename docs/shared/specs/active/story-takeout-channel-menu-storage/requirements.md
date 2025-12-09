# 外卖渠道菜单数据存储需求文档

> 本文档定义外卖渠道菜单数据存储功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-channel-menu-storage.md](../../../../team/proposals/2025-12/takeout-channel-menu-storage.md) |
| **创建日期**      | 2025-12-08                                                                                                 |
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

本功能旨在为 TTPOS 系统构建一个统一、隔离且通用的存储机制，用于保存 TTPOS 商家发送给不同外卖渠道（如 Grab, Lineman）的菜单数据快照。由于不同渠道的数据结构差异巨大，本功能将采用 Schema-less 的 JSON 存储方式，通过 `shop_uuid` 和 `provider_name` 进行索引，以实现对多渠道数据的兼容与追溯。

## 🎯 产品对齐

- **业务灵活性**：支持不同渠道拥有独立的菜单配置，满足各外卖平台特定的业务规则。
- **运维可观测性**：提供菜单数据快照，帮助运维和开发人员在出现同步问题时快速定位是 TTPOS 数据生成问题还是渠道端接收问题。
- **系统解耦**：将渠道特定数据与核心业务数据解耦，降低系统复杂度。

## 📝 用户故事

**作为** 外卖运营/系统开发人员
**我想** 能够保存和查询特定商户针对特定渠道（如 Grab, Lineman）生成的菜单数据快照
**以便于** 在外卖平台菜单显示异常时，确认最后一次同步的数据内容，并进行调试或重新推送。

---

## 功能需求

### Requirement 1: 渠道菜单数据存储

**用户故事**: 作为系统后台服务，我想在生成外卖菜单后，将菜单 JSON 数据保存到数据库中，以便后续查询。

#### 验收标准

1. **WHEN** 调用保存接口传入 `shop_uuid`, `provider_name` 和菜单 JSON 数据 **THEN** 系统应将数据保存到数据库。
2. **IF** 该商户在该渠道已存在菜单数据 **THEN** 系统应覆盖更新为最新的 JSON 数据。
3. **WHEN** 保存 Grab 渠道的复杂嵌套 JSON 结构 **THEN** 系统应能完整存储不丢失。
4. **WHEN** 保存 Lineman 渠道的扁平 JSON 结构 **THEN** 系统应能完整存储不丢失。

#### 具体要求

- [ ] 1.1 数据库表设计应包含 `shop_uuid`, `provider_name`, `menu_data` 字段。
- [ ] 1.2 `shop_uuid` + `provider_name` 应建立唯一索引或作为联合主键。
- [ ] 1.3 `menu_data` 字段应使用 `LONGTEXT` 或足够大的文本类型，以容纳大型菜单 JSON。
- [ ] 1.4 提供内部 Service 接口 `SaveChannelMenu(ctx, shopUUID, providerName, menuData)`。

---

### Requirement 2: 渠道菜单数据读取

**用户故事**: 作为开发人员，我想通过 API 读取指定商户、指定渠道的菜单快照，以便分析数据问题。

#### 验收标准

1. **WHEN** 调用读取接口传入有效的 `shop_uuid` 和 `provider_name` **THEN** 系统应返回对应的菜单 JSON 字符串。
2. **IF** 传入的商户或渠道不存在数据 **THEN** 系统应返回空或明确的 "Not Found" 错误。

#### 具体要求

- [ ] 2.1 提供内部 Service 接口 `GetChannelMenu(ctx, shopUUID, providerName)`。
- [ ] 2.2 读取操作应高效，利用索引直接定位数据。

---

## 非功能需求

### 代码架构和模块化

- **模块位置**: 本功能应实现在 `ttpos-bmp/app/ttpos-takeout` 模块中。
- **分层设计**: 遵循 Controller (可选) -> Logic -> Dao 的分层结构。
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/database.mdc` - 数据库开发规范

### 数据库设计要求

- [ ] 表名建议: `takeout_channel_menu_snapshot` 或类似名称。
- [ ] 必须包含基础字段: `create_time`, `update_time`。
- [ ] 字段名使用 snake_case。

### 性能要求

- [ ] 读写操作应在 200ms 内完成（假设 JSON 大小在合理范围内，如 < 1MB）。
- [ ] 必须使用索引避免全表扫描。

### 安全要求

- [ ] 内部接口调用，无需对外暴露，确保仅授权服务可调用。

---

## 验收标准

### 功能验收

1. **Grab 菜单存储**: 模拟保存一份真实的 Grab 菜单 JSON，验证能否成功保存且读取内容一致。
2. **Lineman 菜单存储**: 模拟保存一份真实的 Lineman 菜单 JSON，验证能否成功保存且读取内容一致。
3. **数据隔离**: 验证修改 Grab 的菜单数据不会影响 Lineman 的数据。

### 测试验收

1. **单元测试**: 覆盖 DAO 层和 Logic 层的读写逻辑。
2. **集成测试**: 模拟完整调用流程，确保数据库交互正确。

### 文档验收

1. **技术文档**: design.md 包含详细的表结构设计。
2. **数据库文档**: 提供 SQL 迁移脚本 (`up.sql` 和 `down.sql`)。

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架。
- 数据库操作应通过 `gdb` 或生成的 DAO 进行。

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3

---

## 依赖关系

### 业务依赖

- 依赖于 `ttpos-takeout` 模块现有的 Grab/Lineman 集成流程（作为流程中的一步调用）。

---

## 风险和缓解

### 风险 1: 菜单数据过大

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 使用 `LONGTEXT` 类型。
- 监控字段大小，必要时在应用层进行压缩（如 GZIP）后再存储（需评估 CPU 开销）。

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-08.md`
