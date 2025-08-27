package model

import (
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// 计算销售订单的金额
func (model *SaleOrder) CalcSaleOrder(setting SaleBillSetting, options ...func(option *CalcOption)) *Calc {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	if option.H5OrderUuid == 0 {
		return model.calcSaleOrder(serviceFeeType, setting.ServiceFeeValue, taxFeeType, 0, option.IsCanceled)
	} else {
		return model.calcSaleOrder(serviceFeeType, setting.ServiceFeeValue, taxFeeType, option.H5OrderUuid, option.IsCanceled)
	}
}

// 计算已送厨商品的订单金额
func (model *SaleOrder) CalcCookingSaleOrder(setting SaleBillSetting) *Calc {
	list := make([]*SaleOrderProduct, 0)
	products := model.GetCookingOrderProductList()
	for _, product := range products {
		if product.IsDelete() {
			continue
		}
		if product.IsPackageSubProduct() {
			continue
		}
		list = append(list, product)
	}
	return model.CalcSaleOrderByProductList(list, setting)
}

// 计算指定商品列表的订单金额
func (model *SaleOrder) CalcSaleOrderByProductList(products []*SaleOrderProduct, setting SaleBillSetting) *Calc {
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	return model.calcCookingSaleOrder(products, serviceFeeType, setting.ServiceFeeValue, taxFeeType)
}

// 计算已送厨商品和已下单商品的订单金额
func (model *SaleOrder) CalcCookingAndOrderSaleOrder(setting SaleBillSetting) *Calc {
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	products := model.GetCookingOrderProductList()
	products = append(products, model.GetH5OrderProductList()...)
	return model.calcCookingSaleOrder(products, serviceFeeType, setting.ServiceFeeValue, taxFeeType)
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
	amount := model.GetAmountValue()
	if hasCommission {
		// 销售订单各个支付单的支付金额之和
		payOrderAmount := model.calcPayOrderAmount()
		// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和
		unPayAmount := decimal.NewFromFloat(amount).Sub(decimal.NewFromFloat(payOrderAmount))
		return unPayAmount.InexactFloat64()
	}
	// 没有手续费时
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	// 销售订单各个支付单的支付金额之和
	payOrderAmount := model.calcPayOrderAmount()
	zeroFee := model.CalcCheckOutZeroFee()
	// 销售订单未付款的金额 = 应收金额-销售订单各个支付单的支付金额之和-结账抹零金额
	unPayAmount := decimal.NewFromFloat(amount).Sub(
		decimal.NewFromFloat(payOrderAmount)).Sub(
		decimal.NewFromFloat(zeroFee))
	return unPayAmount.InexactFloat64()

}

// 设置销售订单结账抹零金额
func (model *SaleOrder) SetCheckOutZeroFee() {
	model.ZeroCheckoutFee = model.CalcCheckOutZeroFee()
}

// 计算销售订单的结账抹零金额。根据订单设置的结账抹零规则金额计算
func (model *SaleOrder) CalcCheckOutZeroFee() float64 {
	amount := model.GetAmountValue()
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

// 重新计算销售订单的金额
func (model *SaleOrder) calcSaleOrder(serviceFeeType int, serviceFeeValue float64, taxFeeType int, h5OrderUuid uint64, isCanceled bool) *Calc {
	calc := Calc{}
	products := model.GetUnCookingAndCookingOrderProductList(h5OrderUuid, isCanceled)
	calc.ProductOriginalAmount = model.calcProductAmount(products, WithOriginPrice())
	model.ProductOriginalAmount = calc.ProductOriginalAmount
	calc.ProductAmount = model.calcProductAmount(products)
	model.ProductAmount = calc.ProductAmount
	calc.ServiceFee = model.calcServiceFee(products, serviceFeeType, serviceFeeValue)
	model.ServiceFee = calc.ServiceFee
	calc.TaxFee = model.calcTaxFee(products)
	model.TaxFee = calc.TaxFee
	calc.MemberDiscountFee = model.calcMemberDiscountFee(products)
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.Amount = model.calcAmount(products, serviceFeeType, serviceFeeValue, taxFeeType)
	model.Amount = calc.Amount
	calc.OriginAmount = model.calcOriginAmount(products, serviceFeeType, serviceFeeValue, taxFeeType)
	model.OriginAmount = calc.OriginAmount
	calc.ZeroFee = model.calcZeroFee(model.Amount)
	model.ZeroFee = calc.ZeroFee
	calc.CustomDiscountFee = model.calcCustomDiscountFee(products, calc.Amount)
	model.CustomDiscountFee = calc.CustomDiscountFee
	calc.PayPointsAmount = model.CaclPointsExchangeAmount()
	model.PayPointsAmount = calc.PayPointsAmount // 有抵扣积分时，抵扣金额才大于0
	// 再重新计算一次应付金额
	calc.Amount = model.calcAmountZero(calc.Amount, calc.ZeroFee)
	model.Amount = calc.Amount
	return &calc
}

// 重新计算已送厨的销售订单的金额
func (model *SaleOrder) calcCookingSaleOrder(products []*SaleOrderProduct, serviceFeeType int, serviceFeeValue float64, taxFeeType int) *Calc {
	calc := Calc{}

	calc.ProductOriginalAmount = model.calcProductOriginalAmount(products)
	calc.ProductAmount = model.calcProductAmount(products)                                // 已送厨的订单商品金额（折后价）
	calc.ServiceFee = model.calcServiceFee(products, serviceFeeType, serviceFeeValue)     // 已送厨的订单商品服务费
	calc.TaxFee = model.calcTaxFee(products)                                              // 已送厨的订单商品消费税
	calc.MemberDiscountFee = model.calcMemberDiscountFee(products)                        // 已送厨的订单商品会员折扣
	calc.Amount = model.calcAmount(products, serviceFeeType, serviceFeeValue, taxFeeType) // 已送厨的订单应付金额
	calc.ZeroFee = model.calcZeroFee(calc.Amount)                                         // 已送厨的订单优惠折扣抹零金额
	calc.CustomDiscountFee = model.calcCustomDiscountFee(products, calc.Amount)           // 已送厨的订单自定义优惠
	// 再重新计算一次应付金额
	calc.Amount = model.calcAmountZero(calc.Amount, calc.ZeroFee) // 已送厨的订单应付金额抹零后的金额
	return &calc
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
// 最终应收=应收金额+支付手续费= 应收金额 +（各个支付订单的手续费之和+当前支付方式的手续费）- 减去结账抹零金额
func (model *SaleOrder) calcFinallyAmount() (float64, bool) {
	hasCommission := false
	amount := decimal.NewFromFloat(model.GetAmountValue())
	commissionFee := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		commissionFee = commissionFee.Add(decimal.NewFromFloat(paymentOrder.PaymentCommissionFee))
	}
	// 未支付的金额的手续费 = 未支付的金额 * 支付手续费费率
	// 最终应收 = 应收金额+已支付的手续费
	finallyAmount := amount.Add(commissionFee).Sub(decimal.NewFromFloat(model.CalcCheckOutZeroFee()))

	if commissionFee.InexactFloat64() > 0 {
		hasCommission = true
	}
	return finallyAmount.InexactFloat64(), hasCommission
}

// 计算订单商品SalePrice之和。等于所有已接单商品SalePrice之和
func (model *SaleOrder) calcSumOrderProductSalePrice(products []*SaleOrderProduct) float64 {
	sumSalePrice := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		if orderProduct.IsGiftProduct() {
			continue
		}
		// SalePrice * Num
		saleProductSalePrice := decimal.NewFromFloat(orderProduct.SalePrice).Mul(orderProduct.GetNumDecimal())
		sumSalePrice = sumSalePrice.Add(saleProductSalePrice)
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

// 计算订单“商品金额”。订单“商品金额”=等于所有已接单商品Price之和
func (model *SaleOrder) calcSumOrderProductPrice(products []*SaleOrderProduct, options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	sumPrice := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		productPrice := orderProduct.GetFinalSalePrice()
		if option.IsOriginPrice {
			productPrice = orderProduct.GetSalePriceUnit()
		}
		// price * num
		price := decimal.NewFromFloat(productPrice).Mul(orderProduct.GetNumDecimal())
		sumPrice = sumPrice.Add(price)
	}
	return sumPrice.Round(2).InexactFloat64()
}

// 计算自助餐顾客价格之和。等于所有自助餐顾客价格之和。折后价
func (model *SaleOrder) calcSumOrderProductCustomerDiscountPrice(options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
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
		if option.IsOriginPrice {
			discountPrice = decimal.NewFromFloat(orderBuffetCustomer.GetOriginPrice())
		}
		sumCustomerPrice = sumCustomerPrice.Add(discountPrice)
	}
	return sumCustomerPrice.InexactFloat64()
}

// 计算自助餐加钟商品价格之和。等于所有自助餐加钟商品价格之和
func (model *SaleOrder) calcSumOrderProductBuffetDelayPrice() float64 {
	sumBuffetDelayPrice := decimal.NewFromFloat(0)
	for _, orderBuffetDelay := range model.SaleOrderBuffetDelayProducts {
		// 已经移动到其他订单的商品不计
		if orderBuffetDelay.SaleOrderUuid != model.Uuid {
			continue
		}
		if orderBuffetDelay.IsDelete() {
			continue
		}
		// 自助餐加钟商品价格之和
		amount := decimal.NewFromFloat(orderBuffetDelay.GetAmount())
		sumBuffetDelayPrice = sumBuffetDelayPrice.Add(amount)
	}
	return sumBuffetDelayPrice.InexactFloat64()
}

// 计算订单商品金额（折前价）。订单商品金额（折前价）= 订单商品SalePrice之和 + 自助餐顾客价格CustomerPrice之和 + 自助餐加钟商品价格Price之和
func (model *SaleOrder) calcProductOriginalAmount(products []*SaleOrderProduct) float64 {
	// 订单商品SalePrice之和
	sumSaleOrderProduct := model.calcSumOrderProductSalePrice(products)
	// 自助餐顾客价格Price之和
	sumCustomerPrice := model.calcSumOrderProductCustomerPrice()
	// 自助餐加钟商品价格之和
	sumBuffetDelayPrice := model.calcSumOrderProductBuffetDelayPrice()
	return decimal.NewFromFloat(sumSaleOrderProduct).Add(
		decimal.NewFromFloat(sumCustomerPrice)).Add(
		decimal.NewFromFloat(sumBuffetDelayPrice)).InexactFloat64()
}

// 计算订单商品金额（折后价）。订单商品金额（折后价）= 订单商品Price之和 + 自助餐顾客价格Price之和 + 自助餐加钟商品价格Price
func (model *SaleOrder) calcProductAmount(products []*SaleOrderProduct, options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	// 订单商品Price之和。折后价
	sumOrderProductPrice := model.calcSumOrderProductPrice(products)
	if option.IsOriginPrice {
		sumOrderProductPrice = model.calcSumOrderProductPrice(products, WithOriginPrice())
	}
	// 自助餐顾客价格SalePrice之和。折前价
	sumCustomerPrice := model.calcSumOrderProductCustomerDiscountPrice()
	if option.IsOriginPrice {
		sumCustomerPrice = model.calcSumOrderProductCustomerDiscountPrice(WithOriginPrice())
	}
	// 自助餐加钟商品价格之和
	sumBuffetDelayPrice := model.calcSumOrderProductBuffetDelayPrice()
	return decimal.NewFromFloat(sumOrderProductPrice).Add(
		decimal.NewFromFloat(sumCustomerPrice)).Add(
		decimal.NewFromFloat(sumBuffetDelayPrice)).Truncate(2).InexactFloat64()
}

// 计算订单产生的税费。订单税费=订单商品TaxFee之和 + 订单商品ServiceTaxFee之和 + 自助餐顾客税费之和
func (model *SaleOrder) calcTaxFee(products []*SaleOrderProduct) float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		taxFee = taxFee.Add(
			decimal.NewFromFloat(orderProduct.GetTaxFee())).Add(
			decimal.NewFromFloat(orderProduct.GetServiceTaxFee()))
		// fmt.Println("订单税费", orderProduct.GetTaxFee(), orderProduct.GetServiceTaxFee(), taxFee.InexactFloat64())
	}
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		taxFee = taxFee.Add(
			decimal.NewFromFloat(buffetCustomer.GetTaxFee())).Add(
			decimal.NewFromFloat(buffetCustomer.GetServiceTaxFee()))
	}
	return taxFee.Truncate(3).Round(2).InexactFloat64()
}

