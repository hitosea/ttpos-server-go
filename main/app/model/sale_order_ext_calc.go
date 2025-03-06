package model

import (
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// 计算销售订单原价金额。
// 商品未含税时，销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（包括消费税税费和服务费税费）
// 商品已含税时，销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
func (model *SaleOrder) CalcOrderOriginAmount(serviceFeeRate float64, serviceFeeValue float64, taxFeeType int) float64 {
	// 商品未含税时,销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（包括消费税税费和服务费税费）
	if taxFeeType == constant.TaxFeeTypeNoTax {
		// 原服务费
		originService := model.calcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
		// 原服务费税费
		serviceTaxFee := model.calcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType)
		// 原商品消费税税费
		originProductTaxFee := model.calcOriginProductTaxFee(taxFeeType)

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
		originService := model.calcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
		// 服务费税费
		serviceTaxFee := decimal.NewFromFloat(model.calcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType))
		//销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
		amount := decimal.NewFromFloat(model.ProductOriginalAmount).Add(
			decimal.NewFromFloat(originService)).Add(
			serviceTaxFee)
		return amount.InexactFloat64()
	}

	// 默认按商品已含税处理。
	// 服务费
	originService := model.calcOrderOriginServiceFee(serviceFeeRate, serviceFeeValue, taxFeeType)
	// 服务费税费
	serviceTaxFee := decimal.NewFromFloat(model.calcOrderOriginServiceTaxFee(serviceFeeRate, taxFeeType))
	//销售订单原价金额=销售订单商品总价（折前价）+ 服务费 + 消费税（只含服务费税费）
	amount := decimal.NewFromFloat(model.ProductOriginalAmount).Add(
		decimal.NewFromFloat(originService)).Add(
		serviceTaxFee)
	return amount.InexactFloat64()
}

// 计算销售订单已经支付的付款单手续费
func (model *SaleOrder) CalcCommissionFee() float64 {
	commissionFee := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		commissionFee = commissionFee.Add(decimal.NewFromFloat(paymentOrder.PaymentCommissionFee))
	}
	return commissionFee.InexactFloat64()
}

// 计算销售订单未付款的金额。
// 支付没有手续费时，销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
// 支付有手续费时，销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和
func (model *SaleOrder) CalcUnPayAmount(hasCommission bool) float64 {
	if hasCommission {
		// 销售订单各个支付单的支付金额之和
		payOrderAmount := model.calcPayOrderAmount()
		// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和
		unPayAmount := decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(payOrderAmount))
		return unPayAmount.InexactFloat64()
	}
	// 没有手续费时
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	// 销售订单各个支付单的支付金额之和
	payOrderAmount := model.calcPayOrderAmount()
	zeroFee := model.CalcCheckOutZeroFee()
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	unPayAmount := decimal.NewFromFloat(model.Amount).Sub(
		decimal.NewFromFloat(payOrderAmount)).Sub(
		decimal.NewFromFloat(zeroFee))
	return unPayAmount.InexactFloat64()

}

