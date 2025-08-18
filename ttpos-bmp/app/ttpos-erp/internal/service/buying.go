// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	IBuying interface {
		GetSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error)
		CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error)
	}
)

var (
	localBuying IBuying
)

func Buying() IBuying {
	if localBuying == nil {
		panic("implement not found for interface IBuying, forgot register?")
	}
	return localBuying
}

func RegisterBuying(i IBuying) {
	localBuying = i
}
