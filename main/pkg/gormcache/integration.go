package gormcache

import (
	"context"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ============ 全局配置管理 ============

var (
	// globalConfig 全局缓存配置
	globalConfig *Config

	// globalCacher 全局缓存实例
	globalCacher Cacher

	// configOnce 确保配置只初始化一次
	configOnce sync.Once

	// enabledDBs 已启用缓存的数据库
	enabledDBs sync.Map
)

// Init 初始化全局 GORM 缓存
// 应在应用启动时调用，在 cache.Init 之后
//
// 示例：
//
//	gormcache.Init(&gormcache.Config{
//	    Easer: true,
//	    TTL:   5 * time.Minute,
//	    Tables: []string{"ttpos_product", "ttpos_category"},
//	})
func Init(conf *Config) {
	configOnce.Do(func() {
		if conf == nil {
			conf = DefaultConfig()
		}
		globalConfig = conf

		// 创建默认 Redis 缓存器
		if conf.Cacher == nil {
			globalCacher = NewRedisCacher(
				WithDefaultTTL(conf.TTL),
			)
			globalConfig.Cacher = globalCacher
		} else {
			globalCacher = conf.Cacher
		}

		logInfo("gormcache: 全局初始化完成",
			zap.Bool("easer", conf.Easer),
			zap.Duration("ttl", conf.TTL),
			zap.Int64("maxRows", conf.MaxRows),
			zap.Int("tables", len(conf.Tables)),
			zap.Int("excludeTables", len(conf.ExcludeTables)),
		)
	})
}

// Enable 为指定的 GORM 数据库启用缓存
// 如果 conf 为 nil，使用全局配置
//
// 示例：
//
//	db := dbManager.GetDB(companyUuid)
//	gormcache.Enable(db, nil)  // 使用全局配置
//
//	// 或使用自定义配置
//	gormcache.Enable(db, &gormcache.Config{
//	    Tables: []string{"ttpos_menu_item"},
//	    TTL: 10 * time.Minute,
//	})
func Enable(db *gorm.DB, conf *Config) error {
	if db == nil {
		return nil
	}

	// 检查是否已启用
	dbPtr := getDBPointer(db)
	if _, loaded := enabledDBs.LoadOrStore(dbPtr, true); loaded {
		return nil // 已启用，跳过
	}

	// 使用传入配置或全局配置
	if conf == nil {
		conf = globalConfig
	}
	if conf == nil {
		conf = DefaultConfig()
		conf.Cacher = NewRedisCacher()
	}

	// 创建并注册插件
	plugin := New(conf)
	if err := db.Use(plugin); err != nil {
		enabledDBs.Delete(dbPtr)
		return err
	}

	logDebug("gormcache: 数据库缓存已启用",
		zap.String("db", db.Name()),
	)

	return nil
}

// EnableWithOptions 使用选项模式启用缓存
func EnableWithOptions(db *gorm.DB, opts ...ConfigOption) error {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return Enable(db, conf)
}

// getDBPointer 获取数据库的唯一标识
func getDBPointer(db *gorm.DB) uintptr {
	// 使用 db.Config 的地址作为唯一标识
	if db.Config != nil {
		return reflect.ValueOf(db.Config).Pointer()
	}
	return 0
}

// ============ 配置选项模式 ============

// ConfigOption 配置选项函数
type ConfigOption func(*Config)

// WithEaser 启用请求合并
func WithEaser(enable bool) ConfigOption {
	return func(c *Config) {
		c.Easer = enable
	}
}

// WithTTL 设置默认缓存时间
func WithTTL(ttl time.Duration) ConfigOption {
	return func(c *Config) {
		c.TTL = ttl
	}
}

// WithMaxRows 设置最大缓存行数
func WithMaxRows(maxRows int64) ConfigOption {
	return func(c *Config) {
		c.MaxRows = maxRows
	}
}

// WithTables 设置缓存表白名单
func WithTables(tables ...string) ConfigOption {
	return func(c *Config) {
		c.Tables = tables
	}
}

// WithExcludeTables 设置排除表黑名单
func WithExcludeTables(tables ...string) ConfigOption {
	return func(c *Config) {
		c.ExcludeTables = tables
	}
}

// WithDebug 启用调试日志
func WithDebug(enable bool) ConfigOption {
	return func(c *Config) {
		c.Debug = enable
	}
}

// WithCacher 设置自定义缓存器
func WithCacher(cacher Cacher) ConfigOption {
	return func(c *Config) {
		c.Cacher = cacher
	}
}

// ============ 便捷操作 ============

// InvalidateTable 失效指定表的缓存
// 可在手动更新数据后调用
func InvalidateTable(ctx context.Context, tableName string) error {
	if globalCacher == nil {
		return nil
	}
	return globalCacher.InvalidateTable(ctx, tableName)
}

// InvalidateAll 失效所有缓存
func InvalidateAll(ctx context.Context) error {
	if globalCacher == nil {
		return nil
	}
	return globalCacher.InvalidateAll(ctx)
}

// GetStats 获取缓存统计信息
func GetStats(ctx context.Context) map[string]int64 {
	if rc, ok := globalCacher.(*RedisCacher); ok {
		return rc.GetStats(ctx)
	}
	return nil
}

// ============ GORM Scopes ============

// NoCache 跳过缓存的 Scope
// 用法: db.Scopes(gormcache.NoCache()).Find(&users)
func NoCache() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		ctx := context.WithValue(db.Statement.Context, skipCacheKey{}, true)
		return db.WithContext(ctx)
	}
}

// ============ DBManager 集成辅助 ============

// WrapDBManager 包装 DBManager 的 GetDB 方法，自动启用缓存
// 这是一个辅助函数，用于在不修改 DBManager 源码的情况下集成缓存
//
// 示例（在 root.go 初始化后使用）：
//
//	// 原始方式
//	db := dbManager.GetDB(companyUuid)
//
//	// 包装后，自动启用缓存
//	getDBWithCache := gormcache.WrapGetDB(dbManager.GetDB)
//	db := getDBWithCache(companyUuid)
func WrapGetDB(getDB func(uint64) *gorm.DB) func(uint64) *gorm.DB {
	return func(index uint64) *gorm.DB {
		db := getDB(index)
		if db != nil {
			// 尝试启用缓存（如果已启用会自动跳过）
			Enable(db, nil)
		}
		return db
	}
}

// AutoEnableOnConnect 创建一个在连接创建时自动启用缓存的钩子
// 返回一个可以在 GORM 配置中使用的 ConnPool 包装器
//
// 注意：此方法需要在创建数据库连接时调用，不适合已存在的连接
func AutoEnableOnConnect(conf *Config) func(*gorm.DB) {
	return func(db *gorm.DB) {
		Enable(db, conf)
	}
}
