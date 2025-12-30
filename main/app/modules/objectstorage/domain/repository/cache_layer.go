package repository

import (
	"context"
	"time"
)

// GetOption GET 方法的选项
type GetOption struct {
	// SkipCache 是否跳过缓存检查，直接执行查询
	// true: 跳过所有缓存（L1/L2），直接从数据库查询（仍会写入缓存）
	// false: 正常流程，按 L1 -> L2 -> L3 的顺序查找（默认）
	SkipCache bool
}

// WithSkipCache 设置是否跳过缓存选项
func WithSkipCache() func(*GetOption) {
	return func(opt *GetOption) {
		opt.SkipCache = true
	}
}

// CacheLayer 三级缓存基础包接口（泛型版本）
type CacheLayer[T any] interface {
	// GET 方法：从缓存获取，未命中时调用 query 函数查询并写入缓存
	// 默认行为：按 L1 -> L2 -> L3 的顺序自动查找
	// 参数：
	//   - key: 缓存 key
	//   - query: 查询函数（当需要从数据库查询时调用）
	//   - opts: 选项函数（可选），如 WithSkipCache() 跳过缓存直接查询
	GET(key string, query func() (T, error), opts ...func(*GetOption)) (T, error)

	// BATCH_GET 方法：批量获取缓存
	// 参数：
	//   - keys: 缓存 key 列表
	//   - query: 批量查询函数（当需要从数据库查询时调用）
	//   - opts: 选项函数（可选），如 WithSkipCache() 跳过缓存直接查询
	BATCH_GET(keys []string, query func([]string) (map[string]T, error), opts ...func(*GetOption)) (map[string]T, error)

	// SET 方法：设置缓存
	SET(key string, value T, ttl time.Duration) error

	// DEL 方法：删除缓存
	DEL(keys ...string) error

	// SCAN 方法：扫描匹配模式的 key（可选，用于批量失效）
	// 如果未实现，返回 nil, nil
	SCAN(ctx context.Context, pattern string) ([]string, error)
}
