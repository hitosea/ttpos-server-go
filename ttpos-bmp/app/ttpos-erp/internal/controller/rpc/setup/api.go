package setup

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	setup.UnimplementedSetupServiceServer
}

func Register(s *grpcx.GrpcServer) {
	setup.RegisterSetupServiceServer(s.Server, &Controller{})
}

func (c *Controller) InitShop(ctx context.Context, req *setup.InitShopReq) (res *api.ResponseInfo, err error) {
	if branchName, err := service.Setup().InitShop(ctx, req); err == nil {
		res = &api.ResponseInfo{
			Code:    "0",
			Message: "success",
		}
		res.Data, _ = anypb.New(&setup.InitShopResp{BranchName: branchName})

	} else {
		res = &api.ResponseInfo{
			Code:    "1",
			Message: err.Error(),
		}
	}
	return
}
