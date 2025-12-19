package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

const (
	// ContextKeyGrabPartnerClaims 用于在 Context 中存储 Grab Partner Token Claims
	ContextKeyGrabPartnerClaims = "GrabPartnerClaims"
)

// MiddlewareGrabJWTAuth Grab Partner API JWT Token 认证中间件
// 从 Authorization 头中提取 Bearer Token，验证后将 Claims 存入 Context
func MiddlewareGrabJWTAuth(r *ghttp.Request) {
	ctx := r.Context()

	// OAuth Token 颁发接口不做 JWT 校验（兼容 /oauth/partner 与 /oauth/partner/webhook）
	if strings.HasSuffix(r.URL.Path, "/oauth/partner") {
		r.Middleware.Next()
		return
	}

	// 提取 Authorization 头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.Log().Warning(ctx, "[grab-jwt-auth] Authorization header missing")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Authorization header is required",
		})
		return
	}

	// 验证 Bearer Token 格式
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		g.Log().Warning(ctx, "[grab-jwt-auth] Invalid Authorization format")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Invalid Authorization header format. Expected: Bearer <token>",
		})
		return
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		g.Log().Warning(ctx, "[grab-jwt-auth] Token is empty")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Token is required",
		})
		return
	}

	// 验证 JWT Token 格式：应该包含三个部分，用点号分隔
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		g.Log().Warningf(ctx, "[grab-jwt-auth] Invalid token format: expected 3 segments, got %d. Token preview: %s...", len(tokenParts), getTokenPreview(token))
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Invalid token format",
		})
		return
	}

	// 调用 Grab Service 验证 Token
	claims, err := service.Grab().ParsePartnerToken(token)
	if err != nil {
		g.Log().Warningf(ctx, "[grab-jwt-auth] Token validation failed: %v. Token preview: %s...", err, getTokenPreview(token))
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Invalid or expired token",
		})
		return
	}

	// 将 Claims 存入 Context
	r.SetCtxVar(ContextKeyGrabPartnerClaims, claims)

	g.Log().Debugf(ctx, "[grab-jwt-auth] Token validated successfully for client_id=%s, scope=%s",
		claims.ClientID, claims.Scope)

	// 继续处理请求
	r.Middleware.Next()
}

func MiddlewareDirectResponse(r *ghttp.Request) {
	r.Middleware.Next()

	// There's custom buffer content, it then exits current handler.
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}
	var (
		msg  string
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		msg = err.Error()
		// 根据错误码设置 HTTP status
		r.Response.Status = mapCodeToHTTPStatus(code)
		r.Response.WriteJsonExit(g.Map{
			"code":    code,
			"message": msg,
		})
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden:
				code = gcode.CodeNotAuthorized
			default:
				code = gcode.CodeUnknown
			}
			// It creates an error as it can be retrieved by other middlewares.
			err = gerror.NewCode(code, msg)
			r.SetError(err)
		} else {
			code = gcode.CodeOK
		}
		msg = code.Message()
	}

	r.Response.WriteJson(res)
}

// mapCodeToHTTPStatus 将 gcode.Code 映射到 HTTP status code
func mapCodeToHTTPStatus(code gcode.Code) int {
	switch code {
	// 4xx 客户端错误
	case gcode.CodeInvalidParameter, gcode.CodeValidationFailed:
		return http.StatusBadRequest // 400
	case gcode.CodeNotFound:
		return http.StatusNotFound // 404
	case gcode.CodeNotAuthorized:
		return http.StatusUnauthorized // 401

	// 5xx 服务器错误
	case gcode.CodeInternalError:
		return http.StatusInternalServerError // 500
	case gcode.CodeNotImplemented:
		return http.StatusNotImplemented // 501
	case gcode.CodeUnknown, gcode.CodeNil:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}

// getTokenPreview 返回 token 的前缀预览（用于日志，不暴露完整 token）
func getTokenPreview(token string) string {
	if len(token) == 0 {
		return ""
	}
	previewLen := 20
	if len(token) < previewLen {
		previewLen = len(token)
	}
	return token[:previewLen]
}
