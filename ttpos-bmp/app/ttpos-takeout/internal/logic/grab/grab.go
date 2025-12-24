// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

const (
	// RedisKeyTokenPrefix Redis Token Key 前缀
	RedisKeyTokenPrefix = "ttpos:takeout:grab:token:"
	// TokenExpireBuffer Token 过期缓冲时间 (秒)，提前刷新避免请求时过期
	TokenExpireBuffer = 60
)

var (
	// Grab Grab服务实例
	Grab     = new(sGrab)
	config   *conf.Grab
	configMu sync.RWMutex
)

type sGrab struct {
	verifier  *SignatureVerifier
	client    *grabfood.APIClient // Grab SDK 客户端
	tokenLock sync.RWMutex        // Token 操作锁
}

func init() {
	service.RegisterGrab(Grab)
}

// MustConf 获取 Grab 配置
func (s *sGrab) MustConf() *conf.Grab {
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

// getClient 获取 SDK 客户端 (懒加载单例)
func (s *sGrab) getClient() *grabfood.APIClient {
	if s.client == nil {
		sdkConfig := grabfood.NewConfiguration()
		s.client = grabfood.NewAPIClient(sdkConfig)
	}
	return s.client
}

// getVerifier 获取签名验证器 (懒加载单例)
func (s *sGrab) getVerifier() *SignatureVerifier {
	if s.verifier == nil {
		conf := s.MustConf()
		s.verifier = newSignatureVerifier(conf.SecretKey)
	}
	return s.verifier
}

// handleSDKError 统一处理 Grab SDK Execute() 返回的错误
// 参数:
//   - ctx: 上下文
//   - err: SDK 返回的错误
//   - operation: 操作名称（如 "AcceptOrder", "GetStoreStatus"）
//
// 返回: 包装后的错误
func (s *sGrab) handleSDKError(ctx context.Context, err error, operation string) error {
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

// getSDKContext 获取带环境配置的 Context
// staging 环境使用 StgEnv (0)，production 环境使用 PrdEnv (1)
func (s *sGrab) getSDKContext(ctx context.Context) context.Context {
	conf := s.MustConf()
	if conf.Environment == "staging" {
		return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.StgEnv)
	}
	return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.PrdEnv)
}

