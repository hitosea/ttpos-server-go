package item

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 物品服务控制器
type Controller struct {
	item.UnimplementedItemServiceServer
}

// Register 注册物品服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
}

// GetItemList 获取物品列表
// 参数：ctx 上下文，req 获取物品列表请求
// 返回：响应信息和错误
func (c *Controller) GetItemList(ctx context.Context, req *item.GetItemListReq) (*api.ResponseInfo, error) {

	// 调用服务层获取数据
	dataList, err := service.Item().GetItemList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品列表成功", dataList), nil
}

// SaveItem 保存物品信息
// 参数：ctx 上下文，req 物品信息
// 返回：响应信息和错误
func (c *Controller) SaveItem(ctx context.Context, req *item.ItemInfo) (*api.ResponseInfo, error) {
	// 参数验证
	if req.ItemName == "" {
		return rpc.ApiError("物品名称不能为空"), nil
	}
	if req.StockUom == "" {
		return rpc.ApiError("库存单位不能为空"), nil
	}
	if len(req.TemplateItemCode) > 0 && req.ItemSpecification == "" {
		return rpc.ApiError("多规格商品时，物品规格不能为空"), nil
	}

	// 调用服务层保存数据
	savedItem, err := service.Item().SaveItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存物品成功", savedItem), nil
}

// GetUomList 获取单位列表
// 参数：ctx 上下文，req 获取单位列表请求
// 返回：响应信息和错误
func (c *Controller) GetUomList(ctx context.Context, req *item.GetUomListReq) (*api.ResponseInfo, error) {
	// 调用服务层获取数据
	dataList, err := service.Stock().GetUomList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取单位列表成功", dataList), nil
}

// SaveUom 保存单位信息
func (*Controller) SaveUom(ctx context.Context, req *item.UomInfo) (*api.ResponseInfo, error) {
	if req.UomName == "" {
		return rpc.ApiError("单位名称不能为空"), nil
	}

	// 调用服务层保存数据
	err := service.Stock().SaveUom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("保存单位成功"), nil
}

// GetAttributeList 获取属性列表
// 参数：ctx 上下文，req 获取属性列表请求
// 返回：响应信息和错误
func (c *Controller) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (*api.ResponseInfo, error) {
	// 调用服务层获取数据
	dataList, err := service.Stock().GetAttributeList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取属性列表成功", dataList), nil
}

// SaveAttribute 保存属性信息
func (*Controller) SaveAttribute(ctx context.Context, req *item.AttributeInfo) (*api.ResponseInfo, error) {
	if req.AttributeName == "" {
		return rpc.ApiError("属性名称不能为空"), nil
	}

	// 调用服务层保存数据
	err := service.Stock().SaveAttribute(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("保存属性成功"), nil
}

// GetItemStock 获取物品库存信息
func (*Controller) GetItemStock(ctx context.Context, req *item.GetItemStockReq) (res *api.ResponseInfo, err error) {
	if req.CompanyAbbr == "" {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if req.Branch == "" {
		return rpc.ApiError("分支机构不能为空"), nil
	}

	// 调用服务层获取数据
	dataList, err := service.Item().GetItemStock(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品库存成功", dataList), nil
}

// GetItem 根据物品编码获取单个物品信息
// 参数：ctx 上下文，req 包含物品编码和可选的公司、分支信息
// 返回：响应信息和错误
func (c *Controller) GetItem(ctx context.Context, req *item.GetItemReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.ItemCode == "" {
		return rpc.ApiError("物品编码不能为空"), nil
	}

	// 调用服务层获取数据
	itemInfo, err := service.Item().GetItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 将 itemInfo 的值赋给 respItem
	respItem := &item.ItemInfo{
		ItemName:           itemInfo.ItemName,
		StockUom:           itemInfo.StockUom,
		ItemCode:           itemInfo.ItemCode,
		ValuationRate:      itemInfo.ValuationRate,
		IsStockItem:        itemInfo.IsStockItem == 1,
		Branch:             itemInfo.CustomBranch,
		Company:            itemInfo.CustomCompany,
		ItemSpecification:  itemInfo.CustomSpecification,
		Disabled:           itemInfo.Disabled == 1,
		Classification:     itemInfo.Classification,
		ClassificationCode: itemInfo.ClassificationCode,
		InternalCode:       itemInfo.CustomInternalCode,
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品信息成功", respItem), nil
}
