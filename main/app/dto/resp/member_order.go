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
	MemberSaleOrderUuid uint64               `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	CompanyName         string               `json:"company_name"`           // 公司名称
	SerialNumber        string               `json:"serial_number"`          // 订单流水号
	Status              uint                 `json:"status"`                 // 订单状态 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	Num                 float64              `json:"num"`                    // 商品数量. 所有商品数量总和，如商品A数量为2，商品B数量为3，则总数量为5
	Amount              float64              `json:"amount"`                 // 订单金额，最终应收. 订单金额=商品金额+配送费-会员折扣金额
	ProductAmount       float64              `json:"product_amount"`         // 商品金额. 所有商品金额总和，如商品A金额为2，商品B金额为3，则总金额为5
	ProductList         []MemberOrderProduct `json:"product_list"`           // 订单商品列表
	Rider               RiderInfo            `json:"rider"`                  // 骑手信息
}

type ExtraMemberCashierOrderListMeta struct {
	UnacceptNum   int64 `json:"unaccept_num"`   // 待接单数量
	AcceptNum     int64 `json:"accept_num"`     // 备餐中数量
	UndeliveryNum int64 `json:"undelivery_num"` // 待配送数量
	DeliveryNum   int64 `json:"delivery_num"`   // 配送中数量
	CompletedNum  int64 `json:"completed_num"`  // 已完成数量
	CancelNum     int64 `json:"cancel_num"`     // 已取消数量
}

type GetMemberCashierOrderListResp struct {
	Meta  dto.PageResponse                `json:"meta"`
	Extra ExtraMemberCashierOrderListMeta `json:"extra"`
	List  []MemberCashierOrder            `json:"list"` // 订单列表
}

type MemberCashierOrder struct {
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
	TotalPrice          float64            `json:"total_price"`           // 总价. 总价=单价*数量。 折后
	OriginTotalPrice    float64            `json:"origin_total_price"`    // 原价总价. 原价总价=原价单价*数量。 折前
	Image               string             `json:"image"`                 // 商品图片
}

type DeliveryResp struct {
	DeliveryDistance   float64 `json:"delivery_distance"`     // 配送距离，单位km
	DeliveryFeeAmount  float64 `json:"delivery_fee_amount"`   // 配送费
	DeliveryFeeMinFee  float64 `json:"delivery_fee_min_fee"`  // 起步配送费
	DeliveryFeeBaseFee float64 `json:"delivery_fee_base_fee"` // 基础配送费
	DeliveryFeePerKm   float64 `json:"delivery_fee_per_km"`   // 每公里配送费
}

// 会员端订单详情
type GetMemberOrderDetailResp struct {
	MemberSaleOrderUuid  uint64                   `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	OrderNo              string                   `json:"order_no"`               // 订单编号
	CompanyName          string                   `json:"company_name"`           // 公司名称
	PayTime              int64                    `json:"pay_time"`               // 支付时间
	RemainingPaymentTime int64                    `json:"remaining_payment_time"` // 剩余支付时间(单位秒)
	FinishTime           int64                    `json:"finish_time"`            // 完成时间
	CancelTime           int64                    `json:"cancel_time"`            // 取消时间
	CreateTime           int64                    `json:"create_time"`            // 创建时间
	CancelReason         string                   `json:"cancel_reason"`          // 取消原因
	Status               uint                     `json:"status"`                 // 订单状态 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	Remark               string                   `json:"remark"`                 // 订单备注
	AmountInfo           MemberOrderAmountInfo    `json:"amount_info"`            // 订单金额信息
	ProductList          MemberProductList        `json:"product_list"`           // 订单商品列表
	AddressInfo          MemberOrderDetailAddress `json:"address_info"`           // 订单地址
	DeliveryConfig       DeliveryResp             `json:"delivery_config"`        // 配送费配置
	Rider                RiderInfo                `json:"rider"`                  // 骑手信息
}

// 收银端“外送”接单页面订单详情
type GetMemberOrderCashierDetailResp struct {
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
	Name              string  `json:"name"`               // 骑手姓名
	Phone             string  `json:"phone"`              // 骑手电话
	Location          string  `json:"location"`           // 骑手位置.格式: 纬度,经度
	RemainingDistance float64 `json:"remaining_distance"` // 剩余距离
	EstimatedTime     string  `json:"estimated_time"`     // 预计送达时间
}

