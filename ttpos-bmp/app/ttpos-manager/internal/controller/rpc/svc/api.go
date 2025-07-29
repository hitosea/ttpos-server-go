package svc

import (
	"context"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-manager/api/manager"
	"ttpos-bmp/app/ttpos-manager/api/svc"
	"ttpos-bmp/app/ttpos-manager/internal/service"
)

type Controller struct {
	svc.UnimplementedSettingSvcServer
}

func Register(s *grpcx.GrpcServer) {
	svc.RegisterSettingSvcServer(s.Server, &Controller{})
}

func (*Controller) GetSetting(ctx context.Context, req *svc.GetReq) (setting *manager.Setting, err error) {
	setting, err = service.Setting().GetSetting(ctx, req)
	if err != nil {
		g.Log().Warning(ctx, "get setting failed", err)
	}
	return
}
