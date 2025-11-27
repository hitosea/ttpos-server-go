// Package health 健康检查逻辑层
package health

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// Service 健康检查服务
type Service struct {
	checkers map[string]Checker
	config   *Config
	cache    *ResultCache
	mu       sync.RWMutex
}

var (
	instance *Service
	once     sync.Once
)

// Instance 获取健康检查服务单例
func Instance() *Service {
	once.Do(func() {
		instance = &Service{
			checkers: make(map[string]Checker),
		}
		instance.init()
	})
	return instance
}

// init 初始化健康检查服务
func (s *Service) init() {
	ctx := gctx.GetInitCtx()

	// 加载配置
	s.config = s.loadConfig(ctx)

	// 初始化缓存
	s.cache = NewResultCache(s.config.CacheTTL)

	// 注册检查器
	s.registerCheckers(ctx)

	g.Log().Info(ctx, "健康检查服务初始化完成")
}

// loadConfig 加载配置
func (s *Service) loadConfig(ctx context.Context) *Config {
	cfg := &Config{
		Enable:     g.Cfg().MustGet(ctx, "health.enable", true).Bool(),
		Path:       g.Cfg().MustGet(ctx, "health.path", "/health").String(),
		Token:      g.Cfg().MustGet(ctx, "health.token", "").String(),
		AllowedIPs: g.Cfg().MustGet(ctx, "health.allowedIps", []string{}).Strings(),
		CacheTTL:   g.Cfg().MustGet(ctx, "health.cacheTtl", "2s").Duration(),
		Timeout:    g.Cfg().MustGet(ctx, "health.timeout", "500ms").Duration(),
	}

	// 加载依赖配置
	cfg.Dependencies.MySQL.Enabled = g.Cfg().MustGet(ctx, "health.dependencies.mysql.enabled", true).Bool()
	cfg.Dependencies.MySQL.PingQuery = g.Cfg().MustGet(ctx, "health.dependencies.mysql.pingQuery", "SELECT 1").String()

	cfg.Dependencies.Redis.Enabled = g.Cfg().MustGet(ctx, "health.dependencies.redis.enabled", true).Bool()
	cfg.Dependencies.Redis.Cluster = g.Cfg().MustGet(ctx, "health.dependencies.redis.cluster", true).Bool()

	cfg.Dependencies.RocketMQ.Enabled = g.Cfg().MustGet(ctx, "health.dependencies.rocketmq.enabled", false).Bool()
	cfg.Dependencies.Nacos.Enabled = g.Cfg().MustGet(ctx, "health.dependencies.nacos.enabled", false).Bool()

	// 加载降级配置
	cfg.Degrade.PausePush = g.Cfg().MustGet(ctx, "health.degrade.pausePush", false).Bool()

	return cfg
}

// registerCheckers 注册检查器
func (s *Service) registerCheckers(ctx context.Context) {
	// 注册 MySQL 检查器
	if s.config.Dependencies.MySQL.Enabled {
		db := g.DB()
		if db != nil {
			checker := NewMySQLChecker(db, s.config.Dependencies.MySQL.PingQuery)
			s.RegisterChecker(checker)
			g.Log().Debug(ctx, "已注册 MySQL 健康检查器")
		}
	}

	// 注册 Redis 检查器
	if s.config.Dependencies.Redis.Enabled {
		redis := g.Redis()
		if redis != nil {
			checker := NewRedisChecker(redis, s.config.Dependencies.Redis.Cluster)
			s.RegisterChecker(checker)
			g.Log().Debug(ctx, "已注册 Redis 健康检查器")
		}
	}

	// 预留：RocketMQ 和 Nacos 检查器
	if s.config.Dependencies.RocketMQ.Enabled {
		g.Log().Info(ctx, "RocketMQ 健康检查器尚未实现，跳过")
	}
	if s.config.Dependencies.Nacos.Enabled {
		g.Log().Info(ctx, "Nacos 健康检查器尚未实现，跳过")
	}
}

// RegisterChecker 注册检查器
func (s *Service) RegisterChecker(checker Checker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers[checker.Name()] = checker
}

// Check 执行健康检查
func (s *Service) Check(ctx context.Context, opts CheckOptions) *HealthResponse {
	// 检查缓存
	if cached, ok := s.cache.Get(); ok && !opts.Detail {
		return cached
	}

	// 执行检查
	response := &HealthResponse{
		Status:     StatusUp,
		Timestamp:  time.Now().UnixMilli(),
		Components: make(map[string]CheckResult),
		Meta:       make(map[string]interface{}),
	}

	// 创建超时上下文
	checkCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// 并发执行所有检查器
	var wg sync.WaitGroup
	var mu sync.Mutex

	s.mu.RLock()
	checkers := make([]Checker, 0, len(s.checkers))
	for _, checker := range s.checkers {
		checkers = append(checkers, checker)
	}
	s.mu.RUnlock()

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			result := c.Check(checkCtx)

			mu.Lock()
			response.Components[c.Name()] = result
			// 任一组件 DOWN，整体状态为 DOWN
			if result.Status == StatusDown {
				response.Status = StatusDown
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	// 如果不需要详细信息，简化响应
	if !opts.Detail {
		for name := range response.Components {
			comp := response.Components[name]
			comp.Detail = ""
			response.Components[name] = comp
		}
	}

	// 添加元数据
	response.Meta["cache_hit"] = false
	response.Meta["checker_count"] = len(response.Components)

	// 缓存结果
	s.cache.Set(response)

	// 记录日志
	if response.Status == StatusDown {
		g.Log().Warningf(ctx, "健康检查失败: %+v", response)
	} else {
		g.Log().Debugf(ctx, "健康检查成功: %+v", response)
	}

	// 记录指标
	RecordMetrics(ctx, response)

	return response
}

// GetConfig 获取配置
func (s *Service) GetConfig() *Config {
	return s.config
}

// ClearCache 清空缓存
func (s *Service) ClearCache() {
	s.cache.Clear()
}
