// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	rpc "ttpos-bmp/app/ttpos-manager/api/manager"
	"ttpos-bmp/app/ttpos-manager/api/svc"
)

type (
	ISetting interface {
		GetSetting(ctx context.Context, req *svc.GetReq) (setting *rpc.Setting, err error)
	}
)

var (
	localSetting ISetting
)

func Setting() ISetting {
	if localSetting == nil {
		panic("implement not found for interface ISetting, forgot register?")
	}
	return localSetting
}

func RegisterSetting(i ISetting) {
	localSetting = i
}
