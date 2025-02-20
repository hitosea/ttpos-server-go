package constant

const (
	CashBoxLogTypeOut = 1 // 取现
	CashBoxLogTypeIn  = 2 // 存现
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
