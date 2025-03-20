package model

// LlPaymentOrder 支付订单 ttpos_ll_payment_order
type LlPaymentOrder struct {
	BaseModel
	SaleBillUuid     uint64  `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid    uint64  `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`
	PaymentOrderUuid uint64  `gorm:"column:payment_order_uuid;type:bigint(20) unsigned;default:0;comment:自己系统的支付订单ID;NOT NULL" json:"payment_order_uuid"`
	RelatedType      int     `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_type"`
	MerchantId       string  `gorm:"column:merchant_id;type:varchar(255);default:'';comment:lianlian商户号;NOT NULL" json:"merchant_id"`
	MerchantOrderId  string  `gorm:"column:merchant_order_id;type:varchar(255);default:'';comment:自己系统的为支付生成的订单号;NOT NULL" json:"merchant_order_id"`
	OrderId          string  `gorm:"column:order_id;type:varchar(255);default:'';comment:lianlian订单ID;NOT NULL" json:"order_id"`
	OrderType        string  `gorm:"column:order_type;type:varchar(50);default:'';comment:订单类型;NOT NULL" json:"order_type"`
	OrderStatus      string  `gorm:"column:order_status;type:varchar(50);default:'';comment:lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期;NOT NULL" json:"order_status"`
	OrderAmount      float64 `gorm:"column:order_amount;type:decimal(12,2);default:0.00;comment:lianlian订单金额;NOT NULL" json:"order_amount"`
	OrderCurrency    string  `gorm:"column:order_currency;type:varchar(50);default:'';comment:lianlian订单货币;NOT NULL" json:"order_currency"`
	FullName         string  `gorm:"column:full_name;type:varchar(50);default:'';comment:订单人名称;NOT NULL" json:"full_name"`
	OrderDesc        string  `gorm:"column:order_desc;type:varchar(50);default:'';comment:订单描述;NOT NULL" json:"order_desc"`
	LinkUrl          string  `gorm:"column:link_url;type:varchar(2000);default:'';comment:lianlian订单支付链接;NOT NULL" json:"link_url"`
	MerchantUserId   string  `gorm:"column:merchant_user_id;type:varchar(255);default:'';comment:自己系统的用户ID;NOT NULL" json:"merchant_user_id"`
	LlCreateTime     string  `gorm:"column:ll_create_time;type:varchar(250);default:'0';comment:lianlian订单创建时间;NOT NULL" json:"ll_create_time"`
	PayTime          int     `gorm:"column:pay_time;type:int(11);default:0;comment:支付时间;NOT NULL" json:"pay_time"`
}

// 获取过期时间
func (model *LlPaymentOrder) GetExpireTime() int {
	// 二维码有效期 微信(90111)-60分 支付宝(90222)-15分 promptPay(90333)-8分
	// $alive_time = [
	//     '90111' =>  60 * 60,
	//     '90222' =>  60 * 15,
	//     '90333' =>  60 * 8,
	// ];
	if model.OrderType == "LIANLIAN_WECHAT" {
		return 60 * 60
	}
	if model.OrderType == "LIANLIAN_ALIPAY" {
		return 60 * 15
	}
	if model.OrderType == "LIANLIAN_PROMPTPAY" {
		return 60 * 8
	}
	return 0
}
