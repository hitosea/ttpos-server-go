# GrabFood 外卖平台对接 (v1.1.3) 任务分解

> 本文档定义 GrabFood 对接功能的详细执行任务清单。

## 📊 进度总览

**总任务数**: 13
**已完成**: 0

---

## Phase 1: 数据库与模型 (ttpos-takeout)

- [ ] 1.1 创建数据库迁移文件

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/{YYYYMMDDHHMMSS}_create_takeout_order_table.up.sql`
  - Purpose: 创建 `takeout_order`, `takeout_order_item`, `takeout_menu_log`, `takeout_order_status_log` 表
  - Requirements: 2.2, 菜单保存, 状态变更记录
  - Prompt: Role: SQL Expert | Task: Create migration SQL for takeout_order, takeout_order_item, takeout_menu_log and takeout_order_status_log tables | Context: takeout_order stores Grab order details, takeout_menu_log stores menu sync snapshots, takeout_order_status_log stores status changes | Restrictions: Use snake_case, InnoDB, utf8mb4

- [ ] 1.2 执行数据库迁移与生成 DAO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/dao/`
  - Purpose: 生成 GoFrame DAO/Entity/DO 代码
  - Requirements: 2.2
  - Command: `gf gen dao` (in ttpos-takeout module)

- [ ] 1.3 定义 Grab DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/grab_dto.go`
  - Purpose: 定义 Grab API 交互的数据结构 (Webhook Payloads, Menu Structure)
  - Requirements: 2.1, 1.1

---

## Phase 2: Webhook 与持久化 (ttpos-takeout)

- [ ] 2.1 实现签名验证逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/auth.go`
  - Purpose: 实现 HMAC-SHA256 签名验证
  - Requirements: 2.1

- [ ] 2.2 实现订单保存逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
  - Purpose: 将 DTO 转换为 Entity 并保存到数据库（事务）
  - Requirements: 2.3
  - Prompt: Role: Go Developer | Task: Implement SaveOrder function | Context: Use transaction to save Order and Items, store raw JSON in raw_data

- [ ] 2.3 实现 MQ 消息发送

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
  - Purpose: 发送标准消息到 RocketMQ
  - Requirements: 2.4

- [ ] 2.4 实现状态变更处理逻辑 (New)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
  - Purpose: 处理订单状态变更回调，保存日志，更新主表，发送 MQ
  - Requirements: 2.3, 状态记录
  - Prompt: Role: Go Developer | Task: Implement HandleOrderStatusUpdate | Context: Save to takeout_order_status_log, update takeout_order.status, send MQ event

- [ ] 2.5 实现 Webhook Controller

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/webhook/grab_controller.go`
  - Purpose: 组合验证、保存、发送逻辑 (支持 Submit Order 和 Push Order State)
  - Requirements: 2.1

- [ ] 2.6 注册 Webhook 路由

  - File: `ttpos-bmp/app/ttpos-takeout/internal/cmd/cmd.go`
  - Purpose: 注册路由
  - Requirements: 2.1

---

## Phase 3: 菜单同步与 API Client (ttpos-takeout)

- [ ] 3.1 实现 Grab API Client

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab/client.go`
  - Purpose: 封装 Grab API endpoints 调用 (UpdateMenu, UpdateStoreStatus, etc.)
  - Requirements: 1.2, 3.1
  - Prompt: Role: Go Developer | Task: Implement GrabClient struct | Context: Handle Auth (Client Credentials), HTTP requests, Error handling, Retries | Restrictions: Use standard http or gclient

- [ ] 3.2 实现菜单数据转换与保存

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
  - Purpose: 将内部 Menu 转换为 Grab 格式，保存快照到 `takeout_menu_log`，调用 Client 推送
  - Requirements: 1.1, 菜单保存
  - Prompt: Role: Go Developer | Task: Implement SyncMenu function | Context: Convert menu, save snapshot to DB, call GrabClient.UpdateMenu

- [ ] 3.3 实现门店状态控制

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/store_service.go`
  - Purpose: 调用 Client 暂停/恢复营业
  - Requirements: 3.1

---

## 提交清单

- [ ] 代码通过 `go fmt` 和 `go vet`
- [ ] 单元测试通过
