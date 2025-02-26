package service

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// IOrderSrv 定义订单服务接口
type IOrderSrv interface {
	CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error)                                                                 // 创建点餐订单
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                                           // 创建桌台订单
	GetOrderLists(dbId uint64, staff model.Staff, source string, req req.OrderListReq) (resp.OrderListPaginationResp, error)                     // 获取订单列表
	GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error)                                                        // 获取订单详情
	CancelOrder(ctx context.Context, req req.OrderCancelReq) error                                                                               // 取消订单
	DeleteOrder(dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error                                                                    // 删除订单
	IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error)                                                          // 判断桌台是否可取消
	OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) // 删除订单商品
	OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error)                                     // 修改订单商品价格
	OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error)                                         // 修改订单人数
	GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error)                                                                             // 通过桌台uuid获取到销售账单信息
	OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq) (*resp.ShopCart, error)                                               // 修改订单商品备注
	CreateSaleBillSetting(ctx context.Context, db *gorm.DB, dbId uint64, saleBillUuid uint64) (model.SaleBillSetting, error)                     // 创建销售账单设置
	GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error)                                                     // 通过设备SN获取点餐购物车信息
	GetOrderCartInfo(ctx context.Context, saleOrderUuid uint64) (*resp.ShopCart, error)                                                          // 获取购物车信息
	InstantOrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error)                                      // 向购物车添加商品
	OrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error)                                             // 修改购物车商品数量
	InstantOrderCartProductNum(ctx context.Context, req req.OrderCartProductNumReq) (*resp.ShopCart, error)                                      // 修改购物车商品数量
	InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, error)                              // 送厨购物车商品
	InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, error)                                         // 获取点餐必点方案
	InstantOrderPaymentInfo(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error)           // 获取结账页面信息
	InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error)              // 给销售订单创建一个支付单
	InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error)                             // 给销售订单创建一个销售订单
	InstantOrderSaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq) (*resp.ShopCart, error)                   // 从一个销售订单移动商品到另一个销售订单
}

// orderSrv 订单服务结构
type orderSrv struct {
	dbm         *database.DBManager // 数据库管理器
	lock        lock.Lock
	localeSrv   ILocaleSrv
	settingSrv  setting.ISrv
	mustPlanSrv IMustPlanSrv
}

// NewOrderSrv 创建订单服务实例
func NewOrderSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, mustPlanSrv IMustPlanSrv) IOrderSrv {
	return NewOrderSrvImpl(dbm, localeSrv, settingSrv, mustPlanSrv)
}

// NewOrderSrvImpl 创建订单服务实例实现
func NewOrderSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, mustPlanSrv IMustPlanSrv) IOrderSrv {
	return &orderSrv{
		dbm:         dbm,
		lock:        lock.NewSystemLock(),
		localeSrv:   localeSrv,
		settingSrv:  settingSrv,
		mustPlanSrv: mustPlanSrv,
	}
}

// CreateInstantOrder 创建点餐订单
func (s *orderSrv) CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error) {
	dbId := ctx.GetDbId()
	var billUuid uint64
	var orderUuid uint64
	db := s.dbm.GetDB(dbId)
	err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {

		// 获取设备uuid
		device, errGetDevice := repository.NewDeviceRepo(db).GetDeviceBySn(ctx, ctx.GetDeviceSn())
		if errGetDevice != nil {
			return errors.New("获取设备uuid失败")
		}

		// 判断是否有待支付、未挂单的订单
		commonRepo := repository.NewCommonRepo()
		orderRepo := repository.NewOrderRepo(tx)
		order, err := orderRepo.GetSaleBill(
			commonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceInstant]),
			commonRepo.WhereByStatus(constant.SaleBillStatusPending),
			commonRepo.WhereByIsHide(false),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil && !errors2.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if order.Uuid > 0 && device.Uuid == order.DeviceUuid {
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
			DeviceUuid:   device.Uuid,
		})
		if err != nil {
			return err
		}

		// 创建销售账单设置
		saleBillSetting, err := s.CreateSaleBillSetting(ctx, tx, dbId, saleBill.Uuid)
		if err != nil {
			return err
		}

		// 创建销售订单
		saleOrder, errCreateSaleOrder := createSaleOrder(tx, &saleBillSetting, saleBill.Uuid, orderNo)
		if errCreateSaleOrder != nil {
			return errCreateSaleOrder
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

func createSaleOrder(db *gorm.DB, saleBillSetting *model.SaleBillSetting, saleBillUuid uint64, saleBillOrderNo string) (*model.SaleOrder, error) {
	// 创建销售订单
	var serviceFee float64
	if saleBillSetting.ServiceFeeType == constant.SaleBillSettingServiceFeeTypeFixed {
		serviceFee = saleBillSetting.ServiceFeeValue
	}
	saleOrder, err := repository.NewOrderRepo(db).CreateSaleOrder(model.SaleOrder{
		SaleBillUuid: saleBillUuid,
		OrderNo:      saleBillOrderNo,
		ServiceFee:   serviceFee,
	})
	if err != nil {
		return nil, err
	}
	return &saleOrder, nil
}

// CreateSaleBillSetting 创建销售账单设置
func (s *orderSrv) CreateSaleBillSetting(ctx context.Context, db *gorm.DB, dbId uint64, saleBillUuid uint64) (model.SaleBillSetting, error) {
	// 获取服务费设置
	serviceFeeSetting, err := s.settingSrv.GetServiceFeeSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, err
	}
	// 获取税率设置
	taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, err
	}
	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, err
	}

	var serviceFeeType uint
	var serviceFeeValue float64
	var taxFeeType uint
	var discountType uint
	var zero uint
	var zeroCheckout uint
	var isStatGift uint = constant.SaleBillSettingIsStatGiftYes
	var isStatFree uint = constant.SaleBillSettingIsStatFreeYes

	// 销售账单服务费
	if serviceFeeSetting.IsOpen == "1" {
		if serviceFeeSetting.ChargeType == "1" {
			serviceFeeType = constant.SaleBillSettingServiceFeeTypeFixed
		}
		if serviceFeeSetting.ChargeType == "2" {
			if serviceFeeSetting.IsOpenTax == "0" {
				serviceFeeType = constant.SaleBillSettingServiceFeeTypePercent
			}
			if serviceFeeSetting.IsOpenTax == "1" {
				serviceFeeType = constant.SaleBillSettingServiceFeeTypePercentTax
			}
		}
		serviceFeeValue, err = strconv.ParseFloat(serviceFeeSetting.ServiceCharge, 64)
		if err != nil {
			return model.SaleBillSetting{}, err
		}
	}

	// 销售账单税率
	if taxRateSetting.IsOpen == "1" {
		if taxRateSetting.CalcType == "1" {
			taxFeeType = constant.SaleBillSettingTaxFeeTypePercentTax
		}
		if taxRateSetting.CalcType == "2" {
			taxFeeType = constant.SaleBillSettingTaxFeeTypePercent
		}
	}

	// 销售账单优惠折扣
	if businessSetting.DiscountMethod == "20" {
		discountType = constant.SaleBillSettingDiscountTypeOff
	}

	// 销售账单优惠折扣自动抹零方式
	zeroingMethod, _ := convertor.ToInt(businessSetting.ZeroingMethod)
	zero = uint(zeroingMethod)

	// 销售账单结账自动抹零方式
	checkoutZeroingMethod, _ := convertor.ToInt(businessSetting.CheckoutZeroingMethod)
	zeroCheckout = uint(checkoutZeroingMethod)

	// 销售账单赠菜计算方式
	if businessSetting.GiftMethod == "20" {
		isStatGift = constant.SaleBillSettingIsStatGiftNone
	}

	// 销售账单免单计算方式
	if businessSetting.FreeMethod == "20" {
		isStatFree = constant.SaleBillSettingIsStatFreeNone
	}

	saleBillSetting, err := repository.NewOrderRepo(db).CreateSaleBillSetting(model.SaleBillSetting{
		SaleBillUuid:    saleBillUuid,
		ServiceFeeType:  serviceFeeType,
		ServiceFeeValue: serviceFeeValue,
		TaxFeeType:      taxFeeType,
		DiscountType:    discountType,
		Zero:            zero,
		ZeroCheckout:    zeroCheckout,
		IsStatGift:      isStatGift,
		IsStatFree:      isStatFree,
	})

	return saleBillSetting, err
}

// CreateDeskOrder 创建桌台订单
func (s *orderSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
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

		// 创建销售账单设置
		saleBillSetting, err := s.CreateSaleBillSetting(ctx, tx, dbId, saleBill.Uuid)
		if err != nil {
			return err
		}

		// 创建销售订单
		var serviceFee float64
		if saleBillSetting.ServiceFeeType == constant.SaleBillSettingServiceFeeTypeFixed {
			serviceFee = saleBillSetting.ServiceFeeValue
		}
		saleOrder, err := repository.NewOrderRepo(tx).CreateSaleOrder(model.SaleOrder{
			SaleBillUuid: saleBill.Uuid,
			OrderNo:      saleBill.OrderNo,
			ServiceFee:   serviceFee,
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
						SaleOrderUuid:               saleOrder.Uuid,
						BuffetPackageUuid:           buffetUuid,
						BuffetCustomerTypePriceUuid: buffetCustomerType.Uuid,
						Num:                         *buffetCustomerType.MealNum,
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

		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			orderNo = ""
			break
		} else {
			break
		}
	}

	return orderNo
}

