package rpc

import (
	"encoding/json"
	"fmt"
	"takeout/api"
	"takeout/api/echo"
	"time"
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

func TestEstimatePrice() (res *api.EstimatePriceResp, err error) {
	client, conn, err := takeout.NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &api.EstimatePriceReq{
		ProviderName: "skootar",
		Address: []*api.Address{
			{
				Lat: "13.721899",
				Lng: "100.52900",
			},
			{
				Lat: "13.747408",
				Lng: "100.540244",
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.EstimatePrice(ctx, in)

	resData, _ := json.Marshal(res)
	fmt.Println("++++")
	fmt.Println(string(resData))
	fmt.Println("++++")

	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}

func TestCreateOrder() (res *api.CreateOrderResp, err error) {
	client, conn, err := takeout.NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &api.CreateOrderReq{
		ProviderName:  "skootar",
		ShopOrderUuid: "8888777",
		CustomerLocation: &api.Location{
			Lat:          "13.721899",
			Lng:          "100.52900",
			AddressName:  "new point",
			Address:      "281/28 บรรทัดทอง เขต ราชเทวี กรุงเทพมหานคร ประเทศไทย 10400",
			ContactName:  "Thanate",
			ContactPhone: "0864923595",
		},
		MerchantLocation: &api.Location{
			Lat:          "13.747408",
			Lng:          "100.540244",
			AddressName:  "ม.รามคำแหง",
			Address:      "2086 ถนนรามคำแหง เขตบางกะปิ กรุงเทพมหานคร ประเทศไทย 10240",
			ContactName:  "มาช่า",
			ContactPhone: "0998888999",
		},
		Remark:      "test",
		CallbackUrl: "https://ttpos-test1.ttpos.com/api/callback/skootar",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.CreateOrder(ctx, in)

	resData, _ := json.Marshal(res)
	fmt.Println("++++")
	fmt.Println(string(resData))
	fmt.Println("++++")

	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}

func TestConfirmOrder() (res *api.ConfirmOrderResp, err error) {
	client, conn, err := takeout.NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &api.ConfirmOrderReq{
		ProviderName: "skootar",
		OrderId:      "J25070856677",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.ConfirmOrder(ctx, in)

	resData, _ := json.Marshal(res)
	fmt.Println("++++")
	fmt.Println(string(resData))
	fmt.Println("++++")

	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}

func TestGetDriverLocation() (res *api.GetDriverLocationResp, err error) {
	client, conn, err := takeout.NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return
	}
	defer conn.Close()
	in := &api.GetDriverLocationReq{
		ProviderName: "skootar",
		OrderId:      "SK0002",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = client.GetDriverLocation(ctx, in)

	resData, _ := json.Marshal(res)
	fmt.Println("++++")
	fmt.Println(string(resData))
	fmt.Println("++++")

	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return
	}
	logger.Logger.Info("外送服务gRPC客户端测试成功", zap.Any("res", res))
	return
}
