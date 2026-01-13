package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	productPackageAttributeCacheControllerInstance *CacheVersionController
	productPackageAttributeCacheControllerOnce     sync.Once
)

// GetProductPackageAttributeCacheController 获取商品包属性缓存控制器单例
func GetProductPackageAttributeCacheController() *CacheVersionController {
	productPackageAttributeCacheControllerOnce.Do(func() {
		productPackageAttributeCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeProductPackageAttribute,
		)
	})
	return productPackageAttributeCacheControllerInstance
}
