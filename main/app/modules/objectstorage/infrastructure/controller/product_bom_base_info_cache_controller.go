package controller

import (
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// 全局单例实例
var (
	productBomBaseInfoCacheControllerInstance *CacheVersionController
	productBomBaseInfoCacheControllerOnce     sync.Once
)

// GetProductBomBaseInfoCacheController 获取商品BOM基础信息缓存控制器单例
func GetProductBomBaseInfoCacheController() *CacheVersionController {
	productBomBaseInfoCacheControllerOnce.Do(func() {
		productBomBaseInfoCacheControllerInstance = InitCacheVersionController(
			cache.Global,
			persistence.ObjectTypeProductBomBaseInfo,
		)
	})
	return productBomBaseInfoCacheControllerInstance
}
