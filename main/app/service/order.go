package service

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	event2 "ttpos-server-go/app/event"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/admin"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/app/service/lianlian"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"

	"go.uber.org/zap"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// IOrderSrv 定义订单服务接口
type IOrderSrv interface {
	CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error)                                                                   // 创建点餐订单
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                                             // 创建桌台订单
	GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error)                                                 // 获取订单列表
	GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error)                                                          // 获取订单详情
	CancelOrder(ctx context.Context, req req.OrderCancelReq) error                                                                                 // 取消订单
	DeleteOrder(ctx context.Context, dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error                                                 // 删除订单
	ReturnOrder(ctx context.Context, req req.OrderReturnReq) error                                                                                 // 退款订单
	GetReturnOrderInfo(ctx context.Context, req req.OrderReturnInfoReq) (*resp.OrderReturnInfoResp, error)                                         // 获取退款信息
	GetReverseSettleInfo(ctx context.Context, req req.OrderReverseSettleInfoReq) (*resp.OrderReverseSettleInfoResp, error)                         // 获取反结账信息
	ReverseSettle(ctx context.Context, req req.OrderReverseSettleReq) error                                                                        // 反结账
	IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error)                                                            // 判断桌台是否可取消
	HideOrder(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error)                                                                    // 挂单
	ShowOrder(ctx context.Context, req req.OrderShowReq) (*resp.ShopCart, error)                                                                   // 显示订单
	InstantHideOrderList(ctx context.Context, req req.HideSaleBillListReq) (*resp.InstantHideOrderListResp, error)                                 // 获取挂单订单列表
	OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error)                                                             // 打包
	OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error)   // 删除订单商品
	OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error)                                       // 修改订单商品价格
	OrderAmountChange(ctx context.Context, req req.OrderAmountChangeReq) (*resp.ShopCart, error)                                                   // 修改订单金额
	OrderDiscount(ctx context.Context, req req.OrderDiscountReq) (*resp.ShopCart, error)                                                           // 修改订单折扣
	OrderZeroRule(ctx context.Context, req req.OrderZeroRuleReq) (*resp.ShopCart, error)                                                           // 修改订单抹零规则
	OrderDiscountCancel(ctx context.Context, req req.OrderDiscountCancelReq) (*resp.ShopCart, error)                                               // 取消点餐订单所有优惠折扣，包括改价、打折、抹零
	OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error)                                           // 修改订单人数
	OrderChangeBuffet(ctx context.Context, req req.OrderChangeBuffetReq) (*resp.ShopCart, error)                                                   // 调整自助餐
	OrderChangeBuffetClock(ctx context.Context, req req.OrderChangeBuffetClockReq) (*resp.ShopCart, error)                                         // 调整自助餐
	OrderDeskBuffetProductList(ctx context.Context, req req.OrderChangeBuffetProductListReq) (*resp.BuffetProductList, error)                      // 获取桌台的自助餐商品列表
	GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error)                                                                               // 通过桌台uuid获取到销售账单信息
	OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq) (*resp.ShopCart, error)                                                 // 修改订单商品备注
	CreateSaleBillSetting(ctx context.Context, db *gorm.DB, dbId uint64, saleBillUuid uint64) (model.SaleBillSetting, error)                       // 创建销售账单设置
	GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error)                                                       // 通过设备SN获取点餐购物车信息
	GetOrderCartInfo(ctx context.Context, saleOrderUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                // 获取购物车信息
	InstantOrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error)                                        // 向购物车添加商品
	GetSaleBillUuidAndSaleOrderUuid(ctx context.Context, deskUuid uint64) (uint64, uint64, error)                                                  // 获取销售账单uuid和销售订单uuid
	OrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error)                                               // 修改购物车商品数量
	OrderCartProductNum(ctx context.Context, req req.OrderCartProductNumReq) (*resp.ShopCart, error)                                               // 修改购物车商品数量
	InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, *resp.OrderCheckServiceRes, error)    // 送厨购物车商品
	InstantOrderCartProductReturning(ctx context.Context, req req.OrderCartProductReturningReq) (*resp.ShopCart, error)                            // 退菜购物车商品
	InstantOrderCartProductCancelReturning(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                  // 退菜购物车商品
	InstantOrderCartProductChangeDesk(ctx context.Context, req req.OrderCartProductChangeDeskReq) (*resp.ShopCart, error)                          // 转菜购物车商品
	InstantOrderCartProductGiving(ctx context.Context, req req.OrderCartProductGivingReq) (*resp.ShopCart, error)                                  // 取消赠菜购物车商品
	InstantOrderCartProductCancelGiving(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                     // 取消赠菜购物车商品
	InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, error)                                           // 获取点餐必点方案
	InstantOrderPaymentInfo(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error)             // 获取结账页面信息
	InstantOrderPaymentQrcode(ctx context.Context, req req.InstantOrderPaymentQrcodeReq) (*resp.InstantOrderPaymentQrcodeInfoResp, error)          // 获取支付二维码
	InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error)                // 给销售订单创建一个支付单
	InstantOrderPaymentCancel(ctx context.Context, req req.InstantOrderPaymentCancelReq) (*resp.InstantOrderPaymentInfoResp, error)                // 撤销一个支付单
	InstantOrderPaymentFinish(ctx context.Context, req req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error)                            // 给销售订单创建一个支付单
	InstantOrderFree(ctx context.Context, req req.InstantOrderFreeReq) (*resp.OrderFinishResp, error)                                              // 免单
	InstantOrderPaymentZeroRule(ctx context.Context, req req.InstantOrderPaymentZeroRuleReq) (*resp.InstantOrderPaymentInfoResp, error)            // 设置结账抹零规则
	InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error)                               // 给销售订单创建一个销售订单
	SaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq, needDeleteSaleOrder bool) (*resp.ShopCart, error)       // 从一个销售订单移动商品到另一个销售订单
	InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq) (bool, error)                                         // 确认必点商品
	OrderCheck(ctx context.Context, req req.InstantOrderCheckReq) (*resp.OrderCheckServiceRes, error)                                              // 订单检查
	InstantOrderSaleOrderDelete(ctx context.Context, req req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error)                               // 删除一个销售订单(删除拆单)
	InstantOrderSaleOrderDeleteAll(ctx context.Context, req req.InstantOrderSaleOrderDeleteAllReq) (*resp.ShopCart, error)                         // 删除所有子销售订单(撤销拆单)
	OrderMemberCancel(ctx context.Context, req req.OrderMemberCancelReq) (*resp.InstantOrderPaymentInfoResp, error)                                // 取消使用会员优惠
	OrderUseMember(ctx context.Context, req req.CheckMemberPasswordReq) (*resp.InstantOrderPaymentInfoResp, error)                                 // 使用会员优惠
	CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error               // 计算并保存销售账单
	OrderPrint(ctx context.Context, req req.OrderPrintReq) (*resp.PrinterData, error)                                                              // 打印
	OrderPrintInvoice(ctx context.Context, req req.OrderPrintInvoiceReq) (*resp.PrinterData, error)                                                // 图片打印
	OrderPrintInvoiceInfo(ctx context.Context, req req.OrderInvoiceInfoReq) resp.SaleOrderInvoiceInfo                                              // 图片打印
	OrderUnlock(ctx context.Context, saleBillUuid uint64) error                                                                                    // 订单解锁
	GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error)                                                    // 必点方案列表
	GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error)   // 获取扫码h5购物车未下单商品列表
	GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error) // 获取扫码h5购物车已下单商品列表
	ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) error                                                           // 下单扫码h5订单
	GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error)             // 未送厨商品列表
	GetSentKitchen(ctx context.Context, saleBillUuid uint64) (resp.SentKitchen, error)                                                             // 已送厨商品列表

	ActionCooking(ctx context.Context, ignoreMust bool, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct) (*resp.OrderCheckServiceRes, error) // 送厨
}

// orderSrv 订单服务结构
type orderSrv struct {
	bus         *event.SystemEventBus
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
		bus:         event.NewSystemBus(),
		dbm:         dbm,
		lock:        lock.NewSystemLock(),
		localeSrv:   localeSrv,
		settingSrv:  settingSrv,
		mustPlanSrv: mustPlanSrv,
	}
}

// HasInstantOrder 判断该收银机是否有未挂单的点餐订单
func HasInstantOrder(ctx context.Context, db *gorm.DB) (*model.SaleBill, bool, error) {
	// 获取设备uuid
	device, err := repository.NewDeviceRepo(db).GetDeviceBySn(ctx.GetDeviceSn())
	if err != nil {
		return nil, false, errors.WithMessage(err, "获取设备uuid失败")
	}

	// 判断是否有待支付、未挂单的订单
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetInstantSaleBill()
	if err != nil && !strings.Contains(err.Error(), "record not found") {
		return nil, false, errors.WithMessage(err, "获取待支付、未挂单的订单失败")
	}
	if saleBill != nil && device.Uuid == saleBill.DeviceUuid {
		return saleBill, true, nil
	}
	return nil, false, nil
}

// CreateInstantOrder 创建点餐订单
func (s *orderSrv) CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error) {
	dbId := ctx.GetDbId()
	var billUuid uint64
	var orderUuid uint64
	db := s.dbm.GetDB(dbId)

	// 判断是否有待支付、未挂单的订单
	_, hasInstantOrder, err := HasInstantOrder(ctx, db)
	if err != nil {
		return resp.CreateInstantOrderResp{}, errors.WithMessage(err)
	}
	if hasInstantOrder {
		return resp.CreateInstantOrderResp{}, errors.New("有待支付、未挂单的订单")
	}
	if err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {

		deviceRepo := repository.NewDeviceRepo(db)
		// 获取设备uuid
		device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
		if err != nil {
			return errors.WithMessage(err, "获取设备uuid失败")
		}

		// todo
		// // 判断是否有待支付、未挂单的订单
		// commonRepo := repository.NewCommonRepo()
		// orderRepo := repository.NewOrderRepo(tx)
		// order, err := orderRepo.GetSaleBill(
		// 	commonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceInstant]),
		// 	commonRepo.WhereByStatus(constant.SaleBillStatusPending),
		// 	commonRepo.WhereByIsHide(false),
		// 	commonRepo.WhereBySoftDelete(),
		// )
		// if err != nil && !strings.Contains(err.Error(), "record not found") {
		// 	return fmt.Errorf("11 GetSaleBill: %s", err)
		// }
		// if order.Uuid > 0 && device.Uuid == order.DeviceUuid {
		// 	return errors.New("有待支付、未挂单的订单")
		// }

		// 创建订单编号
		orderNo, err := s.createOrderNo(tx, constant.OrderSourceInstant)
		if err != nil {
			ctx.Log().Error("订单编号生成失败", zap.Error(err))
			return errors.WithMessage(err, "订单编号生成失败")
		}

		serialNo, err := s.createInstantOrderSerialNo(tx)
		if err != nil {
			ctx.Log().Error("订单序号生成失败", zap.Error(err))
			return errors.WithMessage(err, "订单序号生成失败")
		}
		// 创建销售账单
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			OrderNo:      orderNo,
			SerialNo:     serialNo,
			BillType:     constant.OrderSourceMapToBillType[constant.OrderSourceInstant],
			DiningMethod: constant.SaleBillDiningMethodDineIn,
			DeviceUuid:   device.Uuid,
		})
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售账单设置
		saleBillSetting, err := s.CreateSaleBillSetting(ctx, tx, dbId, saleBill.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售订单
		saleOrder, errCreateSaleOrder := createSaleOrder(tx, &saleBillSetting, saleBill.Uuid, orderNo)
		if errCreateSaleOrder != nil {
			return errCreateSaleOrder
		}

		billUuid = saleBill.Uuid
		orderUuid = saleOrder.Uuid

		return nil
	}); err != nil {
		return resp.CreateInstantOrderResp{}, errors.WithMessage(err)
	}

	return resp.CreateInstantOrderResp{
		SaleBillUuid:  billUuid,
		SaleOrderUuid: orderUuid,
	}, nil
}

func (s *orderSrv) createInstantOrderSerialNo(db *gorm.DB) (string, error) {
	var serialNo string
	saleBillRepo := repository.NewSaleBillRepo(db)
	saleBill, err := saleBillRepo.GetInstantSaleBillLatest()
	if err != nil {
		return "", errors.WithMessage(err)
	}
	// 如果没有查询到账单，则设置为0001
	if saleBill == nil {
		serialNo = "0001"
		return serialNo, nil
	}
	createTime := saleBill.CreateTime
	// 判断账单的创建时间是不是今天
	if !IsToday(createTime) {
		serialNo = "0001"
	} else {
		oldSerialNo := saleBill.SerialNo
		if oldSerialNo == "" {
			// 如果serialNo为空，则设置为0001. 兼容老数据没有serialNo的情况
			serialNo = "0001"
		} else {
			serialNoNum, err := strconv.Atoi(oldSerialNo)
			if err != nil {
				return "", errors.WithMessage(err)
			}
			serialNo = strconv.Itoa(serialNoNum + 1)
		}
		if len(serialNo) < 4 {
			serialNo = strings.Repeat("0", 4-len(serialNo)) + serialNo
		}
	}
	return serialNo, nil
}

func IsToday(timestamp int64) bool {
	return time.Unix(timestamp, 0).Format("20060102") == time.Now().Format("20060102")
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
		return nil, errors.WithMessage(err)
	}
	return &saleOrder, nil
}

// CreateSaleBillSetting 创建销售账单设置
func (s *orderSrv) newSaleBillSetting(ctx context.Context, saleBillUuid uint64) (*model.SaleBillSetting, error) {
	// 获取服务费设置
	serviceFeeSetting, err := s.settingSrv.GetServiceFeeSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取税率设置
	taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
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
		serviceFeeValue, err = utils.ParseFloat(serviceFeeSetting.ServiceCharge)
		if err != nil {
			return nil, errors.WithMessage(err)
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

	saleBillSetting := model.SaleBillSetting{
		SaleBillUuid:     saleBillUuid,
		ServiceFeeType:   serviceFeeType,
		ServiceFeeValue:  serviceFeeValue,
		TaxFeeType:       taxFeeType,
		DiscountType:     discountType,
		ZeroRule:         zero,
		ZeroCheckoutRule: zeroCheckout,
		IsStatGift:       isStatGift,
		IsStatFree:       isStatFree,
	}

	return &saleBillSetting, nil
}

// CreateSaleBillSetting 创建销售账单设置
func (s *orderSrv) CreateSaleBillSetting(ctx context.Context, db *gorm.DB, dbId uint64, saleBillUuid uint64) (model.SaleBillSetting, error) {
	// 获取服务费设置
	serviceFeeSetting, err := s.settingSrv.GetServiceFeeSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, errors.WithMessage(err)
	}
	// 获取税率设置
	taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, errors.WithMessage(err)
	}
	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return model.SaleBillSetting{}, errors.WithMessage(err)
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
		serviceFeeValue, err = utils.ParseFloat(serviceFeeSetting.ServiceCharge)
		if err != nil {
			return model.SaleBillSetting{}, errors.WithMessage(err)
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
		SaleBillUuid:     saleBillUuid,
		ServiceFeeType:   serviceFeeType,
		ServiceFeeValue:  serviceFeeValue,
		TaxFeeType:       taxFeeType,
		DiscountType:     discountType,
		ZeroRule:         zero,
		ZeroCheckoutRule: zeroCheckout,
		IsStatGift:       isStatGift,
		IsStatFree:       isStatFree,
	})

	return saleBillSetting, errors.WithMessage(err)
}

// CreateDeskOrder 创建桌台订单
func (s *orderSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.DeskUuid)
		defer lock.NewSystemLock().UnlockUuid(req.DeskUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	desk, err := repository.NewDeskRepo(db).GetDeskRecord(req.DeskUuid)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "无法找到空闲桌台")
	}
	if !desk.IsAvailableDesk() {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "该桌台非空闲桌台")
	}
	saleBillUuid, _ := utils.GetID()
	desk.SetOpenDesk(saleBillUuid)

	// 创建订单编号
	orderNo, err := s.createOrderNo(db, constant.OrderSourceDesk)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "订单编号生成失败")
	}

	// 构建销售账单
	saleBill := model.NewDeskSaleBill(saleBillUuid, orderNo, req.BuffetUuids, req.GetMealNum(), req.Remark, req.DeskUuid, desk.DeskNo)

	// 构建销售账单设置
	saleBillSetting, err := s.newSaleBillSetting(ctx, saleBill.Uuid)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}
	// 构建销售订单
	saleOrder := model.NewSaleOrder(saleBill.Uuid, saleBill.OrderNo, *saleBillSetting)

	// 获取自助餐信息
	buffetList, err := repository.NewBuffetRepo(db).GetBuffetListByUuids(req.BuffetUuids)
	if err != nil {
		return resp.CreateDeskOrderResp{}, nil
	}

	// 构建自助餐顾客列表
	buffetCustomerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&buffetCustomerTypes, req.BuffetCustomerTypes)
	saleOrderBuffetCustomerTypes, _, _, maxTimeLimit := saleOrder.GetSaleOrderBuffetCustomerTypes(buffetList, req.BuffetUuids, buffetCustomerTypes, saleBillSetting)

	// 开始事务
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 如果是自助餐，有顾客列表的话，创建顾客
		if len(saleOrderBuffetCustomerTypes) > 0 {
			for _, customer := range saleOrderBuffetCustomerTypes {
				if _, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetCustomerType(*customer); err != nil {
					return errors.WithMessage(err)
				}
			}
			if maxTimeLimit == -1 {
				saleBill.BuffetDuration = 0
			} else {
				saleBill.BuffetDuration = uint(maxTimeLimit)
			}
		}

		// 创建销售账单
		if _, errCreateSaleBill := repository.NewOrderRepo(tx).CreateSaleBill(*saleBill); errCreateSaleBill != nil {
			return errCreateSaleBill
		}

		// 创建销售账单设置
		if _, errCreateSaleBillSetting := repository.NewOrderRepo(db).CreateSaleBillSetting(*saleBillSetting); errCreateSaleBillSetting != nil {
			return errCreateSaleBillSetting
		}

		// 创建销售订单
		if _, errCreateSaleOrder := repository.NewOrderRepo(tx).CreateSaleOrder(*saleOrder); errCreateSaleOrder != nil {
			return errCreateSaleOrder
		}

		// 新桌台的状态
		if errUpdate := repository.NewDeskRepo(tx).UpdateDesk(req.DeskUuid, *desk); errUpdate != nil {
			return errUpdate
		}

		// 提交完订单后，重新查询并计算金额。 todo 改为sale_bill保存数据库前计算好金额。
		{
			// 获取销售账单信息
			saleBill, err := repository.NewOrderRepo(tx).GetSaleBillAllInfo(saleBill.Uuid)
			if err != nil {
				return errors.WithMessage(err)
			}
			// 计算销售账单金额
			if err = s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}

	return resp.CreateDeskOrderResp{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: saleOrder.Uuid,
	}, nil
}

// createOrderNo 创建订单编号
func (s *orderSrv) createOrderNo(db *gorm.DB, orderSource string) (string, error) {
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

		if !utils.IsNotFoundRecord(err) {
			orderNo = ""
			break
		} else {
			break
		}
	}
	if orderNo == "" {
		return "", errors.New("订单编号生成失败")
	}
	return orderNo, nil
}