// GetCashierOrderList 获取订单列表
func (s *orderSrv) GetOrderLists(dbId uint64, staff model.Staff, source string, req req.OrderListReq) (resp.OrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	lists, total, err := orderRepo.GetCashierOrderListWithPagination(reqs)
	if err != nil {
		return resp.OrderListPaginationResp{}, err
	}
	// 组合列表源数据
	billList := make([]resp.BillLists, len(lists))
	consumerUuids := []string{}
	for i, bill := range lists {
		totalPayTypeNames := []string{}
		isSplit := len(bill.SaleOrders) > 1 // 拆单
		orderList := make([]resp.BillListsOrder, 0)
		//
		billListsExtra := resp.BillListsExtra{
			IsCellRefund:        false,
			IsCellCancel:        bill.Status == constant.SaleBillStatusPending,
			IsCellReverseSettle: false,
			IsCellPrint:         !isSplit && bill.Status != constant.SaleBillStatusPending,
			IsCellInvoice:       !isSplit && bill.Status == constant.SaleBillStatusComplete,
			IsCellDelete:        bill.Status == constant.SaleBillStatusCanceled,
		}
		// 拆单
		if isSplit {
			for k, order := range bill.SaleOrders {
				payTypeNames := []string{}
				for _, payment := range order.PaymentOrders {
					totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
					payTypeNames = append(payTypeNames, payment.PaymentMethodName)
				}
				orderExtra := resp.BillListsExtra{
					IsCellRefund:        false,
					IsCellCancel:        order.Status == constant.SaleBillStatusPending,
					IsCellReverseSettle: false,
					IsCellPrint:         !isSplit && order.Status != constant.SaleBillStatusPending,
					IsCellInvoice:       !isSplit && order.Status == constant.SaleBillStatusComplete,
					IsCellDelete:        order.Status == constant.SaleBillStatusCanceled,
				}
				// 不等于免单 && 未全退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					orderExtra.IsCellRefund = true
				}
				// 等于主单 && 完成 && 等于当前用户 && 在班次时间内
				if order.Status == constant.SaleBillStatusComplete && staff.Uuid == bill.CashierUuid && order.FinishTime > staff.CashierLoginTime {
					orderExtra.IsCellReverseSettle = true
				}
				//
				orderList = append(orderList, resp.BillListsOrder{
					SaleBillUuid:  order.SaleBillUuid,
					SaleOrderUuid: order.Uuid,
					BillType:      bill.BillType,
					SerialNo:      bill.SerialNo + "-" + strconv.Itoa(k+1),
					ConsumerUuids: func() string {
						if order.ConsumerUuid == 0 {
							return ""
						}
						return strconv.FormatUint(order.ConsumerUuid, 10)
					}(),
					OrderNo:       order.OrderNo,
					Status:        order.Status,
					FinishTime:    order.FinishTime,
					OrderAmount:   order.Amount,
					PaymentAmount: order.PaymentAmount,
					PayTypeName:   strings.Join(payTypeNames, ","),
					Extra:         orderExtra,
				})
				//
				if order.ConsumerUuid > 0 {
					consumerUuids = append(consumerUuids, strconv.FormatUint(order.ConsumerUuid, 10))
				}
			}
		} else {
			// 没有拆单
			if len(bill.SaleOrders) > 0 {
				order := bill.SaleOrders[0]
				if order.ConsumerUuid > 0 {
					consumerUuids = append(consumerUuids, strconv.FormatUint(order.ConsumerUuid, 10))
				}
				//
				for _, payment := range order.PaymentOrders {
					totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
				}
				// 不等于免单 && 未退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					billListsExtra.IsCellRefund = true
				}
				// 等于主单 && 完成 && 等于当前用户 && 在班次时间内
				if order.Status == constant.SaleBillStatusComplete && staff.Uuid == bill.CashierUuid && order.FinishTime > staff.CashierLoginTime {
					billListsExtra.IsCellReverseSettle = true
				}
			}
		}
		//
		billList[i] = resp.BillLists{
			SaleBillUuid:  bill.Uuid,
			BillType:      bill.BillType,
			IsSplit:       len(bill.SaleOrders) > 1,
			SerialNo:      bill.SerialNo,
			OrderNo:       bill.OrderNo,
			Status:        bill.Status,
			FinishTime:    bill.FinishTime,
			OrderAmount:   bill.Amount,
			PaymentAmount: bill.PaymentAmount,
			ConsumerUuids: strings.Join(consumerUuids, ","),
			PayTypeName:   strings.Join(totalPayTypeNames, ","),
			SaleOrders:    orderList,
			Extra:         billListsExtra,
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
func (s *orderSrv) GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取信息源
	info, err := orderRepo.GetSaleBillDetails(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return resp.OrderInfosResp{}, err
	}
	isMain := req.SaleOrderUuid > 0     // 是否查询主单
	isSplit := len(info.SaleOrders) > 1 // 是否拆单

	// 组合信息
	totalMemberNames := []string{}
	totalMemberUuids := []string{}
	payTypes := make([]resp.OrderInfoPayTypes, 0)
	orderList := make([]resp.OrderInfo, len(info.SaleOrders))
	for i, order := range info.SaleOrders {
		payTypeNames := []string{}
		if order.IsFree == 1 {
			payTypes = append(payTypes, resp.OrderInfoPayTypes{
				Uuid:            0,
				PaymentTypeName: i18n.Translate(ctx.GetLanguage(), "免单"),
				CurrencyUnit:    "",
				PaymentAmount:   order.PaymentAmount,
				Status:          2,
				Source:          0,
				SourceText:      "",
			})
			payTypeNames = append(payTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
		} else {
			for _, payment := range order.PaymentOrders {
				payTypes = append(payTypes, resp.OrderInfoPayTypes{
					Uuid:            payment.Uuid,
					PaymentTypeName: payment.PaymentMethodName,
					CurrencyUnit:    payment.CurrencyUnit,
					PaymentAmount:   payment.PaymentAmount,
					Status:          uint(payment.Status),
					Source:          uint(payment.PaymentMethod.Source),
					SourceText:      payment.PaymentMethod.GetSourceText(ctx.GetLanguage()),
				})
				payTypeNames = append(payTypeNames, payment.PaymentMethodName)
			}
		}
		if order.Member.Nickname != "" && !slices.Contains(totalMemberNames, order.Member.Nickname) {
			totalMemberNames = append(totalMemberNames, order.Member.Nickname)
		}
		if order.ConsumerUuid != 0 {
			totalMemberUuids = append(totalMemberUuids, strconv.FormatUint(order.ConsumerUuid, 10))
		}
		//
		products := make([]resp.OrderProduct, len(order.SaleOrderProducts))
		for j, product := range order.SaleOrderProducts {
			products[j] = resp.OrderProduct{
				Uuid:           product.Uuid,
				LocaleName:     product.MultiLanguageName.GetNames(),
				FlavorName:     product.FlavorName,
				Num:            product.Num,
				Price:          product.Price,
				SalePrice:      product.SalePrice,
				TotalPrice:     decimal.NewFromFloat(product.TotalPrice).Mul(decimal.NewFromInt(int64(product.Num))).InexactFloat64(),
				TotalSalePrice: decimal.NewFromFloat(product.SalePrice).Mul(decimal.NewFromInt(int64(product.Num))).InexactFloat64(),
				TaxRate:        product.TaxRate,
				Status:         product.Status,
				Remark:         product.Remark,
				IsGift:         product.IsGiftProduct(),
				GiftReason:     product.GiftReason,
				ImageUrl:       product.ImageFile.GetUrl(),
				Attributes:     product.GetAttributeNames(),
				CancelReason:   product.CancelReason,
				// todo 待完善
				RefundAmount: 0,
			}
		}
		// todo - SerialNo 取值不对
		orderList[i] = resp.OrderInfo{
			SaleOrderUuid: order.Uuid,
			BillType:      info.BillType,
			DiningMethod:  info.DiningMethod,
			SerialNo:      info.SerialNo + "-" + strconv.Itoa(i+1),
			OrderNo:       order.OrderNo,
			Status:        order.Status,
			IsFree:        order.IsFree == 1,
			FreeReason:    order.FreeReason,
			OrderAmount:   order.Amount,
			PaymentAmount: order.PaymentAmount - order.GetTotalRefundAmount(),
			RefundAmount:  order.GetTotalRefundAmount(),
			PayTypeName:   strings.Join(payTypeNames, ","),
			MemberName:    order.Member.Nickname,
			MemberUuid:    order.ConsumerUuid,
			Products:      products,
		}
	}

	// 处理额外信息
	order := info.SaleOrders[0]
	orderExtra := resp.BillListsExtra{
		IsCellRefund: false,
		IsCellPrint:  (!isSplit || !isMain) && order.Status != constant.SaleBillStatusPending,
		IsCellDelete: order.Status == constant.SaleBillStatusCanceled,
	}
	if (!isSplit || !isMain) && order.IsFree == 0 && info.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
		orderExtra.IsCellRefund = true
	}

	// 返回响应对象
	return resp.OrderInfosResp{
		Detail: resp.OrderInfos{
			SaleBillUuid: info.Uuid,
			IsSplit:      isSplit,
			BillType:     info.BillType,
			SerialNo:     info.SerialNo,
			OrderNo: func() string {
				if isMain {
					return info.OrderNo
				}
				return order.OrderNo
			}(),
			Status:        info.Status,
			CreateTime:    info.CreateTime,
			FinishTime:    info.FinishTime,
			OrderAmount:   info.Amount,
			PaymentAmount: info.PaymentAmount - info.GetTotalRefundAmount(),
			RefundAmount:  info.GetTotalRefundAmount(),
			MemberNames:   strings.Join(totalMemberNames, ","),
			MemberUuids:   strings.Join(totalMemberUuids, ","),
			CashierName:   info.Cashier.RealName,
			IsBuffet:      info.IsBuffet == 1,
			BuffetNames:   info.GetBuffetNames(ctx.GetLanguage()),
			CancelReason:  info.Reason,
			PayTypes:      payTypes,
			SaleOrders:    orderList,
			Remark:        info.Remark,
		},
		OperationLog: struct {
			List []resp.OrderOperationLog `json:"list"`
		}{
			List: func() []resp.OrderOperationLog {
				logs, err := s.GetRecordList(ctx.GetDbId(), req.SaleBillUuid, 0)
				if err != nil {
					return []resp.OrderOperationLog{}
				}
				return logs
			}(),
		},
		Extra: orderExtra,
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
		// 解析Data字段的JSON字符串为map
		var data any
		if record.Data != "" {
			if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
				data = struct{}{}
			}
		} else {
			data = struct{}{}
		}
		//
		logs = append(logs, resp.OrderOperationLog{
			Uuid:          record.Uuid,
			Source:        record.Source,
			Action:        record.Action,
			Data:          data,
			Remark:        record.Remark,
			SaleBillUuid:  record.SaleBillUuid,
			SaleOrderUuid: record.SaleOrderUuid,
			CreateTime:    record.CreateTime,
		})
	}
	return logs, nil
}

