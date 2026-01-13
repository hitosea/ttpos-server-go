package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	productListCacheControllerInstance *CacheVersionController
	productListCacheControllerOnce     sync.Once
)

// GetProductListCacheController 获取商品列表缓存控制器单例
func GetProductListCacheController() *CacheVersionController {
	productListCacheControllerOnce.Do(func() {
		productListCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeProductList,
		)
	})
	return productListCacheControllerInstance
}