// GetCashierOrderList 获取订单列表
func (s *orderSrv) GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	lists, total, dbOption, err := orderRepo.GetCashierOrderListWithPagination(reqs)
	if err != nil {
		return resp.OrderListPaginationResp{}, errors.WithMessage(err)
	}

	// 组合列表源数据
	billList := make([]resp.BillLists, len(lists))
	consumerUuids := []string{}
	for i, bill := range lists {
		totalPayTypeNames := []string{}
		isSplit := len(bill.SaleOrders) > 1 // 拆单
		orderList := make([]resp.BillListsOrder, 0)
		var paymentAmounts float64
		//
		billListsExtra := resp.BillListsExtra{
			IsCellRefund:        false,
			IsCellCancel:        bill.Status == constant.SaleBillStatusPending,
			IsCellReverseSettle: bill.IsCellReverseSettle(),
			IsCellPrint:         !isSplit && bill.Status != constant.SaleBillStatusPending,
			IsCellInvoice:       !isSplit && bill.Status == constant.SaleBillStatusComplete,
			IsCellDelete:        bill.Status == constant.SaleBillStatusCanceled,
		}
		// 拆单
		if isSplit {
			for k, order := range bill.SaleOrders {
				if order.IsDelete() {
					continue
				}
				// 获取支付方式
				payTypeNames := []string{}
				if order.IsFree == 1 {
					totalPayTypeNames = append(totalPayTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
					payTypeNames = append(payTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
				} else {
					for _, payment := range order.PaymentOrders {
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
						payTypeNames = append(payTypeNames, payment.PaymentMethodName)
					}
				}

				orderExtra := resp.BillListsExtra{
					IsCellRefund:        false,
					IsCellCancel:        false,
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
				if order.Status == constant.SaleBillStatusComplete && ctx.GetStaff().Uuid == bill.CashierUuid && order.FinishTime > ctx.GetStaff().CashierLoginTime {
					orderExtra.IsCellReverseSettle = true
				}
				//
				paymentAmount := order.GetActualPaymentAmount()
				paymentAmounts += paymentAmount
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
					PaymentAmount: paymentAmount,
					PayTypeName:   strings.Join(utils.RemoveDuplicates(payTypeNames), ","),
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
				if order.IsFree == 1 {
					totalPayTypeNames = append(totalPayTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
				} else {
					for _, payment := range order.PaymentOrders {
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
					}
				}
				// 不等于免单 && 未退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					billListsExtra.IsCellRefund = true
				}
				// 等于主单 && 完成 && 等于当前用户 && 在班次时间内
				if order.Status == constant.SaleBillStatusComplete && ctx.GetStaff().Uuid == bill.CashierUuid && order.FinishTime > ctx.GetStaff().CashierLoginTime {
					billListsExtra.IsCellReverseSettle = true
				}
			}
		}
		//
		saleOrderUuid := uint64(0)
		if !isSplit && len(bill.SaleOrders) > 0 {
			saleOrderUuid = bill.SaleOrders[0].Uuid
		}
		//
		billList[i] = resp.BillLists{
			SaleBillUuid:  bill.Uuid,
			SaleOrderUuid: saleOrderUuid,
			BillType:      bill.BillType,
			IsSplit:       len(bill.SaleOrders) > 1,
			SerialNo:      bill.SerialNo,
			OrderNo:       bill.OrderNo,
			Status:        bill.Status,
			FinishTime:    bill.FinishTime,
			OrderAmount:   bill.Amount,
			PaymentAmount: bill.GetPaymentAmount(),
			ConsumerUuids: strings.Join(consumerUuids, ","),
			PayTypeName:   strings.Join(utils.RemoveDuplicates(totalPayTypeNames), ","),
			SaleOrders:    orderList,
			Extra:         billListsExtra,
		}
	}
	// 获取数量
	getOrderNum := func(status uint) int64 {
		num, _ := orderRepo.GetOrderNum(
			repository.CommonRepo.WhereByStatus(status),
			repository.CommonRepo.WhereBySoftDelete(),
			repository.CommonRepo.WhereByCooking(),
			dbOption,
		)
		return num
	}
	// 返回响应对象
	return resp.OrderListPaginationResp{
		List: billList,
		Meta: resp.OrderListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			UnpaidNum:   getOrderNum(constant.SaleBillStatusPending),
			CompleteNum: getOrderNum(constant.SaleBillStatusComplete),
			CancelNum:   getOrderNum(constant.SaleBillStatusCanceled),
		},
	}, nil
}

// GetOrderInfos 获取收银端订单信息
func (s *orderSrv) GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取信息源
	saleBill, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, 0)
	if err != nil {
		return resp.OrderInfosResp{}, errors.WithMessage(err)
	}
	isMain := req.SaleOrderUuid == 0        // 是否查询主单
	isSplit := len(saleBill.SaleOrders) > 1 // 是否拆单
	isCellCancel := isMain

	// 组合信息
	totalMemberNames := []string{}
	totalMemberUuids := []string{}
	orderList := make([]resp.OrderInfo, 0)
	for i, saleOrder := range saleBill.SaleOrders {
		if req.SaleOrderUuid > 0 && req.SaleOrderUuid != saleOrder.Uuid {
			continue
		}
		if saleOrder.GetMemberName() != "" && !slices.Contains(totalMemberNames, saleOrder.GetMemberName()) {
			totalMemberNames = append(totalMemberNames, saleOrder.GetMemberName())
		}
		if saleOrder.ConsumerUuid != 0 {
			totalMemberUuids = append(totalMemberUuids, strconv.FormatUint(saleOrder.ConsumerUuid, 10))
		}
		//
		products := make([]resp.OrderProduct, 0)

		// 添加自助餐顾客
		{
			for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if orderBuffetCustomer.IsDelete() {
					continue
				}
				// 自助餐顾客价格收费列表
				products = append(products, resp.OrderProduct{
					Uuid:       orderBuffetCustomer.Uuid,
					LocaleName: orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNames(),
					LocaleAttributeName: dto.LocaleResponse{
						ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					},
					Price:            orderBuffetCustomer.Price,
					Num:              orderBuffetCustomer.Num, // 这种类型顾客多少个，如老人这个类型2人
					SalePrice:        orderBuffetCustomer.GetOriginPrice(),
					TotalPrice:       orderBuffetCustomer.GetDiscountPrice(),
					Status:           1,
					Remark:           "",
					IsMust:           false,
					IsGift:           false,
					IsBuffetCustomer: true,
				})
			}
		}

		// 添加加钟商品
		{
			for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if delayProduct.IsDelete() {
					continue
				}
				products = append(products, resp.OrderProduct{
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
					Num:                 delayProduct.Num, // 拆单后不等于桌台人数，但同一个加钟商品的总数等于桌台人数
					Price:               delayProduct.Price,
					SalePrice:           delayProduct.GetAmount(),
					TotalPrice:          delayProduct.GetAmount(),
					Status:              1,  // 添加后标记送厨状态，不可修改
					Remark:              "", // 加钟商品没有备注
					IsMust:              false,
					IsGift:              false,
					IsBuffet:            false,
					IsDelay:             true,
				})
			}
		}

		// 添加正常商品
		{
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() {
					continue
				}
				imageUrl := ""
				if saleOrderProduct.ImageFile != nil {
					imageUrl = saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
				}
				products = append(products, resp.OrderProduct{
					Uuid:                saleOrderProduct.Uuid,
					LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: saleOrderProduct.GetAttributeName(),
					Price:               saleOrderProduct.Price,
					Num:                 saleOrderProduct.Num,
					SalePrice:           saleOrderProduct.GetSalePrice(),
					TotalPrice:          saleOrderProduct.GetTotalPrice(),
					Status:              saleOrderProduct.Status,
					Remark:              saleOrderProduct.Remark,
					IsMust:              saleOrderProduct.IsMustProduct(),
					IsGift:              saleOrderProduct.IsGiftProduct(),
					IsBuffet:            saleOrderProduct.IsBuffetProduct(),
					ImageUrl:            imageUrl,
					CancelReason:        saleOrderProduct.CancelReason,
					GiftReason:          saleOrderProduct.GiftReason,
					RefundAmount:        saleOrderProduct.GetReturnPrice(),
				})
			}
		}

		//
		orderList = append(orderList, resp.OrderInfo{
			SaleOrderUuid: saleOrder.Uuid,
			BillType:      saleBill.BillType,
			DiningMethod:  saleBill.DiningMethod,
			SerialNo:      saleBill.SerialNo + "-" + strconv.Itoa(i+1),
			OrderNo:       saleOrder.OrderNo,
			Status:        saleOrder.Status,
			IsFree:        saleOrder.IsFree == 1,
			FreeReason:    saleOrder.FreeReason,
			OrderAmount:   saleOrder.Amount,
			PaymentAmount: saleOrder.GetActualPaymentAmount(),
			RefundAmount:  saleOrder.GetTotalRefundAmount(),
			PayTypeName:   saleOrder.GetPayTypeNames(ctx.GetLanguage()),
			MemberName:    saleOrder.GetMemberName(),
			MemberUuid:    saleOrder.ConsumerUuid,
			Products:      products,
		})
		//
		if saleOrder.Status != constant.SaleBillStatusPending {
			isCellCancel = false
		}
	}

	// 处理额外信息
	var order *model.SaleOrder
	if len(saleBill.SaleOrders) > 0 {
		order = saleBill.SaleOrders[0]
	}
	orderExtra := resp.BillListsExtra{
		IsCellRefund:        false,
		IsCellCancel:        isCellCancel,
		IsCellReverseSettle: saleBill.IsCellReverseSettle(),
		IsCellPrint:         true,
		IsCellDelete:        order.Status == constant.SaleBillStatusCanceled,
		IsCellInvoice:       false,
	}
	if (!isSplit || !isMain) && order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
		orderExtra.IsCellRefund = true
	}

	// 返回响应对象
	return resp.OrderInfosResp{
		Detail: resp.OrderInfos{
			SaleBillUuid: saleBill.Uuid,
			IsSplit:      isSplit,
			BillType:     saleBill.BillType,
			DiningMethod: saleBill.DiningMethod,
			SerialNo:     saleBill.SerialNo,
			OrderNo: func() string {
				if isMain {
					return saleBill.OrderNo
				}
				return order.OrderNo
			}(),
			Status:        saleBill.Status,
			CreateTime:    saleBill.CreateTime,
			FinishTime:    saleBill.FinishTime,
			OrderAmount:   saleBill.Amount,
			PaymentAmount: saleBill.GetPaymentAmount(),
			RefundAmount:  saleBill.GetTotalRefundAmount(),
			MemberNames:   strings.Join(totalMemberNames, ","),
			MemberUuids:   strings.Join(totalMemberUuids, ","),
			CashierName:   saleBill.Cashier.RealName,
			IsBuffet:      saleBill.IsBuffet == 1,
			BuffetNames:   saleBill.GetBuffetNames(ctx.GetLanguage()),
			CancelReason:  saleBill.Reason,
			PayTypes:      saleBill.GetPayTypes(ctx.GetLanguage(), req.SaleOrderUuid),
			SaleOrders:    orderList,
			Remark:        saleBill.Remark,
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
		return []resp.OrderOperationLog{}, errors.WithMessage(err)
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
		return model.SaleBill{}, errors.WithMessage(err)
	}
	if err := billInfo.ValidateOrderStatus(constant.OrderOrderCancel); err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
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
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取订单信息
	billInfo, err := s.IsCellCancelOrder(ctx, req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	// 验证高级密码
	if s.settingSrv == nil {
		return errors.New("找不到 settingSrv")
	}
	// 如果不需要验证高级密码，则跳过
	if !req.NotNeedPassword {
		if err := s.settingSrv.VerifyAdvancedPassword(ctx, req.Password); err != nil {
			return errors.WithMessage(err)
		}
	}

	// 获取信息源
	db := s.dbm.GetDB(dbId)

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	orderRepo := repository.NewOrderRepo(tx)
	deskRepo := repository.NewDeskRepo(tx)
	qrOrderRepo := repository.NewH5OrderRepo(tx)
	orderRecordRepo := repository.NewOrderOperationRecordRepo(tx)

	// 退回商品库存
	{
		// 获取销售账单信息
		saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		s.reverseSettleWarehouseForm(ctx, saleBill)
		var warehouseForm *model.WarehouseForm
		{
			products := saleBill.GetSaleOrderProductCooking()
			productList, err := s.getDecreaseStockList(ctx, products)
			if err != nil {
				return errors.WithMessage(err)
			}
			warehouseForm = model.NewWarehouseForm(productList, saleBill.Uuid)
		}
		// 创建入库单
		if warehouseForm != nil && len(warehouseForm.WarehouseFormItems) > 0 {
			// 创建入库单
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseFormRecord(*warehouseForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建入库单记录
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseFormItemRecords(warehouseForm.WarehouseFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 发布"加库存"事件
		go func() {
			event2.AddStock(db, saleBill.Uuid)
		}()
	}

	// 如果是桌台订单
	if billInfo.BillType == 0 && billInfo.DeskUuid > 0 {
		// 拒绝所有待接单 - todo 待对应的服务层实现
		err := qrOrderRepo.Reject(billInfo.DeskUuid)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
		// 关闭桌台
		err = deskRepo.CloseDesk(billInfo.DeskUuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	} else {
		err = orderRepo.CancelOrder(req.SaleBillUuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
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
		return errors.WithMessage(err)
	}

	return nil
}

// DeleteOrder 删除订单, saleOrderUuid = 等于0的时候删除主单，并且主单下的子单也会被删除， saleOrderUuid > 0 的时候删除子单
func (s *orderSrv) DeleteOrder(ctx context.Context, dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, constant.OptionalUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	if billInfo.Status != constant.SaleBillStatusCanceled {
		return errors.New("订单状态不允许删除")
	}

	err = orderRepo.DeleteOrder(saleBillUuid, saleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 检查是否有已送厨的商品，如果有，则标记production_order_product.status为消单退菜（制作中消单退菜、制作完成消单退菜）
	// 如果已送厨商品还在制作中，通知厨房取消制作
	doingProductList := make([]uint64, 0) // 制作中的商品uuid列表 sale_order_product_uuid
	// todo 获取还在制作中的商品

	// 发布事件，通知厨房取消制作
	event.NewSystemBus().PublishCancelDoingProductEvent(event.CancelDoingProductPayload{SaleOrderProductUuids: doingProductList})
	return nil
}

// ReturnOrder 退款订单
func (s *orderSrv) ReturnOrder(ctx context.Context, req req.OrderReturnReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return errors.WithMessage(errors.New("找不到销售订单"))
	}

	returnType := constant.ReturnOrderRefundTypeTotal
	saleOrderProducts := make([]*model.SaleOrderProduct, 0) // 退款商品列表
	numMap := make(map[uint64]uint)                         // 每个退款商品的退货数量
	// 整单退款
	if len(req.Products) == 0 {
		returnType = constant.ReturnOrderRefundTypeTotal
		// 整单退款，退款商品列表为销售订单商品列表.
		// 注意：要判断订单商品是否还有可退货数量
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = canReturnNum
			}
		}
	}
	// 部分退款
	if len(req.Products) > 0 {
		returnType = constant.ReturnOrderRefundTypePart
		// 获取退款商品列表
		saleOrderProductUuids := make([]uint64, 0)
		for _, product := range req.Products {
			saleOrderProductUuids = append(saleOrderProductUuids, product.SaleOrderProductUuid)
			numMap[product.SaleOrderProductUuid] = uint(product.Num)
		}
		// 注意：要判断订单商品是否还有可退货数量
		saleOrderProductList := saleOrder.GetSaleOrderProductList(saleOrderProductUuids)
		if len(saleOrderProductList) == 0 {
			return errors.WithMessage(errors.New("找不到退货商品"))
		}
		for _, saleOrderProduct := range saleOrderProductList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if num <= canReturnNum {
					saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("退货数量超过可退货数量"))
				}
			}
		}
	}

	// 如果退款类型为部分退款，则必须有可退货的商品。整单退款则可以没有可退货的商品，可能已经退完商品了但还有手续费没有退
	if len(saleOrderProducts) == 0 && returnType == constant.ReturnOrderRefundTypePart {
		return errors.WithMessage(errors.New("没有可退货的商品"))
	}

	// 创建退款单
	returnOrder := saleOrder.NewReturnOrder(saleOrderProducts, numMap, returnType)

	err = repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建退货单
		if _, err = repository.NewReturnOrderRepo(db).CreateReturnOrderRecord(*returnOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 创建退款金额
		if err = repository.NewReturnOrderRepo(db).CreateReturnOrderAmount(returnOrder.ReturnOrderAmounts); err != nil {
			return errors.WithMessage(err)
		}
		// 创建退货单商品
		if err = repository.NewReturnOrderRepo(db).CreateReturnOrderProduct(returnOrder.ReturnOrderProducts); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
	}
	// 发布退款事件 todo
	// 退款到余额或现金 todo
	return nil
}

// GetReturnOrderInfo 获取退款信息
func (s *orderSrv) GetReturnOrderInfo(ctx context.Context, req req.OrderReturnInfoReq) (*resp.OrderReturnInfoResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("找不到销售订单")
	}

	// 判断订单是否可以退款
	// if !saleBill.IsReturnable() {
	// 	return nil, errors.New("订单状态不允许退款")
	// }

	// 获取销售订单付款单列表

	// 获取销售订单退货单列表

	// 获取销售订单商品列表

	products := make([]resp.OrderReturnProduct, 0)

	// 获取销售订单的每个付款单的可退款金额
	// 要求排好序：退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	paymentRecords, currencyUnit := saleOrder.GetPaymentOrderCanReturnAmount()

	// 获取销售订单商品列表
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(),
			Num:                  saleOrderProduct.GetCanReturnNum(), // 可退货数量=订单商品数量-已退货数量
			Price:                saleOrderProduct.Price,
			CanReturnAmount:      saleOrderProduct.GetCanReturnPrice(),
			CurrencyUnit:         currencyUnit,
		})
	}

	// 获取销售订单付款单列表
	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmount()
	res := &resp.OrderReturnInfoResp{
		CanReturnAmount: canReturnAmount, // 可退款金额. 可退款金额=订单最终应收金额-已退款金额
		PaymentRecords:  paymentRecords,
		Products:        products,
	}

	return res, nil
}

// GetReverseSettleInfo 获取反结账信息
func (s *orderSrv) GetReverseSettleInfo(ctx context.Context, req req.OrderReverseSettleInfoReq) (*resp.OrderReverseSettleInfoResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	var resDesks *resp.OrderReverseSettleDeskList
	if saleBill.IsDeskSaleBill() {
		desk := saleBill.Desk
		desks := make([]resp.OrderReverseSettleDesk, 0)
		// 如果原桌台空闲
		if desk.IsAvailableDesk() {
			desks = append(desks, resp.OrderReverseSettleDesk{
				Uuid:     desk.Uuid,
				SerialNo: desk.DeskNo,
			})
		}
		// 如果原桌台不空闲
		if !desk.IsAvailableDesk() {
			// 获取所有空闲的桌台
			freeDesks, err := repository.NewDeskRepo(db).GetAvailableDeskList()
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			for _, freeDesk := range freeDesks {
				desks = append(desks, resp.OrderReverseSettleDesk{
					Uuid:     freeDesk.Uuid,
					SerialNo: freeDesk.DeskNo,
				})
			}
		}
		resDesks = &resp.OrderReverseSettleDeskList{
			OriginDeskAvailable: desk.IsAvailableDesk(),
			List:                desks,
		}
	}

	var hasInstantOrder *bool
	if !saleBill.IsDeskSaleBill() {
		// 判断该收银机是否有未挂单的点餐订单
		_, hasInstantOrderBool, err := HasInstantOrder(ctx, db)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		hasInstantOrder = &hasInstantOrderBool
	}

	// 获取支付方式名称列表
	payMethods := saleBill.GetPaymentMethodNameList()

	return &resp.OrderReverseSettleInfoResp{
		SaleBillUuid:    saleBill.Uuid,
		SaleBillNo:      saleBill.OrderNo,
		SaleBillType:    saleBill.BillType,
		OrderAmount:     saleBill.Amount,
		PaymentAmount:   saleBill.PaymentAmount,
		PayMethods:      payMethods,
		Desks:           resDesks,
		HasInstantOrder: hasInstantOrder,
	}, nil
}

