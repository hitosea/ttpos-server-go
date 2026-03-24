---
description: 数据库设计与迁移规范 — 编写迁移文件或修改数据模型时自动加载
globs:
  - "admin/database/migrations/**"
  - "main/app/model/**/*.go"
  - "ttpos-bmp/**/manifest/sql/**"
---

# 数据库规范

## 多租户架构

每个商户一个独立数据库（如 `shop8267304538112000`），业务表无需 `company_uuid` 字段。

## 表设计规范

- 表名使用 `ttpos_` 前缀
- **迁移文件中表名不要带 `ttpos_` 前缀**（框架自动添加）
- 每表必需字段：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- 时间字段用 `int`（Unix 时间戳）
- 金额字段用 `decimal`
- 布尔字段用 `int`（0/1），禁止使用 boolean

## 迁移文件 TARGET 常量

| 值 | 说明 |
|---|---|
| `const TARGET = 'all';` | 所有商户库 + saas 主库 |
| `const TARGET = 'main';` | 仅 saas 主库 |
| 不设置 | 仅所有商户数据库 |

## 多语言表

需要多语言的业务表必须包含：
- `name` VARCHAR(1000) — JSON 格式多语言数据
- `multi_language_name_uuid` BIGINT — 关联 `ttpos_multi_language_name` 表

## 迁移后同步

创建迁移文件后必须更新：
- `admin/database/seeds/shop_01.sql`
- `main/app/model/` 对应模型文件
