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
			zap.Int("allowDBs", len(conf.AllowDBs)),
			zap.Int("excludeDBs", len(conf.ExcludeDBs)),
		)
	})
}

// Enable 为指定的 GORM 数据库启用缓存
// dbName: 数据库名称（如 "saas" 或 "shop123456"），用于多租户隔离
// 如果 conf 为 nil，使用全局配置
//
// 示例：
//
//	database.SetOnDBReady(func(db *gorm.DB, dbName string) {
//	    gormcache.Enable(db, dbName, nil)  // 使用全局配置
//	})
//
//	// 或使用自定义配置
//	gormcache.Enable(db, dbName, &gormcache.Config{
//	    Tables: []string{"ttpos_menu_item"},
//	    TTL: 10 * time.Minute,
//	})
func Enable(db *gorm.DB, dbName string, conf *Config) error {
	if db == nil {
		return nil
	}

	// 获取基础配置
	baseConf := conf
	if baseConf == nil {
		baseConf = globalConfig
	}
	if baseConf == nil {
		baseConf = DefaultConfig()
	}

	// 复制配置，避免污染全局配置
	// 每个数据库需要独立的 Config 实例（DBName 不同）
	instanceConf := baseConf.Copy()
	instanceConf.DBName = dbName

	// 如果基础配置没有 Cacher，创建默认的
	if instanceConf.Cacher == nil {
		instanceConf.Cacher = NewRedisCacher()
	}

	// 白名单模式（内测）：只允许指定商户使用缓存
	if len(instanceConf.AllowDBs) > 0 {
		allowed := false
		for _, allowDB := range instanceConf.AllowDBs {
			if dbName == allowDB {
				allowed = true
				break
			}
		}
		if !allowed {
			// 不在白名单中，跳过
			return nil
		}
	} else if len(instanceConf.ExcludeDBs) > 0 {
		// 黑名单模式：排除指定商户
		for _, excludeDB := range instanceConf.ExcludeDBs {
			if dbName == excludeDB {
				logDebug("gormcache: 数据库在黑名单中，跳过启用",
					zap.String("db", dbName),
				)
				return nil
			}
		}
	}

	// 检查是否已启用
	dbPtr := getDBPointer(db)
	if _, loaded := enabledDBs.LoadOrStore(dbPtr, true); loaded {
		return nil // 已启用，跳过
	}

	// 创建并注册插件
	plugin := New(instanceConf)
	if err := db.Use(plugin); err != nil {
		enabledDBs.Delete(dbPtr)
		return err
	}

	logInfo("gormcache: 数据库缓存已启用",
		zap.String("db", dbName),
	)

	return nil
}

// EnableWithOptions 使用选项模式启用缓存
func EnableWithOptions(db *gorm.DB, dbName string, opts ...ConfigOption) error {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return Enable(db, dbName, conf)
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

// WithExcludeDBs 设置排除数据库黑名单（商户级别禁用缓存）
// 用于某些商户不希望使用 GORM 缓存的场景
func WithExcludeDBs(dbs ...string) ConfigOption {
	return func(c *Config) {
		c.ExcludeDBs = dbs
	}
}

// WithAllowDBs 设置允许数据库白名单（商户级别，用于内测）
// 设置后只有白名单中的商户才会启用缓存，ExcludeDBs 配置无效
func WithAllowDBs(dbs ...string) ConfigOption {
	return func(c *Config) {
		c.AllowDBs = dbs
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

// WithLocalCache 使用默认配置的本地内存缓存（BigCache）
// 使用此选项将替代 Redis 缓存
func WithLocalCache() ConfigOption {
	return func(c *Config) {
		c.Cacher = NewLocalCacher(
			WithLocalDefaultTTL(c.TTL),
		)
	}
}

// WithLocalCacheOptions 使用自定义配置的本地内存缓存（BigCache）
func WithLocalCacheOptions(opts ...LocalCacherOption) ConfigOption {
	return func(c *Config) {
		c.Cacher = NewLocalCacher(opts...)
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
	switch c := globalCacher.(type) {
	case *RedisCacher:
		return c.GetStats(ctx)
	case *LocalCacher:
		return c.GetStats(ctx)
	default:
		return nil
	}
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
// 已废弃：推荐使用 database.SetOnDBReady 回调方式，可获取准确的 dbName
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
			// 使用 extractDBName 回退方式获取数据库名
			Enable(db, extractDBName(db), nil)
		}
		return db
	}
}

// AutoEnableOnConnect 创建一个在连接创建时自动启用缓存的钩子
// 返回一个可以在 GORM 配置中使用的 ConnPool 包装器
//
// 已废弃：推荐使用 database.SetOnDBReady 回调方式，可获取准确的 dbName
//
// 注意：此方法需要在创建数据库连接时调用，不适合已存在的连接
func AutoEnableOnConnect(conf *Config) func(*gorm.DB) {
	return func(db *gorm.DB) {
		// 使用 extractDBName 回退方式获取数据库名
		Enable(db, extractDBName(db), conf)
	}
}
