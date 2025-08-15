// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
)

type (
	IBom interface {
		// GetBomList 获取BOM列表
		// 参数：ctx 上下文，req 获取BOM列表请求
		// 返回：BOM列表响应，错误信息
		GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (res *manufacturing.GetBomListResp, err error)
		// SaveBom 保存BOM信息
		// 参数：ctx 上下文，req 保存BOM请求
		// 返回：保存BOM响应，错误信息
		SaveBom(ctx context.Context, req *manufacturing.SaveBomReq) (res *manufacturing.SaveBomResp, err error)
	}
)

var (
	localBom IBom
)

func Bom() IBom {
	if localBom == nil {
		panic("implement not found for interface IBom, forgot register?")
	}
	return localBom
}

func RegisterBom(i IBom) {
	localBom = i
}
