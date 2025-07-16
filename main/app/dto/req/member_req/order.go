package member_req

type SetMemberOrderAddressReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	MemberAddressUuid   uint64 `json:"member_address_uuid" binding:"required"`    // 会员地址UUID
}
