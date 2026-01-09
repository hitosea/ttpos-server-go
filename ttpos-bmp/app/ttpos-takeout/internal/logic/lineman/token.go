// Package lineman 提供 LINE MAN 平台集成服务
//
// 功能包括:
// 1. OAuth Token 生成与验证
// 2. 菜单同步到 Lineman 平台
//
// ⚠️ 临时方案说明:
// Token 管理为临时实现方案,参考 Grab OAuth 架构快速支持 LINE MAN 平台接入。
// 后续将迁移到统一的权限中心单点登录系统 (SSO)。
package lineman

import (
	"context"

	linemanClient "ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

var (
	jwtTokenClient   *linemanClient.JWTTokenClient
	oauthTokenClient *linemanClient.OAuthTokenClient
)

// getJWTTokenClient 获取 JWT Token 客户端（懒加载）
func (s *sLineman) getJWTTokenClient() *linemanClient.JWTTokenClient {
	if jwtTokenClient == nil {
		jwtTokenClient = linemanClient.NewJWTTokenClient()
	}
	return jwtTokenClient
}

// getOAuthTokenClient 获取 OAuth Token 客户端（懒加载）
func (s *sLineman) getOAuthTokenClient() *linemanClient.OAuthTokenClient {
	if oauthTokenClient == nil {
		oauthTokenClient = linemanClient.NewOAuthTokenClient()
	}
	return oauthTokenClient
}

// ============================================================================
// JWT Token 管理（委托给 client）
// ============================================================================

// GenerateToken 根据 client_id / client_secret 生成访问 Token
// 采用 JWT（HS256）实现
// 参数：
//   - ctx: 上下文
//   - clientID: LINE MAN 分配的 client_id
//   - clientSecret: LINE MAN 分配的 client_secret，用于校验身份
//
// 返回：
//   - token: 生成的 JWT Token
//   - expiresIn: Token 有效期（秒）
//   - err: 错误信息
func (s *sLineman) GenerateToken(ctx context.Context, clientID string, clientSecret string) (token string, expiresIn int, err error) {
	return s.getJWTTokenClient().GenerateToken(ctx, clientID, clientSecret)
}

// ParseToken 校验并解析 LINE MAN Token
// 参数：
//   - ctx: 上下文
//   - tokenStr: JWT Token 字符串
//
// 返回：
//   - claims: 解析后的 Claims
//   - err: 错误信息
func (s *sLineman) ParseToken(ctx context.Context, tokenStr string) (*lineman.LinemanTokenClaims, error) {
	return s.getJWTTokenClient().ParseToken(ctx, tokenStr)
}

// GetPartnerConfig 通过 partner code 获取配置
// 参数：
//   - ctx: 上下文
//   - code: Partner 代码
//
// 返回：
//   - partner: Partner 配置
//   - err: 错误信息
func (s *sLineman) GetPartnerConfig(ctx context.Context, code string) (*conf.LinemanPartner, error) {
	return s.getJWTTokenClient().GetPartnerConfig(ctx, code)
}

// GetPartnerConfigByClientID 通过 client_id 获取配置
// 参数：
//   - ctx: 上下文
//   - clientID: Client ID
//
// 返回：
//   - partner: Partner 配置
//   - err: 错误信息
func (s *sLineman) GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.LinemanPartner, error) {
	return s.getJWTTokenClient().GetPartnerConfigByClientID(ctx, clientID)
}

// ============================================================================
// OAuth Access Token 管理（委托给 client）
// ============================================================================

// FetchTokenFromAPI 从 LINE MAN OAuth 服务器获取 Access Token
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - token: OAuth Access Token
//   - expiresIn: Token 有效期（秒）
//   - err: 错误信息
func (s *sLineman) FetchTokenFromAPI(ctx context.Context) (string, int, error) {
	return s.getOAuthTokenClient().FetchTokenFromAPI(ctx)
}

// GetAccessToken 获取或刷新 Access Token（使用 Redis 缓存 + 双重检查锁）
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - token: OAuth Access Token
//   - err: 错误信息
func (s *sLineman) GetAccessToken(ctx context.Context) (string, error) {
	return s.getOAuthTokenClient().GetAccessToken(ctx)
}

// GetAuthorizationHeader 获取 Authorization 请求头（Bearer Token 格式）
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - header: Authorization 请求头值（格式: "Bearer {token}"）
//   - err: 错误信息
func (s *sLineman) GetAuthorizationHeader(ctx context.Context) (string, error) {
	return s.getOAuthTokenClient().GetAuthorizationHeader(ctx)
}
