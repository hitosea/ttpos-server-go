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
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

const (
	defaultLinemanTokenTTL = 3600                   // 秒 (1小时)
	tokenExpireBuffer      = 60                     // Token 过期缓冲时间（秒）
	redisKeyTokenPrefix    = "lineman:oauth:token:" // Redis Key 前缀
)

// sLinemanToken LINE MAN Token 服务
type sLinemanToken struct {
	cfgLoader *PartnerConfigLoader
	secretKey string
	expiresIn int
	tokenLock sync.Mutex // Token 获取互斥锁（用于双重检查锁）
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

// ============================================================================
// OAuth Access Token 管理
// ============================================================================

// FetchTokenFromAPI 从 LINE MAN OAuth 服务器获取 Access Token
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - token: OAuth Access Token
//   - expiresIn: Token 有效期（秒）
//   - err: 错误信息
func (s *sLinemanToken) FetchTokenFromAPI(ctx context.Context) (string, int, error) {
	cfg := MustConfig(ctx)

	// 构造 OAuth Token URL
	tokenURL := cfg.Endpoint + "/oauth/token"

	// 构造请求体
	reqBody := lineman.LinemanOAuthTokenRequest{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		GrantType:    "client_credentials",
	}

	// 发送 POST 请求
	client := g.Client().SetTimeout(cfg.Timeout)
	resp, err := client.ContentJson().Post(ctx, tokenURL, reqBody)
	if err != nil {
		return "", 0, gerror.Wrapf(err, "[LINE MAN] OAuth API 请求失败")
	}
	defer resp.Close()

	// 解析响应
	respBytes := resp.ReadAll()
	g.Log().Debugf(ctx, "[LINE MAN] OAuth API 响应: status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		g.Log().Errorf(ctx, "[LINE MAN] OAuth API 返回错误: status=%d, body=%s", resp.StatusCode, string(respBytes))
		return "", 0, gerror.Newf("OAuth API 返回错误状态码: %d", resp.StatusCode)
	}

	// 解析 JSON 响应
	var tokenResp lineman.LinemanOAuthTokenResponse
	if err := gjson.Unmarshal(respBytes, &tokenResp); err != nil {
		return "", 0, gerror.Wrapf(err, "[LINE MAN] OAuth 响应解析失败")
	}

	// 校验必需字段
	if tokenResp.AccessToken == "" {
		return "", 0, gerror.New("[LINE MAN] OAuth 响应缺少 access_token")
	}
	if tokenResp.ExpiresIn <= 0 {
		return "", 0, gerror.New("[LINE MAN] OAuth 响应的 expires_in 无效")
	}

	g.Log().Infof(ctx, "[LINE MAN] OAuth Token 获取成功, expires_in=%d", tokenResp.ExpiresIn)
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// GetAccessToken 获取或刷新 Access Token（使用 Redis 缓存 + 双重检查锁）
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - token: OAuth Access Token
//   - err: 错误信息
func (s *sLinemanToken) GetAccessToken(ctx context.Context) (string, error) {
	cfg := MustConfig(ctx)
	// 构造 Redis Key
	redisKey := redisKeyTokenPrefix + cfg.ClientID

	// 1. 第一次检查：尝试从 Redis 读取（无锁）
	cachedToken, err := g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[LINE MAN] OAuth Token 缓存命中: %s", redisKey)
		return token, nil
	}

	// 2. Redis Miss 或错误，获取互斥锁
	s.tokenLock.Lock()
	defer s.tokenLock.Unlock()

	// 3. 双重检查：获取锁后再次检查缓存
	cachedToken, err = g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[LINE MAN] OAuth Token 缓存命中（双重检查）: %s", redisKey)
		return token, nil
	}

	// 4. 从 LINE MAN OAuth API 获取新 Token
	g.Log().Infof(ctx, "[LINE MAN] OAuth Token 缓存未命中，从远程获取")
	token, expiresIn, err := s.FetchTokenFromAPI(ctx)
	if err != nil {
		return "", err
	}

	// 5. 写入 Redis 缓存（TTL = expires_in - 缓冲时间）
	ttl := expiresIn - tokenExpireBuffer
	if ttl > 0 {
		if err := g.Redis().SetEX(ctx, redisKey, token, int64(ttl)); err != nil {
			// Redis 写入失败仅记录日志，不影响返回
			g.Log().Warningf(ctx, "[LINE MAN] Token 缓存到 Redis 失败: %v", err)
		} else {
			g.Log().Infof(ctx, "[LINE MAN] OAuth Token 已缓存到 Redis: key=%s, ttl=%ds", redisKey, ttl)
		}
	}

	return token, nil
}

// GetAuthorizationHeader 获取 Authorization 请求头（Bearer Token 格式）
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - header: Authorization 请求头值（格式: "Bearer {token}"）
//   - err: 错误信息
func (s *sLinemanToken) GetAuthorizationHeader(ctx context.Context) (string, error) {
	token, err := s.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}
