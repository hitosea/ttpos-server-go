package req

import "ttpos-server-go/app/model"

// AddMemberReq 添加会员请求
type AddMemberReq struct {
	LevelUuid uint64 `json:"level_uuid" binding:"required"`                    // 会员等级Uuid
	Phone     string `json:"phone" binding:"required,max=20"`                  // 手机号
	Nickname  string `json:"nickname" binding:"omitempty,max=50"`              // 昵称
	Password  string `json:"password" binding:"omitempty,number,min=4,max=16"` // 密码
}

// AddMemberReqMessage 添加会员错误提示
var AddMemberReqMessage = map[string]string{
	"level_uuid.required": "会员等级不存在",
	"phone.max":           "手机号不能超过20个字符",
	"nickname.max":        "昵称不能超过50个字符",
	"password.number":     "密码必须为4-16位纯数字",
	"password.min":        "密码必须为4-16位纯数字",
	"password.max":        "密码必须为4-16位纯数字",
}

// RechargeReq 充值请求
type RechargeReq struct {
	MemberUuid        uint64  `json:"member_uuid" binding:"required"`           // 会员Uuid
	RechargeAmount    float64 `json:"recharge_amount" binding:"required,min=1"` // 充值金额
	RechargeOrderUuid uint64  `json:"recharge_order_uuid"`                      // 进行中的充值订单，如果没有进行中的充值订单，传递0
	GiftAmount        float64 `json:"gift_amount" binding:"omitempty,min=0"`    // 赠送金额
	GiftPoint         float64 `json:"gift_point" binding:"omitempty,min=0"`     // 赠送积分
}

var CreateRechargeOrderReqMessage = map[string]string{
	"recharge_amount.min": "最小充值金额为1",
	"gift_amount.min":     "赠送金额错误",
	"gift_point.min":      "赠送积分错误",
}

// RechargeOrderAddPaymentMethodReq 充值订单添加支付方式
type RechargeOrderAddPaymentMethodReq struct {
	RechargeOrderUuid uint64  `json:"recharge_order_uuid" binding:"required"`     // 充值订单Uuid
	PaymentAmount     float64 `json:"payment_amount" binding:"required,min=0.01"` // 支付金额，使用现金支付，可能大于充值金额，比如充值19，但是会员给了20
	PaymentMethodUuid uint64  `json:"payment_method_uuid" binding:"required"`     // 支付方式Uuid
	PaymentOrderUuid  uint64  `json:"-"`                                          // 在线充值订单支付订单Uuid，在线支付时需要

	CompanySetting model.CompanySetting `json:"-"`
}

var RechargeOrderAddPaymentMethodReqMessage = map[string]string{
	"payment_amount.min": "支付金额最低0.01",
}

// RechargeOrderCancelPaymentMethodReq 充值订单撤销支付方式
type RechargeOrderCancelPaymentMethodReq struct {
	RechargeOrderUuid uint64 `json:"recharge_order_uuid" binding:"required"` // 充值订单Uuid
	PaymentOrderUuid  uint64 `json:"payment_order_uuid" binding:"required"`  // 支付订单Uuid
}

// ConfirmRechargeOrder 确认充值订单
type ConfirmRechargeOrder struct {
	RechargeOrderUuid uint64 `json:"recharge_order_uuid"` // 充值订单Uuid
	MemberUuid        uint64 `json:"member_uuid"`         // 会员Uuid
}

// GetMemberDiscountReq 确认充值订单
type GetMemberDiscountReq struct {
	SaleOrderUuid uint64 `form:"sale_order_uuid"` // 销售订单 Uuid
	SaleBillUuid  uint64 `form:"sale_bill_uuid"`  // 销售账单 Uuid
	MemberUuid    uint64 `form:"member_uuid"`     // 会员 Uuid
}

// CheckMemberPasswordReq 使用会员优惠验证密码
type CheckMemberPasswordReq struct {
	Password   string `json:"password" binding:"omitempty,number,min=4,max=16"` // 会员密码
	MemberUuid uint64 `json:"member_uuid" binding:"required"`                   // 会员 Uuid
}

// CheckMemberPasswordMessage 使用会员优惠验证密码错误提示
var CheckMemberPasswordMessage = map[string]string{
	"password.number": "密码必须为4-16位纯数字",
	"password.min":    "密码必须为4-16位纯数字",
	"password.max":    "密码必须为4-16位纯数字",
}

// PrintRechargeOrderReq 打印充值订单
type PrintRechargeOrderReq struct {
	RechargeOrderUuid uint64 `json:"recharge_order_uuid" binding:"required"` // 充值订单Uuid
	PrintLang         string `json:"print_lang"`                             // 打印语言
}
