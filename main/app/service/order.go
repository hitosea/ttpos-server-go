package service

import (
	contexts "context"
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
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// IOrderSrv 定义订单服务接口
type IOrderSrv interface {
	IMemberOrderSrv

	// base
	CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error)                                                             // 创建点餐订单
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                                       // 创建桌台订单
	SetOrderSource(ctx context.Context, saleBillUuid uint64, orderSourceUuid uint64) (resp.ShopCart, error)                                  // 设置订单来源
	SetNationality(ctx context.Context, saleBillUuid uint64, nationalityUuid uint64) (resp.ShopCart, error)                                  // 设置订单国籍
	IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error)                                                      // 判断桌台是否可取消
	HideOrder(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error)                                                              // 挂单
	ShowOrder(ctx context.Context, req req.OrderShowReq) (*resp.ShopCart, error)                                                             // 显示订单
	InstantHideOrderList(ctx context.Context, req req.HideSaleBillListReq) (*resp.InstantHideOrderListResp, error)                           // 获取挂单订单列表
	OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error)                                                       // 打包
	OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error)                                     // 修改订单人数
	InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error)                         // 给销售订单创建一个销售订单
	SaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq, needDeleteSaleOrder bool) (*resp.ShopCart, error) // 从一个销售订单移动商品到另一个销售订单
	InstantOrderSaleOrderDelete(ctx context.Context, req req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error)                         // 删除一个销售订单(删除拆单)
	InstantOrderSaleOrderDeleteAll(ctx context.Context, req req.InstantOrderSaleOrderDeleteAllReq) (*resp.ShopCart, error)                   // 删除所有子销售订单(撤销拆单)
	OrderUnlock(ctx context.Context, saleBillUuid uint64) error                                                                              // 订单解锁
	GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error)                                                                         // 通过桌台uuid获取到销售账单信息
	GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error)                                                 // 通过设备SN获取点餐购物车信息
	GetSaleBillUuidAndSaleOrderUuid(ctx context.Context, deskUuid uint64) (uint64, uint64, error)                                            // 获取销售账单uuid和销售订单uuid
	GetProductPackageDetail(ctx context.Context, req req.GetProductPackageDetailReq) (*resp.ProductPackageDetailRes, error)                  // 获取商品选购详情

	// manage
	GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error)                                               // 获取订单列表
	ExportOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderExportListPaginationResp, error)                                      // 导出订单列表
	GetPaymentAmount(ctx context.Context, req req.OrderPaymentAmountReq) resp.GetPaymentAmountResp                                               // 获取实付金额
	GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error)                                                        // 获取订单详情
	GetRecordList(ctx context.Context, saleBillUuid uint64, h5OrderUuid uint64) ([]resp.OrderOperationLog, error)                                // 获取订单操作日志
	CancelOrder(ctx context.Context, req req.OrderCancelReq) error                                                                               // 取消订单
	DeleteOrder(ctx context.Context, dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error                                               // 删除订单
	ReturnOrder(ctx context.Context, req req.OrderReturnReq) (error, int)                                                                        // 退款订单
	ReReturnOrder(ctx context.Context, req req.OrderReReturnReq) (error, int)                                                                    // 重新退款
	GetReturnOrderInfo(ctx context.Context, req req.OrderReturnInfoReq) (*resp.OrderReturnInfoResp, error)                                       // 获取退款信息
	GetReverseSettleInfo(ctx context.Context, req req.OrderReverseSettleInfoReq) (*resp.OrderReverseSettleInfoResp, error)                       // 获取反结账信息
	ReverseSettle(ctx context.Context, req req.OrderReverseSettleReq) error                                                                      // 反结账
	OrderRemark(ctx context.Context, req req.OrderRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                 // 修改订单备注
	CreateSaleBillSetting(ctx context.Context, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, isMember bool) (*model.SaleBillSetting, error) // 创建销售账单设置
	CheckAuthorization(ctx context.Context, operationType string) (bool, error)                                                                  // 检查授权（折扣操作）
	VerifyPassword(ctx context.Context, req req.VerifyPasswordForSensitiveOperationReq) (bool, error)                                            // 密码验证（根据operation_type选择折扣操作或退款操作）

	// product
	OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error)              // 删除订单商品
	OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                // 修改订单商品备注
	OrderCartProductAdd(ctx context.Context, request req.ProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                   // 修改购物车商品数量
	OrderCartProductNum(ctx context.Context, req req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)              // 修改购物车商品数量
	AssistantOrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) // 修改购物车商品数量
	InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, *resp.OrderCheckServiceRes, error)               // 送厨购物车商品
	InstantOrderCartProductReturning(ctx context.Context, req req.OrderCartProductReturningReq) (*resp.ShopCart, error)                                       // 退菜购物车商品
	InstantOrderCartProductCancelReturning(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                             // 退菜购物车商品
	InstantOrderCartProductChangeDesk(ctx context.Context, req req.OrderCartProductChangeDeskReq) (*resp.ShopCart, error)                                     // 转菜购物车商品
	OrderCartProductWrap(ctx context.Context, req req.OrderCartProductWrapReq) (*resp.ShopCart, error)                                                        // 打包购物车商品
	OrderCartProductUnwrap(ctx context.Context, req req.OrderCartProductUnwrapReq) (*resp.ShopCart, error)                                                    // 取消打包购物车商品
	InstantOrderCartProductGiving(ctx context.Context, req req.OrderCartProductGivingReq) (*resp.ShopCart, error)                                             // 取消赠菜购物车商品
	InstantOrderCartProductCancelGiving(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error)                                                // 取消赠菜购物车商品
	GetOrderCartInfo(ctx context.Context, saleOrderUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)                           // 获取购物车信息
	OrderCartProductPackageAdd(ctx context.Context, request req.OrderCartProductPackageAddReq) (*resp.ShopCart, error)                                        // 向购物车添加套餐
	OrderCartProductFlavorAndAttribute(ctx context.Context, request req.OrderCartProductFlavorAndAttributeReq) (*resp.ProductFlavorAndAttributeRes, error)    // 查询购物车商品“规格/属性”
	OrderCartProductFlavorAndAttributeChange(ctx context.Context, request req.OrderCartProductFlavorAndAttributeChangeReq) (*resp.ShopCart, error)            // 修改购物车商品“规格/属性”
	InstantOrderCartProductAdd(ctx context.Context, request req.OrderCartProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error)   // 向购物车添加商品
	ChangeBatchTag(ctx context.Context, req req.ChangeBatchTagReq) (*resp.ShopCart, error)                                                                    // 更换分批类型（前置模式）

	// cooking
	InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, bool, error)                                                                                                                                                  // 获取点餐必点方案
	GetProductNameByItemCode(ctx context.Context, itemCode string, saleOrderUuid uint64) ([]ProductInfo, error)                                                                                                                                                 // 通过itemCode获取订单中库存不足的商品名称
	InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq, opts ...func(option *MustPlanConfirmOption)) (bool, *resp.InstantProductMustPlan, error)                                                                           // 确认必点商品
	OrderCheck(ctx context.Context, req req.InstantOrderCheckReq) (*resp.OrderCheckServiceRes, error)                                                                                                                                                           // 订单检查
	CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error                                                                                                                            // 计算并保存销售账单
	GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error)                                                                                                                                                                 // 必点方案列表                                                                                                                                                                                                       // 拒单商家的所有待接单h5订单
	GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error)                                                                                                 // 未送厨商品列表
	GetSentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart) (resp.SentKitchen, error)                                                                                                                                                 // 已送厨商品列表
	ActionCooking(ctx context.Context, ignoreMust bool, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct, h5OrderUuid uint64, isAutoOrder bool, options ...func(option *ActionCookingOption)) (*resp.OrderCheckServiceRes, error) // 送厨
	ActionAddAndCooking(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill, IgnoreMust bool) (*resp.OrderCheckServiceRes, error)                                                                                                          // 加购并送厨
	TabletAddAndCooking(ctx context.Context, request req.TabletOrderCartProductAddReq) (*TabletAddAndCookingRes, error)                                                                                                                                         // 平板端加购并送厨
	GetOrderCartProductBatchCookingList(ctx context.Context, req req.GetOrderCartProductBatchCookingListReq) (*resp.OrderCartProductBatchCookingRes, error)                                                                                                     // 获取分批送厨弹框的销售订单商品列表
	OrderCartProductBatchCooking(ctx context.Context, req req.OrderCartProductBatchCookingReq) (*resp.ShopCart, error)                                                                                                                                          // 分批送厨

	// discount
	OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error) // 修改订单商品价格
	OrderAmountChange(ctx context.Context, req req.OrderAmountChangeReq) (*resp.ShopCart, error)             // 修改订单金额
	OrderDiscount(ctx context.Context, req req.OrderDiscountReq) (*resp.ShopCart, error)                     // 修改订单折扣
	OrderZeroRule(ctx context.Context, req req.OrderZeroRuleReq) (*resp.ShopCart, error)                     // 修改订单抹零规则
	OrderDiscountCancel(ctx context.Context, req req.OrderDiscountCancelReq) (*resp.ShopCart, error)         // 取消点餐订单所有优惠折扣，包括改价、打折、抹零

	// buffet
	GetOrderChangeBuffet(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) (resp.OrderBuffetResp, error)        // 自助餐信息
	OrderChangeBuffet(ctx context.Context, req req.OrderChangeBuffetReq) (*resp.ShopCart, error)                              // 调整自助餐
	OrderChangeBuffetClock(ctx context.Context, req req.OrderChangeBuffetClockReq) (*resp.ShopCart, error)                    // 调整自助餐
	OrderDeskBuffetProductList(ctx context.Context, req req.OrderChangeBuffetProductListReq) (*resp.BuffetProductList, error) // 获取桌台的自助餐商品列表

	// pay
	OrderPaymentCoupon(ctx context.Context, req req.InstantOrderPaymentCouponReq) (*resp.InstantOrderPaymentInfoResp, error)                                     // 使用优惠券 或 取消优惠券
	OrderPaymentPoints(ctx context.Context, req req.InstantOrderPaymentPointsReq) (*resp.InstantOrderPaymentInfoResp, error)                                     // 设置订单的抵扣积分数量
	OrderPaymentActivity(ctx context.Context, req req.InstantOrderPaymentActivityReq) (*resp.InstantOrderPaymentInfoResp, error)                                 // 选择或取消满减活动
	InstantOrderPaymentQrcode(ctx context.Context, req req.InstantOrderPaymentQrcodeReq) (*resp.InstantOrderPaymentQrcodeInfoResp, error)                        // 获取支付二维码
	InstantOrderPaymentCreate(ctx context.Context, req req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error)                              // 给销售订单创建一个支付单
	InstantOrderPaymentCancel(ctx context.Context, req req.InstantOrderPaymentCancelReq) (*resp.InstantOrderPaymentInfoResp, error)                              // 撤销一个支付单
	InstantOrderPaymentFinish(ctx context.Context, req req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error)                                          // 给销售订单创建一个支付单
	InstantOrderFree(ctx context.Context, req req.InstantOrderFreeReq) (*resp.OrderFinishResp, error)                                                            // 免单
	InstantOrderPaymentZeroRule(ctx context.Context, req req.InstantOrderPaymentZeroRuleReq) (*resp.InstantOrderPaymentInfoResp, error)                          // 设置结账抹零规则
	InstantOrderPaymentInfo(ctx context.Context, saleBill *model.SaleBill, saleBillUuid uint64, saleOrderUuid uint64) (*resp.InstantOrderPaymentInfoResp, error) // 获取结账页面信息

	// h5
	GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error)   // 获取扫码h5购物车未下单商品列表
	GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error) // 获取扫码h5购物车已下单商品列表
	ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64, ignoreMust bool) (any, error)                                                            // 下单扫码h5订单
	AcceptH5Order(ctx context.Context, h5OrderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error)                                                            // 接单扫码h5订单
	RejectH5Order(ctx context.Context, h5OrderUuid uint64) error                                                                                                            // 拒单扫码h5订单
	RejectAllH5Order(ctx context.Context, saleBillUuid uint64) error                                                                                                        // 拒绝某销售账单的所有未接单h5订单
	RejectAllH5OrderInShop(ctx context.Context) error                                                                                                                       // 将商家的所有待接单的h5订单都拒单

	// member
	OrderMemberCancel(ctx context.Context, req req.OrderMemberCancelReq) (*resp.InstantOrderPaymentInfoResp, error)      // 取消使用会员优惠
	OrderUseMember(ctx context.Context, req req.CheckMemberPasswordReq) (*resp.InstantOrderPaymentInfoResp, bool, error) // 使用会员优惠
	GetOrderMemberList(ctx context.Context, saleBillUuid uint64) (resp.InstantOrderMemberList, error)                    // 获取订单会员列表

	// print
	OrderPrint(ctx context.Context, req req.OrderPrintReq, needLock bool) (*resp.PrinterData, error)  // 打印
	OrderPrintInvoice(ctx context.Context, req req.OrderPrintInvoiceReq) (*resp.PrinterData, error)   // 图片打印
	OrderPrintInvoiceInfo(ctx context.Context, req req.OrderInvoiceInfoReq) resp.SaleOrderInvoiceInfo // 图片打印

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
func createMemberSaleOrder(_ context.Context, db *gorm.DB, params model.CreateMemberSaleOrderParams) (*model.MemberSaleOrder, error) {
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

	// 分批送厨模式（从业务设置中读取并保存快照）
	batchCookingMode := businessSetting.BatchCookingMode
	if batchCookingMode == "" {
		batchCookingMode = constant.BatchCookingModePost // 默认值
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
		BatchCookingMode:   batchCookingMode,
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
		Source:              constant.MapJwtSourceToSaleBillSource(ctx.GetSource()),
		ClientVersion:       constant.NormalizeClientVersion(ctx.GetVersion()),
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
			LocaleName:           saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
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
	utils.Go(func() {
		event.NewSystemBus().PublishChangeStockEvent(event.ChangeStockPayload{
			BasePayload: event.BasePayload{ // 库存变更
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})

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
func (s *orderSrv) orderProductDelete(ctx context.Context, dbId uint64, _ uint64, _ string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) {

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
	for i := range newSaleBill.SaleOrders {
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

	produtcDetail := formatProducts[0]

	// 注入商品规格和小料的库存
	bomUuids := s.collectProductBomUuids(produtcDetail)
	stockNumMap := s.getProductBomStockNumMap(ctx, bomUuids, productPackageUuid)
	// 为商品注入库存（包括规格和小料）
	product_resp.ProductSlice([]product_resp.Product{produtcDetail}).InjectStockNum(stockNumMap)

	return produtcDetail, nil
}

func EditProduct(ctx context.Context, db *gorm.DB, saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct, request req.EditProductReq, productBomStockNum func(productBomUuid uint64) float64) (*model.SaleOrderProduct, error) {
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
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.ToJson(),
			Price:          flavorProductBom.Price,
			ProductBomUuid: request.FlavorUuid,
			ErpCode:        flavorProductBom.ErpCode,
		})
		// 设置规格名称快照（JSON）
		// Requirement: story-main-product-attribute-snapshot-fix
		if !flavorProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
			flavor.ProductBom = *flavorProductBom
			if err := flavor.SetNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
				ctx.Log().Error("设置规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", request.FlavorUuid))
			}
		}
		flavor.SetUpdate()
		saleOrderProduct.SaleOrderProductBoms = append(saleOrderProduct.SaleOrderProductBoms, flavor)
		saleOrderProduct.ChangeFlavor(flavor)

		// 添加新加料
		sauceProductBoms, errSauceProductBoms := GetSauceInfo(ctx, db, request.SauceUuidList, saleOrderProduct.Num, productBomStockNum)
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
			sauceObj := model.NewSaleOrderProductSauce(saleOrderProduct.Uuid, saleOrder.Uuid, sauce)
			// 设置小料名称快照（JSON）
			// Requirement: story-main-product-attribute-snapshot-fix
			if sauceProductBom, ok := sauceProductBoms[sauce.ProductBomUuid]; ok && !sauceProductBom.ProductSauce.MultiLanguageName.IsNullName() {
				sauceObj.ProductBom = *sauceProductBom
				if err := sauceObj.SetNameSnapshot(sauceProductBom.ProductSauce.MultiLanguageName); err != nil {
					ctx.Log().Error("设置小料名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", sauce.ProductBomUuid))
				}
			}
			sauceObj.SetUpdate()
			saleOrderProduct.SaleOrderProductBoms = append(saleOrderProduct.SaleOrderProductBoms, sauceObj)
		}
		// 添加新属性
		productAttributes, errProductAttributes := GetAttributeInfo(ctx, db, request.AttributeUuidList)
		if errProductAttributes != nil {
			return nil, errors.WithMessage(errProductAttributes)
		}
		attributes := sortProductAttributes(ctx, productAttributes)
		for _, attribute := range attributes {
			attr := model.NewSaleOrderProductAttribute(saleOrderProduct.Uuid, saleOrder.Uuid, attribute)
			// 设置属性名称快照（JSON）
			// Requirement: story-main-product-attribute-snapshot-fix
			if productPackageAttribute, ok := productAttributes[attribute.ProductPackageAttributeUuid]; ok && !productPackageAttribute.Attribute.MultiLanguageName.IsNullName() {
				attr.ProductAttribute = productPackageAttribute.Attribute
				if err := attr.SetNameSnapshot(productPackageAttribute.Attribute.MultiLanguageName); err != nil {
					ctx.Log().Error("设置属性名称快照失败", zap.Error(err), zap.Uint64("product_attribute_uuid", attribute.ProductAttributeUuid))
				}
			}
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
func GetSauceInfo(ctx context.Context, db *gorm.DB, sauceProductBomUuidList []uint64, productNum float64, productBomStockNum func(productBomUuid uint64) float64) (map[uint64]*model.ProductBom, error) {
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
			if bom.GetStockNum(productBomStockNum) < productNum {
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
	IsTabletAddAndCooking  bool    // 是否是平台的加购并送厨
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

	// 创建库存服务实例（用于库存检查）
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	productBomStockNum := s.createProductBomStockNumFunc(ctx, appService)

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
		if flavorProductBom.GetStockNum(productBomStockNum) < float64(product.Num) {
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
					return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
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
				if bom.GetStockNum(productBomStockNum) < product.Num {
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
		isBatch := func() uint8 {
			if businessSetting.OpenIsBatch() {
				if productPackage.IsBatchBool() {
					return 1
				}
			}
			return 0
		}()

		// 前置模式下，处理分批类型UUID（使用快照值）
		batchTagUuid := uint64(0)
		batchCookingMode := params.Setting.BatchCookingMode
		if batchCookingMode == "" {
			batchCookingMode = constant.BatchCookingModePost // 默认值
		}
		if batchCookingMode == constant.BatchCookingModePre && isBatch == 1 {
			// 必须指定分批类型
			if product.BatchTagUuid == 0 {
				return nil, errors.WithMessage(errors.New("请选择分批类型再加购"))
			}
			batchTagRepo := repository.NewBatchTagRepo(db)
			if product.BatchTagUuid > 0 {
				// 验证 batch_tag_uuid 的有效性
				_, err := batchTagRepo.GetBatchTagInfo(product.BatchTagUuid)
				if err != nil {
					return nil, errors.WithMessage(fmt.Errorf("分批类型不存在"), err.Error())
				}
				batchTagUuid = product.BatchTagUuid
			} else {
				// 如果未提供，使用默认分批类型（排序第一的类型）
				// batchTags, err := batchTagRepo.GetBatchTagList()
				// if err != nil {
				// 	return nil, errors.WithMessage(err)
				// }
				// if len(batchTags) > 0 {
				// 	// 按 sort 排序，获取排序第一的类型
				// 	sort.Slice(batchTags, func(i, j int) bool {
				// 		return batchTags[i].Sort < batchTags[j].Sort
				// 	})
				// 	batchTagUuid = batchTags[0].Uuid
				// }
			}
		}

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
			CopyNum:                product.Num,
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
				Name:           flavorProductBom.ProductFlavor.MultiLanguageName.ToJson(), // 填顾客下单时规格的名字
				Price:          flavorPrice,
				ProductBomUuid: product.FlavorProductBomUuid,
				ErpCode:        flavorProductBom.ErpCode,
			},
			Attribute:     attributes,
			IsAcceptOrder: uint(isAcceptOrder),
			Remark:        product.Remark,
			IsBatch:       isBatch,
			BatchTagUuid:  batchTagUuid,
		}, &productPackage, product.Operation)
		saleOrderProduct.SetUnitNum(1) // 设置默认单位数量为1
		saleOrderProduct.Num = decimal.NewFromFloat(saleOrderProduct.GetUnitNum()).Mul(decimal.NewFromFloat(saleOrderProduct.CopyNum)).Round(4).InexactFloat64()

		// 设置规格、小料和属性的快照（JSON）
		// Requirement: story-main-product-attribute-snapshot-fix
		// 设置规格名称快照
		if !flavorProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
			for _, bom := range saleOrderProduct.SaleOrderProductBoms {
				if bom.IsFlavor() {
					bom.ProductBom = *flavorProductBom
					if err := bom.SetNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
						ctx.Log().Error("设置规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", flavorProductBom.Uuid))
					}
					// 同时更新 SaleOrderProduct.FlavorName
					if err := saleOrderProduct.SetFlavorNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
						ctx.Log().Error("设置商品规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", flavorProductBom.Uuid))
					}
					break
				}
			}
		}
		// 设置小料名称快照
		for _, bom := range saleOrderProduct.SaleOrderProductBoms {
			if !bom.IsFlavor() {
				if sauceProductBom, ok := sauceProductBoms[bom.ProductBomUuid]; ok && !sauceProductBom.ProductSauce.MultiLanguageName.IsNullName() {
					bom.ProductBom = *sauceProductBom
					if err := bom.SetNameSnapshot(sauceProductBom.ProductSauce.MultiLanguageName); err != nil {
						ctx.Log().Error("设置小料名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", bom.ProductBomUuid))
					}
				}
			}
		}
		// 设置属性名称快照
		for _, attr := range saleOrderProduct.SaleOrderProductAttributes {
			if productPackageAttribute, ok := productAttributes[attr.ProductPackageAttributeUuid]; ok && !productPackageAttribute.Attribute.MultiLanguageName.IsNullName() {
				attr.ProductAttribute = productPackageAttribute.Attribute
				if err := attr.SetNameSnapshot(productPackageAttribute.Attribute.MultiLanguageName); err != nil {
					ctx.Log().Error("设置属性名称快照失败", zap.Error(err), zap.Uint64("product_attribute_uuid", attr.ProductAttributeUuid))
				}
			}
		}

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
				innerParams.IsTabletAddAndCooking = true
				subProducts, err := s.newPackageSubProducts(ctx, product.GetSubProducts(), innerParams, params, saleOrderProduct.Uuid, saleOrderProduct.DeductStockType, product.Num)
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, subProducts...)
				saleOrderProducts = append(saleOrderProducts, subProducts...)

				// 计算套餐主商品的加价总和 = Σ(子商品每份的加价 × 子商品份数)
				totalAddPrice := decimal.NewFromFloat(0.0)
				for _, subProduct := range subProducts {
					totalAddPrice = totalAddPrice.Add(decimal.NewFromFloat(subProduct.AddPrice).Mul(decimal.NewFromFloat(subProduct.CopyNum)))
				}
				saleOrderProduct.AddPrice = totalAddPrice.InexactFloat64()
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
					appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
					productBomStockNum := s.createProductBomStockNumFunc(ctx, appService)
					// 检查商品是否超过限购
					status, message := orderProduct.CheckCookingProduct(ctx.GetLanguage(), productBomStockNum)
					if status != constant.CodeSuccess {
						return nil, errors.WithMessage(errors.New(message))
					}
					// 如果该商品是套餐，则修改套餐子商品的数量
					if orderProduct.ProductType == constant.ProductTypePackage {
						subProducts := params.SaleOrder.GetPackageSubProductList(orderProduct.Uuid)
						for _, subProduct := range subProducts {
							uintNum := decimal.NewFromFloat(subProduct.GetUnitNum())                                    // 每个套餐该子商品的数量
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
							uintNum := decimal.NewFromFloat(subProduct.GetUnitNum())                                    // 每个套餐该子商品的数量
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
					subProducts, err := s.newPackageSubProducts(ctx, product.GetSubProducts(), innerParams, params, saleOrderProduct.Uuid, saleOrderProduct.DeductStockType, product.Num)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, subProducts...)
					saleOrderProducts = append(saleOrderProducts, subProducts...)

					// 计算套餐主商品的加价总和 = Σ(子商品每份加价 × 子商品份数)
					totalAddPrice := decimal.NewFromFloat(0.0)
					for _, subProduct := range subProducts {
						totalAddPrice = totalAddPrice.Add(decimal.NewFromFloat(subProduct.AddPrice).Mul(decimal.NewFromFloat(subProduct.CopyNum)))
					}
					saleOrderProduct.AddPrice = totalAddPrice.InexactFloat64()
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
	params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint, packageNum float64) ([]*model.SaleOrderProduct, error) {
	subSaleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, subProduct := range subProducts {
		subSaleOrderProduct, err := s.newSaleOrderProductForPackageSubProduct(ctx, subProduct, innerParams, params, packageUuid, deductStockType, packageNum)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		subSaleOrderProducts = append(subSaleOrderProducts, subSaleOrderProduct)
	}
	return subSaleOrderProducts, nil
}