// 计算订单产生的税费(折前价)。订单税费=订单商品TaxFee之和 + 订单商品ServiceTaxFee之和 + 自助餐顾客税费之和
func (model *SaleOrder) calcOriginTaxFee(products []*SaleOrderProduct, serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	taxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		taxFee = taxFee.Add(
			decimal.NewFromFloat(orderProduct.GetOriginTaxFee(taxFeeType))).Add(
			decimal.NewFromFloat(orderProduct.GetOriginServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
	}
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		taxFee = taxFee.Add(
			decimal.NewFromFloat(buffetCustomer.GetOriginTaxFee(taxFeeType))).Add(
			decimal.NewFromFloat(buffetCustomer.GetOriginServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
	}
	return taxFee.Truncate(3).Round(2).InexactFloat64()
}

// 计算已送厨商品的销售订单的自定义优惠折扣金额。
// 没有整单改价时，订单自定义优惠金额=销售订单商品金额（折前价，含赠菜商品的金额之和）- 销售订单商品金额（折后价）- 会员折扣金额 + 订单抹零金额 + 赠菜商品的金额之和 。 总的优惠金额= 销售订单商品金额（折前价）- 销售订单商品金额（折后价）。这样计算的原因是避免四舍五入引起的误差问题
// 有整单改价时，订单自定义优惠金额=销售订单应收金额 - 整单改价金额
func (model *SaleOrder) calcCustomDiscountFee(products []*SaleOrderProduct, amount float64) float64 {
	// 有整单改价时, 订单自定义优惠金额=销售订单应收金额 - 整单改价金额
	if model.CustomAmount != constant.SaleOrderCustomAmountCancel {
		return decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(model.CustomAmount)).Truncate(3).Round(2).InexactFloat64()
	}
	customDiscountFee := decimal.NewFromFloat(0)
	// 销售订单商品金额（折前价）
	productOriginalAmount := model.ProductOriginalAmount
	// 销售订单商品金额（折后价）
	productAmount := model.ProductAmount
	// 会员折扣金额
	memberDiscountFee := model.MemberDiscountFee
	// 总折扣 = 销售订单商品金额（折前价）- 销售订单商品金额（折后价） = 会员折扣 + 优惠折扣
	// 优惠折扣 = 销售订单商品金额（折前价）- 销售订单商品金额（折后价）- 会员折扣
	discount := productOriginalAmount - productAmount - memberDiscountFee
	// 订单抹零金额
	zeroFee := model.calcZeroFee(amount)
	// 订单自定义优惠金额=销售订单商品金额（折前价，含赠菜商品的金额之和）- 销售订单商品金额（折后价）- 会员折扣金额 + 订单抹零金额
	customDiscountFee = customDiscountFee.Add(
		decimal.NewFromFloat(discount)).Add(
		decimal.NewFromFloat(zeroFee))
	return customDiscountFee.Truncate(2).InexactFloat64()
}

// 计算销售订单会员折扣金额。销售订单会员折扣金额=订单商品会员折扣金额之和
func (model *SaleOrder) calcMemberDiscountFee(products []*SaleOrderProduct) float64 {
	memberDiscountFee := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		memberDiscountFee = memberDiscountFee.Add(
			decimal.NewFromFloat(orderProduct.GetMemberDiscountFee()))
	}
	return memberDiscountFee.Truncate(3).Round(2).InexactFloat64()
}

