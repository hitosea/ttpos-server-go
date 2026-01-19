package controller

import (
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// CompanySettingQueryFunc 单个 CompanySetting 查询函数类型
type CompanySettingQueryFunc func(db *gorm.DB, uuid uint64) (*model.CompanySetting, error)

// CompanySettingBatchQueryFunc 批量 CompanySetting 查询函数类型
type CompanySettingBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.CompanySetting, error)

// companySettingControllerInstance CompanySetting 控制器单例实例
var (
	companySettingControllerInstance *CacheObjectController[*model.CompanySetting]
	companySettingControllerOnce     sync.Once
)

// InitCompanySettingController 初始化 CompanySetting 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 CompanySetting 查询函数
//   - batchQueryFunc: 批量 CompanySetting 查询函数
func InitCompanySettingController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc CompanySettingQueryFunc,
	batchQueryFunc CompanySettingBatchQueryFunc,
) {
	companySettingControllerOnce.Do(func() {
		companySettingControllerInstance = InitCacheObjectController[*model.CompanySetting](
			underlyingCache,
			ttl,
			func(db *gorm.DB, uuid uint64) (*model.CompanySetting, error) {
				return queryFunc(db, uuid)
			},
			func(db *gorm.DB, uuids []uint64) ([]*model.CompanySetting, error) {
				return batchQueryFunc(db, uuids)
			},
			func(obj *model.CompanySetting) uint64 {
				return obj.Uuid
			},
			persistence.ObjectTypeCompanySetting,
		)
	})
}

// GetCompanySettingController 获取 CompanySetting 对象的缓存控制器单例实例
// 保证返回非 nil 值，如果未初始化会 panic
// 返回：
//   - *CacheObjectController[*model.CompanySetting]: CompanySetting 缓存对象控制器实例（保证非 nil）
func GetCompanySettingController() *CacheObjectController[*model.CompanySetting] {
	return companySettingControllerInstance
}
