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
	// ClearL1 清空 L1 本地缓存
	ClearL1()
}

// L1CacheClearable 非泛型接口，用于清空 L1 缓存
// 允许在 sync.Map 中直接存储 cacheGroup 实例，方便全局操作
type L1CacheClearable interface {
	// ClearL1 清空 L1 本地缓存
	ClearL1()
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

	// L1TTL L1 本地缓存的过期时间，如果为 0 则使用任务 TTL
	// 建议设置为任务 TTL 的 1/3 到 1/2，例如任务 TTL 为 5 分钟，L1TTL 可设置为 1-2 分钟
	// 阶梯式 TTL 的优势：
	//   - L1 使用较短 TTL，减少内存占用
	//   - L2 使用较长 TTL，保持缓存命中率
	//   - L1 过期后可从 L2 回填，避免直接查数据库
	L1TTL time.Duration

	// L2TTL L2 Redis 缓存的过期时间，如果为 0 则使用任务 TTL
	// 建议设置为任务 TTL 的 1.5 到 2 倍，例如任务 TTL 为 5 分钟，L2TTL 可设置为 7-10 分钟
	L2TTL time.Duration
}
