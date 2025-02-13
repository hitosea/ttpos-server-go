package service

import "ttpos-server-go/pkg/database"

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface{}

// orderProductSrv 订单商品服务结构体
type orderProductSrv struct {
	dbm *database.DBManager
}

// NewOrderProductSrv 创建商品服务
func NewOrderProductSrv(dbm *database.DBManager) IOrderProductSrv {
	return NewOrderProductSrvImpl(dbm)
}

// NewOrderProductSrvImpl 创建商品服务实现
func NewOrderProductSrvImpl(dbm *database.DBManager) IOrderProductSrv {
	return &orderProductSrv{
		dbm: dbm,
	}
}
