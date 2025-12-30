package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IRechargeOrderSrv 定义充值订单服务接口
type IRechargeOrderSrv interface {
	GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder                                                                              // 获取进行中的会员充值订单
	CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error)                                           // 创建充值订单
	AddPaymentMethod(ctx context.Context, addPaymentMethod req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error)                    // 充值订单添加支付方式
	CancelPaymentMethod(ctx context.Context, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error)           // 充值订单撤销支付方式
	ConfirmRechargeOrder(ctx context.Context, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error)              // 确认充值订单
	PrintTicket(ctx context.Context, printRechargeOrderReq req.PrintRechargeOrderReq) (*resp.PrinterData, error)                                // 打印充值订单
	GetRechargeOrderList(ctx context.Context, listReq req.RechargeOrderListReq) (resp.RechargeOrderList, error)                                 // 充值订单列表
	GetRechargeOrderInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderInfo, error)                                                      // 获取充值订单详情
	GetRechargeOrderPaymentQrcode(ctx context.Context, req req.RechargeOrderPaymentQrcodeReq) (*resp.RechargeOrderPaymentQrcodeInfoResp, error) // 获取支付方式的二维码信息
	CancelRechargeOrder(ctx context.Context, uuid uint64) error                                                                                 // 取消充值订单
	GetRechargeOrderRefundInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderRefundInfo, error)                                          // 获取退款信息
	CheckRechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrderReverseSettleInfo, error)                              // 检查反结账信息
	RechargeOrderReverseSettle(ctx context.Context, uuid uint64) error                                                                          // 充值订单反结账
	RechargeOrderRefund(ctx context.Context, refundReq req.RechargeOrderRefundReq) error                                                        // 充值订单退款
	RechargeOrderReReturnOrder(ctx context.Context, req req.RechargeOrderReReturnReq) error                                                     // 充值订单退款
}

type rechargeOrderSrv struct {
	dbm              *database.DBManager // 数据库管理器
	bus              *event.SystemEventBus
	cache            cache.Cache
	paymentMethodSrv IPaymentMethodSrv
	settingSrv       setting.ISrv
	cashBoxSrv       ICashBoxSrv
	memberSrv        IMemberSrv
	smsSrv           ISmsSrv
	staffShiftSrv    IStaffShiftSrv
	lock             lock.Lock
}

func NewRechargeOrderSrv(dbm *database.DBManager, cache cache.Cache, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv, memberSrv IMemberSrv, smsSrv ISmsSrv, staffShiftSrv IStaffShiftSrv) IRechargeOrderSrv {
	return NewRechargeOrderSrvImpl(dbm, cache, paymentMethodSrv, settingSrv, cashBoxSrv, memberSrv, smsSrv, staffShiftSrv)
}

func NewRechargeOrderSrvImpl(dbm *database.DBManager, cache cache.Cache, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv, memberSrv IMemberSrv, smsSrv ISmsSrv, staffShiftSrv IStaffShiftSrv) IRechargeOrderSrv {
	return &rechargeOrderSrv{
		dbm:              dbm,
		bus:              event.NewSystemBus(),
		cache:            cache,
		paymentMethodSrv: paymentMethodSrv,
		settingSrv:       settingSrv,
		cashBoxSrv:       cashBoxSrv,
		memberSrv:        memberSrv,
		smsSrv:           smsSrv,
		staffShiftSrv:    staffShiftSrv,
		lock:             lock.NewSystemLock(),
	}
}

// GetPendingRechargeOrder 进行中的会员充值订单
func (s *rechargeOrderSrv) GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	// 进行中的充值订单
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	if order.Uuid == 0 {
		return resp.RechargeOrder{PaymentOrders: resp.PaymentInfoList{List: make([]resp.PaymentOrder, 0)}}
	}

	respPaymentOrders := make([]resp.PaymentOrder, 0, len(order.PaymentOrders))

	for _, paymentOrder := range order.PaymentOrders {

		var respPaymentOrder resp.PaymentOrder
		copier.Copy(&respPaymentOrder, paymentOrder)

		respPaymentOrder.PaymentMethodCode = paymentOrder.PaymentMethod.Code
		respPaymentOrder.PaymentMethodName = paymentOrder.PaymentMethod.PaymentName
		respPaymentOrder.DisabledCancel = slices.Contains([]int{constant.PaymentMethodCodeLianLianWechatPay,
			constant.PaymentMethodCodeLianLianAliPay,
			constant.PaymentMethodCodeLianLianQRPromptPay}, paymentOrder.PaymentMethod.Code)

		respPaymentOrders = append(respPaymentOrders, respPaymentOrder)
	}

	var respRechargeOrder resp.RechargeOrder
	copier.Copy(&respRechargeOrder, order)

	respRechargeOrder.ChargeDue = s.getChargeDue(order.PaymentOrders)

	respRechargeOrder.PaymentOrders = resp.PaymentInfoList{List: respPaymentOrders}
	return respRechargeOrder
}

// 计算支付订单初始金额
func (s *rechargeOrderSrv) sumPaymentAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = utils.DecimalAdd(sum, paymentOrder.PaymentAmount)
	}
	return sum
}

// 计算应收金额
func (s *rechargeOrderSrv) getRechargeOrderAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = utils.DecimalAdd(sum, paymentOrder.PaymentAmount, paymentOrder.PaymentCommissionFee)
	}
	return sum
}

// 计算找零
func (s *rechargeOrderSrv) getChargeDue(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = utils.DecimalAdd(sum, utils.DecimalSub(paymentOrder.Amount, paymentOrder.PaymentCommissionFee, paymentOrder.PaymentAmount))
	}
	return sum
}

// 计算支付订单实收金额
func (s *rechargeOrderSrv) getActualAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = utils.DecimalAdd(sum, paymentOrder.Amount)
	}
	return sum
}

// 计算支付订单手续费
func (s *rechargeOrderSrv) getPayFee(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = utils.DecimalAdd(sum, paymentOrder.PaymentCommissionFee)
	}
	return sum
}

