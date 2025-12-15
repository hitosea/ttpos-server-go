// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
)

type (
	ITakeout interface {
		// Get 根据 shop_order_uuid 查询订单（兼容旧接口，返回 SkootarJob DTO）
		// 注意：此方法为兼容性方法，实际查询新表结构并转换为 DTO 格式
		Get(ctx context.Context, shopOrderUuid string) (*dto.SkootarJob, error)
		// GetWithDriver 根据 shop_order_uuid 查询订单及司机信息（新方法）
		GetWithDriver(ctx context.Context, shopOrderUuid string) (*dto.OrderWithDriver, error)
		// GetMenuSnapshot 根据 request_id 查询菜单快照
		GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*api.GetMenuSnapshotResp, error)
		// SaveMenuSnapshot 保存菜单快照
		SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*api.SaveMenuSnapshotResp, error)
	}
)

var (
	localTakeout ITakeout
)

func Takeout() ITakeout {
	if localTakeout == nil {
		panic("implement not found for interface ITakeout, forgot register?")
	}
	return localTakeout
}

func RegisterTakeout(i ITakeout) {
	localTakeout = i
}
