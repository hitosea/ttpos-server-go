package takeout

import (
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials/insecure"
	"ttpos-bmp/app/ttpos-takeout/api/echo"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"

	"google.golang.org/grpc"
)

// 根据配置创建外送服务gRPC客户端
func NewEchoClient() (echo.EchoServiceClient, *grpc.ClientConn, error) {
	addr, err := cloud.GetServiceGrpcAddr(string(cloud.TakeOutServiceName))
	if err != nil {
		logger.Logger.Error("获取外送服务gRPC地址失败:", zap.Error(err))
		return nil, nil, err
	}
	// 1. 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, errors.New("连接外送服务gRPC端点失败: %v", addr)
	}

	// 2. 创建echo服务客户端
	return echo.NewEchoServiceClient(conn), conn, nil
}
