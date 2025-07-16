package constant

const (
	MemberSaleOrderStatusSelecting             = 0 // 选购中
	MemberSaleOrderStatusPendingPayment        = 1 // 待支付
	MemberSaleOrderStatusPendingMerchantAccept = 2 // 待商家接单
	MemberSaleOrderStatusCooking               = 3 // 商家备餐中
	MemberSaleOrderStatusPendingRiderPickup    = 4 // 待骑手接单
	MemberSaleOrderStatusPendingRiderDelivery  = 5 // 骑手正在赶往商家
	MemberSaleOrderStatusDeliverying           = 6 // 骑手配送中
	MemberSaleOrderStatusCompleted             = 7 // 已完成
	MemberSaleOrderStatusCancelled             = 8 // 已取消
)
