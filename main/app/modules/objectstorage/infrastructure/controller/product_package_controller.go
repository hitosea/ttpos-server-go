package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// ProductPackageQueryFunc 单个 ProductPackage 查询函数类型
type ProductPackageQueryFunc func(db *gorm.DB, uuid uint64) (*model.ProductPackage, error)

// ProductPackageBatchQueryFunc 批量 ProductPackage 查询函数类型
type ProductPackageBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.ProductPackage, error)

// productPackageControllerInstance ProductPackage 控制器单例实例
var (
	productPackageControllerInstance *CacheObjectController[*model.ProductPackage]
	productPackageControllerOnce     sync.Once
	productPackageControllerMutex    sync.RWMutex
)

// InitProductPackageController 初始化 ProductPackage 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 ProductPackage 查询函数
//   - batchQueryFunc: 批量 ProductPackage 查询函数
func InitProductPackageController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc ProductPackageQueryFunc,
	batchQueryFunc ProductPackageBatchQueryFunc,
) {
	InitCacheObjectController[*model.ProductPackage](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductPackage, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductPackage, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.ProductPackage) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeProductPackage,
		&productPackageControllerInstance,
		&productPackageControllerOnce,
		&productPackageControllerMutex,
	)
}

// GetProductPackageController 获取 ProductPackage 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.ProductPackage]: ProductPackage 缓存对象控制器实例（保证非 nil）
func GetProductPackageController() *CacheObjectController[*model.ProductPackage] {
	return GetCacheObjectController[*model.ProductPackage](
		&productPackageControllerInstance,
		&productPackageControllerMutex,
		"ProductPackage 缓存控制器未初始化，请先调用 InitProductPackageController",
	)
}
