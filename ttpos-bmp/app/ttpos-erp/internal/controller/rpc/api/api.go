package api

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/company"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	company.UnimplementedCompanyServiceServer
}

func Register(s *grpcx.GrpcServer) {
	company.RegisterCompanyServiceServer(s.Server, &Controller{})
}

func (*Controller) GetCompanyList(context.Context, *GetCompanyListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetItemList(context.Context, *GetItemListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetUomList(context.Context, *GetUomListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) SaveUom(context.Context, *UomInfo) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetAttributeList(context.Context, *GetAttributeListReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) SaveAttribute(context.Context, *AttributeInfo) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetPosProfileList(context.Context, *PosProfileReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) CreatePaymentAccount(context.Context, *CreatePaymentAccountReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) InitShop(context.Context, *InitShopReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) CreatePosUser(context.Context, *CreatePosUserReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
