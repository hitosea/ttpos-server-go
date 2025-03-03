package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	errors2 "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IRechargeOrderSrv 定义充值订单服务接口
type IRechargeOrderSrv interface {
	GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder                                                                    // 获取进行中的会员充值订单
	CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error)                                 // 创建充值订单
	AddPaymentMethod(ctx context.Context, addPaymentMethod req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error)          // 充值订单添加支付方式
	CancelPaymentMethod(ctx context.Context, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error) // 充值订单撤销支付方式
	ConfirmRechargeOrder(ctx context.Context, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error)    // 确认充值订单
	PrintTicket(ctx context.Context, printRechargeOrderReq req.PrintRechargeOrderReq) (resp.PrinterLogData, error)                    // 打印充值订单
	GetRechargeOrderList(ctx context.Context, listReq req.RechargeOrderListReq) (resp.RechargeOrderList, error)                       // 充值订单列表
	GetRechargeOrderInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderInfo, error)                                            // 获取充值订单详情
	CancelRechargeOrder(ctx context.Context, uuid uint64) error                                                                       // 取消充值订单
	GetRechargeOrderRefundInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderRefundInfo, error)                                // 获取退款信息
	CheckRechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrderReverseSettleInfo, error)                    // 检查反结账信息
	RechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrder, error)                                          // 充值订单反结账
}

type rechargeOrderSrv struct {
	dbm              *database.DBManager // 数据库管理器
	cache            cache.Cache
	paymentMethodSrv IPaymentMethodSrv
	settingSrv       setting.ISrv
	cashBoxSrv       ICashBoxSrv
	rechargePrintSrv IRechargePrintSrv
	memberSrv        IMemberSrv
}

func NewRechargeOrderSrv(dbm *database.DBManager, cache cache.Cache, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv, rechargePrintSrv IRechargePrintSrv, memberSrv IMemberSrv) IRechargeOrderSrv {
	return NewRechargeOrderSrvImpl(dbm, cache, paymentMethodSrv, settingSrv, cashBoxSrv, rechargePrintSrv, memberSrv)
}

func NewRechargeOrderSrvImpl(dbm *database.DBManager, cache cache.Cache, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv, rechargePrintSrv IRechargePrintSrv, memberSrv IMemberSrv) IRechargeOrderSrv {
	return &rechargeOrderSrv{
		dbm:              dbm,
		cache:            cache,
		paymentMethodSrv: paymentMethodSrv,
		settingSrv:       settingSrv,
		cashBoxSrv:       cashBoxSrv,
		rechargePrintSrv: rechargePrintSrv,
		memberSrv:        memberSrv,
	}
}

// GetPendingRechargeOrder 进行中的会员充值订单
func (s *rechargeOrderSrv) GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	// 进行中的充值订单
	rechargeOrder := rechargeOrderRepo.GetRechargeOrder(
		rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	if rechargeOrder.Uuid == 0 {
		return resp.RechargeOrder{PaymentOrders: make([]resp.PaymentOrder, 0)}
	}

	paymentOrderCount := len(rechargeOrder.PaymentOrders)
	respPaymentOrders := make([]resp.PaymentOrder, 0, paymentOrderCount)
	if paymentOrderCount > 0 {
		var respPaymentOrder resp.PaymentOrder
		for _, paymentOrder := range rechargeOrder.PaymentOrders {
			copier.Copy(&respPaymentOrder, paymentOrder)
			respPaymentOrder.PaymentMethodCode = paymentOrder.PaymentMethod.Code
			respPaymentOrder.PaymentMethodName = paymentOrder.PaymentMethod.PaymentName
			respPaymentOrder.DisabledCancel = slices.Contains([]int{constant.PaymentMethodCodeLianLianWechatPay,
				constant.PaymentMethodCodeLianLianAliPay,
				constant.PaymentMethodCodeLianLianQRPromptPay}, paymentOrder.PaymentMethod.Code)
			respPaymentOrders = append(respPaymentOrders, respPaymentOrder)
		}
	}

	var respRechargeOrder resp.RechargeOrder
	copier.Copy(&respRechargeOrder, rechargeOrder)

	respRechargeOrder.PaymentOrders = respPaymentOrders
	return respRechargeOrder
}

// 计算支付订单初始金额
func (s *rechargeOrderSrv) sumPaymentAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentAmount
	}
	return sum
}

