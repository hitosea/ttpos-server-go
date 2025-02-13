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
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IProductSrv 定义收银服务接口
type IOrderSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                        // 创建点餐订单
	CreateDeskOrder(dbId uint64, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                  // 创建桌台订单
	GetCashierOrderList(dbId uint64, req req.OrderListReq) (resp.CashierOrderListPaginationResp, error)         // 获取收银订单列表
	GetCashierOrderInfo(dbId uint64, req req.OrderInfoReq) (resp.CashierOrderInfoResp, error)                   // 获取收银订单详情
	CancelOrder(dbId uint64, staff model.Staff, source string, saleBillUuid uint64, saleOrderUuid uint64) error // 取消订单
}

// orderSrv 收银服务结构
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
	var billUuid uint64
	var orderUuid uint64
	db := s.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 判断是否有待支付、未挂单的订单
		commonRepo := repository.NewCommonRepo()
		orderRepo := repository.NewOrderRepo(tx)
		order, err := orderRepo.GetSaleBill(
			commonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceInstant]),
			commonRepo.WhereByStatus(constant.SaleBillStatusPending),
			commonRepo.WhereByIsHide(false),
		)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if order.Uuid > 0 {
			return errors.New("有待支付、未挂单的订单")
		}

		// 获取销售账单UUID
		saleBillUuid, err := utils.GetID()
		if err != nil {
			return err
		}

		// 创建订单编号
		orderNo := s.createOrderNo(tx, constant.OrderSourceInstant)
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
		saleOrderUuid, err := utils.GetID()
		if err != nil {
			return err
		}

		// 创建销售订单
		saleOrder, err := repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
			Uuid:         saleOrderUuid,
			SaleBillUuid: saleBill.Uuid,
			OrderNo:      saleBill.OrderNo,
		})
		if err != nil {
			return err
		}

		billUuid = saleBill.Uuid
		orderUuid = saleOrder.Uuid

		return nil
	})
	if err != nil {
		return resp.CreateInstantOrderResp{}, err
	}

	return resp.CreateInstantOrderResp{
		SaleBillUuid:  billUuid,
		SaleOrderUuid: orderUuid,
	}, nil
}

// CreateDeskOrder 创建桌台订单
func (s *orderSrv) CreateDeskOrder(dbId uint64, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	var billUuid uint64
	var orderUuid uint64
	var db = s.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 获取销售账单UUID
		saleBillUuid, err := utils.GetID()
		if err != nil {
			return err
		}

		// 创建订单编号
		orderNo := s.createOrderNo(tx, constant.OrderSourceDesk)
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
		saleOrderUuid, err := utils.GetID()
		if err != nil {
			return err
		}

		// 创建销售订单
		saleOrder, err := repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
			Uuid:         saleOrderUuid,
			SaleBillUuid: saleBill.Uuid,
			OrderNo:      saleBill.OrderNo,
		})
		if err != nil {
			return err
		}

		if *req.IsBuffet {
			commonRepo := repository.NewCommonRepo()
			buffetRepo := repository.NewBuffetRepo(tx)
			// 创建销售订单自助餐顾客类型
			for _, buffetUuid := range req.BuffetUuids {
				for _, buffetCustomerType := range req.BuffetCustomerTypes {
					// 获取自助餐顾客类型价格
					_, err := buffetRepo.GetBuffetCustomerTypePrice(
						commonRepo.WhereByBuffetPackageUuid(buffetUuid),
						commonRepo.WhereByCustomerTypeUuid(buffetCustomerType.Uuid),
					)
					if err != nil {
						continue
					}

					saleOrderBuffetCustomerTypeUuid, err := utils.GetID()
					if err != nil {
						return err
					}

					_, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetCustomerType(model.SaleOrderBuffetCustomerType{
						Uuid:                   saleOrderBuffetCustomerTypeUuid,
						SaleOrderUuid:          saleOrder.Uuid,
						BuffetPackageUuid:      buffetUuid,
						BuffetCustomerTypeUuid: buffetCustomerType.Uuid,
						Num:                    *buffetCustomerType.MealNum,
					})
					if err != nil {
						return err
					}
				}
			}
		}

		billUuid = saleBill.Uuid
		orderUuid = saleOrder.Uuid

		return nil
	})

	if err != nil {
		return resp.CreateDeskOrderResp{}, err
	}

	return resp.CreateDeskOrderResp{
		SaleBillUuid:  billUuid,
		SaleOrderUuid: orderUuid,
	}, nil
}

