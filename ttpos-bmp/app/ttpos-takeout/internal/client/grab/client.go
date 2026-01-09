// Package grab 提供 Grab SDK 客户端封装
package grab

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

const (
	// RedisKeyTokenPrefix Redis Token Key 前缀
	RedisKeyTokenPrefix = "ttpos:takeout:grab:token:"
	// TokenExpireBuffer Token 过期缓冲时间 (秒)，提前刷新避免请求时过期
	TokenExpireBuffer = 60
)

// Client Grab SDK 客户端封装
// 负责 SDK 初始化、OAuth Token 管理、签名验证等技术基础设施
type Client struct {
	config    *conf.Grab
	sdkClient *grabfood.APIClient
	verifier  *SignatureVerifier
	tokenLock sync.RWMutex
}

var (
	defaultClient *Client
	clientMu      sync.RWMutex
	config        *conf.Grab
	configMu      sync.RWMutex
)

// Default 获取默认客户端 (懒加载单例)
func Default() *Client {
	clientMu.RLock()
	if defaultClient != nil {
		defer clientMu.RUnlock()
		return defaultClient
	}
	clientMu.RUnlock()

	clientMu.Lock()
	defer clientMu.Unlock()

	// 双重检查
	if defaultClient != nil {
		return defaultClient
	}

	defaultClient = New(MustConf())
	return defaultClient
}

// New 创建新客户端
func New(config *conf.Grab) *Client {
	return &Client{
		config: config,
	}
}

// MustConf 获取 Grab 配置
func MustConf() *conf.Grab {
	configMu.RLock()
	if config != nil {
		defer configMu.RUnlock()
		return config
	}
	configMu.RUnlock()

	configMu.Lock()
	defer configMu.Unlock()

	// 双重检查
	if config != nil {
		return config
	}

	config = MustConfig()
	return config
}

// GetClient 获取 SDK 客户端 (懒加载)
func (c *Client) GetClient() *grabfood.APIClient {
	if c.sdkClient == nil {
		sdkConfig := grabfood.NewConfiguration()
		c.sdkClient = grabfood.NewAPIClient(sdkConfig)
	}
	return c.sdkClient
}

// GetVerifier 获取签名验证器 (懒加载)
func (c *Client) GetVerifier() *SignatureVerifier {
	if c.verifier == nil {
		c.verifier = newSignatureVerifier(c.config.SecretKey)
	}
	return c.verifier
}

// HandleSDKError 统一处理 Grab SDK Execute() 返回的错误
// 参数:
//   - ctx: 上下文
//   - err: SDK 返回的错误
//   - operation: 操作名称（如 "AcceptOrder", "GetStoreStatus"）
//
// 返回: 包装后的错误
func (c *Client) HandleSDKError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}
	errBody := ""
	// 安全的类型断言，记录 API 详细错误
	if apiErr, ok := err.(*grabfood.GenericOpenAPIError); ok {
		errBody = string(apiErr.Body())
		g.Log().Errorf(ctx, "[Grab] SDK %s 失败: %s", operation, errBody)
	} else {
		g.Log().Errorf(ctx, "[Grab] SDK %s 失败: %v", operation, err)
	}

	return gerror.Wrapf(err, "[Grab] SDK %s 失败 %s", operation, errBody)
}

// GetSDKContext 获取带环境配置的 Context
// staging 环境使用 StgEnv (0)，production 环境使用 PrdEnv (1)
func (c *Client) GetSDKContext(ctx context.Context) context.Context {
	if c.config.Environment == "staging" {
		return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.StgEnv)
	}
	return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.PrdEnv)
}

// fetchTokenFromSDK 通过 SDK 从 Grab OAuth2 服务器获取新 Token
func (c *Client) fetchTokenFromSDK(ctx context.Context) (string, int, error) {
	authReq := grabfood.NewGrabOauthRequest(
		c.config.ClientID,
		c.config.ClientSecret,
		"client_credentials",
		"food.partner_api",
	)

	resp, httpResp, err := c.GetClient().GetOauthGrabAPI.
		GetOauthGrab(c.GetSDKContext(ctx)).
		ContentType("application/json").
		GrabOauthRequest(*authReq).
		Execute()

	if err = c.HandleSDKError(ctx, err, "OAuth"); err != nil {
		return "", 0, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	if resp.AccessToken == nil || resp.ExpiresIn == nil {
		return "", 0, gerror.New("SDK OAuth 响应缺少必需字段")
	}

	g.Log().Infof(ctx, "[Grab] OAuth Token 获取成功, expires_in=%d", *resp.ExpiresIn)
	return *resp.AccessToken, int(*resp.ExpiresIn), nil
}

// GetAccessToken 获取或刷新 Access Token (使用 Redis 缓存)
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	// 构造 Redis Key
	redisKey := RedisKeyTokenPrefix + c.config.ClientID

	// 1. 尝试从 Redis 读取
	cachedToken, err := g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[Grab] OAuth Token 缓存命中: %s", redisKey)
		return token, nil
	}

	// 2. Redis Miss 或错误，从 SDK 远程获取
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	// Double check: 获取锁后再次检查缓存
	cachedToken, err = g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		return cachedToken.String(), nil
	}

	g.Log().Infof(ctx, "[Grab] OAuth Token 缓存未命中，从远程获取")
	token, expiresIn, err := c.fetchTokenFromSDK(ctx)
	if err != nil {
		return "", err
	}

	// 3. 写入 Redis 缓存 (TTL = expires_in - 缓冲时间)
	ttl := expiresIn - TokenExpireBuffer
	if ttl > 0 {
		if err := g.Redis().SetEX(ctx, redisKey, token, int64(ttl)); err != nil {
			// Redis 写入失败仅记录日志，不影响返回
			g.Log().Warningf(ctx, "[Grab] Token 缓存到 Redis 失败: %v", err)
		} else {
			g.Log().Infof(ctx, "[Grab] OAuth Token 已缓存到 Redis: key=%s, ttl=%ds", redisKey, ttl)
		}
	}

	return token, nil
}

// GetAuthorizationHeader 获取 Authorization 请求头
func (c *Client) GetAuthorizationHeader(ctx context.Context) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}
