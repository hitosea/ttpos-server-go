package takeout

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"takeout/api"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"
)

func NewTakeoutClient() (api.TakeoutServiceClient, *grpc.ClientConn, error) {
	// 1. 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	conn, err := grpc.NewClient(config.TakeOutRpcConf.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, errors.New("连接外送服务gRPC端点失败: %v", config.TakeOutRpcConf.Endpoint)
	}

	return api.NewTakeoutServiceClient(conn), conn, nil
}
