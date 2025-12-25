package cache

import (
	"context"
	"time"
)

// Task 定义一个可合并且可缓存的任务
// T 是返回的结果类型
type Task[T any] interface {
	// Hash 返回任务的唯一标识，用于 Singleflight 合并请求
	Hash() string

	// Exec 实际的业务执行逻辑（如查库、调 API 等）
	Exec(ctx context.Context) (T, error)

	// TTL 返回该任务结果的缓存过期时间
	TTL() time.Duration
}

// ICacheGroup 定义缓存组接口
// 提供统一的 Read-Through 访问入口
type ICacheGroup[T any] interface {
	// Do 执行任务，内部自动处理 L1/L2 缓存及 Singleflight 合并
	Do(ctx context.Context, task Task[T]) (T, error)
}

// GroupConfig 缓存组配置
type GroupConfig struct {
	// Name 缓存组名称，用于日志和监控埋点
	Name string

	// EnableLocalCache 是否开启 L1 本地内存缓存
	EnableLocalCache bool

	// EnableRedisCache 是否开启 L2 Redis 分布式缓存
	EnableRedisCache bool

	// NegativeTTL 负缓存（空结果/错误）的过期时间，设为 0 则不开启
	NegativeTTL time.Duration
}
