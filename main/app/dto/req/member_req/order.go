package member_req

// SetMemberOrderAddressReq 设置会员端订单地址
type SetMemberOrderAddressReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	MemberAddressUuid   uint64 `json:"member_address_uuid" binding:"required"`    // 会员地址UUID
}

// VerifyPhoneReq 验证手机号
type VerifyPhoneReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	Phone               string `json:"phone" binding:"required"`                  // 手机号
	Code                string `json:"code" binding:"required"`                   // 验证码
	Register            bool   `json:"register"`                                  // 是否注册会员
	ReferrerPhone       string `json:"referrer_phone"`                            // 推荐人手机号
}

// PayMemberOrderReq 提交支付
type PayMemberOrderReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	PaymentMethodUuid   uint64 `json:"payment_method_uuid" binding:"required"`    // 支付方式UUID
	Remark              string `json:"remark"`                                    // 订单的备注信息。产品说在点击“提交支付”时保存订单备注
}

// PaidMemberOrderReq 支付成功
type PaidMemberOrderReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}

// GetMemberOrderPayInfoReq 获取支付信息
type GetMemberOrderPayInfoReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}
