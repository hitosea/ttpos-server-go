package controller

import (
	goCtx "context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// BatchGetByUuidsOption BatchGetByUuids 方法的选项
type BatchGetByUuidsOption struct {
	// SkipCache 是否跳过缓存检查，直接执行查询
	SkipCache bool
}

// WithSkipCache 设置是否跳过缓存选项（用于 BatchGetByUuids）
func WithSkipCache() func(*BatchGetByUuidsOption) {
	return func(opt *BatchGetByUuidsOption) {
		opt.SkipCache = true
	}
}

// QueryFunc[T] 单个对象查询函数类型（泛型）
type QueryFunc[T any] func(db *gorm.DB, uuid uint64) (T, error)

// BatchQueryFunc[T] 批量对象查询函数类型（泛型）
type BatchQueryFunc[T any] func(db *gorm.DB, uuids []uint64) ([]T, error)

// GetUUIDFunc[T] 从对象中提取 UUID 的函数类型（泛型）
type GetUUIDFunc[T any] func(obj T) uint64

// CacheObjectController[T] 缓存对象控制器实现（泛型）
// 统一管理对象的缓存查询和更新，支持观察者模式更新缓存
// T: 对象类型，如 *model.Desk, *model.SaleBillSetting
type CacheObjectController[T any] struct {
	cacheLayer     repository.CacheLayer[T]
	queryFunc      QueryFunc[T]
	batchQueryFunc BatchQueryFunc[T]
	getUUIDFunc    GetUUIDFunc[T] // 用于从对象中提取 UUID（用于批量查询结果转换）
	objectType     string         // 对象类型（用于构建缓存 key）
	ttl            time.Duration
}

// GetByUuid 根据 UUID 获取对象（带缓存）
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - db: 数据库连接
//   - uuid: 对象 UUID
//
// 返回：
//   - T: 对象
//   - error: 错误信息
func (c *CacheObjectController[T]) GetByUuid(ctx goCtx.Context, db *gorm.DB, uuid uint64) (T, error) {
	var zero T
	if uuid == 0 {
		return zero, fmt.Errorf("对象 UUID 不能为0")
	}

	// 构建缓存 key
	key := persistence.BuildKey(ctx, c.objectType, uuid)

	// 正常流程：按 L1 -> L2 -> L3 的顺序查找
	result, err := c.cacheLayer.GET(key, func() (T, error) {
		return c.queryFunc(db, uuid)
	})

	if err != nil {
		return zero, fmt.Errorf("GetByUuid: %v", err)
	}

	return result, nil
}

// BatchGetByUuids 批量根据 UUID 列表获取对象（带缓存）
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - db: 数据库连接
//   - uuids: 对象 UUID 列表
//   - opts: 选项函数（可选），如 WithSkipCache() 跳过缓存直接查询
//
// 返回：
//   - map[uint64]T: UUID 到对象的映射
//   - error: 错误信息
func (c *CacheObjectController[T]) BatchGetByUuids(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*BatchGetByUuidsOption)) (map[uint64]T, error) {
	if len(uuids) == 0 {
		return make(map[uint64]T), nil
	}

	// 解析选项
	option := &BatchGetByUuidsOption{
		SkipCache: false, // 默认不跳过缓存
	}
	for _, opt := range opts {
		opt(option)
	}

	// 过滤掉 0 值
	validUuids := make([]uint64, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			validUuids = append(validUuids, uuid)
		}
	}
	if len(validUuids) == 0 {
		return make(map[uint64]T), nil
	}

	// 构建批量查询的 keys
	keys := make([]string, 0, len(validUuids))
	for _, uuid := range validUuids {
		keys = append(keys, persistence.BuildKey(ctx, c.objectType, uuid))
	}

	// 批量从缓存获取
	var batchResult map[string]T
	var err error
	if option.SkipCache {
		// 跳过缓存，直接查询数据库（仍会写入缓存）
		batchResult, err = c.cacheLayer.BATCH_GET(keys, func([]string) (map[string]T, error) {
			// 批量查询数据库
			objects, err := c.batchQueryFunc(db, validUuids)
			if err != nil {
				return nil, err
			}

			// 转换为 map[string]T
			result := make(map[string]T)
			for _, obj := range objects {
				uuid := c.getUUIDFunc(obj)
				if uuid > 0 {
					key := persistence.BuildKey(ctx, c.objectType, uuid)
					result[key] = obj
				}
			}
			return result, nil
		}, repository.WithSkipCache())
	} else {
		// 正常流程：按 L1 -> L2 -> L3 的顺序查找
		batchResult, err = c.cacheLayer.BATCH_GET(keys, func([]string) (map[string]T, error) {
			// 批量查询数据库
			objects, err := c.batchQueryFunc(db, validUuids)
			if err != nil {
				return nil, err
			}

			// 转换为 map[string]T
			result := make(map[string]T)
			for _, obj := range objects {
				uuid := c.getUUIDFunc(obj)
				if uuid > 0 {
					key := persistence.BuildKey(ctx, c.objectType, uuid)
					result[key] = obj
				}
			}
			return result, nil
		})
	}

	if err != nil {
		return nil, fmt.Errorf("BatchGetByUuids: %v", err)
	}

	// 转换为 map[uint64]T
	result := make(map[uint64]T)
	for key, obj := range batchResult {
		// 从 key 中提取 UUID
		if uuid, err := extractUUIDFromKey(key); err == nil {
			result[uuid] = obj
		}
	}

	return result, nil
}

