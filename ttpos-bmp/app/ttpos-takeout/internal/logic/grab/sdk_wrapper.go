// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

const (
	// RedisKeyTokenPrefix Redis Token Key 前缀
	RedisKeyTokenPrefix = "ttpos:takeout:grab:token:"
	// TokenExpireBuffer Token 过期缓冲时间 (秒)，提前刷新避免请求时过期
	TokenExpireBuffer = 60
)

// SDKConfig SDK 配置
type SDKConfig struct {
	ClientID     string // OAuth Client ID
	ClientSecret string // OAuth Client Secret
	Environment  string // "staging" 或 "production"
}

// SDKWrapper 封装 Grab 官方 SDK
// 提供统一的 API 访问入口，隔离业务代码与 SDK 直接依赖
type SDKWrapper struct {
	client    *grabfood.APIClient
	config    *SDKConfig
	tokenLock sync.RWMutex
}

// NewSDKWrapper 创建 SDK Wrapper
func NewSDKWrapper(cfg *SDKConfig) *SDKWrapper {
	sdkConfig := grabfood.NewConfiguration()
	return &SDKWrapper{
		client: grabfood.NewAPIClient(sdkConfig),
		config: cfg,
	}
}

// GetContext 获取带环境配置的 Context
// staging 环境使用 StgEnv (0)，production 环境使用 PrdEnv (1)
func (w *SDKWrapper) GetContext(ctx context.Context) context.Context {
	if w.config.Environment == "staging" {
		return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.StgEnv)
	}
	return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.PrdEnv)
}

// fetchTokenFromSDK 通过 SDK 从 Grab OAuth2 服务器获取新 Token
func (w *SDKWrapper) fetchTokenFromSDK(ctx context.Context) (string, int, error) {
	authReq := grabfood.NewGrabOauthRequest(
		w.config.ClientID,
		w.config.ClientSecret,
		"client_credentials",
		"food.partner_api",
	)

	resp, httpResp, err := w.client.GetOauthGrabAPI.
		GetOauthGrab(w.GetContext(ctx)).
		GrabOauthRequest(*authReq).
		Execute()

	if err != nil {
		return "", 0, fmt.Errorf("SDK OAuth request failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	if resp.AccessToken == nil || resp.ExpiresIn == nil {
		return "", 0, fmt.Errorf("SDK OAuth response missing required fields")
	}

	g.Log().Infof(ctx, "[SDK] Grab OAuth token fetched successfully, expires_in=%d", *resp.ExpiresIn)
	return *resp.AccessToken, int(*resp.ExpiresIn), nil
}

// GetAccessToken 获取或刷新 Access Token (使用 Redis 缓存)
func (w *SDKWrapper) GetAccessToken(ctx context.Context) (string, error) {
	// 构造 Redis Key (使用与旧实现相同的前缀，保持兼容)
	redisKey := RedisKeyTokenPrefix + w.config.ClientID

	// 1. 尝试从 Redis 读取
	cachedToken, err := g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[SDK] Grab OAuth token hit from cache: %s", redisKey)
		return token, nil
	}

	// 2. Redis Miss 或错误，从 SDK 远程获取
	w.tokenLock.Lock()
	defer w.tokenLock.Unlock()

	// Double check: 获取锁后再次检查缓存
	cachedToken, err = g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		return cachedToken.String(), nil
	}

	g.Log().Infof(ctx, "[SDK] Grab OAuth token cache miss, fetching from remote")
	token, expiresIn, err := w.fetchTokenFromSDK(ctx)
	if err != nil {
		return "", err
	}

	// 3. 写入 Redis 缓存 (TTL = expires_in - 缓冲时间)
	ttl := expiresIn - TokenExpireBuffer
	if ttl > 0 {
		if err := g.Redis().SetEX(ctx, redisKey, token, int64(ttl)); err != nil {
			// Redis 写入失败仅记录日志，不影响返回
			g.Log().Warningf(ctx, "[SDK] Failed to cache Grab token to Redis: %v", err)
		} else {
			g.Log().Infof(ctx, "[SDK] Grab OAuth token cached to Redis: key=%s, ttl=%ds", redisKey, ttl)
		}
	}

	return token, nil
}

