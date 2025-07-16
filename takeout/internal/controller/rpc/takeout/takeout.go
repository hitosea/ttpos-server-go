package takeout

import (
	"context"
	"takeout/api"
	"takeout/internal/consts"
	"takeout/internal/logic/takeout"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct {
	api.UnimplementedTakeoutServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterTakeoutServiceServer(s.Server, &Controller{})
}

func (c *Controller) getService(providerName string) takeout.Takeout {
	return takeout.GetService(consts.ProviderName(providerName))
}

func (c *Controller) EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error) {
	res, err = c.getService(req.ProviderName).EstimatePrice(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "获取预估价格失败")
	}
	g.Log().Debugf(ctx, "获取预估价格成功:%+v", res)
	return
}

func (c *Controller) CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error) {
	res, err = c.getService(req.ProviderName).CreateOrder(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "下单失败")
	}
	g.Log().Debugf(ctx, "下单成功:%+v", res)
	return
}

func (c *Controller) ConfirmOrder(ctx context.Context, req *api.ConfirmOrderReq) (res *api.ConfirmOrderResp, err error) {
	res, err = c.getService(req.ProviderName).ConfirmOrder(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "商家确认订单失败")
	}
	g.Log().Debugf(ctx, "商家确认订单成功:%+v", res)
	return
}

func (c *Controller) GetDriverLocation(ctx context.Context, req *api.GetDriverLocationReq) (res *api.GetDriverLocationResp, err error) {
	res, err = c.getService(req.ProviderName).GetDriverLocation(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "获取司机位置失败")
	}
	g.Log().Debugf(ctx, "获取司机位置成功:%+v", res)
	return
}
