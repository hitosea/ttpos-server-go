package resp

import "ttpos-server-go/app/dto"

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
	FinishTime    int64   `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
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
	FinishTime    int64          `json:"finish_time"`    // 完成时间（支付时间）（时间戳）
	OrderAmount   float64        `json:"order_amount"`   // 订单总金额
	PaymentAmount float64        `json:"payment_amount"` // 支付金额
	PayTypeName   string         `json:"pay_type_name"`  // 支付类型名称
	SaleOrders    []CashierOrder `json:"sale_orders"`    // 订单列表
}

// 订单列表分页响应
type CashierOrderListPaginationResp struct {
	List []CashierBillList `json:"list"` // 订单列表
	Meta struct {
		dto.PageResponse
		UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
		CompleteNum int64 `json:"complete_num"` // 已完成数量
		CancelNum   int64 `json:"cancel_num"`   // 已取消数量
	} `json:"meta"` // Meta信息
}

type CashierOrderInfoPayTypes struct {
	Uuid              uint64  `json:"sale_bill_uuid"`      // 支付单UUID
	PaymentMethodName string  `json:"payment_method_name"` // 支付类型名称
	CurrencyUnit      string  `json:"currency_unit"`       // 货币单位
	PaymentAmount     float64 `json:"payment_amount"`      // 支付金额
	Status            uint    `json:"status"`              // 支付状态, 0-未支付 1-已支付 2-已退款
	Source            uint    `json:"source"`              // 支付状态, 来源 0-系统 1-手动 2-LianLianPay
}

type CashierOrderProduct struct {
	Uuid                  uint64             `json:"uuid"`                    // 销售订单商品ID
	LocaleName            dto.LocaleResponse `json:"locale_name"`             // 产品名称
	FlavorName            string             `json:"flavor_name"`             // 口味名称
	Num                   uint               `json:"num"`                     // 数量
	CustomPrice           float64            `json:"custom_price"`            // 自定义价格
	UnitPrice             float64            `json:"unit_price"`              // 单价
	Price                 float64            `json:"price"`                   // 最终单价
	TaxRate               uint               `json:"tax_rate"`                // 税率,单位%.下单时单税率,结账时再重新核算
	ProductOriginalAmount float64            `json:"product_original_amount"` // 原价销售额.包含加料、税费.
	Status                uint               `json:"status"`                  // 状态, 0-正常 1-退菜
	Remark                string             `json:"remark"`                  // 备注
	IsGift                bool               `json:"is_gift"`                 // 是否赠品, fasle-否 true-是
	GiftReason            string             `json:"gift_reason"`             // 赠品原因
	Attributes            string             `json:"attributes"`              // 规格属性加料
	ImageUrl              string             `json:"image_url"`               // 图片地址
}

type CashierOrderInfo struct {
	SaleOrderUuid uint64                `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint                  `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	SerialNo      string                `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string                `json:"order_no"`        // 订单编号
	Status        uint                  `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64                 `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64               `json:"order_amount"`    // 订单总金额
	PaymentAmount float64               `json:"payment_amount"`  // 支付金额
	PayTypeName   string                `json:"pay_type_name"`   // 支付类型名称
	MemberName    string                `json:"member_name"`     // 会员名称
	Products      []CashierOrderProduct `json:"products"`        // 产品列表
}

// 订单信息响应
type CashierOrderInfoResp struct {
	SaleBillUuid  uint64                     `json:"sale_bill_uuid"` // 销售账单UUID
	BillType      uint                       `json:"bill_type"`      // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool                       `json:"is_split"`       // 是否拆单	false:否 true:是
	SerialNo      string                     `json:"serial_no"`      // 桌位编号 (点餐流水号)
	OrderNo       string                     `json:"order_no"`       // 订单编号
	Status        uint                       `json:"status"`         // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	CreateTime    int64                      `json:"create_time"`    // 创建时间
	FinishTime    int64                      `json:"finish_time"`    // 完成时间（支付时间）（时间戳）
	OrderAmount   float64                    `json:"order_amount"`   // 订单总金额
	PaymentAmount float64                    `json:"payment_amount"` // 支付金额
	MemberNames   string                     `json:"member_names"`   // 会员名称
	PayTypes      []CashierOrderInfoPayTypes `json:"pay_types"`      // 支付类型
	SaleOrders    []CashierOrderInfo         `json:"sale_orders"`    // 订单列表
}
