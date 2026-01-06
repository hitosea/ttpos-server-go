package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// ProductAttributeQueryFunc 单个 ProductAttribute 查询函数类型
type ProductAttributeQueryFunc func(db *gorm.DB, uuid uint64) (*model.ProductAttribute, error)

// ProductAttributeBatchQueryFunc 批量 ProductAttribute 查询函数类型
type ProductAttributeBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.ProductAttribute, error)

// productAttributeControllerInstance ProductAttribute 控制器单例实例
var (
	productAttributeControllerInstance *CacheObjectController[*model.ProductAttribute]
	productAttributeControllerOnce     sync.Once
	productAttributeControllerMutex    sync.RWMutex
)

// InitProductAttributeController 初始化 ProductAttribute 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 ProductAttribute 查询函数
//   - batchQueryFunc: 批量 ProductAttribute 查询函数
func InitProductAttributeController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc ProductAttributeQueryFunc,
	batchQueryFunc ProductAttributeBatchQueryFunc,
) {
	InitCacheObjectController[*model.ProductAttribute](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductAttribute, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductAttribute, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.ProductAttribute) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeProductAttribute,
		&productAttributeControllerInstance,
		&productAttributeControllerOnce,
		&productAttributeControllerMutex,
	)
}

// GetProductAttributeController 获取 ProductAttribute 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.ProductAttribute]: ProductAttribute 缓存对象控制器实例（保证非 nil）
func GetProductAttributeController() *CacheObjectController[*model.ProductAttribute] {
	return GetCacheObjectController[*model.ProductAttribute](
		&productAttributeControllerInstance,
		&productAttributeControllerMutex,
		"ProductAttribute 缓存控制器未初始化，请先调用 InitProductAttributeController",
	)
}

