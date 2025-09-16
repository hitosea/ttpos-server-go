package cloud

import (
	"fmt"
	"ttpos-server-go/app/errors"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"sync"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/nacos"
)

var nacosClient *nacos.NacosClient

// ServiceName 服务名称
type ServiceName string

const (
	TakeOutServiceName ServiceName = "ttpos-takeout"
	ErpServiceName     ServiceName = "ttpos-erp"
)

func Init() {
	once := sync.Once{}
	once.Do(func() {
		if config.Server.Mode == "debug" {
			go Startup()
		} else {
			Startup()
		}
	})
}

// Shutdown 优雅关闭 Nacos 客户端
func Shutdown() error {
	if nacosClient == nil {
		return nil
	}

	logger.Logger.Info("开始关闭 Nacos 客户端")

	// Nacos SDK 没有提供显式的关闭方法，但我们可以清空引用
	nacosClient = nil

	logger.Logger.Info("Nacos 客户端已关闭")
	return nil
}

// Startup 初始化Nacos客户端服务
// 返回：错误信息
func Startup() error {
	// 移除不必要的 goroutine，统一处理逻辑
	client, err := nacos.NewNacosClient(config.Nacos)
	if err != nil {
		fmt.Println("初始化nacos客户端服务失败:", zap.Error(err))
		logger.Logger.Error("初始化nacos客户端服务失败:", zap.Error(err))
		return err
	}
	nacosClient = client
	fmt.Println("初始化nacos客户端服务成功")
	logger.Logger.Info("初始化nacos客户端服务成功")
	return nil
}

func GetServiceGrpcAddr(serviceName string) (string, error) {
	if nacosClient == nil {
		return "", errors.WithMessage(errors.New("服务中心未启动，请稍等"))
	}
	return nacosClient.GetServiceGrpcAddr(serviceName)
}

func GetRpcConnWithName(serviceName ServiceName) (conn *grpc.ClientConn, err error) {
	addr, err := GetServiceGrpcAddr(string(serviceName))
	if err != nil {
		logger.Logger.Error("获取外送服务gRPC地址失败:", zap.Error(err))
		return nil, err
	}
	// 1. 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("连接服务gRPC端点失败: %v", addr)
	}
	return
}
