package cache

import (
	"context"
	"encoding/json"
	"time"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

const (
	// MaxCacheSizeWarningThreshold 缓存大小警告阈值（1MB）
	// 超过此大小的缓存数据会被记录警告日志
	MaxCacheSizeWarningThreshold = 1024 * 1024 // 1MB

	// MaxCacheSizeErrorThreshold 缓存大小错误阈值（10MB）
	// 超过此大小的缓存数据会被记录错误日志，建议检查数据是否过大
	MaxCacheSizeErrorThreshold = 10 * 1024 * 1024 // 10MB
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
		logger.Logger.Error("缓存序列化失败",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	// 记录缓存大小并检查阈值
	size := len(bytes)
	if size >= MaxCacheSizeErrorThreshold {
		// 超过错误阈值，记录错误日志
		logger.Logger.Error("缓存数据过大（超过错误阈值）",
			zap.String("key", key),
			zap.Int("size_bytes", size),
			zap.Float64("size_mb", float64(size)/(1024*1024)),
			zap.Duration("ttl", ttl),
			zap.String("level", "L2"),
			zap.String("type", "SET"),
		)
	} else if size >= MaxCacheSizeWarningThreshold {
		// 超过警告阈值，记录警告日志
		logger.Logger.Warn("缓存数据较大（超过警告阈值）",
			zap.String("key", key),
			zap.Int("size_bytes", size),
			zap.Float64("size_kb", float64(size)/1024),
			zap.Duration("ttl", ttl),
			zap.String("level", "L2"),
			zap.String("type", "SET"),
		)
	} else {
		// 正常大小，记录调试日志
		logger.Logger.Debug("缓存写入 L2",
			zap.String("key", key),
			zap.Int("size_bytes", size),
			zap.Duration("ttl", ttl),
			zap.String("level", "L2"),
			zap.String("type", "SET"),
		)
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
