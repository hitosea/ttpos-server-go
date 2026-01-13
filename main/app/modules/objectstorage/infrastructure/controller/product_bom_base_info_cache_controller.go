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

// InitProductBomBaseInfoCacheController 初始化商品BOM基础信息缓存控制器（单例模式）
func InitProductBomBaseInfoCacheController() {
	InitCacheVersionController(
		cache.Global,
		persistence.ObjectTypeProductBomBaseInfo,
		&productBomBaseInfoCacheControllerInstance,
		&productBomBaseInfoCacheControllerOnce,
	)
}

// GetProductBomBaseInfoCacheController 获取商品BOM基础信息缓存控制器单例
func GetProductBomBaseInfoCacheController() *CacheVersionController {
	productBomBaseInfoCacheControllerOnce.Do(func() {
		InitProductBomBaseInfoCacheController()
	})
	return GetCacheVersionController(
		&productBomBaseInfoCacheControllerInstance,
		"ProductBomBaseInfo 缓存控制器未初始化，请先调用 InitProductBomBaseInfoCacheController",
	)
}
