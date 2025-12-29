package cache

import (
	"context"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ObjectStorageCacheConfigKey 对象存储缓存配置的 key（用于日志过滤）
// 此常量与 objectstorage 模块中的 SettingObjectStorageCache 保持一致
const ObjectStorageCacheConfigKey = "object_storage_cache"

// cacheGroup 是 ICacheGroup 的具体实现
// 它协调 L1 本地缓存、L2 Redis 缓存和 Singleflight 请求合并
type cacheGroup[T any] struct {
	config GroupConfig
	l1     *localCache[T]
	l2     *redisCacheWrapper[T]
	sf     *singleflightEngine[T]
}

// NewCacheGroup 创建一个泛型缓存组
func NewCacheGroup[T any](config GroupConfig) ICacheGroup[T] {
	g := &cacheGroup[T]{
		config: config,
		sf:     newSingleflightEngine[T](),
	}

	if config.EnableLocalCache {
		g.l1 = newLocalCache[T]()
	}

	if config.EnableRedisCache {
		g.l2 = newRedisCacheWrapper[T]()
	}

	return g
}

// Do 执行任务，实现多级缓存和请求合并的核心逻辑
func (g *cacheGroup[T]) Do(ctx context.Context, task Task[T]) (T, error) {
	key := task.Hash()

	// 1. 尝试从 L1 本地缓存读取
	if g.config.EnableLocalCache {
		if val, ok := g.l1.get(key); ok {
			if key != ObjectStorageCacheConfigKey { // 排除对象存储缓存配置的日志
				logger.Logger.Debug("缓存命中 L1",
					zap.String("key", key),
					zap.String("level", "L1"),
					zap.String("type", "GET"),
				)
			}
			return val, nil
		}
	}

	// 2. 尝试从 L2 Redis 缓存读取
	if g.config.EnableRedisCache {
		if val, ok := g.l2.get(ctx, key); ok {
			// 命中 L2，回填 L1（使用 L1 TTL）
			if g.config.EnableLocalCache {
				l1TTL := task.TTL()
				if g.config.L1TTL > 0 {
					l1TTL = g.config.L1TTL
				}
				g.l1.set(key, val, l1TTL)
				logger.Logger.Debug("缓存写入 L1",
					zap.String("key", key),
					zap.String("level", "L1"),
					zap.String("type", "GET"),
					zap.Duration("ttl", l1TTL),
					zap.Duration("task_ttl", task.TTL()),
				)
			}
			logger.Logger.Debug("缓存命中 L2",
				zap.String("key", key),
				zap.String("level", "L2"),
				zap.String("type", "GET"),
			)
			return val, nil
		}
	}

	// 3. L1/L2 均未命中，使用 Singleflight 合并并发请求
	logger.Logger.Debug("缓存未命中",
		zap.String("key", key),
		zap.String("type", "GET"),
	)
	return g.sf.do(ctx, key, func(ctx context.Context) (T, error) {
		// Double Check: 进入 Singleflight 后再次检查缓存，防止并发下的重复执行
		// (因为在等待 Singleflight 锁的过程中，可能前面的请求已经填好缓存了)
		if g.config.EnableRedisCache {
			if val, ok := g.l2.get(ctx, key); ok {
				logger.Logger.Debug("Singleflight Double Check 缓存命中 L2",
					zap.String("key", key),
					zap.String("level", "L2"),
					zap.String("type", "GET"),
				)
				return val, nil
			}
		}

		// 执行实际的业务逻辑
		val, err := task.Exec(ctx)
		if err != nil {
			// 处理负缓存逻辑：防止缓存穿透
			if g.config.NegativeTTL > 0 {
				var zero T
				if g.config.EnableRedisCache {
					_ = g.l2.set(ctx, key, zero, g.config.NegativeTTL)
				}
				if g.config.EnableLocalCache {
					g.l1.set(key, zero, g.config.NegativeTTL)
				}
			}
			var zero T
			return zero, err
		}

		// 填充缓存（使用阶梯式 TTL）
		taskTTL := task.TTL()

		// 确定 L1 TTL：如果配置了 L1TTL 则使用配置值，否则使用任务 TTL
		l1TTL := taskTTL
		if g.config.L1TTL > 0 {
			l1TTL = g.config.L1TTL
		}

		// 确定 L2 TTL：如果配置了 L2TTL 则使用配置值，否则使用任务 TTL
		l2TTL := taskTTL
		if g.config.L2TTL > 0 {
			l2TTL = g.config.L2TTL
		}

		if g.config.EnableRedisCache {
			if err := g.l2.set(ctx, key, val, l2TTL); err == nil {
				logger.Logger.Debug("缓存写入 L2",
					zap.String("key", key),
					zap.String("level", "L2"),
					zap.String("type", "GET"),
					zap.Duration("ttl", l2TTL),
					zap.Duration("task_ttl", taskTTL),
				)
			}
		}
		if g.config.EnableLocalCache {
			g.l1.set(key, val, l1TTL)
			logger.Logger.Debug("缓存写入 L1",
				zap.String("key", key),
				zap.String("level", "L1"),
				zap.String("type", "GET"),
				zap.Duration("ttl", l1TTL),
				zap.Duration("task_ttl", taskTTL),
			)
		}

		return val, nil
	})
}

// ClearL1 清空 L1 本地缓存
func (g *cacheGroup[T]) ClearL1() {
	if g.config.EnableLocalCache && g.l1 != nil {
		g.l1.flush()
	}
}
