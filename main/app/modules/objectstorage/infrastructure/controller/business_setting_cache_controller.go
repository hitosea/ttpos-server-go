package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	businessSettingCacheControllerInstance *CacheVersionController
	businessSettingCacheControllerOnce     sync.Once
)

// GetBusinessSettingCacheController 获取门店业务设置缓存控制器单例
func GetBusinessSettingCacheController() *CacheVersionController {
	businessSettingCacheControllerOnce.Do(func() {
		businessSettingCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeBusinessSetting,
		)
	})
	return businessSettingCacheControllerInstance
}
