// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
)

type (
	ISelling interface {
		// GetPosProfileList 查询Pos Profile列表
		// 参数：ctx 上下文，req 查询请求
		// 返回：erp.ResponseInfo，错误信息
		GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error)
		CreateModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) (err error)
		// CreatePosProfile CreatePosFile 创建 默认 pos profile  配置默认 posprofile
		CreatePosProfile(ctx context.Context, req *setup.CreatePosProfileInp) (*erp.POSProfile, error)
		// OpenPosEntry 开帐
		OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error)
		// ClosePosEntry 关帐
		ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
		// IsProfileOpening 查询pos profile是否开帐
		IsProfileOpening(ctx context.Context, posProfile string) (bool, error)
	}
)

var (
	localSelling ISelling
)

func Selling() ISelling {
	if localSelling == nil {
		panic("implement not found for interface ISelling, forgot register?")
	}
	return localSelling
}

func RegisterSelling(i ISelling) {
	localSelling = i
}
