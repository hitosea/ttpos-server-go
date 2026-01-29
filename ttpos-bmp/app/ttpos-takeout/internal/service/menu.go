// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
)

type (
	IMenu interface {
		// NotifyMenuUpdate 通知菜单更新（统一路由入口）
		// 根据 provider_name 路由到对应平台的服务
		NotifyMenuUpdate(ctx context.Context, req *menu.NotifyMenuUpdateReq) (*takeout.ApiResponse, error)
	}
)

var (
	localMenu IMenu
)

func Menu() IMenu {
	if localMenu == nil {
		panic("未找到 IMenu 接口实现，是否忘记注册？")
	}
	return localMenu
}

func RegisterMenu(i IMenu) {
	localMenu = i
}
