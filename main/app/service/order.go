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
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                       // 创建点餐订单
	CreateDeskOrder(dbId uint64, req req.CreateDeskOrderReq) (resp.CreateDeskOrderResp, error) // 创建桌台订单
	CreateOrderNo(db *gorm.DB, orderSource string) string                                      // 创建订单编号
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
		// 判断是否有待支付、未挂单的订单
		order, err := repository.NewOrderRepo(tx).GetSaleBill(
			repository.NewCommonRepo().WhereByStatus(constant.SaleBillStatusPending),
			repository.NewCommonRepo().WhereByIsHide(false),
		)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if order.Uuid > 0 {
			return errors.New("有待支付、未挂单的订单")
		}

		// 获取销售账单UUID
		saleBillUuid, err := database.GetID()
		if err != nil {
			return err
		}

		// 创建订单编号
		orderNo := s.CreateOrderNo(tx, constant.OrderSourceInstant)
		if orderNo == "" {
			return errors.New("订单编号生成失败")
		}

		// 创建销售账单
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			Uuid:         saleBillUuid,
			OrderNo:      orderNo,
			BillType:     constant.OrderSourceMapToBillType[constant.OrderSourceInstant],
			DiningMethod: constant.SaleBillDiningMethodDineIn,
		})
		if err != nil {
			return err
		}

		// 获取销售订单UUID
		saleOrderUuid, err := database.GetID()
		if err != nil {
			return err
		}

		// 创建销售订单
		_, err = repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
			Uuid:         saleOrderUuid,
			SaleBillUuid: saleBill.Uuid,
			OrderNo:      saleBill.OrderNo,
		})
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

// CreateDeskOrder 创建桌台订单
func (s *orderSrv) CreateDeskOrder(dbId uint64, req req.CreateDeskOrderReq) (resp.CreateDeskOrderResp, error) {
	var uuid uint64
	var db = s.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 获取销售账单UUID
		saleBillUuid, err := database.GetID()
		if err != nil {
			return err
		}

		// 创建订单编号
		orderNo := s.CreateOrderNo(tx, constant.OrderSourceDesk)
		if orderNo == "" {
			return errors.New("订单编号生成失败")
		}

		// 创建销售账单
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			Uuid:         saleBillUuid,
			OrderNo:      orderNo,
			BillType:     constant.OrderSourceMapToBillType[constant.OrderSourceDesk],
			DiningMethod: constant.SaleBillDiningMethodDineIn,
			IsBuffet:     utils.BoolToUint(req.IsBuffet),
			MealNum:      req.MealNum,
			Remark:       req.Remark,
		})
		if err != nil {
			return err
		}

		// 获取销售订单UUID
		saleOrderUuid, err := database.GetID()
		if err != nil {
			return err
		}

		// 创建销售订单
		_, err = repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
			Uuid:         saleOrderUuid,
			SaleBillUuid: saleBill.Uuid,
			OrderNo:      saleBill.OrderNo,
		})
		if err != nil {
			return err
		}

		uuid = saleBill.Uuid

		return nil
	})

	if err != nil {
		return resp.CreateDeskOrderResp{}, err
	}

	return resp.CreateDeskOrderResp{
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