// ReverseSettle 处理反结账
func (s *orderSrv) ReverseSettle(ctx context.Context, req req.OrderReverseSettleReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if saleBill.IsDeskSaleBill() {
		if req.DeskUuid == 0 {
			return errors.WithMessage(errors.New("桌台UUID不能为0"))
		}
	}

	// 销售账单状态变为未结账状态
	// 销售订单状态变为未结账状态
	// 销售订单的所有付款单都退款，并生成退款单
	saleBill.SetReverseSettle()

	// 如果销售账单是桌台订单，则开桌
	// 开桌
	var desk *model.Desk
	if saleBill.IsDeskSaleBill() {
		deskRepo := repository.NewDeskRepo(db)
		desk, err = deskRepo.GetDeskRecord(req.DeskUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if !desk.IsAvailableDesk() {
			return errors.WithMessage(errors.New("桌台非空闲"))
		}
		desk.SetOpenDesk(saleBill.Uuid)
	}

	// 如果销售账单是点餐订单，则如果存在未挂单的点餐订单，根据参数决定是否挂单
	var hideSaleBill *model.SaleBill
	if !saleBill.IsDeskSaleBill() {
		saleBill, hasInstantOrder, err := HasInstantOrder(ctx, db)
		if err != nil {
			return errors.WithMessage(err)
		}
		// 如果存在未挂单的点餐订单，则根据参数决定是否挂单
		if hasInstantOrder {
			if req.HideOrder {
				hideSaleBill = saleBill
				hideSaleBill.SetHideSaleBill()
			} else {
				// 如果存在未挂单的点餐订单
				return errors.WithMessage(errors.New("存在未挂单的点餐订单，请先挂单"))
			}
		}
	}

	// 构建入库单，将账单的商品重新入库.
	// 出库记录标记为已撤销，并生成入库单将库存退还
	// 构建出库单，将账单下单减库存的商品出库
	if err := s.reverseSettleWarehouseForm(ctx, saleBill); err != nil {
		return errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 如果存在需要挂单的销售账单，则更新该销售账单
		if hideSaleBill != nil {
			if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*hideSaleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新销售账单
		if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售订单
		for _, saleOrder := range saleBill.SaleOrders {
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 退款
		for _, saleOrder := range saleBill.SaleOrders {
			for _, paymentOrder := range saleOrder.PaymentOrders {
				if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*paymentOrder); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 更新桌台
		if saleBill.IsDeskSaleBill() {
			if err := repository.NewDeskRepo(db).UpdateDeskRecord(*desk); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *orderSrv) reverseSettleWarehouseForm(ctx context.Context, saleBill *model.SaleBill) error {
	db := ctx.GetDB()
	// 构建入库单，将账单的商品重新入库.
	// 出库记录标记为已撤销，并生成入库单将库存退还

	// 1.该账单的出库单撤销
	var forms []*model.WarehouseOutForm
	{
		// 获取该账单的所有出库单
		warehouseOutForms, err := repository.NewWarehouseFormRepo(db).GetWarehouseOutFormsBySaleBillUuid(saleBill.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		for _, form := range warehouseOutForms {
			form.RevokeForm()
		}
		forms = warehouseOutForms
	}

	// 2.该账单的商品归还库存，生成入库单
	// 获取账单的所有商品，退菜的商品除外
	var warehouseForm *model.WarehouseForm
	{
		products := saleBill.GetSaleOrderProductCooking()
		productList, err := s.getDecreaseStockList(ctx, products)
		if err != nil {
			return errors.WithMessage(err)
		}
		warehouseForm = model.NewWarehouseForm(productList, saleBill.Uuid)
	}

	// 3.构建出库单，将账单下单减库存的商品出库
	var warehouseOutForm *model.WarehouseOutForm
	{
		products := saleBill.GetSaleOrderProductCookingSubStock()
		productList, err := s.getDecreaseStockList(ctx, products)
		if err != nil {
			return errors.WithMessage(err)
		}
		warehouseOutForm = model.NewWarehouseOutForm(productList, false, saleBill.Uuid)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 撤销出库单
		for _, form := range forms {
			if err := repository.NewWarehouseFormRepo(db).UpdateWarehouseOutFormRecord(*form); err != nil {
				return errors.WithMessage(err)
			}
			// 撤销出库单明细
			for _, item := range form.WarehouseOutFormItems {
				if err := repository.NewWarehouseFormRepo(db).UpdateWarehouseOutFormItemRecord(*item); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 创建入库单
		if warehouseForm != nil && len(warehouseForm.WarehouseFormItems) > 0 {
			// 创建入库单
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseFormRecord(*warehouseForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建入库单记录
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseFormItemRecords(warehouseForm.WarehouseFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}
		fmt.Println("8888888888:")

		// 如果出库单明细不为空，则创建出库单
		if warehouseOutForm != nil && len(warehouseOutForm.WarehouseOutFormItems) > 0 {
			// 创建出库单
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseOutFormRecord(*warehouseOutForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建出库单记录
			if err := repository.NewWarehouseFormRepo(db).CreateWarehouseOutFormItemRecords(warehouseOutForm.WarehouseOutFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	fmt.Println("9999999999:")
	// 发布"加库存"事件
	go func() {
		event2.AddStock(db, saleBill.Uuid)
	}()
	fmt.Println("10000000000:")
	// 发布"减库存"事件
	go func() {
		event2.ReduceStock(db, saleBill.Uuid)
	}()

	fmt.Println("1111111221111:")
	return nil
}

// HideOrder 隐藏订单（挂单）
func (s *orderSrv) HideOrder(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillRecord(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return nil, errors.New("找不到订单")
	}
	if billInfo.Status != constant.SaleBillStatusPending {
		return nil, errors.New("订单状态不允许挂单")
	}

	// 隐藏
	err = orderRepo.HideOrder(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"挂单"事件
	go func() {
		event.NewSystemBus().PublishHideSaleBillEvent(event.HideSaleBillPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		return &resp.ShopCart{SaleOrderList: make([]resp.SaleOrder, 0)}, nil
	}

	return info, nil
}

// ShowOrder 显示订单
func (s *orderSrv) ShowOrder(ctx context.Context, req req.OrderShowReq) (*resp.ShopCart, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	ctx.Log().Debug("取单", zap.Any("request", req))
	db := s.dbm.GetDB(ctx.GetDbId())

	// 判断是否有未挂单的点餐账单
	hasShowOrder, err := repository.NewOrderRepo(db).HasShowOrder(ctx.GetDeviceUuid())
	if err != nil {
		ctx.Log().Error("判断是否有未挂单的点餐账单失败", zap.Error(err))
		return nil, errors.WithMessage(err, "判断是否有未挂单的点餐账单失败")
	}
	if hasShowOrder {
		return nil, errors.New("该设备有未挂单的点餐账单，禁止取单")
	}

	saleBill, err := repository.NewOrderRepo(db).GetSaleBillRecord(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	if saleBill.IsShowSaleBill() {
		return nil, errors.New("该账单已取出")
	}

	// 修改销售账单信息，标记账单取出
	saleBill.SetShowSaleBill(ctx.GetDeviceUuid())
	// 更新销售账单
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
		return nil, errors.WithMessage(err, "更新销售账单失败", fmt.Sprintf("NewSaleBillRepo(db).UpdateSaleBillRecord failed, sale_bill uuid:%d", saleBill.Uuid))
	}

	// 发布"取单"事件
	go func() {
		event.NewSystemBus().PublishShowSaleBillEvent(event.ShowSaleBillPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: req.SaleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return info, nil
}

// InstantHideOrderList 获取挂单订单列表
func (s *orderSrv) InstantHideOrderList(ctx context.Context, req req.HideSaleBillListReq) (*resp.InstantHideOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBillRepo := repository.NewSaleBillRepo(db)

	// 查询所有已挂单的点餐销售账单
	saleBills, total, err := saleBillRepo.GetHideSaleBillList(req.PageNo, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	hideSaleBills := make([]resp.InstantHideSaleBill, 0)
	for _, saleBill := range saleBills {
		if saleBill.IsShowSaleBill() || saleBill.IsDeskSaleBill() || saleBill.IsDelete() {
			continue
		}
		listMap := make(map[string]resp.Product) // 商品列表，key为商品sign
		for _, saleOrder := range saleBill.SaleOrders {
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() || saleOrderProduct.Num == 0 {
					continue
				}
				if product, ok := listMap[saleOrderProduct.Sign]; !ok {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromUint64(uint64(saleOrderProduct.Num))).InexactFloat64()
					newProduct := resp.Product{
						LocaleName:    saleOrderProduct.MultiLanguageName.GetNames(),
						Num:           saleOrderProduct.Num,
						SalePrice:     productPrice,
						DiscountPrice: productPrice,
					}
					listMap[saleOrderProduct.Sign] = newProduct
				} else {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromUint64(uint64(saleOrderProduct.Num)))
					price := productPrice.Add(decimal.NewFromFloat(product.SalePrice)).InexactFloat64()
					product.Num += saleOrderProduct.Num
					product.SalePrice = price
					product.DiscountPrice = price
				}
			}
		}
		list := make([]resp.Product, 0)
		for sign, _ := range listMap {
			list = append(list, listMap[sign])
		}
		productList := resp.InstantHideSaleProductList{List: list}
		hideSaleBill := resp.InstantHideSaleBill{
			SaleBillUuid: saleBill.Uuid,
			SerialNo:     saleBill.SerialNo,
			Amount:       saleBill.Amount,
			HideBillTime: saleBill.HideBillTime,
			Products:     productList,
		}
		hideSaleBills = append(hideSaleBills, hideSaleBill)
	}

	res := &resp.InstantHideOrderListResp{
		List: hideSaleBills,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}
	return res, nil
}

// OrderTakeout 打包
func (s *orderSrv) OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(constant.OrderTakeout, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 修改销售账单状态
	if req.Takeout {
		saleBill.SetTakeoutSaleBill(constant.SaleBillDiningMethodTakeout)
	} else {
		saleBill.SetTakeoutSaleBill(constant.SaleBillDiningMethodDineIn)
	}

	saleBill.CalcAll()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return info, nil
}

// OrderProductDelete 删除订单商品
func (s *orderSrv) OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	return s.orderProductDelete(ctx, dbId, staffUuid, source, req)
}

func (s *orderSrv) orderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) {

	// 获取信息源
	db := s.dbm.GetDB(dbId)

	// 获取操作的销售账单信息
	saleBill, saleOrder, saleOrderProduct, err := getSaleOrderFromDB(ctx, db, req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售订单信息失败")
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDeleteProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 判断订单商品状态
	if saleOrderProduct == nil {
		return nil, errors.New("找不到订单商品")
	}
	if saleOrderProduct.CancelTime == 0 && saleOrderProduct.Status == constant.OrderProductStatusSentKitchen {
		return nil, errors.New("商品已送厨，禁止删除")
	}

	saleOrderProduct.DeleteProduct()

	// 计算订单金额
	afterSaleOrderCalc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("删除商品后,销售订单信息", zap.Any("saleOrder calc", afterSaleOrderCalc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 删除订单商品
		err := repository.NewOrderRepo(db).DeleteOrderProduct(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
			return errUpdate
		}
		// 更新销售账单
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新销售订单失败")
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// 从DB获取销售订单
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
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, saleOrder, saleOrderProduct, err := getSaleOrderFromDB(ctx, db, req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售订单信息失败")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	ctx.Log().Debug("改价前", zap.Any("SalePrice", saleOrderProduct.SalePrice))
	ctx.Log().Debug("改价前", zap.Any("saleOrderProduct calc", saleOrderProduct.BeforeCalc()))
	// 改价
	saleOrderProduct.ChangeProductPrice(req.Price)

	ctx.Log().Debug("改价后", zap.Any("SalePrice", saleOrderProduct.SalePrice))

	// 计算商品数据。折扣、税费、服务
	afterCalc := saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
	ctx.Log().Debug("改价后", zap.Any("saleOrderProduct calc", afterCalc))

	// 计算订单金额
	ctx.Log().Debug("改价前,销售订单信息", zap.Any("saleOrder calc", saleOrder.BeforeCalc()))
	afterSaleOrderCalc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("改价后,销售订单信息", zap.Any("saleOrder calc", afterSaleOrderCalc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdateProduct := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*saleOrderProduct); errUpdateProduct != nil {
			return errUpdateProduct
		}
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errUpdate
		}
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新销售订单失败")
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"改价"事件
	go func() {
		event.NewSystemBus().PublishChangeSaleOrderProductPriceEvent(event.ChangeSaleOrderProductPricePayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.OrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			TotalNum:       saleOrderProduct.Num,
			Price:          req.Price,
		})
	}()

	return info, nil
}

// OrderAmountChange  修改订单金额，整单改价
func (s *orderSrv) OrderAmountChange(ctx context.Context, req req.OrderAmountChangeReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 设置整单改价金额
	saleOrder.SetCustomAmount(req.Price)

	// 整单改价后，整单折扣会取消，需要重新计算订单商品的金额
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"改价"事件
	go func() {
		event.NewSystemBus().PublishChangePriceSaleOrderEvent(event.ChangePriceSaleOrderPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        saleOrder.GetOriginAmount(), // 旧价格为订单的原始应收金额
			NewPrice:        req.Price,
			DiscountType:    constant.DiscountOperationLogTypeChangePriceSaleOrder, // 整单改价的类型值
			SpecialDiscount: decimal.NewFromFloat(saleOrder.GetOriginAmount()).Sub(decimal.NewFromFloat(req.Price)).InexactFloat64(),
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderDiscount  修改订单折扣
func (s *orderSrv) OrderDiscount(ctx context.Context, req req.OrderDiscountReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 在折扣之前计算会员打折后金额。必须在设置折扣之前获取，否则amount值已经改变了
	memberDiscountAmount := saleOrder.GetMemberDiscountAmount()
	// 设置整单折扣率
	saleOrder.SetCustomDiscount(req.GetDiscount())

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"整单打折"事件
	go func() {
		event.NewSystemBus().PublishDiscountSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        memberDiscountAmount, // 旧价格为订单的会员折扣后的金额。如果没有会员折扣，则旧价格为订单应收金额
			NewPrice:        saleOrder.GetAmount(),
			DiscountType:    constant.DiscountOperationLogTypeDiscountSaleOrder,
			RoundingRate:    req.GetOffDiscount(),
			SpecialDiscount: decimal.NewFromFloat(memberDiscountAmount).Sub(decimal.NewFromFloat(saleOrder.GetAmount())).InexactFloat64(),
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderZeroRule  修改订单抹零规则
func (s *orderSrv) OrderZeroRule(ctx context.Context, req req.OrderZeroRuleReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 设置整单改价金额
	saleOrder.SetZeroRule(req.ZeroRule)

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"订单抹零"事件
	go func() {
		event.NewSystemBus().PublishDiscountZeroSaleOrderEvent(event.DiscountZeroSaleOrderPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			DiscountType:    constant.DiscountOperationLogTypeZeroSaleOrder,
			RoundingType:    req.ZeroRule,
			SpecialDiscount: saleOrder.ZeroFee, // ZeroFee这个字段是算好的抹零优惠金额。先计算好订单应付金额，再根据抹零规格进行抹零得到的结果
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderDiscountCancel  取消点餐订单所有优惠折扣，包括改价、打折、抹零。撤销优惠折扣
func (s *orderSrv) OrderDiscountCancel(ctx context.Context, req req.OrderDiscountCancelReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 撤销订单的优惠折扣
	saleOrder.SetAllDiscountCancel()

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布取消优惠折扣事件
	go func() {
		event.NewSystemBus().PublishCancelSaleOrderDiscountEvent(event.CancelSaleOrderDiscountPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			ParentId:  req.SaleBillUuid,
			OrderName: req.SaleOrderUuid,
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderChangePopulation  修改订单人数
func (s *orderSrv) OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error) {
	if req.Population < 0 || req.Population > 999 {
		return nil, errors.New("人数错误")
	}

	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, constant.OptionalUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 点餐助手，拆单后不可以修改人数
	if ctx.GetSource() == constant.SourceAssistant && billInfo.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	oldMealNum := billInfo.MealNum

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderUpdateMealNum, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 修改订单商品人数
	if err := orderRepo.ChangePopulation(req.SaleBillUuid, int(req.Population)); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"修改桌台就餐人数"事件
	go func() {
		event.NewSystemBus().PublishChangeMealNumSaleBillEvent(event.ChangeMealNumSaleBillPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: req.SaleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			OldMealNum: oldMealNum,
			NewMealNum: uint(req.Population),
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderChangeBuffet 调整自助餐
func (s *orderSrv) OrderChangeBuffet(ctx context.Context, req req.OrderChangeBuffetReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillInfoAndProduct(req.SaleBillUuid, 0, 0)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if !saleBill.IsBuffetSaleBill() {
		return nil, errors.New("当前订单不是自助餐类型，无法调整自助餐")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.New("当前订单已拆单，请前去收银机操作")
	}
	if saleBill.IsSplit() {
		return nil, errors.New("当前订单已拆单，无法调整自助餐")
	}
	if err := saleBill.ValidateOrderStatus(constant.OrderUpdateMealNum); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前的所有商品
	saleOrderProductAll := saleBill.GetSaleOrderProductAll()
	// 获取自助餐信息
	buffetList, err := repository.NewBuffetRepo(db).GetBuffetListByUuids(req.BuffetUuids)
	if err != nil || len(buffetList) != len(req.BuffetUuids) {
		return nil, errors.New("自助餐不存在")
	}
	// 销售订单
	saleOrder := saleBill.SaleOrders[0]

	// 是否新增
	oldBuffetIds := []uint64{saleBill.BuffetPackage1Uuid, saleBill.BuffetPackage2Uuid}
	addBuffetIds := slice.Difference(req.BuffetUuids, oldBuffetIds)
	removeBuffetIds := slice.Difference(oldBuffetIds, req.BuffetUuids)

	// 获取自助餐顾客
	customerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&customerTypes, req.BuffetCustomerTypes)
	saleOrderCustomerTypes, buffetUuids, num, maxTimeLimit := saleOrder.GetSaleOrderBuffetCustomerTypes(
		buffetList,
		req.BuffetUuids,
		customerTypes,
		saleBill.SaleBillSetting,
	)

	// 修改
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 删除原来的 CustomerType
		repository.NewOrderRepo(tx).DeleteSaleOrderBuffetCustomerType(saleOrder.Uuid)

		// 创建新的顾客
		if len(saleOrderCustomerTypes) > 0 {
			for _, customer := range saleOrderCustomerTypes {
				if _, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetCustomerType(*customer); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 如果改了套餐信息，需要处理时间 和调整商品是否自助餐
		if len(addBuffetIds) != 0 || len(removeBuffetIds) != 0 {
			// 设置自助餐的开启时间和时长
			saleBill.SetBuffetStartTimeAndDuration(maxTimeLimit)
			// 更新商品为自助餐商品
			buffetProductUuids := []uint64{}
			for _, buffet := range buffetList {
				for _, buffetProduct := range buffet.BuffetProducts {
					buffetProductUuids = append(buffetProductUuids, buffetProduct.ProductPackageUuid)
				}
			}
			for _, product := range saleOrderProductAll {
				if product.IsBuffetProduct() {
					return errors.New("请先清除自助餐套餐内商品")
				}
				if slices.Contains(buffetProductUuids, product.Uuid) {
					product.IsBuffet = 1
					if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProduct(product); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
		}

		// 保存账单。不能用这个方法来创建销售账单，故不使用UpdateOrCreate
		saleBill.MealNum = num
		saleBill.SetBuffetPackage(buffetUuids)
		if err := repository.NewSaleBillRepo(tx).UpdateSaleBillRecord(saleBill); err != nil {
			return errors.WithMessage(err)
		}

		// 重新计算和保存
		saleBill, errSaleBill := repository.NewOrderRepo(tx).GetSaleBillAllInfo(req.SaleBillUuid)
		if errSaleBill != nil {
			return errSaleBill
		}
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// todo 添加操作日志
	// orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderUpdateMealNum, model.SaleBillOperationRecord{
	// 	    Source:        ctx.GetSource(),
	// 	    Remark:        "修改桌台就餐人数",
	// 	    SaleBillUuid:  req.SaleBillUuid,
	// 	    SaleOrderUuid: 0,
	// 	    OperatorUuid:  ctx.GetStaffUuid(),
	// }, map[string]interface{}{
	// 	    "old_meal_num": billInfo.MealNum,
	// 	    "new_meal_num": req.Population,
	// })

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderChangeBuffetClock 调整自助餐加钟
func (s *orderSrv) OrderChangeBuffetClock(ctx context.Context, req req.OrderChangeBuffetClockReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 系统级验证
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return nil, err
	}
	buffetSetting, buffetErr := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if buffetErr != nil {
		return nil, buffetErr
	}
	if buffetSetting.IsAddClock != "1" {
		return nil, errors.New("未开启加钟")
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.New("销售账单不存在")
	}
	if !saleBill.IsBuffetSaleBill() {
		return nil, errors.New("当前不是自助餐类订单，无法调整加钟")
	}
	if saleBill.GetBuffetRemainingSeconds() == -1 {
		return nil, errors.New("当前套餐已经是无限时，无法调整加钟")
	}
	if err := saleBill.ValidateOrderStatus(constant.OrderClock, req.SaleOrderUuid); err != nil {
		return nil, err
	}

	// 获取销售订单
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 获取加钟
	delays, err := base.NewBuffetDelayRepo(db).GetBuffetDelayListByUuids(req.DelayUuids)
	if err != nil || len(delays) != len(req.DelayUuids) {
		return nil, errors.New("加钟不存在")
	}

	// 修改
	if err := db.Transaction(func(tx *gorm.DB) error {
		//
		totalDelayTime := int64(0)
		for _, delay := range delays {
			delayProduct := model.SaleOrderBuffetDelayProduct{
				SaleOrderUuid:   saleOrder.Uuid,
				BuffetDelayUuid: delay.Uuid,
				Price:           delay.Price,
				Name:            delay.Name,
				Num:             saleBill.MealNum,
				DelayTime:       delay.DelayTime,
				Sign:            uuid.New().String(),
			}
			if _, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetDelayProduct(delayProduct); err != nil {
				return err
			}
			//
			totalDelayTime += delay.DelayTime * 60
			//
			saleBill.AddSaleOrderBuffetDelayProduct(saleOrder.Uuid, delayProduct)
		}

		// 设置加钟时长
		saleBill.AddBuffetDelayStartTimeAndDuration(int(totalDelayTime))

		// 计算销售账单金额
		if err = s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// OrderDeskBuffetProductList 获取桌台的自助餐商品列表. 实现思路：通过账单uuid获取该账单的自助餐套餐，然后通过自助餐套餐uuid获取该自助餐套餐的商品列表
func (s *orderSrv) OrderDeskBuffetProductList(ctx context.Context, req req.OrderChangeBuffetProductListReq) (*resp.BuffetProductList, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取销售账单
	saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillBuffetProductList(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	buffetProductList := saleBill.GetBuffetProductList()

	return &buffetProductList, nil
}

// GetSaleBillByDeskId  获取桌台账单信息
func (s *orderSrv) GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error) {
	dbId := ctx.GetDbId()
	deskUuid := ctx.GetDeskUuid()

	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 通过桌台查找到当前桌台的正在进行销售账单
	billInfo, err := orderRepo.GetSaleBillInfoByDesk(deskUuid, constant.OptionalUuid)
	if err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
	}
	return billInfo, nil
}

// OrderProductRemark  修改订单商品备注
func (s *orderSrv) OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfoAndProduct(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if ctx.GetSource() == constant.SourceAssistant && billInfo.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(constant.OrderRemark, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断商品
	if len(billInfo.SaleOrders) == 0 || len(billInfo.SaleOrders[0].SaleOrderProducts) == 0 {
		return nil, errors.New("找不到订单商品")
	}

	// 修改订单商品备注
	if err := orderRepo.ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Remark); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// 通过设备SN获取销售账单uuid
func (s *orderSrv) getSaleBillUuidByDeviceSn(ctx context.Context, deviceSn string) (uint64, error) {
	var saleBillUuid uint64
	// 通过设备sn查询设备ID
	db := s.dbm.GetDB(ctx.GetDbId())
	deviceRepo := repository.NewDeviceRepo(db)
	device, errDevice := deviceRepo.GetDevice(deviceRepo.WhereSn(deviceSn))
	if errDevice != nil {
		return 0, errors.WithMessage(errDevice, "deviceRepo.GetDevice failed")
	}
	ctx.Log().Debug("通过device_sn查询设备uuid", zap.Any("deviceSn", deviceSn), zap.Any("device_uuid", device.Uuid))
	if device.IsDelete() {
		return 0, errors.NewWithCode(constant.CodeParamError, "设备不存在")
	}
	ctx.Log().Debug("通过设备ID查询未挂单的销售账单", zap.Any("device_uuid", device.Uuid))
	// 通过设备ID查询未挂单的销售账单
	if saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(device.Uuid); err != nil {
		if utils.IsNotFoundRecord(err) {
			return 0, nil // 没有点餐账单
		}
		return 0, errors.WithMessage(err)
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
		return nil, errors.WithMessage(errUuid)
	}
	// 没有找到销售账单
	if saleBillUuid == 0 {
		ctx.Log().Info("没有找到销售账单", zap.String("deviceSn", deviceSn))
		// 收银机点餐页面没有销售账单时，检查是否有自动加购的必点方案，如果有，则创建一个销售账单并自动加购商品

		res, err := s.InstantOrderMustPlan2(ctx, deviceSn)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if res == nil {
			return nil, nil
		}
		return res, nil
	}
	// 查询购物车信息
	ctx.Log().Info("查询购物车信息", zap.Uint64("saleBillUuid", saleBillUuid))
	cartInfo, errInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errInfo != nil {
		return nil, errInfo
	}
	return cartInfo, nil
}

// GetOrderCartInfo 获取点餐购物车信息
func (s *orderSrv) GetOrderCartInfo(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	option := &repository.OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	dbId := ctx.GetDbId()
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 通过销售订单ID得到订单商品列表、订单金额信息、账单的销售订单列表

	shopCart, err := orderRepo.GetOrderCartInfo(saleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("saleBillUuid: %d", saleBillUuid))
	}
	if shopCart.SaleBill.IsEndStatus() {
		ctx.Log().Info("销售账单已经结束", zap.Uint64("saleBillUuid", saleBillUuid))
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeDeskOrderEnd, "桌台账单结束"))
	}
	// 重新计算金额
	shopCart.SaleBill.CalcAll()

	// 给订单列表添加订单
	saleOrderList := make([]resp.SaleOrder, 0)
	for _, saleOrder := range shopCart.SaleBill.SaleOrders {
		productList := make([]resp.Product, 0)
		// 给商品列表条件顾客类型
		// 如不是桌台订单、不是自助餐，这个Buffets列表是空的，故不会往productList里加入商品
		{
			for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if orderBuffetCustomer.IsDelete() {
					continue
				}
				// 自助餐顾客价格收费列表
				product := resp.Product{
					Uuid:       orderBuffetCustomer.Uuid,
					LocaleName: orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNames(),
					LocaleAttributeName: dto.LocaleResponse{
						ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					},
					Num:           orderBuffetCustomer.Num, // 这种类型顾客多少个，如老人这个类型2人
					SalePrice:     orderBuffetCustomer.GetOriginPrice(),
					DiscountPrice: orderBuffetCustomer.GetDiscountPrice(),
					Status:        1,
					Remark:        "",
					IsMust:        false,
					IsGift:        false,
					IsCancel:      false,
					IsBuffet:      false,
					AboutBuffet: resp.AboutBuffet{
						IsCustomer:       true,
						IsDelay:          false,
						CustomerTypeUuid: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Uuid,
						BuffetUuid:       orderBuffetCustomer.BuffetPackageUuid,
					},
					SendKitchenTime: orderBuffetCustomer.CreateTime,
					Sign:            cryptor.Md5String(orderBuffetCustomer.GetSign()),
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
					Num:                 delayProduct.Num, // 拆单后不等于桌台人数，但同一个加钟商品的总数等于桌台人数
					SalePrice:           delayProduct.GetAmount(),
					DiscountPrice:       0,  // 加钟商品没有优惠价
					Status:              1,  // 添加后标记送厨状态，不可修改
					Remark:              "", // 加钟商品没有备注
					IsMust:              false,
					IsGift:              false,
					IsCancel:            false,
					IsBuffet:            false,
					AboutBuffet: resp.AboutBuffet{
						IsCustomer: false,
						IsDelay:    true, // 标记该商品是加钟商品
					},
					SendKitchenTime: delayProduct.CreateTime,
					Sign:            cryptor.Md5String(delayProduct.GetSign()),
				}
				productList = append(productList, product)
			}
		}

		// 添加正常商品
		{
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				// 如果查询的是H5已下单的商品和被拒单的商品，则不跳过被删除的商品
				if option.UnorderedH5Product != repository.OrderedH5ProductWithReject {
					if saleOrderProduct.IsDelete() {
						continue
					}
				}

				sendKitchenTime := saleOrderProduct.SendKitchenTime
				if sendKitchenTime == 0 {
					sendKitchenTime = saleOrderProduct.CreateTime
				}
				product := resp.Product{
					Uuid:                saleOrderProduct.Uuid,
					LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: saleOrderProduct.GetAttributeName(),
					Num:                 saleOrderProduct.Num,
					SalePrice:           saleOrderProduct.GetSalePrice(),
					DiscountPrice:       saleOrderProduct.GetPrice(),
					Status:              saleOrderProduct.StatusValue(),
					Remark:              saleOrderProduct.Remark,
					IsMust:              saleOrderProduct.IsMustProduct(),
					IsGift:              saleOrderProduct.IsGiftProduct(),
					IsBuffet:            saleOrderProduct.IsBuffetProduct(),
					IsCancel:            saleOrderProduct.IsCancelProduct(),
					SendKitchenTime:     sendKitchenTime,
					Sign:                cryptor.Md5String(saleOrderProduct.Sign),
					ProductPackageUuid:  saleOrderProduct.ProductPackageUuid,
					AcceptTime:          saleOrderProduct.GetAcceptTime(),
					UnitPrice:           saleOrderProduct.SalePrice,
				}
				if saleOrderProduct.ProductionOrderProduct != nil {
					if saleOrderProduct.ProductionOrderProduct.Status == constant.ProductionOrderProductStatusFinished {
						product.FinishedNum = saleOrderProduct.ProductionOrderProduct.Num
					}
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
			Uuid:               saleOrder.Uuid,
			OrderNo:            saleOrder.OrderNo,
			Status:             saleOrder.Status,
			ProductNum:         productNum,
			ProductList:        productList,
			IsDiscount:         saleOrder.IsDiscount(),
			CustomDiscountRate: saleOrder.CustomDiscountRate,
			ZeroRule:           saleOrder.ZeroRule,
			// 订单金额信息
			AmountInfo: resp.AmountInfo{
				ProductOriginalAmount: saleOrder.ProductOriginalAmount,
				ProductAmount:         saleOrder.ProductAmount,
				ServiceAmount:         saleOrder.ServiceFee,
				TaxAmount:             saleOrder.TaxFee,
				DiscountAmount:        decimal.NewFromFloat(saleOrder.CustomDiscountFee).Round(2).InexactFloat64(),
				MemberDiscountAmount:  saleOrder.MemberDiscountFee,
				Amount:                saleOrder.GetAmount(),
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

	takeout := shopCart.SaleBill.IsTakeout()
	shopCartInfo := &resp.ShopCart{
		SaleBillUuid:  saleBillUuid,
		IsDeskOrder:   shopCart.IsDeskShopCart(),
		IsLock:        shopCart.SaleBill.IsLockStatus(),
		Takeout:       &takeout,
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
			Duration:  time.Now().Unix() - shopCart.SaleBill.CreateTime,
		}
		shopCartInfo.Desk = &deskInfo
		// 如果是自助餐桌台
		if shopCart.SaleBill.IsBuffetSaleBill() {
			shopCartInfo.Buffet = &resp.BuffetInfo{
				RemainingSeconds: shopCart.SaleBill.GetTotalRemainingSeconds(),
				LocaleName:       shopCart.SaleBill.GetBuffetName(),
				IsTimeLimited:    shopCart.SaleBill.IsTimeLimited(),
			}
		}
	}
	return shopCartInfo, nil
}

// GetSaleBillUuidAndSaleOrderUuid 获取销售账单uuid和销售订单uuid
func (s *orderSrv) GetSaleBillUuidAndSaleOrderUuid(ctx context.Context, deskUuid uint64) (uint64, uint64, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取桌台信息
	saleBillUuid, saleOrderUuid, err := repository.NewDeskRepo(db).GetSaleBillUuidAndSaleOrderUuid(deskUuid)
	if err != nil {
		return 0, 0, errors.WithMessage(err)
	}

	return saleBillUuid, saleOrderUuid, nil
}

// 点餐页面，往购物车添加商品。
func (s *orderSrv) InstantOrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error) {
	// 当不填销售账单ID时，表示要新建一个销售账单
	if req.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		_, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if hasInstantOrder {
			return nil, errors.New("参数错误")
		}
		//
		order, err := s.CreateInstantOrder(ctx)
		if err != nil {
			ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
			return nil, errors.WithMessage(err)
		}
		ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
		req.SaleBillUuid = order.SaleBillUuid
		req.SaleOrderUuid = order.SaleOrderUuid
	}

	// 往销售账单里添加商品
	shopCart, err := s.OrderCartProductAdd(ctx, req)
	if err != nil {
		ctx.Log().Info("往点餐账单里添加商品失败", zap.Any("req", req), zap.Any("error", err))
		return nil, errors.WithMessage(err)
	}
	return shopCart, nil
}

func (s *orderSrv) newSaleOrderProduct(ctx context.Context, isDeskSaleBill bool, isH5Product bool, saleBillUuid, saleOrderUuid, deskUuid uint64, flavorProductBomUuid uint64, sauceProductBomUuidList []uint64, productPackageAttributeUuidList []uint64, diningMethod uint, memberDiscountRate, memberCardDiscountRate, customDiscountRate float64) (*model.SaleOrderProduct, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取商品包信息
	productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(flavorProductBomUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if productBom.IsDelete() {
		return nil, errors.New("商品规格已经删除")
	}
	productPackage := productBom.ProductPackage

	// 获取某商品规格信息
	flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(flavorProductBomUuid)
	if errFlavorProductBom != nil {
		return nil, errFlavorProductBom
	}

	// 获取加料信息
	sauceProductBoms := make(map[uint64]*model.ProductBom)
	if len(sauceProductBomUuidList) > 0 {
		sauceProductBomList, errSauceProductBomList := repository.NewProductBomRepo(db).GetSauceProductBomsByUuids(sauceProductBomUuidList)
		if errSauceProductBomList != nil {
			return nil, errSauceProductBomList
		}
		for i, bom := range sauceProductBomList {
			sauceProductBoms[bom.Uuid] = sauceProductBomList[i]
		}
	}

	// 获取属性信息
	productAttributes := make(map[uint64]*model.ProductPackageAttribute)
	if len(productPackageAttributeUuidList) > 0 {
		productAttributeList, errProductAttributeList := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(productPackageAttributeUuidList)
		if errProductAttributeList != nil {
			return nil, errProductAttributeList
		}
		for i, attribute := range productAttributeList {
			productAttributes[attribute.Uuid] = productAttributeList[i]
		}
	}

	// 构建加料信息
	sauces := make([]model.Sauce, 0)
	for sauceProductBomUuid, sauceProductBom := range sauceProductBoms {
		sauce := model.Sauce{
			Name:           sauceProductBom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
			Price:          sauceProductBom.Price,
			ProductBomUuid: sauceProductBomUuid,
		}
		sauces = append(sauces, sauce)
	}

	// 构建属性信息
	attributes := make([]model.Attribute, 0)
	for _, productAttribute := range productAttributes {
		attribute := model.Attribute{
			Name:                 productAttribute.Attribute.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
			ProductAttributeUuid: productAttribute.Attribute.Uuid,
		}
		attributes = append(attributes, attribute)
	}

	isAcceptOrder := constant.OrderProductIsAcceptOrderAccepted // 已接单
	if isH5Product {
		isAcceptOrder = constant.OrderProductIsAcceptOrderUnAccept // 未接单
	}
	saleOrderProduct := model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
		Name:                   productPackage.Name,
		OpenMemberDiscount:     productPackage.OpenDiscount,
		TaxRate:                productPackage.TaxRate(diningMethod),
		DeductStockType:        productPackage.DeductStockType,
		MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
		ImageFileUuid:          productPackage.ImageFileUuid,
		ProductPackageUuid:     productPackage.Uuid,
		SaleBillUuid:           saleBillUuid,
		SaleOrderUuid:          saleOrderUuid,
		MemberDiscountRate:     memberDiscountRate,
		MemberCardDiscountRate: memberCardDiscountRate,
		CustomDiscountRate:     customDiscountRate,
		Sauces:                 sauces,
		Flavor: model.Flavor{
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 填顾客下单时规格的名字 todo preload
			Price:          flavorProductBom.Price,
			ProductBomUuid: flavorProductBomUuid,
		},
		Attribute:     attributes,
		IsAcceptOrder: uint(isAcceptOrder),
	})
	// 设置必点信息
	var mustPlanUuid uint64
	var isRequire bool
	mustPlanUuid, err = s.mustPlanSrv.GetMustPlanUuidByProductPackage(ctx, saleBillUuid, productPackage.Uuid, deskUuid)
	ctx.Log().Debug("获取到必点方案uuid", zap.Any("mustPlanUuid", mustPlanUuid))
	// 判断该必点方案是不是这个sale_biil的
	shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	var mustPlanList []resp.InstantProductMustPlan
	var errMustPlanList error
	if isDeskSaleBill {
		mustPlanList, errMustPlanList = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid)
	} else {
		mustPlanList, errMustPlanList = s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
	}
	if errMustPlanList != nil {
		return nil, errors.WithMessage(errMustPlanList)
	}
	for _, mustPlan := range mustPlanList {
		if mustPlan.Uuid == mustPlanUuid {
			isRequire = true
		}
	}
	if isRequire {
		saleOrderProduct.SetMustPlanInfo(mustPlanUuid)
	}

	return saleOrderProduct, nil
}

// OrderCartProductAdd 向购物车添加商品
func (s *orderSrv) OrderCartProductAdd(ctx context.Context, req req.OrderCartProductAddReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderRemark, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}
	// 录入订单商品数据
	saleOrderProduct, err := s.newSaleOrderProduct(ctx, saleBill.IsDeskSaleBill(), req.IsH5Product(), req.SaleBillUuid, req.SaleOrderUuid, saleBill.DeskUuid, req.FlavorUuid, req.SauceUuidList, req.AttributeUuidList, saleBill.DiningMethod, saleOrder.MemberDiscountRate, saleOrder.MemberCardDiscountRate, saleOrder.CustomDiscountRate)
	if err != nil {
		return nil, errors.WithMessage(err, "构建商品失败")
	}
	// 生成签名
	saleOrderProduct.Sign = saleOrderProduct.GenerateProductSign()
	ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)

	// 判断该商品是不是自助餐商品
	if saleBill.IsBuffetSaleBill() {
		// 获取自助餐商品包uuid列表
		productPackageUuidMap := make(map[uint64]bool) // 自助餐商品包uuid
		if saleBill.BuffetPackage1 != nil {
			for _, buffetProduct := range saleBill.BuffetPackage1.BuffetProducts {
				productPackageUuidMap[buffetProduct.ProductPackageUuid] = true
			}
		}
		if saleBill.BuffetPackage2 != nil {
			for _, buffetProduct := range saleBill.BuffetPackage2.BuffetProducts {
				productPackageUuidMap[buffetProduct.ProductPackageUuid] = true
			}
		}
		// 判断该商品是不是自助餐商品
		if _, ok := productPackageUuidMap[saleOrderProduct.ProductPackageUuid]; ok {
			saleOrderProduct.SetIsBuffet()
		}
	}

	// 查询是否存在签名相同的订单商品
	orderProduct := saleOrder.GetSaleOrderProductBySign(saleOrderProduct.Sign)
	if orderProduct == nil {
		ctx.Log().Debug("不存在相同的sign", zap.Any("sign", saleOrderProduct.Sign))
	}

	// 订单中存在相同签名的商品
	//hasSameSign := false
	if orderProduct != nil {
		// 加上新增的商品数量
		orderProduct.Num += saleOrderProduct.Num
		orderProduct.SetUpdate()
		//hasSameSign = true
	} else {
		// 将新的订单商品加入到订单的商品列表中，用于计算订单金额
		saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, saleOrderProduct)
	}

	errUpdate := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if errUpdate != nil {
		ctx.Log().Error("添加商品失败", zap.Error(errUpdate))
		return nil, errUpdate
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderCartProductNum 修改购物车商品数量
func (s *orderSrv) OrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 检查商品销售库存是否充足
	// todo
	ctx.Log().Debug("获取账单信息")
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	ctx.Log().Debug("获取到账单信息成功")

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderRemark, request.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	ctx.Log().Debug("获取到订单信息成功")

	// 获取销售订单商品信息
	saleOrderProduct, index, errSaleOrderProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
	if errSaleOrderProduct != nil {
		return nil, errSaleOrderProduct
	}
	ctx.Log().Debug("获取到订单商品信息成功")

	if saleOrderProduct.IsCookingProduct() {
		return nil, errors.WithMessage(errors.New("商品已送厨，不能修改数量"))
	}
	// 数量为0删除商品
	if request.Num == 0 {
		res, err := s.OrderProductDelete(ctx, ctx.GetDbId(), ctx.GetStaffUuid(), ctx.GetSource(), req.OrderProductDeleteReq{
			SaleBillUuid:     request.SaleBillUuid,
			SaleOrderUuid:    request.SaleOrderUuid,
			OrderProductUuid: request.SaleOrderProductUuid,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "删除商品失败")
		}
		return res, nil
	}

	// 修改销售订单商品数量
	saleOrderProduct.Num = uint(request.Num)
	ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了商品金额", zap.Any("saleOrderProduct salePrice", saleOrderProduct.SalePrice))
	saleOrder.SaleOrderProducts[index] = saleOrderProduct

	// 计算订单金额
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
			return errUpdate
		}
		ctx.Log().Debug("更新销售订单商品成功")

		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
			return errUpdate
		}
		ctx.Log().Debug("更新销售订单成功")
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "修改商品数量时，保存数据失败")
	}

	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	ctx.Log().Debug("获取新的账单数据")

	return info, nil
}

// checkOrder 检查订单
func (s *orderSrv) checkOrder(ctx context.Context, ignoreMust bool, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, unCookingSaleOrderProducts []*model.SaleOrderProduct, saleOrderProductAll []*model.SaleOrderProduct) (*resp.OrderCheckServiceRes, error) {
	ctx.SetDB(db)
	// 检查必选
	if !ignoreMust {
		// 查询到购物车信息
		shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		var instantMustPlanList []resp.InstantProductMustPlan
		if deskUuid != 0 {
			// 如果是桌台订单
			instantMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		} else {
			// 如果不是桌台订单
			instantMustPlanList, err = s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		}

		mustPlans := make([]resp.InstantProductMustPlan, 0)
		for _, instantProductMustPlan := range instantMustPlanList {
			if instantProductMustPlan.NeedNum > 0 {
				mustPlans = append(mustPlans, instantProductMustPlan)
			}
		}
		if len(mustPlans) > 0 {
			res := &resp.OrderCheckServiceRes{
				Code:          constant.CodeOrderCheckProductMust,
				OrderCheckRes: resp.OrderCheckRes{ProductMustPlanList: &resp.ProductMustPlanList{List: instantMustPlanList}},
			}
			return res, nil
		}
	}

	statusMap := make(map[int][]*model.SaleOrderProduct)
	// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动
	{
		for _, saleOrderProduct := range saleOrderProductAll {
			status, message := saleOrderProduct.CheckProduct()
			ctx.Log().Debug("检查商品", zap.Any("status", status), zap.Any("message", message))
			if status != constant.CodeSuccess {
				statusMap[status] = append(statusMap[status], saleOrderProduct)
				// 如果商品价格变化，更新销售订单商品的价格。都是后台更新价格而未立即更新已选购商品的价格引起的
				// 价格变化包括：
				// 1. 商品规格价格变化
				// 2. 商品小料价格变化
				if status == constant.CodeOrderCheckProductPriceChanged {
					shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					saleBill := shopCartInfo.SaleBill
					s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLastestPrice())
				}
			}
		}
	}

	// 检查限购
	{
		// product_package_uuid => num
		numMap := make(map[uint64]uint) // key为商品包uuid value为已购买数量
		// product_package_uuid => ProductPackage
		productPackageMap := make(map[uint64]*model.ProductPackage) // key为商品包uuid value为已购买数量
		// product_package_uuid => SaleOrderProduct
		saleOrderProductMap := make(map[uint64]*model.SaleOrderProduct) // key为商品包uuid value为订单商品
		for _, saleOrderProduct := range saleOrderProductAll {
			// 限购检查只检查本台的商品，并台过来的商品不记.
			// 跳过非本台的商品
			if !saleOrderProduct.IsCurrentDeskProduct() {
				continue
			}
			productPackageUuid := saleOrderProduct.ProductPackageUuid
			productPackageMap[productPackageUuid] = saleOrderProduct.ProductPackage
			numMap[productPackageUuid] = numMap[productPackageUuid] + saleOrderProduct.Num
			saleOrderProductMap[productPackageUuid] = saleOrderProduct
		}

		for productPackageUuid, num := range numMap {
			limitNum := productPackageMap[productPackageUuid].LimitNum
			// 0表示不限购
			if limitNum == 0 {
				continue
			}
			if num > limitNum {
				statusMap[constant.CodeOrderCheckProductLimitOut] = append(statusMap[constant.CodeOrderCheckProductLimitOut], saleOrderProductMap[productPackageUuid])
			}
		}
	}

	if len(statusMap) > 0 {
		for code, saleOrderProduct := range statusMap {
			products := make([]resp.Product, 0)
			for _, product := range saleOrderProduct {
				products = append(products, resp.Product{
					Uuid:                product.Uuid,
					LocaleName:          product.MultiLanguageName.GetNames(),
					LocaleAttributeName: product.GetAttributeName(),
					Num:                 product.Num,
					SalePrice:           product.SalePrice,
					DiscountPrice:       product.DiscountFee,
					Status:              int(product.Status),
					Remark:              product.Remark,
					IsMust:              product.IsMustProduct(),
					IsGift:              product.IsGiftProduct(),
					IsCancel:            product.IsCancelProduct(),
				})
			}
			res := &resp.OrderCheckServiceRes{
				Code:          code,
				OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: products}},
			}
			return res, nil
		}
	}

	return nil, nil
}

func (s *orderSrv) GetProductDecreaseStockList(ctx context.Context, unCookingSaleOrderProducts []*model.SaleOrderProduct) ([]*model.Product, error) {
	cookingDeductSaleOrderProducts, err := s.getOrderProductForDecreaseStock(ctx, unCookingSaleOrderProducts)
	if err != nil {
		return nil, err
	}
	return s.getDecreaseStockList(ctx, cookingDeductSaleOrderProducts)
}

// 获取下单减库存的商品，从未送厨的商品中获取
func (s *orderSrv) getOrderProductForDecreaseStock(ctx context.Context, unCookingSaleOrderProducts []*model.SaleOrderProduct) ([]*model.SaleOrderProduct, error) {
	products := make([]*model.SaleOrderProduct, 0)
	for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {

		if unCookingSaleOrderProduct.ProductPackage.IsCookingDeductStock() {
			products = append(products, unCookingSaleOrderProduct)
		}
	}
	return products, nil
}

// 获取减库存的清单信息
func (s *orderSrv) getDecreaseStockList(ctx context.Context, cookingDeductSaleOrderProducts []*model.SaleOrderProduct) ([]*model.Product, error) {
	list := make([]*model.Product, 0)
	for _, cookingDeductSaleOrderProduct := range cookingDeductSaleOrderProducts {
		for _, saleOrderProductBom := range cookingDeductSaleOrderProduct.SaleOrderProductBoms {
			// 获取原材料的出库数量
			productBomMaterials := make([]*model.ProductBomMaterials, 0)
			// 如果是规格商品
			if saleOrderProductBom.IsFlavor() {
				for _, productBomMaterial := range saleOrderProductBom.ProductBom.FlavorMaterials {
					productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
						MaterialUuid:  productBomMaterial.MaterialUUID,
						Num:           productBomMaterial.GetDecreaseNum(cookingDeductSaleOrderProduct.Num),
						SaleOrderUuid: cookingDeductSaleOrderProduct.SaleOrderUuid,
					})
				}
			}
			// 如果是小料
			if saleOrderProductBom.IsSauce() {
				for _, material := range saleOrderProductBom.ProductBom.ProductSauce.SauceMaterials {
					productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
						MaterialUuid: material.MaterialUUID,
						Num:          material.GetDecreaseNum(cookingDeductSaleOrderProduct.Num),
					})
				}
			}
			// 获取规格商品的出库数量
			list = append(list, &model.Product{
				ProductBomUuid:       saleOrderProductBom.ProductBomUuid,
				SaleOrderProductUuid: cookingDeductSaleOrderProduct.Uuid,
				SaleOrderUuid:        cookingDeductSaleOrderProduct.SaleOrderUuid,
				Num:                  int(cookingDeductSaleOrderProduct.Num),
				ProductBomMaterials:  productBomMaterials,
			})
		}
	}
	return list, nil
}

// InstantOrderCartProductCooking 送厨购物车商品
func (s *orderSrv) InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, *resp.OrderCheckServiceRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, nil, errSaleBill
	}
	ctx.Log().Debug("获取销售账单信息")

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()
	if len(unCookingSaleOrderProducts) == 0 {
		return nil, nil, errors.New("没有未送厨的商品")
	}

	// 送厨
	checkServiceRes, err := s.ActionCooking(ctx, req.IgnoreMust, saleBill, unCookingSaleOrderProducts)
	if err != nil {
		return nil, nil, err
	}
	if checkServiceRes != nil {
		return nil, checkServiceRes, nil
	}

	ctx.Log().Debug("获取新的购物车信息")
	cartInfo, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取购物车信息失败")
	}
	return cartInfo, nil, nil
}

func newProductionOrder(ctx context.Context, saleOrderUuid, saleBillUuid uint64, unCookingSaleOrderProducts []*model.SaleOrderProduct) *model.ProductionOrder {
	productionOrderUuid, _ := utils.GetID()
	productionOrderProducts := make([]*model.ProductionOrderProduct, 0)
	for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {
		var firstCategoryUuid uint64
		if unCookingSaleOrderProduct.ProductPackage != nil {
			firstCategoryUuid = unCookingSaleOrderProduct.ProductPackage.ProductCategory.GetFirstCategoryUuid()
		}
		attributeName := unCookingSaleOrderProduct.GetAttributeName()
		productionOrderProduct := model.ProductionOrderProduct{
			SaleBillUuid:          saleBillUuid,
			ProductionOrderUuid:   productionOrderUuid,
			SaleOrderProductUuid:  unCookingSaleOrderProduct.Uuid,
			FirstCategoryUuid:     firstCategoryUuid,
			ProductPackageUuid:    unCookingSaleOrderProduct.ProductPackageUuid,
			Num:                   unCookingSaleOrderProduct.Num,
			FlavorName:            unCookingSaleOrderProduct.Name,
			ProductAttributeNames: attributeName.GetLocale(ctx.GetLanguage()),
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

// InstantOrderCartProductReturning 退菜购物车商品
func (s *orderSrv) InstantOrderCartProductReturning(ctx context.Context, req req.OrderCartProductReturningReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 验证高级密码
	if err := s.settingSrv.VerifyAdvancedPassword(ctx, req.Password); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 校验是否合规
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsSendKitchen():
		return nil, errors.New("商品未送厨")
	case req.Num == 0:
		return nil, errors.New("退菜数量不能为0")
	case req.Num > saleOrderProduct.Num:
		return nil, errors.New("退菜数量不能大于当前商品数量")
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	returnFoodReason, err := base.NewReturnFoodReasonRepo(db).GetReturnFoodReasonListByUuids(req.ReturnIds)
	if err != nil {
		return nil, errors.WithMessage(err, "params:", utils.ToJson(req.ReturnIds))
	}
	// 如果查到的原因数量跟提交的原因数量不一致，提示退菜原因不存在
	if len(returnFoodReason) != len(req.ReturnIds) {
		return nil, errors.WithMessage(fmt.Errorf("退菜原因不存在: %v", req.ReturnIds))
	}
	// 构建订单商品退菜原因列表
	returnFoodReasonList := saleOrderProduct.NewSaleOrderProductReasonList(returnFoodReason)

	var returnSaleOrderProduct *model.SaleOrderProduct
	// 如果退菜数量等于该商品的数量，则标记该商品为退菜并在商品的退菜原因列表中添加退菜原因
	if saleOrderProduct.Num == req.Num {
		saleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		returnSaleOrderProduct = saleOrderProduct
	} else {
		// 如果退菜数量小于该商品的数量，则新建一个销售订单商品并在新商品的退菜原因列表中添加退菜原因
		// 1. 修改原商品的数量
		// 2. 新建一个销售订单商品，该商品数量为退菜数量.
		// 3. 判断新建的销售订单商品是否要合并到已有的退菜商品中。当两个退菜商品的签名一致时，将两个商品合并，数量相加
		// 修改原商品的数量
		saleOrderProduct.SetNum(saleOrderProduct.Num - req.Num)
		// 新建一个销售订单商品，该商品数量为退菜数量
		newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderProduct.SaleOrderUuid)
		newSaleOrderProduct.SetNum(req.Num)
		newSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
		if sameSignSaleOrderProduct != nil {
			// 有相同签名的商品。将两个商品合并，数量相加
			sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
			returnSaleOrderProduct = sameSignSaleOrderProduct
		} else {
			// 没有相同签名的商品。将新建的商品添加到销售订单商品列表中
			// CalcAndSaveSaleBill 方法会检查到newSaleOrderProduct没有主键ID，会创建新记录。所以不用另外创建该订单商品，否则会重复创建
			saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, newSaleOrderProduct)
			returnSaleOrderProduct = newSaleOrderProduct
		}
	}

	// 如果退菜商品是下单减库存的商品，则需要创建入库单
	var warehouseForm *model.WarehouseForm
	if returnSaleOrderProduct.DeductStockType == constant.ProductPackageDeductStockTypeCooking {
		productList, err := s.getDecreaseStockList(ctx, []*model.SaleOrderProduct{returnSaleOrderProduct})
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		warehouseForm = model.NewWarehouseForm(productList, req.SaleBillUuid)
	}

	// 新建一个销售订单商品，该商品数量为移动数量
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 创建退菜记录
		if len(returnFoodReasonList) > 0 {
			if err := repository.NewSaleOrderProductReasonRepo(tx).CreateSaleOrderProductReasons(returnFoodReasonList); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		if warehouseForm != nil && len(warehouseForm.WarehouseFormItems) > 0 {
			// 创建入库单
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseFormRecord(*warehouseForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建入库单记录
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseFormItemRecords(warehouseForm.WarehouseFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}

	// 发布“退菜”事件
	go func() {
		s.bus.PublishCancelSaleOrderProductEvent(event.CancelSaleOrderProductPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId:  req.SaleOrderProductUuid,
			ProductId:       saleOrderProduct.ProductPackageUuid,
			ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:     saleOrderProduct.GetAttributeName(),
			ProductAttrList: saleOrderProduct.GetAttributeNameList(),
			TotalNum:        req.Num,
			IsBuffet:        saleOrderProduct.IsBuffet == 1,
			Remark:          saleOrderProduct.Remark,
			Reason:          model.GetReturnFoodReasonNames(returnFoodReason),
			CustomReason:    saleOrderProduct.CancelReason,
		})
	}()
	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderCartProductCancelReturning 取消退菜购物车商品
func (s *orderSrv) InstantOrderCartProductCancelReturning(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品未取消")
	}
	// 更新销售订单商品
	errUpdate := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductReasons(
			saleOrderProduct.SaleOrderUuid,
			saleOrderProduct.Uuid,
			constant.ProductReasonTypeGift,
		); err != nil {
			return errors.WithMessage(err)
		}
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			Status:       0,
			CancelTime:   0,
			CancelReason: "",
		}, map[string]bool{
			"Status":       true,
			"CancelTime":   true,
			"CancelReason": true,
		})
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.New("更新数据失败")
		}
		return nil
	})
	if errUpdate != nil {
		return nil, errors.WithMessage(errUpdate, "操作失败")
	}

	// 发布取消退菜事件
	go func() {
		s.bus.PublishCancelReturnSaleOrderProductEvent(event.CancelReturnSaleOrderProductPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			Num:            saleOrderProduct.Num,
			ParentId:       saleOrder.SaleBillUuid,
			OrderName:      saleOrder.Uuid,
		})
	}()

	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderCartProductChangeDesk 转菜购物车商品
func (s *orderSrv) InstantOrderCartProductChangeDesk(ctx context.Context, req req.OrderCartProductChangeDeskReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderChangeTable, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}

	// 获取目标桌台的信息和销售账单
	targetDesk, err := repository.NewDeskRepo(db).GetDeskAndSaleBillByDeskUuid(req.DeskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if err := targetDesk.CheckProductChangeDesk(); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取目标桌台的销售账单信息
	targetSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(targetDesk.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "目标桌台的销售账单不存在", fmt.Sprintf("targetDesk.SaleBillUuid: %d", targetDesk.SaleBillUuid))
	}

	// 将销售订单商品的sale_bill_uuid和sale_order_uuid更新为新的桌台
	targetSaleOrder := targetSaleBill.SaleOrders[0]
	saleOrderProduct.SaleBillUuid = targetDesk.SaleBillUuid
	saleOrderProduct.SaleOrderUuid = targetSaleOrder.Uuid
	// 将商品的折扣改为使用目标桌台的折扣
	saleOrderProduct.MemberDiscountRate = targetSaleOrder.MemberDiscountRate
	saleOrderProduct.MemberCardDiscountRate = targetSaleOrder.MemberCardDiscountRate
	saleOrderProduct.CustomDiscountRate = targetSaleOrder.CustomDiscountRate
	saleOrderProduct.SetUpdate() // 标记该商品的记录要更新，会在原桌台账单的CalcAndSaveSaleBill方法中更新
	// 将商品添加到目标桌台的销售订单中
	targetSaleOrder.SaleOrderProducts = append(targetSaleOrder.SaleOrderProducts, saleOrderProduct)

	// todo
	// 如果商品已经送厨且未完成出餐，要为这个商品生成一个目标桌台的送厨单并从原桌台的送厨单中删除该商品

	// todo
	// 如果商品已经送厨且已完成出餐，要为这个商品生成一个目标桌台的送厨单并完成出餐，并从原桌台的送厨单中删除该商品。考虑到该商品制作完成后，在厨显撤回制作完成，此时该商品不应该在原桌台的送厨单上

	// todo 重新生成订单商品签名

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 重新计算原桌台的销售账单
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		// 重新计算目标桌台的销售账单
		if err := s.CalcAndSaveSaleBill(ctx, tx, targetSaleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新数据失败")
	}

	// 发布转菜事件
	go func() {
		s.bus.PublishChangeDeskSaleOrderProductEvent(event.ChangeDeskSaleOrderProductPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			TotalNum:       saleOrderProduct.Num,
			ToOrderId:      targetSaleOrder.Uuid,
			ToTableId:      targetDesk.Uuid,
			ToTableNo:      targetDesk.DeskNo,
		})
	}()

	// 获取新的购物车信息
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}

	return cartInfo, nil
}

// InstantOrderCartProductGiving 赠送购物车商品
func (s *orderSrv) InstantOrderCartProductGiving(ctx context.Context, req req.OrderCartProductGivingReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}
	//  验证赠菜标签
	reasons := [][2]uint64{}
	if len(req.GiftIds) > 0 {
		_reasons, notFound, err := base.NewGiftOrFreeOrderReasonRepo(db).ExistsByUuids(req.GiftIds)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if len(notFound) > 0 {
			return nil, fmt.Errorf("以下赠菜原因不存在: %v", notFound)
		}
		reasons = _reasons
	}
	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			GiftTime:   time.Now().Unix(),
			GiftReason: req.Reason,
		})
		// 添加赠菜原因
		if len(reasons) > 0 {
			if err := repository.NewSaleOrderProductRepo(tx).CreateSaleOrderProductReasons(
				saleOrderProduct.SaleOrderUuid,
				saleOrderProduct.Uuid,
				constant.ProductReasonTypeGift,
				reasons,
			); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	// 发布赠菜事件
	go func() {
		s.bus.PublishGiftSaleOrderProductEvent(event.GiftSaleOrderProductPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			ProductPrice:   saleOrderProduct.Price,
			TotalNum:       saleOrderProduct.Num,
			TotalPrice:     decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromInt(int64(saleOrderProduct.Num))).Round(2).InexactFloat64(),
			FreeTagIds:     req.GiftIds,
			FreeRemark:     req.Reason,
		})
	}()
	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo)
	}
	return cartInfo, nil
}

// InstantOrderCartProductCancelGiving 取消赠送购物车商品
func (s *orderSrv) InstantOrderCartProductCancelGiving(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsGiftProduct():
		return nil, errors.New("商品未赠送")
	}

	// 更新销售订单商品
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductReasons(
			saleOrderProduct.SaleOrderUuid,
			saleOrderProduct.Uuid,
			constant.ProductReasonTypeGift,
		); err != nil {
			return errors.WithMessage(err)
		}
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			GiftTime:   0,
			GiftReason: "",
		}, map[string]bool{
			"GiftTime":   true,
			"GiftReason": true,
		})
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.New("更新数据失败")
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "操作失败")
	}

	// 发布取消赠菜事件
	go func() {
		s.bus.PublishCancelGiftSaleOrderProductEvent(event.CancelGiftSaleOrderProductPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			ProductPrice:   saleOrderProduct.Price,
			TotalNum:       saleOrderProduct.Num,
			TotalPrice:     decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromInt(int64(saleOrderProduct.Num))).Round(2).InexactFloat64(),
			ParentId:       saleOrder.SaleBillUuid,
			OrderName:      saleOrder.Uuid,
		})
	}()

	// 获取新的购物车信息
	cartInfo, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderMustPlan 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx, deviceSn)
	if errUuid != nil {
		return nil, errors.WithMessage(errUuid, "无法找到销售账单")
	}
	ctx.Log().Debug("查询必点方案列表", zap.Any("saleBillUuid", saleBillUuid), zap.Any("deviceSn", deviceSn))

	mustPlanList := make([]resp.InstantProductMustPlan, 0)
	// product_bom_uuid => *resp.InstantMustPlanProduct
	autoFlavorProduct := make(map[uint64]*resp.InstantMustPlanProduct) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购

	// 查询到购物车信息
	shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "repository.NewOrderRepo(db).GetOrderCartInfo failed", fmt.Sprintf("saleBillUuid:%d", saleBillUuid))
	}

	planList, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
	if errMustPlan != nil {
		ctx.Log().Info("获取必点列表失败", zap.Error(errMustPlan))
		return nil, errors.New("获取必点列表失败")
	}
	mustPlanList = planList
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
	if len(autoFlavorProduct) > 0 && shopCartInfo.SaleBill.IsAutoAddMustProduct() {
		errTx := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			shopCart, err = autoAddSaleOrderProduct(ctx, db, s, autoFlavorProduct)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			return nil
		})
		if errTx != nil {
			return nil, errors.WithMessage(errTx, "自动添加必点商品失败")
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
	if shopCartInfo.SaleBill.IsShowMustPlan() {
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

	deviceRepo := repository.NewDeviceRepo(db)
	device, errDevice := deviceRepo.GetDevice(deviceRepo.WhereSn(deviceSn))
	if errDevice != nil {
		return nil, errors.WithMessage(errDevice)
	}
	ctx.Log().Debug("通过device_sn查询设备uuid", zap.Any("deviceSn", deviceSn), zap.Any("device_uuid", device.Uuid))
	if device.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeParamError, "设备不存在")
	}
	ctx.Log().Debug("通过设备ID查询未挂单的销售账单", zap.Any("device_uuid", device.Uuid))
	// 通过设备ID查询未挂单的销售账单
	saleBill, errGetSaleBill := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(device.Uuid)
	if errGetSaleBill != nil {
		if !utils.IsNotFoundRecord(errGetSaleBill) {
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

// InstantOrderMustPlan 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan2(ctx context.Context, deviceSn string) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取点餐的必选方案列表
	mustPlanList, err := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, make(ro.MustPlanProductInfo))
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取必点列表失败"), fmt.Sprintf("err:%v", err))
	}
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// product_bom_uuid => *resp.InstantMustPlanProduct
	autoFlavorProduct := make(map[uint64]*resp.InstantMustPlanProduct) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购

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
	if len(autoFlavorProduct) > 0 {
		err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			cart, err := autoAddSaleOrderProduct(ctx, db, s, autoFlavorProduct)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			shopCart = cart
			return nil
		})
		if err != nil {
			return nil, errors.WithMessage(err, "自动添加必点商品失败")
		}
	}

	return shopCart, nil
}