// 计算应收金额
func (s *rechargeOrderSrv) getRechargeOrderAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentAmount + paymentOrder.PaymentCommissionFee
	}
	return sum
}

// 计算找零
func (s *rechargeOrderSrv) getChargeDue(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.Amount - paymentOrder.PaymentCommissionFee - paymentOrder.PaymentAmount
	}
	return sum
}

// 计算支付订单实收金额
func (s *rechargeOrderSrv) getActualAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.Amount
	}
	return sum
}

// 计算支付订单手续费
func (s *rechargeOrderSrv) getPayFee(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentCommissionFee
	}
	return sum
}

// CreateRechargeOrder 创建充值订单
func (s *rechargeOrderSrv) CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
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
		rechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending),
			rechargeOrderRepo.WithPaymentOrders())
		if rechargeOrder.Uuid != 0 {
			oldRechargeAmount := rechargeOrder.Amount

			sumPaymentAmount := s.sumPaymentAmount(rechargeOrder.PaymentOrders)
			if rechargeReq.RechargeAmount < sumPaymentAmount {
				return orderResp, errors.New("充值金额不能小于已充值金额")
			}
			if err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
				// 更新充值订单信息
				err := repository.NewMemberRechargeOrderRepo(tx).Update(rechargeOrder.Uuid, map[string]any{
					"recharge_amount": rechargeReq.RechargeAmount,
					"gift_amount":     rechargeReq.GiftAmount,
					"gift_point":      rechargeReq.GiftPoint,
					"member_uuid":     rechargeReq.MemberUuid,
					"staff_uuid":      ctx.GetStaffUuid(),
				})
				// 会员充值操作日志
				operationData, _ := json.Marshal(map[string]any{
					"recharge_money":     rechargeReq.RechargeAmount,
					"old_recharge_money": oldRechargeAmount,
				})
				err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
					OperatorName:      ctx.GetStaff().RealName,
					OperatorEmail:     ctx.GetStaff().Username,
					Client:            ctx.GetSource(),
					Message:           "变更充值金额",
					Action:            constant.RechargeOrderActionChangeAmount,
					Data:              string(operationData),
					RechargeOrderUuid: rechargeOrder.Uuid,
				})
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				ctx.Log().Error("修改充值订单失败", zap.Error(err))
				return orderResp, errors.New("修改充值订单失败")
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
			RechargeAmount: rechargeReq.RechargeAmount,
			GiftAmount:     rechargeReq.GiftAmount,
			GiftPoint:      rechargeReq.GiftPoint,
			MemberUuid:     rechargeReq.MemberUuid,
			StaffUuid:      ctx.GetStaffUuid(),
		})
		if err != nil {
			return err
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
			return err
		}
		return nil
	})
	if err != nil {
		return orderResp, errors.New("创建充值订单失败")
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// AddPaymentMethod 充值订单添加支付方式
func (s *rechargeOrderSrv) AddPaymentMethod(ctx context.Context, addReq req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(addReq.RechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders())
	if rechargeOrder.Uuid == 0 || rechargeOrder.Status != constant.RechargeOrderStatusPending {
		return orderResp, errors.New("充值订单不存在")
	}

	// todo 在线支付订单暂不处理

	// 根据Uuid 获取支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(companyUuid))
	paymentMethod := paymentMethodRepo.GetPaymentMethod(paymentMethodRepo.WhereUuid(addReq.PaymentMethodUuid))
	if paymentMethod.Uuid == 0 {
		return orderResp, errors.New("支付方式不存在")
	}
	// 支付方式是否可用
	if paymentMethod.IsShowMemberRecharge == 0 || !s.paymentMethodSrv.IsEnabled(ctx, paymentMethod, addReq.CompanySetting) {
		return orderResp, errors.New("支付方式未开启")
	}
	if paymentMethod.Code == constant.PaymentMethodCodeBalance {
		return orderResp, errors.New("不能使用余额支付充值")
	}

	// 计算支付手续费
	paymentCommissionFee := s.paymentMethodSrv.CalculatePaymentCommissionFee(paymentMethod, addReq.PaymentAmount)

	sumPaymentAmount := s.sumPaymentAmount(rechargeOrder.PaymentOrders)
	// 支付订单金额大于充值金额
	if sumPaymentAmount >= rechargeOrder.RechargeAmount {
		return orderResp, errors.New("当前已足额")
	}
	sumPaymentAmountAddCash := sumPaymentAmount + addReq.PaymentAmount
	if paymentMethod.Code != constant.PaymentMethodCodeCash && sumPaymentAmountAddCash > rechargeOrder.RechargeAmount {
		return orderResp, errors.New("非现金支付不能大于应收")
	}

	// 支付订单总金额 = 支付金额 + 支付手续费
	amount := addReq.PaymentAmount + paymentCommissionFee

	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	// 获取已存在的该支付方式的支付订单
	paymentOrder, _ := paymentOrderRepo.GetPaymentOrder(
		paymentOrderRepo.WhereRelatedUuid(rechargeOrder.Uuid), paymentOrderRepo.WherePaymentTypeUuid(paymentMethod.Uuid))

	var rechargeAmountLeft float64
	var cashPaidPaymentAmount float64
	if paymentMethod.Code == constant.PaymentMethodCodeCash {
		// 如果现金支付订单存在
		if paymentOrder.Uuid != 0 {
			cashPaidPaymentAmount = paymentOrder.PaymentAmount
		}
		rechargeAmountLeft = rechargeOrder.RechargeAmount - (sumPaymentAmount - cashPaidPaymentAmount)
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
			return orderResp, errors.New("添加支付方式失败")
		}
	} else { // 不存在则创建
		currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
		if err != nil {
			logger.Logger.Error("添加支付订单-获取货币设置失败", zap.Error(err))
			return orderResp, errors.New("添加支付方式失败")
		}
		_, err = paymentOrderRepo.Create(model.PaymentOrder{
			PaymentMethodName:    paymentMethod.PaymentName,
			PaymentMethodUuid:    paymentMethod.Uuid,
			PaymentFeePercent:    paymentMethod.FeePercent,
			RelatedType:          constant.PaymentOrderRelatedTypeRechargeOrder,
			RelatedUuid:          rechargeOrder.Uuid,
			CurrencyUnit:         currencySetting.Unit, // 留档使用
			PaymentAmount:        addReq.PaymentAmount,
			PaymentCommissionFee: paymentCommissionFee,
			Amount:               amount,
			Status:               constant.PaymentOrderStatusPaid, // ToDo 手动添加在线支付标为0，处理lianlianpay
		})
		if err != nil {
			logger.Logger.Error("添加支付订单-创建支付订单", zap.Error(err))
			return orderResp, errors.New("添加支付方式失败")
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
		return orderResp, errors2.ErrInternal
	}

	// 更新现金支付订单
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(paymentOrder.RelatedUuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	var cashPaymentOrder model.PaymentOrder
	var sumPaymentAmount float64
	for _, order := range rechargeOrder.PaymentOrders {
		if order.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			cashPaymentOrder = order
		}
		sumPaymentAmount = sumPaymentAmount + order.PaymentAmount
	}
	if cashPaymentOrder.Uuid != 0 {
		paymentOrderRepo.Update(cashPaymentOrder.Uuid, map[string]any{
			"payment_amount": rechargeOrder.RechargeAmount - sumPaymentAmount,
		})
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// 计算非现金支付订单金额(payment_amount)
func (s *rechargeOrderSrv) sumPaymentAmountExcludeCash(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		if paymentOrder.PaymentMethod.Code != constant.PaymentMethodCodeCash {
			sum = sum + paymentOrder.PaymentAmount
		}
	}
	return sum
}

// ConfirmRechargeOrder 确认充值订单
func (s *rechargeOrderSrv) ConfirmRechargeOrder(ctx context.Context, confirmReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error) {
	var confirmResp resp.ConfirmRechargeOrder

	companyUuid := ctx.GetCompanyUuid()
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetMember(memberRepo.WhereUuid(confirmReq.MemberUuid))
	if member.Uuid == 0 {
		return confirmResp, errors.New("会员不存在")
	}

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(confirmReq.RechargeOrderUuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
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

	// 更新充值订单
	updates := map[string]any{
		"amount":       s.getRechargeOrderAmount(order.PaymentOrders), // 应收金额
		"status":       constant.RechargeOrderStatusPaid,              // 状态,0-pending待支付 1-paid已支付 2-canceled已取消 3-exp已过期
		"payment_time": time.Now().Unix(),
		"charge_due":   s.getChargeDue(order.PaymentOrders), // 找零
	}

	var memberPointsChanged bool

	err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
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
				return err
			}
			memberPointsChanged = true
		}

		// 处理会员余额
		if order.RechargeAmount > 0 || order.GiftAmount > 0 {
			if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				Uuid:           member.Uuid,
				RechargeAmount: order.RechargeAmount,
				GiftAmount:     order.GiftAmount,
				Scene:          constant.MemberBalanceLogRecharge,
				Describe:       fmt.Sprintf("收银机管理员操作 [%s]", ctx.GetStaff().RealName),
			}); err != nil {
				return err
			}
		}

		if sumPaymentAmount > sumPaymentAmountExcludeCash {
			if err := s.cashBoxSrv.UpdateBalance(ctx, constant.CashBoxLogTypeIn, sumPaymentAmount-sumPaymentAmountExcludeCash, order.Uuid); err != nil {
				return err
			}
		}

		rechargeOperation := map[string]any{
			"order_price":    order.RechargeAmount,             //  订单金额
			"recharge_money": order.RechargeAmount,             //  充值金额
			"pay_price":      order.Amount,                     //  订单应付
			"gift_money":     order.GiftAmount,                 //  赠送金额
			"gift_point":     order.GiftPoint,                  //  赠送积分
			"pay_fee":        s.getPayFee(order.PaymentOrders), //  支付手续费
			"change_due":     order.ChargeDue,                  //  找零
			"pay_type":       order.PaymentOrders,              //  支付方式
		}

		operationData, _ := json.Marshal(rechargeOperation)

		if err := repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "充值",
			Action:            constant.RechargeOrderActionRecharge,
			Data:              string(operationData),
			RechargeOrderUuid: order.Uuid,
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return confirmResp, err
	}

	if memberPointsChanged {
		go s.memberSrv.HandleMemberUpgrade(companyUuid, member.Uuid)
	}

	// 打印充值单
	go func() {
		_, err = s.rechargePrintSrv.PrintTicket(ctx, PrinterTicketReq{
			RechargeOrder: order,
			IsQueue:       false,
			DeviceId:      ctx.GetDeviceSn(),
			PrintLang:     ctx.GetLanguage(),
		})
	}()

	return s.confirmRechargeOrderResp(companyUuid, order.Uuid), nil
}

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

