package service

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/repository/saas"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

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
		utils.Go(func() {
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
		})
	}

	newInfoResp, err := s.InstantOrderPaymentInfo(ctx, nil, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return newInfoResp, nil
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

	// 如果选择了活动，验证活动有效性
	if saleOrder.FullReductionActivityUuid > 0 {
		now := time.Now().Unix()
		activityRepo := repository.NewFullReductionActivityRepo(db)
		activity, err := activityRepo.GetByUuid(saleOrder.FullReductionActivityUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "查询活动失败")
		}
		if activity == nil {
			return nil, errors.New("活动信息已经变更，请重新确认")
		}

		// 验证活动有效性
		status := activity.GetStatus(now, "")
		if status != constant.ActivityStatusInProgress {
			return nil, errors.New("活动信息已经变更，请重新确认")
		}

		// 判断活动是否在适用时段内
		if !s.isActivityInTimeRange(ctx, activity, now) {
			return nil, errors.New("活动信息已经变更，请重新确认")
		}

		// 重新计算活动抵扣金额（确保金额正确）
		discountAmount, activityMessage, err := s.calculateActivityDiscount(ctx, saleOrder, saleOrder.FullReductionActivityUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "计算活动抵扣金额失败")
		}

		if discountAmount == 0 {
			return nil, errors.New("活动可抵扣金额为0，请重新确认")
		}

		// 更新活动信息
		saleOrder.ActivityAmount = discountAmount
		saleOrder.FullReductionActivityMessage = activityMessage
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
		ActivityAmount:       saleOrder.ActivityAmount, // 满减活动抵扣金额
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
		utils.Go(func() {
			s.memberSrv.HandleMemberUpgrade(ctx.GetCompanyUuid(), saleOrder.ConsumerUuid)
		})
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
	utils.Go(func() {
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
	})

	// 该订单的所有出库记录都标记已出库。将预出库的状态改为已出库
	repository.NewWarehouseFormRepo(db).UpdateWarehouseOutFormItemRecordsStatus(saleOrder.Uuid)

	// 发送结账短信
	if saleOrder.ConsumerUuid != 0 {
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
		if err != nil {
			ctx.Log().Info("停止发送短信，获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			utils.Go(func() {
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
			})
		}
	}

	// 发布"结账"事件
	originSaleOrderAmount := saleOrder.GetOriginAmountValue()
	saleOrderPaymentAmount := saleOrder.PaymentAmount
	saleOrderChangeAmount := saleOrder.ChangeAmount
	utils.Go(func() {

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
	})

	// 整单完结时, 发布"统计"事件
	if saleBill.CanFinishSaleBill() {
		utils.Go(func() {
			s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
				BasePayload: event.BasePayload{ // 统计
					Ctx: ctx,
				},
				SaleBillUuid: saleBill.Uuid,
			})
		})
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

	// 清空活动字段（免单时清空活动选择）
	saleOrder.SetActivityCancel()

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
	utils.Go(func() {
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
	})

	// 发布"统计"事件
	utils.Go(func() {
		s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			SaleBillUuid: saleBill.Uuid,
		})
	})

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
	utils.Go(func() {
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
	})

	return infoResp, nil
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

	paymentApp, paymentAppErr := saas.NewPaymentAppRepo(s.dbm.GetDB(constant.DefaultDB)).GetPaymentAppCompanyUuid(ctx.GetCompanyUuid())
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
				ActivityAmount:        saleOrder.ActivityAmount,
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
				ActivityAmount:        saleOrder.ActivityAmount,
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

	// 获取满减活动列表
	activityList, err := s.getFullReductionActivityList(ctx, saleOrder, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err, "查询满减活动列表失败")
	}

	infoResp := &resp.InstantOrderPaymentInfoResp{
		MemberInfo:     memberInfo,
		CouponList:     couponList,
		PaymentOrders:  resp.PaymentInfoList{List: paymentOrders},
		PaymentMethods: resp.PaymentMethodList{List: methodItems},
		Amounts:        resp.PaymentMethodAmountList{List: amounts},
		PointsExchange: pointsExchange,
		ActivityList:   activityList,
	}

	return infoResp, nil
}