// 计算已送厨的销售订单服务费消费税金额。已送厨的销售订单服务费消费税金额=已送厨订单商品服务费消费税金额之和
func (model *SaleOrder) calcCookingServiceTaxFee(products []*SaleOrderProduct) float64 {
	serviceTaxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		serviceTaxFee = serviceTaxFee.Add(
			decimal.NewFromFloat(orderProduct.GetServiceTaxFee()))
	}
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		serviceTaxFee = serviceTaxFee.Add(
			decimal.NewFromFloat(buffetCustomer.GetServiceTaxFee()))
	}
	return serviceTaxFee.InexactFloat64()
}

// 计算已送厨的销售订单服务费消费税金额。已送厨的销售订单服务费消费税金额=已送厨订单商品服务费消费税金额之和
func (model *SaleOrder) calcOriginCookingServiceTaxFee(products []*SaleOrderProduct, serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	serviceTaxFee := decimal.NewFromFloat(0)
	for _, orderProduct := range products {
		originServiceTaxFee := orderProduct.GetOriginServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)
		serviceTaxFee = serviceTaxFee.Add(
			decimal.NewFromFloat(originServiceTaxFee))
	}
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		serviceTaxFee = serviceTaxFee.Add(
			decimal.NewFromFloat(buffetCustomer.GetOriginServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
	}
	return serviceTaxFee.InexactFloat64()
}

