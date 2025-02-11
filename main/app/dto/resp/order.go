package resp

import "ttpos-server-go/app/dto"

// 创建点餐订单响应
type CreateInstantOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

// 创建桌台订单响应
type CreateDeskOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

// 创建订单响应
type CreateOrderResp struct {
	Uuid uint64 `json:"uuid"` // 订单UUID
}

type CashierOrder struct {
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint    `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	SerialNo      string  `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string  `json:"order_no"`        // 订单编号
	Status        uint    `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    uint    `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64 `json:"order_amount"`    // 订单总金额
	PaymentAmount float64 `json:"payment_amount"`  // 支付金额
	PayTypeName   string  `json:"pay_type_name"`   // 支付类型名称
}

// 订单列表响应
type CashierBillList struct {
	SaleBillUuid  uint64         `json:"sale_bill_uuid"` // 销售账单UUID
	BillType      uint           `json:"bill_type"`      // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool           `json:"is_split"`       // 是否拆单	false:否 true:是
	SerialNo      string         `json:"serial_no"`      // 桌位编号 (点餐流水号)
	OrderNo       string         `json:"order_no"`       // 订单编号
	Status        uint           `json:"status"`         // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    uint           `json:"finish_time"`    // 完成时间（支付时间）（时间戳）
	OrderAmount   float64        `json:"order_amount"`   // 订单总金额
	PaymentAmount float64        `json:"payment_amount"` // 支付金额
	PayTypeName   string         `json:"pay_type_name"`  // 支付类型名称
	SaleOrders    []CashierOrder `json:"sale_order"`     // 订单列表
}

// 订单列表Meta信息
type CashierOrderListMeta struct {
	dto.PageResponse
	UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
	CompleteNum int64 `json:"complete_num"` // 已完成数量
	CancelNum   int64 `json:"cancel_num"`   // 已取消数量
}

// 订单列表分页响应
type CashierOrderListPaginationResp struct {
	List []CashierBillList    `json:"list"` // 订单列表
	Meta CashierOrderListMeta `json:"meta"` // Meta信息
}
