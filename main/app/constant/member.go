package constant

const (
	RechargeOrderStatusPending  = 0 // 充值订单待支付
	RechargeOrderStatusPaid     = 1 // 充值订单已支付
	RechargeOrderStatusCanceled = 2 // 充值订单已取消
	RechargeOrderStatusExpired  = 3 // 充值订单已过期
)

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

	PaymentMethodCodeLianLianPayWechat = 90111
	PaymentMethodCodeAliPay            = 90222
	PaymentMethodCodeLianLianQRPrompt  = 90333
)

const (
	PaymentMethodSourceSystem      = 0
	PaymentMethodSourceDefault     = 1
	PaymentMethodSourceLianLianPay = 2
)

// 场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减
const (
	MemberPointLogSceneRecharge        = 10 // 用户充值
	MemberPointLogSceneConsume         = 20 // 消费赠送/订单赠送
	MemberPointLogSceneAdmin           = 30 // 管理员操作
	MemberPointLogSceneRefund          = 40 // 退款扣除
	MemberPointLogSceneReverse         = 60 // 订单反结账
	MemberPointLogSceneRechargeGive    = 70 // 充值赠送
	MemberPointLogSceneRechargeReverse = 80 // 充值反结账
	MemberPointLogSceneDeduct          = 90 // 扣减
)

// 场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-扣减
const (
	MemberBalanceLogRecharge        = 10 // 用户充值
	MemberBalanceLogConsume         = 20 // 用户消费
	MemberBalanceLogAdmin           = 30 // 管理员操作
	MemberBalanceLogRefund          = 40 // 订单退款
	MemberBalanceLogCash            = 50 // 余额提现
	MemberBalanceLogReverse         = 60 // 订单反结账
	MemberBalanceLogRechargeReverse = 70 // 充值反结账
	MemberBalanceLogRechargeRefund  = 80 // 充值退款
	MemberBalanceLogDeduct          = 90 // 扣减
)

const (
	CashBoxLogTypeOut = 1 // 取现
	CashBoxLogTypeIn  = 2 // 存现
)

// 变更类型(10后台管理员设置 20自动升级)

const (
	MemberLevelLogTypeAdminUser   = 10 // 后台管理员设置
	MemberLevelLogTypeAutoUpgrade = 20 // 自动升级
)

// 场景 1-支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值

const (
	CashBoxLogScenePay       = 1 // 支付
	CashBoxLogSceneRefund    = 2 // 退货退款
	CashBoxLogSceneCancelPay = 3 // 取消付款
	CashBoxLogSceneOut       = 4 // 中途取出
	CashBoxLogSceneIn        = 5 // 中途存入
	CashBoxLogSceneRecharge  = 6 // 会员充值
)
