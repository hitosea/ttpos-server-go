package service

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"

	"github.com/duke-git/lancet/v2/cryptor"
)

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

	// 更换会员后自动取消满减活动选择
	saleOrder.SetActivityCancel()

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
		if member.Password != md5Password {
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

	// 更换会员后自动取消满减活动选择
	saleOrder.SetActivityCancel()

	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, false, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	infoResp, err := s.InstantOrderPaymentInfo(ctx, nil, request.SaleBillUuid, request.SaleOrderUuid)
	if err != nil {
		return nil, false, errors.WithMessage(err)
	}
	return infoResp, isCustomAmountAndZero, nil
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