func (s *orderSrv) newSaleOrderProductForPackageSubProduct(ctx context.Context, product req.ProductParams, innerParams InnerParams, params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint, packageNum float64) (*model.SaleOrderProduct, error) {
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

	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	productBomStockNum := s.createProductBomStockNumFunc(ctx, appService)

	if flavorProductBom.GetStockNum(productBomStockNum) < float64(product.Num) {
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
				return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
			}
		}
	}

	// 获取加料信息（使用之前已创建的库存服务实例）
	sauceProductBoms, errSauceProductBoms := GetSauceInfo(ctx, db, product.SauceProductBomUuidList, product.Num, productBomStockNum)
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
		CopyNum:                product.Num,
		UnitNum:                product.UnitNum,
		PackageNum:             packageNum,
		IsTabletAddAndCooking:  innerParams.IsTabletAddAndCooking,
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
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.ToJson(), // 填顾客下单时规格的名字
			Price:          flavorProductBom.Price,
			ProductBomUuid: product.FlavorProductBomUuid,
			ErpCode:        flavorProductBom.ErpCode,
		},
		Attribute:     attributes,
		IsAcceptOrder: uint(isAcceptOrder),
		Remark:        product.Remark,
	}, &productPackage, product.Operation)

	// 设置加价金额（子商品）
	saleOrderProduct.AddPrice = product.AddPrice

	// 生成签名
	saleOrderProduct.UpdateSign()
	ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(params.Setting)

	return saleOrderProduct, nil
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

