package repository

import (
	"context"
	"time"
)

// CacheLayer 三级缓存基础包接口
type CacheLayer interface {
	// GET 方法：从缓存获取，未命中时调用 query 函数查询并写入缓存
	GET(key string, query func() (any, error)) (any, error)

	// SET 方法：设置缓存
	SET(key string, value any, ttl time.Duration) error

	// DEL 方法：删除缓存
	DEL(keys ...string) error

	// BATCH_GET 方法：批量获取缓存
	BATCH_GET(keys []string, query func([]string) (map[string]any, error)) (map[string]any, error)

	// SCAN 方法：扫描匹配模式的 key（可选，用于批量失效）
	// 如果未实现，返回 nil, nil
	SCAN(ctx context.Context, pattern string) ([]string, error)
}

