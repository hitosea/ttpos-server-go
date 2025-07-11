package rpc

import (
	"context"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"takeout/internal/controller/rpc/echo"
)

var Server = new(rpcServer)

type rpcServer struct {
	Grpc *grpcx.GrpcServer
}

func (s *rpcServer) Init(context.Context) {
	s.Grpc = grpcx.Server.New()
}

func (s *rpcServer) Run() {
	s.Grpc.Run()
}

func initRpcServer() {
	// 注册服务
	echo.Register(Server.Grpc)

	go Server.Grpc.Run()
}

func StopRpcServer() {
	Server.Grpc.Stop()
}

func InitRpc(ctx context.Context) {
	//加载grpc服务
	Server.Init(ctx)

	initRpcServer()

}