// InstantOrderPaymentInfo 获取结账页面信息
func (s *orderSrv) InstantOrderPaymentInfo(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error) {
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	// 获取销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	if saleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.New("销售账单已结束"))
	}
	saleOrder := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}
	if err := saleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}
	paymentMethods := repository.NewPaymentMethodRepo(db).GetPaymentMethodsByCtx(ctx)

	var memberInfo *resp.MemberInfo
	if saleOrder.Member != nil {
		memberInfo = &resp.MemberInfo{
			Uuid:     saleOrder.ConsumerUuid,
			Nickname: saleOrder.GetMemberName(),
			Card:     resp.CardInfo{Name: saleOrder.Member.GetMemberCardName()},
			Level:    resp.LevelInfo{Name: saleOrder.Member.GetMemberLevelName()},
			Balance:  saleOrder.Member.GetBalanceAll(),
			Points:   saleOrder.Member.Point,
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

	companySetting := ctx.GetCompanySetting()
	paymentApp, paymentAppErr := admin.NewPaymentAppRepo(s.dbm.GetDB(0)).GetPaymentAppCompanyUuid(ctx.GetCompanyUuid())
	serviceFeeRate := saleBill.SaleBillSetting.GetServiceFeeRate()
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	for _, paymentMethod := range paymentMethods {
		// 不显示免单
		if paymentMethod.Code == constant.PaymentMethodCodeFreePay {
			continue
		}
		// 没有启用会员功能不显示余额
		if companySetting.IsOpenMember != 1 {
			continue
		}
		// LianLianPay 没有配置支付信息 不显示
		if paymentMethod.Code == constant.PaymentMethodCodeLianLianWechatPay ||
			paymentMethod.Code == constant.PaymentMethodCodeLianLianAliPay ||
			paymentMethod.Code == constant.PaymentMethodCodeLianLianQRPromptPay {
			if paymentAppErr != nil || paymentApp == nil || paymentApp.ID == 0 {
				continue
			}
		}

		var logoUrl string
		var qrcodeUrl string
		if paymentMethod.LogoFile != nil {
			logoUrl = paymentMethod.LogoFile.GetUrl(baseUrl)
		}
		if paymentMethod.QrcodeFile != nil {
			qrcodeUrl = paymentMethod.QrcodeFile.GetUrl(baseUrl)
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

		saleOrderAmount := saleOrder.GetAmount()
		saleOrderOriginAmount := saleOrder.CalcOrderOriginAmount(serviceFeeRate, serviceFeeValue, taxFeeType)
		if commissionFee > 0 {
			// 如果有手续费
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrderOriginAmount,
				SaleOrderAmount:       saleOrderAmount,
				CommissionFee:         commissionFee,
				UnpaidAmount:          saleOrder.CalcUnPayAmount(true),
				ZeroAmount:            0, // 只有没有手续费时才会抹零
				ZeroRule:              constant.SaleBillSettingCheckoutZeroingMethodNone,
				PaymentMethodUuid:     methodItem.Uuid,
			}
			amounts = append(amounts, amount)
		} else {
			hasCommission := false
			// 如果没有手续费
			zeroFee := saleOrder.CalcCheckOutZeroFee()
			if methodItem.FeePercent != 0 {
				// 如果支付方式有手续费，则不能抹零，抹零金额为0
				zeroFee = 0
				hasCommission = true
			}
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrderOriginAmount,
				SaleOrderAmount:       saleOrderAmount,
				CommissionFee:         commissionFee,
				UnpaidAmount:          saleOrder.CalcUnPayAmount(hasCommission),
				ZeroAmount:            zeroFee, // 只有没有手续费时且支付方式不需要手续费才会抹零
				ZeroRule:              saleOrder.ZeroCheckoutRule,
				PaymentMethodUuid:     methodItem.Uuid,
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

// InstantOrderPaymentQrcode
func (s *orderSrv) InstantOrderPaymentQrcode(ctx context.Context, req req.InstantOrderPaymentQrcodeReq) (*resp.InstantOrderPaymentQrcodeInfoResp, error) {
	// baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid + req.PaymentMethodUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid + req.PaymentMethodUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 获取销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillInfoAndPaymentOrders(req.SaleBillUuid, req.SaleOrderUuid, 0)
	if errSaleBill != nil {
		return nil, errSaleBill
	}
	if saleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.New("销售账单已结束"))
	}
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}
	if err := saleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断当前是否连连支付
	paymentMethods := repository.NewPaymentMethodRepo(db).GetPaymentMethodsByCtx(ctx)
	var paymentMethod *model.PaymentMethod
	for _, method := range paymentMethods {
		if method.Code != constant.PaymentMethodCodeLianLianWechatPay &&
			method.Code != constant.PaymentMethodCodeLianLianAliPay &&
			method.Code != constant.PaymentMethodCodeLianLianQRPromptPay {
			continue
		}
		if method.Uuid == req.PaymentMethodUuid {
			paymentMethod = method
			break
		}
	}

	// 支付方式不可用
	if paymentMethod == nil {
		return nil, errors.New("支付方式不可用")
	}

	// 计算手续费
	percent := paymentMethod.GetFeePercent()
	commissionFee := decimal.NewFromFloat(req.PaymentAmount).Mul(decimal.NewFromFloat(percent)).InexactFloat64()
	paymentAmount := decimal.NewFromFloat(req.PaymentAmount).Add(decimal.NewFromFloat(commissionFee)).InexactFloat64()

	// 判断支付金额是否大于未收金额.只能现金支付大于未收金额
	unpaidAmount := saleOrder.GetUnpaidAmount()
	if unpaidAmount < req.PaymentAmount {
		return nil, errors.New("超出订单剩余可支付金额")
	}

	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "添加支付订单-获取货币设置失败")
	}

	// 创建支付订单
	payment, err := lianlian.NewPaymentRepo(ctx, s.dbm).CreatePayment(paymentMethod.Code, paymentAmount)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 创建支付订单
	paymentOrder := &model.PaymentOrder{
		PaymentMethodName:    paymentMethod.PaymentName,
		PaymentMethodUuid:    req.PaymentMethodUuid,
		PaymentFeePercent:    percent,
		RelatedType:          constant.PaymentOrderRelatedTypeSaleOrder,
		RelatedUuid:          req.SaleOrderUuid,
		CurrencyUnit:         currencySetting.Unit,
		PaymentAmount:        req.PaymentAmount,
		PaymentCommissionFee: commissionFee,
		Amount:               paymentAmount,
		Status:               constant.PaymentOrderStatusUnPay,
	}

	// 创建或更新支付单
	if err := repository.NewPaymentOrderRepo(db).UpdateOrCreatePaymentOrderRecord(*paymentOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 在 infoResp 初始化之前添加
	qrCodeExpireSec := 480 // 默认8分钟过期
	if expireSec, err := strconv.Atoi(payment.Order.QrCodeExpireSec); err == nil {
		qrCodeExpireSec = expireSec
	}

	infoResp := &resp.InstantOrderPaymentQrcodeInfoResp{
		PaymentOrderUuid: paymentOrder.Uuid,
		QrCode:           payment.Order.QrCode,
		QrCodeExpireSec:  qrCodeExpireSec,
		Status:           constant.PaymentOrderStatusUnPay,
		PaymentAmount:    paymentAmount,
	}

	return infoResp, nil
}