// 计算已送厨的销售订单应付金额。已送厨的销售订单应付金额=商品金额+服务费+消费税。 给前端显示时，已送厨的销售订单应付金额=商品金额+服务费+消费税-订单抹零金额
// 商品未含税时，已送厨的销售订单应付金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。
// 商品已含税时，已送厨的销售订单应付金额=商品金额（包含商品消费税税费）+服务费+服务费税费。
// 商品关闭税费时，已送厨的销售订单应付金额=商品金额ProductAmount(折后)+服务费
func (model *SaleOrder) calcAmount(products []*SaleOrderProduct, serviceFeeType int, serviceFeeValue float64, taxFeeType int) float64 {
	productAmount := model.calcProductAmount(products)
	serviceFee := model.calcServiceFee(products, serviceFeeType, serviceFeeValue)

	amount := decimal.NewFromFloat(0)
	// 商品已含税时 todo
	if taxFeeType == constant.TaxFeeTypeTax {
		serviceTaxFee := model.calcCookingServiceTaxFee(products) // 订单商品服务费税费之和 + 自助餐顾客服务费税费之和
		//商品金额（包含商品消费税税费）+服务费+服务费税费。
		amount = amount.Add(
			decimal.NewFromFloat(productAmount)).Add(
			decimal.NewFromFloat(serviceFee)).Add(
			decimal.NewFromFloat(serviceTaxFee))
		// fmt.Println("销售订单应收金额 productAmount", productAmount, "serviceFee", serviceFee, "serviceTaxFee", serviceTaxFee, "amount", amount.InexactFloat64())
		return amount.Truncate(3).Round(2).InexactFloat64()
	}
	// 商品未含税时
	if taxFeeType == constant.TaxFeeTypeNoTax {
		taxFee := model.calcTaxFee(products) // 订单商品税费之和 + 自助餐顾客税费之和
		// 销售订单应付金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）
		amount = amount.Add(
			decimal.NewFromFloat(productAmount).Add(
				decimal.NewFromFloat(serviceFee).Add(
					decimal.NewFromFloat(taxFee))))
		return amount.Truncate(3).Round(2).InexactFloat64()
	}
	// 商品关闭税费时
	// 销售订单应付金额=商品金额ProductAmount(折后)+服务费
	result := decimal.NewFromFloat(productAmount).Add(decimal.NewFromFloat(serviceFee))
	return result.Truncate(3).Round(2).InexactFloat64()
}

