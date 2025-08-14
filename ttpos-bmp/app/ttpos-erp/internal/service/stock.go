// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	IItem interface {
		// SyncDelay 同步延迟处理
		// 将物品同步任务推送到队列，并设置10秒延迟
		SyncDelay()
		// GetItemList 获取物品列表
		// 根据查询条件过滤并返回物品信息列表
		GetItemList(ctx context.Context, req *item.GetItemListReq) (res *item.GetItemListResp, err error)
		// SaveItem 保存物品信息
		// 如果物品已存在则更新，否则创建新物品
		SaveItem(ctx context.Context, reqInfo *item.ItemInfo) (res *item.ItemInfo, err error)
		// GetItemStock 获取物品库存信息
		// 根据公司简称、分支机构和物品编码查询库存信息
		GetItemStock(ctx context.Context, req *item.GetItemStockReq) (res *item.GetItemStockResp, err error)
	}
	IStock interface {
		// GetUomList 获取单位列表
		// 根据查询条件过滤并返回单位信息列表
		GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error)
		// SaveUom 保存单位信息
		// 如果单位已存在则更新，否则创建新单位
		SaveUom(ctx context.Context, req *item.UomInfo) error
		// GetAttributeList 获取属性列表
		// 根据查询条件过滤并返回属性信息列表
		GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *item.GetAttributeListResp, err error)
		// GetAttributeValuesList 获取属性值列表
		// 根据属性名称查询对应的属性值列表
		GetAttributeValuesList(ctx context.Context, attributeName string) ([]*item.AttributeValueInfo, error)
		// SaveAttribute 保存属性信息
		// 如果属性已存在则更新，否则创建新属性
		SaveAttribute(ctx context.Context, req *item.AttributeInfo) error
	}
	IWarehouse interface {
		// CreateWarehouse 创建仓库
		// 参数：ctx 上下文，req 包含 shop_name、company_abbr
		// 返回：仓库名称，错误信息
		CreateWarehouse(ctx context.Context, req *erp.CreateWarehouseInp) (warehouseName string, err error)
		// GetWarehouseList 获取仓库列表
		// 根据查询条件过滤并返回仓库信息列表
		GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (res *warehouse.GetWarehouseListResp, err error)
		GetDefaultWarehouse(ctx context.Context, company string, branch string) (res *warehouse.WarehouseInfo, err error)
	}
)

var (
	localItem      IItem
	localStock     IStock
	localWarehouse IWarehouse
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

func Warehouse() IWarehouse {
	if localWarehouse == nil {
		panic("implement not found for interface IWarehouse, forgot register?")
	}
	return localWarehouse
}

func RegisterWarehouse(i IWarehouse) {
	localWarehouse = i
}
