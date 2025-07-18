package rpc

import (
	"takeout/api/echo"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
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

func TestEstimateDistance() (*resp.TakeoutDistanceResp, error) {
	res, err := takeout.NewTakeoutSrv().EstimateDistance(context.Background(), &req.TakeoutDistanceReq{
		ProviderName: "skootar",
		Address: []*req.TakeoutAddress{
			{
				AddressName: "new point",
				Address:     "281/28 บรรทัดทอง เขต ราชเทวี กรุงเทพมหานคร ประเทศไทย 10400",
				Lat:         "13.721899",
				Lng:         "100.52900",
			},
			{
				AddressName: "ม.รามคำแหง",
				Address:     "2086 ถนนรามคำแหง เขตบางกะปิ กรุงเทพมหานคร ประเทศไทย 10240",
				Lat:         "13.747408",
				Lng:         "100.540244",
			},
		},
	})
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return nil, err
	}

	logger.Logger.Info("TestEstimateDistance", zap.Any("res", res))

	return res, nil
}

func TestCreateOrder() (*resp.CreateTakeoutOrderResp, error) {
	res, err := takeout.NewTakeoutSrv().CreateOrder(context.Background(), &req.CreateTakeoutOrderReq{
		ProviderName:  "skootar",
		ShopOrderUuid: "3675543933419521",
		CustomerLocation: &req.TakeoutLocation{
			TakeoutAddress: req.TakeoutAddress{
				AddressName: "zomato",
				Address:     "281/28 บรรทัดทอง เขต ราชเทวี กรุงเทพมหานคร ประเทศไทย 10401",
				Lat:         "13.721900",
				Lng:         "100.52900",
			},
			ContactName:  "Thanata",
			ContactPhone: "0864923676",
		},
		MerchantLocation: &req.TakeoutLocation{
			TakeoutAddress: req.TakeoutAddress{
				AddressName: "xiaoxiong",
				Address:     "2086 ถนนรามคำแหง เขตบางกะปิ กรุงเทพมหานคร ประเทศไทย 10241",
				Lat:         "13.747418",
				Lng:         "100.540244",
			},
			ContactName:  "มาช่า",
			ContactPhone: "0998888999",
		},
		Remark:      "test",
		CallbackUrl: "https://ttpos-test1.ttpos.com/api/callback/skootar",
	})
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return nil, err
	}
	logger.Logger.Info("TestCreateOrder", zap.Any("res", res))
	return res, nil
}

func TestConfirmOrder() error {
	err := takeout.NewTakeoutSrv().ConfirmOrder(context.Background(), &req.ConfirmTakeoutOrderReq{
		ShopOrderUuid: "3675227238301699",
	})
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return err
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功")
	return nil
}

func TestGetDriverInfo() (*resp.GetDriverInfoResp, error) {
	res, err := takeout.NewTakeoutSrv().GetDriverInfo(context.Background(), &req.GetDriverInfoReq{
		ShopOrderUuid: "3675227238301699",
	})

	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败", zap.Error(err))
		return nil, err
	}
	logger.Logger.Info("TestGetDriverInfo", zap.Any("res", res))
	return res, nil
}

func TestCancelOrder() error {
	err := takeout.NewTakeoutSrv().CancelOrder(context.Background(), &req.CancelTakeoutOrderReq{
		ShopOrderUuid: "3675534194245633",
		Reason:        "wait too long",
	})
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return err
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功")
	return nil
}