// 计算已送厨的销售订单原始应收金额。已送厨的销售订单原始应收金额=商品金额+服务费+消费税。 给前端显示时，已送厨的销售订单原始应收金额=商品金额+服务费+消费税-订单抹零金额
// 商品未含税时，已送厨的销售订单原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。
// 商品已含税时，已送厨的销售订单原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。
// 商品关闭税费时，已送厨的销售订单原始应收金额=商品金额ProductAmount(折前价)+服务费
func (model *SaleOrder) calcOriginAmount(products []*SaleOrderProduct, serviceFeeType int, serviceFeeValue float64, taxFeeType int) float64 {
	productAmount := model.calcProductAmount(products, WithOriginPrice())
	serviceFee := model.calcOriginServiceFee(products, serviceFeeType, serviceFeeValue, taxFeeType)

	amount := decimal.NewFromFloat(0)
	// 商品已含税时 todo
	if taxFeeType == constant.TaxFeeTypeTax {
		serviceTaxFee := model.calcOriginCookingServiceTaxFee(products, serviceFeeValue, taxFeeType, serviceFeeType) // 订单商品服务费税费之和 + 自助餐顾客服务费税费之和
		//商品金额（包含商品消费税税费）+服务费+服务费税费。
		amount = amount.Add(
			decimal.NewFromFloat(productAmount)).Add(
			decimal.NewFromFloat(serviceFee)).Add(
			decimal.NewFromFloat(serviceTaxFee))
		// fmt.Println("销售订单原价 productAmount", productAmount, "serviceFee", serviceFee, "serviceTaxFee", serviceTaxFee, "amount", amount.InexactFloat64())
		return amount.Truncate(3).Round(2).InexactFloat64()
	}
	// 商品未含税时
	if taxFeeType == constant.TaxFeeTypeNoTax {
		taxFee := model.calcOriginTaxFee(products, serviceFeeValue, taxFeeType, serviceFeeType) // 订单商品税费之和 + 自助餐顾客税费之和
		// 销售订单应付金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）
		amount = amount.Add(
			decimal.NewFromFloat(productAmount).Add(
				decimal.NewFromFloat(serviceFee).Add(
					decimal.NewFromFloat(taxFee))))
		return amount.Truncate(3).Round(2).InexactFloat64()
	}
	// 商品关闭税费时
	// 销售订单应付金额=商品金额ProductAmount(折前价)+服务费
	result := decimal.NewFromFloat(productAmount).Add(decimal.NewFromFloat(serviceFee))
	return result.Truncate(3).Round(2).InexactFloat64()
}

