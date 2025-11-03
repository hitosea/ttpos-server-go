package service

import (
	contexts "context"
	"encoding/json"
	builtinerrors "errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/product_resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/app/repository/saas"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/gin-gonic/gin"

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
	IMemberOrderSrv
	CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error)                                                                                 // 创建点餐订单
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                                                           // 创建桌台订单
	GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error)                                                               // 获取订单列表
	ExportOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderExportListPaginationResp, error)                                                      // 导出订单列表
	GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error)                                                                        // 获取订单详情
	GetRecordList(ctx context.Context, saleBillUuid uint64, h5OrderUuid uint64) ([]resp.OrderOperationLog, error)                                                // 获取订单操作日志
	CancelOrder(ctx context.Context, req req.OrderCancelReq) error                                                                                               // 取消订单
	DeleteOrder(ctx context.Context, dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error                                                               // 删除订单
	ReturnOrder(ctx context.Context, req req.OrderReturnReq) (error, int)                                                                                        // 退款订单
	ReReturnOrder(ctx context.Context, req req.OrderReReturnReq) (error, int)                                                                                    // 重新退款
	GetReturnOrderInfo(ctx context.Context, req req.OrderReturnInfoReq) (*resp.OrderReturnInfoResp, error)                                                       // 获取退款信息
	GetReverseSettleInfo(ctx context.Context, req req.OrderReverseSettleInfoReq) (*resp.OrderReverseSettleInfoResp, error)                                       // 获取反结账信息
	ReverseSettle(ctx context.Context, req req.OrderReverseSettleReq) error                                                                                      // 反结账
	IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error)                                                                          // 判断桌台是否可取消
	HideOrder(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error)                                                                                  // 挂单
	ShowOrder(ctx context.Context, req req.OrderShowReq) (*resp.ShopCart, error)                                                                                 // 显示订单
	InstantHideOrderList(ctx context.Context, req req.HideSaleBillListReq) (*resp.InstantHideOrderListResp, error)                                               // 获取挂单订单列表
	OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error)                                                                           // 打包
	OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error)                 // 删除订单商品
	OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error)                                                     // 修改订单商品价格
	OrderAmountChange(ctx context.Context, req req.OrderAmountChangeReq) (*resp.ShopCart, error)                                                                 // 修改订单金额
	OrderDiscount(ctx context.Context, req req.OrderDiscountReq) (*resp.ShopCart, error)                                                                         // 修改订单折扣
	OrderZeroRule(ctx context.Context, req req.OrderZeroRuleReq) (*resp.ShopCart, error)                                                                         // 修改订单抹零规则
	OrderDiscountCancel(ctx context.Context, req req.OrderDiscountCancelReq) (*resp.ShopCart, error)                                                             // 取消点餐订单所有优惠折扣，包括改价、打折、抹零
	OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error)                                                         // 修改订单人数
	GetOrderChangeBuffet(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (resp.OrderBuffetResp, error)                                           // 自助餐信息
	OrderChangeBuffet(ctx context.Context, req req.OrderChangeBuffetReq) (*resp.ShopCart, error)                                                                 // 调整自助餐
	OrderChangeBuffetClock(ctx context.Context, req req.OrderChangeBuffetClockReq) (*resp.ShopCart, error)                                                       // 调整自助餐
	OrderDeskBuffetProductList(ctx context.Context, req req.OrderChangeBuffetProductListReq) (*resp.BuffetProductList, error)                                    // 获取桌台的自助餐商品列表
	GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error)                                                                                             // 通过桌台uuid获取到销售账单信息
	OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                   // 修改订单商品备注
	OrderRemark(ctx context.Context, req req.OrderRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                                 // 修改订单备注
	CreateSaleBillSetting(ctx context.Context, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, isMember bool) (*model.SaleBillSetting, error)                 // 创建销售账单设置
	GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error)                                                                     // 通过设备SN获取点餐购物车信息
	GetOrderCartInfo(ctx context.Context, saleOrderUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                              // 获取购物车信息
	OrderCartProductPackageAdd(ctx context.Context, request req.OrderCartProductPackageAddReq) (*resp.ShopCart, error)                                           // 向购物车添加套餐
	OrderCartProductFlavorAndAttribute(ctx context.Context, request req.OrderCartProductFlavorAndAttributeReq) (*resp.ProductFlavorAndAttributeRes, error)       // 查询购物车商品“规格/属性”
	OrderCartProductFlavorAndAttributeChange(ctx context.Context, request req.OrderCartProductFlavorAndAttributeChangeReq) (*resp.ShopCart, error)               // 修改购物车商品“规格/属性”
	InstantOrderCartProductAdd(ctx context.Context, request req.OrderCartProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)      // 向购物车添加商品
	GetSaleBillUuidAndSaleOrderUuid(ctx context.Context, deskUuid uint64) (uint64, uint64, error)                                                                // 获取销售账单uuid和销售订单uuid
	OrderCartProductAdd(ctx context.Context, request req.ProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                      // 修改购物车商品数量
	OrderCartProductNum(ctx context.Context, req req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                 // 修改购物车商品数量
	AssistantOrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)    // 修改购物车商品数量
	InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, *resp.OrderCheckServiceRes, error)                  // 送厨购物车商品
	InstantOrderCartProductReturning(ctx context.Context, req req.OrderCartProductReturningReq) (*resp.ShopCart, error)                                          // 退菜购物车商品
	InstantOrderCartProductCancelReturning(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                                // 退菜购物车商品
	InstantOrderCartProductChangeDesk(ctx context.Context, req req.OrderCartProductChangeDeskReq) (*resp.ShopCart, error)                                        // 转菜购物车商品
	OrderCartProductWrap(ctx context.Context, req req.OrderCartProductWrapReq) (*resp.ShopCart, error)                                                           // 打包购物车商品
	OrderCartProductUnwrap(ctx context.Context, req req.OrderCartProductUnwrapReq) (*resp.ShopCart, error)                                                       // 取消打包购物车商品
	InstantOrderCartProductGiving(ctx context.Context, req req.OrderCartProductGivingReq) (*resp.ShopCart, error)                                                // 取消赠菜购物车商品
	InstantOrderCartProductCancelGiving(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                                   // 取消赠菜购物车商品
	InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, bool, error)                                                   // 获取点餐必点方案
	InstantOrderPaymentInfo(ctx context.Context, saleBill *model.SaleBill, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error) // 获取结账页面信息
	GetProductNameByItemCode(ctx context.Context, itemCode string, saleOrderUuid uint64) ([]ProductInfo, error)                                                  // 通过itemCode获取订单中库存不足的商品名称
	OrderPaymentCoupon(ctx context.Context, req req.InstantOrderPaymentCouponReq) (*resp.InstantOrderPaymentInfoResp, error)
	OrderPaymentPoints(ctx context.Context, req req.InstantOrderPaymentPointsReq) (*resp.InstantOrderPaymentInfoResp, error)                                                          // 设置订单的抵扣积分数量
	InstantOrderPaymentQrcode(ctx context.Context, req req.InstantOrderPaymentQrcodeReq) (*resp.InstantOrderPaymentQrcodeInfoResp, error)                                             // 获取支付二维码
	InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error)                                                   // 给销售订单创建一个支付单
	InstantOrderPaymentCancel(ctx context.Context, req req.InstantOrderPaymentCancelReq) (*resp.InstantOrderPaymentInfoResp, error)                                                   // 撤销一个支付单
	InstantOrderPaymentFinish(ctx context.Context, req req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error)                                                               // 给销售订单创建一个支付单
	InstantOrderFree(ctx context.Context, req req.InstantOrderFreeReq) (*resp.OrderFinishResp, error)                                                                                 // 免单
	InstantOrderPaymentZeroRule(ctx context.Context, req req.InstantOrderPaymentZeroRuleReq) (*resp.InstantOrderPaymentInfoResp, error)                                               // 设置结账抹零规则
	InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error)                                                                  // 给销售订单创建一个销售订单
	SaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq, needDeleteSaleOrder bool) (*resp.ShopCart, error)                                          // 从一个销售订单移动商品到另一个销售订单
	InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq, opts ...func(option *MustPlanConfirmOption)) (bool, *resp.InstantProductMustPlan, error) // 确认必点商品
	OrderCheck(ctx context.Context, req req.InstantOrderCheckReq) (*resp.OrderCheckServiceRes, error)                                                                                 // 订单检查
	InstantOrderSaleOrderDelete(ctx context.Context, req req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error)                                                                  // 删除一个销售订单(删除拆单)
	InstantOrderSaleOrderDeleteAll(ctx context.Context, req req.InstantOrderSaleOrderDeleteAllReq) (*resp.ShopCart, error)                                                            // 删除所有子销售订单(撤销拆单)
	OrderMemberCancel(ctx context.Context, req req.OrderMemberCancelReq) (*resp.InstantOrderPaymentInfoResp, error)                                                                   // 取消使用会员优惠
	OrderUseMember(ctx context.Context, req req.CheckMemberPasswordReq) (*resp.InstantOrderPaymentInfoResp, bool, error)                                                              // 使用会员优惠
	CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error                                                  // 计算并保存销售账单
	OrderPrint(ctx context.Context, req req.OrderPrintReq, needLock bool) (*resp.PrinterData, error)                                                                                  // 打印
	OrderPrintInvoice(ctx context.Context, req req.OrderPrintInvoiceReq) (*resp.PrinterData, error)                                                                                   // 图片打印
	OrderPrintInvoiceInfo(ctx context.Context, req req.OrderInvoiceInfoReq) resp.SaleOrderInvoiceInfo                                                                                 // 图片打印
	OrderUnlock(ctx context.Context, saleBillUuid uint64) error                                                                                                                       // 订单解锁
	GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error)                                                                                       // 必点方案列表
	GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error)             // 获取扫码h5购物车未下单商品列表
	GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error)           // 获取扫码h5购物车已下单商品列表
	ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64, ignoreMust bool) (any, error)                                                                      // 下单扫码h5订单
	AcceptH5Order(ctx context.Context, h5OrderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error)                                                                      // 接单扫码h5订单
	RejectH5Order(ctx context.Context, h5OrderUuid uint64) error                                                                                                                      // 拒单扫码h5订单
	RejectAllH5Order(ctx context.Context, saleBillUuid uint64) error
	RejectAllH5OrderInShop(ctx context.Context) error                                                                                                           // 拒单商家的所有待接单h5订单
	GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error) // 未送厨商品列表
	GetSentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart) (resp.SentKitchen, error)                                                 // 已送厨商品列表

	ActionCooking(ctx context.Context, ignoreMust bool, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct, h5OrderUuid uint64, isAutoOrder bool, options ...func(option *ActionCookingOption)) (*resp.OrderCheckServiceRes, error) // 送厨
	ActionAddAndCooking(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill, IgnoreMust bool) (*resp.OrderCheckServiceRes, error)                                                                                                          // 加购并送厨

	TabletAddAndCooking(ctx context.Context, request req.TabletOrderCartProductAddReq) (*TabletAddAndCookingRes, error) // 平板端加购并送厨

	GetOrderMemberList(ctx context.Context, saleBillUuid uint64) (resp.InstantOrderMemberList, error)                       // 获取订单会员列表
	GetProductPackageDetail(ctx context.Context, req req.GetProductPackageDetailReq) (*resp.ProductPackageDetailRes, error) // 获取商品选购详情

	GetOrderCartProductBatchCookingList(ctx context.Context, req req.GetOrderCartProductBatchCookingListReq) (*resp.OrderCartProductBatchCookingRes, error) // 获取分批送厨弹框的销售订单商品列表
	OrderCartProductBatchCooking(ctx context.Context, req req.OrderCartProductBatchCookingReq) (*resp.ShopCart, error)                                      // 分批送厨
}

// orderSrv 订单服务结构
type orderSrv struct {
	bus              *event.SystemEventBus
	dbm              *database.DBManager // 数据库管理器
	lock             lock.Lock
	localeSrv        ILocaleSrv
	settingSrv       setting.ISrv
	mustPlanSrv      IMustPlanSrv
	paymentMethodSrv IPaymentMethodSrv
	memberSrv        IMemberSrv
	cashBoxSrv       ICashBoxSrv
	smsSrv           ISmsSrv
}

type ISrvOption struct {
	SmsSrv ISmsSrv
}

func WithSmsSrv(dbm *database.DBManager) func(option *ISrvOption) {
	return func(option *ISrvOption) {
		option.SmsSrv = NewSMSSrv(dbm)
	}
}

// NewOrderSrv 创建订单服务实例
func NewOrderSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, mustPlanSrv IMustPlanSrv, paymentMethodSrv IPaymentMethodSrv, memberSrv IMemberSrv, cashBoxSrv ICashBoxSrv, opts ...func(option *ISrvOption)) IOrderSrv {
	return NewOrderSrvImpl(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, opts...)
}

// NewOrderSrvImpl 创建订单服务实例实现
func NewOrderSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, mustPlanSrv IMustPlanSrv, paymentMethodSrv IPaymentMethodSrv, memberSrv IMemberSrv, cashBoxSrv ICashBoxSrv, opts ...func(option *ISrvOption)) IOrderSrv {
	option := &ISrvOption{}
	for _, opt := range opts {
		opt(option)
	}
	return &orderSrv{
		bus:              event.NewSystemBus(),
		dbm:              dbm,
		lock:             lock.NewSystemLock(),
		localeSrv:        localeSrv,
		settingSrv:       settingSrv,
		mustPlanSrv:      mustPlanSrv,
		paymentMethodSrv: paymentMethodSrv,
		memberSrv:        memberSrv,
		cashBoxSrv:       cashBoxSrv,
		smsSrv:           option.SmsSrv,
	}
}

// HasInstantOrder 判断该收银机是否有未挂单的点餐订单
func HasInstantOrder(ctx context.Context, db *gorm.DB) (*model.SaleBill, bool, error) {
	deviceUuid := ctx.GetDeviceUuid()
	// 判断是否有待支付、未挂单的订单
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetInstantSaleBill(deviceUuid)
	if err != nil && !strings.Contains(err.Error(), "record not found") {
		return nil, false, errors.WithMessage(err, "获取待支付、未挂单的订单失败")
	}
	if saleBill != nil && deviceUuid == saleBill.DeviceUuid {
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
		// 创建订单编号
		orderNo, err := s.createOrderNo(tx, constant.OrderSourceInstant)
		if err != nil {
			ctx.Log().Error("订单编号生成失败", zap.Error(err))
			return errors.WithMessage(err, "订单编号生成失败")
		}

		serialNo, err := s.createInstantOrderSerialNo(ctx, tx)
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
			DeviceUuid:   ctx.GetDeviceUuid(),
		})
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售账单设置
		saleBillSetting, err := s.CreateSaleBillSetting(ctx, tx, saleBill.Uuid, saleBill.DeskUuid, false)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售订单
		saleOrder, errCreateSaleOrder := createSaleOrder(ctx, tx, saleBillSetting, saleBill.Uuid, orderNo)
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

// createMemberSaleOrderSerialNo 创建会员端订单序号
func (s *orderSrv) createMemberSaleOrderSerialNo(db *gorm.DB, timezone string) (string, error) {
	var serialNo string
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderLatest() // 获取最新的一条会员销售订单
	if err != nil {
		return "", errors.WithMessage(err)
	}
	// 如果没有查询到账单，则设置为0001
	if memberSaleOrder == nil {
		serialNo = "0001"
		return serialNo, nil
	}
	createTime := memberSaleOrder.SubmitPayTime
	// 判断账单的创建时间是不是今天
	if !IsToday(timezone, createTime) {
		serialNo = "0001"
	} else {
		oldSerialNo := memberSaleOrder.SerialNumber
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

func (s *orderSrv) createInstantOrderSerialNo(ctx context.Context, db *gorm.DB) (string, error) {
	var serialNo string
	saleBillRepo := repository.NewSaleBillRepo(db)
	saleBill, err := saleBillRepo.GetInstantSaleBillLatest()
	if err != nil {
		return "", errors.WithMessage(err)
	}

	startSerialNo := "0001"

	// 获取业务设置
	businessSetting, err := setting.NewSrv(s.dbm, ctx.GetCache()).GetBusinessSetting(ctx)
	if err == nil {
		startSerialNo = businessSetting.StartSerialNo
	}

	// 如果没有查询到账单，则设置为0001
	if saleBill == nil {
		serialNo = startSerialNo
		return serialNo, nil
	}
	createTime := saleBill.CreateTime
	setting := ctx.GetCompanySetting()

	if !IsToday(setting.GetTimezone(), createTime) {
		serialNo = startSerialNo
	} else {
		oldSerialNo := saleBill.SerialNo
		if oldSerialNo == "" {
			// 如果serialNo为空，则设置为startSerialNo. 兼容老数据没有serialNo的情况
			serialNo = startSerialNo
		} else {
			serialNoNum, err := strconv.Atoi(oldSerialNo)
			if err != nil {
				return "", errors.WithMessage(err)
			}

			startSerialNoNum, err := strconv.Atoi(startSerialNo)
			if err != nil {
				return "", errors.WithMessage(err)
			}

			// 如果当前序列号少于起始序列号，则设置为起始序列号，否则加1
			if serialNoNum < startSerialNoNum {
				serialNo = startSerialNo
			} else {
				serialNo = strconv.Itoa(serialNoNum + 1)
			}
		}
		if len(serialNo) < 4 {
			serialNo = strings.Repeat("0", 4-len(serialNo)) + serialNo
		}
	}
	return serialNo, nil
}

func IsToday(timezone string, timestamp int64) bool {
	return utils.SetTimezone(timezone).FormatUnixTime(timestamp, "20060102") == utils.SetTimezone(timezone).Now().Format("20060102")
}

func createSaleOrder(ctx context.Context, db *gorm.DB, saleBillSetting *model.SaleBillSetting, saleBillUuid uint64, saleBillOrderNo string) (*model.SaleOrder, error) {
	// 创建销售订单
	deviceSn := ctx.GetDeviceSn()
	if ctx.GetSource() == jwt.SourceH5 {
		deviceSn = jwt.SourceH5 // 扫码h5订单，设备sn为h5
	}
	if ctx.GetSource() == jwt.SourceMember {
		deviceSn = jwt.SourceMember // 会员端订单，设备sn为member
	}

	saleOrderObj := model.NewSaleOrder(deviceSn, saleBillUuid, saleBillOrderNo, *saleBillSetting)
	// 设置收银员信息
	staff := ctx.GetStaff()
	saleOrderObj.SetCashier(staff.Uuid, staff.GetUserName())

	// 设置员工班次信息
	if staff.Uuid != 0 {
		// 获取当前员工班次信息
		staffShiftLog, err := repository.NewShiftLogRepo(db).GetShiftLog(
			func(db *gorm.DB) *gorm.DB {
				return db.Where("staff_uuid = ?", staff.Uuid)
			},
			func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", constant.StaffNotHandedOver)
			},
		)
		if err == nil {
			// 班次信息存在，记录班次UUID
			saleOrderObj.StaffShiftLogUuid = staffShiftLog.Uuid
		}
		// 如果班次不存在，不影响订单创建，继续执行
	}

	// 设置会员折扣
	if ctx.GetMemberUuid() != 0 {
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(ctx.GetMemberUuid())
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		saleOrderObj.SetMemberDiscount(*member)
	}
	//
	saleOrder, err := repository.NewOrderRepo(db).CreateSaleOrder(*saleOrderObj)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &saleOrder, nil
}

// 创建会员端销售订单
func createMemberSaleOrder(ctx context.Context, db *gorm.DB, params model.CreateMemberSaleOrderParams) (*model.MemberSaleOrder, error) {
	saleOrderObj := model.NewMemberSaleOrder(params)

	if err := repository.NewMemberSaleOrderRepo(db).CreateMemberSaleOrder(*saleOrderObj); err != nil {
		return nil, errors.WithMessage(err)
	}
	return saleOrderObj, nil
}

// 解析服务费比例
// 兼容取值范围0-100的情况。本系统中比例的统一取值范围是0-1，所以需要转换
// 取值范围0-1
func parseServiceFeeRate(ServiceChargeRate string) (float64, error) {
	serviceFeeValue, err := utils.ParseFloat(ServiceChargeRate)
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	// 兼容取值范围0-100的情况。本系统中比例的统一取值范围是0-1，所以需要转换
	if serviceFeeValue > 0 {
		// 将取值范围0-100转换为0-1
		serviceFeeValue = decimal.NewFromFloat(serviceFeeValue).Div(decimal.NewFromInt(100)).Truncate(4).InexactFloat64()
		// 取值范围0-1
		serviceFeeValue = math.Min(serviceFeeValue, 1)
		serviceFeeValue = math.Max(serviceFeeValue, 0)
		return serviceFeeValue, nil
	}
	// 取值范围0-1
	serviceFeeValue = math.Min(serviceFeeValue, 1)
	serviceFeeValue = math.Max(serviceFeeValue, 0)
	return serviceFeeValue, nil
}

// NewSaleBillSetting 创建销售账单设置
// isMember: 是否是会员端
func (s *orderSrv) NewSaleBillSetting(ctx context.Context, saleBillUuid uint64, deskUuid uint64, isMember bool) (*model.SaleBillSetting, error) {
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
	// 获取积分设置
	pointsSetting, err := s.settingSrv.GetPointsSetting(ctx)
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
	// 开启服务费时
	if serviceFeeSetting.IsOpen == constant.SaleBillSettingIsOpenServiceFeeYes {
		// 服务费类型，固定金额
		if serviceFeeSetting.ChargeType == constant.SaleBillSettingServiceFeeFixed {
			serviceFeeType = constant.SaleBillSettingServiceFeeTypeFixed
			serviceFeeValue, err = utils.ParseFloat(serviceFeeSetting.ServiceCharge)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		}
		// 服务费类型，按比例
		if serviceFeeSetting.ChargeType == constant.SaleBillSettingServiceFeePercent {
			// 税费类型，不收取税费
			if serviceFeeSetting.IsOpenTax == constant.SaleBillSettingIsOpenTaxNo {
				serviceFeeType = constant.SaleBillSettingServiceFeeTypePercent
			}
			// 税费类型，收取税费
			if serviceFeeSetting.IsOpenTax == constant.SaleBillSettingIsOpenTaxYes {
				serviceFeeType = constant.SaleBillSettingServiceFeeTypePercentTax
			}
			serviceFeeValue, err = parseServiceFeeRate(serviceFeeSetting.ServiceChargeRate)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
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
	serviceApply, err := s.IsServiceFeeApply(ctx, deskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 服务费计算基准，默认是商品惠后价
	serviceFeeBase := uint(constant.SaleBillSettingServiceFeeBasePrice)
	if serviceFeeSetting.ServiceFeeBase == settingResp.ServiceFeeBaseAmount {
		serviceFeeBase = uint(constant.SaleBillSettingServiceFeeBaseAmount)
	}

	saleBillSetting := model.SaleBillSetting{
		SaleBillUuid:       saleBillUuid,
		ServiceFeeType:     serviceFeeType,
		ServiceFeeValue:    serviceFeeValue,
		TaxFeeType:         taxFeeType,
		DiscountType:       discountType,
		ZeroRule:           zero,
		ZeroCheckoutRule:   zeroCheckout,
		IsStatGift:         isStatGift,
		IsStatFree:         isStatFree,
		ServiceApply:       serviceApply,
		ServiceFeeBase:     serviceFeeBase,
		OpenPointsExchange: utils.BoolToUint(pointsSetting.GetOpenPointsExchange()),
		PointsExchangeRate: pointsSetting.GetPointsExchangeRate(),
		AutoPointsExchange: utils.BoolToUint(pointsSetting.IsAutoPointsExchange()),
	}

	// 如果是会员端订单，不收取服务费
	if isMember {
		// 不收取服务费
		saleBillSetting.ServiceApply = 0
		saleBillSetting.ServiceFeeType = constant.SaleBillSettingServiceFeeTypeNone
		saleBillSetting.ServiceFeeValue = 0
		// 不自动抹零
		saleBillSetting.ZeroRule = constant.SaleBillSettingCheckoutZeroingMethodNone
		saleBillSetting.ZeroCheckoutRule = constant.SaleBillSettingCheckoutZeroingMethodNone
		// 关闭积分抵扣
		saleBillSetting.OpenPointsExchange = 0
	}

	return &saleBillSetting, nil
}

// IsServiceFeeApply 判断该销售账单是否在服务费应用范围内
func (s *orderSrv) IsServiceFeeApply(ctx context.Context, deskUuid uint64) (uint, error) {
	// 是否是全部应用
	serviceFeeSetting, err := s.settingSrv.GetServiceFeeSetting(ctx)
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	// 是否是全部应用。是的话，则在范围内
	if serviceFeeSetting.IsApplyScopeAll() {
		return 1, nil
	}
	// 部分应用时
	// 点餐订单
	if deskUuid == 0 {
		// 部分应用时，点餐方式是否勾选
		if serviceFeeSetting.IsApplyScopeOrderingOpen() {
			return 1, nil
		} else {
			return 0, nil
		}
	}
	// 桌台订单
	if serviceFeeSetting.IsApplyScopeTableOpen() {
		// 判断账单的桌台是否在应用服务费的桌台列表中
		saleBillDeskUuid := int64(deskUuid)
		if slices.Contains(serviceFeeSetting.ApplyScopeTableList, saleBillDeskUuid) {
			return 1, nil
		}
	}
	return 0, nil
}

// CreateSaleBillSetting 创建销售账单设置
// isMember 是否是会员端订单
func (s *orderSrv) CreateSaleBillSetting(ctx context.Context, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, isMember bool) (*model.SaleBillSetting, error) {
	saleBillSetting, err := s.NewSaleBillSetting(ctx, saleBillUuid, deskUuid, isMember)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	newSaleBillSetting, err := repository.NewOrderRepo(db).CreateSaleBillSetting(*saleBillSetting)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &newSaleBillSetting, nil
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
	staff := ctx.GetStaff()
	saleBill := model.NewDeskSaleBill(saleBillUuid, orderNo, req.BuffetUuids, req.GetMealNum(), req.Remark, req.DeskUuid, desk.DeskNo, staff.DutyNo, staff.Uuid, staff.GetUserName())

	// 构建销售账单设置
	saleBillSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, req.DeskUuid, false)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}
	// 构建销售订单
	saleOrder := model.NewSaleOrder(ctx.GetDeviceSn(), saleBill.Uuid, saleBill.OrderNo, *saleBillSetting)
	staffShiftLogUuid := uint64(0)
	{
		staffShiftLog, err := GetCurrentStaffShiftLog(db, staff.Uuid)
		if err != nil {
			logger.Logger.Error("获取当前员工班次信息失败", zap.Error(err))
		} else {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
	}

	saleOrder.StaffShiftLogUuid = staffShiftLogUuid

	// 获取自助餐信息
	buffetList, err := repository.NewBuffetRepo(db).GetBuffetListByUuids(req.BuffetUuids)
	if err != nil {
		return resp.CreateDeskOrderResp{}, nil
	}

	// 构建自助餐顾客列表
	buffetCustomerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&buffetCustomerTypes, req.BuffetCustomerTypes)
	saleOrderBuffetCustomerTypes, _, _, maxTimeLimit, nonOrderingTime, reminderOrderTime := saleOrder.GetSaleOrderBuffetCustomerTypes(buffetList, req.BuffetUuids, buffetCustomerTypes, saleBillSetting)

	// 开始事务
	if err := db.Transaction(func(tx *gorm.DB) error {

		// 标记
		tx = tx.WithContext(contexts.WithValue(contexts.Background(), constant.OrderOperateSource, constant.OrderOpenTable))

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
				saleBill.NonOrderingTime = nonOrderingTime
				saleBill.ReminderOrderTime = reminderOrderTime
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

// getMemberSaleOrderAllInfo 获取会员端销售订单
func (s *orderSrv) getMemberSaleOrderAllInfo(ctx context.Context, memberSaleOrderUuid uint64, saleBill *model.SaleBill) (*model.MemberSaleOrder, error) {
	// 获取数据库
	db := ctx.GetDB()
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if saleBill == nil {
		// 当前销售账单数据
		var errSaleBill error
		saleBill, errSaleBill = repository.NewOrderRepo(db).GetSaleBillAllInfo(0, repository.WithMemberSaleOrderUuid(memberSaleOrderUuid))
		if errSaleBill != nil {
			return nil, errors.WithMessage(errSaleBill)
		}
	}
	memberSaleOrder.SaleBill = saleBill

	return memberSaleOrder, nil
}

// createMemberOrder 创建会员端订单
// 1. 创建member_sale_order
// 2. 创建sale_bill
func (s *orderSrv) createMemberOrder(ctx context.Context, request req.CreateMemberOrderReq) (*resp.CreateMemberOrderResp, error) {
	ctxCopy := ctx.Copy()
	db := ctx.GetDB()
	// 创建订单编号
	orderNo, err := s.createOrderNo(db, constant.OrderSourceMember)
	if err != nil {
		ctx.Log().Error("会员端订单编号生成失败", zap.Error(err))
		return nil, errors.WithMessage(err, "会员端订单编号生成失败")
	}

	memberSaleOrderUuid, _ := utils.GetID() // 生成外送订单uuid

	// 创建销售账单
	saleBill, err := repository.NewOrderRepo(db).CreateSaleBill(model.SaleBill{
		OrderNo:             orderNo,
		SerialNo:            "", // 等会员端订单提交支付时再生成订单流水号
		BillType:            constant.OrderSourceMapToBillType[constant.OrderSourceMember],
		DiningMethod:        constant.SaleBillDiningMethodTakeout,
		DeviceUuid:          ctx.GetDeviceUuid(),
		MemberSaleOrderUuid: memberSaleOrderUuid,
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 创建销售账单设置
	saleBillSetting, err := s.CreateSaleBillSetting(ctx, db, saleBill.Uuid, saleBill.DeskUuid, true)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 创建销售订单
	saleOrder, errCreateSaleOrder := createSaleOrder(ctx, db, saleBillSetting, saleBill.Uuid, orderNo)
	if errCreateSaleOrder != nil {
		return nil, errors.WithMessage(errCreateSaleOrder)
	}

	saleBillUuid := saleBill.Uuid
	saleOrderUuid := saleOrder.Uuid

	// 创建外送订单
	memberSaleOrder, errCreateMemberSaleOrder := createMemberSaleOrder(ctx, db, model.CreateMemberSaleOrderParams{
		Uuid:          memberSaleOrderUuid,
		SerialNo:      "", // 等会员端订单提交支付时再生成订单流水号
		OrderNo:       orderNo,
		MemberUuid:    ctx.GetMemberUuid(),
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
	})
	if errCreateMemberSaleOrder != nil {
		return nil, errors.WithMessage(errCreateMemberSaleOrder)
	}
	// 当前销售账单数据
	targetSaleBill, errSaleBill := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	// 添加订单商品。
	productParams := make([]req.ProductParams, 0)
	for _, product := range request.Products {
		param := req.ProductParams{
			FlavorProductBomUuid:            product.FlavorUuid,
			Num:                             product.Num,
			SauceProductBomUuidList:         product.SauceUuidList,
			ProductPackageAttributeUuidList: product.AttributeUuidList,
		}
		productParams = append(productParams, param)
	}
	params := req.ProductAddReq{
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
		Products:      productParams,
		IsMemberAdd:   true,
	}
	if err := s.ActionAdd(ctx, params, targetSaleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	ctxCopy.SetDB(db)
	info, err := s.GetMemberOrderCheckoutInfo(ctxCopy, req.GetMemberOrderCheckoutInfoReq{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
	}, nil)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 用于记录操作日志
	info.MemberSaleOrderInfo.SaleBillUuid = saleBillUuid
	info.MemberSaleOrderInfo.SaleOrderUuid = saleOrderUuid
	return info, nil
}

// GetMemberOrderCheckoutInfo 获取会员端订单结账页面信息
func (s *orderSrv) GetMemberOrderCheckoutInfo(ctx context.Context, req req.GetMemberOrderCheckoutInfoReq, saleBill *model.SaleBill) (*resp.CreateMemberOrderResp, error) {
	if saleBill == nil {
		// 当前销售账单数据
		var errSaleBill error
		saleBill, errSaleBill = repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(0, repository.WithMemberSaleOrderUuid(req.MemberSaleOrderUuid))
		if errSaleBill != nil {
			return nil, errors.WithMessage(errSaleBill)
		}
	}
	memberSaleOrder, err := s.getMemberSaleOrderAllInfo(ctx, req.MemberSaleOrderUuid, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 重新设置订单的member_order_discount_rate, 并计算商品金额
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	rate := businessSetting.GetDeliveryPriceRatio() // 外送商品价格和商品原价比例. 取值范围1-300， 表示原价的1%到300%
	saleBill.SetMemberOrderDiscountRate(rate)
	if err := s.CalcAndSaveSaleBill(ctx, ctx.GetDB(), saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	products := make([]resp.MemberSaleOrderProduct, 0)
	for _, saleOrderProduct := range saleBill.GetFirstSaleOrder().SaleOrderProducts {
		amount := saleOrderProduct.GetTotalPriceOrigin()
		products = append(products, resp.MemberSaleOrderProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			Num:                  saleOrderProduct.Num,
			UnitPrice:            saleOrderProduct.OriginTotalPrice,
			Amount:               amount,
			LocaleName:           saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(),
			Image: func() string {
				if saleOrderProduct.ImageFileUuid != 0 {
					return saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
				}
				return ""
			}(),
		})
	}

	// 获取支付方式
	paymentMethods, err := repository.NewPaymentMethodRepo(ctx.GetDB()).GetLianLianPayPaymentMethodList()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	payList := make([]resp.PaymentMethodItem, 0)
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	for _, paymentMethod := range paymentMethods {
		payList = append(payList, resp.PaymentMethodItem{
			Source:        paymentMethod.Source,
			SourceText:    constant.PaymentMethodSourceTextMap[paymentMethod.Source],
			Uuid:          paymentMethod.Uuid,
			PaymentName:   paymentMethod.GetPaymentName(),
			PaymentMethod: paymentMethod.GetName(),
			FeePercent:    paymentMethod.FeePercent,
			Logo: func() string {
				if paymentMethod.IsWechatPay() {
					return baseUrl + "/image/pay/wechat_pay.png"
				}
				if paymentMethod.IsAliPay() {
					return baseUrl + "/image/pay/alipay.png"
				}
				if paymentMethod.IsQrPromptPay() {
					return baseUrl + "/image/pay/qr_prompt_pay.png"
				}
				return ""
			}(),
			Code: paymentMethod.Code,
		})
	}

	// 如果订单未设置地址，查询会员的默认地址,如果有默认地址，则使用默认地址
	if memberSaleOrder.MemberAddressUuid == 0 {
		memberAddress, err := repository.NewMemberAddressRepo(ctx.GetDB()).GetMemberDefaultAddress(memberSaleOrder.MemberUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if memberAddress != nil {
			memberSaleOrder.SetMemberAddress(*memberAddress)
		}
	}

	var address resp.MemberSaleOrderAddress
	if memberSaleOrder.Address != nil {
		lat, lng, err := memberSaleOrder.Address.GetLocation()
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		address = resp.MemberSaleOrderAddress{
			MemberAddressUuid:  memberSaleOrder.MemberAddressUuid,
			Longitude:          lng,
			Latitude:           lat,
			Address:            memberSaleOrder.ContactAddress,
			DetailAddress:      memberSaleOrder.ContactAddressDetail,
			ContactName:        memberSaleOrder.ContactName,
			ContactPhone:       memberSaleOrder.ContactPhone,
			ContactPhonePrefix: memberSaleOrder.ContactPhonePrefix,
			ContactGender:      memberSaleOrder.ContactGender,
		}
	}

	isOutRange := false
	// 如果未计算距离费，且配置了地址，则查询配送距离
	if !memberSaleOrder.GetIsDistanceCalculated() && memberSaleOrder.MemberAddressUuid != 0 {
		// 查询配送距离
		distance, err := s.QueryDistance(ctx, memberSaleOrder)
		if err != nil {
			ctx.Log().Error("查询配送距离失败", zap.Error(err))
			if strings.Contains(err.Error(), "不在配送范围内，无法下单") {
				isOutRange = true
			}
		} else {
			memberSaleOrder.SetDeliveryDistance(distance)
			if err := repository.NewMemberSaleOrderRepo(ctx.GetDB()).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
				return nil, errors.WithMessage(err)
			}
		}
	}

	// 每次获取会员端订单信息时，重新获取最新的外送费用配置
	companySetting := ctx.GetCompanySetting()
	deliveryConfig, err := companySetting.GetDeliveryConfig(constant.ProviderNameSkootar, memberSaleOrder.DeliveryDistance)
	if err != nil {
		return nil, errors.WithMessage(err, "配送费配置失败")
	}
	memberSaleOrder.UpdateDeliveryConfig(*deliveryConfig)

	// 保存一下
	memberSaleOrder.ProductNum = memberSaleOrder.SaleBill.SaleOrders[0].GetProductNum()
	memberSaleOrder.ProductAmount = memberSaleOrder.SaleBill.SaleOrders[0].GetProductAmount()
	memberSaleOrder.OriginProductAmount = memberSaleOrder.SaleBill.SaleOrders[0].GetOriginProductAmount()
	memberSaleOrder.MemberDiscountFee = memberSaleOrder.CalculateMemberDiscount()
	memberSaleOrder.Amount = memberSaleOrder.CalculateAmount()
	if err := repository.NewMemberSaleOrderRepo(ctx.GetDB()).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	info := &resp.MemberSaleOrderInfo{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		Status:              memberSaleOrder.Status,
		ProductList:         resp.MemberSaleOrderProductList{List: products},
		ProductAmount:       memberSaleOrder.OriginProductAmount,
		MemberDiscount:      memberSaleOrder.MemberDiscountFee,
		Amount:              memberSaleOrder.Amount,
		Remark:              memberSaleOrder.Remark,
		IsVerifiedPhone:     memberSaleOrder.IsVerifiedPhoneBool(),
		IsInDeliveryRange: func() bool {
			if isOutRange {
				return false
			}
			if deliveryConfig != nil {
				return deliveryConfig.IsInDeliveryRange
			}
			return false // 如果配送费配置为空，则认为不在配送范围内
		}(),
		Address: address,
		DeliveryFee: resp.MemberSaleOrderDeliveryFee{
			Amount:   memberSaleOrder.CalculateDeliveryFee(),
			Distance: memberSaleOrder.DeliveryDistance,
			MinFee:   memberSaleOrder.DeliveryFeeMinFee,
			BaseFee:  memberSaleOrder.DeliveryFeeBaseFee,
			FeePerKm: memberSaleOrder.DeliveryFeePerKm,
		},
		PaymentMethods: resp.PaymentMethodList{List: payList},
	}

	return &resp.CreateMemberOrderResp{
		MemberSaleOrderInfo: info,
	}, nil
}

// 查询配送距离
func (s *orderSrv) QueryDistance(ctx context.Context, memberSaleOrder *model.MemberSaleOrder) (float64, error) {
	takeoutSrv := takeout.NewTakeoutSrv()

	// 收货人地址
	lat, lng, err := memberSaleOrder.Address.GetLocation()
	if err != nil {
		return 0, errors.WithMessage(err)
	}

	companySetting := ctx.GetCompanySetting()
	latitude, longitude := companySetting.GetCoordinates()
	if latitude == "" || longitude == "" {
		return 0, errors.WithMessage(errors.New("无法找到商家经纬度"))
	}

	takeoutResp, err := takeoutSrv.EstimateDistance(contexts.Background(), &req.TakeoutDistanceReq{
		ProviderName: constant.ProviderNameSkootar,
		Address: []*req.TakeoutAddress{
			// 商家地址
			{
				AddressName: ctx.GetCompany().Name,
				Address:     companySetting.Address,
				Lat:         latitude,
				Lng:         longitude,
			},
			// 收货人地址
			{
				AddressName: memberSaleOrder.ContactName,
				Address:     memberSaleOrder.ContactAddress,
				Lat:         lat,
				Lng:         lng,
			},
		},
	})
	if err != nil {
		ctx.Log().Error("计算配送距离失败", zap.Error(err))
		return 0, errors.WithMessage(errors.NewWithCode(constant.CodeDistanceError, "不在配送范围内，无法下单"), err.Error())
	}

	return takeoutResp.Distance, nil
}

// updateMemberOrder 更新会员端订单
func (s *orderSrv) updateMemberOrder(ctx context.Context, request req.CreateMemberOrderReq, memberSaleOrder *model.MemberSaleOrder) (*resp.CreateMemberOrderResp, *resp.OrderCheckServiceRes, error) {
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(0, repository.WithMemberSaleOrderUuid(request.MemberSaleOrderUuid))
	if errSaleBill != nil {
		return nil, nil, errors.WithMessage(errSaleBill)
	}
	if saleBill == nil {
		return nil, nil, errors.New("无法找到销售账单")
	}
	memberSaleOrder.SaleBill = saleBill

	// 识别出本次提交订单中"删除的订单商品"、"新增的订单商品"、"修改的订单商品"
	commitProductMap := make(map[string]req.OrderProductAddReq)
	for index := range request.Products {
		product := request.Products[index]
		productPackageAttributes, err := repository.NewProductPackageAttributeRepo(ctx.GetDB()).GetProductPackageAttributesByUuids(product.AttributeUuidList)
		if err != nil {
			return nil, nil, errors.WithMessage(err)
		}
		attributeUuids := make([]uint64, 0)
		for _, attribute := range productPackageAttributes {
			attributeUuids = append(attributeUuids, attribute.AttributeUuid)
		}
		key := product.ProductKey(attributeUuids)
		commitProductMap[key] = product
	}
	olderProductMap := make(map[string]*model.SaleOrderProduct)
	for index := range memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts {
		product := memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts[index]
		key := product.ProductKey()
		olderProductMap[key] = product
	}

	addProductMap := make(map[string]req.OrderProductAddReq)          // 新增商品
	deleteProductMap := make(map[string]*model.SaleOrderProduct)      // 删除商品
	updateProductMap := make(map[string]req.OrderProductAddReq)       // 修改商品
	updateOlderProductMap := make(map[string]*model.SaleOrderProduct) // 修改商品

	// commitProductMap提交中存在，而olderProductMap中不存在的商品，为新增商品
	for key, product := range commitProductMap {
		if _, ok := olderProductMap[key]; !ok {
			// 新增商品
			addProductMap[key] = product
		}
	}

	// commitProductMap提交中不存在，而olderProductMap中存在商品，为删除商品
	for key, product := range olderProductMap {
		if _, ok := commitProductMap[key]; !ok {
			// 删除商品
			deleteProductMap[key] = product
		}
	}

	// commitProductMap提交中存在，而olderProductMap中存在商品，为修改商品
	for key, product := range commitProductMap {
		if _, ok := olderProductMap[key]; ok {
			// 修改商品
			updateProductMap[key] = product
			updateOlderProductMap[key] = olderProductMap[key]
		}
	}

	// 删除商品
	for _, saleOrderProduct := range deleteProductMap {
		saleOrderProduct.DeleteProduct()
	}

	// 修改商品
	for key, product := range updateProductMap {
		saleOrderProduct := updateOlderProductMap[key]
		saleOrderProduct.Num = product.Num
	}

	// 新增商品
	productParams := make([]req.ProductParams, 0)
	for _, product := range addProductMap {
		// 添加订单商品。
		param := req.ProductParams{
			FlavorProductBomUuid:            product.FlavorUuid,
			Num:                             product.Num,
			SauceProductBomUuidList:         product.SauceUuidList,
			ProductPackageAttributeUuidList: product.AttributeUuidList,
		}
		productParams = append(productParams, param)
	}
	params := req.ProductAddReq{
		SaleBillUuid:  memberSaleOrder.SaleBill.Uuid,
		SaleOrderUuid: memberSaleOrder.SaleBill.SaleOrders[0].Uuid,
		Products:      productParams,
		IsMemberAdd:   true,
	}
	if err := s.ActionAdd(ctx, params, memberSaleOrder.SaleBill); err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	info, err := s.GetMemberOrderCheckoutInfo(ctx, req.GetMemberOrderCheckoutInfoReq{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
	}, nil)
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}
	return info, nil, nil
}

// getMemberOrderDetail 获取完整的外送订单信息
func getMemberOrderDetail(ctx context.Context, memberSaleOrderUuid uint64) (*model.MemberSaleOrder, error) {
	db := ctx.GetDB()
	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(0, repository.WithMemberSaleOrderUuid(memberSaleOrderUuid))
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	memberSaleOrder.SaleBill = saleBill

	return memberSaleOrder, nil
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
		noExist, err := repository.NewOrderRepo(db).IsOrderNoExists(orderNo)
		if err != nil {
			return "", errors.WithMessage(err)
		}
		// 如果订单编号存在，则重新生成
		if !noExist {
			orderNo = ""
			continue
		}
		// 如果订单编号不存在，则退出，本次生成的订单编号可用
		break
	}
	if orderNo == "" {
		return "", errors.New("订单编号生成失败")
	}
	return orderNo, nil
}

// GetOrderLists 获取订单列表
func (s *orderSrv) GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	lists, total, dbOption, err := orderRepo.GetCashierOrderListWithPagination(reqs, ctx.GetCompanySetting().Timezone)
	if err != nil {
		return resp.OrderListPaginationResp{}, errors.WithMessage(err)
	}

	// 组合列表源数据
	billList := make([]resp.BillLists, len(lists))
	for i, bill := range lists {
		consumerUuids := []string{}
		totalPayTypeNames := []string{}
		isSplit := len(bill.SaleOrders) > 1 // 拆单
		orderList := make([]resp.BillListsOrder, 0)
		var paymentAmounts float64
		//
		billListsExtra := resp.BillListsExtra{
			IsCellRefund:        false,
			IsCellCancel:        bill.Status == constant.SaleBillStatusPending,
			IsCellReverseSettle: bill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime),
			IsCellPrint:         !isSplit,
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
						if payment.IsDelete() {
							continue
						}
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
						payTypeNames = append(payTypeNames, payment.PaymentMethodName)
					}
				}

				orderExtra := resp.BillListsExtra{
					IsCellRefund:        false,
					IsCellCancel:        false,
					IsCellReverseSettle: false,
					IsCellPrint:         true,
					IsCellInvoice:       order.Status == constant.SaleBillStatusComplete,
					IsCellDelete:        order.Status == constant.SaleBillStatusCanceled,
				}
				// 不等于免单 && 未全退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					orderExtra.IsCellRefund = true
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
						return strconv.FormatUint(uint64(order.Member.ID), 10)
					}(),
					OrderNo:       order.OrderNo,
					Status:        order.Status,
					FinishTime:    order.FinishTime,
					OrderAmount:   order.OriginAmount,
					PaymentAmount: paymentAmount,
					PayTypeName:   strings.Join(utils.RemoveDuplicates(payTypeNames), ","),
					Extra:         orderExtra,
				})
				//
				if order.ConsumerUuid > 0 {
					consumerUuids = append(consumerUuids, strconv.FormatUint(uint64(order.Member.ID), 10))
				}
			}
		} else {
			// 没有拆单
			if len(bill.SaleOrders) > 0 {
				order := bill.SaleOrders[0]
				if order.ConsumerUuid > 0 {
					if order.Member == nil {
						logger.Logger.Info("member is nil", zap.Any("order", order))
					}
					consumerUuids = append(consumerUuids, strconv.FormatUint(uint64(order.Member.ID), 10))
				}
				if order.IsFree == 1 {
					totalPayTypeNames = append(totalPayTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
				} else {
					for _, payment := range order.PaymentOrders {
						if payment.IsDelete() {
							continue
						}
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
					}
				}
				// 不等于免单 && 未退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					billListsExtra.IsCellRefund = true
				}
				// 等于主单 && 完成 && 等于当前用户 && 在班次时间内
				billListsExtra.IsCellReverseSettle = bill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime)
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
			OrderAmount:   bill.OriginAmount,
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
			repository.CommonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
			dbOption,
		)
		return num
	}
	// 获取数量
	unpaidNum := getOrderNum(constant.SaleBillStatusPending)
	completeNum := getOrderNum(constant.SaleBillStatusComplete)
	cancelNum := getOrderNum(constant.SaleBillStatusCanceled)

	// 返回响应对象
	return resp.OrderListPaginationResp{
		List: billList,
		Meta: resp.OrderListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			TotalNum:    unpaidNum + completeNum + cancelNum,
			UnpaidNum:   unpaidNum,
			CompleteNum: completeNum,
			CancelNum:   cancelNum,
		},
	}, nil
}

// ExportOrderLists 导出订单列表
func (s *orderSrv) ExportOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderExportListPaginationResp, error) {
	language := ctx.GetLanguage()

	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
	lists, total, _, err := orderRepo.GetCashierOrderExportListWithPagination(reqs, ctx.GetCompanySetting().Timezone)
	if err != nil {
		return resp.OrderExportListPaginationResp{}, errors.WithMessage(err)
	}

	statusText := map[uint]string{
		constant.SaleBillStatusPending:  i18n.Translate(language, "待付款"),
		constant.SaleBillStatusComplete: i18n.Translate(language, "已完成"),
		constant.SaleBillStatusCanceled: i18n.Translate(language, "已取消"),
	}

	// 组合列表源数据
	saleBillUuids := []uint64{}
	exportLists := make([]resp.OrderExportInfo, 0)
	for _, bill := range lists {
		saleBillUuids = append(saleBillUuids, bill.Uuid)
		isSplit := len(bill.SaleOrders) > 1
		// 拆单
		for index, saleOrder := range bill.SaleOrders {

			var products []*resp.OrderExportInfoProduct
			// 添加自助餐顾客
			for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if orderBuffetCustomer.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(language),
					AttrName:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					Num:        float64(orderBuffetCustomer.Num),
					TotalPrice: orderBuffetCustomer.GetDiscountPriceWithVAT(),
				})
			}
			// 添加加钟商品
			for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if delayProduct.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       delayProduct.Name,
					AttrName:   "",
					Num:        float64(delayProduct.Num),
					TotalPrice: delayProduct.GetAmount(),
				})
			}
			// 添加正常商品
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       saleOrderProduct.MultiLanguageName.GetNameByLang(language),
					AttrName:   saleOrderProduct.GetAttributeNamesByLang(language),
					Num:        saleOrderProduct.Num,
					TotalPrice: saleOrderProduct.GetTotalPrice(),
				})
			}
			//
			exportLists = append(exportLists, resp.OrderExportInfo{
				CreateTime:    saleOrder.CreateTime,
				BillUuid:      bill.Uuid,
				BillType:      utils.IfString(bill.IsInstantBill(), i18n.Translate(language, "点餐订单"), i18n.Translate(language, "桌台订单")),
				Products:      products,
				SerialNo:      utils.IfString(isSplit, bill.SerialNo+"-"+strconv.Itoa(index+1), bill.SerialNo),
				OrderNo:       saleOrder.OrderNo,
				Status:        bill.Status,
				StatusText:    statusText[bill.Status],
				FinishTime:    saleOrder.FinishTime,
				OrderAmount:   saleOrder.OriginAmount,
				ServiceFee:    saleOrder.ServiceFee,
				DiscountFee:   saleOrder.CustomDiscountFee,
				MemberFee:     saleOrder.MemberDiscountFee,
				PaymentAmount: saleOrder.GetActualPaymentAmount(),
				RefundAmount:  saleOrder.GetTotalRefundAmount(),
				MemberNames:   saleOrder.GetMemberName(),
				MemberIds: func() string {
					if saleOrder.Member == nil {
						return ""
					}
					return strconv.FormatUint(uint64(saleOrder.Member.ID), 10)
				}(),
				PayTypeName:  saleOrder.GetPayTypeNames(ctx.GetLanguage()),
				DiningMethod: bill.DiningMethod,
				CashierName:  bill.CashierName,
			})
		}
		// 拆单
		if isSplit && len(exportLists) > 0 {
			mainOrder := exportLists[len(exportLists)-1]
			// 收集当前账单所有会员名称并去重
			var allMemberNames []string
			var allMemberUuids []string
			var allProducts []*resp.OrderExportInfoProduct
			var allPayTypeNames []string
			for _, orderExportInfo := range exportLists {
				if orderExportInfo.BillUuid == bill.Uuid {
					// 添加产品和支付方式
					allProducts = append(allProducts, orderExportInfo.Products...)
					// 处理MemberNames，将字符串拆分为数组并逐个添加
					if orderExportInfo.MemberNames != "" {
						memberNames := strings.Split(orderExportInfo.MemberNames, ",")
						for _, name := range memberNames {
							if name != "" {
								allMemberNames = append(allMemberNames, name)
							}
						}
					}
					if orderExportInfo.MemberIds != "" {
						memberIds := strings.Split(orderExportInfo.MemberIds, ",")
						for _, id := range memberIds {
							if id != "" {
								allMemberUuids = append(allMemberUuids, id)
							}
						}
					}
					if orderExportInfo.PayTypeName != "" {
						payTypeNames := strings.Split(orderExportInfo.PayTypeName, ",")
						for _, name := range payTypeNames {
							if name != "" {
								allPayTypeNames = append(allPayTypeNames, name)
							}
						}
					}
				}
			}
			//
			mainOrder.Products = allProducts
			mainOrder.SerialNo = bill.SerialNo
			mainOrder.OrderNo = bill.OrderNo
			mainOrder.Status = bill.Status
			mainOrder.StatusText = statusText[bill.Status]
			mainOrder.FinishTime = bill.FinishTime
			mainOrder.OrderAmount = bill.OriginAmount
			mainOrder.ServiceFee = bill.ServiceFee
			mainOrder.DiscountFee = bill.CustomDiscountFee
			mainOrder.MemberFee = bill.MemberDiscountFee
			mainOrder.PaymentAmount = bill.GetPaymentAmount()
			mainOrder.RefundAmount = bill.GetTotalRefundAmount()
			mainOrder.MemberIds = strings.Join(utils.RemoveDuplicates(allMemberUuids), ",")
			mainOrder.MemberNames = strings.Join(utils.RemoveDuplicates(allMemberNames), ",")
			mainOrder.PayTypeName = strings.Join(utils.RemoveDuplicates(allPayTypeNames), ",")
			mainOrder.DiningMethod = bill.DiningMethod
			mainOrder.CashierName = bill.CashierName
			exportLists = append(exportLists, mainOrder)
		}
	}
	rankLists, err := orderRepo.GetMonthlyOrderRanks(saleBillUuids)
	if err != nil {
		return resp.OrderExportListPaginationResp{}, errors.WithMessage(err)
	}
	for i, exportList := range exportLists {
		result, ok := slice.FindBy(rankLists, func(index int, rankList repository.MonthlyOrderRank) bool {
			return rankList.OrderNo == exportList.OrderNo
		})
		if ok {
			exportLists[i].OrderID = fmt.Sprintf("OID%s%05d", result.MonthYear, result.MonthlyOrderNumber)
		}
	}
	//
	return resp.OrderExportListPaginationResp{
		List: exportLists,
		Meta: resp.OrderExportMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
		},
	}, errors.WithMessage(err)
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
			totalMemberUuids = append(totalMemberUuids, strconv.FormatUint(uint64(saleOrder.Member.ID), 10))
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
						SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					},
					Price:            orderBuffetCustomer.SalePrice,
					Num:              float64(orderBuffetCustomer.Num), // 这种类型顾客多少个，如老人这个类型2人
					SalePrice:        orderBuffetCustomer.GetDiscountPriceWithVAT(model.WithOriginPrice()),
					TotalPrice:       orderBuffetCustomer.GetDiscountPriceWithVAT(),
					RefundAmount:     -orderBuffetCustomer.GetReturnPrice(),
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
						SV:   delayProduct.Name,
					},
					LocaleAttributeName: dto.LocaleResponse{},
					Num:                 float64(delayProduct.Num), // 拆单后不等于桌台人数，但同一个加钟商品的总数等于桌台人数
					Price:               delayProduct.Price,
					SalePrice:           delayProduct.GetAmount(),
					TotalPrice:          delayProduct.GetAmount(),
					RefundAmount:        -delayProduct.GetReturnPrice(),
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
				// 取消订单时，过滤掉未送厨的商品
				if saleBill.IsCanceled() {
					if saleOrderProduct.IsUnCookingProduct() {
						continue
					}
				}
				// 过滤掉套餐子商品
				if saleOrderProduct.IsPackageSubProduct() {
					continue
				}

				// 过滤掉未接单的商品
				if !saleOrderProduct.IsAcceptOrderProduct() {
					continue
				}
				imageUrl := ""
				if saleOrderProduct.ImageFile != nil {
					imageUrl = saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
				}
				cancelReason := saleOrderProduct.GetCancelReason()
				giftReason := saleOrderProduct.GetGiftReason()

				attributeName := saleOrderProduct.GetAttributeName()
				if saleOrderProduct.IsPackageProduct() {
					// 如果是套餐商品，则获取各个子商品的名称、数量、规格、属性，如：“牛排*1（标准，黑椒汁）；可乐*2（大杯，少冰）；沙拉*1（大份，沙拉酱，蜂蜜酱）”
					subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
					zh := ""
					th := ""
					en := ""
					zhtw := ""
					ja := ""
					ko := ""
					my := ""
					tr := ""
					sv := ""
					for _, subProduct := range subProducts {
						zh += subProduct.GetProductNameAttributes(string(constant.LocaleZH)) + "；"
						th += subProduct.GetProductNameAttributes(string(constant.LocaleTH)) + "；"
						en += subProduct.GetProductNameAttributes(string(constant.LocaleEN)) + "；"
						zhtw += subProduct.GetProductNameAttributes(string(constant.LocaleZHTW)) + "；"
						ja += subProduct.GetProductNameAttributes(string(constant.LocaleJA)) + "；"
						ko += subProduct.GetProductNameAttributes(string(constant.LocaleKO)) + "；"
						my += subProduct.GetProductNameAttributes(string(constant.LocaleMY)) + "；"
						tr += subProduct.GetProductNameAttributes(string(constant.LocaleTR)) + "；"
						sv += subProduct.GetProductNameAttributes(string(constant.LocaleSV)) + "；"
					}
					attributeName = dto.LocaleResponse{
						ZH:   zh,
						TH:   th,
						EN:   en,
						ZHTW: zhtw,
						JA:   ja,
						KO:   ko,
						MY:   my,
						TR:   tr,
						SV:   sv,
					}
				}

				products = append(products, resp.OrderProduct{
					Uuid:                saleOrderProduct.Uuid,
					LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: attributeName,
					Price:               saleOrderProduct.SalePrice,
					Num:                 saleOrderProduct.Num,
					SalePrice:           saleOrderProduct.GetTotalPriceOrigin(),
					TotalPrice:          saleOrderProduct.GetTotalPrice(),
					Status:              saleOrderProduct.Status,
					Remark:              saleOrderProduct.Remark,
					IsMust:              saleOrderProduct.IsMustProduct(),
					IsGift:              saleOrderProduct.IsGiftProduct(),
					IsWrap:              saleOrderProduct.IsWrapProduct(),
					IsBuffet:            saleOrderProduct.IsBuffetProduct(),
					ImageUrl:            imageUrl,
					CancelReason:        cancelReason.GetLocale(ctx.GetLanguage()),
					GiftReason:          giftReason.GetLocale(ctx.GetLanguage()),
					RefundAmount:        -saleOrderProduct.GetReturnPrice(),
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
			FreeReason:    saleOrder.GetFreeReason(),
			OrderAmount:   saleOrder.OriginAmount,
			PaymentAmount: saleOrder.GetActualPaymentAmount(),
			RefundAmount:  saleOrder.GetTotalRefundAmount(),
			PayTypeName:   saleOrder.GetPayTypeNames(ctx.GetLanguage()),
			MemberName:    saleOrder.GetMemberName(),
			MemberUuid: func() uint64 {
				if saleOrder.Member == nil {
					return uint64(0)
				}
				return uint64(saleOrder.Member.ID)
			}(),
			Products: products,
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
		IsCellReverseSettle: saleBill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime),
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
			OrderAmount:   saleBill.OriginAmount,
			PaymentAmount: saleBill.GetPaymentAmount(),
			RefundAmount:  saleBill.GetTotalRefundAmount(),
			MemberNames:   strings.Join(totalMemberNames, ","),
			MemberUuids:   strings.Join(totalMemberUuids, ","),
			CashierName:   saleBill.CashierName,
			IsBuffet:      saleBill.IsBuffet == constant.SaleBillIsBuffetYes,
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
				logs, err := s.GetRecordList(ctx, req.SaleBillUuid, 0)
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
func (s *orderSrv) GetRecordList(ctx context.Context, saleBillUuid uint64, h5OrderUuid uint64) ([]resp.OrderOperationLog, error) {
	orderRecordRepo := repository.NewOrderOperationRecordRepo(ctx.GetDB())
	var dbOptions []repository.DBOption
	dbOptions = append(dbOptions, orderRecordRepo.WithSaleBillUuid(saleBillUuid))
	if h5OrderUuid > 0 {
		dbOptions = append(dbOptions, orderRecordRepo.WithH5OrderUuid(h5OrderUuid))
	}
	orderRecordLists, err := orderRecordRepo.GetRecordLists(dbOptions...)
	if err != nil {
		return []resp.OrderOperationLog{}, errors.WithMessage(err)
	}
	logs := make([]resp.OrderOperationLog, 0)
	language := ctx.GetLanguage()

	for _, record := range orderRecordLists {
		actionDescription := s.getActionDescription(ctx, record, language)
		if actionDescription.HideLog {
			// 隐藏. 日志关联的订单已经被删除，故因此该订单的操作记录
			continue
		}

		// 获取操作描述
		actionText := s.getActionText(record, language)
		if record.Action == constant.OrderCheckoutDiscount && actionDescription.IsAutoCheckoutZero {
			actionText = i18n.Translate(language, "结账自动抹零")
		}
		if actionDescription.SplitMessage != "" {
			actionText = actionDescription.SplitMessage + actionText
		}
		if actionDescription.Desc != "" {
			if record.Source == constant.SourceMember || record.Action == constant.OrderCancelMemberSaleOrder {
				actionText = actionText + actionDescription.Desc
			} else {
				actionText = actionText + ": " + actionDescription.Desc
			}
		}
		realName := record.Operator.RealName
		if record.Source == constant.SourceH5 {
			realName = i18n.Translate(language, "用户")
		}

		// 如果是骑手端的操作
		if record.Source == constant.SourceRider {
			prefix := i18n.Translate(language, "骑手")
			riderName := "--"
			var payload event.RiderAcceptMemberSaleOrderPayload
			err := json.Unmarshal([]byte(record.Data), &payload)
			if err == nil {
				riderName = payload.RiderName
			}
			realName = fmt.Sprintf("%s(%s)", prefix, riderName)
		}

		// 如果是会员端的操作
		if record.Source == constant.SourceMember {
			prefix := i18n.Translate(language, "顾客")
			realName = fmt.Sprintf("%s(%s)", prefix, record.Member.Nickname)
		}

		// 如果是系统自动操作
		if record.Source == "" {
			realName = i18n.Translate(language, "系统自动")
		}

		logs = append(logs, resp.OrderOperationLog{
			Uuid:       record.Uuid,
			RealName:   realName,
			Email:      record.Operator.Username,
			Source:     i18n.Translate(language, constant.SourceTextMap[record.Source]),
			CreateTime: record.CreateTime,
			RefundType: func() int {
				var refundType int
				if record.Action == constant.OrderRefund {
					var refundPayload event.ReturnOrderPayload
					json.Unmarshal([]byte(record.Data), &refundPayload)
					refundType = refundPayload.RefundType
				}
				return refundType
			}(),
			Description: actionText, // 获取描述
			PayType:     s.getRefundPayType(ctx, record, language),
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
	if slices.Contains([]string{constant.SourceShop}, ctx.GetSource()) {
		if orderRepo.IsPartiallyPaid(saleBillUuid) {
			return model.SaleBill{}, errors.New("当前订单已被部分支付，不支持取消")
		}
	}
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderOrderCancel, 0); err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
	}
	if !slices.Contains([]string{constant.SourceShop}, ctx.GetSource()) {
		if orderRepo.IsPartiallyPaid(saleBillUuid) {
			return model.SaleBill{}, errors.New("当前订单已被部分支付，不支持取消")
		}
	}
	return billInfo, nil
}

// RejectAllH5Order 拒绝某销售账单的所有未接单h5订单
func (s *orderSrv) RejectAllH5Order(ctx context.Context, saleBillUuid uint64) error {
	db := ctx.GetDB()
	// 获取所有待接单的h5订单
	h5Orders, err := repository.NewH5OrderRepo(db).GetH5OrderListBySaleBillUuid(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, h5Order := range h5Orders {
		err := s.RejectH5Order(ctx, h5Order.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// RejectAllH5OrderInShop 将商家的所有待接单的h5订单都拒单
func (s *orderSrv) RejectAllH5OrderInShop(ctx context.Context) error {
	db := ctx.GetDB()
	// 查询所有还在进行中的桌台账单
	saleBills, err := repository.NewSaleBillRepo(db).GetDeskSaleBillUnPay()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, saleBill := range saleBills {
		s.RejectAllH5Order(ctx, saleBill.Uuid)
	}
	return nil
}

// CancelOrder 取消订单
func (s *orderSrv) CancelOrder(ctx context.Context, req req.OrderCancelReq) error {
	dbId := ctx.GetDbId()
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

	// 退回商品库存
	{
		// 获取销售账单信息
		saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		s.returnInventory(ctx, saleBill)
	}

	// 如果是桌台订单
	if billInfo.IsDeskBill() && billInfo.DeskUuid > 0 {
		// 拒绝所有待接单
		ctx.SetDB(tx)
		if err := s.RejectAllH5Order(ctx, billInfo.Uuid); err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
		// 关闭桌台
		err = deskRepo.CloseDesk(ctx, billInfo.DeskUuid, billInfo.Uuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	} else {
		err = orderRepo.CancelOrder(ctx, req.SaleBillUuid, 0, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	}

	// 标记送厨单、送厨商品为删除
	productionRepo := repository.NewProductionRepo(tx)
	saleBillUuidOpt := productionRepo.WhereSaleBillUuid(billInfo.Uuid)
	err = productionRepo.UpdateOrder([]repository.DBOption{saleBillUuidOpt}, map[string]any{
		"delete_time": time.Now().Unix(),
	})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单失败"), err.Error())
	}
	// 修改送厨商品数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
	err = productionRepo.UpdateProduct([]repository.DBOption{saleBillUuidOpt}, map[string]any{
		"num": 0,
	})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单商品失败"), err.Error())
	}

	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "销售账单不存在")
	}
	saleBill.SetCanceled()
	// 计算订单商品、订单、账单金额并更新或创建
	if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill, model.WithCanceled()); err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithMessage(err)
	}

	// 发布"整单取消"操作事件
	go func() {
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{ // 整单取消
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: billInfo.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	// 成功后，推送到厨显端更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
	})

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
	// TODO 获取还在制作中的商品

	// 发布事件，通知厨房取消制作
	event.NewSystemBus().PublishCancelDoingProductEvent(event.CancelDoingProductPayload{SaleOrderProductUuids: doingProductList})
	return nil
}

// ReturnOrder 退款订单
func (s *orderSrv) ReturnOrder(ctx context.Context, req req.OrderReturnReq) (error, int) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("ReturnOrder process, GetStoreSetting failed", zap.Error(err))
		return errors.WithMessage(err), constant.CodeFail
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return errors.WithMessage(errors.New("找不到销售订单")), constant.CodeFail
	}

	if req.Points > saleOrder.GetManualReturnPoints() {
		return errors.WithMessage(errors.New("退款积分不能大于最大可退积分")), constant.CodeFail
	}

	returnType := constant.ReturnOrderRefundTypeTotal
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)                       // 退款商品列表
	saleOrderBuffetCustomerTypes := make([]*model.SaleOrderBuffetCustomerType, 0) // 退款自助餐顾客列表
	saleOrderBuffetDelayProducts := make([]*model.SaleOrderBuffetDelayProduct, 0) // 退款自助餐延迟商品列表
	numMap := make(map[uint64]float64)                                            // 每个退款商品的退货数量
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

		for _, saleOrderProduct := range saleOrder.SaleOrderBuffetCustomerTypes {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = float64(canReturnNum)
			}
		}

		for _, saleOrderProduct := range saleOrder.SaleOrderBuffetDelayProducts {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderBuffetDelayProducts = append(saleOrderBuffetDelayProducts, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = float64(canReturnNum)
			}
		}
	}
	// 部分退款
	isPartReturn := false // 部分退款时，如果有赠菜，先不传给erp
	if len(req.Products) > 0 {
		returnType = constant.ReturnOrderRefundTypePart
		isPartReturn = true
		// 获取退款商品列表
		saleOrderProductUuids := make([]uint64, 0)
		for _, product := range req.Products {
			saleOrderProductUuids = append(saleOrderProductUuids, product.SaleOrderProductUuid)
			numMap[product.SaleOrderProductUuid] = product.Num
		}
		// 注意：要判断订单商品是否还有可退货数量
		saleOrderProductList := saleOrder.GetSaleOrderProductList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderProductList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if num <= canReturnNum {
					saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("1退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

		saleOrderBuffetCustomerTypeList := saleOrder.GetSaleOrderBuffetComstomerTypeList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderBuffetCustomerTypeList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if uint(num) <= canReturnNum {
					saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("2退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

		saleOrderBuffetDelayProductsList := saleOrder.GetSaleOrderBuffetDelayList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderBuffetDelayProductsList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if num <= float64(canReturnNum) {
					saleOrderBuffetDelayProducts = append(saleOrderBuffetDelayProducts, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("3退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

	}

	// 如果退款类型为部分退款，则必须有可退货的商品。整单退款则可以没有可退货的商品，可能已经退完商品了但还有手续费没有退
	if len(saleOrderProducts) == 0 && len(saleOrderBuffetCustomerTypes) == 0 && len(saleOrderBuffetDelayProducts) == 0 && returnType == constant.ReturnOrderRefundTypePart {
		return errors.WithMessage(errors.New("没有可退货的商品")), constant.CodeFail
	}

	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmount()
	// 如果是会员端订单，则需要根据配送费计算可退款金额
	deliveryFee := 0.0
	if ctx.GetScene() == constant.SceneMemberOrder {
		memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecordOnlyBySaleBillUuid(req.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err), constant.CodeFail
		}
		deliveryFee = memberSaleOrder.DeliveryFeeAmount
		canReturnAmount = saleOrder.GetCanReturnAmountWithDeliveryFee(deliveryFee)
	}
	// 可退的会员消费金额
	canReturnMemberConsumptionAmount := saleOrder.GetCanReturnMemberConsumptionAmount()

	// 获取当前员工班次信息
	var staffShiftLogUuid uint64
	if ctx.GetStaffUuid() != 0 {
		staffShiftLog, err := GetCurrentStaffShiftLog(db, ctx.GetStaffUuid())
		if err == nil {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
	}

	// 创建退款单
	returnOrder, err := saleOrder.NewReturnOrder(ctx.GetScene(), deliveryFee, ctx.GetStaff().DutyNo, ctx.GetLanguage(), saleOrderProducts, saleOrderBuffetCustomerTypes, saleOrderBuffetDelayProducts, numMap, returnType, canReturnAmount, staffShiftLogUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 本次退款的会员累计消费金额。=退款金额
	returnOrderMemberConsumptionAmount := returnOrder.RefundAmount
	if returnOrderMemberConsumptionAmount > canReturnMemberConsumptionAmount {
		// 本次退款的会员累计消费金额不能大于可退的会员消费金额
		returnOrderMemberConsumptionAmount = canReturnMemberConsumptionAmount
	}

	// 是否存在QrPromptPay支付
	if returnOrder.IsExistQrPromptPay() {
		if req.BankCode == "" || req.AccountNo == "" || req.AccountName == "" {
			return errors.WithMessage(errors.New("请选择银行")), constant.CodeReturnOrderBank
		}
		returnOrder.BankCode = req.BankCode
		returnOrder.AccountNo = req.AccountNo
		returnOrder.AccountName = req.AccountName
	}

	lianLianPayCount := returnOrder.GetLianLianPayCount()

	var publishChangeMemberBalance, publishChangeMemberPoints, isExistCashPay bool
	// 创建
	err = repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		ctx.SetDB(db) // 否则 s.memberSrv.HandleMemberBalance会事务失效

		// 创建退货单
		if _, err = repository.NewReturnOrderRepo(db).CreateReturnOrderRecord(*returnOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 创建连连退款订单
		for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
			if lianLianPayCount > 0 && returnOrderAmount.PaymentMethod.IsLianLianPay() {
				paymentServiceRefundReq := PaymentServiceRefundReq{
					RelatedType:           constant.PaymentOrderRelatedTypeSaleOrder,
					PaymentOrderUuid:      returnOrderAmount.PaymentOrderUuid,
					MerchantRefundOrderNo: returnOrderAmount.MerchantRefundOrderNo,
					RefundAmount:          returnOrderAmount.Amount,
					BankCode:              returnOrder.BankCode,
					AccountNo:             returnOrder.AccountNo,
					AccountName:           returnOrder.AccountName,
				}
				if lianLianPayCount > 1 {
					go func() {
						payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
						if err != nil {
							returnOrderAmount.RefundStatus = 2
							returnOrderAmount.LlReturnOrderid = "0"
						} else {
							returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
						}
						// 更新退款状态
						returnOrderRepo := repository.NewReturnOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
						returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{
							returnOrderRepo.WhereUuid(returnOrderAmount.Uuid),
						}, returnOrderAmount)
						if err != nil {
							fmt.Println("更新退款状态失败", err)
							logger.Logger.Error("更新退款状态失败", zap.Error(err))
						}
					}()
				} else {
					payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
					if err != nil {
						return errors.WithMessage(err)
					}
					// 设置连连退款订单ID
					returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
				}
			} else {
				returnOrderAmount.RefundStatus = 1
			}
			// 创建退款金额
			if err = repository.NewReturnOrderRepo(db).CreateReturnOrderAmount([]model.ReturnOrderAmount{returnOrderAmount}); err != nil {
				return errors.WithMessage(err)
			}
			// 如果退款金额为余额，则退回余额，创建余额变动记录
			if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
				if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
					MemberUuid:  returnOrderAmount.MemberBalanceLog.MemberUuid,
					GiftMoney:   returnOrderAmount.MemberBalanceLog.GiftMoney, // 退款金额。余额退款都是退回到赠送帐户
					Scene:       returnOrderAmount.MemberBalanceLog.Scene,
					Describe:    returnOrderAmount.MemberBalanceLog.Describe,
					RelatedUuid: returnOrderAmount.MemberBalanceLog.RelatedUuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
				publishChangeMemberBalance = true
			}
			// 如果退款金额为现金，则更新钱箱
			if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				isExistCashPay = true
				// 存现金，更新钱箱
				ctx.SetDB(db)
				if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
					Amount:    -returnOrderAmount.Amount,
					Scene:     constant.CashBoxLogSceneRefund,
					OrderUuid: returnOrderAmount.Uuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 创建退货单商品
		if err = repository.NewReturnOrderRepo(db).CreateReturnOrderProduct(returnOrder.ReturnOrderProducts); err != nil {
			return errors.WithMessage(err)
		}
		// 更新高峰时段
		if err := repository.NewSaleOrderPeakTimeRepo(db).Record("dec", saleBill, returnOrder.RefundAmount, storeSetting.TimeZone); err != nil {
			return errors.WithMessage(err)
		}
		// 退积分
		if saleOrder.ConsumerUuid > 0 {
			// 手动退积分
			if saleOrder.CanManualReturnPoints() {
				if req.Points > 0 {
					points := req.Points
					// 开始手动退积分
					member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
					if err != nil {
						return errors.WithMessage(err)
					}
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewRefundMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			} else {
				// 自动退积分
				refundAmount := returnOrder.RefundAmount // 退款金额
				// 积分赠送比例
				integralGiveRate := saleOrder.GiftPointsRate
				//  部分退款时。退积分=退款金额*积分赠送比例
				points := decimal.NewFromFloat(refundAmount).Mul(decimal.NewFromFloat(integralGiveRate)).Round(2).InexactFloat64()
				// 退还积分不能超过可退积分
				if points > saleOrder.GetManualReturnPoints() {
					points = saleOrder.GetManualReturnPoints()
				}
				// 如果退款类型为整单退款，则退还积分剩余未退的积分
				if len(req.Products) == 0 {
					points = saleOrder.GetManualReturnPoints()
				}
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					return errors.WithMessage(err)
				}
				// 更新会员积分
				if points > 0 {
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewRefundMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
			publishChangeMemberPoints = true
		}

		// 退会员的累计消费金额。退款后减少会员的累计消费金额
		if saleOrder.ConsumerUuid > 0 {
			// 减少会员累计消费金额
			if err := repository.NewMemberRepo(db).DecConsumptionAmount(saleOrder.ConsumerUuid, returnOrderMemberConsumptionAmount); err != nil {
				return errors.WithMessage(err)
			}
			// 如果是整单退款，则减少会员累计消费次数
			if len(req.Products) == 0 {

			}
		}

		// 保存发票到erp
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			res, err := s.ReturnPosInvoice(ctx, saleOrder, returnOrder, db, returnType, isPartReturn)
			if err != nil {
				return errors.WithMessage(err)
			}
			returnOrder.ErpInvoiceName = res.InvoiceName
		}
		// 更新退货单erp发票名
		if err := repository.NewReturnOrderRepo(db).UpdateReturnOrderRecordErpInvoiceName(returnOrder.Uuid, returnOrder.ErpInvoiceName); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	})

	// 发送短信
	go func() {
		// 获取最新的会员信息
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
		if err != nil {
			ctx.Log().Info("停止发送短信（退款），获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			refundAmount := float64(0)
			for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
				if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
					refundAmount = returnOrderAmount.Amount
					break
				}
			}
			if refundAmount > 0 {
				if member != nil {
					smsReq := sms.MemberOrderRefundRequest{
						Company:       ctx.GetCompany().Name,
						OrderRefund:   refundAmount,
						Balance:       member.GetBalanceAll(),
						PointsBalance: member.GetPoints(),
					}
					if err := s.smsSrv.SendMemberOrderRefundSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送退款短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送退款短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			}
		}
	}()

	if publishChangeMemberBalance {
		// 发布"会员余额变动"事件
		go func() {
			s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
				BasePayload: event.BasePayload{ // 会员余额变动
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		}()
	}

	if publishChangeMemberPoints {
		// 发布"会员积分变动"事件
		go func() {
			s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
				BasePayload: event.BasePayload{ // 会员积分变动
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		}()
	}

	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	// 发布"退款"事件
	products := make(event.Products, 0)
	for _, saleOrderProduct := range saleOrderProducts {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId:  saleOrderProduct.Uuid,
				ProductId:       saleOrderProduct.ProductPackageUuid,
				ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
				ProductAttr:     saleOrderProduct.GetAttributeName(),
				ProductAttrList: saleOrderProduct.GetAttributeNameList(),
				TotalNum:        num,
				NumType:         saleOrderProduct.NumType,
				IsBuffet:        saleOrderProduct.IsBuffet == 1,
				IsWrap: func() bool {
					if saleBill.IsTakeout() && saleBill.MemberSaleOrderUuid == 0 {
						return true
					}
					return saleOrderProduct.IsWrapProduct()
				}(),
				Remark: saleOrderProduct.Remark,
			})
		}
	}
	for _, saleOrderProduct := range saleOrderBuffetCustomerTypes {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId: saleOrderProduct.Uuid,
				ProductId:      saleOrderProduct.BuffetCustomerTypePriceUuid,
				ProductName:    saleOrderProduct.BuffetPackage.MultiLanguageName.GetNames(),
				ProductAttr: dto.LocaleResponse{
					ZH:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					TH:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					EN:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					ZHTW: saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					JA:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					KO:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					MY:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					TR:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					SV:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				},
				TotalNum: num,
			})
		}
	}
	for _, saleOrderProduct := range saleOrderBuffetDelayProducts {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId: saleOrderProduct.Uuid,
				ProductId:      saleOrderProduct.BuffetDelayUuid,
				ProductName: dto.LocaleResponse{
					ZH:   saleOrderProduct.Name,
					TH:   saleOrderProduct.Name,
					EN:   saleOrderProduct.Name,
					ZHTW: saleOrderProduct.Name,
					JA:   saleOrderProduct.Name,
					KO:   saleOrderProduct.Name,
					MY:   saleOrderProduct.Name,
					TR:   saleOrderProduct.Name,
					SV:   saleOrderProduct.Name,
				},
				TotalNum: num,
			})
		}
	}
	var payTypes []event.RefundPayType
	for _, amount := range returnOrder.ReturnOrderAmounts {
		payTypes = append(payTypes, event.RefundPayType{
			Name:              amount.PaymentMethod.PaymentName,
			Code:              amount.PaymentMethod.Code,
			Amount:            amount.Amount,
			RefundStatus:      amount.RefundStatus,
			ReturnAmountUuid:  amount.Uuid,
			ReturnOrderUuid:   amount.ReturnOrderUuid,
			PaymentOrderUuid:  amount.PaymentOrderUuid,
			PaymentMethodUuid: amount.PaymentMethodUuid,
		})
	}

	// 记录外送订单的退款金额
	if saleBill.MemberSaleOrderUuid > 0 {
		go func() {
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRefundAmount(saleBill.MemberSaleOrderUuid, returnOrder.RefundAmount); err != nil {
				ctx.Log().Error("记录外送订单的退款金额失败", zap.Error(errors.WithMessage(err)))
			}
		}()
	}

	go func() {
		s.bus.PublishReturnOrderEvent(event.ReturnOrderPayload{
			SaleBill: saleBill,
			BasePayload: event.BasePayload{ // 退款
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  saleBill.Uuid,
				SaleOrderUuid: saleOrder.Uuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Products:   products,
			PayTypes:   payTypes,
			RefundType: returnType,
		})
	}()
	// 发布"统计"事件
	go func() {
		s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			SaleBillUuid: saleBill.Uuid,
		})
	}()

	if isExistCashPay {
		return nil, constant.CodeSuccessOpenCashBox
	}

	return nil, 0
}

// GetCurrentStaffShiftLog 获取当前员工班次信息
func GetCurrentStaffShiftLog(db *gorm.DB, staffUuid uint64) (*model.StaffShiftLog, error) {
	shiftLogRepo := repository.NewShiftLogRepo(db)

	// 查询当前员工未交班的班次记录
	shiftLog, err := shiftLogRepo.GetShiftLog(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("staff_uuid = ?", staffUuid)
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", constant.StaffNotHandedOver)
		},
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("员工当前没有进行中的班次")
		}
		return nil, errors.WithMessage(err)
	}

	return &shiftLog, nil
}

// ReReturnOrder 重新退款
func (s *orderSrv) ReReturnOrder(ctx context.Context, req req.OrderReReturnReq) (error, int) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.ReturnOrderUuid)
		defer lock.NewSystemLock().UnlockUuid(req.ReturnOrderUuid)
		ctx.AddLock()
	}
	// 获取退款订单信息
	returnOrderRepo := repository.NewReturnOrderRepo(ctx.GetDB())
	orderAmount, err := returnOrderRepo.GetReturnOrderAmount(
		returnOrderRepo.WithReturnOrder(),
		returnOrderRepo.WithPaymentMethod(),
		returnOrderRepo.WhereUuid(req.ReturnAmountUuid),
	)
	if err != nil || orderAmount.ReturnOrder.Uuid != req.ReturnOrderUuid {
		return errors.New("找不到订单"), constant.CodeFail
	}
	if orderAmount.RefundStatus == 1 {
		return errors.New("该订单已成功退款，无法重复退款"), constant.CodeFail
	}
	if !orderAmount.PaymentMethod.IsLianLianPay() {
		return errors.New("该订单无法重新退款"), constant.CodeFail
	}
	// 判断订单是否正在退款
	if orderAmount.RefundStatus == 0 {
		return errors.New("该订单正在进行退款，无法重复操作"), constant.CodeFail
	}

	refundReq := PaymentServiceRefundReq{
		RelatedType:           constant.PaymentOrderRelatedTypeSaleOrder,
		PaymentOrderUuid:      orderAmount.PaymentOrderUuid,
		MerchantRefundOrderNo: orderAmount.MerchantRefundOrderNo,
		RefundAmount:          orderAmount.Amount,
		RefundOrderId:         orderAmount.LlReturnOrderid,
	}

	// 是否存在QrPromptPay支付
	isChangeBankCode := false
	if orderAmount.PaymentMethod.IsQrPromptPay() {
		if req.BankCode == "" || req.AccountNo == "" || req.AccountName == "" {
			return errors.WithMessage(errors.New("请选择银行")), constant.CodeReturnOrderBank
		}
		if req.BankCode != orderAmount.ReturnOrder.BankCode || req.AccountNo != orderAmount.ReturnOrder.AccountNo || req.AccountName != orderAmount.ReturnOrder.AccountName {
			isChangeBankCode = true
		}
		refundReq.BankCode = orderAmount.ReturnOrder.BankCode
		refundReq.AccountNo = orderAmount.ReturnOrder.AccountNo
		refundReq.AccountName = orderAmount.ReturnOrder.AccountName
	}

	// 发起退款
	refund, err := NewPaymentRepo(ctx, s.dbm).Refund(refundReq)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	if refund.RefundStatus == "RP" {
		return errors.New("该订单正在进行退款，无法重复操作"), constant.CodeFail
	}
	if refund.RefundStatus == "RS" {
		orderAmount.RefundStatus = 1
		err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
		if err != nil {
			return errors.WithMessage(err), constant.CodeFail
		}
		return errors.New("该订单已成功退款，无法重复退款"), constant.CodeFail
	}
	// 更新银行信息 - 重新发起退款
	if isChangeBankCode {
		orderAmount.RefundStatus = 1
		orderAmount.MerchantRefundOrderNo = utils.GenerateMerchantOrderNo("RE")
		// 更新退款订单号
		refundReq.MerchantRefundOrderNo = orderAmount.MerchantRefundOrderNo
		refundReq.BankCode = req.BankCode
		refundReq.AccountNo = req.AccountNo
		refundReq.AccountName = req.AccountName
		// 更新银行信息
		orderAmount.ReturnOrder.BankCode = req.BankCode
		orderAmount.ReturnOrder.AccountNo = req.AccountNo
		orderAmount.ReturnOrder.AccountName = req.AccountName
	}
	// 重新发起退款
	refund, err = NewPaymentRepo(ctx, s.dbm).Refund(refundReq)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	// 更新退款订单号
	orderAmount.LlReturnOrderid = refund.RefundOrderId
	err = returnOrderRepo.UpdateReturnOrder([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.ReturnOrder.Uuid)}, *orderAmount.ReturnOrder)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	//
	return nil, 0
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

	products := make([]resp.OrderReturnProduct, 0)

	// 获取销售订单的每个付款单的可退款金额
	// 要求排好序：退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	paymentRecords, currencyUnit := saleOrder.GetPaymentOrderCanReturnAmount()

	// 获取销售订单的自助餐顾客列表
	buffetCustomers := saleOrder.GetCustomerList()
	for _, buffetCustomer := range buffetCustomers {
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: buffetCustomer.Uuid,
			LocaleName:           buffetCustomer.LocaleName,
			LocaleAttributeName:  buffetCustomer.LocaleAttributeName,
			Num:                  float64(buffetCustomer.CanReturnNum), // 自助餐顾客类型可退货数量
			Price:                buffetCustomer.TotalPrice,            // 自助餐顾客类型总价（单个商品、折后）
			CanReturnAmount:      buffetCustomer.CanReturnAmount,       // 自助餐顾客类型可退款金额
			CurrencyUnit:         currencyUnit,
		})
	}

	// 获取销售订单的加钟商品列表
	delayProducts := saleOrder.GetDelayProductList()
	for _, delayProduct := range delayProducts {
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: delayProduct.Uuid,
			LocaleName:           delayProduct.LocaleName,
			LocaleAttributeName:  delayProduct.LocaleAttributeName,
			Num:                  float64(delayProduct.CanReturnNum), // 加钟商品可退货数量
			Price:                delayProduct.UnitPrice,             // 加钟商品单价
			CanReturnAmount:      delayProduct.CanReturnAmount,       // 加钟商品可退款金额
			CurrencyUnit:         currencyUnit,
		})
	}

	// 获取销售订单商品列表
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		if saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsGiftProduct() || saleOrderProduct.Status == constant.OrderProductStatusUnSending {
			continue
		}
		if saleOrderProduct.IsPackageSubProduct() {
			continue
		}
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(),
			Num:                  saleOrderProduct.GetCanReturnNum(), // 可退货数量=订单商品数量-已退货数量
			NumType:              saleOrderProduct.NumType,
			Price:                saleOrderProduct.TotalPrice,
			CanReturnAmount:      saleOrderProduct.GetCanReturnPrice(),
			CurrencyUnit:         currencyUnit,
		})
	}

	// 过滤掉单价为0的商品
	productList := make([]resp.OrderReturnProduct, 0)
	for _, product := range products {
		if product.Price == 0 {
			continue
		}
		productList = append(productList, product)
	}

	// 获取销售订单付款单列表
	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmount()
	res := &resp.OrderReturnInfoResp{
		ManualReturnPoints: saleOrder.CanManualReturnPoints(), // 是否可以手动退款积分。订单是按比例赠送积分且未发生积分抵扣时，不自动退款。
		DeductiblePoints:   saleOrder.GetManualReturnPoints(), // 可扣除积分。订单赠送的积分-已经退回的积分
		CanReturnAmount:    canReturnAmount,                   // 可退款金额. 可退款金额=订单最终应收金额-已退款金额
		PaymentRecords:     paymentRecords,
		Products:           productList,
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
	payMethods := saleBill.GetPaymentMethodNameList(ctx.GetLanguage())

	return &resp.OrderReverseSettleInfoResp{
		SaleBillUuid:    saleBill.Uuid,
		SaleBillNo:      saleBill.OrderNo,
		SaleBillType:    saleBill.BillType,
		OrderAmount:     saleBill.OriginAmount,
		PaymentAmount:   saleBill.PaymentAmount,
		PayMethods:      payMethods,
		Desks:           resDesks,
		HasInstantOrder: hasInstantOrder,
	}, nil
}

// ReverseSettle 处理反结账
func (s *orderSrv) ReverseSettle(ctx context.Context, request req.OrderReverseSettleReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetStoreSetting failed", zap.Error(err))
		return errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if saleBill.IsDeskSaleBill() {
		if request.DeskUuid == 0 {
			return errors.WithMessage(errors.New("桌台UUID不能为0"))
		}
	}

	// 销售账单状态变为未结账状态
	// 销售订单状态变为未结账状态
	// 销售订单的所有付款单都退款，并生成退款单
	// 反结账次数+1
	saleBill.SetReverseSettle()

	// 如果销售账单是桌台订单，则开桌
	// 开桌
	var desk *model.Desk
	if saleBill.IsDeskSaleBill() {
		deskRepo := repository.NewDeskRepo(db)
		desk, err = deskRepo.GetDeskRecord(request.DeskUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if !desk.IsAvailableDesk() {
			return errors.WithMessage(errors.New("桌台非空闲"))
		}
		desk.SetOpenDesk(saleBill.Uuid)
		saleBill.DeskUuid = desk.Uuid
		saleBill.SerialNo = desk.DeskNo
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
			if request.HideOrder {
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
	if err := s.returnInventory(ctx, saleBill, WithReverseSettle()); err != nil {
		return errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 删除销售订单原料
		if err := repository.NewSaleOrderMaterialRepo(db).DeleteSaleOrderMaterial(saleBill.Uuid); err != nil {
			return errors.WithMessage(err)
		}
		// 如果销售订单是免单，删除免单原因
		for _, saleOrder := range saleBill.SaleOrders {
			if saleOrder.IsFreeSaleOrder() {
				saleOrder.SetCancelFreeOrder()
				// 删除销售订单的免单原因
				if err := repository.NewSaleOrderProductReasonRepo(db).DeleteFreeReason(saleOrder.Uuid); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 如果存在需要挂单的销售账单，则更新该销售账单
		if hideSaleBill != nil {
			if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*hideSaleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新高峰时段
		if err := repository.NewSaleOrderPeakTimeRepo(db).Record("dec", saleBill, 0, storeSetting.TimeZone); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售账单
		if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
			return errors.WithMessage(err)
		}

		giftPointsMap := make(map[uint64]float64) // sale_order_uuid -> gift_points
		// 更新销售订单
		for _, saleOrder := range saleBill.SaleOrders {
			giftPointsMap[saleOrder.Uuid] = saleOrder.GiftPoints // 清空之前记录的赠送积分
			// 将结账才记录的值清空
			saleOrder.ClearSettleInfo()
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新支付订单,状态为已退款
		for _, saleOrder := range saleBill.SaleOrders {
			for _, paymentOrder := range saleOrder.PaymentOrders {
				if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*paymentOrder); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 生成退款单
		for _, saleOrder := range saleBill.SaleOrders {
			isUseMember := saleOrder.ConsumerUuid != 0
			for _, paymentOrder := range saleOrder.PaymentOrders {
				refundOrder := paymentOrder.RefundOrder
				if err := repository.NewPaymentOrderRepo(db).CreateRefundOrderRecord(*refundOrder); err != nil {
					return errors.WithMessage(err)
				}
				// 如果是余额支付，则退款到余额
				if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
					s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
						MemberUuid:  saleOrder.ConsumerUuid,
						Money:       paymentOrder.BalanceAmount,
						GiftMoney:   paymentOrder.GiftBalanceAmount,
						Scene:       constant.MemberBalanceLogReverse,
						Describe:    fmt.Sprintf("订单反结账：%s", saleOrder.OrderNo),
						RelatedUuid: saleOrder.Uuid,
					})
				}
				// 取现金，更新钱箱
				if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
					if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
						Amount:    -paymentOrder.Amount,
						Scene:     constant.CashBoxLogSceneRefund,
						OrderUuid: saleOrder.Uuid,
					}); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
			// 退会员的累计消费金额
			if isUseMember {
				// 减少会员累计消费金额
				if err := repository.NewMemberRepo(db).DecConsumptionAmount(saleOrder.ConsumerUuid, saleOrder.GetCanReturnMemberConsumptionAmountMax()); err != nil {
					return errors.WithMessage(err)
				}
				// 减少会员累计消费次数
				if err := repository.NewMemberRepo(db).DecConsumptionCount(saleOrder.ConsumerUuid); err != nil {
					return errors.WithMessage(err)
				}
			}
			// 退积分
			if isUseMember {
				points := giftPointsMap[saleOrder.Uuid]
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					return errors.WithMessage(err)
				}
				// 如果订单有积分抵扣，则已抵扣的积分
				if saleBill.SaleBillSetting.IsOpenPointsExchange() && saleOrder.PayPoints > 0 {
					// 更新会员积分
					member.FrozenPoint = member.FrozenPoint + saleOrder.PayPoints // 退回已抵扣的积分
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint, // 退回已抵扣的积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewReverseSettleExchangeMemberPointLog(saleOrder.PayPoints)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
				if points > 0 {
					// 如果会员积分余额不足时，仅扣完余额
					if member.GetPoints() < points {
						points = member.GetPoints()
					}
					// 更新会员积分
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewReverseSettleMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			}

			// 退优惠券。如果订单使用了优惠券，需要将优惠券退还给会员。如果使用了通用优惠券，则通用优惠券余量+1并生成记录
			if saleOrder.HasCoupon() {
				// 加锁, 避免并发问题,避免数量+1重复或失效
				lock.NewSystemLock().LockUuid(constant.LockNameActivityConsumption)
				defer lock.NewSystemLock().UnlockUuid(constant.LockNameActivityConsumption)
				for _, coupon := range saleOrder.Coupons {
					if !coupon.IsDelete() {
						if coupon.IsCommonCoupon() {
							commonCoupon, err := repository.NewMarketingCouponRepo(db).GetCouponByUuid(coupon.MarketingCouponUuid)
							if err != nil {
								return errors.WithMessage(err)
							}
							commonCoupon.Count = commonCoupon.Count + 1
							// 取消核销通用优惠券，数量+1
							if err := repository.NewMarketingCouponRepo(db).UpdateCommonCouponCountCancel(coupon.MarketingCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
							// 创建通用优惠券记录，记录类型：反结账退还
							if err := repository.NewMarketingCouponRepo(db).CreateCommonCouponRecordCancel(coupon.MarketingCouponUuid, commonCoupon.Count); err != nil {
								return errors.WithMessage(err)
							}
						}
						if coupon.IsMemberCoupon() {
							// 取消核销会员优惠券
							if err := repository.NewMemberCouponRepo(db).CancelVerifyMemberCoupon(coupon.MemberCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
							// 删除会员优惠券使用记录
							if err := repository.NewMemberCouponRepo(db).DeleteMemberCouponRecord(coupon.MemberCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
						}
					}
				}
			}

			// 发布"会员余额变动"事件
			go func() {
				s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
					BasePayload: event.BasePayload{ // 会员余额变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						SaleBillUuid: request.SaleBillUuid,
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			}()

			// 发布"会员积分变动"事件
			go func() {
				s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
					BasePayload: event.BasePayload{ // 会员积分变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						SaleBillUuid: request.SaleBillUuid,
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			}()

			// 发送短信
			if isUseMember {
				go func() {
					newCtx := ctx.Copy()
					newCtx.SetDB(s.dbm.GetDB(newCtx.GetDbId()))
					// 获取最新的会员信息
					member, err := repository.NewMemberRepo(newCtx.GetDB()).GetMemberByUuid(saleOrder.ConsumerUuid)
					if err != nil {
						ctx.Log().Info("停止发送短信（消费反结账），获取会员失败", zap.Error(errors.WithMessage(err)))
					} else {
						refundAmount := float64(0)
						for _, paymentOrder := range saleOrder.PaymentOrders {
							// 如果是余额支付，则退款到余额
							if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
								refundAmount = decimal.NewFromFloat(paymentOrder.BalanceAmount).Add(decimal.NewFromFloat(paymentOrder.GiftBalanceAmount)).Truncate(2).InexactFloat64()
							}
						}
						if refundAmount > 0 {
							if member != nil {
								smsReq := sms.MemberOrderRefundRequest{
									Company:       ctx.GetCompany().Name,
									OrderRefund:   refundAmount,
									Balance:       member.GetBalanceAll(),
									PointsBalance: member.GetPoints(),
								}
								if err := s.smsSrv.SendMemberOrderRefundSMS(newCtx, member.Phone, &smsReq); err != nil {
									ctx.Log().Info("发送退款短信失败（消费反结账）", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
								} else {
									ctx.Log().Info("发送退款短信成功（消费反结账）", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
								}
							}
						}
					}
				}()
			}
		}

		go func() {
			// 发布"反结账"操作事件
			var payTypes []event.PayType
			for _, order := range saleBill.SaleOrders {
				if order.IsFree == 1 {
					payTypes = append(payTypes, event.PayType{
						Name:  "免单",
						Value: constant.PaymentMethodCodeFreePay,
						Price: order.GetAmount(),
					})
				} else {
					if infoResp, err := s.InstantOrderPaymentInfo(ctx, saleBill, request.SaleBillUuid, order.Uuid); err == nil {
						for _, paymentOrder := range infoResp.PaymentOrders.List {
							payTypes = append(payTypes, event.PayType{
								Name:           paymentOrder.PaymentMethodName,
								Value:          paymentOrder.PaymentMethodCode,
								DisabledCancel: utils.BoolToUint(paymentOrder.DisabledCancel),
								Price:          paymentOrder.Amount,
								FeeMoney:       paymentOrder.PaymentCommissionFee,
							})
						}
					}
				}
			}
			s.bus.PublishOrderReverseSettleEvent(event.OrderReverseSettlePayload{
				BasePayload: event.BasePayload{ // 订单反结账
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
				PayTypes: payTypes,
			})
		}()

		// 更新桌台
		if saleBill.IsDeskSaleBill() {
			if err := repository.NewDeskRepo(db).UpdateDeskRecord(*desk); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐销量
		if saleBill.IsBuffetSaleBill() {
			if saleBill.BuffetPackage1Uuid != 0 {
				saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage1Uuid)
				if err := repository.NewBuffetRepo(db).SubActualSaleNum(saleBill.BuffetPackage1Uuid, saleNum); err != nil {
					fmt.Println(err)
					ctx.Log().Error("SubActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
				}
			}
			if saleBill.BuffetPackage2Uuid != 0 {
				saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage2Uuid)
				if err := repository.NewBuffetRepo(db).SubActualSaleNum(saleBill.BuffetPackage2Uuid, saleNum); err != nil {
					ctx.Log().Error("SubActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
				}
			}
		}

		go func() {
			// 发布"统计"操作事件
			s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
				BasePayload: event.BasePayload{ // 统计
					Ctx: ctx,
				},
				SaleBillUuid: saleBill.Uuid,
				OnlyDelete:   true,
			})
		}()

		// 在ERP取消发票
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			staff := ctx.GetStaff()
			shiftLogRepo := repository.NewShiftLogRepo(db)
			shiftLog, err := shiftLogRepo.GetShiftLog(
				repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
				repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
			)
			if err != nil {
				return errors.WithMessage(err)
			}
			if shiftLog.IsHandedOver() {
				return errors.New("当前班次已交班，无法保存发票")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			for _, saleOrder := range saleBill.SaleOrders {
				if saleOrder.IsDelete() {
					continue
				}
				err := erpSrv.CancelPosInvoice(ctx, req.CancelPosInvoiceReq{
					ProductsInvoiceName: saleOrder.ErpProductsInvoiceName,
					MaterialInvoiceName: saleOrder.ErpMaterialInvoiceName,
					OpenPosEntryName:    shiftLog.ErpnextOpenPosEntryName, //异步模式必填
					OrderNo:             saleOrder.OrderNo,                //异步模式必填
				})
				if err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

type returnInventoryOptions struct {
	IsReverseSettle bool // 是否是反结账
}

func WithReverseSettle() func(opts *returnInventoryOptions) {
	return func(opts *returnInventoryOptions) {
		opts.IsReverseSettle = true
	}
}

// returnInventory 退回库存
// 取消订单时，将所有商品都退回库存
// 反结账时，将先将所有商品都退回库存，再将下单减库存的商品扣除库存(出库)
func (s *orderSrv) returnInventory(ctx context.Context, saleBill *model.SaleBill, opts ...func(opts *returnInventoryOptions)) error {
	opt := &returnInventoryOptions{}
	for _, o := range opts {
		o(opt)
	}
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
		var products []*model.SaleOrderProduct
		if opt.IsReverseSettle {
			products = saleBill.GetSaleOrderProductCooking()
		} else { // 整单取消订单时，需要过滤掉付款减库存的商品。
			products = saleBill.GetSaleOrderProductCooking()
			// 过滤掉付款减库存的商品
			products = model.FilterPaymentDeductStockProduct(products)
		}
		//
		productList, err := s.getDecreaseStockList(ctx, products)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 查询所有出库过的商品
		{
			// 创建一个映射来存储所有出库单中的SaleOrderProductUuid
			existingSaleOrderProductUuids := make(map[uint64]bool)
			for _, form := range forms {
				for _, item := range form.WarehouseOutFormItems {
					existingSaleOrderProductUuids[item.SaleOrderProductUuid] = true
				}
			}

			// 过滤productList，只保留那些SaleOrderProductUuid存在于出库单中的项目
			filteredProductList := make([]*model.Product, 0)
			for _, product := range productList {
				if existingSaleOrderProductUuids[product.SaleOrderProductUuid] {
					filteredProductList = append(filteredProductList, product)
				}
			}

			warehouseForm = model.NewWarehouseForm(filteredProductList, saleBill.Uuid)
		}
	}

	// 3.构建出库单，将账单下单减库存的商品出库
	var warehouseOutForms []*model.WarehouseOutForm
	if opt.IsReverseSettle {
		products := saleBill.GetSaleOrderProductCookingSubStock()
		productList, err := s.getDecreaseStockList(ctx, products)
		if err != nil {
			return errors.WithMessage(err)
		}
		staffShiftLogUuid := uint64(0)
		staffShiftLog, err := GetCurrentStaffShiftLog(db, ctx.GetStaffUuid())
		if err != nil {
			logger.Logger.Error("获取当前未交班的班次列表失败", zap.Uint64("staffUuid", ctx.GetStaffUuid()), zap.Error(err))
		} else {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
		warehouseOutForms = model.NewWarehouseOutForm(productList, false, saleBill.Uuid, ctx.GetStaffUuid(), staffShiftLogUuid)
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

		// 更新销量. 如果是反结账，则减少销量
		if opt.IsReverseSettle {
			ProductBoms, ProducttPackages := GetSalesVolume(saleBill)
			for productBomUuid, saleNum := range ProductBoms {
				if err := repository.NewProductBomRepo(db).SubActualSaleNum(productBomUuid, saleNum); err != nil {
					return errors.WithMessage(err)
				}
			}
			for productPackageUuid, saleNum := range ProducttPackages {
				if err := repository.NewProductPackageRepo(db).SubActualSaleNum(productPackageUuid, saleNum); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 如果出库单明细不为空，则创建出库单
		for _, warehouseOutForm := range warehouseOutForms {
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
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 发布"库存变更"事件
	go func() {
		event.NewSystemBus().PublishChangeStockEvent(event.ChangeStockPayload{
			BasePayload: event.BasePayload{ // 库存变更
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	return nil
}

// 获取销量
func GetSalesVolume(saleBill *model.SaleBill) (map[uint64]float64, map[uint64]float64) {
	ProductBoms := make(map[uint64]float64)      // 规格商品销量 map[规格商品UUID]销量
	ProducttPackages := make(map[uint64]float64) // 套餐商品销量 map[套餐商品UUID]销量
	for _, saleOrder := range saleBill.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			// 删除商品、取消商品、未送厨商品、套餐子商品不增加销量
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsSendKitchen() || saleOrderProduct.IsPackageSubProduct() {
				continue
			}
			ProductBoms[saleOrderProduct.GetFlavorBomUuid()] = decimal.NewFromFloat(ProductBoms[saleOrderProduct.GetFlavorBomUuid()]).Add(decimal.NewFromFloat(saleOrderProduct.Num)).InexactFloat64()           // 增加实际销量
			ProducttPackages[saleOrderProduct.ProductPackageUuid] = decimal.NewFromFloat(ProducttPackages[saleOrderProduct.ProductPackageUuid]).Add(decimal.NewFromFloat(saleOrderProduct.Num)).InexactFloat64() // 增加实际销量
		}
	}
	return ProductBoms, ProducttPackages
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
			BasePayload: event.BasePayload{ // 挂单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	// 获取新的数据
	info, err := s.GetOrderCartInfoByDeviceSn(ctx, ctx.GetDeviceSn())
	if err != nil {
		fmt.Println("获取订单信息失败", err)
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
	currentSaleBillUuid, err := repository.NewOrderRepo(db).HasShowOrder(ctx.GetDeviceUuid())
	if err != nil {
		ctx.Log().Error("判断是否有未挂单的点餐账单失败", zap.Error(err))
		return nil, errors.WithMessage(err, "判断是否有未挂单的点餐账单失败")
	}
	if currentSaleBillUuid != 0 {
		// 如果未挂单的点餐账单没有商品，则删除该订单，并允许取单
		// 当前销售账单数据
		currentSaleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(currentSaleBillUuid)
		if errSaleBill != nil {
			return nil, errSaleBill
		}
		if len(currentSaleBill.GetSaleOrderProductAll()) == 0 {
			// 软删除sale_bill和sale_order
			repository.NewSaleBillRepo(db).DeleteSaleBill(currentSaleBill.Uuid)
			for _, saleOrder := range currentSaleBill.SaleOrders {
				repository.NewSaleOrderRepo(db).DeleteSaleOrder(saleOrder.Uuid)
			}
		} else {
			return nil, errors.New("该设备有未挂单的点餐账单，禁止取单")
		}
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
			BasePayload: event.BasePayload{ // 取单
				Ctx:          ctx,
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
	saleBills, total, err := saleBillRepo.GetHideSaleBillList(req.PageNo, req.PageSize, ctx.GetDeviceUuid())
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
				if saleOrderProduct.IsPackageSubProduct() {
					// 套餐子商品不显示
					continue
				}
				if saleOrderProduct.IsDelete() || saleOrderProduct.Num == 0 {
					continue
				}
				if product, ok := listMap[saleOrderProduct.Sign]; !ok {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(saleOrderProduct.GetNumDecimal()).InexactFloat64()
					newProduct := resp.Product{
						LocaleName:    saleOrderProduct.MultiLanguageName.GetNames(),
						Num:           saleOrderProduct.Num,
						SalePrice:     productPrice,
						DiscountPrice: productPrice,
					}
					// 如果是套餐商品，则设置套餐商品列表
					if saleOrderProduct.IsPackageProduct() {
						subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
						packageProductList := make([]resp.PackageProduct, 0)
						for _, subProduct := range subProducts {
							packageProductList = append(packageProductList, resp.PackageProduct{
								Uuid:       subProduct.Uuid,
								LocaleName: subProduct.MultiLanguageName.GetNames(),
								Num:        subProduct.Num,
								UnitNum:    subProduct.UnitNum,
							})
						}
						newProduct.PackageProductList = resp.PackageProductList{
							List: packageProductList,
						}
					}
					listMap[saleOrderProduct.Sign] = newProduct
				} else {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(saleOrderProduct.GetNumDecimal())
					price := productPrice.Add(decimal.NewFromFloat(product.SalePrice)).InexactFloat64()
					product.Num += saleOrderProduct.Num
					product.SalePrice = price
					product.DiscountPrice = price
					// 如果是套餐商品，则更新套餐商品列表
					if saleOrderProduct.IsPackageProduct() {
						for index, _ := range product.PackageProductList.List {
							unitNum := decimal.NewFromFloat(saleOrderProduct.UnitNum)                       // 每份套餐的子商品数量
							num := decimal.NewFromFloat(product.Num).Mul(unitNum).Round(3).InexactFloat64() // 套餐数量*每份套餐的子商品数量= 子商品的数量
							product.PackageProductList.List[index].Num = num
						}
					}
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

// 删除或拒单h5订单商品
func (s *orderSrv) deleteOrRejectH5OrderProduct(ctx context.Context, db *gorm.DB, saleOrderProduct *model.SaleOrderProduct) error {
	if saleOrderProduct.H5OrderUuid != 0 {
		// 如果这个商品是h5订单的最后一个商品，则删除拒单该h5订单
		// 查询h5订单是否只有一个商品，如果只有一个商品就拒单该h5订单
		h5OrderProductCount, err := repository.NewH5OrderRepo(db).GetH5OrderProductCount(saleOrderProduct.H5OrderUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if h5OrderProductCount == 1 {
			s.RejectH5Order(ctx, saleOrderProduct.H5OrderUuid)
		} else {
			// 删除h5订单商品
			if err := repository.NewH5OrderRepo(db).DeleteH5OrderProduct(saleOrderProduct.H5OrderUuid, saleOrderProduct.Uuid); err != nil {
				return errors.WithMessage(err)
			}
		}
	}
	return nil
}

// OrderTakeout 打包
func (s *orderSrv) OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 当不填销售账单ID时，表示要新建一个销售账单
	if req.SaleBillUuid == 0 {
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			req.SaleBillUuid = billInfo.Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			req.SaleBillUuid = order.SaleBillUuid
		}
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderTakeout, 0); err != nil {
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

	// 发布"整单打包"或"取消整单打包"事件
	go func() {
		if req.Takeout {
			s.bus.PublishWrapSaleBillEvent(event.WrapSaleBillPayload{
				BasePayload: event.BasePayload{ // 整单打包
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		} else {
			s.bus.PublishUnwrapSaleBillEvent(event.UnwrapSaleBillPayload{
				BasePayload: event.BasePayload{ // 取消整单打包
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		}
	}()

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
		if strings.Contains(err.Error(), "销售订单商品不存在") {
			// 获取新的数据
			info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			return info, nil
		}
		return nil, errors.WithMessage(err, "查询销售订单信息失败")
	}
	// 判断订单状态
	if ctx.GetSource() == constant.SourceAssistant {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, req.SaleOrderUuid, model.WithIsAssistant()); err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, req.SaleOrderUuid); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 判断订单商品状态
	if saleOrderProduct == nil {
		return nil, errors.New("找不到订单商品")
	}
	if saleOrderProduct.CancelTime == 0 && saleOrderProduct.Status == constant.OrderProductStatusSentKitchen {
		return nil, errors.New("商品已送厨，禁止删除")
	}

	saleOrderProduct.DeleteProduct()
	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.DeleteProduct()
		}
	}

	if saleOrderProduct.H5OrderUuid != 0 {
		if err := s.deleteOrRejectH5OrderProduct(ctx, db, saleOrderProduct); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

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
		// 删除套餐子商品
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(subProduct); errUpdate != nil {
					return errors.WithMessage(errUpdate)
				}
			}
		}
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
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
		return nil, nil, nil, errors.New("销售订单商品不存在")
	}
	return newSaleBill, newSaleOrder, newSaleOrderProduct, nil
}

// OrderProductChangePrice  修改订单商品价格
func (s *orderSrv) OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error) {
	if req.Price < 0 || req.Price > 100000000 {
		return nil, errors.New("请输入0-100000000间的价格")
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
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
		// 更新套餐商品的子商品的签名
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.UpdateSign()
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*subProduct); errUpdate != nil {
					return errUpdate
				}
			}
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
			BasePayload: event.BasePayload{ // 改价
				Ctx:           ctx,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 设置整单改价金额
	saleOrder.SetCustomAmount(req.Price)
	oldPrice := saleOrder.GetOriginAmount()

	// 整单改价后，整单折扣会取消，需要重新计算订单商品的金额
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"改价"事件
	go func() {
		event.NewSystemBus().PublishDiscountChangePriceSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 改价
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        oldPrice, // 旧价格为订单的原始应收金额
			NewPrice:        req.Price,
			DiscountType:    constant.DiscountOperationLogTypeChangePriceSaleOrder, // 整单改价的类型值
			SpecialDiscount: decimal.NewFromFloat(oldPrice).Sub(decimal.NewFromFloat(req.Price)).InexactFloat64(),
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 取消会员优惠券
	saleOrder.SetAllCouponCancel()
	if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
		return nil, errors.WithMessage(err, "取消销售订单优惠券失败")
	}

	// 在折扣之前计算会员打折后金额。必须在设置折扣之前获取，否则amount值已经改变了
	memberDiscountAmount := saleOrder.GetMemberDiscountAmount()
	// 设置整单折扣率
	saleOrder.SetCustomDiscount(req.GetDiscount())

	// 获取最新的设置
	newSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, saleBill.DeskUuid, false)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting)); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"整单打折"事件
	go func() {
		discountAmount := saleOrder.CustomDiscountFee
		event.NewSystemBus().PublishDiscountSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 整单打折
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        memberDiscountAmount, // 旧价格为订单的会员折扣后的金额。如果没有会员折扣，则旧价格为订单应收金额
			NewPrice:        saleOrder.GetAmount(),
			DiscountType:    constant.DiscountOperationLogTypeDiscountSaleOrder,
			RoundingRate:    req.GetPercentDiscount(),
			SpecialDiscount: discountAmount,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
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
		event.NewSystemBus().PublishDiscountZeroSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 订单抹零
				Ctx:           ctx,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 撤销订单的优惠折扣
	saleOrder.SetAllDiscountCancel()

	// 获取最新的设置
	newSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, saleBill.DeskUuid, false)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting)); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布取消优惠折扣事件
	go func() {
		event.NewSystemBus().PublishCancelSaleOrderDiscountEvent(event.CancelSaleOrderDiscountPayload{
			BasePayload: event.BasePayload{ // 取消折扣
				Ctx:           ctx,
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
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateMealNum, 0); err != nil {
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
	if err := orderRepo.ChangePopulation(req.SaleBillUuid, req.Population); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"修改桌台就餐人数"事件
	go func() {
		event.NewSystemBus().PublishChangeMealNumSaleBillEvent(event.ChangeMealNumSaleBillPayload{
			BasePayload: event.BasePayload{ // 修改桌台人数
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: req.SaleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			OldMealNum: oldMealNum,
			NewMealNum: uint(req.Population),
		})
	}()

	// 推送桌台更新
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
		"desk_uuid":   billInfo.DeskUuid,
		"update_time": time.Now().Unix(),
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// GetOrderChangeBuffet 桌台订单自助餐信息
func (s *orderSrv) GetOrderChangeBuffet(ctx context.Context, saleBillUuid, saleOrderUuid uint64) (resp.OrderBuffetResp, error) {
	res := resp.OrderBuffetResp{
		BuffetUuids:         make([]uint64, 0),
		BuffetCustomerTypes: make([]resp.DeskBuffetCustomerType, 0),
	}
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetOrderBuffetInfo(saleBillUuid, saleOrderUuid)
	if err != nil {
		return res, errors.ErrInternal
	}
	if !saleBill.IsBuffetSaleBill() {
		return res, nil
	}
	if len(saleBill.SaleOrders) == 0 {
		return res, nil
	}
	var customerTypes []resp.DeskBuffetCustomerType
	var buffetUuids []uint64
	for _, customerType := range saleBill.SaleOrders[0].SaleOrderBuffetCustomerTypes {
		customerTypes = append(customerTypes, resp.DeskBuffetCustomerType{
			Uuid:    customerType.BuffetCustomerTypePrice.BuffetCustomerType.Uuid,
			MealNum: customerType.Num,
		})
		buffetUuids = append(buffetUuids, customerType.BuffetPackageUuid)
	}
	return resp.OrderBuffetResp{
		BuffetUuids:         buffetUuids,
		BuffetCustomerTypes: customerTypes,
	}, nil
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
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateMealNum, 0); err != nil {
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
	oldBuffetIds := []uint64{}
	if saleBill.BuffetPackage1Uuid != 0 {
		oldBuffetIds = append(oldBuffetIds, saleBill.BuffetPackage1Uuid)
	}
	if saleBill.BuffetPackage2Uuid != 0 {
		oldBuffetIds = append(oldBuffetIds, saleBill.BuffetPackage2Uuid)
	}
	addBuffetIds := slice.Difference(req.BuffetUuids, oldBuffetIds)
	removeBuffetIds := slice.Difference(oldBuffetIds, req.BuffetUuids)

	// 获取自助餐顾客
	customerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&customerTypes, req.BuffetCustomerTypes)
	saleOrderCustomerTypes, buffetUuids, mealNum, maxTimeLimit, _, _ := saleOrder.GetSaleOrderBuffetCustomerTypes(
		buffetList,
		req.BuffetUuids,
		customerTypes,
		saleBill.SaleBillSetting,
	)

	// 修改
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 删除原来的 CustomerType
		repository.NewOrderRepo(tx).DeleteSaleOrderBuffetCustomerType(saleOrder.Uuid)
		saleBill.DeleteSaleOrderBuffetCustomerTypeAll(saleOrder.Uuid)

		// 创建新的顾客
		if len(saleOrderCustomerTypes) > 0 {
			for _, customer := range saleOrderCustomerTypes {
				saleBill.AddSaleOrderBuffetCustomerType(saleOrder.Uuid, customer)
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
				if slices.Contains(buffetProductUuids, product.ProductPackageUuid) {
					product.IsBuffet = 1
					if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProduct(product); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
		}

		// 保存账单。不能用这个方法来创建销售账单，故不使用UpdateOrCreate
		saleBill.SetMealNum(mealNum)
		saleBill.SetBuffetPackage(buffetUuids)
		saleBill.SetDelayProductMealNum(mealNum)

		//
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 重新计算销售账单.
	{
		newSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "repository.NewOrderRepo(db).GetSaleBillAllInfo failed")
		}
		if err := s.CalcAndSaveSaleBill(ctx, db, newSaleBill); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

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
		return nil, errors.New("当前不是自助餐类订单，无法添加加钟")
	}
	if saleBill.GetBuffetRemainingSeconds() == -1 {
		return nil, errors.New("当前套餐已经是无限时，无法添加加钟")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderClock, req.SaleOrderUuid); err != nil {
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
func (s *orderSrv) OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if ctx.GetSource() == constant.SourceAssistant && billInfo.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderProductRemark, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单
	saleOrder := billInfo.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 修改订单商品备注
	var saleOrderProduct *model.SaleOrderProduct
	for _, product := range saleOrder.SaleOrderProducts {
		if product.Uuid == req.OrderProductUuid {
			saleOrderProduct = product
			break
		}
	}
	if saleOrderProduct == nil {
		return nil, errors.New("订单商品不存在")
	}

	// 更新订单商品备注
	repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		saleOrderProduct.Remark = req.Remark
		saleOrderProduct.UpdateSign()
		sign := saleOrderProduct.Sign
		if err := repository.NewOrderRepo(db).ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Remark, sign); err != nil {
			return errors.WithMessage(err)
		}
		// 更新套餐商品的子商品的签名
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.UpdateSign()
				if err := repository.NewOrderRepo(db).ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, subProduct.Uuid, req.Remark, subProduct.Sign); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		return nil
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderRemark 修改订单备注
func (s *orderSrv) OrderRemark(ctx context.Context, req req.OrderRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	if req.SaleBillUuid == 0 {
		return nil, errors.New("请先下单再整单备注")
	}
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderOrderRemark, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 查询整单备注信息
	orderRemarkList, err := base.NewOrderRemarkRepo(db).GetOrderRemarkListByUuids(req.RemarkUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "查询整单备注信息失败")
	}
	if len(orderRemarkList) != len(req.RemarkUuids) {
		return nil, errors.WithMessage(errors.New("整单备注信息不存在"), "整单备注信息不存在")
	}

	// 整单备注信息
	orderRemark, err := billInfo.GetOrderRemark()
	if err != nil {
		return nil, errors.WithMessage(err, "获取整单备注信息失败")
	}

	if req.IsNullRemark() && orderRemark != nil { // 历史备注信息存在，但是没有选择整单备注或输入整单备注文本
		// 划线所有历史备注
		for i := range orderRemark.List {
			orderRemark.List[i].IsLatest = false
		}
		billInfo.OrderRemark = orderRemark.ToJson()
	} else {
		// 创建新的备注信息
		orderRemarkItem := resp.OrderRemarkItem{
			IsLatest: true,
			Uuids:    req.RemarkUuids,
			Remark:   req.Remark,
			Remarks: func() []dto.LocaleResponse {
				remarks := make([]dto.LocaleResponse, 0)
				for _, remark := range orderRemarkList {
					remarks = append(remarks, remark.MultiLanguageName.GetNames())
				}
				return remarks
			}(),
			CreateTime: time.Now().Unix(),
		}
		if orderRemark != nil {
			// 有历史备注信息
			// 修改历史备注信息为不是最新
			for i := range orderRemark.List {
				orderRemark.List[i].IsLatest = false
			}
			orderRemark.List = append(orderRemark.List, orderRemarkItem)
			billInfo.OrderRemark = orderRemark.ToJson()
		} else {
			// 没有历史备注信息
			orderRemarkInfo := &resp.OrderRemarkInfo{
				List: []resp.OrderRemarkItem{orderRemarkItem},
			}
			billInfo.OrderRemark = orderRemarkInfo.ToJson()
		}
	}
	// 修改订单备注
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := repository.NewOrderRepo(db).UpdateSaleBillOrderRemark(req.SaleBillUuid, billInfo.OrderRemark); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// 通过设备SN获取销售账单uuid
func (s *orderSrv) getSaleBillUuidByDeviceSn(ctx context.Context) (uint64, error) {
	var saleBillUuid uint64
	// 通过设备sn查询设备ID
	db := s.dbm.GetDB(ctx.GetDbId())
	// 通过设备ID查询未挂单的销售订单
	if saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(ctx.GetDeviceUuid()); err != nil {
		if utils.IsNotFoundRecord(err) {
			return 0, nil // 没有点餐订单
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
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx)
	if errUuid != nil {
		return nil, errors.WithMessage(errUuid)
	}
	// 没有找到销售账单
	if saleBillUuid == 0 {
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
	cartInfo, errInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errInfo != nil {
		return nil, errInfo
	}
	return cartInfo, nil
}

// GetOrderCartInfo 获取点餐购物车信息
func (s *orderSrv) GetOrderCartInfo(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 追加请求头参数，从http的header中获取h5_order_uuid
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		opts = append(opts, repository.WithH5OrderUuid(h5OrderUuid))
	}

	option := &repository.OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	db := ctx.GetDB()
	orderRepo := repository.NewOrderRepo(db)

	// 通过销售订单ID得到订单商品列表、订单金额信息、账单的销售订单列表
	shopCart, err := orderRepo.GetOrderCartInfo(saleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("saleBillUuid: %d", saleBillUuid))
	}
	if !option.FilterEndStatus && shopCart.SaleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeDeskOrderEnd, "桌台订单结束"))
	}
	// 重新计算金额
	if option.H5OrderUuid == 0 {
		shopCart.SaleBill.CalcAll()
	} else {
		// 如果从待接单进入桌台时，要把h5订单商品计算在内
		shopCart.SaleBill.CalcAll(model.WithAllAndOneH5Order(option.H5OrderUuid))
	}

	// 给订单列表添加订单
	saleOrderList := make([]resp.SaleOrder, 0)
	for _, saleOrder := range shopCart.SaleBill.SaleOrders {
		productList := make([]resp.Product, 0)
		// 给商品列表条件顾客类型
		// 如不是桌台订单、不是自助餐，这个Buffets列表是空的，故不会往productList里加入商品
		{
			customerList := saleOrder.GetCustomerList()
			productList = append(productList, customerList...)
		}

		// 添加加钟商品
		{
			delayProductList := saleOrder.GetDelayProductList()
			productList = append(productList, delayProductList...)
		}

		// 添加正常商品
		{
			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			products := saleOrder.GetProductList(option.UnorderedH5Product == repository.OrderedH5ProductWithReject, businessSetting.OpenIsBatch())
			productList = append(productList, products...)
		}

		// 商品计数
		productNum := 0.0
		for _, product := range productList {
			// 退菜的商品不计入
			if product.IsCancel {
				continue
			}
			productNum += product.Num
		}

		// 填写订单信息
		order := resp.SaleOrder{
			Uuid:                saleOrder.Uuid,
			OrderNo:             saleOrder.OrderNo,
			Status:              saleOrder.Status,
			ProductNum:          decimal.NewFromFloat(productNum).Truncate(3).InexactFloat64(),
			ProductList:         productList,
			IsDiscount:          saleOrder.IsManualDiscount(uint8(shopCart.SaleBill.SaleBillSetting.ZeroRule)),
			IsMemberDiscount:    saleOrder.IsMemberDiscount(),
			CustomDiscountRate:  saleOrder.CustomDiscountRate,
			ZeroRule:            saleOrder.ZeroRule,
			AutoDiscountMessage: saleOrder.GetAutoDiscountMessage(*shopCart.SaleBill.SaleBillSetting, ctx.GetLanguage()),
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

	takeout := shopCart.SaleBill.IsTakeout()
	orderRemark, _ := shopCart.SaleBill.GetOrderRemarkRes()
	shopCartInfo := &resp.ShopCart{
		SaleBillUuid:  saleBillUuid,
		IsDeskOrder:   shopCart.IsDeskShopCart(),
		IsLock:        shopCart.SaleBill.IsLockStatus(),
		Takeout:       &takeout,
		Desk:          nil,
		Buffet:        nil,
		DiningMethod:  shopCart.SaleBill.DiningMethod,
		SaleOrderList: saleOrderList,
		UpdateTime:    shopCart.SaleBill.UpdateTime,
		OrderRemark:   orderRemark,
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
				RemainingSeconds:      shopCart.SaleBill.GetTotalRemainingSeconds(),
				IsTimeLimited:         shopCart.SaleBill.IsTimeLimited(),
				LocaleName:            shopCart.SaleBill.GetBuffetName(),
				RemainingOrderingTime: shopCart.SaleBill.GetRemainingOrderingSeconds(),
				ReminderOrderTime:     int64(shopCart.SaleBill.ReminderOrderTime) * 60,
				IsTabletH5TimeSet:     shopCart.SaleBill.IsBuffetTabletH5TimeSet(),
			}
		}
	}

	// 获取必点方案
	if !option.NoQueryMustPlan {
		saleOrder := shopCart.SaleBill.GetFirstSaleOrder()
		var productMustPlanList *resp.ProductMustPlanList
		// 如果销售账单需要显示必点方案
		if shopCart.SaleBill.IsShowMustPlan() {
			// 获取必点方案列表
			var mustPlan *resp.InstantProductMustPlanResp
			var isAutoAdd bool
			if shopCart.SaleBill.IsDeskSaleBill() {
				mustPlan, isAutoAdd, err = s.DeskOrderMustPlan(ctx, saleBillUuid, saleOrder.Uuid, shopCart.SaleBill.MealNum, option.H5AutoAdd, option.NoAutoAdd)
				if err != nil {
					ctx.Log().Info("获取桌台必点方案列表失败", zap.Error(errors.WithMessage(err)))
					if !shopCart.SaleBill.IsEndStatus() {
						return nil, errors.WithMessage(errors.New("获取桌台必点方案列表失败"), err.Error())
					}
				}
			} else {
				mustPlan, isAutoAdd, err = s.InstantOrderMustPlan(ctx, ctx.GetDeviceSn())
				if err != nil {
					ctx.Log().Info("获取点餐必点方案列表失败", zap.Error(errors.WithMessage(err)))
					if !shopCart.SaleBill.IsEndStatus() {
						return nil, errors.WithMessage(errors.New("获取点餐必点方案列表失败"), err.Error())
					}
				}
			}

			if mustPlan != nil {
				// 如果已经自动加购完成，则不在显示必点方案
				finish := true
				for _, mustPlan := range mustPlan.List {
					// 如果必点方案中还有需要加购的商品，则显示必点方案
					if mustPlan.NeedNum != 0 {
						finish = false
					}
				}
				if isAutoAdd {
					opts = append(opts, repository.WithCanCloseMustPlanView()) // 设置可以关闭必点弹窗
					return s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
				} else {
					productMustPlanList = &resp.ProductMustPlanList{
						List: mustPlan.List,
					}

					// 如果在可以关闭必点弹窗的场景下，且已经自动加购完成
					if option.CanCloseMustPlanView && finish {
						// 如果已经自动加购完成，则不在显示必点方案.并更新sale_bill为已完成必点
						err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(saleBillUuid)
						if err != nil {
							return nil, errors.WithMessage(err)
						}
						// 清空必点方案, 不给前端返回必点数据
						productMustPlanList = nil

					}
				}
			}
		}
		// 如果要显示必点信息
		if productMustPlanList != nil {
			shopCartInfo.MustPlans = productMustPlanList
		}

		// 判断是否需要弹出分批送厨弹窗。只有收银机和助手端需要判断
		if ctx.GetSource() == constant.SourceAssistant || ctx.GetSource() == constant.SourceCashier {
			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			if businessSetting.OpenIsBatch() {
				if shopCart.SaleBill.IsNeedBatchSendCooking() {
					shopCartInfo.SetCode(constant.CodeOrderCheckProductBatch)
				}
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

// 获取商品详情
func (s *orderSrv) GetProductDetail(ctx context.Context, productPackageUuid uint64) (product_resp.Product, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	productRepo := repository.NewProductRepo(db)
	commonRepo := repository.NewCommonRepo()
	products, _, err := productRepo.GetProductListWithPagination(
		1,
		10,
		commonRepo.WhereByUuid(productPackageUuid),
	)

	if err != nil {
		return product_resp.Product{}, errors.WithMessage(err)
	}
	if len(products) == 0 {
		return product_resp.Product{}, errors.WithMessage(errors.New("商品不存在"), "商品不存在")
	}

	formatProducts := FormatProducts(ctx, products)
	return formatProducts[0], nil
}

// OrderCartProductPackageAdd 往购物车添加套餐
func (s *orderSrv) OrderCartProductPackageAdd(ctx context.Context, request req.OrderCartProductPackageAddReq) (*resp.ShopCart, error) {

	// 当不填销售账单ID时，表示要新建一个销售账单
	if request.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			request.SaleBillUuid = billInfo.Uuid
			request.SaleOrderUuid = billInfo.SaleOrders[0].Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			request.SaleBillUuid = order.SaleBillUuid
			request.SaleOrderUuid = order.SaleOrderUuid
		}
	}

	// 上锁防止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 往销售账单里添加商品
	productParam := req.ProductParams{
		FlavorProductBomUuid: request.ProductPackageUuid,
		Num:                  1,
		Operation:            "add",
	}
	// 记录相关的子商品。
	subProducts := make([]req.ProductParams, 0)
	for _, productReq := range request.Products {
		subProduct := req.ProductParams{
			FlavorProductBomUuid:            productReq.FlavorUuid,
			Num:                             productReq.Num,
			ProductPackageAttributeUuidList: productReq.AttributeUuidList,
			ProductPackageGroupUuid:         productReq.ProductPackageGroupUuid,
			Operation:                       "add",
		}
		subProducts = append(subProducts, subProduct)
	}
	productParam.SetIsPackageProduct(subProducts) // 设置为套餐商品

	shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products: []req.ProductParams{
			productParam,
		},
		IsH5Product: request.IsH5Product(),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return shopCart, nil
}

func (s *orderSrv) OrderCartProductFlavorAndAttribute(ctx context.Context, request req.OrderCartProductFlavorAndAttributeReq) (*resp.ProductFlavorAndAttributeRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("销售订单不存在"), "销售订单不存在")
	}

	saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
	if errProduct != nil {
		return nil, errors.WithMessage(errProduct)
	}

	if saleOrderProduct.IsPackageProduct() {
		product, errProduct := s.GetProductDetail(ctx, saleOrderProduct.ProductPackageUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		productDetail := saleOrderProduct.GetPackageDetail()
		return &resp.ProductFlavorAndAttributeRes{
			ProductType: uint(saleOrderProduct.ProductType),
			ProductInfo: &resp.ProductInfo{
				Product: product,
				SelectedProductPackage: &resp.ProductSelectedInfoList{
					List: productDetail,
				},
			},
		}, nil
	} else {
		product, errProduct := s.GetProductDetail(ctx, saleOrderProduct.ProductPackageUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		productDetail := saleOrderProduct.GetProductPackageDetail()
		return &resp.ProductFlavorAndAttributeRes{
			ProductType: uint(saleOrderProduct.ProductType),
			ProductInfo: &resp.ProductInfo{
				Product:         product,
				SelectedProduct: &productDetail,
			},
		}, nil
	}
}

func (s *orderSrv) OrderCartProductFlavorAndAttributeChange(ctx context.Context, request req.OrderCartProductFlavorAndAttributeChangeReq) (*resp.ShopCart, error) {

	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("销售订单不存在"), "销售订单不存在")
	}

	// 如果是修改普通商品
	if request.ProductType == 0 {
		// 普通商品
		saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		if !saleOrderProduct.IsCanEdit() {
			return nil, errors.WithMessage(errors.New("商品不可编辑"), "商品不可编辑")
		}

		saleOrderProduct, err := EditProduct(ctx, db, saleOrder, saleOrderProduct, request.EditProductReq)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		// 套餐商品
		saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		if !saleOrderProduct.IsCanEdit() {
			return nil, errors.WithMessage(errors.New("商品不可编辑"), "商品不可编辑")
		}
		// 获取套餐子商品
		subProducts := saleOrder.GetPackageSubProductList(request.SaleOrderProductUuid)
		subProductParamMap := make(map[string]req.ProductRequest) // 套餐子商品参数. key为商品规格uuid+套餐分组uuid
		for i, _ := range request.Products {
			params := request.Products[i]
			subProductParamMap[fmt.Sprintf("%d-%d", params.FlavorUuid, params.ProductPackageGroupUuid)] = params
		}
		if len(subProducts) != len(subProductParamMap) {
			return nil, errors.WithMessage(errors.New("修改前后套餐子商品数量不一致"), "修改前后套餐子商品数量不一致")
		}
		for _, subProduct := range subProducts {
			key := fmt.Sprintf("%d-%d", subProduct.GetFlarvorSaleOrderProductBom().ProductBomUuid, subProduct.PackageGroupUuid)
			params, ok := subProductParamMap[key]
			if !ok {
				return nil, errors.WithMessage(errors.New("套餐子商品不存在"), "套餐子商品不存在")
			}
			_, err := EditProduct(ctx, db, saleOrder, subProduct, params.EditProductReq)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		}
		// 重新记录套餐商品的子商品参数
		saleOrderProduct.PackageSubProductParams = func() string {
			subProductList := make([]req.SubProduct, 0)
			for i := range subProducts {
				params := request.Products[i]
				subProductList = append(subProductList, req.SubProduct{
					FlavorUuid:              params.FlavorUuid,
					AttributeUuid:           params.AttributeUuidList,
					ProductPackageGroupUuid: params.ProductPackageGroupUuid,
				})
			}
			return utils.ToJson(subProductList)
		}()
		// 重新计算套餐商品的签名
		saleOrderProduct.UpdateSign()
	}

	// 从新计算订单并保存
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

func EditProduct(ctx context.Context, db *gorm.DB, saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct, request req.EditProductReq) (*model.SaleOrderProduct, error) {
	// 删除所有规格、加料和属性
	saleOrderProduct.DeleteAllSaleOrderProductBomsAndAttributes()
	// 添加新的规格、加料和属性
	{
		// 添加新规格
		flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(request.FlavorUuid)
		if errFlavorProductBom != nil {
			return nil, errors.WithMessage(errFlavorProductBom)
		}
		flavor := model.NewSaleOrderProductFlavor(saleOrderProduct.Uuid, saleOrder.Uuid, model.Flavor{
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			Price:          flavorProductBom.Price,
			ProductBomUuid: request.FlavorUuid,
			ErpCode:        flavorProductBom.ErpCode,
		})
		flavor.SetUpdate()
		saleOrderProduct.SaleOrderProductBoms = append(saleOrderProduct.SaleOrderProductBoms, flavor)
		saleOrderProduct.ChangeFlavor(flavor)

		// 添加新加料
		sauceProductBoms, errSauceProductBoms := GetSauceInfo(ctx, db, request.SauceUuidList, saleOrderProduct.Num)
		if errSauceProductBoms != nil {
			return nil, errors.WithMessage(errSauceProductBoms)
		}
		sauces := make([]model.Sauce, 0)
		for sauceProductBomUuid, sauceProductBom := range sauceProductBoms {
			sauce := model.Sauce{
				Name:           sauceProductBom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
				Price:          sauceProductBom.Price,
				ProductBomUuid: sauceProductBomUuid,
			}
			sauces = append(sauces, sauce)
		}
		for _, sauce := range sauces {
			sauce := model.NewSaleOrderProductSauce(saleOrderProduct.Uuid, saleOrder.Uuid, sauce)
			sauce.SetUpdate()
			saleOrderProduct.SaleOrderProductBoms = append(saleOrderProduct.SaleOrderProductBoms, sauce)
		}
		// 添加新属性
		productAttributes, errProductAttributes := GetAttributeInfo(ctx, db, request.AttributeUuidList)
		if errProductAttributes != nil {
			return nil, errors.WithMessage(errProductAttributes)
		}
		attributes := sortProductAttributes(ctx, productAttributes)
		for _, attribute := range attributes {
			attr := model.NewSaleOrderProductAttribute(saleOrderProduct.Uuid, saleOrder.Uuid, attribute)
			attr.SetUpdate()
			saleOrderProduct.SaleOrderProductAttributes = append(saleOrderProduct.SaleOrderProductAttributes, attr)
		}
	}

	// 重新计算商品的签名
	saleOrderProduct.UpdateSign()
	return saleOrderProduct, nil
}

// 获取商品规格信息。用于加购商品时作为订单商品数据的数据来源
func GetAttributeInfo(ctx context.Context, db *gorm.DB, productPackageAttributeUuidList []uint64) (map[uint64]*model.ProductPackageAttribute, error) {
	// 获取属性信息
	productAttributes := make(map[uint64]*model.ProductPackageAttribute)
	if len(productPackageAttributeUuidList) > 0 {
		productAttributeList, errProductAttributeList := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(productPackageAttributeUuidList)
		if errProductAttributeList != nil {
			return nil, errors.WithMessage(errProductAttributeList)
		}
		for i, attribute := range productAttributeList {
			productAttributes[attribute.Uuid] = productAttributeList[i]
		}
	}
	return productAttributes, nil
}

// 获取加料信息。用于加购商品时作为订单商品数据的数据来源
func GetSauceInfo(ctx context.Context, db *gorm.DB, sauceProductBomUuidList []uint64, productNum float64) (map[uint64]*model.ProductBom, error) {
	// 获取加料信息
	sauceProductBoms := make(map[uint64]*model.ProductBom)
	if len(sauceProductBomUuidList) > 0 {
		sauceProductBomList, errSauceProductBomList := repository.NewProductBomRepo(db).GetSauceProductBomsByUuids(sauceProductBomUuidList)
		if errSauceProductBomList != nil {
			return nil, errors.WithMessage(errSauceProductBomList)
		}
		if len(sauceProductBomList) != len(sauceProductBomUuidList) {
			sauceUuidMap := make(map[uint64]struct{})
			for _, uuid := range sauceProductBomUuidList {
				sauceUuidMap[uuid] = struct{}{}
			}
			for _, bom := range sauceProductBomList {
				delete(sauceUuidMap, bom.Uuid)
			}
			names := make([]string, 0)
			for uuid := range sauceUuidMap {
				bom, err := repository.NewProductBomRepo(db).GetSauceProductBomByUuid(uuid)
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				names = append(names, sauceName)
			}
			tipStrPrefix := i18n.Translate(ctx.GetLanguage(), "加料")
			tipStr := i18n.Translate(ctx.GetLanguage(), "已下架，请重新选择其他加料")
			return nil, errors.New(tipStrPrefix + " " + strings.Join(names, ",") + " " + tipStr)
		}
		for i, bom := range sauceProductBomList {
			sauceProductBoms[bom.Uuid] = sauceProductBomList[i]
			sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
			if bom.GetStockNum() < productNum {
				return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
			}
			// 检查加料材料库存是否充足
			if len(bom.ProductSauce.SauceMaterials) > 0 {
				for _, sauceMaterial := range bom.ProductSauce.SauceMaterials {
					materialStockNum := sauceMaterial.Material.GetStockNum()
					if materialStockNum < sauceMaterial.GetDecreaseNum(productNum) {
						return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "加料材料库存不足")))
					}
				}
			}
		}
	}
	return sauceProductBoms, nil
}

// InstantOrderCartProductAdd 点餐页面，往购物车添加商品。
func (s *orderSrv) InstantOrderCartProductAdd(ctx context.Context, request req.OrderCartProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 当不填销售账单ID时，表示要新建一个销售账单
	if request.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			request.SaleBillUuid = billInfo.Uuid
			request.SaleOrderUuid = billInfo.SaleOrders[0].Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			request.SaleBillUuid = order.SaleBillUuid
			request.SaleOrderUuid = order.SaleOrderUuid
		}
	}

	// 判断商品价格是否与后台设置的最新价格不一致
	// 查询商品规格的最新价格
	// 查询所选加料的最新价格
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	if request.Price != nil {
		productInfo, err := s.getInfo(ctx, req.ProductParams{
			FlavorProductBomUuid:    request.FlavorUuid,
			Price:                   request.Price,
			IsBuffet:                request.IsBuffet,
			SauceProductBomUuidList: request.SauceUuidList,
		}, db)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if productInfo != nil {
			return &resp.ShopCart{
				Product: productInfo,
			}, errors.ErrProductPriceChanged
		}
	}

	num := 1.0
	if request.Num != nil {
		num = *request.Num
	}

	// 往销售账单里添加商品
	shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products: []req.ProductParams{
			{
				FlavorProductBomUuid:            request.FlavorUuid,
				Num:                             num,
				SauceProductBomUuidList:         request.SauceUuidList,
				ProductPackageAttributeUuidList: request.AttributeUuidList,
				Operation:                       request.Operation,
				MustPlanUuid:                    request.MustPlanUuid,
			},
		},
		IsH5Product: request.IsH5Product(),
	}, opts...)

	// 往销售账单里添加商品
	// shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
	// 	SaleBillUuid:  request.SaleBillUuid,
	// 	SaleOrderUuid: request.SaleOrderUuid,
	// 	Products: []req.ProductParams{
	// 		{
	// 			FlavorProductBomUuid:            request.FlavorUuid,
	// 			Num:                             num,
	// 			SauceProductBomUuidList:         request.SauceUuidList,
	// 			ProductPackageAttributeUuidList: request.AttributeUuidList,
	// 			Operation:                       request.Operation,
	// 			MustPlanUuid:                    request.MustPlanUuid,
	// 		},
	// 	},
	// 	IsH5Product: request.IsH5Product(),
	// }, opts...)
	if err != nil {
		ctx.Log().Info("往点餐账单里添加商品失败", zap.Any("req", request), zap.Any("error", err))
		return nil, errors.WithMessage(err)
	}
	return shopCart, nil
}

type CreateSaleOrderProductParams struct {
	IsH5Product bool                  // 是否是H5商品
	Products    []req.ProductParams   // 要加购的商品列表
	Setting     model.SaleBillSetting // 销售账单设置
	SaleBill    *model.SaleBill       // 销售账单
	SaleOrder   *model.SaleOrder      // 销售订单
}

type InnerParams struct {
	IsDeskSaleBill         bool    // 是否是桌台销售账单
	SaleBillUuid           uint64  // 销售账单uuid
	SaleOrderUuid          uint64  // 销售订单uuid
	DeskUuid               uint64  // 桌台uuid
	DiningMethod           uint    // 就餐方式
	MemberDiscountRate     float64 // 会员折扣率
	MemberCardDiscountRate float64 // 会员卡折扣率
	CustomDiscountRate     float64 // 自定义折扣率
}

func (s *orderSrv) newSaleOrderProduct(ctx context.Context, params CreateSaleOrderProductParams, options ...func(option *ActionAddOption)) ([]*model.SaleOrderProduct, error) {
	option := &ActionAddOption{}
	for _, opt := range options {
		opt(option)
	}

	innerParams := InnerParams{
		IsDeskSaleBill:         params.SaleBill.IsDeskSaleBill(),
		SaleBillUuid:           params.SaleBill.Uuid,
		SaleOrderUuid:          params.SaleOrder.Uuid,
		DeskUuid:               params.SaleBill.DeskUuid,
		DiningMethod:           params.SaleBill.DiningMethod,
		MemberDiscountRate:     params.SaleOrder.MemberDiscountRate,
		MemberCardDiscountRate: params.SaleOrder.MemberCardDiscountRate,
		CustomDiscountRate:     params.SaleOrder.CustomDiscountRate,
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, product := range params.Products {
		// 获取商品包信息
		productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.FlavorProductBomUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		productPackage := productBom.ProductPackage
		productName := productPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())

		if productBom.IsDelete() {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品规格已经删除")))
		}
		// 商品已经下架
		if productBom.IsProductPackageDown() {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品已经下架")))
		}

		// 获取某商品规格信息
		flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(product.FlavorProductBomUuid)
		if errFlavorProductBom != nil {
			return nil, errors.WithMessage(errFlavorProductBom)
		}
		if flavorProductBom.GetStockNum() < float64(product.Num) {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
		}
		// 如果商品规格关联了材料，检查材料库存是否充足
		if len(flavorProductBom.FlavorMaterials) > 0 {
			for _, flavorMaterial := range flavorProductBom.FlavorMaterials {
				if flavorMaterial.IsDelete() {
					continue
				}
				materialStockNum := flavorMaterial.Material.GetStockNum()
				if materialStockNum < flavorMaterial.GetDecreaseNum(product.Num) {
					return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "材料库存不足")))
				}
			}
		}

		// 获取加料信息
		sauceProductBoms := make(map[uint64]*model.ProductBom)
		if len(product.SauceProductBomUuidList) > 0 {
			sauceProductBomList, errSauceProductBomList := repository.NewProductBomRepo(db).GetSauceProductBomsByUuids(product.SauceProductBomUuidList)
			if errSauceProductBomList != nil {
				return nil, errors.WithMessage(errSauceProductBomList)
			}
			if len(sauceProductBomList) != len(product.SauceProductBomUuidList) {
				sauceUuidMap := make(map[uint64]struct{})
				for _, uuid := range product.SauceProductBomUuidList {
					sauceUuidMap[uuid] = struct{}{}
				}
				for _, bom := range sauceProductBomList {
					delete(sauceUuidMap, bom.Uuid)
				}
				names := make([]string, 0)
				for uuid := range sauceUuidMap {
					bom, err := repository.NewProductBomRepo(db).GetSauceProductBomByUuid(uuid)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					names = append(names, sauceName)
				}
				tipStrPrefix := i18n.Translate(ctx.GetLanguage(), "加料")
				tipStr := i18n.Translate(ctx.GetLanguage(), "已下架，请重新选择其他加料")
				return nil, errors.New(tipStrPrefix + " " + strings.Join(names, ",") + " " + tipStr)
			}
			for i, bom := range sauceProductBomList {
				sauceProductBoms[bom.Uuid] = sauceProductBomList[i]
				sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				if bom.GetStockNum() < product.Num {
					return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
				}
				// 检查加料材料库存是否充足
				if len(bom.ProductSauce.SauceMaterials) > 0 {
					for _, sauceMaterial := range bom.ProductSauce.SauceMaterials {
						materialStockNum := sauceMaterial.Material.GetStockNum()
						if materialStockNum < sauceMaterial.GetDecreaseNum(product.Num) {
							return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "加料材料库存不足")))
						}
					}
				}
			}
		}

		// 获取属性信息
		productAttributes := make(map[uint64]*model.ProductPackageAttribute)
		if len(product.ProductPackageAttributeUuidList) > 0 {
			productAttributeList, errProductAttributeList := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(product.ProductPackageAttributeUuidList)
			if errProductAttributeList != nil {
				return nil, errors.WithMessage(errProductAttributeList)
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
		attributes := sortProductAttributes(ctx, productAttributes)

		isAcceptOrder := constant.OrderProductIsAcceptOrderAccepted // 已接单
		if params.IsH5Product {
			isAcceptOrder = constant.OrderProductIsAcceptOrderUnAccept // 未接单
		}
		deviceSn := ctx.GetDeviceSn()
		if ctx.GetSource() == jwt.SourceH5 {
			deviceSn = jwt.SourceH5 // 扫码h5订单，设备sn为h5
		}

		flavorPrice := flavorProductBom.Price
		saleOrderProduct := model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
			DeviceId:               deviceSn,
			Name:                   productPackage.Name,
			OpenMemberDiscount:     productPackage.OpenDiscount,
			TaxRate:                productPackage.TaxRate(innerParams.DiningMethod),
			DeductStockType:        productPackage.DeductStockType,
			MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
			ImageFileUuid:          productPackage.ImageFileUuid,
			ProductPackageUuid:     productPackage.Uuid,
			SaleBillUuid:           innerParams.SaleBillUuid,
			SaleOrderUuid:          innerParams.SaleOrderUuid,
			MemberDiscountRate:     innerParams.MemberDiscountRate,
			MemberCardDiscountRate: innerParams.MemberCardDiscountRate,
			CustomDiscountRate:     innerParams.CustomDiscountRate,
			Sauces:                 sauces,
			Num:                    product.Num,
			NumType:                productPackage.NumType,
			PackageSubProductParams: func() string {
				if product.GetIsPackageProduct() {
					return utils.ToJson(product.GetSubProductList())
				}
				return ""
			}(),
			ProductType: func() uint8 {
				if product.GetIsPackageProduct() {
					return 1
				}
				return 0
			}(),
			Flavor: model.Flavor{
				Name:           flavorProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 填顾客下单时规格的名字 todo preload
				Price:          flavorPrice,
				ProductBomUuid: product.FlavorProductBomUuid,
				ErpCode:        flavorProductBom.ErpCode,
			},
			Attribute:     attributes,
			IsAcceptOrder: uint(isAcceptOrder),
			Remark:        product.Remark,
			IsBatch: func() uint8 {
				if businessSetting.OpenIsBatch() {
					if productPackage.IsBatchBool() {
						return 1
					}
				}
				return 0
			}(),
		}, &productPackage, product.Operation)

		// 设置必点信息
		if !option.IsMemberAdd { // 会员端加购不设置必点信息
			var mustPlanUuid uint64
			var isRequire bool
			mustPlanUuid, err = s.mustPlanSrv.GetMustPlanUuidByProductPackage(ctx, innerParams.SaleBillUuid, productPackage.Uuid, innerParams.DeskUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("获取到必点方案uuid", zap.Any("mustPlanUuid", mustPlanUuid))
			// 判断该必点方案是不是这个sale_bill的
			shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(innerParams.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			var mustPlanList []resp.InstantProductMustPlan
			var errMustPlanList error
			if innerParams.IsDeskSaleBill {
				mustPlanList, errMustPlanList = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), innerParams.DeskUuid)
			} else {
				mustPlanList, errMustPlanList = s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
			}
			if errMustPlanList != nil {
				return nil, errors.WithMessage(errMustPlanList)
			}
			mustPlanUuidMap := make(map[uint64]bool)
			for _, mustPlan := range mustPlanList {
				if mustPlan.Uuid == mustPlanUuid {
					isRequire = true
				}
				mustPlanUuidMap[mustPlan.Uuid] = true
			}

			if isRequire {
				if mustPlanUuidMap[product.MustPlanUuid] {
					// 如果请求中填写的必点方案uuid是该桌台的必点方案之一，则标记该商品是该方案的
					mustPlanUuid = product.MustPlanUuid
				}
				saleOrderProduct.SetMustPlanInfo(mustPlanUuid)
			}
		}
		// 如果是打包订单，则更新商品打包状态
		if params.SaleBill.IsTakeout() {
			saleOrderProduct.SetWrap()
		}
		// 生成签名
		saleOrderProduct.UpdateSign()
		ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

		// 计算商品数据。折扣、税费、服务
		saleOrderProduct.CalcSaleOrderProduct(params.Setting)

		// 判断该商品是不是自助餐商品
		if params.SaleBill.IsBuffetSaleBill() {
			// 获取自助餐商品包uuid列表
			productPackageUuidMap := params.SaleBill.GetBuffetProductMap()
			// 判断该商品是不是自助餐商品
			if _, ok := productPackageUuidMap[saleOrderProduct.ProductPackageUuid]; ok {
				saleOrderProduct.SetIsBuffet()
			}
		}

		// 暂时废弃，product.Price 一直都会是nil
		// 判断前端传入的商品价格是否与后台设置的最新价格一致，如果不一致则加购失败，并返回最新的价格
		if product.Price != nil && *product.Price != saleOrderProduct.ProductPrice {
			return nil, errors.WithMessage(errors.ErrProductPriceChanged)
		}

		// 如果是平板端加购，用于不会因为签名相同而合并商品
		if option.IsTableAdd {
			// 将新的订单商品加入到订单的商品列表中，用于计算订单金额
			params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, saleOrderProduct)
			saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
			// 商品数量不能超过999个
			if saleOrderProduct.Num > constant.ProductNumMax {
				return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品数量不能超过999个")))
			}
			// 如果该商品是套餐，则新建套餐子商品
			if saleOrderProduct.ProductType == constant.ProductTypePackage {
				subProducts, err := s.newPackageSubProducts(ctx, product.GetSubProducts(), innerParams, params, saleOrderProduct.Uuid, saleOrderProduct.DeductStockType)
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, subProducts...)
				saleOrderProducts = append(saleOrderProducts, subProducts...)
			}
		} else {
			// 查询是否存在签名相同的订单商品
			orderProduct := params.SaleOrder.GetSaleOrderProductBySign(saleOrderProduct.Sign)
			if orderProduct == nil {
				ctx.Log().Debug("不存在相同的sign", zap.Any("sign", saleOrderProduct.Sign))
			}
			// 订单中存在相同签名的商品
			//hasSameSign := false
			if orderProduct != nil {
				if saleOrderProduct.IsAddOperation() {
					// 加上新增的商品数量
					orderProduct.Num += saleOrderProduct.Num
					orderProduct.SetUpdate()
					saleOrderProducts = append(saleOrderProducts, orderProduct)
					if saleOrderProduct.Num > constant.ProductNumMax {
						return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品数量不能超过999个")))
					}
					// 检查商品是否超过限购
					status, message := orderProduct.CheckCookingProduct(ctx.GetLanguage())
					if status != constant.CodeSuccess {
						return nil, errors.WithMessage(errors.New(message))
					}
					// 如果该商品是套餐，则修改套餐子商品的数量
					if orderProduct.ProductType == constant.ProductTypePackage {
						subProducts := params.SaleOrder.GetPackageSubProductList(orderProduct.Uuid)
						for _, subProduct := range subProducts {
							uintNum := decimal.NewFromFloat(subProduct.UnitNum)                                         // 每个套餐该子商品的数量
							addNum := decimal.NewFromFloat(saleOrderProduct.Num).Mul(uintNum).Round(3).InexactFloat64() // 新增的套餐该子商品的数量
							subProduct.Num += addNum
							subProduct.SetUpdate()
						}
					}
				} else if saleOrderProduct.IsSubOperation() {
					// 减去新增的商品数量
					orderProduct.Num -= saleOrderProduct.Num
					// 如果该商品是套餐，则修改套餐子商品的数量
					if orderProduct.ProductType == constant.ProductTypePackage {
						subProducts := params.SaleOrder.GetPackageSubProductList(orderProduct.Uuid)
						for _, subProduct := range subProducts {
							uintNum := decimal.NewFromFloat(subProduct.UnitNum)                                         // 每个套餐该子商品的数量
							addNum := decimal.NewFromFloat(saleOrderProduct.Num).Mul(uintNum).Round(3).InexactFloat64() // 新增的套餐该子商品的数量
							subProduct.Num -= addNum
							if subProduct.Num <= 0 {
								subProduct.SetDelete()
							}
							subProduct.SetUpdate()
						}
					}
					if orderProduct.Num <= 0 {
						orderProduct.SetDelete()
						// 如果该商品是套餐，则删除套餐子商品
						if orderProduct.ProductType == constant.ProductTypePackage {
							subProducts := params.SaleOrder.GetPackageSubProductList(orderProduct.Uuid)
							for _, subProduct := range subProducts {
								subProduct.SetDelete()
								subProduct.SetUpdate()
							}
						}
					}
					orderProduct.SetUpdate()
					saleOrderProducts = append(saleOrderProducts, orderProduct)
				}
			} else {
				// 将新的订单商品加入到订单的商品列表中，用于计算订单金额
				params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, saleOrderProduct)
				saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				// 如果该商品是套餐，则新建套餐子商品
				if saleOrderProduct.ProductType == constant.ProductTypePackage {
					subProducts, err := s.newPackageSubProducts(ctx, product.GetSubProducts(), innerParams, params, saleOrderProduct.Uuid, saleOrderProduct.DeductStockType)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, subProducts...)
					saleOrderProducts = append(saleOrderProducts, subProducts...)
				}
				// 商品数量不能超过999个
				if saleOrderProduct.Num > constant.ProductNumMax {
					return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品数量不能超过999个")))
				}
			}
		}
	}
	return saleOrderProducts, nil
}

// 排序商品属性
func sortProductAttributes(ctx context.Context, productAttributes map[uint64]*model.ProductPackageAttribute) []model.Attribute {
	attributes := make([]model.Attribute, 0)
	// 将map转换为切片，然后按AttributeGroupUuid分组排序
	productAttributeSlice := make([]*model.ProductPackageAttribute, 0, len(productAttributes))
	for _, productAttribute := range productAttributes {
		productAttributeSlice = append(productAttributeSlice, productAttribute)
	}
	if len(productAttributeSlice) > 0 {
		// 按AttributeGroupUuid分组
		groupMap := make(map[uint64][]*model.ProductPackageAttribute)
		for _, attr := range productAttributeSlice {
			groupUuid := attr.Attribute.AttributeGroupUuid
			groupMap[groupUuid] = append(groupMap[groupUuid], attr)
		}
		// 对每个分组内的属性按Attribute.ID排序
		for groupUuid, groupAttrs := range groupMap {
			sort.Slice(groupAttrs, func(i, j int) bool {
				return groupAttrs[i].ID < groupAttrs[j].ID
			})
			groupMap[groupUuid] = groupAttrs
		}
		// 获取所有分组，并按第一个分组的Attribute.ID排序
		var groups [][]*model.ProductPackageAttribute
		for _, groupAttrs := range groupMap {
			groups = append(groups, groupAttrs)
		}
		// 按第一个分组的Attribute.ID排序分组
		sort.Slice(groups, func(i, j int) bool {
			if len(groups[i]) == 0 || len(groups[j]) == 0 {
				return false
			}
			return groups[i][0].ID < groups[j][0].ID
		})
		// 合并所有分组
		productAttributeSlice = productAttributeSlice[:0] // 清空原切片
		for _, group := range groups {
			productAttributeSlice = append(productAttributeSlice, group...)
		}
	}
	for _, productAttribute := range productAttributeSlice {
		attribute := model.Attribute{
			Name:                        productAttribute.Attribute.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
			ProductAttributeUuid:        productAttribute.Attribute.Uuid,
			ProductPackageAttributeUuid: productAttribute.Uuid,
		}
		attributes = append(attributes, attribute)
	}
	return attributes
}

// 新建套餐子商品
func (s *orderSrv) newPackageSubProducts(ctx context.Context, subProducts []req.ProductParams, innerParams InnerParams,
	params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint) ([]*model.SaleOrderProduct, error) {
	subSaleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, subProduct := range subProducts {
		subSaleOrderProduct, err := s.newSaleOrderProductForPackageSubProduct(ctx, subProduct, innerParams, params, packageUuid, deductStockType)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		subSaleOrderProducts = append(subSaleOrderProducts, subSaleOrderProduct)
	}
	return subSaleOrderProducts, nil
}

func (s *orderSrv) newSaleOrderProductForPackageSubProduct(ctx context.Context, product req.ProductParams, innerParams InnerParams, params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint) (*model.SaleOrderProduct, error) {
	db := ctx.GetDB()
	// 获取商品包信息
	productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.FlavorProductBomUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	productPackage := productBom.ProductPackage
	productName := productPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())

	if productBom.IsDelete() {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品规格已经删除")))
	}
	// 商品已经下架
	if productBom.IsProductPackageDown() {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品已经下架")))
	}

	// 获取某商品规格信息
	flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(product.FlavorProductBomUuid)
	if errFlavorProductBom != nil {
		return nil, errors.WithMessage(errFlavorProductBom)
	}
	if flavorProductBom.GetStockNum() < float64(product.Num) {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
	}
	// 如果商品规格关联了材料，检查材料库存是否充足
	if len(flavorProductBom.FlavorMaterials) > 0 {
		for _, flavorMaterial := range flavorProductBom.FlavorMaterials {
			if flavorMaterial.IsDelete() {
				continue
			}
			materialStockNum := flavorMaterial.Material.GetStockNum()
			if materialStockNum < flavorMaterial.GetDecreaseNum(product.Num) {
				return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "材料库存不足")))
			}
		}
	}

	// 获取加料信息
	sauceProductBoms, errSauceProductBoms := GetSauceInfo(ctx, db, product.SauceProductBomUuidList, product.Num)
	if errSauceProductBoms != nil {
		return nil, errors.WithMessage(errSauceProductBoms)
	}

	// 获取属性信息
	productAttributes, errProductAttributes := GetAttributeInfo(ctx, db, product.ProductPackageAttributeUuidList)
	if errProductAttributes != nil {
		return nil, errors.WithMessage(errProductAttributes)
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
	// 将map转换为切片，然后按AttributeGroupUuid分组排序
	productAttributeSlice := make([]*model.ProductPackageAttribute, 0, len(productAttributes))
	for _, productAttribute := range productAttributes {
		productAttributeSlice = append(productAttributeSlice, productAttribute)
	}
	if len(productAttributeSlice) > 0 {
		// 按AttributeGroupUuid分组
		groupMap := make(map[uint64][]*model.ProductPackageAttribute)
		for _, attr := range productAttributeSlice {
			groupUuid := attr.Attribute.AttributeGroupUuid
			groupMap[groupUuid] = append(groupMap[groupUuid], attr)
		}
		// 对每个分组内的属性按Attribute.ID排序
		for groupUuid, groupAttrs := range groupMap {
			sort.Slice(groupAttrs, func(i, j int) bool {
				return groupAttrs[i].ID < groupAttrs[j].ID
			})
			groupMap[groupUuid] = groupAttrs
		}
		// 获取所有分组，并按第一个分组的Attribute.ID排序
		var groups [][]*model.ProductPackageAttribute
		for _, groupAttrs := range groupMap {
			groups = append(groups, groupAttrs)
		}
		// 按第一个分组的Attribute.ID排序分组
		sort.Slice(groups, func(i, j int) bool {
			if len(groups[i]) == 0 || len(groups[j]) == 0 {
				return false
			}
			return groups[i][0].ID < groups[j][0].ID
		})
		// 合并所有分组
		productAttributeSlice = productAttributeSlice[:0] // 清空原切片
		for _, group := range groups {
			productAttributeSlice = append(productAttributeSlice, group...)
		}
	}
	for _, productAttribute := range productAttributeSlice {
		attribute := model.Attribute{
			Name:                        productAttribute.Attribute.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
			ProductAttributeUuid:        productAttribute.Attribute.Uuid,
			ProductPackageAttributeUuid: productAttribute.Uuid,
		}
		attributes = append(attributes, attribute)
	}

	isAcceptOrder := constant.OrderProductIsAcceptOrderAccepted // 已接单
	if params.IsH5Product {
		isAcceptOrder = constant.OrderProductIsAcceptOrderUnAccept // 未接单
	}
	deviceSn := ctx.GetDeviceSn()
	if ctx.GetSource() == jwt.SourceH5 {
		deviceSn = jwt.SourceH5 // 扫码h5订单，设备sn为h5
	}

	saleOrderProduct := model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
		DeviceId:               deviceSn,
		Name:                   productPackage.Name,
		OpenMemberDiscount:     productPackage.OpenDiscount,
		TaxRate:                productPackage.TaxRate(innerParams.DiningMethod),
		DeductStockType:        deductStockType,
		MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
		ImageFileUuid:          productPackage.ImageFileUuid,
		ProductPackageUuid:     productPackage.Uuid,
		SaleBillUuid:           innerParams.SaleBillUuid,
		SaleOrderUuid:          innerParams.SaleOrderUuid,
		MemberDiscountRate:     innerParams.MemberDiscountRate,
		MemberCardDiscountRate: innerParams.MemberCardDiscountRate,
		CustomDiscountRate:     innerParams.CustomDiscountRate,
		Sauces:                 sauces,
		Num:                    product.Num,
		NumType:                productPackage.NumType,
		PackageSubProductParams: func() string {
			if product.GetIsPackageProduct() {
				return utils.ToJson(product.GetSubProductList())
			}
			return ""
		}(),
		ProductType:      constant.ProductTypePackageSubProduct, // 套餐子商品
		PackageUuid:      packageUuid,
		PackageGroupUuid: product.ProductPackageGroupUuid,
		Flavor: model.Flavor{
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 填顾客下单时规格的名字 todo preload
			Price:          flavorProductBom.Price,
			ProductBomUuid: product.FlavorProductBomUuid,
			ErpCode:        flavorProductBom.ErpCode,
		},
		Attribute:     attributes,
		IsAcceptOrder: uint(isAcceptOrder),
		Remark:        product.Remark,
	}, &productPackage, product.Operation)

	// 生成签名
	saleOrderProduct.UpdateSign()
	ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(params.Setting)

	return saleOrderProduct, nil
}

// OrderCartProductAdd 向购物车添加商品
func (s *orderSrv) OrderCartProductAdd(ctx context.Context, request req.ProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 判断订单状态
	if ctx.GetSource() == constant.SourceAssistant {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid, model.WithIsAssistant()); err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 设置添加来源
	saleBill.SetOperateSource(ctx.GetSource())

	// 加购
	if err := s.ActionAdd(ctx, request, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderCartProductNum 修改购物车商品数量
func (s *orderSrv) OrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}
	option := &repository.OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.Log().Info("修改购物车商品数量", zap.Any("request", request))
	// 商品数量不能超过999个
	if request.Num > 999 {
		ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
		return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
	}

	// 检查商品销售库存是否充足
	ctx.Log().Debug("获取账单信息")
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	ctx.Log().Debug("获取到账单信息成功")

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateProductNum, request.SaleOrderUuid); err != nil {
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
		return nil, errors.WithMessage(errSaleOrderProduct)
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
	beforeNum := saleOrderProduct.Num
	saleOrderProduct.Num = request.Num
	ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

	// 检查商品销售库存是否充足
	if request.Num > beforeNum {
		status, message := saleOrderProduct.CheckCookingProduct(ctx.GetLanguage())
		if status != constant.CodeSuccess {
			return nil, errors.WithMessage(errors.New(message))
		}
	}
	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了商品金额", zap.Any("saleOrderProduct salePrice", saleOrderProduct.SalePrice))
	saleOrder.SaleOrderProducts[index] = saleOrderProduct

	// 计算订单金额
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	// 检查限购
	{
		// 如果是减数量，则不检查限购. 只有加数量时，才检查限购
		if request.Num > beforeNum {
			limitProducts, err := s.getBuffetProductLimitList(ctx, request.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			var overLimitProducts []*model.SaleOrderProduct
			if option.UnorderedH5Product == repository.UnorderedH5Product {
				overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithH5CheckLimit(), model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
			} else {
				overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
			}
			if len(overLimitProducts) > 0 {
				return nil, errors.WithMessage(errors.New("商品超过限购"))
			}
		}
	}

	// 商品数量不能超过999个
	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购999个
	if request.Num > beforeNum {
		if request.Num > constant.ProductNumMax {
			ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
			return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
		}
	}

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			unitNum := decimal.NewFromFloat(subProduct.UnitNum)
			subProduct.Num = decimal.NewFromFloat(request.Num).Mul(unitNum).Round(3).InexactFloat64()
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单商品成功")
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(subProduct); errUpdate != nil {
					return errors.WithMessage(errUpdate)
				}
				ctx.Log().Debug("更新销售订单套餐子商品成功")
			}
		}
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单成功")
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errors.WithMessage(errUpdateSaleBill)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "修改商品数量时，保存数据失败")
	}
	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	ctx.Log().Debug("获取新的账单数据")
	return info, nil
}

// AssistantOrderCartProductNum 助手端修改购物车商品数量
func (s *orderSrv) AssistantOrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.Log().Info("修改购物车商品数量", zap.Any("request", request))
	// 商品数量不能超过999个
	if request.Num > 999 {
		ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
		return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	ctx.Log().Debug("获取到账单信息成功")

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateProductNum, 0, model.WithIsAssistant()); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	ctx.Log().Debug("获取到订单信息成功")

	// 获取销售订单商品信息
	saleOrderProduct := saleBill.GetSaleOrderProductByUuid(request.SaleOrderProductUuid)
	if saleOrderProduct == nil {
		return nil, errors.New("销售订单商品不存在")
	}
	ctx.Log().Debug("获取到订单商品信息成功")

	// 商品的签名
	sign := saleOrderProduct.Sign
	saleOrderProductList := saleBill.GetSaleOrderProductBySign(sign)
	productNum := float64(0)
	for _, saleOrderProduct := range saleOrderProductList {
		productNum += saleOrderProduct.Num
	}
	operation := "add"
	if productNum > request.Num {
		operation = "sub"
	}

	// 排序. 从子单开始减商品
	slices.SortFunc(saleBill.SaleOrders, func(a, b *model.SaleOrder) int {
		return int(b.CreateTime) - int(a.CreateTime)
	})
	for _, saleOrder := range saleBill.SaleOrders {
		for _, product := range saleOrderProductList {
			if product.SaleOrderUuid == saleOrder.Uuid {
				saleOrderProduct = product
				break
			}
		}
	}

	if saleOrderProduct.IsCookingProduct() {
		return nil, errors.WithMessage(errors.New("商品已送厨，不能修改数量"))
	}

	// 修改销售订单商品数量
	beforeNum := saleOrderProduct.Num

	// 检查商品销售库存是否充足
	if operation == "add" {
		saleOrderProduct.Num = saleOrderProduct.Num + 1
		request.Num = saleOrderProduct.Num
		ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

		// FIXME 暂时废弃，只在送厨和结账时检查
		// status, message := saleOrderProduct.CheckCookingProduct(ctx.GetLanguage())
		// if status != constant.CodeSuccess {
		// 	return nil, errors.WithMessage(errors.New(message))
		// }
	} else if operation == "sub" {
		num := saleOrderProduct.Num - 1
		// 数量为0删除商品
		if num == 0 {
			res, err := s.OrderProductDelete(ctx, ctx.GetDbId(), ctx.GetStaffUuid(), ctx.GetSource(), req.OrderProductDeleteReq{
				SaleBillUuid:     request.SaleBillUuid,
				SaleOrderUuid:    saleOrderProduct.SaleOrderUuid,
				OrderProductUuid: saleOrderProduct.Uuid,
			})
			if err != nil {
				return nil, errors.WithMessage(err, "删除商品失败")
			}
			return res, nil
		}
		saleOrderProduct.Num = num
		request.Num = saleOrderProduct.Num
		ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))
	}
	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			unitNum := decimal.NewFromFloat(subProduct.UnitNum)
			subProduct.Num = decimal.NewFromFloat(saleOrderProduct.Num).Mul(unitNum).Round(3).InexactFloat64()
			subProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
		}
	}

	// 计算订单金额
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	// FIXME 暂时废弃，只在送厨和结账时检查
	// 检查限购
	// {
	// 	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购
	// 	if request.Num > beforeNum {
	// 		limitProducts, err := s.getBuffetProductLimitList(ctx, request.SaleBillUuid)
	// 		if err != nil {
	// 			return nil, errors.WithMessage(err)
	// 		}
	// 		overLimitProducts := saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
	// 		if len(overLimitProducts) > 0 {
	// 			return nil, errors.WithMessage(errors.New("商品超过限购"))
	// 		}
	// 	}
	// }

	// 商品数量不能超过999个
	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购999个
	if request.Num > beforeNum {
		if request.Num > constant.ProductNumMax {
			ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
			return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单商品成功")
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(subProduct); errUpdate != nil {
					return errors.WithMessage(errUpdate)
				}
			}
			ctx.Log().Debug("更新销售订单套餐子商品成功")
		}

		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单成功")
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errors.WithMessage(errUpdateSaleBill)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "修改商品数量时，保存数据失败")
	}

	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	ctx.Log().Debug("获取新的账单数据")

	return info, nil
}

// 获取自助餐商品的限购数量
func (s *orderSrv) getBuffetProductLimitList(ctx context.Context, saleBillUuid uint64) (map[uint64]uint, error) {
	buffetProductList, err := s.OrderDeskBuffetProductList(ctx, req.OrderChangeBuffetProductListReq{
		SaleBillUuid: saleBillUuid,
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	limitProducts := make(map[uint64]uint)
	for _, buffetProduct := range buffetProductList.List {
		limitProducts[buffetProduct.Uuid] = buffetProduct.Limit
	}
	return limitProducts, nil
}

func (s *orderSrv) skipCheck(saleOrderProduct *model.SaleOrderProduct) bool {
	// 如果商品是新加购的商品，则不检查。用于加购并送厨的场景
	if saleOrderProduct.NoPrimaryKey() {
		return true
	}
	// 如果商品是取消的商品，则不检查
	if saleOrderProduct.IsCancelProduct() {
		return true
	}
	return false
}

type CheckOrderOptions struct {
	CheckType                int                     // 1:送厨检查 2:结账检查
	IsH5Check                bool                    // 是否是H5检查
	SelectedMustPlanProducts *ro.MustPlanProductInfo // 桌台已经选择的必点商品。使用场景仅用于平板加购并送厨时，将新加购的商品构建为该对象
	SaleOrderProductUuids    []uint64                // h5端下单场景下，只检查本次下单的商品是否超过限购
	H5OrderUuid              uint64                  // h5端下单场景下，当h5订单接单时，当h5订单商品金额有变化时更新该h5订单商品金额
}

const (
	CheckTypeCooking  = 1 // 1:送厨检查
	CheckTypeCheckout = 2 // 2:结账检查
)

// WithCheckTypeCooking 送厨检查
func WithCheckTypeCooking() func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.CheckType = CheckTypeCooking
	}
}

// WithCheckTypeCheckout 结账检查
func WithCheckTypeCheckout() func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.CheckType = CheckTypeCheckout
	}
}

func WithIsH5Check() func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.IsH5Check = true
	}
}

func WithSelectedMustPlanProducts(selectedMustPlanProducts *ro.MustPlanProductInfo) func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.SelectedMustPlanProducts = selectedMustPlanProducts
	}
}

func WithSaleOrderProductUuid(saleOrderProductUuids ...uint64) func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.SaleOrderProductUuids = saleOrderProductUuids
	}
}

func WithH5OrderUuid(h5OrderUuid uint64) func(*CheckOrderOptions) {
	return func(options *CheckOrderOptions) {
		options.H5OrderUuid = h5OrderUuid
	}
}

type FlavorNum struct {
	SaleOrderProduct *model.SaleOrderProduct
	Num              float64 // 销售订单中该规格商品的数量
}

func (f *FlavorNum) IsStockShortage() bool {
	for _, saleOrderProductBom := range f.SaleOrderProduct.SaleOrderProductBoms {
		if saleOrderProductBom.IsDelete() {
			continue
		}
		if saleOrderProductBom.IsFlavor() {
			return saleOrderProductBom.ProductBom.IsStockShortageWithMaterial(f.Num)
		}
	}
	return false
}

// checkOrder 检查订单
func (s *orderSrv) checkOrder(ctx context.Context, ignoreMust bool, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, saleOrderProductAll []*model.SaleOrderProduct, opts ...func(*CheckOrderOptions)) (*resp.OrderCheckServiceRes, error) {
	options := &CheckOrderOptions{}
	for _, opt := range opts {
		opt(options)
	}
	ctx.SetDB(db)
	// 检查必选
	if !ignoreMust {
		var shopCartInfo *ro.ShopCartRepo
		var err error
		if options.IsH5Check {
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid, repository.WithNotDeleted())
		} else {
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid, repository.WithH5OrderUuid(context.GetH5OrderUuid(ctx)))
		}
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		var instantMustPlanList []resp.InstantProductMustPlan
		if deskUuid != 0 {
			// 如果是桌台订单
			if options.CheckType == CheckTypeCheckout {
				// 检查结账
				instantMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid, WithCheckSceneCheckout())
			} else {
				// 检查送厨
				if options.SelectedMustPlanProducts != nil {
					instantMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid, WithCheckSceneCheckout(), WithSelectedMustPlanProductsCheckOption(options.SelectedMustPlanProducts))
				} else {
					// 检查结账
					instantMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid, WithCheckSceneCooking())
				}
			}
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		} else {
			// 如果不是桌台订单
			if options.CheckType == CheckTypeCheckout {
				instantMustPlanList, err = s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo(), WithCheckSceneCheckout())
			} else {
				instantMustPlanList, err = s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo(), WithCheckSceneCooking())
			}
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
				OrderCheckRes: resp.OrderCheckRes{ProductMustPlanList: &resp.ProductMustPlanList{List: mustPlans}},
			}
			return res, nil
		}
	}

	statusMap := make(map[int][]*model.SaleOrderProduct)
	// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动
	{
		// 某个规格的商品 => 商品数量
		productBomNumMap := make(map[uint64]*FlavorNum)

		var saleOrderProductsList []*model.SaleOrderProduct
		if ctx.GetSource() == constant.SourceTablet {
			// 如果是平板端，去掉非本次加购的商品. 非本次加购的商品不进行送厨检查，如库存检查等
			saleOrderProductsList = filterOtherClientProducts(saleOrderProductAll)
		} else {
			saleOrderProductsList = saleOrderProductAll
		}
		for _, saleOrderProduct := range saleOrderProductsList {
			// 跳过检查
			if s.skipCheck(saleOrderProduct) {
				continue
			}
			// 如果是h5下单场景，只检查h5下单的商品, 跳过不是未下单的商品
			if options.IsH5Check && !saleOrderProduct.IsUnOrderH5OrderProduct() {
				continue
			}

			var status int
			var message string
			if options.CheckType == CheckTypeCheckout {
				// 如果是结账检查
				status, message = saleOrderProduct.CheckOutProduct()
				// 只检查未送厨的商品
				if saleOrderProduct.DeductStockType == constant.ProductPackageDeductStockTypePay {
					{
						flavorBomUuid := saleOrderProduct.GetFlavorBomUuid()
						if _, ok := productBomNumMap[flavorBomUuid]; !ok {
							productBomNumMap[flavorBomUuid] = &FlavorNum{
								SaleOrderProduct: saleOrderProduct,
								Num:              saleOrderProduct.Num,
							}
						} else {
							flavorNum := productBomNumMap[flavorBomUuid]
							flavorNum.Num += saleOrderProduct.Num
						}
					}
				} else {
					if !saleOrderProduct.IsCookingProduct() {
						{
							flavorBomUuid := saleOrderProduct.GetFlavorBomUuid()
							if _, ok := productBomNumMap[flavorBomUuid]; !ok {
								productBomNumMap[flavorBomUuid] = &FlavorNum{
									SaleOrderProduct: saleOrderProduct,
									Num:              saleOrderProduct.Num,
								}
							} else {
								flavorNum := productBomNumMap[flavorBomUuid]
								flavorNum.Num += saleOrderProduct.Num
							}
						}
					}
				}
			} else {
				// 如果是送厨检查
				status, message = saleOrderProduct.CheckCookingProduct(ctx.GetLanguage())
				if status == constant.CodeOrderCheckProductSauceDown {
					return nil, errors.New(message)
				}

				// 只检查未送厨的商品
				if !saleOrderProduct.IsCookingProduct() {
					// 只要是未送厨的商品，都需要检查
					{
						flavorBomUuid := saleOrderProduct.GetFlavorBomUuid()
						if _, ok := productBomNumMap[flavorBomUuid]; !ok {
							productBomNumMap[flavorBomUuid] = &FlavorNum{
								SaleOrderProduct: saleOrderProduct,
								Num:              saleOrderProduct.Num,
							}
						} else {
							flavorNum := productBomNumMap[flavorBomUuid]
							flavorNum.Num += saleOrderProduct.Num
						}
					}
				}
			}

			ctx.Log().Debug("检查商品", zap.Any("status", status), zap.Any("message", message))
			if status != constant.CodeSuccess {
				statusMap[status] = append(statusMap[status], saleOrderProduct)
				// 如果商品价格变化，更新销售订单商品的价格。都是后台更新价格而未立即更新已选购商品的价格引起的
				// 价格变化包括：
				// 1. 商品规格价格变化
				// 2. 商品小料价格变化
				if status == constant.CodeOrderCheckProductPriceChanged {
					var shopCartInfo *ro.ShopCartRepo
					var err error
					if options.IsH5Check {
						shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid, repository.WithUnorderedH5Product())
					} else {
						if options.H5OrderUuid != 0 {
							// 接单场景下，检查h5订单商品金额
							shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid, repository.WithH5OrderUuid(options.H5OrderUuid))
						} else {
							shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
						}
					}
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					saleBill := shopCartInfo.SaleBill
					s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice())
				}
			}
		}

		for _, flavorNum := range productBomNumMap {
			if flavorNum.IsStockShortage() {
				exist := false
				for _, saleOrderProduct := range statusMap[constant.CodeOrderCheckProductStockZero] {
					if saleOrderProduct.Uuid == flavorNum.SaleOrderProduct.Uuid {
						exist = true
					}
				}
				// 如果没有存在，添加
				if !exist {
					statusMap[constant.CodeOrderCheckProductStockZero] = append(statusMap[constant.CodeOrderCheckProductStockZero], flavorNum.SaleOrderProduct)
				}
			}
		}
	}

	// 检查限购
	{
		// product_package_uuid => num
		numMap := make(map[uint64]float64) // key为商品包uuid value为已购买数量
		// product_package_uuid => ProductPackage
		productPackageMap := make(map[uint64]*model.ProductPackage) // key为商品包uuid, value为ProductPackage
		// product_package_uuid => SaleOrderProduct
		saleOrderProductMap := make(map[uint64]*model.SaleOrderProduct) // key为商品包uuid value为订单商品

		// 只检查本次下单的商品是否超过限购
		productPackageUuids := make([]uint64, 0)
		// 过滤掉套餐子商品, 子商品不限购
		productList := model.FilterPackageSubProduct(saleOrderProductAll)
		for _, saleOrderProduct := range productList {
			if options.SaleOrderProductUuids != nil && slices.Contains(options.SaleOrderProductUuids, saleOrderProduct.Uuid) {
				productPackageUuids = append(productPackageUuids, saleOrderProduct.ProductPackageUuid)
			}
		}

		for _, saleOrderProduct := range productList {
			// 限购检查只检查本台的商品，并台过来的商品不记.
			// 跳过非本台的商品
			if !saleOrderProduct.IsCurrentDeskProduct() {
				continue
			}
			// 跳过退菜商品
			if saleOrderProduct.IsCancelProduct() {
				continue
			}
			productPackageUuid := saleOrderProduct.ProductPackageUuid
			// 在h5端下单场景下，只检查本次下单的商品是否超过限购
			// 在收银端送厨场景下，只检查本次送厨的商品是否超过限购
			if options.SaleOrderProductUuids != nil && len(productPackageUuids) > 0 && !slices.Contains(productPackageUuids, saleOrderProduct.ProductPackageUuid) {
				continue
			}
			productPackageMap[productPackageUuid] = saleOrderProduct.ProductPackage
			numMap[productPackageUuid] = numMap[productPackageUuid] + saleOrderProduct.Num
			saleOrderProductMap[productPackageUuid] = saleOrderProduct
		}

		limitProducts, err := s.getBuffetProductLimitList(ctx, saleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}

		for productPackageUuid, num := range numMap {
			limitNum := productPackageMap[productPackageUuid].LimitNum
			// 如果商品是自助餐商品的话，使用自助餐商品的限购规则
			if productPackageLimitNum, ok := limitProducts[productPackageUuid]; ok {
				limitNum = productPackageLimitNum
			}
			// 0表示不限购
			if limitNum == 0 {
				continue
			}
			if num > float64(limitNum) {
				statusMap[constant.CodeOrderCheckProductLimitOut] = append(statusMap[constant.CodeOrderCheckProductLimitOut], saleOrderProductMap[productPackageUuid])
			}
		}
	}

	if len(statusMap) > 0 {
		for code, saleOrderProduct := range statusMap {
			products := make([]resp.Product, 0)
			for _, product := range saleOrderProduct {
				localeName := product.GetNameAndFlavorName()
				// 如果商品包没有多语言名称，则查询数据库获取
				if product.MultiLanguageName.IsNullName() {
					uuid := product.MultiLanguageNameUuid
					names, err := repository.NewMultiLanguageNameRepositoryImpl(db).GetMultiLanguageNameByUuid(uuid)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					bom, _ := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.SaleOrderProductBoms[0].ProductBomUuid)
					localeName = product.GetNameAndFlavorNameFrom(bom, &names)
				}
				products = append(products, resp.Product{
					Uuid:          product.Uuid,
					LocaleName:    localeName,
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

			// 按商品名LocaleName进行去重。如果购物车中有两个同名LocaleName的商品，那当库存不足时会有两个重复的
			list := make([]resp.Product, 0)
			{
				listMap := make(map[string]resp.Product)
				for _, product := range products {
					key := product.LocaleName.ToJson()
					listMap[key] = product
				}
				for _, product := range listMap {
					list = append(list, product)
				}
			}

			res := &resp.OrderCheckServiceRes{
				Code:          code,
				OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: list}},
			}
			return res, nil
		}
	}

	return nil, nil
}

// 检查自助餐顾客类型价格是否变动
func (s *orderSrv) checkBuffetCustomerTypePriceChanged(ctx context.Context, saleBill *model.SaleBill) *resp.OrderCheckServiceRes {
	res := &resp.OrderCheckServiceRes{
		Code: constant.CodeOrderCheckProductPriceChanged,
		OrderCheckRes: resp.OrderCheckRes{
			Products: &resp.CartProductList{
				List: make([]resp.Product, 0),
			},
		},
	}
	// 检查自助餐顾客类型价格是否变动
	for _, saleOrder := range saleBill.SaleOrders {
		if saleOrder.IsDelete() {
			continue
		}
		for _, buffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
			if buffetCustomer.IsDelete() {
				continue
			}
			if buffetCustomer.IsBuffetCustomerTypePriceChanged() || buffetCustomer.GetOpenOverallDiscountChanged() {
				// 自助餐顾客类型价格变动
				customer := resp.Product{
					Uuid:       buffetCustomer.Uuid,
					LocaleName: buffetCustomer.BuffetPackage.MultiLanguageName.GetNames(),
					LocaleAttributeName: dto.LocaleResponse{
						ZH:   buffetCustomer.Name,
						TH:   buffetCustomer.Name,
						EN:   buffetCustomer.Name,
						ZHTW: buffetCustomer.Name,
						JA:   buffetCustomer.Name,
						KO:   buffetCustomer.Name,
						MY:   buffetCustomer.Name,
						TR:   buffetCustomer.Name,
						SV:   buffetCustomer.Name,
					},
					Num:       float64(buffetCustomer.Num),
					SalePrice: buffetCustomer.SalePrice,
				}
				res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, customer)
				return res
			}
		}
	}
	return nil
}

// 检查账单快照信息是否改变
func (s *orderSrv) checkSaleBillSettingChanged(ctx context.Context, saleBill *model.SaleBill) (*resp.OrderCheckServiceRes, *model.SaleBillSetting, error) {
	res := &resp.OrderCheckServiceRes{
		Code: constant.CodeOrderCheckProductPriceChanged,
		OrderCheckRes: resp.OrderCheckRes{
			Products: &resp.CartProductList{
				List: make([]resp.Product, 0),
			},
		},
	}

	oldSetting := saleBill.SaleBillSetting
	newSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, saleBill.DeskUuid, false)
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}
	// 检查税费类型是否改变：商品含税、商品未含税
	if oldSetting.TaxFeeType != newSetting.TaxFeeType {
		oldTaxFeeType := parseTaxFeeType(ctx.GetLanguage(), oldSetting.TaxFeeType)
		newTaxFeeType := parseTaxFeeType(ctx.GetLanguage(), newSetting.TaxFeeType)
		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   oldTaxFeeType + " -> " + newTaxFeeType,
				TH:   oldTaxFeeType + " -> " + newTaxFeeType,
				EN:   oldTaxFeeType + " -> " + newTaxFeeType,
				ZHTW: oldTaxFeeType + " -> " + newTaxFeeType,
				JA:   oldTaxFeeType + " -> " + newTaxFeeType,
				KO:   oldTaxFeeType + " -> " + newTaxFeeType,
				MY:   oldTaxFeeType + " -> " + newTaxFeeType,
				TR:   oldTaxFeeType + " -> " + newTaxFeeType,
				SV:   oldTaxFeeType + " -> " + newTaxFeeType,
			},
		})
		return res, newSetting, nil
	}
	// 检查服务费类型是否改变
	if oldSetting.ServiceFeeType != newSetting.ServiceFeeType {
		oldServiceFeeType := parseServiceFeeType(ctx.GetLanguage(), oldSetting.ServiceFeeType)
		newServiceFeeType := parseServiceFeeType(ctx.GetLanguage(), newSetting.ServiceFeeType)
		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   oldServiceFeeType + " -> " + newServiceFeeType,
				TH:   oldServiceFeeType + " -> " + newServiceFeeType,
				EN:   oldServiceFeeType + " -> " + newServiceFeeType,
				ZHTW: oldServiceFeeType + " -> " + newServiceFeeType,
				JA:   oldServiceFeeType + " -> " + newServiceFeeType,
				KO:   oldServiceFeeType + " -> " + newServiceFeeType,
				MY:   oldServiceFeeType + " -> " + newServiceFeeType,
				TR:   oldServiceFeeType + " -> " + newServiceFeeType,
				SV:   oldServiceFeeType + " -> " + newServiceFeeType,
			},
		})
		return res, newSetting, nil
	}
	// 检查固定服务费是否改变
	if oldSetting.ServiceFeeType == newSetting.ServiceFeeType && oldSetting.ServiceFeeType == constant.SaleBillSettingServiceFeeTypeFixed && oldSetting.ServiceFeeValue != newSetting.ServiceFeeValue {
		oldServiceFeeType := parseServiceFeeType(ctx.GetLanguage(), oldSetting.ServiceFeeType)
		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				EN:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				ZHTW: oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				JA:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				KO:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				MY:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TR:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				SV:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
			},
		})
		return res, newSetting, nil
	}
	// 检查"按比例-不收取税费"比例是否改变
	if oldSetting.ServiceFeeType == newSetting.ServiceFeeType && oldSetting.ServiceFeeType == constant.SaleBillSettingServiceFeeTypePercent && oldSetting.ServiceFeeValue != newSetting.ServiceFeeValue {
		oldServiceFeeType := parseServiceFeeType(ctx.GetLanguage(), oldSetting.ServiceFeeType)
		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				EN:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				ZHTW: oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				JA:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				KO:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				MY:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TR:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				SV:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
			},
		})
		return res, newSetting, nil
	}
	// 检查"按比例-收取税费"比例是否改变
	if oldSetting.ServiceFeeType == newSetting.ServiceFeeType && oldSetting.ServiceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax && oldSetting.ServiceFeeValue != newSetting.ServiceFeeValue {
		oldServiceFeeType := parseServiceFeeType(ctx.GetLanguage(), oldSetting.ServiceFeeType)
		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TH:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				EN:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				ZHTW: oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				JA:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				KO:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				MY:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				TR:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
				SV:   oldServiceFeeType + fmt.Sprintf(": %v -> %v", oldSetting.ServiceFeeValue, newSetting.ServiceFeeValue),
			},
		})
		return res, newSetting, nil
	}
	// 检查订单是否在服务费应用范围内是否改变
	if oldSetting.ServiceApply != newSetting.ServiceApply {

		res.OrderCheckRes.Products.List = append(res.OrderCheckRes.Products.List, resp.Product{
			Uuid: oldSetting.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				TH:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				EN:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				ZHTW: fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				JA:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				KO:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				MY:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				TR:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
				SV:   fmt.Sprintf("%v -> %v", parseServiceApply(ctx.GetLanguage(), oldSetting.ServiceApply), parseServiceApply(ctx.GetLanguage(), newSetting.ServiceApply)),
			},
		})
		return res, newSetting, nil
	}
	return nil, nil, nil
}

func parseServiceFeeType(language string, serviceFeeType uint) string {
	switch serviceFeeType {
	case constant.SaleBillSettingServiceFeeTypeNone:
		return i18n.Translate(language, "不收取服务费")
	case constant.SaleBillSettingServiceFeeTypeFixed:
		return i18n.Translate(language, "固定服务费")
	case constant.SaleBillSettingServiceFeeTypePercent:
		return i18n.Translate(language, "按比例收取服务费（服务费不收税）")
	case constant.SaleBillSettingServiceFeeTypePercentTax:
		return i18n.Translate(language, "按比例收取服务费（服务费收税）")
	default:
		return i18n.Translate(language, "未知")
	}
}

func parseTaxFeeType(language string, taxFeeType uint) string {
	switch taxFeeType {
	case constant.TaxFeeTypeTax:
		return i18n.Translate(language, "商品已含税")
	case constant.TaxFeeTypeNoTax:
		return i18n.Translate(language, "商品未含税")
	default:
		return i18n.Translate(language, "关闭消费税")
	}
}

func parseServiceApply(language string, serviceApply uint) string {
	switch serviceApply {
	case 1:
		return i18n.Translate(language, "订单在服务费应用范围内")
	case 0:
		return i18n.Translate(language, "订单不在服务费应用范围内")
	default:
		return i18n.Translate(language, "未知")
	}
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

		if unCookingSaleOrderProduct.IsCookingDeductStock() {
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
			if saleOrderProductBom.IsDelete() {
				continue
			}
			// 获取原材料的出库数量
			productBomMaterials := make([]*model.ProductBomMaterials, 0)
			// 如果是规格商品
			if saleOrderProductBom.IsFlavor() {
				// 如果没有成本卡，则使用规格商品的原材料
				flavorMaterials := saleOrderProductBom.ProductBom.FlavorMaterials
				// 如果有成本卡，则使用成本卡的原材料
				if saleOrderProductBom.ProductBom.HasProductBomCard() {
					ProductBomCard := saleOrderProductBom.ProductBom.ProductBomCard
					flavorMaterials = ProductBomCard.RelatedMaterials
				}
				// 遍历原材料
				for _, productBomMaterial := range flavorMaterials {
					if productBomMaterial.IsDelete() {
						continue
					}
					// 如果材料被禁用，则跳过，不扣减库存
					if productBomMaterial.Material.Status == false {
						continue
					}
					if num := productBomMaterial.GetDecreaseNum(cookingDeductSaleOrderProduct.Num); num > 0 {
						productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
							MaterialUuid:  productBomMaterial.MaterialUuid,
							WarehouseUuid: productBomMaterial.Material.WarehouseUuid,
							Num:           num,
							SaleOrderUuid: cookingDeductSaleOrderProduct.SaleOrderUuid,
						})
					}
				}
			}
			// 如果是小料
			if saleOrderProductBom.IsSauce() {
				// 如果没有成本卡，则使用小料的原材料
				sauceMaterials := saleOrderProductBom.ProductBom.ProductSauce.SauceMaterials
				// 如果有成本卡，则使用成本卡的原材料
				if saleOrderProductBom.ProductBom.ProductSauce.HasProductBomCard() {
					ProductBomCard := saleOrderProductBom.ProductBom.ProductSauce.ProductBomCard
					sauceMaterials = ProductBomCard.RelatedMaterials
				}
				// 遍历原材料
				for _, material := range sauceMaterials {
					// 如果材料被禁用，则跳过，不扣减库存
					if material.Material.Status == false {
						continue
					}
					if num := material.GetDecreaseNum(cookingDeductSaleOrderProduct.Num); num > 0 {
						productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
							MaterialUuid:  material.MaterialUuid,
							WarehouseUuid: material.Material.WarehouseUuid,
							Num:           num,
						})
					}
				}
			}
			// 获取规格商品的出库数量
			if cookingDeductSaleOrderProduct.Num > 0 {
				list = append(list, &model.Product{
					ProductBomUuid:       saleOrderProductBom.ProductBomUuid,
					PackageUuid:          cookingDeductSaleOrderProduct.PackageUuid,
					SaleOrderProductUuid: cookingDeductSaleOrderProduct.Uuid,
					SaleOrderUuid:        cookingDeductSaleOrderProduct.SaleOrderUuid,
					Num:                  cookingDeductSaleOrderProduct.Num,
					ProductBomMaterials:  productBomMaterials,
				})
			}
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

	// 从http的header中获取h5_order_uuid
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		req.H5OrderUuid = h5OrderUuid
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, nil, errSaleBill
	}
	ctx.Log().Debug("获取销售账单信息")

	h5OrderProductUnAccept := make([]*model.SaleOrderProduct, 0)
	if req.H5OrderUuid != 0 {
		h5OrderProductUnAccept = saleBill.GetH5OrderProductUnAccept(req.H5OrderUuid)
	}

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()
	nonNeedBatchSendCooking := false
	// 没有未送厨商品时，判断是否需要弹出分批送厨弹窗。只有收银机和助手端需要判断
	if ctx.GetSource() == constant.SourceAssistant || ctx.GetSource() == constant.SourceCashier {
		// 获取门店业务设置
		businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
		if err != nil {
			return nil, nil, errors.WithMessage(err)
		}
		if businessSetting.OpenIsBatch() {
			if saleBill.IsNeedBatchSendCooking() {
				nonNeedBatchSendCooking = true
			}
		}
	}

	if len(unCookingSaleOrderProducts) == 0 && len(h5OrderProductUnAccept) == 0 && !nonNeedBatchSendCooking {
		return nil, nil, errors.New("没有未送厨的商品")
	} else if len(unCookingSaleOrderProducts) == 0 && len(h5OrderProductUnAccept) == 0 && nonNeedBatchSendCooking {
		if !req.IgnoreMust { // 如果不是在结账检查时送厨，才返回-209，否则直接不分批送厨而是使用正常送厨
			return nil, &resp.OrderCheckServiceRes{
				Code: constant.CodeOrderCheckProductBatch,
			}, nil
		} else {
			// 获取预送厨的商品
			preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
			if len(preCookingSaleOrderProducts) > 0 {
				nonBathchUuids := make([]uint64, 0) // 预送厨的商品uuid列表
				for _, saleOrderProduct := range preCookingSaleOrderProducts {
					saleOrderProduct.IsBatch = 0
					nonBathchUuids = append(nonBathchUuids, saleOrderProduct.Uuid)
				}
				if len(nonBathchUuids) > 0 {
					// 将预送厨的商品变为未分批商品
					db := s.dbm.GetDB(ctx.GetDbId())
					if err := db.Model(&model.SaleOrderProduct{}).Where("uuid IN (?)", nonBathchUuids).Update("is_batch", 0).Error; err != nil {
						return nil, nil, errors.WithMessage(err)
					}
					// 将预送厨的生产单商品变为未分批商品
					if err := db.Model(&model.ProductionOrderProduct{}).Where("sale_order_product_uuid IN (?)", nonBathchUuids).Update("create_time", time.Now().Unix()).Update("is_batch", 0).Error; err != nil {
						return nil, nil, errors.WithMessage(err)
					}
				}
			}
		}
	}

	// 获取某个h5订单的已下单但未接单的商品
	if req.H5OrderUuid != 0 {
		checkRes, err := s.AcceptH5Order(ctx, req.H5OrderUuid, false)
		if err != nil {
			return nil, nil, err
		}
		if checkRes != nil {
			return nil, checkRes, nil
		}
	}

	// 送厨
	if len(unCookingSaleOrderProducts) > 0 {
		var checkServiceRes *resp.OrderCheckServiceRes
		var err error
		// 助手端，仅检查送厨时
		if ctx.GetSource() == constant.SourceAssistant && req.IsCheckCooking {
			checkServiceRes, err = s.ActionCooking(ctx, req.IgnoreMust, saleBill, unCookingSaleOrderProducts, 0, false, WithOnlyCheckCooking()) // 购物车送厨检查
		} else {
			checkServiceRes, err = s.ActionCooking(ctx, req.IgnoreMust, saleBill, unCookingSaleOrderProducts, 0, false, WithIsBatch()) // 购物车送厨商品
		}
		if err != nil {
			return nil, nil, err
		}
		if checkServiceRes != nil {
			return nil, checkServiceRes, nil
		}
	}

	// 会员端接单，不返回购物车信息
	if req.IsMemberOrderAccept {
		return nil, nil, nil
	}

	ctx.Log().Debug("获取新的购物车信息")
	cartInfo, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取购物车信息失败")
	}
	return cartInfo, nil, nil
}

func newProductionOrder(ctx context.Context, saleOrderUuid, saleBillUuid, deskUuid uint64, unCookingSaleOrderProducts []*model.SaleOrderProduct) *model.ProductionOrder {
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
			InitNum:               unCookingSaleOrderProduct.Num,
			FlavorName:            unCookingSaleOrderProduct.Name,
			ProductAttributeNames: attributeName.ToJson(),
			Status:                constant.ProductionOrderProductStatusCooking,
			Remark:                unCookingSaleOrderProduct.Remark,
			// TODO 植焕
			//HasMaterial:              unCookingSaleOrderProduct,
			ProductionOrderMaterials: unCookingSaleOrderProduct.GetMaterialBom(), // 获取这个商品各个材料的用量
			IsBatch: func() uint8 {
				if unCookingSaleOrderProduct.IsBatchBool() {
					// 如果是结账送厨时，标记为不是分批商品

					return 1
				}
				return 0
			}(),
		}
		productionOrderProducts = append(productionOrderProducts, &productionOrderProduct)
	}
	productionOrder := model.ProductionOrder{
		BaseModel:               model.BaseModel{Uuid: productionOrderUuid},
		DeskUuid:                deskUuid,
		SaleOrderUuid:           saleOrderUuid,
		SaleBillUuid:            saleBillUuid,
		ProductionOrderProducts: productionOrderProducts,
		Source:                  ctx.GetSource(),
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderRefundProduct, req.SaleOrderUuid); err != nil {
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
	// var returnSubSaleOrderProducts []*model.SaleOrderProduct // 套餐子商品退菜列表
	var keepNum float64

	// 如果退菜数量等于该商品的数量，则标记该商品为退菜并在商品的退菜原因列表中添加退菜原因
	if saleOrderProduct.Num == req.Num {
		// 尝试获取相同签名的商品，如果有，将两个商品合并。该商品可能已经退过菜
		newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderProduct.SaleOrderUuid)
		newSaleOrderProduct.SetNum(req.Num)
		newSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
		if sameSignSaleOrderProduct != nil {
			// 有相同签名的商品。将两个商品合并，数量相加
			sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
			returnSaleOrderProduct = sameSignSaleOrderProduct
			saleOrderProduct.SetDelete() // 标记该商品为删除
			saleOrderProduct.SetUpdate() // 标记该商品需要更新
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					subProduct.SetDelete() // 标记该商品为删除
					subProduct.SetUpdate() // 标记该商品需要更新
				}
			}
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if sameSignSaleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(sameSignSaleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					unitNum := decimal.NewFromFloat(subProduct.UnitNum)
					addNum := decimal.NewFromFloat(req.Num).Mul(unitNum).Round(3).InexactFloat64()
					subProduct.SetNum(subProduct.Num + addNum)
				}
			}

		} else {
			saleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
			returnSaleOrderProduct = saleOrderProduct
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					subProduct.SetCancelInfo(req.Reason, nil)
				}
			}
		}
	} else {
		// 如果退菜数量小于该商品的数量，则新建一个销售订单商品并在新商品的退菜原因列表中添加退菜原因
		// 1. 修改原商品的数量
		// 2. 新建一个销售订单商品，该商品数量为退菜数量.
		// 3. 判断新建的销售订单商品是否要合并到已有的退菜商品中。当两个退菜商品的签名一致时，将两个商品合并，数量相加
		// 修改原商品的数量
		keepNum = saleOrderProduct.Num - req.Num
		saleOrderProduct.SetNum(keepNum)
		// 新建一个销售订单商品，该商品数量为退菜数量
		newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderProduct.SaleOrderUuid)
		newSaleOrderProduct.SetNum(req.Num)
		newSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
		if sameSignSaleOrderProduct != nil {
			// 有相同签名的商品。将两个商品合并，数量相加
			sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
			returnSaleOrderProduct = sameSignSaleOrderProduct
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(sameSignSaleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					unitNum := decimal.NewFromFloat(subProduct.UnitNum)
					num := decimal.NewFromFloat(sameSignSaleOrderProduct.Num).Mul(unitNum).Round(3).InexactFloat64()
					subProduct.SetNum(num)
				}
			}
		} else {
			// 没有相同签名的商品。将新建的商品添加到销售订单商品列表中
			// CalcAndSaveSaleBill 方法会检查到newSaleOrderProduct没有主键ID，会创建新记录。所以不用另外创建该订单商品，否则会重复创建
			saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, newSaleOrderProduct)
			returnSaleOrderProduct = newSaleOrderProduct
			// 补全newSaleOrderProduct对象的ProductBom对象
			for _, saleOrderProductBom := range newSaleOrderProduct.SaleOrderProductBoms {
				productBom, err := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(saleOrderProductBom.ProductBomUuid)
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				saleOrderProductBom.ProductBom = *productBom
			}

			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					newSubSaleOrderProduct := subProduct.CopyOrderProduct(subProduct.SaleOrderUuid)
					newSubSaleOrderProduct.PackageUuid = newSaleOrderProduct.Uuid // 设置为新的套餐商品uuid
					unitNum := decimal.NewFromFloat(subProduct.UnitNum)
					num := decimal.NewFromFloat(req.Num).Mul(unitNum).Round(3).InexactFloat64()
					newSubSaleOrderProduct.SetNum(num)
					newSubSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
					saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, newSubSaleOrderProduct)
				}
			}
		}
		// 更新旧的子商品的数量
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				unitNum := decimal.NewFromFloat(subProduct.UnitNum)
				num := decimal.NewFromFloat(keepNum).Mul(unitNum).Round(3).InexactFloat64()
				subProduct.SetNum(num)
			}
		}
	}

	// 如果退菜商品是下单减库存的商品，则需要创建入库单
	var warehouseForm *model.WarehouseForm
	if returnSaleOrderProduct.DeductStockType == constant.ProductPackageDeductStockTypeCooking {
		currentReturnSaleOrderProduct := *returnSaleOrderProduct
		currentReturnSaleOrderProduct.Num = req.Num // 退菜数量,仅入库本次退菜的数量

		cookingDeductSaleOrderProducts := []*model.SaleOrderProduct{&currentReturnSaleOrderProduct}

		// 如果是套餐商品，则加上子商品的退菜信息
		if currentReturnSaleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(currentReturnSaleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				product := *subProduct
				product.Num = decimal.NewFromFloat(req.Num).Mul(decimal.NewFromFloat(subProduct.UnitNum)).Round(3).InexactFloat64() // 退菜数量,仅入库本次退菜的数量
				cookingDeductSaleOrderProducts = append(cookingDeductSaleOrderProducts, &product)
			}
		}

		productList, err := s.getDecreaseStockList(ctx, cookingDeductSaleOrderProducts)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		warehouseForm = model.NewWarehouseForm(productList, req.SaleBillUuid)
	}

	// 新建一个销售订单商品，该商品数量为移动数量
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 如果退菜商品还关联着分批标签，则解除
		if returnSaleOrderProduct.IsBatchBool() {
			returnSaleOrderProduct.BatchTagUuid = 0 // 手动置0，否则可能又被updateSaleOrderProduct方法更新回非零值
			returnSaleOrderProduct.BatchTime = 0
			if err := tx.Model(&model.SaleOrderProduct{}).Where("uuid = ?", returnSaleOrderProduct.Uuid).Updates(map[string]any{
				"batch_tag_uuid": 0,
				"batch_time":     0,
			}).Error; err != nil {
				return errors.WithMessage(err)
			}
		}
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

		// 修改送厨商品
		productionRepo := repository.NewProductionRepo(tx)
		if keepNum > 0 { // 修改数量
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)},
				map[string]any{"num": keepNum}); err != nil {
				return errors.WithMessage(err)
			}
		} else { // 修改数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)},
				map[string]any{"num": 0}); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 修改送厨子商品的数量
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				unitNum := decimal.NewFromFloat(subProduct.UnitNum)
				num := decimal.NewFromFloat(keepNum).Mul(unitNum).Round(3).InexactFloat64()
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(subProduct.Uuid)},
					map[string]any{"num": num}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}

	// 发布"退菜"事件
	go func() {
		var convertToEventOrderProduct func(saleOrderProduct *model.SaleOrderProduct, isSubProduct bool) event.CancelSaleOrderProductProduct
		convertToEventOrderProduct = func(saleOrderProduct *model.SaleOrderProduct, isSubProduct bool) event.CancelSaleOrderProductProduct {
			return event.CancelSaleOrderProductProduct{
				OrderProductId:  req.SaleOrderProductUuid,
				ProductId:       saleOrderProduct.ProductPackageUuid,
				ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
				ProductAttr:     saleOrderProduct.GetAttributeName(),
				ProductAttrList: saleOrderProduct.GetAttributeNameList(),
				TotalNum: func() float64 {
					if !isSubProduct {
						return req.Num
					}
					return decimal.NewFromFloat(req.Num).Mul(decimal.NewFromFloat(saleOrderProduct.UnitNum)).Round(3).InexactFloat64()
				}(),
				IsBuffet:     saleOrderProduct.IsBuffet == 1,
				Remark:       saleOrderProduct.Remark,
				Reason:       model.GetReturnFoodReasonNames(returnFoodReason),
				CustomReason: saleOrderProduct.CancelReason,
				Sign:         saleOrderProduct.Sign,
				SubProducts: func() []event.CancelSaleOrderProductProduct {
					if saleOrderProduct.IsPackageProduct() {
						subProducts := make([]event.CancelSaleOrderProductProduct, 0)
						for _, subProduct := range saleOrder.GetPackageSubProductList(returnSaleOrderProduct.Uuid) {
							subProducts = append(subProducts, convertToEventOrderProduct(subProduct, true))
						}
						return subProducts
					}
					return nil
				}(),
			}
		}
		s.bus.PublishCancelSaleOrderProductEvent(event.CancelSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 退菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			CancelSaleOrderProductProduct: convertToEventOrderProduct(returnSaleOrderProduct, false),
		})
	}()

	// 送厨成功后，推送更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
	})

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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
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
			constant.ProductReasonTypeReturnFood,
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
		// 更新套餐子商品
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					Status:       0,
					CancelTime:   0,
					CancelReason: "",
				}, map[string]bool{
					"Status":       true,
					"CancelTime":   true,
					"CancelReason": true,
				})
			}
		}
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
			BasePayload: event.BasePayload{ // 取消退菜
				Ctx:           ctx,
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
			Sign:           saleOrderProduct.Sign,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderChangeTable, req.SaleOrderUuid); err != nil {
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
	// 不能转菜到自助餐桌台
	if targetSaleBill.IsBuffetSaleBill() {
		return nil, errors.WithMessage(errors.New("不能转菜到自助餐桌台"))
	}

	// 将销售订单商品的sale_bill_uuid和sale_order_uuid更新为新的桌台
	targetSaleOrder := targetSaleBill.GetFirstSaleOrder()

	// 判断订单状态
	if err := targetSaleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 更新数据
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 处理订单商品
		var handleOrderProducts func(saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct) error
		handleOrderProducts = func(saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct) error {
			// 获取原生产单Uuid
			oldProductionOrderUuid := saleOrderProduct.ProductionOrderUuid

			var newProductionOrderUuid uint64
			if oldProductionOrderUuid != 0 {
				newProductionOrderUuid, _ = utils.GetID()
			}

			// 修改生产单Uuid
			saleOrderProduct.ProductionOrderUuid = newProductionOrderUuid
			// 设置已接单
			saleOrderProduct.SetAcceptOrderProduct()
			// h5订单只有一个商品时拒单
			if saleOrderProduct.H5OrderUuid != 0 {
				// 如果这个商品是h5订单的最后一个商品，则删除拒单该h5订单
				s.deleteOrRejectH5OrderProduct(ctx, db, saleOrderProduct)
				// 删除掉h5订单uuid
				saleOrderProduct.H5OrderUuid = 0
			}
			// 更新销售订单商品
			saleOrderProduct.SaleBillUuid = targetDesk.SaleBillUuid
			saleOrderProduct.SaleOrderUuid = targetSaleOrder.Uuid
			// 将商品的折扣改为使用目标桌台的折扣
			saleOrderProduct.MemberDiscountRate = targetSaleOrder.MemberDiscountRate
			saleOrderProduct.MemberCardDiscountRate = targetSaleOrder.MemberCardDiscountRate
			saleOrderProduct.CustomDiscountRate = targetSaleOrder.CustomDiscountRate
			// 判断商品是否在目标桌台是否是自助餐商品
			productPackageUuidMap := targetSaleBill.GetBuffetProductMap()
			if _, ok := productPackageUuidMap[saleOrderProduct.ProductPackageUuid]; ok {
				saleOrderProduct.SetIsBuffet()
			} else {
				saleOrderProduct.SetNotBuffet()
			}
			saleOrderProduct.SetUpdate() // 标记该商品的记录要更新，会在原桌台账单的CalcAndSaveSaleBill方法中更新

			// 将商品添加到目标桌台的销售订单中
			targetSaleOrder.SaleOrderProducts = append(targetSaleOrder.SaleOrderProducts, saleOrderProduct)

			// 重新计算原桌台的销售账单
			if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
				return errors.WithMessage(err)
			}

			// 重新计算目标桌台的销售账单
			if err := s.CalcAndSaveSaleBill(ctx, tx, targetSaleBill); err != nil {
				return errors.WithMessage(err)
			}

			// 更新生产单
			productionRepo := repository.NewProductionRepo(db)
			oldProductionOrder, _ := productionRepo.GetProductionOrder(productionRepo.WhereUuid(oldProductionOrderUuid))
			if oldProductionOrder != nil {
				if err := productionRepo.CreateProductionOrder(&model.ProductionOrder{
					BaseModel:     model.BaseModel{Uuid: newProductionOrderUuid},
					DeskUuid:      targetDesk.Uuid,
					SaleOrderUuid: targetSaleOrder.Uuid,
					SaleBillUuid:  targetSaleBill.Uuid,
					Source:        oldProductionOrder.Source,
				}); err != nil {
					return errors.WithMessage(err)
				}
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)}, map[string]any{
					"sale_bill_uuid":        targetSaleBill.Uuid,
					"production_order_uuid": newProductionOrderUuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}

			// 如果是套餐商品，则更新套餐子商品
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					handleOrderProducts(saleOrder, subProduct)
				}
			}

			return nil
		}
		return handleOrderProducts(saleOrder, saleOrderProduct)

	}); err != nil {
		return nil, errors.WithMessage(err, "更新数据失败")
	}

	// 发布转菜事件
	go func() {
		s.bus.PublishChangeDeskSaleOrderProductEvent(event.ChangeDeskSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 转菜
				Ctx:           ctx,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
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
	// 设置赠菜时间
	saleOrderProduct.SetGiftProduct(req.Reason)
	// 如果是套餐商品，则更新套餐子商品赠菜时间
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetGiftProduct(req.Reason)
		}
	}

	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			GiftTime:   saleOrderProduct.GiftTime,
			Sign:       saleOrderProduct.Sign,
			GiftReason: saleOrderProduct.GiftReason,
		})
		// 更新套餐子商品赠菜时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					GiftTime:   subProduct.GiftTime,
					Sign:       subProduct.Sign,
					GiftReason: subProduct.GiftReason,
				})
			}
		}
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
			BasePayload: event.BasePayload{ // 赠菜
				Ctx:           ctx,
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

// OrderCartProductWrap 打包购物车商品
func (s *orderSrv) OrderCartProductWrap(ctx context.Context, req req.OrderCartProductWrapReq) (*resp.ShopCart, error) {
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
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

	// 设置打包时间
	saleOrderProduct.SetWrap()

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetWrap()
		}
	}

	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			WrapTime: saleOrderProduct.WrapTime,
			Sign:     saleOrderProduct.Sign,
		})
		// 更新套餐子商品打包时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					WrapTime: subProduct.WrapTime,
					Sign:     subProduct.Sign,
				})
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
	// 发布打包事件
	go func() {
		s.bus.PublishWrapSaleOrderProductEvent(event.WrapSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 打包
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleOrderProductUuid: req.SaleOrderProductUuid,
			ProductPackageUuid:   saleOrderProduct.ProductPackageUuid,
			ProductName:          saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:          saleOrderProduct.GetAttributeName(),
			ProductPrice:         saleOrderProduct.Price,
			Num:                  saleOrderProduct.Num,
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

// OrderCartProductUnwrap 取消打包购物车商品
func (s *orderSrv) OrderCartProductUnwrap(ctx context.Context, req req.OrderCartProductUnwrapReq) (*resp.ShopCart, error) {
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
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

	// 设置打包时间
	saleOrderProduct.SetUnwrap()
	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetUnwrap()
		}
	}
	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			WrapTime: saleOrderProduct.WrapTime,
		})
		// 更新套餐子商品打包时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					WrapTime: subProduct.WrapTime,
				})
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
	// 发布打包事件
	go func() {
		s.bus.PublishUnwrapSaleOrderProductEvent(event.UnwrapSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 打包
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleOrderProductUuid: req.SaleOrderProductUuid,
			ProductPackageUuid:   saleOrderProduct.ProductPackageUuid,
			ProductName:          saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:          saleOrderProduct.GetAttributeName(),
			ProductPrice:         saleOrderProduct.Price,
			Num:                  saleOrderProduct.Num,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
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

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
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
		// 更新套餐子商品赠菜时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					GiftTime:   0,
					GiftReason: "",
				}, map[string]bool{
					"GiftTime":   true,
					"GiftReason": true,
				})
			}
		}
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
			BasePayload: event.BasePayload{ // 取消赠菜
				Ctx:           ctx,
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

type InstanceAutoFlavorProduct map[uint64]resp.ProductAutoAddReq

// InstantOrderMustPlan 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, bool, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx)
	if errUuid != nil {
		return nil, false, errors.WithMessage(errUuid, "无法找到销售账单")
	}
	ctx.Log().Debug("查询必点方案列表", zap.Any("saleBillUuid", saleBillUuid), zap.Any("deviceSn", deviceSn))

	mustPlanList := make([]resp.InstantProductMustPlan, 0)
	// must_plan_uuid => autoFlavorProduct
	planAutoFlavorProduct := make(map[uint64]InstanceAutoFlavorProduct) // 必点方案ID => 自动加购的商品列表. 用于记录每个必点方案的自动加购商品

	// 查询到购物车信息
	shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, false, errors.WithMessage(err, "repository.NewOrderRepo(db).GetOrderCartInfo failed", fmt.Sprintf("saleBillUuid:%d", saleBillUuid))
	}

	planList, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
	if errMustPlan != nil {
		ctx.Log().Info("获取必点列表失败", zap.Error(errMustPlan))
		return nil, false, errors.New("获取必点列表失败")
	}
	mustPlanList = planList
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// 遍历得到要自动加购的商品
	for i, plan := range mustPlanList {
		// product_bom_uuid => *resp.InstantMustPlanProduct
		autoFlavorProduct := make(map[uint64]resp.ProductAutoAddReq) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购
		for j, product := range plan.Products.List {
			if product.IsAutoAdd {
				planProduct := mustPlanList[i].Products.List[j].Product
				productFlavorBomUuid := planProduct.Flavors.List[0].Uuid
				productFlavorStockNum := planProduct.Flavors.List[0].StockNum
				if productFlavorStockNum > 0 {
					autoFlavorProduct[productFlavorBomUuid] = product.ProductAutoAddReq
				}
			}
		}
		if len(autoFlavorProduct) > 0 {
			planAutoFlavorProduct[plan.Uuid] = autoFlavorProduct
		}
	}

	var shopCart *resp.ShopCart
	var isAutoAdd bool
	// 判断是否需要给点餐账单自动加购商品。当map列表中有商品时，表示需要自动加购
	if len(planAutoFlavorProduct) > 0 && shopCartInfo.SaleBill.IsAutoAddMustProduct() {
		errTx := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			shopCart, err = autoAddSaleOrderProduct(ctx, tx, s, planAutoFlavorProduct)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			// 设置已经完成自动加购
			if err := repository.NewSaleBillRepo(tx).UpdateSaleBillAutoAddMustProduct(saleBillUuid); err != nil {
				return errors.WithMessage(err, "标记自动加购完成失败")
			}
			return nil
		})
		if errTx != nil {
			return nil, false, errors.WithMessage(errTx, "自动添加必点商品失败")
		}
		isAutoAdd = true
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
	mustPlan := &resp.InstantProductMustPlanResp{List: list, ShopCartInfo: cartInfo}
	return mustPlan, isAutoAdd, nil
}

type AutoFlavorProduct map[uint64]resp.ProductAutoAddReq

func (s *orderSrv) DeskOrderMustPlan(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64, mealNum uint, h5AutoAdd bool, noAutoAdd bool) (*resp.InstantProductMustPlanResp, bool, error) {
	db := ctx.GetDB()

	mustPlanList := make([]resp.InstantProductMustPlan, 0)
	// must_plan_uuid => autoFlavorProduct
	planAutoFlavorProduct := make(map[uint64]AutoFlavorProduct) // 必点方案ID => 自动加购的商品列表. 用于记录每个必点方案的自动加购商品

	// 查询到购物车信息
	var shopCartInfo *ro.ShopCartRepo
	var err error
	if h5AutoAdd {
		// 如果是H5自动加购时，查询未删除的所有商品（主要是要包含h5加购但未下单的商品）
		shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid, repository.WithNotDeleted())
	} else {
		shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	}
	if err != nil {
		return nil, false, errors.WithMessage(err, "repository.NewOrderRepo(db).GetOrderCartInfo failed", fmt.Sprintf("saleBillUuid:%d", saleBillUuid))
	}

	if shopCartInfo.SaleBill.IsBuffetSaleBill() {
		// 如果是自助餐桌台，将必点商品弹框中的自助餐商品价格标记为0元
		mustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, mealNum, shopCartInfo.GetMustPlanProductInfo(), shopCartInfo.SaleBill.DeskUuid, WithSaleBillUuid(shopCartInfo.SaleBill.Uuid))
	} else {
		mustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, mealNum, shopCartInfo.GetMustPlanProductInfo(), shopCartInfo.SaleBill.DeskUuid)
	}
	if err != nil {
		ctx.Log().Info("获取必点列表失败", zap.Error(err))
		return nil, false, errors.WithMessage(err, "获取必点列表失败")
	}
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// 遍历得到要自动加购的商品
	for i, plan := range mustPlanList {
		// product_bom_uuid => *resp.InstantMustPlanProduct
		autoFlavorProduct := make(map[uint64]resp.ProductAutoAddReq) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购
		for j, product := range plan.Products.List {
			if product.IsAutoAdd {
				planProduct := mustPlanList[i].Products.List[j].Product
				productFlavorBomUuid := planProduct.Flavors.List[0].Uuid
				productFlavorStockNum := planProduct.Flavors.List[0].StockNum
				if productFlavorStockNum > 0 {
					product.ProductAutoAddReq.Num = mustPlanList[i].Products.List[j].MustNum // 商品数量
					autoFlavorProduct[productFlavorBomUuid] = product.ProductAutoAddReq
				}
			}
		}
		if len(autoFlavorProduct) > 0 {
			planAutoFlavorProduct[plan.Uuid] = autoFlavorProduct
		}
	}

	isAutoAdd := false

	// 1. 是否自动加购。平板不自动加购
	// 2. 判断是否需要给点餐账单自动加购商品。当map列表中有商品时，表示需要自动加购
	if !noAutoAdd && len(planAutoFlavorProduct) > 0 && shopCartInfo.SaleBill.IsAutoAddMustProduct() {
		errTx := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			ctx.SetDB(tx)
			// 自动加购
			err = autoAddSaleOrderProductToDesk(ctx, s, planAutoFlavorProduct, saleBillUuid, saleOrderUuid, shopCartInfo.SaleBill, h5AutoAdd)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			return nil
		})
		if errTx != nil {
			return nil, false, errors.WithMessage(errTx, "自动添加必点商品失败")
		}
		isAutoAdd = true
	}

	// 获取新的购物车商品数据
	ctx.SetDB(db)
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid, repository.WithNoQueryMustPlan())
	if err != nil {
		return nil, false, errors.WithMessage(err)
	}

	cartInfo := &resp.InstantShopCart{
		SaleBillUuid:  shopCart.SaleBillUuid,
		DiningMethod:  shopCart.DiningMethod,
		SaleOrderList: shopCart.SaleOrderList,
	}

	list := make([]resp.InstantProductMustPlan, 0)
	if shopCartInfo.SaleBill.IsShowMustPlan() {
		list = mustPlanList
	}
	return &resp.InstantProductMustPlanResp{List: list, ShopCartInfo: cartInfo}, isAutoAdd, nil
}

func autoAddSaleOrderProduct(ctx context.Context, db *gorm.DB, s *orderSrv, planAutoFlavorProduct map[uint64]InstanceAutoFlavorProduct) (*resp.ShopCart, error) {
	var saleBillUuid uint64
	deviceUuid := ctx.GetDeviceUuid()
	if deviceUuid == 0 {
		ctx.Log().Debug("自动加购必选商品失败，上下文中没有device_uuid")
		return nil, errors.New("自动加购必选商品失败，上下文中没有device_uuid")
	}
	// 通过设备ID查询未挂单的销售账单
	saleBill, errGetSaleBill := repository.NewSaleBillRepo(db).GetSaleBillByDeviceUuid(deviceUuid)
	if errGetSaleBill != nil {
		if !utils.IsNotFoundRecord(errGetSaleBill) {
			return nil, errors.New(errGetSaleBill.Error())
		}
	}
	if saleBill != nil {
		saleBillUuid = saleBill.Uuid
	}

	var shopCartInfo *resp.ShopCart
	if saleBillUuid == 0 {
		// 如果没有未挂单的点餐销售账单，则新建一个点餐销售账单并接入自动必点商品
		ctx.Log().Debug("没有未挂单的点餐销售账单，新建一个点餐销售账单并接入自动必点商品")
		saleBillUuid := uint64(0)
		saleOrderUuid := uint64(0)
		for mustPlanUuid, autoFlavorProduct := range planAutoFlavorProduct {
			for flavorUuid, autoAddReq := range autoFlavorProduct {
				if saleBillUuid != 0 {
					ctx.Log().Debug("新建的销售账单号", zap.Any("saleBillUuid", saleBillUuid))
				}
				ctx.Log().Debug("添加商品", zap.Any("flavorUuid", flavorUuid))
				shopCart, errAdd := s.InstantOrderCartProductAdd(ctx, req.OrderCartProductAddReq{
					SaleBillUuid:      saleBillUuid,
					SaleOrderUuid:     saleOrderUuid,
					FlavorUuid:        autoAddReq.FlavorUuid,
					AttributeUuidList: autoAddReq.AttributeUuidList,
					SauceUuidList:     autoAddReq.SauceUuidList,
					MustPlanUuid:      mustPlanUuid,
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
		}
	} else {
		// 如果有未挂单的点餐销售账单。未有这个需求，暂时不做
	}

	// 如果购物车信息为空，则返回空
	if shopCartInfo == nil {
		return nil, nil
	}

	// 如果已经完成了必点，则关闭必点弹窗
	finish := true
	if shopCartInfo.MustPlans != nil {
		for _, plan := range shopCartInfo.MustPlans.List {
			if plan.NeedNum != 0 {
				finish = false
			}
		}
	}

	if finish {
		// 关闭必点弹窗
		shopCartInfo.MustPlans = nil
		// 如果已经自动加购完成，则不在显示必点方案.并更新sale_bill为已完成必点
		err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(shopCartInfo.SaleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	return shopCartInfo, nil
}

func autoAddSaleOrderProductToDesk(ctx context.Context, s *orderSrv, planAutoFlavorProduct map[uint64]AutoFlavorProduct, saleBillUuid, saleOrderUuid uint64, saleBill *model.SaleBill, isH5AutoAdd bool) error {
	productParams := make([]req.ProductParams, 0)

	for mustPlanUuid, autoFlavorProduct := range planAutoFlavorProduct {
		for flavorUuid, autoAddReq := range autoFlavorProduct {
			productParams = append(productParams, req.ProductParams{
				FlavorProductBomUuid:            flavorUuid,
				ProductPackageAttributeUuidList: autoAddReq.AttributeUuidList,
				SauceProductBomUuidList:         autoAddReq.SauceUuidList,
				MustPlanUuid:                    mustPlanUuid,
				Num:                             autoAddReq.Num,
			})
		}
	}
	// 加购
	db := ctx.GetDB()
	errAdd := s.ActionAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
		Products:      productParams,
		IsH5Product:   isH5AutoAdd,
	}, saleBill)
	if errAdd != nil {
		ctx.Log().Info("自动加购必选商品失败", zap.Error(errAdd))
	}
	// 更新sale_bill为自动加购完成
	saleBill.AutoAddMustProduct = constant.AutoAddMustProductNo
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// InstantOrderMustPlan2 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan2(ctx context.Context, deviceSn string) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取点餐的必选方案列表
	mustPlanList, err := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, make(ro.MustPlanProductInfo))
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取必点列表失败"), fmt.Sprintf("err:%v", err))
	}
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// must_plan_uuid => autoFlavorProduct
	planAutoFlavorProduct := make(map[uint64]InstanceAutoFlavorProduct) // 必点方案ID => 自动加购的商品列表. 用于记录每个必点方案的自动加购商品

	// 遍历得到要自动加购的商品
	for i, plan := range mustPlanList {
		// product_bom_uuid => *resp.InstantMustPlanProduct
		autoFlavorProduct := make(map[uint64]resp.ProductAutoAddReq) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购
		for j, product := range plan.Products.List {
			if product.IsAutoAdd {
				planProduct := mustPlanList[i].Products.List[j].Product
				productFlavorBomUuid := planProduct.Flavors.List[0].Uuid
				productFlavorStockNum := planProduct.Flavors.List[0].StockNum
				if productFlavorStockNum > 0 {
					autoFlavorProduct[productFlavorBomUuid] = product.ProductAutoAddReq
				}
			}
		}
		if len(autoFlavorProduct) > 0 {
			planAutoFlavorProduct[plan.Uuid] = autoFlavorProduct
		}
	}

	var shopCart *resp.ShopCart
	// 判断是否需要给点餐账单自动加购商品。当map列表中有商品时，表示需要自动加购
	if len(planAutoFlavorProduct) > 0 {
		err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			cart, err := autoAddSaleOrderProduct(ctx, tx, s, planAutoFlavorProduct)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			shopCart = cart
			return nil
		})
		if err != nil {
			return nil, errors.WithMessage(err, "自动添加必点商品失败")
		}
	} else {
		shopCart = &resp.ShopCart{
			SaleOrderList: []resp.SaleOrder{},
			MustPlans:     &resp.ProductMustPlanList{List: mustPlanList},
		}
	}

	return shopCart, nil
}

// GetValidMemberCouponList 获取有效的会员优惠券列表,包括通用优惠券和会员优惠券
func (s *orderSrv) GetValidMemberCouponList(ctx context.Context, memberUuid uint64, selectedCouponUuid uint64, hasPay, hasCommonCoupon bool, saleOrderAmount float64) (*resp.CouponList, error) {
	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if !companySetting.GetIsOpenCoupon() {
		// 未开启优惠券
		return nil, nil
	}

	shopTimeZone := storeSetting.TimeZone
	layout := "2006-01-02"

	timezone := utils.SetTimezone(shopTimeZone)
	coupons := make(map[uint64]*resp.Coupon, 0) // key为coupon_uuid value为Coupon有效期内的优惠券，包括通用优惠券和会员优惠券
	// 查询本店所有none类型的优惠券
	commonCouponList, err := repository.NewMarketingCouponRepo(ctx.GetDB()).GetValidCouponList()
	if err != nil {
		return nil, errors.WithMessage(err, "查询通用优惠券列表失败")
	}

	for _, coupon := range commonCouponList {
		endTime := timezone.FormatUnixTime(int64(coupon.ValidEndTime), layout)
		coupons[coupon.Uuid] = &resp.Coupon{
			Uuid:           coupon.Uuid,
			Name:           coupon.Name,
			Requirement:    coupon.Requirement,
			Amount:         coupon.Amount,
			Count:          1,     // 通用优惠券数量为1
			IsSelected:     false, // 默认未选中，另外在判断是否被选中
			IsAvailable:    false, // 另外再判断是否在使用时段内
			DayStartTime:   coupon.DayStartTime,
			DayEndTime:     coupon.DayEndTime,
			ValidStartTime: timezone.FormatUnixTime(int64(coupon.ValidStartTime), layout),
			ValidEndTime:   endTime,
			Sort: resp.SortParam{
				Type:              coupon.Requirement,
				ValidEndTimestamp: int64(coupon.ValidEndTime),
				Sort:              coupon.Sort,
				CouponUuid:        coupon.Uuid,
			},
		}
	}

	// 如果订单使用了会员，查询会员有效期内的优惠券列表
	if memberUuid != 0 {
		memberCouponList, err := repository.NewMemberCouponRepo(ctx.GetDB()).GetValidMemberCouponList(memberUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "查询会员优惠券列表失败")
		}
		// 根据CouponUuid分类，然后再根据有效期和时段分类（有效期开始时间、结束时间、使用时段开始时间、使用时段结束时间相同的合并）
		// key规则， CouponUuid_有效期开始时间_有效期结束时间_使用时段开始时间_使用时段结束时间
		couponMap := make(map[string][]*model.MemberCoupon, 0)

		for _, coupon := range memberCouponList {
			// 过滤掉已经被禁用的营销优惠券
			if coupon.MarketingCoupon.Status == 0 { // 0禁用 1开启
				continue
			}
			startTime := timezone.FormatUnixTime(coupon.StartTime, layout)
			endTime := timezone.FormatUnixTime(coupon.EndTime, layout)
			key := fmt.Sprintf("%d_%s_%s_%s_%s", coupon.CouponUuid, startTime, endTime, coupon.DayStartTime, coupon.DayEndTime)
			if _, ok := couponMap[key]; !ok {
				couponMap[key] = make([]*model.MemberCoupon, 0)
			}

			couponMap[key] = append(couponMap[key], coupon)

		}

		for _, memberCouponList := range couponMap {
			var targetCoupon *model.MemberCoupon // 同一个key的优惠券列表中，先到期的优惠券
			for _, coupon := range memberCouponList {
				if targetCoupon == nil {
					targetCoupon = coupon
				} else {
					if coupon.EndTime < targetCoupon.EndTime {
						targetCoupon = coupon
					}
				}
			}
			startTime := timezone.FormatUnixTime(targetCoupon.StartTime, layout)
			endTime := timezone.FormatUnixTime(targetCoupon.EndTime, layout)

			// sampleMemberCouponList 列表指收音机中某个会员优惠券是由这个列表中的会员优惠券聚合而成
			sampleMemberCouponList := make([]*resp.SampleMemberCoupon, 0)
			for _, memberCoupon := range memberCouponList {
				sampleMemberCouponList = append(sampleMemberCouponList, &resp.SampleMemberCoupon{
					Uuid:         memberCoupon.Uuid,
					Name:         memberCoupon.Name,
					CouponUuid:   memberCoupon.CouponUuid,
					StartTime:    memberCoupon.StartTime,
					EndTime:      memberCoupon.EndTime,
					DayStartTime: memberCoupon.DayStartTime,
					DayEndTime:   memberCoupon.DayEndTime,
				})
			}

			// 添加一个聚合好的会员优惠券
			coupons[targetCoupon.Uuid] = &resp.Coupon{
				Uuid:           targetCoupon.Uuid,
				Name:           targetCoupon.MarketingCoupon.Name,
				Requirement:    targetCoupon.MarketingCoupon.Requirement,
				Amount:         targetCoupon.Amount,
				Count:          len(memberCouponList), // 会员优惠券数量为1
				IsSelected:     false,                 // 默认未选中，另外在判断是否被选中
				IsAvailable:    false,                 // 另外再判断是否在使用时段内
				DayStartTime:   targetCoupon.MarketingCoupon.DayStartTime,
				DayEndTime:     targetCoupon.MarketingCoupon.DayEndTime,
				ValidStartTime: startTime,
				ValidEndTime:   endTime,
				CouponList:     sampleMemberCouponList,
				Sort: resp.SortParam{
					Type:              targetCoupon.MarketingCoupon.Requirement,
					ValidEndTimestamp: targetCoupon.EndTime,
					Sort:              targetCoupon.MarketingCoupon.Sort,
					CouponUuid:        targetCoupon.Uuid,
				},
			}
		}
	}
	// 判断是否被选中
	for _, coupon := range coupons {
		if coupon.Uuid == selectedCouponUuid {
			coupon.IsSelected = true
			coupon.IsAvailable = true // 被选择的优惠券即使不在使用时段内也应该可以操作
			break
		}
	}
	// 判断是否在使用时段内
	nowTime := timezone.FormatUnixTime(timezone.NowUnix(), "15:04")
	for _, coupon := range coupons {
		isAvailable, err := utils.IsTimeInRange(nowTime, coupon.DayStartTime, coupon.DayEndTime)
		if err != nil {
			ctx.Log().Info("判断是否在使用时段内失败", zap.Error(err))
			continue
		}
		if isAvailable {
			coupon.IsAvailable = true
			coupon.Sort.IsAvailable = true
		}
	}

	// 积分抵扣后，订单应收为0时，则全部优惠券不可选，置灰
	if saleOrderAmount <= 0 {
		for _, coupon := range coupons {
			coupon.IsAvailable = false
		}
	}

	// 如果已经产生了支付，则全部优惠券不可选，置灰
	if hasPay {
		for _, coupon := range coupons {
			coupon.IsAvailable = false
		}
	}

	// 如果sale_bill已经使用通用优惠券（所有人可用），则通用优惠券不可选，置灰
	if hasCommonCoupon {
		for _, coupon := range coupons {
			if coupon.Requirement == constant.CouponRequirementNone {
				coupon.IsAvailable = false
			}
		}
	}

	// 重新排序
	commonCouponArray := make([]resp.Coupon, 0)
	memberCouponArray := make([]resp.Coupon, 0)
	for _, coupon := range coupons {
		if coupon.Requirement == constant.CouponRequirementNone {
			commonCouponArray = append(commonCouponArray, *coupon)
		} else if coupon.Requirement == constant.CouponRequirementMember {
			memberCouponArray = append(memberCouponArray, *coupon)
		}
	}
	couponList := s.SortCouponList(&resp.CouponList{List: append(commonCouponArray, memberCouponArray...)})

	return couponList, nil
}

// SortCouponList 排序优惠券列表
// 规则1. 按是否在可用时间段内分组：可用、不可用，可用在前。 可用：指优惠券在有效期内，且在使用时段内
// 规则2. 按是否为通用优惠券分组：通用、会员，通用在前。 通用：指优惠券的requirement为none
// 规则3. 同一组内按有效期排序：先到期的在前
// 规则4. 到期时间相同的，按sort排序
// 规则5. sort相同的，按CouponUuid排序（创建时间排序）
func (s *orderSrv) SortCouponList(couponList *resp.CouponList) *resp.CouponList {
	// 可用组
	availableCouponList := make([]resp.Coupon, 0)
	// 不可用组
	unavailableCouponList := make([]resp.Coupon, 0)
	for _, coupon := range couponList.List {
		if coupon.Sort.IsAvailable {
			availableCouponList = append(availableCouponList, coupon)
		} else {
			unavailableCouponList = append(unavailableCouponList, coupon)
		}
	}

	// 规则2. 按是否为通用优惠券分组：通用、会员，通用在前。 通用：指优惠券的requirement为none
	filterFn := func(couponList []resp.Coupon) ([]resp.Coupon, []resp.Coupon) {
		commonCouponList := make([]resp.Coupon, 0)
		memberCouponList := make([]resp.Coupon, 0)
		for _, coupon := range couponList {
			if coupon.Requirement == constant.CouponRequirementNone {
				commonCouponList = append(commonCouponList, coupon)
			} else if coupon.Requirement == constant.CouponRequirementMember {
				memberCouponList = append(memberCouponList, coupon)
			}
		}
		return commonCouponList, memberCouponList
	}

	// 排序函数. 先规格3，再规格4，再规格5
	sortFn := func(couponList []resp.Coupon) []resp.Coupon {
		// 对"可用组中的通用优惠券组"进行排序。先按ValidEndTimestamp时间戳排序，小的在前；再对ValidEndTimestamp相同的情况，按Sort排序；再对Sort相同的情况，按CouponUuid排序
		sort.Slice(couponList, func(i, j int) bool {
			if couponList[i].Sort.ValidEndTimestamp == couponList[j].Sort.ValidEndTimestamp {
				if couponList[i].Sort.Sort == couponList[j].Sort.Sort {
					return couponList[i].Sort.CouponUuid < couponList[j].Sort.CouponUuid
				}
				return couponList[i].Sort.Sort < couponList[j].Sort.Sort
			}
			return couponList[i].Sort.ValidEndTimestamp < couponList[j].Sort.ValidEndTimestamp
		})
		return couponList
	}

	// 对可用组进行规格2分组
	// commonAvailableCouponList 可用组中的通用优惠券组
	// memberAvailableCouponList 可用组中的会员优惠券组
	commonAvailableCouponList, memberAvailableCouponList := filterFn(availableCouponList)

	commonAvailableCouponList = sortFn(commonAvailableCouponList)
	memberAvailableCouponList = sortFn(memberAvailableCouponList)

	// 对不可用组进行规格2分组
	// commonUnavailableCouponList 不可用组中的通用优惠券组
	// memberUnavailableCouponList 不可用组中的会员优惠券组
	commonUnavailableCouponList, memberUnavailableCouponList := filterFn(unavailableCouponList)

	commonUnavailableCouponList = sortFn(commonUnavailableCouponList)
	memberUnavailableCouponList = sortFn(memberUnavailableCouponList)

	list := append(commonAvailableCouponList, memberAvailableCouponList...)
	list = append(list, commonUnavailableCouponList...)
	list = append(list, memberUnavailableCouponList...)
	return &resp.CouponList{
		List: list,
	}
}

// InstantOrderPaymentInfo 获取结账页面信息
func (s *orderSrv) InstantOrderPaymentInfo(ctx context.Context, saleBill *model.SaleBill, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error) {
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	// 获取销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	if saleBill == nil {
		var errSaleBill error
		saleBill, errSaleBill = repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
		if errSaleBill != nil {
			return nil, errSaleBill
		}
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
			Uuid:          saleOrder.ConsumerUuid,
			Nickname:      saleOrder.GetMemberName(),
			Card:          resp.CardInfo{Name: saleOrder.Member.GetMemberCardName()},
			Level:         resp.LevelInfo{Name: saleOrder.Member.GetMemberLevelName()},
			Balance:       saleOrder.Member.GetBalanceAll(),
			Points:        saleOrder.Member.GetPoints(),
			RechargeMoney: saleOrder.Member.GetRechargeMoney(),
		}
	}
	selectedCouponUuid := saleOrder.GetSelectedCouponUuid()
	// saleBill 是否使用了通用优惠券
	hasCommonCoupon, selectedCouponSaleOrderUuid := saleBill.IsCommonCouponUsed()
	// 如果使用通用优惠券的销售订单是当前订单，则通用优惠券可选，可以切换或取消通用优惠券。
	if selectedCouponSaleOrderUuid == saleOrderUuid {
		hasCommonCoupon = false
	}
	couponList, err := s.GetValidMemberCouponList(ctx, saleOrder.ConsumerUuid, selectedCouponUuid, len(saleOrder.PaymentOrders) > 0, hasCommonCoupon, saleOrder.GetPointsExchangeAmount())
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员优惠券列表失败")
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

	var pointsExchange resp.PointsExchangeInfo
	if saleOrder.Member != nil && saleBill.SaleBillSetting.IsOpenPointsExchange() {
		// 积分抵扣信息。
		maxPoints := saleOrder.CaclMaxPoints()

		// 如果自动抵扣积分，且未创建付款单，则更新销售订单的抵扣积分和抵扣金额
		if saleBill.SaleBillSetting.IsOpenPointsExchange() && saleOrder.AutoPointsExchange == 1 && len(saleOrder.PaymentOrders) == 0 {
			// 自动抵扣积分，更新销售订单的抵扣积分和抵扣金额
			saleOrder.PayPoints = maxPoints
			saleOrder.PayPointsAmount = saleOrder.CaclPointsExchangeAmount()

			// 更新销售订单的积分抵扣信息
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderPointsExchange(saleOrder.Uuid, saleOrder.PayPoints, saleOrder.PayPointsAmount, saleOrder.PointsExchangeRate, 1); err != nil {
				return nil, errors.WithMessage(err)
			}

			// 自动抵扣积分,取消所有优惠券
			saleOrder.SetPointsCouponCancel()
			if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
				return nil, errors.WithMessage(err, "取消销售订单所有优惠券失败")
			}
		}
		canChangePoints := true
		if saleBill.SaleBillSetting.IsOpenPointsExchange() && len(saleOrder.PaymentOrders) > 0 {
			// 已创建付款单，则不能修改抵扣积分
			canChangePoints = false
		}
		pointsExchange = resp.PointsExchangeInfo{
			MaxPoints:          maxPoints,
			PointsExchangeRate: saleOrder.PointsExchangeRate,
			PayPoints:          saleOrder.PayPoints, // 手动抵扣积分或已经生效的自动抵扣积分
			PayPointsAmount:    saleOrder.PayPointsAmount,
			OpenPointsExchange: saleBill.SaleBillSetting.IsOpenPointsExchange(),
			CanChangePoints:    canChangePoints,
		}
	}

	methodItems := make([]resp.PaymentMethodItem, 0)
	amounts := make([]resp.PaymentMethodAmount, 0)

	paymentApp, paymentAppErr := saas.NewPaymentAppRepo(s.dbm.GetDB(0)).GetPaymentAppCompanyUuid(ctx.GetCompanyUuid())
	for _, paymentMethod := range paymentMethods {
		// 不显示免单
		if paymentMethod.Code == constant.PaymentMethodCodeFreePay {
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
		if logoUrl == "" && paymentMethod.DefaultImg != "" {
			logoUrl = strings.TrimRight(baseUrl, "/") + paymentMethod.DefaultImg
		}
		if paymentMethod.QrcodeFile != nil {
			qrcodeUrl = paymentMethod.QrcodeFile.GetUrl(baseUrl)
		}
		methodItem := resp.PaymentMethodItem{
			Source:        paymentMethod.Source,
			SourceText:    paymentMethod.GetSourceText(ctx.GetLanguage()),
			Uuid:          paymentMethod.Uuid,
			PaymentName:   paymentMethod.GetPaymentName(),
			PaymentMethod: paymentMethod.GetName(),
			FeePercent:    paymentMethod.FeePercent,
			Logo:          logoUrl,
			Qrcode:        qrcodeUrl,
			Code:          paymentMethod.Code,
		}
		methodItems = append(methodItems, methodItem)

		commissionFee := saleOrder.CalcCommissionFee()

		saleOrderAmount := saleOrder.GetAmountValue() // 积分抵扣后的应收金额
		saleOrderOriginAmount := saleOrder.GetOriginAmountValue()
		if commissionFee > 0 {
			// 如果有手续费
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrderOriginAmount,
				SaleOrderCartAmount:   saleOrder.GetAmount(),
				SaleOrderAmount:       saleOrderAmount,
				CommissionFee:         commissionFee,
				CouponExchangeAmount:  saleOrder.CalcCouponExchangeAmount(),
				UnpaidAmount:          saleOrder.CalcUnPayAmount(true),
				ZeroAmount:            0, // 只有没有手续费时才会抹零
				ZeroRule:              constant.SaleBillSettingCheckoutZeroingMethodNone,
				PaymentMethodUuid:     methodItem.Uuid,
				Code:                  methodItem.Code,
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
			unpaidAmount := saleOrder.CalcCouponExchangeAmount()
			amount := resp.PaymentMethodAmount{
				SaleOrderOriginAmount: saleOrderOriginAmount,
				SaleOrderCartAmount:   saleOrder.GetAmount(),
				SaleOrderAmount:       saleOrderAmount,
				CommissionFee:         commissionFee,
				CouponExchangeAmount:  unpaidAmount,
				UnpaidAmount:          saleOrder.CalcUnPayAmount(hasCommission),
				ZeroAmount:            zeroFee, // 只有没有手续费时且支付方式不需要手续费才会抹零
				IsAutoZero:            saleOrder.IsAutoCheckoutZeroDiscount(*saleBill.SaleBillSetting),
				ZeroRule:              saleOrder.ZeroCheckoutRule,
				PaymentMethodUuid:     methodItem.Uuid,
				Code:                  methodItem.Code,
			}
			amounts = append(amounts, amount)
		}
	}

	infoResp := &resp.InstantOrderPaymentInfoResp{
		MemberInfo:     memberInfo,
		CouponList:     couponList,
		PaymentOrders:  resp.PaymentInfoList{List: paymentOrders},
		PaymentMethods: resp.PaymentMethodList{List: methodItems},
		Amounts:        resp.PaymentMethodAmountList{List: amounts},
		PointsExchange: pointsExchange,
	}

	return infoResp, nil
}

type ProductInfo struct {
	ProductName dto.LocaleResponse
}

// 通过itemCode获取订单中库存不足的商品名称
func (s *orderSrv) GetProductNameByItemCode(ctx context.Context, itemCode string, saleOrderUuid uint64) ([]ProductInfo, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取物品
	material, err := repository.NewMaterialRepo(db).GetMaterialByErpCode(itemCode)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取成本卡uuid列表
	productBomCardUuids, err := repository.NewRelatedMaterialRepo(db).GetProductBomCardUuidsByMaterialUuid(material.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 通过成本卡uuid列表获取product_bom列表
	productBomUuids, err := repository.NewProductBomRepo(db).GetFlavorProductBomUuidsByCardUuids(productBomCardUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 通过product_bom uuid列表查询sale_order_product_bom表获取sale_order_product_uuid列表
	saleOrderProductUuids, err := repository.NewSaleOrderProductRepo(db).GetSaleOrderProductUuidsByProductBomUuids(saleOrderUuid, productBomUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取商品信息
	products, err := repository.NewSaleOrderProductRepo(db).GetSaleOrderProductBySaleOrderProductUuids(saleOrderProductUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	productInfos := make([]ProductInfo, 0)
	for _, product := range products {
		localeName := product.GetNameAndFlavorName()
		// 如果商品包没有多语言名称，则查询数据库获取
		if product.MultiLanguageName.IsNullName() {
			uuid := product.MultiLanguageNameUuid
			names, err := repository.NewMultiLanguageNameRepositoryImpl(db).GetMultiLanguageNameByUuid(uuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			bom, _ := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.SaleOrderProductBoms[0].ProductBomUuid)
			localeName = product.GetNameAndFlavorNameFrom(bom, &names)
		}
		productInfos = append(productInfos, ProductInfo{
			ProductName: localeName,
		})
	}

	return productInfos, nil
}

// OrderPaymentCoupon 使用优惠券 或 取消优惠券
func (s *orderSrv) OrderPaymentCoupon(ctx context.Context, req req.InstantOrderPaymentCouponReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}

	// 获取优惠券信息，判断该优惠券是否是属于该会员的
	couponOriginAmount := 0.0
	if req.CouponRequirement == constant.CouponRequirementNone {
		marketingCoupon, err := repository.NewMarketingCouponRepo(db).GetCouponByUuid(req.CouponUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		couponOriginAmount = marketingCoupon.Amount
	} else if req.CouponRequirement == constant.CouponRequirementMember {
		memberUuid, errSaleOrder := repository.NewSaleOrderRepo(db).GetSaleOrderMemberUuid(req.SaleOrderUuid)
		if errSaleOrder != nil {
			return nil, errors.WithMessage(errSaleBill)
		}
		memberCoupon, errMemberCoupon := repository.NewMemberCouponRepo(db).GetMemberCouponByUuid(req.CouponUuid)
		if errMemberCoupon != nil {
			return nil, errors.WithMessage(errMemberCoupon)
		}
		if memberCoupon.MemberUuid != memberUuid {
			return nil, errors.WithMessage(errors.New("优惠券不属于该会员"))
		}
		couponOriginAmount = memberCoupon.Amount
	}

	hasCoupon := saleOrder.HasCoupon()
	// 判断该销售订单是否已经使用了优惠券，一个订单只能使用一个优惠券
	// 如果使用了优惠券，
	if hasCoupon {
		if saleOrder.HasCouponByUuid(req.CouponUuid, req.CouponRequirement) {
			// 判断是否是同一个优惠券，如果是，删除该优惠券使用记录，表示取消选择
			coupon := saleOrder.GetCouponByUuid(req.CouponUuid, req.CouponRequirement)
			coupon.SetDelete()
		} else {
			// 否则则修改sale_order_coupon表中的记录为新选择的优惠券
			coupon := saleOrder.Coupons[0]
			coupon.ReplaceCoupon(req.CouponUuid, req.CouponRequirement, couponOriginAmount)
		}
	} else {
		// 如果未使用优惠券，则在sale_order_coupon表中增加一条记录
		var couponAmount float64
		if req.CouponRequirement == constant.CouponRequirementMember {
			memberCoupon, err := repository.NewMemberCouponRepo(db).GetMemberCouponByUuid(req.CouponUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			couponAmount = memberCoupon.Amount
		} else if req.CouponRequirement == constant.CouponRequirementNone {
			marketingCoupon, err := repository.NewMarketingCouponRepo(db).GetCouponByUuid(req.CouponUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			couponAmount = marketingCoupon.Amount
		}

		saleOrder.AddCoupon(req.CouponUuid, req.CouponRequirement, couponAmount)
	}

	db.Transaction(func(tx *gorm.DB) error {
		// 选择优惠券后，将积分自动抵扣失效改为手动抵扣
		saleOrder.AutoPointsExchange = 0

		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		if hasCoupon {
			// 更新优惠券使用记录: 更换应用于订单的优惠券 或 软删除优惠券
			if err := repository.NewSaleOrderCouponRepo(tx).UpdateSaleOrderCoupon(*saleOrder.Coupons[0]); err != nil {
				return err
			}
		} else {
			// 新增优惠券使用记录
			if err := repository.NewSaleOrderCouponRepo(tx).CreateSaleOrderCoupon(*saleOrder.Coupons[0]); err != nil {
				return err
			}
		}
		return nil
	})

	// 获取支付结账页面信息
	return s.InstantOrderPaymentInfo(ctx, saleBill, req.SaleBillUuid, req.SaleOrderUuid)
}

func (s *orderSrv) OrderPaymentPoints(ctx context.Context, req req.InstantOrderPaymentPointsReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}

	if !saleBill.SaleBillSetting.IsOpenPointsExchange() {
		return nil, errors.New("未开启积分抵扣功能")
	}

	if saleOrder.Member == nil {
		return nil, errors.New("订单没有会员")
	}
	if len(saleOrder.PaymentOrders) > 0 {
		return nil, errors.New("订单已付款，无法修改积分抵扣数量")
	}

	if req.Points > saleOrder.Member.GetPoints() {
		return nil, errors.New("会员可用积分不足")
	}

	// 检查积分数量是否超过最大抵扣数
	if saleOrder.Member != nil && saleBill.SaleBillSetting.IsOpenPointsExchange() {
		maxPoints := saleOrder.CaclMaxPoints()
		if req.Points > maxPoints {
			return nil, errors.New("积分数量超过最大抵扣数")
		}

		// 如果未创建付款单，则更新销售订单的抵扣积分和抵扣金额
		if len(saleOrder.PaymentOrders) == 0 {
			// 手动抵扣积分，更新销售订单的抵扣积分和抵扣金额
			saleOrder.PayPoints = req.Points
			saleOrder.AutoPointsExchange = 0
			saleOrder.PayPointsAmount = saleOrder.CaclPointsExchangeAmount()

			if err := db.Transaction(func(tx *gorm.DB) error {
				saleOrder.SetCheckoutZeroRuleCancel() // 取消抹零，修改saleBill中的数据
				if err := repository.NewSaleOrderRepo(tx).SetCheckoutZeroRuleCancel(saleOrder.Uuid); err != nil {
					return errors.WithMessage(err)
				}
				// 取消所有优惠券
				saleOrder.SetPointsCouponCancel()
				if err := repository.NewSaleOrderCouponRepo(tx).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
					return errors.WithMessage(err, "取消销售订单所有优惠券失败")
				}
				// 更新销售订单的积分抵扣信息
				if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderPointsExchange(saleOrder.Uuid, saleOrder.PayPoints, saleOrder.PayPointsAmount, saleOrder.PointsExchangeRate, 0); err != nil {
					return errors.WithMessage(err)
				}
				return nil
			}); err != nil {
				return nil, errors.WithMessage(err)
			}
		}
	}

	// 获取订单的付款信息
	infoResp, err := s.InstantOrderPaymentInfo(ctx, saleBill, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return infoResp, nil
}

func (s *orderSrv) InstantOrderPaymentQrcode(ctx context.Context, req req.InstantOrderPaymentQrcodeReq) (*resp.InstantOrderPaymentQrcodeInfoResp, error) {
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

	// 验证订单是否可操作
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	if saleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.New("销售账单已结束"))
	}

	// 获取销售订单
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}
	if err := saleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断当前是否连连支付
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethod := paymentMethodRepo.GetPaymentMethod(paymentMethodRepo.WhereUuid(req.PaymentMethodUuid))
	if paymentMethod.Uuid == 0 {
		return nil, errors.New("支付方式不存在")
	}
	// 支付方式是否可用
	if !paymentMethod.IsLianLianPay() {
		return nil, errors.New("支付方式不可用")
	}

	// 判断支付方式是否已支付
	orderRepo := repository.NewPaymentOrderRepo(db)
	paymentOrder, err := orderRepo.GetPaymentOrderInfo(
		repository.CommonRepo.WhereBySoftDelete(),
		orderRepo.WhereRelatedUuid(saleOrder.Uuid),
		orderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
		orderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
	)
	if err == nil {
		if paymentOrder.Status == constant.PaymentOrderStatusPaid {
			infoResp := &resp.InstantOrderPaymentQrcodeInfoResp{
				PaymentOrderUuid: paymentOrder.Uuid,
				QrCode:           "",
				QrCodeExpireSec:  10000,
				Status:           paymentOrder.Status,
				PaymentAmount:    paymentOrder.PaymentAmount,
			}
			return infoResp, nil
		}
	}

	// 计算手续费
	commissionFee := paymentMethod.CalculatePaymentCommissionFee(req.PaymentAmount)
	paymentAmount := paymentMethod.CalculatePaymentAmount(req.PaymentAmount)
	if commissionFee > 0 {
		saleOrder.SetCheckoutZeroRuleCancel()
	}
	unpaidAmount := saleOrder.GetUnpaidAmount()

	// 判断支付金额是否大于未收金额.只能现金支付大于未收金额
	if unpaidAmount < req.PaymentAmount {
		return nil, errors.WithMessage(errors.New(fmt.Sprintf("支付金额不能大于未收金额 %.2f", unpaidAmount)))
	}

	// 创建连连支付订单
	payment, err := NewPaymentRepo(ctx, s.dbm).CreatePayment(CreatePaymentReq{
		RelatedType:       constant.PaymentOrderRelatedTypeSaleOrder,
		RelatedUuid:       saleOrder.Uuid,
		PaymentMethodUuid: paymentMethod.Uuid,
		PaymentMethodCode: paymentMethod.Code,
		PaymentAmount:     decimal.NewFromFloat(paymentAmount).Add(decimal.NewFromFloat(commissionFee)).InexactFloat64(),
		CommissionFee:     commissionFee,
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 在 infoResp 初始化之前添加
	infoResp := &resp.InstantOrderPaymentQrcodeInfoResp{
		PaymentOrderUuid: payment.PaymentOrderUuid,
		QrCode:           payment.LinkUrl,
		QrCodeExpireSec:  payment.GetRemainingPayableTime(),
		Status:           payment.GetStatus(),
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
	// if !saleBill.IsCookingStatus() {
	// 	return nil, errors.WithMessage(errors.New("订单没有商品，请选购商品"))
	// }
	// 判断销售订单是否可操作
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("无法查询到销售订单"))
	}

	paymentMethod, err := repository.NewPaymentMethodRepo(db).GetPaymentMethodByUuid(req.PaymentMethodUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("支付方式未开启"))
	}

	// 支付方式是否可用
	if ctx.GetSource() == jwt.SourceAssistant {
		if paymentMethod.IsShowAssistant == 0 {
			return nil, errors.WithMessage(errors.New("支付方式未开启"))
		}
	} else if ctx.GetSource() == jwt.SourceCashier {
		if paymentMethod.IsShowCashier == 0 {
			return nil, errors.WithMessage(errors.New("支付方式未开启"))
		}
	}
	if !s.paymentMethodSrv.IsEnabled(ctx, *paymentMethod, ctx.GetCompanySetting()) {
		return nil, errors.WithMessage(errors.New("支付方式未开启"))
	}

	// 默认支付订单状态
	paymentOrderStatus := constant.PaymentOrderStatusPaid

	// 获取支付订单
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)

	if paymentMethod.IsBalance() {
		// 检查会员余额是否充足
		if saleOrder.Member == nil {
			return nil, errors.New("会员不存在")
		}
		if saleOrder.Member.GetBalanceAll() < req.PaymentAmount {
			return nil, errors.New("会员余额不足")
		}
	}

	//  在线支付订单
	if paymentMethod.IsLianLianPay() {
		paymentOrder, _ := paymentOrderRepo.GetPaymentOrder(
			paymentOrderRepo.WhereRelatedUuid(saleOrder.Uuid),
			paymentOrderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
		)
		// 如果已经存在直接返回
		if paymentOrder.Uuid != 0 {
			if paymentOrder.PaymentAmount != req.PaymentAmount {
				return nil, errors.New("不能重复支付")
			}
			newInfoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			return newInfoResp, nil
		}
	} else if req.PaymentOrderUuid != 0 {
		return nil, errors.New("非在线支付无需传支付订单ID")
	}

	// 非在线支付订单
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "添加支付订单-获取货币设置失败")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	commissionAmount := infoResp.GetCommissionAmount()

	// 判断支付金额是否大于未收金额.只能现金支付大于未收金额
	unpaidAmount := infoResp.GetUnpaidAmount(paymentMethod.Uuid)
	if unpaidAmount < req.PaymentAmount {
		if paymentMethod.Code != constant.PaymentMethodCodeCash {
			return nil, errors.WithMessage(errors.New(fmt.Sprintf("支付金额不能大于未收金额 %.2f", unpaidAmount)))
		}
	}

	percent := paymentMethod.GetFeePercent()
	commissionFee := paymentMethod.CalculatePaymentCommissionFee(req.PaymentAmount)
	amount := paymentMethod.CalculatePaymentAmount(req.PaymentAmount)
	paymentOrder := &model.PaymentOrder{
		BaseModel:            model.BaseModel{Uuid: req.PaymentOrderUuid},
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
		Status:               paymentOrderStatus,
	}

	// 判断这个支付方式是否已经支付过，如果已经支付过，则更新支付单
	paymentOrderList, err := paymentOrderRepo.GetPaymentOrderListBySaleOrderUuid(req.SaleOrderUuid)
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
			s.bus.PublishCheckoutZeroSaleOrderEvent(event.CheckoutZeroSaleOrderPayload{
				BasePayload: event.BasePayload{ // 自动抹零
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  req.SaleBillUuid,
					SaleOrderUuid: req.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				Operation: constant.OrderCheckoutDiscountCancel,
				Reason:    "选择含手续费的支付方式",
			})
		}()
	}

	newInfoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
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
	// 当不是收银端的时候，拆单不可操作结账
	if ctx.GetSource() != constant.SourceCashier && saleBill.IsSplit() {
		return errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}
	// 判断销售账单是否结束
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
	paymentOrder.SetNil()
	// 更新支付单
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*paymentOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}
	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
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
		if saleOrderProduct.IsUnAcceptOrderBool() {
			continue
		}
		if saleOrderProduct.IsCancelProduct() {
			continue
		}
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

// FinishSaleBill 完成销售账单
func (s *orderSrv) FinishSaleBill(ctx context.Context, saleBill *model.SaleBill, businessSetting settingResp.Business, db *gorm.DB) error {
	// 更新销售账单
	updateSaleBill := false
	if saleBill.CanFinishSaleBill() {
		staff := ctx.GetStaff()
		saleBill.SetFinishSaleBill(staff.DutyNo, staff.Uuid, staff.GetUserName())
		saleBill.CalcAll()
		updateSaleBill = true
	}
	if updateSaleBill {
		// 更新销售账单
		saleBill.SetCookingStatus()
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
		// 拒绝所有待接单的h5订单
		if err := s.RejectAllH5Order(ctx, saleBill.Uuid); err != nil {
			return err
		}
	}
	// 更新自助餐销量
	if saleBill.IsFinish() && saleBill.IsBuffetSaleBill() {
		if saleBill.BuffetPackage1Uuid != 0 {
			saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage1Uuid)
			if err := repository.NewBuffetRepo(db).AddActualSaleNum(saleBill.BuffetPackage1Uuid, saleNum); err != nil {
				ctx.Log().Error("AddActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
			}
		}
		if saleBill.BuffetPackage2Uuid != 0 {
			saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage2Uuid)
			if err := repository.NewBuffetRepo(db).AddActualSaleNum(saleBill.BuffetPackage2Uuid, saleNum); err != nil {
				ctx.Log().Error("AddActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
			}
		}
	}
	//
	return nil
}

// 获取订单的积分发放规格信息
func (s *orderSrv) GetPointsRuleInfo(ctx context.Context, isBufferOrder bool, memberLevelUuid uint64) (*settingResp.PointsRule, error) {
	// 获取积分设置
	pointsSetting, err := s.settingSrv.GetPointsSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	rule := pointsSetting.GetPointsGiftRule(isBufferOrder, memberLevelUuid)
	return &rule, nil
}

// 核销优惠券
func (s *orderSrv) VerifyCoupon(ctx context.Context, saleOrder *model.SaleOrder, db *gorm.DB) error {
	if !saleOrder.HasCoupon() {
		return nil
	}
	coupon := saleOrder.Coupons[0] // 目前一个订单只支持一个优惠券

	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	timezone := utils.SetTimezone(storeSetting.TimeZone)
	nowTime := timezone.Now().Format("15:04")
	// 使用营销会员优惠券时
	if coupon.CouponRequirement == constant.CouponRequirementMember {
		if saleOrder.ConsumerUuid == 0 {
			return errors.New("订单未使用会员，无法核销营销会员优惠券")
		}
		memberCoupon, errMemberCoupon := repository.NewMemberCouponRepo(db).GetMemberCouponByUuid(coupon.MemberCouponUuid)
		if errMemberCoupon != nil {
			return errors.WithMessage(errMemberCoupon)
		}
		// 判断优惠券是否可用
		if err := memberCoupon.IsAvailable(saleOrder.ConsumerUuid, nowTime); err != nil {
			return errors.WithMessage(err)
		}
		// 核销优惠券
		if err := repository.NewMemberCouponRepo(db).VerifyMemberCoupon(coupon.MemberCouponUuid); err != nil {
			return errors.WithMessage(err)
		}
		// 添加会员优惠券使用记录
		if err := repository.NewMemberCouponRepo(db).CreateMemberCouponRecord(model.MemberCouponUseRecord{
			MemberUuid:     saleOrder.ConsumerUuid,
			CouponUuid:     coupon.MemberCouponUuid,
			UseOrderUuid:   saleOrder.Uuid,
			UseOrderAmount: saleOrder.CouponAmount,
		}); err != nil {
			return errors.WithMessage(err)
		}
		// 添加会员优惠券使用记录
		if err := repository.NewMarketingCouponRepo(db).CreateMarketingCouponRecord(model.MarketingCouponRecord{
			BaseModel:    model.BaseModel{},
			CouponUuid:   coupon.MemberCouponUuid,
			SerialNo:     "",
			ActivityUuid: 0,
			MemberUuid:   saleOrder.ConsumerUuid,
			Type:         constant.CouponRecordTypeUsed,
			Count:        0,
			LeftCount:    0,
		}); err != nil {
			return errors.WithMessage(err)
		}
		// 添加营销会员优惠券使用记录
		// 查询营销优惠券的剩余数量
		marketingCouponUuid := coupon.MemberCoupon.CouponUuid
		marketingCouponLeftCount, errMarketingCouponLeftCount := repository.NewMarketingCouponRepo(db).GetCouponLeftCount(marketingCouponUuid)
		if errMarketingCouponLeftCount != nil {
			return errors.WithMessage(errMarketingCouponLeftCount)
		}
		if err := repository.NewMarketingCouponRepo(db).CreateMemberCouponRecord(marketingCouponUuid, saleOrder.ConsumerUuid, marketingCouponLeftCount); err != nil {
			return errors.WithMessage(err)
		}
	}

	// 使用通用优惠券（所有人可用）时
	if coupon.CouponRequirement == constant.CouponRequirementNone {
		commonCoupon, err := repository.NewMarketingCouponRepo(db).GetCouponByUuid(coupon.MarketingCouponUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if err := commonCoupon.IsAvailable(nowTime); err != nil {
			return errors.WithMessage(err)
		}
		commonCoupon.Count = commonCoupon.Count - 1
		if err := repository.NewMarketingCouponRepo(db).CreateCommonCouponRecord(commonCoupon.Uuid, commonCoupon.Count); err != nil {
			return errors.WithMessage(err, "创建通用优惠券使用记录失败")
		}
		if err := repository.NewMarketingCouponRepo(db).UpdateCommonCouponCount(commonCoupon.Uuid); err != nil {
			return errors.WithMessage(err, "减1通用优惠券数量失败")
		}
	}
	return nil
}

// 保存发票到erp
func (s *orderSrv) SavePosInvoice(ctx context.Context, saleOrder *model.SaleOrder, saleBill *model.SaleBill, db *gorm.DB) (*selling.SavePosInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	staff := ctx.GetStaff()
	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLog, err := shiftLogRepo.GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if shiftLog.IsHandedOver() {
		return nil, errors.New("当前班次已交班，无法保存发票")
	}

	// 订单商品列表
	items := make([]*selling.PosInvoiceItem, 0)
	isFreeOrder := saleOrder.IsFreeSaleOrder()
	for _, product := range saleOrder.SaleOrderBuffetCustomerTypes {
		// 自助餐名称
		buffetName := product.BuffetPackage.MultiLanguageName.EnName
		items = append(items, &selling.PosInvoiceItem{
			ItemCode:    "ZZC001",
			Qty:         float64(product.Num),
			Rate:        product.GetFinalSalePriceNoneTax(),                                                                                                             // 商品未含税价格（折后）
			Amount:      decimal.NewFromFloat(product.GetFinalSalePriceNoneTax()).Mul(decimal.NewFromFloat(float64(product.Num))).Truncate(3).Round(2).InexactFloat64(), // 商品未含税价格（折后）* 数量
			Description: fmt.Sprintf("%s-%s", buffetName, product.Name),
		})
	}
	buffetPackageName := saleBill.GetBuffetPackageName()
	for _, product := range saleOrder.SaleOrderBuffetDelayProducts {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode:    "ZZCJZ001",
			Qty:         float64(product.Num),
			Rate:        product.Price,                                                                                                             // 商品未含税价格（折后）
			Amount:      decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(float64(product.Num))).Truncate(3).Round(2).InexactFloat64(), // 商品未含税价格（折后）* 数量
			Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, product.Name),
		})
	}
	for _, product := range saleOrder.SaleOrderProducts {
		// 如果商品已删除，则跳过
		if product.IsDelete() || product.IsCancelProduct() {
			continue
		}
		if product.IsPackageSubProduct() {
			continue
		}
		if product.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(product.Uuid)
			for _, subProduct := range subProducts {
				productBom := subProduct.GetFlarvorSaleOrderProductBom()
				erpCode := productBom.ProductBom.ErpCode
				packageName := language.JsonToLocaleResponse(product.Name) // 套餐名称
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    erpCode,
					Qty:         subProduct.Num,
					Rate:        0,                                                  // 套餐子商品没有单价
					Amount:      0,                                                  // 套餐子商品没有金额
					Description: fmt.Sprintf("Sales in package:%s", packageName.EN), // 套餐子商品描述
					IsFreeItem:  true,
				})
			}
		}
		productBom := product.GetFlarvorSaleOrderProductBom()
		erpCode := productBom.ProductBom.ErpCode
		// 是否是赠菜
		if product.IsGiftProduct() {
			if product.IsPackageProduct() {
				packageName := language.JsonToLocaleResponse(product.Name) // 套餐名称
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "TC001",
					Qty:         product.Num,
					Rate:        0,              // 商品未含税价格（折后）
					Amount:      0,              // 商品未含税价格（折后）* 数量
					Description: packageName.EN, // 套餐子商品描述
					IsFreeItem:  true,
				})
			} else {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:   erpCode,
					Qty:        product.Num,
					Rate:       0,    // 商品未含税价格（折后）
					Amount:     0,    // 商品未含税价格（折后）* 数量
					IsFreeItem: true, // 赠菜
				})
			}
		} else if product.SalePrice == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:   erpCode,
				Qty:        product.Num,
				Rate:       0,    // 商品未含税价格（折后）
				Amount:     0,    // 商品未含税价格（折后）* 数量
				IsFreeItem: true, // 零元商品当作赠菜
			})
		} else {
			if product.IsPackageProduct() { // 套餐主商品不添加到发票
				packageName := language.JsonToLocaleResponse(product.Name) // 套餐名称
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "TC001",
					Qty:         product.Num,
					Rate:        product.GetFinalSalePriceNoneTax(),        // 商品未含税价格（折后）
					Amount:      product.GetProductFinalSalePriceNoneTax(), // 商品未含税价格（折后）* 数量
					Description: packageName.EN,                            // 套餐子商品描述
					IsFreeItem:  false,
				})
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode:   erpCode,
					Qty:        product.Num,
					Rate:       product.GetFinalSalePriceNoneTax(),        // 商品未含税价格（折后）
					Amount:     product.GetProductFinalSalePriceNoneTax(), // 商品未含税价格（折后）* 数量
					IsFreeItem: isFreeOrder,
				}
				if isFreeOrder {
					item.Rate = 0
					item.Amount = 0
				}
				items = append(items, item)
			}
		}
		// 如果有小料，则需要添加小料
		sauceBoms := product.GetSauceSaleOrderProductBom()
		for _, sauceBom := range sauceBoms {
			erpCode := sauceBom.ProductBom.ProductSauce.ErpCode
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:   erpCode,
				Qty:        product.Num,
				Rate:       0, // 加料没有单价
				Amount:     0, // 加料没有金额
				IsFreeItem: true,
			})
		}
	}
	materialItems := make([]*selling.PosInvoiceItem, 0)
	erpProductBomMaterials := saleOrder.GetErpProductBomMaterials()
	for _, material := range erpProductBomMaterials {
		materialItems = append(materialItems, &selling.PosInvoiceItem{
			ItemCode: material.ErpCode,
			Qty:      material.Num, // 原材料数量
			Uom:      material.Uom, // 原材料单位
			Rate:     0,            // 原材料没有单价
			Amount:   0,            // 原材料没有金额
		})
	}

	taxes := make([]*selling.PosInvoiceTax, 0)
	// Tax 消费税、Service Fee 服务费、Payment Processing Fee 支付手续费、Delivery Fee 配送费
	if saleOrder.TaxFee > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   saleOrder.TaxFee,
			Description: "Tax", // 消费税
		})
	}
	// 如果有服务费，则添加一个虚拟商品来记录服务费
	if saleOrder.ServiceFee > 0 {
		serviceFeeItem := &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeServiceFee,
			Qty:      saleOrder.ServiceFee,
			Rate:     1,
			Amount:   saleOrder.ServiceFee,
		}
		// 如果是免单
		if isFreeOrder {
			serviceFeeItem.IsFreeItem = true
		}
		items = append(items, serviceFeeItem)
	}
	// 如果有支付手续费，则添加一个虚拟商品来记录支付手续费
	if saleOrder.PaymentCommissionFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee,
			Qty:      saleOrder.PaymentCommissionFee,
			Rate:     1,
			Amount:   saleOrder.PaymentCommissionFee,
		})
		// taxes = append(taxes, &selling.PosInvoiceTax{
		// 	TaxAmount:   saleOrder.PaymentCommissionFee,
		// 	Description: "Payment Processing Fee", // 支付手续费
		// })
	}

	// 整单改价 - Whole Order Price Adjustment
	// 优惠折扣抹零 - Discount Rounding Off
	// 结账抹零 - Checkout Rounding Off
	// 优惠券抵扣 - Coupon Deduction
	// 积分抵扣 - Points Deduction
	// 如果有订单应收优惠的话，赠加一个taxes元素
	erpDiscountAmount := saleOrder.GetErpCustomAmount()
	if erpDiscountAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -erpDiscountAmount,
			Description: "Whole Order Price Adjustment", // 订单应收优惠
		})
		// 更新saleOrder的erp_discount_amount字段
		if err := repository.NewOrderRepo(db).UpdateErpDiscountAmount(saleOrder.Uuid, erpDiscountAmount); err != nil {
			return nil, errors.WithMessage(err)
		}
	}
	if saleOrder.ZeroFee != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.ZeroFee,
			Description: "Discount Rounding Off", // 优惠折扣抹零
		})
	}
	if saleOrder.ZeroCheckoutFee != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.ZeroCheckoutFee,
			Description: "Checkout Rounding Off", // 结账抹零
		})
	}
	if saleOrder.CouponAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.CouponAmount,
			Description: "Coupon Deduction", // 优惠券抵扣
		})
	}
	if saleOrder.PayPointsAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.PayPointsAmount,
			Description: "Points Deduction", // 积分抵扣
		})
	}

	// 免单订单不收税费、服务费、支付手续费
	if isFreeOrder {
		taxes = make([]*selling.PosInvoiceTax, 0)
	}

	payments := make([]*selling.PosInvoicePayment, 0)
	if saleOrder.IsFreeSaleOrder() {
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: "Free Meal", // 免单
			Amount:        0,
		})
	} else if saleOrder.GetAmountValue() == 0 { // 如果订单应收为0元时
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: "Cash", // 现金支付
			Amount:        0,
		})
	} else {
		// 获取所有支付方式
		paymentMethodRepo := repository.NewPaymentMethodRepo(db)
		paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
		methodMap := make(map[int]string)
		for _, paymentMethod := range paymentMethods {
			if paymentMethod.ErpnextPayment != "" {
				methodMap[paymentMethod.Code] = paymentMethod.ErpnextPayment
			}
		}
		for _, payment := range saleOrder.PaymentOrders {
			if payment.IsDelete() {
				continue
			}
			var modeOfPayment string
			if method, ok := methodMap[payment.PaymentMethod.Code]; ok {
				modeOfPayment = method
			} else {
				// modeOfPayment =  "Cash" // 其他支付方式，默认现金支付
				return nil, errors.WithMessage(errors.New("不支持的支付方式"))
			}
			payments = append(payments, &selling.PosInvoicePayment{
				ModeOfPayment: modeOfPayment,
				Amount:        payment.Amount,
			})
		}
	}

	customerUuid := fmt.Sprintf("%d", saleOrder.ConsumerUuid)
	if customerUuid == "0" {
		customerUuid = ""
	}
	erpSrv := erp.NewIErpSrv(s.dbm)
	param := req.SavePosInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          saleOrder.OrderNo,
		OpenPosEntryName: shiftLog.ErpnextOpenPosEntryName,
		PostingDatetime:  saleOrder.FinishTime,
		CustomerUuid:     customerUuid,
		Items:            items,         // 订单商品列表
		MaterialItems:    materialItems, // 订单原材料列表
		Taxes:            taxes,         // 订单税费列表
		Payments:         payments,      // 订单付款列表
	}
	if saleOrder.IsErpReverseSettle() {
		param.AmendedProductsInvoiceName = saleOrder.ErpProductsInvoiceName
		param.AmendedMaterialInvoiceName = saleOrder.ErpMaterialInvoiceName
	}
	response, err := erpSrv.SavePosInvoice(ctx, param)
	if err != nil {
		erpSrv := erp.NewIErpSrv(s.dbm)
		getPosInvoiceErrorResp, err := erpSrv.ParseSavePosInvoiceError(err)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if getPosInvoiceErrorResp.ErrorScene == constant.ErpItemStockNotEnough {
			return nil, errors.WithMessage(errors.New("物品库存不足," + getPosInvoiceErrorResp.ItemCode))
		}
		return nil, errors.WithMessage(err)
	}
	return response, nil
}

// 退款发票到erp
func (s *orderSrv) ReturnPosInvoice(ctx context.Context, saleOrder *model.SaleOrder, returnOrder *model.ReturnOrder, db *gorm.DB, returnType int, isPartReturn bool) (*selling.ReturnPosInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	staff := ctx.GetStaff()
	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLog, err := shiftLogRepo.GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if shiftLog.IsHandedOver() {
		return nil, errors.New("当前班次已交班，无法保存发票")
	}

	// 订单商品列表
	items := make([]*selling.PosInvoiceItem, 0)
	totalTaxFee := decimal.NewFromFloat(0)
	totalServiceFee := decimal.NewFromFloat(0)
	for _, product := range returnOrder.ReturnOrderProducts {
		if product.ProductType == constant.ReturnOrderProductTypeSaleOrderProduct {
			saleOrderProduct, _, _ := saleOrder.GetSaleOrderProduct(product.SaleOrderProductUuid)
			if saleOrderProduct.IsPackageSubProduct() {
				continue // 跳过子商品，因为在套餐商品中已经录入了子商品
			}
			if isPartReturn && saleOrderProduct.IsGiftProduct() {
				continue // 部分退款时，如果有赠菜，先不传给erp
			}
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					packageName := language.JsonToLocaleResponse(saleOrderProduct.Name) // 套餐名称
					num := decimal.NewFromFloat(product.Num).Mul(decimal.NewFromFloat(subProduct.UnitNum)).Round(3).InexactFloat64()
					items = append(items, &selling.PosInvoiceItem{
						ItemCode:    subProduct.ErpCode,
						Qty:         -num,
						Rate:        0,                                                  // 套餐子商品没有单价
						Amount:      0,                                                  // 套餐子商品没有金额
						Description: fmt.Sprintf("Sales in package:%s", packageName.EN), // 套餐子商品描述
						IsFreeItem:  true,
					})
				}
			}
			taxFee := saleOrderProduct.TaxFee // 商品税费,仅消费税
			// 无论商品是否“已含税”或“未含税”都是要累积税费
			{
				// 累计本次退款操作中退款商品的税费
				tax := decimal.NewFromFloat(taxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()                                   // 仅消费税
				serviceTaxFee := decimal.NewFromFloat(saleOrderProduct.ServiceTaxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64() // 仅服务费税费
				totalTaxFee = totalTaxFee.Add(decimal.NewFromFloat(tax)).Add(decimal.NewFromFloat(serviceTaxFee))                                      // +消费税+服务费税费
				serviceFee := decimal.NewFromFloat(saleOrderProduct.ServiceFee).Mul(decimal.NewFromFloat(product.Num))
				totalServiceFee = totalServiceFee.Add(serviceFee)
			}
			if saleOrderProduct.SalePrice == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:   product.ErpCode,
					Qty:        -product.Num,
					Rate:       0,    // 商品未含税价格（折后）
					Amount:     0,    // 商品未含税价格（折后）* 数量
					IsFreeItem: true, // 零元商品当作赠菜
				})
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode: product.ErpCode,
					Qty:      -product.Num,
					Rate:     product.GetProductPriceNoneTax(taxFee, saleOrderProduct.HasTax()),        // 商品未含税价格（折后）
					Amount:   -product.GetProductTotalAmountNoneTax(taxFee, saleOrderProduct.HasTax()), // 商品未含税价格（折后）* 数量
				}
				if saleOrderProduct.IsGiftProduct() {
					item.IsFreeItem = true
				}
				items = append(items, item)

			}
		} else if product.ProductType == constant.ReturnOrderProductTypeSaleOrderBuffetCustomer {
			buffetCustomer, _, _ := saleOrder.GetSaleOrderBuffetCustomerType(product.SaleOrderProductUuid)
			// 无论商品是否“已含税”或“未含税”都是要累积税费
			{
				// 累计本次退款操作中退款商品的税费
				taxFee := buffetCustomer.TaxFee                                                                                                      // 商品税费,仅消费税
				tax := decimal.NewFromFloat(taxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()                                 // 仅消费税
				serviceTaxFee := decimal.NewFromFloat(buffetCustomer.ServiceTaxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64() // 仅服务费税费
				totalTaxFee = totalTaxFee.Add(decimal.NewFromFloat(tax)).Add(decimal.NewFromFloat(serviceTaxFee))                                    // +消费税+服务费税费
				serviceFee := decimal.NewFromFloat(buffetCustomer.ServiceFee).Mul(decimal.NewFromFloat(product.Num))
				totalServiceFee = totalServiceFee.Add(serviceFee)
			}
			if buffetCustomer.SalePrice == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
				item := &selling.PosInvoiceItem{
					ItemCode:   "ZZC001",
					Qty:        -product.Num,
					Rate:       0,
					Amount:     0,
					IsFreeItem: true,
				}
				items = append(items, item)
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode: "ZZC001",
					Qty:      -product.Num,
					Rate:     buffetCustomer.GetFinalSalePriceNoneTax(),
					Amount:   -decimal.NewFromFloat(buffetCustomer.GetFinalSalePriceNoneTax()).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
				}
				items = append(items, item)
			}
		} else if product.ProductType == constant.ReturnOrderProductTypeBuffetAddTimeProduct {
			buffetDelayProduct, _, _ := saleOrder.GetSaleOrderBuffetDelayProduct(product.SaleOrderProductUuid)
			if buffetDelayProduct.Price == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
				item := &selling.PosInvoiceItem{
					ItemCode:   "ZZC001",
					Qty:        -product.Num,
					Rate:       0,
					Amount:     0,
					IsFreeItem: true,
				}
				items = append(items, item)
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode: "ZZC001",
					Qty:      -product.Num,
					Rate:     buffetDelayProduct.Price,
					Amount:   -decimal.NewFromFloat(buffetDelayProduct.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
				}
				items = append(items, item)
			}
		}
	}

	taxes := make([]*selling.PosInvoiceTax, 0)
	// Tax 消费税、Service Fee 服务费、Payment Processing Fee 支付手续费、Delivery Fee 配送费
	if totalTaxFee.GreaterThan(decimal.NewFromFloat(0)) {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -totalTaxFee.InexactFloat64(),
			Description: "Tax", // 消费税
		})
	}
	// 如果是部分退款，则需要添加服务费(从各个saleOrderProduct中累计的按比例收取的服务费)
	if returnType == constant.ReturnOrderRefundTypePart {
		if totalServiceFee.GreaterThan(decimal.NewFromFloat(0)) {
			items = append(items, &selling.PosInvoiceItem{
				ItemCode: constant.PosInvoiceItemCodeServiceFee,
				Qty:      -totalServiceFee.InexactFloat64(),
				Rate:     1,
				Amount:   -totalServiceFee.InexactFloat64(),
			})
			// taxes = append(taxes, &selling.PosInvoiceTax{
			// 	TaxAmount:   -totalServiceFee.InexactFloat64(),
			// 	Description: "Service Fee", // 服务费
			// })
		}
	}
	// 如果是整单退款，则退支付手续费、固定服务费
	if returnType == constant.ReturnOrderRefundTypeTotal {
		if saleOrder.PaymentCommissionFee > 0 { // 有支付手续费，才退
			items = append(items, &selling.PosInvoiceItem{
				ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee,
				Qty:      -saleOrder.PaymentCommissionFee,
				Rate:     1,
				Amount:   -saleOrder.PaymentCommissionFee,
			})
			// taxes = append(taxes, &selling.PosInvoiceTax{
			// 	TaxAmount:   -saleOrder.PaymentCommissionFee,
			// 	Description: "Payment Processing Fee", // 支付手续费
			// })
		}
		//如果是固定服务费
		if saleOrder.IsFixedServiceFee() {
			if saleOrder.ServiceFee > 0 { // 有固定服务费，才退
				items = append(items, &selling.PosInvoiceItem{
					ItemCode: constant.PosInvoiceItemCodeServiceFee,
					Qty:      -saleOrder.ServiceFee,
					Rate:     1,
					Amount:   -saleOrder.ServiceFee,
				})
				// taxes = append(taxes, &selling.PosInvoiceTax{
				// 	TaxAmount:   -saleOrder.ServiceFee,
				// 	Description: "Service Fee", // 固定服务费
				// })
			}
		} else {
			// 按比例收取服务费
			if totalServiceFee.GreaterThan(decimal.NewFromFloat(0)) {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode: constant.PosInvoiceItemCodeServiceFee,
					Qty:      -totalServiceFee.InexactFloat64(),
					Rate:     1,
					Amount:   -totalServiceFee.InexactFloat64(),
				})
				// taxes = append(taxes, &selling.PosInvoiceTax{
				// 	TaxAmount:   -totalServiceFee.InexactFloat64(),
				// 	Description: "Service Fee", // 服务费(按比例收取)
				// })
			}
		}

		// 整单改价 - Whole Order Price Adjustment
		// 优惠折扣抹零 - Discount Rounding Off
		// 结账抹零 - Checkout Rounding Off
		// 优惠券抵扣 - Coupon Deduction
		// 积分抵扣 - Points Deduction
		// 如果是整单退款，有订单应收优惠的话，退款时加上订单应收优惠tax项
		if saleOrder.ErpDiscountAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ErpDiscountAmount,    // 该item要是正数表示返还
				Description: "Whole Order Price Adjustment", // 整单改价
			})
		}
		// 如果是整单退款，有整单改价的话，退款时加上整单改价tax项
		if saleOrder.ZeroFee != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ZeroFee,
				Description: "Discount Rounding Off", // 优惠折扣抹零
			})
		}
		if saleOrder.ZeroCheckoutFee != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ZeroCheckoutFee,
				Description: "Checkout Rounding Off", // 结账抹零
			})
		}
		if saleOrder.CouponAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.CouponAmount,
				Description: "Coupon Deduction", // 优惠券抵扣
			})
		}
		if saleOrder.PayPointsAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.PayPointsAmount,
				Description: "Points Deduction", // 积分抵扣
			})
		}
	}

	// 获取所有支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
	methodMap := make(map[int]string)
	for _, paymentMethod := range paymentMethods {
		if paymentMethod.ErpnextPayment != "" {
			methodMap[paymentMethod.Code] = paymentMethod.ErpnextPayment
		}
	}
	payments := make([]*selling.PosInvoicePayment, 0)
	for _, payment := range returnOrder.ReturnOrderAmounts {
		var modeOfPayment string
		if method, ok := methodMap[payment.PaymentMethod.Code]; ok {
			modeOfPayment = method
		} else {
			// modeOfPayment =  "Cash" // 其他支付方式，默认现金支付
			return nil, errors.WithMessage(errors.New("不支持的支付方式"))
		}
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: modeOfPayment,
			Amount:        -payment.Amount,
		})
	}

	erpSrv := erp.NewIErpSrv(s.dbm)
	param := req.ReturnPosInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          saleOrder.OrderNo,
		OpenPosEntryName: shiftLog.ErpnextOpenPosEntryName,
		PostingDatetime:  returnOrder.CreateTime, // 退款单时间
		CompanyAbbr:      companySetting.ErpnextCompanyAbbr,
		InvoiceName:      saleOrder.ErpProductsInvoiceName, // 发票名称
		Items:            items,                            // 订单退款商品列表
		Taxes:            taxes,                            // 订单退税费列表
		Payments:         payments,                         // 订单退款列表
	}
	response, err := erpSrv.ReturnPosInvoice(ctx, param)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return response, nil
}

// InstantOrderPaymentFinish 完成销售订单的付款结账
func (s *orderSrv) InstantOrderPaymentFinish(ctx context.Context, request req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("GetSaleBillAllInfo", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), errSaleBill)))
		return nil, errSaleBill
	}

	// 重新计算销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 当不是收银端的时候，拆单不可操作结账
	if ctx.GetSource() != constant.SourceCashier && saleBill.IsSplit() {
		return nil, errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 如果开启积分抵扣，则检查会员积分是否足够
	if saleBill.SaleBillSetting.IsOpenPointsExchange() {
		if saleOrder.Member != nil {
			if saleOrder.PayPoints > 0 && saleOrder.Member.GetPoints() < saleOrder.PayPoints {
				return nil, errors.New("当前会员抵扣积分不足，需撤销支付后重新抵扣")
			}
		}
	}

	var unpaidAmount float64  // 未付款金额
	var commissionFee float64 // 手续费，付款已经产生的手续费
	// 获取最小的那个未付款金额。因为可能结账抹零后已经没有未付款金额了
	for index, amountItem := range infoResp.Amounts.List {
		if index == 0 {
			unpaidAmount = amountItem.UnpaidAmount
			commissionFee = amountItem.CommissionFee
			continue
		}
		if amountItem.UnpaidAmount < unpaidAmount {
			unpaidAmount = amountItem.UnpaidAmount
			commissionFee = amountItem.CommissionFee
		}
	}
	if unpaidAmount > 0 {
		return nil, errors.WithMessage(errors.New("销售订单未结清"))
	}

	// 检查是否有未送厨的商品。场景：当收银机1结账时，收银机2加购了新的商品。
	if len(saleBill.GetSaleOrderProductUnCooking()) > 0 {
		return nil, errors.New("有未送厨的商品")
	}

	// 计算抹零金额. 只有没有手续费时，才能抹零
	if commissionFee == 0 {
		saleOrder.SetCheckOutZeroFee()
	}

	// 最终应收=应收金额+手续费-结账抹零金额
	finalAmount := decimal.NewFromFloat(saleOrder.GetAmountValue()).Add(decimal.NewFromFloat(commissionFee)).Sub(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee)).InexactFloat64()

	totalPay := float64(0) // 总付款金额=各个付款单的实收金额之和
	for _, paymentOrder := range infoResp.PaymentOrders.List {
		// 如果订单没有会员，但又支付了会员余额，则提示先撤销会员余额支付
		if paymentOrder.PaymentMethodCode == constant.PaymentMethodCodeBalance && saleOrder.ConsumerUuid == 0 {
			return nil, errors.New("订单没有会员，请撤销会员余额支付")
		}
		totalPay = decimal.NewFromFloat(totalPay).Add(decimal.NewFromFloat(paymentOrder.Amount)).InexactFloat64()
	}
	originTotalPay := totalPay // 结账完成后的弹窗要显示的金额。需要包含找零金额

	// 现金支付的金额，未减掉找零的金额
	cashAmount := saleOrder.GetCashAmount()
	outMoney := totalPay - finalAmount // 超付金额=支付金额-最终应收
	// 如果超付金额大于现金支付金额，则拒绝完成订单，提示"收款金额大于最终应收，请先修改收款金额"
	if outMoney > cashAmount {
		return nil, errors.New("收款金额大于最终应收，请先修改收款金额")
	}

	// 计算找零金额。
	changeAmount := float64(0)
	if totalPay > finalAmount {
		changeAmount = decimal.NewFromFloat(totalPay).Sub(decimal.NewFromFloat(finalAmount)).InexactFloat64()
	}

	// 注意在检查超付金额大于现金支付金额之后再修改现金付款金额
	// 如果找零金额大于0，则修改现金付款单的payment_amount和amount字段。amount = payment_amount = amount - changeAmount
	var cashPaymentOrder *model.PaymentOrder
	if changeAmount > 0 {
		for index, paymentOrder := range saleOrder.PaymentOrders {
			if paymentOrder.IsDelete() {
				continue
			}
			if paymentOrder.PaymentMethod.IsCash() {
				saleOrder.PaymentOrders[index].PaymentAmount = paymentOrder.Amount - changeAmount
				saleOrder.PaymentOrders[index].Amount = paymentOrder.Amount - changeAmount
				cashPaymentOrder = saleOrder.PaymentOrders[index]
			}
		}
		// 总付款金额=各个付款单的实收金额之和。总付款金额=总付款金额-找零金额
		totalPay = totalPay - changeAmount
	}

	// 现金支付的金额，已减掉找零的金额
	cashAmount = saleOrder.GetCashAmount()

	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 修改订单为支付完成，并记录找零金额、最终付款金额等结算后才计算的字段
	final := model.FinalAmount{
		CouponAmount:         saleOrder.CalcCouponExchangeAmount(),
		PaymentAmount:        totalPay,
		ChangeAmount:         changeAmount,
		ZeroCheckoutFee:      saleOrder.CalcCheckOutZeroFee(),
		FinalPrice:           finalAmount,
		PaymentCommissionFee: commissionFee,
		GiftAmount:           saleOrder.CalcGiftAmount(saleOrder.SaleOrderProducts),
		Unit:                 currencySetting.Unit,
	}
	saleOrder.SetFinishStatus(final) // 设置销售订单状态为已结清

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取商家的会员设置
	pointsSetting, err := s.settingSrv.GetPointsSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 计算本单获取的积分. 如果订单没有会员，则不计算
	if saleOrder.ConsumerUuid != 0 {
		// 计算积分
		// 根据订单类型（自助餐订单或非自助餐订单）选择积分策略（按比例或按人数）
		pointsRule, err := s.GetPointsRuleInfo(ctx, saleBill.IsBuffetSaleBill(), saleOrder.Member.MemberLevelUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		saleOrder.SetGiftPointsRate(int(saleBill.MealNum), *pointsRule)
	}

	// 会员余额扣费相关
	memberBalanceAmount, memberGiftBalanceAmount := float64(0), float64(0)
	var balancePaymentOrder *model.PaymentOrder // 余额支付的付款单
	if saleOrder.ConsumerUuid != 0 {
		// 加锁. 避免会员余额并发操作
		s.lock.LockUuid(saleOrder.ConsumerUuid)
		defer s.lock.UnlockUuid(saleOrder.ConsumerUuid)
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.Member.Uuid)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取会员信息失败"), err.Error())
		}
		saleOrder.Member = member // 将最新的会员信息赋值给销售订单. 避免并发问题
		// 扣减会员余额
		// 获取该销售订单使用会员余额支付的金额
		balanceAmount := saleOrder.GetMemberBalanceAmount()
		if member.GetBalanceAll() < memberBalanceAmount {
			return nil, errors.New("会员余额不足,请先充值")
		}
		if balanceAmount > 0 {
			// 扣减会员余额
			deductRatioMain, deductRatioGift := pointsSetting.GetDeductRatioMainAndGift()
			memberBalanceAmount, memberGiftBalanceAmount = member.SetFrozenBalance(balanceAmount, deductRatioMain, deductRatioGift)
			// 更新付款单，记录退款金额。主账户扣款多少、赠送帐户扣款多少
			for _, paymentOrder := range saleOrder.PaymentOrders {
				if paymentOrder.PaymentMethod.IsBalance() {
					paymentOrder.SetUpdate()
					paymentOrder.BalanceAmount = memberBalanceAmount         // 主账户扣款多少
					paymentOrder.GiftBalanceAmount = memberGiftBalanceAmount // 赠送帐户扣款多少
					balancePaymentOrder = paymentOrder
				}
			}
		}
		// 更新会员消费金额和消费次数
		consumptionAmount := decimal.NewFromFloat(saleOrder.GetAmountValue()).Sub(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee)).Truncate(2).InexactFloat64()
		repository.NewMemberRepo(db).IncConsumptionAmount(saleOrder.ConsumerUuid, consumptionAmount)
		repository.NewMemberRepo(db).IncConsumptionCount(saleOrder.ConsumerUuid)
		// 处理会员升级 todo 如果后面的逻辑报错，这个升级没有回滚，应该放在事务中升级
		go s.memberSrv.HandleMemberUpgrade(ctx.GetCompanyUuid(), saleOrder.ConsumerUuid)
	}

	// 记录会员余额
	saleOrder.SetMemberBalance()

	needCancelCoupon := false // 是否需要取消优惠券

	// 加锁, 避免并发问题
	if saleOrder.HasCoupon() {
		lock.NewSystemLock().LockUuid(constant.LockNameActivityConsumption)
		defer lock.NewSystemLock().UnlockUuid(constant.LockNameActivityConsumption)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		ctx.SetDB(db)
		if cashPaymentOrder != nil {
			// 更新现金支付单
			if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*cashPaymentOrder); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新销售订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
			return errors.WithMessage(err)
		}

		// 更新销售账单. 如果可以结束销售账单的话
		if err := s.FinishSaleBill(ctx, saleBill, businessSetting, db); err != nil {
			return errors.WithMessage(err)
		}

		// 更新会员的余额
		if saleOrder.ConsumerUuid != 0 {
			if saleOrder.Member.GetUpdate() {
				if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
					MemberUuid:  saleOrder.Member.Uuid,
					Money:       -memberBalanceAmount,     // 扣减会员余额
					GiftMoney:   -memberGiftBalanceAmount, // 扣减会员赠送余额
					Scene:       constant.MemberBalanceLogConsume,
					Describe:    fmt.Sprintf("用户消费：%s", saleOrder.OrderNo),
					RelatedUuid: saleOrder.Uuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 如果开启积分抵扣，且使用了积分抵扣时，则更新会员的积分余额
		if saleBill.SaleBillSetting.IsOpenPointsExchange() {
			if saleOrder.ConsumerUuid != 0 && saleOrder.PayPoints > 0 {
				if err := s.memberSrv.HandleMemberPoints(ctx, MemberPointsChangeReq{
					Uuid:     saleOrder.Member.Uuid,
					Points:   -saleOrder.PayPoints,
					Scene:    constant.MemberPointLogScenePointsExchange,
					Describe: fmt.Sprintf("订单积分抵扣：%s", saleOrder.OrderNo),
				}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 更新余额支付的付款单. 记录会员余额支付时主账户和赠送账户扣款金额
		if balancePaymentOrder != nil {
			if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*balancePaymentOrder); err != nil {
				return errors.WithMessage(err)
			}
		}

		if cashAmount > 0 {
			// 存现金，更新钱箱
			ctx.SetDB(db)
			if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
				Amount:    cashAmount,
				Scene:     constant.CashBoxLogScenePay,
				OrderUuid: saleOrder.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 核销优惠券
		if err := s.VerifyCoupon(ctx, saleOrder, db); err != nil {
			needCancelCoupon = true
			return errors.WithMessage(err)
		}

		// 更新优惠券抵扣金额
		if saleOrder.HasCoupon() {
			if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponAmount(saleOrder.Uuid, saleOrder.CouponAmount); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新发票信息
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
			if err != nil {
				return errors.WithMessage(err)
			}
			saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
			saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderErpInvoice(saleOrder.Uuid, saleOrder.ErpProductsInvoiceName, saleOrder.ErpMaterialInvoiceName); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); err != nil {
		if needCancelCoupon {
			// 取消优惠券
			saleOrder.SetAllCouponCancel()
			if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
				return nil, errors.WithMessage(err, "取消销售订单会员优惠券失败")
			}
			return nil, errors.WithMessage(err, "请刷新优惠券列表")
		}
		return nil, errors.WithMessage(err)
	}

	// 事务结束了，从新使用回db，而不是tx
	ctx.SetDB(db)

	// 出库
	go func() {
		// 判断销售订单的每个商品是否都已有对应的出库记录
		// 获取没有出库记录的销售订单商品
		db := s.dbm.GetDB(ctx.GetDbId())
		ctx := ctx.Copy()
		ctx.SetDB(db)
		withoutWarehouseOutFormSaleOrderProducts, err := s.getSaleOrderProductWithoutWarehouseOutForm(ctx, saleOrder.Uuid, saleOrder.SaleOrderProducts)
		if err != nil {
			logger.Logger.Error("出库失败 - 01", zap.Error(err))
			return
		}
		// 获取减库存的清单信息
		decreaseStockList, err := s.getDecreaseStockList(ctx, withoutWarehouseOutFormSaleOrderProducts)
		if err != nil {
			logger.Logger.Error("出库失败 - 02", zap.Error(err))
			return
		}

		staffShiftLogUuid := uint64(0)
		staffShiftLog, err := GetCurrentStaffShiftLog(db, ctx.GetStaffUuid())
		if err != nil {
			logger.Logger.Error("出库失败 - 02.1", zap.Error(err))
		} else {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
		// 构建出库单

		warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, true, request.SaleBillUuid, ctx.GetStaffUuid(), staffShiftLogUuid)
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			for _, warehouseOutForm := range warehouseOutForms {
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
			}
			return nil
		}); err != nil {
			logger.Logger.Error("出库失败 - 03", zap.Error(err))
		}
	}()

	// 该订单的所有出库记录都标记已出库。将预出库的状态改为已出库
	repository.NewWarehouseFormRepo(db).UpdateWarehouseOutFormItemRecordsStatus(saleOrder.Uuid)

	// 发送结账短信
	if saleOrder.ConsumerUuid != 0 {
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
		if err != nil {
			ctx.Log().Info("停止发送短信，获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			go func() {
				var memberPaymentOrder *resp.PaymentOrder
				for _, paymentOrder := range infoResp.PaymentOrders.List {
					if paymentOrder.PaymentMethodCode == constant.PaymentMethodCodeBalance {
						memberPaymentOrder = &paymentOrder
						break
					}
				}

				if member != nil {
					smsReq := sms.MemberConsumptionRequest{
						Company:        ctx.GetCompany().Name,
						Consumption:    saleOrder.FinalPrice,
						IncreasePoints: saleOrder.GiftPoints,
						Balance:        member.GetBalanceAll(),
						PointsBalance:  decimal.NewFromFloat(member.GetPoints()).Add(decimal.NewFromFloat(saleOrder.GiftPoints)).Round(2).InexactFloat64(), // 会员积分=会员积分+本次增加的积分。 此时积分还未增加到会员表中
					}
					if memberPaymentOrder != nil {
						smsReq.MemberPay = memberPaymentOrder.Amount
					} else {
						// 没有余额支付单，则认为没有会员余额支付,MemberPay=0
						smsReq.MemberPay = 0
					}

					if err := s.smsSrv.SendMemberConsumptionSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送结账短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送结账短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			}()
		}
	}

	// 发布"结账"事件
	originSaleOrderAmount := saleOrder.GetOriginAmountValue()
	saleOrderPaymentAmount := saleOrder.PaymentAmount
	saleOrderChangeAmount := saleOrder.ChangeAmount
	go func() {

		// 结账前，发布"抹零"事件。如果优惠折扣自动抹零且抹零金额不为0，则发布"抹零"事件。
		if saleOrder.IsAutoZeroDiscount(*saleBill.SaleBillSetting) && saleOrder.ZeroFee != 0 {
			event.NewSystemBus().PublishDiscountZeroSaleOrderEvent(event.DiscountSaleOrderPayload{
				BasePayload: event.BasePayload{ // 订单抹零
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  request.SaleBillUuid,
					SaleOrderUuid: request.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				DiscountType:    constant.DiscountOperationLogTypeZeroSaleOrder,
				RoundingType:    int(saleOrder.ZeroRule),
				SpecialDiscount: saleOrder.ZeroFee, // ZeroFee这个字段是算好的抹零优惠金额。先计算好订单应付金额，再根据抹零规格进行抹零得到的结果
				IsAuto:          true,
			})
		}

		// 结账前，发布"结账抹零"事件。如果结账自动抹零且抹零金额不为0，则发布"结账抹零"事件
		if saleOrder.IsAutoCheckoutZeroDiscount(*saleBill.SaleBillSetting) && saleOrder.ZeroCheckoutFee != 0 {
			s.bus.PublishCheckoutZeroSaleOrderEvent(event.CheckoutZeroSaleOrderPayload{
				BasePayload: event.BasePayload{ // 结账抹零
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  request.SaleBillUuid,
					SaleOrderUuid: request.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				Operation:       constant.OrderCheckoutDiscountAdd,
				RoundingType:    int(saleOrder.ZeroCheckoutRule),
				SpecialDiscount: saleOrder.ZeroCheckoutFee,
				IsAuto:          true,
			})
		}

		payTypes := make([]event.PayType, 0)
		for _, paymentOrder := range infoResp.PaymentOrders.List {
			payTypes = append(payTypes, event.PayType{
				Name:           paymentOrder.PaymentMethodName,
				Value:          paymentOrder.PaymentMethodCode,
				DisabledCancel: utils.BoolToUint(paymentOrder.DisabledCancel),
				Price:          paymentOrder.Amount,
				FeeMoney:       paymentOrder.PaymentCommissionFee,
			})
		}
		s.bus.PublishCheckoutSaleOrderEvent(event.CheckoutSaleOrderPayload{
			BasePayload: event.BasePayload{ // 结账
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  request.SaleBillUuid,
				SaleOrderUuid: request.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleBill:    saleBill,
			OrderPrice:  originSaleOrderAmount,
			PayPrice:    saleOrderPaymentAmount,
			ActualPrice: totalPay, // 最终实付金额=每笔付款单的付款金额之和（含手续费）- 找零金额
			ChangeDue:   saleOrderChangeAmount,
			PayType:     payTypes,
		})
	}()

	// 整单完结时, 发布"统计"事件
	if saleBill.CanFinishSaleBill() {
		go func() {
			s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
				BasePayload: event.BasePayload{ // 统计
					Ctx: ctx,
				},
				SaleBillUuid: saleBill.Uuid,
			})
		}()
	}

	// 返回结果
	payMethods := make([]resp.PayMethod, 0)
	for _, paymentOrder := range infoResp.PaymentOrders.List {
		method := resp.PayMethod{
			Uuid: paymentOrder.PaymentMethodUuid,
			Name: paymentOrder.PaymentMethodName,
		}
		payMethods = append(payMethods, method)
	}
	return &resp.OrderFinishResp{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		AmountInfo: resp.PayAmountInfo{
			OrderAmount:  saleOrder.FinalPrice, // 最终应收
			PayAmount:    originTotalPay,       // 原总付款=总付款-找零金额
			ChangeAmount: saleOrderChangeAmount,
		},
		PayMethodList: resp.PayMethodList{
			List: payMethods,
		},
	}, nil
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("无法查询到销售订单"))
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
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

	// 取消积分抵扣
	saleOrder.SetPayPointsCancel()
	// 记录会员余额
	saleOrder.SetMemberBalance()

	updateSaleBill := false
	// 如果销售账单中只有一个销售订单，则可以结束销售账单
	if saleBill.CanFinishSaleBill() {
		staff := ctx.GetStaff()
		saleBill.SetFinishSaleBill(staff.DutyNo, staff.Uuid, staff.GetUserName())
		saleBill.CalcAll()
		updateSaleBill = true
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 取消优惠券
	saleOrder.SetAllCouponCancel()
	if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
		return nil, errors.WithMessage(err, "取消销售订单优惠券失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建免单原因
		if len(freeOrderReasons) > 0 {
			if err := repository.NewSaleOrderProductReasonRepo(db).CreateSaleOrderProductReasons(freeOrderReasons); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 保存发票到erp
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
			if err != nil {
				return errors.WithMessage(err)
			}
			saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
			saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
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

		ctx.SetDB(db)
		// 拒绝所有待接单的h5订单
		if err := s.RejectAllH5Order(ctx, saleBill.Uuid); err != nil {
			return err
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
					Name: i18n.Translate(ctx.GetLanguage(), "免单"),
				},
			},
		},
	}

	// 发布"免单"事件
	go func() {
		// 结账前，发布"抹零"事件。如果优惠折扣自动抹零且抹零金额不为0，则发布"抹零"事件。
		if saleOrder.IsAutoZeroDiscount(*saleBill.SaleBillSetting) && saleOrder.ZeroFee != 0 {
			event.NewSystemBus().PublishDiscountZeroSaleOrderEvent(event.DiscountSaleOrderPayload{
				BasePayload: event.BasePayload{ // 订单抹零
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  req.SaleBillUuid,
					SaleOrderUuid: req.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				DiscountType:    constant.DiscountOperationLogTypeZeroSaleOrder,
				RoundingType:    int(saleOrder.ZeroRule),
				SpecialDiscount: saleOrder.ZeroFee, // ZeroFee这个字段是算好的抹零优惠金额。先计算好订单应付金额，再根据抹零规格进行抹零得到的结果
				IsAuto:          true,
			})
		}

		s.bus.PublishFreeSaleOrderEvent(event.FreeSaleOrderPayload{
			BasePayload: event.BasePayload{ // 免单
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleBill:      saleBill,
			OrderPrice:    saleOrder.GetOriginAmountValue(),
			PayPrice:      0, // 免单时，支付金额为0
			ActualPrice:   0, // 免单时，实际支付金额为0
			ChangeDue:     0, // 免单时，找零金额为0
			IsFree:        utils.BoolToUint(true),
			DiscountMoney: saleOrder.GetAmount(),
		})
	}()

	// 发布"统计"事件
	go func() {
		s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			SaleBillUuid: saleBill.Uuid,
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, req.SaleOrderUuid); err != nil {
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

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	zeroAmount := infoResp.GetZeroAmount()

	// 发布"结账抹零"事件
	go func() {
		s.bus.PublishCheckoutZeroSaleOrderEvent(event.CheckoutZeroSaleOrderPayload{
			BasePayload: event.BasePayload{ // 结账抹零
				Ctx:           ctx,
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

	// 最大只能创建10个
	if len(saleBill.SaleOrders) == 10 {
		return nil, errors.New("销售账单最多只能创建10个销售订单")
	}

	// 如果销售账单目前只有一个销售订单，增加一个销售订单后要求撤销订单1的优惠折扣
	// 这是产品的特殊要求，可能后续会改。
	// 撤销订单的优惠折扣
	if len(saleBill.SaleOrders) == 1 {
		saleOrder := saleBill.GetFirstSaleOrder()
		// 撤销订单1的优惠折扣
		if saleOrder.IsManualDiscount(uint8(saleBill.SaleBillSetting.ZeroRule)) {
			saleOrder.SetAllDiscountCancel()
		}
		// 撤销订单1的会员折扣
		saleOrder.SetMemberDiscountCancel()
	}

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 生成订单编号
	var orderSourceType string
	if saleBill.IsDeskSaleBill() {
		orderSourceType = constant.OrderSourceDesk
	} else {
		orderSourceType = constant.OrderSourceInstant
	}
	orderNo, err := s.createOrderNo(db, orderSourceType)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 设置拆单
	saleBill.SetIsSplitOrder(true)

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建销售订单
		if _, errCreateSaleOrder := createSaleOrder(ctx, db, saleBill.SaleBillSetting, saleBill.Uuid, orderNo); errCreateSaleOrder != nil {
			return errors.WithMessage(errCreateSaleOrder, fmt.Sprintf("新建拆单失败,saleBill.Uuid:%v, orderNo:%v", saleBill.Uuid, orderNo))
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

	var orders []event.Order
	for i, order := range cartInfo.SaleOrderList {
		orders = append(orders, event.Order{
			SaleOrderUuid: order.Uuid,
			OrderName:     fmt.Sprintf("%d", i+1),
			Amount:        order.AmountInfo.Amount,
		})
	}

	// 发布"拆单"操作事件
	go func() {
		s.bus.PublishSplitOrderEvent(event.SplitOrderPayload{
			BasePayload: event.BasePayload{ // 拆单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			Orders: orders,
		})
	}()

	return cartInfo, nil
}

func MoreThanMoveNum(saleOrderProductNum, moveNum float64) bool {
	return saleOrderProductNum > moveNum
}

func LessThanMoveNum(saleOrderProduct *model.SaleOrderProduct, moveNum float64) bool {
	return saleOrderProduct.Num < moveNum
}

func EqualMoveNum(saleOrderProductNum, moveNum float64) bool {
	return saleOrderProductNum == moveNum
}
func IsSameSignature[T any](sign string, toSaleOrderProductSignMap map[string]*T) bool {
	return toSaleOrderProductSignMap[sign] != nil
}

func (s *orderSrv) CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error {
	// 保存到数据库
	if db == nil {
		db = s.dbm.GetDB(ctx.GetDbId())
	}
	return CalcAndSaveSaleBill(ctx, db, saleBill, options...)
}

func CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error {
	option := &model.CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	// 计算订单商品、订单、账单
	saleBill.CalcAll(options...)
	// 设置收银员信息
	staff := ctx.GetStaff()
	saleBill.SetCashier(staff.DutyNo, staff.Uuid, staff.GetUserName())

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
				if saleOrderProduct == nil {
					continue
				}
				// 保存订单商品。只有标记更新的商品才会更新
				if err := repository.NewSaleOrderProductRepo(db).UpdateOrCreateSaleOrderProductRecord(*saleOrderProduct); err != nil {
					return errors.WithMessage(err)
				}
				for _, saleOrderProductBom := range saleOrderProduct.SaleOrderProductBoms {
					if saleOrderProductBom.GetUpdate() {
						if err := repository.NewOrderProductBomRepo(db).UpdateOrCreateSaleOrderProductBomRecord(*saleOrderProductBom); err != nil {
							return errors.WithMessage(err)
						}
					}
				}
				for _, saleOrderProductAttribute := range saleOrderProduct.SaleOrderProductAttributes {
					if saleOrderProductAttribute.GetUpdate() {
						if err := repository.NewSaleOrderProductAttributeRepo(db).UpdateOrCreateSaleOrderProductAttributeRecord(*saleOrderProductAttribute); err != nil {
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
			// 保存自助餐加钟商品
			for _, buffetDelayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if err := repository.NewSaleOrderBuffetDelayProductRepo(db).UpdateOrCreateSaleOrderBuffetDelayProductRecord(*buffetDelayProduct); err != nil {
					return errors.WithMessage(err)
				}
			}
			// 更新账单设置
			if option.SaleBillSetting != nil {
				option.SaleBillSetting.Uuid = saleBill.SaleBillSetting.Uuid
				if _, err := repository.NewOrderRepo(db).UpdateSaleBillSetting(*option.SaleBillSetting); err != nil {
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
func (s *orderSrv) moveSaleOrderProduct(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, saleOrderProducts []*model.SaleOrderProduct, moveNumMap map[uint64]float64) (map[uint64]*model.SaleOrderProduct, map[uint64]*model.SaleOrderProduct, error) {
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
		// 处理套餐商品
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrderFrom.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				if subProduct.IsDelete() || subProduct.SaleOrderUuid != saleOrderTo.Uuid {
					continue
				}
				toSaleOrderProductSignMap[subProduct.Sign] = subProduct
			}
		}
	}

	// 遍历要移动的订单商品，移动到目标订单中
	var handleProduct func(saleOrderProducts []*model.SaleOrderProduct, packageProductUuid uint64, moveNum float64) error
	handleProduct = func(saleOrderProducts []*model.SaleOrderProduct, packageProductUuid uint64, moveNum float64) error {
		for _, saleOrderProduct := range saleOrderProducts {
			ctx.Log().Debug("移动商品", zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))

			// 记录套餐商品的uuid
			subPackageUuid := saleOrderProduct.Uuid

			// 获取移动数量。套餐商品的移动数量为子商品的单位数量乘以移动数量
			var moveProductNum float64
			if packageProductUuid == 0 {
				moveProductMapNum, ok := moveNumMap[saleOrderProduct.Uuid]
				if !ok {
					return errors.WithMessage(errors.New("商品可能移动到其他销售订单中"), fmt.Sprintf("sale_order_product_uuid:%d", saleOrderProduct.Uuid))
				}
				if moveProductMapNum > saleOrderProduct.Num {
					return errors.WithMessage(errors.New("移动数量大于销售订单商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", saleOrderProduct.Uuid))
				}
				moveProductNum = moveProductMapNum
			} else {
				unitNum := decimal.NewFromFloat(saleOrderProduct.UnitNum)
				moveProductNum = decimal.NewFromFloat(moveNum).Mul(unitNum).Round(3).InexactFloat64()
			}

			hasHandle := false // 是否已经处理过。因为一个商品被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
			// 第一种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中有签名一样的商品，该商品数量增加移动数量
			if !hasHandle && MoreThanMoveNum(saleOrderProduct.Num, moveProductNum) && IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
				hasHandle = true
				ctx.Log().Debug("移动商品，第一种移动方式", zap.Any("from", saleOrderProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
				// 修改原销售订单商品数量，更新记录，重新计算订单金额
				saleOrderProduct.Num -= moveProductNum
				// 更新套餐商品的子商品数量
				saleOrderProduct.PackageUuid = packageProductUuid
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
				// 更新套餐商品的子商品数量
				newSaleOrderProduct.PackageUuid = packageProductUuid
				// 计算商品数据。折扣、税费、服务
				discountInfo := saleOrderTo.GetDiscountInfo()
				newSaleOrderProduct.SetDiscountInfo(discountInfo.MemberDiscountRate, discountInfo.MemberCardDiscountRate, discountInfo.CustomDiscountRate)
				newSaleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
				// 记录套餐商品的uuid
				subPackageUuid = newSaleOrderProduct.Uuid
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

			// 处理套餐商品
			if saleOrderProduct.IsPackageProduct() {
				err := handleProduct(saleOrderFrom.GetPackageSubProductList(saleOrderProduct.Uuid), subPackageUuid, moveProductNum)
				if err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		return nil
	}

	// 处理商品
	err := handleProduct(saleOrderProducts, 0, 0)
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	return waitUpdateSaleOrderProductMap, waitCreateSaleOrderProductMap, nil
}

// moveBuffetCustomer 移动自助餐顾客
func (s *orderSrv) moveBuffetCustomer(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, buffetCustomers []*model.SaleOrderBuffetCustomerType, moveNumMap map[uint64]float64) (map[uint64]*model.SaleOrderBuffetCustomerType, map[uint64]*model.SaleOrderBuffetCustomerType, error) {
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
		if moveCustomerNum > float64(buffetCustomer.Num) {
			return nil, nil, errors.WithMessage(errors.New("移动数量大于销售订单商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", buffetCustomer.Uuid))
		}
		hasHandle := false // 是否已经处理过。因为一个顾客被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
		// 第一种移动方式：原销售订单顾客数量大于移动数量，则原销售订单顾客数量减少移动数量，目标销售订单中有签名一样的顾客，该顾客数量增加移动数量
		if !hasHandle && MoreThanMoveNum(float64(buffetCustomer.Num), float64(moveCustomerNum)) && IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动顾客，第一种移动方式", zap.Any("from", buffetCustomer.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			// 修改原销售订单顾客数量，更新记录，重新计算订单金额
			buffetCustomer.Num -= uint(moveCustomerNum)
			// 修改目标销售订单顾客数量，更新记录，重新计算订单金额
			toBuffetCustomerSignMap[buffetCustomer.GetSign()].Num += uint(moveCustomerNum)
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
			waitUpdateBuffetCustomerMap[toBuffetCustomerSignMap[buffetCustomer.GetSign()].Uuid] = toBuffetCustomerSignMap[buffetCustomer.GetSign()]
		}

		// 第二种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && MoreThanMoveNum(float64(buffetCustomer.Num), float64(moveCustomerNum)) && !IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动顾客，第二种移动方式", zap.Any("from", buffetCustomer.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			ctx.Log().Debug("移动顾客", zap.Any("原销售订单商品修改前数量", buffetCustomer.Num))
			// 修改原销售订单商品数量，更新记录，重新计算订单金额
			buffetCustomer.Num -= uint(moveCustomerNum)
			// 新建一个销售订单商品，该商品数量为移动数量
			newBuffetCustomer := buffetCustomer.CopyBuffetCustomer(saleOrderTo.Uuid)
			newBuffetCustomer.Num = uint(moveCustomerNum)
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
		if !hasHandle && EqualMoveNum(float64(buffetCustomer.Num), float64(moveCustomerNum)) && IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动商品，第三种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", buffetCustomer.Uuid), zap.Any("buffetCustomer", buffetCustomer.Name))
			// 删除原销售订单商品，更新表记录，重新计算原订单金额；
			buffetCustomer.DeleteTime = time.Now().Unix()
			// 修改目标销售订单商品数量，更新记录，重新计算订单金额
			toBuffetCustomerSignMap[buffetCustomer.GetSign()].Num += uint(moveCustomerNum)
			// 记录到待更新列表中
			waitUpdateBuffetCustomerMap[buffetCustomer.Uuid] = buffetCustomer
			waitUpdateBuffetCustomerMap[toBuffetCustomerSignMap[buffetCustomer.GetSign()].Uuid] = toBuffetCustomerSignMap[buffetCustomer.GetSign()]
		}

		// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
		if !hasHandle && EqualMoveNum(float64(buffetCustomer.Num), float64(moveCustomerNum)) && !IsSameSignature(buffetCustomer.GetSign(), toBuffetCustomerSignMap) {
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
func (s *orderSrv) moveBuffetDelayProduct(ctx context.Context, saleBill *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, delayProducts []*model.SaleOrderBuffetDelayProduct, moveNumMap map[uint64]float64) (map[uint64]*model.SaleOrderBuffetDelayProduct, map[uint64]*model.SaleOrderBuffetDelayProduct, error) {
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
		if moveCustomerNum > float64(delayProduct.Num) {
			return nil, nil, errors.WithMessage(errors.New("移动数量大于加钟商品数量"), fmt.Sprintf("sale_order_product_uuid:%d", delayProduct.Uuid))
		}

		hasHandle := false // 是否已经处理过。因为一个加钟商品被一个处理方式处理过后，可能又满足多种移动方式，所以需要一个标志来判断是否已经处理过
		// 第一种移动方式：原销售订单加钟商品数量大于移动数量，则原销售订单加钟商品数量减少移动数量，目标销售订单中有签名一样的加钟商品，该加钟商品数量增加移动数量
		if !hasHandle && MoreThanMoveNum(float64(delayProduct.Num), float64(moveCustomerNum)) && IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第一种移动方式", zap.Any("from", delayProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			// 修改原销售订单加钟商品数量，更新记录，重新计算订单金额
			delayProduct.Num -= uint(moveCustomerNum)
			// 修改目标销售订单加钟商品数量，更新记录，重新计算订单金额
			toBuffetDelayProductSignMap[delayProduct.GetSign()].Num += uint(moveCustomerNum)
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitUpdateBuffetDelayProductMap[toBuffetDelayProductSignMap[delayProduct.GetSign()].Uuid] = toBuffetDelayProductSignMap[delayProduct.GetSign()]
		}

		// 第二种移动方式：原加钟商品数量大于移动数量，则原加钟商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个加钟商品，该商品数量为移动数量
		if !hasHandle && MoreThanMoveNum(float64(delayProduct.Num), float64(moveCustomerNum)) && !IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第二种移动方式", zap.Any("from", delayProduct.SaleOrderUuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			ctx.Log().Debug("移动加钟商品", zap.Any("原加钟商品修改前数量", delayProduct.Num))
			// 修改原加钟商品数量，更新记录，重新计算订单金额
			delayProduct.Num -= uint(moveCustomerNum)
			// 新建一个加钟商品，该商品数量为移动数量
			newBuffetCustomer := delayProduct.CopyBuffetDelayProduct(saleOrderTo.Uuid)
			newBuffetCustomer.Num = uint(moveCustomerNum)
			// 在目标销售订单中新建一个加钟商品
			saleOrderTo.SaleOrderBuffetDelayProducts = append(saleOrderTo.SaleOrderBuffetDelayProducts, newBuffetCustomer)
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitCreateBuffetDelayProductMap[newBuffetCustomer.Uuid] = newBuffetCustomer
			ctx.Log().Debug("移动加钟商品", zap.Any("原加钟商品数量", delayProduct.Num), zap.Any("目标加钟商品数量", newBuffetCustomer.Num))
		}

		// 第三种移动方式：原加钟商品数量等于移动数量，则原加钟商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
		if !hasHandle && EqualMoveNum(float64(delayProduct.Num), float64(moveCustomerNum)) && IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
			hasHandle = true
			ctx.Log().Debug("移动加钟商品，第三种移动方式", zap.Any("from", saleOrderFrom.Uuid), zap.Any("to", saleOrderTo.Uuid), zap.Any("product uuid", delayProduct.Uuid), zap.Any("delayProduct", delayProduct.Name))
			// 删除原加钟商品，更新表记录，重新计算原订单金额；
			delayProduct.DeleteTime = time.Now().Unix()
			// 修改目标加钟商品数量，更新记录，重新计算订单金额
			toBuffetDelayProductSignMap[delayProduct.GetSign()].Num += uint(moveCustomerNum)
			// 记录到待更新列表中
			waitUpdateBuffetDelayProductMap[delayProduct.Uuid] = delayProduct
			waitUpdateBuffetDelayProductMap[toBuffetDelayProductSignMap[delayProduct.GetSign()].Uuid] = toBuffetDelayProductSignMap[delayProduct.GetSign()]
		}

		// 第四种移动方式：原加钟商品数量等于移动数量，则原加钟商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个加钟商品，该商品数量为移动数量
		if !hasHandle && EqualMoveNum(float64(delayProduct.Num), float64(moveCustomerNum)) && !IsSameSignature(delayProduct.GetSign(), toBuffetDelayProductSignMap) {
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

// SaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
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
	moveProductMap := make(map[uint64]float64)
	for _, moveProduct := range req.Products {
		moveProductMap[moveProduct.Uuid] = moveProduct.Num
	}

	saleOrderProducts, saleOrderBuffetCustomers, buffetDelayProducts, err := s.getMoveProductInfo(ctx, saleOrderFrom, req)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	moveNumMap := make(map[uint64]float64)
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

type MustPlanConfirmOption struct {
	IsH5Order bool // 是否是H5订单
}

func WithIsH5Order() func(option *MustPlanConfirmOption) {
	return func(option *MustPlanConfirmOption) {
		option.IsH5Order = true
	}
}

// InstantOrderMustPlanConfirm 确认必点商品
func (s *orderSrv) InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq, opts ...func(option *MustPlanConfirmOption)) (bool, *resp.InstantProductMustPlan, error) {
	option := &MustPlanConfirmOption{}
	for _, opt := range opts {
		opt(option)
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	var mustPlanList []resp.InstantProductMustPlan

	// 点餐页面未创建销售账单时，但前端已显示出必点弹窗
	if req.SaleBillUuid == 0 {
		// 获取点餐的必选方案列表
		mustPlanLists, err := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, make(ro.MustPlanProductInfo))
		if err != nil {
			return false, nil, errors.WithMessage(err, "GetInstantMustPlanList failed")
		}
		if len(mustPlanLists) == 0 {
			return false, nil, errors.New("没有必点方案")
		}
		mustPlanList = mustPlanLists
	} else {
		// 获取销售账单信息
		saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
		if errSaleBill != nil {
			ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
			return false, nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
		}

		// 查询到购物车信息
		var shopCartInfo *ro.ShopCartRepo
		var err error
		if option.IsH5Order {
			// 如果是h5订单时要查询出未下单的商品
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid, repository.WithNotDeleted())
		} else {
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
		}
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return false, nil, errors.WithMessage(err, "获取购物车信息失败")
		}

		if saleBill.IsDeskSaleBill() {
			mustPlan, errMustPlan := s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), saleBill.DeskUuid)
			if errMustPlan != nil {
				return false, nil, errMustPlan
			}
			mustPlanList = mustPlan
		} else {
			mustPlan, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
			if errMustPlan != nil {
				return false, nil, errMustPlan
			}
			mustPlanList = mustPlan
		}
	}

	// 判断必点商品是否售罄
	if mustPlanList != nil && len(mustPlanList) > 0 {
		for _, plan := range mustPlanList {
			if plan.NeedNum > 0 {
				// 判断商品是否售罄。如果售罄，则允许"确认必点"
				isSoldOut, err := planProductSoldOut(ctx, &plan)
				if err != nil {
					return false, nil, errors.WithMessage(err)
				}
				if isSoldOut {
					break // 所有的未满足必点的商品都没有库存时，允许"确认必点"
				} else {
					ctx.Log().Info("确认必点商品失败，必点商品未点", zap.Any("plan name", plan.Name))
					return false, &plan, nil
				}
			}
		}
	}

	// 点餐订单需要创建
	if req.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return false, nil, errors.WithMessage(err)
		}
		if billInfo != nil && hasInstantOrder {
			req.SaleBillUuid = billInfo.Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return false, nil, errors.WithMessage(err)
			}
			req.SaleBillUuid = order.SaleBillUuid
		}
	}

	// 修改sale_bill表的show_must_plan
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(req.SaleBillUuid); err != nil {
		return false, nil, errors.WithMessage(err, "确认必点商品失败")
	}

	return true, nil, nil
}

func planProductSoldOut(ctx context.Context, plan *resp.InstantProductMustPlan) (bool, error) {
	// 如果是可选商品. 只有必点方案的所有商品都无库存时才返回true
	if plan.MustRule == constant.ProductMustPlanMustRuleAny {
		isSaleOut := true
		for _, product := range plan.Products.List {
			// 未满足必点的商品包
			productPackage, err := repository.NewProductPackageRepo(ctx.GetDB()).GetProductPackageBoms(product.Product.Uuid)
			if err != nil {
				return false, errors.WithMessage(err)
			}
			sisSaleOut := productPackage.IsSaleout()
			if !sisSaleOut {
				isSaleOut = false
				break
			}
		}
		// 可选商品. 只有必点方案的所有商品都无库存时才返回true
		return isSaleOut, nil
	}
	if len(plan.Products.List) == 0 {
		return false, nil
	}
	isSaleOut := true
	for _, product := range plan.Products.List {
		if product.NeedNum <= 0 {
			continue
		}
		// 未满足必点的商品包
		productPackage, err := repository.NewProductPackageRepo(ctx.GetDB()).GetProductPackageBoms(product.Product.Uuid)
		if err != nil {
			return false, errors.WithMessage(err)
		}
		if !productPackage.IsSaleout() {
			isSaleOut = false
			break
		}
	}
	// 所有的未满足必点的商品都没有库存时才返回true
	return isSaleOut, nil
}

// OrderCheck 订单检查
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

	// 助手拆单之后不能操作
	if ctx.GetSource() == constant.SourceAssistant && len(saleBill.SaleOrders) > 1 {
		return nil, errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	ctx.Log().Debug("获取销售账单信息")

	// 从http的header中获取h5_order_uuid
	h5OrderProductUnAccept := make([]*model.SaleOrderProduct, 0) // 未接单的h5订单商品
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		h5OrderProductUnAccept = saleBill.GetH5OrderProductUnAccept(h5OrderUuid)
	}

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()

	// 获取所有商品,用于检查限购
	var saleOrderProductAll []*model.SaleOrderProduct
	if h5OrderUuid != 0 {
		// 如果从接到"进入桌台"时操作结账检查，要加入本h5订单的商品
		saleOrderProductAll = saleBill.GetSaleOrderProductAll(model.WithH5CheckLimit(), model.WithH5OrderUuid(h5OrderUuid))
	} else {
		saleOrderProductAll = saleBill.GetSaleOrderProductAll()
	}
	// 结账检查时，只检查未送厨的商品是否超过限购
	saleOrderProductUuids := make([]uint64, 0)
	for _, saleOrderProduct := range unCookingSaleOrderProducts {
		saleOrderProductUuids = append(saleOrderProductUuids, saleOrderProduct.Uuid)
	}
	for _, saleOrderProduct := range h5OrderProductUnAccept {
		saleOrderProductUuids = append(saleOrderProductUuids, saleOrderProduct.Uuid)
	}

	// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
	var deskUuid uint64
	if saleBill.IsDeskSaleBill() {
		deskUuid = saleBill.DeskUuid
	}

	var checkServiceRes *resp.OrderCheckServiceRes
	var errCheck error
	if h5OrderUuid != 0 {
		// 接单场景下，检查h5订单商品金额
		checkServiceRes, errCheck = s.checkOrder(ctx, req.IgnoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithSaleOrderProductUuid(saleOrderProductUuids...), WithH5OrderUuid(h5OrderUuid))
	} else {
		checkServiceRes, errCheck = s.checkOrder(ctx, req.IgnoreMust, db, req.SaleBillUuid, deskUuid, saleOrderProductAll, WithCheckTypeCheckout(), WithSaleOrderProductUuid(saleOrderProductUuids...))
	}
	if errCheck != nil {
		return nil, errors.WithMessage(errCheck, "订单检查失败")
	}
	if checkServiceRes != nil {
		return checkServiceRes, nil
	}

	// 检查自助餐顾客类型价格是否变动
	res := s.checkBuffetCustomerTypePriceChanged(ctx, saleBill)
	if res != nil {
		{
			// 如果价格变化，更新销售订单顾客的价格。都是后台更新价格而未立即更新已选购商品的价格引起的
			// 价格变化包括：
			// 1. 自助餐顾客类型价格变化
			shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			saleBill := shopCartInfo.SaleBill
			s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice())
		}
		return res, nil
	}

	// 检查含税未含税是否改变、服务费类型是否改变、服务费是否改变、服务费比例是否改变
	res, newSetting, err := s.checkSaleBillSettingChanged(ctx, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res != nil {
		{
			// 如果账单快照设置变化，更新销售订单的金额。都是后台更新后而未立即更新账单设置引起的
			// 账单快照设置变化包括：
			// 1. 含税未含税是否改变
			// 2. 服务费类型是否改变
			// 3. 固定服务费是否改变
			// 4. 服务费比例是否改变
			shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			saleBill := shopCartInfo.SaleBill
			s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting))
		}
		return res, nil
	}

	// 检查是否有未送厨的商品
	if len(unCookingSaleOrderProducts) > 0 || len(h5OrderProductUnAccept) > 0 {
		products := make([]resp.Product, 0)
		for _, product := range unCookingSaleOrderProducts {
			if product.IsPackageSubProduct() {
				continue
			}
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
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
		for _, product := range h5OrderProductUnAccept {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
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

		// 获取预送厨的商品
		preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
		for _, product := range preCookingSaleOrderProducts {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
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
	} else {
		products := make([]resp.Product, 0)
		// 获取预送厨的商品
		preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
		for _, product := range preCookingSaleOrderProducts {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
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
		if len(products) > 0 {
			res := &resp.OrderCheckServiceRes{
				Code:          constant.CodeOrderCheckProductUnCooking,
				OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: products}},
			}
			return res, nil
		}
	}

	return nil, nil
}

// getMoveProductList 获取删除某个子单要移动的商品列表,包括订单商品、订单顾客、订单加钟商品
func (s *orderSrv) getMoveProductList(saleOrderFrom *model.SaleOrder) []req.MoveProduct {
	moveProductList := make([]req.MoveProduct, 0) // 移动商品列表,包括订单商品、订单顾客、订单加钟商品
	// 获取要移动的商品
	for _, saleOrderProduct := range saleOrderFrom.SaleOrderProducts {
		if saleOrderProduct.IsDelete() || saleOrderProduct.Num == 0 || saleOrderProduct.ProductType == constant.ProductTypePackageSubProduct {
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
			Num:  float64(saleOrderBuffetCustomer.Num),
		})
	}
	// 获取要移动的加钟商品
	for _, buffetDelayProduct := range saleOrderFrom.SaleOrderBuffetDelayProducts {
		if buffetDelayProduct.IsDelete() || buffetDelayProduct.Num == 0 {
			continue
		}
		moveProductList = append(moveProductList, req.MoveProduct{
			Uuid: buffetDelayProduct.Uuid,
			Num:  float64(buffetDelayProduct.Num),
		})
	}
	return moveProductList
}

// InstantOrderSaleOrderDelete 删除一个销售订单(删除拆单)
func (s *orderSrv) InstantOrderSaleOrderDelete(ctx context.Context, request req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error) {
	var shopCart *resp.ShopCart
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

	// 如果第一个销售订单已经结账且要删除的订单还有商品的话，则提示"拆单1已结账，请结账当前拆单或删除商品后再删除拆单"。已送厨的商品也要先退菜再删除后才能删除拆单
	if firstSaleOrder.IsSettled() && len(moveProductList) > 0 {
		return nil, errors.New("拆单1已结账，请结账当前拆单或删除商品后再删除拆单")
	}

	// 如果第一个销售订单已经结账且要删除的订单没有商品且销售订单数量大于2时，则删除该拆单
	if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) > 2 {
		// 如果销售订单中没有商品，则直接删除订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
			ctx.Log().Error("删除订单失败", zap.Error(err))
			return nil, errors.New("删除订单失败")
		}

		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 如果第一个销售订单已经结账且要删除的订单没有商品且销售订单数量等于2时，则删除该拆单并完成该销售账单
	if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) == 2 {
		// 如果销售订单中没有商品，则直接删除订单
		saleOrderFrom.SetDelete()
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				ctx.Log().Error("删除订单失败", zap.Error(err))
				return errors.New("删除订单失败")
			}

			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return errors.WithMessage(err)
			}

			// 更新销售账单. 如果可以结束销售账单的话
			if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
				return errors.WithMessage(err)
			}
			return nil
		}); err != nil {
			return nil, errors.WithMessage(err)
		}
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 如果销售订单中有商品，则先移动商品到第一个销售订单再删除该子单
	if !firstSaleOrder.IsSettled() && len(moveProductList) > 0 {
		moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
			SaleBillUuid: request.SaleBillUuid,
			From:         request.SaleOrderUuid,
			To:           firstSaleOrder.Uuid,
			Products:     moveProductList,
		}
		var err error
		shopCart, err = s.SaleOrderMoveProduct(ctx, moveProductReq, true)
		if err != nil {
			ctx.Log().Error("移动商品失败", zap.Error(err))
			return nil, errors.WithMessage(err)
		}
	}

	if !firstSaleOrder.IsSettled() && len(moveProductList) == 0 {
		// 如果销售订单中没有商品，则直接删除订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
			ctx.Log().Error("删除订单失败", zap.Error(err))
			return nil, errors.New("删除订单失败")
		}

		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 更新销售账单
	saleBill.SetIsSplitOrder(len(saleBill.SaleOrders)-1 > 1)
	if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
		return nil, errors.WithMessage(errUpdateSaleBill)
	}

	// 发布"拆单"操作事件
	go func() {
		var orders []event.Order
		for i, order := range shopCart.SaleOrderList {
			orders = append(orders, event.Order{
				SaleOrderUuid: order.Uuid,
				OrderName:     fmt.Sprintf("%d", i+1),
				Amount:        order.AmountInfo.Amount,
			})
		}
		if len(orders) == 1 {
			// 发布"撤销拆单"操作事件
			s.bus.PublishCancelSplitOrderEvent(event.CancelSplitOrderPayload{
				BasePayload: event.BasePayload{ // 撤销拆单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		} else {
			// 发布"拆单"操作事件
			s.bus.PublishSplitOrderEvent(event.SplitOrderPayload{
				BasePayload: event.BasePayload{ // 拆单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
				Orders: orders,
			})
		}
	}()

	return shopCart, nil

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

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	firstSaleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)

	// 如果第一个销售订单已经结账，则提示"当前订单已结账，无法撤销"
	if firstSaleOrder.IsSettled() {
		return nil, errors.New("当前订单已结账，无法撤销")
	}

	// 判断订单是否已被部分支付
	if repository.NewOrderRepo(db).IsPartiallyPaid(request.SaleBillUuid) {
		return nil, errors.New("当前订单已被部分支付，不支持撤销拆单")
	}

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
			// NOTE 优化减少重复查询
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

	// 发布"撤销拆单"操作事件
	go func() {
		s.bus.PublishCancelSplitOrderEvent(event.CancelSplitOrderPayload{
			BasePayload: event.BasePayload{ // 撤销拆单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	}()

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}
	if request.MemberUuid > 0 {
		// 设置会员折扣
		member, errMember := repository.NewMemberRepo(db).GetMemberInfoForSaleOrder(ctx, request.MemberUuid)
		if errMember != nil {
			return nil, errors.WithMessage(errMember)
		}
		saleOrder.SetMemberDiscount(*member)
	}
	saleOrder.SetAllDiscountCancel()
	// 更新销售账单是否拆单的字段
	saleBill.SetIsSplitOrder(len(saleBill.SaleOrders)-1 > 1)

	// 重新计算销售订单金额
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
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

	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, request.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	saleOrder.SetMemberDiscountCancel()
	// 取消优惠券
	saleOrder.SetAllCouponCancel()
	if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
		return nil, errors.WithMessage(err, "取消销售订单会员优惠券失败")
	}

	// 取消会员余额支付
	if saleOrder.ConsumerUuid == 0 {
		memberBalancePayment := saleOrder.GetMemberBalancePayment()
		if memberBalancePayment != nil {
			memberBalancePayment.SetDelete()
			repository.NewPaymentOrderRepo(db).DeletePaymentOrderRecord(memberBalancePayment.Uuid)
		}
	}

	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return infoResp, nil
}

// OrderUseMember 使用会员优惠
func (s *orderSrv) OrderUseMember(ctx context.Context, request req.CheckMemberPasswordReq) (*resp.InstantOrderPaymentInfoResp, bool, error) {
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
		return nil, false, errMember
	}

	// 如果会员有密码的话，验证会员密码
	if member.HasPassword() {
		md5Password := cryptor.Md5String(request.Password)
		ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
		if member.Password != md5Password {
			ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
			return nil, false, errors.New("会员密码错误")
		}
	}

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, false, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, false, errors.New("销售订单不存在")
	}

	// 验证订单是否可操作
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderSettle, request.SaleOrderUuid); err != nil {
		return nil, false, errors.WithMessage(err)
	}

	// 判断订单是否进行了整单改价或抹零
	isCustomAmountAndZero := false
	if saleOrder.IsCustomAmount() || saleOrder.IsZeroRule() {
		isCustomAmountAndZero = true
	}

	saleOrder.SetMemberDiscount(*member)
	// 取消优惠券
	saleOrder.SetAllCouponCancel()
	if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
		return nil, false, errors.WithMessage(err, "取消销售订单会员优惠券失败")
	}

	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, false, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, false, errors.WithMessage(err)
	}
	return infoResp, isCustomAmountAndZero, nil
}

// OrderPrint 打印
func (s *orderSrv) OrderPrint(ctx context.Context, request req.OrderPrintReq, needLock bool) (*resp.PrinterData, error) {
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

	// 当不是收银端的时候，拆单不可操作结账
	if ctx.GetSource() != constant.SourceCashier && saleBill.IsSplit() {
		return nil, errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 更新收银员信息。当收银员信息发生变化时，才更新。且只更新收银员信息相关字段
	needUpdateCashier := false
	if saleOrder.CashierUuid != ctx.GetStaffUuid() {
		saleOrder.CashierUuid = ctx.GetStaffUuid()
		needUpdateCashier = true
	}
	staff := ctx.GetStaff()
	if saleOrder.CashierName != staff.GetUserName() {
		saleOrder.CashierName = staff.GetUserName()
		needUpdateCashier = true
	}
	if needUpdateCashier {
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderCashier(ctx, saleOrder.Uuid, saleOrder.CashierUuid, saleOrder.CashierName); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 判断是否已支付
	printType := constant.PrinterTemplatePreBilling
	if saleOrder.IsPaid() {
		printType = constant.PrinterTemplateBilling
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, request.PrintLang).PrintingStatementOrder(
		printType,
		saleBill,
		saleOrder.Uuid,
		utils.IfInt(ctx.GetSource() == constant.SourceAssistant, 0, 1),
		request.PayMethodUuid,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 保存销售账单
	if needLock && !saleOrder.IsPaid() {
		if err := repository.NewOrderRepo(db).SetLock(saleBill.Uuid, true); err != nil {
			return nil, errors.WithMessage(err, "设置锁单失败")
		}
		// 推送桌台更新
		go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
			"desk_uuid":   saleBill.DeskUuid,
			"update_time": time.Now().Unix(),
		})
	}

	// 如果是点餐助手，不能直接打印
	if ctx.GetSource() == constant.SourceAssistant {
		return &resp.PrinterData{}, nil
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

	// 更新收银员信息。当收银员信息发生变化时，才更新。且只更新收银员信息相关字段
	needUpdateCashier := false
	if saleOrder.CashierUuid != ctx.GetStaffUuid() {
		saleOrder.CashierUuid = ctx.GetStaffUuid()
		needUpdateCashier = true
	}
	staff := ctx.GetStaff()
	if saleOrder.CashierName != staff.GetUserName() {
		saleOrder.CashierName = staff.GetUserName()
		needUpdateCashier = true
	}
	if needUpdateCashier {
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderCashier(ctx, saleOrder.Uuid, saleOrder.CashierUuid, saleOrder.CashierName); err != nil {
			return nil, errors.WithMessage(err)
		}
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
		// 更新打印次数
		saleOrder.InvoiceInfo.PrintNum = saleOrder.InvoiceInfo.PrintNum + 1
		repository.NewOrderRepo(db).SaveOrUpdateInvoiceInfo(saleOrder.Uuid, model.SaleOrderInvoiceInfo{
			SaleOrderUuid: saleOrder.Uuid,
			PrintNum:      saleOrder.InvoiceInfo.PrintNum,
		})
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, req.PrintLang).PrintingInvoice(
		saleBill,
		saleOrder.Uuid,
		utils.IfInt(ctx.GetSource() == constant.SourceAssistant, 0, 1),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 如果是点餐助手，不能直接打印
	if ctx.GetSource() == constant.SourceAssistant {
		return &resp.PrinterData{}, nil
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
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUnlock, 0); err != nil {
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

	// 推送桌台更新
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
		"desk_uuid":   saleBill.DeskUuid,
		"update_time": time.Now().Unix(),
	})

	return nil
}

// GetMustPlanList 点餐助手、平板端获取必点商品方案列表
func (s *orderSrv) GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error) {
	saleBillRepo := repository.NewSaleBillRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.ErrInternal
	}
	list, err := s.mustPlanSrv.GetDeskMustPlanList(ctx, saleBill.MealNum, make(map[uint64]map[uint64]float64), saleBill.DeskUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.WithMessage(err)
	}
	return resp.ProductMustPlanList{
		List: list,
	}, nil
}

// GetUnOrderedH5ProductList 获取扫码h5购物车未下单商品列表
func (s *orderSrv) GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error) {
	res, err := s.GetUnsentKitchen(ctx, saleBillUuid, shopCart, opts...)
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
func (s *orderSrv) GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error) {
	if shopCart == nil {
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取点餐购物车信息"), err.Error())
		}
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
			AcceptTime:         products[0].AcceptTime,
			IsAccept:           products[0].IsAccept,
			H5OrderTime:        products[0].H5OrderTime,
			IsH5OrderNeedAudit: products[0].IsH5OrderNeedAudit,
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
func (s *orderSrv) ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64, ignoreMust bool) (any, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	res := make(map[string]any)
	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return res, errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderH5Confirm, saleOrderUuid); err != nil {
		return res, errors.WithMessage(err)
	}

	// h5下单限制
	h5Setting, _ := s.settingSrv.GetH5Setting(ctx, []dto.LanguageItem{})
	if h5Setting.IsBuffetOrderLimit == "1" || h5Setting.IsOrderLimit == "1" {
		// 读取上一单下单时间
		h5Repo := repository.NewH5OrderRepo(db)
		lastH5Order, _ := h5Repo.GetH5Order(h5Repo.WhereSaleBillUuid(saleBillUuid))
		var lastOrderTime int64 = 0
		if lastH5Order != nil {
			lastOrderTime = lastH5Order.CreateTime
		}
		if h5Setting.IsBuffetOrderLimit == "1" && saleBill.IsBuffetSaleBill() { // 自助餐下单限制
			if h5Setting.BuffetOrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(h5Setting.BuffetOrderLimit.LimitTime)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					return gin.H{"value": nextTime - now}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if h5Setting.BuffetOrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(h5Setting.BuffetOrderLimit.LimitNum)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				if saleBill.GetUnOrderH5OrderProductNum() > float64(numLimit) {
					return gin.H{"value": numLimit}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
		if h5Setting.IsOrderLimit == "1" && !saleBill.IsBuffetSaleBill() { // 非自助餐下单限制
			if h5Setting.OrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(h5Setting.OrderLimit.LimitTime)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					return gin.H{"value": nextTime - now}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if h5Setting.OrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(h5Setting.OrderLimit.LimitNum)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				if saleBill.GetUnOrderH5OrderProductNum() > float64(numLimit) {
					return gin.H{"value": numLimit}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
	}

	h5Order, err := saleBill.NewH5Order(ctx.GetCompanySetting())
	if err != nil {
		return res, errors.WithMessage(err)
	}
	// 获取未下单的h5订单商品
	h5OrderProducts := saleBill.GetUnOrderH5OrderProduct()
	subProductList := make([]*model.SaleOrderProduct, 0)
	// 将未下单的h5订单商品变为已下单的h5订单商品
	for _, h5OrderProduct := range h5OrderProducts {
		h5OrderProduct.SetH5OrderProduct(h5Order.Uuid)
		// 如果商品是套餐商品，则设置套餐子商品变为已下单的h5订单商品
		if h5OrderProduct.IsPackageProduct() {
			subProducts := saleBill.GetSubProducts(h5OrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.SetH5OrderProduct(h5Order.Uuid)
			}
			subProductList = append(subProductList, subProducts...)
		}
	}

	// 检查超时不能加购
	if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, h5OrderProducts); err != nil {
		return res, errors.WithMessage(err)
	}

	// 只检查本次下单的商品是否超过限购
	uuids := make([]uint64, 0)
	for _, h5OrderProduct := range h5OrderProducts {
		uuids = append(uuids, h5OrderProduct.Uuid)
	}

	// 检查必点
	saleOrderProductAll := saleBill.GetSaleOrderProductAll(model.WithH5CheckLimit())
	checkServiceRes, errCheck := s.checkOrder(ctx, ignoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithIsH5Check(), WithSaleOrderProductUuid(uuids...))
	if errCheck != nil {
		ctx.Log().Error("检查商品失败", zap.Error(errCheck))
		return nil, errors.New("检查商品失败")
	}
	if checkServiceRes != nil {
		if checkServiceRes.Code == constant.CodeOrderCheckProductMust && ignoreMust {
			// 必点方案未选择，且忽略必点方案
		} else {
			return checkServiceRes.OrderCheckRes, errors.WithMessage(errors.New(constant.ParseCodeOrderCheck(checkServiceRes.Code, constant.WithIsH5())))
		}
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
		// 更新销售订单商品，将套餐子商品改为已下单商品。记录上saleOrderProduct的h5OrderUuid
		for _, saleOrderProduct := range subProductList {
			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err, "更新销售订单商品失败")
			}
		}
		return nil
	}); err != nil {
		return res, errors.WithMessage(err, "下单扫码h5订单失败")
	}

	// 自动接单
	{
		companySetting := ctx.GetCompanySetting()
		if companySetting.GetIsOpenH5Order() {
			// 判断
			acceptOrderSetting, err := s.settingSrv.GetAcceptOrderSetting(ctx)
			if err != nil {
				ctx.Log().Error("获取接单设置失败", zap.Error(err))
			}
			totalPrice := saleBill.GetUnAcceptH5OrderProductTotalPrice(h5OrderProducts) // 未接单的h5订单商品的商品金额之和
			if acceptOrderSetting.CanAutoOrder(totalPrice) {
				// 自动接单
				s.AcceptH5Order(ctx, h5Order.Uuid, true)
			}
		} else {
			// 如果关闭h5接单功能
			// 自动接单
			s.AcceptH5Order(ctx, h5Order.Uuid, true)
		}
	}

	return res, nil
}

func (s *orderSrv) AcceptH5Order(ctx context.Context, h5OrderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	companySetting := ctx.GetCompanySetting()
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	h5Order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid, companySetting.GetIsOpenH5Order())
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取h5订单失败"), err.Error())
	}
	if h5Order.Status != constant.H5OrderStatusOrder {
		return nil, nil
	}

	if ctx.NoLock() {
		s.lock.LockUuid(h5Order.SaleOrder.SaleBillUuid)
		defer s.lock.UnlockUuid(h5Order.SaleOrder.SaleBillUuid)
		ctx.AddLock()
	}

	// 接单,保证h5订单的商品快照信息
	h5Order.Accept(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 将已下单的h5订单商品变为已接单单的h5订单商品
	h5Order.ChangeToAccepted()

	// 标记为是自动接单
	if isAutoOrder {
		h5Order.IsAutoAccept = 1
	}

	{
		ignoreMust := true // 接单，送厨忽略必点方案
		// 获取销售账单信息
		saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(h5Order.SaleOrder.SaleBillUuid)
		if errSaleBill != nil {
			return nil, errors.WithMessage(errSaleBill, "repository.NewOrderRepo(db).GetSaleBillAllInfo")
		}
		ctx.Log().Debug("获取销售账单信息")

		// 获取本次接单的商品列表
		unCookingSaleOrderProducts := h5Order.SaleOrderProducts

		// 检查超时不能加购
		if !isAutoOrder {
			if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, unCookingSaleOrderProducts); err != nil {
				return nil, errors.WithMessage(err)
			}
		}

		// 将h5订单商品插入到销售订单中
		saleOrder := saleBill.GetFirstSaleOrder()
		saleOrder.InsertSaleOrderProduct(unCookingSaleOrderProducts)

		// 送厨
		checkServiceRes, err := s.ActionCooking(ctx, ignoreMust, saleBill, unCookingSaleOrderProducts, h5OrderUuid, isAutoOrder, withCalcAndSaveSaleBill()) // 接单
		if err != nil {
			return nil, errors.WithMessage(err, "ActionCooking")
		}
		if checkServiceRes != nil {
			return checkServiceRes, nil
		}
	}
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*h5Order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 更新销售订单商品.将该h5订单的商品变为已接单
		// for _, saleOrderProduct := range h5Order.SaleOrderProducts {
		// 	if err := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
		// 		return errors.WithMessage(err, "将已下单的h5订单商品变为已接单单的h5订单商品失败")
		// 	}
		// }

		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "接单失败")
	}
	return nil, nil
}

func (s *orderSrv) RejectH5Order(ctx context.Context, h5OrderUuid uint64) error {
	db := ctx.GetDB()
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	h5Order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid, true)
	if err != nil {
		if builtinerrors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.WithMessage(errors.New("获取h5订单失败"), err.Error())
	}
	// 非待处理状态不可操作
	if h5Order.Status != constant.H5OrderStatusOrder {
		return errors.WithMessage(errors.New("当前状态不可操作"))
	}

	// 拒单,保证h5订单的商品快照信息
	h5Order.Reject(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 删除销售订单商品
	h5Order.DeleteSaleOrderProduct()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*h5Order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 删除销售订单商品.将该h5订单的商品删除
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductList(h5Order.SaleOrderProducts); err != nil {
			return errors.WithMessage(err, "删除销售订单商品失败")
		}

		// 发布"拒单"操作事件
		go func() {
			s.bus.PublishRejectH5OrderEvent(event.RejectH5OrderPayload{
				BasePayload: event.BasePayload{ // 拒单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: h5Order.SaleBillUuid,
					H5OrderUuid:  h5OrderUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		}()
		return nil
	}); err != nil {
		return errors.WithMessage(err, "拒单失败")
	}
	return nil
}

// GetUnsentKitchen 未送厨商品列表
func (s *orderSrv) GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error) {
	// 初始化返回结果
	res := resp.UnsentKitchen{
		Products:   resp.CartProductList{List: make([]resp.Product, 0)},
		AmountInfo: resp.SimpleAmountInfo{},
	}

	if shopCart == nil {
		var err error
		// 获取购物车信息
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
		if err != nil {
			return res, errors.WithMessage(err, "获取点餐购物车信息: "+err.Error())
		}
	}

	// 按商品签名分组并合并相同商品
	signProduct := make(map[string]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			// 只处理未送厨且非赠菜的商品
			if product.Status != constant.SaleOrderProductStatusNormal {
				continue
			}
			if product.IsGift {
				product.DiscountPrice = 0
			}
			res.AmountInfo.ProductNum += product.Num
			// 合并相同商品的数量和折扣价格
			if p, exists := signProduct[product.Sign]; exists {
				product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
				product.Num = p.Num + product.Num
			}
			signProduct[product.Sign] = product
		}
	}

	// 将合并后的商品添加到结果列表
	res.Products.List = make([]resp.Product, 0, len(signProduct))
	for _, product := range signProduct {
		res.Products.List = append(res.Products.List, product)
	}

	// 按送厨时间排序
	sort.Slice(res.Products.List, func(i, j int) bool {
		if res.Products.List[i].CreateTime == res.Products.List[j].CreateTime {
			return res.Products.List[i].Uuid < res.Products.List[j].Uuid
		}
		return res.Products.List[i].CreateTime < res.Products.List[j].CreateTime
	})

	// 获取销售账单信息并计算未送厨商品总金额
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return res, errors.WithMessage(errors.New("获取销售账单所有信息"), err.Error())
	}

	for _, order := range saleBill.SaleOrders {
		res.AmountInfo.ProductAmount = utils.DecimalAdd(res.AmountInfo.ProductAmount, order.GetUnCookingProductAmount())
	}

	return res, nil
}

// GetSentKitchen 已送厨商品列表
func (s *orderSrv) GetSentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart) (resp.SentKitchen, error) {
	if shopCart == nil {
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid)
		if err != nil {
			return resp.SentKitchen{}, errors.WithMessage(err, "获取点餐购物车信息: "+err.Error())
		}
	}

	// 统计已送厨商品并按送厨时间分组
	productNum := float64(0)
	productGroup := make(map[int64]map[string]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			if product.Status != constant.SaleOrderProductStatusCooking {
				continue
			}

			productNum += product.Num

			// 按送厨时间分组,并在组内按商品签名合并
			if _, exists := productGroup[product.SendKitchenTime]; !exists {
				productGroup[product.SendKitchenTime] = make(map[string]resp.Product)
			}

			if p, exists := productGroup[product.SendKitchenTime][product.Sign]; exists {
				product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
				product.Num = p.Num + product.Num
			}
			productGroup[product.SendKitchenTime][product.Sign] = product
		}
	}

	// 构建分组列表
	groups := make([]resp.SentKitchenProductGroup, 0, len(productGroup))
	for sendKitchenTime, signProducts := range productGroup {
		products := make([]resp.Product, 0, len(signProducts))
		for _, product := range signProducts {
			products = append(products, product)
		}

		groups = append(groups, resp.SentKitchenProductGroup{
			SendKitchenTime: sendKitchenTime,
			Products: resp.GroupProductList{
				List: products,
			},
		})
	}

	// 按送厨时间倒序排序
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SendKitchenTime > groups[j].SendKitchenTime
	})

	// 获取销售账单并计算金额
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return resp.SentKitchen{}, errors.WithMessage(errors.New("获取销售账单所有信息"), err.Error())
	}

	amount := resp.AmountInfo{ProductNum: productNum}
	for _, order := range saleBill.SaleOrders {
		calc := order.CalcCookingSaleOrder(*saleBill.SaleBillSetting)
		amount.ProductOriginalAmount = utils.DecimalAdd(amount.ProductOriginalAmount, calc.ProductOriginalAmount)
		amount.ProductAmount = utils.DecimalAdd(amount.ProductAmount, calc.ProductAmount)
		amount.ServiceAmount = utils.DecimalAdd(amount.ServiceAmount, calc.ServiceFee)
		amount.TaxAmount = utils.DecimalAdd(amount.TaxAmount, calc.TaxFee)
		amount.DiscountAmount = utils.DecimalAdd(amount.DiscountAmount, calc.CustomDiscountFee)
		amount.MemberDiscountAmount = utils.DecimalAdd(amount.MemberDiscountAmount, calc.MemberDiscountFee)
		amount.Amount = utils.DecimalAdd(amount.Amount, utils.IfFloat64(order.CustomAmount >= 0, order.CustomAmount, calc.Amount))
	}

	return resp.SentKitchen{
		Groups:     resp.GroupList{List: groups},
		AmountInfo: amount,
	}, nil
}

// GetOrderMemberList 获取订单会员列表
func (s *orderSrv) GetOrderMemberList(ctx context.Context, saleBillUuid uint64) (resp.InstantOrderMemberList, error) {
	db := ctx.GetDB()
	//
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillInfoAndMember(saleBillUuid)
	if err != nil {
		return resp.InstantOrderMemberList{}, errors.WithMessage(errors.New("获取销售账单信息"), err.Error())
	}
	//
	list := make([]resp.InstantOrderMember, 0)
	uuidList := make([]uint64, 0)
	for _, v := range saleBill.SaleOrders {
		if v.Member == nil {
			continue
		}
		if slices.Contains(uuidList, v.Member.Uuid) {
			continue
		}
		uuidList = append(uuidList, v.Member.Uuid)
		list = append(list, resp.InstantOrderMember{
			Uuid:     v.Member.Uuid,
			Nickname: v.Member.Nickname,
			Phone:    v.Member.Phone,
		})
	}
	//
	extra := resp.InstantOrderMemberExtra{
		IsCheckout:        saleBill.IsExistPaid(),
		IsPartialCheckout: saleBill.IsExistPaid(),
	}
	if !saleBill.IsExistPaid() && repository.NewOrderRepo(db).IsPartiallyPaid(saleBillUuid) {
		extra.IsPartialCheckout = true
	}
	//
	return resp.InstantOrderMemberList{
		List:  list,
		Extra: extra,
	}, nil
}

// GetProductPackageDetail 获取商品包详情
func (s *orderSrv) GetProductPackageDetail(ctx context.Context, req req.GetProductPackageDetailReq) (*resp.ProductPackageDetailRes, error) {
	db := ctx.GetDB()
	// 获取销售订单中h5未下单的销售订单商品
	saleOrderProducts, err := repository.NewSaleOrderProductRepo(db).GetProductPackageDetail(req.SaleBillUuid, req.SaleOrderUuid, req.ProductPackageUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品包详情失败")
	}

	productPackageDetailList := make([]resp.ProductPackageDetail, 0)

	for _, saleOrderProduct := range saleOrderProducts {
		productPackageDetail := saleOrderProduct.GetProductPackageDetail()
		productPackageDetailList = append(productPackageDetailList, productPackageDetail)
	}

	return &resp.ProductPackageDetailRes{List: productPackageDetailList}, nil
}

func (s *orderSrv) GetOrderCartProductBatchCookingList(ctx context.Context, req req.GetOrderCartProductBatchCookingListReq) (*resp.OrderCartProductBatchCookingRes, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	batchCookingSaleOrderProducts := saleBill.GetSaleOrderProductBatchCooking()

	batchCookingSaleOrderProductsList := make([]resp.OrderCartProductBatchCooking, 0)
	for _, saleOrderProduct := range batchCookingSaleOrderProducts {
		baseURL := utils.GetBaseURL(ctx.GetGin().Request)
		batchCookingSaleOrderProductsList = append(batchCookingSaleOrderProductsList, resp.OrderCartProductBatchCooking{
			Uuid:       saleOrderProduct.Uuid,
			LocaleName: saleOrderProduct.MultiLanguageName.GetNames(),
			// LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			LocaleAttributeName: saleOrderProduct.GetFlavorName(),
			Image: func() string {
				if saleOrderProduct.ImageFile != nil {
					url := saleOrderProduct.ImageFile.GetUrl(baseURL)
					return url
				}
				return ""
			}(),
			BatchTagUuid:    saleOrderProduct.BatchTagUuid,
			BatchTime:       saleOrderProduct.BatchTime,
			SendKitchenTime: saleOrderProduct.SendKitchenTime,
			CreateTime:      saleOrderProduct.CreateTime,
		})
	}

	tagMap := make(map[uint64]int)
	for _, batchCookingSaleOrderProduct := range batchCookingSaleOrderProducts {
		tagMap[batchCookingSaleOrderProduct.BatchTagUuid]++
	}

	// 获取分批类型列表
	batchTags, errBatchTags := repository.NewBatchTagRepo(db).GetBatchTagList()
	if errBatchTags != nil {
		return nil, errors.WithMessage(errBatchTags)
	}
	batchCookingSaleOrderProductsTags := make([]resp.OrderCartProductBatchCookingTag, 0)
	for _, batchTag := range batchTags {
		batchCookingSaleOrderProductsTags = append(batchCookingSaleOrderProductsTags, resp.OrderCartProductBatchCookingTag{
			Uuid:       batchTag.Uuid,
			LocaleName: batchTag.MultiLanguageName.GetNames(),
			Color:      batchTag.Color,
			Sort:       uint(batchTag.Sort),
			Count:      uint(tagMap[batchTag.Uuid]),
		})
	}

	// 排序
	batchCookingSaleOrderProductsList = sortBatchCookingSaleOrderProducts(batchCookingSaleOrderProductsList)

	return &resp.OrderCartProductBatchCookingRes{List: batchCookingSaleOrderProductsList, Tags: batchCookingSaleOrderProductsTags}, nil
}

// 排序分批送厨弹出的商品列表。根据送厨时间分组，后送厨的排在前面。同一组内，下单时间早的排在前面。
func sortBatchCookingSaleOrderProducts(batchCookingSaleOrderProducts []resp.OrderCartProductBatchCooking) []resp.OrderCartProductBatchCooking {
	sort.Slice(batchCookingSaleOrderProducts, func(i, j int) bool {
		if batchCookingSaleOrderProducts[i].SendKitchenTime == batchCookingSaleOrderProducts[j].SendKitchenTime { // 同一组内，下单时间早的排在前面。
			return batchCookingSaleOrderProducts[i].CreateTime < batchCookingSaleOrderProducts[j].CreateTime
		}
		// 不同组内，后送厨的排在前面。
		return batchCookingSaleOrderProducts[i].SendKitchenTime > batchCookingSaleOrderProducts[j].SendKitchenTime
	})
	return batchCookingSaleOrderProducts
}

// 分批送厨
func (s *orderSrv) OrderCartProductBatchCooking(ctx context.Context, req req.OrderCartProductBatchCookingReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	// 获取saleBill中的分批商品
	saleBillProducts := saleBill.GetSaleOrderProductBatchCookingBySaleOrderUuid(req.SaleOrderProductUuids)

	// 标记分批送厨
	batchTime := time.Now().Unix()
	for _, saleBillProduct := range saleBillProducts {
		saleBillProduct.SetCookingBatch(req.BatchTagUuid, batchTime)
	}

	// 更新saleBill中的最新分批类型的颜色。（每次分批送厨后，都改变一次）
	saleBill.BatchTagUuid = req.BatchTagUuid

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新销售订单商品状态从预送厨变为已送厨
		if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(saleBillProducts); err != nil {
			return errors.WithMessage(err)
		}
		// 更新送厨单商品的batch_time、batch_tag_uuid
		// 将product_order_product中的分批商品标记为已送厨
		if err := repository.NewProductionRepo(tx).UpdateProductionOrderProductBatchTimeAndBatchTagUuid(req.SaleBillUuid, req.SaleOrderProductUuids, batchTime, req.BatchTagUuid); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售账单中的分批类型UUID
		if err := repository.NewSaleBillRepo(tx).UpdateSaleBillBatchTagUuid(req.SaleBillUuid, req.BatchTagUuid); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发起“预送厨”操作的事件
	go s.bus.PublishSentCookingPreEvent(event.SentCookingPrePayload{
		BasePayload: event.BasePayload{ // 送厨
			Ctx:           ctx,
			CompanyUuid:   ctx.GetCompanyUuid(),
			Source:        ctx.GetSource(),
			SaleBillUuid:  saleBill.Uuid,
			SaleOrderUuid: req.SaleOrderUuid,
			OperatorUuid:  int64(ctx.GetStaffUuid()),
		},
		Products: func() event.ProductsPre {
			products := make(event.ProductsPre, 0)
			for _, unCookingSaleOrderProduct := range saleBillProducts {
				unCookingSaleOrderProduct.BatchTagUuid = req.BatchTagUuid
				products = append(products, s.convertToEventOrderProductPre(
					unCookingSaleOrderProduct,
					saleBill,
				))
			}
			return products
		}(),
	})

	shopCart, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return shopCart, nil
}
