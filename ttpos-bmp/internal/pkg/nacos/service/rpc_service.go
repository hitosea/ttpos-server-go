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

// GetNacosAddress 获取 Nacos 地址（支持多实例，返回逗号分隔的地址字符串）
// GoFrame nacos.New() 支持逗号分隔的多个地址，格式：host1:port1,host2:port2,host3:port3
// Nacos SDK 会自动处理多实例连接、故障转移和负载均衡
func GetNacosAddress(ctx context.Context) string {
	// 优先使用 addresses 配置（多实例）
	if addresses := g.Cfg().MustGetWithEnv(ctx, "nacos.server.addresses", "").String(); len(addresses) > 0 {
		// 验证地址格式
		addressList := gstr.Split(addresses, ",")
		validAddresses := make([]string, 0, len(addressList))
		for _, addr := range addressList {
			addr = gstr.Trim(addr)
			if isValidAddress(addr) {
				validAddresses = append(validAddresses, addr)
			} else {
				g.Log().Warningf(ctx, "无效的 Nacos 地址格式，将忽略: %s", addr)
			}
		}
		if len(validAddresses) > 0 {
			result := gstr.Join(validAddresses, ",")
			g.Log().Infof(ctx, "使用多实例 Nacos 配置: %s", result)
			return result
		}
		g.Log().Warningf(ctx, "多实例配置中没有有效地址，将使用单实例配置")
	}

	// 兼容单实例配置
	nacosServerIp := g.Cfg().MustGetWithEnv(ctx, "nacos.server.ip")
	nacosServerPort := g.Cfg().MustGetWithEnv(ctx, "nacos.server.port")
	address := fmt.Sprintf("%v:%v", nacosServerIp, nacosServerPort)
	g.Log().Debugf(ctx, "使用单实例 Nacos 配置: %s", address)
	return address
}

// isValidAddress 验证 Nacos 地址格式是否正确
// 地址格式应为 host:port
func isValidAddress(addr string) bool {
	if len(addr) == 0 {
		return false
	}
	parts := gstr.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}
	// 验证 host 不为空
	if len(gstr.Trim(parts[0])) == 0 {
		return false
	}
	// 验证 port 不为空
	port := gstr.Trim(parts[1])
	if len(port) == 0 {
		return false
	}
	return true
}