// CreateRechargeOrder 创建充值订单
func (s *rechargeOrderSrv) CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	staff := ctx.GetStaff()
	var orderResp resp.RechargeOrder

	// 判断会员是否存在
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetMember(memberRepo.WhereUuid(rechargeReq.MemberUuid))
	if member.ID == 0 {
		return orderResp, errors.New("会员不存在")
	}

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))

	if rechargeReq.RechargeOrderUuid != 0 {
		// 如果已经存在已支付的payment_order，则充值金额不能小于现有的 "充值金额不能小于已充值金额"
		order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending),
			rechargeOrderRepo.WithPaymentOrders())
		if order.Uuid != 0 {
			oldRechargeAmount := order.RechargeAmount

			sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
			if rechargeReq.RechargeAmount < sumPaymentAmount {
				return orderResp, errors.New("充值金额不能小于已充值金额")
			}
			if err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
				// 更新充值订单信息
				err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, map[string]any{
					"recharge_amount": rechargeReq.RechargeAmount,
					"gift_amount":     rechargeReq.GiftAmount,
					"gift_point":      rechargeReq.GiftPoint,
					"member_uuid":     rechargeReq.MemberUuid,
					"staff_uuid":      ctx.GetStaffUuid(),
				})
				if err != nil {
					return errors.WithMessage(err)
				}
				err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
					OperatorName:      ctx.GetStaff().RealName,
					OperatorEmail:     ctx.GetStaff().Username,
					Client:            ctx.GetSource(),
					Message:           "变更充值金额",
					Action:            constant.RechargeOrderActionChangeAmount,
					Data:              s.getChangeAmountLogData(rechargeReq.RechargeAmount, oldRechargeAmount),
					RechargeOrderUuid: order.Uuid,
				})
				if err != nil {
					return errors.WithMessage(err)
				}
				return nil
			}); err != nil {
				ctx.Log().Error("修改充值订单失败", zap.Error(err))
				return orderResp, errors.WithMessage(err, "修改充值订单失败")
			}

			// 返回充值订单信息
			return s.GetPendingRechargeOrder(companyUuid), nil
		}
	}

	// 如果已存在进行中的充值订单，且未传递进行中的充值订单Uuid，则直接返回
	pendingRechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending))
	if pendingRechargeOrder.Uuid != 0 {
		return s.GetPendingRechargeOrder(companyUuid), nil
	}

	err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
		// 创建充值订单
		order, err := repository.NewMemberRechargeOrderRepo(tx).Create(model.MemberRechargeOrder{
			OrderNo:        s.generateRechargeOrderNo(),
			DutyNo:         staff.DutyNo,
			RechargeAmount: rechargeReq.RechargeAmount,
			GiftAmount:     rechargeReq.GiftAmount,
			GiftPoint:      rechargeReq.GiftPoint,
			MemberUuid:     rechargeReq.MemberUuid,
			StaffUuid:      ctx.GetStaffUuid(),
		})
		if err != nil {
			return errors.WithMessage(err)
		}
		// 会员充值操作日志
		err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "生成订单",
			Action:            constant.RechargeOrderActionGenerateOrder,
			RechargeOrderUuid: order.Uuid,
		})
		if err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if err != nil {
		return orderResp, errors.WithMessage(err, "创建充值订单失败")
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// AddPaymentMethod 充值订单添加支付方式
func (s *rechargeOrderSrv) AddPaymentMethod(ctx context.Context, addReq req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereUuid(addReq.RechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders(),
	)
	if order.Uuid == 0 || order.Status != constant.RechargeOrderStatusPending {
		return orderResp, errors.New("充值订单不存在")
	}

	// 根据Uuid 获取支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(companyUuid))
	paymentMethod := paymentMethodRepo.GetPaymentMethod(paymentMethodRepo.WhereUuid(addReq.PaymentMethodUuid))
	if paymentMethod.Uuid == 0 {
		return orderResp, errors.New("支付方式不存在")
	}
	if paymentMethod.Code == constant.PaymentMethodCodeBalance {
		return orderResp, errors.New("不能使用余额支付充值")
	}

	isValid, err := s.staffShiftSrv.ValidatePaymentMethod(ctx, order.DutyNo, paymentMethod.Uuid)
	if err != nil || !isValid {
		return orderResp, errors.WithMessage(err, "支付方式不可用")
	}

	// 默认支付订单状态
	paymentOrderStatus := constant.PaymentOrderStatusPaid

	// 获取已存在的该支付方式的支付订单
	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	paymentOrder, _ := paymentOrderRepo.GetPaymentOrder(
		paymentOrderRepo.WhereRelatedUuid(order.Uuid),
		paymentOrderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
	)

	//  在线支付订单 - 如果已经存在直接返回
	if paymentMethod.IsLianLianPay() {
		if paymentOrder.Uuid != 0 {
			if paymentOrder.PaymentAmount != addReq.PaymentAmount {
				return orderResp, errors.New("不能重复支付")
			}
			return s.GetPendingRechargeOrder(companyUuid), nil
		}
		if addReq.PaymentOrderUuid == 0 {
			return orderResp, errors.New("支付订单ID不能为空")
		}
	} else if addReq.PaymentOrderUuid != 0 {
		return orderResp, errors.New("非在线支付无需传支付订单ID")
	}

	// 支付方式是否可用
	if paymentMethod.IsShowMemberRecharge == 0 || !s.paymentMethodSrv.IsEnabled(ctx, paymentMethod, addReq.CompanySetting) {
		return orderResp, errors.New("支付方式未开启")
	}

	// 计算支付手续费
	paymentCommissionFee := paymentMethod.CalculatePaymentCommissionFee(addReq.PaymentAmount)
	sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
	// 支付订单金额大于充值金额
	if sumPaymentAmount >= order.RechargeAmount {
		return orderResp, errors.New("当前已足额")
	}
	sumPaymentAmountAddCash := utils.DecimalAdd(sumPaymentAmount, addReq.PaymentAmount)
	if paymentMethod.Code != constant.PaymentMethodCodeCash && sumPaymentAmountAddCash > order.RechargeAmount {
		return orderResp, errors.New("非现金支付不能大于应收")
	}

	// 支付订单总金额 = 支付金额 + 支付手续费
	amount := paymentMethod.CalculatePaymentAmount(addReq.PaymentAmount)

	var rechargeAmountLeft float64
	var cashPaidPaymentAmount float64
	if paymentMethod.Code == constant.PaymentMethodCodeCash {
		// 如果现金支付订单存在
		if paymentOrder.Uuid != 0 {
			cashPaidPaymentAmount = paymentOrder.PaymentAmount
		}
		rechargeAmountLeft = utils.DecimalSub(order.RechargeAmount, utils.DecimalSub(sumPaymentAmount, cashPaidPaymentAmount))
		if rechargeAmountLeft > 0 && addReq.PaymentAmount > rechargeAmountLeft {
			addReq.PaymentAmount = rechargeAmountLeft
		}
	}

	if paymentOrder.Uuid != 0 { // 存在则更新
		err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
			"payment_amount":         addReq.PaymentAmount,
			"amount":                 amount,
			"payment_commission_fee": paymentCommissionFee,
		})
		if err != nil {
			logger.Logger.Error("添加支付订单-更新支付方式", zap.Error(err))
			return orderResp, errors.WithMessage(err, "添加支付方式失败")
		}
	} else { // 不存在则创建
		currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
		if err != nil {
			logger.Logger.Error("添加支付订单-获取货币设置失败", zap.Error(err))
			return orderResp, errors.WithMessage(err, "添加支付方式失败")
		}
		_, err = paymentOrderRepo.Create(model.PaymentOrder{
			BaseModel:            model.BaseModel{Uuid: addReq.PaymentOrderUuid},
			PaymentMethodName:    paymentMethod.PaymentName,
			PaymentMethodUuid:    paymentMethod.Uuid,
			PaymentFeePercent:    paymentMethod.FeePercent,
			RelatedType:          constant.PaymentOrderRelatedTypeRechargeOrder,
			RelatedUuid:          order.Uuid,
			CurrencyUnit:         currencySetting.Unit, // 留档使用
			PaymentAmount:        addReq.PaymentAmount,
			PaymentCommissionFee: paymentCommissionFee,
			Amount:               amount,
			Status:               paymentOrderStatus,
		})
		if err != nil {
			logger.Logger.Error("添加支付订单-创建支付订单", zap.Error(err))
			return orderResp, errors.WithMessage(err, "添加支付方式失败")
		}
	}

	return s.GetPendingRechargeOrder(companyUuid), nil
}

// CancelPaymentMethod 充值订单撤销支付方式
func (s *rechargeOrderSrv) CancelPaymentMethod(ctx context.Context, cancelReq req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	// 判断支付订单
	paymentOrder, err := paymentOrderRepo.GetPaymentOrder(paymentOrderRepo.WhereUuid(cancelReq.PaymentOrderUuid),
		paymentOrderRepo.WithPaymentMethod(), paymentOrderRepo.WithMemberRechargeOrder())
	if err != nil || paymentOrder.Uuid == 0 {
		return orderResp, errors.New("支付订单不存在")
	}
	if paymentOrder.MemberRechargeOrder == nil || paymentOrder.RelatedUuid != cancelReq.RechargeOrderUuid {
		return orderResp, errors.New("充值订单不存在")
	}
	// lianlianpay不支持撤销
	if paymentOrder.PaymentMethod == nil || paymentOrder.PaymentMethod.Source == constant.PaymentMethodSourceLianLianPay {
		return orderResp, errors.New("支付方式不可撤销")
	}
	// 标记支付订单删除
	if err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
		"delete_time": time.Now().Unix(),
	}); err != nil {
		logger.Logger.Error("标记支付订单失败", zap.Error(err))
		return orderResp, errors.ErrInternal
	}

	// 更新现金支付订单
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(paymentOrder.RelatedUuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	var cashPaymentOrder model.PaymentOrder
	var sumPaymentAmount float64
	for _, po := range order.PaymentOrders {
		if po.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			cashPaymentOrder = po
		}
		sumPaymentAmount = utils.DecimalAdd(sumPaymentAmount, po.PaymentAmount)
	}
	if cashPaymentOrder.Uuid != 0 {
		paymentOrderRepo.Update(cashPaymentOrder.Uuid, map[string]any{
			"payment_amount": utils.DecimalSub(order.RechargeAmount, sumPaymentAmount),
		})
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// 计算非现金支付订单金额(payment_amount)
func (s *rechargeOrderSrv) sumPaymentAmountExcludeCash(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		if paymentOrder.PaymentMethod.Code != constant.PaymentMethodCodeCash {
			sum = utils.DecimalAdd(sum, paymentOrder.PaymentAmount)
		}
	}
	return sum
}

