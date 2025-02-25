package model

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// SaleBill 销售账单 `ttpos_sale_bill`
type SaleBill struct {
	BaseModel
	// 主键和标识字段
	OrderNo  string `gorm:"column:order_no;type:varchar(255);default:'';comment:销售账单编号" json:"order_no"`
	DutyNo   string `gorm:"column:duty_no;type:varchar(255);default:'';comment:当班编号,用于标记该账单属于哪个当班" json:"duty_no"`
	SerialNo string `gorm:"column:serial_no;type:varchar(255);default:'';comment:桌位编号 (点餐流水号)" json:"serial_no"`

	// 状态相关字段
	Status uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-待付款、1-已完成、2-已取消" json:"status"`
	IsLock uint `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`

	// 订单类型字段
	BillType       uint `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-桌台订单、1-点餐订单" json:"bill_type"`
	DiningMethod   uint `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	IsBuffet       uint `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	BuffetDuration uint `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒）" json:"buffet_duration"`

	// 订单基本信息
	MealNum uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Remark  string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	Reason  string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 金额字段 - 主要金额
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
	ProductAmount float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额,关联销售订单的商品金额之和" json:"product_amount"`

	// 金额字段 - 支付相关
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,多次支付的支付手续费之和" json:"payment_commission_fee"`

	// 金额字段 - 费用相关
	ServiceFee float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费,关联销售订单的服务费之和" json:"service_fee"`
	TaxFee     float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费,关联销售订单的税费之和" json:"tax_fee"`

	// 金额字段 - 优惠相关
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用,关联销售订单的折扣费用之和" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用,关联销售订单的会员折扣费用之和" json:"member_discount_fee"`
	GiftAmount        float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,关联销售订单的赠菜金额之和" json:"gift_amount"`
	FreeAmount        float64 `gorm:"column:free_amount;type:decimal(12,2);default:0;comment:免单金额,关联销售订单的免单金额之和" json:"free_amount"`

	// 时间相关字段
	FinishTime   int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	HideBillTime int64 `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`

	// 关联ID字段
	ConsumerUuid       uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid        uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	DeskUuid           uint64 `gorm:"column:desk_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"desk_uuid"`
	BuffetPackage1Uuid uint64 `gorm:"column:buffet_package1_uuid;type:bigint(20);default:0;comment:自助餐套餐1ID" json:"buffet_package1_uuid"`
	BuffetPackage2Uuid uint64 `gorm:"column:buffet_package2_uuid;type:bigint(20);default:0;comment:自助餐套餐2ID" json:"buffet_package2_uuid"`
	DeviceUuid         uint64 `gorm:"column:device_uuid;type:bigint(20);default:0;comment:设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询" json:"device_uuid"`

	// 关联模型
	SaleOrders      []*SaleOrder     `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	SaleBillSetting *SaleBillSetting `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	Cashier         Staff            `gorm:"foreignKey:CashierUuid;references:uuid"`
	Desk            *Desk            `gorm:"foreignKey:DeskUuid;references:uuid"`
	BuffetPackage1  BuffetPackage    `gorm:"foreignKey:BuffetPackage1Uuid;references:uuid"`
	BuffetPackage2  BuffetPackage    `gorm:"foreignKey:BuffetPackage2Uuid;references:uuid"`
}

// 返回新的销售账单
func (model *SaleBill) GetSaleOrder(saleOrderUuid uint64) *SaleOrder {
	for index, saleOrder := range model.SaleOrders {
		if saleOrderUuid == saleOrder.Uuid {
			return model.SaleOrders[index]
		}
	}
	return nil
}

func (model *SaleBill) GetBuffetName() (name dto.LocaleResponse) {
	name1 := model.BuffetPackage1.MultiLanguageName.GetNames()
	name2 := model.BuffetPackage2.MultiLanguageName.GetNames()
	if model.BuffetPackage1.Uuid != 0 && model.BuffetPackage2.Uuid != 0 {
		name = dto.LocaleResponse{
			ZH:   fmt.Sprintf("%s+%s", name1.ZH, name2.ZH),
			TH:   fmt.Sprintf("%s+%s", name1.TH, name2.TH),
			EN:   fmt.Sprintf("%s+%s", name1.EN, name2.EN),
			ZHTW: fmt.Sprintf("%s+%s", name1.ZHTW, name2.ZHTW),
			JA:   fmt.Sprintf("%s+%s", name1.JA, name2.JA),
			KO:   fmt.Sprintf("%s+%s", name1.KO, name2.KO),
			MY:   fmt.Sprintf("%s+%s", name1.MY, name2.MY),
			TR:   fmt.Sprintf("%s+%s", name1.TR, name2.TR),
		}
		return
	}
	// 只有一个自助餐时都是只填在BuffetPackage1
	if model.BuffetPackage1.Uuid != 0 {
		name = dto.LocaleResponse{
			ZH:   fmt.Sprintf("%s", name1.ZH),
			TH:   fmt.Sprintf("%s", name1.TH),
			EN:   fmt.Sprintf("%s", name1.EN),
			ZHTW: fmt.Sprintf("%s", name1.ZHTW),
			JA:   fmt.Sprintf("%s", name1.JA),
			KO:   fmt.Sprintf("%s", name1.KO),
			MY:   fmt.Sprintf("%s", name1.MY),
			TR:   fmt.Sprintf("%s", name1.TR),
		}
		return
	}
	return name
}

func (model *SaleBill) IsDeskSaleBill() bool {
	return model.DeskUuid != 0 // 桌台账单肯定是有桌台ID
}

func (model *SaleBill) IsBuffetSaleBill() bool {
	return model.IsBuffet == 1 // 桌台账单肯定是有桌台ID
}

func (model *SaleBill) BuffetEndTime() int64 {
	endTime := model.BaseModel.CreateTime + int64(model.BuffetDuration)
	return endTime
}

// 自助餐还剩余多少秒。可以为负数，表示自助餐已经结束了多少秒
func (model *SaleBill) BuffetRemainingSeconds() int64 {
	remainingTime := model.BuffetEndTime() - time.Now().Unix()
	return remainingTime
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleBill) ValidateOrderStatus(operation string, saleOrderUuid ...uint64) error {
	if operation != constant.OrderSettle && model.IsLock == 1 {
		return errors.New("订单已被锁定，请解锁后重新操作")
	}
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	if len(model.SaleOrders) > 0 {
		// 拆单没有取消权限
		if operation == constant.OrderOrderCancel && len(model.SaleOrders) > 1 {
			return errors.New("拆单不可操作")
		}
		// 单个订单不能操作
		for _, so := range model.SaleOrders {
			if len(saleOrderUuid) == 0 || slices.Contains(saleOrderUuid, so.Uuid) {
				if err := so.ValidateOrderStatus(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 获取总的退款金额
func (model *SaleBill) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, saleOrder := range model.SaleOrders {
		for _, refundOrder := range saleOrder.ReturnOrders {
			refundAmount += refundOrder.RefundAmount
		}
	}
	return refundAmount
}

// 获取所有自助餐名称
func (model *SaleBill) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, order := range model.SaleOrders {
		for _, buffet := range order.SaleOrderBuffetCustomerTypes {
			name := buffet.BuffetPackageMultiLanguageName.GetNameByLang(language)
			if !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
			}
		}
	}
	return strings.Join(buffets, "+")
}

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

// 计算销售订单原服务费金额。销售订单原服务费金额= 销售订单商品的原服务费之和。
// 不受服务费、按固定服务费费收时，销售订单商品的原服务费=0 或 销售订单商品的原服务费=固定费用
// 按比例收服务费时，商品未含税时，销售订单商品的原服务费=销售订单商品的SalePrice * 服务费费率
// 按比例收服务费时，商品已含税时，销售订单商品的原服务费=销售订单商品的未含税价格 * 服务费费率 = （销售订单商品的SalePrice - 销售订单商品的商品消费税税费） * 服务费费率 = （销售订单商品的SalePrice - （销售订单商品的商品消费税税费）） * 服务费费率
func (model *SaleOrder) CalcOrderOriginServiceFee(serviceFeeRate float64, serviceFeeValue float64, taxFeeType int) float64 {
	originServiceFee := decimal.NewFromFloat(0)
	// 不是按比例收取服务费时
	if serviceFeeRate <= 0 {
		return serviceFeeValue
	} else {
		// 如果按比例收取
		for _, saleOrderProduct := range model.SaleOrderProducts {
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
				continue
			}
			serviceFee := saleOrderProduct.CalcOriginServiceFee(serviceFeeRate, taxFeeType)
			originServiceFee = originServiceFee.Add(decimal.NewFromFloat(serviceFee))
		}
		return originServiceFee.InexactFloat64()
	}

	// 默认，不是按比例收取服务费
	return serviceFeeValue
}

// 计算销售订单的原商品服务费税费。销售订单的原商品服务费税费 = 销售订单的各个商品的原服务费税费之和
// 不是按比例收取服务费时，服务费税费=0
// 按比例收取服务费且收取服务费税费时，销售订单的原商品服务费税费 = 销售订单的各个商品的原服务费税费之和
// 商品的原服务费税费= 原服务费 * 服务费税率
func (model *SaleOrder) CalcOrderOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int) float64 {
	originServiceFee := decimal.NewFromFloat(0)
	// 不是按比例收取服务费时
	if serviceFeeRate <= 0 {
		return 0
	} else {
		// 如果按比例收取
		for _, saleOrderProduct := range model.SaleOrderProducts {
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
				continue
			}
			serviceFee := saleOrderProduct.CalcOriginServiceFee(serviceFeeRate, taxFeeType)
			// 商品的原服务费税费= 原服务费 * 服务费税率
			originServiceTaxFee := decimal.NewFromFloat(serviceFee).Mul(decimal.NewFromFloat(saleOrderProduct.TaxRate))
			// 累加各个订单商品的服务费税费
			originServiceFee = originServiceFee.Add(originServiceTaxFee)
		}
		return originServiceFee.InexactFloat64()
	}
}

