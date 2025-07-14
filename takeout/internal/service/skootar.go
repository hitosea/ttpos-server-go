// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"takeout/api"
	"takeout/internal/model/conf"
	"takeout/internal/model/input/skootar"
)

type (
	ISkootar interface {
		EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error)
		MustConf() *conf.Skootar
		GetUrl(apiPath string) string
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
