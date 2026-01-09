# Grab 菜单快照数据查询 gRPC 服务 设计文档

> 本文档定义 Grab 菜单快照数据查询 gRPC 服务 的技术设计和实现方案。

## 📋 概述

本设计旨在 `ttpos-takeout` (BMP 模块) 中实现一个 gRPC 接口 `GetMenuSnapshot`，用于查询 Grab 等渠道的菜单快照原始数据。涉及数据库表结构的调整（增加字段和重命名字段）、Proto 定义更新、以及相应的业务逻辑实现。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **微服务架构**: 基于 GoFrame 2.x 框架开发。
- **分层设计**: 严格遵循 Controller (RPC) -> Logic -> DAO 分层。
- **禁止修改**: 严禁手动修改 `internal/dao`, `internal/model/entity`, `internal/model/do` 下的自动生成代码。
- **服务注册**: gRPC 服务需注册到 Nacos。

### API 设计规范 (api.mdc)

- **gRPC 定义**: 使用 proto3 语法，遵循 Google API Design Guide。
- **错误处理**: 使用 gRPC 标准错误码 (NotFound, InvalidArgument 等)。

### 数据库规范 (database.mdc)

- **字段规范**: 时间字段使用 int (Unix Timestamp)，必须包含 `created_at`, `updated_at`, `deleted_at` (软删除)。
- **迁移管理**: 使用 migrate 工具管理数据库变更。

---

## 🔄 代码复用分析

### 可复用的现有组件

- **现有表 `channel_menu_snapshot`**: 将基于现有表结构进行迁移修改，而非新建。
- **GoFrame 工具链**: 复用 `gf gen dao` 生成数据访问层代码。
- **BMP 基础库**: 复用 `ttpos-bmp/utility` 中的工具函数。

---

## 🏗️ 架构设计

### 架构图

```mermaid
graph TD
    Client[gRPC Client (Debug/Ops Tool)] --> RPC[RPC Controller (GetMenuSnapshot)]
    RPC --> Logic[Logic Layer (MenuSnapshot)]
    Logic --> DAO[DAO Layer (ChannelMenuSnapshot)]
    DAO --> DB[(MySQL)]
```

### 模块划分

#### Go BMP 模块 (ttpos-takeout)

- **RPC Controller**: `ttpos-takeout/internal/controller/grab/` (或新建 `snapshot` 相关 controller) - 处理 gRPC 请求。
- **Logic 层**: `ttpos-takeout/internal/logic/grab/` - 实现快照查询逻辑。
- **DAO 层**: `ttpos-takeout/internal/dao/` - 自动生成的数据库访问代码。
- **Proto**: `ttpos-takeout/manifest/protobuf/takeout/` - gRPC 接口定义。

---

## 🗄️ 数据库设计

### 数据表设计

#### 表: `channel_menu_snapshot` (修改)

基于现有表结构进行修改。

**变更操作**:

1.  重命名 `create_time` -> `created_at`
2.  重命名 `update_time` -> `updated_at`
3.  新增 `deleted_at` (int, DEFAULT NULL) - 软删除支持
4.  新增 `sync_state` (varchar(32), DEFAULT 'QUEUEING')
5.  新增 `ttpos_menu_data` (longtext)
6.  新增 `ttpos_updated_at` (int, DEFAULT 0)

**最终表结构预览**:

