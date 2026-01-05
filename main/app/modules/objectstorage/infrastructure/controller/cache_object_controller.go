package controller

import (
	goCtx "context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// DeskQueryFunc 单个 Desk 查询函数类型
type DeskQueryFunc func(db *gorm.DB, uuid uint64) (*model.Desk, error)

// DeskBatchQueryFunc 批量 Desk 查询函数类型
type DeskBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.Desk, error)

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

// ICacheObjectController 缓存对象控制器接口
// 统一管理对象的缓存查询和更新，支持观察者模式更新缓存
// 可用于 Desk、ProductPackage 等各种对象类型
type ICacheObjectController interface {
	// GetByUuid 根据 UUID 获取对象（带缓存）
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuid: 对象 UUID
	// 返回：
	//   - *model.Desk: 对象指针
	//   - error: 错误信息
	GetByUuid(ctx goCtx.Context, db *gorm.DB, uuid uint64) (*model.Desk, error)

	// BatchGetByUuids 批量根据 UUID 列表获取对象（带缓存）
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuids: 对象 UUID 列表
	//   - opts: 选项函数（可选），如 WithSkipCache() 跳过缓存直接查询
	// 返回：
	//   - map[uint64]*model.Desk: UUID 到对象的映射
	//   - error: 错误信息
	BatchGetByUuids(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*BatchGetByUuidsOption)) (map[uint64]*model.Desk, error)

	// Update 更新对象的缓存（用于观察者模式）
	// 当对象发生变化时，调用此方法更新缓存
	// 通过重新调用 BatchGetByUuids 并跳过缓存，从数据库重新查询并更新缓存
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuids: 对象 UUID 列表
	// 返回：
	//   - error: 错误信息
	Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64) error
}

// CacheObjectController 缓存对象控制器实现
// 统一管理对象的缓存查询和更新，支持观察者模式更新缓存
// 当前实现用于 Desk 对象，未来可扩展支持其他对象类型
type CacheObjectController struct {
	cacheLayer     repository.CacheLayer[*model.Desk]
	queryFunc      DeskQueryFunc
	batchQueryFunc DeskBatchQueryFunc
	objectType     string // 对象类型（用于构建缓存 key）
	ttl            time.Duration
}

// 确保 CacheObjectController 实现了 ICacheObjectController 接口
var _ ICacheObjectController = (*CacheObjectController)(nil)

// cacheObjectControllerInstance 缓存对象控制器单例实例
var (
	cacheObjectControllerInstance ICacheObjectController
	cacheObjectControllerOnce     sync.Once
	cacheObjectControllerMutex    sync.RWMutex
)

// InitDeskController 初始化 Desk 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 Desk 查询函数
//   - batchQueryFunc: 批量 Desk 查询函数
func InitDeskController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc DeskQueryFunc,
	batchQueryFunc DeskBatchQueryFunc,
) {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	cacheObjectControllerOnce.Do(func() {
		// 创建缓存层
		cacheLayer := adapter.GetOrderObjectCache[*model.Desk](underlyingCache, ttl)

		cacheObjectControllerInstance = &CacheObjectController{
			cacheLayer:     cacheLayer,
			queryFunc:      queryFunc,
			batchQueryFunc: batchQueryFunc,
			objectType:     persistence.ObjectTypeDesk,
			ttl:            ttl,
		}
	})
}

// GetDeskController 获取 Desk 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - ICacheObjectController: 缓存对象控制器实例（保证非 nil）
func GetDeskController() ICacheObjectController {
	cacheObjectControllerMutex.RLock()
	defer cacheObjectControllerMutex.RUnlock()
	if cacheObjectControllerInstance == nil {
		panic("Desk 缓存控制器未初始化，请先调用 InitDeskController")
	}
	return cacheObjectControllerInstance
}

// GetByUuid 根据 UUID 获取对象（带缓存）
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - db: 数据库连接
//   - uuid: 对象 UUID
//
// 返回：
//   - *model.Desk: 对象指针
//   - error: 错误信息
func (c *CacheObjectController) GetByUuid(ctx goCtx.Context, db *gorm.DB, uuid uint64) (*model.Desk, error) {
	if uuid == 0 {
		return nil, fmt.Errorf("对象 UUID 不能为0")
	}

	// 构建缓存 key
	key := persistence.BuildKey(ctx, c.objectType, uuid)

	// 正常流程：按 L1 -> L2 -> L3 的顺序查找
	result, err := c.cacheLayer.GET(key, func() (*model.Desk, error) {
		return c.queryFunc(db, uuid)
	})

	if err != nil {
		return nil, fmt.Errorf("GetByUuid: %v", err)
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
//   - map[uint64]*model.Desk: UUID 到对象的映射
//   - error: 错误信息
func (c *CacheObjectController) BatchGetByUuids(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*BatchGetByUuidsOption)) (map[uint64]*model.Desk, error) {
	if len(uuids) == 0 {
		return make(map[uint64]*model.Desk), nil
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
		return make(map[uint64]*model.Desk), nil
	}

	// 构建批量查询的 keys
	keys := make([]string, 0, len(validUuids))
	for _, uuid := range validUuids {
		keys = append(keys, persistence.BuildKey(ctx, c.objectType, uuid))
	}

	// 批量从缓存获取
	var batchResult map[string]*model.Desk
	var err error
	if option.SkipCache {
		// 跳过缓存，直接查询数据库（仍会写入缓存）
		batchResult, err = c.cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.Desk, error) {
			// 批量查询数据库
			desks, err := c.batchQueryFunc(db, validUuids)
			if err != nil {
				return nil, err
			}

			// 转换为 map[string]*model.Desk
			result := make(map[string]*model.Desk)
			for _, desk := range desks {
				if desk != nil {
					key := persistence.BuildKey(ctx, c.objectType, desk.Uuid)
					result[key] = desk
				}
			}
			return result, nil
		}, repository.WithSkipCache())
	} else {
		// 正常流程：按 L1 -> L2 -> L3 的顺序查找
		batchResult, err = c.cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.Desk, error) {
			// 批量查询数据库
			desks, err := c.batchQueryFunc(db, validUuids)
			if err != nil {
				return nil, err
			}

			// 转换为 map[string]*model.Desk
			result := make(map[string]*model.Desk)
			for _, desk := range desks {
				if desk != nil {
					key := persistence.BuildKey(ctx, c.objectType, desk.Uuid)
					result[key] = desk
				}
			}
			return result, nil
		})
	}

	if err != nil {
		return nil, fmt.Errorf("BatchGetByUuids: %v", err)
	}

	// 转换为 map[uint64]*model.Desk
	result := make(map[uint64]*model.Desk)
	for key, desk := range batchResult {
		if desk != nil {
			// 从 key 中提取 UUID
			if uuid, err := extractUUIDFromKey(key); err == nil {
				result[uuid] = desk
			}
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
func (c *CacheObjectController) Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64) error {
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

