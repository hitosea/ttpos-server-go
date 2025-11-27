// Package health 健康检查逻辑层
package health

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MySQLChecker MySQL 健康检查器
type MySQLChecker struct {
	db    gdb.DB
	query string
}

// NewMySQLChecker 创建 MySQL 检查器
func NewMySQLChecker(db gdb.DB, query string) *MySQLChecker {
	if query == "" {
		query = "SELECT 1"
	}
	return &MySQLChecker{
		db:    db,
		query: query,
	}
}

// Name 返回检查器名称
func (c *MySQLChecker) Name() string {
	return "mysql"
}

// Check 执行健康检查
func (c *MySQLChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Status: StatusUp,
	}

	// 执行 Ping 查询
	_, err := c.db.Ctx(ctx).Query(ctx, c.query)
	latency := time.Since(start).Milliseconds()
	result.Latency = latency

	if err != nil {
		result.Status = StatusDown
		result.Error = "MySQL 连接失败"
		result.Detail = err.Error()
		g.Log().Errorf(ctx, "MySQL 健康检查失败: %v, 耗时: %dms", err, latency)
		return result
	}

	g.Log().Debugf(ctx, "MySQL 健康检查成功, 耗时: %dms", latency)
	return result
}