// 计算销售订单原价金额。
// 商品未含税时，销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（包括消费税税费和服务费税费）
// 商品已含税时，销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
func (model *SaleOrder) CalcOrderOriginAmount(serviceFeeRate float64, serviceFeeValue float64, taxFeeType int) float64 {
	// 商品未含税时,销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（包括消费税税费和服务费税费）
	if taxFeeType == constant.TaxFeeTypeNoTax {
		// 原服务费
		originService := model.CalcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
		// 原服务费税费
		serviceTaxFee := model.CalcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType)
		// 原商品消费税税费
		originProductTaxFee := model.CalcOriginProductTaxFee(taxFeeType)

		// 销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（包括消费税税费和服务费税费）
		amount := decimal.NewFromFloat(model.ProductOriginalAmount).Add( // 销售订单商品总价（折前价）
			decimal.NewFromFloat(originService)).Add( //  服务费
			decimal.NewFromFloat(serviceTaxFee)).Add( //  服务费消费税金额
			decimal.NewFromFloat(originProductTaxFee)) // 商品消费税税费金额
		return amount.InexactFloat64()
	}
	// 商品已含税时，销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
	if taxFeeType == constant.TaxFeeTypeTax {
		// 服务费
		originService := model.CalcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
		// 服务费税费
		serviceTaxFee := decimal.NewFromFloat(model.CalcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType))
		//销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
		amount := decimal.NewFromFloat(model.ProductOriginalAmount).Add(
			decimal.NewFromFloat(originService)).Add(
			serviceTaxFee)
		return amount.InexactFloat64()
	}

	// 默认按商品已含税处理。
	// 服务费
	originService := model.CalcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
	// 服务费税费
	serviceTaxFee := decimal.NewFromFloat(model.CalcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType))
	//销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
	amount := decimal.NewFromFloat(model.ProductOriginalAmount).Add(
		decimal.NewFromFloat(originService)).Add(
		serviceTaxFee)
	return amount.InexactFloat64()
}

// 计算销售订单已经支付的金额。 销售订单已经支付的金额= 销售订单的所有支付单的支付金额之和
func (model *SaleOrder) CalcPayOrderAmount() float64 {
	payAmount := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		payAmount = payAmount.Add(decimal.NewFromFloat(paymentOrder.PaymentAmount))
	}
	return payAmount.InexactFloat64()
}

// 计算销售订单已经支付的付款单手续费
func (model *SaleOrder) CalcCommissionFee() float64 {
	commissionFee := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		commissionFee = commissionFee.Add(decimal.NewFromFloat(paymentOrder.PaymentCommissionFee))
	}
	return commissionFee.InexactFloat64()
}

// 计算销售订单的最终应收金额。
// 最终应收=应收金额+支付手续费= 应收金额 +（各个支付订单的手续费之和+当前支付方式的手续费）
func (model *SaleOrder) CalcFinallyAmount() (float64, bool) {
	hasCommission := false
	amount := decimal.NewFromFloat(model.Amount)
	commissionFee := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		commissionFee = commissionFee.Add(decimal.NewFromFloat(paymentOrder.PaymentCommissionFee))
	}
	// 未支付的金额的手续费 = 未支付的金额 * 支付手续费费率
	// 最终应收 = 应收金额+已支付的手续费
	finallyAmount := amount.Add(commissionFee)

	if commissionFee.InexactFloat64() > 0 {
		hasCommission = true
	}
	return finallyAmount.InexactFloat64(), hasCommission
}

// 计算销售订单未付款的金额。
// 支付没有手续费时，销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
// 支付有手续费时，销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和
func (model *SaleOrder) CalcUnPayAmount(hasCommission bool) float64 {
	if hasCommission {
		// 销售订单各个支付单的支付金额之和
		payOrderAmount := model.CalcPayOrderAmount()
		// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和
		unPayAmount := decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(payOrderAmount))
		return unPayAmount.InexactFloat64()
	}
	// 没有手续费时
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	// 销售订单各个支付单的支付金额之和
	payOrderAmount := model.CalcPayOrderAmount()
	zeroFee := model.CalcCheckOutZeroFee()
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	unPayAmount := decimal.NewFromFloat(model.Amount).Sub(
		decimal.NewFromFloat(payOrderAmount)).Sub(
		decimal.NewFromFloat(zeroFee))
	return unPayAmount.InexactFloat64()

}

// 计算订单商品SalePrice之和。等于所有已接单商品SalePrice之和
func (model *SaleOrder) calcSumOrderProductSalePrice() float64 {
	sumSalePrice := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 销售订单商品已接单且未删除商品
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？要累计上
			//退菜？不累计
			if orderProduct.IsCancelBool() {
				continue
			}
			// SalePrice * Num
			fmt.Println("orderProduct.SalePrice", orderProduct.SalePrice, "orderProduct.uuid", orderProduct.Uuid)
			saleProductSalePrice := decimal.NewFromFloat(orderProduct.SalePrice).Mul(decimal.NewFromUint64(uint64(orderProduct.Num)))
			sumSalePrice = sumSalePrice.Add(saleProductSalePrice)
			fmt.Println("sumSalePrice:", sumSalePrice.InexactFloat64())
		}
	}
	return sumSalePrice.InexactFloat64()
}

// 计算订单商品Price之和。等于所有已接单商品Price之和
func (model *SaleOrder) calcSumOrderProductPrice() float64 {
	sumPrice := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 销售订单商品已接单且未删除商品
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			// price * num
			price := decimal.NewFromFloat(orderProduct.Price).Mul(decimal.NewFromUint64(uint64(orderProduct.Num)))
			sumPrice = sumPrice.Add(price)
		}
	}
	return sumPrice.InexactFloat64()
}

// 计算订单商品金额（折前价）。订单商品金额（折前价）= 订单商品SalePrice之和 + 自助餐顾客价格CustomerPrice之和 + 自助餐加钟商品价格Price之和
func (model *SaleOrder) CalcProductOriginalAmount() float64 {
	return model.calcSumOrderProductSalePrice() // todo + 自助餐顾客价格之和 + 自助餐加钟商品价格之和
}

// 计算订单商品金额（折后价）。订单商品金额（折后价）= 订单商品Price之和 + 自助餐顾客价格Price之和 + 自助餐加钟商品价格Price
func (model *SaleOrder) CalcProductAmount() float64 {
	return model.calcSumOrderProductPrice() // todo + 自助餐顾客价格之和 + 自助餐加钟商品价格之和
}

// 计算订单产生的税费。订单税费=订单商品TaxFee之和 + 订单商品ServiceTaxFee之和
func (model *SaleOrder) CalcTaxFee() float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			taxFee = taxFee.Add(
				decimal.NewFromFloat(orderProduct.TaxFee)).Add(
				decimal.NewFromFloat(orderProduct.ServiceTaxFee))
		}
	}
	return taxFee.InexactFloat64()
}

// 计算销售订单的原商品消费税金额。 销售订单的原商品消费税金额= 销售订单的各个商品的原商品消费税金额之和。
func (model *SaleOrder) CalcOriginProductTaxFee(taxFeeType int) float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			taxFee = taxFee.Add(
				decimal.NewFromFloat(orderProduct.CalcOriginTaxFee(taxFeeType)))
		}
	}
	return taxFee.InexactFloat64()
}

