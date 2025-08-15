package manufacturing

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Controller BOM服务控制器
type Controller struct {
	manufacturing.UnimplementedBomServiceServer
}

// Register 注册BOM服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	manufacturing.RegisterBomServiceServer(s.Server, &Controller{})
}

func (s Controller) GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (*api.ResponseInfo, error) {

	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