func (s *rechargeOrderSrv) PrintTicket(ctx context.Context, printRechargeOrderReq req.PrintRechargeOrderReq) (resp.PrinterLogData, error) {
	var res resp.PrinterLogData
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(printRechargeOrderReq.RechargeOrderUuid))
	if order.Uuid == 0 {
		return res, errors.New("充值订单不存在")
	}
	printLang := printRechargeOrderReq.PrintLang
	if printLang == "" {
		printLang = ctx.GetLanguage()
	}
	printerData, err := s.rechargePrintSrv.PrintTicket(ctx, PrinterTicketReq{
		RechargeOrder: order,
		IsQueue:       false,
		DeviceId:      ctx.GetDeviceSn(),
		PrintLang:     printLang,
	})
	if err != nil {
		return res, err
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

func (s *rechargeOrderSrv) GetRechargeOrderList(ctx context.Context, listReq req.RechargeOrderListReq) (resp.RechargeOrderList, error) {

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	var dbOptions []repository.DBOption
	if listReq.OrderNo != "" {
		dbOptions = append(dbOptions, rechargeOrderRepo.WhereOrderNoLike(listReq.OrderNo))
	}
	if listReq.Status != -1 {
		dbOptions = append(dbOptions, rechargeOrderRepo.WhereStatus(listReq.Status))
	}

	if listReq.DateType >= 0 && listReq.DateType <= 3 {
		now := time.Now()
		var startTime, endTime time.Time
		switch listReq.DateType {
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
		dbOptions = append(dbOptions, rechargeOrderRepo.WhereCreateTimeBetween(startTime.Unix(), endTime.Unix()))
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

		dbOptions = append(dbOptions, rechargeOrderRepo.WhereTimeBetween(repository.TimeQueryParams{
			TimeRanges: timeRanges,
			Operator:   "OR",
		}))
	}

	// 关联查询
	dbOptions = append(dbOptions, rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	rechargeOrders, total, err := rechargeOrderRepo.PaginateGetRechargeOrder(listReq.PageNo, listReq.PageSize, dbOptions...)

	if err != nil {
		return resp.RechargeOrderList{}, errors2.ErrInternal
	}

	items := make([]resp.RechargeOrderItem, 0, len(rechargeOrders))

	for _, order := range rechargeOrders {

		var paymentMethods []string
		for _, paymentOrder := range order.PaymentOrders {
			paymentMethods = append(paymentMethods, paymentOrder.PaymentMethodName)
		}

		items = append(items, resp.RechargeOrderItem{
			Uuid:           order.Uuid,
			OrderNo:        order.OrderNo,
			Status:         order.Status,
			PaymentTime:    order.PaymentTime,
			RechargeAmount: order.RechargeAmount,
			Amount:         order.Amount,
			PaymentMethods: paymentMethods,
			Extra: resp.RechargeOrderItemExtra{
				IsCellRefund:        order.Status == constant.RechargeOrderStatusPaid,
				IsCellCancel:        order.Status == constant.RechargeOrderStatusPending,
				IsCellReverseSettle: order.Status == constant.RechargeOrderStatusPaid, // ToDo 同一个收银员工、同一个班次、已支付、未退款
			},
		})
	}

	// 获取数量
	getOrderNum := func(status int) int64 {
		num, _ := rechargeOrderRepo.GetOrderCount(rechargeOrderRepo.WhereStatus(status))
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
			UnpaidNum:   getOrderNum(0),
			CompleteNum: getOrderNum(1),
			CancelNum:   getOrderNum(2),
		},
	}, nil
}

func (s *rechargeOrderSrv) getActionDescription(log model.MemberRechargeOrderOperationLog) string {
	return ""
}

func (s *rechargeOrderSrv) GetRechargeOrderInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderInfo, error) {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithMember(), rechargeOrderRepo.WithStaff(), rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithRechargeOrderOperationLogs())
	if order.Uuid == 0 {
		return resp.RechargeOrderInfo{}, errors.New("充值订单不存在")
	}
	var cashierName string
	if order.Staff != nil {
		cashierName = order.Staff.RealName
	}

	var paymentMethods []resp.RechargeOrderPaymentMethod
	for _, paymentOrder := range order.PaymentOrders {
		paymentMethods = append(paymentMethods, resp.RechargeOrderPaymentMethod{
			Name:  paymentOrder.PaymentMethodName,
			Price: paymentOrder.Amount,
		})
	}

	var logs []resp.RechargeOrderOperationLogItem
	for _, log := range order.RechargeOrderOperationLogs {
		logs = append(logs, resp.RechargeOrderOperationLogItem{
			RealName:    log.OperatorName,
			Username:    log.OperatorEmail,
			Client:      log.Client,
			CreateTime:  log.CreateTime,
			Description: s.getActionDescription(log),
		})
	}
	return resp.RechargeOrderInfo{
		Uuid:    order.Uuid,
		OrderNo: order.OrderNo,
		Member: resp.RechargeOrderMember{
			Uuid:     order.MemberUuid,
			Nickname: order.Member.Nickname,
		},
		Status: order.Status,
		Cashier: resp.RechargeOrderCashier{
			RealName: cashierName,
		},
		RechargeAmount: order.RechargeAmount,
		Amount:         order.Amount,
		ChargeDue:      order.ChargeDue,
		PaymentTime:    order.PaymentTime,
		CreateTime:     order.CreateTime,
		GiftAmount:     order.GiftAmount,
		GiftPoint:      order.GiftPoint,
		PaymentMethods: paymentMethods,
		OperationLog:   resp.RechargeOrderOperationLog{}, // ToDo 处理日志
		Extra: resp.RechargeOrderItemExtra{
			IsCellRefund:        order.Status == constant.RechargeOrderStatusPaid,
			IsCellCancel:        order.Status == constant.RechargeOrderStatusPending,
			IsCellReverseSettle: order.Status == constant.RechargeOrderStatusPaid, // ToDo 同一个收银员工、同一个班次、已支付、未退款
		},
	}, nil
}

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
			return err
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
			return err
		}
		return nil
	})
	if err != nil {
		return errors2.ErrInternal
	}
	return nil
}

