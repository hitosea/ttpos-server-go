package takeout

import (
	"context"
	"takeout/api"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	api.UnimplementedTakeoutServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterTakeoutServiceServer(s.Server, &Controller{})
}

func (*Controller) EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) ConfirmOrder(ctx context.Context, req *api.ConfirmOrderReq) (res *api.ConfirmOrderResp, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetDriverLocation(ctx context.Context, req *api.GetDriverLocationReq) (res *api.GetDriverLocationResp, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
