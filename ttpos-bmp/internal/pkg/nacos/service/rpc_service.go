package service

import (
	"context"
	"fmt"
	"github.com/gogf/gf/contrib/registry/nacos/v2"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gsvc"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/text/gstr"
)

var RpcServer = new(rpcServer)

type rpcServer struct {
	GRpc *grpcx.GrpcServer
}

func (s *rpcServer) Init(ctx context.Context) {
	grpcConf := grpcx.Server.NewConfig()
	//特殊处理 grpc.endpoints ,支持从环境变量中获取
	if endpoints := genv.Get("GRPC_ENDPOINTS", "").String(); len(endpoints) > 0 {
		g.Log().Infof(ctx, "使用环境变量注册服务 GRPC_ENDPOINTS: %v", endpoints)
		grpcConf.Endpoints = gstr.Split(endpoints, ",")
	}
	grpcx.Resolver.Register(nacos.New(GetNacosAddress(ctx)))
	//使用默认组和集群名称
	//.SetClusterName("DEFAULT")
	//.SetGroupName("DEFAULT_GROUP"))
	s.GRpc = grpcx.Server.New(grpcConf)
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
