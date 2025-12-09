# 外卖渠道菜单数据存储设计文档

> 本文档定义外卖渠道菜单数据存储功能的技术设计和实现方案。

## 📋 概述

本功能旨在为 TTPOS 系统构建一个统一、隔离且通用的存储机制，用于保存 TTPOS 商家发送给不同外卖渠道（如 Grab, Lineman）的菜单数据快照。该功能将在 `ttpos-bmp` 的 `ttpos-takeout` 模块中实现，提供内部 Logic 接口供业务流程调用。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **GoFrame 框架**: 使用 GoFrame 2.x 框架和工具链。
- **分层架构**: 遵循 Controller -> Logic -> DAO 分层。
- **代码生成**: 使用 `gf gen dao` 生成 DAO/Entity/DO 代码，禁止手动修改。
- **项目结构**: 遵循 `ttpos-bmp` 的标准目录结构。

### 数据库规范 (database.mdc)

- **表命名**: 使用 `takeout_` 前缀（微服务内表名）。
- **字段规范**: 包含 `id`, `uuid`, `create_time`, `update_time`。
- **类型规范**: 时间使用 `int`，大文本使用 `longtext`。

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Takeout Logic**: `ttpos-bmp/app/ttpos-takeout/internal/logic/` - 现有的外卖逻辑层，将在其中新增 `channel_menu` 逻辑。
- **DB Component**: GoFrame 的 `gdb` 组件用于数据库操作。

### 集成点

- **Grab 集成**: 在 Grab 菜单同步流程中调用保存接口。
- **Lineman 集成**: 在 Lineman 菜单同步流程中调用保存接口。

---

## 🏗️ 架构设计

### 分层设计原则

**Go BMP 架构**:

```
Logic 层 (Internal Service)
  ↓ 依赖
DAO 层 (Data Access Object)
  ↓ 依赖
Database (MySQL)
```

### 模块划分

#### Go BMP 模块 (`ttpos-takeout`)

- **Logic 层**: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/` - 负责菜单快照的存储和读取逻辑。
- **DAO 层**: `ttpos-bmp/app/ttpos-takeout/internal/dao/` - 自动生成的数据库访问层。
- **Model 层**: `ttpos-bmp/app/ttpos-takeout/internal/model/` - 数据模型。
- **Service 接口**: `ttpos-bmp/app/ttpos-takeout/internal/service/` - 定义对外暴露的接口。

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: takeout_channel_menu_snapshot

```sql
CREATE TABLE IF NOT EXISTS `takeout_channel_menu_snapshot` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `shop_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '商户UUID',
    `provider_name` varchar(32) NOT NULL DEFAULT '' COMMENT '渠道名称 (grab, lineman)',
    `menu_data` longtext COMMENT '菜单数据快照 (JSON)',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_shop_provider` (`shop_uuid`, `provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外卖渠道菜单数据快照表';
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | bigint unsigned | 主键 ID | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | DEFAULT 0, UNIQUE |
| shop_uuid | bigint unsigned | 商户UUID | DEFAULT 0 |
| provider_name | varchar(32) | 渠道名称 | DEFAULT '' |
| menu_data | longtext | 菜单数据JSON | NULL |
| create_time | int | 创建时间 | DEFAULT 0 |
| update_time | int | 更新时间 | DEFAULT 0 |

**索引设计**:
- 主键索引: `PRIMARY KEY (id)`
- 唯一索引: `UNIQUE KEY uk_uuid (uuid)`
- 唯一索引: `UNIQUE KEY uk_shop_provider (shop_uuid, provider_name)` - 确保每个商户每个渠道只有一份快照

**迁移文件**: `ttpos-bmp/app/ttpos-takeout/manifest/sql/{YYYYMMDDHHMMSS}_create_takeout_channel_menu_snapshot_table.up.sql`

---

## 🧩 组件和接口

### Service 层

#### Service 接口

```go
// ttpos-bmp/app/ttpos-takeout/internal/service/channel_menu.go
type IChannelMenu interface {
    SaveChannelMenu(ctx context.Context, shopUUID uint64, providerName string, menuData string) error
    GetChannelMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error)
}
```

#### Logic 实现

```go
// ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go
type sChannelMenu struct {}

func (s *sChannelMenu) SaveChannelMenu(ctx context.Context, shopUUID uint64, providerName string, menuData string) error {
    // 1. 检查是否存在
    // 2. 如果存在，更新 menu_data 和 update_time
    // 3. 如果不存在，插入新记录
    // 使用 DAO 的 Save 或 InsertOnDuplicateKeyUpdate 功能
}

func (s *sChannelMenu) GetChannelMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error) {
    // 1. 查询 DAO
    // 2. 返回 menu_data
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 数据库写入失败
- **处理方式**: 返回 error，上层业务记录错误日志，但不阻断核心外卖业务（可选，视业务重要性而定）。
- **用户影响**: 运维无法查看到最新的菜单快照。

#### 场景 2: 查询不存在的数据
- **处理方式**: 返回特定的 NotFound error 或空字符串。

---

## 🔒 安全设计

### 数据安全
- **内部接口**: 该功能仅通过内部 Service 接口暴露，不直接对外提供 HTTP API。

---

## 🧪 测试策略

### 单元测试
- **Service 测试**: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu_test.go`
- **覆盖率**: > 80%

### 集成测试
- 模拟调用 `SaveChannelMenu` 保存 Grab 和 Lineman 格式的 JSON 数据。
- 验证 `GetChannelMenu` 返回的数据是否与保存的一致。
- 验证重复保存是否正确更新。

---

## 📚 实现清单

### Phase 1: 数据库和模型
- [ ] 创建 SQL 迁移文件 `up.sql` 和 `down.sql`
- [ ] 执行数据库迁移 (手动或通过工具)
- [ ] 使用 `gf gen dao` 生成 Go 代码

### Phase 2: 核心实现
- [ ] 定义 Service 接口 `internal/service/channel_menu.go`
- [ ] 实现 Logic `internal/logic/channel_menu/channel_menu.go`
- [ ] 注册 Service `internal/logic/logic.go`

### Phase 3: 测试
- [ ] 编写单元测试
- [ ] 验证 Grab/Lineman 数据兼容性

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志
- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-08.md`