// 计算销售订单的自定义优惠折扣金额。订单自定义优惠金额=销售订单商品自定义优惠金额之和 + 自助餐顾客自定义优惠金额之和
// todo 考虑每个价格循环一次商品列表计算带来的性能损耗与计算业务更清晰抉择哪个
func (model *SaleOrder) CalcCustomDiscountFee() float64 {
	customDiscountFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			customDiscountFee = customDiscountFee.Add(
				decimal.NewFromFloat(orderProduct.CustomDiscountFee))
		}
	}
	// todo  + 自助餐顾客自定义优惠金额之和
	return customDiscountFee.InexactFloat64()
}

// 计算销售订单会员折扣金额。销售订单会员折扣金额=订单商品会员折扣金额之和
func (model *SaleOrder) CalcMemberDiscountFee() float64 {
	memberDiscountFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			memberDiscountFee = memberDiscountFee.Add(
				decimal.NewFromFloat(orderProduct.MemberDiscountFee))
		}
	}
	return memberDiscountFee.InexactFloat64()
}

// 计算销售订单服务费消费税金额。销售订单服务费消费税金额=订单商品服务费消费税金额之和
func (model *SaleOrder) CalcServiceTaxFee() float64 {
	serviceTaxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			serviceTaxFee = serviceTaxFee.Add(
				decimal.NewFromFloat(orderProduct.ServiceTaxFee))
		}
	}
	return serviceTaxFee.InexactFloat64()
}

// 计算销售订单应付金额。销售订单应付金额=商品金额+服务费+消费税。 给前端显示时，销售订单应付金额=商品金额+服务费+消费税-订单抹零金额
// 商品未含税时，销售订单应付金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。
// 商品已含税时，销售订单应付金额=商品金额（包含商品消费税税费）+服务费+服务费税费。
// 商品关闭税费时，销售订单应付金额=商品金额ProductAmount(折后)+服务费
func (model *SaleOrder) CalcAmount(taxFeeType int) float64 {
	amount := decimal.NewFromFloat(0)
	// 商品已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		serviceTaxFee := model.CalcServiceTaxFee()
		//商品金额（包含商品消费税税费）+服务费+服务费税费。
		amount = amount.Add(
			decimal.NewFromFloat(model.ProductAmount)).Add(
			decimal.NewFromFloat(model.ServiceFee)).Add(
			decimal.NewFromFloat(serviceTaxFee))
		return amount.InexactFloat64()
	}
	// 商品未含税时
	if taxFeeType == constant.TaxFeeTypeNoTax {
		// 销售订单应付金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）
		amount = amount.Add(
			decimal.NewFromFloat(model.ProductAmount).Add(
				decimal.NewFromFloat(model.ServiceFee).Add(
					decimal.NewFromFloat(model.TaxFee))))
		return amount.InexactFloat64()
	}
	// 商品关闭税费时
	// 销售订单应付金额=商品金额ProductAmount(折后)+服务费
	result := decimal.NewFromFloat(model.ProductAmount).Add(decimal.NewFromFloat(model.ServiceFee))

	inexactFloat64 := result.InexactFloat64()
	return inexactFloat64
}

// 计算销售订单的订单优惠折扣抹零金额。根据订单设置的优惠折扣抹零规则金额计算
func (model *SaleOrder) CalcZeroFee() float64 {
	amount := model.Amount
	switch model.ZeroRule {
	// 实款实收
	case constant.SaleBillSettingDiscountZeroingMethodNone:
		return 0
	// 抹分
	case constant.SaleBillSettingDiscountZeroingMethodPercent:
		// 抹分后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(1)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 抹角
	case constant.SaleBillSettingDiscountZeroingMethodFixed:
		// 抹角后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(0)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 四舍五入保留一位小数
	case constant.SaleBillSettingDiscountZeroingMethodRound:
		discountAmount := decimal.NewFromFloat(amount).Round(1)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 四舍五入保留整数
	case constant.SaleBillSettingDiscountZeroingMethodInteger:
		discountAmount := decimal.NewFromFloat(amount).Round(0)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	default:
		return 0
	}
}

// 计算销售订单的结账抹零金额。根据订单设置的结账抹零规则金额计算
func (model *SaleOrder) CalcCheckOutZeroFee() float64 {
	amount := model.Amount
	switch model.ZeroCheckoutRule {
	// 实款实收
	case constant.SaleBillSettingCheckoutZeroingMethodNone:
		return 0
	// 抹分
	case constant.SaleBillSettingCheckoutZeroingMethodPercent:
		// 抹分后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(1)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 抹角
	case constant.SaleBillSettingCheckoutZeroingMethodFixed:
		// 抹角后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(0)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 抹元
	case constant.SaleBillSettingCheckoutZeroingMethodYuan:
		// 原订单金额/10 后抹去小数，再✖乘10，以此来实现抹元
		truncate := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(10)).Truncate(0)
		discountAmount := truncate.Mul(decimal.NewFromFloat(10))
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	default:
		return 0
	}
}

// 计算订单服务费。
// 当服务费关闭时，订单服务费为0
// 当服务费为固定费用时，订单服务费为固定费用
// 当服务费为按比例收费时，订单服务费为所有订单商品的服务费之和
func (model *SaleOrder) CalcServiceFee(serviceFeeType int, serviceFeeValue float64) float64 {
	// 当服务费关闭时，订单服务费为0
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypeNone {
		return 0
	}
	// 当服务费为固定费用时，订单服务费为固定费用
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypeFixed {
		return serviceFeeValue
	}
	// 当服务费为按比例收费时，订单服务费为所有订单商品的服务费之和
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercent || serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		serviceFee := decimal.NewFromFloat(0)
		for _, orderProduct := range model.SaleOrderProducts {
			// 销售商品已接单且为删除
			if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
				// 不计入赠菜、不计入退菜
				if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
					continue
				}
				serviceFee = serviceFee.Add(decimal.NewFromFloat(orderProduct.ServiceFee))
			}
		}
	}
	// 默认不收取服务费
	return 0
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

// 重新计算销售订单的金额
func (model *SaleOrder) CalcSaleOrder(serviceFeeType int, serviceFeeValue float64, taxFeeType int) *Calc {
	calc := Calc{}
	calc.ProductOriginalAmount = model.CalcProductOriginalAmount()
	model.ProductOriginalAmount = calc.ProductOriginalAmount
	calc.ProductAmount = model.CalcProductAmount()
	model.ProductAmount = calc.ProductAmount
	calc.ServiceFee = model.CalcServiceFee(serviceFeeType, serviceFeeValue)
	model.ServiceFee = calc.ServiceFee
	calc.TaxFee = model.CalcTaxFee()
	model.TaxFee = calc.TaxFee
	calc.CustomDiscountFee = model.CalcCustomDiscountFee()
	model.CustomDiscountFee = calc.CustomDiscountFee
	calc.MemberDiscountFee = model.CalcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.Amount = model.CalcAmount(taxFeeType)
	model.Amount = calc.Amount
	calc.ZeroFee = model.CalcCheckOutZeroFee()
	model.ZeroFee = calc.ZeroFee
	return &calc
}

