// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

type (
	IGrabMenu interface {
		// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
		// 签名验证已由中间件完成
		HandleGetMenu(ctx context.Context, partnerMerchantID string) (*grabfood.GetMenuNewResponse, error)
		// HandleMenuSyncState 处理菜单同步状态回调
		// 使用 SDK grabfood.MenuSyncWebhookRequest
		HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error
		// SyncMenu 主动同步菜单到 Grab
		SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse, notifier grabDto.MenuNotifier) error
		// SaveMenuSnapshot 保存菜单快照到数据库
		// 使用 shop_uuid + provider_name 作为唯一键，存在则更新，不存在则插入
		SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (uint64, error)
		// NotifyMenuUpdate 发送菜单更新通知 (RocketMQ)
		NotifyMenuUpdate(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error
	}
)

var (
	localGrabMenu IGrabMenu
)

func GrabMenu() IGrabMenu {
	if localGrabMenu == nil {
		panic("implement not found for interface IGrabMenu, forgot register?")
	}
	return localGrabMenu
}

func RegisterGrabMenu(i IGrabMenu) {
	localGrabMenu = i
}
