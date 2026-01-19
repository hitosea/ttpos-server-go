package resp

import (
	"ttpos-server-go/app/dto"
)

// RechargeOrderItem 订单列表响应
type RechargeOrderItem struct {
	Uuid           uint64                 `json:"uuid"`            // 充值订单Uuid
	OrderNo        string                 `json:"order_no"`        // 订单编号
	Status         int                    `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	PaymentTime    int64                  `json:"payment_time"`    // 完成时间（支付时间）（时间戳）
	RechargeAmount float64                `json:"recharge_amount"` // 充值金额
	Amount         float64                `json:"amount"`          // 实付金额
	PaymentMethods []string               `json:"payment_methods"` // 支付方式
	GiftAmount     float64                `json:"gift_amount"`     // 赠送金额
	GiftPoint      float64                `json:"gift_point"`      // 赠送积分
	MemberUuid     uint64                 `json:"member_uuid"`     // 会员uuid
	RefundMoney    float64                `json:"refund_money"`    // 退款金额
	Cashier        RechargeOrderCashier   `json:"cashier"`         // 收银员
	Extra          RechargeOrderItemExtra `json:"extra,omitempty"` // 通过当前数据控制按钮是否显示
}

// RechargeOrderItemExtra  列表额外信息响应
type RechargeOrderItemExtra struct { // 通过当前数据控制按钮是否显示
	IsCellRefund        bool `json:"is_cell_refund"`         // 是否可退款
	IsCellCancel        bool `json:"is_cell_cancel"`         // 是否可取消
	IsCellReverseSettle bool `json:"is_cell_reverse_settle"` // 是否可反结账：同一个收银员工、同一个班次、已支付、未退款
}

// RechargeOrderList 订单列表分页响应
type RechargeOrderList struct {
	List []RechargeOrderItem   `json:"list"` // 订单列表
	Meta RechargeOrderListMeta `json:"meta"` // Meta信息
}
type RechargeOrderListMeta struct {
	dto.PageResponse
	UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
	CompleteNum int64 `json:"complete_num"` // 已完成数量
	CancelNum   int64 `json:"cancel_num"`   // 已取消数量
}

type RechargeOrderPaymentMethod struct {
	Name       string  `json:"name"`        // 支付方式名称
	Price      float64 `json:"price"`       // 支付金额
	Code       int     `json:"code"`        // 支付方式代号
	SourceText string  `json:"source_text"` // 来源
}

type RefundPayType struct {
	Price            string `json:"price"`              // 支付金额
	Code             int    `json:"value"`              // 支付方式代号
	Name             string `json:"name"`               // 支付方式名称
	Remark           string `json:"remark"`             // 备注
	PaymentOrderUuid uint64 `json:"payment_order_id"`   // 付款单ID
	RefundMoney      string `json:"refund_money"`       // 退款金额
	RefundStatus     int    `json:"refund_status"`      // 退款状态 0-退款失败 1-退款成功
	ReturnOrderUuid  uint64 `json:"return_order_uuid"`  // 退款单ID
	ReturnAmountUuid uint64 `json:"return_amount_uuid"` // 退款金额ID
	BankCode         string `json:"bank_code"`          // 银行代码
	AccountNo        string `json:"account_no"`         // 账号
	AccountName      string `json:"account_name"`       // 账号姓名
	Unit             string `json:"unit"`               // 现金单位
}

type RechargeOrderOperationLogItem struct {
	RealName       string          `json:"real_name"`        // 操作员工真实姓名
	Username       string          `json:"username"`         // 操作员工账号
	Client         string          `json:"client"`           // 来源
	CreateTime     int64           `json:"create_time"`      // 创建时间
	Description    string          `json:"description"`      // 描述
	RefundType     uint            `json:"refund_type"`      // 退款方式：1-整单退款；2-部分退款
	RefundPayTypes []RefundPayType `json:"refund_pay_types"` // 退款支付方式
}

type RechargeOrderOperationLog struct {
	List []RechargeOrderOperationLogItem `json:"list"`
}
type RechargeOrderMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员Uuid
	Nickname string `json:"nickname"` // 会员昵称
	Phone    string `json:"phone"`    // 会员手机号
}

type RechargeOrderCashier struct {
	RealName string `json:"real_name"` // 收银员真实姓名
}

type RechargeOrderInfo struct {
	Uuid           uint64                       `json:"uuid"`            // 充值订单Uuid
	OrderNo        string                       `json:"order_no"`        // 充值订单编号
	Member         RechargeOrderMember          `json:"member"`          // 会员信息
	Status         int                          `json:"status"`          // 充值订单状态
	Cashier        RechargeOrderCashier         `json:"cashier"`         // 收银员
	RechargeAmount float64                      `json:"recharge_amount"` // 充值金额
	Amount         float64                      `json:"amount"`          // 实付金额：充值金额+手续费
	ChargeDue      float64                      `json:"charge_due"`      // 找零
	PaymentTime    int64                        `json:"payment_time"`    // 支付时间
	CreateTime     int64                        `json:"create_time"`     // 下单时间
	GiftAmount     float64                      `json:"gift_amount"`     // 赠送金额
	GiftPoint      float64                      `json:"gift_point"`      // 赠送积分
	PaymentMethods []RechargeOrderPaymentMethod `json:"payment_methods"` // 支付方式
	OperationLog   RechargeOrderOperationLog    `json:"operation_log"`   // 操作日志
	Extra          RechargeOrderItemExtra       `json:"extra"`           // 通过当前数据控制按钮是否显示
}

type RefundRechargeOrderMemberInfo struct {
	Balance     float64 `json:"balance"`      // 主账户
	GiftBalance float64 `json:"gift_balance"` // 赠送账户
	Points      float64 `json:"points"`       // 积分
}

type RefundRechargeOrderPaymentRecord struct {
	PaymentOrderUuid  uint64  `json:"-"`                 // 支付订单Uuid
	PaymentMethodUuid uint64  `json:"-"`                 // 支付方式Uuid
	PaymentMethodCode int     `json:"-"`                 // 支付方式代号
	PaymentName       string  `json:"payment_name"`      // 支付方式名称
	PaymentAmount     float64 `json:"payment_amount"`    // 支付金额
	RefundableAmount  float64 `json:"refundable_amount"` // 剩余可退款金额
	CurrencyUnit      string  `json:"currency_unit"`     // 金额单位
}

type RechargeOrderRefundInfo struct {
	Uuid               uint64                             `json:"uuid"`                 // 充值订单Uuid
	RefundableAmount   float64                            `json:"refundable_amount"`    // 可退款金额
	RechargeAmount     float64                            `json:"recharge_amount"`      // 充值金额
	GiftAmount         float64                            `json:"gift_amount"`          // 赠送金额
	GiftPoint          float64                            `json:"gift_point"`           // 赠送积分
	RechargeMemberInfo RefundRechargeOrderMemberInfo      `json:"recharge_member_info"` // 充值会员信息
	PaymentRecords     []RefundRechargeOrderPaymentRecord `json:"payment_records"`      // 支付订单信息
}

type ReverseSettleRechargeOrderMemberInfo struct {
	Uuid        uint64  `json:"uuid"`         // 会员Uuid
	Nickname    string  `json:"nickname"`     // 会员昵称
	Balance     float64 `json:"balance"`      // 主账户
	GiftBalance float64 `json:"gift_balance"` // 赠送账户
	Points      float64 `json:"points"`       // 积分
}

type RechargeOrderReverseSettleInfo struct {
	MemberInfo ReverseSettleRechargeOrderMemberInfo `json:"member_info"` // 会员信息
	Status     uint                                 `json:"status"`      // 订单状态：0-正常；1-会员主账户余额不足；2-存在待支付订单
	Message    string                               `json:"message"`     // 提示信息
}

// 支付页的二维码信息
type RechargeOrderPaymentQrcodeInfoResp struct {
	PaymentOrderUuid uint64  `json:"payment_order_uuid"` // 支付单uuid (当/cashier/recharge_order/info 接口存在相同的uuid时证明已经支付)
	QrCode           string  `json:"qr_code"`            // 支付单二维码 (永远都是返回base64图片给前端直接显示)
	QrCodeExpireSec  int64   `json:"qr_code_expire_sec"` // 支付单二维码剩余时间（秒）(少于等于0的时候 需要重新生成二维码, 再次请求当前接口就行)
	Status           int     `json:"status"`             // 支付单状态 支付状态, 0-未支付 1-已支付 (可选择轮询当前接口，获取支付状态)
	PaymentAmount    float64 `json:"payment_amount"`     // 支付金额
}