// TableName 指定表名
func (SaleOrder) TableName() string {
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

// SaleOrderProduct 销售订单产品 `ttpos_sale_order_product`
type SaleOrderProduct struct {
	// 基础字段
	BaseModel

	// 基本信息字段
	Name       string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品名称'" json:"name"`
	FlavorName string `gorm:"column:flavor_name;type:varchar(255);not null;default:'';comment:'规格名称'" json:"flavor_name"`
	Num        uint   `gorm:"column:num;type:int(11);not null;default:0;comment:'商品数量。不能减为0，当数量为1再减时，标记删除'" json:"num"`
	Remark     string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'备注，顾客对商品的备注信息'" json:"remark"`

	// 状态相关字段
	Status        uint `gorm:"column:status;type:tinyint(1);not null;default:0;comment:'状态, 0-未送厨 1-已送厨 2-退菜'" json:"status"`
	IsRequire     uint `gorm:"column:is_require;type:tinyint(1);not null;default:0;comment:'是否必点商品 0-否 1-是。用于在前端显示必点图标'" json:"is_require"`
	IsAcceptOrder uint `gorm:"column:is_accept_order;type:tinyint(1);not null;default:0;comment:'是否已接单, 0-否 1-是'" json:"is_accept_order"`

	// 价格相关字段
	FlavorPrice  float64 `gorm:"column:flavor_price;type:decimal(12,2);not null;default:0.00;comment:'规格原价（单商品）,仅某规格商品的原价'" json:"flavor_price"`
	SaucePrice   float64 `gorm:"column:sauce_price;type:decimal(12,2);not null;default:0.00;comment:'小料价（单商品）,所有小料的价格之和'" json:"sauce_price"`
	ProductPrice float64 `gorm:"column:product_price;type:decimal(12,2);not null;default:0.00;comment:'原始单价（单商品）,规格原价+小料价'" json:"product_price"`
	SalePrice    float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0.00;comment:'销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价'" json:"sale_price"`
	Price        float64 `gorm:"column:price;type:decimal(12,2);not null;default:0.00;comment:'最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率'" json:"price"`
	TotalPrice   float64 `gorm:"column:total_price;type:decimal(12,2);not null;default:0.00;comment:'应收金额(单商品)=最终单价+服务费+总税费'" json:"total_price"`

	// 折扣相关字段
	ChangePriceTime        int64   `gorm:"column:change_price_time;type:int(10);not null;default:0;comment:'改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并'" json:"change_price_time"`
	OpenMemberDiscount     uint    `gorm:"column:open_member_discount;type:tinyint(1);not null;default:0;comment:'是否开启会员折扣, 0-否 1-是'" json:"open_member_discount"` // 快照设置相关，不受后台改变，结账时检查
	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员折扣率(0-100%)'" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员卡折扣率(0-100%)'" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣率(0-100%)'" json:"custom_discount_rate"`

	// 折扣金额字段
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);not null;default:0.00;comment:'打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额'" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);not null;default:0.00;comment:'会员折扣金额（单商品）=销售价*会员折扣率*会员卡折扣率'" json:"member_discount_fee"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣金额（单商品）=销售价-最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)'" json:"custom_discount_fee"`

	// 税费和服务费字段
	TaxRate       float64 `gorm:"column:tax_rate;type:decimal(12,2);not null;default:0;comment:'税率,单位%.加购时记录税率,结账时再重新核算'" json:"tax_rate"`
	ServiceTaxFee float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率'" json:"service_tax_fee"`
	TaxFee        float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0.00;comment:'商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率'" json:"tax_fee"`
	ServiceFee    float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例'" json:"service_fee"`

	// 库存相关字段
	DeductStockType uint  `gorm:"column:deduct_stock_type;type:tinyint(1);not null;default:0;comment:'库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数'" json:"deduct_stock_type"`
	DeductStockTime int64 `gorm:"column:deduct_stock_time;type:int(10);not null;default:0;comment:'减库存的时间(时间戳），0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存'" json:"deduct_stock_time"`

	// 赠品相关字段
	GiftTime     int64  `gorm:"column:gift_time;type:int(10);not null;default:0;comment:'赠菜时间(时间戳),用于判断不同时间赠送的商品不合并'" json:"gift_time"`
	CancelTime   int64  `gorm:"column:cancel_time;type:int(10);not null;default:0;comment:'退菜时间(时间戳),用于判断不同时间退菜的商品不合并'" json:"cancel_time"`
	GiftReason   string `gorm:"column:gift_reason;type:varchar(255);not null;default:'';comment:'赠菜原因'" json:"gift_reason"`
	CancelReason string `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:'退菜原因'" json:"refund_reason"`

	// 关联ID字段
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称UUID'" json:"multi_language_name_uuid"`
	ImageFileUuid         uint64 `gorm:"column:image_file_uuid;not null;default:0;comment:'商品图片ID'" json:"image_file_uuid"`
	ProductionOrderUuid   uint64 `gorm:"column:production_order_uuid;type:bigint(20);not null;default:0;comment:'生产订单ID'" json:"production_order_uuid"`
	ProductPackageUuid    uint64 `gorm:"column:product_package_uuid;type:bigint(20);not null;default:0;comment:'商品包ID'" json:"product_package_uuid"`
	SaleBillUuid          uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);not null;default:0;comment:'销售账单ID'" json:"sale_bill_uuid"`
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:'销售订单ID'" json:"sale_order_uuid"`

	// 其他字段
	Sign             string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:'商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品'" json:"sign"`
	IsH5OrderProduct uint   `gorm:"column:is_h5_order_product;type:tinyint(1);not null;default:0;comment:'是否为扫码订单商品, 0-否 1-是'" json:"is_qrcode_order_product"`

	// 关联对象
	MultiLanguageName          MultiLanguageName           `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	ImageFile                  File                        `gorm:"foreignKey:image_file_uuid;references:uuid"`
	SaleOrderProductBoms       []SaleOrderProductBom       `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
	SaleOrderProductAttributes []SaleOrderProductAttribute `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ReturnOrderProducts        []ReturnOrderProduct        `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ProductPackage             ProductPackage              `gorm:"foreignKey:ProductPackageUuid;references:Uuid"`
}

// 更新销售订单商品的折扣信息
func (model *SaleOrderProduct) SetDiscountInfo(memberDiscountRate, memberCardDiscountRate, customDiscountRate float64) {
	model.MemberDiscountRate = memberDiscountRate
	model.MemberCardDiscountRate = memberCardDiscountRate
	model.CustomDiscountRate = customDiscountRate
}

// 标记该订单商品相关的资源为删除
func (model *SaleOrderProduct) DeleteProduct() {
	deleteTime := time.Now().Unix()
	model.DeleteTime = deleteTime
	for index, _ := range model.SaleOrderProductBoms {
		saleOrderProductBom := model.SaleOrderProductBoms[index]
		saleOrderProductBom.DeleteTime = deleteTime
	}
	for index, _ := range model.SaleOrderProductAttributes {
		saleOrderProductAttribute := model.SaleOrderProductAttributes[index]
		saleOrderProductAttribute.DeleteTime = deleteTime
	}
}

func (model *SaleOrderProduct) IsAcceptOrderBool() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

func (model *SaleOrderProduct) IsGiftBool() bool {
	return model.GiftTime > 0
}
func (model *SaleOrderProduct) IsCancelBool() bool {
	return model.CancelTime > 0
}

type Sauce struct {
	Name           string
	Price          float64
	ProductBomUuid uint64
}
type Flavor struct {
	Name           string
	Price          float64
	ProductBomUuid uint64
}
type Attribute struct {
	Name                 string
	ProductAttributeUuid uint64
}

// 点餐时录入的原始数据
type DefaultSaleOrderProduct struct {
	Name                   string
	IsRequire              uint
	OpenMemberDiscount     uint
	TaxRate                float64
	DeductStockType        uint
	MultiLanguageNameUuid  uint64
	ImageFileUuid          uint64
	ProductPackageUuid     uint64
	SaleBillUuid           uint64
	SaleOrderUuid          uint64
	MemberDiscountRate     float64
	MemberCardDiscountRate float64
	CustomDiscountRate     float64
	Sauces                 []Sauce
	Flavor                 Flavor
	Attribute              []Attribute
}

func NewDefaultSaleOrderProduct(def DefaultSaleOrderProduct) *SaleOrderProduct {
	saleOrderProductUuid, _ := utils.GetID()
	saleOrderProductBoms := make([]SaleOrderProductBom, 0)
	for _, bom := range def.Sauces {
		saleOrderProductBom := SaleOrderProductBom{
			Name:                 bom.Name,
			Price:                bom.Price,
			IsFlavorBom:          0,
			SaleOrderProductUuid: saleOrderProductUuid,
			SaleOrderUuid:        def.SaleOrderUuid,
			ProductBomUuid:       bom.ProductBomUuid,
		}
		saleOrderProductBoms = append(saleOrderProductBoms, saleOrderProductBom)
	}
	saleOrderProductBoms = append(saleOrderProductBoms, SaleOrderProductBom{
		Name:                 def.Flavor.Name,
		Price:                def.Flavor.Price,
		IsFlavorBom:          1,
		SaleOrderUuid:        def.SaleOrderUuid,
		SaleOrderProductUuid: saleOrderProductUuid,
		ProductBomUuid:       def.Flavor.ProductBomUuid,
	})

	saleOrderProductAttributes := []SaleOrderProductAttribute{}
	for _, attribute := range def.Attribute {
		saleOrderProductAttribute := SaleOrderProductAttribute{
			Name:                 attribute.Name,
			SaleOrderUuid:        def.SaleOrderUuid,
			SaleOrderProductUuid: saleOrderProductUuid,
			ProductAttributeUuid: attribute.ProductAttributeUuid,
		}
		saleOrderProductAttributes = append(saleOrderProductAttributes, saleOrderProductAttribute)
	}
	product := SaleOrderProduct{
		BaseModel: BaseModel{
			Uuid: saleOrderProductUuid,
		},
		Name:                       def.Name,
		FlavorName:                 def.Flavor.Name,
		Num:                        1,
		Status:                     constant.OrderProductStatusUnSending,
		IsAcceptOrder:              1,
		IsRequire:                  def.IsRequire,
		FlavorPrice:                def.Flavor.Price,
		OpenMemberDiscount:         def.OpenMemberDiscount,
		MemberDiscountRate:         def.MemberDiscountRate,
		MemberCardDiscountRate:     def.MemberCardDiscountRate,
		CustomDiscountRate:         def.CustomDiscountRate,
		TaxRate:                    def.TaxRate,
		DeductStockType:            def.DeductStockType,
		MultiLanguageNameUuid:      def.MultiLanguageNameUuid,
		ImageFileUuid:              def.ImageFileUuid,
		ProductPackageUuid:         def.ProductPackageUuid,
		SaleBillUuid:               def.SaleBillUuid,
		SaleOrderUuid:              def.SaleOrderUuid,
		SaleOrderProductBoms:       saleOrderProductBoms,
		SaleOrderProductAttributes: saleOrderProductAttributes,
	}
	return &product
}