// IsCellCancelOrder 判断订单是否可以取消
func (s *orderSrv) IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, constant.OptionalUuid)
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
func (s *orderSrv) CancelOrder(ctx context.Context, req req.OrderCancelReq) error {
	dbId := ctx.GetDbId()
	staff := ctx.GetStaff()
	source := ctx.GetSource()
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	productRepo := repository.NewOrderProductRepo(db)
	deskRepo := repository.NewDeskRepo(db)
	qrcodeOrderRepo := repository.NewH5OrderRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := s.IsCellCancelOrder(ctx, req.SaleBillUuid)
	if err != nil {
		return err
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	// 验证高级密码
	if err := s.settingSrv.VerifyAdvancedPassword(ctx, req.Password); err != nil {
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
		SaleBillUuid:  billInfo.SaleOrders[0].Uuid,
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
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, constant.OptionalUuid)
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

	// 检查是否有已送厨的商品，如果有，则标记production_order_product.status为消单退菜（制作中消单退菜、制作完成消单退菜）
	// 如果已送厨商品还在制作中，通知厨房取消制作
	doingProductList := make([]uint64, 0) // 制作中的商品uuid列表 sale_order_product_uuid
	// todo 获取还在制作中的商品

	// 发布事件，通知厨房取消制作
	event.NewSystemBus().PublishCancelDoingProductEvent(event.CancelDoingProductPayload{SaleOrderProductUuids: doingProductList})
	return nil
}

// OrderProductDelete 删除订单商品
func (s *orderSrv) OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	//orderRepo := repository.NewOrderRepo(db)

	// 获取操作的销售账单信息
	saleBill, saleOrder, saleOrderProduct, errGetSaleOrder := getSaleOrderFromDB(ctx, db, req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if errGetSaleOrder != nil {
		ctx.Log().Error("改价商品时，查询销售订单信息失败", zap.Error(errGetSaleOrder))
		return nil, errors.New("查询销售订单信息失败")
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDeleteProduct, req.SaleOrderUuid); err != nil {
		return nil, err
	}

	// 判断订单商品状态
	if len(saleBill.SaleOrders) == 0 || len(saleBill.SaleOrders[0].SaleOrderProducts) == 0 {
		return nil, errors.New("找不到订单商品")
	}
	for _, product := range saleBill.SaleOrders[0].SaleOrderProducts {
		if product.Uuid == req.OrderProductUuid && product.Status == constant.OrderProductStatusSentKitchen {
			return nil, errors.New("商品已送厨，禁止删除")
		}
	}

	saleOrderProduct.DeleteProduct()

	// 计算订单金额
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	ctx.Log().Debug("删除商品前,销售订单信息", zap.Any("saleOrder calc", saleOrder.BeforeCalc()))
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	afterSaleOrderCalc := saleOrder.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	ctx.Log().Debug("删除商品后,销售订单信息", zap.Any("saleOrder calc", afterSaleOrderCalc))

	errUpdate := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 删除订单商品
		err := repository.NewOrderRepo(db).DeleteOrderProduct(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
		if err != nil {
			return err
		}
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
			return errUpdate
		}
		return nil
	})
	if errUpdate != nil {
		ctx.Log().Info("更新销售订单失败", zap.Error(errUpdate))
		return nil, errors.New("更新销售订单失败")
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func getSaleOrderFromDB(ctx context.Context, db *gorm.DB, saleBillUuid, saleOrderUuid, saleOrderProductUuid uint64) (*model.SaleBill, *model.SaleOrder, *model.SaleOrderProduct, error) {
	newSaleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, nil, nil, errSaleBill
	}
	var newSaleOrder *model.SaleOrder
	var newSaleOrderProduct *model.SaleOrderProduct
	for i, _ := range newSaleBill.SaleOrders {
		order := newSaleBill.SaleOrders[i]
		if order.Uuid == saleOrderUuid {
			newSaleOrder = newSaleBill.SaleOrders[i]
			break
		}
	}
	for i, order := range newSaleBill.SaleOrders {
		for j, product := range order.SaleOrderProducts {
			if product.Uuid == saleOrderProductUuid {
				newSaleOrderProduct = newSaleBill.SaleOrders[i].SaleOrderProducts[j]
				break
			}
		}
	}
	if newSaleOrder == nil {
		ctx.Log().Error("改价商品时无法查询到销售订单信息", zap.Any("saleBillUuid", saleBillUuid), zap.Any("saleOrderUuid", saleOrderUuid))
		return nil, nil, nil, errors.New("业务错误")
	}
	if newSaleOrderProduct == nil {
		ctx.Log().Error("改价商品时无法查询到销售订单商品信息", zap.Any("saleBillUuid", saleBillUuid), zap.Any("saleOrderUuid", saleOrderUuid), zap.Any("saleOrderProductUuid", saleOrderProductUuid))
		return nil, nil, nil, errors.New("业务错误")
	}
	return newSaleBill, newSaleOrder, newSaleOrderProduct, nil
}

// OrderProductChangePrice  修改订单商品价格
func (s *orderSrv) OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error) {
	if req.Price < 0 || req.Price > 1000000 {
		return nil, errors.New("价格错误")
	}

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, saleOrder, saleOrderProduct, errGetSaleOrder := getSaleOrderFromDB(ctx, db, req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if errGetSaleOrder != nil {
		ctx.Log().Error("改价商品时，查询销售订单信息失败", zap.Error(errGetSaleOrder))
		return nil, errors.New("查询销售订单信息失败")
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
		return nil, err
	}

	ctx.Log().Debug("改价前", zap.Any("SalePrice", saleOrderProduct.SalePrice))
	ctx.Log().Debug("改价前", zap.Any("saleOrderProduct calc", saleOrderProduct.BeforeCalc()))
	// 改价
	saleOrderProduct.ChangeProductPrice(req.Price)
	ctx.Log().Debug("改价后", zap.Any("SalePrice", saleOrderProduct.SalePrice))

	// 计算商品数据。折扣、税费、服务
	serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	afterCalc := saleOrderProduct.CalcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)
	ctx.Log().Debug("改价后", zap.Any("saleOrderProduct calc", afterCalc))

	// 计算订单金额
	ctx.Log().Debug("改价前,销售订单信息", zap.Any("saleOrder calc", saleOrder.BeforeCalc()))
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	afterSaleOrderCalc := saleOrder.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	ctx.Log().Debug("改价后,销售订单信息", zap.Any("saleOrder calc", afterSaleOrderCalc))

	errUpdate := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
			return errUpdate
		}
		return nil
	})
	if errUpdate != nil {
		ctx.Log().Info("更新销售订单失败", zap.Error(errUpdate))
		return nil, errors.New("更新销售订单失败")
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	// 发布事件，记录销售账单操作记录
	event.NewSystemBus().PublishAddSaleBillRecordEvent(event.AddSaleBillRecordPayload{
		SaleBillUuid:     req.SaleBillUuid,
		SaleOrderUuid:    req.SaleOrderUuid,
		OrderProductUuid: req.OrderProductUuid,
		OrderProductName: saleOrderProduct.Name,
		OrderProductNum:  saleOrderProduct.Num,
		CompanyUuid:      ctx.GetCompanyUuid(),
		StaffUuid:        ctx.GetStaffUuid(),
		Price:            req.Price,
		AttributeNames:   saleOrderProduct.GetAttributeNames(),
		Source:           ctx.GetSource(),
	})

	return info, nil
}

// OrderChangePopulation  修改订单人数
func (s *orderSrv) OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error) {
	if req.Population < 0 || req.Population > 999 {
		return nil, errors.New("人数错误")
	}

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, constant.OptionalUuid)
	if err != nil {
		return nil, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderChangePrice, 0); err != nil {
		return nil, err
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
		return nil, err
	}

	// todo - 重算价格 - 等王总的逻辑
	// (new OrderModel)->reloadPrice($order_id);

	// 添加操作日志
	orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderUpdateMealNum, model.SaleBillOperationRecord{
		Source:        ctx.GetSource(),
		Remark:        "修改桌台就餐人数",
		SaleBillUuid:  req.SaleBillUuid,
		SaleOrderUuid: 0,
		OperatorUuid:  ctx.GetStaffUuid(),
	}, map[string]interface{}{
		"old_meal_num": billInfo.MealNum,
		"new_meal_num": req.Population,
	})

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// GetSaleBillByDeskId  获取桌台账单信息
func (s *orderSrv) GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error) {
	dbId := ctx.GetDbId()
	deskUuid := ctx.GetDeskUuid()

	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 通过桌台查找到当前桌台的正在进行销售账单
	billInfo, err := orderRepo.GetSaleBillInfoByDesk(deskUuid, constant.OptionalUuid)
	if err != nil {
		return model.SaleBill{}, err
	}
	return billInfo, nil
}

// OrderProductRemark  修改订单商品备注
func (s *orderSrv) OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	_ = ctx.GetStaff().Uuid // 员工ID
	_ = ctx.GetSource()     // 操作来源
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, err
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderRemark, req.SaleOrderUuid); err != nil {
		return nil, err
	}

	// 判断商品
	if len(billInfo.SaleOrders) == 0 || len(billInfo.SaleOrders[0].SaleOrderProducts) == 0 {
		return nil, errors.New("找不到订单商品")
	}

	// 修改订单商品备注
	if err := orderRepo.ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Remark); err != nil {
		return nil, err
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// 通过设备SN获取销售账单uuid
func (s *orderSrv) getSaleBillUuidByDeviceSn(ctx context.Context, deviceSn string) (uint64, error) {
	var saleBillUuid uint64
	// 通过设备sn查询设备ID
	db := s.dbm.GetDB(ctx.GetDbId())
	device, errDevice := repository.NewDeviceRepo(db).GetDeviceBySn(ctx, deviceSn)
	if errDevice != nil {
		return 0, errors.New(errDevice.Error())
	}
	ctx.Log().Debug("通过device_sn查询设备uuid", zap.Any("deviceSn", deviceSn), zap.Any("device_uuid", device.Uuid))
	if device.IsDelete() {
		return 0, errors.NewWithCode(constant.CodeParamError, "设备不存在")
	}
	ctx.Log().Debug("通过设备ID查询未挂单的销售账单", zap.Any("device_uuid", device.Uuid))
	// 通过设备ID查询未挂单的销售账单
	if saleBill, errGetSaleBill := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(device.Uuid); errGetSaleBill != nil {
		if errors2.Is(errGetSaleBill, gorm.ErrRecordNotFound) {
			return 0, nil // 没有点餐账单
		}
		return 0, errors.New(errGetSaleBill.Error())
	} else {
		saleBillUuid = saleBill.Uuid
	}
	ctx.Log().Debug("查询购物车信息", zap.Any("saleBillUuid", saleBillUuid))
	return saleBillUuid, nil
}
func (s *orderSrv) GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error) {
	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx, deviceSn)
	if errUuid != nil {
		ctx.Log().Info("无法找到销售账单", zap.Error(errUuid))
		return nil, errors.New("无法找到销售账单")
	}
	// 查询购物车信息
	cartInfo, errInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errInfo != nil {
		return nil, errInfo
	}
	return cartInfo, nil
}

