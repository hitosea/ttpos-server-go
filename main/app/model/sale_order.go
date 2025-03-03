package model

import (
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
)

// SaleOrder 销售订单 `ttpos_sale_order`
type SaleOrder struct {
	BaseModel
	// 基础标识字段
	OrderNo    string `gorm:"column:order_no;comment:订单编号" json:"order_no"`
	Status     uint   `gorm:"column:status;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	IsFree     uint   `gorm:"column:is_free;comment:是否免单, 0-否 1-是" json:"is_free"`
	FreeReason string `gorm:"column:free_reason;comment:免单原因" json:"free_reason"`

	// 关联ID字段
	ConsumerUuid uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid  uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 商品金额相关字段
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:商品原始金额" json:"product_original_amount"`

	// 费用相关字段
	ServiceFee        float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费" json:"service_fee"`
	TaxFee            float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费" json:"tax_fee"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);default:0;comment:自定义折扣金额" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣金额" json:"member_discount_fee"`

	// 订单总额相关字段
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"amount"`
	CustomAmount  float64 `gorm:"column:custom_amount;type:decimal(12,2);default:0;comment:整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额" json:"custom_amount"`
	PaymentAmount float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`

	// 时间相关字段
	FinishTime int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`

	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);default:0;comment:会员折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1	" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);default:0;comment:会员卡折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);default:0;comment:自定义折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"custom_discount_rate"`

	// 抹零相关
	ZeroRule         uint8   `gorm:"column:zero_rule;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero_rule"`
	ZeroFee          float64 `gorm:"column:zero_fee;type:decimal(12,2);default:0;comment:优惠折扣抹零金额" json:"zero_fee"`
	ZeroCheckoutRule uint8   `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout_rule"`
	ZeroCheckoutFee  float64 `gorm:"column:zero_checkout_fee;type:decimal(12,2);default:0;comment:结账抹零金额" json:"zero_checkout_fee"`

	// 关联对象
	PaymentOrders                []PaymentOrder                `gorm:"foreignKey:RelatedUuid;references:uuid"`
	Member                       Member                        `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts            []*SaleOrderProduct           `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders                 []ReturnOrder                 `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetCustomerTypes []SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetDelayProducts []SaleOrderBuffetDelayProduct `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
}

func (model *SaleOrder) SetNil() {
	model.PaymentOrders = nil
	model.Member = Member{}
	model.SaleOrderProducts = nil
	model.ReturnOrders = nil
	model.SaleOrderBuffetCustomerTypes = nil
	model.SaleOrderBuffetDelayProducts = nil
}

func (model *SaleOrder) GetAmount() float64 {
	// 整单改价金额大于等于0时，返回整单改价金额
	if model.CustomAmount >= 0 {
		return model.CustomAmount
	}
	// 默认返回订单总金额
	return model.Amount
}

// 取消整单改价金额
func (model *SaleOrder) CancelCustomAmount() {
	model.CustomAmount = constant.SaleOrderCustomAmountCancel
}

// 设置整单改价金额
func (model *SaleOrder) SetCustomAmount(amount float64) {
	defer model.CancelZero() // 取消订单抹零

	model.CustomAmount = amount
}

// 取消订单抹零
func (model *SaleOrder) CancelZero() {
	// 将订单的抹零规则设置为实款实收
	model.ZeroRule = constant.DiscountZeroRuleNone
}

// 设置整单折扣，并修改订单商品的折扣
func (model *SaleOrder) SetDiscount(discount float64) {
	defer model.CancelCustomAmount() // 取消整单改价金额
	defer model.CancelZero()         // 取消订单抹零

	model.CustomDiscountRate = discount
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除或已取消，则不修改折扣
		if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() {
			continue
		}
		saleOrderProduct.CustomDiscountRate = discount
	}
}

func (model *SaleOrder) CancelDiscount() {
	model.CustomDiscountRate = 0
}

func (model *SaleOrder) GetSaleOrderProductBySign(sign string) *SaleOrderProduct {
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.Sign == sign {
			return saleOrderProduct
		}
	}
	return nil
}

func (model *SaleOrder) IsFreeSaleOrder() bool {
	return model.IsFree == constant.SaleOrderProductStatusNormal
}

type DiscountInfo struct {
	MemberDiscountRate     float64 `json:"member_discount_rate"`
	MemberCardDiscountRate float64 `json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `json:"custom_discount_rate"`
}

func (model *SaleOrder) GetDiscountInfo() DiscountInfo {
	return DiscountInfo{
		MemberDiscountRate:     model.MemberDiscountRate,
		MemberCardDiscountRate: model.MemberCardDiscountRate,
		CustomDiscountRate:     model.CustomDiscountRate,
	}
}

// 返回新的销售订单商品
func (model *SaleOrder) GetSaleOrderProduct(saleOrderProductUuid uint64) (*SaleOrderProduct, int, error) {
	for i, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProductUuid == saleOrderProduct.Uuid {
			return saleOrderProduct, i, nil
		}
	}
	return nil, 0, errors.New("销售订单商品不存在")
}

type Calc struct {
	ProductOriginalAmount float64 `json:"product_original_amount"`
	ProductAmount         float64 `json:"product_amount"`
	ServiceFee            float64 `json:"service_fee"`
	TaxFee                float64 `json:"tax_fee"`
	CustomDiscountFee     float64 `json:"custom_discount_fee"`
	MemberDiscountFee     float64 `json:"member_discount_fee"`
	Amount                float64 `json:"amount"`
	ZeroFee               float64 `json:"zero_fee"`
	//ZeroCheckoutFee       float64
	//PaymentAmount         float64
}

func (model *SaleOrder) BeforeCalc() Calc {
	return Calc{
		ProductOriginalAmount: model.ProductOriginalAmount,
		ProductAmount:         model.ProductAmount,
		ServiceFee:            model.ServiceFee,
		TaxFee:                model.TaxFee,
		CustomDiscountFee:     model.CustomDiscountFee,
		MemberDiscountFee:     model.MemberDiscountFee,
		Amount:                model.Amount,
		ZeroFee:               model.ZeroFee,
	}
}

// TableName 指定表名
func (model *SaleOrder) TableName() string {
	return "ttpos_sale_order"
}

// 获取总的退款金额
func (model *SaleOrder) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, refundOrder := range model.ReturnOrders {
		refundAmount += refundOrder.RefundAmount
	}
	return refundAmount
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleOrder) ValidateOrderStatus() error {
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	return nil
}

// 获取所有自助餐名称
func (model *SaleOrder) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, buffet := range model.SaleOrderBuffetCustomerTypes {
		buffets = append(buffets, buffet.BuffetPackageMultiLanguageName.GetNameByLang(language))
	}
	return strings.Join(buffets, "+")
}
