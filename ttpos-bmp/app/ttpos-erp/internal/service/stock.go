// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
)

type (
	IItem interface {
		SyncDelay()
	}
	IStock interface {
		GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error)
		SaveUom(ctx context.Context, req *item.UomInfo) (err error)
		GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *item.GetAttributeListResp, err error)
		GetAttributeValuesList(ctx context.Context, attributeName string) (res []*item.AttributeValueInfo, err error)
		SaveAttribute(ctx context.Context, req *item.AttributeInfo) (err error)
	}
)

var (
	localItem  IItem
	localStock IStock
)

func Item() IItem {
	if localItem == nil {
		panic("implement not found for interface IItem, forgot register?")
	}
	return localItem
}

func RegisterItem(i IItem) {
	localItem = i
}

func Stock() IStock {
	if localStock == nil {
		panic("implement not found for interface IStock, forgot register?")
	}
	return localStock
}

func RegisterStock(i IStock) {
	localStock = i
}
