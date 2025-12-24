// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"time"
	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

type (
	IGrab interface {
		// MustConf 获取 Grab 配置
		MustConf() *conf.Grab
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
		SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse) error
		// HandleIntegrationStatus 处理门店集成状态回调
		// 签名验证已由中间件完成
		HandleIntegrationStatus(ctx context.Context, body []byte) error
		// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
		HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error
		// AcceptOrder 接受订单
		AcceptOrder(ctx context.Context, orderID string) error
		// RejectOrder 拒绝订单
		RejectOrder(ctx context.Context, orderID string, rejectCode int) error
		// CancelOrder 取消订单
		CancelOrder(ctx context.Context, orderID string, cancelCode int) error
		// MarkOrderReady 标记订单准备完成
		MarkOrderReady(ctx context.Context, orderID string, markStatus string) error
		// UpdateDeliveryState 更新配送状态 (自配送)
		UpdateDeliveryState(ctx context.Context, orderID string, fromState string, toState string) error
		// UpdateOrderReadyTime 更新订单准备时间
		UpdateOrderReadyTime(ctx context.Context, orderID string, newReadyTime time.Time) error
		// CheckOrderCancelable 检查订单是否可取消
		CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (bool, string, error)
		// PauseStore 暂停门店
		PauseStore(ctx context.Context, merchantID string, duration int) error
		// ResumeStore 恢复门店营业
		ResumeStore(ctx context.Context, merchantID string) error
		// GetStoreStatus 获取门店状态
		GetStoreStatus(ctx context.Context, merchantID string) (*grabfood.StoreStatusResponse, error)
		// GetStoreHours 获取门店营业时间
		GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error)
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
		// 参数：
		//   - ctx: 上下文对象
		//   - merchantID: Grab Merchant ID
		//   - req: 批量更新请求（GrabFood SDK 结构体）
		//
		// 返回：
		//   - resp: 批量更新响应，包含状态和错误列表
		//   - err: 错误信息
		BatchUpdateMenu(ctx context.Context, merchantID string, req *grabfood.BatchUpdateMenuItem) (*grabfood.BatchUpdateMenuResponse, error)
		// CreateSelfServeJourney 创建自助激活链接
		// merchantID: Grab Merchant ID
		// 返回: activation_url, request_id
		CreateSelfServeJourney(ctx context.Context, merchantID string) (string, string, error)
		// GetPartnerToken 生成 Grab Partner Token，提供给 Grab 调用
		// 校验 client_id 和 client_secret，使用请求中的 scope 生成 Token
		GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (accessToken string, expiresIn int, err error)
		// ParsePartnerToken 校验并解析 Partner Token
		// 用于中间件验证 Grab 发送的请求中携带的 Token
		ParsePartnerToken(token string) (*grabDto.PartnerTokenClaims, error)
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