// ConfirmRechargeOrder 确认充值订单
func (s *rechargeOrderSrv) ConfirmRechargeOrder(ctx context.Context, confirmReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(confirmReq.RechargeOrderUuid)
		defer s.lock.UnlockUuid(confirmReq.RechargeOrderUuid)
		ctx.AddLock()
	}
	var confirmResp resp.ConfirmRechargeOrder

	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)
	memberRepo := repository.NewMemberRepo(db)
	member := memberRepo.GetMember(memberRepo.WhereUuid(confirmReq.MemberUuid))
	if member.Uuid == 0 {
		return confirmResp, errors.New("会员不存在")
	}

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereUuid(confirmReq.RechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders(),
		rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
	)
	if order.Uuid == 0 || order.Status != constant.RechargeOrderStatusPending {
		return confirmResp, errors.New("充值订单不存在")
	}

	// 非现金超额
	sumPaymentAmountExcludeCash := s.sumPaymentAmountExcludeCash(order.PaymentOrders)
	if sumPaymentAmountExcludeCash > order.RechargeAmount {
		return confirmResp, errors.New("收款金额大于充值金额，请先修改收款金额")
	}
	// 所有支付订单金额总数小于充值订单充值金额
	sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
	if sumPaymentAmount < order.RechargeAmount {
		return confirmResp, errors.New("未足额支付")
	}

	chargeDue := s.getChargeDue(order.PaymentOrders)
	// 更新充值订单
	paymentTime := time.Now().Unix()
	amount := s.getRechargeOrderAmount(order.PaymentOrders) // 应收金额
	order.Amount = amount
	updates := map[string]any{
		"amount":            amount,
		"status":            constant.RechargeOrderStatusPaid, // 状态,0-pending待支付 1-paid已支付 2-canceled已取消 3-exp已过期
		"payment_time":      paymentTime,
		"charge_due":        chargeDue,                                                                        // 找零
		"balance":           member.GetBalanceAll(),                                                           // 充值前会员余额
		"balance_recharged": utils.DecimalAdd(member.GetBalanceAll(), order.RechargeAmount, order.GiftAmount), // 充值后会员余额
	}

	var memberPointsChanged bool

	err := db.Transaction(func(tx *gorm.DB) error {
		ctx.SetDB(tx)
		err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, updates)
		if err != nil {
			return errors.New("更新充值订单失败")
		}

		// 处理会员积分
		if order.GiftPoint > 0 {
			if err := s.memberSrv.HandleMemberPoints(ctx, MemberPointsChangeReq{
				Uuid:     order.MemberUuid,
				Points:   order.GiftPoint,
				Scene:    constant.MemberPointLogSceneRechargeGive,
				Describe: fmt.Sprintf("收银机管理员充值赠送操作 [%s]", ctx.GetStaff().RealName),
			}); err != nil {
				return errors.WithMessage(err)
			}
			memberPointsChanged = true
		}

		// 处理会员余额
		if order.RechargeAmount > 0 || order.GiftAmount > 0 {
			if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				MemberUuid:  member.Uuid,
				Money:       order.RechargeAmount,
				GiftMoney:   order.GiftAmount,
				Scene:       constant.MemberBalanceLogRecharge,
				Describe:    fmt.Sprintf("收银机管理员操作 [%s]", ctx.GetStaff().RealName),
				RelatedUuid: order.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}

		for _, paymentOrder := range order.PaymentOrders {
			// 存在现金支付，更新钱箱
			if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
					Amount:    utils.DecimalSub(sumPaymentAmount, sumPaymentAmountExcludeCash),
					Scene:     constant.CashBoxLogSceneRecharge,
					OrderUuid: order.Uuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		if err := repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "充值",
			Action:            constant.RechargeOrderActionRecharge,
			Data:              s.getRechargeLogData(order),
			RechargeOrderUuid: order.Uuid,
		}); err != nil {
			return errors.WithMessage(err)
		}

		// 保存发票到erp
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			order.PaymentTime = paymentTime
			invoiceResp, err := s.SavePosInvoice(ctx, &order, tx)
			if err != nil {
				return errors.WithMessage(err)
			}
			// TODO: 要是“更新充值订单的商品发票名称”失败，事务回滚后用户重新确认充值订单，会导致ERP系统中该笔订单有两个发票
			order.ErpProductsInvoiceName = invoiceResp.ProductsInvoiceName
			// 更新充值订单的商品发票名称
			if err := repository.NewMemberRechargeOrderRepo(tx).UpdateErpProductsInvoiceName(order.Uuid, order.ErpProductsInvoiceName); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	})
	if err != nil {
		return confirmResp, errors.WithMessage(err)
	}

	if memberPointsChanged {
		utils.Go(func() {
			s.memberSrv.HandleMemberUpgrade(companyUuid, member.Uuid)
		})
	}

	// 打印充值单
	utils.Go(func() {
		order := rechargeOrderRepo.GetRechargeOrder(
			rechargeOrderRepo.WithMember(),
			rechargeOrderRepo.WithStaff(),
			rechargeOrderRepo.WithPaymentOrders(),
			rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
			rechargeOrderRepo.WhereUuid(order.Uuid),
		)
		_, err := printer.NewPrinterRepo(ctx).PrintingRechargeOrder(order, 0)
		if err != nil {
			logger.Logger.Error("打印充值单失败", zap.Error(err))
		}
	})

	// 发送结账短信
	{
		// 获取最新的会员信息
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(confirmReq.MemberUuid)
		if err != nil {
			ctx.Log().Info("停止发送短信（充值），获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			utils.Go(func() {
				if member != nil {
					smsReq := sms.MemberRechargeRequest{
						Company:       ctx.GetCompany().Name,
						Recharge:      order.RechargeAmount,
						BonusMoney:    order.GiftAmount,
						BonusPoints:   order.GiftPoint,
						Balance:       member.GetBalanceAll(),
						PointsBalance: member.GetPoints(),
					}
					if err := s.smsSrv.SendMemberRechargeSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送充值短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送充值短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			})
		}
	}
	utils.Go(func() {
		// 发布“会员余额变动”事件
		s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
			BasePayload: event.BasePayload{ // 会员余额变动
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
		// 发布“会员积分变动”事件
		s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
			BasePayload: event.BasePayload{ // 会员积分变动
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})
	// 发布“统计”事件
	utils.Go(func() {
		s.bus.PublishStatisticsMemberEvent(event.StatisticsMemberPayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			MemberRechargeOrderUuid: order.Uuid,
		})
	})

	return s.confirmRechargeOrderResp(companyUuid, order.Uuid), nil
}

// 保存发票到erp
func (s *rechargeOrderSrv) SavePosInvoice(ctx context.Context, memberRechargeOrder *model.MemberRechargeOrder, db *gorm.DB) (*selling.SavePosInvoiceResp, error) {
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

	items = append(items, &selling.PosInvoiceItem{
		ItemCode: constant.PosInvoiceItemCodeMembershipRecharge,
		Qty:      memberRechargeOrder.RechargeAmount,
		Rate:     1,
		Amount:   memberRechargeOrder.RechargeAmount,
	})
	// 如果有支付手续费，则添加一个虚拟商品来记录支付手续费
	commissionFee := memberRechargeOrder.GetCommissionFee()
	if commissionFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee,
			Qty:      commissionFee,
			Rate:     1,
			Amount:   commissionFee,
		})
	}

	payments := make([]*selling.PosInvoicePayment, 0)

	// 获取所有支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
	methodMap := make(map[int]string)
	paymentIdMap := make(map[int]string)
	for _, paymentMethod := range paymentMethods {
		if paymentMethod.ErpnextPayment != "" {
			methodMap[paymentMethod.Code] = paymentMethod.ErpnextPayment
		}
		if paymentMethod.ErpnextPaymentId != "" {
			paymentIdMap[paymentMethod.Code] = paymentMethod.ErpnextPaymentId
		}
	}
	getPaymentId := func(paymentMethodCode int) string {
		if paymentId, ok := paymentIdMap[paymentMethodCode]; ok {
			return paymentId
		}
		return ""
	}
	for _, payment := range memberRechargeOrder.PaymentOrders {
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
		var paymentID *string
		paymentId := getPaymentId(payment.PaymentMethod.Code)
		if paymentId != "" {
			paymentID = &paymentId
		}
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: modeOfPayment,
			PaymentId:     paymentID,
			Amount:        payment.Amount,
		})
	}

	customerUuid := fmt.Sprintf("%d", memberRechargeOrder.MemberUuid)
	if customerUuid == "0" {
		customerUuid = ""
	}
	erpSrv := erp.NewIErpSrv(s.dbm)
	param := req.SavePosInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          memberRechargeOrder.OrderNo,
		OpenPosEntryName: shiftLog.ErpnextOpenPosEntryName,
		PostingDatetime:  memberRechargeOrder.PaymentTime,
		CustomerUuid:     customerUuid,
		Items:            items,    // 订单商品列表
		Payments:         payments, // 订单付款列表
	}
	if memberRechargeOrder.IsErpReverseSettle() {
		param.AmendedProductsInvoiceName = memberRechargeOrder.ErpProductsInvoiceName
	}
	response, err := erpSrv.SavePosInvoice(ctx, param)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return response, nil
}

