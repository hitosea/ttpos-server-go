package boot

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/company"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/setup"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc/warehouse"
	"ttpos-bmp/internal/pkg/nacos/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
)

var (
	managerConn *grpc.ClientConn
)

func InitRpcClient(ctx context.Context) {
	managerConn = grpcx.Client.MustNewGrpcClientConn(g.Cfg().MustGet(ctx, "grpc.name").String())
}

func initRpcServer() {
	company.Register(service.RpcServer.GRpc)
	item.Register(service.RpcServer.GRpc)
	setup.Register(service.RpcServer.GRpc)
	selling.Register(service.RpcServer.GRpc)
	warehouse.Register(service.RpcServer.GRpc)
	stock.Register(service.RpcServer.GRpc)
	go service.RpcServer.GRpc.Run()
}

func InitRpc(ctx context.Context) {
	//加载注册nacos grpc服务
	service.RpcServer.Init(ctx)
	//加载nacos grpc客户端
	service.RpcClient.Init(ctx)

	initRpcServer()

	InitRpcClient(ctx)

}