// 计算小料的价格。累计销售订单商品的所有小料的价格
func (model *SaleOrderProduct) CalcSaucePrice() float64 {
	saucePrice := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if !bom.IsFlavor() {
			// 累加每个小料的价格
			saucePrice.Add(decimal.NewFromFloat(bom.Price))
		}
	}
	return saucePrice.InexactFloat64()
}

// 计算商品价格。某个规格商品价+小料价
// 当商品没有改价时,ProductPrice= 某个规格商品价+小料价
// 当商品改价时，ProductPrice= ProductPrice 。 改价后不会修改这个字段的值，只会修改salePrice的值
func (model *SaleOrderProduct) CalcProductPrice() float64 {
	//if model.ChangePriceTime == constant.CustomPriceOn {
	//	return model.ProductPrice
	//}
	productPrice := decimal.NewFromFloat(model.FlavorPrice).Add(decimal.NewFromFloat(model.CalcSaucePrice()))
	return productPrice.InexactFloat64()
}

// 判断商品有没有改价
func (model *SaleOrderProduct) IsCustomPriceBool() bool {
	return model.ChangePriceTime > 0
}

// 计算商品销售价。如果商品改价，则直接修改SalePrice。如果没有改价，销售价=ProductPrice
func (model *SaleOrderProduct) CalcSalePrice() float64 {
	if model.IsCustomPriceBool() {
		return model.SalePrice
	}
	return model.ProductPrice
}

// 计算商品的折扣率。 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
func (model *SaleOrderProduct) CalcDiscountRate() float64 {
	rate := decimal.NewFromFloat(1)
	memberDiscountRate := model.MemberDiscountRate
	memberCardDiscountRate := model.MemberCardDiscountRate
	customDiscountRate := model.CustomDiscountRate

	if memberDiscountRate != 0 {
		// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
		rate = rate.Mul(decimal.NewFromFloat(memberDiscountRate))
	}
	if memberCardDiscountRate != 0 {
		// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
		rate = rate.Mul(decimal.NewFromFloat(memberCardDiscountRate))
	}
	if customDiscountRate != 0 {
		// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
		rate = rate.Mul(decimal.NewFromFloat(customDiscountRate))
	}
	return rate.InexactFloat64()
}

// 计算商品折后价。最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率
func (model *SaleOrderProduct) CalcPrice() float64 {
	discountRate := model.CalcDiscountRate()
	if discountRate == 0 {
		fmt.Println("-3333333----model.SalePrice:", model.SalePrice)
		return model.SalePrice
	}
	// 销售价*折扣率
	price := decimal.NewFromFloat(model.SalePrice).Mul(
		decimal.NewFromFloat(discountRate))
	return price.InexactFloat64()
}

// 计算会员折扣率。会员折扣率=会员等级折扣率*会员卡折扣率
// 如果商品不参与会员打折的话，会员折扣率=0
func (model *SaleOrderProduct) CalcMemberDiscountRate() float64 {
	if model.OpenMemberDiscount == constant.ProductMemberDiscountOff {
		return 0
	}
	if model.MemberDiscountRate == 0 && model.MemberCardDiscountRate != 0 {
		return model.MemberCardDiscountRate
	} else if model.MemberCardDiscountRate == 0 && model.MemberDiscountRate != 0 {
		return model.MemberDiscountRate
	} else if model.MemberCardDiscountRate != 0 && model.MemberDiscountRate != 0 {
		memberDiscountRate := decimal.NewFromFloat(model.MemberDiscountRate).Mul(decimal.NewFromFloat(model.MemberCardDiscountRate))
		return memberDiscountRate.InexactFloat64()
	}
	// 不匹配时默认为0
	return 0
}

// todo 会员折扣率 要除100才是%为单位。无折扣时，会员折扣率为0。会员折扣率的取值范围0-1，0表示没有折扣，1表示全免、满折扣、无需付钱
// 计算商品的会员折扣费用。商品销售价-商品销售价*会员折扣率=商品销售价*（1-会员折扣率）
func (model *SaleOrderProduct) CalcMemberDiscountFee() float64 {
	// 当会员折扣率为0时，会员折扣费用=0
	memberDiscountRate := model.CalcMemberDiscountRate()
	fmt.Println("memberDiscountRate1111111::", memberDiscountRate)

	if memberDiscountRate == 0 {
		return 0
	}
	fmt.Println("memberDiscountRate::", memberDiscountRate)
	// 1-会员折扣率
	discount := decimal.NewFromFloat(1).Sub(decimal.NewFromFloat(memberDiscountRate))
	// 商品销售价*（1-会员折扣率）
	memberDiscountFee := decimal.NewFromFloat(model.CalcSalePrice()).Mul(discount)
	return memberDiscountFee.InexactFloat64()
}

// 当有会员折扣时，自定义折扣费  = 会员折扣价-会员折扣价*自定义折扣率 = 会员折扣价*（1-自定义折扣率）=（商品销售价-会员折扣费）*（1-自定义折扣率）；
// 当没有会员折扣时，自定义折扣费= 商品销售价- 商品销售价*自定义折扣率 = 商品销售价*（1-自定义折扣率）= （商品销售价-会员折扣费0）*（1-自定义折扣率）
// 当没有会员折扣时，会员折扣费为0，则两个情况的算法可以都用 自定义折扣费=会员折扣价*（1-自定义折扣率）
func (model *SaleOrderProduct) CalcCustomDiscountFee() float64 {
	customDiscountRate := model.CustomDiscountRate
	if customDiscountRate == 0 {
		return 0
	}
	// 会员折扣价 = 商品销售价-会员折扣费。没有会员时，会员折扣费为0。
	memberDiscountPrice := decimal.NewFromFloat(model.CalcSalePrice()).Sub(decimal.NewFromFloat(model.CalcMemberDiscountFee()))
	//（1-自定义折扣率）
	discount := decimal.NewFromFloat(1).Sub(decimal.NewFromFloat(customDiscountRate))
	// 会员折扣价*（1-自定义折扣率）
	customDiscountFee := memberDiscountPrice.Mul(discount)
	return customDiscountFee.InexactFloat64()
}

// 计算某规格商品未含税原价。当商品未含税时，未含税原价=商品原价；当商品已含税时，未含税原价=某规格商品原价/（1+消费税税率）
// 计算过程：某规格商品未含税原价=某规格商品原价-消费税税费 = 某规格商品原价-（某规格商品未含税原价*消费税税率）
// 某规格商品原价= 某规格商品未含税原价+（某规格商品未含税原价*消费税税率）= 某规格商品未含税原价 * （ 1+ 1*消费税税率）= 某规格商品未含税原价 * （1+消费税税率）
// 某规格商品未含税原价 = 某规格商品原价/（1+消费税税率）
func (model *SaleOrderProduct) CalcProductPriceNoneTax(price float64, taxFeeType int) float64 {
	//price := model.Price // 当计算商品的最终商品消费税税费时，按商品的最终单价（折后价）来计算
	//price := model.Price // 当计算商品折扣前的商品消费税税费时，按商品的salePrice（折前价）来计算

	// 不收取税费时、关闭税费时, 某规格商品未含税原价=某规格商品原价
	if taxFeeType == constant.TaxFeeTypeNone {
		return price
	}
	// 商品未含税, 某规格商品未含税原价=某规格商品原价
	if taxFeeType == constant.TaxFeeTypeNoTax {
		return price
	}
	// 商品已含税,某规格商品未含税原价= 某规格商品原价/（1+消费税税率）
	// 计算过程：某规格商品未含税原价=某规格商品原价-消费税税费 = 某规格商品原价-（某规格商品未含税原价*消费税税率）
	// 某规格商品原价= 某规格商品未含税原价+（某规格商品未含税原价*消费税税率）= 某规格商品未含税原价 * （ 1+ 1*消费税税率）= 某规格商品未含税原价 * （1+消费税税率）
	// 某规格商品未含税原价 = 某规格商品原价/（1+消费税税率）
	if taxFeeType == constant.TaxFeeTypeTax {
		// （1+消费税税率）
		percent := decimal.NewFromFloat(1).Add(decimal.NewFromFloat(model.TaxRate))
		// 某规格商品原价/（1+消费税税率）
		productPriceNoneTax := decimal.NewFromFloat(price).Div(percent)
		return productPriceNoneTax.InexactFloat64()
	}
	// 不收取税费时
	return price
}

