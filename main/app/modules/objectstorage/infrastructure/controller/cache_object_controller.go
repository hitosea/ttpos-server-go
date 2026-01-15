package controller

import (
	goCtx "context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// ICacheObjectController Desk 对象缓存控制器接口（向后兼容）
// 统一管理对象的缓存查询和更新，支持观察者模式更新缓存
type ICacheObjectController interface {
	// GetByUuid 根据 UUID 获取对象（带缓存）
	GetByUuid(ctx goCtx.Context, db *gorm.DB, uuid uint64) (*model.Desk, error)

	// BatchGetByUuids 批量根据 UUID 列表获取对象（带缓存）
	BatchGetByUuids(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*BatchGetByUuidsOption)) (map[uint64]*model.Desk, error)

	// Update 更新对象的缓存（用于观察者模式）
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuids: 对象 UUID 列表
	//   - opts: 选项函数（可选），如 WithUpdateValue() 直接提供值更新缓存
	Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*UpdateOption)) error

	// Invalidate 使对象缓存失效
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - uuid: 对象的 UUID（如果为 0，则使所有该类型对象的缓存失效）
	//
	// 返回：
	//   - error: 错误信息
	Invalidate(ctx goCtx.Context, uuid uint64) error
}

// BatchGetByUuidsOption BatchGetByUuids 方法的选项
type BatchGetByUuidsOption struct {
	// SkipCache 是否跳过缓存检查，直接执行查询
	SkipCache bool

	// Value 当 SkipCache=true 时，如果提供了 Value（非空），则直接使用此值更新缓存，不执行查询
	// 如果为 nil，则正常执行查询
	// map[uint64]T: UUID 到对象的映射
	Value map[uint64]interface{}
}

// WithSkipCache 设置是否跳过缓存选项（用于 BatchGetByUuids）
// 参数：
//   - val: 可选的值映射（map[uint64]interface{}），如果提供且非空，则直接使用此值更新缓存，不执行查询
//     如果为 nil，则正常执行查询
func WithSkipCache(val ...map[uint64]interface{}) func(*BatchGetByUuidsOption) {
	return func(opt *BatchGetByUuidsOption) {
		opt.SkipCache = true
		if len(val) > 0 && val[0] != nil && len(val[0]) > 0 {
			opt.Value = val[0]
		}
	}
}

// UpdateOption Update 方法的选项
type UpdateOption struct {
	// Value 如果提供了 Value（非空），则直接使用此值更新缓存，不执行查询
	// 如果为 nil，则正常执行查询
	// map[uint64]interface{}: UUID 到对象的映射
	Value map[uint64]interface{}
}

