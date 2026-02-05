# task-main-order-operation-duration 技术设计

## 📋 概述

| 项目       | 内容                               |
| ---------- | ---------------------------------- |
| Spec ID    | task-main-order-operation-duration |
| 设计人     | xiezhihuan                         |
| 设计日期   | 2026-02-05                         |
| 总 SP      | 3                                  |

---

## 🔄 代码复用分析

### 可复用代码

| 文件                              | 说明               | 复用方式 |
| --------------------------------- | ------------------ | -------- |
| `main/pkg/utils/instance.go`      | GetInstanceID()    | 直接调用 |
| `main/pkg/database/db_manager.go` | GetDB() 获取数据库 | 直接调用 |
| `main/pkg/utils/uuid.go`          | GenUuid() 生成 ID  | 直接调用 |
| `main/pkg/logger/logger.go`       | 日志记录           | 直接调用 |

### 需要新建

| 文件                                            | 说明                   |
| ----------------------------------------------- | ---------------------- |
| `main/app/model/order_operation_duration.go`    | 数据模型               |
| `main/app/dto/req/operation_duration.go`        | DTO 数据结构           |
| `main/app/queue/operation_duration_queue.go`    | 内存队列 + 消费者      |
| `main/app/api/helper/duration_tracker.go`       | Handler 层辅助函数     |
| `admin/database/migrations/xxx.php`             | 数据库迁移文件         |

---

## 🏗️ 架构设计

### 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Handler Layer                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐             │
│  │ OrderHandler│    │ BillHandler │    │ OtherHandler│             │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘             │
│         │                  │                  │                     │
│         └──────────────────┼──────────────────┘                     │
│                            ▼                                        │
│                 ┌─────────────────────┐                             │
│                 │  DurationTracker    │                             │
│                 │  (helper)           │                             │
│                 └──────────┬──────────┘                             │
└────────────────────────────┼────────────────────────────────────────┘
                             │ Push (非阻塞)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Queue Layer                                   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              OperationDurationQueue                          │   │
│  │  ┌──────────────────┐    ┌──────────────────┐              │   │
│  │  │ chan (10000)     │───►│ consume()        │              │   │
│  │  │ 内存缓冲          │    │ 消费者协程        │              │   │
│  │  └──────────────────┘    └────────┬─────────┘              │   │
│  └───────────────────────────────────┼──────────────────────────┘   │
└──────────────────────────────────────┼──────────────────────────────┘
                                       │ 批量写入 (100条/5秒)
                                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Database Layer                                │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              SaaS 主库 (constant.DefaultDB)                  │   │
│  │  ┌──────────────────────────────────────────────────────┐  │   │
│  │  │  ttpos_order_operation_duration                       │  │   │
│  │  └──────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 数据流

```
1. Handler 开始 → StartTrack("action")
2. Handler 执行业务逻辑
3. Handler 结束 → tracker.End(ctx, err)
4. End() 计算耗时 → Push 到 channel (非阻塞)
5. 消费者累积 100 条或 5 秒超时 → 批量 INSERT
```

---

## 🧩 组件和接口

### Component 1: OperationDurationQueue

**位置**: `main/app/queue/operation_duration_queue.go`

```go
type OperationDurationQueue struct {
    ch         chan *req.OperationDurationRecord
    db         *gorm.DB
    batchSize  int
    flushTime  time.Duration
    instanceId string
}

// 公开方法
func NewOperationDurationQueue(dbm *database.DBManager) *OperationDurationQueue
func (q *OperationDurationQueue) Push(record *req.OperationDurationRecord)
func (q *OperationDurationQueue) Start()  // 启动消费者
func (q *OperationDurationQueue) Stop()   // 优雅关闭

// 私有方法
func (q *OperationDurationQueue) consume()
func (q *OperationDurationQueue) flush(records []*req.OperationDurationRecord)
```

### Component 2: DurationTracker

**位置**: `main/app/api/helper/duration_tracker.go`

```go
type DurationTracker struct {
    startTime     int64
    action        string
    requestPath   string
    saleBillUuid  uint64
    saleOrderUuid uint64
}

// 公开方法
func StartTrack(action string) *DurationTracker
func (t *DurationTracker) WithBill(saleBillUuid, saleOrderUuid uint64) *DurationTracker
func (t *DurationTracker) WithPath(path string) *DurationTracker
func (t *DurationTracker) End(ctx *context.Context, err error)
```

---

## 📊 数据模型

### Model: OrderOperationDuration

**位置**: `main/app/model/order_operation_duration.go`