func (model *SaleOrder) SetCheckOutZeroFee() {
	model.ZeroCheckoutFee = model.CalcCheckOutZeroFee()
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

func (model *SaleOrder) CalcSaleOrder(setting SaleBillSetting) *Calc {
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	return model.calcSaleOrder(serviceFeeType, setting.ServiceFeeValue, taxFeeType)
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
func (model *SaleOrder) calcSaleOrder(serviceFeeType int, serviceFeeValue float64, taxFeeType int) *Calc {
	calc := Calc{}
	calc.ProductOriginalAmount = model.calcProductOriginalAmount()
	model.ProductOriginalAmount = calc.ProductOriginalAmount
	calc.ProductAmount = model.calcProductAmount()
	model.ProductAmount = calc.ProductAmount
	calc.ServiceFee = model.calcServiceFee(serviceFeeType, serviceFeeValue)
	model.ServiceFee = calc.ServiceFee
	calc.TaxFee = model.calcTaxFee()
	model.TaxFee = calc.TaxFee
	calc.MemberDiscountFee = model.calcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.Amount = model.calcAmount(taxFeeType)
	model.Amount = calc.Amount
	calc.ZeroFee = model.calcZeroFee()
	model.ZeroFee = calc.ZeroFee
	calc.CustomDiscountFee = model.calcCustomDiscountFee()
	model.CustomDiscountFee = calc.CustomDiscountFee
	// 再重新计算一次应付金额
	calc.Amount = model.calcAmountZero()
	model.Amount = calc.Amount
	return &calc
}

// 计算销售订单原服务费金额。销售订单原服务费金额= 销售订单商品的原服务费之和。
// 不受服务费、按固定服务费费收时，销售订单商品的原服务费=0 或 销售订单商品的原服务费=固定费用
// 按比例收服务费时，商品未含税时，销售订单商品的原服务费=销售订单商品的SalePrice * 服务费费率
// 按比例收服务费时，商品已含税时，销售订单商品的原服务费=销售订单商品的未含税价格 * 服务费费率 = （销售订单商品的SalePrice - 销售订单商品的商品消费税税费） * 服务费费率 = （销售订单商品的SalePrice - （销售订单商品的商品消费税税费）） * 服务费费率
func (model *SaleOrder) calcOrderOriginServiceFee(serviceFeeRate float64, serviceFeeValue float64, taxFeeType int) float64 {
	originServiceFee := decimal.NewFromFloat(0)
	// 不是按比例收取服务费时
	if serviceFeeRate <= 0 {
		return serviceFeeValue
	} else {
		// 如果按比例收取
		for _, saleOrderProduct := range model.SaleOrderProducts {
			if saleOrderProduct.SaleOrderUuid != model.Uuid {
				continue
			}
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
				continue
			}
			serviceFee := saleOrderProduct.calcOriginServiceFee(serviceFeeRate, taxFeeType)
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
func (model *SaleOrder) calcOrderOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int) float64 {
	originServiceFee := decimal.NewFromFloat(0)
	// 不是按比例收取服务费时
	if serviceFeeRate <= 0 {
		return 0
	} else {
		// 如果按比例收取
		for _, saleOrderProduct := range model.SaleOrderProducts {
			if saleOrderProduct.SaleOrderUuid != model.Uuid {
				continue
			}
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
				continue
			}
			serviceFee := saleOrderProduct.calcOriginServiceFee(serviceFeeRate, taxFeeType)
			// 商品的原服务费税费= 原服务费 * 服务费税率
			originServiceTaxFee := decimal.NewFromFloat(serviceFee).Mul(decimal.NewFromFloat(saleOrderProduct.TaxRate))
			// 累加各个订单商品的服务费税费
			originServiceFee = originServiceFee.Add(originServiceTaxFee)
		}
		return originServiceFee.InexactFloat64()
	}
}

// 计算销售订单已经支付的金额。 销售订单已经支付的金额= 销售订单的所有支付单的支付金额之和
func (model *SaleOrder) calcPayOrderAmount() float64 {
	payAmount := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		payAmount = payAmount.Add(decimal.NewFromFloat(paymentOrder.PaymentAmount))
	}
	return payAmount.InexactFloat64()
}

// 计算销售订单的最终应收金额。
// 最终应收=应收金额+支付手续费= 应收金额 +（各个支付订单的手续费之和+当前支付方式的手续费）
func (model *SaleOrder) calcFinallyAmount() (float64, bool) {
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

// 计算订单商品SalePrice之和。等于所有已接单商品SalePrice之和
func (model *SaleOrder) calcSumOrderProductSalePrice() float64 {
	sumSalePrice := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
		// 销售订单商品已接单且未删除商品
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？要累计上
			//退菜？不累计
			if orderProduct.IsCancelBool() {
				continue
			}
			// SalePrice * Num
			saleProductSalePrice := decimal.NewFromFloat(orderProduct.SalePrice).Mul(decimal.NewFromUint64(uint64(orderProduct.Num)))
			sumSalePrice = sumSalePrice.Add(saleProductSalePrice)
		}
	}
	return sumSalePrice.InexactFloat64()
}

// 计算自助餐顾客价格之和。等于所有自助餐顾客价格之和
func (model *SaleOrder) calcSumOrderProductCustomerPrice() float64 {
	sumCustomerPrice := decimal.NewFromFloat(0)
	for _, orderBuffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		// 已经移动到其他订单的商品不计
		if orderBuffetCustomer.SaleOrderUuid != model.Uuid {
			continue
		}
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		// 自助餐顾客价格之和
		customerPrice := decimal.NewFromFloat(orderBuffetCustomer.GetOriginPrice())
		sumCustomerPrice = sumCustomerPrice.Add(customerPrice)
	}
	return sumCustomerPrice.InexactFloat64()
}

// 计算订单商品Price之和。等于所有已接单商品Price之和
func (model *SaleOrder) calcSumOrderProductPrice() float64 {
	sumPrice := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
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
	return sumPrice.Round(2).InexactFloat64()
}

// 计算自助餐顾客价格之和。等于所有自助餐顾客价格之和
func (model *SaleOrder) calcSumOrderProductCustomerDiscountPrice() float64 {
	sumCustomerPrice := decimal.NewFromFloat(0)
	for _, orderBuffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		// 已经移动到其他订单的商品不计
		if orderBuffetCustomer.SaleOrderUuid != model.Uuid {
			continue
		}
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		// 自助餐顾客价格之和
		discountPrice := decimal.NewFromFloat(orderBuffetCustomer.GetDiscountPrice())
		sumCustomerPrice = sumCustomerPrice.Add(discountPrice)
	}
	return sumCustomerPrice.InexactFloat64()
}

// 计算订单商品金额（折前价）。订单商品金额（折前价）= 订单商品SalePrice之和 + 自助餐顾客价格CustomerPrice之和 + 自助餐加钟商品价格Price之和
func (model *SaleOrder) calcProductOriginalAmount() float64 {
	sumSaleOrderProduct := model.calcSumOrderProductSalePrice()

	return decimal.NewFromFloat(sumSaleOrderProduct).Add(
		decimal.NewFromFloat(model.calcSumOrderProductCustomerPrice())).InexactFloat64() // todo  + 自助餐加钟商品价格之和
}

