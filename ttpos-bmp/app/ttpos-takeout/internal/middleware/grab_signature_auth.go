package middleware

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"ttpos-bmp/app/ttpos-takeout/internal/logic/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// MiddlewareGrabSignatureAuth Grab Webhook HMAC-SHA256 签名验证中间件
// 从请求头获取 X-Grab-Signature 和 X-Grab-Timestamp，验证签名有效性
// 仅对 /partner/* 路径生效，OAuth 路径跳过验证
func MiddlewareGrabSignatureAuth(r *ghttp.Request) {
	ctx := r.Context()

	// OAuth 接口不做签名验证
	if strings.Contains(r.URL.Path, "/oauth/partner") {
		r.Middleware.Next()
		return
	}

	// 只对 /partner/* 路径做签名验证
	if !strings.Contains(r.URL.Path, "/partner") {
		r.Middleware.Next()
		return
	}

	// 1. 从 Header 获取签名和时间戳
	signature := r.Header.Get(grab.HeaderXGrabSignature)
	timestamp := r.Header.Get(grab.HeaderXGrabTimestamp)

	// 2. 获取请求体（GoFrame 自动缓存，后续 handler 可再次读取）
	body := r.GetBody()

	// 3. 调用 service 验证签名
	if err := service.Grab().VerifyWebhookSignature(ctx, signature, timestamp, body); err != nil {
		g.Log().Warningf(ctx, "[grab-signature-auth] Signature verification failed: %v", err)

		// 根据错误类型返回不同的错误信息
		errorMsg := "Signature verification failed"
		if errors.Is(err, grab.ErrMissingHeader) {
			errorMsg = "Missing required header"
		} else if errors.Is(err, grab.ErrExpiredTimestamp) {
			errorMsg = "Timestamp expired"
		} else if errors.Is(err, grab.ErrInvalidSignature) {
			errorMsg = "Invalid signature"
		}

		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": errorMsg,
		})
		return
	}

	g.Log().Debugf(ctx, "[grab-signature-auth] Signature verified successfully")

	// 继续处理请求
	r.Middleware.Next()
}
