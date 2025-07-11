package logic

import (
	"context"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

var RpcServer = new(rpcServer)

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
	go RpcServer.Grpc.Run()
}

func StopRpcServer() {
	RpcServer.Grpc.Stop()
}

func InitRpc(ctx context.Context) {
	//加载grpc服务
	RpcServer.Init(ctx)

	initRpcServer()

}