// func (f *FlavorNum) IsStockShortage() bool {
// 	for _, saleOrderProductBom := range f.SaleOrderProduct.SaleOrderProductBoms {
// 		if saleOrderProductBom.IsDelete() {
// 			continue
// 		}
// 		if saleOrderProductBom.IsFlavor() {
// 			return saleOrderProductBom.ProductBom.IsStockShortageWithMaterial(f.Num)
// 		}
// 	}
// 	return false
// }

// collectProductBomUuids 收集商品的所有规格和小料的 BomUuid（包括套餐中的）
// 返回 BomUuid 集合
func (s *orderSrv) collectProductBomUuids(product product_resp.Product) map[uint64]bool {
	bomUuids := make(map[uint64]bool)
	// 收集商品规格的 BomUuid
	for _, flavor := range product.Flavors.List {
		if flavor.BomUuid > 0 {
			bomUuids[flavor.BomUuid] = true
		}
	}
	// 收集商品小料的 BomUuid
	for _, sauce := range product.Sauces.List {
		if sauce.BomUuid > 0 {
			bomUuids[sauce.BomUuid] = true
		}
	}
	// 如果是套餐，收集套餐分组中商品的规格和小料 BomUuid
	if product.PackageGroupList != nil {
		for _, group := range product.PackageGroupList.List {
			for _, pkgProduct := range group.Products.List {
				// 收集套餐商品规格的 BomUuid
				for _, flavor := range pkgProduct.Detail.Flavors.List {
					if flavor.BomUuid > 0 {
						bomUuids[flavor.BomUuid] = true
					}
				}
				// 收集套餐商品小料的 BomUuid
				for _, sauce := range pkgProduct.Detail.Sauces.List {
					if sauce.BomUuid > 0 {
						bomUuids[sauce.BomUuid] = true
					}
				}
			}
		}
	}
	return bomUuids
}

