package shop

import (
	api "ttpos-bmp/app/ttpos-takeout/api/shop"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	api.UnimplementedShopServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterShopServer(s.Server, &Controller{})
}
