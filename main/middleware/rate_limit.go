package middleware

import (
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
)

// RateLimiterByKey 基于 key 的限流器管理
type RateLimiterByKey struct {
	limiters sync.Map   // key -> *rate.Limiter
	rate     rate.Limit // 每秒允许的请求数
	burst    int        // 突发容量
}

// NewRateLimiterByKey 创建基于 key 的限流器
func NewRateLimiterByKey(ratePerSecond float64, burst int) *RateLimiterByKey {
	return &RateLimiterByKey{
		rate:  rate.Limit(ratePerSecond),
		burst: burst,
	}
}

// GetLimiter 获取指定 key 的限流器
func (r *RateLimiterByKey) GetLimiter(key string) *rate.Limiter {
	if limiter, ok := r.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(r.rate, r.burst)
	actual, _ := r.limiters.LoadOrStore(key, limiter)
	return actual.(*rate.Limiter)
}

// Allow 检查指定 key 是否允许请求
func (r *RateLimiterByKey) Allow(key string) bool {
	return r.GetLimiter(key).Allow()
}

// RateLimitByCompanyUuid 基于 company_uuid 的限流中间件
// ratePerSecond: 每秒允许的请求数
// burst: 突发容量（允许的最大突发请求数）
func RateLimitByCompanyUuid(ratePerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiterByKey(ratePerSecond, burst)

	return func(c *gin.Context) {
		// 从 query 参数获取 company_uuid
		companyUuidStr := c.Query("company_uuid")
		if companyUuidStr == "" {
			// 如果没有 company_uuid，使用 IP 作为 fallback
			companyUuidStr = c.ClientIP()
		}

		if !limiter.Allow(companyUuidStr) {
			helper.Fail(c, constant.CodeFail, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByIP 基于 IP 的限流中间件
func RateLimitByIP(ratePerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiterByKey(ratePerSecond, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			helper.Fail(c, constant.CodeFail, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByQueryParam 基于指定 query 参数的限流中间件
func RateLimitByQueryParam(paramName string, ratePerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiterByKey(ratePerSecond, burst)

	return func(c *gin.Context) {
		key := c.Query(paramName)
		if key == "" {
			key = c.ClientIP() // fallback to IP
		}

		if !limiter.Allow(key) {
			helper.Fail(c, constant.CodeFail, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitGlobal 全局限流中间件
func RateLimitGlobal(ratePerSecond float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			helper.Fail(c, constant.CodeFail, "服务繁忙，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CompanyUuidFromQuery 从 query 参数提取 company_uuid 作为限流 key
func CompanyUuidFromQuery(c *gin.Context) string {
	if uuid := c.Query("company_uuid"); uuid != "" {
		return uuid
	}
	return c.ClientIP()
}

// CompanyUuidFromUint64 将 uint64 转换为字符串作为限流 key
func CompanyUuidFromUint64(uuid uint64) string {
	return strconv.FormatUint(uuid, 10)
}
