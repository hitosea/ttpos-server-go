// Package grab_token 提供 Grab Partner Token 生成与验证服务
package grab

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

const (
	defaultPartnerTokenTTL = 900 // 秒
	defaultGrabScope       = "food.partner_api"
)

var (
	tokenCfgLoader *PartnerConfigLoader
	tokenSecretKey string
	tokenExpiresIn = defaultPartnerTokenTTL
	tokenMu        sync.RWMutex
)

// getConfigLoader 获取配置加载器（懒加载）
func getTokenConfigLoader() *PartnerConfigLoader {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	if tokenCfgLoader == nil {
		tokenCfgLoader = &PartnerConfigLoader{}
	}
	return tokenCfgLoader
}

// getTokenSecretKey 获取密钥（懒加载）
func getTokenSecretKey(ctx context.Context) string {
	tokenMu.RLock()
	if tokenSecretKey != "" {
		defer tokenMu.RUnlock()
		return tokenSecretKey
	}
	tokenMu.RUnlock()

	tokenMu.Lock()
	defer tokenMu.Unlock()
	if tokenSecretKey == "" {
		cfg := MustConfig(ctx)
		tokenSecretKey = cfg.SecretKey
	}
	return tokenSecretKey
}

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
func (s *sGrab) GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (token string, expiresIn int, err error) {
	if clientID == "" {
		return "", 0, gerror.New("client_id 不能为空")
	}
	if clientSecret == "" {
		return "", 0, gerror.New("client_secret 不能为空")
	}

	// 根据 client_id 获取配置
	var partnerCfg *conf.GrabPartner
	partnerCfg, err = getTokenConfigLoader().GetByClientID(ctx, clientID)
	if err != nil {
		return "", 0, gerror.Wrap(err, "根据 client_id 获取配置失败")
	}

	// 校验 client_secret
	if partnerCfg.ClientSecret != clientSecret {
		return "", 0, gerror.New("client_secret 不匹配")
	}

	// 使用请求中的 scope，若为空则使用默认值
	if scope == "" {
		scope = defaultGrabScope
	}

	now := time.Now()
	expireAt := now.Add(time.Duration(tokenExpiresIn) * time.Second)
	claims := grab.PartnerTokenClaims{
		ClientID:    clientID,
		Scope:       scope,
		PartnerCode: clientID, // 使用 clientID 作为 partnerCode
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(getTokenSecretKey(context.Background())))
	if err != nil {
		return "", 0, gerror.Wrap(err, "签发 Token 失败")
	}

	return signed, tokenExpiresIn, nil
}

// ParsePartnerToken 校验并解析 Partner Token
// 参数：
//   - ctx: 上下文
//   - tokenStr: JWT Token 字符串
//
// 返回：
//   - claims: Token 中的声明信息
//   - err: 错误信息
func (s *sGrab) ParsePartnerToken(tokenStr string) (*grab.PartnerTokenClaims, error) {
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

	secretKey := getTokenSecretKey(context.Background())
	token, err := jwt.ParseWithClaims(tokenStr, &grab.PartnerTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, gerror.Newf("不支持的签名方法: %s", t.Method.Alg())
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, gerror.Wrap(err, "Token 解析失败")
	}
	claims, ok := token.Claims.(*grab.PartnerTokenClaims)
	if !ok || !token.Valid {
		return nil, gerror.New("Token 无效")
	}
	return claims, nil
}
