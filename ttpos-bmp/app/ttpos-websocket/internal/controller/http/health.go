// Package http HTTP 控制器层
package http

import (
	"net"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"ttpos-bmp/app/ttpos-websocket/internal/logic/health"
)

// HealthCheck 健康检查处理器
func HealthCheck(r *ghttp.Request) {
	ctx := r.Context()

	// 获取健康检查服务
	svc := health.Instance()
	cfg := svc.GetConfig()

	// Token 校验（如果配置了）
	if cfg.Token != "" {
		token := r.Header.Get("X-Health-Token")
		if token != cfg.Token {
			g.Log().Warningf(ctx, "健康检查 Token 校验失败, IP: %s", r.GetClientIp())
			r.Response.WriteStatusExit(http.StatusUnauthorized)
			return
		}
	}

	// IP 白名单校验（如果配置了）
	if len(cfg.AllowedIPs) > 0 {
		clientIP := r.GetClientIp()
		if !isIPAllowed(clientIP, cfg.AllowedIPs) {
			g.Log().Warningf(ctx, "健康检查 IP 校验失败, IP: %s", clientIP)
			r.Response.WriteStatusExit(http.StatusForbidden)
			return
		}
	}

	// 解析查询参数
	opts := health.CheckOptions{
		Detail: r.Get("detail", true).Bool(),
		Scope:  r.Get("scope", "").String(),
	}

	// 执行健康检查
	result := svc.Check(ctx, opts)

	// 设置响应状态码
	statusCode := http.StatusOK
	if result.Status == health.StatusDown {
		statusCode = http.StatusServiceUnavailable
		r.Response.Header().Set("Retry-After", "10")
	}

	// 手动序列化 JSON 并写入响应
	jsonBytes, err := gjson.Marshal(result)
	if err != nil {
		g.Log().Error(ctx, "序列化健康检查结果失败", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.Write("{\"error\":\"internal server error\"}")
		return
	}

	// 设置响应头和状态码，然后写入 JSON
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Status = statusCode
	r.Response.Write(jsonBytes)
}

// isIPAllowed 检查 IP 是否在白名单中
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, allowed := range allowedIPs {
		// 支持 CIDR 格式
		if strings.Contains(allowed, "/") {
			_, ipnet, err := net.ParseCIDR(allowed)
			if err != nil {
				continue
			}
			if ipnet.Contains(ip) {
				return true
			}
		} else {
			// 精确匹配
			if clientIP == allowed {
				return true
			}
		}
	}

	return false
}