// 退款发票到erp
func (s *rechargeOrderSrv) ReturnPosInvoice(ctx context.Context, memberRechargeOrder *model.MemberRechargeOrder, returnOrder *model.ReturnOrder, db *gorm.DB, returnType uint) (*selling.ReturnPosInvoiceResp, error) {
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

	//服务费 - Service Fee - item code: VP001
	//会员充值 - Membership Recharge- item code: VP002
	//配送费 - Delivery Fee- item code: VP003
	//支付手续费 - Payment Processing Fee- item code: VP004
	// 订单商品列表
	items := make([]*selling.PosInvoiceItem, 0)
	remainingAmount := memberRechargeOrder.GetRemainingAmount() // 剩余的充值金额
	if remainingAmount > 0 && returnOrder.RefundAmount > remainingAmount {
		// 当“剩余的充值金额”小于退款金额时，将剩余的充值金额作为商品退款
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeMembershipRecharge, // 会员充值
			Qty:      -remainingAmount,
			Rate:     1,
			Amount:   -remainingAmount,
		})
		// 剩余的退款金额，作为支付手续费，添加一个虚拟商品来记录退款的支付手续费
		commissionFee := decimal.NewFromFloat(returnOrder.RefundAmount).Sub(decimal.NewFromFloat(remainingAmount)).Round(2).InexactFloat64()
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee, // 支付手续费
			Qty:      -commissionFee,
			Rate:     1,
			Amount:   -commissionFee,
		})
	} else if remainingAmount > 0 && returnOrder.RefundAmount <= remainingAmount {
		// 当“剩余的充值金额”大于等于退款金额时，将退款金额作为商品退款
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeMembershipRecharge, // 会员充值
			Qty:      -returnOrder.RefundAmount,
			Rate:     1,
			Amount:   -returnOrder.RefundAmount,
		})
	} else if remainingAmount <= 0 && returnOrder.RefundAmount > 0 {
		// 当“剩余的充值金额”小于等于0时，还要继续退款时，退的是手续费
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee, // 支付手续费
			Qty:      -returnOrder.RefundAmount,
			Rate:     1,
			Amount:   -returnOrder.RefundAmount,
		})
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
		OrderNo:          memberRechargeOrder.OrderNo,
		OpenPosEntryName: shiftLog.ErpnextOpenPosEntryName,
		PostingDatetime:  returnOrder.CreateTime, // 退款单时间
		CompanyAbbr:      companySetting.ErpnextCompanyAbbr,
		InvoiceName:      memberRechargeOrder.ErpProductsInvoiceName, // 发票名称
		Items:            items,                                      // 订单退款商品列表
		Payments:         payments,                                   // 订单退款列表
	}
	response, err := erpSrv.ReturnPosInvoice(ctx, param)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return response, nil
}

// 确认充值订单
func (s *rechargeOrderSrv) confirmRechargeOrderResp(companyUuid uint64, rechargeOrderUuid uint64) resp.ConfirmRechargeOrder {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(rechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())

	paymentMethods := make([]string, 0, len(order.PaymentOrders))
	for _, order := range order.PaymentOrders {
		paymentMethods = append(paymentMethods, order.PaymentMethodName)
	}
	return resp.ConfirmRechargeOrder{
		Amount:         order.Amount,
		ActualAmount:   s.getActualAmount(order.PaymentOrders),
		ChargeDue:      order.ChargeDue,
		PaymentMethods: paymentMethods,
	}
}

// PrintTicket 打印充值订单
func (s *rechargeOrderSrv) PrintTicket(ctx context.Context, printRechargeOrderReq req.PrintRechargeOrderReq) (*resp.PrinterData, error) {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WithMember(),
		rechargeOrderRepo.WithStaff(),
		rechargeOrderRepo.WithPaymentOrders(),
		rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WhereUuid(printRechargeOrderReq.RechargeOrderUuid),
	)
	if order.Uuid == 0 {
		return nil, errors.New("充值订单不存在")
	}
	//
	printLang := printRechargeOrderReq.PrintLang
	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, printLang).PrintingRechargeOrder(
		order,
		1,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

// 生成充值订单编号
func (s *rechargeOrderSrv) generateRechargeOrderNo() string {
	// 定义类型编号
	typeNum := "3"

	// 获取当前时间
	now := time.Now()
	datePart := now.Format("20060102") // 对应PHP的 date('Ymd')

	// 获取微秒部分
	nanos := now.Nanosecond()
	microSeconds := fmt.Sprintf("%06d", nanos/1000) // 转换为6位微秒

	// 生成随机数部分
	randBytes := make([]byte, 3)
	rand.Read(randBytes)
	randNum := fmt.Sprintf("%03d", int(randBytes[0])+int(randBytes[1])+int(randBytes[2]))

	// 组合订单号
	orderNo := datePart + typeNum + microSeconds + randNum

	// 使用redis检查订单号是否存在
	key := "__CREATE_NEW_RC_ORDERNO__" + orderNo
	if _, exits := s.cache.Get(key); exits {
		return s.generateRechargeOrderNo()
	}

	// 设置缓存，5秒过期
	err := s.cache.Set(key, 1, 5*time.Second)
	if err != nil {
		// 处理错误，这里简单地递归重试
		return s.generateRechargeOrderNo()
	}
	return orderNo
}

// GetRechargeOrderList 获取充值订单列表
func (s *rechargeOrderSrv) GetRechargeOrderList(ctx context.Context, listReq req.RechargeOrderListReq) (resp.RechargeOrderList, error) {

	staff := ctx.GetStaff()
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	var options, countOptions []repository.DBOption
	if listReq.OrderNo != "" {
		orderNoOption := rechargeOrderRepo.WhereOrderNoLike(listReq.OrderNo)
		options = append(options, orderNoOption)
		countOptions = append(countOptions, orderNoOption)
	}
	if listReq.Status != -1 {
		options = append(options, rechargeOrderRepo.WhereStatus(listReq.Status))
	}

	// -1=全都、 0=今天、 1=昨天、 2=本周
	if listReq.DateType >= 0 && listReq.DateType <= 2 {
		var startTime, endTime int64
		tz := ctx.GetCompanySetting().Timezone
		switch listReq.DateType {
		case 0: // 今天
			startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeToday)
		case 1: // 昨天
			startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeYesterday)
		case 2: // 本周
			startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisWeek)
		}
		dateTypeOption := rechargeOrderRepo.WhereCreateTimeBetween(startTime, endTime)
		options = append(options, dateTypeOption)
		countOptions = append(countOptions, dateTypeOption)
	}

	// 处理日期时间字符串参数
	if listReq.QueryStartDate != "" && listReq.QueryEndDate != "" {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(listReq.QueryStartDate)
		if err == nil {
			listReq.QueryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(listReq.QueryEndDate)
		if err == nil {
			listReq.QueryEndTime = endTime
		}
	}

	// 日期范围
	if listReq.QueryStartTime != 0 || listReq.QueryEndTime != 0 {
		var timeRanges []repository.TimeRange
		if listReq.EnableCreateTime {
			timeRanges = append(timeRanges, repository.TimeRange{
				Field:     "create_time",
				StartTime: listReq.QueryStartTime,
				EndTime:   listReq.QueryEndTime,
			})
		}
		if listReq.EnablePaymentTime {
			timeRanges = append(timeRanges, repository.TimeRange{
				Field:     "payment_time",
				StartTime: listReq.QueryStartTime,
				EndTime:   listReq.QueryEndTime,
			})
		}

		timeRangeOption := rechargeOrderRepo.WhereTimeBetween(repository.TimeQueryParams{
			TimeRanges: timeRanges,
			Operator:   "OR",
		})
		options = append(options, timeRangeOption)
		countOptions = append(countOptions, timeRangeOption)
	}

	// 关联查询
	options = append(options,
		rechargeOrderRepo.WithMember(),
		rechargeOrderRepo.WithPaymentOrders(),
		rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WithStaff(),
	)
	rechargeOrders, total, err := rechargeOrderRepo.PaginateGetRechargeOrder(listReq.PageNo, listReq.PageSize, options...)

	if err != nil {
		return resp.RechargeOrderList{}, errors.ErrInternal
	}

	items := make([]resp.RechargeOrderItem, 0, len(rechargeOrders))

	for _, order := range rechargeOrders {

		paymentMethods := make([]string, 0, len(order.PaymentOrders))
		for _, paymentOrder := range order.PaymentOrders {
			paymentMethods = append(paymentMethods, paymentOrder.PaymentMethodName)
		}

		var cashierName string
		if order.Staff != nil {
			cashierName = order.Staff.RealName
		}

		// 是否可反结账：同一个收银员工、同一个班次、已支付、未退款
		isCellReverseSettle := order.Status == constant.RechargeOrderStatusPaid &&
			order.DutyNo == staff.DutyNo && order.StaffUuid == staff.Uuid && order.RefundMoney == 0
		items = append(items, resp.RechargeOrderItem{
			Uuid:           order.Uuid,
			OrderNo:        order.OrderNo,
			Status:         order.Status,
			PaymentTime:    order.PaymentTime,
			RechargeAmount: order.RechargeAmount,
			Amount:         utils.DecimalSub(order.Amount, order.RefundMoney), // 实付金额要减去退款
			PaymentMethods: paymentMethods,
			GiftAmount:     order.GiftAmount,
			GiftPoint:      order.GiftPoint,
			MemberUuid: func() uint64 {
				if order.Member != nil {
					return uint64(order.Member.ID)
				}
				return 0
			}(),
			RefundMoney: order.RefundMoney,
			Cashier: resp.RechargeOrderCashier{
				RealName: cashierName,
			},
			Extra: resp.RechargeOrderItemExtra{
				IsCellRefund:        order.Status == constant.RechargeOrderStatusPaid && order.Amount > order.RefundMoney, // 退款完不可再退款
				IsCellCancel:        order.Status == constant.RechargeOrderStatusPending,
				IsCellReverseSettle: isCellReverseSettle,
			},
		})
	}

	// 获取数量
	getOrderNum := func(status int) int64 {
		num, _ := rechargeOrderRepo.GetOrderCount(append(countOptions, rechargeOrderRepo.WhereStatus(status))...)
		return num
	}

	return resp.RechargeOrderList{
		List: items,
		Meta: resp.RechargeOrderListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   listReq.PageNo,
				PageSize: listReq.PageSize,
				Total:    total,
			},
			UnpaidNum:   getOrderNum(constant.RechargeOrderStatusPending),
			CancelNum:   getOrderNum(constant.RechargeOrderStatusCanceled),
			CompleteNum: getOrderNum(constant.RechargeOrderStatusPaid),
		},
	}, nil
}

