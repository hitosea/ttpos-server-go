package takeout

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"takeout/api"
	"takeout/internal/consts"
	"takeout/internal/logic/takeout"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct {
	api.UnimplementedTakeoutServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterTakeoutServiceServer(s.Server, &Controller{})
}

func (*Controller) EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error) {
	res, err = takeout.GetService(consts.ProviderName(req.ProviderName)).EstimatePrice(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "获取预估价格失败")
	}
	g.Log().Debugf(ctx, "获取预估价格成功:%+v", res)
	return
}
