// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	v1 "ttpos-bmp/app/ttpos-takeout/api/callback/v1"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
)

type (
	ISkootar interface {
		// CancelOrder 取消订单
		CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error)
		// ConfirmOrder 商家确认订单
		ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error)
		// CreateOrder 创建订单
		CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error)
		// EstimateDistance 获取预估距离
		EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error)
		// GetDriverInfo 获取司机位置
		GetDriverInfo(ctx context.Context, req *dto.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error)
		JobDetail4Food(ctx context.Context, req *skootar.JobDetailInp) (jobDetail *skootar.JobDetail, err error)
		JobStatusChange(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error)
		MustConf() *conf.Skootar
		// GetUrl 获取Skootar API URL
		GetUrl(apiPath string) string
		// ReqBase 获取请求基础参数
		ReqBase() skootar.ReqBase
	}
)

var (
	localSkootar ISkootar
)

func Skootar() ISkootar {
	if localSkootar == nil {
		panic("implement not found for interface ISkootar, forgot register?")
	}
	return localSkootar
}

func RegisterSkootar(i ISkootar) {
	localSkootar = i
}
