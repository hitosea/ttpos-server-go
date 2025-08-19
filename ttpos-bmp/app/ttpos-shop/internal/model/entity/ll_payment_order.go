// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// LlPaymentOrder is the golang structure for table ll_payment_order.
type LlPaymentOrder struct {
	Id                uint    `json:"id"                orm:"id"                  description:"自增ID"`                                                           // 自增ID
	Uuid              uint64  `json:"uuid"              orm:"uuid"                description:"UUID"`                                                           // UUID
	PaymentOrderUuid  uint64  `json:"paymentOrderUuid"  orm:"payment_order_uuid"  description:"自己系统的支付订单ID"`                                                    // 自己系统的支付订单ID
	PaymentMethodUuid uint64  `json:"paymentMethodUuid" orm:"payment_method_uuid" description:"支付方式ID"`                                                         // 支付方式ID
	RelatedType       int     `json:"relatedType"       orm:"related_type"        description:"关联订单类型：0-销售订单；1-充值订单"`                                           // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid       uint64  `json:"relatedUuid"       orm:"related_uuid"        description:"关联的充值订单、销售订单ID"`                                                 // 关联的充值订单、销售订单ID
	MerchantId        string  `json:"merchantId"        orm:"merchant_id"         description:"lianlian商户号"`                                                    // lianlian商户号
	MerchantOrderId   string  `json:"merchantOrderId"   orm:"merchant_order_id"   description:"自己系统的为支付生成的订单号"`                                                 // 自己系统的为支付生成的订单号
	OrderId           string  `json:"orderId"           orm:"order_id"            description:"lianlian订单ID"`                                                   // lianlian订单ID
	OrderType         string  `json:"orderType"         orm:"order_type"          description:"订单类型"`                                                           // 订单类型
	OrderStatus       string  `json:"orderStatus"       orm:"order_status"        description:"lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期"` // lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期
	OrderAmount       float64 `json:"orderAmount"       orm:"order_amount"        description:"lianlian订单金额"`                                                   // lianlian订单金额
	OrderCurrency     string  `json:"orderCurrency"     orm:"order_currency"      description:"lianlian订单货币"`                                                   // lianlian订单货币
	CommissionFee     float64 `json:"commissionFee"     orm:"commission_fee"      description:"支付手续费,支付金额*支付手续费百分比"`                                            // 支付手续费,支付金额*支付手续费百分比
	FullName          string  `json:"fullName"          orm:"full_name"           description:"订单人名称"`                                                          // 订单人名称
	OrderDesc         string  `json:"orderDesc"         orm:"order_desc"          description:"订单描述"`                                                           // 订单描述
	LinkUrl           string  `json:"linkUrl"           orm:"link_url"            description:"lianlian订单支付链接"`                                                 // lianlian订单支付链接
	MerchantUserId    string  `json:"merchantUserId"    orm:"merchant_user_id"    description:"自己系统的用户ID"`                                                      // 自己系统的用户ID
	LlCreateTime      string  `json:"llCreateTime"      orm:"ll_create_time"      description:"lianlian订单创建时间"`                                                 // lianlian订单创建时间
	ExpiredTime       int     `json:"expiredTime"       orm:"expired_time"        description:"过期时间"`                                                           // 过期时间
	PayTime           int     `json:"payTime"           orm:"pay_time"            description:"支付时间"`                                                           // 支付时间
	CreateTime        int     `json:"createTime"        orm:"create_time"         description:"创建时间"`                                                           // 创建时间
	UpdateTime        int     `json:"updateTime"        orm:"update_time"         description:"更新时间"`                                                           // 更新时间
	DeleteTime        uint    `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"`                                                      // 删除时间(时间戳)
}
