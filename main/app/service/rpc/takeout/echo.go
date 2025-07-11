package takeout

import (
	"google.golang.org/grpc/credentials/insecure"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"

	"google.golang.org/grpc"
	"takeout/api/echo"
)

// NewTakeoutGRPCClient 根据配置创建外送服务gRPC客户端
func NewTakeoutGRPCClient() (echo.EchoServiceClient, *grpc.ClientConn, error) {

	// 1. 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	conn, err := grpc.NewClient(config.TakeOutRpcConf.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, errors.New("连接外送服务gRPC端点失败: %v", config.TakeOutRpcConf.Endpoint)
	}

	// 2. 创建echo服务客户端
	return echo.NewEchoServiceClient(conn), conn, nil
}