// getFullReductionActivityList 获取满减活动列表
func (s *orderSrv) getFullReductionActivityList(ctx context.Context, saleOrder *model.SaleOrder, saleBill *model.SaleBill) (resp.FullReductionActivityList, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	now := time.Now().Unix()

	// 查询有效日期内的活动（进行中的活动）
	activityRepo := repository.NewFullReductionActivityRepo(db)
	activities, _, err := activityRepo.GetList(
		repository.CommonRepo.WhereBySoftDelete(),
		activityRepo.WhereStatus(constant.ActivityStatusInProgress, now),
	)
	if err != nil {
		return resp.FullReductionActivityList{}, errors.WithMessage(err, "查询活动列表失败")
	}

	// 计算订单金额（用于判断是否满足满减条件）
	orderAmount := saleOrder.GetAmountValue() // 积分抵扣后的应收金额

	// 获取当前选中的活动UUID
	selectedActivityUuid := saleOrder.FullReductionActivityUuid

	// 判断订单是否已部分支付
	hasPartialPayment := len(saleOrder.PaymentOrders) > 0

	// 判断积分抵扣后最终应收是否为0
	finalAmount := saleOrder.GetAmountValue() - saleOrder.PayPointsAmount
	isFinalAmountZero := finalAmount <= 0

	// 构建活动列表响应
	activityItems := make([]resp.FullReductionActivityItem, 0, len(activities))

	for _, activity := range activities {
		// 判断活动是否在适用时段内
		isInTimeRange := s.isActivityInTimeRange(ctx, activity, now)

		// 判断订单金额是否达到满减条件
		meetsThreshold := s.checkActivityThreshold(activity, orderAmount)

		// 判断活动是否可用
		isAvailable := isInTimeRange && meetsThreshold && !hasPartialPayment && !isFinalAmountZero

		// 判断活动是否已选中
		isSelected := activity.Uuid == selectedActivityUuid

		// 计算活动抵扣金额（如果已选中）
		discountAmount := 0.0
		if isSelected {
			var err error
			discountAmount, _, err = s.calculateActivityDiscount(ctx, saleOrder, activity.Uuid)
			if err != nil {
				// 如果计算失败，忽略错误，继续处理其他活动
				discountAmount = 0.0
			}
		}

		// 构建活动规则列表
		rules := make([]resp.ActivityRule, 0, len(activity.Rules))
		for _, rule := range activity.Rules {
			rules = append(rules, resp.ActivityRule{
				Threshold: rule.Threshold,
				Discount:  rule.ReductionAmount,
			})
		}

		// 获取多语言名称
		var localeName dto.LocaleResponse
		if activity.MultiLanguageName.Uuid > 0 {
			localeName = activity.MultiLanguageName.GetNames()
		}

		// 格式化日期
		startDate := time.Unix(activity.StartDate, 0).Format("2006-01-02")
		endDate := time.Unix(activity.EndDate, 0).Format("2006-01-02")

		activityItem := resp.FullReductionActivityItem{
			Uuid:           activity.Uuid,
			LocaleName:     localeName,
			ActivityType:   uint(activity.ReductionType + 1), // 1-阶梯满减，2-循环满减
			StartDate:      startDate,
			EndDate:        endDate,
			StartTime:      activity.StartTime,
			EndTime:        activity.EndTime,
			IsAllDay:       activity.IsAllDay == 1,
			Rules:          rules,
			IsAvailable:    isAvailable,
			IsSelected:     isSelected,
			DiscountAmount: discountAmount,
		}

		activityItems = append(activityItems, activityItem)
	}

	// 排序：可用时间范围内显示在前，不在可用时间范围内的活动在后
	sort.Slice(activityItems, func(i, j int) bool {
		// 如果可用性不同，可用的在前
		if activityItems[i].IsAvailable != activityItems[j].IsAvailable {
			return activityItems[i].IsAvailable
		}
		// 如果都可用或都不可用，按创建时间排序（这里简化处理，可以后续优化）
		return activityItems[i].Uuid < activityItems[j].Uuid
	})

	return resp.FullReductionActivityList{
		List: activityItems,
	}, nil
}

