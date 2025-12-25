package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// localCache 提供了基于内存的 L1 缓存实现
// 包装了 github.com/patrickmn/go-cache 并支持泛型
type localCache[T any] struct {
	cache *gocache.Cache
}

// newLocalCache 创建一个新的本地缓存实例
func newLocalCache[T any]() *localCache[T] {
	// 默认设置 5 分钟过期，10 分钟进行一次清理周期
	// 具体的过期时间会在 set 时通过 task.TTL() 覆盖
	return &localCache[T]{
		cache: gocache.New(5*time.Minute, 10*time.Minute),
	}
}

// set 存入本地缓存
func (c *localCache[T]) set(key string, val T, ttl time.Duration) {
	c.cache.Set(key, val, ttl)
}

// get 从本地缓存读取
func (c *localCache[T]) get(key string) (T, bool) {
	val, found := c.cache.Get(key)
	if !found {
		var zero T
		return zero, false
	}

	// 泛型类型断言
	t, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return t, true
}

// delete 从本地缓存删除
func (c *localCache[T]) delete(key string) {
	c.cache.Delete(key)
}
