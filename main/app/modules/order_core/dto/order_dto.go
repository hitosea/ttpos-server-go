package dto

import "ttpos-server-go/app/dto"

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	OrderNo  string        `json:"order_no"`  // 订单编号
	BillType uint          `json:"bill_type"` // 账单类型
	Amount   float64       `json:"amount"`    // 订单总金额
	Orders   []CreateOrder `json:"orders"`    // 子订单列表
}

type CreateOrder struct {
	OrderNo  string               `json:"order_no"` // 子订单编号
	Amount   float64              `json:"amount"`   // 子订单金额
	Products []CreateOrderProduct `json:"products"` // 商品列表
}

type CreateOrderProduct struct {
	Name       dto.LocaleResponse `json:"name"`        // 商品名称
	FlavorName dto.LocaleResponse `json:"flavor_name"` // 商品规格(多语言)
	Num        float64            `json:"num"`         // 商品数量
	SalePrice  float64            `json:"sale_price"`  // 销售单价
	Price      float64            `json:"price"`       // 原价/标价
	TotalPrice float64            `json:"total_price"` // 商品总价
}

// CreateOrderResp 创建订单响应
type CreateOrderResp struct {
	BillUuid   uint64   `json:"bill_uuid"`   // 账单唯一标识
	OrderUuids []uint64 `json:"order_uuids"` // 销售订单唯一标识列表
}

// PayOrderReq 支付订单请求
type PayOrderReq struct {
	BillUuid          uint64  `json:"bill_uuid"`           // 账单UUID
	OrderUuid         uint64  `json:"order_uuid"`          // 订单UUID
	PaymentMethodUuid uint64  `json:"payment_method_uuid"` // 支付方式UUID
	Amount            float64 `json:"amount"`              // 支付金额
}

// CancelOrderReq 取消订单请求
type CancelOrderReq struct {
	BillUuid  uint64 `json:"bill_uuid"`  // 账单UUID
	OrderUuid uint64 `json:"order_uuid"` // 订单UUID
	Reason    string `json:"reason"`     // 取消原因
}
