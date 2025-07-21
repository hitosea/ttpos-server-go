package model

import (
	"strings"
	"time"
	"ttpos-server-go/app/constant"
)

// LlPaymentOrder 支付订单 ttpos_ll_payment_order
type LlPaymentOrder struct {
	BaseModel
	RelatedUuid       uint64  `gorm:"column:related_uuid;type:bigint(20);default:0;comment:关联订单ID" json:"related_uuid"`
	PaymentOrderUuid  uint64  `gorm:"column:payment_order_uuid;type:bigint(20) unsigned;default:0;comment:自己系统的支付订单ID;NOT NULL" json:"payment_order_uuid"`
	PaymentMethodUuid uint64  `gorm:"column:payment_method_uuid;type:bigint(20);default:0;comment:支付方式ID;NOT NULL" json:"payment_method_uuid"`
	RelatedType       int     `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_type"`
	MerchantId        string  `gorm:"column:merchant_id;type:varchar(255);default:'';comment:lianlian商户号;NOT NULL" json:"merchant_id"`
	MerchantOrderId   string  `gorm:"column:merchant_order_id;type:varchar(255);default:'';comment:自己系统的为支付生成的订单号;NOT NULL" json:"merchant_order_id"`
	OrderId           string  `gorm:"column:order_id;type:varchar(255);default:'';comment:lianlian订单ID;NOT NULL" json:"order_id"`
	OrderType         string  `gorm:"column:order_type;type:varchar(50);default:'';comment:订单类型;NOT NULL" json:"order_type"`
	OrderStatus       string  `gorm:"column:order_status;type:varchar(50);default:'';comment:lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期;NOT NULL" json:"order_status"`
	OrderAmount       float64 `gorm:"column:order_amount;type:decimal(12,2);default:0.00;comment:lianlian订单金额;NOT NULL" json:"order_amount"`
	OrderCurrency     string  `gorm:"column:order_currency;type:varchar(50);default:'';comment:lianlian订单货币;NOT NULL" json:"order_currency"`
	CommissionFee     float64 `gorm:"column:commission_fee;type:decimal(12,2);default:0.00;comment:支付手续费;NOT NULL" json:"commission_fee"`
	FullName          string  `gorm:"column:full_name;type:varchar(50);default:'';comment:订单人名称;NOT NULL" json:"full_name"`
	OrderDesc         string  `gorm:"column:order_desc;type:varchar(50);default:'';comment:订单描述;NOT NULL" json:"order_desc"`
	LinkUrl           string  `gorm:"column:link_url;type:text;comment:lianlian订单支付链接" json:"link_url"`
	MerchantUserId    string  `gorm:"column:merchant_user_id;type:varchar(255);default:'';comment:自己系统的用户ID;NOT NULL" json:"merchant_user_id"`
	LlCreateTime      string  `gorm:"column:ll_create_time;type:varchar(250);default:'0';comment:lianlian订单创建时间;NOT NULL" json:"ll_create_time"`
	PayTime           int64   `gorm:"column:pay_time;type:int(11);default:0;comment:支付时间;NOT NULL" json:"pay_time"`
	ExpiredTime       int64   `gorm:"column:expired_time;type:int(11);default:0;comment:过期时间;NOT NULL" json:"expired_time"`

	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
}

// 获取有效时间
func (model *LlPaymentOrder) GetAliveTime() int64 {
	// 二维码有效期 微信(90111)-60分 支付宝(90222)-15分 promptPay(90333)-8分
	if model.OrderType == "LIANLIAN_WECHAT" {
		// h5支付
		if strings.Contains(model.LinkUrl, "https://wx.tenpay.com") {
			return 60 * 5
		}
		// 二维码支付
		return 60 * 60
	}
	if model.OrderType == "LIANLIAN_ALI_OFFLINE_PAY" {
		return 60 * 15
	}
	if model.OrderType == "LIANLIAN_QR_PROMPT_PAY" {
		return 60 * 8
	}
	return 99999999
}

// 获取剩余可支付时间
func (model *LlPaymentOrder) GetRemainingPayableTime() int64 {
	// 计算从创建时间到现在经过的秒数
	elapsedSeconds := time.Now().Unix() - model.BaseModel.CreateTime
	// 剩余可支付时间 = 总有效时间 - 已经过时间
	remainingTime := model.GetAliveTime() - elapsedSeconds
	// 如果剩余时间小于0，说明已过期，返回0
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

// 获取支付状态
func (model *LlPaymentOrder) GetStatus() int {
	if model.OrderStatus == "PS" {
		return constant.PaymentOrderStatusPaid
	}
	return constant.PaymentOrderStatusUnPay
}
