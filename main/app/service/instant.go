package service

import (
	"errors"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
)

// IInstantSrv 点餐订单服务接口
type IInstantSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                                         // 创建点餐订单
	GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error)                       // 获取点餐订单详情
	AddProductToInstantOrder(dbId uint64, lang string, req req.InstantOrderAddProductReq) (*resp.GetInstantOrderInfoResp, error) // 添加商品
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
func (s *instantSrv) CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error) {
	return s.orderSrv.CreateInstantOrder(dbId)
}

func (s *instantSrv) GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error) {
	return resp.GetInstantOrderInfoResp{}, nil
}

// AddProductToInstantOrder 添加商品
func (s *instantSrv) AddProductToInstantOrder(dbId uint64, lang string, req req.InstantOrderAddProductReq) (*resp.GetInstantOrderInfoResp, error) {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	db := s.dbm.GetDB(dbId)

	// 验证参数
	if req.SaleBillUuid == 0 || req.SaleOrderUuid == 0 {
		return nil, errors.New("销售账单uuid或销售订单uuid不能为空")
	}
	if req.Product.Uuid == 0 {
		return nil, errors.New("商品uuid不能为空")
	}
	if req.Product.FlavorUuid == 0 {
		return nil, errors.New("商品规格uuid不能为空")
	}

	// 检查销售账单或销售订单
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillInfo(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.New("销售账单不存在")
	}
	if len(saleBill.SaleOrders) == 0 {
		return nil, errors.New("销售订单不存在")
	}

	// 检查订单是否可操作
	if err = saleBill.ValidateOrderStatus(constant.OrderAddProduct); err != nil {
		return nil, err
	}

	// 检查创建订单商品
	productPackage, err := s.orderProductSrv.CheckCreateOrderProduct(dbId, req.Product)
	if err != nil {
		return nil, err
	}

	// 生成订单商品
	_, err = s.orderProductSrv.CreateOrderProduct(dbId, CreateOrderProductReq{
		Lang:           lang,
		SaleBill:       saleBill,
		SaleOrder:      saleBill.SaleOrders[0],
		ProductPackage: *productPackage,
		SauceUuids:     req.Product.SauceUuids,
		Num:            1,
	})

	if err != nil {
		return nil, errors.New("创建订单商品失败")
	}

	return &resp.GetInstantOrderInfoResp{}, nil
}