// 计算商品的税费（单商品的消费税税费，不含服务费的税费）
// 商品未含税时，商品原价=商品未含税原价。消费税税费=商品未含税原价*消费税税率
// 商品已含税时，商品原价=商品未含税原价+消费税税费；故，消费税税费=商品原价-商品未含税原价。
// 商品原价=商品未含税原价+（商品未含税原价*消费税税率）= 商品未含税原价*（1+消费税税率）
// 商品未含税时，价格中消费税税费为0，故 商品原价=商品未含税原价+消费税税费0，两种情况中可以用公式：消费税税费=商品原价-商品未含税原价
func (model *SaleOrderProduct) CalcTaxFee(price float64, taxFeeType int) float64 {
	// 商品已含税时，消费税税费=商品销售价-商品未含税销售价
	if taxFeeType == constant.TaxFeeTypeTax {
		taxFee := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(model.CalcProductPriceNoneTax(price, taxFeeType)))
		return taxFee.InexactFloat64()
	} else if taxFeeType == constant.TaxFeeTypeNoTax {
		// 商品未含税时，消费税税费=商品销售价*消费税税率
		taxFee := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(model.TaxRate))
		return taxFee.InexactFloat64()
	}

	// 默认为商品未含税时
	taxFee := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(model.TaxRate))
	return taxFee.InexactFloat64()
}

// 计算商品原税费。
// 商品未含税时，商品原税费= salePrice * 消费税税率
// 商品已含税时，商品原税费= 商品未含税金额 * 消费税税率
func (model *SaleOrderProduct) CalcOriginTaxFee(taxFeeType int) float64 {
	price := model.SalePrice
	return model.CalcTaxFee(price, taxFeeType)
}

// 服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例
func (model *SaleOrderProduct) CalcServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
	price := model.CalcPrice()

	// 服务费关闭时，不收取服务费。视为服务费比例为0，不收取
	// 服务费按固定费用收取时，不收取单收订单商品的服务费。视为服务费比例为0，不收取
	// 服务费按比例收取时，根据服务费比例收取
	if serviceFeeRate <= 0 {
		return 0
	}
	// 服务费按比例收取时，根据服务费比例收取
	// 服务费比例大于0时，
	// 商品未含税时，服务费=最终单价（折扣价）*服务费比例
	// 商品已含税时，服务费=商品未含税价格*服务费比例=（最终单价-商品税费）*服务费比例
	if taxFeeType == constant.TaxFeeTypeNoTax {
		// 未含税时
		serviceFee := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.InexactFloat64()
	}
	// 已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		// 商品未含税价格=（最终单价-商品税费）
		priceNoneTax := model.CalcProductPriceNoneTax(price, taxFeeType) // decimal.NewFromFloat(model.CalcPrice()).Sub(decimal.NewFromFloat(model.CalcTaxFee(taxFeeType)))
		priceNoneTaxDecimal := decimal.NewFromFloat(priceNoneTax)
		//  服务费=（最终单价-商品税费）*服务费比例
		serviceFee := priceNoneTaxDecimal.Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.InexactFloat64()
	}
	// 默认商品不收取服务费
	return 0
}

// 计算商品原服务费，即为打折前的服务费。0-固定服务费 大于0-按比例收服务费；
// 商品已含税时，服务费=(销售价-商品税费)*服务费比例；
// 商品未含税时，服务费=销售价*服务费比例
func (model *SaleOrderProduct) CalcOriginServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
	// 服务费关闭时，不收取服务费。视为服务费比例为0，不收取
	// 服务费按固定费用收取时，不收取单收订单商品的服务费。视为服务费比例为0，不收取
	// 服务费按比例收取时，根据服务费比例收取
	if serviceFeeRate <= 0 {
		return 0
	}
	// 服务费按比例收取时，根据服务费比例收取
	// 服务费比例大于0时，
	// 商品未含税时，服务费=销售价（折前价）*服务费比例
	if taxFeeType == constant.TaxFeeTypeNoTax {
		// 未含税时
		serviceFee := decimal.NewFromFloat(model.CalcSalePrice()).Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.InexactFloat64()
	}
	// 已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		// 商品未含税价格=（销售价-商品原税费）
		priceNoneTax := decimal.NewFromFloat(model.CalcSalePrice()).Sub(decimal.NewFromFloat(model.CalcTaxFee(model.SalePrice, taxFeeType)))
		//  服务费=（最终单价-商品税费）*服务费比例
		serviceFee := priceNoneTax.Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.InexactFloat64()
	}
	// 默认商品不收取服务费
	return 0
}

// 计算订单商品的服务费税费。
// 当不收取服务费税费时，服务费税费为0
// 当收取服务费税费时，服务费税费=订单商品服务费*商品消费税税率
func (model *SaleOrderProduct) CalcServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 当服务费收费税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceTaxFee := decimal.NewFromFloat(model.CalcServiceFee(serviceFeeRate, taxFeeType)).Mul(decimal.NewFromFloat(model.TaxRate))
		return serviceTaxFee.InexactFloat64()
	}
	return 0
}

// 计算订单商品的原服务费税费（打折前）。
// 当不收取服务费税费时，服务费税费为0
// 当收取服务费税费时，服务费税费=订单商品服务费（打折前）*商品消费税税率
func (model *SaleOrderProduct) CalcOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 当服务费收费税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceTaxFee := decimal.NewFromFloat(model.CalcOriginServiceFee(serviceFeeRate, taxFeeType)).Mul(decimal.NewFromFloat(model.TaxRate))
		return serviceTaxFee.InexactFloat64()
	}
	return 0
}

// 计算商品的总折扣费用。单个商品总折扣费用=会员折扣费用+自定义打折折扣费用
func (model *SaleOrderProduct) CalcDiscountFee() float64 {
	// 会员折扣费用+自定义打折折扣费用
	discountFee := decimal.NewFromFloat(model.CalcMemberDiscountFee()).Add(decimal.NewFromFloat(model.CalcCustomDiscountFee()))
	return discountFee.InexactFloat64()
}

// 计算单个商品最终应收金额。
// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+总税费=最终价格（折后价）+服务费+（商品税费+服务费税费）
// 商品已含税时，单个商品最终应收金额=商品折后不含税价格+服务费+总税费=（最终价格（折扣价）-商品税费） + 服务费 + 总税费 = （最终价格（折后价）-商品税费） + 服务费 + （商品税费+服务费税费）= 最终价格（折后价）+ 服务费 + 服务费税费
// 总结，商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
func (model *SaleOrderProduct) CalcTotalPrice(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+商品税费+服务费税费
	price := model.CalcPrice()
	if taxFeeType == constant.TaxFeeTypeNoTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.CalcServiceFee(serviceFeeRate, taxFeeType))).Add(
			decimal.NewFromFloat(model.CalcTaxFee(price, taxFeeType))).Add(
			decimal.NewFromFloat(model.CalcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
		return totalPrice.InexactFloat64()
	}
	// 当商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
	if taxFeeType == constant.TaxFeeTypeTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.CalcServiceFee(serviceFeeRate, taxFeeType))).Add(
			decimal.NewFromFloat(model.CalcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
		return totalPrice.InexactFloat64()
	}
	// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
	totalPrice := decimal.NewFromFloat(price).Add(
		decimal.NewFromFloat(model.CalcServiceFee(serviceFeeRate, taxFeeType)))
	return totalPrice.InexactFloat64()
}

func (model *SaleOrderProduct) ChangeProductPrice(price float64) {
	model.ChangePriceTime = time.Now().Unix()
	model.SalePrice = price
}

type SaleOrderProductCalc struct {
	SaucePrice        float64 `json:"sauce_price"`
	ProductPrice      float64 `json:"product_price"`
	SalePrice         float64 `json:"sale_price"`
	Price             float64 `json:"price"`
	MemberDiscountFee float64 `json:"member_discount_fee"`
	CustomDiscountFee float64 `json:"custom_discount_fee"`
	DiscountFee       float64 `json:"discount_fee"`
	TaxFee            float64 `json:"tax_fee"`
	ServiceFee        float64 `json:"service_fee"`
	ServiceTaxFee     float64 `json:"service_tax_fee"`
	TotalPrice        float64 `json:"total_price"`
}

// 获取价格变动前的信息
func (model *SaleOrderProduct) BeforeCalc() SaleOrderProductCalc {
	calc := SaleOrderProductCalc{}
	calc.SaucePrice = model.SaucePrice
	calc.ProductPrice = model.ProductPrice
	calc.SalePrice = model.SalePrice
	calc.Price = model.Price
	calc.MemberDiscountFee = model.MemberDiscountFee
	calc.CustomDiscountFee = model.CustomDiscountFee
	calc.DiscountFee = model.DiscountFee
	calc.TaxFee = model.TaxFee
	calc.ServiceFee = model.ServiceFee
	calc.ServiceTaxFee = model.ServiceTaxFee
	calc.TotalPrice = model.TotalPrice
	return calc
}

