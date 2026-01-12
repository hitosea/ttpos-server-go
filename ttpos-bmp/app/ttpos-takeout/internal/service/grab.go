// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"time"
	"ttpos-bmp/app/ttpos-takeout/api/grab"
	api "ttpos-bmp/app/ttpos-takeout/api/order"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

type (
	IGrab interface {
		// ========================================================================
		// 配置管理
		// ========================================================================

		// MustConf 获取 Grab 配置
		MustConf() *conf.Grab

		// ========================================================================
		// Webhook 处理
		// ========================================================================

		// VerifyWebhookSignature 验证 Grab Webhook 签名 (公开方法，供其他服务调用)
		// signature: X-Grab-Signature 请求头值
		// timestamp: X-Grab-Timestamp 请求头值
		// body: 请求体原始字节
		VerifyWebhookSignature(ctx context.Context, signature string, timestamp string, body []byte) error
		// HandleSubmitOrder 处理 Grab 提交订单 Webhook
		// 签名验证已由中间件完成
		HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error
		// HandlePushOrderState 处理订单状态变更 Webhook
		// 签名验证已由中间件完成
		HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error
		// HandleGetMenu 处理 Grab 获取菜单请求
		// 签名验证已由中间件完成
		HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error)
		// HandleMenuSyncState 处理菜单同步状态回调
		HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error
		// SyncMenu 主动同步菜单到 Grab
		// SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse) error
		// HandleIntegrationStatus 处理门店集成状态回调
		// 签名验证已由中间件完成
		HandleIntegrationStatus(ctx context.Context, body []byte) error
		// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
		HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error

		// ========================================================================
		// 订单操作 API
		// ========================================================================

		// AcceptOrder 接受订单
		AcceptOrder(ctx context.Context, orderID string) error
		// RejectOrder 拒绝订单
		RejectOrder(ctx context.Context, orderID string, rejectCode int) error
		// CancelOrder 取消订单
		CancelOrder(ctx context.Context, merchantID string, orderID string, cancelCode string) error
		// MarkOrderReady 标记订单准备完成
		MarkOrderReady(ctx context.Context, orderID string, markStatus string) error
		// UpdateDeliveryState 更新配送状态 (自配送)
		UpdateDeliveryState(ctx context.Context, orderID string, fromState string, toState string) error
		// UpdateOrderReadyTime 更新订单准备时间
		UpdateOrderReadyTime(ctx context.Context, orderID string, newReadyTime time.Time) error
		// CheckOrderCancelable 检查订单是否可取消
		// 返回 Grab SDK 的完整响应对象
		CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (*grabfood.CheckOrderCancelableResponse, error)
		
		// ========================================================================
		// 订单业务逻辑 (原 IGrabOrder)
		// ========================================================================
		
		// PrepareOrder 准备订单（接受/拒绝）
		PrepareOrder(ctx context.Context, orderEntityInterface interface{}, toState string) error
		// MarkOrderReadyEntity 标记订单准备完成 (使用实体对象)
		MarkOrderReadyEntity(ctx context.Context, orderEntity *entity.Order) error
		// CheckOrderCancelableEntity 检查订单是否可取消 (使用实体对象)
		CheckOrderCancelableEntity(ctx context.Context, orderEntity *entity.Order) (*api.CheckOrderCancelableResp, error)
		// CancelOrderEntity 取消订单 (使用实体对象)
		CancelOrderEntity(ctx context.Context, orderEntity *entity.Order, cancelCode string) (res *api.CancelOrderResp, err error)
		
		// ========================================================================
		// 门店管理 API
		// ========================================================================
		
		// PauseStore 暂停门店
		PauseStore(ctx context.Context, merchantID string, duration int) error
		// ResumeStore 恢复门店营业
		ResumeStore(ctx context.Context, merchantID string) error
		// GetStoreStatus 获取门店状态
		GetStoreStatus(ctx context.Context, merchantID string) (*grabfood.StoreStatusResponse, error)
		// GetStoreHours 获取门店营业时间
		GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error)
		
		// ========================================================================
		// 菜单管理 API
		// ========================================================================
		
		// NotifyMenuUpdate 通知 Grab 菜单已更新
		// 返回 requestID 用于追踪同步状态
		// 此方法实现 MenuNotifier 接口
		NotifyMenuUpdate(ctx context.Context, merchantID string) (string, error)
		// TraceMenuSync 追踪菜单同步状态
		TraceMenuSync(ctx context.Context, merchantID string) (*grabfood.MenuSyncResponse, error)
		// UpdateMenuRecord 更新单个菜单记录 (商品或修饰符)
		// 调用 GrabFood API PUT /partner/v1/merchants/menu/record
		// req 可以是 UpdateMenuItem 或 UpdateMenuModifier
		UpdateMenuRecord(ctx context.Context, merchantID string, req grabfood.UpdateMenuRequest) error
		// BatchUpdateMenu 批量更新菜单记录 (商品或修饰符)
		// 调用 GrabFood API POST /partner/v1/batch/menu
		BatchUpdateMenu(ctx context.Context, merchantID string, req *grabfood.BatchUpdateMenuItem) (*grabfood.BatchUpdateMenuResponse, error)
		
		// ========================================================================
		// 菜单业务逻辑 (原 IGrabMenu)
		// ========================================================================
		
		// HandleGetMenuInternal 处理 Grab 获取菜单请求 (Partner Endpoint)
		HandleGetMenuInternal(ctx context.Context, partnerMerchantID string) (*grabfood.GetMenuNewResponse, error)
		// SaveMenuSnapshot 保存菜单快照到数据库
		SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (uint64, error)
		// NotifyMenuUpdateEvent 发送菜单更新通知 (RocketMQ)
		NotifyMenuUpdateEvent(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error
		// UpdateMenuItem 更新单个菜单项 (商品)
		UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error
		// UpdateMenuModifier 更新单个修饰符
		UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error
		// BatchUpdateMenuItems 批量更新菜单记录 (商品或修饰符)
		BatchUpdateMenuItems(ctx context.Context, req *grabDto.BatchUpdateMenuReq) (*grabDto.BatchUpdateMenuResp, error)
		
		// ========================================================================
		// 自助激活 API
		// ========================================================================
		
		// CreateSelfServeJourney 创建自助激活链接
		// merchantID: Grab Merchant ID
		// 返回: activation_url, request_id
		CreateSelfServeJourney(ctx context.Context, merchantID string) (string, string, error)
		
		// ========================================================================
		// 自助激活业务逻辑 (原 IGrabSelfServe)
		// ========================================================================
		
		// CreateSelfServeJourneyWithReq 创建自助激活链接 (使用请求对象)
		CreateSelfServeJourneyWithReq(ctx context.Context, req *grab.CreateSelfServeJourneyReq) (*grab.CreateSelfServeJourneyResp, error)

		// ========================================================================
		// Token 管理
		// ========================================================================

		// GetPartnerToken 生成 Grab Partner Token，提供给 Grab 调用
		GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (accessToken string, expiresIn int, err error)
		// ParsePartnerToken 校验并解析 Partner Token
		ParsePartnerToken(token string) (*grabDto.PartnerTokenClaims, error)
		// GetPartnerConfig 通过 partner code 获取配置
		GetPartnerConfig(ctx context.Context, code string) (*conf.GrabPartner, error)
		// GetPartnerConfigByClientID 通过 client_id 获取配置
		GetPartnerConfigByClientID(ctx context.Context, clientID string) (*conf.GrabPartner, error)
		
		// ========================================================================
		// 门店配置
		// ========================================================================
		
		// GetShopProviderCfg 查询门店第三方配置
		GetShopProviderCfg(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error)
	}
)

var (
	localGrab IGrab
)

func Grab() IGrab {
	if localGrab == nil {
		panic("implement not found for interface IGrab, forgot register?")
	}
	return localGrab
}

func RegisterGrab(i IGrab) {
	localGrab = i
}
