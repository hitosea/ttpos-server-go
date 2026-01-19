# DDD 业务模块

## 概述

`app/modules/` 目录存放采用 DDD（领域驱动设计）架构的业务模块。每个子目录代表一个**限界上下文（Bounded Context）**。

---

## 目录结构

```
app/modules/
├── inventory/              # 库存模块
│   ├── domain/             # 领域层
│   │   ├── entity/         # 聚合根
│   │   ├── valueobject/    # 值对象
│   │   ├── repository/     # Repository 接口
│   │   ├── service/        # 领域服务
│   │   ├── specification/  # 规格模式
│   │   └── event/          # 领域事件
│   ├── application/        # 应用服务层
│   └── infrastructure/     # 基础设施层
│       ├── persistence/    # 持久化实现
│       └── integration/    # 外部集成
│
├── order/                  # 订单模块（计划中）
├── product/                # 商品模块（计划中）
└── shared/                 # 共享内核（计划中）
```

---

## 模块列表

| 模块 | 状态 | 说明 |
|------|------|------|
| **inventory** | ✅ 已完成 | 仓库管理、库存物品 |
| order | 📋 计划中 | 订单管理 |
| product | 📋 计划中 | 商品管理 |
| shared | 📋 计划中 | 跨模块共享 |

---

## 分层职责

### 领域层 (Domain)

- **实体 (Entity)**：聚合根，封装核心业务逻辑
- **值对象 (Value Object)**：不可变的业务概念
- **Repository 接口**：数据访问契约（只定义接口）
- **领域服务 (Domain Service)**：跨聚合的业务逻辑
- **领域事件 (Domain Event)**：领域内的事件通知

### 应用层 (Application)

- **应用服务**：编排领域服务，处理事务
- **DTO 转换**：领域对象与外部 DTO 转换
- **适配器**：兼容旧接口

### 基础设施层 (Infrastructure)

- **持久化**：Repository 实现
- **外部集成**：第三方服务调用

---

## 快速开始

### 使用库存模块

```go
import (
    "ttpos-server-go/app/modules/inventory/application"
    "ttpos-server-go/app/modules/inventory/domain/service"
    "ttpos-server-go/app/modules/inventory/infrastructure/persistence"
)

// 创建仓储
warehouseRepo := persistence.NewWarehouseRepository(dbm)
warehouseItemRepo := persistence.NewWarehouseItemRepository(dbm)

// 创建领域服务
warehouseDomainSrv := service.NewWarehouseDomainService(warehouseRepo)
itemDomainSrv := service.NewWarehouseItemDomainService(warehouseItemRepo)

// 创建应用服务
itemAppSrv := application.NewWarehouseItemAppService(warehouseItemRepo, itemDomainSrv, dbm)
```

---

## 新增模块

添加新模块时，按以下结构创建：

```bash
app/modules/{module_name}/
├── domain/
│   ├── entity/
│   ├── valueobject/
│   ├── repository/
│   ├── service/
│   └── event/
├── application/
├── infrastructure/
│   ├── persistence/
│   └── integration/
└── README.md           # 模块使用说明
```

---

## 与传统代码的关系

```
app/
├── api/              # API 层（传统）
├── dto/              # 数据传输对象（共享）
├── model/            # 数据库模型（共享）
├── service/          # 传统服务（逐步废弃）
├── repository/       # 传统仓储（逐步废弃）
│
└── modules/          # DDD 业务模块（新）
    └── inventory/
```

---

## 相关文档

- [库存模块详细文档](inventory/README.md)
- [Go Main 核心约束](../../.cursor/rules/go-main.mdc)

---

**最后更新**: 2025-12-04
**维护者**: TTPOS Team

