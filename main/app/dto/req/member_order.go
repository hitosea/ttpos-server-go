package req

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
)

// MemberOrderListReq 会员端订单列表
type MemberOrderListReq struct {
	dto.PageReq        // 分页参数
	Status      string `form:"status"` // 状态: "unaccept" 待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
}

func (req *MemberOrderListReq) GetStatusList() []uint {
	switch req.Status {
	case constant.CashierMemberSaleOrderStatusUnaccept: // 待接单。对应 2-待商家接单
		return []uint{constant.MemberSaleOrderStatusPendingMerchantAccept}
	case constant.CashierMemberSaleOrderStatusAccept: // 备餐中。对应 3-商家备餐中
		return []uint{constant.MemberSaleOrderStatusCooking}
	case constant.CashierMemberSaleOrderStatusUndelivery: // 待配送。对应 4-待骑手接单、5-骑手正在赶往商家
		return []uint{constant.MemberSaleOrderStatusPendingRiderPickup, constant.MemberSaleOrderStatusPendingRiderDelivery}
	case constant.CashierMemberSaleOrderStatusDelivery: // 配送中。对应 6-骑手配送中
		return []uint{constant.MemberSaleOrderStatusDeliverying}
	case constant.CashierMemberSaleOrderStatusDelivered: // 已完成。对应 7-已完成
		return []uint{constant.MemberSaleOrderStatusCompleted}
	case constant.CashierMemberSaleOrderStatusCancel: // 已取消。对应 8-已取消
		return []uint{constant.MemberSaleOrderStatusCancelled}
	default:
		return []uint{}
	}
}

// GetMemberOrderDetailReq 外送订单详情
type GetMemberOrderDetailReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid"` // 会员端销售订单UUID
}
