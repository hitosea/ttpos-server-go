package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// CacheAdapter 适配现有的 cache.Cache 接口到 CacheLayer 接口（泛型版本）
type CacheAdapter[T any] struct {
	cache cache.Cache
}

// NewCacheAdapter 创建缓存适配器
func NewCacheAdapter[T any](c cache.Cache) repository.CacheLayer[T] {
	return &CacheAdapter[T]{cache: c}
}

// GET 从缓存获取，未命中时调用 query 函数查询并写入缓存
func (a *CacheAdapter[T]) GET(key string, query func() (T, error)) (T, error) {
	var zero T
	// 先尝试从缓存获取（使用 GetBytes 直接获取 []byte）
	if bytes, found := a.cache.GetBytes(key); found {
		// 反序列化 JSON
		var obj T
		if err := json.Unmarshal(bytes, &obj); err != nil {
			// 反序列化失败，返回零值
			logger.Logger.Warn("缓存反序列化失败", zap.String("key", key), zap.Error(err))
			return zero, fmt.Errorf("反序列化失败: %w", err)
		}
		// 记录缓存命中
		logger.Logger.Debug("缓存命中", zap.String("key", key), zap.String("type", "GET"))
		return obj, nil
	}

	// 缓存未命中，调用查询函数
	logger.Logger.Debug("缓存未命中", zap.String("key", key), zap.String("type", "GET"))
	result, err := query()
	if err != nil {
		return zero, err
	}

	// 将结果序列化为 JSON 并写入缓存（使用默认 TTL）
	jsonData, err := json.Marshal(result)
	if err != nil {
		logger.Logger.Warn("缓存序列化失败", zap.String("key", key), zap.Error(err))
		return result, nil // 序列化失败，返回原始结果但不缓存
	}

	// 写入缓存（直接存储 []byte，不转换为 string）
	if err := a.cache.Set(key, jsonData, 5*time.Minute); err != nil {
		logger.Logger.Warn("缓存写入失败", zap.String("key", key), zap.Error(err))
		return result, nil // 写入失败，返回原始结果
	}

	logger.Logger.Debug("缓存写入成功", zap.String("key", key), zap.String("type", "GET"))
	return result, nil
}

// SET 设置缓存
// 直接存储 []byte，不转换为 string
func (a *CacheAdapter[T]) SET(key string, value T, ttl time.Duration) error {
	// 序列化为 JSON 字节数组，直接存储 []byte
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	// 直接传递 []byte，不转换为 string
	return a.cache.Set(key, jsonData, ttl)
}

// DEL 删除缓存
func (a *CacheAdapter[T]) DEL(keys ...string) error {
	a.cache.Del(keys...)
	return nil
}

// BATCH_GET 批量获取缓存
func (a *CacheAdapter[T]) BATCH_GET(keys []string, query func([]string) (map[string]T, error)) (map[string]T, error) {
	result := make(map[string]T)
	var missedKeys []string

	// 尝试从缓存获取（使用 GetBatchBytes 批量获取 []byte）
	batchBytes, missedKeys := a.cache.GetBatchBytes(keys)
	hitCount := len(batchBytes)
	missCount := len(missedKeys)

	// 记录批量缓存命中情况
	logger.Logger.Debug("批量缓存查询",
		zap.Int("total", len(keys)),
		zap.Int("hit", hitCount),
		zap.Int("miss", missCount),
		zap.String("type", "BATCH_GET"),
	)

	for key, bytes := range batchBytes {
		// 反序列化 JSON
		var obj T
		if err := json.Unmarshal(bytes, &obj); err == nil {
			result[key] = obj
		} else {
			// 反序列化失败时，该 key 会被加入 missedKeys，后续会重新查询
			logger.Logger.Warn("批量缓存反序列化失败", zap.String("key", key), zap.Error(err))
		}
	}

	// 如果有未命中的 key，调用查询函数
	if len(missedKeys) > 0 {
		queryResult, err := query(missedKeys)
		if err != nil {
			return result, err
		}

		// 将查询结果合并到结果中，并写入缓存
		writeSuccessCount := 0
		for key, val := range queryResult {
			result[key] = val
			// 写入缓存（直接存储 []byte，不转换为 string，使用默认 TTL）
			jsonData, err := json.Marshal(val)
			if err == nil {
				if err := a.cache.Set(key, jsonData, 5*time.Minute); err == nil {
					writeSuccessCount++
				} else {
					logger.Logger.Warn("批量缓存写入失败", zap.String("key", key), zap.Error(err))
				}
			} else {
				logger.Logger.Warn("批量缓存序列化失败", zap.String("key", key), zap.Error(err))
			}
		}
		logger.Logger.Debug("批量缓存写入完成",
			zap.Int("total", len(queryResult)),
			zap.Int("success", writeSuccessCount),
			zap.String("type", "BATCH_GET"),
		)
	}

	return result, nil
}

// SCAN 扫描匹配模式的 key（用于批量失效）
func (a *CacheAdapter[T]) SCAN(ctx context.Context, pattern string) ([]string, error) {
	// 注意：现有的 cache.Cache 接口可能不支持 SCAN
	// 这里返回 nil 表示不支持，调用方会使用其他方式处理
	// 如果需要支持，需要扩展 cache.Cache 接口或使用 Redis 客户端直接操作
	return nil, nil
}