// GetOrderCartInfo 获取点餐购物车信息
func (s *orderSrv) GetOrderCartInfo(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 通过销售订单ID得到订单商品列表、订单金额信息、账单的销售订单列表

	shopCart, err := orderRepo.GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, err
	}

	// 给订单列表添加订单
	saleOrderList := make([]resp.SaleOrder, 0)
	for _, saleOrder := range shopCart.SaleBill.SaleOrders {
		productList := make([]resp.Product, 0)
		// 给商品列表条件顾客类型
		// 如不是桌台订单、不是自助餐，这个Buffets列表是空的，故不会往productList里加入商品
		{
			for _, buffet := range saleOrder.SaleOrderBuffetCustomerTypes {
				if buffet.IsDelete() {
					continue
				}
				// 自助餐顾客价格收费列表
				product := resp.Product{
					Uuid:       buffet.Uuid,
					LocaleName: buffet.BuffetPackageMultiLanguageName.GetNames(),
					LocaleAttributeName: dto.LocaleResponse{
						ZH:   buffet.Name,
						TH:   buffet.Name,
						EN:   buffet.Name,
						ZHTW: buffet.Name,
						JA:   buffet.Name,
						KO:   buffet.Name,
						MY:   buffet.Name,
						TR:   buffet.Name,
					},
					Num:           buffet.Num, // 这种类型顾客多少个，如老人这个类型2人
					SalePrice:     buffet.GetOriginPrice(),
					DiscountPrice: buffet.GetOriginPrice(),
					Status:        1,
					Remark:        "",
					IsMust:        false,
					IsGift:        false,
					IsCancel:      false,
					AboutBuffet: resp.AboutBuffet{
						IsCustomer: true,
						IsDelay:    false,
					},
				}
				productList = append(productList, product)
			}
		}

		// 添加加钟商品
		{
			for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if delayProduct.IsDelete() {
					continue
				}
				product := resp.Product{
					Uuid: delayProduct.Uuid,
					LocaleName: dto.LocaleResponse{
						ZH:   delayProduct.Name,
						TH:   delayProduct.Name,
						EN:   delayProduct.Name,
						ZHTW: delayProduct.Name,
						JA:   delayProduct.Name,
						KO:   delayProduct.Name,
						MY:   delayProduct.Name,
						TR:   delayProduct.Name,
					},
					LocaleAttributeName: dto.LocaleResponse{},
					Num:                 shopCart.SaleBill.MealNum, // 等于桌台人数
					SalePrice:           delayProduct.GetPrice(shopCart.SaleBill.MealNum),
					DiscountPrice:       0,  // 加钟商品没有优惠价
					Status:              1,  // 添加后标记送厨状态，不可修改
					Remark:              "", // 加钟商品没有备注
					IsMust:              false,
					IsGift:              false,
					IsCancel:            false,
					AboutBuffet: resp.AboutBuffet{
						IsCustomer: false,
						IsDelay:    true, // 标记该商品是加钟商品
					},
				}
				productList = append(productList, product)
			}
		}

		// 添加正常商品
		{
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() {
					continue
				}
				language := ctx.GetLanguage()
				product := resp.Product{
					Uuid:                saleOrderProduct.Uuid,
					LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: *saleOrderProduct.AttributeName(language),
					Num:                 saleOrderProduct.Num,
					SalePrice:           saleOrderProduct.GetSalePrice(),
					DiscountPrice:       saleOrderProduct.GetPrice(),
					Status:              saleOrderProduct.StatusValue(),
					Remark:              saleOrderProduct.Remark,
					IsMust:              saleOrderProduct.IsMustProduct(),
					IsGift:              saleOrderProduct.IsGiftProduct(),
					IsCancel:            saleOrderProduct.IsCancelProduct(),
				}
				productList = append(productList, product)
			}
		}

		// 商品计数
		productNum := 0
		for _, product := range productList {
			productNum += int(product.Num)
		}

		// 填写订单信息
		order := resp.SaleOrder{
			Uuid:        saleOrder.Uuid,
			OrderNo:     saleOrder.OrderNo,
			ProductNum:  productNum,
			ProductList: productList,
			// 订单金额信息
			AmountInfo: resp.AmountInfo{
				ProductOriginalAmount: saleOrder.ProductOriginalAmount,
				ProductAmount:         saleOrder.ProductAmount,
				ServiceAmount:         saleOrder.ServiceFee,
				TaxAmount:             saleOrder.TaxFee,
				DiscountAmount:        decimal.NewFromFloat(saleOrder.CustomDiscountFee).Add(decimal.NewFromFloat(saleOrder.ZeroFee)).Round(2).InexactFloat64(),
				MemberDiscountAmount:  saleOrder.MemberDiscountFee,
				Amount:                saleOrder.Amount,
			},
		}
		saleOrderList = append(saleOrderList, order)
	}

	// 获取必点方案列表
	mustPlan, errPlanMust := s.InstantOrderMustPlan(ctx, ctx.GetDeviceSn())
	if errPlanMust != nil {
		ctx.Log().Info("获取必点方案列表失败", zap.Error(errPlanMust))
	}
	var productMustPlanList *resp.ProductMustPlanList
	if mustPlan != nil {
		productMustPlanList = &resp.ProductMustPlanList{
			List: mustPlan.List,
		}
	}

	shopCartInfo := &resp.ShopCart{
		SaleBillUuid:  saleBillUuid,
		IsDeskOrder:   shopCart.IsDeskShopCart(),
		IsLock:        shopCart.SaleBill.IsLock == 1,
		Desk:          nil,
		Buffet:        nil,
		DiningMethod:  shopCart.SaleBill.DiningMethod,
		SaleOrderList: saleOrderList,
	}
	// 如果要显示必点信息
	if productMustPlanList != nil {
		shopCartInfo.MustPlans = productMustPlanList
	}
	// 如果是桌台购物车
	if shopCart.IsDeskShopCart() {
		deskInfo := resp.DeskInfo{
			Uuid:      shopCart.SaleBill.Desk.Uuid,
			DeskNo:    shopCart.SaleBill.Desk.DeskNo,
			MealNum:   shopCart.SaleBill.MealNum,
			StartTime: shopCart.SaleBill.CreateTime,
			Duration:  time.Now().Unix() - shopCart.SaleBill.Desk.CreateTime,
		}
		shopCartInfo.Desk = &deskInfo
		// 如果是自助餐桌台
		if shopCart.SaleBill.IsBuffetSaleBill() {
			shopCartInfo.Buffet = &resp.BuffetInfo{
				RemainingSeconds: shopCart.SaleBill.BuffetRemainingSeconds(),
				LocaleName:       shopCart.SaleBill.GetBuffetName(),
			}
		}
	}
	return shopCartInfo, nil
}

// 点餐页面，往购物车添加商品。
func (s *orderSrv) InstantOrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error) {
	// 当不填销售账单ID时，表示要新建一个销售账单
	if req.SaleBillUuid == 0 {
		order, err := s.CreateInstantOrder(ctx)
		if err != nil {
			ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
			return nil, errors.New(err.Error())
		}
		ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
		req.SaleBillUuid = order.SaleBillUuid
		req.SaleOrderUuid = order.SaleOrderUuid
	}

	// 往销售账单里添加商品
	shopCart, err := s.OrderCartProductAdd(ctx, req)
	if err != nil {
		ctx.Log().Info("往点餐账单里添加商品失败", zap.Any("req", req), zap.Any("error", err))
		return nil, errors.New(err.Error())
	}
	return shopCart, nil
}

