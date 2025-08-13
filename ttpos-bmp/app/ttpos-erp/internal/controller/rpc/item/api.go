package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	item.UnimplementedItemServiceServer
}

func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
}

func (*Controller) GetItemList(ctx context.Context, req *item.GetItemListReq) (res *api.ResponseInfo, err error) {
	if dataList, err := service.Item().GetItemList(ctx, req); err == nil {
		res = rpc.ApiSuccess("获取商品列表成功")
		res.Data, _ = anypb.New(dataList)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}

func (*Controller) SaveItem(ctx context.Context, req *item.ItemInfo) (res *api.ResponseInfo, err error) {
	if req.ItemName == "" {
		res = rpc.ApiError("商品名称不能为空")
		return
	}
	if req.StockUom == "" {
		res = rpc.ApiError("库存单位不能为空")
		return
	}
	if len(req.TemplateItemCode) > 0 && req.ItemSpecification == "" {
		res = rpc.ApiError("多规格商品时，物品规格不能为空")
		return
	}
	item, err := service.Item().SaveItem(ctx, req)
	if err == nil {
		res = rpc.ApiSuccess("保存物品成功")
		res.Data, _ = anypb.New(item)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}

func (*Controller) GetUomList(ctx context.Context, req *item.GetUomListReq) (res *api.ResponseInfo, err error) {
	if dataList, err := service.Stock().GetUomList(ctx, req); err == nil {
		res = rpc.ApiSuccess("获取单位列表成功")
		res.Data, _ = anypb.New(dataList)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}

func (*Controller) SaveUom(ctx context.Context, req *item.UomInfo) (res *api.ResponseInfo, err error) {
	if req.UomName == "" {
		res = rpc.ApiError("单位名称不能为空")
		return
	}
	err = service.Stock().SaveUom(ctx, req)
	if err != nil {
		res = rpc.ApiError(err.Error())
		return
	}
	res = rpc.ApiSuccess("保存单位成功")
	return
}

func (*Controller) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *api.ResponseInfo, err error) {
	if dataList, err := service.Stock().GetAttributeList(ctx, req); err == nil {
		res = rpc.ApiSuccess("获取属性列表成功")
		res.Data, _ = anypb.New(dataList)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}

func (*Controller) SaveAttribute(ctx context.Context, req *item.AttributeInfo) (res *api.ResponseInfo, err error) {
	if req.AttributeName == "" {
		res = rpc.ApiError("属性名称不能为空")
		return
	}
	err = service.Stock().SaveAttribute(ctx, req)
	if err != nil {
		res = rpc.ApiError(err.Error())
		return
	}
	res = rpc.ApiSuccess("保存单位成功")
	return
}
