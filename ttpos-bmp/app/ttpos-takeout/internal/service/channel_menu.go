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
		// GetMenuSnapshot 根据 request_id 查询菜单快照
		GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*api.GetMenuSnapshotResp, error)
		// SaveMenuSnapshot 保存菜单快照
		SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*api.SaveMenuSnapshotResp, error)
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
