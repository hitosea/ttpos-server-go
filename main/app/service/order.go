package service

import (
	"errors"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/cashier_resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IProductSrv 定义收银服务接口
type IOrderSrv interface {
	CreateOrderNo(db *gorm.DB, orderSource string) string                        // 创建订单编号
	CreateInstantOrder(dbId uint64) (cashier_resp.CreateInstantOrderResp, error) // 创建点餐订单
	CreateSaleBill(db *gorm.DB, orderSource string) (uint64, error)              // 创建销售账单
	CreateSaleOrder(db *gorm.DB, saleBillUuid uint64) (uint64, error)            // 创建销售订单
}

// orderSrv 收银服务结构体
type orderSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewOrderSrv 创建新的收银产品类别服务
func NewOrderSrv(dbm *database.DBManager) IOrderSrv {
	return NewOrderSrvImpl(dbm)
}

// NewOrderSrvImpl 创建新的收银服务实现
func NewOrderSrvImpl(dbm *database.DBManager) IOrderSrv {
	return &orderSrv{
		dbm: dbm,
	}
}

// CreateOrderNo 创建订单编号
func (s *orderSrv) CreateOrderNo(db *gorm.DB, orderSource string) string {
	var orderNo string

	// 前八位是年月日
	datePart := time.Now().Format("20060102")
	// 第九位是订单来源
	orderSourceType := constant.OrderSourceMapToOrderNoType[orderSource]

	// 如果订单编号存在, 则重新生成, 重试10次, 否则退出
	for i := 0; i < 10; i++ {
		// 后九位是随机生成
		n := utils.RandomNumber(9)

		// 订单编号
		orderNo = datePart + orderSourceType + n

		// 检查订单编号是否存在
		order, err := repository.NewOrderRepo(db).GetSaleBill(repository.NewCommonRepo().WhereByOrderNo(orderNo))
		if order.Uuid > 0 {
			orderNo = ""
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			orderNo = ""
			break
		} else {
			break
		}
	}

	return orderNo
}

// CreateInstantOrder 创建点餐订单
func (s *orderSrv) CreateInstantOrder(dbId uint64) (cashier_resp.CreateInstantOrderResp, error) {
	var uuid uint64
	db := s.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建销售账单
		saleBillUuid, err := s.CreateSaleBill(tx, constant.OrderSourceInstant)
		if err != nil {
			return err
		}

		// 创建销售订单
		_, err = s.CreateSaleOrder(tx, saleBillUuid)
		if err != nil {
			return err
		}

		uuid = saleBillUuid

		return nil
	})
	if err != nil {
		return cashier_resp.CreateInstantOrderResp{}, err
	}

	return cashier_resp.CreateInstantOrderResp{
		SaleBillUuid: uuid,
	}, nil
}

// CreateSaleBill 创建销售账单
func (s *orderSrv) CreateSaleBill(db *gorm.DB, orderSource string) (uint64, error) {
	// 获取销售账单UUID
	uuid, err := database.GetID()
	if err != nil {
		return 0, err
	}

	// 创建订单编号
	orderNo := s.CreateOrderNo(db, orderSource)
	if orderNo == "" {
		return 0, errors.New("订单编号生成失败")
	}

	// 创建销售账单
	saleBillUuid, err := repository.NewOrderRepo(db).CreateSaleBill(model.SaleBill{
		Uuid:         uuid,
		OrderNo:      orderNo,
		BillType:     constant.OrderSourceMapToBillType[orderSource],
		DiningMethod: constant.SaleBillDiningMethodDineIn,
	})
	if err != nil {
		return 0, err
	}

	return saleBillUuid, nil
}

// CreateSaleOrder 创建销售订单
func (s *orderSrv) CreateSaleOrder(db *gorm.DB, saleBillUuid uint64) (uint64, error) {
	// 获取销售订单UUID
	uuid, err := database.GetID()
	if err != nil {
		return 0, err
	}

	// 创建销售订单
	saleOrderUuid, err := repository.NewOrderRepo(db).CreateSaleOrder(model.SaleOrder{
		Uuid:         uuid,
		SaleBillUuid: saleBillUuid,
	})

	if err != nil {
		return 0, err
	}

	return saleOrderUuid, nil
}
