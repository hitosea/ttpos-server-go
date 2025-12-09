// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Order is the golang structure for table order.
type Order struct {
	Id                int64       `json:"id"                orm:"id"                  description:"主键"`                                                              // 主键
	Uuid              string      `json:"uuid"              orm:"uuid"                description:"系统唯一ID"`                                                          // 系统唯一ID
	MerchantId        string      `json:"merchantId"        orm:"merchant_id"         description:"商户ID (Partner Merchant ID)"`                                      // 商户ID (Partner Merchant ID)
	PartnerOrderId    string      `json:"partnerOrderId"    orm:"partner_order_id"    description:"平台订单号 (Grab Order ID)"`                                           // 平台订单号 (Grab Order ID)
	ShopRefNo         string      `json:"shopRefNo"         orm:"shop_ref_no"         description:"TTPOS内部订单号"`                                                      // TTPOS内部订单号
	ShortOrderNumber  string      `json:"shortOrderNumber"  orm:"short_order_number"  description:"短单号"`                                                             // 短单号
	ProviderName      string      `json:"providerName"      orm:"provider_name"       description:"渠道: grab, foodpanda"`                                             // 渠道: grab, foodpanda
	OrderType         string      `json:"orderType"         orm:"order_type"          description:"订单类型: DeliveryByGrab, DeliveryByRestaurant, TakeAway, DineIn"`    // 订单类型: DeliveryByGrab, DeliveryByRestaurant, TakeAway, DineIn
	OrderTime         *gtime.Time `json:"orderTime"         orm:"order_time"          description:"下单时间"`                                                            // 下单时间
	OrderStatus       string      `json:"orderStatus"       orm:"order_status"        description:"订单状态: PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED"` // 订单状态: PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED
	ScheduledTime     *gtime.Time `json:"scheduledTime"     orm:"scheduled_time"      description:"预约送达时间"`                                                          // 预约送达时间
	Currency          string      `json:"currency"          orm:"currency"            description:"货币代码"`                                                            // 货币代码
	Subtotal          float64     `json:"subtotal"          orm:"subtotal"            description:"商品小计"`                                                            // 商品小计
	TotalAmount       float64     `json:"totalAmount"       orm:"total_amount"        description:"总金额"`                                                             // 总金额
	MerchantCharge    float64     `json:"merchantCharge"    orm:"merchant_charge"     description:"商户收取金额"`                                                          // 商户收取金额
	TaxAmount         float64     `json:"taxAmount"         orm:"tax_amount"          description:"税费"`                                                              // 税费
	DiscountAmount    float64     `json:"discountAmount"    orm:"discount_amount"     description:"折扣金额"`                                                            // 折扣金额
	MerchantFundPromo float64     `json:"merchantFundPromo" orm:"merchant_fund_promo" description:"商户承担促销金额"`                                                        // 商户承担促销金额
	PaymentType       string      `json:"paymentType"       orm:"payment_type"        description:"支付方式: CASH, CASHLESS"`                                            // 支付方式: CASH, CASHLESS
	IsMexEditOrder    int         `json:"isMexEditOrder"    orm:"is_mex_edit_order"   description:"是否为商户编辑订单"`                                                       // 是否为商户编辑订单
	Cutlery           int         `json:"cutlery"           orm:"cutlery"             description:"是否需要餐具"`                                                          // 是否需要餐具
	EaterCount        int         `json:"eaterCount"        orm:"eater_count"         description:"用餐人数"`                                                            // 用餐人数
	CustomerName      string      `json:"customerName"      orm:"customer_name"       description:"顾客姓名"`                                                            // 顾客姓名
	CustomerPhone     string      `json:"customerPhone"     orm:"customer_phone"      description:"顾客电话 (虚拟号)"`                                                      // 顾客电话 (虚拟号)
	DeliveryAddress   string      `json:"deliveryAddress"   orm:"delivery_address"    description:"配送地址 (JSON)"`                                                     // 配送地址 (JSON)
	DriverEta         int         `json:"driverEta"         orm:"driver_eta"          description:"骑手预计到达时间(分钟)"`                                                    // 骑手预计到达时间(分钟)
	ReadyTime         *gtime.Time `json:"readyTime"         orm:"ready_time"          description:"商家预估准备完成时间"`                                                      // 商家预估准备完成时间
	Note              string      `json:"note"              orm:"note"                description:"订单备注"`                                                            // 订单备注
	RawData           string      `json:"rawData"           orm:"raw_data"            description:"原始JSON数据"`                                                        // 原始JSON数据
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:"创建时间"`                                                            // 创建时间
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:"更新时间"`                                                            // 更新时间
	DeletedAt         *gtime.Time `json:"deletedAt"         orm:"deleted_at"          description:"软删除"`                                                             // 软删除
}
