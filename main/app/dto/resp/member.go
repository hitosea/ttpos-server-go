package resp

type MemberLevel struct {
	Uuid       uint64 `json:"uuid"`        // 等级Uuid
	Name       string `json:"name"`        // 等级名称
	Priority   int    `json:"priority"`    // 等级优先级
	CreateTime int64  `json:"create_time"` // 创建时间
}

type MemberLevelList struct {
	List []MemberLevel `json:"list"`
}

type MemberCardType struct {
	Uuid  uint64  `json:"uuid"`  // 等级Uuid
	Name  string  `json:"name"`  // 等级名称
	Price float64 `json:"price"` //
}

type MemberCardTypeList struct {
	List []MemberCardType `json:"list"`
}

type SearchMember struct {
	Uuid         uint64 `json:"uuid"`           // 会员Uuid
	Nickname     string `json:"nickname"`       // 会员昵称
	Phone        string `json:"phone"`          // 手机
	MemberCardNo string `json:"member_card_no"` // 会员卡号
}

type SearchMemberList struct {
	List []SearchMember `json:"list"`
}

type Card struct {
	Name string `json:"name"` // 会员卡名称
}
type Level struct {
	Name string `json:"name"` // 会员等级名称
}

// RechargeMember 充值会员信息
type RechargeMember struct {
	Uuid          uint64  `json:"uuid"`           // 会员Uuid
	Nickname      string  `json:"nickname"`       // 会员昵称
	Card          Card    `json:"card"`           // 会员卡
	Level         Level   `json:"level"`          // 会员等级名称
	Balance       float64 `json:"balance"`        // 会员余额
	Points        float64 `json:"points"`         // 会员积分
	Phone         string  `json:"phone"`          // 手机号
	RechargeMoney float64 `json:"recharge_money"` // 累计充值金额
}

// RechargeOrder 进行中的充值订单
type RechargeOrder struct {
	MemberUuid     uint64          `json:"member_uuid"`     // 会员Uuid
	Uuid           uint64          `json:"uuid"`            // 充值订单Uuid
	RechargeAmount float64         `json:"recharge_amount"` // 充值金额
	GiftAmount     float64         `json:"gift_amount"`     // 赠送金额
	GiftPoint      float64         `json:"gift_point"`      // 赠送积分
	ChargeDue      float64         `json:"charge_due"`      // 找零
	PaymentOrders  PaymentInfoList `json:"payment_orders"`  // 充值订单支付类型列表
}

type PaymentOrder struct {
	Uuid                 uint64  `json:"uuid"`                   // 充值订单支付订单Uuid
	PaymentMethodUuid    uint64  `json:"payment_method_uuid"`    // 支付方式Uuid
	PaymentMethodName    string  `json:"payment_method_name"`    // 支付方式名称
	PaymentMethodCode    int     `json:"payment_method_code"`    // 支付方式代号
	PaymentAmount        float64 `json:"payment_amount"`         // 支付订单金额
	PaymentCommissionFee float64 `json:"payment_commission_fee"` // 手续费
	Amount               float64 `json:"amount"`                 // 实收金额（含手续费）
	DisabledCancel       bool    `json:"disabled_cancel"`        // 禁止撤销，false-不禁止；true-禁止
	PaymentInfo          string  `json:"payment_info"`           // 支付信息.json格式,存储第三方支付返回的详细信息,kbank的信息
}

// ConfirmRechargeOrder 确认充值订单响应
type ConfirmRechargeOrder struct {
	Amount         float64  `json:"amount"`          // 应收
	ActualAmount   float64  `json:"actual_amount"`   // 实收
	ChargeDue      float64  `json:"charge_due"`      // 找零
	PaymentMethods []string `json:"payment_methods"` // 支付方式
}

// MemberDiscountResp 会员优惠响应
type MemberDiscountResp struct {
	Member          RechargeMember `json:"member"`
	HasPassword     bool           `json:"has_password"`     // 是否设置了密码
	DiscountedPrice float64        `json:"discounted_price"` // 折扣后价格
}
