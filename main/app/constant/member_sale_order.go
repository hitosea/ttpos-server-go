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

// 自助点餐机订单操作类型
const (
	OrderPayFinishKioskOrder = "PAY_FINISH_KIOSK_ORDER" // 自助点餐机订单支付成功
)

// 会员端订单状态列表
const (
	MemberSaleOrderStatusSelecting             = 0 // 选购中
	MemberSaleOrderStatusPendingPayment        = 1 // 待付款
	MemberSaleOrderStatusPendingMerchantAccept = 2 // 待商家接单
	MemberSaleOrderStatusCooking               = 3 // 商家备餐中
	MemberSaleOrderStatusPendingRiderPickup    = 4 // 待骑手接单
	MemberSaleOrderStatusPendingRiderDelivery  = 5 // 骑手正在赶往商家
	MemberSaleOrderStatusDelivering            = 6 // 骑手配送中
	MemberSaleOrderStatusCompleted             = 7 // 已完成
	MemberSaleOrderStatusCancelled             = 8 // 已取消
)

// 收银机“外送”页面订单状态分类列表
const (
	CashierMemberSaleOrderStatusUnaccept   = "unaccept"   // 待接单。对应 2-待商家接单
	CashierMemberSaleOrderStatusAccept     = "accept"     // 备餐中。对应 3-商家备餐中
	CashierMemberSaleOrderStatusUndelivery = "undelivery" // 待配送。对应 4-待骑手接单、5-骑手正在赶往商家
	CashierMemberSaleOrderStatusDelivery   = "delivery"   // 配送中。对应 6-骑手配送中
	CashierMemberSaleOrderStatusDelivered  = "completed"  // 已完成。对应 7-已完成
	CashierMemberSaleOrderStatusCancel     = "cancel"     // 已取消。对应 8-已取消
)

// 收银机“订单管理”页面订单状态分类列表
const (
	CashierMemberOrderStatusAll        = ""                                     // 全部
	CashierSaleMemberOrderStatusUnpaid = "unpaid"                               // 待付款。对应 1-待付款
	CashierMemberOrderStatusUndelivery = CashierMemberSaleOrderStatusUndelivery // 待配送。对应 4-待骑手接单、5-骑手正在赶往商家
	CashierMemberOrderStatusDelivery   = CashierMemberSaleOrderStatusDelivery   // 配送中。对应 6-骑手配送中
	CashierMemberOrderStatusCompleted  = CashierMemberSaleOrderStatusDelivered  // 已完成。对应 7-已完成
	CashierMemberOrderStatusCancelled  = CashierMemberSaleOrderStatusCancel     // 已取消。对应 8-已取消
)

// 会员端“订单管理”页面订单状态分类列表
const (
	MemberOrderStatusAll        = ""                  // 全部
	MemberOrderStatusUnpaid     = "member_unpaid"     // 待付款。对应 1-待付款
	MemberOrderStatusUndelivery = "member_undelivery" // 待配送。对应 2-待商家接单、3-商家备餐中、4-待骑手接单
	MemberOrderStatusDelivery   = "member_delivery"   // 配送中。对应 5-骑手正在赶往商家、6-骑手配送中
	MemberOrderStatusCompleted  = "member_completed"  // 已完成。对应 7-已完成
	MemberOrderStatusCancelled  = "member_cancel"     // 已取消。对应 8-已取消
)

// 收银机订单状态分组
func ParseToStatusGroup(status uint) string {
	switch status {
	case MemberSaleOrderStatusPendingPayment: // 待付款
		return CashierSaleMemberOrderStatusUnpaid // 待付款
	case MemberSaleOrderStatusPendingMerchantAccept: // 待商家接单
		return CashierMemberSaleOrderStatusUnaccept // 待接单
	case MemberSaleOrderStatusCooking: // 商家备餐中
		return CashierMemberSaleOrderStatusAccept // 备餐中
	case MemberSaleOrderStatusPendingRiderPickup: // 待骑手接单
		return CashierMemberSaleOrderStatusUndelivery // 待配送
	case MemberSaleOrderStatusPendingRiderDelivery: // 骑手正在赶往商家
		return CashierMemberSaleOrderStatusUndelivery // 待配送
	case MemberSaleOrderStatusDelivering: // 骑手配送中
		return CashierMemberSaleOrderStatusDelivery // 配送中
	case MemberSaleOrderStatusCompleted: // 已完成
		return CashierMemberSaleOrderStatusDelivered // 已完成
	case MemberSaleOrderStatusCancelled: // 已取消
		return CashierMemberSaleOrderStatusCancel // 已取消
	default:
		return ""
	}
}

