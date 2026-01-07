// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

type (
	ILinemanToken interface {
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
		GenerateToken(ctx context.Context, clientID string, clientSecret string) (token string, expiresIn int, err error)
		// ParseToken 校验并解析 LINE MAN Token
		// 参数：
		//   - ctx: 上下文
		//   - tokenStr: JWT Token 字符串
		//
		// 返回：
		//   - claims: 解析后的 Claims
		//   - err: 错误信息
		ParseToken(ctx context.Context, tokenStr string) (*lineman.LinemanTokenClaims, error)
		// GetPartnerConfig 通过 partner code 获取配置
		// 参数：
		//   - ctx: 上下文
		//   - code: Partner 代码
		//
		// 返回：
		//   - partner: Partner 配置
		//   - err: 错误信息
		GetPartnerConfig(ctx context.Context, code string) (*conf.LinemanPartner, error)
		// GetPartnerConfigByClientID 通过 client_id 获取配置
		// 参数：
		//   - ctx: 上下文
		//   - clientID: Client ID
		//
		// 返回：
		//   - partner: Partner 配置
		//   - err: 错误信息
		GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.LinemanPartner, error)
		// FetchTokenFromAPI 从 LINE MAN OAuth 服务器获取 Access Token
		// 参数：
		//   - ctx: 上下文
		//
		// 返回：
		//   - token: OAuth Access Token
		//   - expiresIn: Token 有效期（秒）
		//   - err: 错误信息
		FetchTokenFromAPI(ctx context.Context) (string, int, error)
		// GetAccessToken 获取或刷新 Access Token（使用 Redis 缓存 + 双重检查锁）
		// 参数：
		//   - ctx: 上下文
		//
		// 返回：
		//   - token: OAuth Access Token
		//   - err: 错误信息
		GetAccessToken(ctx context.Context) (string, error)
		// GetAuthorizationHeader 获取 Authorization 请求头（Bearer Token 格式）
		// 参数：
		//   - ctx: 上下文
		//
		// 返回：
		//   - header: Authorization 请求头值（格式: "Bearer {token}"）
		//   - err: 错误信息
		GetAuthorizationHeader(ctx context.Context) (string, error)
	}
)

var (
	localLinemanToken ILinemanToken
)

func LinemanToken() ILinemanToken {
	if localLinemanToken == nil {
		panic("implement not found for interface ILinemanToken, forgot register?")
	}
	return localLinemanToken
}

func RegisterLinemanToken(i ILinemanToken) {
	localLinemanToken = i
}