type MemberOrderAmountInfo struct {
	Amount            float64 `json:"amount"`              // 订单总金额. 订单总金额=实际付款金额=商品金额-会员折扣金额+运费
	MemberDiscountFee float64 `json:"member_discount_fee"` // 会员折扣金额
}

type MemberOrderDetailAddress struct {
	ContactName       string `json:"contact_name"`         // 联系人
	Phone             string `json:"phone"`                // 联系电话
	PhonePrefix       string `json:"phone_prefix"`         // 联系电话前缀. 例如：+86
	Address           string `json:"address"`              // 详细地址
	IsInDeliveryRange bool   `json:"is_in_delivery_range"` // 是否在配送范围内
}

type MemberOrderPaymentInfoResp struct {
	MemberSaleOrderUuid uint64  `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	PaymentOrderUuid    uint64  `json:"payment_order_uuid"`     // 支付单uuid (当/cashier/desk/order/payment/info接口的payment_orders中的存在相同的uuid时证明已经支付)
	PaymentMethodName   string  `json:"payment_method_name"`    // 支付方式名称
	IsWechatPay         bool    `json:"is_wechat_pay"`          // 是否是微信支付
	QrCode              string  `json:"qr_code"`                // 支付单二维码 (返回base64图片)
	LinkUrl             string  `json:"link_url"`               // 支付单链接 (返回跳转地址 window.location.href = LinkUrl; )
	Status              int     `json:"status"`                 // 支付单状态 支付状态, 0-未支付 1-已支付
	PaymentAmount       float64 `json:"payment_amount"`         // 支付金额
}

type MemberOrderPaymentStatusResp struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	Status              uint   `json:"status"`                 // 支付单状态 支付状态, 0-未支付 1-已支付 2-支付失败
}

type GetMemberOrderPaymentMethodListResp struct {
	List                 []PaymentMethodItem `json:"list"`                   // 支付方式列表
	RemainingPaymentTime int64               `json:"remaining_payment_time"` // 剩余支付时间(单位秒)
	Amount               float64             `json:"amount"`                 // 订单总金额. 订单总金额=实际付款金额=商品金额-会员折扣金额+运费
	PaymentMethodUuid    uint64              `json:"payment_method_uuid"`    // 当前订单的支付方式UUID(可用来默认选中)
}

// 名称、地址、坐标
type OrderCoordinate struct {
	Name    string `json:"name"`    // 名称
	Address string `json:"address"` // 地址
	Lat     string `json:"lat"`     // 纬度
	Lng     string `json:"lng"`     // 经度
}

type DriverInfoResp struct {
	Name          string  `json:"name"`           // 骑手姓名
	Phone         string  `json:"phone"`          // 骑手电话
	Avatar        string  `json:"avatar"`         // 骑手头像
	Rating        float64 `json:"rating"`         // 骑手评分
	Lat           string  `json:"lat"`            // 纬度
	Lng           string  `json:"lng"`            // 经度
	EstimatedTime string  `json:"estimated_time"` // 预计送达时间
}

// 订单相关坐标信息
type MemberOrderCoordinates struct {
	Merchant   OrderCoordinate `json:"merchant"`    // 商家
	Customer   OrderCoordinate `json:"customer"`    // 顾客
	DriverInfo DriverInfoResp  `json:"driver_info"` // 骑手
}

// ==================== 会员端堂食订单 ====================

// GetMemberDineInOrderListResp 会员端堂食订单列表响应
type GetMemberDineInOrderListResp struct {
	Meta dto.PageResponse    `json:"meta"`
	List []MemberDineInOrder `json:"list"` // 订单列表
}

// MemberDineInOrderStatusInfo 堂食订单状态信息
type MemberDineInOrderStatusInfo struct {
	Status     string `json:"status"`      // 订单状态：unpaid-待支付 pending-待接单 preparing-备餐中 completed-已完成 partial_refund-部分退款 full_refund-全部退款 cancelled-已取消 rejected-已拒单
	StatusText string `json:"status_text"` // 状态文字（多语言）
}

// MemberDineInOrder 会员端堂食订单列表项
type MemberDineInOrder struct {
	SaleBillUuid  uint64                      `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64                      `json:"sale_order_uuid"` // 销售订单UUID
	CompanyName   string                      `json:"company_name"`    // 商家名称
	SerialNo      string                      `json:"serial_no"`       // 取餐号
	OrderNo       string                      `json:"order_no"`        // 订单编号
	StatusInfo    MemberDineInOrderStatusInfo `json:"status_info"`     // 状态信息
	DiningMethod  uint                        `json:"dining_method"`   // 用餐方式：0-堂食 1-打包
	Num           float64                     `json:"num"`             // 商品数量
	Amount        float64                     `json:"amount"`          // 订单金额（应付金额）
	ProductAmount float64                     `json:"product_amount"`  // 商品金额
	CreateTime    int64                       `json:"create_time"`     // 下单时间
	SubmitPayTime int64                       `json:"submit_pay_time"` // 提交支付时间
	ProductList   []MemberOrderProduct        `json:"product_list"`    // 商品列表（前3个）
}

