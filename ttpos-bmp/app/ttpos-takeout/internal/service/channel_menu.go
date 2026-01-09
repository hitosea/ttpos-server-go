// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/menu"
)

type (
	IChannelMenu interface {
		// SaveChannelMenu 保存外卖渠道菜单快照
		SaveChannelMenu(ctx context.Context, shopUUID uint64, providerName string, menuData string) error
		// GetChannelMenu 读取外卖渠道菜单快照
		GetChannelMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error)
		// GetTtposMenu 读取TTPOS菜单快照
		GetTtposMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error)
		// GetMenuSnapshot
		GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*api.GetMenuSnapshotResp, error)
		// SaveMenuSnapshot 保存菜单快照
		SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*api.SaveMenuSnapshotResp, error)
		// LogMenuSync 记录菜单同步日志
		// 通用方法，供各个渠道（grab、lineman 等）调用
		// 参数：
		//   - ctx: 上下文
		//   - merchantID: 商户ID
		//   - providerName: 渠道名称（grab/lineman）
		//   - syncType: 同步类型（FULL/PARTIAL/BATCH_UPDATE_ITEM等）
		//   - requestID: 请求ID（来自第三方API响应）
		//   - success: 是否成功
		//   - menuSnapshot: 菜单快照（JSON 字符串，可选）
		//   - errMsg: 错误信息（失败时）
		LogMenuSync(ctx context.Context, merchantID string, providerName string, syncType string, requestID string, success bool, menuSnapshot string, errMsg string) error
	}
)

var (
	localChannelMenu IChannelMenu
)

func ChannelMenu() IChannelMenu {
	if localChannelMenu == nil {
		panic("implement not found for interface IChannelMenu, forgot register?")
	}
	return localChannelMenu
}

func RegisterChannelMenu(i IChannelMenu) {
	localChannelMenu = i
}
