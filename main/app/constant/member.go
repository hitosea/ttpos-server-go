package constant

const (
	RechargeOrderStatusPending  = 0 // 充值订单待支付
	RechargeOrderStatusPaid     = 1 // 充值订单已支付
	RechargeOrderStatusCanceled = 2 // 充值订单已取消
	RechargeOrderStatusExpired  = 3 // 充值订单已过期
)

// status '支付状态, 0-未支付 1-已支付 2-已退款',

const (
	PaymentOrderStatusUnPay  = 0 // 支付订单未支付
	PaymentOrderStatusPaid   = 1 // 支付订单已支付
	PaymentOrderStatusRefund = 2 // 支付订单已退款
)

const RechargeOrderActionGenerateOrder = "GENERATE_ORDER" // 生成订单
const RechargeOrderActionChangeAmount = "CHANGE_AMOUNT"   // 变更充值金额
const RechargeOrderActionOrderCancel = "ORDER_CANCEL"     // 取消
const RechargeOrderActionRecharge = "RECHARGE"            // 充值
const RechargeOrderActionReverseSettle = "REVERSE_SETTLE" // 反结账
const RechargeOrderActionRefund = "REFUND"                // 退款

// 系统默认支付方式
const (
	PaymentMethodCodeBalance = 10 // 余额
	PaymentMethodCodeCash    = 40 // 现金
)
