// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	IChannelMenu interface {
		// SaveChannelMenu 保存外卖渠道菜单快照
		SaveChannelMenu(ctx context.Context, shopUUID uint64, providerName string, menuData string) error
		// GetChannelMenu 读取外卖渠道菜单快照
		GetChannelMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error)
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
