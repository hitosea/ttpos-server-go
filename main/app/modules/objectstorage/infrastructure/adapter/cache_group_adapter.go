package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// cacheTask 实现 cache.Task[T] 接口
type cacheTask[T any] struct {
	key   string
	query func() (T, error)
	ttl   time.Duration
}

func (t *cacheTask[T]) Hash() string {
	return t.key
}

func (t *cacheTask[T]) Exec(ctx context.Context) (T, error) {
	return t.query()
}

func (t *cacheTask[T]) TTL() time.Duration {
	return t.ttl
}

// CacheGroupAdapter 使用 ICacheGroup 实现的缓存适配器（泛型版本）
// 同时保留对底层 cache.Cache 的引用，用于 SET/DEL 操作
type CacheGroupAdapter[T any] struct {
	group      cache.ICacheGroup[T]
	underlying cache.Cache // 底层缓存，用于 SET/DEL 操作
	defaultTTL time.Duration
}

// NewCacheGroupAdapter 创建基于 ICacheGroup 的缓存适配器
func NewCacheGroupAdapter[T any](groupConfig cache.GroupConfig, underlyingCache cache.Cache, defaultTTL time.Duration) repository.CacheLayer[T] {
	return &CacheGroupAdapter[T]{
		group:      cache.NewCacheGroup[T](groupConfig),
		underlying: underlyingCache,
		defaultTTL: defaultTTL,
	}
}

// NewCacheGroupAdapterWithGroup 使用已有的 cacheGroup 实例创建适配器
// 用于单例模式，复用 cacheGroup 实例以共享 L1 缓存
func NewCacheGroupAdapterWithGroup[T any](group cache.ICacheGroup[T], underlyingCache cache.Cache, defaultTTL time.Duration) repository.CacheLayer[T] {
	return &CacheGroupAdapter[T]{
		group:      group,
		underlying: underlyingCache,
		defaultTTL: defaultTTL,
	}
}

// GET 从缓存获取，未命中时调用 query 函数查询并写入缓存
func (a *CacheGroupAdapter[T]) GET(key string, query func() (T, error)) (T, error) {
	task := &cacheTask[T]{
		key:   key,
		query: query,
		ttl:   a.defaultTTL,
	}
	return a.group.Do(context.Background(), task)
}

// SET 设置缓存
// 使用底层缓存直接写入，同时 ICacheGroup 会在下次 GET 时自动使用
// 直接存储 []byte，不转换为 string
func (a *CacheGroupAdapter[T]) SET(key string, value T, ttl time.Duration) error {
	if a.underlying == nil {
		return fmt.Errorf("底层缓存未设置，无法执行 SET 操作")
	}
	// 序列化为 JSON 字节数组，直接存储 []byte
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	// 直接传递 []byte，不转换为 string
	return a.underlying.Set(key, jsonData, ttl)
}

// DEL 删除缓存
// 使用底层缓存直接删除
func (a *CacheGroupAdapter[T]) DEL(keys ...string) error {
	if a.underlying == nil {
		return fmt.Errorf("底层缓存未设置，无法执行 DEL 操作")
	}
	a.underlying.Del(keys...)
	return nil
}

// BATCH_GET 批量获取缓存
func (a *CacheGroupAdapter[T]) BATCH_GET(keys []string, query func([]string) (map[string]T, error)) (map[string]T, error) {
	if len(keys) == 0 {
		return make(map[string]T), nil
	}

	// 为每个 key 单独调用 GET，利用 ICacheGroup 的 Singleflight 合并并发请求
	result := make(map[string]T)
	missCount := 0

	for _, key := range keys {
		task := &cacheTask[T]{
			key: key,
			query: func() (T, error) {
				// 从批量查询结果中提取单个 key 的值
				batchResult, err := query([]string{key})
				if err != nil {
					var zero T
					return zero, err
				}
				return batchResult[key], nil
			},
			ttl: a.defaultTTL,
		}
		val, err := a.group.Do(context.Background(), task)
		if err != nil {
			missCount++
			continue // 单个失败不影响其他
		}
		result[key] = val
		// 注意：这里无法准确判断是命中还是未命中，因为 ICacheGroup 内部处理了缓存逻辑
		// 实际的命中/未命中日志会在 cacheGroup.Do 中记录
	}

	// 记录批量查询的总体情况
	logger.Logger.Debug("批量缓存查询完成",
		zap.Int("total", len(keys)),
		zap.Int("success", len(result)),
		zap.Int("failed", missCount),
		zap.String("type", "BATCH_GET"),
	)

	return result, nil
}

// SCAN 扫描匹配模式的 key（用于批量失效）
func (a *CacheGroupAdapter[T]) SCAN(ctx context.Context, pattern string) ([]string, error) {
	// ICacheGroup 不支持 SCAN，需要扩展接口或使用底层 Redis 客户端
	// 这里返回 nil 表示不支持，调用方会使用其他方式处理
	return nil, nil
}
