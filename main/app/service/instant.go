package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IInstantSrv 点餐订单服务接口
type IInstantSrv interface {
	CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error)                           // 创建点餐订单
	GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error) // 获取点餐订单详情
}

// instantSrv 点餐订单服务结构体
type instantSrv struct {
	dbm             *database.DBManager // 数据库管理器
	orderSrv        IOrderSrv           // 订单服务
	orderProductSrv IOrderProductSrv    // 订单商品服务
}

// NewInstantSrv 创建点餐订单服务
func NewInstantSrv(dbm *database.DBManager, orderSrv IOrderSrv, orderProductSrv IOrderProductSrv) IInstantSrv {
	return NewInstantSrvImpl(dbm, orderSrv, orderProductSrv)
}

// NewInstantSrvImpl 创建点餐订单服务实现
func NewInstantSrvImpl(dbm *database.DBManager, orderSrv IOrderSrv, orderProductSrv IOrderProductSrv) IInstantSrv {
	return &instantSrv{
		dbm:             dbm,
		orderSrv:        orderSrv,
		orderProductSrv: orderProductSrv,
	}
}

// CreateInstantOrder 创建点餐订单
func (s *instantSrv) CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error) {
	return s.orderSrv.CreateInstantOrder(ctx)
}

func (s *instantSrv) GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error) {
	return resp.GetInstantOrderInfoResp{}, nil
}
