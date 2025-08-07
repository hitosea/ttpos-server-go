package echo

import (
	"context"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"ttpos-bmp/app/ttpos-takeout/api/echo"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

type Controller struct {
	echo.UnimplementedEchoServiceServer
}

func Register(s *grpcx.GrpcServer) {
	echo.RegisterEchoServiceServer(s.Server, &Controller{})
}

func (*Controller) Echo(ctx context.Context, in *echo.EchoRequest) (res *echo.EchoResponse, err error) {
	rst, err := service.Echo().Msg(ctx, &dto.EchoMsgInput{
		Message:      in.Message,
		OtherMessage: "other message",
	})
	if err != nil {
		return nil, gerror.Wrap(err, "echo failed")
	}
	res = &echo.EchoResponse{
		Message: rst.Message,
	}
	return
}
