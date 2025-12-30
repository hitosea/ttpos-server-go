package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// RechargeOrderListReq 充值订单列表查询
type RechargeOrderListReq struct {
	dto.PageReq              // 分页参数
	OrderNo           string `form:"order_no"`             // 订单编号
	DateType          int    `form:"date_type,default=-1"` // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	EnableCreateTime  bool   `form:"enable_create_time"`   // 启用添加时间 false-不启用，true-启用
	EnablePaymentTime bool   `form:"enable_payment_time"`  // 启用支付时间 false-不启用，true-启用
	QueryStartTime    int64  `form:"query_start_time"`     // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time"`       // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date"`     // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date"`       // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	Status            int    `form:"status,default=-1"`    // 充值订单状态, -1=全都、 0=待付款、1=已完成、2=已取消
}

type RechargeOrderUuidReq struct {
	Uuid uint64 `json:"uuid"` // 充值订单uuid
}

type RechargeOrderRefundReq struct {
	Uuid        uint64  `json:"uuid" binding:"required"`                                   // 充值订单uuid
	RefundType  uint    `json:"refund_type" binding:"required,oneof=1 2"`                  // 退款类型: 1-整单退款, 2-部分退款
	RefundMoney float64 `json:"refund_money" binding:"omitempty,required_if=RefundType 2"` // 部分退款金额
	// QrPromptPay
	BankCode    string `json:"bank_code"`    // 银行代码
	AccountNo   string `json:"account_no"`   // 账号
	AccountName string `json:"account_name"` // 账号名称
}

type RechargeOrderReReturnReq struct {
	ReturnOrderUuid  uint64 `json:"return_order_uuid"`  // 退款订单UUID
	ReturnAmountUuid uint64 `json:"return_amount_uuid"` // 退款金额UUID
	// 退款账户信息
	BankCode    string `json:"bank_code"`    // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo   string `json:"account_no"`   // 账号 - 当存在QR PromptPay的时候需要传
	AccountName string `json:"account_name"` // 账户名称 - 当存在QR PromptPay的时候需要传
}

// RechargeOrderPaymentQrcode
type RechargeOrderPaymentQrcodeReq struct {
	RechargeOrderUuid uint64  `form:"recharge_order_uuid"` // 充值订单UUID, 必填
	PaymentMethodUuid uint64  `form:"payment_method_uuid"` // 支付方式UUID, 必填
	PaymentAmount     float64 `form:"payment_amount"`      // 支付金额, 必填
}

func (r *RechargeOrderPaymentQrcodeReq) Validate() error {
	if r.RechargeOrderUuid == 0 {
		return errors.New("RechargeOrderUuid不能为0")
	}
	if r.PaymentMethodUuid == 0 {
		return errors.New("PaymentMethodUuid不能为空")
	}
	if r.PaymentAmount <= 0 {
		return errors.New("支付金额错误")
	}
	if r.PaymentAmount > 200000 {
		return errors.New("最大支付金额为200000")
	}
	if r.PaymentAmount < 1 {
		return errors.New("最小支付金额为1")
	}
	return nil
}