func (s *rechargeOrderSrv) GetRechargeOrderRefundInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderRefundInfo, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
		rechargeOrderRepo.WithStaff(), rechargeOrderRepo.WithMember())

	if order.Uuid == 0 {
		return resp.RechargeOrderRefundInfo{}, errors.New("充值订单不存在")
	}

	return resp.RechargeOrderRefundInfo{
		Uuid:             order.Uuid,
		RefundableAmount: order.Amount,
		RechargeAmount:   order.RechargeAmount,
		GiftAmount:       order.GiftAmount,
		GiftPoint:        order.GiftPoint,
		RechargeMemberInfo: resp.RefundRechargeOrderMemberInfo{
			Balance:     order.Member.Balance,
			GiftBalance: order.Member.GiftBalance,
			Points:      order.Member.Point,
		},
		PaymentRecords: make([]resp.RefundRechargeOrderPaymentRecord, 0), // ToDo 待充值订单退款等模型完善
	}, nil
}

func (s *rechargeOrderSrv) CheckRechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrderReverseSettleInfo, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid), rechargeOrderRepo.WithMember())

	if order.Uuid == 0 || order.Member == nil {
		return resp.RechargeOrderReverseSettleInfo{}, errors.New("充值订单不存在")
	}

	// ToDo 退款后不可反结账

	var toRechargeOrderUuid uint64
	var message string
	var isOccupy bool
	pendingRechargeOrder := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending))
	if pendingRechargeOrder.Uuid > 0 {
		toRechargeOrderUuid = pendingRechargeOrder.Uuid
		message = i18n.Translate(ctx.GetLanguage(), "当前充值有待付款订单，请先完成订单再进行反结账")
		isOccupy = true
	}

	member := order.Member
	return resp.RechargeOrderReverseSettleInfo{
		RechargeOrderUuid:   order.Uuid,
		ToRechargeOrderUuid: toRechargeOrderUuid,
		MemberInfo: resp.ReverseSettleRechargeOrderMemberInfo{
			Uuid:        member.Uuid,
			Nickname:    member.Nickname,
			Balance:     member.Balance,
			GiftBalance: member.GiftBalance,
			Points:      member.Point,
		},
		Message:  message,
		IsOccupy: isOccupy,
	}, nil
}

