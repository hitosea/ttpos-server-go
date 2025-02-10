package constant

const (
	OrderSourceInstant  = "instant"  // 点餐
	OrderSourceDesk     = "desk"     // 桌台
	OrderSourceRecharge = "recharge" // 充值
)

const (
	SaleBillTypeInstant = 1 // 点餐
	SaleBillTypeDesk    = 0 // 桌台
)

const (
	SaleBillDiningMethodTakeout = 0 // 外卖
	SaleBillDiningMethodDineIn  = 1 // 堂食
)

const (
	SaleBillStatusPending  = 0 // 待付款
	SaleBillStatusComplete = 1 // 已完成
	SaleBillStatusCanceled = 2 // 已取消
)

// 订单来源映射到订单编号类型
var OrderSourceMapToOrderNoType = map[string]string{
	OrderSourceInstant:  "1", // 点餐
	OrderSourceDesk:     "2", // 桌台
	OrderSourceRecharge: "3", // 充值
}

// 订单来源映射到销售账单类型
var OrderSourceMapToBillType = map[string]uint{
	OrderSourceInstant: SaleBillTypeInstant, // 点餐
	OrderSourceDesk:    SaleBillTypeDesk,    // 桌台
}
