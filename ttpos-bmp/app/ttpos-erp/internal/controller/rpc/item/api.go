package item

import (
	"ttpos-bmp/app/ttpos-erp/api/item"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	item.UnimplementedItemServiceServer
}

func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
}
