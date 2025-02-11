package service

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IProductSrv 定义收银服务接口
type IOrderSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                   // 创建点餐订单
	CreateOrderNo(db *gorm.DB, orderSource string) string                                                  // 创建订单编号
	CreateSaleBill(db *gorm.DB, orderSource string) (*model.SaleBill, error)                               // 创建销售账单
	CreateSaleOrder(db *gorm.DB, saleBill model.SaleBill) (*model.SaleOrder, error)                        // 创建销售订单
	GetCashierOrderList(dbId uint64, req req.GetOrderListReq) (resp.CashierOrderListPaginationResp, error) // 获取销售订单列表
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

// CreateInstantOrder 创建点餐订单
func (s *orderSrv) CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error) {
	var uuid uint64
	db := s.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建销售账单
		saleBill, err := s.CreateSaleBill(tx, constant.OrderSourceInstant)
		if err != nil {
			return err
		}

		// 创建销售订单
		_, err = s.CreateSaleOrder(tx, *saleBill)
		if err != nil {
			return err
		}

		uuid = saleBill.Uuid

		return nil
	})
	if err != nil {
		return resp.CreateInstantOrderResp{}, err
	}

	return resp.CreateInstantOrderResp{
		SaleBillUuid: uuid,
	}, nil
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

// CreateSaleBill 创建销售账单
func (s *orderSrv) CreateSaleBill(db *gorm.DB, orderSource string) (*model.SaleBill, error) {
	// 获取销售账单UUID
	uuid, err := database.GetID()
	if err != nil {
		return nil, err
	}

	// 创建订单编号
	orderNo := s.CreateOrderNo(db, orderSource)
	if orderNo == "" {
		return nil, errors.New("订单编号生成失败")
	}

	// 创建销售账单
	saleBill, err := repository.NewOrderRepo(db).CreateSaleBill(model.SaleBill{
		Uuid:         uuid,
		OrderNo:      orderNo,
		BillType:     constant.OrderSourceMapToBillType[orderSource],
		DiningMethod: constant.SaleBillDiningMethodDineIn,
	})
	if err != nil {
		return nil, err
	}

	return &saleBill, nil
}

// CreateSaleOrder 创建销售订单
func (s *orderSrv) CreateSaleOrder(db *gorm.DB, saleBill model.SaleBill) (*model.SaleOrder, error) {
	// 获取销售订单UUID
	uuid, err := database.GetID()
	if err != nil {
		return nil, err
	}

	// 创建销售订单
	saleOrder, err := repository.NewOrderRepo(db).CreateSaleOrder(model.SaleOrder{
		Uuid:         uuid,
		SaleBillUuid: saleBill.Uuid,
		OrderNo:      saleBill.OrderNo,
	})

	if err != nil {
		return nil, err
	}

	return &saleOrder, nil
}

// GetOrderList 获取订单列表
func (s *orderSrv) GetCashierOrderList(dbId uint64, req req.GetOrderListReq) (resp.CashierOrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))
	commonRepo := repository.NewCommonRepo()

	// 获取列表
	lists, total, err := orderRepo.GetOrderListWithPagination(
		req.PageNo,
		req.PageSize,
		commonRepo.Preload(
			repository.WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("delete_time = ?", 0)
					},
				},
			},
			repository.WithPreload{
				Query: "SaleOrders.PaymentOrders",
			},
		),
		commonRepo.WhereByStatus(1),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithID("DESC"),
	)
	if err != nil {
		return resp.CashierOrderListPaginationResp{}, err
	}

	//
	billList := make([]resp.CashierBillList, len(lists))
	for i, bill := range lists {
		totalPayTypeNames := []string{}
		orderList := make([]resp.CashierOrder, len(bill.SaleOrders))
		for i, order := range bill.SaleOrders {
			payTypeNames := []string{}
			for _, payment := range order.PaymentOrders {
				totalPayTypeNames = append(totalPayTypeNames, payment.PaymentTypeName)
				payTypeNames = append(payTypeNames, payment.PaymentTypeName)
			}
			orderList[i] = resp.CashierOrder{
				SaleOrderUuid: order.Uuid,
				BillType:      order.Type,
				SerialNo:      bill.SerialNo + "-" + strconv.Itoa(i+1),
				OrderNo:       order.OrderNo,
				Status:        order.Status,
				FinishTime:    order.FinishTime,
				OrderAmount:   order.OrderAmount,
				PaymentAmount: order.PaymentAmount,
				PayTypeName:   strings.Join(payTypeNames, ","),
			}
		}
		//
		billList[i] = resp.CashierBillList{
			SaleBillUuid:  bill.Uuid,
			BillType:      bill.BillType,
			IsSplit:       len(bill.SaleOrders) > 1,
			SerialNo:      bill.SerialNo,
			OrderNo:       bill.OrderNo,
			Status:        bill.Status,
			FinishTime:    bill.FinishTime,
			OrderAmount:   bill.OrderAmount,
			PaymentAmount: bill.PaymentAmount,
			PayTypeName:   strings.Join(totalPayTypeNames, ","),
			SaleOrders:    orderList,
		}
	}

	// 获取数量
	getOrderNum := func(status uint) (int64, error) {
		return orderRepo.GetOrderNum(
			commonRepo.WhereByStatus(status),
			commonRepo.WhereBySoftDelete(),
		)
	}
	unpaidNum, _ := getOrderNum(0)
	completeNum, _ := getOrderNum(1)
	cancelNum, _ := getOrderNum(2)

	// 返回响应对象
	return resp.CashierOrderListPaginationResp{
		List: billList,
		Meta: resp.CashierOrderListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			UnpaidNum:   unpaidNum,
			CancelNum:   cancelNum,
			CompleteNum: completeNum,
		},
	}, nil
}