// WithUpdateValue 设置直接更新的值（用于 Update）
// 参数：
//   - val: 值映射（map[uint64]interface{}），如果提供且非空，则直接使用此值更新缓存，不执行查询
//     如果为 nil，则正常执行查询
func WithUpdateValue(val map[uint64]interface{}) func(*UpdateOption) {
	return func(opt *UpdateOption) {
		opt.Value = val
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
	cacheLayer      repository.CacheLayer[T]
	queryFunc       QueryFunc[T]
	batchQueryFunc  BatchQueryFunc[T]
	getUUIDFunc     GetUUIDFunc[T] // 用于从对象中提取 UUID（用于批量查询结果转换）
	objectType      string         // 对象类型（用于构建缓存 key）
	ttl             time.Duration
	underlyingCache cache.Cache // 底层缓存实例（用于版本时间戳管理）
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
		// 如果提供了 Value，直接使用这些值更新缓存，不执行查询
		if option.Value != nil && len(option.Value) > 0 {
			// 直接使用提供的值更新缓存
			result := make(map[string]T)
			for _, uuid := range validUuids {
				if val, ok := option.Value[uuid]; ok {
					// 类型断言为 T
					if typedVal, ok := val.(T); ok {
						key := persistence.BuildKey(ctx, c.objectType, uuid)
						// 直接设置缓存
						if err := c.cacheLayer.SET(key, typedVal, c.ttl); err != nil {
							return nil, fmt.Errorf("直接更新缓存失败 (uuid=%d): %v", uuid, err)
						}
						result[key] = typedVal
					}
				}
			}
			batchResult = result
		} else {
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
		}
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
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - db: 数据库连接
//   - uuids: 对象 UUID 列表
//   - opts: 选项函数（可选），如 WithUpdateValue() 直接提供值更新缓存
//
// 返回：
//   - error: 错误信息
func (c *CacheObjectController[T]) Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*UpdateOption)) error {
	if len(uuids) == 0 {
		return nil
	}

	// 解析选项
	option := &UpdateOption{
		Value: nil, // 默认不提供值，正常执行查询
	}
	for _, opt := range opts {
		opt(option)
	}

	// 如果提供了 Value，直接使用这些值更新缓存
	if option.Value != nil && len(option.Value) > 0 {
		_, err := c.BatchGetByUuids(ctx, db, uuids, WithSkipCache(option.Value))
		if err != nil {
			return fmt.Errorf("Update: %v", err)
		}
	} else {
		// 重新调用 BatchGetByUuids 并跳过缓存，从数据库重新查询并自动更新缓存
		_, err := c.BatchGetByUuids(ctx, db, uuids, WithSkipCache())
		if err != nil {
			return fmt.Errorf("Update: %v", err)
		}
	}

	return nil
}

// Invalidate 使对象缓存失效
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - uuid: 对象的 UUID（如果为 0，则使所有该类型对象的缓存失效）
//
// 返回：
//   - error: 错误信息
func (c *CacheObjectController[T]) Invalidate(ctx goCtx.Context, uuid uint64) error {
	// 从上下文提取 companyUuid
	cctx := ctx.(context.Context)
	companyUuid := cctx.GetCompanyUuid()

	// 创建缓存版本时间戳管理器
	versionManager := persistence.NewCacheVersionManager(c.underlyingCache)

	// 更新缓存版本时间戳（在二级缓存中记录对象的最新版本时间戳）
	// Key 格式：{cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
	// object_uuid 为 0 表示全局版本时间戳，要更新所有的该类型对象缓存
	// object_uuid 不为 0 表示具体对象的版本时间戳，要更新具体对象的缓存
	// 版本时间戳的过期时间设置为 L2TTL（5分钟），与最长有效的缓存时间一致
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	if err := versionManager.UpdateCacheVersionTimestamp(companyUuid, c.objectType, uuid); err != nil {
		return fmt.Errorf("更新缓存版本时间戳失败: %w", err)
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
//   - ttl: 缓存 TTL（默认 5 分钟）
//   - queryFunc: 单个对象查询函数
//   - batchQueryFunc: 批量对象查询函数
//   - getUUIDFunc: 从对象中提取 UUID 的函数
//   - objectType: 对象类型（用于构建缓存 key）
//   - instance: 单例实例指针（用于存储控制器实例）
//   - once: sync.Once 实例（确保只初始化一次）
//   - opts: 可选参数，如 adapter.WithEnableL1Cache(false) 禁用 L1 缓存，adapter.WithEnableL2Cache(false) 禁用 L2 缓存
func InitCacheObjectController[T any](
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc QueryFunc[T],
	batchQueryFunc BatchQueryFunc[T],
	getUUIDFunc GetUUIDFunc[T],
	objectType string,
	opts ...func(*adapter.CacheLayerOption),
) *CacheObjectController[T] {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	// 创建缓存层
	cacheLayer := adapter.GetOrderObjectCache[T](underlyingCache, ttl, opts...)

	return &CacheObjectController[T]{
		cacheLayer:      cacheLayer,
		queryFunc:       queryFunc,
		batchQueryFunc:  batchQueryFunc,
		getUUIDFunc:     getUUIDFunc,
		objectType:      objectType,
		ttl:             ttl,
		underlyingCache: underlyingCache,
	}
}
