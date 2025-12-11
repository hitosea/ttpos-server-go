// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/api/grab"
)

type (
	IGrabSelfServe interface {
		// CreateSelfServeJourney 创建自助激活链接
		// 根据 shop_uuid 获取 Grab 配置，调用 SDK 生成激活链接
		CreateSelfServeJourney(ctx context.Context, req *grab.CreateSelfServeJourneyReq) (*grab.CreateSelfServeJourneyResp, error)
	}
)

var (
	localGrabSelfServe IGrabSelfServe
)

func GrabSelfServe() IGrabSelfServe {
	if localGrabSelfServe == nil {
		panic("implement not found for interface IGrabSelfServe, forgot register?")
	}
	return localGrabSelfServe
}

func RegisterGrabSelfServe(i IGrabSelfServe) {
	localGrabSelfServe = i
}
