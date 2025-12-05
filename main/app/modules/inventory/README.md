# 库存领域（Inventory Domain）

## 概述

库存领域负责管理仓库和库存物品，是 TTPOS 系统的核心业务领域之一。

---

## 聚合列表

| 聚合 | 文件 | 说明 |
|------|------|------|
| **Warehouse** | `entity/warehouse.go` | 仓库管理 |
| **WarehouseItem** | `entity/warehouse_item.go` | 库存物品管理 |

---

## 目录结构

```
domain/inventory/
├── entity/                    # 聚合根实体
│   ├── warehouse.go           # 仓库聚合根
│   ├── warehouse_test.go      # 仓库单元测试
│   └── warehouse_item.go      # 库存物品聚合根
├── valueobject/               # 值对象（不可变）
│   ├── warehouse_code.go      # 仓库编码
│   ├── warehouse_status.go    # 仓库状态
│   ├── warehouse_type.go      # 仓库类型
│   ├── contact_info.go        # 联系信息
│   ├── multi_language_name.go # 多语言名称
│   └── stock.go               # 库存数量
├── repository/                # Repository 接口
│   ├── warehouse_repository.go
│   ├── warehouse_item_repository.go
│   └── multi_language_name_repository.go
├── service/                   # 领域服务
│   ├── warehouse_domain_service.go
│   ├── warehouse_domain_service_test.go
│   ├── warehouse_item_domain_service.go
│   └── erp_integration_service.go
├── specification/             # 规格模式
│   └── warehouse_specification.go
└── event/                     # 领域事件
    ├── event.go
    ├── warehouse_created.go
    ├── warehouse_deleted.go
    └── default_warehouse_changed.go
```

---

## ⚠️ Context 约束

**在 modules 目录中，所有方法必须使用自定义的 `pkg/context.Context`，禁止使用标准库的 `context.Context`。**

```go
import "ttpos-server-go/pkg/context"

// ✅ 正确：使用自定义 Context
func (r *WarehouseRepositoryImpl) Save(ctx context.Context, warehouse *entity.Warehouse) error {
    db := ctx.GetDB() // 直接使用 ctx 获取数据库连接
    // ...
}

// ❌ 错误：使用标准库 context
import "context"
```

**调用外部服务时**，如 ERP、RPC，使用 `ctx.GetContext()` 转换：
```go
erpCode, err := erpSrv.CreateWarehouse(ctx.GetContext(), req)
```

详细规范参见：`.cursor/rules/go-modules.mdc`

---

## Warehouse 聚合

### 核心能力

- 仓库创建、编辑、删除
- 设置默认仓库
- 仓库启用/禁用
- 仓库类型管理（普通/在途）

### 业务规则

1. **默认仓库**：系统只能有一个默认仓库，且不能被禁用或删除
2. **在途仓库**：不能被禁用、删除或设置为默认
3. **编码规则**：自动转大写，长度 2-20 字符，必须唯一

### 使用示例

```go
import (
    "ttpos-server-go/app/modules/inventory/domain/entity"
    "ttpos-server-go/app/modules/inventory/domain/valueobject"
)

// 创建仓库
code, _ := valueobject.NewWarehouseCode("WH001")
name := valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "主仓库"})
warehouse := entity.NewWarehouse(name, valueobject.TypeNormal, code)

// 设置为默认仓库
warehouse.SetAsDefault()

// 禁用仓库
err := warehouse.Deactivate()  // 默认仓库会返回错误

// 检查是否可删除
canDelete, reason := warehouse.CanBeDeleted()
```

---

## WarehouseItem 聚合

### 核心能力

- 库存查询（单个/批量/按仓库）
- 入库操作
- 出库操作
- 库存预留/释放
- 仓库间调拨

### 业务规则

1. **库存不能为负**：出库时检查库存是否充足
2. **预留不超过可用**：预留库存不能超过可用库存
3. **调拨需足够库存**：源仓库可用库存必须大于调拨数量
4. **精度控制**：库存数量保留两位小数

### 使用示例

```go
import (
    "ttpos-server-go/app/modules/inventory/domain/entity"
    "ttpos-server-go/app/modules/inventory/domain/valueobject"
)

// 创建库存物品
item := entity.NewWarehouseItem(warehouseUuid, materialUuid, "MAT001", 10.5)

// 增加库存
err := item.AddStock(100.0)

// 减少库存
err := item.ReduceStock(50.0)

// 预留库存
err := item.ReserveStock(30.0)

// 释放预留
err := item.ReleaseReservedStock(10.0)

// 获取可用库存
available := item.AvailableStock()  // 库存 - 预留库存

// 获取库存总价值
totalValue := item.TotalValue()  // 库存 * 估值单价
```