func (s *rechargeOrderSrv) RechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrder, error) {
	var res resp.RechargeOrder
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	order := rechargeOrderRepo.GetRechargeOrder(rechargeOrderRepo.WhereUuid(uuid),
		rechargeOrderRepo.WithMember(), rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())

	if order.Uuid == 0 {
		return res, errors.New("充值订单不存在")
	}

	if order.Status != constant.RechargeOrderStatusPaid {
		return res, errors.New("当前状态不可操作")
	}

	// todo 退款后不可反结账

	if order.Member == nil {
		return res, errors.New("会员不存在")
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
				return err
			}
			memberPointsChanged = true
		}
		// 处理会员余额
		if order.RechargeAmount > 0 || order.GiftAmount > 0 {
			if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				Uuid:           order.MemberUuid,
				RechargeAmount: -order.RechargeAmount,
				GiftAmount:     -order.GiftAmount,
				Scene:          constant.MemberBalanceLogRechargeReverse,
				Describe:       fmt.Sprintf("充值反结账：%s", order.OrderNo),
			}); err != nil {
				return err
			}
		}

		paymentOrderRepo := repository.NewPaymentOrderRepo(tx)
		// 存在现金支付订单
		for _, paymentOrder := range order.PaymentOrders {
			if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				if err := s.cashBoxSrv.UpdateBalance(ctx, constant.CashBoxLogTypeOut, paymentOrder.Amount-order.ChargeDue, order.Uuid); err != nil {
					return err
				}
			}

			// 标记删除
			if err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
				"status":      constant.PaymentOrderStatusRefund,
				"delete_time": time.Now().Unix(),
			}); err != nil {
				return errors2.ErrInternal
			}
		}

		rechargeOperation := map[string]any{
			"pay_price":  order.Amount,        //  订单应付
			"change_due": order.ChargeDue,     //  找零
			"pay_type":   order.PaymentOrders, //  支付方式
		}

		operationData, _ := json.Marshal(rechargeOperation)

		if err := repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "反结账",
			Action:            constant.RechargeOrderActionReverseSettle,
			Data:              string(operationData),
			RechargeOrderUuid: order.Uuid,
		}); err != nil {
			return err
		}

		if err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, map[string]any{
			"status":       constant.RechargeOrderStatusPending,
			"payment_time": time.Now().Unix(),
			"amount":       0,
		}); err != nil {
			return errors2.ErrInternal
		}

		return nil
	})

	if err != nil {
		return res, err
	}

	if memberPointsChanged {
		go s.memberSrv.HandleMemberUpgrade(ctx.GetCompanyUuid(), order.MemberUuid)
	}

	return s.GetPendingRechargeOrder(ctx.GetCompanyUuid()), nil
}
