/*
Package gormcache 提供 GORM 查询缓存功能，支持表级精确缓存失效。

# 核心特性

  - 表级缓存：按表名组织缓存，写操作只失效对应表的缓存
  - Easer 请求合并：相同查询并发时只执行一次数据库查询
  - 与项目现有 cache 包无缝集成
  - 支持单机和 Redis 集群模式
  - 白名单/黑名单控制缓存范围

# 快速开始

1. 在应用启动时初始化（通常在 root.go 的 initCache 之后）：

	import "ttpos-server-go/pkg/gormcache"

	// 方式一：使用默认配置
	gormcache.Init(nil)

	// 方式二：自定义配置
	gormcache.Init(&gormcache.Config{
	    Easer:   true,                    // 启用请求合并
	    TTL:     5 * time.Minute,         // 缓存 5 分钟
	    MaxRows: 1000,                    // 最大缓存 1000 行
	    Tables:  []string{                // 只缓存这些表
	        "ttpos_product",
	        "ttpos_category",
	        "ttpos_menu_item",
	    },
	    Debug: false,
	})

2. 为数据库启用缓存：

	// 在 DBManager.GetDB 返回后启用
	db := dbManager.GetDB(companyUuid)
	gormcache.Enable(db, nil)  // 使用全局配置

	// 或者使用包装函数自动启用
	getDBWithCache := gormcache.WrapGetDB(dbManager.GetDB)
	db := getDBWithCache(companyUuid)

3. 正常使用 GORM，缓存自动生效：

	// 查询会自动使用缓存
	db.Where("status = ?", "active").Find(&products)

	// 写操作会自动失效对应表的缓存
	db.Create(&product)
	db.Model(&product).Update("name", "new name")
	db.Delete(&product)

# 跳过缓存

某些场景需要跳过缓存直接查询数据库：

	// 使用 Scope
	db.Scopes(gormcache.NoCache()).Find(&products)

	// 实时性要求高的查询
	db.Scopes(gormcache.NoCache()).Where("updated_at > ?", time.Now().Add(-1*time.Minute)).Find(&orders)

# 手动失效缓存

当通过原生 SQL 或其他方式修改数据时，需要手动失效缓存：

	// 失效单个表
	gormcache.InvalidateTable(ctx, "ttpos_product")

	// 失效所有缓存
	gormcache.InvalidateAll(ctx)

# 缓存统计

获取各表的缓存键数量：

	stats := gormcache.GetStats(ctx)
	for tableName, keyCount := range stats {
	    fmt.Printf("表 %s: %d 个缓存键\n", tableName, keyCount)
	}

# 配置选项

使用选项模式配置：

	gormcache.EnableWithOptions(db,
	    gormcache.WithEaser(true),
	    gormcache.WithTTL(10*time.Minute),
	    gormcache.WithMaxRows(500),
	    gormcache.WithTables("ttpos_product", "ttpos_category"),
	    gormcache.WithDebug(true),
	)

# 最佳实践

1. 只缓存更新频率低的表：

	gormcache.Init(&gormcache.Config{
	    Tables: []string{
	        "ttpos_product",      // 商品表，更新不频繁
	        "ttpos_category",     // 分类表，很少更新
	        "ttpos_table_area",   // 区域表，基本不更新
	    },
	})

2. 排除频繁更新的表：

	gormcache.Init(&gormcache.Config{
	    ExcludeTables: []string{
	        "ttpos_order",        // 订单表，频繁写入
	        "ttpos_order_item",   // 订单项，频繁写入
	        "ttpos_stock_log",    // 库存日志，高频写入
	    },
	})

3. 合理设置 TTL：

	// 配置类数据，可以设置较长 TTL
	gormcache.WithTTL(30 * time.Minute)

	// 业务数据，建议较短 TTL
	gormcache.WithTTL(5 * time.Minute)

4. 限制缓存结果大小：

	// 防止大查询占用过多内存
	gormcache.WithMaxRows(100)

# 架构说明

	┌─────────────────────────────────────────────────────────────┐
	│                     Application                              │
	│                          │                                   │
	│                    db.Find(&users)                           │
	└─────────────────────────────────────────────────────────────┘
	                           │
	                           ▼
	┌─────────────────────────────────────────────────────────────┐
	│                   GORM Callbacks                             │
	│  ┌──────────┐   ┌──────────┐   ┌──────────┐                 │
	│  │ gormcache│ → │  Query   │ → │ gormcache│                 │
	│  │  Before  │   │          │   │  After   │                 │
	│  └──────────┘   └──────────┘   └──────────┘                 │
	└─────────────────────────────────────────────────────────────┘
	         │                              │
	         ▼                              ▼
	┌─────────────────┐            ┌─────────────────┐
	│  Check Cache    │            │  Store Cache    │
	│  + Easer        │            │  + Index        │
	└─────────────────┘            └─────────────────┘
	         │                              │
	         ▼                              ▼
	┌─────────────────────────────────────────────────────────────┐
	│                      Redis                                   │
	│  gorm:cache:{table}:{sql}:{params} → QueryResult            │
	│  gorm:cache:_idx:{table}           → SET{keys...}           │
	└─────────────────────────────────────────────────────────────┘

# 表级失效原理

写操作时只失效对应表的缓存，不影响其他表：

	UPDATE users SET name = 'new'
	         │
	         ▼
	┌─────────────────────────────────────┐
	│ 1. 提取表名: users                   │
	│ 2. 获取索引: gorm:cache:_idx:users   │
	│ 3. 删除所有 users 相关的缓存键        │
	│ 4. 其他表（orders, products）不受影响 │
	└─────────────────────────────────────┘

# 注意事项

  - 缓存仅对 Query 操作生效，Row 操作不缓存
  - JOIN 查询的缓存失效需要特殊处理（两个表都需要失效）
  - 原生 SQL 更新后需要手动调用 InvalidateTable
  - 事务中的查询也会使用缓存，注意一致性问题
*/
package gormcache