// InstantOrderPaymentCreate 给销售订单创建一个支付单
func (s *orderSrv) InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 判断订单是否已经结束，若订单结束则拒绝操作
	if err := s.checkCanOperateOrder(ctx, req.SaleBillUuid, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	if !saleBill.IsCookingStatus() {
		return nil, errors.WithMessage(errors.New("请先送厨商品"))
	}
	// 判断销售订单是否可操作
	if err := saleBill.ValidateOrderStatus(constant.OrderSettle, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("无法查询到销售订单"))
	}

	paymentMethod, err := repository.NewPaymentMethodRepo(db).GetPaymentMethodByUuid(req.PaymentMethodUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "添加支付订单-获取货币设置失败")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	commissionAmount := infoResp.GetCommissionAmount()

	// 判断支付金额是否大于未收金额.只能现金支付大于未收金额
	unpaidAmount := infoResp.GetUnpaidAmount()
	if unpaidAmount < req.PaymentAmount {
		if paymentMethod.Code != constant.PaymentMethodCodeCash {
			return nil, errors.WithMessage(errors.New(fmt.Sprintf("支付金额不能大于未收金额 %.2f", unpaidAmount)))
		}
	}

	percent := paymentMethod.GetFeePercent()
	commissionFee := decimal.NewFromFloat(req.PaymentAmount).Mul(decimal.NewFromFloat(percent)).InexactFloat64()
	amount := decimal.NewFromFloat(req.PaymentAmount).Add(decimal.NewFromFloat(commissionFee)).InexactFloat64()
	paymentOrder := &model.PaymentOrder{
		PaymentMethodName:    paymentMethod.PaymentName,
		PaymentMethodUuid:    req.PaymentMethodUuid,
		PaymentFeePercent:    percent,
		RelatedType:          constant.PaymentOrderRelatedTypeSaleOrder,
		RelatedUuid:          req.SaleOrderUuid,
		CurrencyUnit:         currencySetting.Unit,
		PaymentAmount:        req.PaymentAmount,
		PaymentCommissionFee: commissionFee,
		Amount:               amount, // 实收金额
		TransactionNumber:    "",
		Status:               constant.PaymentOrderStatusPaid,
	}

	// 判断这个支付方式是否已经支付过，如果已经支付过，则更新支付单
	paymentOrderList, err := repository.NewPaymentOrderRepo(db).GetPaymentOrderListBySaleOrderUuid(req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	for _, oldPaymentOrder := range paymentOrderList {
		if oldPaymentOrder.PaymentMethodUuid == req.PaymentMethodUuid {
			paymentOrder.SetBaseModel(oldPaymentOrder.BaseModel) // 将旧付款单的ID、uuid赋值给新付款单，让旧的付款单记录被新的付款单更新
			break
		}
	}

	// 如果支付方式是含手续费的支付方式且该订单之前未产生过含手续且该订单设置了结账抹零，则自动取消结账抹零
	needCancelCheckoutZeroRule := paymentMethod.HasCommission() && commissionAmount == 0 && saleOrder.HasCheckoutZeroRule()
	if needCancelCheckoutZeroRule {
		saleOrder.SetCheckoutZeroRuleCancel() // 将订单的结账抹零规格设置为实款实收，并清空结账抹零金额
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建或更新支付单
		if err := repository.NewPaymentOrderRepo(db).UpdateOrCreatePaymentOrderRecord(*paymentOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售订单
		if saleOrder.GetUpdate() {
			// 如果销售订单有更新，则更新销售订单
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"自动取消结账抹零"事件
	// 如果支付方式是含手续费的支付方式且该订单之前未产生过含手续且该订单设置了结账抹零，则自动取消结账抹零
	if needCancelCheckoutZeroRule {
		go func() {
			s.bus.PublishCheckoutZeroCancelSaleOrderEvent(event.CheckoutZeroCancelSaleOrderPayload{
				BasePayload: event.BasePayload{
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  req.SaleBillUuid,
					SaleOrderUuid: req.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
			})
		}()
	}

	newInfoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return newInfoResp, nil
}

// 检查是否可以操作
func (s *orderSrv) checkCanOperateOrder(ctx context.Context, saleBillUuid, saleOrderUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 判断订单是否已经结束，若订单结束不能创建支付单、不能撤销付款单、不能查看结账页信息
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return errors.WithMessage(errSaleBill)
	}
	if saleBill.IsEndStatus() {
		return errors.WithMessage(errors.New("销售账单已结束"))
	}
	saleOrder := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return errors.New("无法查询到销售订单")
	}
	if err := saleOrder.ValidateOrderStatus(); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// InstantOrderPaymentCancel 撤销一个支付单
func (s *orderSrv) InstantOrderPaymentCancel(ctx context.Context, req req.InstantOrderPaymentCancelReq) (*resp.InstantOrderPaymentInfoResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 判断订单是否已经结束，若订单结束则拒绝操作
	if err := s.checkCanOperateOrder(ctx, req.SaleBillUuid, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	paymentOrder, err := repository.NewPaymentOrderRepo(db).GetPaymentOrderRecord(req.PaymentOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 撤销支付单
	paymentOrder.Cancel()
	// 更新支付单
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*paymentOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}
	infoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return infoResp, nil
}

// 获取没有出库记录的销售订单商品
func (s *orderSrv) getSaleOrderProductWithoutWarehouseOutForm(ctx context.Context, saleOrderUuid uint64, allSaleOrderProducts []*model.SaleOrderProduct) ([]*model.SaleOrderProduct, error) {

	// 获取该订单所有的出库记录
	warehouseOutFormItems, err := repository.NewWarehouseFormRepo(ctx.GetDB()).GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	productMap := make(map[uint64]*model.SaleOrderProduct)
	for _, saleOrderProduct := range allSaleOrderProducts {
		productMap[saleOrderProduct.Uuid] = saleOrderProduct
	}
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		if _, ok := productMap[warehouseOutFormItem.SaleOrderProductUuid]; ok {
			delete(productMap, warehouseOutFormItem.SaleOrderProductUuid)
		}
	}

	list := make([]*model.SaleOrderProduct, 0)
	for _, saleOrderProduct := range productMap {
		list = append(list, saleOrderProduct)
	}
	return list, nil
}

// InstantOrderPaymentFinish 完成销售订单的付款结账
func (s *orderSrv) InstantOrderPaymentFinish(ctx context.Context, req req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("GetSaleBillAllInfo", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), errSaleBill)))
		return nil, errSaleBill
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	amount := infoResp.Amounts.List[0]
	unpaidAmount := amount.UnpaidAmount
	if unpaidAmount > 0 {
		return nil, errors.New("销售订单未结清")
	}

	// 检查是否有未送厨的商品。场景：当收银机1结账时，收银机2加购了新的商品。
	if len(saleBill.GetSaleOrderProductUnCooking()) > 0 {
		return nil, errors.New("有未送厨的商品")
	}

	// 最终应收=应收金额+手续费
	finalAmount := decimal.NewFromFloat(saleOrder.Amount).Add(decimal.NewFromFloat(amount.CommissionFee)).InexactFloat64()

	totalPay := float64(0)
	for _, paymentOrder := range infoResp.PaymentOrders.List {
		totalPay = decimal.NewFromFloat(totalPay).Add(decimal.NewFromFloat(paymentOrder.Amount)).InexactFloat64()
	}

	// 计算找零金额。
	changeAmount := float64(0)
	if totalPay > finalAmount {
		changeAmount = decimal.NewFromFloat(totalPay).Sub(decimal.NewFromFloat(finalAmount)).InexactFloat64()
	}

	// 计算抹零金额. 只有没有手续费时，才能抹零
	if amount.CommissionFee == 0 {
		saleOrder.SetCheckOutZeroFee()
	}

	// 修改订单为支付完成，并记录找零金额、最终付款金额等结算后才计算的字段
	final := model.FinalAmount{
		PaymentAmount:        totalPay,
		ChangeAmount:         changeAmount,
		ZeroCheckoutFee:      saleOrder.CalcCheckOutZeroFee(),
		FinalPrice:           finalAmount,
		PaymentCommissionFee: amount.CommissionFee,
		GiftAmount:           saleOrder.CalcGiftAmount(saleOrder.SaleOrderProducts),
	}
	saleOrder.SetFinishStatus(final) // 设置销售订单状态为已结清

	// 更新销售账单
	updateSaleBill := false
	if saleBill.CanFinishSaleBill() {
		saleBill.SetFinishSaleBill()
		saleBill.CalcAll()
		updateSaleBill = true
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新销售订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
			return errors.WithMessage(err)
		}

		// 更新账单
		if updateSaleBill {
			if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
				return errUpdateSaleBill
			}
			// 如果是桌台账单，则将桌台状态改为待清台或者空闲
			// 待清台，将桌台信息中的sale_bill_uuid设为0、状态为开台状态
			// 空闲，将桌台信息中的sale_bill_uuid设为0、状态为未开台状态
			// 完成销售账单后，桌台是待清台还是空闲状态由系统是否设置了自动清台决定。若不自动清台，则桌台为待清台桌台。若自动清台，则桌台为空闲桌台
			if saleBill.IsDeskSaleBill() && businessSetting.IsAutoClearDesk() {
				// 结账自动清台，将桌台状态设置为空闲
				saleBill.Desk.SetCloseDesk()
				if err := repository.NewDeskRepo(db).UpdateDeskRecord(*saleBill.Desk); err != nil {
					return err
				}
			}
			if saleBill.IsDeskSaleBill() && !businessSetting.IsAutoClearDesk() {
				// 结账不自动清台，将桌台状态设置为待清台
				saleBill.Desk.SetWaitClearDesk()
				if err := repository.NewDeskRepo(db).UpdateDeskRecord(*saleBill.Desk); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断销售订单的每个商品是否都已有对应的出库记录
	// 获取没有出库记录的销售订单商品
	withoutWarehouseOutFormSaleOrderProducts, err := s.getSaleOrderProductWithoutWarehouseOutForm(ctx, saleOrder.Uuid, saleOrder.SaleOrderProducts)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取减库存的清单信息
	decreaseStockList, err := s.getDecreaseStockList(ctx, withoutWarehouseOutFormSaleOrderProducts)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 构建出库单
	warehouseOutForm := model.NewWarehouseOutForm(decreaseStockList, true, req.SaleBillUuid)
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if len(warehouseOutForm.WarehouseOutFormItems) > 0 {
			// 创建出库单
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormRecord(*warehouseOutForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建出库单记录
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormItemRecords(warehouseOutForm.WarehouseOutFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新该销售订单的所有出库单记录为已出库
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecordsStatus(saleOrder.Uuid); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"结账"事件
	saleOrderAmount := saleOrder.GetAmount()
	saleOrderPaymentAmount := saleOrder.PaymentAmount
	saleOrderChangeAmount := saleOrder.ChangeAmount
	go func() {
		payTypes := make([]event.PayType, 0)
		paymentAmount := decimal.NewFromFloat(0)
		for _, paymentOrder := range infoResp.PaymentOrders.List {
			payTypes = append(payTypes, event.PayType{
				Name:           paymentOrder.PaymentMethodName,
				Value:          paymentOrder.PaymentMethodCode,
				DisabledCancel: utils.BoolToUint(paymentOrder.DisabledCancel),
				Price:          paymentOrder.Amount,
				FeeMoney:       paymentOrder.PaymentCommissionFee,
			})
			paymentAmount = paymentAmount.Add(decimal.NewFromFloat(paymentOrder.Amount))
		}
		s.bus.PublishCheckoutSaleOrderEvent(event.CheckoutSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleBill:    saleBill,
			OrderPrice:  saleOrderAmount,
			PayPrice:    saleOrderPaymentAmount,
			ActualPrice: paymentAmount.InexactFloat64(), // 最终实付金额=每笔付款单的付款金额之和（含手续费）
			ChangeDue:   saleOrderChangeAmount,
			PayType:     payTypes,
		})
	}()

	payMethods := make([]resp.PayMethod, 0)
	for _, paymentOrder := range infoResp.PaymentOrders.List {
		method := resp.PayMethod{
			Uuid: paymentOrder.PaymentMethodUuid,
			Name: paymentOrder.PaymentMethodName,
		}
		payMethods = append(payMethods, method)
	}
	orderFinishResp := &resp.OrderFinishResp{
		SaleBillUuid:  req.SaleBillUuid,
		SaleOrderUuid: req.SaleOrderUuid,
		AmountInfo: resp.PayAmountInfo{
			OrderAmount:  saleOrderAmount,
			PayAmount:    saleOrderPaymentAmount,
			ChangeAmount: saleOrderChangeAmount,
		},
		PayMethodList: resp.PayMethodList{
			List: payMethods,
		},
	}

	return orderFinishResp, nil
}

// InstantOrderFree 免单
func (s *orderSrv) InstantOrderFree(ctx context.Context, req req.InstantOrderFreeReq) (*resp.OrderFinishResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 销售账单已经结束
	if err := saleBill.ValidateOrderStatus(constant.OrderSettle, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("无法查询到销售订单"))
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 已经部分支付，无法进行免单
	if len(infoResp.PaymentOrders.List) > 0 {
		return nil, errors.WithMessage(errors.New("订单已部分支付，无法进行免单"))
	}

	// 获取免单原因
	freeReasons, err := base.NewGiftOrFreeOrderReasonRepo(db).GetFreeOrderReasonListByUuids(req.ReasonIds)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if len(freeReasons) != len(req.ReasonIds) {
		return nil, errors.WithMessage(errors.New("免单原因不存在"), fmt.Sprintf("原因ids：%v", req.ReasonIds))
	}

	freeOrderReasons := saleOrder.NewFreeOrderReason(freeReasons)

	// 设置该销售订单为免单
	saleOrder.SetFreeOrder(req.Reason, freeOrderReasons)

	updateSaleBill := false
	// 如果销售账单中只有一个销售订单，则可以结束销售账单
	if saleBill.CanFinishSaleBill() {
		saleBill.SetFinishSaleBill()
		saleBill.CalcAll()
		updateSaleBill = true
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建免单原因
		if len(freeOrderReasons) > 0 {
			if err := repository.NewSaleOrderProductReasonRepo(db).CreateSaleOrderProductReasons(freeOrderReasons); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新销售订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
			return errors.WithMessage(err)
		}

		// 更新账单
		if updateSaleBill {
			if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
				return errUpdateSaleBill
			}
			// 如果是桌台账单，则将桌台状态改为待清台或者空闲
			// 待清台，将桌台信息中的sale_bill_uuid设为0、状态为开台状态
			// 空闲，将桌台信息中的sale_bill_uuid设为0、状态为未开台状态
			// 完成销售账单后，桌台是待清台还是空闲状态由系统是否设置了自动清台决定。若不自动清台，则桌台为待清台桌台。若自动清台，则桌台为空闲桌台
			if saleBill.IsDeskSaleBill() && businessSetting.IsAutoClearDesk() {
				// 结账自动清台，将桌台状态设置为空闲
				saleBill.Desk.SetCloseDesk()
				if err := repository.NewDeskRepo(db).UpdateDeskRecord(*saleBill.Desk); err != nil {
					return err
				}
			}
			// 如果是桌台订单，且不自动清台
			if saleBill.IsDeskSaleBill() && !businessSetting.IsAutoClearDesk() {
				// 结账不自动清台，将桌台状态设置为待清台
				saleBill.Desk.SetWaitClearDesk()
				if err := repository.NewDeskRepo(db).UpdateDeskRecord(*saleBill.Desk); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	orderFinishResp := &resp.OrderFinishResp{
		SaleBillUuid:  req.SaleBillUuid,
		SaleOrderUuid: req.SaleOrderUuid,
		AmountInfo: resp.PayAmountInfo{
			OrderAmount: saleOrder.GetAmount(),
		},
		PayMethodList: resp.PayMethodList{
			List: []resp.PayMethod{
				{
					Name: "免单",
				},
			},
		},
	}

	// 发布"免单"事件
	go func() {
		s.bus.PublishFreeSaleOrderEvent(event.FreeSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderPrice:    saleOrder.GetAmount(),
			PayPrice:      0, // 免单时，支付金额为0
			ActualPrice:   0, // 免单时，实际支付金额为0
			ChangeDue:     0, // 免单时，找零金额为0
			IsFree:        utils.BoolToUint(true),
			DiscountMoney: saleOrder.GetAmount(),
		})
	}()

	return orderFinishResp, nil
}

// InstantOrderPaymentZeroRule 设置结账抹零规则
func (s *orderSrv) InstantOrderPaymentZeroRule(ctx context.Context, req req.InstantOrderPaymentZeroRuleReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 验证订单是否可操作
	if err := saleBill.ValidateOrderStatus(constant.OrderSettle); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("无法查询到销售订单"))
	}

	// 设置结账抹零规则
	saleOrder.SetCheckoutZeroingMethod(req.ZeroRule)

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	zeroAmount := infoResp.GetZeroAmount()

	// 发布“结账抹零”事件
	go func() {
		s.bus.PublishCheckoutZeroSaleOrderEvent(event.CheckoutZeroSaleOrderPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Operation:       constant.OrderCheckoutDiscountAdd,
			RoundingType:    req.ZeroRule,
			SpecialDiscount: zeroAmount,
		})
	}()

	return infoResp, nil
}

// InstantOrderSaleOrderCreate 给销售账单创建一个销售订单。（创建新拆单）
func (s *orderSrv) InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 加锁
	saleBillUuid := req.SaleBillUuid
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 如果销售账单目前只有一个销售订单，增加一个销售订单后要求撤销订单1的优惠折扣
	// 这是产品的特殊要求，可能后续会改。
	// 撤销订单的优惠折扣
	if len(saleBill.SaleOrders) == 1 {
		saleOrder := saleBill.GetFirstSaleOrder()
		// 撤销订单1的优惠折扣
		saleOrder.SetAllDiscountCancel()
	}

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建销售订单
		if _, errCreateSaleOrder := createSaleOrder(db, saleBill.SaleBillSetting, saleBill.Uuid, saleBill.OrderNo); errCreateSaleOrder != nil {
			return errors.WithMessage(errCreateSaleOrder, fmt.Sprintf("新建拆单失败,saleBill.Uuid:%v, saleBill.OrderNo:%v", saleBill.Uuid, saleBill.OrderNo))
		}

		// 计算并保存销售账单
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	cartInfo, errCartInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errCartInfo != nil {
		ctx.Log().Error("查询购物车信息失败", zap.Any("errCartInfo", errCartInfo))
		return nil, errors.WithMessage(errCartInfo, "查询购物车信息失败")
	}
	return cartInfo, nil
}

