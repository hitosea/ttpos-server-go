// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

var (
	// Grab Grab服务实例
	Grab     = new(sGrab)
	config   *conf.Grab
	configMu sync.RWMutex
)

type sGrab struct {
	verifier   *SignatureVerifier
	sdkWrapper *SDKWrapper // 官方 SDK Wrapper
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

// getSdkWrapper 获取 SDK Wrapper (官方 SDK 实现)
func (s *sGrab) getSdkWrapper() *SDKWrapper {
	if s.sdkWrapper == nil {
		conf := s.MustConf()
		s.sdkWrapper = NewSDKWrapper(&SDKConfig{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			Environment:  conf.Environment,
		})
	}
	return s.sdkWrapper
}

// getVerifier 获取签名验证器 (懒加载单例)
func (s *sGrab) getVerifier() *SignatureVerifier {
	if s.verifier == nil {
		conf := s.MustConf()
		s.verifier = newSignatureVerifier(conf.SecretKey)
	}
	return s.verifier
}

// VerifyWebhookSignature 验证 Grab Webhook 签名 (公开方法，供其他服务调用)
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (s *sGrab) VerifyWebhookSignature(ctx context.Context, signature, timestamp string, body []byte) error {
	return s.getVerifier().VerifySignature(signature, timestamp, body)
}

// HandleSubmitOrder 处理 Grab 提交订单 Webhook
// 签名验证已由中间件完成
func (s *sGrab) HandleSubmitOrder(ctx context.Context, body []byte) error {
	return service.GrabOrder().HandleSubmitOrder(ctx, body)
}

// HandlePushOrderState 处理订单状态变更 Webhook
// 签名验证已由中间件完成
func (s *sGrab) HandlePushOrderState(ctx context.Context, body []byte) error {
	return service.GrabOrder().HandlePushOrderState(ctx, body)
}

// HandleGetMenu 处理 Grab 获取菜单请求
// 签名验证已由中间件完成
func (s *sGrab) HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error) {
	return service.GrabMenu().HandleGetMenu(ctx, merchantID)
}

// HandleMenuSyncState 处理菜单同步状态回调
func (s *sGrab) HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error {
	return service.GrabMenu().HandleMenuSyncState(ctx, req)
}

// SyncMenu 主动同步菜单到 Grab
func (s *sGrab) SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse) error {
	// SDKWrapper 已实现 MenuNotifier 接口
	return service.GrabMenu().SyncMenu(ctx, merchantID, menu, s.getSdkWrapper())
}

// HandleIntegrationStatus 处理门店集成状态回调
// 签名验证已由中间件完成
func (s *sGrab) HandleIntegrationStatus(ctx context.Context, body []byte) error {
	return service.GrabStore().HandleIntegrationStatus(ctx, body)
}

// PauseStore 暂停门店
func (s *sGrab) PauseStore(ctx context.Context, merchantID string, duration int) error {
	return s.getSdkWrapper().UpdateStoreStatus(ctx, merchantID, true, duration)
}

// ResumeStore 恢复门店营业
func (s *sGrab) ResumeStore(ctx context.Context, merchantID string) error {
	return s.getSdkWrapper().UpdateStoreStatus(ctx, merchantID, false, 0)
}

// AcceptOrder 接受订单
func (s *sGrab) AcceptOrder(ctx context.Context, orderID string) error {
	return s.getSdkWrapper().AcceptOrder(ctx, orderID)
}

// RejectOrder 拒绝订单
func (s *sGrab) RejectOrder(ctx context.Context, orderID string, rejectCode int) error {
	return s.getSdkWrapper().RejectOrder(ctx, orderID, rejectCode)
}

// MarkOrderReady 标记订单准备完成
func (s *sGrab) MarkOrderReady(ctx context.Context, orderID string, markStatus string) error {
	// SDK 使用 int32 markStatus，需要转换
	// markStatus: "1" = ready for pickup
	var status int32 = 1
	if markStatus == "1" {
		status = 1
	}
	return s.getSdkWrapper().MarkOrderReady(ctx, orderID, status)
}

// CancelOrder 取消订单
func (s *sGrab) CancelOrder(ctx context.Context, orderID string, cancelCode int) error {
	// SDK 需要 merchantID，但当前接口没有传入，从订单查询
	// TODO: 优化接口，传入 merchantID
	return s.getSdkWrapper().CancelOrder(ctx, orderID, "", cancelCode)
}

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

// GetShopProviderCfg 查询门店第三方配置
func (s *sGrab) GetShopProviderCfg(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error) {
	return service.ShopProviderCfg().GetShopProviderCfgResp(ctx, req)
}