// isActivityInTimeRange 判断活动是否在适用时段内
func (s *orderSrv) isActivityInTimeRange(ctx context.Context, activity *model.FullReductionActivity, now int64) bool {
	// 如果是全天，直接返回true
	if activity.IsAllDay == 1 {
		return true
	}

	// 获取商家时区
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	timeUtil := utils.Timezone(timezone)

	// 使用商家时区获取当前时间（HH:mm格式）
	currentTimeStr := timeUtil.FormatUnixTime(now, "15:04")

	// 比较时间字符串
	return currentTimeStr >= activity.StartTime && currentTimeStr <= activity.EndTime
}

// checkActivityThreshold 判断订单金额是否达到满减条件
func (s *orderSrv) checkActivityThreshold(activity *model.FullReductionActivity, orderAmount float64) bool {
	if len(activity.Rules) == 0 {
		return false
	}

	// 找到最小的阈值
	minThreshold := activity.Rules[0].Threshold
	for _, rule := range activity.Rules {
		if rule.Threshold < minThreshold {
			minThreshold = rule.Threshold
		}
	}

	return orderAmount >= minThreshold
}

// calculateActivityDiscount 计算活动抵扣金额
func (s *orderSrv) calculateActivityDiscount(ctx context.Context, saleOrder *model.SaleOrder, activityUuid uint64) (float64, string, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 查询活动详情
	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(activityUuid)
	if err != nil {
		return 0, "", errors.WithMessage(err, "查询活动失败")
	}
	if activity == nil {
		return 0, "", errors.New("活动不存在")
	}

	// 计算订单金额（积分抵扣后的应收金额）
	orderAmount := saleOrder.GetAmountValue()

	// 根据活动类型计算抵扣金额
	var discountAmount float64
	var activityMessage string

	if activity.ReductionType == constant.FullReductionTypeStep {
		// 阶梯满减：找到满足条件的最大规则
		maxDiscount := 0.0
		maxThreshold := 0.0
		for _, rule := range activity.Rules {
			if orderAmount >= rule.Threshold && rule.Threshold > maxThreshold {
				maxThreshold = rule.Threshold
				maxDiscount = rule.ReductionAmount
			}
		}
		discountAmount = maxDiscount
		if maxDiscount > 0 {
			activityMessage = fmt.Sprintf("满%.2f减%.2f", maxThreshold, maxDiscount)
		}
	} else if activity.ReductionType == constant.FullReductionTypeCycle {
		// 循环满减：计算循环次数
		if len(activity.Rules) > 0 {
			rule := activity.Rules[0] // 循环满减通常只有一个规则
			cycles := int(orderAmount / rule.Threshold)
			discountAmount = float64(cycles) * rule.ReductionAmount
			if cycles > 0 {
				activityMessage = fmt.Sprintf("每满%.2f减%.2f", rule.Threshold, rule.ReductionAmount)
			}
		}
	}

	// 如果扣减金额大于订单金额，则最终扣减金额为订单金额
	if discountAmount > orderAmount {
		discountAmount = orderAmount
	}

	return discountAmount, activityMessage, nil
}

