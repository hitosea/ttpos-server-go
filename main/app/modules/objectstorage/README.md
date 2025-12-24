# Object Storage 模块

## 概述

Object Storage 模块是 TTPOS 系统中的对象存储层模块，采用 DDD（领域驱动设计）架构设计。该模块负责统一管理对象的获取、缓存和生命周期，减少代码重复，提高可维护性。

## 架构设计

### DDD 分层结构

```
app/modules/objectstorage/
├── domain/                          # 领域层（核心业务逻辑）
│   ├── entity/                     # 领域实体
│   │   └── association.go          # 关联配置实体
│   ├── repository/                 # 仓储接口
│   │   └── cache_layer.go         # 缓存层接口
│   └── service/                    # 领域服务
│       └── object_storage.go       # 对象存储领域服务
├── application/                     # 应用层（编排与协调）
│   └── object_storage_app_service.go # 应用服务
├── infrastructure/                  # 基础设施层
│   ├── adapter/                    # 适配器
│   │   └── cache_adapter.go       # 缓存适配器
│   └── persistence/                # 持久化实现
│       └── object_storage_impl.go  # 对象存储实现
└── module.go                       # 模块入口
```

## 核心功能

1. **统一对象获取接口**：提供类型安全的对象获取方法，自动处理缓存查询和回填
2. **批量操作支持**：支持批量获取对象，自动去重和批量查询优化
3. **生命周期管理**：支持缓存失效、更新、预热等操作
4. **多租户支持**：Key 格式包含 company UUID，支持按公司粒度管理缓存
5. **自动关联注入**：类似 GORM Preload 的配置映射方式，自动注入关联对象

## 使用示例

```go
import (
    "ttpos-server-go/app/modules/objectstorage"
    "ttpos-server-go/pkg/cache"
    "ttpos-server-go/pkg/database"
)

// 1. 初始化模块
objectStorageAppService := objectstorage.NewObjectStorageAppService(
    cache.Global,
    dbm,
)

// 2. 获取 SaleBill（不使用 Preload）
saleBill, err := repo.GetSaleBill(CommonRepo.WhereByUuid(saleBillUuid))

// 3. 使用对象存储层自动注入关联对象
err = objectStorageAppService.PreloadSaleBillAssociations(ctx, &saleBill, db)
```

## 相关文档

- [需求文档](../../../docs/shared/specs/active/story-main-object-storage-layer/requirements.md)
- [设计文档](../../../docs/shared/specs/active/story-main-object-storage-layer/design.md)
- [提案文档](../../../docs/team/proposals/2025-12/object-storage-layer.md)

---

**最后更新**: 2025-12-24  
**维护者**: TTPOS Team