func MoreThanMoveNum(saleOrderProductNum, moveNum uint) bool {
	return saleOrderProductNum > moveNum
}

func LessThanMoveNum(saleOrderProduct *model.SaleOrderProduct, moveNum uint) bool {
	return saleOrderProduct.Num < moveNum
}

func EqualMoveNum(saleOrderProductNum, moveNum uint) bool {
	return saleOrderProductNum == moveNum
}
func IsSameSignature[T any](sign string, toSaleOrderProductSignMap map[string]*T) bool {
	return toSaleOrderProductSignMap[sign] != nil
}

func (s *orderSrv) CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error {
	// 计算订单商品、订单、账单
	saleBill.CalcAll(options...)
	// 保存到数据库
	if db == nil {
		db = s.dbm.GetDB(ctx.GetDbId())
	}
	err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		for _, saleOrder := range saleBill.SaleOrders {
			// 保存订单
			if len(saleBill.SaleOrders) == 1 {
				// 账单中只有一个订单时，只能更新订单，不能再创建订单。因为业务上默认就是一个账单一个订单，没有有账单而没有订单的情况
				if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
					return errors.WithMessage(err)
				}
			} else {
				// 有拆单时使用
				if err := repository.NewSaleOrderRepo(db).UpdateOrCreateSaleOrderRecord(*saleOrder); err != nil {
					return errors.WithMessage(err)
				}
			}
			// 保存订单商品
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				// 保存订单商品。只有标记更新的商品才会更新
				if err := repository.NewSaleOrderProductRepo(db).UpdateOrCreateSaleOrderProductRecord(*saleOrderProduct); err != nil {
					return errors.WithMessage(err)
				}
				for _, saleOrderProductBom := range saleOrderProduct.SaleOrderProductBoms {
					if saleOrderProductBom.GetUpdate() {
						if err := repository.NewOrderProductBomRepo(db).UpdateSaleOrderProductBomRecord(*saleOrderProductBom); err != nil {
							return errors.WithMessage(err)
						}
					}
				}
			}
			// 保存自助餐顾客
			for _, buffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if err := repository.NewSaleOrderBuffetCustomerTypeRepo(db).UpdateOrCreateSaleOrderBuffetCustomerTypeRecord(*buffetCustomer); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 保存账单。不能用这个方法来创建销售账单，故不使用UpdateOrCreate
		if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if err != nil {
		ctx.Log().Error("更新金额失败", zap.Error(err))
		return errors.WithMessage(err)
	}

	return nil
}

