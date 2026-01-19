// Package health 健康检查逻辑层
package health

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

// RedisChecker Redis 健康检查器
type RedisChecker struct {
	redis   *gredis.Redis
	cluster bool
}

// NewRedisChecker 创建 Redis 检查器
func NewRedisChecker(redis *gredis.Redis, cluster bool) *RedisChecker {
	return &RedisChecker{
		redis:   redis,
		cluster: cluster,
	}
}

// Name 返回检查器名称
func (c *RedisChecker) Name() string {
	return "redis"
}

// Check 执行健康检查
func (c *RedisChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Status: StatusUp,
	}

	// 执行 PING 命令
	_, err := c.redis.Do(ctx, "PING")
	latency := time.Since(start).Milliseconds()
	result.Latency = latency

	if err != nil {
		result.Status = StatusDown
		result.Error = "Redis 连接失败"
		result.Detail = err.Error()
		g.Log().Errorf(ctx, "Redis 健康检查失败: %v, 耗时: %dms", err, latency)
		return result
	}

	// 如果是集群模式，额外检查 CLUSTER INFO
	if c.cluster {
		_, err = c.redis.Do(ctx, "CLUSTER", "INFO")
		if err != nil {
			result.Status = StatusDown
			result.Error = "Redis 集群状态异常"
			result.Detail = err.Error()
			g.Log().Errorf(ctx, "Redis 集群健康检查失败: %v, 耗时: %dms", err, latency)
			return result
		}
	}

	g.Log().Debugf(ctx, "Redis 健康检查成功, 耗时: %dms", latency)
	return result
}