func GetMemberOrderStatusList(status string) []uint {
	switch status {
	case CashierSaleMemberOrderStatusUnpaid: // 待付款。对应 1-待付款
		return []uint{MemberSaleOrderStatusPendingPayment}
	case CashierMemberSaleOrderStatusUndelivery: // 待配送。对应 2-待商家接单、3-商家备餐中、4-待骑手接单、5-骑手正在赶往商家
		return []uint{MemberSaleOrderStatusPendingMerchantAccept, MemberSaleOrderStatusCooking, MemberSaleOrderStatusPendingRiderPickup}
	case CashierMemberSaleOrderStatusDelivery: // 配送中。对应 5-骑手正在赶往商家、6-骑手配送中
		return []uint{MemberSaleOrderStatusPendingRiderDelivery, MemberSaleOrderStatusDelivering}
	case CashierMemberSaleOrderStatusDelivered: // 已完成。对应 7-已完成
		return []uint{MemberSaleOrderStatusCompleted}
	case CashierMemberSaleOrderStatusCancel: // 已取消。对应 8-已取消
		return []uint{MemberSaleOrderStatusCancelled}
	default:
		return []uint{}
	}
}

// 仅收银机“订单管理”页面使用，会员端禁用这个方案
func GetStatusList(status string) []uint {
	switch status {
	case CashierSaleMemberOrderStatusUnpaid: // 待付款。对应 1-待付款
		return []uint{MemberSaleOrderStatusPendingPayment}
	case CashierMemberSaleOrderStatusUnaccept: // 待接单。对应 2-待商家接单
		return []uint{MemberSaleOrderStatusPendingMerchantAccept}
	case CashierMemberSaleOrderStatusAccept: // 备餐中。对应 3-商家备餐中
		return []uint{MemberSaleOrderStatusCooking}
	case CashierMemberSaleOrderStatusUndelivery: // 待配送。对应 4-待骑手接单、5-骑手正在赶往商家
		return []uint{MemberSaleOrderStatusPendingRiderPickup, MemberSaleOrderStatusPendingRiderDelivery}
	case CashierMemberSaleOrderStatusDelivery: // 配送中。对应 6-骑手配送中
		return []uint{MemberSaleOrderStatusDelivering}
	case CashierMemberSaleOrderStatusDelivered: // 已完成。对应 7-已完成
		return []uint{MemberSaleOrderStatusCompleted}
	case CashierMemberSaleOrderStatusCancel: // 已取消。对应 8-已取消
		return []uint{MemberSaleOrderStatusCancelled}
	case CashierMemberOrderStatusAll: // 全部
		return []uint{MemberSaleOrderStatusPendingPayment, MemberSaleOrderStatusPendingMerchantAccept, MemberSaleOrderStatusCooking, MemberSaleOrderStatusPendingRiderPickup, MemberSaleOrderStatusPendingRiderDelivery, MemberSaleOrderStatusDelivering, MemberSaleOrderStatusCompleted, MemberSaleOrderStatusCancelled}
	default:
		return []uint{}
	}
}

// 取消场景：merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单
const (
	MemberSaleOrderSceneMerchantCancel       = "merchant_cancel"        // 商家取消。备餐中取消
	MemberSaleOrderSceneMemberCancel         = "member_cancel_paid"     // 用户取消。已支付时用户取消
	MemberSaleOrderSceneMerchantReject       = "merchant_reject"        // 商家拒单
	MemberSaleOrderSceneSelectingTimeout     = "selecting_timeout"      // 选购超时
	MemberSaleOrderScenePaymentTimeout       = "payment_timeout"        // 支付超时。未支付时用户取消
	MemberSaleOrderSceneMemberCancelUnpaid   = "member_cancel_unpaid"   // 用户取消。未支付时用户取消
	MemberSaleOrderSceneRiderPickupTimeout   = "rider_pickup_timeout"   // 骑手接单超时。取消订单
	MemberSaleOrderSceneDineInPaymentTimeout = "dinein_payment_timeout" // 堂食订单支付超时
)

// 取消原因
var MemberSaleOrderSceneReason = map[string]string{
	MemberSaleOrderSceneMerchantCancel:       "商家取消",
	MemberSaleOrderSceneMemberCancel:         "用户取消",
	MemberSaleOrderSceneMerchantReject:       "商家拒单",
	MemberSaleOrderSceneSelectingTimeout:     "选购超时",
	MemberSaleOrderScenePaymentTimeout:       "支付超时",
	MemberSaleOrderSceneMemberCancelUnpaid:   "用户取消",
	MemberSaleOrderSceneRiderPickupTimeout:   "骑手超时未接单",
	MemberSaleOrderSceneDineInPaymentTimeout: "支付超时",
}

// 外送订单是否已计算距离费
const (
	DistanceNotCalculated = -1 // 未计算
	DistanceCalculated    = 1  // 已计算
)

// 会员端销售订单排序.  0-其他状态，1-骑手正在赶往商家，2-骑手配送中，降序排序
const (
	MemberSaleOrderSortDefault         = 0 // 默认排序
	MemberSaleOrderSortRiderAccepting  = 1 // 骑手正在赶往商家
	MemberSaleOrderSortRiderDelivering = 2 // 骑手配送中
)