// OrderCartProductAdd 向购物车添加商品
func (s *orderSrv) OrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error) {
	s.lock.LockUuid(req.SaleBillUuid)
	defer s.lock.UnlockUuid(req.SaleBillUuid)
	var saleOrderProduct *model.SaleOrderProduct
	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	//var saleBill *model.SaleBill
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	var saleOrder *model.SaleOrder
	for i, _ := range saleBill.SaleOrders {
		order := saleBill.SaleOrders[i]
		if order.Uuid == req.SaleOrderUuid {
			saleOrder = saleBill.SaleOrders[i]
			break
		}
	}
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 录入订单商品数据
	{
		// 获取商品包信息
		productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(req.FlavorUuid)
		if err != nil {
			return nil, err
		}
		if productBom.IsDelete() {
			return nil, errors.New("商品规格已经删除")
		}

		// 获取某商品规格信息
		flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(req.FlavorUuid)
		if errFlavorProductBom != nil {
			return nil, errFlavorProductBom
		}

		// 获取加料信息
		sauceProductBoms := make(map[uint64]*model.ProductBom)
		if len(req.SauceUuidList) > 0 {
			sauceProductBomList, errSauceProductBomList := repository.NewProductBomRepo(db).GetSauceProductBomsByUuids(req.SauceUuidList)
			if errSauceProductBomList != nil {
				return nil, errSauceProductBomList
			}
			for i, bom := range sauceProductBomList {
				sauceProductBoms[bom.Uuid] = sauceProductBomList[i]
			}
		}

		// 获取属性信息
		var productAttributes map[uint64]*model.ProductPackageAttribute
		if len(productAttributes) > 0 {
			productAttributeList, errProductAttributeList := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(req.AttributeUuidList)
			if errProductAttributeList != nil {
				return nil, errProductAttributeList
			}
			for i, attribute := range productAttributeList {
				productAttributes[attribute.Uuid] = productAttributeList[i]
			}
		}

		// 判断销售账单是否存在且未结账

		productPackage := productBom.ProductPackage
		sauces := make([]model.Sauce, 0)
		for sauceProductBomUuid, sauceProductBom := range sauceProductBoms {
			sauce := model.Sauce{
				Name:           sauceProductBom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
				Price:          sauceProductBom.Price,
				ProductBomUuid: sauceProductBomUuid,
			}
			sauces = append(sauces, sauce)
		}

		attributes := make([]model.Attribute, 0)
		for productAttributeUuid, productAttribute := range productAttributes {
			attribute := model.Attribute{
				Name:                 productAttribute.Attribute.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
				ProductAttributeUuid: productAttributeUuid,
			}
			attributes = append(attributes, attribute)
		}
		saleOrderProduct = model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
			Name:                   productPackage.Name,
			OpenMemberDiscount:     productPackage.OpenDiscount,
			TaxRate:                productPackage.TaxRate(saleBill.DiningMethod),
			DeductStockType:        productPackage.DeductStockType,
			MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
			ImageFileUuid:          productPackage.ImageFileUuid,
			ProductPackageUuid:     productPackage.Uuid,
			SaleBillUuid:           req.SaleBillUuid,
			SaleOrderUuid:          req.SaleOrderUuid,
			MemberDiscountRate:     saleOrder.MemberDiscountRate,
			MemberCardDiscountRate: saleOrder.MemberCardDiscountRate,
			CustomDiscountRate:     saleOrder.CustomDiscountRate,
			Sauces:                 sauces,
			Flavor: model.Flavor{
				Name:           flavorProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 填顾客下单时规格的名字 todo preload
				Price:          flavorProductBom.Price,
				ProductBomUuid: req.FlavorUuid,
			},
			Attribute: attributes,
		})
		// 设置必点信息
		var mustPlanUuid uint64
		mustPlanUuid, err = s.mustPlanSrv.GetMustPlanUuidByProductPackage(ctx, req.SaleBillUuid, productPackage.Uuid, saleBill.DeskUuid)
		ctx.Log().Debug("获取到必点方案uuid", zap.Any("mustPlanUuid", mustPlanUuid))
		saleOrderProduct.SetMustPlanInfo(mustPlanUuid)
	}
	// 计算商品数据。折扣、税费、服务
	serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	saleOrderProduct.CalcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)

	sign := saleOrderProduct.GenerateProductSign()
	saleOrderProduct.Sign = sign
	ctx.Log().Debug("生成商品签名", zap.Any("sign", sign))

	// 查询是否存在该商品
	ctx.Log().Debug("查询是否存在该商品", zap.Any("sign", sign), zap.Any("saleBillUuid", req.SaleBillUuid), zap.Any("saleOrderUuid", req.SaleOrderUuid))
	orderProduct, err := repository.NewOrderProductRepo(db).GetProductInfo(
		repository.CommonRepo.WhereBySign(sign),
		repository.CommonRepo.WhereBySaleBillUuid(req.SaleBillUuid),
		repository.CommonRepo.WhereBySaleOrderUuid(req.SaleOrderUuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		ctx.Log().Debug("查询数据失败", zap.Error(err))
		return nil, errors.New("查询数据失败")
	}

	if orderProduct.Uuid != 0 {
		ctx.Log().Debug("查询到数据，更新数据")
		orderProduct.Num += saleOrderProduct.Num
		orderProduct.SalePrice = saleOrderProduct.SalePrice
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(orderProduct); errUpdate != nil {
			ctx.Log().Error("添加商品失败，更新数据失败", zap.Error(errUpdate))
			return nil, errors.New("添加商品失败")
		}
	} else {
		ctx.Log().Debug("没查询到数据，新建销售订单商品数据")
		// 创建销售订单商品
		_, errCreate := repository.NewSaleOrderProductRepo(db).CreateSaleOrderProduct(*saleOrderProduct)
		if errCreate != nil {
			return nil, errCreate
		}
	}

	newSaleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	var newSaleOrder *model.SaleOrder
	for i, _ := range newSaleBill.SaleOrders {
		order := newSaleBill.SaleOrders[i]
		if order.Uuid == req.SaleOrderUuid {
			newSaleOrder = newSaleBill.SaleOrders[i]
			break
		}
	}

	productNum := len(newSaleBill.SaleOrders[0].SaleOrderProducts)
	ctx.Log().Info("订单产品数量", zap.Any("productNum", productNum))
	// 计算订单金额
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	newSaleOrder.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	// 更新订单记录
	if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(newSaleOrder); errUpdate != nil {
		return nil, errUpdate
	}

	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// OrderCartProductNum 修改购物车商品数量
func (s *orderSrv) InstantOrderCartProductNum(ctx context.Context, req req.OrderCartProductNumReq) (*resp.ShopCart, error) {
	s.lock.LockUuid(req.SaleBillUuid)
	defer s.lock.UnlockUuid(req.SaleBillUuid)
	db := s.dbm.GetDB(ctx.GetDbId())

	// 检查商品销售库存是否充足
	// todo
	ctx.Log().Debug("获取账单信息")
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	ctx.Log().Debug("获取到账单信息成功")

	saleOrder, errSaleOrder := getSaleOrder(req.SaleOrderUuid, saleBill)
	if errSaleOrder != nil {
		return nil, errSaleOrder
	}
	ctx.Log().Debug("获取到订单信息成功")

	// 获取销售订单商品信息
	saleOrderProduct, index, errSaleOrderProduct := getSaleOrderProduct(req.SaleOrderProductUuid, saleOrder)
	if errSaleOrderProduct != nil {
		return nil, errSaleOrderProduct
	}
	ctx.Log().Debug("获取到订单商品信息成功")

	// 修改销售订单商品数量
	saleOrderProduct.Num = uint(req.Num)
	ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

	// 计算商品数据。折扣、税费、服务
	serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	saleOrderProduct.CalcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)
	ctx.Log().Debug("重新计算了商品金额", zap.Any("saleOrderProduct salePrice", saleOrderProduct.SalePrice))
	saleOrder.SaleOrderProducts[index] = saleOrderProduct

	// 计算订单金额
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	calc := saleOrder.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))

	if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
		return nil, errUpdate
	}
	ctx.Log().Debug("更新销售订单商品成功")

	if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
		return nil, errUpdate
	}
	ctx.Log().Debug("更新销售订单成功")

	// // 更新销售账单
	// if errUpdate := repository.NewSaleBillRepo(db).UpdateSaleBill(saleBill); errUpdate != nil {
	// 	return nil, errUpdate
	// }
	// ctx.Log().Debug("更加数据")

	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}
	ctx.Log().Debug("获取新的账单数据")

	return info, nil
}

func getSaleOrder(saleOrderUuid uint64, saleBill *model.SaleBill) (*model.SaleOrder, error) {
	for i, _ := range saleBill.SaleOrders {
		order := saleBill.SaleOrders[i]
		if order.Uuid == saleOrderUuid {
			return saleBill.SaleOrders[i], nil
		}
	}
	return nil, errors.New("销售订单不存在")
}

func getSaleOrderProduct(saleOrderProductUuid uint64, saleOrder *model.SaleOrder) (*model.SaleOrderProduct, int, error) {
	for i, _ := range saleOrder.SaleOrderProducts {
		orderProduct := saleOrder.SaleOrderProducts[i]
		if orderProduct.Uuid == saleOrderProductUuid {
			return saleOrder.SaleOrderProducts[i], i, nil
		}
	}
	return nil, 0, errors.New("销售订单商品不存在")
}

func getSaleOrderProductUnCooking(saleOrder *model.SaleOrder) ([]*model.SaleOrderProduct, error) {
	unCookingSaleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for i, _ := range saleOrder.SaleOrderProducts {
		orderProduct := saleOrder.SaleOrderProducts[i]
		if orderProduct.Status == constant.SaleOrderProductStatusNormal {
			unCookingSaleOrderProducts = append(unCookingSaleOrderProducts, saleOrder.SaleOrderProducts[i])
		}
	}
	return unCookingSaleOrderProducts, nil
}

// InstantOrderCartProductCooking 送厨购物车商品
func (s *orderSrv) InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, error) {
	s.lock.LockUuid(req.SaleBillUuid)
	defer s.lock.UnlockUuid(req.SaleBillUuid)
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	ctx.Log().Debug("获取销售账单信息")
	// 获取销售订单信息
	saleOrder, errSaleOrder := getSaleOrder(req.SaleOrderUuid, saleBill)
	if errSaleOrder != nil {
		return nil, errSaleOrder
	}

	ctx.Log().Debug("获取销售订单信息")
	// 获取销售订单商品信息
	unCookingSaleOrderProducts, errUnCookingSaleOrderProducts := getSaleOrderProductUnCooking(saleOrder)
	if errUnCookingSaleOrderProducts != nil {
		return nil, errUnCookingSaleOrderProducts
	}

	for index, _ := range unCookingSaleOrderProducts {
		product := unCookingSaleOrderProducts[index]
		product.Status = constant.SaleOrderProductStatusCooking
	}

	productionOrder := newProductionOrder(ctx, req.SaleOrderUuid, req.SaleBillUuid, unCookingSaleOrderProducts)

	ctx.Log().Debug("准备开始更新")
	errUpdate := db.Transaction(func(tx *gorm.DB) error {
		// 修改订单商品状态为已送厨
		errUpdateSaleProductStatus := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(unCookingSaleOrderProducts)
		if errUpdateSaleProductStatus != nil {
			ctx.Log().Debug("商品状态更新失败", zap.Error(errUpdateSaleProductStatus))
			return errors.New(errUpdateSaleProductStatus.Error())
		}
		ctx.Log().Debug("商品状态成功")
		errCreateProduction := repository.NewProductionOrderRepo(tx).CreateProductionOrder(productionOrder)
		if errCreateProduction != nil {
			ctx.Log().Debug("创建送厨单失败", zap.Error(errCreateProduction))
			return errors.New(errCreateProduction.Error())
		}
		return nil
	})
	if errUpdate != nil {
		ctx.Log().Debug("更新数据失败", zap.Any("error", errUpdate))
		return nil, errors.New(errUpdate.Error())
	}

	ctx.Log().Debug("获取新的购物车信息")
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.New(errGetCartInfo.Error())
	}
	return cartInfo, nil
}

