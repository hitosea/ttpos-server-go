package resp

import "ttpos-server-go/app/dto"

// 订单商品信息
type OrderExportInfoProduct struct {
	Name       string  `json:"name"`        // 商品名称
	AttrName   string  `json:"attr_name"`   // 商品规格(属性名称)
	Num        float64 `json:"num"`         // 商品数量
	TotalPrice float64 `json:"total_price"` // 商品总价
}

// OrderExportInfo 订单列表响应
type OrderExportInfo struct {
	BillUuid      uint64                    `json:"bill_uuid"`      // 销售账单编号
	BillType      string                    `json:"bill_type"`      // 订单类型
	Products      []*OrderExportInfoProduct `json:"product"`        // 订单商品信息
	SerialNo      string                    `json:"serial_no"`      // 桌位编号 (点餐流水号)
	OrderNo       string                    `json:"order_no"`       // 订单编号
	Status        uint                      `json:"status"`         // 订单状态, 0-待付款、1-已完成、2-已取消
	StatusText    string                    `json:"status_text"`    // 订单状态文本
	FinishTime    int64                     `json:"finish_time"`    // 完成时间（支付时间）（时间戳）
	OrderAmount   float64                   `json:"order_amount"`   // 订单总金额
	ServiceFee    float64                   `json:"service_fee"`    // 服务费
	DiscountFee   float64                   `json:"discount_fee"`   // 优惠折扣
	MemberFee     float64                   `json:"member_fee"`     // 会员折扣
	PaymentAmount float64                   `json:"payment_amount"` // 实付金额
	RefundAmount  float64                   `json:"refund_amount"`  // 退款金额
	MemberIds     string                    `json:"member_ids"`     // 会员id
	MemberNames   string                    `json:"member_names"`   // 会员名称
	PayTypeName   string                    `json:"pay_type_name"`  // 支付类型名称
	DiningMethod  uint                      `json:"dining_method"`  // 用餐方式
	CashierName   string                    `json:"cashier_name"`   // 收银员名称
	CreateTime    int64                     `json:"create_time"`    // 创建时间
	OrderID       string                    `json:"order_id"`       // 订单id
}

type OrderExportMeta struct {
	dto.PageResponse
}

// OrderListPaginationResp 订单列表分页响应
type OrderExportListPaginationResp struct {
	List []OrderExportInfo `json:"list"` // 订单列表
	Meta OrderExportMeta   `json:"meta"` // Meta信息
}
