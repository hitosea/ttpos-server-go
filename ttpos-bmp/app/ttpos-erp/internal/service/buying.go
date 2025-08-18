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
		// GetSupplierList 获取供应商列表
		GetSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error)
		// CreatePurchaseFromMq 根据材料请求创建采购订单
		CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error)
		// CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
		CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error)
		// GetPurchaseOrder 获取采购订单
		GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error)
		// CreatePurchaseReceiptFromOrder 创建采购收货订单
		CreatePurchaseReceiptFromOrder(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*erp.PurchaseReceipt, error)
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
