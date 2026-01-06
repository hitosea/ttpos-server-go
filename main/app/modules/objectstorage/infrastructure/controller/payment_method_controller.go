package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// PaymentMethodQueryFunc 单个 PaymentMethod 查询函数类型
type PaymentMethodQueryFunc func(db *gorm.DB, uuid uint64) (*model.PaymentMethod, error)

// PaymentMethodBatchQueryFunc 批量 PaymentMethod 查询函数类型
type PaymentMethodBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.PaymentMethod, error)

// paymentMethodControllerInstance PaymentMethod 控制器单例实例
var (
	paymentMethodControllerInstance *CacheObjectController[*model.PaymentMethod]
	paymentMethodControllerOnce     sync.Once
	paymentMethodControllerMutex    sync.RWMutex
)

// InitPaymentMethodController 初始化 PaymentMethod 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 PaymentMethod 查询函数
//   - batchQueryFunc: 批量 PaymentMethod 查询函数
func InitPaymentMethodController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc PaymentMethodQueryFunc,
	batchQueryFunc PaymentMethodBatchQueryFunc,
) {
	InitCacheObjectController[*model.PaymentMethod](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.PaymentMethod, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.PaymentMethod, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.PaymentMethod) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypePaymentMethod,
		&paymentMethodControllerInstance,
		&paymentMethodControllerOnce,
		&paymentMethodControllerMutex,
	)
}

// GetPaymentMethodController 获取 PaymentMethod 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.PaymentMethod]: PaymentMethod 缓存对象控制器实例（保证非 nil）
func GetPaymentMethodController() *CacheObjectController[*model.PaymentMethod] {
	return GetCacheObjectController[*model.PaymentMethod](
		&paymentMethodControllerInstance,
		&paymentMethodControllerMutex,
		"PaymentMethod 缓存控制器未初始化，请先调用 InitPaymentMethodController",
	)
}

