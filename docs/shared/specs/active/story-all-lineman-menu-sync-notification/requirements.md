# LINE MAN 菜单同步通知入库需求文档

> 本文档定义 LINE MAN 菜单同步通知入库功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                   |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14-all-lineman-menu-sync-notification.md](../../../../team/proposals/2026-01/v2.14-all-lineman-menu-sync-notification.md) |
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

LINE MAN 菜单同步完成后会回调通知结果（Menu Sync Notification），当前系统未记录该通知，导致无法追踪同步结果与失败原因。本功能实现 MenuSyncNotification 接口，接收通知请求并写入 `menu_log` 表，记录 `menuSyncRequestId`、`status`、`error` 等关键信息，以便追溯同步历史与问题排查。

接口协议参考：[Lineman API 定义及 TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=571121603#gid=571121603)

## 🎯 产品对齐

该功能支持外卖渠道集成的可观测性目标，通过记录菜单同步通知结果，使运维支持人员能够快速定位同步失败原因，降低运维沟通成本，为后续监控和告警提供数据基础。

## 📝 用户故事

**作为** 运维支持人员  
**我想** 记录 LINE MAN 菜单同步通知结果  
**以便于** 快速定位同步失败原因并追踪处理情况

---

## 功能需求

### Requirement 1: MenuSyncNotification 接口实现

**用户故事**: 作为系统，我想接收 LINE MAN 菜单同步通知回调，以便于记录同步结果。

#### 验收标准

1. **WHEN** LINE MAN 平台调用 MenuSyncNotification 接口 **THEN** 系统 **SHALL** 解析 `partnerId`、`storeId` 路径参数
2. **WHEN** 解析请求体 **THEN** 系统 **SHALL** 提取 `menuSyncRequestId`、`updatedAt`、`status`、`error` 字段
3. **WHEN** `status` 为 `SUCCESS` **THEN** 系统 **SHALL** 在 `menu_log` 中记录状态为 `SUCCESS`
4. **WHEN** `status` 为 `FAILED` **THEN** 系统 **SHALL** 在 `menu_log` 中记录状态为 `FAIL`，并记录 `error` 字段内容
5. **WHEN** 记录写入成功 **THEN** 系统 **SHALL** 返回 HTTP 200 状态码，`status=ok`，`code=200`
6. **WHEN** 参数缺失或无效 **THEN** 系统 **SHALL** 返回 HTTP 400 状态码
7. **WHEN** `partnerId` 或 `storeId` 不存在 **THEN** 系统 **SHALL** 返回 HTTP 404 状态码
8. **WHEN** 系统内部错误 **THEN** 系统 **SHALL** 返回 HTTP 500 状态码

#### 具体要求

- [ ] 1.1 接口路径：`POST /v1/partners/{partnerId}/stores/{storeId}/menus/notification`
- [ ] 1.2 Header 必须包含 `Authorization: Bearer {access_token}`
- [ ] 1.3 Header 必须包含 `Content-Type: application/json`
- [ ] 1.4 请求体字段：
  - `menuSyncRequestId` (string, required): 菜单同步请求 ID
  - `updatedAt` (string, required): 更新时间（ISO 8601 格式，如 2022-11-01T13:08:06+07:00）
  - `status` (string, required): 同步状态（SUCCESS / FAILED）
  - `error` (string, optional): 错误信息（status=FAILED 时）
- [ ] 1.5 响应体格式：`{ "status": "ok/fail", "code": "string", "message": "string" }`
- [ ] 1.6 Controller 路径：`ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_menu_sync_notification.go`
- [ ] 1.7 Logic 路径：`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification.go`

---

### Requirement 2: menu_log 记录写入

**用户故事**: 作为系统运维人员，我想在 `menu_log` 中记录每次菜单同步通知，以便于追踪和审计。

#### 验收标准

1. **WHEN** 接收到 MenuSyncNotification 请求 **THEN** 系统 **SHALL** 在 `menu_log` 中创建一条记录
2. **WHEN** 写入 `menu_log` **THEN** 系统 **SHALL** 设置 `sync_type=NOTIFY`
3. **WHEN** `status` 为 `SUCCESS` **THEN** 系统 **SHALL** 设置 `status=SUCCESS`
4. **WHEN** `status` 为 `FAILED` **THEN** 系统 **SHALL** 设置 `status=FAIL` 并记录 `error_msg`
5. **WHEN** 写入 `menu_log` **THEN** 系统 **SHALL** 记录 `merchant_id`（storeId）、`request_id`（menuSyncRequestId）等必要字段
6. **IF** 写入 `menu_log` 失败 **THEN** 系统 **SHALL** 返回错误响应并记录失败日志

#### 具体要求

- [ ] 2.1 使用 `service.ChannelMenu().LogMenuSync()` 写入 `menu_log` 表
- [ ] 2.2 `sync_type` 使用常量 `consts.MenuSyncTypeNotify`（值为 `NOTIFY`）
- [ ] 2.3 `status` 映射：`SUCCESS` → `SUCCESS`，`FAILED` → `FAIL`
- [ ] 2.4 `error_msg` 字段记录 `error` 字段内容（仅 status=FAILED 时）
- [ ] 2.5 写入失败时记录错误日志（包含 `merchant_id`）
- [ ] 2.6 新增常量 `MenuSyncTypeNotify` 定义在 `internal/consts/consts.go`

---

### Requirement 3: 错误处理与日志记录

**用户故事**: 作为系统开发人员，我想确保错误被正确处理和记录，以便于问题排查。

#### 验收标准

1. **WHEN** 参数校验失败 **THEN** 系统 **SHALL** 使用 `gcode.CodeInvalidParameter` 返回 400 错误
2. **WHEN** `partnerId` 或 `storeId` 不存在 **THEN** 系统 **SHALL** 使用 `gcode.CodeNotFound` 返回 404 错误
3. **WHEN** 系统内部错误 **THEN** 系统 **SHALL** 使用 `gcode.CodeInternalError` 返回 500 错误
4. **WHEN** 发生错误 **THEN** 系统 **SHALL** 使用 `g.Log().Errorf()` 记录错误日志
5. **WHEN** 记录成功 **THEN** 系统 **SHALL** 使用 `g.Log().Infof()` 记录成功日志

#### 具体要求

- [ ] 3.1 使用 `gerror` 处理错误（不用标准库 errors）
- [ ] 3.2 错误日志包含 `merchant_id`、`request_id`、`status`、`error` 等关键信息
- [ ] 3.3 成功日志包含 `merchant_id`、`request_id`、`status`
- [ ] 3.4 所有日志使用中文描述

---

## 非功能需求

### 性能要求

- **响应时间**: MenuSyncNotification 接口响应时间 < 500ms (P95)
- **并发能力**: 支持 10 QPS（单店）

### 可观测性

- **日志级别**: 
  - 成功：Info
  - 失败：Error
- **日志内容**: 包含 `merchant_id`、`request_id`、`status`、`error`
- **监控指标**: 无需新增（复用现有日志监控）

### 安全要求

- **认证**: 使用 Bearer Token 认证（由网关/中间件处理）
- **授权**: 校验 `partnerId` 和 `storeId` 有效性
- **审计**: 所有请求记录到 `menu_log` 表

---

## 技术约束

- **框架**: GoFrame 2.x
- **架构**: Controller → Logic → Service → DAO
- **错误处理**: 使用 `gerror` 包
- **日志**: 使用 `g.Log()` 包
- **禁止修改**: `dao/`, `model/entity/`, `model/do/`, `service/` 目录（自动生成）
- **数据库表**: 复用现有 `takeout_menu_log` 表

---

## 验收测试场景

### 测试场景 1: 成功通知

**前置条件**: 
- storeId 已配置 Lineman
- menuSyncRequestId 存在

**测试步骤**:
1. 调用 MenuSyncNotification 接口
2. 传入 `status=SUCCESS`

**预期结果**:
- HTTP 200
- 响应 `status=ok`
- `menu_log` 中记录 `sync_type=NOTIFY`, `status=SUCCESS`

---

### 测试场景 2: 失败通知

**前置条件**: 
- storeId 已配置 Lineman
- menuSyncRequestId 存在

**测试步骤**:
1. 调用 MenuSyncNotification 接口
2. 传入 `status=FAILED`, `error="Invalid menu format"`

**预期结果**:
- HTTP 200
- 响应 `status=ok`
- `menu_log` 中记录 `sync_type=NOTIFY`, `status=FAIL`, `error_msg="Invalid menu format"`

---

### 测试场景 3: 参数缺失

**测试步骤**:
1. 调用 MenuSyncNotification 接口
2. 缺少 `menuSyncRequestId` 字段

**预期结果**:
- HTTP 400
- 响应 `status=fail`, `code=400`, `message` 包含错误描述

---

### 测试场景 4: 门店不存在

**测试步骤**:
1. 调用 MenuSyncNotification 接口
2. 使用不存在的 `storeId`

**预期结果**:
- HTTP 404
- 响应 `status=fail`, `code=404`, `message` 包含错误描述

---

## 相关文档

- [Go BMP 开发规范](../../../../.cursor/rules/go-bmp.mdc)
- [API 设计规范](../../../../.cursor/rules/api.mdc)
- [数据库开发规范](../../../../.cursor/rules/database.mdc)
- [Lineman API 协议](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=571121603#gid=571121603)

---

## 附录

### 状态映射表

| Lineman Status | menu_log Status | 说明 |
|----------------|-----------------|------|
| SUCCESS        | SUCCESS         | 同步成功 |
| FAILED         | FAIL            | 同步失败 |

### 字段映射表

| 接口字段 (Lineman) | 数据库字段 (menu_log) | 类型 | 说明 |
|-------------------|----------------------|------|------|
| menuSyncRequestId | request_id           | string | 请求 ID |
| updatedAt         | - (不存储)           | string | 更新时间 |
| status            | status               | string | 同步状态 |
| error             | error_msg            | string | 错误信息 |
| storeId (路径)    | merchant_id          | string | 商户 ID |
| partnerId (路径)  | - (不存储)           | string | 合作伙伴 ID |

---

**版本**: v1.0.0  
**创建日期**: 2026-01-15  
**最后更新**: 2026-01-15