---

## 值对象说明

### WarehouseCode（仓库编码）

```go
code, err := valueobject.NewWarehouseCode("wh001")  // 自动转大写 "WH001"
code.String()  // "WH001"
code.Equals(otherCode)  // 比较
```

### WarehouseType（仓库类型）

```go
warehouseType := valueobject.NewWarehouseType("normal")
warehouseType.IsNormal()   // true
warehouseType.IsTransit()  // false
warehouseType.String()     // "normal"
```

### WarehouseStatus（仓库状态）

```go
status := valueobject.NewWarehouseStatus(1)
status.IsActive()    // true
status.IsDisabled()  // false
status.ToInt()       // 1
```

### Stock（库存数量）

```go
stock := valueobject.NewStock(100.555)  // 自动四舍五入为 100.56
stock.Value()      // 100.56
stock.IsZero()     // false
stock.IsPositive() // true

// 不可变操作，返回新对象
newStock := stock.Add(50.0)       // 150.56
newStock := stock.Subtract(30.0)  // 70.56
```

### ContactInfo（联系信息）

```go
contact := valueobject.NewContactInfo("张三", "13800138000", "北京市朝阳区")
contact.Name()     // "张三"
contact.Phone()    // "13800138000"
contact.Address()  // "北京市朝阳区"
```

---

## Repository 接口

### IWarehouseRepository

```go
type IWarehouseRepository interface {
    Save(ctx context.Context, warehouse *entity.Warehouse) error
    FindByUuid(ctx context.Context, uuid uint64) (*entity.Warehouse, error)
    FindByCode(ctx context.Context, code valueobject.WarehouseCode) (*entity.Warehouse, error)
    Remove(ctx context.Context, uuid uint64) error
    ExistsCode(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error)
    FindDefault(ctx context.Context) (*entity.Warehouse, error)
    FindAllNormal(ctx context.Context) ([]*entity.Warehouse, error)
    FindTransit(ctx context.Context) (*entity.Warehouse, error)
    SetAsDefault(ctx context.Context, uuid uint64) error
    FindWithPagination(ctx context.Context, spec WarehouseSpecification, pageNo, pageSize int) ([]*entity.Warehouse, int64, error)
}
```

### IWarehouseItemRepository

```go
type IWarehouseItemRepository interface {
    Save(ctx context.Context, item *entity.WarehouseItem) error
    FindByUuid(ctx context.Context, uuid uint64) (*entity.WarehouseItem, error)
    FindByWarehouseAndMaterial(ctx context.Context, warehouseUuid, materialUuid uint64) (*entity.WarehouseItem, error)
    FindByMaterialUuid(ctx context.Context, materialUuid uint64) ([]*entity.WarehouseItem, error)
    FindByWarehouseUuid(ctx context.Context, warehouseUuid uint64) ([]*entity.WarehouseItem, error)
    FindOrCreate(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, valuation float64) (*entity.WarehouseItem, error)
    Remove(ctx context.Context, uuid uint64) error
    GetMaterialStockInNormalWarehouses(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error)
    GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]WarehouseStockInfo, error)
    FindWithPagination(ctx context.Context, spec *WarehouseItemQuerySpec, pageNo, pageSize int) ([]*entity.WarehouseItem, int64, error)
    // 原子操作
    AddStock(ctx context.Context, uuid uint64, quantity float64) error
    ReduceStock(ctx context.Context, uuid uint64, quantity float64) error
    AddReservedStock(ctx context.Context, uuid uint64, quantity float64) error
    ReduceReservedStock(ctx context.Context, uuid uint64, quantity float64) error
}
```

---

## 领域服务

### IWarehouseDomainService

跨聚合的仓库业务逻辑：

```go
type IWarehouseDomainService interface {
    CreateWarehouse(ctx context.Context, req CreateWarehouseRequest) (*entity.Warehouse, error)
    ValidateForUpdate(ctx context.Context, warehouse *entity.Warehouse, newCode valueobject.WarehouseCode, newStatus valueobject.WarehouseStatus) error
    ValidateForDelete(ctx context.Context, warehouse *entity.Warehouse) error
    SetDefaultWarehouse(ctx context.Context, uuid uint64) error
}
```

