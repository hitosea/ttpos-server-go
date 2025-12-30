package request

import "ttpos-server-go/app/model"

// PrintingStatementOrderReq 结账单打印请求
type PrintingStatementOrderReq struct {
	PrintType      int             `json:"print_type"`      // 打印类型 1-预结账单 2-结账单 11-外送单
	SaleBill       *model.SaleBill `json:"sale_bill"`       // 结账单
	SaleOrderUuid  uint64          `json:"sale_order_uuid"` // 订单UUID
	FirstExecution int             `json:"first_execution"` // 是否首次执行 0-否 1-是
	PayMethodUuid  uint64          `json:"pay_method_uuid"` // 支付方式UUID
	PayQrcode      string          `json:"pay_qrcode"`      // 支付二维码
}
