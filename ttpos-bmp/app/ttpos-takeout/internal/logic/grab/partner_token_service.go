package grab

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

const defaultPartnerTokenTTL = 900 // 秒

// PartnerTokenService 生成 Partner Token
type PartnerTokenService struct {
	configLoader *PartnerConfigLoader
	secretKey    string
	expiresIn    int
}

// NewPartnerTokenService 创建 PartnerTokenService
func NewPartnerTokenService(loader *PartnerConfigLoader, secretKey string, expiresIn int) *PartnerTokenService {
	if expiresIn <= 0 {
		expiresIn = defaultPartnerTokenTTL
	}
	return &PartnerTokenService{
		configLoader: loader,
		secretKey:    secretKey,
		expiresIn:    expiresIn,
	}
}

// GeneratePartnerToken 根据 client_id / client_secret 生成访问 Token
// 采用 JWT（HS256）实现，参考 GoFrame 示例
// 参数：
//   - clientID: Grab 分配的 client_id
//   - clientSecret: Grab 分配的 client_secret，用于校验身份
//   - scope: 请求的权限范围
func (s *PartnerTokenService) GeneratePartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (token string, expiresIn int, err error) {
	if clientID == "" {
		return "", 0, gerror.New("client_id 不能为空")
	}
	if clientSecret == "" {
		return "", 0, gerror.New("client_secret 不能为空")
	}

	// 根据 client_id 获取配置
	var partnerCfg *conf.GrabPartner
	partnerCfg, err = s.configLoader.GetByClientID(ctx, clientID)
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
	expireAt := now.Add(time.Duration(s.expiresIn) * time.Second)
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
	signed, err := jwtToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", 0, gerror.Wrap(err, "签发 Token 失败")
	}

	return signed, s.expiresIn, nil
}

// ParsePartnerToken 校验并解析 Partner Token
func (s *PartnerTokenService) ParsePartnerToken(tokenStr string) (*grab.PartnerTokenClaims, error) {
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

	token, err := jwt.ParseWithClaims(tokenStr, &grab.PartnerTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, gerror.Newf("不支持的签名方法: %s", t.Method.Alg())
		}
		return []byte(s.secretKey), nil
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