// 计算已送厨的订单应付金额抹零后的金额。订单应付金额抹零后的金额=订单应付金额-订单抹零金额
func (model *SaleOrder) calcAmountZero(amountNum, zeroFee float64) float64 {
	amount := decimal.NewFromFloat(amountNum)
	amount = amount.Sub(decimal.NewFromFloat(zeroFee))
	return amount.Round(2).InexactFloat64()
}

// CaclMaxPoints 计算最大可抵扣积分
// 1. 当会员积分余额充足时，最大可抵扣积分=订单应收/积分抵扣比例
// 2. 当会员积分余额不足时，最大可抵扣积分=会员积分余额
func (model *SaleOrder) CaclMaxPoints() float64 {
	if model.Member == nil {
		return 0 // 非会员订单，不支持积分抵扣
	}

	memberPoints := model.Member.GetPoints()

	// 1. 计算最大可抵扣积分. 舍去小数点
	maxPoints := model.GetAmount()
	if model.PointsExchangeRate != 0 {
		maxPoints = decimal.NewFromFloat(model.GetAmount()).Div(decimal.NewFromFloat(model.PointsExchangeRate)).Truncate(0).InexactFloat64() // 不能除以0
	}

	// 2. 当会员积分余额充足时，最大可抵扣积分=订单应收/积分抵扣比例
	if memberPoints >= maxPoints {
		return maxPoints
	}

	// 3. 当会员积分余额不足时，最大可抵扣积分=会员积分余额。只返回整数
	if memberPoints > 0 {
		return decimal.NewFromFloat(memberPoints).Truncate(0).InexactFloat64()
	}

	return 0
}

// 计算积分抵扣金额
func (model *SaleOrder) CaclPointsExchangeAmount() float64 {
	if model.Member == nil {
		return 0 // 非会员订单，不支持积分抵扣
	}

	if model.PayPoints == 0 {
		return 0 // 未抵扣积分，则抵扣金额为0
	}

	payPointsAmount := decimal.NewFromFloat(model.PayPoints).Mul(decimal.NewFromFloat(model.PointsExchangeRate)).Round(2).InexactFloat64()

	return payPointsAmount
}

// 计算优惠券抵扣金额
func (model *SaleOrder) CalcCouponExchangeAmount() float64 {
	if len(model.Coupons) == 0 {
		return 0 // 未使用优惠券，则抵扣金额为0
	}
	coupon := model.Coupons[0]
	if coupon.IsDelete() {
		return 0 // 优惠券已删除，则抵扣金额为0
	}
	couponOriginAmount := coupon.CouponOriginAmount // 优惠券原始金额(面值)
	// 如果积分抵扣之后的订单金额大于优惠券面值，则抵扣金额为优惠券面值，否则抵扣金额为积分抵扣之后的订单金额
	amount := model.GetPointsExchangeAmount() // 积分抵扣后的订单金额
	if amount > couponOriginAmount {
		return couponOriginAmount
	} else {
		return amount
	}
}