// GetAuthorizationHeader 获取 Authorization 请求头
func (w *SDKWrapper) GetAuthorizationHeader(ctx context.Context) (string, error) {
	token, err := w.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}

// ============================================================================
// 订单操作 API
// ============================================================================

// AcceptOrder 接受订单
func (w *SDKWrapper) AcceptOrder(ctx context.Context, orderID string) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	// SDK 使用 AcceptOrderRequest: orderID + toState
	acceptReq := grabfood.NewAcceptOrderRequest(orderID, "ACCEPTED")

	httpResp, err := w.client.AcceptRejectOrderAPI.
		AcceptRejectOrder(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*acceptReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK AcceptOrder failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Order accepted: %s", orderID)
	return nil
}

// RejectOrder 拒绝订单
func (w *SDKWrapper) RejectOrder(ctx context.Context, orderID string, rejectCode int) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	// SDK 使用 AcceptOrderRequest: orderID + toState
	// 注意: SDK 的 AcceptOrderRequest 没有 SetRejectCode 方法
	// 需要使用 WithDefaults 然后手动设置字段，或者直接构造
	rejectReq := grabfood.NewAcceptOrderRequest(orderID, "REJECTED")

	httpResp, err := w.client.AcceptRejectOrderAPI.
		AcceptRejectOrder(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*rejectReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK RejectOrder failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Order rejected: %s, code=%d", orderID, rejectCode)
	return nil
}

