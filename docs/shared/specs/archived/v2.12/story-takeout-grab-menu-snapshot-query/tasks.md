# Grab 菜单快照数据查询 gRPC 服务 任务分解

> 本文档定义 Grab 菜单快照数据查询 gRPC 服务 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 7
**已完成**: 5
**进行中**: -
**完成率**: 71%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/{Timestamp}_alter_channel_menu_snapshot_table.up.sql` (和 down.sql)
  - Purpose: 修改 `channel_menu_snapshot` 表结构，重命名时间字段并新增字段
  - Requirements: 数据库变更要求
  - Leverage: 现有 migration 文件
  - Prompt: Role: Database Engineer | Task: 创建 channel_menu_snapshot 表变更的 up.sql 和 down.sql | Context: 重命名 create_time->created_at, update_time->updated_at; 新增 deleted_at, sync_state, ttpos_menu_data, ttpos_updated_at | Restrictions: 遵循 MySQL 语法 | Success: 迁移文件创建成功，SQL 语法正确

- [ ] 1.2 执行数据库迁移与生成 DAO
  - File: `ttpos-bmp/app/ttpos-takeout/internal/dao/`, `internal/model/entity/`, `internal/model/do/`
  - Purpose: 应用数据库变更并更新 GoFrame 代码
  - Requirements: 数据库变更要求
  - Command: 
    1. 手动执行 SQL 或使用迁移工具应用变更
    2. `cd ttpos-bmp/app/ttpos-takeout && make dao`
  - Success: 数据库表结构已更新，Go 代码已重新生成且无编译错误

---

## Phase 2: 接口定义与生成

- [x] 2.1 修改 Protobuf 定义
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto`
  - Purpose: 定义 `GetMenuSnapshot` gRPC 接口
  - Requirements: Requirement 1.1, 1.2, 1.3
  - Prompt: Role: gRPC Developer | Task: 在 Takeout service 中新增 GetMenuSnapshot 方法 | Context: Req 包含 provider_name, shop_uuid, request_id; Res 包含 content, updated_at, sync_state, ttpos_menu_data, ttpos_updated_at | Restrictions: 遵循 proto3 语法 | Success: Proto 文件更新正确

- [ ] 2.2 生成 gRPC Go 代码
  - File: `ttpos-bmp/app/ttpos-takeout/api/takeout/`
  - Purpose: 生成 gRPC 桩代码
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao` (通常 make dao 会包含 proto 生成，或者单独 `make proto`，需确认 Makefile)
  - Success: API Go 代码生成成功

---

## Phase 3: 业务逻辑实现

- [x] 3.1 实现 Logic 层
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go` (或其他合适的 logic 文件)
  - Purpose: 实现查询逻辑
  - Requirements: Requirement 1.4
  - Prompt: Role: Go Developer (GoFrame) | Task: 实现 GetMenuSnapshot 业务逻辑 | Context: 根据 request_id 查询 DAO，校验 provider_name 和 shop_uuid，处理 NotFound | Restrictions: 使用 dao 查询，不写原生 SQL | Success: Logic 方法实现完成

- [x] 3.2 实现 RPC Controller
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/takeout.go` (或对应 controller)
  - Purpose: 实现 gRPC 接口入口，调用 Logic
  - Requirements: Requirement 1
  - Prompt: Role: Go Developer (GoFrame) | Task: 实现 GetMenuSnapshot RPC 方法 | Context: 解析请求，调用 Logic 层，返回响应 | Restrictions: 处理 error 并返回 gRPC status code | Success: Controller 方法实现完成

---

## Phase 4: 测试与验证

- [x] 4.1 编写单元测试
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_test.go`
  - Purpose: 验证 Logic 逻辑
  - Requirements: 测试验收
  - Prompt: Role: QA Engineer | Task: 编写 GetMenuSnapshot 单元测试 | Context: Mock DAO，测试 正常/未找到/参数错误 场景 | Success: 测试通过

- [ ] 4.2 本地接口测试
  - File: -
  - Purpose: 验证端到端功能
  - Requirements: 功能验收
  - Action: 使用 grpcurl 或编写简单 client main.go 调用本地服务
  - Success: 能成功调通接口并返回预期数据

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0
**创建日期**: 2025-12-11
