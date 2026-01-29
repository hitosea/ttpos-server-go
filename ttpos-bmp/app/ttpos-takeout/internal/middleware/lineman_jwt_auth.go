package middleware

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

const (
	// ContextKeyLinemanPartnerClaims 用于在 Context 中存储 LINE MAN Partner Token Claims
	ContextKeyLinemanPartnerClaims = "LinemanPartnerClaims"
)

var (
	linemanSecretKey string
	linemanMu        sync.RWMutex
)

// MiddlewareLinemanJWTAuth LINE MAN Partner API JWT Token 认证中间件
// 从 Authorization 头中提取 Bearer Token，验证后将 Claims 存入 Context
func MiddlewareLinemanJWTAuth(r *ghttp.Request) {
	ctx := r.Context()

	// OAuth Token 颁发接口不做 JWT 校验
	if strings.HasSuffix(r.URL.Path, "/oauth2/token") {
		r.Middleware.Next()
		return
	}

	// 提取 Authorization 头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.Log().Warning(ctx, "[lineman-jwt-auth] Authorization header missing")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Authorization header is required",
		})
		return
	}

	// 验证 Bearer Token 格式
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		g.Log().Warning(ctx, "[lineman-jwt-auth] Invalid Authorization format")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Invalid Authorization header format. Expected: Bearer <token>",
		})
		return
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		g.Log().Warning(ctx, "[lineman-jwt-auth] Token is empty")
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Token is required",
		})
		return
	}

	// 调用本地函数验证 Token
	claims, err := parseLinemanToken(token)
	if err != nil {
		g.Log().Warningf(ctx, "[lineman-jwt-auth] Token validation failed: %v. Token preview: %s...", err, getTokenPreview(token))
		r.Response.WriteJsonExit(g.Map{
			"error":             "unauthorized",
			"error_description": "Invalid or expired token",
		})
		return
	}

	// 将 Claims 存入 Context
	r.SetCtxVar(ContextKeyLinemanPartnerClaims, claims)

	g.Log().Debugf(ctx, "[lineman-jwt-auth] Token validated successfully for client_id=%s, scope=%s",
		claims.ClientID, claims.Scope)

	// 继续处理请求
	r.Middleware.Next()
}

// getLinemanSecretKey 获取 LINE MAN 密钥（懒加载）
func getLinemanSecretKey(ctx context.Context) string {
	linemanMu.RLock()
	if linemanSecretKey != "" {
		defer linemanMu.RUnlock()
		return linemanSecretKey
	}
	linemanMu.RUnlock()

	linemanMu.Lock()
	defer linemanMu.Unlock()
	if linemanSecretKey == "" {
		var linemanCfg conf.Lineman
		if err := g.Cfg().MustGet(ctx, "app.provider.lineman.platform").Scan(&linemanCfg); err != nil {
			g.Log().Fatal(ctx, err)
		}
		linemanSecretKey = linemanCfg.SecretKey
	}
	return linemanSecretKey
}

// parseLinemanToken 校验并解析 LINE MAN Partner Token
func parseLinemanToken(tokenStr string) (*lineman.PartnerTokenClaims, error) {
	// 清理 token 字符串
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, gerror.New("Token 不能为空")
	}

	// 验证 JWT Token 格式：应该包含三个部分，用点号分隔
	tokenParts := strings.Split(tokenStr, ".")
	if len(tokenParts) != 3 {
		return nil, gerror.Newf("Token 格式错误: 期望 3 个部分，实际 %d 个部分", len(tokenParts))
	}

	secretKey := getLinemanSecretKey(context.Background())
	token, err := jwt.ParseWithClaims(tokenStr, &lineman.PartnerTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, gerror.Newf("不支持的签名方法: %s", t.Method.Alg())
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, gerror.Wrap(err, "Token 解析失败")
	}
	claims, ok := token.Claims.(*lineman.PartnerTokenClaims)
	if !ok || !token.Valid {
		return nil, gerror.New("Token 无效")
	}
	return claims, nil
}
