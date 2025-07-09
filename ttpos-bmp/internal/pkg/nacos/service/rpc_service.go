package service

import (
	"context"
	"fmt"
	"github.com/gogf/gf/contrib/registry/nacos/v2"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gsvc"
)

var RpcServer = new(rpcServer)

type rpcServer struct {
	GRpc *grpcx.GrpcServer
}

func (s *rpcServer) Init(ctx context.Context) {
	grpcx.Resolver.Register(nacos.New(GetNacosAddress(ctx)))
	//使用默认组和集群名称
	//.SetClusterName("DEFAULT")
	//.SetGroupName("DEFAULT_GROUP"))
	s.GRpc = grpcx.Server.New()
	//注册服务
	//go s.GRpc.Run()
}

func (s *rpcServer) InitHttp(ctx context.Context) {
	gsvc.SetRegistry(nacos.New(GetNacosAddress(ctx)))
	//使用默认组和集群名称
	//.SetClusterName("DEFAULT")
	//.SetGroupName("DEFAULT_GROUP"))
}

func GetNacosAddress(ctx context.Context) string {
	nacosServerIp := g.Cfg().MustGetWithEnv(ctx, "nacos.server.ip")
	nacosServerPort := g.Cfg().MustGetWithEnv(ctx, "nacos.server.port")
	address := fmt.Sprintf("%v:%v", nacosServerIp, nacosServerPort)
	g.Log().Debugf(ctx, "注册中心配置: %v:%v", nacosServerIp, nacosServerPort)
	return address
}
