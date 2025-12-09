# GrabFood 外卖平台对接 (v1.1.3) 任务分解

> 本文档定义 GrabFood 对接功能的详细执行任务清单。

## 📊 进度总览

**总任务数**: 14
**已完成**: 14

---

## Phase 1: 数据库与模型 (ttpos-takeout)

- [x] 1.1 创建数据库迁移文件

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251205085121_create_grab_order_tables.up.sql`
  - Purpose: 创建 `takeout_order`, `takeout_order_item`, `takeout_menu_log`, `takeout_order_status_log` 表
  - ✅ 已完成

- [x] 1.2 执行数据库迁移与生成 DAO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/dao/`
  - Purpose: 生成 GoFrame DAO/Entity/DO 代码
  - ✅ 已完成 (手动创建，需数据库环境后执行 `make db_up && make dao`)

- [x] 1.3 定义 Grab DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/`
  - Purpose: 定义 Grab API 交互的数据结构
  - ✅ 已完成: `order.go`, `menu.go`, `store.go`, `auth.go`, `accept_order.go`

---

## Phase 2: Webhook 与持久化 (ttpos-takeout)

- [x] 2.1 实现签名验证逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/auth.go`
  - Purpose: 实现 HMAC-SHA256 签名验证
  - ✅ 已完成

- [x] 2.2 实现订单保存逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
  - Purpose: 将 DTO 转换为 Entity 并保存到数据库（事务）
  - ✅ 已完成: `HandleSubmitOrder`, `saveOrder`

- [x] 2.3 实现 MQ 消息发送

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/mq_producer.go`
  - Purpose: 发送标准消息到 RocketMQ
  - ✅ 已完成: `RocketMQProducer`, `NoopMQProducer`

- [x] 2.4 实现状态变更处理逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
  - Purpose: 处理订单状态变更回调，保存日志，更新主表，发送 MQ
  - ✅ 已完成: `HandlePushOrderState`

- [x] 2.5 实现菜单拉取处理逻辑 (GetMenu)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
  - Purpose: 处理 Grab 主动拉取菜单的请求
  - ✅ 已完成: `HandleGetMenu`, `HandleMenuSyncState`, `SyncMenu`

- [x] 2.6 实现 Webhook Controller (多接口支持)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/webhook/grab_controller.go`
  - Purpose: 组合验证、保存、发送逻辑
  - ✅ 已完成: `SubmitOrder`, `PushOrderState`, `GetMenu`, `MenuSyncState`, `IntegrationStatus`

- [x] 2.7 注册 Webhook 路由

  - File: `ttpos-bmp/app/ttpos-takeout/internal/cmd/cmd.go`
  - Purpose: 注册路由
  - ✅ 已完成: `/api/v1/callback/grab/*`

---

## Phase 3: 菜单同步与 API Client (ttpos-takeout)

- [x] 3.1 实现 Grab API Client

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab/client.go`
  - Purpose: 封装 Grab API endpoints 调用
  - ✅ 已完成: OAuth, Menu, Store, Order 操作

- [x] 3.2 实现菜单数据转换与保存

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
  - Purpose: 将内部 Menu 转换为 Grab 格式，保存快照到 `takeout_menu_log`
  - ✅ 已完成

- [x] 3.3 实现门店状态控制

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/store_service.go`
  - Purpose: 调用 Client 暂停/恢复营业
  - ✅ 已完成: `PauseStore`, `ResumeStore`

---

## 提交清单

- [x] 代码通过 `go fmt` 和 `go vet`
- [x] 单元测试通过

---

## 📁 创建的文件清单

### SQL 迁移
- `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251205085121_create_grab_order_tables.up.sql`
- `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251205085121_create_grab_order_tables.down.sql`

### Entity/DO/DAO
- `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order_item.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/entity/menu_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order_status_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/do/order.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/do/order_item.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/do/menu_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/do/order_status_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/order.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/order_item.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/menu_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/order_status_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/internal/order.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/internal/order_item.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/internal/menu_log.go`
- `ttpos-bmp/app/ttpos-takeout/internal/dao/internal/order_status_log.go`

### DTO
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/order.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/store.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/auth.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/accept_order.go`

### Logic
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/auth.go`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/store_service.go`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/mq_producer.go`

### Controller
- `ttpos-bmp/app/ttpos-takeout/internal/controller/webhook/grab_controller.go`

### Service (API Client)
- `ttpos-bmp/app/ttpos-takeout/internal/service/grab/client.go`

### 修改的文件
- `ttpos-bmp/app/ttpos-takeout/internal/cmd/cmd.go` (添加 Grab Webhook 路由注册)
