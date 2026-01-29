# LINE MAN 订单更新 Webhook 任务分解

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移脚本
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/update_order_add_order_updated_time.sql`
  - Purpose: 添加 `order_updated_time` 字段
  - Requirements: Req 2.2

- [ ] 1.2 执行数据库迁移
  - Purpose: 在数据库中添加字段
  - Success: 字段添加成功

- [ ] 1.3 更新 Entity（自动生成）
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen dao`
  - Purpose: 重新生成 DAO 和 Entity

---

## Phase 2: 常量和接口定义

- [ ] 2.1 添加常量
  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - Purpose: 添加 `OrderActionUpdate = "update"`

- [ ] 2.2 更新 Service 接口
  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman_order.go`
  - Purpose: 添加 `HandleOrderUpdate(ctx, req) error` 方法

---

## Phase 3: Logic 层实现

- [ ] 3.1 实现 `HandleOrderUpdate` 方法
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 实现订单更新业务逻辑
  - 步骤:
    1. 查询现有订单
    2. 幂等性检查
    3. 调用 `updateOrder()` 更新
    4. 发送 RocketMQ 事件

- [ ] 3.2 实现 `updateOrder` 方法
  - File: 同上
  - Purpose: 事务更新订单数据

---

## Phase 4: Controller 层实现

- [ ] 4.1 实现 Controller
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_order_update.go`
  - Purpose: 接收 Webhook，调用 Service

---

## Phase 5: 测试

- [ ] 5.1 单元测试 - Logic 层
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go`
  - 测试用例:
    - `TestHandleOrderUpdate_Success`
    - `TestHandleOrderUpdate_OrderNotFound`
    - `TestHandleOrderUpdate_Idempotent`

- [ ] 5.2 集成测试
  - Purpose: 端到端测试

- [ ] 5.3 手动测试
  - Tool: Postman

---

## Phase 6: 文档和发布

- [ ] 6.1 更新 API 文档
- [ ] 6.2 代码审查
- [ ] 6.3 发布到测试环境

---

**创建日期**: 2026-01-12  
**作者**: rikugun