// GetRechargeOrderInfo 获取充值订单信息
func (s *rechargeOrderSrv) GetRechargeOrderInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderInfo, error) {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithMember(), rechargeOrderRepo.WithStaff(), rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WithRechargeOrderOperationLogs())
	if order.Uuid == 0 {
		return resp.RechargeOrderInfo{}, errors.New("充值订单不存在")
	}
	var cashierName string
	if order.Staff != nil {
		cashierName = order.Staff.RealName
	}

	paymentMethods := make([]resp.RechargeOrderPaymentMethod, 0, len(order.PaymentOrders))
	for _, paymentOrder := range order.PaymentOrders {
		amount := paymentOrder.Amount
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			amount = paymentOrder.PaymentAmount
		}
		paymentMethods = append(paymentMethods, resp.RechargeOrderPaymentMethod{
			Name:       paymentOrder.PaymentMethodName,
			Price:      amount,
			Code:       paymentOrder.PaymentMethod.Code,
			SourceText: i18n.Translate(ctx.GetLanguage(), constant.PaymentMethodSourceTextMap[paymentOrder.PaymentMethod.Source]),
		})
	}

	language := ctx.GetLanguage()
	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)
	logs := make([]resp.RechargeOrderOperationLogItem, 0, len(order.RechargeOrderOperationLogs))
	for _, log := range order.RechargeOrderOperationLogs {
		actionDesc := s.getActionDescription(ctx, log, ctx.GetLanguage())
		actionText := s.getActionText(log, ctx.GetLanguage())
		refundPayTypes := make([]resp.RefundPayType, 0)
		var refundType uint
		if log.Action == constant.RechargeOrderActionRefund {
			var refundLog RefundLog
			json.Unmarshal([]byte(log.Data), &refundLog)
			refundType = refundLog.RefundType
			//
			for _, payType := range refundLog.RefundPayTypes {
				// 支付方式名称
				payTypeName := payType.Name
				if slices.Contains([]int{
					constant.PaymentMethodCodeFreePay,
					constant.PaymentMethodCodeBalance,
					constant.PaymentMethodCodeCash,
				}, payType.Code) {
					payTypeName = i18n.Translate(language, payTypeName)
				}
				// 退款支付类型
				data := resp.RefundPayType{
					Price:            "0",
					Code:             payType.Code,
					Name:             payTypeName,
					RefundMoney:      utils.FormatFloat(payType.Amount),
					RefundStatus:     1,
					ReturnOrderUuid:  payType.ReturnOrderUuid,
					ReturnAmountUuid: payType.ReturnAmountUuid,
					PaymentOrderUuid: payType.PaymentOrderUuid,
					Unit:             currencySetting.Unit,
				}
				// 银行支付
				if slices.Contains([]int{
					constant.PaymentMethodCodeLianLianWechatPay,
					constant.PaymentMethodCodeLianLianAliPay,
					constant.PaymentMethodCodeLianLianQRPromptPay,
				}, payType.Code) {
					returnOrderRepo := repository.NewReturnOrderRepo(ctx.GetDB())
					orderAmount, err := returnOrderRepo.GetReturnOrderAmount(returnOrderRepo.WithReturnOrder(), returnOrderRepo.WhereUuid(payType.ReturnAmountUuid))
					if err == nil {
						data.RefundStatus = utils.IfInt(orderAmount.RefundStatus == 2, 0, 1)
					}
					if err == nil && payType.Code == constant.PaymentMethodCodeLianLianQRPromptPay {
						data.BankCode = orderAmount.ReturnOrder.BankCode
						data.AccountNo = orderAmount.ReturnOrder.AccountNo
						data.AccountName = orderAmount.ReturnOrder.AccountName
					}
				}
				refundPayTypes = append(refundPayTypes, data)
			}
		}
		var desc string
		if actionDesc != "" {
			desc = actionText + ": " + actionDesc
		} else {
			desc = actionText
		}
		logs = append(logs, resp.RechargeOrderOperationLogItem{
			RealName:       log.OperatorName,
			Username:       log.OperatorEmail,
			Client:         i18n.Translate(language, constant.SourceTextMap[log.Client]),
			CreateTime:     log.CreateTime,
			Description:    desc,
			RefundType:     refundType,
			RefundPayTypes: refundPayTypes,
		})
	}

	staff := ctx.GetStaff()
	// 是否可反结账：同一个收银员工、同一个班次、已支付、未退款
	isCellReverseSettle := order.Status == constant.RechargeOrderStatusPaid &&
		order.DutyNo == staff.DutyNo && order.StaffUuid == staff.Uuid && order.RefundMoney == 0
	return resp.RechargeOrderInfo{
		Uuid:    order.Uuid,
		OrderNo: order.OrderNo,
		Member: resp.RechargeOrderMember{
			Uuid:     uint64(order.Member.ID),
			Nickname: order.Member.Nickname,
			Phone:    order.Member.Phone,
		},
		Status: order.Status,
		Cashier: resp.RechargeOrderCashier{
			RealName: cashierName,
		},
		RechargeAmount: order.RechargeAmount,
		Amount:         utils.DecimalSub(order.Amount, order.RefundMoney),
		ChargeDue:      order.ChargeDue,
		PaymentTime:    order.PaymentTime,
		CreateTime:     order.CreateTime,
		GiftAmount:     order.GiftAmount,
		GiftPoint:      order.GiftPoint,
		PaymentMethods: paymentMethods,
		OperationLog:   resp.RechargeOrderOperationLog{List: logs},
		Extra: resp.RechargeOrderItemExtra{
			IsCellRefund:        order.Status == constant.RechargeOrderStatusPaid && order.Amount > order.RefundMoney,
			IsCellCancel:        order.Status == constant.RechargeOrderStatusPending,
			IsCellReverseSettle: isCellReverseSettle,
		},
	}, nil
}

