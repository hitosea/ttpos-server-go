package boot

import (
	"context"

	v1 "ttpos-bmp/app/ttpos-message/api/message"
	"ttpos-bmp/app/ttpos-message/internal/controller/rpc/message"
	"ttpos-bmp/internal/pkg/nacos/service"
)

// initRpcServer 初始化 gRPC 服务器
func initRpcServer() {
	// 注册消息服务
	v1.RegisterMessageServer(service.RpcServer.GRpc.Server, message.New())
	// 启动 gRPC 服务
	go service.RpcServer.GRpc.Run()
}

// InitRpc 初始化 RPC 服务
// 包括 gRPC 服务器和客户端
func InitRpc(ctx context.Context) {
	// 加载注册 Nacos gRPC 服务
	service.RpcServer.Init(ctx)
	// 加载 Nacos gRPC 客户端
	service.RpcClient.Init(ctx)

	// 初始化 gRPC 服务器
	initRpcServer()
}