// 计算销售订单商品的所有计算值字段
func (model *SaleOrderProduct) CalcSaleOrderProduct(serviceFeeRate float64, taxFeeType int, serviceFeeType int) SaleOrderProductCalc {
	calc := SaleOrderProductCalc{}
	// 开始计算
	calc.SaucePrice = model.CalcSaucePrice()
	model.SaucePrice = calc.SaucePrice
	calc.ProductPrice = model.CalcProductPrice()
	model.ProductPrice = calc.ProductPrice
	calc.SalePrice = model.CalcSalePrice()
	model.SalePrice = calc.SalePrice
	calc.Price = model.CalcPrice()
	model.Price = calc.Price
	calc.MemberDiscountFee = model.CalcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.CustomDiscountFee = model.CalcCustomDiscountFee()
	model.CustomDiscountFee = calc.CustomDiscountFee
	calc.DiscountFee = model.CalcDiscountFee()
	model.DiscountFee = calc.DiscountFee
	calc.TaxFee = model.CalcTaxFee(model.Price, taxFeeType)
	model.TaxFee = calc.TaxFee
	calc.ServiceFee = model.CalcServiceFee(serviceFeeRate, taxFeeType)
	model.ServiceFee = calc.ServiceFee
	calc.ServiceTaxFee = model.CalcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)
	model.ServiceTaxFee = calc.ServiceTaxFee
	calc.TotalPrice = model.CalcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType)
	model.TotalPrice = calc.TotalPrice
	return calc
}

// 获取商品销售价(折前价)
func (model *SaleOrderProduct) GetSalePrice() float64 {
	// 销售价*数量
	salePrice := decimal.NewFromFloat(model.SalePrice).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return salePrice
}

// 获取最终价格（折后价）
func (model *SaleOrderProduct) GetPrice() float64 {
	// 最终价格*数量
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

func (model *SaleOrderProduct) IsMustProduct() bool {
	return model.IsRequire == 1
}

func (model *SaleOrderProduct) IsGiftProduct() bool {
	return model.GiftTime > 0
}

func (model *SaleOrderProduct) IsCancelProduct() bool {
	return model.CancelTime > 0
}

// 判断商品是否有打折
func (model *SaleOrderProduct) IsDiscount() bool {
	return model.Price != model.SalePrice // 折前价格不等于折后价格时，说明有折扣
}

// 判断是哪个业务状态
func (model *SaleOrderProduct) StatusValue() int {
	return int(model.Status)
}

// 获取该订单商品的材料组成及用量。
// 如一个珍珠奶茶加料珍珠，则计算成分珍珠、奶、茶等各个原材料等用量
func (model *SaleOrderProduct) GetMaterialBom() []*ProductionOrderMaterial {
	return nil // todo
}

func (model *SaleOrderProduct) AttributeName(language string) *dto.LocaleResponse {
	var flavorName dto.LocaleResponse
	var sauceNames []dto.LocaleResponse
	var attributeNames []dto.LocaleResponse
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if saleOrderProductBom.IsFlavor() {
			flavorName = saleOrderProductBom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
		} else {
			sauceName := saleOrderProductBom.ProductBom.ProductSauce.MultiLanguageName.GetNames()
			sauceNames = append(sauceNames, sauceName)
		}
	}
	// 获取商品属性
	for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
		attributeName := saleOrderProductAttribute.ProductAttribute.MultiLanguageName.GetNames()
		attributeNames = append(attributeNames, attributeName)
	}
	// 根据规格生成字符串。`(规格；属性；小料)`
	nameList := make([]dto.LocaleResponse, 0)
	nameList = append(nameList, flavorName)
	if len(attributeNames) > 0 {
		nameList = append(nameList, attributeNames...)
	}
	if len(sauceNames) > 0 {
		nameList = append(nameList, sauceNames...)
	}
	if len(nameList) == 0 {
		return &dto.LocaleResponse{}
	}

	attributeResultNames := dto.LocaleResponse{}
	for index, name := range nameList {
		attributeResultNames.ZH += name.ZH
		attributeResultNames.TH += name.TH
		attributeResultNames.EN += name.EN
		attributeResultNames.ZHTW += name.ZHTW
		attributeResultNames.JA += name.JA
		attributeResultNames.KO += name.KO
		attributeResultNames.MY += name.MY
		attributeResultNames.TR += name.TR
		if index != len(nameList)-1 {
			attributeResultNames.ZH += ";"
			attributeResultNames.TH += ";"
			attributeResultNames.EN += ";"
			attributeResultNames.ZHTW += ";"
			attributeResultNames.JA += ";"
			attributeResultNames.KO += ";"
			attributeResultNames.MY += ";"
			attributeResultNames.TR += ";"
		}
	}

	return &attributeResultNames
}

// SaleOrderProductAttribute 销售订单产品属性 `ttpos_sale_order_product_attribute`
type SaleOrderProductAttribute struct {
	BaseModel
	Name                 string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品属性名称,不随后台更新'"`
	SaleOrderUuid        uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductAttributeUuid uint64 `gorm:"column:product_attribute_uuid;not null;default:0;comment:'商品属性ID'"`

	ProductAttribute ProductAttribute `gorm:"foreignKey:ProductAttributeUuid;references:uuid"`
}

// GetAttributeNames 获取属性名称字符串
func (model *SaleOrderProduct) GetAttributeNames() string {
	attributeNames := []string{}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom == 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	for _, attribute := range model.SaleOrderProductAttributes {
		attributeNames = append(attributeNames, attribute.Name)
	}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom != 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	return strings.Join(attributeNames, "; ")
}