// 计算订单商品金额（折后价）。订单商品金额（折后价）= 订单商品Price之和 + 自助餐顾客价格Price之和 + 自助餐加钟商品价格Price
func (model *SaleOrder) calcProductAmount() float64 {
	// 订单商品Price之和
	sumOrderProductPrice := model.calcSumOrderProductPrice()
	// 自助餐顾客价格Price之和
	sumCustomerPrice := model.calcSumOrderProductCustomerDiscountPrice()
	return decimal.NewFromFloat(sumOrderProductPrice).Add(
		decimal.NewFromFloat(sumCustomerPrice)).InexactFloat64() // todo + 自助餐加钟商品价格之和
}

// 计算订单产生的税费。订单税费=订单商品TaxFee之和 + 订单商品ServiceTaxFee之和
func (model *SaleOrder) calcTaxFee() float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
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
func (model *SaleOrder) calcOriginProductTaxFee(taxFeeType int) float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			taxFee = taxFee.Add(
				decimal.NewFromFloat(orderProduct.calcOriginTaxFee(taxFeeType)))
		}
	}
	return taxFee.InexactFloat64()
}

// 计算销售订单的自定义优惠折扣金额。订单自定义优惠金额=销售订单商品自定义优惠金额之和 + 自助餐顾客自定义优惠金额之和
// todo 考虑每个价格循环一次商品列表计算带来的性能损耗与计算业务更清晰抉择哪个
func (model *SaleOrder) calcCustomDiscountFee() float64 {
	customDiscountFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
		if orderProduct.IsAcceptOrderBool() && !orderProduct.IsDelete() {
			// 赠菜？免费了不计入
			// 退菜？退了不计入
			if orderProduct.IsGiftBool() || orderProduct.IsCancelBool() {
				continue
			}
			// 优惠折扣金额 = 销售订单商品自定义折扣优惠金额之和 + 自助餐顾客自定义折扣优惠金额之和 + 订单抹零金额
			customDiscountFee = customDiscountFee.Add(
				decimal.NewFromFloat(orderProduct.CustomDiscountFee))
		}
	}
	// 优惠折扣金额 = 销售订单商品自定义折扣优惠金额之和 + 自助餐顾客自定义折扣优惠金额之和 + 订单抹零金额
	customDiscountFee = customDiscountFee.Add(
		decimal.NewFromFloat(model.calcZeroFee()))
	// todo  + 自助餐顾客自定义优惠金额之和
	return customDiscountFee.InexactFloat64()
}

// 计算销售订单会员折扣金额。销售订单会员折扣金额=订单商品会员折扣金额之和
func (model *SaleOrder) calcMemberDiscountFee() float64 {
	memberDiscountFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
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
	return memberDiscountFee.Round(2).InexactFloat64()
}

// 计算销售订单服务费消费税金额。销售订单服务费消费税金额=订单商品服务费消费税金额之和
func (model *SaleOrder) calcServiceTaxFee() float64 {
	serviceTaxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range model.SaleOrderProducts {
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}
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
func (model *SaleOrder) calcAmount(taxFeeType int) float64 {
	amount := decimal.NewFromFloat(0)
	// 商品已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		serviceTaxFee := model.calcServiceTaxFee()
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

// 计算订单应付金额抹零后的金额。订单应付金额抹零后的金额=订单应付金额-订单抹零金额
func (model *SaleOrder) calcAmountZero() float64 {
	amount := decimal.NewFromFloat(model.Amount)
	amount = amount.Sub(decimal.NewFromFloat(model.ZeroFee))
	return amount.InexactFloat64()
}

// 计算销售订单的订单优惠折扣抹零金额。根据订单设置的优惠折扣抹零规则金额计算
func (model *SaleOrder) calcZeroFee() float64 {
	amount := model.Amount
	switch model.ZeroRule {
	// 实款实收
	case constant.DiscountZeroRuleNone:
		return 0
	// 抹分
	case constant.DiscountZeroRulePercent:
		// 抹分后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(1)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 抹角
	case constant.DiscountZeroRuleFixed:
		// 抹角后的订单金额
		discountAmount := decimal.NewFromFloat(amount).Truncate(0)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 四舍五入保留一位小数
	case constant.DiscountZeroRuleRound:
		discountAmount := decimal.NewFromFloat(amount).Round(1)
		// 抹零金额 = 原订单金额-抹零后的订单金额
		zeroFee := decimal.NewFromFloat(amount).Sub(discountAmount)
		return zeroFee.InexactFloat64()
	// 四舍五入保留整数
	case constant.DiscountZeroRuleInteger:
		discountAmount := decimal.NewFromFloat(amount).Round(0)
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
func (model *SaleOrder) calcServiceFee(serviceFeeType int, serviceFeeValue float64) float64 {
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
			// 已经移动到其他订单的商品不计
			if orderProduct.SaleOrderUuid != model.Uuid {
				continue
			}
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