// GetRechargeOrderPaymentQrcode 获取支付方式的二维码信息
func (s *rechargeOrderSrv) GetRechargeOrderPaymentQrcode(ctx context.Context, req req.RechargeOrderPaymentQrcodeReq) (*resp.RechargeOrderPaymentQrcodeInfoResp, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取充值订单
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereUuid(req.RechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders(),
		rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
	)
	if order.Uuid == 0 || order.Status != constant.RechargeOrderStatusPending {
		return nil, errors.New("充值订单不存在")
	}

	// 判断当前是否连连支付
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethod := paymentMethodRepo.GetPaymentMethod(paymentMethodRepo.WhereUuid(req.PaymentMethodUuid))
	if paymentMethod.Uuid == 0 || !paymentMethod.IsLianLianPay() {
		return nil, errors.New("支付方式不可用")
	}

	// 验证支付配置
	paymentRepo := NewPaymentRepo(ctx, s.dbm)
	if err := paymentRepo.ValidateConfigError(ctx.GetCompanyUuid()); err != nil {
		return nil, errors.New("请先到支付管理中完善后使用")
	}

	// 判断支付方式是否已支付
	orderRepo := repository.NewPaymentOrderRepo(db)
	paymentOrder, err := orderRepo.GetPaymentOrderInfo(
		repository.CommonRepo.WhereBySoftDelete(),
		orderRepo.WhereRelatedUuid(order.Uuid),
		orderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeRechargeOrder),
		orderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
	)
	if err == nil {
		if paymentOrder.Status == constant.PaymentOrderStatusPaid {
			infoResp := &resp.RechargeOrderPaymentQrcodeInfoResp{
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
	sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
	sumPaymentAmountAddCash := utils.DecimalAdd(sumPaymentAmount, req.PaymentAmount)

	// 支付订单金额大于充值金额
	if sumPaymentAmount >= order.RechargeAmount {
		return nil, errors.New("当前已足额")
	}
	if paymentMethod.Code != constant.PaymentMethodCodeCash && sumPaymentAmountAddCash > order.RechargeAmount {
		return nil, errors.New("非现金支付不能大于应收")
	}

	// 创建连连支付订单
	payment, err := paymentRepo.CreatePayment(CreatePaymentReq{
		RelatedType:       constant.PaymentOrderRelatedTypeRechargeOrder,
		RelatedUuid:       order.Uuid,
		PaymentMethodUuid: paymentMethod.Uuid,
		PaymentMethodCode: paymentMethod.Code,
		PaymentAmount:     decimal.NewFromFloat(paymentAmount).Add(decimal.NewFromFloat(commissionFee)).InexactFloat64(),
		CommissionFee:     commissionFee,
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 在 infoResp 初始化之前添加
	infoResp := &resp.RechargeOrderPaymentQrcodeInfoResp{
		PaymentOrderUuid: payment.PaymentOrderUuid,
		QrCode:           payment.LinkUrl,
		QrCodeExpireSec:  payment.GetRemainingPayableTime(),
		Status:           payment.GetStatus(),
		PaymentAmount:    paymentAmount,
	}

	return infoResp, nil
}

// CancelRechargeOrder 取消充值订单
func (s *rechargeOrderSrv) CancelRechargeOrder(ctx context.Context, uuid uint64) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid), rechargeOrderRepo.WithPaymentOrders())
	if order.Uuid == 0 {
		return errors.New("充值订单不存在")
	}
	if order.Status != constant.RechargeOrderStatusPending {
		return errors.New("当前状态不可操作")
	}
	if len(order.PaymentOrders) > 0 {
		return errors.New("当前订单已被部分支付，不支持取消")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 修改状态为已取消
		err := rechargeOrderRepo.Update(order.Uuid, map[string]any{"status": constant.RechargeOrderStatusCanceled})
		if err != nil {
			return errors.WithMessage(err)
		}
		// 会员充值操作日志
		err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "取消",
			Action:            constant.RechargeOrderActionOrderCancel,
			RechargeOrderUuid: order.Uuid,
		})
		if err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if err != nil {
		return errors.ErrInternal
	}
	return nil
}

// CheckRechargeOrderReverseSettle 检查充值订单是否可以反结账
func (s *rechargeOrderSrv) CheckRechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrderReverseSettleInfo, error) {
	var reverseSettleInfoResp resp.RechargeOrderReverseSettleInfo
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid), rechargeOrderRepo.WithMember())

	if order.Uuid == 0 || order.Member == nil {
		return reverseSettleInfoResp, errors.New("充值订单不存在")
	}

	if order.RefundMoney > 0 {
		return reverseSettleInfoResp, errors.New("退款后不能反结账")
	}

	var (
		message string
		status  uint
	)
	pendingRechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending))
	if pendingRechargeOrder.Uuid > 0 {
		message = i18n.Translate(ctx.GetLanguage(), "当前充值有待付款订单，请先完成订单再进行反结账")
		status = 2
	}
	member := order.Member
	if member.GetBalance() < order.RechargeAmount || member.GetPoints() < order.GiftPoint || member.GetGiftBalance() < order.GiftAmount {
		message = i18n.Translate(ctx.GetLanguage(), "当前会员账户不足反结账")
		status = 1
	}
	return resp.RechargeOrderReverseSettleInfo{
		MemberInfo: resp.ReverseSettleRechargeOrderMemberInfo{
			Uuid:        member.Uuid,
			Nickname:    member.Nickname,
			Balance:     member.GetBalance(),
			GiftBalance: member.GetGiftBalance(),
			Points:      member.GetPoints(),
		},
		Status:  status,
		Message: message,
	}, nil
}

// RechargeOrderReverseSettle 反结账
func (s *rechargeOrderSrv) RechargeOrderReverseSettle(ctx context.Context, uuid uint64) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithMember(), rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())

	if order.Uuid == 0 {
		return errors.WithMessage(errors.New("充值订单不存在"))
	}

	if order.Status != constant.RechargeOrderStatusPaid {
		return errors.WithMessage(errors.New("当前状态不可操作"))
	}

	if order.RefundMoney > 0 {
		return errors.WithMessage(errors.New("退款后不能反结账"))
	}

	if order.Member == nil {
		return errors.WithMessage(errors.New("会员不存在"))
	}

	var memberPointsChanged bool
	err := db.Transaction(func(tx *gorm.DB) error {
		ctx.SetDB(tx)
		// 处理会员积分
		if order.GiftPoint > 0 {
			if err := s.memberSrv.HandleMemberPoints(ctx, MemberPointsChangeReq{
				Uuid:     order.MemberUuid,
				Points:   -order.GiftPoint,
				Scene:    constant.MemberPointLogSceneRechargeReverse,
				Describe: fmt.Sprintf("充值反结账：%s", order.OrderNo),
			}); err != nil {
				return errors.WithMessage(err)
			}
			memberPointsChanged = true
		}
		// 处理会员余额
		if order.RechargeAmount > 0 || order.GiftAmount > 0 {
			if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				MemberUuid:  order.MemberUuid,
				Money:       -order.RechargeAmount,
				GiftMoney:   -order.GiftAmount,
				Scene:       constant.MemberBalanceLogRechargeReverse,
				Describe:    fmt.Sprintf("充值反结账：%s", order.OrderNo),
				RelatedUuid: order.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}

		returnOrderAmounts := make([]model.ReturnOrderAmount, 0, len(order.PaymentOrders))
		// paymentOrderRepo := repository.NewPaymentOrderRepo(tx)

		// 退款现金金额
		var refundCashAmount float64
		var currencyUnit string
		for _, paymentOrder := range order.PaymentOrders {
			// 标记删除
			// if err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
			// 	"status":      constant.PaymentOrderStatusRefund,
			// 	"delete_time": time.Now().Unix(),
			// }); err != nil {
			// 	return errors.WithMessage(errors.ErrInternal, err.Error())
			// }

			amount := paymentOrder.Amount
			if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				amount = paymentOrder.PaymentAmount
				refundCashAmount = utils.DecimalSub(paymentOrder.Amount, order.ChargeDue)
			}
			returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
				PaymentMethodUuid: paymentOrder.PaymentMethodUuid,
				Amount:            amount,
			})

			currencyUnit = paymentOrder.CurrencyUnit
		}

		if err := repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "反结账",
			Action:            constant.RechargeOrderActionReverseSettle,
			Data:              s.getReverseSettleLogData(order),
			RechargeOrderUuid: order.Uuid,
		}); err != nil {
			return errors.WithMessage(err)
		}

		returnOrder, err := repository.NewReturnOrderRepo(tx).CreateReturnOrder(model.ReturnOrder{
			RelatedOrderType:    constant.ReturnOrderRelatedOrderTypeRechargeOrder, // 充值订单退款
			RelatedOrderUuid:    order.Uuid,
			RelatedOrderNo:      order.OrderNo,
			ReturnType:          constant.ReturnOrderRefundTypeTotal, // 整单退
			RefundAmount:        order.Amount,
			ReturnOrderAmounts:  returnOrderAmounts,
			IsReverseSettlement: 1,
			DutyNo:              ctx.GetStaff().DutyNo,
			Unit:                currencyUnit,
			RefundReason:        "反结账",
		})
		if err != nil {
			return errors.WithMessage(errors.ErrInternal, err.Error())
		}

		if refundCashAmount > 0 {
			if err = s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
				Amount:    -refundCashAmount,
				Scene:     constant.CashBoxLogSceneRefund,
				OrderUuid: returnOrder.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}

		if err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, map[string]any{
			"status":       constant.RechargeOrderStatusPending,
			"payment_time": 0,
			"amount":       0,
			"refund_money": 0,
			"charge_due":   0,
		}); err != nil {
			return errors.WithMessage(errors.ErrInternal)
		}

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
			err = erpSrv.CancelPosInvoice(ctx, req.CancelPosInvoiceReq{
				ProductsInvoiceName: order.ErpProductsInvoiceName,
				OpenPosEntryName:    shiftLog.ErpnextOpenPosEntryName,
				OrderNo:             order.OrderNo, //异步模式必填
			})
			if err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	})

	if err != nil {
		return errors.WithMessage(err)
	}

	if memberPointsChanged {
		utils.Go(func() {
			s.memberSrv.HandleMemberUpgrade(ctx.GetCompanyUuid(), order.MemberUuid)
		})
	}

	// 发送短信
	{
		// 获取最新的会员信息
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(order.Member.Uuid)
		if err != nil {
			ctx.Log().Info("停止发送短信（充值反结账），获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			utils.Go(func() {
				if member != nil && member.Phone != "" {
					smsReq := sms.MemberRechargeRefundRequest{
						Company:        ctx.GetCompany().Name,
						RechargeRefund: decimal.NewFromFloat(order.RechargeAmount).Add(decimal.NewFromFloat(order.GiftAmount)).Truncate(2).InexactFloat64(),
						Balance:        member.GetBalanceAll(),
						PointsBalance:  member.GetPoints(),
					}
					if err := s.smsSrv.SendMemberRechargeRefundSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送充值反结账短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送充值反结账短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			})
		}
	}

	// 发布“统计”事件
	utils.Go(func() {
		s.bus.PublishStatisticsMemberEvent(event.StatisticsMemberPayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			MemberRechargeOrderUuid: order.Uuid,
			OnlyDelete:              true,
		})
	})

	utils.Go(func() {
		// 发布“会员余额变动”事件
		s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
			BasePayload: event.BasePayload{ // 会员余额变动
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
		// 发布“会员积分变动”事件
		s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
			BasePayload: event.BasePayload{ // 会员积分变动
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})

	return nil
}

// 组装支付记录
func (s *rechargeOrderSrv) getPaymentRecords(paymentOrders []model.PaymentOrder, returnOrders []model.ReturnOrder) []resp.RefundRechargeOrderPaymentRecord {
	// 根据支付方式source和code排序支付订单
	sort.Slice(paymentOrders, func(i, j int) bool {
		if paymentOrders[i].PaymentMethod.Source != paymentOrders[j].PaymentMethod.Source {
			return paymentOrders[i].PaymentMethod.Source < paymentOrders[j].PaymentMethod.Source
		}
		if paymentOrders[i].PaymentMethod.Code != paymentOrders[j].PaymentMethod.Code {
			return paymentOrders[i].PaymentMethod.Code < paymentOrders[j].PaymentMethod.Code
		}
		return paymentOrders[i].Uuid < paymentOrders[j].Uuid
	})
	// 组装支付记录
	paymentRecords := make([]resp.RefundRechargeOrderPaymentRecord, 0, len(paymentOrders))
	for _, paymentOrder := range paymentOrders {
		refundableAmount := paymentOrder.Amount
		paymentAmount := paymentOrder.Amount
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash { // 现金支付，去掉找零
			refundableAmount = paymentOrder.PaymentAmount
			paymentAmount = paymentOrder.PaymentAmount
		}
		for _, returnOrder := range returnOrders { // 减去退款金额
			for _, amount := range returnOrder.ReturnOrderAmounts {
				if amount.PaymentMethodUuid == paymentOrder.PaymentMethodUuid {
					refundableAmount = utils.DecimalSub(refundableAmount, amount.Amount)
				}
			}
		}
		paymentRecords = append(paymentRecords, resp.RefundRechargeOrderPaymentRecord{
			PaymentOrderUuid:  paymentOrder.Uuid,
			PaymentMethodUuid: paymentOrder.PaymentMethodUuid,
			PaymentMethodCode: paymentOrder.PaymentMethod.Code,
			PaymentName:       paymentOrder.PaymentMethodName,
			PaymentAmount:     paymentAmount,
			RefundableAmount:  refundableAmount,
			CurrencyUnit:      paymentOrder.CurrencyUnit,
		})
	}

	return paymentRecords
}

// GetRechargeOrderRefundInfo 获取退款信息
func (s *rechargeOrderSrv) GetRechargeOrderRefundInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderRefundInfo, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WithStaff(), rechargeOrderRepo.WithMember(), rechargeOrderRepo.WithReturnOrders(), rechargeOrderRepo.WithReturnOrderAmount())

	if order.Uuid == 0 {
		return resp.RechargeOrderRefundInfo{}, errors.New("充值订单不存在")
	}

	paymentRecords := s.getPaymentRecords(order.PaymentOrders, order.ReturnOrders)
	return resp.RechargeOrderRefundInfo{
		Uuid:             order.Uuid,
		RefundableAmount: utils.DecimalSub(order.Amount, order.RefundMoney),
		RechargeAmount:   order.RechargeAmount,
		GiftAmount:       order.GiftAmount,
		GiftPoint:        order.GiftPoint,
		RechargeMemberInfo: resp.RefundRechargeOrderMemberInfo{
			Balance:     order.Member.GetBalance(),
			GiftBalance: order.Member.GetGiftBalance(),
			Points:      order.Member.GetPoints(),
		},
		PaymentRecords: paymentRecords,
	}, nil
}

