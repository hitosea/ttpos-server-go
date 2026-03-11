package cloud

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"ttpos-server-go/app/errors"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	// Check for environment variable override first (for testing)
	// Environment variable format: GRPC_{SERVICE_NAME}_ADDR
	// e.g., GRPC_ERP_ADDR, GRPC_TAKEOUT_ADDR, GRPC_MESSAGE_ADDR, GRPC_WEBSOCKET_ADDR
	envVarName := fmt.Sprintf("GRPC_%s_ADDR", strings.ToUpper(strings.TrimPrefix(serviceName, "ttpos-")))
	if envAddr := os.Getenv(envVarName); envAddr != "" {
		logger.Logger.Debug("Using gRPC address from environment variable",
			zap.String("service", serviceName),
			zap.String("env_var", envVarName),
			zap.String("addr", envAddr))
		return envAddr, nil
	}

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
	// Choose transport credentials: TLS with InsecureSkipVerify when GRPC_TLS_INSECURE=true
	// (used in integration tests where gRPC mock server uses a self-signed cert),
	// otherwise use plaintext (insecure) for internal microservice communication.
	var creds credentials.TransportCredentials
	if os.Getenv("GRPC_TLS_INSECURE") == "true" {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}) //nolint:gosec
	} else {
		creds = insecure.NewCredentials()
	}
	conn, err = grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(creds),
		// 已使用 NewClientHandler 统一处理，增加调用链跟踪
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, errors.New("连接服务gRPC端点失败: %v", addr)
	}
	return
}
