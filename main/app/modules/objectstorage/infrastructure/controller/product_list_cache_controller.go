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

// InitProductListCacheController 初始化商品列表缓存控制器（单例模式）
func InitProductListCacheController() {
	InitCacheVersionController(
		cache.Global,
		persistence.ObjectTypeProductList,
		&productListCacheControllerInstance,
		&productListCacheControllerOnce,
	)
}

// GetProductListCacheController 获取商品列表缓存控制器单例
func GetProductListCacheController() *CacheVersionController {
	productListCacheControllerOnce.Do(func() {
		InitProductListCacheController()
	})
	return GetCacheVersionController(
		&productListCacheControllerInstance,
		"ProductList 缓存控制器未初始化，请先调用 InitProductListCacheController",
	)
}
