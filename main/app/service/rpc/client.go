package rpc

import (
	"time"
	"ttpos-bmp/app/ttpos-takeout/api/echo"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func init() {
}

func TestEcho(c *gin.Context) (res *echo.EchoResponse, err error) {
	gcc := helper.GetContext(c)
	// 关键修复：增加客户端和连接的判空检查
	client, conn, err := takeout.NewEchoClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &echo.EchoRequest{
		Message: "ttpos no1",
	}
	ctx, cancel := context.WithTimeout(gcc.GetContext(), 5*time.Second)
	defer cancel()

	res, err = client.Echo(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}
