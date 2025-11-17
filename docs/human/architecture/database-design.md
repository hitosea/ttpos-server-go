# 数据库架构设计

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 深入理解数据库架构设计理念和多租户实现

---

## 核心设计理念

### 多租户隔离

采用**独立数据库模式**，每个商户拥有独立的数据库：

```
saas                      # 系统主库
shop8267304538112000      # 商户1数据库
shop8609817471094784      # 商户2数据库
shop8609817471094785      # 商户3数据库
```

**为什么选择独立数据库?**

| 方案 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| 共享库+tenant_id | 成本低 | 数据安全风险、扩展困难 | ❌ |
| 独立Schema | 中等成本 | 管理复杂 | ❌ |
| 独立数据库 | 完全隔离、易扩展 | 成本高 | ✅ |

---

## 数据库架构

### 系统库（saas）

**职责**: 存储全局数据和商户信息

**核心表**:
```sql
-- 商户表
CREATE TABLE ttpos_company (
    id bigint PRIMARY KEY,
    uuid bigint UNIQUE,
    company_name varchar(100),
    db_name varchar(50),         -- shop{company_id}
    status tinyint,
    create_time int
);

-- 系统配置表
CREATE TABLE ttpos_config (
    id bigint PRIMARY KEY,
    config_key varchar(50),
    config_value text,
    create_time int
);

-- 字典表
CREATE TABLE ttpos_dictionary (
    id bigint PRIMARY KEY,
    dict_type varchar(50),
    dict_label varchar(100),
    dict_value varchar(100),
    sort_order int
);
```

---

### 商户库（shop{company_id}）

每个商户的数据库包含相同的表结构：

**核心表设计**:

```sql
-- 商品表
CREATE TABLE ttpos_product (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    uuid bigint UNIQUE,
    product_name varchar(100),
    category_id bigint,
    price decimal(14,2),
    stock int,
    status tinyint,
    create_time int,
    update_time int,
    delete_time int
);

-- 订单表
CREATE TABLE ttpos_order (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    uuid bigint UNIQUE,
    order_no varchar(50) UNIQUE,
    member_uuid bigint,
    total_amount decimal(14,2),
    paid_amount decimal(14,2),
    status tinyint,
    create_time int,
    update_time int
);

-- 订单明细表
CREATE TABLE ttpos_order_item (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    order_uuid bigint,
    product_uuid bigint,
    product_name varchar(100),
    quantity int,
    price decimal(14,2),
    amount decimal(14,2)
);
```

---

## 数据库连接管理

### DBManager 设计

**目的**: 动态管理多个商户数据库连接

```go
type DBManager struct {
    dbPool map[string]*gorm.DB  // dbId -> DB连接
    mutex  sync.RWMutex
}

func (m *DBManager) GetDB(dbId string) *gorm.DB {
    m.mutex.RLock()
    if db, ok := m.dbPool[dbId]; ok {
        m.mutex.RUnlock()
        return db
    }
    m.mutex.RUnlock()
    
    // 动态创建连接
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    db := m.createConnection(dbId)
    m.dbPool[dbId] = db
    return db
}

func (m *DBManager) createConnection(dbId string) *gorm.DB {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.DB.Username,
        config.DB.Password,
        config.DB.Host,
        config.DB.Port,
        "shop"+dbId,
    )
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // 配置连接池
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db
}
```

---

## 主从架构

### 读写分离

```
┌──────────┐        ┌──────────┐
│  Master  │◄───────│  Slave1  │
│  (写)    │        │  (读)    │
└────┬─────┘        └──────────┘
     │
     │ 复制
     │
     ▼
┌──────────┐
│  Slave2  │
│  (读)    │
└──────────┘
```

**实现**:
```go
// 写操作 - Master
db.Create(&order)
db.Update(&order)
db.Delete(&order)

// 读操作 - Slave（自动路由）
db.Where("id = ?", orderId).First(&order)
db.Find(&orders)
```

**配置**:
```yaml
database:
  master:
    host: 192.168.1.10
    port: 3306
  slaves:
    - host: 192.168.1.11
      port: 3306
    - host: 192.168.1.12
      port: 3306
```

---

## 索引设计

### 索引原则

1. **主键索引**: 每个表必须有主键（id）
2. **唯一索引**: 唯一字段添加唯一索引（uuid、订单号）
3. **普通索引**: 查询频繁的字段添加索引
4. **复合索引**: 组合查询的字段建立复合索引

### 索引示例

```sql
-- 主键
PRIMARY KEY (id)

-- 唯一索引
UNIQUE KEY uk_uuid (uuid)
UNIQUE KEY uk_order_no (order_no)

-- 普通索引
KEY idx_create_time (create_time)
KEY idx_status (status)
KEY idx_member_uuid (member_uuid)

-- 复合索引（注意顺序）
KEY idx_company_status (company_uuid, status)
KEY idx_member_time (member_uuid, create_time)
```

### 复合索引最左前缀原则

```sql
-- 索引: idx_company_status_time (company_uuid, status, create_time)

-- ✅ 可以使用索引
WHERE company_uuid = 123
WHERE company_uuid = 123 AND status = 1
WHERE company_uuid = 123 AND status = 1 AND create_time > xxx

-- ❌ 无法使用索引
WHERE status = 1
WHERE create_time > xxx
WHERE status = 1 AND create_time > xxx
```

---

## 分库分表（规划）

### 当前架构

每个商户一个数据库，但单表存储所有数据。

### 未来规划

当单表数据量超过 1000万 时，考虑分表：

```
shop8267304538112000
├── ttpos_order_202301      # 2023年1月订单
├── ttpos_order_202302      # 2023年2月订单
├── ttpos_order_202303      # 2023年3月订单
└── ...
```

**分表策略**:
- 按时间分表（月度/年度）
- 按订单ID哈希分表
- 历史数据归档

---

## 备份策略

### 备份方案

1. **全量备份**: 每天凌晨 2:00
2. **增量备份**: 每小时一次
3. **binlog备份**: 实时复制

### 备份脚本

```bash
#!/bin/bash
# 全量备份
mysqldump -u root -p --all-databases > backup_$(date +%Y%m%d).sql

# 备份到远程服务器
rsync -avz backup_*.sql backup@remote:/backups/
```

### 恢复流程

1. 停止应用服务
2. 恢复数据库
```bash
mysql -u root -p < backup_20231117.sql
```
3. 应用 binlog
4. 验证数据
5. 启动服务

---

## 性能优化

### 查询优化

```sql
-- ✅ 使用索引
EXPLAIN SELECT * FROM ttpos_order WHERE uuid = 123;

-- ✅ 避免全表扫描
EXPLAIN SELECT * FROM ttpos_order WHERE status = 1;

-- ✅ 使用覆盖索引
SELECT id, uuid, order_no FROM ttpos_order WHERE status = 1;
```

### 慢查询日志

```ini
# my.cnf
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 1  # 超过1秒的查询
```

---

## 相关文档

- [数据库开发指南](../guides/database-guide.md) - 数据库操作详细指南
- [系统架构总览](./overview.md) - 整体架构设计
- [Go Main 架构](./go-main-architecture.md) - Main 模块数据库使用

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

