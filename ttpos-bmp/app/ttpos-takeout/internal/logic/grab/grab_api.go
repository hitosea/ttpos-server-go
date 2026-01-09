// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/client/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// ============================================================================
// 订单操作 API
// ============================================================================

// AcceptOrder 接受订单
func (s *sGrab) AcceptOrder(ctx context.Context, orderID string) error {
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	acceptReq := grabfood.NewAcceptOrderRequest(orderID, "Accepted")

	httpResp, err := client.GetClient().AcceptRejectOrderAPI.
		AcceptRejectOrder(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*acceptReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "AcceptOrder"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	rejectReq := grabfood.NewAcceptOrderRequest(orderID, "Rejected")

	httpResp, err := client.GetClient().AcceptRejectOrderAPI.
		AcceptRejectOrder(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		AcceptOrderRequest(*rejectReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "RejectOrder"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已拒绝: %s, code=%d", orderID, rejectCode)
	return nil
}

// CancelOrder 取消订单
func (s *sGrab) CancelOrder(ctx context.Context, merchantID string, orderID string, cancelCode string) error {
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	// 将 cancelCode 字符串转换为 int（Grab SDK CancelCode 是 int32 类型）
	cancelCodeInt, err := strconv.Atoi(cancelCode)
	if err != nil {
		return gerror.Wrapf(err, "取消原因码格式错误: %s", cancelCode)
	}

	// SDK CancelCode 是 int32 类型
	cancelCodeEnum := grabfood.CancelCode(cancelCodeInt)
	cancelReq := grabfood.NewCancelOrderRequest(orderID, merchantID, cancelCodeEnum)

	_, httpResp, err := client.GetClient().CancelOrderAPI.
		CancelOrder(client.GetSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		CancelOrderRequest(*cancelReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "CancelOrder"); err != nil {
		return err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	g.Log().Infof(ctx, "[Grab] 订单已取消: merchant_id=%s, order_id=%s, code=%s", merchantID, orderID, cancelCode)
	return nil
}

// MarkOrderReady 标记订单准备完成
func (s *sGrab) MarkOrderReady(ctx context.Context, orderID string, markStatus string) error {
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	// SDK 使用 int32 markStatus
	var status int32 = 1
	if markStatus == "1" {
		status = 1
	}
	if markStatus == "2" {
		status = 2
	}

	markReq := grabfood.NewMarkOrderRequest(orderID, status)

	httpResp, err := client.GetClient().MarkOrderReadyAPI.
		MarkOrderReady(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		MarkOrderRequest(*markReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "MarkOrderReady"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	deliveryReq := grabfood.NewOrderDeliveryRequest(orderID, fromState, toState)

	httpResp, err := client.GetClient().UpdateDeliveryStateAPI.
		UpdateDeliveryState(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		OrderDeliveryRequest(*deliveryReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "UpdateDeliveryState"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	readyTimeReq := grabfood.NewNewOrderTimeRequest(orderID, newReadyTime)

	httpResp, err := client.GetClient().UpdateOrderReadyTimeAPI.
		UpdateOrderReadyTime(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		NewOrderTimeRequest(*readyTimeReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "UpdateOrderReadyTime"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := client.GetClient().CheckOrderCancelableAPI.
		CheckOrderCancelable(client.GetSDKContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		OrderID(orderID).
		Execute()

	if err = client.HandleSDKError(ctx, err, "CheckOrderCancelable"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	pauseReq := grabfood.NewPauseStoreRequest(merchantID, isPause)
	if isPause && duration > 0 {
		// SDK Duration 是 string 类型
		pauseReq.SetDuration(strconv.Itoa(duration))
	}

	httpResp, err := client.GetClient().PauseStoreAPI.
		PauseStore(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		PauseStoreRequest(*pauseReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "UpdateStoreStatus"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := client.GetClient().GetStoreStatusAPI.
		GetStoreStatus(client.GetSDKContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err = client.HandleSDKError(ctx, err, "GetStoreStatus"); err != nil {
		return nil, err
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	return resp, nil
}

// GetStoreHours 获取门店营业时间
func (s *sGrab) GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error) {
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := client.GetClient().GetStoreHourAPI.
		GetStoreHour(client.GetSDKContext(ctx), merchantID).
		Authorization(auth).
		Execute()

	if err = client.HandleSDKError(ctx, err, "GetStoreHours"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return "", gerror.Wrap(err, "获取授权信息失败")
	}

	notifReq := grabfood.NewUpdateMenuNotifRequest(merchantID)

	httpResp, err := client.GetClient().UpdateMenuNotificationAPI.
		UpdateMenuNotification(client.GetSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		UpdateMenuNotifRequest(*notifReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "NotifyMenuUpdate"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	resp, httpResp, err := client.GetClient().TraceMenuSyncAPI.
		TraceMenuSync(client.GetSDKContext(ctx)).
		Authorization(auth).
		MerchantID(merchantID).
		Execute()

	if err = client.HandleSDKError(ctx, err, "TraceMenuSync"); err != nil {
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
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return gerror.Wrap(err, "获取授权信息失败")
	}

	httpResp, err := client.GetClient().UpdateMenuRecordAPI.
		UpdateMenu(client.GetSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		UpdateMenuRequest(req).
		Execute()

	if err = client.HandleSDKError(ctx, err, "UpdateMenuRecord"); err != nil {
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
	client := grab.Default()
	// 1. 获取授权信息
	auth, err := client.GetAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "获取授权信息失败")
	}

	// 2. 记录开始日志
	g.Log().Infof(ctx, "[Grab] BatchUpdateMenu 开始: merchantID=%s, field=%s, count=%d",
		merchantID, req.GetField(), len(req.GetMenuEntities()))

	// 3. 调用 GrabFood SDK BatchUpdateMenu API
	resp, httpResp, err := client.GetClient().UpdateMenuRecordAPI.
		BatchUpdateMenu(client.GetSDKContext(ctx)).
		ContentType("application/json").
		Authorization(auth).
		BatchUpdateMenuItem(*req).
		Execute()

	// 4. 错误处理
	if err = client.HandleSDKError(ctx, err, "BatchUpdateMenu"); err != nil {
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
// merchantID: Grab Merchant ID (shop_uuid)
// 返回: activation_url, request_id
func (s *sGrab) CreateSelfServeJourney(ctx context.Context, merchantID string) (string, string, error) {
	// 1. 配置验证
	conf := MustConfig(ctx)
	if conf == nil {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "Grab 平台配置未加载，请检查配置文件 app.provider.grab.platform")
	}
	if conf.ClientID == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 ClientID，请检查配置文件 app.provider.grab.platform.clientId")
	}
	if conf.ClientSecret == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 ClientSecret，请检查配置文件 app.provider.grab.platform.clientSecret")
	}
	if conf.Environment == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 Environment，请检查配置文件 app.provider.grab.platform.environment")
	}

	// 2. 调用 SDK API
	client := grab.Default()
	auth, err := client.GetAuthorizationHeader(ctx)
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
	resp, httpResp, err := client.GetClient().CreateSelfServeJourneyAPI.
		CreateSelfServeJourney(client.GetSDKContext(ctx)).
		Authorization(auth).
		ContentType("application/json").
		CreateSelfServeJourneyRequest(selfServeReq).
		Execute()

	if err = client.HandleSDKError(ctx, err, "CreateSelfServeJourney"); err != nil {
		g.Log().Errorf(ctx, "[Grab] 创建自助激活链接失败: merchant=%s, error=%v", merchantID, err)
		// 错误映射：根据错误类型返回不同的错误码
		return "", "", s.mapGrabError(err)
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

	// 3. 旅程创建成功，落库 shop_provider_cfg（状态 SYNCING）
	shopUUID := g.NewVar(merchantID).Uint64()
	if shopUUID > 0 {
		if upsertErr := service.ShopProviderCfg().UpsertShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab), "", consts.ProviderShopStatusSyncing); upsertErr != nil {
			// 记录错误但不中断流程，旅程创建已成功
			g.Log().Warningf(ctx, "[Grab] 创建自助激活链接时更新门店第三方配置失败: shop_uuid=%d, error=%v", shopUUID, upsertErr)
		}
	}

	g.Log().Infof(ctx, "[Grab] 自助激活链接已创建: merchant=%s, requestId=%s", merchantID, requestID)
	return activationURL, requestID, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// mapGrabError 映射 Grab SDK 错误到业务错误
func (s *sGrab) mapGrabError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	// 授权错误
	if contains(errStr, "401") || contains(errStr, "unauthorized") || contains(errStr, "authentication") {
		return gerror.NewCode(gcode.CodeNotAuthorized, fmt.Sprintf("Grab 授权失败: %v", err))
	}

	// 参数错误
	if contains(errStr, "400") || contains(errStr, "bad request") || contains(errStr, "invalid") {
		return gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("Grab 参数错误: %v", err))
	}

	// 网络超时
	if contains(errStr, "timeout") || contains(errStr, "deadline exceeded") {
		return gerror.NewCode(gcode.CodeInternalError, fmt.Sprintf("Grab API 调用超时: %v", err))
	}

	// 其他错误
	return gerror.Wrap(err, "Grab API 调用失败")
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