// getProductBomStockNumMap 使用 ProductInventoryAppService 批量查询库存
// 返回库存映射表，key 为 BomUuid，value 为库存数量
func (s *orderSrv) getProductBomStockNumMap(ctx context.Context, bomUuids map[uint64]bool, productPackageUuid uint64) map[uint64]float64 {
	stockNumMap := make(map[uint64]float64)
	if len(bomUuids) == 0 {
		return stockNumMap
	}
	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)

	for bomUuid := range bomUuids {
		inventory, err := appService.GetProductInventory(ctx, bomUuid)
		if err != nil {
			// 如果查询失败，记录日志但继续处理其他规格/小料
			logger.Logger.Warn("查询商品规格/小料库存失败",
				zap.Error(err),
				zap.Uint64("bom_uuid", bomUuid),
				zap.Uint64("product_package_uuid", productPackageUuid),
			)
			// 失败时使用无限库存作为默认值
			inventory = constant.ProductBomInfiniteStock
		}
		stockNumMap[bomUuid] = inventory
	}
	return stockNumMap
}

// createProductBomStockNumFunc 创建一个函数，用于根据productBomUuid获取商品库存
// 使用ProductInventoryAppService的GetProductInventory方法获取库存
// 返回一个func(productBomUuid uint64) float64类型的函数
func (s *orderSrv) createProductBomStockNumFunc(ctx context.Context, appService *inventoryApp.ProductInventoryAppService) func(productBomUuid uint64) float64 {
	return func(productBomUuid uint64) float64 {
		inventory, err := appService.GetProductInventory(ctx, productBomUuid)
		if err != nil {
			// 如果获取库存失败，记录错误并返回0（表示无库存）
			ctx.Log().Error("获取商品库存失败", zap.Uint64("productBomUuid", productBomUuid), zap.Error(err))
			return 0
		}
		return inventory
	}
}