// fetchTokenFromSDK 通过 SDK 从 Grab OAuth2 服务器获取新 Token
func (s *sGrab) fetchTokenFromSDK(ctx context.Context) (string, int, error) {
	conf := s.MustConf()
	authReq := grabfood.NewGrabOauthRequest(
		conf.ClientID,
		conf.ClientSecret,
		"client_credentials",
		"food.partner_api",
	)

	resp, httpResp, err := s.getClient().GetOauthGrabAPI.
		GetOauthGrab(s.getSDKContext(ctx)).
		ContentType("application/json").
		GrabOauthRequest(*authReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "OAuth"); err != nil {
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

// getAccessToken 获取或刷新 Access Token (使用 Redis 缓存)
func (s *sGrab) getAccessToken(ctx context.Context) (string, error) {
	conf := s.MustConf()
	// 构造 Redis Key
	redisKey := RedisKeyTokenPrefix + conf.ClientID

	// 1. 尝试从 Redis 读取
	cachedToken, err := g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		token := cachedToken.String()
		g.Log().Debugf(ctx, "[Grab] OAuth Token 缓存命中: %s", redisKey)
		return token, nil
	}

	// 2. Redis Miss 或错误，从 SDK 远程获取
	s.tokenLock.Lock()
	defer s.tokenLock.Unlock()

	// Double check: 获取锁后再次检查缓存
	cachedToken, err = g.Redis().Get(ctx, redisKey)
	if err == nil && !cachedToken.IsEmpty() {
		return cachedToken.String(), nil
	}

	g.Log().Infof(ctx, "[Grab] OAuth Token 缓存未命中，从远程获取")
	token, expiresIn, err := s.fetchTokenFromSDK(ctx)
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

// getAuthorizationHeader 获取 Authorization 请求头
func (s *sGrab) getAuthorizationHeader(ctx context.Context) (string, error) {
	token, err := s.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}

// ============================================================================
// Webhook 处理
// ============================================================================

// VerifyWebhookSignature 验证 Grab Webhook 签名 (公开方法，供其他服务调用)
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (s *sGrab) VerifyWebhookSignature(ctx context.Context, signature, timestamp string, body []byte) error {
	return s.getVerifier().VerifySignature(signature, timestamp, body)
}

// HandleSubmitOrder 处理 Grab 提交订单 Webhook
// 签名验证已由中间件完成
func (s *sGrab) HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error {
	return service.GrabOrder().HandleSubmitOrder(ctx, req)
}

// HandlePushOrderState 处理订单状态变更 Webhook
// 签名验证已由中间件完成
func (s *sGrab) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error {
	return service.GrabOrder().HandlePushOrderState(ctx, req)
}

// HandleGetMenu 处理 Grab 获取菜单请求
// 签名验证已由中间件完成
func (s *sGrab) HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error) {
	// 1. 将 merchantID (partnerMerchantID) 转换为 shopUUID
	shopUUID := g.NewVar(merchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] partnerMerchantID 格式无效: %s", merchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "partnerMerchantID 格式无效")
	}

	// 2. 调用 GrabMenu 服务获取菜单数据（不包含 MerchantID 和 PartnerMerchantID）
	menuResp, err := service.GrabMenu().HandleGetMenu(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	// 3. 查询 ShopProviderCfg
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取门店第三方配置失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取门店第三方配置失败")
	}
	if cfg == nil {
		g.Log().Errorf(ctx, "[Grab] 门店第三方配置不存在: shopUUID=%d, provider=%s", shopUUID, consts.ProviderGrab)
		return nil, gerror.NewCode(gcode.CodeNotFound, "门店第三方配置不存在")
	}

	// 4. 设置 MerchantID 和 PartnerMerchantID
	merchantIDStr := cfg.ProviderMerchantId
	partnerMerchantIDStr := fmt.Sprintf("%d", shopUUID)
	menuResp.MerchantID = &merchantIDStr
	menuResp.PartnerMerchantID = &partnerMerchantIDStr

	g.Log().Infof(ctx, "[Grab] 获取菜单成功: merchantID=%v, partnerMerchantID=%v, categories=%d",
		menuResp.MerchantID, menuResp.PartnerMerchantID, len(menuResp.Categories))

	return menuResp, nil
}

// HandleMenuSyncState 处理菜单同步状态回调
func (s *sGrab) HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error {
	return service.GrabMenu().HandleMenuSyncState(ctx, req)
}

// SyncMenu 主动同步菜单到 Grab
func (s *sGrab) SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse) error {
	// sGrab 实现了 MenuNotifier 接口 (NotifyMenuUpdate 方法)
	return service.GrabMenu().SyncMenu(ctx, merchantID, menu, s)
}

// HandleIntegrationStatus 处理门店集成状态回调
// 签名验证已由中间件完成
func (s *sGrab) HandleIntegrationStatus(ctx context.Context, body []byte) error {
	return service.GrabStore().HandleIntegrationStatus(ctx, body)
}

// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
func (s *sGrab) HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error {
	// 1. Save Snapshot
	menuUuid, err := service.GrabMenu().SaveMenuSnapshot(ctx, dto)
	if err != nil {
		return err
	}

	// 2. Notify
	event := &grabDto.ProviderMenuUpdateEvent{
		ProviderName:      string(consts.ProviderGrab),
		MerchantID:        dto.MerchantID,
		PartnerMerchantID: dto.PartnerMerchantID,
		ShopUuid:          dto.PartnerMerchantID,
		Uuid:              menuUuid,
		ReceivedAt:        gtime.Timestamp(),
	}
	return service.GrabMenu().NotifyMenuUpdate(ctx, event)
}

// ============================================================================
// 订单操作 API
// ============================================================================

