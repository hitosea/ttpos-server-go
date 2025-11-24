package nacos

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
)

// NacosClient Nacos客户端结构体
type NacosClient struct {
	configClient config_client.IConfigClient // 配置客户端
	namingClient naming_client.INamingClient // 服务发现客户端
	nacosConfig  config.NacosConf            // nacos配置
}

// NewNacosClient 创建nacos客户端
// 参数：nacosConfig nacos配置
// 返回：NacosClient实例，错误信息
func NewNacosClient(nacosConfig config.NacosConf) (*NacosClient, error) {
	// 创建nacos客户端配置
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(nacosConfig.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("./log/nacos"),
		constant.WithCacheDir("./tmp/nacos/cache"),
		constant.WithLogLevel("info"),
		constant.WithUsername(nacosConfig.Username),
		constant.WithPassword(nacosConfig.Password),
	)

	// 解析多个地址，创建多个 ServerConfig
	var serverConfigs []constant.ServerConfig
	if len(nacosConfig.Addresses) > 0 {
		// 多实例配置
		addresses := strings.Split(nacosConfig.Addresses, ",")
		validCount := 0
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			parts := strings.Split(addr, ":")
			if len(parts) == 2 {
				host := strings.TrimSpace(parts[0])
				portStr := strings.TrimSpace(parts[1])
				if len(host) > 0 && len(portStr) > 0 {
					port, err := strconv.ParseUint(portStr, 10, 64)
					if err == nil {
						serverConfigs = append(serverConfigs,
							*constant.NewServerConfig(host, port))
						validCount++
					} else {
						logger.Logger.Warn("无效的 Nacos 端口格式，将忽略",
							zap.String("address", addr), zap.Error(err))
					}
				}
			} else {
				logger.Logger.Warn("无效的 Nacos 地址格式，将忽略",
					zap.String("address", addr))
			}
		}
		if validCount > 0 {
			logger.Logger.Info("使用多实例 Nacos 配置",
				zap.Int("count", validCount),
				zap.String("addresses", nacosConfig.Addresses))
		} else {
			logger.Logger.Warn("多实例配置中没有有效地址，将使用单实例配置")
			// 回退到单实例配置
			serverConfigs = []constant.ServerConfig{
				*constant.NewServerConfig(nacosConfig.Host, uint64(nacosConfig.Port)),
			}
		}
	} else {
		// 单实例配置（向后兼容）
		serverConfigs = []constant.ServerConfig{
			*constant.NewServerConfig(nacosConfig.Host, uint64(nacosConfig.Port)),
		}
		logger.Logger.Debug("使用单实例 Nacos 配置",
			zap.String("host", nacosConfig.Host),
			zap.Int("port", nacosConfig.Port))
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}
	// 创建服务发现客户端
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		logger.Logger.Error("创建nacos服务发现客户端失败:", zap.Error(err))
		return nil, err
	}
	return &NacosClient{
		configClient: configClient,
		namingClient: namingClient,
		nacosConfig:  nacosConfig,
	}, nil
}

// GetServiceGrpcAddr 通过nacos服务发现获取rpc服务地址
// 参数：serviceName 服务名
// 返回：服务地址，错误信息
func (c *NacosClient) GetServiceGrpcAddr(serviceName string) (string, error) {
	// 只取一个健康实例
	instance, err := c.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   c.nacosConfig.Group,
	})
	if err != nil {
		logger.Logger.Error("获取nacos服务发现客户端失败:", zap.Error(err))
		return "", err
	}
	return fmt.Sprintf("%s:%d", instance.Ip, instance.Port), nil
}
