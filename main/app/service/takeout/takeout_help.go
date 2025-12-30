package takeout

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
)

type ITakeoutHelpSrv interface {
	// 验证支付方式是否在班次中
	ValidatePaymentMethod(ctx context.Context, req *request.TakeoutOrderAcceptReq) error
}

// ValidatePaymentMethod  验证支付方式是否在班次中
func (s *takeoutSrv) ValidatePaymentMethod(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
	db := ctx.GetDB()
	staff := ctx.GetStaff()
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	// 查询订单
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}
	// 检查支付方式是否已存在（通过 payment_name 或 code）
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	existPayment := repository.NewPaymentMethodRepo(db).GetPaymentMethod(paymentMethodRepo.WhereCode(func() int {
		if order.IsGrabOrder() {
			return constant.PaymentMethodCodeGrab
		}
		return constant.PaymentMethodCodeLineMan
	}()))
	if existPayment.Uuid == 0 {
		return errors.New("支付方式未开启")
	}
	// 验证支付方式是否在班次中
	isValid, err := service.NewStaffShiftSrv(cache.Global, s.dbm, nil, nil).ValidatePaymentMethod(ctx, staff.DutyNo, existPayment.Uuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if !isValid {
		return errors.New("请交班后再重新选择该支付方式")
	}
	return nil
}