// SaleOrderProductBom 销售订单产品原料 `ttpos_sale_order_product_bom`
type SaleOrderProductBom struct {
	BaseModel
	Name                 string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:'规格或小料规格名称,不随后台更新'"`
	Price                float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:'单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动'"`
	IsFlavorBom          uint    `gorm:"column:is_flavor_bom;type:tinyint(1);not null;default:0;comment:'是否为规格商品BOM, 0-否,加料商品 1-是,规格商品'"`
	SaleOrderUuid        uint64  `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64  `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductBomUuid       uint64  `gorm:"column:product_bom_uuid;not null;default:0;comment:'商品BOM ID'"`

	ProductBom       ProductBom       `gorm:"foreignKey:product_bom_uuid;references:uuid"`
	SaleOrderProduct SaleOrderProduct `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
}

func (model *SaleOrderProductBom) IsFlavor() bool {
	return model.IsFlavorBom == 1
}

// SaleBillSetting 销售账单设置 ttpos_sale_bill_setting
type SaleBillSetting struct {
	// 基础字段
	BaseModel
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 费用计算设置
	ServiceFeeType  uint    `gorm:"column:service_fee_type;type:tinyint(1);default:0;comment:服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费" json:"service_fee_type"`
	ServiceFeeValue float64 `gorm:"column:service_fee_value;type:decimal(12,2);default:0;comment:服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例" json:"service_fee_value"`
	TaxFeeType      uint    `gorm:"column:tax_fee_type;type:tinyint(1);default:0;comment:税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税" json:"tax_fee_type"`

	// 优惠和抹零设置
	DiscountType uint `gorm:"column:discount_type;type:tinyint(1);default:0;comment:打折类型, 0-百分比打折% 1-百分比直接减免% off" json:"discount_type"`
	Zero         uint `gorm:"column:zero;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero"`
	ZeroCheckout uint `gorm:"column:zero_checkout;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout"`

	// 统计设置
	IsStatGift uint `gorm:"column:is_stat_gift;type:tinyint(1);default:0;comment:是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣" json:"is_stat_gift"`
	IsStatFree uint `gorm:"column:is_stat_free;type:tinyint(1);default:0;comment:是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费" json:"is_stat_free"`
}

// 获取销售账单商品税费类型
func (model *SaleBillSetting) GetTaxFeeType() int {
	switch model.TaxFeeType {
	// 关闭税费、不收取税费
	case constant.TaxFeeTypeNone:
		return constant.TaxFeeTypeNone
	//	商品已含税
	case constant.TaxFeeTypeTax:
		return constant.TaxFeeTypeTax
	// 商品未含税
	case constant.TaxFeeTypeNoTax:
		return constant.TaxFeeTypeNoTax
	// 默认为关闭税费
	default:
		return constant.TaxFeeTypeNone
	}
}

// 获取销售账单商品税费类型
func (model *SaleBillSetting) GetServiceFeeType() int {
	switch model.ServiceFeeType {
	// 不收取服务费
	case constant.SaleBillSettingServiceFeeTypeNone:
		return constant.SaleBillSettingServiceFeeTypeNone
	//	固定服务费
	case constant.SaleBillSettingServiceFeeTypeFixed:
		return constant.SaleBillSettingServiceFeeTypeFixed
	// 按比例-不收取税费
	case constant.SaleBillSettingServiceFeeTypePercent:
		return constant.SaleBillSettingServiceFeeTypePercent
	// 按比例-收取税费
	case constant.SaleBillSettingServiceFeeTypePercentTax:
		return constant.SaleBillSettingServiceFeeTypePercentTax
	// 默认为不收取服务费
	default:
		return constant.TaxFeeTypeNone
	}
}

// 获取该销售账单的服务费比例。
// 当服务费关闭时，服务费比例为0，即不收取销售订单商品的服务费
// 当服务费按固定金额收取时，服务费比例为0，即不收取销售订单商品的服务费，只在销售订单加上固定金额的服务费
// 当服务费按比例收取时，服务费比例为ServiceFeeValue字段记录的值
func (model *SaleBillSetting) GetServiceFeeRate() float64 {
	switch model.ServiceFeeType {
	// 不收取服务费时，服务费比率为0
	case constant.SaleBillSettingServiceFeeTypeNone:
		return 0
	// 收固定服务费时，服务费比率为0
	case constant.SaleBillSettingServiceFeeTypeFixed:
		return 0
	//	按比例收时，服务费比率为ServiceFeeValue
	case constant.SaleBillSettingServiceFeeTypePercent:
		return model.ServiceFeeValue
	//	按比例收时，服务费比率为ServiceFeeValue
	case constant.SaleBillSettingServiceFeeTypePercentTax:
		return model.ServiceFeeValue
	//	服务费比率为0
	default:
		return 0

	}
}

// SaleOrderBuffetCustomerType 销售订单自助餐顾客类型
type SaleOrderBuffetCustomerType struct {
	// 主键字段
	BaseModel

	Name string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'名称'"`
	// 关联ID字段
	SaleOrderUuid                      uint64 `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	BuffetPackageUuid                  uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
	BuffetPackageMultiLanguageNameUuid uint64 `gorm:"column:buffet_package_multi_language_name_uuid;comment:自助餐套餐多语言ID" json:"buffet_package_multi_language_name_uuid"`
	BuffetCustomerTypePriceUuid        uint64 `gorm:"column:buffet_customer_type_price_uuid;comment:顾客类型定价ID" json:"buffet_customer_type_price_uuid"`

	// 数值字段
	Num                uint    `gorm:"column:num;type:int(11);default:0;comment:人数" json:"num"`
	CustomerPrice      float64 `gorm:"column:customer_price;type:decimal(12,2);not null;default:0;comment:原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变" json:"customer_price"`
	Price              float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:价格（折后价），只进行自定义打折，不进行会员打折" json:"price"`
	CustomDiscountRate float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:1;comment:自定义折扣率(0-100%)" json:"custom_discount_rate"`
	CustomDiscountFee  float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0;comment:自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*(1-自定义折扣率)" json:"custom_discount_fee"`
	TaxRate            float64 `gorm:"column:tax_rate;type:decimal(10,2);not null;default:0;comment:税率,单位%.加购时记录税率,结账时再重新核算" json:"tax_rate"`
	ServiceTaxFee      float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0;comment:服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率" json:"service_tax_fee"`
	TaxFee             float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0;comment:自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率" json:"tax_fee"`
	ServiceFee         float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0;comment:服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例" json:"service_fee"`
	Amount             float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:应收金额(单人)。自助餐顾客类型已含税时，应收金额(单人)=(自助餐顾客类型原价-自助餐顾客类型税费)+服务费+自助餐顾客类型税费；自助餐顾客类型未含税时，应收金额(单人)=自助餐顾客类型原价+服务费+自助餐顾客类型税费" json:"amount"`

	// 关联字段
	BuffetPackageMultiLanguageName MultiLanguageName       `gorm:"foreignKey:BuffetPackageMultiLanguageNameUuid;references:uuid"`
	BuffetCustomerTypePrice        BuffetCustomerTypePrice `gorm:"foreignKey:BuffetCustomerTypePriceUuid;references:uuid"` // 用于关联后台设置的顾客类型定价。在结账时，判断价格是否改变
}

// 获取顾客原价
func (model *SaleOrderBuffetCustomerType) GetOriginPrice() float64 {
	price := decimal.NewFromFloat(model.CustomerPrice).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// 获取顾客折后价
func (model *SaleOrderBuffetCustomerType) GetDiscountPrice() float64 {
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// SaleOrderBuffetDelayProduct 销售订单加钟价格商品表 `ttpos_sale_order_buffet_delay_product`
type SaleOrderBuffetDelayProduct struct {
	BaseModel

	// 数值字段
	Name string `gorm:"default:'';column:name;comment:'自助餐加钟商品名称，下单时固定不受后台改变'"`
	// 废弃，直接使用桌台人数即可
	//Num   uint    `gorm:"default:0;column:num;comment:'数量'"`
	Price float64 `gorm:"default:0;column:price;comment:'价格（单价）,下单时固定不受后台改变，结账时再检查是否改变'"`

	// 关联ID字段
	SaleOrderUuid   uint64 `gorm:"default:0;column:sale_order_uuid;comment:'销售订单ID'"`
	BuffetDelayUuid uint64 `gorm:"default:0;column:buffet_delay_uuid;comment:'自助餐加钟价格ID'"`
}

// 获取商品的价格。单价*数量
func (model *SaleOrderBuffetDelayProduct) GetPrice(num uint) float64 {
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromInt(int64(num))).Round(2).InexactFloat64()
	return price
}

// SaleBillOperationRecord 桌台账单操作记录
type SaleBillOperationRecord struct {
	BaseModel
	// 基本信息
	Data   string `gorm:"column:data;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"data"`
	Source string `gorm:"column:source;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"source"`
	Action string `gorm:"column:action;comment:操作行为" json:"action"`
	Remark string `gorm:"column:remark;comment:备注" json:"remark"`
	// 关联ID字段
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作员ID" json:"operator_uuid"`
}

// 销售订单优惠策略表
type SaleOrderDiscountStrategy struct {
	BaseModel
	// 基本信息
	Type      uint    `gorm:"column:type;type:tinyint(2);not null;default:0;comment:优惠策略类型,0-整单折扣、1-会员折扣" json:"type"`
	Name      string  `gorm:"column:name;type:varchar(50);not null;default:'';comment:优惠策略名称" json:"name"`
	Value     float64 `gorm:"column:value;type:decimal(12,2);not null;default:0.00;comment:优惠策略值" json:"value"`
	JsonField string  `gorm:"column:json_field;type:text;default:null;comment:JSON字段" json:"json_field"`
	// 关联ID字段
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
}

// GenerateProductSign 生成商品包签名. 相同的商品，商品签名相同,用于取消拆单时合并商品。
// 格式：物料,物料,物料-属性,属性,属性-备注内容-送厨批次-改价时间-赠菜时间-退菜时间
// 更新签名的场景：
// 1 改价销售订单商品价格后要重新生成签名
// 2 修改备注
// 3 送厨
// 4 赠菜
// 5 退菜
func (model *SaleOrderProduct) GenerateProductSign() string {
	bomIdList := make([]string, 0)
	attributeIdList := make([]string, 0)

	// 物料ID列表
	for _, bom := range model.SaleOrderProductBoms {
		bomIdList = append(bomIdList, strconv.FormatUint(bom.ProductBomUuid, 10))
	}
	// 属性ID列表
	for _, attributeGroup := range model.SaleOrderProductAttributes {
		attributeIdList = append(attributeIdList, strconv.FormatUint(attributeGroup.ProductAttributeUuid, 10))
	}
	// 物料ID列表和属性ID列表排序
	sort.Slice(bomIdList, func(i, j int) bool {
		return bomIdList[i] < bomIdList[j]
	})
	sort.Slice(attributeIdList, func(i, j int) bool {
		return attributeIdList[i] < attributeIdList[j]
	})
	// 物料ID列表和属性ID列表拼接。格式：物料,物料,物料-属性,属性,属性-备注内容-送厨批次-改价时间-赠菜时间-退菜时间
	bomIdListStr := strings.Join(bomIdList, ",")
	attributeIdListStr := strings.Join(attributeIdList, ",")
	return fmt.Sprintf("%s-%s-%s-%d-%d-%d-%d",
		bomIdListStr,
		attributeIdListStr,
		model.Remark,
		model.ProductionOrderUuid,
		model.ChangePriceTime,
		model.GiftTime,
		model.CancelTime)
}