```sql
CREATE TABLE `channel_menu_snapshot` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `shop_uuid` bigint unsigned NOT NULL COMMENT '商户UUID',
    `provider_name` varchar(32) NOT NULL COMMENT '渠道名称',
    `menu_data` longtext COMMENT '渠道原始菜单数据(JSON)',
    `request_id` varchar(64) NOT NULL COMMENT '请求ID',
    `sync_state` varchar(32) NOT NULL DEFAULT 'QUEUEING' COMMENT '同步状态',
    `ttpos_menu_data` longtext COMMENT 'TTPOS侧原始菜单数据',
    `ttpos_updated_at` int NOT NULL DEFAULT 0 COMMENT 'TTPOS侧数据更新时间',
    `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updated_at` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `deleted_at` int DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_request_id` (`request_id`),
    KEY `idx_shop_provider` (`shop_uuid`, `provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='渠道菜单快照表';
```

### 数据库迁移

**迁移脚本**:

需创建一个新的 `.up.sql` 和 `.down.sql` 文件。

```sql
-- up.sql
ALTER TABLE `channel_menu_snapshot` CHANGE `create_time` `created_at` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间';
ALTER TABLE `channel_menu_snapshot` CHANGE `update_time` `updated_at` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间';
ALTER TABLE `channel_menu_snapshot` ADD COLUMN `deleted_at` int(11) DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `channel_menu_snapshot` ADD COLUMN `sync_state` varchar(32) NOT NULL DEFAULT 'QUEUEING' COMMENT '同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED' AFTER `request_id`;
ALTER TABLE `channel_menu_snapshot` ADD COLUMN `ttpos_menu_data` longtext COMMENT 'TTPOS 侧菜单原始数据 (JSON)' AFTER `sync_state`;
ALTER TABLE `channel_menu_snapshot` ADD COLUMN `ttpos_updated_at` int(11) NOT NULL DEFAULT 0 COMMENT 'TTPOS 侧菜单数据更新时间' AFTER `ttpos_menu_data`;
```

---

## 🔌 API 设计

### gRPC API

#### Protobuf 定义

文件: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto`

```protobuf
service Takeout {
  // ... existing rpcs
  rpc GetMenuSnapshot (GetMenuSnapshotReq) returns (GetMenuSnapshotRes);
}

message GetMenuSnapshotReq {
  string provider_name = 1; // 渠道名称: grab
  string shop_uuid = 2;     // 店铺 UUID
  string request_id = 3;    // 请求 ID
}

message GetMenuSnapshotRes {
  string content = 1;          // Provider 侧原始菜单 JSON
  int64 updated_at = 2;        // 快照更新时间
  string sync_state = 3;       // 同步状态
}
```

---

## 🧩 组件和接口

### Logic 层

#### Logic 实现

文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go` (或新建 `snapshot` logic)

需实现 `GetMenuSnapshot` 方法：

1.  根据 `request_id` 查询 `channel_menu_snapshot` 表。
2.  校验 `provider_name` 和 `shop_uuid` 是否匹配。
3.  如果未找到，返回 `code.CodeNotFound`。
4.  如果找到，组装 `GetMenuSnapshotRes` 并返回。

---

## 🔒 安全设计

- **内部调用**: 该接口主要供内部排查使用，应限制调用来源或仅在内网环境暴露（通过网关配置）。
- **参数校验**: 严格校验 `request_id`, `shop_uuid` 格式。

---

## 🧪 测试策略

### 单元测试

- **Logic 层**: Mock DAO 层，测试 `GetMenuSnapshot` 的各种场景（找到、未找到、参数错误）。

### API 测试

- **gRPC 测试**: 使用 `grpcurl` 或编写测试代码调用 gRPC 接口，验证端到端功能。

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件 (SQL)
- [ ] 执行数据库迁移
- [ ] 重新生成 GoFrame DAO/Entity/DO 代码

### Phase 2: 接口定义与生成

- [ ] 修改 `.proto` 文件，添加 `GetMenuSnapshot` 定义
- [ ] 执行 `make dao` (包含 proto 编译) 生成 Go 代码

### Phase 3: 业务逻辑实现

- [ ] 在 Logic 层实现 `GetMenuSnapshot` 方法
- [ ] 在 Controller 层实现 gRPC 方法并调用 Logic

### Phase 4: 测试

- [ ] 编写 Logic 层单元测试
- [ ] 进行本地 gRPC 接口测试

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0
**创建日期**: 2025-12-11
**作者**: rikugun
**审核者**: {审核者}
