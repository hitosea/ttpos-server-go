package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// DeskQueryFunc 单个 Desk 查询函数类型
type DeskQueryFunc func(db *gorm.DB, uuid uint64) (*model.Desk, error)

// DeskBatchQueryFunc 批量 Desk 查询函数类型
type DeskBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.Desk, error)

// deskControllerInstance Desk 控制器单例实例
var (
	deskControllerInstance *CacheObjectController[*model.Desk]
	deskControllerOnce     sync.Once
)

// InitDeskController 初始化 Desk 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 Desk 查询函数
//   - batchQueryFunc: 批量 Desk 查询函数
func InitDeskController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc DeskQueryFunc,
	batchQueryFunc DeskBatchQueryFunc,
) {
	deskControllerOnce.Do(func() {
		deskControllerInstance = InitCacheObjectController[*model.Desk](
			underlyingCache,
			ttl,
			func(db *gorm.DB, uuid uint64) (*model.Desk, error) {
				return queryFunc(db, uuid)
			},
			func(db *gorm.DB, uuids []uint64) ([]*model.Desk, error) {
				return batchQueryFunc(db, uuids)
			},
			func(obj *model.Desk) uint64 {
				return obj.Uuid
			},
			persistence.ObjectTypeDesk,
		)
	})
}

// GetDeskController 获取 Desk 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.Desk]: Desk 缓存对象控制器实例（保证非 nil）
func GetDeskController() *CacheObjectController[*model.Desk] {
	return deskControllerInstance
}

// 确保 CacheObjectController 实现了 ICacheObjectController 接口
var _ ICacheObjectController = (*CacheObjectController[*model.Desk])(nil)
