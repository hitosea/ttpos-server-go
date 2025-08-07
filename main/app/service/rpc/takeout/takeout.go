package takeout

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewTakeoutClient() (api.TakeoutServiceClient, *grpc.ClientConn, error) {
	// 1. 建立gRPC连接（开发环境使用Insecure，生产环境建议配置TLS）
	conn, err := cloud.GetRpcConnWithName(cloud.TakeOutServiceName)
	if err != nil {
		return nil, nil, err
	}
	return api.NewTakeoutServiceClient(conn), conn, nil
}

// ITakeoutSrv 定义外送服务接口
type ITakeoutSrv interface {
	// 预估距离
	EstimateDistance(ctx context.Context, req *req.TakeoutDistanceReq) (*resp.TakeoutDistanceResp, error)
	// 创建订单
	CreateOrder(ctx context.Context, req *req.CreateTakeoutOrderReq) (*resp.CreateTakeoutOrderResp, error)
	// 确认订单。即“商家备餐完成，联系骑手”
	ConfirmOrder(ctx context.Context, req *req.ConfirmTakeoutOrderReq) error
	// 获取骑手信息
	GetDriverInfo(ctx context.Context, req *req.GetDriverInfoReq) (*resp.GetDriverInfoResp, error)
	// 取消订单
	CancelOrder(ctx context.Context, req *req.CancelTakeoutOrderReq) error
}

// takeoutSrv 外送服务结构体
type takeoutSrv struct {
}

func NewTakeoutSrv() ITakeoutSrv {
	return &takeoutSrv{}
}

// EstimateDistance 预估距离
func (s *takeoutSrv) EstimateDistance(ctx context.Context, req *req.TakeoutDistanceReq) (*resp.TakeoutDistanceResp, error) {
	client, conn, err := NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return nil, err
	}
	defer conn.Close()
	in := &api.EstimateDistanceReq{
		ProviderName: req.ProviderName,
		Address: []*api.Address{
			{
				AddressName: req.Address[0].AddressName,
				Address:     req.Address[0].Address,
				Lat:         req.Address[0].Lat,
				Lng:         req.Address[0].Lng,
			},
			{
				AddressName: req.Address[1].AddressName,
				Address:     req.Address[1].Address,
				Lat:         req.Address[1].Lat,
				Lng:         req.Address[1].Lng,
			},
		},
	}
	res, err := client.EstimateDistance(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return nil, err
	}
	return &resp.TakeoutDistanceResp{
		Distance:     float64(res.Distance),
		TripDuration: int(res.TripDuration),
	}, nil
}

// CreateOrder 创建订单
func (s *takeoutSrv) CreateOrder(ctx context.Context, req *req.CreateTakeoutOrderReq) (*resp.CreateTakeoutOrderResp, error) {
	client, conn, err := NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return nil, err
	}
	defer conn.Close()
	in := &api.CreateOrderReq{
		ProviderName:  req.ProviderName,
		ShopOrderUuid: req.ShopOrderUuid,
		CustomerLocation: &api.Location{
			Lat:          req.CustomerLocation.Lat,
			Lng:          req.CustomerLocation.Lng,
			AddressName:  req.CustomerLocation.AddressName,
			Address:      req.CustomerLocation.Address,
			ContactName:  req.CustomerLocation.ContactName,
			ContactPhone: req.CustomerLocation.ContactPhone,
		},
		MerchantLocation: &api.Location{
			Lat:          req.MerchantLocation.Lat,
			Lng:          req.MerchantLocation.Lng,
			AddressName:  req.MerchantLocation.AddressName,
			Address:      req.MerchantLocation.Address,
			ContactName:  req.MerchantLocation.ContactName,
			ContactPhone: req.MerchantLocation.ContactPhone,
		},
		Remark:      req.Remark,
		CallbackUrl: req.CallbackUrl,
	}
	res, err := client.CreateOrder(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return nil, err
	}
	return &resp.CreateTakeoutOrderResp{
		TakeoutJobUuid: res.TakeoutJobUuid,
		Status:         res.Status,
		ShopOrderUuid:  res.ShopOrderUuid,
		TakeoutRefNo:   res.TakeoutRefNo,
		FinishTime:     res.FinishTime,
	}, nil
}

// ConfirmOrder 确认订单
func (s *takeoutSrv) ConfirmOrder(ctx context.Context, req *req.ConfirmTakeoutOrderReq) error {
	client, conn, err := NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return err
	}
	defer conn.Close()
	in := &api.ConfirmOrderReq{
		ShopOrderUuid: req.ShopOrderUuid,
	}
	_, err = client.ConfirmOrder(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return err
	}
	return nil
}

// CancelOrder 取消订单
func (s *takeoutSrv) CancelOrder(ctx context.Context, req *req.CancelTakeoutOrderReq) error {
	client, conn, err := NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return err
	}
	defer conn.Close()
	in := &api.CancelOrderReq{
		ShopOrderUuid: req.ShopOrderUuid,
	}
	_, err = client.CancelOrder(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return err
	}
	return nil
}

// GetDriverInfo 获取骑手信息
func (s *takeoutSrv) GetDriverInfo(ctx context.Context, req *req.GetDriverInfoReq) (*resp.GetDriverInfoResp, error) {
	client, conn, err := NewTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建外送服务gRPC客户端失败:", zap.Error(err))
		return nil, err
	}
	defer conn.Close()
	in := &api.GetDriverInfoReq{
		ShopOrderUuid: req.ShopOrderUuid,
	}
	res, err := client.GetDriverInfo(ctx, in)
	if err != nil {
		logger.Logger.Error("调用外送服务gRPC客户端失败: %v", zap.Error(err))
		return nil, err
	}
	return &resp.GetDriverInfoResp{
		Name:   res.Name,
		Phone:  res.Phone,
		Avatar: res.Avatar,
		Rating: res.Rating,
		Lat:    res.Lat,
		Lng:    res.Lng,
	}, nil

}