func newProductionOrder(ctx context.Context, saleOrderUuid, saleBillUuid uint64, unCookingSaleOrderProducts []*model.SaleOrderProduct) *model.ProductionOrder {
	productionOrderUuid, _ := utils.GetID()
	productionOrderProducts := make([]*model.ProductionOrderProduct, 0)
	for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {
		productionOrderProduct := model.ProductionOrderProduct{
			ProductionOrderUuid:   productionOrderUuid,
			SaleOrderProductUuid:  unCookingSaleOrderProduct.Uuid,
			FirstCategoryUuid:     unCookingSaleOrderProduct.ProductPackage.ProductCategory.GetFirstCategoryUuid(),
			Num:                   unCookingSaleOrderProduct.Num,
			FlavorName:            unCookingSaleOrderProduct.Name,
			ProductAttributeNames: unCookingSaleOrderProduct.AttributeName(ctx.GetLanguage()).GetLocale(ctx.GetLanguage()),
			Status:                constant.ProductionOrderProductStatusCooking,
			Remark:                unCookingSaleOrderProduct.Remark,
			//HasMaterial:              unCookingSaleOrderProduct, todo
			ProductionOrderMaterials: unCookingSaleOrderProduct.GetMaterialBom(), // 获取这个商品各个材料的用量
		}
		productionOrderProducts = append(productionOrderProducts, &productionOrderProduct)
	}
	productionOrder := model.ProductionOrder{
		BaseModel:               model.BaseModel{Uuid: productionOrderUuid},
		SaleOrderUuid:           saleOrderUuid,
		SaleBillUuid:            saleBillUuid,
		ProductionOrderProducts: productionOrderProducts,
	}
	return &productionOrder
}

// 获取点餐的必点方案的商品列表
// 只获取要在“必点”弹框中显示的必点方案
//func getInstantMustPlanProductList(mustPlan *model.ProductMustPlan) []resp.InstantMustPlanProductStat {
//	// 不是点餐的必单方案
//	if !mustPlan.IsInstantMustPlan() {
//		return []resp.InstantMustPlanProductStat{}
//	}
//	// 不是自动加购的必点方案都不显示中“必点”弹框中
//	if !mustPlan.IsAutoCart() {
//		return []resp.InstantMustPlanProductStat{}
//	}
//	productList := make([]resp.InstantMustPlanProductStat, 0)
//	// 点餐的必点方案都是“每单必点”，没有“每人必点”类型的
//	// 如果是固定商品
//	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
//		for _, planItem := range mustPlan.ProductMustPlanItems {
//			productPackage := planItem.GetProductInfo()
//			if productPackage == nil {
//				continue
//			}
//			productStat := resp.InstantMustPlanProductStat{
//				Product:     *productPackage,
//				IsAutoAdd:   mustPlan.IsAutoCart() && planItem.ProductPackage.IsNoSelectProduct(), // 必点方案勾选自动加购且商品是无选择商品
//				SelectedNum: 0,                                                                    // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
//				MustNum:     1,                                                                    // 该商品要求必点数量
//				NeedNum:     1,                                                                    // 该商品还需点点数量。展示给前端之前判断购物车内是否已经点了该商品，减已点数量
//			}
//			productList = append(productList, productStat)
//		}
//	}
//	// 如果是可选商品
//	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
//		for _, planItem := range mustPlan.ProductMustPlanItems {
//			productPackage := planItem.GetProductInfo()
//			if productPackage == nil {
//				continue
//			}
//			productStat := resp.InstantMustPlanProductStat{
//				Product:     *productPackage,
//				IsAutoAdd:   false, // 可选商品的必点方案都不自动加购
//				SelectedNum: 0,     // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
//				MustNum:     0,
//				NeedNum:     0,
//			}
//			productList = append(productList, productStat)
//		}
//	}
//	return productList
//}
//
//// 获取点餐的必点方案的商品列表。用于加购商品时判断该商品是不是这些点餐列表里的商品
//func getInstantMustPlanProductList2(mustPlan *model.ProductMustPlan) []resp.InstantMustPlanProductStat {
//	// 不是点餐的必单方案
//	if !mustPlan.IsInstantMustPlan() {
//		return []resp.InstantMustPlanProductStat{}
//	}
//	// 不是自动加购的必点方案都不显示中“必点”弹框中
//	if !mustPlan.IsAutoCart() {
//		return []resp.InstantMustPlanProductStat{}
//	}
//	productList := make([]resp.InstantMustPlanProductStat, 0)
//	// 点餐的必点方案都是“每单必点”，没有“每人必点”类型的
//	// 如果是固定商品
//	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
//		for _, planItem := range mustPlan.ProductMustPlanItems {
//			productPackage := planItem.GetProductInfo()
//			if productPackage == nil {
//				continue
//			}
//			productStat := resp.InstantMustPlanProductStat{
//				Product:     *productPackage,
//				IsAutoAdd:   mustPlan.IsAutoCart() && planItem.ProductPackage.IsNoSelectProduct(), // 必点方案勾选自动加购且商品是无选择商品
//				SelectedNum: 0,                                                                    // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
//				MustNum:     1,                                                                    // 该商品要求必点数量
//				NeedNum:     1,                                                                    // 该商品还需点点数量。展示给前端之前判断购物车内是否已经点了该商品，减已点数量
//			}
//			productList = append(productList, productStat)
//		}
//	}
//	// 如果是可选商品
//	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
//		for _, planItem := range mustPlan.ProductMustPlanItems {
//			productPackage := planItem.GetProductInfo()
//			if productPackage == nil {
//				continue
//			}
//			productStat := resp.InstantMustPlanProductStat{
//				Product:     *productPackage,
//				IsAutoAdd:   false, // 可选商品的必点方案都不自动加购
//				SelectedNum: 0,     // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
//				MustNum:     0,
//				NeedNum:     0,
//			}
//			productList = append(productList, productStat)
//		}
//	}
//	return productList
//}

// 获取点餐的必点方案列表，用于自动加购和显示在“必点”弹框中
//func (s *orderSrv) getInstantMustPlanList(ctx context.Context, db *gorm.DB, shopCart *ro.ShopCartRepo) (*[]resp.InstantProductMustPlan, error) {
//	mustPlanList := make([]resp.InstantProductMustPlan, 0)
//
//	shopCartMustProductInfo := shopCart.GetMustPlanProductInfo()
//
//	// 获取必选方案信息
//	productMustPlans, err := repository.NewProductMustPlanRepo(db).GetProductMustPlanListAllInfos(ctx)
//	if err != nil {
//		return nil, errors.New(err.Error())
//	}
//
//	// 构建必点方案响应列表
//	for _, plan := range productMustPlans {
//		// 不自动加购的必点方案不用显示必选弹框，所以跳过
//		if !plan.IsAutoCart() {
//			continue
//		}
//		if plan.IsDeskMustPlan() {
//			// 忽略桌台必点方案
//			continue
//		}
//		if plan.IsDelete() {
//			continue
//		}
//		//productPackageList := make([]resp.InstantMustPlanProductStat, 0)
//		productPackageList := getInstantMustPlanProductList(plan)
//		// 如果列表为空，跳过不显示
//		if len(productPackageList) == 0 {
//			continue
//		}
//		autoAddProductNum := 0 // 自动加购商品包的数量
//		for _, product := range productPackageList {
//			if product.IsAutoAdd {
//				autoAddProductNum++
//			}
//		}
//
//		// 获取购物车中已经点了多个这个必点方案的商品
//		selectedNum := uint(0)
//		productPackageMap, ok := shopCartMustProductInfo[plan.Uuid]
//		if ok {
//			for _, productPackage := range productPackageList {
//				// 不统计自动加购的商品
//				if productPackage.IsAutoAdd {
//					continue
//				}
//				productPackageUuid := productPackage.Product.Uuid
//				num := productPackageMap[productPackageUuid]
//				selectedNum += num
//			}
//		}
//		// 获取购物车中各个必点商品已经点了多少个，还差多少个
//		mustMap := make(map[uint64]uint) // product_package_uuid => num 每个必点商品还差多少个
//		if ok {
//			for _, productPackage := range productPackageList {
//				// 不统计自动加购的商品
//				if productPackage.IsAutoAdd {
//					continue
//				}
//				productPackageUuid := productPackage.Product.Uuid
//				num := productPackageMap[productPackageUuid] // 该商品已点xx个
//				mustNum := productPackage.MustNum            // 该商品要求点的数量
//				result := mustNum - num
//				if result > 0 {
//					mustMap[productPackageUuid] = result
//				}
//			}
//		}
//
//		// 如果必点方案是可选商品，NeedNum的取值要么是1 要么是0
//		// 当selectedNum>0时，NeedNum为0
//		needNum := uint(0)
//		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
//			if selectedNum > 0 {
//				needNum = 0
//			} else {
//				needNum = 1
//			}
//		}
//		// 如果必点方案是固定商品时，NeedNum的取值为“必选弹框”列表中商品还差数量之和
//		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
//			for _, num := range mustMap {
//				needNum += num
//			}
//		}
//
//		mustPlan := resp.InstantProductMustPlan{
//			Name:     plan.Name,
//			MustType: plan.GetMustType(),
//			MustRule: plan.GetMustRule(),
//			//IsAutoCart:   plan.IsAutoCart(),
//			CanChangeNum: plan.IsCustomerCanChange(),
//			SelectedNum:  selectedNum, // 已选择xx份
//			NeedNum:      needNum,     // 还差xx份。不应该算上自动加购商品的数量
//			Products:     resp.ProductPackageList{List: productPackageList},
//		}
//		mustPlanList = append(mustPlanList, mustPlan)
//	}
//
//	return &mustPlanList, nil
//}

