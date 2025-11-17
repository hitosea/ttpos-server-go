# 数据库迁移工作流

> 本文档定义数据库迁移的完整流程（PHP Phinx + Go Model 同步）

---

## 📋 概述

### 适用场景

- 新增数据库表
- 修改表结构（添加/删除/修改字段）
- 创建索引
- 数据迁移

### 预计时间

- 简单迁移（新增表）: 0.5-1 小时
- 复杂迁移（多表关联）: 1-2 小时

### 核心原则

**所有数据库迁移文件统一在 PHP 中管理（admin/database/migrations/），同时同步更新 Go model**

---

## 完整流程

```
设计表结构 → 创建迁移文件 (PHP) → 编写迁移代码 →
更新 Go model → 更新 seeds 文件 → 执行迁移 → 验证数据 → 提交代码
```

---

## 执行流程

### Step 1: 设计表结构 (15-30 分钟)

#### 设计表字段

```markdown
## 订单表 (order)

| 字段名      | 类型        | 长度 | 默认值 | 注释     |
| ----------- | ----------- | ---- | ------ | -------- |
| id          | biginteger  | 主键 | 自增   | 主键 ID  |
| uuid        | biginteger  | -    | 0      | UUID     |
| order_no    | string      | 50   | ''     | 订单号   |
| amount      | decimal     | 20,8 | 0      | 订单金额 |
| status      | tinyinteger | -    | 0      | 订单状态 |
| create_time | integer     | -    | 0      | 创建时间 |
| update_time | integer     | -    | 0      | 更新时间 |
| delete_time | integer     | -    | 0      | 删除时间 |
```

#### 字段规范（必须遵守）

```yaml
时间字段:
  - 类型: integer (不用 datetime)
  - 命名: 以 _time 结尾
  - 默认值: 0
  - 示例: create_time, update_time, delete_time

金额字段:
  - 类型: decimal(20,8)
  - 命名: 以 amount/price/money 结尾
  - 示例: total_amount, unit_price

UUID字段:
  - 类型: biginteger (不用 bigint)
  - 默认值: 0
  - 必须字段

必须字段:
  - uuid (biginteger, 默认 0)
  - create_time (integer, 默认 0)
  - update_time (integer, 默认 0)
  - delete_time (integer, 默认 0)
```

参考: `.cursor/rules/php.mdc` - 迁移文件规范

---

### Step 2: 创建迁移文件 (5 分钟)

#### 生成迁移文件

```bash
cd path/ttpos-server-go/admin

# 创建迁移文件（文件名前缀会自动生成时间戳）
php think make:migration AddOrderTable
```

**文件命名规范**:

- 前缀: `date +%Y%m%d%H%M%S` 的结果
- 格式: `{timestamp}_add_{table_name}_table.php`
- 示例: `20250721370408_add_order_table.php`

#### 检查清单

- [ ] 迁移文件已创建
- [ ] 文件名符合规范
- [ ] 文件位置: `admin/database/migrations/`

---

### Step 3: 编写迁移代码 (15-30 分钟)

#### 迁移文件模板

```php
<?php

use Phinx\Migration\AbstractMigration;

class AddOrderTable extends AbstractMigration
{
    /**
     * 执行迁移
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('order')) {
            $table = $this->table('order', [
                'id' => false, // 不使用默认主键
                'primary_key' => ['id'],
                'comment' => '订单表', // 表注释（必须）
            ]);

            $table->addColumn('id', 'biginteger', ['identity' => true, 'comment' => '主键ID'])
                  ->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => 'UUID'])
                  ->addColumn('order_no', 'string', ['limit' => 50, 'default' => '', 'comment' => '订单号'])
                  ->addColumn('amount', 'decimal', ['precision' => 20, 'scale' => 8, 'default' => 0, 'comment' => '订单金额'])
                  ->addColumn('status', 'integer', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY, 'default' => 0, 'comment' => '订单状态'])
                  ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                  ->addIndex(['uuid'], ['unique' => true])
                  ->addIndex(['order_no'], ['unique' => true])
                  ->addIndex(['create_time'])
                  ->create();
        }
    }
}
```

#### 修改字段示例

```php
public function change()
{
    $table = $this->table('order');

    // 检查字段是否存在
    if (!$table->hasColumn('remark')) {
        $table->addColumn('remark', 'text', ['comment' => '备注', 'after' => 'status'])
              ->update();
    }
}
```

#### 删除字段示例

```php
public function change()
{
    $table = $this->table('order');

    // 检查字段是否存在
    if ($table->hasColumn('old_field')) {
        $table->removeColumn('old_field')
              ->update();
    }
}
```

#### 关键规范

- [ ] 表名不加前缀 "ttpos\_"
- [ ] 必须添加表注释
- [ ] 迁移前检查表/字段是否存在
- [ ] 所有 addColumn 不要换行
- [ ] text 类型不设置默认值
- [ ] 所有注释使用中文

---

### Step 4: 更新 Go Model (15-30 分钟)

#### 创建/更新 Model 文件

```go
// main/app/model/order.go
package model

import "time"

type Order struct {
    ID         uint64  `gorm:"primaryKey;column:id" json:"id"`
    UUID       uint64  `gorm:"column:uuid;index:idx_uuid,unique" json:"uuid"`
    OrderNo    string  `gorm:"column:order_no;size:50;index:idx_order_no,unique" json:"order_no"`
    Amount     float64 `gorm:"column:amount;type:decimal(20,8)" json:"amount"`
    Status     uint8   `gorm:"column:status" json:"status"`
    CreateTime int64   `gorm:"column:create_time;index:idx_create_time" json:"create_time"`
    UpdateTime int64   `gorm:"column:update_time" json:"update_time"`
    DeleteTime int64   `gorm:"column:delete_time" json:"delete_time"`
}

func (Order) TableName() string {
    return "order"
}
```

