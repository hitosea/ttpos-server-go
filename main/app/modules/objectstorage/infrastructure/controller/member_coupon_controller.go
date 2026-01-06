package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// MemberCouponQueryFunc 单个 MemberCoupon 查询函数类型
type MemberCouponQueryFunc func(db *gorm.DB, uuid uint64) (*model.MemberCoupon, error)

// MemberCouponBatchQueryFunc 批量 MemberCoupon 查询函数类型
type MemberCouponBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.MemberCoupon, error)

// memberCouponControllerInstance MemberCoupon 控制器单例实例
var (
	memberCouponControllerInstance *CacheObjectController[*model.MemberCoupon]
	memberCouponControllerOnce     sync.Once
	memberCouponControllerMutex    sync.RWMutex
)

// InitMemberCouponController 初始化 MemberCoupon 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 MemberCoupon 查询函数
//   - batchQueryFunc: 批量 MemberCoupon 查询函数
func InitMemberCouponController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc MemberCouponQueryFunc,
	batchQueryFunc MemberCouponBatchQueryFunc,
) {
	InitCacheObjectController[*model.MemberCoupon](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.MemberCoupon, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.MemberCoupon, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.MemberCoupon) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeMemberCoupon,
		&memberCouponControllerInstance,
		&memberCouponControllerOnce,
		&memberCouponControllerMutex,
	)
}

// GetMemberCouponController 获取 MemberCoupon 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.MemberCoupon]: MemberCoupon 缓存对象控制器实例（保证非 nil）
func GetMemberCouponController() *CacheObjectController[*model.MemberCoupon] {
	return GetCacheObjectController[*model.MemberCoupon](
		&memberCouponControllerInstance,
		&memberCouponControllerMutex,
		"MemberCoupon 缓存控制器未初始化，请先调用 InitMemberCouponController",
	)
}

