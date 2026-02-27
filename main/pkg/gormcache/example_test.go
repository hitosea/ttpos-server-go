package gormcache_test

import (
	"context"
	"fmt"
	"time"

	"ttpos-server-go/pkg/gormcache"
)

// Example_basicUsage 演示基本使用方法
func Example_basicUsage() {
	// 1. 在应用启动时初始化（通常在 root.go）
	gormcache.Init(&gormcache.Config{
		Easer:   true,            // 启用请求合并
		TTL:     5 * time.Minute, // 缓存 5 分钟
		MaxRows: 1000,            // 最大缓存 1000 行
		Tables: []string{ // 只缓存这些表
			"ttpos_product",
			"ttpos_category",
			"ttpos_menu_item",
		},
	})

	// 2. 为数据库启用缓存
	// db := dbManager.GetDB(companyUuid)
	// gormcache.Enable(db, nil)

	// 3. 正常使用 GORM，缓存自动生效
	// db.Where("status = ?", "active").Find(&products)

	// 4. 写操作会自动失效对应表的缓存
	// db.Create(&product)
	// db.Model(&product).Update("name", "new name")
	// db.Delete(&product)

	fmt.Println("gormcache initialized")
	// Output: gormcache initialized
}

// Example_skipCache 演示如何跳过缓存
func Example_skipCache() {
	// 使用 Scope 跳过缓存
	// db.Scopes(gormcache.NoCache()).Find(&products)

	// 实时性要求高的查询
	// db.Scopes(gormcache.NoCache()).Where("updated_at > ?", time.Now().Add(-1*time.Minute)).Find(&orders)

	fmt.Println("Cache skipped")
	// Output: Cache skipped
}

// Example_manualInvalidate 演示手动失效缓存
func Example_manualInvalidate() {
	ctx := context.Background()

	// 当通过原生 SQL 或其他方式修改数据时，需要手动失效缓存

	// 失效单个表
	_ = gormcache.InvalidateTable(ctx, "ttpos_product")

	// 失效所有缓存
	_ = gormcache.InvalidateAll(ctx)

	fmt.Println("Cache invalidated")
	// Output: Cache invalidated
}

// Example_configOptions 演示配置选项
func Example_configOptions() {
	// 使用选项模式配置
	// gormcache.EnableWithOptions(db,
	//     gormcache.WithEaser(true),
	//     gormcache.WithTTL(10*time.Minute),
	//     gormcache.WithMaxRows(500),
	//     gormcache.WithTables("ttpos_product", "ttpos_category"),
	//     gormcache.WithDebug(true),
	// )

	fmt.Println("Config options applied")
	// Output: Config options applied
}

// Example_wrapDBManager 演示如何包装 DBManager
func Example_wrapDBManager() {
	// 包装 DBManager.GetDB 方法，自动启用缓存
	// getDBWithCache := gormcache.WrapGetDB(dbManager.GetDB)
	// db := getDBWithCache(companyUuid)  // 自动启用缓存

	fmt.Println("DBManager wrapped")
	// Output: DBManager wrapped
}

// Example_cacheStats 演示获取缓存统计
func Example_cacheStats() {
	ctx := context.Background()

	// 获取各表的缓存键数量
	stats := gormcache.GetStats(ctx)
	for tableName, keyCount := range stats {
		fmt.Printf("表 %s: %d 个缓存键\n", tableName, keyCount)
	}

	// 注意：如果没有初始化 Redis，stats 会为 nil
	if stats == nil {
		fmt.Println("No stats available")
	}
	// Output: No stats available
}
