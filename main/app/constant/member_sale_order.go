package constant

// 会员端订单状态列表
const (
	MemberSaleOrderStatusSelecting             = 0 // 选购中
	MemberSaleOrderStatusPendingPayment        = 1 // 待付款
	MemberSaleOrderStatusPendingMerchantAccept = 2 // 待商家接单
	MemberSaleOrderStatusCooking               = 3 // 商家备餐中
	MemberSaleOrderStatusPendingRiderPickup    = 4 // 待骑手接单
	MemberSaleOrderStatusPendingRiderDelivery  = 5 // 骑手正在赶往商家
	MemberSaleOrderStatusDeliverying           = 6 // 骑手配送中
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
	case MemberSaleOrderStatusDeliverying: // 骑手配送中
		return CashierMemberSaleOrderStatusDelivery // 配送中
	case MemberSaleOrderStatusCompleted: // 已完成
		return CashierMemberSaleOrderStatusDelivered // 已完成
	case MemberSaleOrderStatusCancelled: // 已取消
		return CashierMemberSaleOrderStatusCancel // 已取消
	default:
		return ""
	}
}

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
		return []uint{MemberSaleOrderStatusDeliverying}
	case CashierMemberSaleOrderStatusDelivered: // 已完成。对应 7-已完成
		return []uint{MemberSaleOrderStatusCompleted}
	case CashierMemberSaleOrderStatusCancel: // 已取消。对应 8-已取消
		return []uint{MemberSaleOrderStatusCancelled}
	default:
		return []uint{}
	}
}
