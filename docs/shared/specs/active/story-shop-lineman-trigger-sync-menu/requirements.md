# Lineman 触发菜单同步（TriggerSyncMenu）需求文档

> 本文档定义 Lineman 触发菜单同步功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                   |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14-shop-lineman-trigger-sync-menu.md](../../../../team/proposals/2026-01/v2.14-shop-lineman-trigger-sync-menu.md) |
| **创建日期**      | 2026-01-15                                                                                                             |
| **负责人**        | rikugun                                                                                                                |
| **目标 Sprint**   | Sprint TBD                                                                                                             |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                             |

## 📋 审核状态

| 项目         | 内容                 |
| ------------ | -------------------- |
| **审核状态** | 已通过               |
| **审核人**   | rikugun              |
| **审核日期** | 2026-01-15           |
| **审核意见** | 需求明确，可进入开发 |

---

## 📋 概述

当前 Lineman 的菜单同步流程状态不可观测，触发请求未记录到 `menu_log`，导致无法追踪与审计。本功能新增 TriggerSyncMenu 接口，接收到请求后先写入 `menu_log` 记录（`sync_type=NOTIFY`，`status=QUEUED`），随后调用 `service.Lineman().SyncMenu(ctx, shopUUID)` 发起同步，从而实现同步流程可追踪、可审计。

