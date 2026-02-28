package gormcache

import (
	"context"
	"encoding/json"
	"time"

	"ttpos-server-go/pkg/cache"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCacher 基于 Redis 的缓存实现，支持表级精确失效
//
// 缓存 Key 结构：
//
//	gorm:cache:{dbName}:{tableName}:{sql}:{params}
//
// 表索引 Key 结构（ZSET 类型，score 为时间戳）：
//
//	gorm:cache:_idx:{dbName}:{tableName} → ZSET{(key1, timestamp1), (key2, timestamp2), ...}
//
// 注意：多实例部署时，索引始终从 Redis 获取，确保缓存一致性
type RedisCacher struct {
	client        redis.UniversalClient // Redis 客户端（支持单机和集群）
	defaultTTL    time.Duration         // 默认过期时间
	indexTTL      time.Duration         // 索引过期时间（比数据 TTL 长）
	maxIndexSize  int64                 // 索引最大数量，超过时淘汰最早的缓存
	enableMetrics bool                  // 是否启用指标统计
}

const (
	// indexKeyPrefix 表索引键前缀
	indexKeyPrefix = "gorm:cache:_idx:"
)

// RedisCacherOption 配置选项
type RedisCacherOption func(*RedisCacher)

// WithDefaultTTL 设置默认过期时间
func WithDefaultTTL(ttl time.Duration) RedisCacherOption {
	return func(c *RedisCacher) {
		c.defaultTTL = ttl
	}
}

// WithIndexTTL 设置索引过期时间
func WithIndexTTL(ttl time.Duration) RedisCacherOption {
	return func(c *RedisCacher) {
		c.indexTTL = ttl
	}
}

// WithMetrics 启用指标统计
func WithMetrics(enable bool) RedisCacherOption {
	return func(c *RedisCacher) {
		c.enableMetrics = enable
	}
}

// WithMaxIndexSize 设置索引最大数量
// 超过此数量时，会淘汰最早的缓存键
// 默认值 1000，设置为 0 表示不限制
func WithMaxIndexSize(size int64) RedisCacherOption {
	return func(c *RedisCacher) {
		c.maxIndexSize = size
	}
}

// NewRedisCacher 创建 Redis 缓存实例
// 使用项目全局的 cache.Global 作为底层存储
func NewRedisCacher(opts ...RedisCacherOption) *RedisCacher {
	c := &RedisCacher{
		defaultTTL:   5 * time.Minute,
		indexTTL:     6 * time.Minute, // 索引比数据多存 1 分钟
		maxIndexSize: 1000,            // 默认最大索引数量
	}

	for _, opt := range opts {
		opt(c)
	}

	// 获取 Redis 客户端
	if cache.Global != nil {
		if client := cache.Global.GetClient(); client != nil {
			c.client = client
		} else if clusterClient := cache.Global.GetClusterClient(); clusterClient != nil {
			c.client = clusterClient
		}
	}

	return c
}

// NewRedisCacherWithClient 使用指定的 Redis 客户端创建缓存实例
func NewRedisCacherWithClient(client redis.UniversalClient, opts ...RedisCacherOption) *RedisCacher {
	c := &RedisCacher{
		client:       client,
		defaultTTL:   5 * time.Minute,
		indexTTL:     6 * time.Minute,
		maxIndexSize: 1000, // 默认最大索引数量
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get 从缓存获取查询结果
func (c *RedisCacher) Get(ctx context.Context, key string) (*QueryResult, bool) {
	if c.client == nil {
		return nil, false
	}

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			logDebug("gormcache: Redis Get 失败",
				zap.String("key", key),
				zap.Error(err),
			)
		}
		return nil, false
	}

	var result QueryResult
	if err := json.Unmarshal(data, &result); err != nil {
		logWarn("gormcache: 反序列化失败",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, false
	}

	return &result, true
}

// Store 将查询结果存入缓存
// tableKey: 表索引键，格式为 {dbName}:{tableName}，用于多租户隔离
func (c *RedisCacher) Store(ctx context.Context, tableKey string, key string, result *QueryResult, ttl time.Duration) error {
	if c.client == nil {
		return nil
	}

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		logWarn("gormcache: 序列化失败",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	// indexKey 格式: gorm:cache:_idx:{dbName}:{tableName}
	indexKey := indexKeyPrefix + tableKey
	now := float64(time.Now().UnixNano())

	// 使用 Pipeline 批量操作
	pipe := c.client.Pipeline()

	// 1. 存储缓存数据
	pipe.Set(ctx, key, data, ttl)

	// 2. 更新表索引（ZSET 类型，score 为时间戳）
	pipe.ZAdd(ctx, indexKey, redis.Z{Score: now, Member: key})
	pipe.Expire(ctx, indexKey, c.indexTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		logWarn("gormcache: Pipeline 执行失败",
			zap.String("key", key),
			zap.String("tableKey", tableKey),
			zap.Error(err),
		)
		return err
	}

	// 3. 检查索引大小，超限时淘汰最早的缓存
	if c.maxIndexSize > 0 {
		c.evictOldestIfNeeded(ctx, indexKey, tableKey)
	}

	logDebug("gormcache: 缓存存储成功",
		zap.String("key", key),
		zap.String("tableKey", tableKey),
		zap.Duration("ttl", ttl),
		zap.Int("size", len(data)),
	)

	return nil
}

// evictOldestIfNeeded 如果索引超限，淘汰最早的缓存
func (c *RedisCacher) evictOldestIfNeeded(ctx context.Context, indexKey string, tableKey string) {
	// 获取当前索引大小
	size, err := c.client.ZCard(ctx, indexKey).Result()
	if err != nil || size <= c.maxIndexSize {
		return
	}

	// 计算需要淘汰的数量
	evictCount := size - c.maxIndexSize

	// 获取最早的 N 个缓存键（score 最小的）
	oldKeys, err := c.client.ZRange(ctx, indexKey, 0, evictCount-1).Result()
	if err != nil || len(oldKeys) == 0 {
		return
	}

	// 删除缓存数据和索引
	for _, key := range oldKeys {
		c.client.Del(ctx, key)
		c.client.ZRem(ctx, indexKey, key)
	}

	logDebug("gormcache: 淘汰最早的缓存",
		zap.String("tableKey", tableKey),
		zap.Int("evictCount", len(oldKeys)),
		zap.Int64("currentSize", size),
		zap.Int64("maxSize", c.maxIndexSize),
	)
}

// InvalidateTable 失效指定表的所有缓存
// tableKey: 表索引键，格式为 {dbName}:{tableName}，用于多租户隔离
// 注意：始终从 Redis 获取索引，确保多实例部署时缓存一致性
func (c *RedisCacher) InvalidateTable(ctx context.Context, tableKey string) error {
	if c.client == nil {
		return nil
	}

	// indexKey 格式: gorm:cache:_idx:{dbName}:{tableName}
	indexKey := indexKeyPrefix + tableKey

	// 获取索引大小
	keyCount, err := c.client.ZCard(ctx, indexKey).Result()
	if err != nil && err != redis.Nil {
		logWarn("gormcache: 获取索引大小失败",
			zap.String("tableKey", tableKey),
			zap.Error(err),
		)
	}

	if keyCount == 0 {
		logDebug("gormcache: 表缓存为空，跳过失效",
			zap.String("tableKey", tableKey),
		)
		return nil
	}

	// 从 ZSET 获取所有缓存键
	keys, err := c.client.ZRange(ctx, indexKey, 0, -1).Result()
	if err != nil && err != redis.Nil {
		logWarn("gormcache: 获取表索引失败",
			zap.String("tableKey", tableKey),
			zap.Error(err),
		)
	}

	if len(keys) == 0 {
		return nil
	}

	// 删除缓存数据
	// 注意：Redis 集群模式下，不同 key 可能在不同 slot，不能用 Pipeline 批量删除
	// 逐个删除以兼容集群模式
	var delErr error
	for _, key := range keys {
		if err := c.client.Del(ctx, key).Err(); err != nil && err != redis.Nil {
			delErr = err
		}
	}

	// 删除索引
	if err := c.client.Del(ctx, indexKey).Err(); err != nil && err != redis.Nil {
		delErr = err
	}

	if delErr != nil {
		logWarn("gormcache: 缓存删除部分失败",
			zap.String("tableKey", tableKey),
			zap.Int("keyCount", len(keys)),
			zap.Error(delErr),
		)
		return delErr
	}

	logDebug("gormcache: 表缓存失效成功",
		zap.String("tableKey", tableKey),
		zap.Int("keyCount", len(keys)),
	)

	return nil
}

// InvalidateAll 失效所有缓存
func (c *RedisCacher) InvalidateAll(ctx context.Context) error {
	if c.client == nil {
		return nil
	}

	// 使用 SCAN 查找所有缓存键
	pattern := IdentifierPrefix + "*"
	keys, err := cache.ScanRedisKeysDefault(ctx, c.client, pattern)
	if err != nil {
		logWarn("gormcache: Scan 失败",
			zap.String("pattern", pattern),
			zap.Error(err),
		)
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 逐个删除（Redis 集群模式下不同 key 可能在不同 slot）
	for _, key := range keys {
		if err := c.client.Del(ctx, key).Err(); err != nil && err != redis.Nil {
			logWarn("gormcache: 删除失败",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}

	logInfo("gormcache: 全量缓存失效完成",
		zap.Int("keyCount", len(keys)),
	)

	return nil
}

// GetStats 获取缓存统计信息
func (c *RedisCacher) GetStats(ctx context.Context) map[string]int64 {
	if c.client == nil {
		return nil
	}

	stats := make(map[string]int64)

	// 扫描所有索引键
	pattern := indexKeyPrefix + "*"
	keys, err := cache.ScanRedisKeysDefault(ctx, c.client, pattern)
	if err != nil {
		return stats
	}

	for _, key := range keys {
		tableKey := key[len(indexKeyPrefix):]
		count, _ := c.client.ZCard(ctx, key).Result()
		stats[tableKey] = count
	}

	return stats
}
