# Object Storage 模块

## 概述

Object Storage 模块是 TTPOS 系统中的对象存储层模块，采用 DDD（领域驱动设计）架构设计。该模块负责统一管理对象的获取、缓存和生命周期，通过泛型实现类型安全，基于三级缓存（L1 本地缓存 + L2 Redis + Singleflight）提供高性能的对象存储服务。

## 架构设计

### DDD 分层结构

```
app/modules/objectstorage/
├── domain/                          # 领域层（核心业务逻辑）
│   ├── entity/                     # 领域实体
│   │   └── association.go          # 关联配置实体
│   ├── repository/                 # 仓储接口
│   │   └── cache_layer.go          # 缓存层接口（泛型）
│   └── service/                    # 领域服务
│       └── object_storage.go       # 对象存储领域服务接口和配置
├── infrastructure/                 # 基础设施层
│   ├── adapter/                    # 适配器
│   │   └── cache_group_adapter.go  # 基于 ICacheGroup 的缓存适配器
│   └── persistence/                # 持久化实现
│       ├── object_storage_impl.go  # 对象存储实现
│       └── utils.go                # 工具函数（BuildKey、deduplicate等）
└── module.go                       # 模块入口
```

## 核心功能

### 1. 三级缓存架构

基于 `pkg/cache.ICacheGroup` 实现的三级缓存：

- **L1 本地缓存**：进程内内存缓存，最快访问速度
- **L2 Redis 缓存**：分布式缓存，支持多实例共享
- **Singleflight**：请求合并，防止缓存击穿

### 2. 类型安全的泛型实现

使用 Go 泛型实现类型安全的缓存层，避免类型转换和运行时错误：

```go
// 为不同对象类型创建独立的缓存适配器
cacheAdapter := adapter.NewCacheGroupAdapter[model.SaleBill](groupConfig, cache.Global, ttl)
```

### 3. 统一对象获取接口

提供类型安全的对象获取方法，自动处理缓存查询和回填：

- `Get()`: 单个对象获取
- `BatchGet()`: 批量对象获取（自动去重）
- `Update()`: 更新缓存
- `Invalidate()`: 使缓存失效

### 4. 多租户支持

Key 格式：`ttpos4:{company_uuid}:{object_type}:{object_uuid}`

- 自动从 context 提取 company UUID
- 支持按公司粒度批量失效缓存
- 支持按公司+对象类型粒度批量失效缓存

### 5. 自动关联注入

类似 GORM Preload 的配置映射方式，自动注入关联对象：

- 支持嵌套路径（如 `SaleOrders.SaleOrderProducts.ProductPackage`）
- 支持批量查询优化
- 支持单个对象和切片类型

## 核心接口

### CacheLayer 接口

```go
type CacheLayer[T any] interface {
    GET(key string, query func() (T, error)) (T, error)
    SET(key string, value T, ttl time.Duration) error
    DEL(keys ...string) error
    BATCH_GET(keys []string, query func([]string) (map[string]T, error)) (map[string]T, error)
    SCAN(ctx context.Context, pattern string) ([]string, error)
}
```

### IObjectStorage 接口

```go
type IObjectStorage[T any] interface {
    Get(ctx context.Context, key string, query func() (T, error)) (T, error)
    BatchGet(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) (map[string]T, error)
    Invalidate(ctx context.Context, key string) error
    Update(ctx context.Context, key string, value T) error
    Warmup(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) error
    InvalidateByCompany(ctx context.Context, companyUuid uint64) error
    InvalidateByCompanyAndType(ctx context.Context, companyUuid uint64, objectType string) error
    UpdateByCompany(ctx context.Context, companyUuid uint64, objectType string, values map[string]T) error
    PreloadWithConfig(ctx context.Context, obj interface{}, associations []entity.Association) error
}
```

### Association 配置结构

```go
type Association struct {
    Path          string  // 关联路径，支持嵌套，如 "SaleBillSetting"、"SaleOrders.SaleOrderProducts.ProductPackage"
    ObjectType    string  // 对象类型，用于构建缓存 key
    GetUUID       func(obj interface{}) uint64  // 从对象中提取 UUID 的函数
    QueryFunc     func(ctx context.Context, uuid uint64) (interface{}, error)  // 单个对象查询函数
    BatchQueryFunc func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error)  // 批量查询函数（可选）
}
```

## 使用示例

### 1. 创建缓存适配器

```go
import (
    "ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
    "ttpos-server-go/pkg/cache"
    "time"
)

// 配置缓存组
groupConfig := cache.GroupConfig{
    Name:             "object-storage",
    EnableLocalCache: true,   // 启用 L1 本地缓存
    EnableRedisCache: true,   // 启用 L2 Redis 缓存
    NegativeTTL:      30 * time.Second,  // 负缓存 30 秒
}

// 创建缓存适配器（泛型版本）
cacheAdapter := adapter.NewCacheGroupAdapter[model.SaleBill](
    groupConfig,
    cache.Global,  // 底层缓存实例（用于 DEL 操作）
    5 * time.Minute,  // 默认 TTL
)
```