```go
type OrderOperationDuration struct {
    Id            uint64 `gorm:"primaryKey;autoIncrement"`
    Uuid          uint64 `gorm:"uniqueIndex"`
    CompanyUuid   uint64 `gorm:"index"`
    SaleBillUuid  uint64 `gorm:"index"`
    SaleOrderUuid uint64
    Action        string `gorm:"size:100;index"`
    Source        string `gorm:"size:50"`
    StaffUuid     uint64
    DeviceSn      string `gorm:"size:255"`
    InstanceId    string `gorm:"size:128;index"`
    StartTime     int64
    EndTime       int64
    DurationMs    int
    RequestPath   string `gorm:"size:255"`
    Status        int
    ErrorMsg      string `gorm:"size:500"`
    CreateTime    int    `gorm:"autoCreateTime"`
    UpdateTime    int    `gorm:"autoUpdateTime"`
    DeleteTime    int    `gorm:"default:0"`
}

func (OrderOperationDuration) TableName() string {
    return "ttpos_order_operation_duration"
}
```

### DTO: OperationDurationRecord

**位置**: `main/app/dto/req/operation_duration.go`

```go
type OperationDurationRecord struct {
    CompanyUuid   uint64
    SaleBillUuid  uint64
    SaleOrderUuid uint64
    Action        string
    Source        string
    StaffUuid     uint64
    DeviceSn      string
    InstanceId    string
    StartTime     int64
    EndTime       int64
    DurationMs    int64
    RequestPath   string
    Status        int
    ErrorMsg      string
}
```

---

## 🔌 集成点

### QueueManager 集成

**位置**: `main/app/service/queue_manager.go`

```go
type QueueManager struct {
    // 现有队列...
    OperationDurationQueue *queue.OperationDurationQueue  // 新增
}

func NewQueueManager(dbm *database.DBManager) *QueueManager {
    qm := &QueueManager{
        // 现有初始化...
        OperationDurationQueue: queue.NewOperationDurationQueue(dbm),
    }
    qm.OperationDurationQueue.Start()
    return qm
}
```

### Handler 使用示例

```go
func (h *OrderHandler) CancelOrder(c *gin.Context) {
    ctx := helper.GetContext(c)
    req := req.OrderCancelReq{}
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.HandleValidationError(c, err, req, nil)
        return
    }

    // 开始记录
    tracker := helper.StartTrack("cancel_order").
        WithBill(req.SaleBillUuid, req.SaleOrderUuid).
        WithPath(c.Request.URL.Path)

    err := h.orderSrv.CancelOrder(ctx, req)

    // 结束记录（异步推送，不阻塞）
    tracker.End(ctx, err)

    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    helper.Success(c, gin.H{})
}
```

---

## ⚠️ 风险识别

| 风险                   | 影响 | 缓解措施                                     |
| ---------------------- | ---- | -------------------------------------------- |
| 内存队列满时数据丢失   | 低   | 队列容量 10000，丢弃时输出 WARN 日志         |
| 大量写入影响 SaaS 库   | 中   | 批量写入 + 单消费者控制速率                  |
| 服务重启数据丢失       | 低   | 辅助数据，可接受少量丢失                     |
| 消费者协程 panic       | 中   | 使用 utils.Go 包裹，内置 recover             |

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**: 80%+

| 测试文件                                      | 测试内容               |
| --------------------------------------------- | ---------------------- |
| `main/app/queue/operation_duration_queue_test.go` | 队列 Push/消费/批量写入 |
| `main/app/api/helper/duration_tracker_test.go`    | Tracker 耗时计算       |

### 测试用例

1. **队列 Push 测试**
   - 正常 Push 成功
   - 队列满时不阻塞

2. **批量写入测试**
   - 达到 100 条触发写入
   - 超时 5 秒触发写入

3. **DurationTracker 测试**
   - 耗时计算正确
   - 错误信息截断到 500 字符

### 测试命令

```bash
cd main && go test -coverprofile=coverage.out ./app/queue/...
cd main && go test -coverprofile=coverage.out ./app/api/helper/...
cd main && go tool cover -html=coverage.out
```

---

## 📁 文件清单

| 文件路径                                                     | 类型     | 说明               |
| ------------------------------------------------------------ | -------- | ------------------ |
| `main/app/model/order_operation_duration.go`                 | 新建     | 数据模型           |
| `main/app/dto/req/operation_duration.go`                     | 新建     | DTO 结构           |
| `main/app/queue/operation_duration_queue.go`                 | 新建     | 队列组件           |
| `main/app/api/helper/duration_tracker.go`                    | 新建     | Handler 辅助函数   |
| `main/app/service/queue_manager.go`                          | 修改     | 注册新队列         |
| `admin/database/migrations/20260205_create_order_operation_duration.php` | 新建     | 迁移文件           |

---

**版本**: v1.0.0
**设计日期**: 2026-02-05
