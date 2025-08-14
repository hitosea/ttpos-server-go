package item

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"

	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

type Controller struct {
	item.UnimplementedItemServiceServer
}

func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
}

// GetItemList 获取物品列表
func (*Controller) GetItemList(ctx context.Context, req *item.GetItemListReq) (res *api.ResponseInfo, err error) {
	dataList, err := service.Item().GetItemList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res = rpc.ApiSuccess("获取物品列表成功")
	res.Data, _ = anypb.New(dataList)
	return res, nil
}

// SaveItem 保存物品信息
func (*Controller) SaveItem(ctx context.Context, req *item.ItemInfo) (res *api.ResponseInfo, err error) {
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

	item, err := service.Item().SaveItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res = rpc.ApiSuccess("保存物品成功")
	res.Data, _ = anypb.New(item)
	return res, nil
}

// GetUomList 获取单位列表
func (*Controller) GetUomList(ctx context.Context, req *item.GetUomListReq) (res *api.ResponseInfo, err error) {
	dataList, err := service.Stock().GetUomList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res = rpc.ApiSuccess("获取单位列表成功")
	res.Data, _ = anypb.New(dataList)
	return res, nil
}

// SaveUom 保存单位信息
func (*Controller) SaveUom(ctx context.Context, req *item.UomInfo) (res *api.ResponseInfo, err error) {
	if req.UomName == "" {
		return rpc.ApiError("单位名称不能为空"), nil
	}

	err = service.Stock().SaveUom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	return rpc.ApiSuccess("保存单位成功"), nil
}

// GetAttributeList 获取属性列表
func (*Controller) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *api.ResponseInfo, err error) {
	dataList, err := service.Stock().GetAttributeList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res = rpc.ApiSuccess("获取属性列表成功")
	res.Data, _ = anypb.New(dataList)
	return res, nil
}

// SaveAttribute 保存属性信息
func (*Controller) SaveAttribute(ctx context.Context, req *item.AttributeInfo) (res *api.ResponseInfo, err error) {
	if req.AttributeName == "" {
		return rpc.ApiError("属性名称不能为空"), nil
	}

	err = service.Stock().SaveAttribute(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

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

	dataList, err := service.Item().GetItemStock(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res = rpc.ApiSuccess("获取物品库存成功")
	res.Data, _ = anypb.New(dataList)
	return res, nil
}
