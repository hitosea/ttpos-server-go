package cache

import (
	"context"
	"encoding/json"
	"time"
)

// redisCacheWrapper 提供了基于 Redis 的 L2 缓存实现
// 包装了项目现有的 Cache 接口并支持泛型序列化
type redisCacheWrapper[T any] struct {
	// 使用项目全局定义的 Global Cache
}

// newRedisCacheWrapper 创建一个新的 Redis 缓存包装器
func newRedisCacheWrapper[T any]() *redisCacheWrapper[T] {
	return &redisCacheWrapper[T]{}
}

// set 存入 Redis 缓存
func (c *redisCacheWrapper[T]) set(ctx context.Context, key string, val T, ttl time.Duration) error {
	if Global == nil {
		return nil
	}

	// 序列化为 JSON 存储
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return Global.Set(key, bytes, ttl)
}

// get 从 Redis 缓存读取
func (c *redisCacheWrapper[T]) get(ctx context.Context, key string) (T, bool) {
	if Global == nil {
		var zero T
		return zero, false
	}

	// 使用 GetBytes 获取原始数据
	bytes, found := Global.GetBytes(key)
	if !found {
		var zero T
		return zero, false
	}

	// 反序列化为目标泛型类型
	var val T
	if err := json.Unmarshal(bytes, &val); err != nil {
		var zero T
		return zero, false
	}

	return val, true
}

// delete 从 Redis 缓存删除
func (c *redisCacheWrapper[T]) delete(ctx context.Context, key string) {
	if Global == nil {
		return
	}
	Global.Del(key)
}
