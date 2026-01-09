> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 叫号系统-订单已完成自动消失（时间）需求文档

> 本文档定义 叫号系统-订单已完成自动消失（时间） 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/callboard-order-auto-hide-after-completion-timeout.md](../../../../team/proposals/2025-12/callboard-order-auto-hide-after-completion-timeout.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | 王昱                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

为叫号系统添加订单完成后自动消失功能，使用现有的 `timeout_limit` 配置项。当订单完成后超过设定的时间，自动从叫号系统显示屏上移除该订单的显示内容。设置为 0 时表示订单完成后不自动消失。

该功能可以：
- 提升用户体验：显示屏只显示需要关注的订单，信息更清晰
- 提高运营效率：减少视觉干扰，帮助店员更快定位待处理订单
- 灵活配置：支持按门店需求设置订单显示时长，满足不同业务场景

**技术说明**：
- 使用现有的 `timeout_limit` 字段（单位：分钟），无需新增字段
- 订单完成时间从 `PreparedQueue` 的 score（Redis Sorted Set 的时间戳）获取
- 在 `/callboard/data` 接口中根据 `timeout_limit` 过滤 `PreparedQueue` 中的订单

## 🎯 产品对齐

该功能支持叫号系统显示屏的信息管理，通过自动移除已完成的订单，提升显示屏的可读性和可用性，帮助操作人员更高效地处理订单。

## 📝 用户故事

**作为** 叫号系统操作人员  
**我想** 配置订单完成后的自动消失时间  
**以便于** 显示屏上只显示需要关注的订单，提高信息可读性和操作效率

---

## 功能需求

### Requirement 1: 根据 timeout_limit 过滤已完成订单

**用户故事**: 作为叫号系统操作人员，我想已完成的订单在超过设定时间后自动消失，以便于显示屏上只显示需要关注的订单

#### 验收标准

1. **WHEN** 订单完成时间超过配置的 `timeout_limit`（分钟） **THEN** 系统 **SHALL** 自动从叫号显示屏移除该订单
2. **IF** 订单完成后未超过配置的 `timeout_limit` **THEN** 系统 **SHALL** 继续在叫号显示屏上显示该订单
3. **IF** 配置的 `timeout_limit` 为 0 或 nil **THEN** 系统 **SHALL** 不进行过滤，返回所有已完成订单
4. **WHEN** 设备端请求 `/callboard/data` 接口时 **THEN** 系统 **SHALL** 过滤掉 `PreparedQueue` 中已超过超时时间的订单

#### 具体要求

- [ ] 1.1 在 `GetQueueData` 方法中，从 `bindInfo.TimeoutLimit` 获取超时时间配置（单位：分钟）
- [ ] 1.2 在获取 `PreparedQueue` 时，根据 `timeout_limit` 过滤订单：
  - 如果 `timeout_limit` 为 0 或 nil，返回所有订单
  - 如果 `timeout_limit` > 0，过滤掉 `当前时间 - 订单完成时间(score) > timeout_limit * 60` 的订单
- [ ] 1.3 使用服务器时间（`time.Now().Unix()`）进行判断，确保时间同步准确
- [ ] 1.4 订单完成时间从 Redis Sorted Set 的 score 获取（Unix 时间戳，单位：秒）
- [ ] 1.5 过滤逻辑在 `getCallBoardQueue` 方法或调用处实现，确保高效执行

---

### Requirement 2: 配置兼容性

**用户故事**: 作为系统管理员，我想现有配置能够正常工作，以便于不影响现有功能

#### 验收标准

1. **WHEN** 设备配置中 `timeout_limit` 为 nil 时 **THEN** 系统 **SHALL** 默认视为 0（不自动消失）
2. **WHEN** 设备配置中 `timeout_limit` 为 0 时 **THEN** 系统 **SHALL** 不进行过滤，返回所有订单
3. **WHEN** 修改设备配置的 `timeout_limit` 时 **THEN** 系统 **SHALL** 立即生效，影响后续 `/callboard/data` 接口的返回结果

#### 具体要求

- [ ] 2.1 使用现有的 `timeout_limit` 字段，无需新增字段或数据库迁移
- [ ] 2.2 在 `GetQueueData` 方法中处理 `timeout_limit` 为 nil 的情况，默认视为 0
- [ ] 2.3 确保现有设备配置（`timeout_limit` 为 nil 或 0）的行为保持不变

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/callboard/data`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 无需新增数据库字段，使用现有的 `timeout_limit` 配置（存储在 Redis）
- [x] 订单完成时间从 Redis Sorted Set 的 score 获取，无需额外存储
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 订单过滤逻辑在 Service 层高效实现，避免影响 Redis 查询性能
- [ ] 过滤逻辑应在获取队列数据后、返回前进行，减少不必要的计算
- [ ] 考虑在 Redis 查询时使用 `ZRangeByScore` 直接过滤，提升性能

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程（订单完成 → 超时过滤）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] SQL 注入防护（使用参数化查询）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 时间同步问题处理（使用服务器时间）

---

## 验收标准

### 功能验收

1. **超时时间配置**: 使用现有的 `timeout_limit` 配置（单位：分钟），默认值为 0 或 nil（不自动消失）
2. **自动移除机制**: `PreparedQueue` 中订单完成时间超过 `timeout_limit` 分钟，自动从叫号显示屏移除
3. **显示逻辑**: `PreparedQueue` 中订单完成时间未超过 `timeout_limit` 分钟，继续显示在叫号显示屏上
4. **配置兼容性**: 现有设备配置（`timeout_limit` 为 nil 或 0）的行为保持不变，不进行过滤
5. **时间准确性**: 使用服务器时间（`time.Now().Unix()`）进行判断，订单完成时间从 Redis score 获取

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%）
2. **API 测试**: `/callboard/data` 接口测试通过，验证过滤逻辑
3. **集成测试**: 端到端流程测试通过（订单完成制作 → 进入 PreparedQueue → 超时过滤 → 显示更新）
4. **边界测试**: 
   - `timeout_limit` 为 0、nil、负数、超大值等边界情况测试通过
   - 订单完成时间刚好等于超时时间的情况测试通过
   - `PreparedQueue` 为空的情况测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整（如有变更）
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

**说明**：无需数据库文档，因为不涉及数据库变更

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 必须保持向后兼容：现有设备配置（`timeout_limit` 为 nil 或 0）的行为保持不变
- 订单完成时间从 Redis Sorted Set 的 score 获取，使用服务器时间戳
- 过滤逻辑必须高效，不能影响接口性能
- `timeout_limit` 单位：分钟，需要转换为秒进行比较（`timeout_limit * 60`）

### 资源约束

- 开发时间: 3-5 天
- Story Point: {SP 值} (必须 ≤ 5，待技术评审确认)

---

## 依赖关系

### 技术依赖

- 现有叫号系统模块 (`main/app/service/callboard`)
- Redis Sorted Set（存储订单队列，score 为完成时间）

### 服务依赖

- 无需额外服务依赖，使用现有 Redis 数据结构

### 业务依赖

- 依赖 `story-callboard-data-config`：`timeout_limit` 配置已在该功能中实现
- 订单完成时间从 `PreparedQueue` 的 score 获取（订单完成制作时调用 `pushToPreparedQueue`，score 为 `time.Now().Unix()`）

---

## 风险和缓解

### 风险 1: 时间同步问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用服务器时间而非客户端时间进行判断
- 在 Service 层统一处理时间计算逻辑
- 记录订单完成时间时使用服务器时间戳

### 风险 2: 实时性要求

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 采用 WebSocket 或轮询机制确保状态实时同步
- 在订单状态变更时及时更新完成时间
- 优化查询性能，确保接口响应时间 < 200ms

### 风险 3: 配置兼容性

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用现有的 `timeout_limit` 字段，无需数据库迁移
- 现有设备配置（`timeout_limit` 为 nil 或 0）默认不进行过滤，保持向后兼容
- 配置更新接口已存在（`UpdateBindInfo`），支持后续调整

### 风险 4: 性能影响

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在数据库查询层面或 Service 层高效实现过滤逻辑
- 考虑使用索引优化查询性能
- 缓存设备配置信息，减少数据库查询

---

## 时间表

- **Phase 1 - 业务逻辑实现**: 1-2 天（Service 层过滤逻辑实现）
- **Phase 2 - 测试和优化**: 1 天（单元测试、集成测试、性能优化）
- **总计**: 2-3 天（SP = {值}，待技术评审确认）

**说明**：无需数据库迁移，无需新增字段，开发工作量较小

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 相关功能

- `story-callboard-data-config` - 叫号系统数据配置相关功能
- `docs/team/proposals/2025-12/callboard-data-config.md` - 叫号系统数据配置提案

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- [外部参考链接]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: 王昱  
**审核者**: {审核者}
