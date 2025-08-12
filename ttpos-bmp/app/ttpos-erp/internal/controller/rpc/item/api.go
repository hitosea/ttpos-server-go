package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	item.UnimplementedItemServiceServer
}

func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
}

func (*Controller) GetItemList(context.Context, *item.GetItemListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
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
	err = service.Stock().SaveUom(ctx, req)
	if err != nil {
		res = rpc.ApiError(err.Error())
		return
	}
	res = rpc.ApiSuccess("保存单位成功")
	return
}

func (*Controller) GetAttributeList(context.Context, *item.GetAttributeListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) SaveAttribute(context.Context, *item.AttributeInfo) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
