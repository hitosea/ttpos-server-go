package sms

// MemberConsumptionRequest 会员消费请求参数
type MemberConsumptionRequest struct {
	Company        string  `json:"company"`         // 公司名称
	Consumption    float64 `json:"consumption"`     // 消费金额
	MemberPay      float64 `json:"member_pay"`      // 会员支付的金额
	IncreasePoints float64 `json:"increase_points"` // 增加的积分
	Balance        float64 `json:"balance"`         // 消费后，会员余额
	PointsBalance  float64 `json:"points_balance"`  // 消费后，积分余额
}

// MemberRechargeRequest 会员充值请求参数
type MemberRechargeRequest struct {
	Company       string  `json:"company"`        // 公司名称
	Recharge      float64 `json:"recharge"`       // 充值金额
	BonusMoney    float64 `json:"bonus_money"`    // 赠送金额
	BonusPoints   float64 `json:"bonus_points"`   // 赠送积分
	Balance       float64 `json:"balance"`        // 充值后，会员余额
	PointsBalance float64 `json:"points_balance"` // 充值后，积分余额
}

// MemberRechargeRefundRequest 会员充值退款请求参数
type MemberRechargeRefundRequest struct {
	Company        string  `json:"company"`
	RechargeRefund float64 `json:"recharge_refund"`
	Balance        float64 `json:"balance"`
	PointsBalance  float64 `json:"points_balance"`
}

// MemberOrderRefundRequest 会员用餐订单退款请求参数
type MemberOrderRefundRequest struct {
	Company       string  `json:"company"`
	OrderRefund   float64 `json:"order_refund"`
	Balance       float64 `json:"balance"`
	PointsBalance float64 `json:"points_balance"`
}