// checkProductBomInventory 检查productBomNumMap中的商品库存是否充足
// 使用ProductInventoryAppService的GetProductInventory方法获取库存
// 返回库存不足的商品列表
func (s *orderSrv) checkProductBomInventory(ctx context.Context, productBomNumMap map[uint64]*FlavorNum) []*model.SaleOrderProduct {
	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)

	stockShortageProducts := make([]*model.SaleOrderProduct, 0)
	for productBomUuid, flavorNum := range productBomNumMap {
		// 获取商品库存
		inventory, err := appService.GetProductInventory(ctx, productBomUuid)
		if err != nil {
			// 如果获取库存失败，记录错误但不中断检查流程
			ctx.Log().Error("获取商品库存失败", zap.Uint64("productBomUuid", productBomUuid), zap.Error(err))
			continue
		}
		// 检查库存是否充足：库存需要大于等于所需数量
		if inventory < flavorNum.Num {
			stockShortageProducts = append(stockShortageProducts, flavorNum.SaleOrderProduct)
		}
	}
	return stockShortageProducts
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

		// 使用工厂方法创建库存应用服务实例
		appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
		productBomStockNum := s.createProductBomStockNumFunc(ctx, appService)

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
				status, message = saleOrderProduct.CheckOutProduct(productBomStockNum)
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
				// 使用工厂方法创建库存应用服务实例
				appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
				productBomStockNum := s.createProductBomStockNumFunc(ctx, appService)
				// 如果是送厨检查
				status, message = saleOrderProduct.CheckCookingProduct(ctx.GetLanguage(), productBomStockNum)
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
						if err != nil {
							return nil, errors.WithMessage(err)
						}
						// 先计算一次未下单的商品价格, 用于更新h5订单中未下单的商品价格
						saleBill := shopCartInfo.SaleBill
						s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice())
						// 重新获取销售账单信息, 用于更新桌台已下单的商品价格
						shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
						if err != nil {
							return nil, errors.WithMessage(err)
						}
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

		// 检查各个规格商品的库存是否充足
		stockShortageProducts := s.checkProductBomInventory(ctx, productBomNumMap)
		for _, saleOrderProduct := range stockShortageProducts {
			exist := false
			for _, existingProduct := range statusMap[constant.CodeOrderCheckProductStockZero] {
				if existingProduct.Uuid == saleOrderProduct.Uuid {
					exist = true
					break
				}
			}
			// 如果没有存在，添加
			if !exist {
				statusMap[constant.CodeOrderCheckProductStockZero] = append(statusMap[constant.CodeOrderCheckProductStockZero], saleOrderProduct)
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
func (s *orderSrv) checkBuffetCustomerTypePriceChanged(_ context.Context, saleBill *model.SaleBill) *resp.OrderCheckServiceRes {
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
				// 优先使用快照字段，降级使用关联表数据
				// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
				buffetLocaleName := buffetCustomer.GetLocaleBuffetPackageName()
				// 自助餐顾客类型价格变动
				// Requirement: story-main-buffet-customer-type-name-snapshot-fix
				customerTypeLocaleName := buffetCustomer.GetLocaleName()
				customer := resp.Product{
					Uuid:                buffetCustomer.Uuid,
					LocaleName:          buffetLocaleName,
					LocaleAttributeName: customerTypeLocaleName,
					Num:                 float64(buffetCustomer.Num),
					SalePrice:           buffetCustomer.SalePrice,
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
func (s *orderSrv) getOrderProductForDecreaseStock(_ context.Context, unCookingSaleOrderProducts []*model.SaleOrderProduct) ([]*model.SaleOrderProduct, error) {
	products := make([]*model.SaleOrderProduct, 0)
	for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {

		if unCookingSaleOrderProduct.IsCookingDeductStock() {
			products = append(products, unCookingSaleOrderProduct)
		}
	}
	return products, nil
}

// 获取减库存的清单信息
func (s *orderSrv) getDecreaseStockList(_ context.Context, cookingDeductSaleOrderProducts []*model.SaleOrderProduct) ([]*model.Product, error) {
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
					// 防止 Material 为空
					if productBomMaterial.Material == nil {
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
					// 防止 Material 为空
					if material.Material == nil {
						continue
					}
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

func newProductionOrder(ctx context.Context, saleOrderUuid, saleBillUuid, deskUuid uint64, unCookingSaleOrderProducts []*model.SaleOrderProduct) *model.ProductionOrder {
	productionOrderUuid, _ := utils.GetID()
	productionOrderProducts := make([]*model.ProductionOrderProduct, 0)
	for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {
		var firstCategoryUuid uint64
		if unCookingSaleOrderProduct.ProductPackage != nil {
			firstCategoryUuid = unCookingSaleOrderProduct.ProductPackage.ProductCategory.GetFirstCategoryUuid()
		}
		attributeName := unCookingSaleOrderProduct.GetAttributeName()

		// 送厨的商品数量, sale_order_product.num 该字段已经计算了单位数量和份数,表示送厨的商品数量
		productionOrderProductNum := unCookingSaleOrderProduct.Num
		productionOrderProduct := model.ProductionOrderProduct{
			SaleBillUuid:          saleBillUuid,
			ProductionOrderUuid:   productionOrderUuid,
			SaleOrderProductUuid:  unCookingSaleOrderProduct.Uuid,
			FirstCategoryUuid:     firstCategoryUuid,
			ProductPackageUuid:    unCookingSaleOrderProduct.ProductPackageUuid,
			Num:                   productionOrderProductNum,
			InitNum:               productionOrderProductNum,
			Name:                  unCookingSaleOrderProduct.Name,
			FlavorName:            unCookingSaleOrderProduct.FlavorName,
			ProductBomUuid:        unCookingSaleOrderProduct.GetFlavorBomUuid(),
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
			BatchTagUuid: unCookingSaleOrderProduct.BatchTagUuid,
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

type InstanceAutoFlavorProduct map[uint64]resp.ProductAutoAddReq

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

type ProductInfo struct {
	ProductName dto.LocaleResponse
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
		// 优先使用快照字段，降级使用关联表数据
		// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
		buffetLocaleName := product.GetLocaleBuffetPackageName()
		buffetName := buffetLocaleName.EN
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
				productBom := subProduct.GetFlavorSaleOrderProductBom()
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
		productBom := product.GetFlavorSaleOrderProductBom()
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
			item := &selling.PosInvoiceItem{
				ItemCode:   erpCode,
				Qty:        product.Num,
				Rate:       0,    // 商品未含税价格（折后）
				Amount:     0,    // 商品未含税价格（折后）* 数量
				IsFreeItem: true, // 零元商品当作赠菜
			}
			if product.IsPackageProduct() { // 0 元的套餐固定使用TC001
				packageName := language.JsonToLocaleResponse(product.Name) // 套餐名称
				item.ItemCode = "TC001"
				item.Description = packageName.EN
			}
			items = append(items, item)
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
func (s *orderSrv) ReturnPosInvoice(ctx context.Context, saleOrder *model.SaleOrder, returnOrder *model.ReturnOrder, saleBill *model.SaleBill, db *gorm.DB, returnType int, isPartReturn bool) (*selling.ReturnPosInvoiceResp, error) {
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
					num := decimal.NewFromFloat(product.Num).Mul(decimal.NewFromFloat(subProduct.GetUnitNum())).Round(3).InexactFloat64()
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
				if saleOrderProduct.IsPackageProduct() {
					product.ErpCode = "TC001" // 当退款套餐商品时，商品编码为TC001
				}
				packageName := language.JsonToLocaleResponse(product.ProductName) // 套餐名称
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    product.ErpCode,
					Qty:         -product.Num,
					Rate:        0,              // 商品未含税价格（折后）
					Amount:      0,              // 商品未含税价格（折后）* 数量
					IsFreeItem:  true,           // 零元商品当作赠菜
					Description: packageName.EN, // 套餐商品描述
				})
			} else {
				packageName := language.JsonToLocaleResponse(product.ProductName) // 套餐名称
				item := &selling.PosInvoiceItem{
					ItemCode:    product.ErpCode,
					Qty:         -product.Num,
					Rate:        product.GetProductPriceNoneTax(taxFee, saleOrderProduct.HasTax()),        // 商品未含税价格（折后）
					Amount:      -product.GetProductTotalAmountNoneTax(taxFee, saleOrderProduct.HasTax()), // 商品未含税价格（折后）* 数量
					Description: packageName.EN,                                                           // 套餐商品描述
				}
				if saleOrderProduct.IsGiftProduct() {
					item.IsFreeItem = true
				}
				if saleOrderProduct.IsPackageProduct() {
					item.ItemCode = "TC001" // 当退款套餐商品时，商品编码为TC001
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
			// 优先使用快照字段，降级使用关联表数据
			// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
			buffetLocaleName := buffetCustomer.GetLocaleBuffetPackageName()
			buffetName := buffetLocaleName.EN
			if buffetCustomer.SalePrice == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
				item := &selling.PosInvoiceItem{
					ItemCode:    "ZZC001",
					Qty:         -product.Num,
					Rate:        0,
					Amount:      0,
					IsFreeItem:  true,
					Description: fmt.Sprintf("%s-%s", buffetName, buffetCustomer.Name),
				}
				items = append(items, item)
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode:    "ZZC001",
					Qty:         -product.Num,
					Rate:        buffetCustomer.GetFinalSalePriceNoneTax(),
					Amount:      -decimal.NewFromFloat(buffetCustomer.GetFinalSalePriceNoneTax()).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
					Description: fmt.Sprintf("%s-%s", buffetName, buffetCustomer.Name),
				}
				items = append(items, item)
			}
		} else if product.ProductType == constant.ReturnOrderProductTypeBuffetAddTimeProduct {
			buffetPackageName := saleBill.GetBuffetPackageName()
			buffetDelayProduct, _, _ := saleOrder.GetSaleOrderBuffetDelayProduct(product.SaleOrderProductUuid)
			if buffetDelayProduct.Price == 0 { // 当商品是0元商品时，可能是通过商品改价为0或原本售价就是0
				item := &selling.PosInvoiceItem{
					ItemCode:    "ZZCJZ001",
					Qty:         -product.Num,
					Rate:        0,
					Amount:      0,
					IsFreeItem:  true,
					Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, buffetDelayProduct.Name),
				}
				items = append(items, item)
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode:    "ZZCJZ001",
					Qty:         -product.Num,
					Rate:        buffetDelayProduct.Price,
					Amount:      -decimal.NewFromFloat(buffetDelayProduct.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
					Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, buffetDelayProduct.Name),
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
func (s *orderSrv) getMoveProductInfo(_ context.Context, saleOrderFrom *model.SaleOrder, req req.InstantOrderSaleOrderMoveProductReq) ([]*model.SaleOrderProduct, []*model.SaleOrderBuffetCustomerType, []*model.SaleOrderBuffetDelayProduct, error) {
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
				unitNum := decimal.NewFromFloat(saleOrderProduct.GetUnitNum())
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
func (s *orderSrv) moveBuffetDelayProduct(ctx context.Context, _ *model.SaleBill, saleOrderFrom, saleOrderTo *model.SaleOrder, delayProducts []*model.SaleOrderBuffetDelayProduct, moveNumMap map[uint64]float64) (map[uint64]*model.SaleOrderBuffetDelayProduct, map[uint64]*model.SaleOrderBuffetDelayProduct, error) {
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

type MustPlanConfirmOption struct {
	IsH5Order bool // 是否是H5订单
}

func WithIsH5Order() func(option *MustPlanConfirmOption) {
	return func(option *MustPlanConfirmOption) {
		option.IsH5Order = true
	}
}

func planProductSoldOut(ctx context.Context, dbm *database.DBManager, plan *resp.InstantProductMustPlan) (bool, error) {
	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(dbm, cache.Global)

	// 如果是可选商品. 只有必点方案的所有商品都无库存时才返回true
	if plan.MustRule == constant.ProductMustPlanMustRuleAny {
		isSaleOut := true
		for _, product := range plan.Products.List {
			// 未满足必点的商品包 - 使用 ProductInventoryAppService 查询库存
			inventory, err := appService.GetProductPackageInventory(ctx, product.Product.Uuid)
			if err != nil {
				return false, errors.WithMessage(err)
			}
			// 库存 > 0 表示有库存，未缺货
			if inventory > 0 {
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
		// 未满足必点的商品包 - 使用 ProductInventoryAppService 查询库存
		inventory, err := appService.GetProductPackageInventory(ctx, product.Product.Uuid)
		if err != nil {
			return false, errors.WithMessage(err)
		}
		// 库存 > 0 表示有库存，未缺货
		if inventory > 0 {
			isSaleOut = false
			break
		}
	}
	// 所有的未满足必点的商品都没有库存时才返回true
	return isSaleOut, nil
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

// generateInvoiceNumber 生成当日递增的发票编号（使用saas库序列表）
func (s *orderSrv) generateInvoiceNumber(ctx context.Context) (string, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	companyUuid := ctx.GetCompanySetting().HeadquarterUuid
	if companyUuid == 0 {
		companyUuid = ctx.GetCompanyUuid()
	}
	seq, err := repository.NewNumberSequenceRepo(saasDB).GetNextSequence(companyUuid, constant.NumberTypeInvoice, dateStr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%04d-%02d-%02d-%03d", now.Year(), now.Month(), now.Day(), seq), nil
}
