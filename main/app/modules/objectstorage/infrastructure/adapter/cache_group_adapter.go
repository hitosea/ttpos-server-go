package adapter

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
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
	group          cache.ICacheGroup[T]
	underlying     cache.Cache // 底层缓存，用于 DEL 操作
	defaultTTL     time.Duration
	objectType     string                           // 对象类型（用于版本时间戳检查，必选）
	versionManager *persistence.CacheVersionManager // 缓存版本时间戳管理器
}

// NewCacheGroupAdapterWithGroup 使用已有的 cacheGroup 实例创建适配器
// 用于单例模式，复用 cacheGroup 实例以共享 L1 缓存
func NewCacheGroupAdapterWithGroup[T any](group cache.ICacheGroup[T], underlyingCache cache.Cache, defaultTTL time.Duration, objectType ...string) repository.CacheLayer[T] {
	adapter := &CacheGroupAdapter[T]{
		group:          group,
		underlying:     underlyingCache,
		defaultTTL:     defaultTTL,
		versionManager: persistence.NewCacheVersionManager(underlyingCache),
	}
	if len(objectType) > 0 {
		adapter.objectType = objectType[0]
	}
	return adapter
}

// cacheGroupSingletonManager 缓存组单例管理器
// 使用 sync.Map 直接存储 cacheGroup 实例，确保 L1 缓存可以跨请求共享
// 存储为 L1CacheClearable 接口类型，方便全局操作（如清除所有 L1 缓存）
var (
	cacheGroupSingletons sync.Map // map[string]cache.L1CacheClearable
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

	// EnableL1Cache 是否开启 L1 本地缓存，默认为 true
	EnableL1Cache bool
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

// WithEnableL1Cache 设置是否开启 L1 本地缓存
// 参数：
//   - enable: true 开启 L1 缓存（默认），false 禁用 L1 缓存
//
// 示例：WithEnableL1Cache(false) // 禁用 L1 缓存，只使用 L2 Redis 缓存
func WithEnableL1Cache(enable bool) func(*CacheLayerOption) {
	return func(opt *CacheLayerOption) {
		opt.EnableL1Cache = enable
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

	// 确定单例 objectType
	var objectType string
	if option.SingletonKey != "" {
		objectType = option.SingletonKey
	} else {
		// 使用类型名称作为 key
		objectType = reflect.TypeOf((*T)(nil)).Elem().String()
	}

	// 尝试从单例池中获取 cacheGroup 实例
	var group cache.ICacheGroup[T]
	if cached, ok := cacheGroupSingletons.Load(objectType); ok {
		// 使用类型断言获取实例
		if cachedGroup, ok := cached.(cache.ICacheGroup[T]); ok {
			group = cachedGroup
		}
	}

	// 如果不存在，创建新的 cacheGroup 实例
	if group == nil {
		cacheGroupMutex.Lock()
		// Double check
		if cached, ok := cacheGroupSingletons.Load(objectType); ok {
			if cachedGroup, ok := cached.(cache.ICacheGroup[T]); ok {
				group = cachedGroup
			}
		}
		if group == nil {
			groupConfig.Name = groupConfig.Name + ":" + objectType
			// 应用选项中的 L1/L2 TTL 配置
			if option.L1TTL > 0 {
				groupConfig.L1TTL = option.L1TTL
			}
			if option.L2TTL > 0 {
				groupConfig.L2TTL = option.L2TTL
			}
			// 设置统计回调函数（使用全局统计管理器）
			globalStats := GetGlobalStatsManager()
			groupConfig.HitStatsCallback = func(key string, hit bool, level string) {
				if hit {
					globalStats.RecordHit(key)
				} else {
					globalStats.RecordMiss(key)
				}
			}
			group = cache.NewCacheGroup[T](groupConfig)
			// 直接存储实例（作为 L1CacheClearable 接口类型）
			cacheGroupSingletons.Store(objectType, group)
		}
		cacheGroupMutex.Unlock()
	}

	// 使用已有的 cacheGroup 实例创建 adapter
	return NewCacheGroupAdapterWithGroup[T](group, underlyingCache, defaultTTL, objectType)
}

// GET 从缓存获取，未命中时调用 query 函数查询并写入缓存
// 如果配置了 objectType，会检查对象的 UpdateTime 字段是否小于最新版本时间戳
func (a *CacheGroupAdapter[T]) GET(key string, query func() (T, error), opts ...func(*repository.GetOption)) (T, error) {
	// 解析选项
	option := &repository.GetOption{
		SkipCache: false, // 默认不跳过缓存
	}
	for _, opt := range opts {
		opt(option)
	}

	// 构建任务
	task := &cacheTask[T]{
		key:   key,
		query: query,
		ttl:   a.defaultTTL,
	}

	// 如果配置了跳过缓存，直接执行查询
	if option.SkipCache {
		return a.group.Do(context.Background(), task, cache.WithSkipCache())
	}

	// 正常检查缓存（L1 -> L2 -> L3）
	result, err := a.group.Do(context.Background(), task)
	if err != nil {
		var zero T
		return zero, err
	}

	// 如果配置了 objectType，需要检查版本时间戳
	if a.objectType != "" {
		result = a.checkAndRefreshIfNeeded(key, result, query)
	}

	return result, nil
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
	_, err := a.group.Do(context.Background(), task, cache.WithSkipCache())
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
// 如果配置了 objectType，会检查对象的 UpdateTime 字段是否小于最新版本时间戳
func (a *CacheGroupAdapter[T]) BATCH_GET(keys []string, query func([]string) (map[string]T, error), opts ...func(*repository.GetOption)) (map[string]T, error) {
	if len(keys) == 0 {
		return make(map[string]T), nil
	}

	// 解析选项
	option := &repository.GetOption{
		SkipCache: false, // 默认不跳过缓存
	}
	for _, opt := range opts {
		opt(option)
	}

	// 为每个 key 单独调用 GET，利用 ICacheGroup 的 Singleflight 合并并发请求
	result := make(map[string]T)
	missCount := 0

	for _, key := range keys {
		querySingleKey := func() (T, error) {
			// 从批量查询结果中提取单个 key 的值
			batchResult, err := query([]string{key})
			if err != nil {
				var zero T
				return zero, err
			}
			return batchResult[key], nil
		}
		task := &cacheTask[T]{
			key:   key,
			query: querySingleKey,
			ttl:   a.defaultTTL,
		}

		// 根据选项决定是否跳过缓存
		var val T
		var err error
		if option.SkipCache {
			val, err = a.group.Do(context.Background(), task, cache.WithSkipCache())
		} else {
			val, err = a.group.Do(context.Background(), task)
		}

		if err != nil {
			missCount++
			continue // 单个失败不影响其他
		}

		// 如果配置了 objectType，需要检查版本时间戳
		if a.objectType != "" {
			val = a.checkAndRefreshIfNeeded(key, val, querySingleKey)
		}

		result[key] = val
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

// ClearAllL1Cache 清空所有 cacheGroup 单例的 L1 本地缓存
// 用于紧急情况下立即清除所有 L1 缓存，而不需要等待 TTL 过期
func ClearAllL1Cache() int {
	count := 0
	cacheGroupSingletons.Range(func(key, value interface{}) bool {
		// value 是 cache.L1CacheClearable 接口类型，直接调用 ClearL1 方法
		if clearable, ok := value.(cache.L1CacheClearable); ok {
			clearable.ClearL1()
			count++
		}
		return true
	})
	return count
}

// ClearL1CacheByPattern 根据匹配模式清空 L1 本地缓存
// pattern: 匹配模式，如 "ttpos4:*", "ttpos4:7709131161600000:*" 等
// 注意：由于 L1 缓存是内存缓存，无法直接按模式匹配，此方法会清空所有 L1 缓存
// 如果需要按模式清除，建议先清除 L2 缓存，然后清空所有 L1 缓存
func ClearL1CacheByPattern(pattern string) int {
	// L1 缓存是内存缓存，无法按模式匹配，直接清空所有
	// 这是合理的，因为 L1 缓存通常数据量不大，且清空后会自动从 L2 回填
	return ClearAllL1Cache()
}

// checkAndRefreshIfNeeded 检查版本时间戳，如果需要则刷新缓存
// 参数：
//   - key: 缓存 key
//   - result: 当前缓存结果
//   - query: 查询函数
//
// 返回：
//   - T: 更新后的结果（如果需要刷新则返回新查询的结果，否则返回原结果）
func (a *CacheGroupAdapter[T]) checkAndRefreshIfNeeded(key string, result T, query func() (T, error)) T {
	// 从 key 中提取 companyUuid
	ck, err := persistence.NewCacheKey(key)
	if err != nil {
		logger.Logger.Error("解析缓存 key 失败", zap.Error(err), zap.String("key", key))
		// 如果无法提取 companyUuid，直接返回原结果
		return result
	}
	companyUuid := ck.CompanyUuid
	objectType := ck.ObjectType
	objectUuid, _ := ck.GetObjectUuid() // 如果是product_list 对象，objectUuid 为 0. 这是特性,表示全局版本时间戳,要更新所有的product_list缓存. 其他对象的objectUuid不为0,表示具体对象的版本时间戳,要更新具体对象的缓存.

	// 获取最新版本时间戳
	versionTimestamp, versionExists := a.versionManager.GetCacheVersionTimestamp(companyUuid, objectType, objectUuid, persistence.WithCompanyWide(true))

	// 如果版本时间戳不存在（versionExists=false），说明版本时间戳已过期，
	// 此时应该认为缓存也已过期，需要重新查询并设置新的版本时间戳
	needRefresh := false
	if !versionExists {
		needRefresh = true
		if err := a.versionManager.UpdateCacheVersionTimestamp(companyUuid, objectType, objectUuid); err != nil {
			logger.Logger.Error("更新缓存版本时间戳失败", zap.Error(err))
		}
	} else if versionExists && versionTimestamp == 0 {
		// 如果版本时间戳存在但为0，说明版本时间戳已过期，
		needRefresh = true
		if err := a.versionManager.UpdateCacheVersionTimestamp(companyUuid, objectType, objectUuid); err != nil {
			logger.Logger.Error("更新缓存版本时间戳失败", zap.Error(err))
		}
	} else {
		// 如果存在版本时间戳，检查对象的 UpdateTime 字段
		updateTime := persistence.GetUpdateTime(result)
		// 版本时间戳和 UpdateTime 都是秒
		if updateTime < versionTimestamp {
			needRefresh = true
		}
	}

	if needRefresh {
		// 缓存已失效，重新查询
		newResult, err := query()
		if err != nil {
			// 查询失败，返回原结果
			return result
		}
		// 更新对象的 UpdateTime 字段为当前时间戳（秒）
		// 使用反射判断 T 是否是指针类型，如果是则直接传入，否则传入指针
		setUpdateTimeForGeneric(newResult, time.Now().Unix())
		// 更新缓存
		updateTask := &cacheTask[T]{
			key: key,
			query: func() (T, error) {
				return newResult, nil
			},
			ttl: a.defaultTTL,
		}
		_, _ = a.group.Do(context.Background(), updateTask, cache.WithSkipCache())
		return newResult
	}

	return result
}

// setUpdateTimeForGeneric 设置泛型值的 UpdateTime 字段
// 自动处理指针类型和值类型
func setUpdateTimeForGeneric[T any](val T, updateTime int64) {
	valReflect := reflect.ValueOf(val)
	// 如果 val 本身是指针类型（如 *model.Staff），需要解引用
	for valReflect.Kind() == reflect.Ptr {
		if valReflect.IsNil() {
			return
		}
		valReflect = valReflect.Elem()
	}
	// 现在 valReflect 应该是结构体类型
	if valReflect.Kind() == reflect.Struct {
		field := valReflect.FieldByName("UpdateTime")
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.SetInt(updateTime)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.SetUint(uint64(updateTime))
			}
		}
	}
}
