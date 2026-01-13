package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	productBomSauceCacheControllerInstance *CacheVersionController
	productBomSauceCacheControllerOnce     sync.Once
)

// GetProductBomSauceCacheController 获取小料商品BOM缓存控制器单例
func GetProductBomSauceCacheController() *CacheVersionController {
	productBomSauceCacheControllerOnce.Do(func() {
		productBomSauceCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeProductBomSauce,
		)
	})
	return productBomSauceCacheControllerInstance
}
