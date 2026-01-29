# LINE MAN 订单更新 Webhook 技术设计文档

⚠️ **提醒**: 本设计文档在需求审核通过前提前准备，待 requirements.md 审核通过后正式实施。

## 📋 概述

实现 LINE MAN 订单更新 Webhook 接口，使用 GoFrame 2.x 框架。

**核心流程**: 
```
LINE MAN → Controller → Service → Logic → DAO → MySQL
                                        ↓
                                   RocketMQ → Main 模块
```

## 🏗️ 架构设计

### 分层架构

**Controller 层** (`lineman_v1_order_update.go`)
- 接收请求，调用 Service，返回响应

**Service 层** (`lineman_order.go`)
- 接口定义: `HandleOrderUpdate(ctx, req) error`

**Logic 层** (`lineman_order.go`)
- 查询订单
- 幂等性检查（`order_updated_time`）
- 更新订单（事务）
- 发送 RocketMQ 事件

**DAO 层** (自动生成)
- `dao.Order` / `dao.OrderItem`

## 🗄️ 数据库设计

### 新增字段

```sql
ALTER TABLE `order` 
ADD COLUMN `order_updated_time` TIMESTAMP NULL 
COMMENT '订单更新时间（LINE MAN）';
```

## 🔄 代码复用

1. LINE MAN 认证中间件
2. Grab OrderEvent 结构体
3. RocketMQ: `queue.PushWithContext()`
4. 参考: `lineman_order.go` 的 `HandlePlaceOrder()`

## 核心实现

详见 `tasks.md` 任务分解。

---

**创建日期**: 2026-01-12  
**作者**: rikugun
