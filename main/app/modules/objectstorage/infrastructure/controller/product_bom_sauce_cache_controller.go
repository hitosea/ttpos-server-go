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

// InitProductBomSauceCacheController 初始化小料商品BOM缓存控制器（单例模式）
func InitProductBomSauceCacheController() {
	InitCacheVersionController(
		cache.Global,
		persistence.ObjectTypeProductBomSauce,
		&productBomSauceCacheControllerInstance,
		&productBomSauceCacheControllerOnce,
	)
}

// GetProductBomSauceCacheController 获取小料商品BOM缓存控制器单例
func GetProductBomSauceCacheController() *CacheVersionController {
	productBomSauceCacheControllerOnce.Do(func() {
		InitProductBomSauceCacheController()
	})
	return GetCacheVersionController(
		&productBomSauceCacheControllerInstance,
		"ProductBomSauce 缓存控制器未初始化，请先调用 InitProductBomSauceCacheController",
	)
}