// createOrderNo 创建订单编号
func (s *orderSrv) createOrderNo(db *gorm.DB, orderSource string) string {
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
func (s *orderSrv) GetCashierOrderList(dbId uint64, req req.OrderListReq) (resp.CashierOrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	copier.Copy(&reqs, req)
	lists, total, err := orderRepo.GetCashierOrderListWithPagination(reqs)
	if err != nil {
		return resp.CashierOrderListPaginationResp{}, err
	}
	// 组合列表源数据
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
	getOrderNum := func(status uint) int64 {
		num, _ := orderRepo.GetOrderNum(
			repository.CommonRepo.WhereByStatus(status),
			repository.CommonRepo.WhereBySoftDelete(),
		)
		return num
	}
	// 返回响应对象
	return resp.CashierOrderListPaginationResp{
		List: billList,
		Meta: struct {
			dto.PageResponse
			UnpaidNum   int64 `json:"unpaid_num"`
			CompleteNum int64 `json:"complete_num"`
			CancelNum   int64 `json:"cancel_num"`
		}{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			UnpaidNum:   getOrderNum(0),
			CancelNum:   getOrderNum(1),
			CompleteNum: getOrderNum(2),
		},
	}, nil
}

// GetOrderList 获取订单信息
func (s *orderSrv) GetCashierOrderInfo(dbId uint64, req req.OrderInfoReq) (resp.CashierOrderInfoResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取信息源
	info, err := orderRepo.GetSaleBillDetails(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return resp.CashierOrderInfoResp{}, err
	}

	// 组合信息
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
		//
		products := make([]resp.CashierOrderProduct, len(order.SaleOrderProducts))
		for j, product := range order.SaleOrderProducts {
			attributeNames := []string{}
			for _, bom := range product.SaleOrderProductBoms {
				if bom.IsFlavorBom == 1 {
					attributeNames = append(attributeNames, bom.Name)
				}
			}
			for _, attribute := range product.SaleOrderProductAttributes {
				attributeNames = append(attributeNames, attribute.Name)
			}
			for _, bom := range product.SaleOrderProductBoms {
				if bom.IsFlavorBom != 1 {
					attributeNames = append(attributeNames, bom.Name)
				}
			}
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
				ImageUrl:              product.ImageFile.GetUrl(),
				Attributes:            strings.Join(attributeNames, ";"),
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

	// 返回响应对象
	return resp.CashierOrderInfoResp{
		Detail: resp.CashierOrderInfos{
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
		},
		OperationLog: struct {
			List []resp.CashierOrderOperationLog
		}{
			List: []resp.CashierOrderOperationLog{},
		},
	}, nil
}

// CancelOrder 取消订单
func (s *orderSrv) CancelOrder(dbId uint64, staff model.Staff, source string, saleBillUuid uint64, saleOrderUuid uint64) error {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(saleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	productRepo := repository.NewOrderProductRepo(db)
	deskRepo := repository.NewDeskRepo(db)
	qrcodeOrderRepo := repository.NewQrcodeOrderRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, 0)
	if err != nil {
		return err
	}

	// 验证状态
	if err = billInfo.ValidateOrderStatus("cancel"); err != nil {
		return err
	}

	// 检查是否部分支付
	isPartially := orderRepo.IsPartiallyPaid(billInfo)
	if isPartially {
		return errors.New("当前订单已被部分支付，不支持取消")
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 获取订单已送厨产品，退回商品库存
	products, err := productRepo.GetProductList(
		repository.CommonRepo.WhereByStatus(1),
		productRepo.WhereSaleBillUuids([]uint64{saleBillUuid}),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// todo 未完成 - 退回商品库存
	// for _, po := range products {
	// 	// ProductFactory::getFactory($detail['order_source'])->backProductStock([$orderProduct], $isPay);
	// }

	// if ($force) {
	// 	$detail->force()->delete();
	// 	// 子数据删除
	// 	$orderBuffetList = OrderBuffet::where('order_id', '=', $order_id)->select();
	// 	foreach ($orderBuffetList as $item) {
	// 		$item->delete();
	// 	}
	// 	$orderBuffetCustomerList = OrderBuffetCustomer::where('order_id', '=', $order_id)->select();
	// 	foreach ($orderBuffetCustomerList as $item) {
	// 		$item->delete();
	// 	}
	// 	$orderDelayList = OrderDelay::where('order_id', '=', $order_id)->select();
	// 	foreach ($orderDelayList as $item) {
	// 		$item->delete();
	// 	}
	// 	// 子单数据
	// 	$subOrderList = (new self)->where('parent_id', $order_id)->select();
	// 	foreach ($subOrderList as $item) {
	// 		$item->force()->delete();
	// 	}
	// 	$opList = OrderProduct::where('order_id', '=', $order_id)->select();
	// 	foreach ($opList as $item) {
	// 		$item->force()->delete();
	// 	}
	// } else {
	// 	/** @var OrderModel $detail */
	// 	$detail->CashierOrderCancels($remark);
	// }

	// 如果是桌台订单
	if billInfo.BillType == 0 && billInfo.DeskUuid > 0 {
		// 拒绝所有待接单 - doto 待对应的服务层实现
		err := qrcodeOrderRepo.Reject(billInfo.DeskUuid)
		if err != nil {
			tx.Rollback()
			return err
		}
		// 关闭桌台
		err = deskRepo.CloseDesk(billInfo.DeskUuid)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		err = orderRepo.CancelOrder(saleBillUuid)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// todo - 添加操作记录
	// 'data' => json_encode($data, JSON_UNESCAPED_UNICODE),
	orderRecordRepo.CreateRecord(model.SaleBillOperationRecord{
		Source:        source,
		Action:        constant.OrderOrderCancel,
		Message:       "整单取消",
		Remark:        "取消订单",
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: billInfo.SaleOrders[0].Uuid,
		OperatorUuid:  staff.Uuid,
	})

	fmt.Println(products)
	fmt.Println(billInfo)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
