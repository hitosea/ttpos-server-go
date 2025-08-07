package takeout

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

type ITakeout interface {
	// EstimatePrice 预估距离
	EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error)
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error)
	// ConfirmOrder 商家确认订单
	ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error)
	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error)
	// GetDriverInfo 获取司机信息
	GetDriverInfo(ctx context.Context, req *dto.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error)
}

func GetService(name consts.ProviderName) ITakeout {
	switch name {
	case consts.ProviderSkootar:
		return service.Skootar()
	default:
		return service.Skootar()
	}
}

type sTakeout struct {
}

var Takeout = new(sTakeout)

func (s *sTakeout) Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error) {
	var takeoutJob *entity.Job
	err := dao.Job.Ctx(ctx).Where(dao.Job.Columns().ShopRefNo, shopOrderUuid).Scan(&takeoutJob)
	if err != nil {
		if e, ok := err.(*gerror.Error); ok {
			return nil, gerror.Wrap(e.Cause(), "获取外送订单失败")
		}
		return nil, gerror.Wrap(err, "获取外送订单失败")
	}
	return takeoutJob, nil
}
