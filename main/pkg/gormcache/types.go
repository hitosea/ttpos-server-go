// Package gormcache 提供 GORM 查询缓存功能，支持表级精确缓存失效
//
// 核心特性：
//   - 表级缓存：按表名组织缓存，写操作只失效对应表的缓存
//   - Easer 请求合并：相同查询并发时只执行一次
//   - 与项目现有 cache 包无缝集成
//   - 支持单机和 Redis 集群模式
package gormcache

import (
	"context"
	"time"
)

// QueryResult 缓存的查询结果
type QueryResult struct {
	Dest         any   `json:"dest"`          // 查询结果数据
	RowsAffected int64 `json:"rows_affected"` // 影响行数
}

// Cacher 缓存存储接口
// 实现此接口以支持不同的缓存后端（Redis、内存等）
type Cacher interface {
	// Get 从缓存获取查询结果
	// key: 由 buildIdentifier 生成的缓存键
	// 返回: 查询结果和是否命中
	Get(ctx context.Context, key string) (*QueryResult, bool)

	// Store 将查询结果存入缓存
	// tableKey: 表索引键，格式为 {dbName}:{tableName}，用于多租户隔离和表级失效
	// key: 缓存键
	// result: 查询结果
	// ttl: 过期时间
	Store(ctx context.Context, tableKey string, key string, result *QueryResult, ttl time.Duration) error

	// InvalidateTable 失效指定表的所有缓存
	// tableKey: 表索引键，格式为 {dbName}:{tableName}
	// 在 INSERT/UPDATE/DELETE 后调用
	InvalidateTable(ctx context.Context, tableKey string) error

	// InvalidateAll 失效所有缓存
	InvalidateAll(ctx context.Context) error
}

// Config 缓存插件配置
type Config struct {
	// Easer 是否启用请求合并（singleflight）
	// 当多个相同查询并发时，只执行一次数据库查询
	Easer bool

	// Cacher 缓存存储实现
	// 如果为 nil，则不启用缓存存储，但 Easer 仍可单独使用
	Cacher Cacher

	// TTL 默认缓存过期时间
	// 如果为 0，使用默认值 5 分钟
	TTL time.Duration

	// MaxRows 最大缓存行数
	// 查询结果超过此行数时不缓存，防止大结果集占用过多内存
	// 如果为 0，使用默认值 1000
	MaxRows int64

	// Tables 需要缓存的表名列表（白名单模式）
	// 如果为空，则缓存所有表
	// 建议只缓存更新频率低、查询频率高的表
	Tables []string

	// ExcludeTables 排除缓存的表名列表（黑名单模式）
	// 如果设置了 Tables 白名单，此配置无效
	ExcludeTables []string

	// ExcludeDBs 排除缓存的数据库名列表（商户黑名单）
	// 在黑名单中的数据库不会启用缓存
	// 格式示例: []string{"shop1234567890", "shop9876543210"}
	ExcludeDBs []string

	// InvalidateOnWrite 写操作后是否自动失效缓存
	// 默认为 true
	InvalidateOnWrite *bool

	// Debug 是否开启调试日志
	Debug bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	invalidate := true
	return &Config{
		Easer:             true,
		TTL:               5 * time.Minute,
		MaxRows:           1000,
		InvalidateOnWrite: &invalidate,
		Debug:             false,
	}
}

// HitStats 缓存命中统计
type HitStats struct {
	TableName string // 表名
	Hits      int64  // 命中次数
	Misses    int64  // 未命中次数
	Eased     int64  // Easer 合并次数
}

// StatsCallback 统计回调函数
type StatsCallback func(stats HitStats)
