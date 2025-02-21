package resp

import "ttpos-server-go/app/dto"

// 创建订单响应
type CreateOrderResp struct {
	Uuid uint64 `json:"uuid"` // 订单UUID
}

// 账单列表订单响应
type BillListsOrder struct {
	SaleBillUuid  uint64         `json:"sale_bill_uuid"`  // 销售订单UUID
	SaleOrderUuid uint64         `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint           `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	SerialNo      string         `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string         `json:"order_no"`        // 订单编号
	Status        uint           `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	ConsumerUuids string         `json:"consumer_uuids"`  // 会员id
	FinishTime    int64          `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64        `json:"order_amount"`    // 订单总金额
	PaymentAmount float64        `json:"payment_amount"`  // 支付金额
	PayTypeName   string         `json:"pay_type_name"`   // 支付类型名称
	Extra         BillListsExtra `json:"extra,omitempty"`
}

// 账单列表额外信息响应
type BillListsExtra struct { // 通过当前数据控制按钮是否显示
	IsCellRefund        bool `json:"is_cell_refund"`         // 是否可退款
	IsCellCancel        bool `json:"is_cell_cancel"`         // 是否可取消
	IsCellReverseSettle bool `json:"is_cell_reverse_settle"` // 是否可反结账
	IsCellPrint         bool `json:"is_cell_print"`          // 是否可打印小票
	IsCellDelete        bool `json:"is_cell_delete"`         // 是否可删除
	IsCellInvoice       bool `json:"is_cell_invoice"`        // 是否可打印发票
}

// 订单列表响应
type BillLists struct {
	SaleBillUuid  uint64           `json:"sale_bill_uuid"`  // 销售账单UUID
	BillType      uint             `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool             `json:"is_split"`        // 是否拆单	false:否 true:是
	SerialNo      string           `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string           `json:"order_no"`        // 订单编号
	Status        uint             `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64            `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64          `json:"order_amount"`    // 订单总金额
	PaymentAmount float64          `json:"payment_amount"`  // 支付金额
	PayTypeName   string           `json:"pay_type_name"`   // 支付类型名称
	ConsumerUuids string           `json:"consumer_uuids"`  // 会员id
	SaleOrders    []BillListsOrder `json:"sale_orders"`     // 订单列表
	Extra         BillListsExtra   `json:"extra,omitempty"` // 通过当前数据控制按钮是否显示
}

// 订单列表分页响应
type OrderListPaginationResp struct {
	List []BillLists `json:"list"` // 订单列表
	Meta struct {
		dto.PageResponse
		UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
		CompleteNum int64 `json:"complete_num"` // 已完成数量
		CancelNum   int64 `json:"cancel_num"`   // 已取消数量
	} `json:"meta"` // Meta信息
}

type OrderInfoPayTypes struct {
	Uuid            uint64  `json:"sale_bill_uuid"`    // 支付单UUID
	PaymentTypeName string  `json:"payment_type_name"` // 支付类型名称
	CurrencyUnit    string  `json:"currency_unit"`     // 货币单位
	PaymentAmount   float64 `json:"payment_amount"`    // 支付金额
	Code            string  `json:"code"`              // 支付Code
	Status          uint    `json:"status"`            // 支付状态, 0-未支付 1-已支付 2-已退款
	Source          uint    `json:"source"`            // 支付状态, 来源 0-系统 1-手动 2-LianLianPay
	SourceText      string  `json:"source_text"`       // 支付状态, 来源 0-系统 1-手动 2-LianLianPay
}

type OrderProduct struct {
	Uuid           uint64             `json:"uuid"`             // 销售订单商品ID
	LocaleName     dto.LocaleResponse `json:"locale_name"`      // 产品名称
	FlavorName     string             `json:"flavor_name"`      // 口味名称
	Num            uint               `json:"num"`              // 数量
	SalePrice      float64            `json:"sale_price"`       // 销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价
	Price          float64            `json:"price"`            // 最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率'
	TotalPrice     float64            `json:"total_price"`      // 应收金额(单商品)=最终单价+服务费+总税费
	TotalSalePrice float64            `json:"total_sale_price"` // 销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价'
	TaxRate        float64            `json:"tax_rate"`         // 税率,单位%.下单时单税率,结账时再重新核算
	RefundAmount   float64            `json:"refund_amount"`    // 退款金额
	Status         uint               `json:"status"`           // 状态, 0-正常 1-退菜
	Remark         string             `json:"remark"`           // 备注
	IsGift         bool               `json:"is_gift"`          // 是否赠品, fasle-否 true-是
	GiftReason     string             `json:"gift_reason"`      // 赠品原因
	Attributes     string             `json:"attributes"`       // 规格属性加料
	ImageUrl       string             `json:"image_url"`        // 图片地址
	RefundReason   string             `json:"refund_reason"`    // 退菜原因
}

// 订单信息响应
type OrderInfo struct {
	SaleOrderUuid uint64         `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint           `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	DiningMethod  uint           `json:"dining_method"`   // 用餐方式,0-堂食 1-打包
	SerialNo      string         `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string         `json:"order_no"`        // 订单编号
	Status        uint           `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64          `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64        `json:"order_amount"`    // 订单总金额
	PaymentAmount float64        `json:"payment_amount"`  // 支付金额
	RefundAmount  float64        `json:"refund_amount"`   // 退款金额
	PayTypeName   string         `json:"pay_type_name"`   // 支付类型名称
	MemberName    string         `json:"member_name"`     // 会员名称
	MemberUuid    uint64         `json:"member_uuid"`     // 会员名称
	IsFree        bool           `json:"is_free"`         // 是否免单
	FreeReason    string         `json:"free_reason"`     // 免单原因
	Products      []OrderProduct `json:"products"`        // 产品列表
}

