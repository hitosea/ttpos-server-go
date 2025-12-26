package adapter

import (
	"context"
	"fmt"
	"reflect"
	"sync"
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
// 同时保留对底层 cache.Cache 的引用，用于 DEL 操作
type CacheGroupAdapter[T any] struct {
	group      cache.ICacheGroup[T]
	underlying cache.Cache // 底层缓存，用于 DEL 操作
	defaultTTL time.Duration
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

// cacheGroupSingletonManager 缓存组单例管理器
// 使用 sync.Map 存储 cacheGroup 创建函数，确保 L1 缓存可以跨请求共享
var (
	cacheGroupSingletons sync.Map // map[string]func() any
	cacheGroupMutex      sync.Mutex
)

// CacheLayerOption 缓存层选项
type CacheLayerOption struct {
	// SingletonKey 单例 key，如果为空则使用类型名称作为 key
	SingletonKey string

	// L1TTL L1 本地缓存的过期时间，如果为 0 则使用任务 TTL
	// 建议设置为任务 TTL 的 1/3 到 1/2
	L1TTL time.Duration

	// L2TTL L2 Redis 缓存的过期时间，如果为 0 则使用任务 TTL
	// 建议设置为任务 TTL 的 1.5 到 2 倍
	L2TTL time.Duration
}

// WithSingletonKey 设置单例 key
func WithSingletonKey(key string) func(*CacheLayerOption) {
	return func(opt *CacheLayerOption) {
		opt.SingletonKey = key
	}
}

// WithL1TTL 设置 L1 本地缓存的过期时间
// 阶梯式 TTL：L1 使用较短 TTL，减少内存占用
// 示例：WithL1TTL(1 * time.Minute) // L1 缓存 1 分钟
func WithL1TTL(ttl time.Duration) func(*CacheLayerOption) {
	return func(opt *CacheLayerOption) {
		opt.L1TTL = ttl
	}
}

// WithL2TTL 设置 L2 Redis 缓存的过期时间
// 阶梯式 TTL：L2 使用较长 TTL，保持缓存命中率
// 示例：WithL2TTL(10 * time.Minute) // L2 缓存 10 分钟
func WithL2TTL(ttl time.Duration) func(*CacheLayerOption) {
	return func(opt *CacheLayerOption) {
		opt.L2TTL = ttl
	}
}

// GetOrCreateCacheLayer 获取或创建缓存层（使用单例模式）
// 确保相同类型的 cacheGroup 实例是唯一的，从而让 L1 缓存可以跨请求共享
//
// 参数：
//   - groupConfig: 缓存组配置
//   - underlyingCache: 底层缓存实例（用于 DEL 操作）
//   - defaultTTL: 默认 TTL（用于负缓存等场景）
//   - opts: 可选参数
//   - WithSingletonKey: 设置单例 key
//   - WithL1TTL: 设置 L1 本地缓存 TTL（阶梯式 TTL）
//   - WithL2TTL: 设置 L2 Redis 缓存 TTL（阶梯式 TTL）
//
// 返回：
//   - CacheLayer[T]: 缓存层接口
//
// 示例：
//
//	// 使用类型名称作为 key（默认）
//	cacheLayer := GetOrCreateCacheLayer[MyType](groupConfig, cache.Global, 5*time.Minute)
//
//	// 使用自定义 key
//	cacheLayer := GetOrCreateCacheLayer[MyType](groupConfig, cache.Global, 5*time.Minute, WithSingletonKey("my-custom-key"))
//
//	// 使用阶梯式 TTL（推荐）
//	// L1 缓存 1 分钟，L2 缓存 5 分钟
//	// 优势：L1 过期后可从 L2 回填，减少内存占用同时保持缓存命中率
//	cacheLayer := GetOrCreateCacheLayer[MyType](
//		groupConfig,
//		cache.Global,
//		5*time.Minute,
//		WithL1TTL(1*time.Minute),  // L1 缓存 1 分钟
//		WithL2TTL(5*time.Minute),   // L2 缓存 5 分钟
//	)
func GetOrCreateCacheLayer[T any](groupConfig cache.GroupConfig, underlyingCache cache.Cache, defaultTTL time.Duration, opts ...func(*CacheLayerOption)) repository.CacheLayer[T] {
	// 解析选项
	option := &CacheLayerOption{}
	for _, opt := range opts {
		opt(option)
	}

	// 确定单例 key
	var key string
	if option.SingletonKey != "" {
		key = option.SingletonKey
	} else {
		// 使用类型名称作为 key
		key = reflect.TypeOf((*T)(nil)).Elem().String()
	}

	// 尝试从单例池中获取 cacheGroup 创建函数
	var group cache.ICacheGroup[T]
	if cached, ok := cacheGroupSingletons.Load(key); ok {
		// 使用类型断言获取创建函数
		if createFunc, ok := cached.(func() cache.ICacheGroup[T]); ok {
			group = createFunc()
		}
	}

	// 如果不存在，创建新的 cacheGroup 实例和创建函数
	if group == nil {
		cacheGroupMutex.Lock()
		// Double check
		if cached, ok := cacheGroupSingletons.Load(key); ok {
			if createFunc, ok := cached.(func() cache.ICacheGroup[T]); ok {
				group = createFunc()
			}
		}
		if group == nil {
			groupConfig.Name = groupConfig.Name + ":" + key
			// 应用选项中的 L1/L2 TTL 配置
			if option.L1TTL > 0 {
				groupConfig.L1TTL = option.L1TTL
			}
			if option.L2TTL > 0 {
				groupConfig.L2TTL = option.L2TTL
			}
			group = cache.NewCacheGroup[T](groupConfig)
			// 存储创建函数（返回同一个实例）
			createFunc := func() cache.ICacheGroup[T] {
				return group
			}
			cacheGroupSingletons.Store(key, createFunc)
		}
		cacheGroupMutex.Unlock()
	}

	// 使用已有的 cacheGroup 实例创建 adapter
	return NewCacheGroupAdapterWithGroup[T](group, underlyingCache, defaultTTL)
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
// 使用 group.Do 方式实现，通过闭包传入 value 值
// ICacheGroup 会自动处理 L1/L2 缓存的写入
func (a *CacheGroupAdapter[T]) SET(key string, value T, ttl time.Duration) error {
	task := &cacheTask[T]{
		key: key,
		query: func() (T, error) {
			// 通过闭包直接返回传入的 value
			return value, nil
		},
		ttl: ttl,
	}
	_, err := a.group.Do(context.Background(), task)
	return err
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
