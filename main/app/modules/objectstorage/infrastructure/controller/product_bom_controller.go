package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// ProductBomQueryFunc 单个 ProductBom 查询函数类型
type ProductBomQueryFunc func(db *gorm.DB, uuid uint64) (*model.ProductBom, error)

// ProductBomBatchQueryFunc 批量 ProductBom 查询函数类型
type ProductBomBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.ProductBom, error)

// productBomControllerInstance ProductBom 控制器单例实例
var (
	productBomControllerInstance *CacheObjectController[*model.ProductBom]
	productBomControllerOnce     sync.Once
	productBomControllerMutex    sync.RWMutex
)

// InitProductBomController 初始化 ProductBom 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 ProductBom 查询函数
//   - batchQueryFunc: 批量 ProductBom 查询函数
func InitProductBomController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc ProductBomQueryFunc,
	batchQueryFunc ProductBomBatchQueryFunc,
) {
	InitCacheObjectController[*model.ProductBom](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductBom, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductBom, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.ProductBom) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeProductBom,
		&productBomControllerInstance,
		&productBomControllerOnce,
		&productBomControllerMutex,
	)
}

// GetProductBomController 获取 ProductBom 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.ProductBom]: ProductBom 缓存对象控制器实例（保证非 nil）
func GetProductBomController() *CacheObjectController[*model.ProductBom] {
	return GetCacheObjectController[*model.ProductBom](
		&productBomControllerInstance,
		&productBomControllerMutex,
		"ProductBom 缓存控制器未初始化，请先调用 InitProductBomController",
	)
}