// 订单信息响应
type OrderInfos struct {
	SaleBillUuid  uint64              `json:"sale_bill_uuid"` // 销售账单UUID
	BillType      uint                `json:"bill_type"`      // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool                `json:"is_split"`       // 是否拆单	false:否 true:是
	IsBuffet      bool                `json:"is_buffet"`      // 是否自助餐	false:否 true:是
	SerialNo      string              `json:"serial_no"`      // 桌位编号 (点餐流水号)
	OrderNo       string              `json:"order_no"`       // 订单编号
	Status        uint                `json:"status"`         // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	CreateTime    int64               `json:"create_time"`    // 创建时间
	FinishTime    int64               `json:"finish_time"`    // 完成时间（支付时间）（时间戳）
	OrderAmount   float64             `json:"order_amount"`   // 订单总金额
	PaymentAmount float64             `json:"payment_amount"` // 支付金额
	RefundAmount  float64             `json:"refund_amount"`  // 退款金额
	MemberNames   string              `json:"member_names"`   // 会员名称
	MemberUuids   string              `json:"member_uuids"`   // 会员名称
	BuffetNames   string              `json:"buffet_names"`   // 自助餐名称
	CancelReason  string              `json:"cancel_reason"`  // 取消原因
	CashierName   string              `json:"cashier_name"`   // 收银员名称
	Remark        string              `json:"remark"`         // 备注
	PayTypes      []OrderInfoPayTypes `json:"pay_types"`      // 支付类型
	SaleOrders    []OrderInfo         `json:"sale_orders"`    // 订单列表
}

type OrderOperationLog struct {
	Uuid          uint64 `json:"uuid"`            // 账单操作记录ID
	Source        string `json:"source"`          // 操作来源 cashier-收银 assistant-助手 shop-商家后台
	Action        string `json:"action"`          // 操作行为
	Data          any    `json:"data"`            // 消息数据
	Remark        string `json:"remark"`          // 备注
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单ID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单ID
	CreateTime    int64  `json:"create_time"`     // 创建时间(时间戳)
}

type OrderInfosResp struct {
	Detail       OrderInfos `json:"detail"` // 明细
	OperationLog struct {
		List []OrderOperationLog `json:"list"`
	} `json:"operation_log"` // 操作日志
	Extra BillListsExtra `json:"extra"` // 通过当前数据控制按钮是否显示
}
