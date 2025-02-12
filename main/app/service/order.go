package service

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IProductSrv 定义收银服务接口
type IOrderSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                   // 创建点餐订单
	CreateDeskOrder(dbId uint64, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)             // 创建桌台订单
	CreateOrderNo(db *gorm.DB, orderSource string) string                                                  // 创建订单编号
	GetCashierOrderList(dbId uint64, req req.GetOrderListReq) (resp.CashierOrderListPaginationResp, error) // 获取收银订单列表
	GetCashierOrderInfo(dbId uint64, req req.GetOrderInfoReq) (resp.CashierOrderInfoResp, error)           // 获取收银订单详情
}

// orderSrv 收银服务结构体
type orderSrv struct {
	dbm       *database.DBManager // 数据库管理器
	localeSrv ILocaleSrv
	cache     cache.Cache
}

// NewOrderSrv 创建新的收银产品类别服务
func NewOrderSrv(dbm *database.DBManager, localeSrv ILocaleSrv, cache cache.Cache) IOrderSrv {
	return NewOrderSrvImpl(dbm, localeSrv, cache)
}

// NewOrderSrvImpl 创建新的收银服务实现
func NewOrderSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, cache cache.Cache) IOrderSrv {
	return &orderSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
		cache:     cache,
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
func (s *orderSrv) CreateDeskOrder(dbId uint64, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
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
			IsBuffet:     utils.BoolToUint(*req.IsBuffet),
			MealNum:      *req.MealNum,
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
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
		),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithID("DESC"),
		// 额外条件
		func() repository.DBOption {
			return func(db *gorm.DB) *gorm.DB {
				// 订单编号
				if req.OrderNo != "" {
					db = db.Where("order_no = ?", req.OrderNo)
				}
				// 账单类型
				if req.BillType != -1 {
					db = db.Where("bill_type = ?", req.BillType)
				}
				//  账单状态
				if req.Status != -1 {
					db = db.Where("status = ?", uint(req.Status))
				}
				//  日期类型 -1-全都 1-今天 2-昨天 3-本周
				if req.DateType >= 0 && req.DateType <= 3 {
					now := time.Now()
					var startTime, endTime time.Time
					switch req.DateType {
					case 1: // 今天
						startTime = now.Truncate(24 * time.Hour)
						endTime = startTime.Add(24*time.Hour - time.Second)
					case 2: // 昨天
						startTime = now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
						endTime = startTime.Add(24*time.Hour - time.Second)
					case 3: // 本周
						weekday := int(now.Weekday())
						if weekday == 0 {
							weekday = 7
						}
						startTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
						endTime = startTime.AddDate(0, 0, 7).Add(-time.Second)
					}
					db = db.Where("create_time BETWEEN ? AND ?", startTime.Unix(), endTime.Unix())
				}
				// 日期范围
				if req.QueryStartTime != 0 || req.QueryEndTime != 0 {
					timeFields := []string{}
					if req.EnableCreateTime || !req.EnablePayTime {
						timeFields = append(timeFields, "create_time")
					}
					if req.EnablePayTime {
						timeFields = append(timeFields, "finish_time")
					}
					// 开始时间
					endTime := uint(0)
					if req.QueryEndTime != 0 {
						endTime = req.QueryEndTime + 86399
					}
					//
					query := ""
					args := []interface{}{}
					for i, field := range timeFields {
						if i > 0 {
							query += " OR "
						}
						if req.QueryStartTime > 0 && endTime > 0 {
							query += fmt.Sprintf("(%s BETWEEN ? AND ?)", field)
							args = append(args, req.QueryStartTime, endTime)
						} else if req.QueryStartTime > 0 {
							query += fmt.Sprintf("(%s > ?)", field)
							args = append(args, req.QueryStartTime)
						} else if endTime > 0 {
							query += fmt.Sprintf("(%s < ? AND %s > 0)", field, field)
							args = append(args, endTime)
						}
					}
					if query != "" {
						db = db.Where(query, args...)
					}
				}
				//
				return db
			}
		}(),
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
				totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
				payTypeNames = append(payTypeNames, payment.PaymentMethodName)
			}
			orderList[i] = resp.CashierOrder{
				SaleOrderUuid: order.Uuid,
				BillType:      bill.BillType,
				SerialNo:      bill.SerialNo + "-" + strconv.Itoa(i+1),
				OrderNo:       order.OrderNo,
				Status:        order.Status,
				FinishTime:    order.FinishTime,
				OrderAmount:   order.Amount,
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
			OrderAmount:   bill.Amount,
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
		Meta: struct {
			dto.PageResponse
			UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
			CompleteNum int64 `json:"complete_num"` // 已完成数量
			CancelNum   int64 `json:"cancel_num"`   // 已取消数量
		}{
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

// GetOrderList 获取订单信息
func (s *orderSrv) GetCashierOrderInfo(dbId uint64, req req.GetOrderInfoReq) (resp.CashierOrderInfoResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))
	commonRepo := repository.NewCommonRepo()
	// 获取列表
	info, err := orderRepo.GetSaleBill(
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
			repository.WithPreload{
				Query: "SaleOrders.Member",
			},
			repository.WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
			},
		),
		// 额外条件
		func() repository.DBOption {
			return func(db *gorm.DB) *gorm.DB {
				db = db.Where("uuid = ?", req.SaleBillUuid)
				//
				return db
			}
		}(),
	)
	if err != nil {
		return resp.CashierOrderInfoResp{}, err
	}
	//
	totalMemberNames := []string{}
	payTypes := []resp.CashierOrderInfoPayTypes{}
	orderList := make([]resp.CashierOrderInfo, len(info.SaleOrders))
	for i, order := range info.SaleOrders {
		payTypeNames := []string{}
		for _, payment := range order.PaymentOrders {
			payTypes = append(payTypes, resp.CashierOrderInfoPayTypes{
				Uuid:              payment.Uuid,
				PaymentMethodName: payment.PaymentMethodName,
				CurrencyUnit:      payment.CurrencyUnit,
				PaymentAmount:     payment.PaymentAmount,
				Status:            payment.Status,
				Source:            payment.PaymentMethod.Source,
			})
			payTypeNames = append(payTypeNames, payment.PaymentMethodName)
		}
		if order.Member.Nickname != "" && !slices.Contains(totalMemberNames, order.Member.Nickname) {
			totalMemberNames = append(totalMemberNames, order.Member.Nickname)
		}
		// todo 未完成
		products := make([]resp.CashierOrderProduct, len(order.SaleOrderProducts))
		for j, product := range order.SaleOrderProducts {
			products[j] = resp.CashierOrderProduct{
				Uuid:                  product.Uuid,
				LocaleName:            s.localeSrv.GetLocaleNames(product.MultiLanguageName),
				FlavorName:            product.FlavorName,
				Num:                   product.Num,
				CustomPrice:           product.CustomPrice,
				UnitPrice:             product.UnitPrice,
				Price:                 product.Price,
				TaxRate:               product.TaxRate,
				ProductOriginalAmount: product.ProductOriginalAmount,
				Status:                product.Status,
				Remark:                product.Remark,
				IsGift:                product.IsGift == 1,
				GiftReason:            product.GiftReason,
				AttributeName:         "",
			}
		}
		//
		orderList[i] = resp.CashierOrderInfo{
			SaleOrderUuid: order.Uuid,
			BillType:      info.BillType,
			SerialNo:      info.SerialNo + "-" + strconv.Itoa(i+1),
			OrderNo:       order.OrderNo,
			Status:        order.Status,
			FinishTime:    order.FinishTime,
			OrderAmount:   order.Amount,
			PaymentAmount: order.PaymentAmount,
			PayTypeName:   strings.Join(payTypeNames, ","),
			MemberName:    order.Member.Nickname,
			Products:      products,
		}
	}
	//
	return resp.CashierOrderInfoResp{
		SaleBillUuid:  info.Uuid,
		BillType:      info.BillType,
		IsSplit:       len(info.SaleOrders) > 1,
		SerialNo:      info.SerialNo,
		OrderNo:       info.OrderNo,
		Status:        info.Status,
		CreateTime:    info.CreateTime,
		FinishTime:    info.FinishTime,
		OrderAmount:   info.Amount,
		PaymentAmount: info.PaymentAmount,
		MemberNames:   strings.Join(totalMemberNames, ","),
		PayTypes:      payTypes,
		SaleOrders:    orderList,
	}, nil
}
