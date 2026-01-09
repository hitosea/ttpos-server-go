package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// SaleBillSettingQueryFunc 单个 SaleBillSetting 查询函数类型
type SaleBillSettingQueryFunc func(db *gorm.DB, saleBillUuid uint64) (*model.SaleBillSetting, error)

// SaleBillSettingBatchQueryFunc 批量 SaleBillSetting 查询函数类型
type SaleBillSettingBatchQueryFunc func(db *gorm.DB, saleBillUuids []uint64) ([]*model.SaleBillSetting, error)

// saleBillSettingControllerInstance SaleBillSetting 控制器单例实例
var (
	saleBillSettingControllerInstance *CacheObjectController[*model.SaleBillSetting]
	saleBillSettingControllerOnce     sync.Once
	saleBillSettingControllerMutex    sync.RWMutex
)

// InitSaleBillSettingController 初始化 SaleBillSetting 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 SaleBillSetting 查询函数
//   - batchQueryFunc: 批量 SaleBillSetting 查询函数
func InitSaleBillSettingController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc SaleBillSettingQueryFunc,
	batchQueryFunc SaleBillSettingBatchQueryFunc,
) {
	InitCacheObjectController[*model.SaleBillSetting](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.SaleBillSetting, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.SaleBillSetting, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.SaleBillSetting) uint64 {
			return obj.SaleBillUuid
		},
		persistence.ObjectTypeSaleBillSetting,
		&saleBillSettingControllerInstance,
		&saleBillSettingControllerOnce,
		&saleBillSettingControllerMutex,
	)
}

// GetSaleBillSettingController 获取 SaleBillSetting 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.SaleBillSetting]: SaleBillSetting 缓存对象控制器实例（保证非 nil）
func GetSaleBillSettingController() *CacheObjectController[*model.SaleBillSetting] {
	return GetCacheObjectController[*model.SaleBillSetting](
		&saleBillSettingControllerInstance,
		&saleBillSettingControllerMutex,
		"SaleBillSetting 缓存控制器未初始化，请先调用 InitSaleBillSettingController",
	)
}