// AcceptOrder 接受订单
func (s *sGrab) AcceptOrder(ctx context.Context, orderID string) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	acceptReq := grabfood.NewAcceptOrderRequest(orderID, "Accepted")

	httpResp, err := s.getClient().AcceptRejectOrderAPI.
		AcceptRejectOrder(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*acceptReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "AcceptOrder"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已接受: %s", orderID)
	return nil
}

// RejectOrder 拒绝订单
func (s *sGrab) RejectOrder(ctx context.Context, orderID string, rejectCode int) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	rejectReq := grabfood.NewAcceptOrderRequest(orderID, "Rejected")

	httpResp, err := s.getClient().AcceptRejectOrderAPI.
		AcceptRejectOrder(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*rejectReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "RejectOrder"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已拒绝: %s, code=%d", orderID, rejectCode)
	return nil
}

// CancelOrder 取消订单
func (s *sGrab) CancelOrder(ctx context.Context, orderID string, cancelCode int) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	// SDK CancelCode 是 int32 类型
	cancelCodeEnum := grabfood.CancelCode(cancelCode)
	// TODO: 优化接口，传入 merchantID
	cancelReq := grabfood.NewCancelOrderRequest(orderID, "", cancelCodeEnum)

	_, httpResp, err := s.getClient().CancelOrderAPI.
		CancelOrder(s.getSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		CancelOrderRequest(*cancelReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "CancelOrder"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已取消: %s, code=%d", orderID, cancelCode)
	return nil
}

// MarkOrderReady 标记订单准备完成
func (s *sGrab) MarkOrderReady(ctx context.Context, orderID string, markStatus string) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	// SDK 使用 int32 markStatus
	var status int32 = 1
	if markStatus == "1" {
		status = 1
	}

	markReq := grabfood.NewMarkOrderRequest(orderID, status)

	httpResp, err := s.getClient().MarkOrderReadyAPI.
		MarkOrderReady(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		MarkOrderRequest(*markReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "MarkOrderReady"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已标记为准备完成: %s, status=%d", orderID, status)
	return nil
}

// UpdateDeliveryState 更新配送状态 (自配送)
func (s *sGrab) UpdateDeliveryState(ctx context.Context, orderID string, fromState string, toState string) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	deliveryReq := grabfood.NewOrderDeliveryRequest(orderID, fromState, toState)

	httpResp, err := s.getClient().UpdateDeliveryStateAPI.
		UpdateDeliveryState(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		OrderDeliveryRequest(*deliveryReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "UpdateDeliveryState"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 配送状态已更新: %s, %s -> %s", orderID, fromState, toState)
	return nil
}

// UpdateOrderReadyTime 更新订单准备时间
func (s *sGrab) UpdateOrderReadyTime(ctx context.Context, orderID string, newReadyTime time.Time) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	readyTimeReq := grabfood.NewNewOrderTimeRequest(orderID, newReadyTime)

	httpResp, err := s.getClient().UpdateOrderReadyTimeAPI.
		UpdateOrderReadyTime(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		NewOrderTimeRequest(*readyTimeReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "UpdateOrderReadyTime"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单准备时间已更新: %s, newTime=%s", orderID, newReadyTime.Format(time.RFC3339))
	return nil
}

// CheckOrderCancelable 检查订单是否可取消
// 返回 Grab SDK 的完整响应对象
func (s *sGrab) CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (*grabfood.CheckOrderCancelableResponse, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := s.getClient().CheckOrderCancelableAPI.
		CheckOrderCancelable(s.getSDKContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		OrderID(orderID).
		Execute()

	if err = s.handleSDKError(ctx, err, "CheckOrderCancelable"); err != nil {
		return nil, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// ============================================================================
// 门店管理 API
// ============================================================================

// PauseStore 暂停门店
func (s *sGrab) PauseStore(ctx context.Context, merchantID string, duration int) error {
	return s.updateStoreStatus(ctx, merchantID, true, duration)
}

// ResumeStore 恢复门店营业
func (s *sGrab) ResumeStore(ctx context.Context, merchantID string) error {
	return s.updateStoreStatus(ctx, merchantID, false, 0)
}

// updateStoreStatus 更新门店状态 (暂停/恢复)
func (s *sGrab) updateStoreStatus(ctx context.Context, merchantID string, isPause bool, duration int) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	pauseReq := grabfood.NewPauseStoreRequest(merchantID, isPause)
	if isPause && duration > 0 {
		// SDK Duration 是 string 类型
		pauseReq.SetDuration(strconv.Itoa(duration))
	}

	httpResp, err := s.getClient().PauseStoreAPI.
		PauseStore(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		PauseStoreRequest(*pauseReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "UpdateStoreStatus"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	action := "resumed"
	if isPause {
		action = fmt.Sprintf("paused for %d minutes", duration)
	}
	g.Log().Infof(ctx, "[Grab] 门店状态已更新: merchant=%s, action=%s", merchantID, action)
	return nil
}

// GetStoreStatus 获取门店状态
func (s *sGrab) GetStoreStatus(ctx context.Context, merchantID string) (*grabfood.StoreStatusResponse, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := s.getClient().GetStoreStatusAPI.
		GetStoreStatus(s.getSDKContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err = s.handleSDKError(ctx, err, "GetStoreStatus"); err != nil {
		return nil, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// GetStoreHours 获取门店营业时间
func (s *sGrab) GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := s.getClient().GetStoreHourAPI.
		GetStoreHour(s.getSDKContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err = s.handleSDKError(ctx, err, "GetStoreHours"); err != nil {
		return nil, err
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
// 此方法实现 MenuNotifier 接口
func (s *sGrab) NotifyMenuUpdate(ctx context.Context, merchantID string) (string, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return "", gerror.Wrap(err, "获取授权信息失败")
	}

	notifReq := grabfood.NewUpdateMenuNotifRequest(merchantID)

	httpResp, err := s.getClient().UpdateMenuNotificationAPI.
		UpdateMenuNotification(s.getSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		UpdateMenuNotifRequest(*notifReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "NotifyMenuUpdate"); err != nil {
		return "", err
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

	g.Log().Infof(ctx, "[Grab] 菜单更新通知已发送: merchant=%s, requestId=%s", merchantID, requestID)
	return requestID, nil
}

// TraceMenuSync 追踪菜单同步状态
func (s *sGrab) TraceMenuSync(ctx context.Context, merchantID string) (*grabfood.MenuSyncResponse, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := s.getClient().TraceMenuSyncAPI.
		TraceMenuSync(s.getSDKContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		Execute()

	if err = s.handleSDKError(ctx, err, "TraceMenuSync"); err != nil {
		return nil, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// UpdateMenuRecord 更新单个菜单记录 (商品或修饰符)
// 调用 GrabFood API PUT /partner/v1/merchants/menu/record
// req 可以是 UpdateMenuItem 或 UpdateMenuModifier
func (s *sGrab) UpdateMenuRecord(ctx context.Context, merchantID string, req grabfood.UpdateMenuRequest) error {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	httpResp, err := s.getClient().UpdateMenuRecordAPI.
		UpdateMenu(s.getSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		UpdateMenuRequest(req).
		Execute()

	if err = s.handleSDKError(ctx, err, "UpdateMenuRecord"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 检查 HTTP 状态码
	if httpResp != nil && httpResp.StatusCode >= 400 {
		return gerror.Newf("Grab API 返回错误: HTTP %d", httpResp.StatusCode)
	}

	g.Log().Infof(ctx, "[Grab] UpdateMenuRecord 成功: merchantID=%s", merchantID)
	return nil
}

// BatchUpdateMenu 批量更新菜单记录 (商品或修饰符)
// 调用 GrabFood API POST /partner/v1/batch/menu
// 参数：
//   - ctx: 上下文对象
//   - merchantID: Grab Merchant ID
//   - req: 批量更新请求（GrabFood SDK 结构体）
//
// 返回：
//   - resp: 批量更新响应，包含状态和错误列表
//   - err: 错误信息
func (s *sGrab) BatchUpdateMenu(ctx context.Context, merchantID string, req *grabfood.BatchUpdateMenuItem) (*grabfood.BatchUpdateMenuResponse, error) {
	// 1. 获取授权信息
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	// 2. 记录开始日志
	g.Log().Infof(ctx, "[Grab] BatchUpdateMenu 开始: merchantID=%s, field=%s, count=%d",
		merchantID, req.GetField(), len(req.GetMenuEntities()))

	// 3. 调用 GrabFood SDK BatchUpdateMenu API
	resp, httpResp, err := s.getClient().UpdateMenuRecordAPI.
		BatchUpdateMenu(s.getSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		BatchUpdateMenuItem(*req).
		Execute()

	// 4. 错误处理
	if err = s.handleSDKError(ctx, err, "BatchUpdateMenu"); err != nil {
		return nil, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 5. 检查 HTTP 状态码
	if httpResp != nil && httpResp.StatusCode >= 400 {
		return nil, gerror.Newf("Grab API 返回错误: HTTP %d", httpResp.StatusCode)
	}

	// 6. 记录成功日志
	status := resp.GetStatus()
	g.Log().Infof(ctx, "[Grab] BatchUpdateMenu 成功: merchantID=%s, status=%s, errorCount=%d",
		merchantID, status, len(resp.GetErrors()))

	return resp, nil
}

// ============================================================================
// 自助激活链接 API
// ============================================================================

// CreateSelfServeJourney 创建自助激活链接
// merchantID: Grab Merchant ID
// 返回: activation_url, request_id
func (s *sGrab) CreateSelfServeJourney(ctx context.Context, merchantID string) (string, string, error) {
	auth, err := s.getAuthorizationHeader(ctx)
	if err != nil {
		return "", "", gerror.Wrap(err, "获取授权信息失败")
	}

	// 构建 Partner 信息
	partnerInfo := grabfood.NewCreateSelfServeJourneyRequestPartner(merchantID)

	// 构建请求
	selfServeReq := grabfood.CreateSelfServeJourneyRequest{
		Partner: *partnerInfo,
	}

	// 调用 SDK API
	resp, httpResp, err := s.getClient().CreateSelfServeJourneyAPI.
		CreateSelfServeJourney(s.getSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		CreateSelfServeJourneyRequest(selfServeReq).
		Execute()

	if err = s.handleSDKError(ctx, err, "CreateSelfServeJourney"); err != nil {
		return "", "", err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 提取 request_id
	requestID := ""
	if httpResp != nil {
		requestID = httpResp.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = httpResp.Header.Get("Request-Id")
		}
	}

	// 提取 activation_url
	activationURL := ""
	if resp != nil {
		if url := resp.GetActivationUrl(); url != "" {
			activationURL = url
		}
	}

	g.Log().Infof(ctx, "[Grab] 自助激活链接已创建: merchant=%s, requestId=%s", merchantID, requestID)
	return activationURL, requestID, nil
}

// ============================================================================
// Token 管理
// ============================================================================

// GetPartnerToken 生成 Grab Partner Token，提供给 Grab 调用
// 校验 client_id 和 client_secret，使用请求中的 scope 生成 Token
func (s *sGrab) GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (accessToken string, expiresIn int, err error) {
	return service.GrabToken().GeneratePartnerToken(ctx, clientID, clientSecret, scope)
}

// ParsePartnerToken 校验并解析 Partner Token
// 用于中间件验证 Grab 发送的请求中携带的 Token
func (s *sGrab) ParsePartnerToken(token string) (*grabDto.PartnerTokenClaims, error) {
	return service.GrabToken().ParsePartnerToken(context.Background(), token)
}

// ============================================================================
// 门店配置
// ============================================================================

// GetShopProviderCfg 查询门店第三方配置
func (s *sGrab) GetShopProviderCfg(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error) {
	return service.ShopProviderCfg().GetShopProviderCfgResp(ctx, req)
}