// 获取点餐的获取点餐的必点方案列表，用于检查加购的商品是否是必点商品并且属于哪个必点方案
//func (s *orderSrv) getDeskMustPlanList(ctx context.Context, db *gorm.DB, shopCart *ro.ShopCartRepo) (*[]resp.InstantProductMustPlan, error) {
//	mustPlanList := make([]resp.InstantProductMustPlan, 0)
//
//	shopCartMustProductInfo := shopCart.GetMustPlanProductInfo()
//
//	// 获取必选方案信息
//	productMustPlans, err := repository.NewProductMustPlanRepo(db).GetProductMustPlanListAllInfos(ctx)
//	if err != nil {
//		return nil, errors.New(err.Error())
//	}
//
//	// 构建必点方案响应列表
//	for _, plan := range productMustPlans {
//		if plan.IsDeskMustPlan() {
//			// 忽略桌台必点方案
//			continue
//		}
//		if plan.IsDelete() {
//			continue
//		}
//		//productPackageList := make([]resp.InstantMustPlanProductStat, 0)
//		productPackageList := getInstantMustPlanProductList(plan)
//		// 如果列表为空，跳过不显示
//		if len(productPackageList) == 0 {
//			continue
//		}
//
//		// 获取购物车中已经点了多个这个必点方案的商品
//		selectedNum := uint(0)
//		productPackageMap, ok := shopCartMustProductInfo[plan.Uuid]
//		if ok {
//			for _, productPackage := range productPackageList {
//				// 不统计自动加购的商品
//				if productPackage.IsAutoAdd {
//					continue
//				}
//				productPackageUuid := productPackage.Product.Uuid
//				num := productPackageMap[productPackageUuid]
//				selectedNum += num
//			}
//		}
//		// 获取购物车中各个必点商品已经点了多少个，还差多少个
//		mustMap := make(map[uint64]uint) // product_package_uuid => num 每个必点商品还差多少个
//		if ok {
//			for _, productPackage := range productPackageList {
//				productPackageUuid := productPackage.Product.Uuid
//				num := productPackageMap[productPackageUuid] // 该商品已点xx个
//				mustNum := productPackage.MustNum            // 该商品要求点的数量
//				result := mustNum - num
//				if result > 0 {
//					mustMap[productPackageUuid] = result
//				} else {
//					mustMap[productPackageUuid] = 0
//				}
//			}
//		}
//
//		// 如果必点方案是可选商品，NeedNum的取值要么是1 要么是0
//		// 当selectedNum>0时，NeedNum为0
//		needNum := uint(0)
//		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
//			if selectedNum > 0 {
//				needNum = 0
//			} else {
//				needNum = 1
//			}
//		}
//		// 如果必点方案是固定商品时，NeedNum的取值为“必选弹框”列表中商品还差数量之和
//		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
//			for _, num := range mustMap {
//				needNum += num
//			}
//		}
//
//		mustPlan := resp.InstantProductMustPlan{
//			Name:         plan.Name,
//			MustType:     plan.GetMustType(),
//			MustRule:     plan.GetMustRule(),
//			CanChangeNum: plan.IsCustomerCanChange(),
//			SelectedNum:  selectedNum, // 已选择xx份
//			NeedNum:      needNum,     // 还差xx份。不应该算上自动加购商品的数量
//			Products:     resp.ProductPackageList{List: productPackageList},
//		}
//		mustPlanList = append(mustPlanList, mustPlan)
//	}
//
//	return &mustPlanList, nil
//}

// InstantOrderMustPlan 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx, deviceSn)
	if errUuid != nil {
		ctx.Log().Info("无法找到销售账单", zap.Error(errUuid))
		return nil, errors.New("无法找到销售账单")
	}
	ctx.Log().Debug("查询必点方案列表", zap.Any("saleBillUuid", saleBillUuid), zap.Any("deviceSn", deviceSn))

	mustPlanList := make([]resp.InstantProductMustPlan, 0)
	// product_bom_uuid => *resp.InstantMustPlanProduct
	autoFlavorProduct := make(map[uint64]*resp.InstantMustPlanProduct) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购

	// 查询到购物车信息
	shopCart2, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, err
	}

	if !shopCart2.SaleBill.IsShowMustPlan() {

	}

	planList, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCart2)
	if errMustPlan != nil {
		ctx.Log().Info("获取必点列表失败", zap.Error(errMustPlan))
		return nil, errors.New("获取必点列表失败")
	}
	mustPlanList = *planList
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// 遍历得到要自动加购的商品
	for i, plan := range mustPlanList {
		for j, product := range plan.Products.List {
			if product.IsAutoAdd {
				planProduct := mustPlanList[i].Products.List[j].Product
				productFlavorBomUuid := planProduct.Flavors.List[0].Uuid
				autoFlavorProduct[productFlavorBomUuid] = &planProduct
			}
		}
	}

	var shopCart *resp.ShopCart
	// 判断是否需要给点餐账单自动加购商品。当map列表中有商品时，表示需要自动加购
	if len(autoFlavorProduct) > 0 && shopCart2.SaleBill.IsAutoAddMustProduct() {
		errTx := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			shopCart, err = autoAddSaleOrderProduct(ctx, db, s, autoFlavorProduct)
			if err != nil {
				ctx.Log().Debug("自动添加必点商品失败", zap.Error(err))
				return errors.New(err.Error())
			}
			return nil
		})
		if errTx != nil {
			ctx.Log().Debug("自动添加必点商品失败", zap.Error(err))
			return nil, errors.New(errTx.Error())
		}
	}

	var cartInfo *resp.InstantShopCart
	if shopCart != nil {
		cartInfo = &resp.InstantShopCart{
			SaleBillUuid:  shopCart.SaleBillUuid,
			DiningMethod:  shopCart.DiningMethod,
			SaleOrderList: shopCart.SaleOrderList,
		}
	}

	list := make([]resp.InstantProductMustPlan, 0)
	if shopCart2.SaleBill.IsShowMustPlan() {
		list = mustPlanList
	}
	return &resp.InstantProductMustPlanResp{List: list, ShopCartInfo: cartInfo}, nil
}

func autoAddSaleOrderProduct(ctx context.Context, db *gorm.DB, s *orderSrv, autoFlavorProduct map[uint64]*resp.InstantMustPlanProduct) (*resp.ShopCart, error) {
	var saleBillUuid uint64
	deviceSn := ctx.GetDeviceSn()
	ctx.Log().Debug("autoAddSaleOrderProduct", zap.Any("deviceSn", deviceSn))
	if deviceSn == "" {
		ctx.Log().Debug("自动加购必选商品失败，上下文中没有device_sn")
		return nil, errors.New("自动加购必选商品失败，上下文中没有device_sn")
	}

	device, errDevice := repository.NewDeviceRepo(db).GetDeviceBySn(ctx, deviceSn)
	if errDevice != nil {
		return nil, errors.New(errDevice.Error())
	}
	ctx.Log().Debug("通过device_sn查询设备uuid", zap.Any("deviceSn", deviceSn), zap.Any("device_uuid", device.Uuid))
	if device.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeParamError, "设备不存在")
	}
	ctx.Log().Debug("通过设备ID查询未挂单的销售账单2222", zap.Any("device_uuid", device.Uuid))
	// 通过设备ID查询未挂单的销售账单
	saleBill, errGetSaleBill := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(device.Uuid)
	if errGetSaleBill != nil {
		if !errors2.Is(errGetSaleBill, gorm.ErrRecordNotFound) {
			return nil, errors.New(errGetSaleBill.Error())
		}
	}
	if saleBill != nil {
		saleBillUuid = saleBill.Uuid
	}

	ctx.Log().Debug("查询到的账单", zap.Any("saleBillUuid", saleBillUuid))

	var shopCartInfo *resp.ShopCart
	if saleBillUuid == 0 {
		// 如果没有未挂单的点餐销售账单，则新建一个点餐销售账单并接入自动必点商品
		ctx.Log().Debug("没有未挂单的点餐销售账单，新建一个点餐销售账单并接入自动必点商品")
		saleBillUuid := uint64(0)
		saleOrderUuid := uint64(0)
		for flavorUuid, _ := range autoFlavorProduct {
			if saleBillUuid != 0 {
				ctx.Log().Debug("新建的销售账单号", zap.Any("saleBillUuid", saleBillUuid))
			}
			ctx.Log().Debug("添加商品", zap.Any("flavorUuid", flavorUuid))
			shopCart, errAdd := s.InstantOrderCartProductAdd(ctx, req.OrderCartProductAddReq{
				SaleBillUuid:  saleBillUuid,
				SaleOrderUuid: saleOrderUuid,
				FlavorUuid:    flavorUuid,
			})
			if errAdd != nil {
				return nil, errAdd
			}
			saleBillUuid = shopCart.SaleBillUuid
			if len(shopCart.SaleOrderList) != 1 {
				return nil, errors.New("业务错误")
			}
			saleOrderUuid = shopCart.SaleOrderList[0].Uuid
			shopCartInfo = shopCart
		}

	} else {
		// 如果有未挂单的点餐销售账单。未有这个需求，暂时不做
	}
	return shopCartInfo, nil
}

