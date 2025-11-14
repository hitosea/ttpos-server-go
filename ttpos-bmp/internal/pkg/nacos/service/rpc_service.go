package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/contrib/registry/nacos/v2"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gsvc"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/text/gstr"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	grpcConf.Options = append(grpcConf.Options,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(unaryRecordInterceptor()),
	)
	grpcx.Resolver.Register(nacos.New(GetNacosAddress(ctx)))
	//使用默认组和集群名称
	//.SetClusterName("DEFAULT")
	//.SetGroupName("DEFAULT_GROUP"))
	s.GRpc = grpcx.Server.New(grpcConf)
	//注册服务
	//go s.GRpc.Run()
}

func unaryRecordInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 获取当前 span，这个 span 应该是由 otelgrpc stats handler 创建的
		span := trace.SpanFromContext(ctx)

		// 只在 span 有效时记录事件
		if span.SpanContext().IsValid() {
			// 序列化请求体
			var requestStr string
			if pm, ok := req.(proto.Message); ok {
				if b, err := protojson.Marshal(pm); err == nil {
					requestStr = string(b)
				} else {
					requestStr = fmt.Sprintf("Failed to marshal proto message: %v", err)
				}
			} else if b, err := json.Marshal(req); err == nil {
				requestStr = string(b)
			} else {
				requestStr = fmt.Sprintf("Failed to marshal request: %v", err)
			}

			// 限制请求体大小
			if len(requestStr) > 2048 {
				requestStr = requestStr[:2048] + "...[truncated]"
			}

			// 记录请求体事件
			span.AddEvent("grpc.request.body", trace.WithAttributes(
				attribute.String("grpc.request", requestStr),
			))
		}

		// 执行实际的请求处理
		resp, err := handler(ctx, req)

		// 记录响应事件
		if span.SpanContext().IsValid() {
			if err != nil {
				span.AddEvent("grpc.response.error", trace.WithAttributes(
					attribute.String("grpc.error", err.Error()),
				))
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.AddEvent("grpc.response.success")
				span.SetStatus(codes.Ok, "success")
			}
		}

		return resp, err
	}
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