// RechargeOrderRefund 退款
func (s *rechargeOrderSrv) RechargeOrderRefund(ctx context.Context, refundReq req.RechargeOrderRefundReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereUuid(refundReq.Uuid),
		rechargeOrderRepo.WithPaymentOrders(),
		rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WithStaff(),
		rechargeOrderRepo.WithMember(),
		rechargeOrderRepo.WithReturnOrders(),
		rechargeOrderRepo.WithReturnOrderAmount(),
	)

	if order.Uuid == 0 {
		return errors.New("充值订单不存在")
	}
	if order.Status != constant.RechargeOrderStatusPaid {
		return errors.New("该订单不合法")
	}
	if order.Member == nil {
		return errors.New("用户不存在")
	}
	if refundReq.RefundType == constant.ReturnOrderRefundTypeTotal { // 整单退款
		refundReq.RefundMoney = utils.DecimalSub(order.Amount, order.RefundMoney)
	}
	if refundReq.RefundMoney <= 0 {
		return errors.New("退款金额错误")
	}
	paymentRecords := s.getPaymentRecords(order.PaymentOrders, order.ReturnOrders)
	// 可退款金额
	var refundableAmount float64
	var currencyUnit string
	for _, record := range paymentRecords {
		refundableAmount = utils.DecimalAdd(refundableAmount, record.RefundableAmount)
		currencyUnit = record.CurrencyUnit
	}
	if refundableAmount <= 0 {
		return errors.New("无法退款")
	}
	// 要退款金额
	refundMoney := refundReq.RefundMoney
	if refundMoney > refundableAmount {
		return errors.New("退款金额不能大于实付金额")
	}
	// 本次退款要从会员主账户扣除的金额
	deductionMoney := decimal.NewFromFloat(order.RechargeAmount).Sub(decimal.NewFromFloat(order.RefundAmount)).Truncate(2).InexactFloat64()
	if refundMoney < deductionMoney {
		deductionMoney = refundMoney
	}
	// 要从会员主账户扣除的金额 > 用户主账户余额，不可退款
	if deductionMoney > order.Member.GetBalance() {
		return errors.New("当前会员主账户余额不足以退款")
	}
	// 处理退款金额
	var returnOrderAmounts []model.ReturnOrderAmount
	var refundCashMoney float64
	for _, record := range paymentRecords {
		if record.RefundableAmount == 0 {
			continue
		}
		// 遍历所有支付方式，如果小于扣款金额，则继续扣款，否则跳过
		if record.RefundableAmount < refundMoney {
			returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
				BaseModel: model.BaseModel{
					Uuid: func() uint64 {
						id, _ := utils.GetID()
						return id
					}(),
				},
				PaymentMethodUuid:     record.PaymentMethodUuid,
				Amount:                record.RefundableAmount,
				PaymentOrderUuid:      record.PaymentOrderUuid,
				MerchantRefundOrderNo: utils.GenerateMerchantOrderNo("RE"),
				PaymentMethod:         &model.PaymentMethod{Code: record.PaymentMethodCode, PaymentName: record.PaymentName},
			})
			refundMoney = utils.DecimalSub(refundMoney, record.RefundableAmount)
			if record.PaymentMethodCode == constant.PaymentMethodCodeCash {
				refundCashMoney = record.RefundableAmount
			}
		} else {
			returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
				BaseModel: model.BaseModel{
					Uuid: func() uint64 {
						id, _ := utils.GetID()
						return id
					}(),
				},
				PaymentMethodUuid:     record.PaymentMethodUuid,
				Amount:                refundMoney,
				PaymentOrderUuid:      record.PaymentOrderUuid,
				MerchantRefundOrderNo: utils.GenerateMerchantOrderNo("RE"),
				PaymentMethod:         &model.PaymentMethod{Code: record.PaymentMethodCode, PaymentName: record.PaymentName},
			})
			if record.PaymentMethodCode == constant.PaymentMethodCodeCash {
				refundCashMoney = refundMoney
			}
			break
		}
	}

	// 创建退货单
	returnOrder := model.ReturnOrder{
		BaseModel: model.BaseModel{
			Uuid: func() uint64 {
				id, _ := utils.GetID()
				return id
			}(),
			CreateTime: time.Now().Unix(),
		},
		RelatedOrderType:   constant.ReturnOrderRelatedOrderTypeRechargeOrder,
		RelatedOrderUuid:   order.Uuid,
		RelatedOrderNo:     order.OrderNo,
		ReturnType:         refundReq.RefundType,
		RefundAmount:       refundReq.RefundMoney,
		BankCode:           refundReq.BankCode,
		AccountNo:          refundReq.AccountNo,
		AccountName:        refundReq.AccountName,
		ReturnOrderAmounts: returnOrderAmounts, // 关联创建退款金额
		DutyNo:             ctx.GetStaff().DutyNo,
		Unit:               currencyUnit,
		RefundReason:       "退款",
	}

	// 是否存在QrPromptPay支付
	if returnOrder.IsExistQrPromptPay() {
		if refundReq.BankCode == "" || refundReq.AccountNo == "" || refundReq.AccountName == "" {
			return errors.NewWithCode(constant.CodeReturnOrderBank, "请选择银行")
		}
	}
	refundPayTypes := make([]event.RefundPayType, 0)
	for _, amount := range returnOrderAmounts {
		for _, paymentRecord := range paymentRecords {
			if amount.PaymentMethodUuid == paymentRecord.PaymentMethodUuid {
				refundPayTypes = append(refundPayTypes, event.RefundPayType{
					Name:              paymentRecord.PaymentName,
					Code:              paymentRecord.PaymentMethodCode,
					Amount:            amount.Amount,
					RefundStatus:      amount.RefundStatus,
					ReturnAmountUuid:  amount.BaseModel.Uuid,
					ReturnOrderUuid:   returnOrder.BaseModel.Uuid,
					PaymentOrderUuid:  amount.PaymentOrderUuid,
					PaymentMethodUuid: amount.PaymentMethodUuid,
				})
				break
			}
		}
	}

	lianLianPayCount := returnOrder.GetLianLianPayCount()

	var isExistCashPay bool

	err := db.Transaction(func(tx *gorm.DB) error {
		ctx.SetDB(tx)

		// 更新充值订单
		err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, map[string]any{
			"refund_money":  utils.DecimalAdd(order.RefundMoney, refundReq.RefundMoney),
			"refund_amount": utils.DecimalAdd(order.RefundAmount, deductionMoney),
		})
		if err != nil {
			return errors.ErrInternal
		}

		// 添加操作日志
		err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "退款",
			Action:            constant.RechargeOrderActionRefund,
			Data:              s.getRefundData(refundReq.RefundType, refundReq.RefundMoney, refundPayTypes),
			RechargeOrderUuid: order.Uuid,
		})
		if err != nil {
			return errors.ErrInternal
		}

		var paymentOrderUuid uint64
		// 创建退货单
		if paymentOrderUuid, err = repository.NewReturnOrderRepo(tx).CreateReturnOrderRecord(returnOrder); err != nil {
			return errors.WithMessage(err)
		}

		// 创建连连退款订单
		for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
			if lianLianPayCount > 0 && returnOrderAmount.PaymentMethod.IsLianLianPay() {
				paymentServiceRefundReq := PaymentServiceRefundReq{
					RelatedType:           constant.PaymentOrderRelatedTypeRechargeOrder,
					PaymentOrderUuid:      returnOrderAmount.PaymentOrderUuid,
					MerchantRefundOrderNo: returnOrderAmount.MerchantRefundOrderNo,
					RefundAmount:          returnOrderAmount.Amount,
					BankCode:              returnOrder.BankCode,
					AccountNo:             returnOrder.AccountNo,
					AccountName:           returnOrder.AccountName,
				}
				if lianLianPayCount > 1 {
					utils.Go(func() {
						payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
						if err != nil {
							returnOrderAmount.RefundStatus = 2
							returnOrderAmount.LlReturnOrderid = "0"
						} else {
							returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
						}
						// 更新退款状态
						returnOrderRepo := repository.NewReturnOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
						err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{
							returnOrderRepo.WhereUuid(returnOrderAmount.Uuid),
						}, returnOrderAmount)
						if err != nil {
							fmt.Println("更新退款状态失败", err)
							logger.Logger.Error("更新退款状态失败", zap.Error(err))
						}
					})
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
			returnOrderAmount.ReturnOrderUuid = paymentOrderUuid
			// 创建退款金额
			if err = repository.NewReturnOrderRepo(db).CreateReturnOrderAmount([]model.ReturnOrderAmount{returnOrderAmount}); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 退还余额
		if deductionMoney > 0 {
			err = s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				MemberUuid:  order.MemberUuid,
				Money:       -deductionMoney,
				Scene:       constant.MemberBalanceLogRechargeRefund,
				Describe:    fmt.Sprintf("退款：%s", order.OrderNo),
				RelatedUuid: order.Uuid,
			})
			if err != nil {
				return errors.WithMessage(err)
			}
		}

		// 退还现金
		if refundCashMoney > 0 {
			isExistCashPay = true
			if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
				Amount:    -refundCashMoney,
				Scene:     constant.CashBoxLogSceneRefund,
				OrderUuid: returnOrder.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 退款发票到erp
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			res, err := s.ReturnPosInvoice(ctx, &order, &returnOrder, tx, refundReq.RefundType)
			if err != nil {
				return errors.WithMessage(err)
			}
			returnOrder.ErpInvoiceName = res.InvoiceName
			if err := repository.NewReturnOrderRepo(tx).UpdateReturnOrderRecordErpInvoiceName(returnOrder.Uuid, returnOrder.ErpInvoiceName); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	// 发送短信
	if deductionMoney > 0 {
		utils.Go(func() {
			// 获取最新的会员信息
			member, err := repository.NewMemberRepo(db).GetMemberByUuid(order.Member.Uuid)
			if err != nil {
				ctx.Log().Info("停止发送短信（充值退款），获取会员失败", zap.Error(errors.WithMessage(err)))
			} else {
				ctx.SetDB(db)
				if member != nil {
					smsReq := sms.MemberRechargeRefundRequest{
						Company:        ctx.GetCompany().Name,
						RechargeRefund: deductionMoney,
						Balance:        member.GetBalanceAll(),
						PointsBalance:  member.GetPoints(),
					}
					if err := s.smsSrv.SendMemberRechargeRefundSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送充值退款短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送充值退款短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			}
		})
	}

	// 发布“统计”事件
	utils.Go(func() {
		s.bus.PublishStatisticsMemberEvent(event.StatisticsMemberPayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			MemberRechargeOrderUuid: order.Uuid,
		})
	})

	if isExistCashPay {
		return errors.NewWithCode(constant.CodeSuccessOpenCashBox, "请求成功")
	}

	return nil
}