### 2. 创建对象存储实例

```go
import (
    "ttpos-server-go/app/modules/objectstorage/domain/service"
    "ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
)

// 创建对象存储配置
config := &service.Config[model.SaleBill]{
    TTL:          5 * time.Minute,
    DisableCache: false,
    CacheLayer:   cacheAdapter,
}

// 创建对象存储实例
objectStorage := persistence.NewObjectStorage[model.SaleBill](config)
```

### 3. 使用对象存储获取对象

```go
import "ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"

// 构建缓存 key（自动从 context 提取 company UUID）
key := persistence.BuildKey(ctx, "SaleBill", saleBillUuid)

// 获取对象（自动处理缓存查询和回填）
saleBill, err := objectStorage.Get(ctx, key, func() (model.SaleBill, error) {
    // 缓存未命中时的查询逻辑
    return repo.GetSaleBill(CommonRepo.WhereByUuid(saleBillUuid))
})
```

### 4. 使用 PreloadWithConfig 注入关联对象

```go
import "ttpos-server-go/app/modules/objectstorage/domain/entity"

// 定义关联配置
associations := []entity.Association{
    {
        Path:       "SaleBillSetting",
        ObjectType: "SaleBillSetting",
        GetUUID: func(obj interface{}) uint64 {
            saleBill := obj.(*model.SaleBill)
            return saleBill.SaleBillSettingUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetSaleBillSetting(CommonRepo.WhereByUuid(uuid))
        },
        BatchQueryFunc: func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error) {
            // 批量查询优化
            settings, err := repo.GetSaleBillSettingsByUuids(uuids)
            if err != nil {
                return nil, err
            }
            result := make(map[uint64]interface{})
            for _, setting := range settings {
                result[setting.Uuid] = setting
            }
            return result, nil
        },
    },
    {
        Path:       "SaleOrders.SaleOrderProducts.ProductPackage",
        ObjectType: "ProductPackage",
        GetUUID: func(obj interface{}) uint64 {
            product := obj.(*model.SaleOrderProduct)
            return product.ProductPackageUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetProductPackage(CommonRepo.WhereByUuid(uuid))
        },
    },
}

// 注入关联对象
saleBill := &model.SaleBill{}
err := objectStorage.PreloadWithConfig(ctx, saleBill, associations)
```

### 5. 批量操作示例

```go
// 批量获取对象
keys := []string{
    persistence.BuildKey(ctx, "SaleBill", uuid1),
    persistence.BuildKey(ctx, "SaleBill", uuid2),
    persistence.BuildKey(ctx, "SaleBill", uuid3),
}

saleBills, err := objectStorage.BatchGet(ctx, keys, func(uuids []string) (map[string]model.SaleBill, error) {
    // 批量查询逻辑
    result := make(map[string]model.SaleBill)
    for _, key := range uuids {
        uuid := extractUuidFromKey(key)
        bill, err := repo.GetSaleBill(CommonRepo.WhereByUuid(uuid))
        if err != nil {
            continue
        }
        result[key] = bill
    }
    return result, nil
})

// 批量失效缓存
err := objectStorage.InvalidateByCompany(ctx, companyUuid)
err := objectStorage.InvalidateByCompanyAndType(ctx, companyUuid, "SaleBill")

// 批量更新缓存
values := map[string]model.SaleBill{
    key1: saleBill1,
    key2: saleBill2,
}
err := objectStorage.UpdateByCompany(ctx, companyUuid, "SaleBill", values)
```

## 实现细节

### CacheGroupAdapter

`CacheGroupAdapter` 是基于 `pkg/cache.ICacheGroup` 的适配器实现：

- **GET**: 使用 `group.Do()` 执行任务，自动处理 L1/L2 缓存和 Singleflight
- **SET**: 使用 `group.Do()` 方式实现，通过闭包传入 value 值，自动触发缓存写入
- **DEL**: 使用底层缓存直接删除
- **BATCH_GET**: 为每个 key 单独调用 GET，利用 Singleflight 合并并发请求

### Key 格式

```
ttpos4:{company_uuid}:{object_type}:{object_uuid}
```

- `ttpos4`: 系统前缀
- `company_uuid`: 公司 UUID（从 context 自动提取）
- `object_type`: 对象类型（如 "SaleBill"、"SaleOrder"）
- `object_uuid`: 对象 UUID

### 缓存策略

- **默认 TTL**: 可通过 `Config.TTL` 配置
- **类型特定 TTL**: 可通过 `Config.SetTTL(objectType, ttl)` 为不同对象类型设置不同的 TTL
- **负缓存**: 查询失败时缓存空结果，防止缓存穿透（通过 `GroupConfig.NegativeTTL` 配置）

## 相关文档

- [需求文档](../../../docs/shared/specs/active/story-main-object-storage-layer/requirements.md)
- [设计文档](../../../docs/shared/specs/active/story-main-object-storage-layer/design.md)
- [提案文档](../../../docs/team/proposals/2025-12/object-storage-layer.md)

---

**最后更新**: 2025-01-16  
**维护者**: TTPOS Team
