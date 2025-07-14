package rpc

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"takeout/api"
	"takeout/api/echo"
	"time"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/pkg/logger"
)

func init() {
}

func TestEcho() (res *echo.EchoResponse, err error) {
	// 关键修复：增加客户端和连接的判空检查
	client, conn, err := takeout.NewEchoClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &echo.EchoRequest{
		Message: "ttpos no1",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.Echo(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}

func TestEstimatePrice() (res *api.EstimatePriceResp, err error) {
	client, conn, err := takeout.NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &api.EstimatePriceReq{
		ProviderName: "skootar",
		Address: []*api.Address{
			{
				Lat: "13.721899",
				Lng: "100.52900",
			},
			{
				Lat: "13.747408",
				Lng: "100.540244",
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.EstimatePrice(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}
