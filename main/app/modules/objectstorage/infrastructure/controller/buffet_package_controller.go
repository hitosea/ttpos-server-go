package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// BuffetPackageQueryFunc 单个 BuffetPackage 查询函数类型
type BuffetPackageQueryFunc func(db *gorm.DB, uuid uint64) (*model.BuffetPackage, error)

// BuffetPackageBatchQueryFunc 批量 BuffetPackage 查询函数类型
type BuffetPackageBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.BuffetPackage, error)

// buffetPackageControllerInstance BuffetPackage 控制器单例实例
var (
	buffetPackageControllerInstance *CacheObjectController[*model.BuffetPackage]
	buffetPackageControllerOnce     sync.Once
	buffetPackageControllerMutex    sync.RWMutex
)

// InitBuffetPackageController 初始化 BuffetPackage 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 BuffetPackage 查询函数
//   - batchQueryFunc: 批量 BuffetPackage 查询函数
func InitBuffetPackageController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc BuffetPackageQueryFunc,
	batchQueryFunc BuffetPackageBatchQueryFunc,
) {
	InitCacheObjectController[*model.BuffetPackage](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.BuffetPackage, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.BuffetPackage, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.BuffetPackage) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeBuffetPackage,
		&buffetPackageControllerInstance,
		&buffetPackageControllerOnce,
		&buffetPackageControllerMutex,
	)
}

// GetBuffetPackageController 获取 BuffetPackage 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.BuffetPackage]: BuffetPackage 缓存对象控制器实例（保证非 nil）
func GetBuffetPackageController() *CacheObjectController[*model.BuffetPackage] {
	return GetCacheObjectController[*model.BuffetPackage](
		&buffetPackageControllerInstance,
		&buffetPackageControllerMutex,
		"BuffetPackage 缓存控制器未初始化，请先调用 InitBuffetPackageController",
	)
}
