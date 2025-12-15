package grab

import (
	api "ttpos-bmp/app/ttpos-takeout/api/grab"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	api.UnimplementedGrabServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterGrabServer(s.Server, &Controller{})
}