// InstantOrderPaymentInfo 获取结账页面信息
func (s *orderSrv) InstantOrderPaymentInfo(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	s.lock.LockUuid(saleBillUuid)
	defer s.lock.UnlockUuid(saleBillUuid)
	// 获取销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	var saleOrder *model.SaleOrder
	saleOrder = saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}

	var paymentMethods []*model.PaymentMethod
	paymentMethods = repository.NewPaymentMethodRepo(db).GetPaymentMethodsByCtx(ctx)
	if len(paymentMethods) <= 0 {
		return nil, errors.New("系统没有支付方式")
	}

	memberInfo := resp.MemberInfo{}
	if saleOrder != nil && saleOrder.Member.MemberCard != nil && saleOrder.Member.MemberLevel != nil {
		memberInfo = resp.MemberInfo{
			Uuid:    saleOrder.Member.Uuid,
			Name:    saleOrder.Member.Nickname,
			Card:    resp.CardInfo{Name: saleOrder.Member.MemberCard.MemberCardType.Name},
			Level:   resp.LevelInfo{Name: saleOrder.Member.MemberLevel.Name},
			Balance: saleOrder.Member.Balance,
			Points:  saleOrder.Member.Point,
		}
	}

	paymentOrders := make([]resp.PaymentOrder, 0)
	for _, paymentOrder := range saleOrder.PaymentOrders {
		order := resp.PaymentOrder{
			Uuid:                 paymentOrder.Uuid,
			PaymentMethodUuid:    paymentOrder.PaymentMethodUuid,
			PaymentMethodName:    paymentOrder.PaymentMethodName,
			PaymentMethodCode:    paymentOrder.PaymentMethod.Code,
			PaymentAmount:        paymentOrder.PaymentAmount,
			PaymentCommissionFee: paymentOrder.PaymentCommissionFee,
			Amount:               paymentOrder.Amount,
			DisabledCancel:       paymentOrder.PaymentMethod.IsDisabledCancel(),
		}
		paymentOrders = append(paymentOrders, order)
	}

	methodItems := make([]resp.PaymentMethodItem, 0)
	amounts := make([]resp.PaymentMethodAmount, 0)

	serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	for _, paymentMethod := range paymentMethods {
		var logoUrl string
		var qrcodeUrl string
		if paymentMethod.LogoFile != nil {
			logoUrl = paymentMethod.LogoFile.GetUrl()
		}
		if paymentMethod.QrcodeFile != nil {
			qrcodeUrl = paymentMethod.QrcodeFile.GetUrl()
		}
		methodItem := resp.PaymentMethodItem{
			Source:      paymentMethod.Source,
			SourceText:  paymentMethod.GetSourceText(ctx.GetLanguage()),
			Uuid:        paymentMethod.Uuid,
			PaymentName: paymentMethod.PaymentName,
			FeePercent:  paymentMethod.FeePercent,
			Logo:        logoUrl,
			Qrcode:      qrcodeUrl,
			Code:        paymentMethod.Code,
		}
		methodItems = append(methodItems, methodItem)

		commissionFee := saleOrder.CalcCommissionFee()

		if commissionFee > 0 {
			// 如果有手续费
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrder.CalcOrderOriginAmount(serviceFeeRate, serviceFeeValue, taxFeeType),
				SaleOrderAmount:       saleOrder.Amount,
				UnpaidAmount:          saleOrder.CalcUnPayAmount(true),
				ZeroAmount:            0, // 只有没有手续费时才会抹零
				ZeroRule:              constant.SaleBillSettingCheckoutZeroingMethodNone,
				PaymentMethodUuid:     methodItem.Uuid,
				CommissionFee:         commissionFee,
			}
			amounts = append(amounts, amount)
		} else {
			// 如果没有手续费
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrder.CalcOrderOriginAmount(serviceFeeRate, serviceFeeValue, taxFeeType),
				SaleOrderAmount:       saleOrder.Amount,
				UnpaidAmount:          saleOrder.CalcUnPayAmount(false),
				ZeroAmount:            saleOrder.CalcCheckOutZeroFee(), // 只有没有手续费时才会抹零
				ZeroRule:              saleOrder.ZeroCheckoutRule,
				PaymentMethodUuid:     methodItem.Uuid,
				CommissionFee:         commissionFee,
			}
			amounts = append(amounts, amount)
		}
	}

	infoResp := &resp.InstantOrderPaymentInfoResp{
		MemberInfo:     memberInfo,
		PaymentOrders:  resp.PaymentInfoList{List: paymentOrders},
		PaymentMethods: resp.PaymentMethodList{List: methodItems},
		Amounts:        resp.PaymentMethodAmountList{List: amounts},
	}

	return infoResp, nil
}

// InstantOrderPaymentCreate 给销售订单创建一个支付单
func (s *orderSrv) InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error) {
	//db := s.dbm.GetDB(ctx.GetDbId())
	return nil, nil
}

// InstantOrderSaleOrderCreate 给销售订单创建一个销售订单
func (s *orderSrv) InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 加锁
	saleBillUuid := req.SaleBillUuid
	s.lock.LockUuid(saleBillUuid)
	defer s.lock.UnlockUuid(saleBillUuid)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 创建销售订单
	_, errCreateSaleOrder := createSaleOrder(db, saleBill.SaleBillSetting, saleBill.Uuid, saleBill.OrderNo)
	if errCreateSaleOrder != nil {
		ctx.Log().Error("新建拆单失败", zap.Any("errCreateSaleOrder", errCreateSaleOrder))
		return nil, errors.New("新建拆单失败")
	}

	cartInfo, errCartInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errCartInfo != nil {
		ctx.Log().Error("查询购物车信息失败", zap.Any("errCartInfo", errCartInfo))
		return nil, errors.New("查询购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderSaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
func (s *orderSrv) InstantOrderSaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq) (*resp.ShopCart, error) {
	// 需要更新的销售订单商品
	waitUpdateSaleOrderProductMap := make(map[uint64]*model.SaleOrderProduct)

	db := s.dbm.GetDB(ctx.GetDbId())
	// 加锁
	saleBillUuid := req.SaleBillUuid
	s.lock.LockUuid(saleBillUuid)
	defer s.lock.UnlockUuid(saleBillUuid)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	// 获取销售订单信息
	saleOrderFrom := saleBill.GetSaleOrder(req.From)
	saleOrderTo := saleBill.GetSaleOrder(req.To)

	// 构建移动到订单商品的map结构
	moveProductMap := make(map[uint64]uint)
	for _, moveProduct := range req.Products {
		moveProductMap[moveProduct.Uuid] = moveProduct.Num
	}

	// 构建原销售订单中的商品map
	fromProductMap := make(map[uint64]*model.SaleOrderProduct)
	for index, saleOrderProduct := range saleOrderFrom.SaleOrderProducts {
		fromProductMap[saleOrderProduct.Uuid] = saleOrderFrom.SaleOrderProducts[index]
	}

	// 构建目标销售订单中的商品签名map
	toSaleOrderProductSignMap := make(map[string]*model.SaleOrderProduct)
	for i, saleOrderProduct := range saleOrderTo.SaleOrderProducts {
		toSaleOrderProductSignMap[saleOrderProduct.Sign] = saleOrderTo.SaleOrderProducts[i]
	}

	// 检查购物车商品是否有变动
	// 选择移动的商品中有商品已经不在原销售订单中
	for _, moveProduct := range req.Products {
		if _, ok := fromProductMap[moveProduct.Uuid]; !ok {
			return nil, errors.New("购物车商品变化，从重新操作")
		}
	}

	// 遍历要移动的订单商品，移动到目标订单中
	for _, moveProduct := range req.Products {
		saleOrderProduct := fromProductMap[moveProduct.Uuid]
		// 记录到待更新列表中
		waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
		// 如果原销售订单商品数量还有剩余，则在目标销售订单中新建一个销售订单商品
		if saleOrderProduct.Num > moveProduct.Num {
			// 复制一个销售订单商品
			var newSaleOrderProduct model.SaleOrderProduct
			newSaleOrderProduct = *saleOrderProduct
			uuid, _ := utils.GetID()
			newSaleOrderProduct.BaseModel = model.BaseModel{}
			newSaleOrderProduct.Uuid = uuid
			newSaleOrderProduct.SaleOrderUuid = req.To
			newSaleOrderProduct.Num = moveProduct.Num
			sign := newSaleOrderProduct.GenerateProductSign()
			newSaleOrderProduct.Sign = sign
			ctx.Log().Debug("生产销售订单商品签名", zap.Any("saleOrderProduct uuid", newSaleOrderProduct.Uuid), zap.Any("sign", sign))
			// 检查to销售账单中是否有与该商品签名一样的销售订单商品，若有则合并他们
			if orderProduct, exit := toSaleOrderProductSignMap[sign]; exit {
				// 目标销售账单中有商品签名一样的商品，将数量累加到这个商品上
				orderProduct.Num += moveProduct.Num
				// 单价不变，不用重新计算。

				// 记录到待更新列表中
				waitUpdateSaleOrderProductMap[orderProduct.Uuid] = orderProduct
			} else {
				// 目标销售账单中没有商品签名一样的商品，则在目标销售账单中新建一个销售订单商品
				// 修改订单商品的折扣优惠，并重新计算金额
				discountInfo := saleOrderTo.GetDiscountInfo()
				newSaleOrderProduct.SetDiscountInfo(discountInfo.MemberDiscountRate, discountInfo.MemberCardDiscountRate, discountInfo.CustomDiscountRate)
				// 计算商品数据。折扣、税费、服务
				serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
				taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
				serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
				newSaleOrderProduct.CalcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)
				// 在目标销售账单中新建一个销售订单商品
				//saleOrderTo.SaleOrderProducts = append(saleOrderTo.SaleOrderProducts, &newSaleOrderProduct)

				// 记录到待更新列表中
				waitUpdateSaleOrderProductMap[newSaleOrderProduct.Uuid] = &newSaleOrderProduct
			}
		} else {
			// 如果原销售订单商品的数量没有剩余，则修改该销售订单商品的销售订单uuid为目标销售订单的uuid
			saleOrderProduct.SaleOrderUuid = req.To
			sign := saleOrderProduct.GenerateProductSign()
			// 检查to销售账单中是否有与该商品签名一样的销售订单商品，若有则合并他们
			if orderProduct, exit := toSaleOrderProductSignMap[sign]; exit {
				// 目标销售账单中有商品签名一样的商品，将数量累加到这个商品上
				orderProduct.Num += saleOrderProduct.Num
				// 单价没有改变，不需要重新计算

				// 记录到待更新列表中
				waitUpdateSaleOrderProductMap[orderProduct.Uuid] = orderProduct
			} else {
				// 目标销售账单中没有商品签名一样的商品，则在目标销售账单中新建一个销售订单商品
				// saleOrderProduct更新这个记录即可
				discountInfo := saleOrderTo.GetDiscountInfo()
				saleOrderProduct.SetDiscountInfo(discountInfo.MemberDiscountRate, discountInfo.MemberCardDiscountRate, discountInfo.CustomDiscountRate)
				// 计算商品数据。折扣、税费、服务
				serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
				taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
				serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
				saleOrderProduct.CalcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)

				// 记录到待更新列表中
				waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
			}
		}
	}

	// 计算订单金额
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderTo calc", saleOrderTo.BeforeCalc()))
	afterSaleOrderCalc := saleOrderTo.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderTo calc", afterSaleOrderCalc))

	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderFrom calc", saleOrderFrom.BeforeCalc()))
	afterSaleOrderFromCalc := saleOrderFrom.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderFrom calc", afterSaleOrderFromCalc))

	var cartInfo *resp.ShopCart
	errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, saleOrderProduct := range waitUpdateSaleOrderProductMap {
			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProduct(saleOrderProduct); err != nil {
				return err
			}
		}
		if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrder(saleOrderFrom); err != nil {
			return err
		}
		if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrder(saleOrderTo); err != nil {
			return err
		}

		info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
		if err != nil {
			return err
		}
		cartInfo = info
		return nil
	})
	if errUpdateDB != nil {
		return nil, errors.New("更新数据失败")
	}

	return cartInfo, nil
}