// Update 更新对象的缓存（用于观察者模式）
// 当对象发生变化时，调用此方法更新缓存
// 通过重新调用 BatchGetByUuids 并跳过缓存，从数据库重新查询并更新缓存
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - db: 数据库连接
//   - uuids: 对象 UUID 列表
//
// 返回：
//   - error: 错误信息
func (c *CacheObjectController[T]) Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}

	// 重新调用 BatchGetByUuids 并跳过缓存，从数据库重新查询并自动更新缓存
	_, err := c.BatchGetByUuids(ctx, db, uuids, WithSkipCache())
	if err != nil {
		return fmt.Errorf("Update: %v", err)
	}

	return nil
}

// extractUUIDFromKey 从缓存 key 中提取 UUID
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
func extractUUIDFromKey(key string) (uint64, error) {
	parts := strings.Split(key, ":")
	if len(parts) < 4 {
		return 0, fmt.Errorf("无效的缓存 key 格式: %s", key)
	}
	uuidStr := parts[3]
	uuid, err := strconv.ParseUint(uuidStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析 UUID: %v", err)
	}
	return uuid, nil
}

// InitCacheObjectController[T] 初始化缓存对象控制器（泛型，单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个对象查询函数
//   - batchQueryFunc: 批量对象查询函数
//   - getUUIDFunc: 从对象中提取 UUID 的函数
//   - objectType: 对象类型（用于构建缓存 key）
//   - instance: 单例实例指针（用于存储控制器实例）
//   - once: sync.Once 实例（确保只初始化一次）
//   - mutex: sync.RWMutex 实例（用于线程安全访问）
func InitCacheObjectController[T any](
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc QueryFunc[T],
	batchQueryFunc BatchQueryFunc[T],
	getUUIDFunc GetUUIDFunc[T],
	objectType string,
	instance **CacheObjectController[T],
	once *sync.Once,
	mutex *sync.RWMutex,
) {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	once.Do(func() {
		// 创建缓存层
		cacheLayer := adapter.GetOrderObjectCache[T](underlyingCache, ttl)

		*instance = &CacheObjectController[T]{
			cacheLayer:     cacheLayer,
			queryFunc:      queryFunc,
			batchQueryFunc: batchQueryFunc,
			getUUIDFunc:    getUUIDFunc,
			objectType:     objectType,
			ttl:            ttl,
		}
	})
}

// GetCacheObjectController[T] 获取缓存对象控制器单例实例（泛型）
// 保证返回非 nil 值，如果未初始化会 panic
// 参数：
//   - instance: 单例实例指针
//   - mutex: sync.RWMutex 实例（用于线程安全访问）
//   - panicMsg: panic 消息
//
// 返回：
//   - *CacheObjectController[T]: 缓存对象控制器实例（保证非 nil）
func GetCacheObjectController[T any](
	instance **CacheObjectController[T],
	mutex *sync.RWMutex,
	panicMsg string,
) *CacheObjectController[T] {
	mutex.RLock()
	defer mutex.RUnlock()
	return *instance
}
