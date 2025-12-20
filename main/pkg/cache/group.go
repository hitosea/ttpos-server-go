package cache

import (
	"context"
)

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
			return val, nil
		}
	}

	// 2. 尝试从 L2 Redis 缓存读取
	if g.config.EnableRedisCache {
		if val, ok := g.l2.get(ctx, key); ok {
			// 命中 L2，回填 L1
			if g.config.EnableLocalCache {
				g.l1.set(key, val, task.TTL())
			}
			return val, nil
		}
	}

	// 3. L1/L2 均未命中，使用 Singleflight 合并并发请求
	return g.sf.do(ctx, key, func(ctx context.Context) (T, error) {
		// Double Check: 进入 Singleflight 后再次检查缓存，防止并发下的重复执行
		// (因为在等待 Singleflight 锁的过程中，可能前面的请求已经填好缓存了)
		if g.config.EnableRedisCache {
			if val, ok := g.l2.get(ctx, key); ok {
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

		// 填充缓存
		ttl := task.TTL()
		if g.config.EnableRedisCache {
			_ = g.l2.set(ctx, key, val, ttl)
		}
		if g.config.EnableLocalCache {
			g.l1.set(key, val, ttl)
		}

		return val, nil
	})
}
