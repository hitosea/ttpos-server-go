package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 窗口时间（秒）
	WindowSeconds int
	// 最大请求数
	MaxRequests int
	// 超出限制时的响应状态码
	StatusCode int
	// 超出限制时的响应消息
	Message string
	// 是否启用
	Enabled bool
}

// DefaultRateLimits 默认速率限制配置
// 可通过调用 SetConfig 方法来自定义特定路径的限制规则
var DefaultRateLimits = map[string]RateLimitConfig{
	// 登录接口 - 每分钟最多5次 (防止暴力破解)
	"/api/v1/cashier/passport/login": {
		WindowSeconds: 60,
		MaxRequests:   5,
		StatusCode:    http.StatusTooManyRequests,
		Message:       "登录请求过于频繁，请1分钟后再试",
		Enabled:       true,
	},
	"/api/v1/passport/login": {
		WindowSeconds: 60,
		MaxRequests:   5,
		StatusCode:    http.StatusTooManyRequests,
		Message:       "登录请求过于频繁，请1分钟后再试",
		Enabled:       true,
	},
	// 获取验证码 - 每分钟最多5次 (防止短信轰炸)
	"/api/v1/passport/captcha": {
		WindowSeconds: 60,
		MaxRequests:   20,
		StatusCode:    http.StatusTooManyRequests,
		Message:       "验证码获取过于频繁，请1分钟后再试",
		Enabled:       true,
	},
	// 获取服务器公钥 - 每分钟最多10次 (防止密钥枚举)
	"/api/v1/passport/server_public_key": {
		WindowSeconds: 60,
		MaxRequests:   20,
		StatusCode:    http.StatusTooManyRequests,
		Message:       "请求过于频繁，请稍后再试",
		Enabled:       true,
	},
	// 默认限制 - 每分钟最多30次 (防止API滥用)
	"default": {
		WindowSeconds: 60,
		MaxRequests:   60,
		StatusCode:    http.StatusTooManyRequests,
		Message:       "请求过于频繁，请稍后再试",
		Enabled:       true,
	},
}

// RequestRecord 请求记录
type RequestRecord struct {
	Count       int
	WindowStart time.Time
}

// RateLimiter 速率限制器
type RateLimiter struct {
	mu      sync.RWMutex
	records map[string]*RequestRecord
	configs map[string]RateLimitConfig
	logger  *zap.Logger
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		records: make(map[string]*RequestRecord),
		configs: DefaultRateLimits,
		logger:  logger,
	}
}

// SetConfig 设置特定路径的速率限制配置
func (rl *RateLimiter) SetConfig(path string, config RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.configs[path] = config
}

// GetConfig 获取路径的速率限制配置
func (rl *RateLimiter) GetConfig(path string) RateLimitConfig {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if config, exists := rl.configs[path]; exists {
		return config
	}
	return rl.configs["default"]
}

// IsAllowed 检查请求是否被允许
func (rl *RateLimiter) IsAllowed(clientIP, path string) bool {
	config := rl.GetConfig(path)
	if !config.Enabled {
		return true
	}

	key := fmt.Sprintf("%s:%s", clientIP, path)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	record, exists := rl.records[key]
	now := time.Now()

	if !exists || now.Sub(record.WindowStart) >= time.Duration(config.WindowSeconds)*time.Second {
		// 新的时间窗口
		rl.records[key] = &RequestRecord{
			Count:       1,
			WindowStart: now,
		}
		return true
	}

	// 检查是否超过限制
	if record.Count >= config.MaxRequests {
		return false
	}

	// 增加计数
	record.Count++
	return true
}

// Cleanup 清理过期的记录（可选的维护方法）
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, record := range rl.records {
		// 如果记录超过最长窗口时间的2倍，认为已过期
		maxWindow := 120 * time.Second // 2分钟
		for _, config := range rl.configs {
			if time.Duration(config.WindowSeconds)*time.Second > maxWindow {
				maxWindow = time.Duration(config.WindowSeconds) * time.Second
			}
		}

		if now.Sub(record.WindowStart) > maxWindow*2 {
			delete(rl.records, key)
		}
	}
}

// Middleware 创建Gin中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		if !rl.IsAllowed(clientIP, path) {
			config := rl.GetConfig(path)

			// 记录被限制的请求
			rl.logger.Warn("请求被速率限制",
				zap.String("ip", clientIP),
				zap.String("path", path),
				zap.String("method", c.Request.Method),
				zap.String("user_agent", c.Request.UserAgent()))

			c.AbortWithStatusJSON(config.StatusCode, gin.H{
				"code":    config.StatusCode,
				"message": config.Message,
			})
			return
		}

		c.Next()
	}
}

// GlobalRateLimiter 全局速率限制器实例
var GlobalRateLimiter *RateLimiter

// InitGlobalRateLimiter 初始化全局速率限制器
func InitGlobalRateLimiter(logger *zap.Logger) {
	GlobalRateLimiter = NewRateLimiter(logger)

	// 启动清理协程，每5分钟清理一次过期记录
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			GlobalRateLimiter.Cleanup()
		}
	}()
}

// RateLimit 返回全局速率限制中间件
func RateLimit() gin.HandlerFunc {
	if GlobalRateLimiter == nil {
		// 如果未初始化，使用默认配置但不进行限制
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return GlobalRateLimiter.Middleware()
}
