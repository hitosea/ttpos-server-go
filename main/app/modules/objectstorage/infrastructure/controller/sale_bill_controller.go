package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// SaleBillQueryFunc 单个 SaleBill 查询函数类型
type SaleBillQueryFunc func(db *gorm.DB, uuid uint64) (*model.SaleBill, error)

// SaleBillBatchQueryFunc 批量 SaleBill 查询函数类型
type SaleBillBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.SaleBill, error)

// saleBillControllerInstance SaleBill 控制器单例实例
var (
	saleBillControllerInstance *CacheObjectController[*model.SaleBill]
	saleBillControllerOnce     sync.Once
)

// InitSaleBillController 初始化 SaleBill 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 SaleBill 查询函数
//   - batchQueryFunc: 批量 SaleBill 查询函数
func InitSaleBillController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc SaleBillQueryFunc,
	batchQueryFunc SaleBillBatchQueryFunc,
) {
	// 构建缓存层选项,禁用 L1 本地缓存
	cacheOpts := []func(*adapter.CacheLayerOption){adapter.WithEnableL1Cache(false)}
	saleBillControllerOnce.Do(func() {
		saleBillControllerInstance = InitCacheObjectController[*model.SaleBill](
			underlyingCache,
			ttl,
			func(db *gorm.DB, uuid uint64) (*model.SaleBill, error) {
				return queryFunc(db, uuid)
			},
			func(db *gorm.DB, uuids []uint64) ([]*model.SaleBill, error) {
				return batchQueryFunc(db, uuids)
			},
			func(obj *model.SaleBill) uint64 {
				return obj.Uuid
			},
			persistence.ObjectTypeSaleBill,
			cacheOpts...,
		)
	})
}

// GetSaleBillController 获取 SaleBill 对象的缓存控制器单例实例
func GetSaleBillController() *CacheObjectController[*model.SaleBill] {
	return saleBillControllerInstance
}
