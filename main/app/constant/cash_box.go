package constant

// 场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值  7-结账找零 8-交班

const (
	CashBoxLogScenePay       = 1 // 销售订单支付
	CashBoxLogSceneRefund    = 2 // 退货退款、反结账
	CashBoxLogSceneCancelPay = 3 // 取消付款
	CashBoxLogSceneOut       = 4 // 中途取出
	CashBoxLogSceneIn        = 5 // 中途存入
	CashBoxLogSceneRecharge  = 6 // 会员充值
	CashBoxLogSceneChange    = 7 // 结账找零
	CashBoxLogSceneShift     = 8 // 交班
)
