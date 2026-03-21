package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// WebhookAuth ERPNext Webhook 认证中间件
// 验证请求头 X-Frappe-Webhook-Secret 与配置的 webhook secret 是否匹配
// 仅对 /callback/{siteCode}/* 路径生效，/callback/retry/* 跳过验证
func WebhookAuth(r *ghttp.Request) {
	// 内部管理接口不需要 webhook 认证
	if strings.HasPrefix(r.URL.Path, "/callback/retry") {
		r.Middleware.Next()
		return
	}

	ctx := r.Context()

	secret := g.Cfg().MustGet(ctx, "app.erpnext.webhook.secret").String()
	if secret == "" {
		// 未配置 secret 则跳过认证（开发环境）
		g.Log().Debug(ctx, "[webhook-auth] secret 未配置，跳过认证")
		r.Middleware.Next()
		return
	}

	headerSecret := r.Header.Get("X-Frappe-Webhook-Secret")
	if headerSecret == "" {
		g.Log().Warningf(ctx, "[webhook-auth] 缺少 X-Frappe-Webhook-Secret 头: path=%s, ip=%s",
			r.URL.Path, r.GetClientIp())
		r.Response.WriteJsonExit(g.Map{
			"code": "401",
			"msg":  "missing webhook secret",
		})
		return
	}

	// 恒定时间比较，防止时序攻击
	if subtle.ConstantTimeCompare([]byte(secret), []byte(headerSecret)) != 1 {
		g.Log().Warningf(ctx, "[webhook-auth] webhook secret 不匹配: path=%s, ip=%s",
			r.URL.Path, r.GetClientIp())
		r.Response.WriteJsonExit(g.Map{
			"code": "401",
			"msg":  "invalid webhook secret",
		})
		return
	}

	r.Middleware.Next()
}
