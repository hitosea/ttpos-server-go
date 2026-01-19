package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// MarketingCouponQueryFunc 单个 MarketingCoupon 查询函数类型
type MarketingCouponQueryFunc func(db *gorm.DB, uuid uint64) (*model.MarketingCoupon, error)

// MarketingCouponBatchQueryFunc 批量 MarketingCoupon 查询函数类型
type MarketingCouponBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.MarketingCoupon, error)

// marketingCouponControllerInstance MarketingCoupon 控制器单例实例
var (
	marketingCouponControllerInstance *CacheObjectController[*model.MarketingCoupon]
	marketingCouponControllerOnce     sync.Once
)

// InitMarketingCouponController 初始化 MarketingCoupon 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 MarketingCoupon 查询函数
//   - batchQueryFunc: 批量 MarketingCoupon 查询函数
func InitMarketingCouponController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc MarketingCouponQueryFunc,
	batchQueryFunc MarketingCouponBatchQueryFunc,
) {
	marketingCouponControllerOnce.Do(func() {
		marketingCouponControllerInstance = InitCacheObjectController[*model.MarketingCoupon](
			underlyingCache,
			ttl,
			func(db *gorm.DB, uuid uint64) (*model.MarketingCoupon, error) {
				return queryFunc(db, uuid)
			},
			func(db *gorm.DB, uuids []uint64) ([]*model.MarketingCoupon, error) {
				return batchQueryFunc(db, uuids)
			},
			func(obj *model.MarketingCoupon) uint64 {
				return obj.Uuid
			},
			persistence.ObjectTypeMarketingCoupon,
		)
	})
}

// GetMarketingCouponController 获取 MarketingCoupon 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.MarketingCoupon]: MarketingCoupon 缓存对象控制器实例（保证非 nil）
func GetMarketingCouponController() *CacheObjectController[*model.MarketingCoupon] {
	return marketingCouponControllerInstance
}
