// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
)

type (
	ISelling interface {
		// GetPosProfileList 查询Pos Profile列表
		// 参数：ctx 上下文，req 查询请求
		// 返回：erp.ResponseInfo，错误信息
		GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error)
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
