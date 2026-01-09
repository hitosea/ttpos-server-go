// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

type (
	ILineman interface {
		// SyncMenu 同步菜单到 Lineman
		// 参数:
		//   - ctx: 上下文
		//   - shopUUID: 门店UUID
		//
		// 返回:
		//   - error: 错误信息
		SyncMenu(ctx context.Context, shopUUID uint64) error
		// BuildMenuPayload 构建菜单数据
		// 参数:
		//   - ctx: 上下文
		//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串
		//
		// 返回:
		//   - *lineman.MenuSyncRequest: 菜单数据
		//   - error: 错误信息
		BuildMenuPayload(ctx context.Context, ttposMenuJSON string) (*lineman.MenuSyncRequest, error)
	}
)

var (
	localLineman ILineman
)

func Lineman() ILineman {
	if localLineman == nil {
		panic("implement not found for interface ILineman, forgot register?")
	}
	return localLineman
}

func RegisterLineman(i ILineman) {
	localLineman = i
}
