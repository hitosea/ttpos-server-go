package api

import (
	"context"
	"takeout/api/echo"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	echo.UnimplementedEchoServiceServer
}

func Register(s *grpcx.GrpcServer) {
	echo.RegisterEchoServiceServer(s.Server, &Controller{})
}

func (*Controller) Echo(ctx context.Context, req *echo.EchoRequest) (res *echo.EchoResponse, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
