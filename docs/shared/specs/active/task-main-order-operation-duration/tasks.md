# task-main-order-operation-duration 任务清单

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 3    |
| 总任务数 | 7    |
| 已完成   | 7    |
| 完成率   | 100% |

---

## Phase 1: 基础设施

### 1.1 创建数据库迁移文件

| 项目         | 内容                                                              |
| ------------ | ----------------------------------------------------------------- |
| File         | `admin/database/migrations/20260205_create_order_operation_duration.php` |
| Purpose      | 在 SaaS 主库创建 ttpos_order_operation_duration 表                |
| Requirements | Req1 记录字段、数据库约束                                         |
| Leverage     | 参考现有迁移文件格式                                              |

**注意事项**:
- `const TARGET = 'main';` 仅应用到 saas 主库
- 表名不带 `ttpos_` 前缀（迁移文件会自动添加）

- [x] 完成

---

### 1.2 创建数据模型

| 项目         | 内容                                           |
| ------------ | ---------------------------------------------- |
| File         | `main/app/model/order_operation_duration.go`   |
| Purpose      | 定义 OrderOperationDuration GORM 模型          |
| Requirements | Req1 记录字段                                  |
| Leverage     | 参考现有 model 格式                            |

**字段清单**:
- id, uuid, company_uuid, sale_bill_uuid, sale_order_uuid
- action, source, staff_uuid, device_sn, instance_id
- start_time, end_time, duration_ms
- request_path, status, error_msg
- create_time, update_time, delete_time

- [x] 完成

---

### 1.3 创建 DTO 结构

| 项目         | 内容                                     |
| ------------ | ---------------------------------------- |
| File         | `main/app/dto/req/operation_duration.go` |
| Purpose      | 定义队列传输的数据结构                   |
| Requirements | Req1 记录字段                            |
| Leverage     | -                                        |

- [x] 完成

---

### 1.4 创建队列组件

| 项目         | 内容                                            |
| ------------ | ----------------------------------------------- |
| File         | `main/app/queue/operation_duration_queue.go`    |
| Purpose      | 内存队列 + 消费者协程 + 批量写入                |
| Requirements | Req2 内存队列缓冲、Req3 异步批量写入、Req4 分布式追踪 |
| Leverage     | `main/pkg/utils/instance.go`, `main/pkg/database/db_manager.go` |

**核心逻辑**:
1. `NewOperationDurationQueue()` - 初始化队列
2. `Push()` - 非阻塞推送（select + default）
3. `consume()` - 消费者循环（使用 utils.Go）
4. `flush()` - 批量写入数据库

**配置参数**:
- 队列容量: 10000
- 批量阈值: 100
- 刷新间隔: 5 秒

- [x] 完成

---

### 1.5 注册队列到 QueueManager

| 项目         | 内容                                 |
| ------------ | ------------------------------------ |
| File         | `main/app/service/queue_manager.go`  |
| Purpose      | 将新队列注册到全局队列管理器         |
| Requirements | -                                    |
| Leverage     | 参考现有队列注册方式                 |

**修改内容**:
1. 添加 `OperationDurationQueue` 字段
2. 在 `NewQueueManager()` 中初始化
3. 启动消费者协程

- [x] 完成

---

## Phase 2: Handler 集成

### 2.1 创建 DurationTracker Helper

| 项目         | 内容                                        |
| ------------ | ------------------------------------------- |
| File         | `main/app/api/helper/duration_tracker.go`   |
| Purpose      | 提供 Handler 层简洁的耗时记录 API           |
| Requirements | Req1 耗时记录器                             |
| Leverage     | -                                           |

**公开方法**:
```go
func StartTrack(action string) *DurationTracker
func (t *DurationTracker) WithBill(saleBillUuid, saleOrderUuid uint64) *DurationTracker
func (t *DurationTracker) WithPath(path string) *DurationTracker
func (t *DurationTracker) End(ctx *context.Context, err error)
```

**使用示例**:
```go
tracker := helper.StartTrack("cancel_order").WithBill(req.SaleBillUuid, 0)
err := service.DoSomething()
tracker.End(ctx, err)  // 异步推送，不阻塞
```

- [x] 完成

---

### 2.2 集成到关键 Handler（可选）

| 项目         | 内容                                           |
| ------------ | ---------------------------------------------- |
| File         | `main/app/api/v1/cashier/cashier_order.go` 等  |
| Purpose      | 在关键订单操作中添加耗时记录                   |
| Requirements | 验收标准                                       |
| Leverage     | DurationTracker                                |

**已集成的 Handler**:
- [x] CancelOrder - 取消订单 (`cashier_order.go`)
- [x] ReturnOrder - 退款订单 (`cashier_order.go`)
- [x] ReverseSettle - 反结账 (`cashier_order.go`)
- [x] OrderPaymentFinish - 结账 (`cashier_desk.go`, `cashier_instant.go`)
- [x] OrderDiscount - 订单折扣 (`cashier_desk.go`, `cashier_instant.go`)
- [x] MergeDesk - 并单 (`cashier_desk.go`)

**集成方式（3 行代码）**:
```go
tracker := helper.StartTrack("action").WithBill(uuid1, uuid2).WithPath(path)
err := h.service.Method(ctx, req)
tracker.End(ctx, err)
```

- [x] 完成

---

## 提交清单

### 代码质量

- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性

- [x] Req1 耗时记录器：记录开始/结束时间、计算耗时、记录错误信息
- [x] Req2 内存队列缓冲：非阻塞推送、队列满丢弃并告警
- [x] Req3 异步批量写入：100 条或 5 秒触发、使用 SaaS 库连接
- [x] Req4 分布式追踪：自动填充 instance_id

### 迁移同步

- [x] 迁移文件已创建
- [ ] 执行 `php think migrate:run` 验证

### 日志规范

- [x] WARN 日志包含 company_uuid（队列满时）
- [x] ERROR 日志包含 count（写入失败时）

---

## 验收测试

### 手动测试步骤

1. **启动服务**
   ```bash
   cd main && go run main.go
   ```

2. **触发订单操作**（调用已集成 DurationTracker 的 API）

3. **检查数据库**
   ```sql
   SELECT * FROM ttpos_order_operation_duration ORDER BY id DESC LIMIT 10;
   ```

4. **验证字段**
   - company_uuid 正确
   - start_time, end_time, duration_ms 正确
   - instance_id 非空
   - status 正确（成功 1 / 失败 0）

5. **测试队列满场景**（可选）
   - 模拟高并发，检查是否输出 WARN 日志
   - 检查业务不受影响

---

**版本**: v1.0.0
**创建日期**: 2026-02-05
