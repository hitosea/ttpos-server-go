package resp

import "ttpos-server-go/app/dto"

// RechargeOrderItem 订单列表响应
type RechargeOrderItem struct {
	Uuid           uint64                 `json:"uuid"`            // 充值订单Uuid
	OrderNo        string                 `json:"order_no"`        // 订单编号
	Status         int                    `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	PaymentTime    int64                  `json:"payment_time"`    // 完成时间（支付时间）（时间戳）
	RechargeAmount float64                `json:"recharge_amount"` // 充值金额
	Amount         float64                `json:"amount"`          // 实付金额
	PaymentMethods []string               `json:"payment_methods"` // 支付方式
	Extra          RechargeOrderItemExtra `json:"extra,omitempty"` // 通过当前数据控制按钮是否显示
}

// RechargeOrderItemExtra  列表额外信息响应
type RechargeOrderItemExtra struct { // 通过当前数据控制按钮是否显示
	IsCellRefund        bool `json:"is_cell_refund"`         // 是否可退款
	IsCellCancel        bool `json:"is_cell_cancel"`         // 是否可取消
	IsCellReverseSettle bool `json:"is_cell_reverse_settle"` // 是否可反结账
	IsCellPrint         bool `json:"is_cell_print"`          // 是否可打印小票
	IsCellDelete        bool `json:"is_cell_delete"`         // 是否可删除
	IsCellInvoice       bool `json:"is_cell_invoice"`        // 是否可打印发票
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
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type RechargeOrderOperationLogItem struct {
	Nickname    string `json:"nickname"`    // 会员昵称
	Phone       string `json:"phone"`       // 会员手机号
	Source      string `json:"source"`      // 来源
	CreateTime  string `json:"create_time"` // 创建时间
	Description string `json:"description"` // 描述
}

type RechargeOrderOperationLog struct {
	List []RechargeOrderOperationLogItem `json:"list"`
}
type RechargeOrderMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员Uuid
	Nickname string `json:"nickname"` // 会员昵称
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
