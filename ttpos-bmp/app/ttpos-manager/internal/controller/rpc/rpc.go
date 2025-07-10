package rpc

import (
	"context"
	"ttpos-bmp/app/ttpos-manager/api/rpc/manager"
	"ttpos-bmp/app/ttpos-manager/api/rpc/svc"
	"ttpos-bmp/app/ttpos-manager/internal/dao"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	svc.UnimplementedSettingSvcServer
}

func Register(s *grpcx.GrpcServer) {
	svc.RegisterSettingSvcServer(s.Server, &Controller{})
}

func (*Controller) GetSetting(ctx context.Context, req *svc.GetReq) (setting *manager.Setting, err error) {
	err = dao.Setting.Ctx(ctx).Fields(setting).Where(dao.Setting.Columns().Key, req.Key).Scan(&setting)
	return
}