// GetMemberDineInOrderDetailResp 会员端堂食订单详情响应
type GetMemberDineInOrderDetailResp struct {
	SaleBillUuid         uint64                       `json:"sale_bill_uuid"`         // 销售账单UUID
	SaleOrderUuid        uint64                       `json:"sale_order_uuid"`        // 销售订单UUID
	CompanyName          string                       `json:"company_name"`           // 商家名称
	SerialNo             string                       `json:"serial_no"`              // 取餐号
	OrderNo              string                       `json:"order_no"`               // 订单编号
	StatusInfo           MemberDineInOrderStatusInfo  `json:"status_info"`            // 状态信息
	DiningMethod         uint                         `json:"dining_method"`          // 用餐方式：0-堂食 1-打包
	Remark               string                       `json:"remark"`                 // 订单备注
	CreateTime           int64                        `json:"create_time"`            // 下单时间
	SubmitPayTime        int64                        `json:"submit_pay_time"`        // 提交支付时间
	PayTime              int64                        `json:"pay_time"`               // 支付时间
	CancelTime           int64                        `json:"cancel_time"`            // 取消时间
	RemainingPaymentTime int64                        `json:"remaining_payment_time"` // 剩余支付时间（秒）
	RefundAmount         float64                      `json:"refund_amount"`          // 退款金额（用于显示"已退款 ¥xx"）
	IsOrderFirstPayLater bool                         `json:"is_order_first_pay_later"` // 是否先下单后付款（true时前端调submit，false时调pay）
	AmountInfo           MemberDineInOrderAmountInfo  `json:"amount_info"`            // 金额信息
	ProductList          MemberDineInOrderProductList `json:"product_list"`           // 商品列表
	PaymentMethods       PaymentMethodList            `json:"payment_methods"`        // 支付方式列表（待支付时返回）
}

// MemberDineInOrderProductList 堂食订单商品列表
type MemberDineInOrderProductList struct {
	List []MemberDineInOrderProduct `json:"list"` // 商品列表
}

// MemberDineInOrderProduct 堂食订单商品（包含退款信息）
type MemberDineInOrderProduct struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 float64            `json:"num"`                   // 数量
	Price               float64            `json:"price"`                 // 单价（折前）
	TotalPrice          float64            `json:"total_price"`           // 总价（折前）= 单价 * 数量
	Image               string             `json:"image"`                 // 商品图片
	RefundAmount        float64            `json:"refund_amount"`         // 退款金额（0表示未退款）
	ProductType         uint               `json:"product_type"`          // 商品类型 0-商品 1-套餐
	PackageProductList  PackageProductList `json:"package_product_list"`  // 套餐子商品列表（仅套餐商品有值）
}

// MemberDineInOrderAmountInfo 堂食订单金额信息
type MemberDineInOrderAmountInfo struct {
	ProductAmount     float64 `json:"product_amount"`      // 商品金额（合计）
	DiscountAmount    float64 `json:"discount_amount"`     // 优惠折扣金额（整单打折），与收银机购物车 discount_amount 一致
	ServiceFee        float64 `json:"service_fee"`         // 服务费
	TaxFee            float64 `json:"tax_fee"`             // 税费
	Amount            float64 `json:"amount"`              // 应付金额
	PaymentMethodName string  `json:"payment_method_name"` // 支付方式名称
}
