package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/database"
)

// IInstantSrv 点餐订单服务接口
type IInstantSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                   // 创建点餐订单
	GetInstantOrderInfo(dbId uint64, req req.GetInstantOrderInfoReq) (resp.GetInstantOrderInfoResp, error) // 获取点餐订单详情
}

// instantSrv 点餐订单服务结构体
type instantSrv struct {
	dbm      *database.DBManager // 数据库管理器
	orderSrv IOrderSrv           // 订单服务
}

// NewInstantSrv 创建点餐订单服务
func NewInstantSrv(dbm *database.DBManager, orderSrv IOrderSrv) IInstantSrv {
	return NewInstantSrvImpl(dbm, orderSrv)
}

// NewInstantSrvImpl 创建点餐订单服务实现
func NewInstantSrvImpl(dbm *database.DBManager, orderSrv IOrderSrv) IInstantSrv {
	return &instantSrv{
		dbm:      dbm,
		orderSrv: orderSrv,
	}
}

// CreateInstantOrder 创建点餐订单
func (s *instantSrv) CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error) {
	return s.orderSrv.CreateInstantOrder(dbId)
}

func (s *instantSrv) GetInstantOrderInfo(dbId uint64, req req.GetInstantOrderInfoReq) (resp.GetInstantOrderInfoResp, error) {
	return resp.GetInstantOrderInfoResp{}, nil
}
