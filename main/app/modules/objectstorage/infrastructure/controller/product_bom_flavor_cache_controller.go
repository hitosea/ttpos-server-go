package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	productBomFlavorCacheControllerInstance *CacheVersionController
	productBomFlavorCacheControllerOnce     sync.Once
)

// GetProductBomFlavorCacheController 获取规格商品BOM缓存控制器单例
func GetProductBomFlavorCacheController() *CacheVersionController {
	productBomFlavorCacheControllerOnce.Do(func() {
		productBomFlavorCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeProductBomFlavor,
		)
	})
	return productBomFlavorCacheControllerInstance
}
