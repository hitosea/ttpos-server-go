package takeout

import (
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/pkg/context"
)

type ITakeoutHelpSrv interface {
	// ValidatePaymentMethod 验证支付方式是否在班次中（已废弃，始终返回nil）
	ValidatePaymentMethod(ctx context.Context, req *request.TakeoutOrderAcceptReq, action string) error
}

// ValidatePaymentMethod 验证支付方式是否在班次中（已废弃，不再限制）
func (s *takeoutSrv) ValidatePaymentMethod(ctx context.Context, req *request.TakeoutOrderAcceptReq, action string) error {
	return nil
}
