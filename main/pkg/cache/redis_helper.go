package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// GetRedisKeys 获取匹配模式的所有 Redis keys，支持集群和单机模式
// 参数:
//   - ctx: 上下文
//   - client: Redis客户端（UniversalClient 类型，兼容单机和集群）
//   - pattern: 匹配模式，如 "prefix:*"
//
// 返回:
//   - []string: 所有匹配的 key 列表
//   - error: 错误信息
func GetRedisKeys(ctx context.Context, client redis.UniversalClient, pattern string) ([]string, error) {
	var allKeys []string

	// 判断是否为集群模式
	if clusterClient, ok := client.(*redis.ClusterClient); ok {
		// 集群模式：需要对每个分片节点执行 Keys 命令
		err := clusterClient.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			keys, err := shard.Keys(ctx, pattern).Result()
			if err != nil {
				return err
			}
			allKeys = append(allKeys, keys...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 单机模式：直接执行 Keys 命令
		keys, err := client.Keys(ctx, pattern).Result()
		if err != nil {
			return nil, err
		}
		allKeys = keys
	}

	return allKeys, nil
}

const (
	// DefaultScanCount 默认的 Scan 命令 count 值
	// 适用于大多数场景，平衡性能和内存占用
	DefaultScanCount int64 = 1000
)

// ScanRedisKeys 使用 Scan 命令获取匹配模式的所有 Redis keys，支持集群和单机模式
// Scan 命令比 Keys 更安全，不会阻塞 Redis，推荐在生产环境使用
// 参数:
//   - ctx: 上下文
//   - client: Redis客户端（UniversalClient 类型，兼容单机和集群）
//   - pattern: 匹配模式，如 "prefix:*"
//   - count: 每次扫描的数量建议值（实际返回数量可能不同）
//     如果 count <= 0，将使用默认值 DefaultScanCount (1000)
//     建议值范围：
//   - 小规模（预计 < 1000 个key）：100
//   - 中规模（预计 1000-10000 个key）：1000
//   - 大规模（预计 > 10000 个key）：5000-10000
//
// 返回:
//   - []string: 所有匹配的 key 列表
//   - error: 错误信息
func ScanRedisKeys(ctx context.Context, client redis.UniversalClient, pattern string, count int64) ([]string, error) {
	// 如果 count <= 0，使用默认值
	if count <= 0 {
		count = DefaultScanCount
	}
	var allKeys []string

	// 判断是否为集群模式
	if clusterClient, ok := client.(*redis.ClusterClient); ok {
		// 集群模式：需要对每个分片节点执行 Scan 命令
		err := clusterClient.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			var cursor uint64
			for {
				var scanKeys []string
				var err error
				scanKeys, cursor, err = shard.Scan(ctx, cursor, pattern, count).Result()
				if err != nil {
					return err
				}

				allKeys = append(allKeys, scanKeys...)

				if cursor == 0 {
					break
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 单机模式：直接执行 Scan 命令
		var cursor uint64
		for {
			var scanKeys []string
			var err error
			scanKeys, cursor, err = client.Scan(ctx, cursor, pattern, count).Result()
			if err != nil {
				return nil, err
			}

			allKeys = append(allKeys, scanKeys...)

			if cursor == 0 {
				break
			}
		}
	}

	return allKeys, nil
}

// ScanRedisKeysDefault 使用默认 count 值扫描 Redis keys
// 这是 ScanRedisKeys 的简化版本，使用默认的 count 值 (1000)
// 适用于不确定 key 数量的场景
// 参数:
//   - ctx: 上下文
//   - client: Redis客户端（UniversalClient 类型，兼容单机和集群）
//   - pattern: 匹配模式，如 "prefix:*"
//
// 返回:
//   - []string: 所有匹配的 key 列表
//   - error: 错误信息
func ScanRedisKeysDefault(ctx context.Context, client redis.UniversalClient, pattern string) ([]string, error) {
	return ScanRedisKeys(ctx, client, pattern, DefaultScanCount)
}
