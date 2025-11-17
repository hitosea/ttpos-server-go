# 数据库开发详细指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 数据库设计、迁移和操作的详细指南

---

## 多租户架构

### 独立数据库模式

每个商户一个独立数据库：

```
shop8267304538112000     # 商户1
shop8609817471094784     # 商户2
saas                     # 系统主库
```

**优点**:
- 数据完全隔离，安全性高
- 便于数据迁移和备份
- 便于扩展到不同服务器

---

## 命名规范

### 数据库命名

```sql
shop8267304538112000    -- ✅ 商户数据库（shop + 商户ID）
saas                    -- ✅ 系统数据库
```

### 表命名

```sql
ttpos_product              -- ✅ 使用 ttpos_ 前缀
ttpos_sale_order           -- ✅ 小写下划线
ttpos_product_category     -- ✅ 清晰的命名
```

### 字段命名

```sql
user_id                    -- ✅ 小写下划线
product_name               -- ✅ 清晰明确
create_time                -- ✅ 统一命名
```

### 索引命名

```sql
PRIMARY KEY (`id`)                      -- 主键
UNIQUE KEY `uk_uuid` (`uuid`)          -- 唯一索引
KEY `idx_company_id` (`company_id`)    -- 普通索引
KEY `idx_status_type` (`status`, `type`) -- 复合索引
```

---

## 字段规范

### 必需字段

每个表必须包含：

```sql
CREATE TABLE `ttpos_example` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例表';
```

### 时间字段 - 使用 int 类型

```sql
`create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间'
`update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间'
`delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间'
```

**原因**:
- int 类型占用空间小（4字节）
- 不受时区影响
- 便于计算和比较
- 默认值为 0 表示未设置

### 金额字段 - 使用 decimal

```sql
-- 高精度金额（8位小数）
`amount` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '金额'

-- 普通金额（2位小数）
`price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '价格'
`total` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '总计'
```

**为什么不用 float/double?**
- float/double 会有精度丢失问题
- 金额计算必须精确

### UUID 字段 - 使用 bigint

```sql
`uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识'
`company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '公司UUID'
```

### 状态字段 - 使用 tinyint

```sql
`status` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0-禁用 1-启用'
`type` tinyint NOT NULL DEFAULT 0 COMMENT '类型：1-类型A 2-类型B'
```

### 文本字段

```sql
-- 短文本（<255字符）
`name` varchar(100) NOT NULL DEFAULT '' COMMENT '名称'

-- 中等文本（<65535字符）
`description` text COMMENT '描述'

-- 长文本（>65535字符）
`content` longtext COMMENT '内容'
```

⚠️ **注意**: text 类型不能设置默认值

---

## 索引设计

### 索引类型

```sql
-- 主键索引
PRIMARY KEY (`id`)

-- 唯一索引
UNIQUE KEY `uk_uuid` (`uuid`)

-- 普通索引
KEY `idx_company_id` (`company_id`)

-- 复合索引
KEY `idx_company_status` (`company_id`, `status`)
```

### 索引设计原则

1. 为 WHERE、ORDER BY、GROUP BY 中的字段添加索引
2. 复合索引遵循最左前缀原则
3. 选择性高的字段放在复合索引前面
4. 避免过多索引（影响写入性能）
5. 定期分析和优化索引

### 索引使用示例

```sql
-- ✅ 使用索引
SELECT * FROM ttpos_order WHERE company_uuid = 123;

-- ✅ 使用复合索引
SELECT * FROM ttpos_order WHERE company_uuid = 123 AND status = 1;

-- ❌ 无法使用复合索引（没有前导列）
SELECT * FROM ttpos_order WHERE status = 1;
```

---

## 迁移文件

### PHP 迁移（admin/）

```php
<?php
use think\migration\Migrator;
use think\migration\db\Column;

class CreateProductTable extends Migrator
{
    public function change()
    {
        $table = $this->table('product', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_general_ci',
            'comment' => '商品表',
        ]);
        
        $table->addColumn('id', 'biginteger', [
            'signed' => false,
            'identity' => true,
            'comment' => '主键ID',
        ])
        ->addColumn('uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '唯一标识',
        ])
        ->addColumn('name', 'string', [
            'limit' => 100,
            'default' => '',
            'comment' => '商品名称',
        ])
        ->addColumn('price', 'decimal', [
            'precision' => 14,
            'scale' => 2,
            'default' => 0.00,
            'comment' => '价格',
        ])
        ->addColumn('create_time', 'integer', [
            'signed' => false,
            'default' => 0,
            'comment' => '创建时间',
        ])
        ->addIndex(['uuid'], ['unique' => true])
        ->addIndex(['name'])
        ->create();
    }
}
```

### GoFrame 迁移（ttpos-bmp/）

```sql
-- manifest/sql/xxx_up.sql
CREATE TABLE IF NOT EXISTS `product` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一标识',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '商品名称',
    `price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '价格',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';
```

---

## 查询优化

### 使用索引

```sql
-- ✅ 使用索引字段查询
SELECT * FROM ttpos_user WHERE uuid = 123;

-- ❌ 避免全表扫描
SELECT * FROM ttpos_user WHERE email LIKE '%@example.com';
```

### 避免 SELECT *

```sql
-- ✅ 只查询需要的字段
SELECT id, name, price FROM ttpos_product WHERE status = 1;

-- ❌ 查询所有字段（浪费）
SELECT * FROM ttpos_product WHERE status = 1;
```

### 分页查询

```sql
-- ✅ 使用 LIMIT 分页
SELECT * FROM ttpos_order 
WHERE company_uuid = 123 
ORDER BY create_time DESC 
LIMIT 20 OFFSET 0;
```

### 预加载避免 N+1

```go
// ✅ 预加载
db.Preload("Products").Find(&orders)

// ❌ N+1 查询
for _, order := range orders {
    db.Where("order_id = ?", order.Id).Find(&products)
}
```

---

## 事务处理

### Go (GORM)

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&profile).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

### PHP (ThinkPHP)

```php
Db::startTrans();
try {
    User::create($userData);
    Profile::create($profileData);
    Db::commit();
} catch (\Exception $e) {
    Db::rollback();
    throw $e;
}
```

---

## 软删除

### 实现软删除

```sql
-- 软删除：更新 delete_time
UPDATE ttpos_user SET delete_time = 1699999999 WHERE id = 1;

-- 查询时排除已删除
SELECT * FROM ttpos_user WHERE delete_time = 0;
```

### Go 实现

```go
// 软删除
db.Model(&User{}).Where("id = ?", userId).Update("delete_time", time.Now().Unix())

// 查询未删除
db.Where("delete_time = ?", 0).Find(&users)
```

### PHP 实现

```php
// 软删除
User::where('id', $userId)->update(['delete_time' => time()]);

// 模型自动处理软删除
$user = User::find($userId);
$user->delete();  // 自动设置 delete_time
```

---

## 相关文档

- [数据库设计](../architecture/database-design.md) - 数据库架构设计理念
- [Go Main 开发指南](./go-main-development.md) - Go 数据库操作
- [PHP 开发指南](./php-development.md) - PHP 数据库操作

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

