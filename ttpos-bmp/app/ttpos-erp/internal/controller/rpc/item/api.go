package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
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

func (*Controller) GetUomList(context.Context, *item.GetUomListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) AddUom(context.Context, *item.AddUomReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetAttributeList(context.Context, *item.GetAttributeListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) AddAttribute(context.Context, *item.AddAttributeReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
