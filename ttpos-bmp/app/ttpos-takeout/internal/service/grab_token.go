// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

type (
	IGrabToken interface {
		// GeneratePartnerToken 根据 client_id / client_secret 生成访问 Token
		// 采用 JWT（HS256）实现
		// 参数：
		//   - ctx: 上下文
		//   - clientID: Grab 分配的 client_id
		//   - clientSecret: Grab 分配的 client_secret，用于校验身份
		//   - scope: 请求的权限范围
		//
		// 返回：
		//   - token: 生成的 JWT Token
		//   - expiresIn: Token 有效期（秒）
		//   - err: 错误信息
		GeneratePartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (token string, expiresIn int, err error)
		// ParsePartnerToken 校验并解析 Partner Token
		// 参数：
		//   - ctx: 上下文
		//   - tokenStr: JWT Token 字符串
		//
		// 返回：
		//   - claims: Token 中的声明信息
		//   - err: 错误信息
		ParsePartnerToken(ctx context.Context, tokenStr string) (*grab.PartnerTokenClaims, error)
		// GetPartnerConfig 通过 partner code 获取配置
		// 参数：
		//   - ctx: 上下文
		//   - code: Partner 代码
		//
		// 返回：
		//   - partner: Partner 配置
		//   - err: 错误信息
		GetPartnerConfig(ctx context.Context, code string) (*conf.GrabPartner, error)
		// GetPartnerConfigByClientID 通过 client_id 获取配置
		// 参数：
		//   - ctx: 上下文
		//   - clientID: Client ID
		//
		// 返回：
		//   - partner: Partner 配置
		//   - err: 错误信息
		GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.GrabPartner, error)
	}
)

var (
	localGrabToken IGrabToken
)

func GrabToken() IGrabToken {
	if localGrabToken == nil {
		panic("implement not found for interface IGrabToken, forgot register?")
	}
	return localGrabToken
}

func RegisterGrabToken(i IGrabToken) {
	localGrabToken = i
}
