# 模块关系说明

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 理解 main/admin/ttpos-bmp 模块的关系和通信方式

---

## 模块概览

TTPOS 系统由三大模块组成：

| 模块 | 技术栈 | 职责 | 通信方式 |
|------|--------|------|----------|
| **main** | Go + Gin | 核心业务API | HTTP REST, gRPC |
| **admin** | PHP + ThinkPHP | 管理后台API | HTTP REST |
| **ttpos-bmp** | Go + GoFrame | 微服务群 | gRPC |

---

## 模块关系图

```
┌─────────────┐       ┌─────────────┐
│    main     │       │    admin    │
│   (Go/Gin) │       │ (PHP/TP6)   │
└──────┬──────┘       └──────┬──────┘
       │                     │
       │ gRPC                │ HTTP REST
       │                     │
       └─────────┬───────────┘
                 │
        ┌────────▼────────┐
        │   ttpos-bmp     │
        │ (Go/GoFrame)    │
        │ 微服务群        │
        └─────────────────┘
```

---

## main 模块

**定位**: 核心业务引擎

**主要功能**:
- 收银端 API (`/api/v1/cashier/*`)
- 店铺端 API (`/api/v1/shop/*`)
- 会员管理
- 订单处理
- 支付对接

**通信**:
- **对外**: 提供 HTTP REST API
- **对内**: 调用 ttpos-bmp 的 gRPC 服务

---

## admin 模块

**定位**: 管理后台支撑

**主要功能**:
- 管理后台 API (`/api/admin/*`)
- 店铺后台 API (`/api/shop/*`)
- 系统配置
- 权限管理
- 数据报表

**通信**:
- **对外**: 提供 HTTP REST API
- **现状**: 独立运行，未来可能调用 main 或 ttpos-bmp

---

## ttpos-bmp 模块

**定位**: 业务中台微服务群

**子服务**:
- `ttpos-erp`: ERP 服务（进销存）
- `ttpos-message`: 消息服务（邮件、短信）
- `ttpos-shop`: 店铺服务
- `ttpos-takeout`: 外卖服务
- `ttpos-websocket`: WebSocket 服务

**通信**:
- **对外**: 提供 gRPC 接口
- **对内**: 服务间 gRPC 调用

---

## 通信方式

### 1. HTTP REST API

**客户端 → main/admin**

```
前端应用 ─── HTTP REST ──→ main/admin
```

- 使用 JSON 格式
- JWT 认证
- 统一响应格式

### 2. gRPC

**main → ttpos-bmp**

```
main ─── gRPC ──→ ttpos-bmp
```

- Protocol Buffers 序列化
- 高性能、类型安全
- 服务注册发现（Nacos）

---

## 典型调用链路

### 订单创建

```
1. 收银端 → main API (HTTP)
2. main → ttpos-erp (gRPC) - 查询库存
3. main → ttpos-message (gRPC) - 发送通知
4. main → 收银端 (HTTP) - 返回结果
```

### 库存查询

```
1. 店铺后台 → admin API (HTTP)
2. admin → 数据库查询
3. admin → 店铺后台 (HTTP) - 返回数据
```

---

## 数据隔离

### 数据库

每个模块使用独立的数据库连接：

```
main   → shop{company_id} (商户库)
admin  → shop{company_id} (商户库)
ttpos-bmp → shop{company_id} (商户库)
```

**多租户**:
- 每个商户一个独立数据库
- 完全数据隔离
- 便于迁移和备份

### 缓存

共享 Redis 集群：

```
main   → Redis (共享)
admin  → Redis (共享)
ttpos-bmp → Redis (共享)
```

---

## 部署关系

```
┌────────────────────────────────────┐
│         Nginx (负载均衡)           │
└────────┬───────────────────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼───┐
│ main │  │admin │
└───┬──┘  └──────┘
    │
    │ gRPC
    │
┌───▼──────────┐
│  ttpos-bmp   │
│  (微服务群)   │
└──────────────┘
```

---

## 未来演进

### 阶段1: 现状
- main 和 admin 独立运行
- main 调用 ttpos-bmp
- admin 独立查询数据库

### 阶段2: 整合（进行中）
- 逐步迁移 admin 功能到 main
- 统一 API 入口
- 简化架构

### 阶段3: 微服务化（规划中）
- main 拆分为多个服务
- 统一网关
- 服务治理

---

## 相关文档

- [系统架构总览](./overview.md) - 整体架构设计
- [Go Main 架构](./go-main-architecture.md) - main 模块详细设计
- [Go BMP 架构](./go-bmp-architecture.md) - ttpos-bmp 详细设计
- [PHP 架构](./php-architecture.md) - admin 模块详细设计

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