### IWarehouseItemDomainService

库存物品业务逻辑：

```go
type IWarehouseItemDomainService interface {
    GetMaterialStock(ctx context.Context, materialUuid uint64) (float64, error)
    GetMaterialStockBatch(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error)
    GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]WarehouseStockInfo, error)
    TransferStock(ctx context.Context, fromWarehouse, toWarehouse, material uint64, quantity float64) error
    ReserveStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error
    ReleaseReservedStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error
    ConsumeStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error
    AddStock(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, quantity, valuation float64) error
}
```

---

## 领域事件

| 事件 | 说明 | 触发时机 |
|------|------|----------|
| `warehouse.created` | 仓库创建 | 新仓库创建后 |
| `warehouse.deleted` | 仓库删除 | 仓库删除后 |
| `warehouse.default_changed` | 默认仓库变更 | 默认仓库切换时 |

### 使用示例

#### 1. 创建订阅者

```go
package subscriber

import (
    "fmt"
    "ttpos-server-go/app/modules/inventory/domain/event"
)

// WarehouseEventSubscriber 仓库事件订阅者
type WarehouseEventSubscriber struct{}

// NewWarehouseEventSubscriber 创建订阅者
func NewWarehouseEventSubscriber() *WarehouseEventSubscriber {
    return &WarehouseEventSubscriber{}
}

// SubscribedEvents 返回订阅的事件列表
func (s *WarehouseEventSubscriber) SubscribedEvents() []string {
    return []string{
        "warehouse.created",
        "warehouse.deleted",
        "warehouse.default_changed",
    }
}

// Handle 处理事件
func (s *WarehouseEventSubscriber) Handle(e event.DomainEvent) error {
    switch evt := e.(type) {
    case *event.WarehouseCreatedEvent:
        fmt.Printf("仓库创建: UUID=%d, Code=%s\n", evt.AggregateID(), evt.Code)
        // 执行业务逻辑，如：同步到 ERP、发送通知等
    case *event.WarehouseDeletedEvent:
        fmt.Printf("仓库删除: UUID=%d\n", evt.AggregateID())
    case *event.DefaultWarehouseChangedEvent:
        fmt.Printf("默认仓库变更: Old=%d, New=%d\n", evt.PreviousDefaultUuid, evt.NewDefaultUuid)
    }
    return nil
}
```

#### 2. 注册和发布事件

```go
import "ttpos-server-go/app/modules/inventory/domain/event"

// 创建事件分发器（通常作为单例）
dispatcher := event.NewEventDispatcher()

// 注册订阅者
subscriber := NewWarehouseEventSubscriber()
dispatcher.Register(subscriber)

// 创建并发布事件
createdEvent := event.NewWarehouseCreatedEvent(uuid, code, name, warehouseType)
dispatcher.Publish(createdEvent)
```

#### 3. 在应用服务中使用

```go
type WarehouseAppService struct {
    dispatcher *event.EventDispatcher
    // ...
}

func (s *WarehouseAppService) CreateWarehouse(ctx context.Context, req CreateReq) error {
    // 1. 创建仓库
    warehouse, err := s.domainService.CreateWarehouse(ctx, ...)
    if err != nil {
        return err
    }
    
    // 2. 保存到数据库
    if err := s.warehouseRepo.Save(ctx, warehouse); err != nil {
        return err
    }
    
    // 3. 发布领域事件
    evt := event.NewWarehouseCreatedEvent(
        warehouse.Uuid(),
        warehouse.Code().String(),
        warehouse.Name().ZhName(),
        warehouse.Type().String(),
    )
    s.dispatcher.Publish(evt)
    
    return nil
}
```

---

## 测试

```bash
# 运行实体测试
go test -v ./app/modules/inventory/domain/entity/... -count=1

# 运行领域服务测试
go test -v ./app/modules/inventory/domain/service/... -count=1

# 运行所有库存领域测试
go test -v ./app/modules/inventory/domain/... -count=1
```

---

## 相关文档

- [DDD 架构总览](../../README.md)
- [应用服务层](../../application/inventory/)
- [基础设施层](../../infrastructure/persistence/inventory/)

---

**最后更新**: 2025-12-04
**维护者**: TTPOS Team

