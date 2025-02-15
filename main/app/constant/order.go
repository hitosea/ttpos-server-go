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
	SaleBillDiningMethodDineIn  = 0 // 堂食
	SaleBillDiningMethodTakeout = 1 // 打包
)

const (
	SaleBillStatusPending  = 0 // 待付款
	SaleBillStatusComplete = 1 // 已完成
	SaleBillStatusCanceled = 2 // 已取消
)

// OrderSourceMapToOrderNoType 订单来源映射到订单编号类型
var OrderSourceMapToOrderNoType = map[string]string{
	OrderSourceInstant:  "1", // 点餐
	OrderSourceDesk:     "2", // 桌台
	OrderSourceRecharge: "3", // 充值
}

// OrderSourceMapToBillType 订单来源映射到销售账单类型
var OrderSourceMapToBillType = map[string]uint{
	OrderSourceInstant: SaleBillTypeInstant, // 点餐
	OrderSourceDesk:    SaleBillTypeDesk,    // 桌台
}

// 订单操作类型
const (
	OrderOpenTable           = "OPEN"                // 开台
	OrderSendKitchen         = "SEND"                // 送厨
	OrderRefundProduct       = "REFUND"              // 退菜
	OrderCancelRefundProduct = "CANCEL_REFUND"       // 取消退菜
	OrderChangeTable         = "CHANGE_TABLE"        // 转台
	OrderChangePrice         = "CHANGE_PRICE"        // 改价
	OrderUpdateMealNum       = "UPDATE_MEAL_NUM"     // 修改桌台就餐人数
	OrderStayOrder           = "STAY_ORDER"          // 挂单
	OrderPickOrder           = "PICK_ORDER"          // 取单
	OrderProductFree         = "PRODUCT_FREE"        // 赠菜
	OrderCancelProductFree   = "CANCEL_PRODUCT_FREE" // 取消赠菜
	OrderProductMove         = "PRODUCT_MOVE"        // 转菜
	OrderDiscount            = "DISCOUNT"            // 优惠折扣
	OrderCancelDiscount      = "CANCEL_DISCOUNT"     // 撤销优惠折扣
	OrderSettle              = "SETTLE"              // 结账
	OrderReverseSettle       = "REVERSE_SETTLE"      // 反结账
	OrderRefund              = "REFUND"              // 退款
	OrderOrderTaking         = "ORDER_TAKING"        // 接单
	OrderOrderReject         = "ORDER_REJECT"        // 拒单
	OrderMergeTable          = "MERGE_TABLE"         // 并台
	OrderOrderCancel         = "ORDER_CANCEL"        // 整单取消
	OrderCheckoutDiscount    = "CHECKOUT_DISCOUNT"   // 结账手动抹零
	OrderSplitOrder          = "SPLIT_ORDER"         // 拆单
	OrderCancelSplitOrder    = "CANCEL_SPLIT_ORDER"  // 撤销拆单
	OrderAddProduct          = "ADD_PRODUCT"         // 增加菜品
	OrderDeleteProduct       = "DELETE_PRODUCT"      // 删除菜品
)
