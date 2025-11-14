# WebSocket 服务数据库迁移文件

## 概述
本目录包含 ttpos-websocket 服务的数据库迁移文件，用于管理数据库表结构的版本控制。

## 迁移文件列表

### 20251114091944_init_websocket_schema
- **up.sql**: 创建 WebSocket 服务初始数据库表结构
- **down.sql**: 回滚 WebSocket 服务数据库表结构

#### 包含的表
1. **websocket_msg**: WebSocket 消息记录表
   - 存储所有 WebSocket 消息的详细记录
   - 支持消息状态追踪和离线消息处理
   - 包含完整的索引设计用于性能优化

## 使用方法

### 本地开发环境
```bash
# 在 ttpos-websocket 目录下执行
cd ttpos-bmp/app/ttpos-websocket

# 执行迁移
make db_up

# 回滚迁移
make db_down
```

### Docker 环境
```bash
# 在 ttpos-websocket 目录下执行
make db_up.docker
```

### 根目录全量迁移
```bash
# 在 ttpos-bmp 根目录执行
make migrate
```

## 表结构说明

### websocket_msg 表
用于存储 WebSocket 消息记录，支持：
- 消息唯一标识和去重
- 多公司数据隔离
- 消息状态管理
- 离线消息处理
- 性能优化索引

#### 字段说明
- `id`: 主键ID，自增长
- `uuid`: 消息唯一标识，用于去重和追踪
- `company_uuid`: 公司UUID，数据隔离
- `uid`: 用户/设备标识
- `msg`: 消息内容（JSON格式）
- `type`: 消息类型（heartbeat/order/notification/system/broadcast）
- `source_client`: 来源客户端（pos/tablet/kitchen/h5/mobile）
- `status`: 消息状态（0-待发送，1-发送中，2-发送成功，3-发送失败）
- `is_offline`: 是否离线消息（0-在线，1-离线）
- `create_time`: 创建时间戳
- `update_time`: 更新时间戳
- `delete_time`: 删除时间戳（软删除）

#### 索引设计
- `PRIMARY KEY (id)`: 主键索引
- `UNIQUE KEY idx_uuid (uuid)`: 消息唯一性索引
- `KEY idx_company_uid (company_uuid, uid)`: 公司用户复合索引
- `KEY idx_status (status)`: 状态查询索引
- `KEY idx_create_time (create_time)`: 时间查询索引
- `KEY idx_type (type)`: 消息类型索引
- `KEY idx_source_client (source_client)`: 客户端类型索引

## 注意事项

1. **数据库连接**: 确保数据库连接配置正确
2. **权限要求**: 需要数据库创建表和索引的权限
3. **字符集**: 使用 utf8mb4 字符集支持完整的 Unicode
4. **存储引擎**: 使用 InnoDB 引擎支持事务和外键
5. **备份建议**: 在生产环境执行迁移前请先备份数据库

## 相关文档
- [WebSocket 服务文档](../../docs/app/ttpos-websocket/)
- [数据库迁移规范](../../../docs/DATABASE_MIGRATION_RULES.md)
- [GoFrame 迁移工具](https://goframe.org/docs/cli)
