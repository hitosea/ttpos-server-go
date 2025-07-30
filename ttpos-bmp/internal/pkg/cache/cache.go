package cache

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
)

// cache 缓存驱动
var cache *gcache.Cache

// Instance 缓存实例
func Instance() *gcache.Cache {
	if cache == nil {
		panic("cache uninitialized.")
	}
	return cache
}

// SetAdapter 设置缓存适配器
func SetAdapter(ctx context.Context) {
	var adapter gcache.Adapter

	switch g.Cfg().MustGet(ctx, "cache.adapter").String() {
	case "redis":
		adapter = gcache.NewAdapterRedis(g.Redis())

	default:
		adapter = gcache.NewAdapterMemory()
	}

	// 数据库缓存，默认和通用缓冲驱动一致，如果你不想使用默认的，可以自行调整
	g.DB().GetCache().SetAdapter(adapter)

	// 通用缓存
	cache = gcache.New()
	cache.SetAdapter(adapter)
}
