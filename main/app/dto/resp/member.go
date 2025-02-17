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

type SearchMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员Uuid
	Nickname string `json:"nickname"` // 会员昵称
	Phone    string `json:"phone"`    // 手机
}

type SearchMemberList struct {
	List []SearchMember `json:"list"`
}

// RechargeMember 充值会员信息
type RechargeMember struct {
	Uuid      uint64  `json:"uuid"`       // 会员Uuid
	Nickname  string  `json:"nickname"`   // 会员昵称
	CardName  string  `json:"card_name"`  // 会员卡名称
	LevelName string  `json:"level_name"` // 会员等级
	Balance   float64 `json:"balance"`    // 会员余额
	Points    float64 `json:"points"`     // 会员积分
}

// PendingRechargeOrder 进行中的充值订单
type PendingRechargeOrder struct {
	MemberUuid    uint64         `json:"member_uuid"`    // 会员Uuid
	Uuid          uint64         `json:"uuid"`           // 充值订单Uuid
	RechargeMoney float64        `json:"recharge_money"` // 充值金额
	GiftMoney     float64        `json:"gift_money"`     // 赠送金额
	GiftPoint     float64        `json:"gift_point"`     // 赠送积分
	PaymentOrders []PaymentOrder `json:"payment_orders"` // 充值订单支付类型列表
}

type PaymentOrder struct {
	Uuid              uint64  `json:"uuid"`                // 充值订单支付订单Uuid
	PaymentMethodUuid uint64  `json:"payment_method_uuid"` // 支付方式Uuid
	PaymentAmount     float64 `json:"payment_amount"`      // 支付订单金额
	Amount            float64 `json:"amount"`              // 支付订单总金额 = 支付订单金额 + 手续费 + 找零
}

// ConfirmRechargeOrder 确认充值订单响应
type ConfirmRechargeOrder struct {
	Amount         float64  `json:"amount"`          // 实收
	RechargeAmount float64  `json:"recharge_amount"` // 应收
	ChargeDue      float64  `json:"charge_due"`      // 找零
	PaymentMethods []string `json:"payment_methods"` // 支付方式
}