// 解析要移动的商品，识别为销售订单商品、顾客、加钟
func (s *orderSrv) getMoveProductInfo(ctx context.Context, saleOrderFrom *model.SaleOrder, req req.InstantOrderSaleOrderMoveProductReq) ([]*model.SaleOrderProduct, []*model.SaleOrderBuffetCustomerType, []*model.SaleOrderBuffetDelayProduct, error) {
	// 构建销售订单商品map
	saleOrderProductMap := make(map[uint64]*model.SaleOrderProduct)
	for index, saleOrderProduct := range saleOrderFrom.SaleOrderProducts {
		if saleOrderProduct.IsDelete() || saleOrderProduct.SaleOrderUuid != saleOrderFrom.Uuid {
			continue
		}
		saleOrderProductMap[saleOrderProduct.Uuid] = saleOrderFrom.SaleOrderProducts[index]
	}

	// 构建顾客map
	buffetCustomerMap := make(map[uint64]*model.SaleOrderBuffetCustomerType)
	for index, buffetCustomer := range saleOrderFrom.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() || buffetCustomer.SaleOrderUuid != saleOrderFrom.Uuid {
			continue
		}
		buffetCustomerMap[buffetCustomer.Uuid] = saleOrderFrom.SaleOrderBuffetCustomerTypes[index]
	}

	// 构建加钟map
	buffetDelayProductMap := make(map[uint64]*model.SaleOrderBuffetDelayProduct)
	for index, buffetDelayProduct := range saleOrderFrom.SaleOrderBuffetDelayProducts {
		if buffetDelayProduct.IsDelete() || buffetDelayProduct.SaleOrderUuid != saleOrderFrom.Uuid {
			continue
		}
		buffetDelayProductMap[buffetDelayProduct.Uuid] = saleOrderFrom.SaleOrderBuffetDelayProducts[index]
	}

	// 遍历要移动的商品，识别为销售订单商品、顾客、加钟
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)
	buffetCustomers := make([]*model.SaleOrderBuffetCustomerType, 0)
	buffetDelayProducts := make([]*model.SaleOrderBuffetDelayProduct, 0)
	for _, moveProduct := range req.Products {
		if saleOrderProduct, ok := saleOrderProductMap[moveProduct.Uuid]; ok {
			saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
			continue
		}
		if buffetCustomer, ok := buffetCustomerMap[moveProduct.Uuid]; ok {
			buffetCustomers = append(buffetCustomers, buffetCustomer)
			continue
		}
		if buffetDelayProduct, ok := buffetDelayProductMap[moveProduct.Uuid]; ok {
			buffetDelayProducts = append(buffetDelayProducts, buffetDelayProduct)
			continue
		}
		return nil, nil, nil, errors.WithMessage(errors.New("商品可能移动到其他销售订单中"), fmt.Sprintf("sale_order_product_uuid:%d", moveProduct.Uuid))
	}

	return saleOrderProducts, buffetCustomers, buffetDelayProducts, nil
}

// moveSaleOrderProduct 从一个销售订单移动商品到另一个销售订单
// 第一种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第二种移动方式：原销售订单商品数量小于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 第三种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 数据处理：
// 第一种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第二种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；新建目标销售订单商品，计算金额，表插入记录，数组增加这条记录，计算订单金额
// 第三种移动方式：删除原销售订单商品，更新表记录，重新计算原订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第四种移动方式：修改原销售订单商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；目标销售订单的商品数组增加这条记录，重新计算订单金额
func (s *orderSrv) moveSaleOrderProduct(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, saleOrderProducts []*model.SaleOrderProduct, moveNumMap map[uint64]uint) (map[uint64]*model.SaleOrderProduct, map[uint64]*model.SaleOrderProduct, error) {
	// 需要更新的销售订单商品
	waitUpdateSaleOrderProductMap := make(map[uint64]*model.SaleOrderProduct)
	// 需要新建的销售订单商品
	waitCreateSaleOrderProductMap := make(map[uint64]*model.SaleOrderProduct)

	// 构建目标销售订单中的商品签名map
	toSaleOrderProductSignMap := make(map[string]*model.SaleOrderProduct)
	for i, saleOrderProduct := range saleOrderTo.SaleOrderProducts {
		if saleOrderProduct.IsDelete() || saleOrderProduct.SaleOrderUuid != saleOrderTo.Uuid {
			continue
		}
		toSaleOrderProductSignMap[saleOrderProduct.Sign] = saleOrderTo.SaleOrderProducts[i]
	}

	// 遍历要移动的订单商品，移动到目标订单中
	for _, saleOrderProduct := range saleOrderProducts {
		ctx.Log().Debug("移动商品", zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
		moveProductNum, ok := moveNumMap[saleOrderProduct.Uuid]
		if !ok {
			return nil, nil, errors.WithMessage(errors.New("商品可能移动到其他销售订单中"), fmt.Sprintf("sale_order_product_uuid:%d", saleOrderProduct.Uuid))
		}
		if moveProductNum > saleOrderProduct.Num {
			return nil, nil, errors.WithMessage(errors.New("移动数量大于销售订单商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", saleOrderProduct.Uuid))
		}
		hasHandle := false // 是否已经处理过。因为一个商品被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
		// 第一种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中有签名一样的商品，该商品数量增加移动数量
		if !hasHandle && MoreThanMoveNum(saleOrderProduct.Num, moveProductNum) && IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第一种移动方式", zap.Any("from", saleOrderProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			// 修改原销售订单商品数量，更新记录，重新计算订单金额
			saleOrderProduct.Num -= moveProductNum
			// 修改目标销售订单商品数量，更新记录，重新计算订单金额
			toSaleOrderProductSignMap[saleOrderProduct.Sign].Num += moveProductNum
			// 记录到待更新列表中
			waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
			waitUpdateSaleOrderProductMap[toSaleOrderProductSignMap[saleOrderProduct.Sign].Uuid] = toSaleOrderProductSignMap[saleOrderProduct.Sign]
		}

		// 第二种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && MoreThanMoveNum(saleOrderProduct.Num, moveProductNum) && !IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第二种移动方式", zap.Any("from", saleOrderProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			ctx.Log().Debug("移动商品", zap.Any("原销售订单商品修改前数量", saleOrderProduct.Num))
			// 修改原销售订单商品数量，更新记录，重新计算订单金额
			saleOrderProduct.Num -= moveProductNum
			// 新建一个销售订单商品，该商品数量为移动数量
			newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderTo.Uuid)
			newSaleOrderProduct.Num = moveProductNum
			// 计算商品数据。折扣、税费、服务
			discountInfo := saleOrderTo.GetDiscountInfo()
			newSaleOrderProduct.SetDiscountInfo(discountInfo.MemberDiscountRate, discountInfo.MemberCardDiscountRate, discountInfo.CustomDiscountRate)
			newSaleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
			// 在目标销售订单中新建一个销售订单商品
			saleOrderTo.SaleOrderProducts = append(saleOrderTo.SaleOrderProducts, newSaleOrderProduct)
			// 记录到待更新列表中
			waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
			waitCreateSaleOrderProductMap[newSaleOrderProduct.Uuid] = newSaleOrderProduct
			ctx.Log().Debug("移动商品", zap.Any("原销售订单商品数量", saleOrderProduct.Num), zap.Any("目标销售订单商品数量", newSaleOrderProduct.Num))
		}

		// 第三种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
		if !hasHandle && EqualMoveNum(saleOrderProduct.Num, moveProductNum) && IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第三种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			// 删除原销售订单商品，更新表记录，重新计算原订单金额；
			saleOrderProduct.DeleteTime = time.Now().Unix()
			// 修改目标销售订单商品数量，更新记录，重新计算订单金额
			toSaleOrderProductSignMap[saleOrderProduct.Sign].Num += moveProductNum
			// 记录到待更新列表中
			waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
			waitUpdateSaleOrderProductMap[toSaleOrderProductSignMap[saleOrderProduct.Sign].Uuid] = toSaleOrderProductSignMap[saleOrderProduct.Sign]
		}

		// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && EqualMoveNum(saleOrderProduct.Num, moveProductNum) && !IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第四种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			// 修改原销售订单商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；
			discountInfo := saleOrderTo.GetDiscountInfo()
			saleOrderProduct.SaleOrderUuid = saleOrderTo.Uuid
			saleOrderProduct.SetDiscountInfo(discountInfo.MemberDiscountRate, discountInfo.MemberCardDiscountRate, discountInfo.CustomDiscountRate)
			// 计算商品数据。折扣、税费、服务
			saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
			// 目标销售订单的商品数组增加这条记录，重新计算订单金额
			saleOrderTo.SaleOrderProducts = append(saleOrderTo.SaleOrderProducts, saleOrderProduct)
			// 记录到待更新列表中
			waitUpdateSaleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
		}
	}

	return waitUpdateSaleOrderProductMap, waitCreateSaleOrderProductMap, nil
}

// moveBuffetCustomer 移动自助餐顾客
func (s *orderSrv) moveBuffetCustomer(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, buffetCustomers []*model.SaleOrderBuffetCustomerType, moveNumMap map[uint64]uint) (map[uint64]*model.SaleOrderBuffetCustomerType, map[uint64]*model.SaleOrderBuffetCustomerType, error) {
	// 需要更新的销售订单顾客
	waitUpdateBuffetCustomerMap := make(map[uint64]*model.SaleOrderBuffetCustomerType)
	// 需要新建的销售订单顾客
	waitCreateBuffetCustomerMap := make(map[uint64]*model.SaleOrderBuffetCustomerType)

	toBuffetCustomerSignMap := make(map[string]*model.SaleOrderBuffetCustomerType)
	for i, buffetCustomer := range saleOrderTo.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() || buffetCustomer.SaleOrderUuid != saleOrderTo.Uuid {
			continue
		}
		toBuffetCustomerSignMap[buffetCustomer.GetSign()] = saleOrderTo.SaleOrderBuffetCustomerTypes[i]
	}

	// 遍历要移动的订单顾客，移动到目标订单中
	for _, buffetCustomer := range buffetCustomers {
		ctx.Log().Debug("移动顾客", zap.Any("buffetCustomer", buffetCustomer.Name))
		moveCustomerNum, ok := moveNumMap[buffetCustomer.Uuid]
		if !ok {
			return nil, nil, errors.WithMessage(errors.New("顾客可能移动到其他销售订单中"), fmt.Sprintf("buffetCustomer_uuid:%d", buffetCustomer.Uuid))
		}
		if moveCustomerNum > buffetCustomer.Num {
			return nil, nil, errors.WithMessage(errors.New("移动数量大于销售订单商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", buffetCustomer.Uuid))
		}
		hasHandle := false // 是否已经处理过。因为一个顾客被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
		// 第一种移动方式：原销售订单顾客数量大于移动数量，则原销售订单顾客数量减少移动数量，目标销售订单中有签名一样的顾客，该顾客数量增加移动数量
		if !hasHandle && MoreThanMoveNum(buffetCustomer.Num, moveCustomerNum) && IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动顾客，第一种移动方式", zap.Any("from", buffetCustomer.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			// 修改原销售订单顾客数量，更新记录，重新计算订单金额
			buffetCustomer.Num -= moveCustomerNum
			// 修改目标销售订单顾客数量，更新记录，重新计算订单金额
			toBuffetCustomerSignMap[buffetCustomer.GetSign()].Num += moveCustomerNum
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
			waitUpdateBuffetCustomerMap[toBuffetCustomerSignMap[buffetCustomer.GetSign()].Uuid] = toBuffetCustomerSignMap[buffetCustomer.GetSign()]
		}

		// 第二种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && MoreThanMoveNum(buffetCustomer.Num, moveCustomerNum) && !IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动顾客，第二种移动方式", zap.Any("from", buffetCustomer.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			ctx.Log().Debug("移动顾客", zap.Any("原销售订单商品修改前数量", buffetCustomer.Num))
			// 修改原销售订单商品数量，更新记录，重新计算订单金额
			buffetCustomer.Num -= moveCustomerNum
			// 新建一个销售订单商品，该商品数量为移动数量
			newBuffetCustomer := buffetCustomer.CopyBuffetCustomer(saleOrderTo.Uuid)
			newBuffetCustomer.Num = moveCustomerNum
			// 计算商品数据。折扣、税费、服务
			newBuffetCustomer.CustomDiscountRate = saleOrderTo.CustomDiscountRate
			newBuffetCustomer.CalcSaleOrderBuffetCustomerType(*saleBill.SaleBillSetting)
			// 在目标销售订单中新建一个销售订单商品
			saleOrderTo.SaleOrderBuffetCustomerTypes = append(saleOrderTo.SaleOrderBuffetCustomerTypes, newBuffetCustomer)
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
			waitCreateBuffetCustomerMap[newBuffetCustomer.Uuid] = newBuffetCustomer
			ctx.Log().Debug("移动商品", zap.Any("原销售订单商品数量", buffetCustomer.Num), zap.Any("目标销售订单商品数量", newBuffetCustomer.Num))
		}

		// 第三种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
		if !hasHandle && EqualMoveNum(buffetCustomer.Num, moveCustomerNum) && IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第三种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			// 删除原销售订单商品，更新表记录，重新计算原订单金额；
			buffetCustomer.DeleteTime = time.Now().Unix()
			// 修改目标销售订单商品数量，更新记录，重新计算订单金额
			toBuffetCustomerSignMap[buffetCustomer.GetSign()].Num += moveCustomerNum
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
			waitUpdateBuffetCustomerMap[toBuffetCustomerSignMap[buffetCustomer.GetSign()].Uuid] = toBuffetCustomerSignMap[buffetCustomer.GetSign()]
		}

		// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && EqualMoveNum(buffetCustomer.Num, moveCustomerNum) && !IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第四种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			// 修改原销售订单商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；
			buffetCustomer.SaleOrderUuid = saleOrderTo.Uuid
			buffetCustomer.CustomDiscountRate = saleOrderTo.CustomDiscountRate
			// 计算商品数据。折扣、税费、服务
			buffetCustomer.CalcSaleOrderBuffetCustomerType(*saleBill.SaleBillSetting)
			// 目标销售订单的商品数组增加这条记录，重新计算订单金额
			saleOrderTo.SaleOrderBuffetCustomerTypes = append(saleOrderTo.SaleOrderBuffetCustomerTypes, buffetCustomer)
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
		}
	}

	return waitUpdateBuffetCustomerMap, waitCreateBuffetCustomerMap, nil
}

// moveBuffetCustomer 移动加钟商品
func (s *orderSrv) moveBuffetDelayProduct(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, delayProducts []*model.SaleOrderBuffetDelayProduct, moveNumMap map[uint64]uint) (map[uint64]*model.SaleOrderBuffetDelayProduct, map[uint64]*model.SaleOrderBuffetDelayProduct, error) {
	// 需要更新的销售订单加钟商品
	waitUpdateBuffetDelayProductMap := make(map[uint64]*model.SaleOrderBuffetDelayProduct)
	// 需要新建的销售订单加钟商品
	waitCreateBuffetDelayProductMap := make(map[uint64]*model.SaleOrderBuffetDelayProduct)

	toBuffetDelayProductSignMap := make(map[string]*model.SaleOrderBuffetDelayProduct)
	for i, buffetDelayProduct := range saleOrderTo.SaleOrderBuffetDelayProducts {
		if buffetDelayProduct.IsDelete() || buffetDelayProduct.SaleOrderUuid != saleOrderTo.Uuid {
			continue
		}
		toBuffetDelayProductSignMap[buffetDelayProduct.GetSign()] = saleOrderTo.SaleOrderBuffetDelayProducts[i]
	}

	// 遍历要移动的订单加钟商品，移动到目标订单中
	for _, delayProduct := range delayProducts {
		ctx.Log().Debug("移动加钟商品", zap.Any("delayProduct", delayProduct.Name))
		moveCustomerNum, ok := moveNumMap[delayProduct.Uuid]
		if !ok {
			return nil, nil, errors.WithMessage(errors.New("加钟商品可能移动到其他销售订单中"), fmt.Sprintf("buffetCustomer_uuid:%d", delayProduct.Uuid))
		}
		if moveCustomerNum > delayProduct.Num {
			return nil, nil, errors.WithMessage(errors.New("移动数量大于加钟商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", delayProduct.Uuid))
		}

		hasHandle := false // 是否已经处理过。因为一个加钟商品被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
		// 第一种移动方式：原销售订单加钟商品数量大于移动数量，则原销售订单加钟商品数量减少移动数量，目标销售订单中有签名一样的加钟商品，该加钟商品数量增加移动数量
		if !hasHandle && MoreThanMoveNum(delayProduct.Num, moveCustomerNum) && IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第一种移动方式", zap.Any("from", delayProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			// 修改原销售订单加钟商品数量，更新记录，重新计算订单金额
			delayProduct.Num -= moveCustomerNum
			// 修改目标销售订单加钟商品数量，更新记录，重新计算订单金额
			toBuffetDelayProductSignMap[delayProduct.GetSign()].Num += moveCustomerNum
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitUpdateBuffetDelayProductMap[toBuffetDelayProductSignMap[delayProduct.GetSign()].Uuid] = toBuffetDelayProductSignMap[delayProduct.GetSign()]
		}

		// 第二种移动方式：原加钟商品数量大于移动数量，则原加钟商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个加钟商品，该商品数量为移动数量
		if !hasHandle && MoreThanMoveNum(delayProduct.Num, moveCustomerNum) && !IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第二种移动方式", zap.Any("from", delayProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			ctx.Log().Debug("移动加钟商品", zap.Any("原加钟商品修改前数量", delayProduct.Num))
			// 修改原加钟商品数量，更新记录，重新计算订单金额
			delayProduct.Num -= moveCustomerNum
			// 新建一个加钟商品，该商品数量为移动数量
			newBuffetCustomer := delayProduct.CopyBuffetDelayProduct(saleOrderTo.Uuid)
			newBuffetCustomer.Num = moveCustomerNum
			// 在目标销售订单中新建一个加钟商品
			saleOrderTo.SaleOrderBuffetDelayProducts = append(saleOrderTo.SaleOrderBuffetDelayProducts, newBuffetCustomer)
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitCreateBuffetDelayProductMap[newBuffetCustomer.Uuid] = newBuffetCustomer
			ctx.Log().Debug("移动加钟商品", zap.Any("原加钟商品数量", delayProduct.Num), zap.Any("目标加钟商品数量", newBuffetCustomer.Num))
		}

		// 第三种移动方式：原加钟商品数量等于移动数量，则原加钟商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
		if !hasHandle && EqualMoveNum(delayProduct.Num, moveCustomerNum) && IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第三种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			// 删除原加钟商品，更新表记录，重新计算原订单金额；
			delayProduct.DeleteTime = time.Now().Unix()
			// 修改目标加钟商品数量，更新记录，重新计算订单金额
			toBuffetDelayProductSignMap[delayProduct.GetSign()].Num += moveCustomerNum
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitUpdateBuffetDelayProductMap[toBuffetDelayProductSignMap[delayProduct.GetSign()].Uuid] = toBuffetDelayProductSignMap[delayProduct.GetSign()]
		}

		// 第四种移动方式：原加钟商品数量等于移动数量，则原加钟商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个加钟商品，该商品数量为移动数量
		if !hasHandle && EqualMoveNum(delayProduct.Num, moveCustomerNum) && !IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第四种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			// 修改原加钟商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；
			delayProduct.SaleOrderUuid = saleOrderTo.Uuid
			// 目标销售订单的商品数组增加这条记录，重新计算订单金额
			saleOrderTo.SaleOrderBuffetDelayProducts = append(saleOrderTo.SaleOrderBuffetDelayProducts, delayProduct)
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
		}
	}

	return waitUpdateBuffetDelayProductMap, waitCreateBuffetDelayProductMap, nil
}

// InstantOrderSaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
// 第一种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第二种移动方式：原销售订单商品数量小于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 第三种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 数据处理：
// 第一种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第二种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；新建目标销售订单商品，计算金额，表插入记录，数组增加这条记录，计算订单金额
// 第三种移动方式：删除原销售订单商品，更新表记录，重新计算原订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第四种移动方式：修改原销售订单商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；目标销售订单的商品数组增加这条记录，重新计算订单金额
func (s *orderSrv) SaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq, needDeleteSaleOrder bool) (*resp.ShopCart, error) {
	saleBillUuid := req.SaleBillUuid
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 获取销售订单信息
	saleOrderFrom := saleBill.GetSaleOrder(req.From)
	saleOrderTo := saleBill.GetSaleOrder(req.To)

	// 构建移动到订单商品的map结构
	moveProductMap := make(map[uint64]uint)
	for _, moveProduct := range req.Products {
		moveProductMap[moveProduct.Uuid] = moveProduct.Num
	}

	saleOrderProducts, saleOrderBuffetCustomers, buffetDelayProducts, err := s.getMoveProductInfo(ctx, saleOrderFrom, req)

	moveNumMap := make(map[uint64]uint)
	for _, moveProduct := range req.Products {
		moveNumMap[moveProduct.Uuid] = moveProduct.Num
	}

	waitUpdateSaleOrderProductMap, waitCreateSaleOrderProductMap, err := s.moveSaleOrderProduct(ctx, saleBill, saleOrderFrom, saleOrderTo, saleOrderProducts, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	waitUpdateBuffetCustomerMap, waitCreateBuffetCustomerMap, err := s.moveBuffetCustomer(ctx, saleBill, saleOrderFrom, saleOrderTo, saleOrderBuffetCustomers, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	waitUpdateBuffetDelayProductMap, waitCreateBuffetDelayProductMap, err := s.moveBuffetDelayProduct(ctx, saleBill, saleOrderFrom, saleOrderTo, buffetDelayProducts, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if len(waitUpdateSaleOrderProductMap) == 0 && len(waitUpdateBuffetCustomerMap) == 0 && len(waitUpdateBuffetDelayProductMap) == 0 {
		ctx.Log().Debug("移动商品失败，没有需要更新的销售订单商品")
		return nil, errors.WithMessage(errors.New("移动商品失败"))
	}

	// 计算订单金额
	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderTo calc", saleOrderTo.BeforeCalc()))
	afterSaleOrderCalc := saleOrderTo.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderTo calc", afterSaleOrderCalc))

	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderFrom calc", saleOrderFrom.BeforeCalc()))
	afterSaleOrderFromCalc := saleOrderFrom.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderFrom calc", afterSaleOrderFromCalc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	var cartInfo *resp.ShopCart
	errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		ctx.Log().Debug("更新销售订单商品", zap.Any("waitUpdateSaleOrderProductMap len", len(waitUpdateSaleOrderProductMap)))
		for _, saleOrderProduct := range waitUpdateSaleOrderProductMap {
			ctx.Log().Debug("更新销售订单商品", zap.Any("saleOrderProduct saleOrder uuid", saleOrderProduct.SaleOrderUuid), zap.Any("saleOrderProduct uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))

			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 创建销售订单商品及BOM、属性
		for _, saleOrderProduct := range waitCreateSaleOrderProductMap {
			ctx.Log().Debug("新建销售订单商品", zap.Any("saleOrderProduct saleOrder uuid", saleOrderProduct.SaleOrderUuid), zap.Any("saleOrderProduct uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			if _, err := repository.NewSaleOrderProductRepo(tx).CreateSaleOrderProductAndBomAndAttribute(*saleOrderProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐顾客
		for _, buffetCustomer := range waitUpdateBuffetCustomerMap {
			if err := repository.NewSaleOrderBuffetCustomerTypeRepo(tx).UpdateSaleOrderBuffetCustomerTypeRecord(*buffetCustomer); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 创建自助餐顾客
		for _, buffetCustomer := range waitCreateBuffetCustomerMap {
			if err := repository.NewSaleOrderBuffetCustomerTypeRepo(tx).CreateSaleOrderBuffetCustomerTypeRecord(*buffetCustomer); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐加钟商品
		for _, buffetDelayProduct := range waitUpdateBuffetDelayProductMap {
			if err := repository.NewOrderRepo(tx).UpdateSaleOrderBuffetDelayProductRecord(*buffetDelayProduct); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 创建自助餐加钟商品
		for _, buffetDelayProduct := range waitCreateBuffetDelayProductMap {
			if _, err := repository.NewOrderRepo(tx).CreateSaleOrderBuffetDelayProduct(*buffetDelayProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 当删除拆单时. needDeleteSaleOrder使用场景：1.删除某个子单，移动完商品后，需要删除该子单；2.撤销拆单，移动完商品后，需要删除所有子单
		if needDeleteSaleOrder {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				return errors.WithMessage(err)
			}
		} else {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrderFrom); err != nil {
				return errors.WithMessage(err)
			}
		}
		if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrderTo); err != nil {
			return errors.WithMessage(err)
		}

		// 更新账单
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	})
	if errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	cartInfo = info

	return cartInfo, nil
}

// InstantOrderMustPlanConfirm 确认必点商品
func (s *orderSrv) InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq) (bool, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return false, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	// 查询到购物车信息
	shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("获取购物车信息失败", zap.Error(err))
		return false, errors.WithMessage(err, "获取购物车信息失败")
	}

	var mustPlanList []resp.InstantProductMustPlan
	if saleBill.IsDeskSaleBill() {
		mustPlan, errMustPlan := s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), saleBill.DeskUuid)
		if errMustPlan != nil {
			return false, errMustPlan
		}
		mustPlanList = mustPlan
	} else {
		mustPlan, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
		if errMustPlan != nil {
			return false, errMustPlan
		}
		mustPlanList = mustPlan
	}

	if mustPlanList != nil && len(mustPlanList) > 0 {
		for _, plan := range mustPlanList {
			if plan.NeedNum > 0 {
				ctx.Log().Info("确认必点商品失败，必点商品未点", zap.Any("plan name", plan.Name))
				return false, nil
			}
		}
	}

	// 修改sale_bill表的show_must_plan
	saleBill.ShowMustPlan = constant.SaleBillShowMustPlanNo
	ctx.Log().Debug("修改sale_bill表的show_must_plan", zap.Any("saleBill.ShowMustPlan", saleBill.ShowMustPlan))
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(req.SaleBillUuid); err != nil {
		ctx.Log().Error("修改sale_bill表的show_must_plan失败", zap.Error(err))
		return false, errors.WithMessage(err, "确认必点商品失败")
	}

	return true, nil
}

// InstantOrderCheck 订单检查
func (s *orderSrv) OrderCheck(ctx context.Context, req req.InstantOrderCheckReq) (*resp.OrderCheckServiceRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	ctx.Log().Debug("获取销售账单信息")

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()

	// 获取所有商品,用于检查限购
	saleOrderProductAll := saleBill.GetSaleOrderProductAll()

	// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
	var deskUuid uint64
	if saleBill.IsDeskSaleBill() {
		deskUuid = saleBill.DeskUuid
	}
	checkServiceRes, errCheck := s.checkOrder(ctx, req.IgnoreMust, db, req.SaleBillUuid, deskUuid, unCookingSaleOrderProducts, saleOrderProductAll)
	if errCheck != nil {
		return nil, errors.WithMessage(errCheck, "订单检查失败")
	}
	if checkServiceRes != nil {
		return checkServiceRes, nil
	}

	// 检查是否有未送厨的商品
	if len(unCookingSaleOrderProducts) > 0 {
		products := make([]resp.Product, 0)
		for _, product := range unCookingSaleOrderProducts {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.MultiLanguageName.GetNames(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.DiscountFee,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsCancel:      product.IsCancelProduct(),
			})
		}
		res := &resp.OrderCheckServiceRes{
			Code:          constant.CodeOrderCheckProductUnCooking,
			OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: products}},
		}
		return res, nil
	}

	return nil, nil
}

// getMoveProductList 获取删除某个子单要移动的商品列表,包括订单商品、订单顾客、订单加钟商品
func (s *orderSrv) getMoveProductList(saleOrderFrom *model.SaleOrder) []req.MoveProduct {
	moveProductList := make([]req.MoveProduct, 0) // 移动商品列表,包括订单商品、订单顾客、订单加钟商品
	// 获取要移动的商品
	for _, saleOrderProduct := range saleOrderFrom.SaleOrderProducts {
		if saleOrderProduct.IsDelete() || saleOrderProduct.Num == 0 {
			continue
		}
		moveProductList = append(moveProductList, req.MoveProduct{
			Uuid: saleOrderProduct.Uuid,
			Num:  saleOrderProduct.Num,
		})
	}
	// 获取要移动的顾客
	for _, saleOrderBuffetCustomer := range saleOrderFrom.SaleOrderBuffetCustomerTypes {
		if saleOrderBuffetCustomer.IsDelete() || saleOrderBuffetCustomer.Num == 0 {
			continue
		}
		moveProductList = append(moveProductList, req.MoveProduct{
			Uuid: saleOrderBuffetCustomer.Uuid,
			Num:  saleOrderBuffetCustomer.Num,
		})
	}
	// 获取要移动的加钟商品
	for _, buffetDelayProduct := range saleOrderFrom.SaleOrderBuffetDelayProducts {
		if buffetDelayProduct.IsDelete() || buffetDelayProduct.Num == 0 {
			continue
		}
		moveProductList = append(moveProductList, req.MoveProduct{
			Uuid: buffetDelayProduct.Uuid,
			Num:  buffetDelayProduct.Num,
		})
	}
	return moveProductList
}

// InstantOrderSaleOrderDelete 删除一个销售订单(删除拆单)
func (s *orderSrv) InstantOrderSaleOrderDelete(ctx context.Context, request req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		// 加锁
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}
	ctx.Log().Debug("删除一个销售订单(删除拆单)", zap.Any("request", request))
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return nil, errors.New("获取销售账单信息失败")
	}

	// 不能删除第一个销售订单
	if len(saleBill.SaleOrders) > 0 {
		if saleBill.SaleOrders[0].Uuid == request.SaleOrderUuid {
			return nil, errors.New("不能删除第一个销售订单")
		}
	}

	firstSaleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)

	saleOrderFrom := saleBill.GetSaleOrder(request.SaleOrderUuid)

	// 获取要移动的商品列表
	moveProductList := s.getMoveProductList(saleOrderFrom)

	moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
		SaleBillUuid: request.SaleBillUuid,
		From:         request.SaleOrderUuid,
		To:           firstSaleOrder.Uuid,
		Products:     moveProductList,
	}

	if len(moveProductList) > 0 {
		shopCart, err := s.SaleOrderMoveProduct(ctx, moveProductReq, true)
		if err != nil {
			ctx.Log().Error("移动商品失败", zap.Error(err))
			return nil, errors.WithMessage(err)
		}
		return shopCart, nil
	}

	// 如果销售订单中没有商品，则直接删除订单
	if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
		ctx.Log().Error("删除订单失败", zap.Error(err))
		return nil, errors.New("删除订单失败")
	}

	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("获取购物车信息失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取购物车信息失败")
	}

	return info, nil

}

// InstantOrderSaleOrderDeleteAll 删除所有子销售订单(撤销拆单)
func (s *orderSrv) InstantOrderSaleOrderDeleteAll(ctx context.Context, request req.InstantOrderSaleOrderDeleteAllReq) (*resp.ShopCart, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 将除了第一个销售订单的所有商品都移动到第一个销售订单里

	ctx.Log().Debug("删除所有子销售订单(撤销拆单)", zap.Any("request", request))

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	firstSaleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)

	saleOrderFromList := make([]*model.SaleOrder, 0)
	for _, saleOrder := range saleBill.SaleOrders {
		if saleOrder.Uuid == firstSaleOrder.Uuid {
			continue
		}
		saleOrderFromList = append(saleOrderFromList, saleOrder)
	}

	for _, saleOrderFrom := range saleOrderFromList {
		moveProductList := s.getMoveProductList(saleOrderFrom)
		moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
			SaleBillUuid: request.SaleBillUuid,
			From:         saleOrderFrom.Uuid,
			To:           firstSaleOrder.Uuid,
			Products:     moveProductList,
		}

		if len(moveProductList) > 0 {
			// todo 优化减少重复查询
			_, err := s.SaleOrderMoveProduct(ctx, moveProductReq, true)
			if err != nil {
				ctx.Log().Error("移动商品失败", zap.Error(err))
				return nil, errors.WithMessage(err)
			}
		} else {
			// 如果销售订单中没有商品，则直接删除订单
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				ctx.Log().Error("删除订单失败", zap.Error(err))
				return nil, errors.WithMessage(err, "删除订单失败")
			}
		}
	}

	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("获取购物车信息失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取购物车信息失败")
	}

	return info, nil
}

// OrderMemberCancel 不使用此会员
func (s *orderSrv) OrderMemberCancel(ctx context.Context, request req.OrderMemberCancelReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	saleOrder.SetMemberDiscountCancel()

	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return infoResp, nil
}

// OrderUseMember 使用会员优惠
func (s *orderSrv) OrderUseMember(ctx context.Context, request req.CheckMemberPasswordReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取会员信息
	member, errMember := repository.NewMemberRepo(db).GetMemberInfoForSaleOrder(ctx, request.MemberUuid)
	if errMember != nil {
		return nil, errMember
	}

	// 如果会员有密码的话，验证会员密码
	if member.HasPassword() {
		md5Password := cryptor.Md5String(request.Password)
		ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
		if member.Password != md5Password {
			ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
			return nil, errors.New("密码错误")
		}
	}

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	saleOrder.SetMemberDiscount(*member)

	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return infoResp, nil
}

// OrderPrint 打印
func (s *orderSrv) OrderPrint(ctx context.Context, request req.OrderPrintReq) (*resp.PrinterData, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 更新用户
	saleOrder.CashierUuid = ctx.GetStaffUuid()
	saleOrder.CashierName = ctx.GetStaff().RealName
	if saleOrder.CashierName == "" {
		saleOrder.CashierName = ctx.GetStaff().Username
	}
	if err := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断是否已支付
	printType := constant.PrinterTemplatePreBilling
	firstExecution := 1
	if saleOrder.IsPaid() {
		printType = constant.PrinterTemplateBilling
		firstExecution = 0
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, request.PrintLang).PrintingStatementOrder(
		printType,
		saleBill,
		saleOrder.Uuid,
		firstExecution,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 保存销售账单
	if err := repository.NewOrderRepo(db).SetLock(saleBill.Uuid, true); err != nil {
		return nil, errors.WithMessage(err, "设置锁单失败")
	}

	return printerData, nil
}

// OrderPrintInvoice 打印发票
func (s *orderSrv) OrderPrintInvoice(ctx context.Context, req req.OrderPrintInvoiceReq) (*resp.PrinterData, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 更新用户
	saleOrder.CashierUuid = ctx.GetStaffUuid()
	saleOrder.CashierName = ctx.GetStaff().RealName
	if saleOrder.CashierName == "" {
		saleOrder.CashierName = ctx.GetStaff().Username
	}
	if err := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 设置发票信息
	if req.CompanyName != "" {
		// 创建发票信息对象
		invoiceInfo := model.SaleOrderInvoiceInfo{
			SaleOrderUuid:    saleOrder.Uuid,
			CompanyName:      req.CompanyName,
			CompanyAddr:      req.CompanyAddr,
			CompanyTaxNumber: req.CompanyTaxNumber,
			CompanyPhone:     req.CompanyPhone,
		}
		// 保存发票信息（不存在则创建，存在则更新）
		invoiceInfos, err := repository.NewOrderRepo(db).SaveOrUpdateInvoiceInfo(saleOrder.Uuid, invoiceInfo)
		if err != nil {
			return nil, errors.WithMessage(err, "保存发票信息失败")
		}
		// 更新内存中的发票信息
		saleOrder.InvoiceInfo = invoiceInfos
	} else {
		invoiceInfo, err := repository.NewOrderRepo(db).GetInvoiceInfo(saleOrder.Uuid)
		if err != nil {
			saleOrder.InvoiceInfo = &model.SaleOrderInvoiceInfo{}
		} else {
			saleOrder.InvoiceInfo = invoiceInfo
		}
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, req.PrintLang).PrintingInvoice(
		saleBill,
		saleOrder.Uuid,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

// OrderPrintInvoiceInfo 获取发票信息
func (s *orderSrv) OrderPrintInvoiceInfo(ctx context.Context, req req.OrderInvoiceInfoReq) resp.SaleOrderInvoiceInfo {
	db := s.dbm.GetDB(ctx.GetDbId())
	invoiceInfo, err := repository.NewOrderRepo(db).GetInvoiceInfo(req.SaleOrderUuid)
	if err != nil {
		return resp.SaleOrderInvoiceInfo{}
	}
	return resp.SaleOrderInvoiceInfo{
		CompanyName:      invoiceInfo.CompanyName,
		CompanyAddr:      invoiceInfo.CompanyAddr,
		CompanyTaxNumber: invoiceInfo.CompanyTaxNumber,
		CompanyPhone:     invoiceInfo.CompanyPhone,
	}
}

// OrderUnlock 订单解锁
func (s *orderSrv) OrderUnlock(ctx context.Context, saleBillUuid uint64) error {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "查询销售账单失败")
	}

	// 验证订单是否可操作
	if err := saleBill.ValidateOrderStatus(constant.OrderUnlock); err != nil {
		return errors.WithMessage(err)
	}

	// 验证销售账单是否已锁定
	if !saleBill.IsLockStatus() {
		return errors.New("销售账单未锁定")
	}

	// 保存销售账单
	if err := repository.NewOrderRepo(db).SetLock(saleBill.Uuid, false); err != nil {
		return errors.WithMessage(err, "解锁失败")
	}

	return nil
}

// GetMustPlanList 点餐助手、平板端获取必点商品方案列表
func (s *orderSrv) GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error) {
	saleBillRepo := repository.NewSaleBillRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.ErrInternal
	}
	list, err := s.mustPlanSrv.GetDeskMustPlanList(ctx, saleBill.MealNum, make(map[uint64]map[uint64]uint), saleBill.DeskUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.WithMessage(err)
	}
	return resp.ProductMustPlanList{
		List: list,
	}, nil
}

// GetUnOrderedH5ProductList 获取扫码h5购物车未下单商品列表
func (s *orderSrv) GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error) {
	res, err := s.GetUnsentKitchen(ctx, saleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 重新计算商品金额。商品金额=商品列表中各个商品的金额之和
	productAmount := decimal.NewFromFloat(0)
	for _, product := range res.Products.List {
		productAmount = productAmount.Add(decimal.NewFromFloat(product.GetPrice()))
	}
	res.AmountInfo.ProductAmount = productAmount.InexactFloat64()

	return &res, nil
}

// GetOrderedH5ProductList 获取扫码h5购物车已下单商品列表
func (s *orderSrv) GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error) {
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(errors.ErrInternal, "获取点餐购物车信息: "+err.Error())
	}
	productGroup := make(map[int64][]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			productGroup[product.SendKitchenTime] = append(productGroup[product.SendKitchenTime], product)
		}
	}
	groups := make([]resp.H5Group, 0)
	for sendKitchenTime, products := range productGroup {
		groups = append(groups, resp.H5Group{
			SentKitchenProductGroup: resp.SentKitchenProductGroup{
				SendKitchenTime: sendKitchenTime,
				Products: resp.GroupProductList{
					List: products,
				},
			},
			AcceptTime: products[0].AcceptTime,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SendKitchenTime > groups[j].SendKitchenTime
	})

	return &resp.H5CartSendProduct{
		Groups: resp.H5GroupList{List: groups},
	}, nil
}

// ConfirmH5Order 下单扫码h5订单
func (s *orderSrv) ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(constant.OrderH5Confirm, saleOrderUuid); err != nil {
		return errors.WithMessage(err)
	}

	h5Order := saleBill.NewH5Order()
	// 获取未下单的h5订单商品
	h5OrderProducts := saleBill.GetUnOrderH5OrderProduct()
	// 将未下单的h5订单商品变为已下单的h5订单商品
	for _, h5OrderProduct := range h5OrderProducts {
		h5OrderProduct.SetH5OrderProduct(h5Order.Uuid)
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 创建h5订单
		if _, err := repository.NewH5OrderRepo(tx).CreateH5Order(*h5Order); err != nil {
			return errors.WithMessage(err, "保存h5订单失败")
		}
		// 创建h5订单商品
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			h5OrderProduct.H5OrderUuid = h5Order.Uuid
			if _, err := repository.NewH5OrderRepo(tx).CreateH5OrderProduct(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "保存h5订单商品失败")
			}
		}
		// 更新销售订单商品，将未下单商品改为已下单商品。记录上saleOrderProduct的h5OrderUuid
		for _, saleOrderProduct := range h5OrderProducts {
			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err, "更新销售订单商品失败")
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err, "下单扫码h5订单失败")
	}
	return nil
}

// GetUnsentKitchen 未送厨商品列表
func (s *orderSrv) GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error) {
	res := resp.UnsentKitchen{
		Products:   resp.CartProductList{List: make([]resp.Product, 0)},
		AmountInfo: resp.SimpleAmountInfo{},
	}
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
	if err != nil {
		return res, errors.WithMessage(errors.ErrInternal, "获取点餐购物车信息: "+err.Error())
	}
	signProduct := make(map[string]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			// 未送厨，且不是赠菜
			if product.Status == constant.SaleOrderProductStatusNormal && !product.IsGift {
				if p, exists := signProduct[product.Sign]; exists {
					product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
					product.Num = p.Num + product.Num
				}
				signProduct[product.Sign] = product
				res.AmountInfo.ProductNum = res.AmountInfo.ProductNum + product.Num
			}
		}
	}
	for _, product := range signProduct {
		res.Products.List = append(res.Products.List, product)
	}
	sort.Slice(res.Products.List, func(i, j int) bool {
		return res.Products.List[i].SendKitchenTime < res.Products.List[j].SendKitchenTime
	})
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return res, errors.WithMessage(errors.ErrInternal, "获取销售账单所有信息: "+err.Error())
	}
	for _, order := range saleBill.SaleOrders {
		res.AmountInfo.ProductAmount = utils.DecimalAdd(res.AmountInfo.ProductAmount, order.GetUnCookingProductAmount())
	}
	return res, nil
}

// GetSentKitchen 已送厨商品列表
func (s *orderSrv) GetSentKitchen(ctx context.Context, saleBillUuid uint64) (resp.SentKitchen, error) {
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		return resp.SentKitchen{}, errors.WithMessage(errors.ErrInternal, "获取点餐购物车信息: "+err.Error())
	}
	var productNum uint
	productGroup := make(map[int64][]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			if product.Status == constant.SaleOrderProductStatusCooking {
				productNum = productNum + product.Num
				productGroup[product.SendKitchenTime] = append(productGroup[product.SendKitchenTime], product)
			}
		}
	}

	groups := make([]resp.SentKitchenProductGroup, 0, len(productGroup))
	for sendKitchenTime, products := range productGroup {
		// 分组内的商品合并
		signProduct := make(map[string]resp.Product)
		for _, product := range products {
			if p, exists := signProduct[product.Sign]; exists {
				product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
				product.Num = p.Num + product.Num
			}
			signProduct[product.Sign] = product
		}
		var mergeProducts []resp.Product
		for _, product := range signProduct {
			mergeProducts = append(mergeProducts, product)
		}

		groups = append(groups, resp.SentKitchenProductGroup{
			SendKitchenTime: sendKitchenTime,
			Products: resp.GroupProductList{
				List: mergeProducts,
			},
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SendKitchenTime > groups[j].SendKitchenTime
	})
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return resp.SentKitchen{}, errors.WithMessage(errors.ErrInternal, "获取销售账单所有信息: "+err.Error())
	}
	var amount resp.AmountInfo
	for _, order := range saleBill.SaleOrders {
		calc := order.CalcCookingSaleOrder(*saleBill.SaleBillSetting)
		amount.ProductOriginalAmount = utils.DecimalAdd(amount.ProductOriginalAmount, calc.ProductOriginalAmount)
		amount.ProductAmount = utils.DecimalAdd(amount.ProductAmount, calc.ProductAmount)
		amount.ServiceAmount = utils.DecimalAdd(amount.ServiceAmount, calc.ServiceFee)
		amount.TaxAmount = utils.DecimalAdd(amount.TaxAmount, calc.TaxFee)
		amount.DiscountAmount = utils.DecimalAdd(amount.DiscountAmount, calc.CustomDiscountFee)
		amount.MemberDiscountAmount = utils.DecimalAdd(amount.MemberDiscountAmount, calc.MemberDiscountFee)
		amount.Amount = utils.DecimalAdd(amount.Amount, calc.Amount)
	}
	amount.ProductNum = productNum
	return resp.SentKitchen{
		Groups:     resp.GroupList{List: groups},
		AmountInfo: amount,
	}, nil
}
