package constant

// 场景,10-用户充值 20-订单赠送 30-管理员操作 40-订单退款 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减 100-积分抵扣 110-积分抵扣反结账
const (
	MemberPointLogSceneRecharge              = 10  // 用户充值
	MemberPointLogSceneConsume               = 20  // 消费赠送/订单赠送
	MemberPointLogSceneAdmin                 = 30  // 管理员操作
	MemberPointLogSceneRefund                = 40  // 订单退款
	MemberPointLogSceneReverse               = 60  // 订单反结账
	MemberPointLogSceneRechargeGive          = 70  // 充值赠送
	MemberPointLogSceneRechargeReverse       = 80  // 充值反结账
	MemberPointLogSceneDeduct                = 90  // 扣减
	MemberPointLogScenePointsExchange        = 100 // 积分抵扣
	MemberPointLogScenePointsExchangeReverse = 110 // 积分抵扣反结账
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

// 变更类型(10后台管理员设置 20自动升级)

const (
	MemberLevelLogTypeAdminUser   = 10 // 后台管理员设置
	MemberLevelLogTypeAutoUpgrade = 20 // 自动升级
)

const (
	RechargeOrderStatusPending  = 0 // 充值订单待支付
	RechargeOrderStatusPaid     = 1 // 充值订单已支付
	RechargeOrderStatusCanceled = 2 // 充值订单已取消
)

const (
	RechargeOrderActionGenerateOrder = "GENERATE_ORDER" // 生成订单
	RechargeOrderActionChangeAmount  = "CHANGE_AMOUNT"  // 变更充值金额
	RechargeOrderActionOrderCancel   = "ORDER_CANCEL"   // 取消
	RechargeOrderActionRecharge      = "RECHARGE"       // 充值
	RechargeOrderActionReverseSettle = "REVERSE_SETTLE" // 反结账
	RechargeOrderActionRefund        = "REFUND"         // 退款
)

const (
	MemberGenderFemale  = 0 // 女
	MemberGenderMale    = 1 // 男
	MemberGenderUnknown = 2 // 未知
)
