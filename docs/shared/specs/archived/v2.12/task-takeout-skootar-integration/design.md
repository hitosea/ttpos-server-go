# 整合 Skootar 订单逻辑到现有订单模型 设计文档

> 本文档定义 整合 Skootar 订单逻辑到现有订单模型 的技术设计和实现方案。

## 📋 概述

本方案旨在重构 Skootar 订单逻辑，将原 `takeout_job` 表中混合的"通用订单信息"和"Skootar特有信息"进行拆分。通用信息合并到 `takeout_order` 主表，特有信息（如 `skootar_id`、`skootar_rating`）保留在扩展表 `takeout_order_skootar` 中。通过 `order_uuid` 进行关联，实现数据模型的统一和扩展性。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **目录结构**: 遵循 `ttpos-bmp` 模块结构，Logic 层处理业务，DAO/Model 由框架生成。
- **数据库规范**: 使用 GoFrame ORM 和 Migrations 进行数据库变更。
- **不可变性**: `dao`, `model/do`, `model/entity` 目录由工具生成，手动修改仅限于 `logic` 和 `dto`。

### 数据库规范 (database.mdc)

- **表设计**:
  - `takeout_order_skootar` 包含标准字段 `id`, `uuid`, `create_time`, `update_time`, `delete_time`。
  - 使用 `order_uuid` 作为逻辑外键关联主表。
  - 字段命名使用 snake_case。

---

## 🔄 代码复用分析

### 可复用的现有组件

- **`internal/logic/takeout`**: 现有的通用外卖服务接口定义。
- **`takeout_order` 表**: Grab 集成引入的通用订单表，直接复用。

### 集成点

- **`CreateOrder`**: 逻辑层需同时写入 `takeout_order` 和 `takeout_order_skootar`。
- **`GetDriverInfo`**: Controller 层需聚合主表和扩展表数据。

---

## 🏗️ 架构设计

### 模块划分

#### Go BMP 模块 (`ttpos-takeout`)

- **Logic 层**: `internal/logic/skootar/` - 负责拆解 CreateOrder 请求，分别调用不同 DAO 写入数据；负责聚合查询。
- **DAO 层**:
  - `internal/dao/order.go` (复用) - 操作 `takeout_order`
  - `internal/dao/order_skootar.go` (新) - 操作 `takeout_order_skootar`
- **DTO 层**: `internal/model/dto/skootar/` - 定义转换对象。

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: takeout_order_skootar (Skootar 订单扩展表)

```sql
CREATE TABLE IF NOT EXISTS `takeout_order_skootar` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` varchar(100) NOT NULL COMMENT '唯一标识',
    `order_uuid` varchar(100) NOT NULL COMMENT '关联主订单UUID',
    `skootar_id` varchar(100) DEFAULT NULL COMMENT '骑手ID',
    `skootar_name` varchar(100) DEFAULT NULL COMMENT '骑手名称',
    `skootar_phone` varchar(100) DEFAULT NULL COMMENT '骑手电话',
    `skootar_rating` decimal(10,2) DEFAULT NULL COMMENT '骑手评分',
    `skootar_image_url` text DEFAULT NULL COMMENT '骑手头像',
    `create_time` datetime DEFAULT NULL COMMENT '创建时间',
    `update_time` datetime DEFAULT NULL COMMENT '更新时间',
    `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Skootar订单扩展表';
```

### 数据库迁移

**迁移脚本**: `manifest/sql/20251205_migrate_skootar_data.up.sql`

逻辑：
1. 创建 `takeout_order_skootar` 表。
2. 从 `takeout_job` (旧表) 迁移数据：
   - 通用字段 -> `takeout_order` (注意：需处理主键冲突和数据映射)
   - 特有字段 -> `takeout_order_skootar`
3. (可选) 重命名/备份旧表 `takeout_job`。

---

## 📊 数据模型

### DTO 定义

#### CreateOrder 内部传输

```go
type SkootarOrderData struct {
    // 通用部分
    BaseOrder *entity.Order
    // 特有部分
    ExtInfo   *entity.OrderSkootar
}
```

---

## 🧩 组件和接口

### Logic 层

#### Skootar Service (`internal/logic/skootar/create_order.go`)

```go
func (s *sSkootar) CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error) {
    // 1. 调用 Skootar API 创建订单
    // ...

    // 2. 开启事务
    err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        // 3. 写入主表 takeout_order
        // ...

        // 4. 写入扩展表 takeout_order_skootar
        // ...
        return nil
    })
    return
}
```

#### Controller 层适配 (`internal/controller/rpc/takeout/takeout.go`)

```go
func (c *Controller) GetDriverInfo(...) {
    // 1. 查询主表获取 Provider
    // 2. 如果是 Skootar，联查扩展表
    // 3. 组装返回
}
```

---

## 🧪 测试策略

### 单元测试

- **DAO 测试**: 验证 `takeout_order_skootar` 的 CRUD。
- **Logic 测试**: 验证事务回滚（模拟主表成功、扩展表失败场景）。

### 集成测试

- **数据迁移测试**: 准备一份包含旧 `takeout_job` 数据的数据库快照，运行迁移脚本，验证数据是否正确进入新表。

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建 SQL 迁移脚本：创建新表 + 数据迁移逻辑
- [ ] 执行 Migration
- [ ] 生成 GoFrame DAO/Entity

### Phase 2: 核心实现

- [ ] 修改 `CreateOrder` Logic：写入双表
- [ ] 修改 `GetDriverInfo` Logic：聚合查询
- [ ] 修改 `JobStatusChange` Logic：更新状态到主表，更新司机信息到扩展表

### Phase 3: 测试

- [ ] 验证 Skootar 下单流程
- [ ] 验证 API 响应兼容性
- [ ] 验证历史数据查询

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: User  
**审核者**: TBD

