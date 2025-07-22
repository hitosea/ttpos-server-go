# 数据库迁移快速开始指南

## 1. 安装 golang-migrate

### macOS
```bash
brew install golang-migrate
```

### Linux
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
```

### Windows
下载 [golang-migrate releases](https://github.com/golang-migrate/migrate/releases) 并添加到 PATH

## 2. 配置数据库连接

编辑 `.env` 文件（如果没有则复制 `.env.example`）：
```bash
cp .env.example .env
```

修改数据库配置：
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=ttpos_bmp
```

## 3. 创建第一个迁移

### 使用 Makefile（推荐）
```bash
make migrate-create
# 输入迁移描述，例如：create_users_table
```

### 手动创建
```bash
migrate create -ext sql -dir app/ttpos-manager/manifest/sql -seq create_users_table
```

这会创建两个文件：
- `20250122143000_create_users_table.up.sql` （升级脚本）
- `20250122143000_create_users_table.down.sql` （回滚脚本）

## 4. 编写迁移脚本

编辑 `up.sql` 文件：
```sql
-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'UUID',
    `username` varchar(50) NOT NULL COMMENT '用户名',
    `email` varchar(100) NOT NULL COMMENT '邮箱',
    `password` varchar(255) NOT NULL COMMENT '密码',
    `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

编辑 `down.sql` 文件：
```sql
-- 删除用户表
DROP TABLE IF EXISTS `users`;
```

## 5. 执行迁移

### 执行升级
```bash
make migrate-up
```

### 查看当前版本
```bash
make migrate-version
```

### 回滚一个版本
```bash
make migrate-down
```

## 6. 常用命令

| 命令 | 说明 |
|------|------|
| `make help` | 显示所有可用命令 |
| `make migrate-create` | 创建新的迁移文件 |
| `make migrate-up` | 执行升级迁移 |
| `make migrate-down` | 回滚一个版本 |
| `make migrate-version` | 查看当前版本 |
| `make migrate-status` | 查看迁移状态 |
| `make migrate-fresh` | 重置数据库并重新执行所有迁移 |

## 7. 为不同应用执行迁移

```bash
# 管理端迁移
make migrate-manager

# ERP系统迁移
make migrate-erp

# 商城系统迁移
make migrate-shop
```

## 8. 最佳实践

### 迁移脚本编写
1. **一个迁移只做一件事**：每个迁移脚本应该只完成一个特定的数据库变更
2. **测试回滚**：确保每个迁移都能正确回滚
3. **使用事务**：在可能的情况下使用事务包装迁移
4. **避免数据丢失**：在删除数据前先备份

### 命名规范
- 文件名：`{timestamp}_{description}.{up|down}.sql`
- 表名：使用下划线命名法，如 `user_profiles`
- 字段名：使用下划线命名法，如 `user_name`
- 索引名：`idx_字段名` 或 `uk_字段名`

### 字段规范
- **时间字段**：使用 `int` 类型，以 `_time` 结尾
- **金额字段**：使用 `decimal(14,2)` 类型
- **必填字段**：`uuid`, `create_time`, `update_time`, `delete_time`

## 9. 故障排除

### 迁移失败
```bash
# 查看详细错误信息
migrate -path app/ttpos-manager/manifest/sql -database "$(DB_URL)" up -verbose

# 强制设置版本
make migrate-force
```

### 版本不一致
```bash
# 查看迁移状态
make migrate-status

# 强制设置版本
make migrate-force
```

### 语法错误
```bash
# 验证迁移脚本语法
make migrate-validate
```

## 10. 示例项目结构

```
ttpos-bmp/
├── app/
│   ├── ttpos-manager/
│   │   └── manifest/
│   │       └── sql/
│   │           ├── 20250122143000_create_users_table.up.sql
│   │           ├── 20250122143000_create_users_table.down.sql
│   │           ├── 20250122144500_add_user_email_index.up.sql
│   │           └── 20250122144500_add_user_email_index.down.sql
│   ├── ttpos-erp/
│   │   └── manifest/
│   │       └── sql/
│   └── ttpos-shop/
│       └── manifest/
│           └── sql/
├── Makefile
├── DATABASE_MIGRATION_RULES.md
├── MIGRATION_QUICK_START.md
└── migrate_template.sql
```

## 11. 下一步

1. 阅读 [DATABASE_MIGRATION_RULES.md](./DATABASE_MIGRATION_RULES.md) 了解详细规范
2. 查看 [migrate_template.sql](./migrate_template.sql) 获取更多模板
3. 开始创建你的第一个迁移！

## 需要帮助？

- 查看 [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- 阅读项目中的迁移规则文档
- 联系项目维护者 