package boot

import (
	"context"
	"net"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"

	v1 "ttpos-bmp/app/ttpos-websocket/api/websocket"
	"ttpos-bmp/app/ttpos-websocket/internal/controller/rpc/websocket"
)

var grpcServer *grpc.Server

// InitRpc 初始化 RPC 服务
// 包括 gRPC 服务器和客户端
func InitRpc(ctx context.Context) {
	// 创建 gRPC 服务器
	grpcServer = grpc.NewServer()

	// 注册 WebSocket 服务
	v1.RegisterWebSocketServer(grpcServer, websocket.New())

	// 获取 gRPC 监听地址
	grpcAddr := g.Cfg().MustGet(ctx, "grpc.address", ":9090").String()

	g.Log().Info(ctx, "gRPC 服务器启动", "address", grpcAddr)

	// 启动 gRPC 服务器
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			g.Log().Fatalf(ctx, "gRPC 服务器监听失败: %v", err)
		}

		g.Log().Info(ctx, "gRPC 服务器已启动", "address", grpcAddr)

		if err := grpcServer.Serve(lis); err != nil {
			g.Log().Fatalf(ctx, "gRPC 服务器启动失败: %v", err)
		}
	}()
}

// ShutdownRpc 关闭 gRPC 服务
func ShutdownRpc(ctx context.Context) {
	if grpcServer != nil {
		g.Log().Info(ctx, "正在关闭 gRPC 服务器...")
		grpcServer.GracefulStop()
		g.Log().Info(ctx, "gRPC 服务器已关闭")
	}
}
