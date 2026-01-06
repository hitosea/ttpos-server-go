package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// NationalityQueryFunc 单个 Nationality 查询函数类型
type NationalityQueryFunc func(db *gorm.DB, uuid uint64) (*model.Nationality, error)

// NationalityBatchQueryFunc 批量 Nationality 查询函数类型
type NationalityBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.Nationality, error)

// nationalityControllerInstance Nationality 控制器单例实例
var (
	nationalityControllerInstance *CacheObjectController[*model.Nationality]
	nationalityControllerOnce     sync.Once
	nationalityControllerMutex    sync.RWMutex
)

// InitNationalityController 初始化 Nationality 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 Nationality 查询函数
//   - batchQueryFunc: 批量 Nationality 查询函数
func InitNationalityController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc NationalityQueryFunc,
	batchQueryFunc NationalityBatchQueryFunc,
) {
	InitCacheObjectController[*model.Nationality](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.Nationality, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.Nationality, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.Nationality) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeNationality,
		&nationalityControllerInstance,
		&nationalityControllerOnce,
		&nationalityControllerMutex,
	)
}

// GetNationalityController 获取 Nationality 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.Nationality]: Nationality 缓存对象控制器实例（保证非 nil）
func GetNationalityController() *CacheObjectController[*model.Nationality] {
	return GetCacheObjectController[*model.Nationality](
		&nationalityControllerInstance,
		&nationalityControllerMutex,
		"Nationality 缓存控制器未初始化，请先调用 InitNationalityController",
	)
}
