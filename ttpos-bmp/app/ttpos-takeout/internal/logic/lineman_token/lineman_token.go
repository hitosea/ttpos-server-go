// Package lineman_token 提供 LINE MAN OAuth Token 生成与验证服务
//
// ⚠️ 临时方案说明:
// 本包为临时实现方案,参考 Grab OAuth 架构快速支持 LINE MAN 平台接入。
// 后续将迁移到统一的权限中心单点登录系统 (SSO),届时本包将被废弃。
// 开发者在维护时应保持代码简洁,避免过度复杂化,以便未来迁移。
package lineman_token

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

const (
	defaultLinemanTokenTTL = 3600 // 秒 (1小时)
)

// sLinemanToken LINE MAN Token 服务
type sLinemanToken struct {
	cfgLoader *PartnerConfigLoader
	secretKey string
	expiresIn int
}

func init() {
	service.RegisterLinemanToken(New())
}

// New 创建 LinemanToken 服务实例
func New() *sLinemanToken {
	return &sLinemanToken{
		expiresIn: defaultLinemanTokenTTL,
	}
}

// getConfigLoader 获取配置加载器（懒加载）
func (s *sLinemanToken) getConfigLoader() *PartnerConfigLoader {
	if s.cfgLoader == nil {
		s.cfgLoader = &PartnerConfigLoader{}
	}
	return s.cfgLoader
}

// getSecretKey 获取密钥（懒加载）
func (s *sLinemanToken) getSecretKey(ctx context.Context) string {
	if s.secretKey == "" {
		cfg := MustConfig(ctx)
		s.secretKey = cfg.SecretKey
	}
	return s.secretKey
}

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
func (s *sLinemanToken) GenerateToken(ctx context.Context, clientID string, clientSecret string) (token string, expiresIn int, err error) {
	if clientID == "" {
		return "", 0, gerror.New("client_id 不能为空")
	}
	if clientSecret == "" {
		return "", 0, gerror.New("client_secret 不能为空")
	}

	// 根据 client_id 获取配置
	var partnerCfg *conf.LinemanPartner
	partnerCfg, err = s.getConfigLoader().GetByClientID(ctx, clientID)
	if err != nil {
		return "", 0, gerror.Wrap(err, "根据 client_id 获取配置失败")
	}

	// 校验 client_secret
	if partnerCfg.ClientSecret != clientSecret {
		return "", 0, gerror.New("client_secret 不匹配")
	}

	now := time.Now()
	expireAt := now.Add(time.Duration(s.expiresIn) * time.Second)
	claims := lineman.LinemanTokenClaims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(s.getSecretKey(ctx)))
	if err != nil {
		return "", 0, gerror.Wrap(err, "签发 Token 失败")
	}

	return signed, s.expiresIn, nil
}

// ParseToken 校验并解析 LINE MAN Token
// 参数：
//   - ctx: 上下文
//   - tokenStr: JWT Token 字符串
//
// 返回：
//   - claims: 解析后的 Claims
//   - err: 错误信息
func (s *sLinemanToken) ParseToken(ctx context.Context, tokenStr string) (*lineman.LinemanTokenClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenStr, &lineman.LinemanTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, gerror.Newf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(s.getSecretKey(ctx)), nil
	})

	if err != nil {
		return nil, gerror.Wrap(err, "Token 解析失败")
	}

	if claims, ok := token.Claims.(*lineman.LinemanTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, gerror.New("Token 无效")
}

// GetPartnerConfig 通过 partner code 获取配置
// 参数：
//   - ctx: 上下文
//   - code: Partner 代码
//
// 返回：
//   - partner: Partner 配置
//   - err: 错误信息
func (s *sLinemanToken) GetPartnerConfig(ctx context.Context, code string) (*conf.LinemanPartner, error) {
	return s.getConfigLoader().GetByCode(ctx, code)
}

// GetPartnerConfigByClientID 通过 client_id 获取配置
// 参数：
//   - ctx: 上下文
//   - clientID: Client ID
//
// 返回：
//   - partner: Partner 配置
//   - err: 错误信息
func (s *sLinemanToken) GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.LinemanPartner, error) {
	return s.getConfigLoader().GetByClientID(ctx, clientID)
}
