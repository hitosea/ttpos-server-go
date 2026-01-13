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

// InitProductBomFlavorCacheController 初始化规格商品BOM缓存控制器（单例模式）
func InitProductBomFlavorCacheController() {
	InitCacheVersionController(
		cache.Global,
		persistence.ObjectTypeProductBomFlavor,
		&productBomFlavorCacheControllerInstance,
		&productBomFlavorCacheControllerOnce,
	)
}

// GetProductBomFlavorCacheController 获取规格商品BOM缓存控制器单例
func GetProductBomFlavorCacheController() *CacheVersionController {
	productBomFlavorCacheControllerOnce.Do(func() {
		InitProductBomFlavorCacheController()
	})
	return GetCacheVersionController(
		&productBomFlavorCacheControllerInstance,
		"ProductBomFlavor 缓存控制器未初始化，请先调用 InitProductBomFlavorCacheController",
	)
}
