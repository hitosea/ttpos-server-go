package gormcache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"go.uber.org/zap"
)

// LocalCacher 基于 BigCache 的本地内存缓存实现，支持表级精确失效
//
// 与 RedisCacher 平级，都实现 Cacher 接口，通过配置可切换使用
//
// 缓存 Key 结构与 RedisCacher 一致：
//
//	gorm:cache:{tableName}:{sql}:{params}
//
// 表索引使用 sync.Map 在内存中维护
type LocalCacher struct {
	cache      *bigcache.BigCache // bigcache 实例
	defaultTTL time.Duration      // 默认过期时间
	tableIndex sync.Map           // 表名 → *sync.Map{key → struct{}}
}

// LocalCacherOption 配置选项
type LocalCacherOption func(*localCacherConfig)

// localCacherConfig 本地缓存配置
type localCacherConfig struct {
	defaultTTL       time.Duration
	cleanWindow      time.Duration
	hardMaxCacheSize int // MB
	maxEntrySize     int // bytes
	shards           int
}

func defaultLocalCacherConfig() *localCacherConfig {
	return &localCacherConfig{
		defaultTTL:       5 * time.Minute,
		cleanWindow:      1 * time.Minute,
		hardMaxCacheSize: 256,        // 256 MB
		maxEntrySize:     500 * 1024, // 500 KB
		shards:           1024,
	}
}

// WithLocalDefaultTTL 设置默认过期时间
func WithLocalDefaultTTL(ttl time.Duration) LocalCacherOption {
	return func(c *localCacherConfig) {
		c.defaultTTL = ttl
	}
}

// WithLocalCleanWindow 设置过期条目清理间隔
func WithLocalCleanWindow(d time.Duration) LocalCacherOption {
	return func(c *localCacherConfig) {
		c.cleanWindow = d
	}
}

// WithLocalMaxSize 设置最大内存占用（MB）
func WithLocalMaxSize(mb int) LocalCacherOption {
	return func(c *localCacherConfig) {
		c.hardMaxCacheSize = mb
	}
}

// WithLocalMaxEntrySize 设置单条目最大大小（bytes）
func WithLocalMaxEntrySize(size int) LocalCacherOption {
	return func(c *localCacherConfig) {
		c.maxEntrySize = size
	}
}

// WithLocalShards 设置分片数（必须是 2 的幂）
func WithLocalShards(shards int) LocalCacherOption {
	return func(c *localCacherConfig) {
		c.shards = shards
	}
}

// NewLocalCacher 创建本地内存缓存实例
func NewLocalCacher(opts ...LocalCacherOption) *LocalCacher {
	cfg := defaultLocalCacherConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	bcConfig := bigcache.DefaultConfig(cfg.defaultTTL)
	bcConfig.CleanWindow = cfg.cleanWindow
	bcConfig.HardMaxCacheSize = cfg.hardMaxCacheSize
	bcConfig.MaxEntrySize = cfg.maxEntrySize
	bcConfig.Shards = cfg.shards
	// 静默丢弃过期条目，不打印日志
	bcConfig.Verbose = false

	bc, err := bigcache.New(context.Background(), bcConfig)
	if err != nil {
		logError("gormcache: 创建 BigCache 失败", zap.Error(err))
		// 回退到最简配置
		bc, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.defaultTTL))
	}

	return &LocalCacher{
		cache:      bc,
		defaultTTL: cfg.defaultTTL,
	}
}

// Get 从本地缓存获取查询结果
func (c *LocalCacher) Get(ctx context.Context, key string) (*QueryResult, bool) {
	data, err := c.cache.Get(key)
	if err != nil {
		return nil, false
	}

	var result QueryResult
	if err := json.Unmarshal(data, &result); err != nil {
		logWarn("gormcache: 本地缓存反序列化失败",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, false
	}

	return &result, true
}

// Store 将查询结果存入本地缓存
func (c *LocalCacher) Store(ctx context.Context, tableName string, key string, result *QueryResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		logWarn("gormcache: 本地缓存序列化失败",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	// 存储到 bigcache（bigcache 使用全局 TTL，由初始化时的 LifeWindow 控制）
	if err := c.cache.Set(key, data); err != nil {
		logWarn("gormcache: 本地缓存存储失败",
			zap.String("key", key),
			zap.String("table", tableName),
			zap.Error(err),
		)
		return err
	}

	// 更新表索引
	c.addToTableIndex(tableName, key)

	logDebug("gormcache: 本地缓存存储成功",
		zap.String("key", key),
		zap.String("table", tableName),
		zap.Int("size", len(data)),
	)

	return nil
}

// InvalidateTable 失效指定表的所有缓存
func (c *LocalCacher) InvalidateTable(ctx context.Context, tableName string) error {
	keys := c.getAndClearTableIndex(tableName)

	if len(keys) == 0 {
		logDebug("gormcache: 本地缓存表索引为空，跳过失效",
			zap.String("table", tableName),
		)
		return nil
	}

	// 逐个删除缓存条目
	for _, key := range keys {
		c.cache.Delete(key)
	}

	logDebug("gormcache: 本地缓存表失效成功",
		zap.String("table", tableName),
		zap.Int("keyCount", len(keys)),
	)

	return nil
}

// InvalidateAll 失效所有缓存
func (c *LocalCacher) InvalidateAll(ctx context.Context) error {
	// 清空表索引
	c.tableIndex.Range(func(key, value any) bool {
		c.tableIndex.Delete(key)
		return true
	})

	// 清空 bigcache
	if err := c.cache.Reset(); err != nil {
		logWarn("gormcache: 本地缓存全量清空失败", zap.Error(err))
		return err
	}

	logInfo("gormcache: 本地缓存全量失效完成")

	return nil
}

// addToTableIndex 添加到表索引
func (c *LocalCacher) addToTableIndex(tableName string, key string) {
	actual, _ := c.tableIndex.LoadOrStore(tableName, &sync.Map{})
	keySet := actual.(*sync.Map)
	keySet.Store(key, struct{}{})
}

// getAndClearTableIndex 获取并清空表索引
func (c *LocalCacher) getAndClearTableIndex(tableName string) []string {
	actual, ok := c.tableIndex.LoadAndDelete(tableName)
	if !ok {
		return nil
	}

	keySet := actual.(*sync.Map)
	var keys []string
	keySet.Range(func(k, v any) bool {
		keys = append(keys, k.(string))
		return true
	})

	return keys
}

// GetStats 获取缓存统计信息
func (c *LocalCacher) GetStats(ctx context.Context) map[string]int64 {
	stats := make(map[string]int64)

	c.tableIndex.Range(func(key, value any) bool {
		tableName := key.(string)
		keySet := value.(*sync.Map)
		var count int64
		keySet.Range(func(k, v any) bool {
			count++
			return true
		})
		stats[tableName] = count
		return true
	})

	return stats
}