// RechargeOrderReReturnOrder 重新退款
func (s *rechargeOrderSrv) RechargeOrderReReturnOrder(ctx context.Context, req req.RechargeOrderReReturnReq) error {
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
		return errors.New("找不到订单")
	}
	if orderAmount.RefundStatus == 1 {
		return errors.New("该订单已成功退款，无法重复退款")
	}
	if !orderAmount.PaymentMethod.IsLianLianPay() {
		return errors.New("该订单无法重新退款")
	}
	// 判断订单是否正在退款
	if orderAmount.RefundStatus == 0 {
		return errors.New("该订单正在进行退款，无法重复操作")
	}

	refundReq := PaymentServiceRefundReq{
		RelatedType:           constant.PaymentOrderRelatedTypeRechargeOrder,
		PaymentOrderUuid:      orderAmount.PaymentOrderUuid,
		MerchantRefundOrderNo: orderAmount.MerchantRefundOrderNo,
		RefundAmount:          orderAmount.Amount,
		RefundOrderId:         orderAmount.LlReturnOrderid,
	}

	// 是否存在QrPromptPay支付
	isChangeBankCode := false
	if orderAmount.PaymentMethod.IsQrPromptPay() {
		if req.BankCode == "" || req.AccountNo == "" || req.AccountName == "" {
			return errors.NewWithCode(constant.CodeReturnOrderBank, "请选择银行")
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
		return errors.WithMessage(err)
	}
	if refund.RefundStatus == "RP" {
		return errors.New("该订单正在进行退款，无法重复操作")
	}
	if refund.RefundStatus == "RS" {
		orderAmount.RefundStatus = 1
		err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
		if err != nil {
			return errors.WithMessage(err)
		}
		return errors.New("该订单已成功退款，无法重复退款")
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
		return errors.WithMessage(err)
	}
	// 更新退款订单号
	orderAmount.LlReturnOrderid = refund.RefundOrderId
	err = returnOrderRepo.UpdateReturnOrder([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.ReturnOrder.Uuid)}, *orderAmount.ReturnOrder)
	if err != nil {
		return errors.WithMessage(err)
	}
	err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
	if err != nil {
		return errors.WithMessage(err)
	}
	//
	return nil
}
