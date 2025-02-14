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
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IOrderSrv 定义订单服务接口
type IOrderSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                                              // 创建点餐订单
	CreateDeskOrder(dbId uint64, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                                        // 创建桌台订单
	GetOrderLists(dbId uint64, req req.OrderListReq) (resp.OrderListPaginationResp, error)                                            // 获取订单列表
	GetOrderInfos(dbId uint64, req req.OrderInfoReq) (resp.OrderInfosResp, error)                                                     // 获取订单详情
	CancelOrder(dbId uint64, staff model.Staff, source string, req req.OrderCancelReq) error                                          // 取消订单
	DeleteOrder(dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error                                                         // 删除订单
	IsCellCancelOrder(dbId uint64, saleBillUuid uint64) (model.SaleBill, error)                                                       // 判断桌台是否可取消
	OrderProductDelete(dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (model.SaleBill, error)           // 删除订单商品
	OrderProductChangePrice(dbId uint64, staffUuid uint64, source string, req req.OrderProductChangePriceReq) (model.SaleBill, error) // 修改订单商品价格
	OrderChangePopulation(dbId uint64, staffUuid uint64, source string, req req.OrderChangePopulationReq) (model.SaleBill, error)     // 修改订单商品人数
	OrderProductRemark(dbId uint64, staffUuid uint64, source string, req req.OrderProductRemarkReq) (model.SaleBill, error)           // 修改订单商品备注
}

// orderSrv 订单服务结构
type orderSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv
	settingSrv setting.ISrv
}

// NewOrderSrv 创建订单服务实例
func NewOrderSrv(dbm *database.DBManager, localeSrv ILocaleSrv, cache setting.ISrv) IOrderSrv {
	return NewOrderSrvImpl(dbm, localeSrv, cache)
}

// NewOrderSrvImpl 创建订单服务实例实现
func NewOrderSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IOrderSrv {
	return &orderSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
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

		// 创建订单编号
		orderNo := s.createOrderNo(tx, constant.OrderSourceInstant)
		if orderNo == "" {
			return errors.New("订单编号生成失败")
		}

		// 创建销售账单
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			OrderNo:      orderNo,
			BillType:     constant.OrderSourceMapToBillType[constant.OrderSourceInstant],
			DiningMethod: constant.SaleBillDiningMethodDineIn,
		})
		if err != nil {
			return err
		}

		// 创建销售订单
		saleOrder, err := repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
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

		// 创建订单编号
		orderNo := s.createOrderNo(tx, constant.OrderSourceDesk)
		if orderNo == "" {
			return errors.New("订单编号生成失败")
		}

		// 创建销售账单
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
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

		// 创建销售订单
		saleOrder, err := repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
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
					_, err = buffetRepo.GetBuffetCustomerTypePrice(
						commonRepo.WhereByBuffetPackageUuid(buffetUuid),
						commonRepo.WhereByCustomerTypeUuid(buffetCustomerType.Uuid),
					)
					if err != nil {
						continue
					}

					_, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetCustomerType(model.SaleOrderBuffetCustomerType{
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
		saleBill, err := repository.NewOrderRepo(db).GetSaleBill(repository.NewCommonRepo().WhereByOrderNo(orderNo))
		if err == nil && saleBill.Uuid > 0 {
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

// GetCashierOrderList 获取订单列表
func (s *orderSrv) GetOrderLists(dbId uint64, req req.OrderListReq) (resp.OrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	lists, total, err := orderRepo.GetCashierOrderListWithPagination(reqs)
	if err != nil {
		return resp.OrderListPaginationResp{}, err
	}
	// 组合列表源数据
	billList := make([]resp.BillList, len(lists))
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
		billList[i] = resp.BillList{
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
	return resp.OrderListPaginationResp{
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

// GetOrderInfos 获取收银端订单信息
func (s *orderSrv) GetOrderInfos(dbId uint64, req req.OrderInfoReq) (resp.OrderInfosResp, error) {
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)

	// 获取信息源
	info, err := orderRepo.GetSaleBillDetails(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return resp.OrderInfosResp{}, err
	}

	// 组合信息
	totalMemberNames := []string{}
	payTypes := make([]resp.OrderInfoPayTypes, 0)
	orderList := make([]resp.OrderInfo, len(info.SaleOrders))
	for i, order := range info.SaleOrders {
		payTypeNames := []string{}
		for _, payment := range order.PaymentOrders {
			payTypes = append(payTypes, resp.OrderInfoPayTypes{
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
		// todo - 待完善
		products := make([]resp.OrderProduct, len(order.SaleOrderProducts))
		for j, product := range order.SaleOrderProducts {
			products[j] = resp.OrderProduct{
				Uuid:       product.Uuid,
				LocaleName: s.localeSrv.GetLocaleNames(product.MultiLanguageName),
				FlavorName: product.FlavorName,
				Num:        product.Num,
				// CustomPrice:           product.CustomPrice,
				// UnitPrice:             product.UnitPrice,
				Price:   product.Price,
				TaxRate: product.TaxRate,
				// ProductOriginalAmount: product.ProductOriginalAmount,
				Status:     product.Status,
				Remark:     product.Remark,
				IsGift:     product.IsGift == 1,
				GiftReason: product.GiftReason,
				ImageUrl:   product.ImageFile.GetUrl(),
				Attributes: product.GetAttributeNames(),
			}
		}

		orderList[i] = resp.OrderInfo{
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
	return resp.OrderInfosResp{
		Detail: resp.OrderInfos{
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
			List []resp.OrderOperationLog
		}{
			List: func() []resp.OrderOperationLog {
				logs, err := s.GetRecordList(dbId, req.SaleBillUuid, 0)
				if err != nil {
					return []resp.OrderOperationLog{}
				}
				return logs
			}(),
		},
	}, nil
}

// GetRecordList 获取操作记录
func (s *orderSrv) GetRecordList(dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) ([]resp.OrderOperationLog, error) {
	orderRecordRepo := repository.NewOrderOperationRecordRepo(s.dbm.GetDB(dbId))
	orderRecordLists, err := orderRecordRepo.GetRecordLists(saleBillUuid)
	if err != nil {
		return []resp.OrderOperationLog{}, err
	}
	// todo - 数据格式待处理
	logs := make([]resp.OrderOperationLog, 0)
	for _, record := range orderRecordLists {
		logs = append(logs, resp.OrderOperationLog{
			Uuid:          record.Uuid,
			Source:        record.Source,
			Action:        record.Action,
			Data:          record.Data,
			Remark:        record.Remark,
			SaleBillUuid:  record.SaleBillUuid,
			SaleOrderUuid: record.SaleOrderUuid,
			CreateTime:    record.CreateTime,
		})
	}
	return logs, nil
}

// IsCellCancelOrder 判断订单是否可以取消
func (s *orderSrv) IsCellCancelOrder(dbId uint64, saleBillUuid uint64) (model.SaleBill, error) {
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, 0)
	if err != nil {
		return model.SaleBill{}, err
	}
	if err := billInfo.ValidateOrderStatus(constant.OrderOrderCancel); err != nil {
		return model.SaleBill{}, err
	}
	if orderRepo.IsPartiallyPaid(saleBillUuid) {
		return model.SaleBill{}, errors.New("当前订单已被部分支付，不支持取消")
	}
	return billInfo, nil
}

// CancelOrder 取消订单
func (s *orderSrv) CancelOrder(dbId uint64, staff model.Staff, source string, req req.OrderCancelReq) error {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	productRepo := repository.NewOrderProductRepo(db)
	deskRepo := repository.NewDeskRepo(db)
	qrcodeOrderRepo := repository.NewQrcodeOrderRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := s.IsCellCancelOrder(dbId, req.SaleBillUuid)
	if err != nil {
		return err
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	// 验证高级密码
	if err := s.settingSrv.VerifyAdvancedPassword(dbId, req.Password); err != nil {
		return err
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
		productRepo.WhereSaleBillUuids([]uint64{req.SaleBillUuid}),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// todo 未完成 - 退回商品库存
	for _, po := range products {
		fmt.Println(po)
		// ProductFactory::getFactory($detail['order_source'])->backProductStock([$orderProduct], $isPay);
	}

	// 如果是桌台订单
	if billInfo.BillType == 0 && billInfo.DeskUuid > 0 {
		// 拒绝所有待接单 - todo 待对应的服务层实现
		err := qrcodeOrderRepo.Reject(billInfo.DeskUuid)
		if err != nil {
			tx.Rollback()
			return err
		}
		// 关闭桌台
		err = deskRepo.CloseDesk(billInfo.DeskUuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		err = orderRepo.CancelOrder(req.SaleBillUuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 添加操作日志
	orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderOrderCancel, model.SaleBillOperationRecord{
		Source:        source,
		Remark:        "取消订单",
		SaleOrderUuid: billInfo.SaleOrders[0].Uuid,
		OperatorUuid:  staff.Uuid,
	}, nil)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// DeleteOrder 删除订单, saleOrderUuid = 等于0的时候删除主单，并且主单下的子单也会被删除， saleOrderUuid > 0 的时候删除子单
func (s *orderSrv) DeleteOrder(dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(saleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, 0)
	if err != nil {
		return err
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	if billInfo.Status != constant.SaleBillStatusCanceled {
		return errors.New("订单状态不允许删除")
	}

	err = orderRepo.DeleteOrder(saleBillUuid, saleOrderUuid)
	if err != nil {
		return err
	}

	return nil
}

// OrderProductDelete 删除订单商品
func (s *orderSrv) OrderProductDelete(dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (model.SaleBill, error) {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfoAndProduct(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return model.SaleBill{}, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderDeleteProduct, req.SaleOrderUuid); err != nil {
		return model.SaleBill{}, err
	}

	// 判断订单商品状态
	if len(billInfo.SaleOrders) == 0 || len(billInfo.SaleOrders[0].SaleOrderProducts) == 0 {
		return model.SaleBill{}, errors.New("找不到订单商品")
	}
	for _, product := range billInfo.SaleOrders[0].SaleOrderProducts {
		if product.Uuid == req.OrderProductUuid && product.Status == constant.OrderProductStatusSentKitchen {
			return model.SaleBill{}, errors.New("商品已送厨，禁止删除")
		}
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 删除订单商品
	err = orderRepo.DeleteOrderProduct(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return model.SaleBill{}, err
	}

	// todo - 重算价格 - 等王总的逻辑
	// (new OrderModel)->reloadPrice($order_id);

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return model.SaleBill{}, err
	}

	return billInfo, nil
}

// OrderProductChangePrice  修改订单商品价格
func (s *orderSrv) OrderProductChangePrice(dbId uint64, staffUuid uint64, source string, req req.OrderProductChangePriceReq) (model.SaleBill, error) {
	if req.Price < 0 || req.Price > 1000000 {
		return model.SaleBill{}, errors.New("价格错误")
	}

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	orderProductRepo := repository.NewOrderProductRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return model.SaleBill{}, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
		return model.SaleBill{}, err
	}

	// 判断商品
	product, err := orderProductRepo.GetProductInfoByUuid(req.OrderProductUuid)
	if err != nil {
		return model.SaleBill{}, errors.New("找不到订单商品")
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 修改订单商品价格
	if err := orderRepo.ChangeProductPrice(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Price); err != nil {
		return model.SaleBill{}, err
	}

	// todo - 重算价格 - 等王总的逻辑
	// (new OrderModel)->reloadPrice($order_id);

	// 添加操作日志
	orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderChangePrice, model.SaleBillOperationRecord{
		Source:        source,
		Remark:        "改价",
		SaleBillUuid:  req.SaleBillUuid,
		SaleOrderUuid: req.SaleOrderUuid,
		OperatorUuid:  staffUuid,
	}, map[string]interface{}{
		"order_product_id": req.OrderProductUuid,
		"product_id":       product.Uuid,
		"product_name":     product.Name,
		"total_num":        product.Num,
		"price":            req.Price,
		"product_attr":     product.GetAttributeNames(),
	})

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return model.SaleBill{}, err
	}

	return billInfo, nil
}

// OrderChangePopulation  修改订单人数
func (s *orderSrv) OrderChangePopulation(dbId uint64, staffUuid uint64, source string, req req.OrderChangePopulationReq) (model.SaleBill, error) {
	if req.Population < 0 || req.Population > 999 {
		return model.SaleBill{}, errors.New("人数错误")
	}

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, 0)
	if err != nil {
		return model.SaleBill{}, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderChangePrice, 0); err != nil {
		return model.SaleBill{}, err
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 修改订单商品人数
	if err := orderRepo.ChangePopulation(req.SaleBillUuid, req.Population); err != nil {
		return model.SaleBill{}, err
	}

	// todo - 重算价格 - 等王总的逻辑
	// (new OrderModel)->reloadPrice($order_id);

	// 添加操作日志
	orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderUpdateMealNum, model.SaleBillOperationRecord{
		Source:        source,
		Remark:        "修改桌台就餐人数",
		SaleBillUuid:  req.SaleBillUuid,
		SaleOrderUuid: 0,
		OperatorUuid:  staffUuid,
	}, map[string]interface{}{
		"old_meal_num": billInfo.MealNum,
		"new_meal_num": req.Population,
	})

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return model.SaleBill{}, err
	}

	return billInfo, nil
}

// OrderProductRemark  修改订单商品备注
func (s *orderSrv) OrderProductRemark(dbId uint64, staffUuid uint64, source string, req req.OrderProductRemarkReq) (model.SaleBill, error) {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return model.SaleBill{}, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
		return model.SaleBill{}, err
	}

	// 判断商品
	if len(billInfo.SaleOrders) == 0 || len(billInfo.SaleOrders[0].SaleOrderProducts) == 0 {
		return model.SaleBill{}, errors.New("找不到订单商品")
	}

	// 修改订单商品备注
	if err := orderRepo.ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Remark); err != nil {
		return model.SaleBill{}, err
	}

	return billInfo, nil
}
