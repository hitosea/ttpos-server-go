package constant

// 会员端订单操作类型
const (
	OrderCreateMemberSaleOrder    = "CREATE_MEMBER_SALE_ORDER"         // 创建订单
	OrderPayFinishMemberSaleOrder = "PAY_FINISH_MEMBER_SALE_ORDER"     // 订单支付成功
	OrderCancelMemberSaleOrder    = "CANCEL_MEMBER_SALE_ORDER"         // 订单取消
	OrderAcceptMemberSaleOrder    = "ACCEPT_MEMBER_SALE_ORDER"         // 商家接单
	OrderPickMemberSaleOrder      = "COMPLETE_ORDER_MEMBER_SALE_ORDER" // 出餐完成，呼叫骑手
	OrderPickUpMemberSaleOrder    = "PICK_UP_MEMBER_SALE_ORDER"        // 骑手已接单，正在赶往商家
	OrderDeliveryMemberSaleOrder  = "DELIVERY_MEMBER_SALE_ORDER"       // 骑手取货完成，开始配送
	OrderFinishMemberSaleOrder    = "FINISH_MEMBER_SALE_ORDER"         // 配送完成，订单完成
)
