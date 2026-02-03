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

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

var nacosClient *nacos.NacosClient

// ServiceName 服务名称
type ServiceName string

const (
	TakeOutServiceName   ServiceName = "ttpos-takeout"
	ErpServiceName       ServiceName = "ttpos-erp"
	MessageServiceName   ServiceName = "ttpos-message"
	WebSocketServiceName ServiceName = "ttpos-websocket"
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

// getDirectGrpcAddr 获取BMP服务的直连地址（调试模式使用）
func getDirectGrpcAddr(serviceName ServiceName) string {
	switch serviceName {
	case ErpServiceName:
		return config.Bmp.ErpGrpcAddr
	case TakeOutServiceName:
		return config.Bmp.TakeoutGrpcAddr
	case MessageServiceName:
		return config.Bmp.MessageGrpcAddr
	case WebSocketServiceName:
		return config.Bmp.WebsocketGrpcAddr
	default:
		return ""
	}
}

func GetRpcConnWithName(serviceName ServiceName) (conn *grpc.ClientConn, err error) {
	var addr string

	// 调试模式下，优先使用直连地址
	if config.Server.Mode == "debug" {
		if directAddr := getDirectGrpcAddr(serviceName); directAddr != "" {
			addr = directAddr
			logger.Logger.Debug("使用BMP服务直连地址", zap.String("service", string(serviceName)), zap.String("addr", addr))
		}
	}

	// 如果没有直连地址，则通过Nacos服务发现获取
	if addr == "" {
		addr, err = GetServiceGrpcAddr(string(serviceName))
		if err != nil {
			// 调试模式下，如果Nacos服务发现失败，给出更友好的提示
			if config.Server.Mode == "debug" {
				logger.Logger.Error("获取BMP服务gRPC地址失败，请检查：1.BMP服务是否已启动 2.是否配置了直连地址(BMP_ERP_GRPC_ADDR等)", zap.String("service", string(serviceName)), zap.Error(err))
			} else {
				logger.Logger.Error("获取BMP服务gRPC地址失败:", zap.String("service", string(serviceName)), zap.Error(err))
			}
			return nil, err
		}
	}

	// 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	// 同时接入 OpenTelemetry 客户端拦截器以自动创建与传播 trace
	conn, err = grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 已使用 NewClientHandler 统一处理，增加调用链跟踪
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, errors.New("连接服务gRPC端点失败: %v", addr)
	}
	return
}