#### 类型映射

| PHP (Phinx)   | MySQL   | Go (GORM) |
| ------------- | ------- | --------- |
| biginteger    | BIGINT  | uint64    |
| integer       | INT     | int32     |
| tinyinteger   | TINYINT | uint8     |
| string        | VARCHAR | string    |
| text          | TEXT    | string    |
| decimal(20,8) | DECIMAL | float64   |

#### 检查清单

- [ ] Go Model 已创建/更新
- [ ] 字段类型映射正确
- [ ] 添加了 gorm 标签
- [ ] 添加了 json 标签
- [ ] TableName() 方法已实现

---

### Step 5: 更新 Seeds 文件 (10-20 分钟)

#### 更新 shop_01.sql

```bash
# 文件位置
admin/database/seeds/shop_01.sql
```

#### 添加建表语句

```sql
-- 订单表
DROP TABLE IF EXISTS `order`;
CREATE TABLE `order` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'UUID',
  `order_no` varchar(50) NOT NULL DEFAULT '' COMMENT '订单号',
  `amount` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '订单金额',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '订单状态',
  `create_time` int(11) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `delete_time` int(11) NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uuid` (`uuid`),
  UNIQUE KEY `idx_order_no` (`order_no`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
```

#### 添加测试数据（可选）

```sql
INSERT INTO `order` (`uuid`, `order_no`, `amount`, `status`, `create_time`)
VALUES
  (1001, 'ORD20250116001', 100.50000000, 1, 1737014400),
  (1002, 'ORD20250116002', 200.75000000, 1, 1737014400);
```

---

### Step 6: 执行迁移 (5 分钟)

#### 运行迁移命令

```bash
cd path/ttpos-server-go/admin

# 执行迁移
php think migrate:run

# 检查迁移状态
php think migrate:status
```

#### 验证结果

```bash
# 登录数据库
mysql -u root -p

# 检查表结构
SHOW CREATE TABLE `order`;

# 检查字段
DESC `order`;
```

#### 回滚（如需要）

```bash
# 回滚上一次迁移
php think migrate:rollback

# 回滚到指定版本
php think migrate:rollback -t {timestamp}
```

---

### Step 7: 验证数据 (5-10 分钟)

#### Go 代码验证

```go
// 测试 Model
func TestOrderModel(t *testing.T) {
    db := database.GetDB()

    order := &model.Order{
        UUID:       1001,
        OrderNo:    "TEST001",
        Amount:     100.50,
        Status:     1,
        CreateTime: time.Now().Unix(),
    }

    err := db.Create(order).Error
    assert.NoError(t, err)

    // 查询验证
    var result model.Order
    err = db.Where("uuid = ?", 1001).First(&result).Error
    assert.NoError(t, err)
    assert.Equal(t, "TEST001", result.OrderNo)
}
```

#### 检查清单

- [ ] 表已创建
- [ ] 字段类型正确
- [ ] 索引已创建
- [ ] Go Model 可以正常操作
- [ ] 测试数据可以插入

---

### Step 8: 提交代码 (5-10 分钟)

#### 提交迁移文件

```bash
git add admin/database/migrations/{timestamp}_add_order_table.php
git add admin/database/seeds/shop_01.sql
git add main/app/model/order.go

git commit -m "feat(database): 新增订单表

- 创建 order 表
- 添加必须字段 (uuid, create_time, update_time, delete_time)
- 更新 Go Model
- 更新 seeds 文件

Migration: {timestamp}_add_order_table"
```

---

## 检查清单

### 设计阶段

- [ ] 表结构已设计
- [ ] 字段规范已确认
- [ ] 索引已设计

### 迁移文件

- [ ] 迁移文件已创建
- [ ] 迁移前检查表/字段存在性
- [ ] 字段规范正确（时间 int，金额 decimal）
- [ ] 必须字段已添加 (uuid, \*\_time)
- [ ] 表注释已添加
- [ ] 所有注释使用中文

### Go Model

- [ ] Go Model 已创建/更新
- [ ] 字段类型映射正确
- [ ] gorm 标签正确
- [ ] TableName() 已实现

### Seeds

- [ ] shop_01.sql 已更新
- [ ] 建表语句正确
- [ ] 测试数据已添加（可选）

### 执行验证

- [ ] 迁移执行成功
- [ ] 表结构正确
- [ ] Go Model 可以正常操作
- [ ] 测试通过

---

## 常见问题

### Q: 为什么所有迁移都在 PHP 中？

**A**:

- 统一管理，避免混乱
- PHP Phinx 功能强大
- Go model 只需同步更新

### Q: 时间字段为什么用 int 不用 datetime？

**A**:

- 跨时区兼容性好
- 存储效率高
- 便于计算和比较

### Q: 金额字段为什么用 decimal(20,8)？

**A**:

- 避免浮点数精度问题
- 支持大金额和高精度
- 行业标准

### Q: 如何处理旧数据迁移？

**A**:

1. 创建单独的迁移文件
2. 在 change() 方法中编写数据处理逻辑
3. 使用 SQL 批量更新
4. 测试验证后再执行

---

## 相关资源

### 规范文件

- `.cursor/rules/php.mdc` - PHP 迁移文件规范 ⭐⭐⭐
- `.cursor/rules/golang.mdc` - Go model 规范

### 模板

- `docs/agent/templates/database-migration-template.md` - 迁移模板

### 文档

- `docs/human/architecture/entities/` - 实体模型文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 建议：复杂迁移或需要特殊回滚方案时沉淀 Episode，并在 Spec/design 中互链。

---

**最后更新**: 2025-11-16  
**维护者**: 后端开发组
