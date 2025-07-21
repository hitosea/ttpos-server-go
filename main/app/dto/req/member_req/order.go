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
	PaymentMethodUuid   uint64 `form:"payment_method_uuid"`                       // 支付方式UUID - 如果为0，则使用订单的支付方式, 否则使用指定的支付方式
}

// GetMemberOrderPayStatusReq 获取支付状态
type GetMemberOrderPayStatusReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}

type CallbackReq struct {
	JobStatusAfter  string `json:"jobStatusAfter"`  // 变化后的状态，即当前状态。 枚举请查看”skootar 订单状态“
	JobStatusBefore string `json:"jobStatusBefore"` // 变化前的状态。 枚举请查看”skootar 订单状态“
	ProviderName    string `json:"providerName"`    // 外送渠道名称。如：skootar。 枚举请查看”外送渠道方“
	ShopRefNo       string `json:"shopRefNo"`       // 订单号。即member_sale_order_uuid
	TakeoutRefNo    string `json:"takeoutRefNo"`    // 外送渠道订单号。如skootar的order_id
}

// CancelOrderReq 取消订单
type CancelOrderReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	CancelReason        string `form:"cancel_reason"`                             // 取消原因
}
