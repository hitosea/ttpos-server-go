package rpc

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-takeout/api/echo"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service/rpc/erp"
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

func TestEstimateDistance(ctx context.Context) (*resp.TakeoutDistanceResp, error) {
	res, err := takeout.NewTakeoutSrv().EstimateDistance(ctx, &req.TakeoutDistanceReq{
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

			// 13.769116, 100.605733
			TakeoutAddress: req.TakeoutAddress{
				AddressName: "Sarllive Hair Spa",
				Address:     "พลับพลา 456/10 Biz town in town Srivara Rd, Bangkok 10310",
				Lat:         "13.769116",
				Lng:         "100.605733",
			},
			ContactName:  "aaa",
			ContactPhone: "0864923676",
		},
		MerchantLocation: &req.TakeoutLocation{

			// 13.772280704650836, 100.60901012388038

			TakeoutAddress: req.TakeoutAddress{
				AddressName: "Ob Aroi Town In Town",
				Address:     "1329 53 Lat Phrao 94 Alley, Phlabphla, วังทองหลา Bangkok 10310",
				Lat:         "13.772280",
				Lng:         "100.609010",
			},
			ContactName:  "bbb",
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
		ShopOrderUuid: "3675543933419521",
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

func TestCompanyList(ctx context.Context) error {
	client, conn, err := erp.NewErpSellingClient()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	result, err := client.GetPosProfileList(erp.WithSiteCode(ctx, "2"), &selling.PosProfileReq{
		CompanyAbbr: "wallace-pjd",
	})
	if err != nil {
		panic(err)
	}
	var posProfileListResp selling.PosProfileListResp
	if err := result.Data.UnmarshalTo(&posProfileListResp); err != nil {
		logger.Logger.Error("GetPosProfileList-UnmarshalTo", zap.Any("err", err))
		return err
	}
	ccccc, _ := json.Marshal(posProfileListResp.ProfileList)
	fmt.Println(string(ccccc))

	return nil
}