// 计算销售订单的订单优惠折扣抹零金额。根据订单设置的优惠折扣抹零规则金额计算
func (model *SaleOrder) calcZeroFee(amount float64) float64 {
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

// CalcGiftAmount 计算赠菜金额. 赠菜金额=销售订单赠菜商品.总最终单价之和
func (model *SaleOrder) CalcGiftAmount(products []*SaleOrderProduct) float64 {
	amount := float64(0)
	for _, saleOrderProduct := range products {
		if saleOrderProduct.IsCancelProduct() {
			continue
		}
		if saleOrderProduct.IsPackageSubProduct() {
			continue
		}
		if saleOrderProduct.IsGiftProduct() {
			// 商品的最终金额
			giftFee := saleOrderProduct.GetSalePrice()
			// 累计各个赠品的最终金额
			amount = decimal.NewFromFloat(amount).Add(decimal.NewFromFloat(giftFee)).InexactFloat64()
		}
	}
	return amount
}

// 计算订单服务费。
// 当服务费关闭时，订单服务费为0
// 当服务费为固定费用时，订单服务费为固定费用
// 当服务费为按比例收费时，订单服务费=所有订单商品的服务费之和+所有自助餐顾客的服务费之和
func (model *SaleOrder) calcServiceFee(products []*SaleOrderProduct, serviceFeeType int, serviceFeeValue float64) float64 {
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
		// 订单商品服务费之和
		for _, orderProduct := range products {
			serviceFee = serviceFee.Add(decimal.NewFromFloat(orderProduct.GetServiceFee()))
		}
		// 自助餐顾客服务费之和
		for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
			serviceFee = serviceFee.Add(decimal.NewFromFloat(buffetCustomer.GetServiceFee()))
		}
		return serviceFee.Truncate(3).Round(2).InexactFloat64()
	}
	// 默认不收取服务费
	return 0
}

// 计算订单服务费。
// 当服务费关闭时，订单服务费为0
// 当服务费为固定费用时，订单服务费为固定费用
// 当服务费为按比例收费时，订单服务费=所有订单商品的服务费之和+所有自助餐顾客的服务费之和
func (model *SaleOrder) calcOriginServiceFee(products []*SaleOrderProduct, serviceFeeType int, serviceFeeValue float64, taxFeeType int) float64 {
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
		// 订单商品服务费之和
		for _, orderProduct := range products {
			serviceFee = serviceFee.Add(decimal.NewFromFloat(orderProduct.GetOriginServiceFee(serviceFeeValue, taxFeeType)))
		}
		// 自助餐顾客服务费之和
		for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
			serviceFee = serviceFee.Add(decimal.NewFromFloat(buffetCustomer.GetOriginServiceFee(serviceFeeValue, taxFeeType)))
		}
		return serviceFee.Truncate(3).Round(2).InexactFloat64()
	}
	// 默认不收取服务费
	return 0
}

func (model *SaleOrder) BeforeCalc() Calc {
	return Calc{
		ProductOriginalAmount: model.ProductOriginalAmount,
		ProductAmount:         model.ProductAmount,
		ServiceFee:            model.ServiceFee,
		TaxFee:                model.TaxFee,
		CustomDiscountFee:     model.CustomDiscountFee,
		MemberDiscountFee:     model.MemberDiscountFee,
		Amount:                model.GetAmount(),
		ZeroFee:               model.ZeroFee,
	}
}

type Calc struct {
	ProductOriginalAmount float64 `json:"product_original_amount"` // 订单商品金额（折前价）
	ProductAmount         float64 `json:"product_amount"`          // 订单商品金额（折后价）
	ServiceFee            float64 `json:"service_fee"`             // 订单服务费
	TaxFee                float64 `json:"tax_fee"`                 // 订单消费税
	CustomDiscountFee     float64 `json:"custom_discount_fee"`     // 订单优惠折扣
	MemberDiscountFee     float64 `json:"member_discount_fee"`     // 订单会员折扣
	PayPointsAmount       float64 `json:"pay_points_amount"`       // 订单积分抵扣金额
	Amount                float64 `json:"amount"`                  // 订单应付金额
	OriginAmount          float64 `json:"origin_amount"`           // 订单原始应收金额
	ZeroFee               float64 `json:"zero_fee"`                // 订单优惠折扣抹零金额
}
