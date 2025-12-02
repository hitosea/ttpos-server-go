// Package health 健康检查逻辑层
package health

import (
	"context"
	"time"
)

// Checker 依赖检查器接口
type Checker interface {
	// Name 返回检查器名称
	Name() string
	// Check 执行健康检查
	Check(ctx context.Context) CheckResult
}

// CheckResult 检查结果
type CheckResult struct {
	Status  string `json:"status"`           // "UP" 或 "DOWN"
	Latency int64  `json:"latency_ms"`       // 检查耗时（毫秒）
	Error   string `json:"error,omitempty"`  // 失败时的简短描述
	Detail  string `json:"detail,omitempty"` // 可选的详细堆栈/错误码
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     string                 `json:"status"`               // 总体状态："UP" 或 "DOWN"
	Timestamp  int64                  `json:"timestamp"`            // 时间戳（毫秒）
	Components map[string]CheckResult `json:"components,omitempty"` // 各组件状态
	Meta       map[string]interface{} `json:"meta,omitempty"`       // 元数据
}

// Config 健康检查配置
type Config struct {
	Enable       bool               `json:"enable"`
	Path         string             `json:"path"`
	Token        string             `json:"token"`
	AllowedIPs   []string           `json:"allowedIps"`
	CacheTTL     time.Duration      `json:"cacheTtl"`
	Timeout      time.Duration      `json:"timeout"`
	Dependencies DependenciesConfig `json:"dependencies"`
	Degrade      DegradeConfig      `json:"degrade"`
}

// DependenciesConfig 依赖检测配置
type DependenciesConfig struct {
	MySQL    MySQLConfig    `json:"mysql"`
	Redis    RedisConfig    `json:"redis"`
	RocketMQ RocketMQConfig `json:"rocketmq"`
	Nacos    NacosConfig    `json:"nacos"`
}

// MySQLConfig MySQL 检测配置
type MySQLConfig struct {
	Enabled   bool   `json:"enabled"`
	PingQuery string `json:"pingQuery"`
}

// RedisConfig Redis 检测配置
type RedisConfig struct {
	Enabled bool `json:"enabled"`
	Cluster bool `json:"cluster"`
}

// RocketMQConfig RocketMQ 检测配置
type RocketMQConfig struct {
	Enabled bool `json:"enabled"`
}

// NacosConfig Nacos 检测配置
type NacosConfig struct {
	Enabled bool `json:"enabled"`
}

// DegradeConfig 降级配置
type DegradeConfig struct {
	PausePush bool `json:"pausePush"`
}

// CheckOptions 检查选项
type CheckOptions struct {
	Detail bool   `json:"detail"` // 是否返回详细信息
	Scope  string `json:"scope"`  // liveness/readiness/startup
}

const (
	// StatusUp 健康状态
	StatusUp = "UP"
	// StatusDown 不健康状态
	StatusDown = "DOWN"
	// StatusSkip 跳过检查
	StatusSkip = "SKIP"
)
