package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/pkg/cache"
)

// CacheAdapter 适配现有的 cache.Cache 接口到 CacheLayer 接口
type CacheAdapter struct {
	cache cache.Cache
}

// NewCacheAdapter 创建缓存适配器
func NewCacheAdapter(c cache.Cache) repository.CacheLayer {
	return &CacheAdapter{cache: c}
}

// GET 从缓存获取，未命中时调用 query 函数查询并写入缓存
func (a *CacheAdapter) GET(key string, query func() (any, error)) (any, error) {
	// 先尝试从缓存获取
	if val, found := a.cache.Get(key); found {
		// 反序列化 JSON
		var obj any
		if strVal, ok := val.(string); ok {
			if err := json.Unmarshal([]byte(strVal), &obj); err == nil {
				return obj, nil
			}
		}
		// 如果不是字符串或反序列化失败，直接使用原始值
		return val, nil
	}

	// 缓存未命中，调用查询函数
	result, err := query()
	if err != nil {
		return nil, err
	}

	// 将结果序列化为 JSON 并写入缓存（使用默认 TTL）
	jsonData, err := json.Marshal(result)
	if err != nil {
		return result, nil // 序列化失败，返回原始结果但不缓存
	}

	// 写入缓存（使用字符串形式）
	if err := a.cache.Set(key, string(jsonData), 5*time.Minute); err != nil {
		return result, nil // 写入失败，返回原始结果
	}

	return result, nil
}

// SET 设置缓存
func (a *CacheAdapter) SET(key string, value any, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	return a.cache.Set(key, string(jsonData), ttl)
}

// DEL 删除缓存
func (a *CacheAdapter) DEL(keys ...string) error {
	a.cache.Del(keys...)
	return nil
}

// BATCH_GET 批量获取缓存
func (a *CacheAdapter) BATCH_GET(keys []string, query func([]string) (map[string]any, error)) (map[string]any, error) {
	result := make(map[string]any)
	var missedKeys []string

	// 尝试从缓存获取
	for _, key := range keys {
		if val, found := a.cache.Get(key); found {
			// 反序列化 JSON
			var obj any
			if strVal, ok := val.(string); ok {
				if err := json.Unmarshal([]byte(strVal), &obj); err == nil {
					result[key] = obj
					continue
				}
			}
			// 如果不是字符串或反序列化失败，直接使用原始值
			result[key] = val
		} else {
			missedKeys = append(missedKeys, key)
		}
	}

	// 如果有未命中的 key，调用查询函数
	if len(missedKeys) > 0 {
		queryResult, err := query(missedKeys)
		if err != nil {
			return result, err
		}

		// 将查询结果合并到结果中，并写入缓存
		for key, val := range queryResult {
			result[key] = val
			// 写入缓存（使用默认 TTL）
			jsonData, err := json.Marshal(val)
			if err == nil {
				a.cache.Set(key, string(jsonData), 5*time.Minute)
			}
		}
	}

	return result, nil
}

// SCAN 扫描匹配模式的 key（用于批量失效）
func (a *CacheAdapter) SCAN(ctx context.Context, pattern string) ([]string, error) {
	// 注意：现有的 cache.Cache 接口可能不支持 SCAN
	// 这里返回 nil 表示不支持，调用方会使用其他方式处理
	// 如果需要支持，需要扩展 cache.Cache 接口或使用 Redis 客户端直接操作
	return nil, nil
}

