package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// OrderSourceQueryFunc 单个 OrderSource 查询函数类型
type OrderSourceQueryFunc func(db *gorm.DB, uuid uint64) (*model.OrderSource, error)

// OrderSourceBatchQueryFunc 批量 OrderSource 查询函数类型
type OrderSourceBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.OrderSource, error)

// orderSourceControllerInstance OrderSource 控制器单例实例
var (
	orderSourceControllerInstance *CacheObjectController[*model.OrderSource]
	orderSourceControllerOnce     sync.Once
	orderSourceControllerMutex    sync.RWMutex
)

// InitOrderSourceController 初始化 OrderSource 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 OrderSource 查询函数
//   - batchQueryFunc: 批量 OrderSource 查询函数
func InitOrderSourceController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc OrderSourceQueryFunc,
	batchQueryFunc OrderSourceBatchQueryFunc,
) {
	InitCacheObjectController[*model.OrderSource](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.OrderSource, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.OrderSource, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.OrderSource) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeOrderSource,
		&orderSourceControllerInstance,
		&orderSourceControllerOnce,
		&orderSourceControllerMutex,
	)
}

// GetOrderSourceController 获取 OrderSource 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.OrderSource]: OrderSource 缓存对象控制器实例（保证非 nil）
func GetOrderSourceController() *CacheObjectController[*model.OrderSource] {
	return GetCacheObjectController[*model.OrderSource](
		&orderSourceControllerInstance,
		&orderSourceControllerMutex,
		"OrderSource 缓存控制器未初始化，请先调用 InitOrderSourceController",
	)
}

