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
		GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (res *manufacturing.GetBomListResp, err error)
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
