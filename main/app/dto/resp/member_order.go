package resp

import "ttpos-server-go/app/dto"

/***
  收银机"外送"模块 接口返回结构体
***/

type GetMemberOrderListResp struct {
	Meta dto.PageResponse `json:"meta"`
	List []MemberOrder    `json:"list"` // 订单列表
}

type MemberOrder struct {
	MemberSaleOrderUuid uint64  `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	SerialNumber        string  `json:"serial_number"`          // 订单流水号
	Status              uint    `json:"status"`                 // 订单状态.0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	StatusGroup         string  `json:"status_group"`           // 订单状态分组. "unaccept" 待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
	Num                 float64 `json:"num"`                    // 商品数量. 所有商品数量总和，如商品A数量为2，商品B数量为3，则总数量为5
	ProductAmount       float64 `json:"product_amount"`         // 商品金额. 所有商品金额总和，如商品A金额为2，商品B金额为3，则总金额为5
}

type MemberProductList struct {
	List          []MemberOrderProduct `json:"list"`           // 订单商品列表
	ProductAmount float64              `json:"product_amount"` // 商品金额. 所有商品金额总和，如商品A金额为2，商品B金额为3，则总金额为5
}

type MemberOrderProduct struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 float64            `json:"num"`                   // 数量
	TotalPrice          float64            `json:"total_price"`           // 总价. 总价=单价*数量
}

// 收银端“外送”接单页面订单详情
type GetMemberOrderDetailResp struct {
	MemberSaleOrderUuid uint64                   `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	PayTime             int64                    `json:"pay_time"`               // 支付时间
	FinishTime          int64                    `json:"finish_time"`            // 完成时间
	CancelTime          int64                    `json:"cancel_time"`            // 取消时间
	CancelReason        string                   `json:"cancel_reason"`          // 取消原因
	Remark              string                   `json:"remark"`                 // 订单备注
	AmountInfo          MemberOrderAmountInfo    `json:"amount_info"`            // 订单金额信息
	ProductList         MemberProductList        `json:"product_list"`           // 订单商品列表
	AddressInfo         MemberOrderDetailAddress `json:"address_info"`           // 订单地址
	Rider               RiderInfo                `json:"rider"`                  // 骑手信息
}

type RiderInfo struct {
	Name  string `json:"name"`  // 骑手姓名
	Phone string `json:"phone"` // 骑手电话
}

type MemberOrderAmountInfo struct {
	Amount float64 `json:"amount"` // 订单总金额. 订单总金额=实际付款金额=商品金额-会员折扣金额+运费
}

type MemberOrderDetailAddress struct {
	ContactName string `json:"contact_name"` // 联系人
	Phone       string `json:"phone"`        // 联系电话
	PhonePrefix string `json:"phone_prefix"` // 联系电话前缀. 例如：+86
	Address     string `json:"address"`      // 详细地址
}

type MemberOrderPaymentInfoResp struct {
	MemberSaleOrderUuid uint64  `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	PaymentOrderUuid    uint64  `json:"payment_order_uuid"`     // 支付单uuid (当/cashier/desk/order/payment/info接口的payment_orders中的存在相同的uuid时证明已经支付)
	PaymentMethodName   string  `json:"payment_method_name"`    // 支付方式名称
	QrCode              string  `json:"qr_code"`                // 支付单二维码 (返回base64图片)
	LinkUrl             string  `json:"link_url"`               // 支付单链接 (返回跳转地址 window.location.href = LinkUrl; )
	Status              int     `json:"status"`                 // 支付单状态 支付状态, 0-未支付 1-已支付
	PaymentAmount       float64 `json:"payment_amount"`         // 支付金额
}

type MemberOrderPaymentStatusResp struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	Status              uint   `json:"status"`                 // 支付单状态 支付状态, 0-未支付 1-已支付 2-支付失败
}
