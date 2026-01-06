package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// ProductBomCardQueryFunc 单个 ProductBomCard 查询函数类型
type ProductBomCardQueryFunc func(db *gorm.DB, uuid uint64) (*model.ProductBomCard, error)

// ProductBomCardBatchQueryFunc 批量 ProductBomCard 查询函数类型
type ProductBomCardBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.ProductBomCard, error)

// productBomCardControllerInstance ProductBomCard 控制器单例实例
var (
	productBomCardControllerInstance *CacheObjectController[*model.ProductBomCard]
	productBomCardControllerOnce     sync.Once
	productBomCardControllerMutex    sync.RWMutex
)

// InitProductBomCardController 初始化 ProductBomCard 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 ProductBomCard 查询函数
//   - batchQueryFunc: 批量 ProductBomCard 查询函数
func InitProductBomCardController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc ProductBomCardQueryFunc,
	batchQueryFunc ProductBomCardBatchQueryFunc,
) {
	InitCacheObjectController[*model.ProductBomCard](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductBomCard, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductBomCard, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.ProductBomCard) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeProductBomCard,
		&productBomCardControllerInstance,
		&productBomCardControllerOnce,
		&productBomCardControllerMutex,
	)
}

// GetProductBomCardController 获取 ProductBomCard 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.ProductBomCard]: ProductBomCard 缓存对象控制器实例（保证非 nil）
func GetProductBomCardController() *CacheObjectController[*model.ProductBomCard] {
	return GetCacheObjectController[*model.ProductBomCard](
		&productBomCardControllerInstance,
		&productBomCardControllerMutex,
		"ProductBomCard 缓存控制器未初始化，请先调用 InitProductBomCardController",
	)
}

