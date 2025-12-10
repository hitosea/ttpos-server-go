// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
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
		HandleSubmitOrder(ctx context.Context, signature string, timestamp string, body []byte) error
		// HandlePushOrderState 处理订单状态变更 Webhook
		HandlePushOrderState(ctx context.Context, signature string, timestamp string, body []byte) error
		// HandleGetMenu 处理 Grab 获取菜单请求
		HandleGetMenu(ctx context.Context, signature string, timestamp string, merchantID string) (*grabDto.GetMenuResponse, error)
		// HandleMenuSyncState 处理菜单同步状态回调
		HandleMenuSyncState(ctx context.Context, signature string, timestamp string, body []byte) error
		// SyncMenu 主动同步菜单到 Grab
		SyncMenu(ctx context.Context, merchantID string, menu *grabDto.GetMenuResponse) error
		// HandleIntegrationStatus 处理门店集成状态回调
		HandleIntegrationStatus(ctx context.Context, signature string, timestamp string, body []byte) error
		// PauseStore 暂停门店
		PauseStore(ctx context.Context, merchantID string, duration int) error
		// ResumeStore 恢复门店营业
		ResumeStore(ctx context.Context, merchantID string) error
		// AcceptOrder 接受订单
		AcceptOrder(ctx context.Context, orderID string) error
		// RejectOrder 拒绝订单
		RejectOrder(ctx context.Context, orderID string, rejectCode int) error
		// MarkOrderReady 标记订单准备完成
		MarkOrderReady(ctx context.Context, orderID string, markStatus string) error
		// CancelOrder 取消订单
		CancelOrder(ctx context.Context, orderID string, cancelCode int) error
		// GetPartnerToken 生成 Grab Partner Token，提供给 Grab 调用
		// 校验 client_id 和 client_secret，使用请求中的 scope 生成 Token
		GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (accessToken string, expiresIn int, err error)
		// ParsePartnerToken 校验并解析 Partner Token
		// 用于中间件验证 Grab 发送的请求中携带的 Token
		ParsePartnerToken(token string) (*grabDto.PartnerTokenClaims, error)
		// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
		HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error
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
