// Package lineman Lineman API 客户端
package lineman

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
)

const (
	defaultLinemanTokenTTL = 3600                   // 秒 (1小时)
	tokenExpireBuffer      = 60                     // Token 过期缓冲时间（秒）
	redisKeyTokenPrefix    = "lineman:oauth:token:" // Redis Key 前缀
)

// ============================================================================
// JWT Token 客户端
// ============================================================================

// JWTTokenClient JWT Token 生成和解析客户端
type JWTTokenClient struct {
	cfgLoader *PartnerConfigLoader
	secretKey string
	expiresIn int
}

// NewJWTTokenClient 创建 JWT Token 客户端
func NewJWTTokenClient() *JWTTokenClient {
	return &JWTTokenClient{
		cfgLoader: &PartnerConfigLoader{},
		expiresIn: defaultLinemanTokenTTL,
	}
}

// getSecretKey 获取密钥（懒加载）
func (c *JWTTokenClient) getSecretKey(ctx context.Context) string {
	if c.secretKey == "" {
		cfg := MustConfig(ctx)
		c.secretKey = cfg.SecretKey
	}
	return c.secretKey
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
func (c *JWTTokenClient) GenerateToken(ctx context.Context, clientID string, clientSecret string) (token string, expiresIn int, err error) {
	if clientID == "" {
		return "", 0, gerror.New("client_id 不能为空")
	}
	if clientSecret == "" {
		return "", 0, gerror.New("client_secret 不能为空")
	}

	// 根据 client_id 获取配置
	var partnerCfg *conf.LinemanPartner
	partnerCfg, err = c.cfgLoader.GetByClientID(ctx, clientID)
	if err != nil {
		return "", 0, gerror.Wrap(err, "根据 client_id 获取配置失败")
	}

	// 校验 client_secret
	if partnerCfg.ClientSecret != clientSecret {
		return "", 0, gerror.New("client_secret 不匹配")
	}

	now := time.Now()
	expireAt := now.Add(time.Duration(c.expiresIn) * time.Second)
	claims := lineman.LinemanTokenClaims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(c.getSecretKey(ctx)))
	if err != nil {
		return "", 0, gerror.Wrap(err, "签发 Token 失败")
	}

	return signed, c.expiresIn, nil
}

// ParseToken 校验并解析 LINE MAN Token
// 参数：
//   - ctx: 上下文
//   - tokenStr: JWT Token 字符串
//
// 返回：
//   - claims: 解析后的 Claims
//   - err: 错误信息
func (c *JWTTokenClient) ParseToken(ctx context.Context, tokenStr string) (*lineman.LinemanTokenClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenStr, &lineman.LinemanTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, gerror.Newf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(c.getSecretKey(ctx)), nil
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
func (c *JWTTokenClient) GetPartnerConfig(ctx context.Context, code string) (*conf.LinemanPartner, error) {
	return c.cfgLoader.GetByCode(ctx, code)
}

// GetPartnerConfigByClientID 通过 client_id 获取配置
func (c *JWTTokenClient) GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.LinemanPartner, error) {
	return c.cfgLoader.GetByClientID(ctx, clientID)
}

// ============================================================================
// OAuth Token 客户端
// ============================================================================

// OAuthTokenClient OAuth Access Token 管理客户端
type OAuthTokenClient struct {
	tokenLock sync.Mutex // Token 获取互斥锁（用于双重检查锁）
}

// NewOAuthTokenClient 创建 OAuth Token 客户端
func NewOAuthTokenClient() *OAuthTokenClient {
	return &OAuthTokenClient{}
}

// FetchTokenFromAPI 从 LINE MAN OAuth 服务器获取 Access Token
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - token: OAuth Access Token
//   - expiresIn: Token 有效期（秒）
//   - err: 错误信息
func (c *OAuthTokenClient) FetchTokenFromAPI(ctx context.Context) (string, int, error) {
	cfg := MustConfig(ctx)

	// 构造 OAuth Token URL
	tokenURL := cfg.Endpoint + "/v1/oauth/token"

	// 构造请求体（使用表单格式）
	reqBody := g.Map{
		"grant_type":    "client_credentials",
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
	}

	// 发送 POST 请求（使用 application/x-www-form-urlencoded 格式）
	client := g.Client().SetTimeout(cfg.Timeout)
	resp, err := client.ContentType("application/x-www-form-urlencoded").Post(ctx, tokenURL, reqBody)
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
func (c *OAuthTokenClient) GetAccessToken(ctx context.Context) (string, error) {
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
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	// 3. 双重检查：获取锁后再次检查缓存
	cachedToken, err = g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[LINE MAN] OAuth Token 缓存命中（双重检查）: %s", redisKey)
		return token, nil
	}

	// 4. 从 LINE MAN OAuth API 获取新 Token
	g.Log().Infof(ctx, "[LINE MAN] OAuth Token 缓存未命中，从远程获取")
	token, expiresIn, err := c.FetchTokenFromAPI(ctx)
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
func (c *OAuthTokenClient) GetAuthorizationHeader(ctx context.Context) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}