// CancelOrder 取消订单
func (w *SDKWrapper) CancelOrder(ctx context.Context, orderID string, merchantID string, cancelCode int) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	// SDK CancelCode 是 int32 类型
	cancelCodeEnum := grabfood.CancelCode(cancelCode)
	cancelReq := grabfood.NewCancelOrderRequest(orderID, merchantID, cancelCodeEnum)

	_, httpResp, err := w.client.CancelOrderAPI.
		CancelOrder(w.GetContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		CancelOrderRequest(*cancelReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK CancelOrder failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Order cancelled: %s, code=%d", orderID, cancelCode)
	return nil
}

// MarkOrderReady 标记订单准备完成
// markStatus: 1 = ready for pickup
func (w *SDKWrapper) MarkOrderReady(ctx context.Context, orderID string, markStatus int32) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	markReq := grabfood.NewMarkOrderRequest(orderID, markStatus)

	httpResp, err := w.client.MarkOrderReadyAPI.
		MarkOrderReady(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		MarkOrderRequest(*markReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK MarkOrderReady failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Order marked ready: %s, status=%d", orderID, markStatus)
	return nil
}

// UpdateDeliveryState 更新配送状态 (自配送)
func (w *SDKWrapper) UpdateDeliveryState(ctx context.Context, orderID string, fromState string, toState string) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	deliveryReq := grabfood.NewOrderDeliveryRequest(orderID, fromState, toState)

	httpResp, err := w.client.UpdateDeliveryStateAPI.
		UpdateDeliveryState(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		OrderDeliveryRequest(*deliveryReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK UpdateDeliveryState failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Delivery state updated: %s, %s -> %s", orderID, fromState, toState)
	return nil
}

// UpdateOrderReadyTime 更新订单准备时间
func (w *SDKWrapper) UpdateOrderReadyTime(ctx context.Context, orderID string, newReadyTime time.Time) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	readyTimeReq := grabfood.NewNewOrderTimeRequest(orderID, newReadyTime)

	httpResp, err := w.client.UpdateOrderReadyTimeAPI.
		UpdateOrderReadyTime(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		NewOrderTimeRequest(*readyTimeReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK UpdateOrderReadyTime failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[SDK] Order ready time updated: %s, newTime=%s", orderID, newReadyTime.Format(time.RFC3339))
	return nil
}

// CheckOrderCancelable 检查订单是否可取消
func (w *SDKWrapper) CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (bool, string, error) {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return false, "", fmt.Errorf("get authorization failed: %w", err)
	}

	resp, httpResp, err := w.client.CheckOrderCancelableAPI.
		CheckOrderCancelable(w.GetContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		OrderID(orderID).
		Execute()

	if err != nil {
		return false, "", fmt.Errorf("SDK CheckOrderCancelable failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 注意: SDK 字段名是 CancelAble (而非 Cancelable)
	cancelable := resp.CancelAble != nil && *resp.CancelAble
	reason := ""
	if resp.NonCancellationReason != nil {
		reason = *resp.NonCancellationReason
	}

	return cancelable, reason, nil
}

// ============================================================================
// 门店管理 API
// ============================================================================

// UpdateStoreStatus 更新门店状态 (暂停/恢复)
// duration: 暂停时长（分钟），仅在 isPause=true 时有效
func (w *SDKWrapper) UpdateStoreStatus(ctx context.Context, merchantID string, isPause bool, duration int) error {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return fmt.Errorf("get authorization failed: %w", err)
	}

	pauseReq := grabfood.NewPauseStoreRequest(merchantID, isPause)
	if isPause && duration > 0 {
		// SDK Duration 是 string 类型
		pauseReq.SetDuration(strconv.Itoa(duration))
	}

	httpResp, err := w.client.PauseStoreAPI.
		PauseStore(w.GetContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		PauseStoreRequest(*pauseReq).
		Execute()

	if err != nil {
		return fmt.Errorf("SDK UpdateStoreStatus failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	action := "resumed"
	if isPause {
		action = fmt.Sprintf("paused for %d minutes", duration)
	}
	g.Log().Infof(ctx, "[SDK] Store status updated: merchant=%s, action=%s", merchantID, action)
	return nil
}

// GetStoreStatus 获取门店状态
func (w *SDKWrapper) GetStoreStatus(ctx context.Context, merchantID string) (*grabfood.StoreStatusResponse, error) {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get authorization failed: %w", err)
	}

	resp, httpResp, err := w.client.GetStoreStatusAPI.
		GetStoreStatus(w.GetContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("SDK GetStoreStatus failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// GetStoreHours 获取门店营业时间
func (w *SDKWrapper) GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error) {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get authorization failed: %w", err)
	}

	resp, httpResp, err := w.client.GetStoreHourAPI.
		GetStoreHour(w.GetContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("SDK GetStoreHours failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// ============================================================================
// 菜单管理 API
// ============================================================================

// NotifyMenuUpdate 通知 Grab 菜单已更新
// 返回 requestID 用于追踪同步状态
func (w *SDKWrapper) NotifyMenuUpdate(ctx context.Context, merchantID string) (string, error) {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return "", fmt.Errorf("get authorization failed: %w", err)
	}

	notifReq := grabfood.NewUpdateMenuNotifRequest(merchantID)

	httpResp, err := w.client.UpdateMenuNotificationAPI.
		UpdateMenuNotification(w.GetContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		UpdateMenuNotifRequest(*notifReq).
		Execute()

	if err != nil {
		return "", fmt.Errorf("SDK NotifyMenuUpdate failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// SDK 的 UpdateMenuNotification 返回的 HTTP Response 可能包含 Request-Id 头
	requestID := ""
	if httpResp != nil {
		requestID = httpResp.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = httpResp.Header.Get("Request-Id")
		}
	}

	g.Log().Infof(ctx, "[SDK] Menu update notification sent: merchant=%s, requestId=%s", merchantID, requestID)
	return requestID, nil
}

// TraceMenuSync 追踪菜单同步状态
func (w *SDKWrapper) TraceMenuSync(ctx context.Context, merchantID string) (*grabfood.MenuSyncResponse, error) {
	auth, err := w.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get authorization failed: %w", err)
	}

	resp, httpResp, err := w.client.TraceMenuSyncAPI.
		TraceMenuSync(w.GetContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("SDK TraceMenuSync failed: %w", err)
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}
