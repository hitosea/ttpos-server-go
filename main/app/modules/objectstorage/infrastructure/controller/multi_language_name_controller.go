package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// MultiLanguageNameQueryFunc 单个 MultiLanguageName 查询函数类型
type MultiLanguageNameQueryFunc func(db *gorm.DB, uuid uint64) (*model.MultiLanguageName, error)

// MultiLanguageNameBatchQueryFunc 批量 MultiLanguageName 查询函数类型
type MultiLanguageNameBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.MultiLanguageName, error)

// multiLanguageNameControllerInstance MultiLanguageName 控制器单例实例
var (
	multiLanguageNameControllerInstance *CacheObjectController[*model.MultiLanguageName]
	multiLanguageNameControllerOnce     sync.Once
	multiLanguageNameControllerMutex    sync.RWMutex
)

// InitMultiLanguageNameController 初始化 MultiLanguageName 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 MultiLanguageName 查询函数
//   - batchQueryFunc: 批量 MultiLanguageName 查询函数
func InitMultiLanguageNameController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc MultiLanguageNameQueryFunc,
	batchQueryFunc MultiLanguageNameBatchQueryFunc,
) {
	InitCacheObjectController[*model.MultiLanguageName](
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.MultiLanguageName, error) {
			return queryFunc(db, uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.MultiLanguageName, error) {
			return batchQueryFunc(db, uuids)
		},
		func(obj *model.MultiLanguageName) uint64 {
			return obj.Uuid
		},
		persistence.ObjectTypeMultiLanguageName,
		&multiLanguageNameControllerInstance,
		&multiLanguageNameControllerOnce,
		&multiLanguageNameControllerMutex,
	)
}

// GetMultiLanguageNameController 获取 MultiLanguageName 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.MultiLanguageName]: MultiLanguageName 缓存对象控制器实例（保证非 nil）
func GetMultiLanguageNameController() *CacheObjectController[*model.MultiLanguageName] {
	return GetCacheObjectController[*model.MultiLanguageName](
		&multiLanguageNameControllerInstance,
		&multiLanguageNameControllerMutex,
		"MultiLanguageName 缓存控制器未初始化，请先调用 InitMultiLanguageNameController",
	)
}
