package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// MemberQueryFunc 单个 Member 查询函数类型
type MemberQueryFunc func(db *gorm.DB, uuid uint64) (*model.Member, error)

// MemberBatchQueryFunc 批量 Member 查询函数类型
type MemberBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.Member, error)

// memberControllerInstance Member 控制器单例实例
var (
	memberControllerInstance *CacheObjectController[*model.Member]
	memberControllerOnce     sync.Once
	memberControllerMutex    sync.RWMutex
)

// InitMemberController 初始化 Member 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 Member 查询函数
//   - batchQueryFunc: 批量 Member 查询函数
func InitMemberController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc MemberQueryFunc,
	batchQueryFunc MemberBatchQueryFunc,
) {
	InitCacheObjectController[*model.Member](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.Member, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.Member, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.Member) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeMember,
		&memberControllerInstance,
		&memberControllerOnce,
		&memberControllerMutex,
	)
}

// GetMemberController 获取 Member 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.Member]: Member 缓存对象控制器实例（保证非 nil）
func GetMemberController() *CacheObjectController[*model.Member] {
	return GetCacheObjectController[*model.Member](
		&memberControllerInstance,
		&memberControllerMutex,
		"Member 缓存控制器未初始化，请先调用 InitMemberController",
	)
}

