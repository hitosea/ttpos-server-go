// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
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
		// GetItem 根据物品编码获取单个物品信息
		// 参数：ctx 上下文，req 包含物品编码和可选的公司、分支信息
		// 返回：物品详细信息，错误信息
		GetItem(ctx context.Context, req *item.GetItemReq) (res *erp.Item, err error)
		// SavePosAttribute 保存 POS 系统中的属性物品
		// 参数：ctx 上下文，req 保存属性物品请求
		// 返回：保存后的物品信息，错误信息
		SavePosAttribute(ctx context.Context, req *item.SavePosAttributeReq) (res *item.ItemInfo, err error)
		// SavePosAddon 保存 POS 系统中的加料物品
		// 参数：ctx 上下文，req 保存加料物品请求
		// 返回：保存后的物品信息，错误信息
		SavePosAddon(ctx context.Context, req *item.SavePosAddonReq) (res *item.ItemInfo, err error)
		// CreateSingleVariantItem 创建多规格商品的单个规格商品
		// 参数：ctx 上下文，req 物品信息，code 物品编码
		// 返回：创建结果
		CreateSingleVariantItem(ctx context.Context, req *erp.CreateSingleVariantItemReq, templateItemInfo *item.ItemInfo) (string, error)
		// DeleteItem 删除物品
		// 删除模板商品时，变体商品都会被移除
		DeleteItem(ctx context.Context, req *item.DeleteItemReq) (*item.DeleteItemResp, error)
	}
	IItemGroup interface {
		// GetItemGroupList 获取物品分组列表
		// 根据查询条件过滤并返回物品分组信息列表
		GetItemGroupList(ctx context.Context, req *item.GetItemGroupListReq) (res *item.GetItemGroupListResp, err error)
		// GetItemGroup 根据分组代码获取单个物品分组信息
		// 参数：ctx 上下文，req 包含分组代码的请求
		// 返回：物品分组详细信息，错误信息
		GetItemGroup(ctx context.Context, req *item.GetItemGroupReq) (res *erp.ItemGroupInfo, err error)
		// SaveItemGroup 保存物品分组信息
		// 如果物品分组已存在则更新，否则创建新物品分组
		SaveItemGroup(ctx context.Context, req *item.SaveItemGroupReq) (res *erp.ItemGroupInfo, err error)
		// DeleteItemGroup 删除物品分组
		// 参数：ctx 上下文，req 删除物品分组请求
		// 返回：错误信息
		DeleteItemGroup(ctx context.Context, req *item.DeleteItemGroupReq) error
		// SaveAttributeGroup 保存物品属性分组
		// 参数：ctx 上下文，req 保存物品属性分组请求
		// 返回：物品属性分组响应，错误信息
		SaveAttributeGroup(ctx context.Context, req *item.SaveAttributeGroupReq) (*item.SaveAttributeGroupResp, error)
		DeleteAttributeGroup(ctx context.Context, req *item.DeleteAttributeGroupReq) (*item.DeleteAttributeGroupReq, error)
		// SaveAddonGroup 保存加料组
		// 关联门店时，每个门店都会自动创建一个加料组,
		SaveAddonGroup(ctx context.Context, req *item.SaveAddonGroupReq) (*item.SaveAddonGroupResp, error)
	}
	IProduct interface {
		UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*item.UpdateProductResp, error)
	}
	IStock interface {
		// GetAttributeList 获取属性列表
		// 根据查询条件过滤并返回属性信息列表
		GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *item.GetAttributeListResp, err error)
		// GetAttributeValuesList 获取属性值列表
		// 根据属性名称查询对应的属性值列表
		GetAttributeValuesList(ctx context.Context, attributeName string) ([]*item.AttributeValueInfo, error)
		// SaveAttribute 保存属性信息
		// 如果属性已存在则更新，否则创建新属性
		SaveAttribute(ctx context.Context, req *item.AttributeInfo) error
		CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *stock.SaveMaterialRequestResp, err error)
		// GetMaterialRequestList 获取物料请求列表
		GetMaterialRequestList(ctx context.Context, req *stock.GetMaterialRequestListReq) (res *stock.GetMaterialRequestListResp, err error)
		// GetItemAttribute 根据属性名称获取单个属性详细信息
		// 参数：ctx 上下文，attributeName 属性名称
		// 返回：属性详细信息，错误信息
		GetItemAttribute(ctx context.Context, attributeName string) (res *erp.ItemAttribute, err error)
		// GetStockLedger 获取库存分类账信息
		// 根据查询条件过滤并返回库存分类账记录列表
		GetStockLedger(ctx context.Context, req *stock.GetStockLedgerReq) (res *stock.GetStockLedgerResp, err error)
	}
	IUom interface {
		DeleteUom(ctx context.Context, req *item.DeleteUomReq) (*item.DeleteUomResp, error)
		// SaveUom 保存单位信息
		// 如果单位已存在则更新，否则创建新单位
		SaveUom(ctx context.Context, req *item.UomInfo) error
		// GetUomList 获取单位列表
		// 根据查询条件过滤并返回单位信息列表
		GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error)
		// GetUom 根据单位名称获取单个单位详细信息
		// 参数：ctx 上下文，req 包含单位名称
		// 返回：单位详细信息，错误信息
		GetUom(ctx context.Context, req *item.GetUomReq) (res *erp.UOM, err error)
	}
	IWarehouse interface {
		// CreateWarehouse 创建仓库
		// 参数：ctx 上下文，req 包含 shop_name、company_abbr
		// 返回：仓库名称，错误信息
		CreateWarehouse(ctx context.Context, req *setup.CreateWarehouseInp) (warehouseName string, err error)
		// GetWarehouseList 获取仓库列表
		// 根据查询条件过滤并返回仓库信息列表
		GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (res *warehouse.GetWarehouseListResp, err error)
		GetDefaultWarehouse(ctx context.Context, company string, branch string) (res *warehouse.WarehouseInfo, err error)
		// GetWarehouse 获取单个仓库详情
		// 根据仓库名称获取仓库详细信息
		GetWarehouse(ctx context.Context, req *warehouse.GetWarehouseReq) (res *warehouse.GetWarehouseResp, err error)
		// UpdateWarehouse 更新仓库信息
		// 根据仓库名称更新仓库的相关信息
		UpdateWarehouse(ctx context.Context, req *warehouse.UpdateWarehouseReq) (res *warehouse.UpdateWarehouseResp, err error)
		// DeleteWarehouse 删除仓库
		// 根据仓库名称删除指定仓库
		DeleteWarehouse(ctx context.Context, req *warehouse.DeleteWarehouseReq) (res *warehouse.DeleteWarehouseResp, err error)
	}
)

var (
	localItem      IItem
	localItemGroup IItemGroup
	localProduct   IProduct
	localStock     IStock
	localUom       IUom
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

func ItemGroup() IItemGroup {
	if localItemGroup == nil {
		panic("implement not found for interface IItemGroup, forgot register?")
	}
	return localItemGroup
}

func RegisterItemGroup(i IItemGroup) {
	localItemGroup = i
}

func Product() IProduct {
	if localProduct == nil {
		panic("implement not found for interface IProduct, forgot register?")
	}
	return localProduct
}

func RegisterProduct(i IProduct) {
	localProduct = i
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

func Uom() IUom {
	if localUom == nil {
		panic("implement not found for interface IUom, forgot register?")
	}
	return localUom
}

func RegisterUom(i IUom) {
	localUom = i
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
