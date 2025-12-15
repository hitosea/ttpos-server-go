package takeout

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/logic/takeout"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"

	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
)

type Controller struct {
	api.UnimplementedTakeoutServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterTakeoutServiceServer(s.Server, &Controller{})
}

func (c *Controller) getService(providerName string) takeout.ITakeout {
	return takeout.GetService(consts.ProviderName(providerName))
}

func (c *Controller) EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error) {
	res, err = c.getService(req.ProviderName).EstimateDistance(ctx, req)
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
	takeoutJob, err := takeout.Takeout.Get(ctx, req.ShopOrderUuid)
	if err != nil {
		return nil, gerror.Wrap(err, "确认订单失败")
	}
	if takeoutJob == nil {
		return nil, gerror.Wrap(gerror.New("订单不存在"), "确认订单失败")
	}
	res, err = c.getService(takeoutJob.ProviderName).ConfirmOrder(ctx, &dto.ConfirmOrderInp{JobId: takeoutJob.TakeoutRefNo})
	if err != nil {
		return nil, gerror.Wrap(err, "确认订单失败")
	}
	g.Log().Debugf(ctx, "商家确认订单成功:%+v", res)
	return
}

func (c *Controller) GetDriverInfo(ctx context.Context, req *api.GetDriverInfoReq) (res *api.GetDriverInfoResp, err error) {
	takeoutJob, err := takeout.Takeout.Get(ctx, req.ShopOrderUuid)
	if err != nil {
		return nil, gerror.Wrap(err, "获取司机位置失败")
	}
	if takeoutJob == nil {
		return nil, gerror.Wrap(gerror.New("订单不存在"), "获取司机位置失败")
	}
	if takeoutJob.SkootarId == "" {
		return nil, gerror.Wrap(gerror.New("未有骑手接单"), "获取司机位置失败")
	}
	res, err = c.getService(takeoutJob.ProviderName).GetDriverInfo(ctx, &dto.GetDriverInfoInp{
		SKootarId: takeoutJob.SkootarId,
		Name:      takeoutJob.SkootarName,
		Phone:     takeoutJob.SkootarPhone,
		Avatar:    takeoutJob.SkootarImageUrl,
		Rating:    takeoutJob.SkootarRating,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "获取司机位置失败")
	}
	g.Log().Debugf(ctx, "获取司机位置成功:%+v", res)
	return
}

func (c *Controller) CancelOrder(ctx context.Context, req *api.CancelOrderReq) (res *api.CancelOrderResp, err error) {
	takeoutJob, err := takeout.Takeout.Get(ctx, req.ShopOrderUuid)
	if err != nil {
		return nil, gerror.Wrap(err, "取消订单失败")
	}
	if takeoutJob == nil {
		return nil, gerror.Wrap(gerror.New("订单不存在"), "取消订单失败")
	}
	res, err = c.getService(takeoutJob.ProviderName).CancelOrder(ctx, &dto.CancelOrderInp{JobId: takeoutJob.TakeoutRefNo, Reason: req.Reason})
	if err != nil {
		return nil, gerror.Wrap(err, "取消订单失败")
	}
	g.Log().Debugf(ctx, "取消订单成功:%+v", res)
	return
}
