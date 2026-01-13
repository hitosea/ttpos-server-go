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

// InitProductPackageAttributeCacheController 初始化商品包属性缓存控制器（单例模式）
func InitProductPackageAttributeCacheController() {
	InitCacheVersionController(
		cache.Global,
		persistence.ObjectTypeProductPackageAttribute,
		&productPackageAttributeCacheControllerInstance,
		&productPackageAttributeCacheControllerOnce,
	)
}

// GetProductPackageAttributeCacheController 获取商品包属性缓存控制器单例
func GetProductPackageAttributeCacheController() *CacheVersionController {
	productPackageAttributeCacheControllerOnce.Do(func() {
		InitProductPackageAttributeCacheController()
	})
	return GetCacheVersionController(
		&productPackageAttributeCacheControllerInstance,
		"ProductPackageAttribute 缓存控制器未初始化，请先调用 InitProductPackageAttributeCacheController",
	)
}