接口协议参考：[Lineman API 定义及 TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1404604549#gid=1404604549)

## 🎯 产品对齐

该功能支持外卖渠道集成的可靠性和可观测性目标，通过记录同步触发日志，使商户管理员能够追踪菜单同步状态，快速定位问题，降低运维成本。

## 📝 用户故事

**作为** 商户管理员  
**我想** 通过 TriggerSyncMenu 触发菜单同步并记录日志  
**以便于** 同步流程可追踪、可审计

---

## 功能需求

### Requirement 1: TriggerSyncMenu 接口实现

**用户故事**: 作为商户管理员，我想通过 TriggerSyncMenu 接口触发菜单同步，以便于 Lineman 平台主动通知菜单更新。

#### 验收标准

1. **WHEN** Lineman 平台调用 TriggerSyncMenu 接口 **THEN** 系统 **SHALL** 解析 `partnerId` 和 `storeId` 参数
2. **WHEN** 参数解析成功 **THEN** 系统 **SHALL** 写入 `menu_log` 记录，`sync_type=NOTIFY`，`status=QUEUED`
3. **WHEN** `menu_log` 写入成功 **THEN** 系统 **SHALL** 调用 `service.Lineman().SyncMenu(ctx, shopUUID)`
4. **WHEN** 同步触发成功 **THEN** 系统 **SHALL** 返回 HTTP 200 状态码，`status=ok`
5. **WHEN** 参数缺失或无效 **THEN** 系统 **SHALL** 返回 HTTP 400 状态码
6. **WHEN** `partnerId` 或 `storeId` 不存在 **THEN** 系统 **SHALL** 返回 HTTP 404 状态码
7. **WHEN** 系统内部错误 **THEN** 系统 **SHALL** 返回 HTTP 500 状态码

#### 具体要求

- [ ] 1.1 接口路径：`POST /v1/partners/{partnerId}/stores/{storeId}/menus/trigger-sync`
- [ ] 1.2 Header 必须包含 `Authorization: Bearer {access_token}`
- [ ] 1.3 Header 必须包含 `Content-Type: application/json`
- [ ] 1.4 响应体格式：`{ "status": "ok/fail", "code": "string", "message": "string" }`
- [ ] 1.5 Controller 路径：`ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_trigger_sync_menu.go`
- [ ] 1.6 Logic 路径：`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/`

---

### Requirement 2: menu_log 记录写入

**用户故事**: 作为系统运维人员，我想在 `menu_log` 中记录每次触发同步的请求，以便于追踪和审计。

#### 验收标准

1. **WHEN** 接收到 TriggerSyncMenu 请求 **THEN** 系统 **SHALL** 在 `menu_log` 中创建一条记录
2. **WHEN** 写入 `menu_log` **THEN** 系统 **SHALL** 设置 `sync_type=NOTIFY`
3. **WHEN** 写入 `menu_log` **THEN** 系统 **SHALL** 设置 `status=QUEUED`
4. **WHEN** 写入 `menu_log` **THEN** 系统 **SHALL** 记录 `shop_uuid`、`create_time` 等必要字段
5. **IF** 写入 `menu_log` 失败 **THEN** 系统 **SHALL** 返回错误响应并记录失败日志

#### 具体要求

- [ ] 2.1 使用 DAO 层写入 `menu_log` 表
- [ ] 2.2 `sync_type` 常量定义在 `internal/consts/consts.go`
- [ ] 2.3 `status` 常量定义在 `internal/consts/consts.go`
- [ ] 2.4 写入失败时记录错误日志（包含 `shop_uuid`）

---

### Requirement 3: 调用同步服务

**用户故事**: 作为系统，我想在记录日志后立即调用同步服务，以便于触发实际的菜单同步流程。

#### 验收标准

1. **WHEN** `menu_log` 写入成功 **THEN** 系统 **SHALL** 调用 `service.Lineman().SyncMenu(ctx, shopUUID)`
2. **IF** 调用 `SyncMenu` 失败 **THEN** 系统 **SHALL** 记录错误日志
3. **WHEN** 调用 `SyncMenu` 成功 **THEN** 系统 **SHALL** 返回成功响应

#### 具体要求

- [ ] 3.1 在 Logic 层调用 `service.Lineman().SyncMenu(ctx, shopUUID)`
- [ ] 3.2 调用失败时记录详细错误信息
- [ ] 3.3 不阻塞响应返回（异步处理或快速返回）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Controller → Logic → Service → DAO
- **单一职责原则**: Controller 仅负责参数解析和响应，Logic 负责业务编排
- **模块化设计**: Logic 和 Service 应独立且可复用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 RESTful 风格
- [ ] 响应格式符合 Lineman API 定义规范
- [ ] 错误码清晰明确（400/401/404/500）
- [ ] 参考: [Lineman API 定义及 TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1404604549#gid=1404604549)

### 数据库设计要求

- [ ] `menu_log` 表已存在，无需新增表
- [ ] 写入时使用 DAO 层方法
- [ ] 必须包含 `shop_uuid`、`sync_type`、`status`、`create_time` 字段
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 接口响应时间 < 300ms
- [ ] 数据库写入优化（批量插入或异步）
- [ ] 日志记录不阻塞主流程

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 70%
- [ ] 集成测试覆盖 TriggerSyncMenu 完整流程
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要 Bearer Token 验证
- [ ] 参数校验防止 SQL 注入
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 写入 `menu_log` 失败时优雅降级
- [ ] 错误日志记录（使用 `g.Log()`）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **TriggerSyncMenu 接口正常响应**: 返回 HTTP 200 状态码，`status=ok`
2. **menu_log 正常写入**: 记录包含正确的 `sync_type=NOTIFY` 和 `status=QUEUED`
3. **SyncMenu 正常调用**: 同步服务被触发

### 测试验收

1. **单元测试**: Logic 层覆盖率达标
2. **集成测试**: TriggerSyncMenu → menu_log → SyncMenu 完整流程测试通过
3. **手动测试**: Postman 测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 接口协议已记录在 Google Sheets
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 使用 `g.Log()` 记录日志
- 使用 `gerror` 处理错误

### 业务约束

- `sync_type=NOTIFY` 表示由外部平台主动触发
- `status=QUEUED` 表示已加入同步队列

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2

---

## 依赖关系

### 技术依赖

- `service.Lineman().SyncMenu(ctx, shopUUID)` - 菜单同步服务
- `internal/dao` - 数据访问层
- `internal/consts` - 常量定义

### 服务依赖

- **TriggerSyncMenu → menu_log**: 数据库写入
- **TriggerSyncMenu → SyncMenu**: 服务调用

### 业务依赖

- 依赖 Lineman 平台配置（`partnerId`, `storeId` 映射）
- 依赖 `service.Lineman().SyncMenu()` 已实现

---

## 风险和缓解

### 风险 1: 写入 `menu_log` 失败导致同步不可追踪

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 失败时返回错误响应并记录详细日志
- 监控 `menu_log` 写入失败率

### 风险 2: `SyncMenu` 调用失败导致同步未触发

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 记录 `SyncMenu` 调用失败日志
- 支持手动重试或告警通知

---

## 时间表

- **Phase 1 - 接口实现**: 0.5 天
- **Phase 2 - 日志记录**: 0.5 天
- **Phase 3 - 测试与文档**: 1 天
- **总计**: 2 天（SP = 2）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- [Lineman API 定义及 TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1404604549#gid=1404604549)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-15.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-15  
**作者**: rikugun  
**审核者**: 待审核
