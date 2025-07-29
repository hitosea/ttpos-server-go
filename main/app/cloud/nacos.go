package cloud

import (
	"go.uber.org/zap"
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
)

func Init() {
	once := sync.Once{}
	once.Do(func() {
		var err error
		if nacosClient, err = nacos.NewNacosClient(config.Nacos); err != nil {
			logger.Logger.Error("初始化nacos客户端服务失败:", zap.Error(err))
		}
	})
}

func GetServiceGrpcAddr(serviceName string) (string, error) {
	return nacosClient.GetServiceGrpcAddr(serviceName)
}
