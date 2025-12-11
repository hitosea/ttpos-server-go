package boot

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/takeout"
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
	takeout.Register(service.RpcServer.GRpc)
	grab.Register(service.RpcServer.GRpc)
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