// OrderPaymentActivity 选择或取消满减活动
func (s *orderSrv) OrderPaymentActivity(ctx context.Context, req req.InstantOrderPaymentActivityReq) (*resp.InstantOrderPaymentInfoResp, error) {
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

	// 验证订单状态
	if err := saleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断订单是否已部分支付
	if len(saleOrder.PaymentOrders) > 0 {
		return nil, errors.New("订单已部分支付，不可选择活动")
	}

	// 判断积分抵扣后最终应收是否为0
	finalAmount := saleOrder.GetAmountValue() - saleOrder.PayPointsAmount
	if finalAmount <= 0 {
		return nil, errors.New("积分抵扣后最终应收为0，不可选择满减活动")
	}

	// 如果选择活动，验证活动有效性
	if req.FullReductionActivityUuid > 0 {
		// 查询活动详情
		activityRepo := repository.NewFullReductionActivityRepo(db)
		activity, err := activityRepo.GetByUuid(req.FullReductionActivityUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "查询活动失败")
		}
		if activity == nil {
			return nil, errors.New("活动不存在")
		}

		// 验证活动有效性
		now := time.Now().Unix()
		status := activity.GetStatus(now, "")
		if status != constant.ActivityStatusInProgress {
			return nil, errors.New("活动不在有效期内")
		}

		// 判断活动是否在适用时段内
		if !s.isActivityInTimeRange(ctx, activity, now) {
			return nil, errors.New("活动不在适用时段内")
		}

		// 判断订单金额是否达到满减条件
		orderAmount := saleOrder.GetAmountValue()
		if !s.checkActivityThreshold(activity, orderAmount) {
			return nil, errors.New("订单金额未达到满减条件")
		}

		// 验证活动与优惠券互斥：如果已使用优惠券，则不能选择活动
		if saleOrder.HasCoupon() {
			return nil, errors.New("活动与优惠券只能二选一")
		}

		// 计算活动抵扣金额
		discountAmount, activityMessage, err := s.calculateActivityDiscount(ctx, saleOrder, req.FullReductionActivityUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "计算活动抵扣金额失败")
		}

		// 更新订单的活动信息
		saleOrder.FullReductionActivityUuid = req.FullReductionActivityUuid
		saleOrder.FullReductionActivityMessage = activityMessage
		saleOrder.ActivityAmount = discountAmount

		// 选择活动后，将积分自动抵扣失效改为手动抵扣
		saleOrder.AutoPointsExchange = 0
	} else {
		// 取消活动
		saleOrder.SetActivityCancel()
	}

	// 记录使用活动前的订单金额
	oldPrice := saleOrder.GetAmountValue()

	// 使用事务更新订单
	err := db.Transaction(func(tx *gorm.DB) error {
		// 更新订单的活动相关字段
		saleOrderRepo := repository.NewSaleOrderRepo(tx)
		if err := saleOrderRepo.UpdateSaleOrderActivity(
			saleOrder.Uuid,
			saleOrder.FullReductionActivityUuid,
			saleOrder.FullReductionActivityMessage,
			saleOrder.ActivityAmount,
			saleOrder.AutoPointsExchange,
		); err != nil {
			return errors.WithMessage(err, "更新订单活动信息失败")
		}

		// 重新计算账单金额
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	})

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 重新获取订单信息（因为金额可能已更新）
	saleBill, errSaleBill = repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 获取更新后的订单金额
	saleOrder = saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("无法查询到销售订单")
	}
	newPrice := saleOrder.GetAmountValue()

	// 发布活动事件
	utils.Go(func() {
		event.NewSystemBus().PublishActivitySaleOrderEvent(event.ActivitySaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			FullReductionActivityUuid:    saleOrder.FullReductionActivityUuid,
			FullReductionActivityMessage: saleOrder.FullReductionActivityMessage,
			ActivityAmount:               saleOrder.ActivityAmount,
			OldPrice:                     oldPrice,
			NewPrice:                     newPrice,
		})
	})

	// 获取支付结账页面信息
	return s.InstantOrderPaymentInfo(ctx, saleBill, req.SaleBillUuid, req.SaleOrderUuid)
}
