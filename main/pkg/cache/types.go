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

// DoOption Do 方法的选项
type DoOption struct {
	// SkipCache 是否跳过 L1/L2 缓存检查，直接执行查询逻辑
	// true: 跳过缓存检查，直接执行 g.sf.do 逻辑（仍会写入缓存）
	// false: 正常流程，先检查 L1/L2 缓存，未命中再执行查询
	SkipCache bool
}

// WithSkipCache 设置跳过缓存选项
func WithSkipCache() func(*DoOption) {
	return func(opt *DoOption) {
		opt.SkipCache = true
	}
}

// ICacheGroup 定义缓存组接口
// 提供统一的 Read-Through 访问入口
type ICacheGroup[T any] interface {
	// Do 执行任务，内部自动处理 L1/L2 缓存及 Singleflight 合并
	// 参数：
	//   - ctx: 上下文
	//   - task: 任务
	//   - opts: 选项函数（可选），如 WithSkipCache(true) 跳过缓存直接查询
	Do(ctx context.Context, task Task[T], opts ...func(*DoOption)) (T, error)

	// ClearL1 清空 L1 本地缓存
	ClearL1()

	// DeleteL1 删除 L1 本地缓存中指定的 key
	DeleteL1(key string)
}

// L1CacheClearable 非泛型接口，用于清空 L1 缓存
// 允许在 sync.Map 中直接存储 cacheGroup 实例，方便全局操作
type L1CacheClearable interface {
	// ClearL1 清空 L1 本地缓存
	ClearL1()

	// DeleteL1 删除 L1 本地缓存中指定的 key
	DeleteL1(key string)
}

// HitStatsCallback 缓存命中率统计回调函数
// 参数：
//   - key: 缓存 key
//   - hit: 是否命中（true=命中，false=未命中）
//   - level: 缓存层级（"L1"、"L2" 或 ""，空字符串表示未命中）
type HitStatsCallback func(key string, hit bool, level string)

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
	// TTL 优先级：L1TTL > taskTTL
	L1TTL time.Duration

	// L2TTL L2 Redis 缓存的过期时间，如果为 0 则使用任务 TTL
	// 建议设置为任务 TTL 的 1.5 到 2 倍，例如任务 TTL 为 5 分钟，L2TTL 可设置为 7-10 分钟
	// TTL 优先级：L2TTL > taskTTL
	L2TTL time.Duration

	// HitStatsCallback 缓存命中率统计回调函数（可选）
	// 当缓存命中或未命中时，会调用此回调函数进行统计
	// 如果为 nil，则不进行统计
	HitStatsCallback HitStatsCallback

	// AsyncL2Write 是否异步写入 L2 Redis 缓存
	// true: 异步写入 L2，提升响应速度（适用于 L2 写入延迟较高的场景）
	// false: 同步写入 L2，保证数据一致性（默认）
	// 注意：
	//   - L1 本地缓存始终同步写入（内存操作延迟低）
	//   - 负缓存（NegativeTTL）始终同步写入（防止缓存穿透）
	//   - 异步写入失败时，后续请求可能无法命中缓存，但不会影响当前请求的响应
	AsyncL2Write bool
}
