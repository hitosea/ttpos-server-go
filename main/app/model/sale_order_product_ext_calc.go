package model

import (
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

func (model *SaleOrderProduct) CalcSaleOrderProduct(setting SaleBillSetting) SaleOrderProductCalc {
	defer model.SetUpdate() // 标记该记录要更新
	serviceFeeRate := setting.GetServiceFeeRate()
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	return model.calcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType)
}

// 计算销售订单商品的所有计算值字段
func (model *SaleOrderProduct) calcSaleOrderProduct(serviceFeeRate float64, taxFeeType int, serviceFeeType int) SaleOrderProductCalc {
	calc := SaleOrderProductCalc{}
	// 开始计算
	calc.SaucePrice = model.calcSaucePrice()
	model.SaucePrice = calc.SaucePrice
	calc.ProductPrice = model.calcProductPrice()
	model.ProductPrice = calc.ProductPrice
	calc.SalePrice = model.calcSalePrice()
	model.SalePrice = calc.SalePrice
	calc.Price = model.calcPrice()
	model.Price = calc.Price
	calc.MemberDiscountFee = model.calcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.CustomDiscountFee = model.calcCustomDiscountFee()
	model.CustomDiscountFee = calc.CustomDiscountFee
	calc.DiscountFee = model.calcDiscountFee()
	model.DiscountFee = calc.DiscountFee
	calc.TaxFee = model.calcTaxFee(model.Price, taxFeeType)
	model.TaxFee = calc.TaxFee
	calc.ServiceFee = model.calcServiceFee(serviceFeeRate, taxFeeType)
	model.ServiceFee = calc.ServiceFee
	calc.ServiceTaxFee = model.calcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)
	model.ServiceTaxFee = calc.ServiceTaxFee
	calc.TotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType)
	model.TotalPrice = calc.TotalPrice
	return calc
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

// 计算商品折后价。最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率
func (model *SaleOrderProduct) calcPrice() float64 {
	discountRate := model.calcDiscountRate()
	if discountRate == constant.NoDiscount {
		return model.SalePrice
	}
	// 销售价*折扣率
	price := decimal.NewFromFloat(model.SalePrice).Mul(
		decimal.NewFromFloat(discountRate))
	return price.InexactFloat64()
}

// 计算会员折扣率。会员折扣率=会员等级折扣率*会员卡折扣率
// 如果商品不参与会员打折的话，会员折扣率=0
func (model *SaleOrderProduct) calcMemberDiscountRate() float64 {
	if model.OpenMemberDiscount == constant.ProductMemberDiscountOff {
		return constant.NoDiscount
	}
	//if model.MemberDiscountRate == 0 && model.MemberCardDiscountRate != 0 {
	//	return model.MemberCardDiscountRate
	//} else if model.MemberCardDiscountRate == 0 && model.MemberDiscountRate != 0 {
	//	return model.MemberDiscountRate
	//} else if model.MemberCardDiscountRate != 0 && model.MemberDiscountRate != 0 {
	//	memberDiscountRate := decimal.NewFromFloat(model.MemberDiscountRate).Mul(decimal.NewFromFloat(model.MemberCardDiscountRate))
	//	return memberDiscountRate.InexactFloat64()
	//}
	//// 不匹配时默认为0
	//return 0

	memberDiscountRate := decimal.NewFromFloat(model.MemberDiscountRate).Mul(decimal.NewFromFloat(model.MemberCardDiscountRate))
	return memberDiscountRate.InexactFloat64()
}

// 计算商品的会员折扣费用。会员折扣费用=(商品销售价-商品销售价*会员折扣率) * 商品数量 =商品销售价 * 商品数量 *（1-会员折扣率）
func (model *SaleOrderProduct) calcMemberDiscountFee() float64 {
	// 当会员折扣率为0时，会员折扣费用=0
	memberDiscountRate := model.calcMemberDiscountRate()

	// 会员折扣率的取值范围是大于0、小于等于1。如果值大于或等于1，则强制解释为不打折，折扣费用为0
	if memberDiscountRate >= constant.NoDiscount {
		return 0
	}
	// 1-会员折扣率
	discount := decimal.NewFromFloat(1).Sub(decimal.NewFromFloat(memberDiscountRate))
	// 商品销售价 * 商品数量 *（1-会员折扣率）
	memberDiscountFee := decimal.NewFromFloat(model.calcSalePrice()).Mul(decimal.NewFromUint64(uint64(model.Num))).Mul(discount)
	return memberDiscountFee.InexactFloat64()
}

// 当有会员折扣时，自定义折扣费  = 会员折扣价-会员折扣价*自定义折扣率 = 会员折扣价*（1-自定义折扣率）=（商品销售价-会员折扣费）*（1-自定义折扣率）；
// 当没有会员折扣时，自定义折扣费= 商品销售价- 商品销售价*自定义折扣率 = 商品销售价*（1-自定义折扣率）= （商品销售价-会员折扣费0）*（1-自定义折扣率）
// 当没有会员折扣时，会员折扣费为0，则两个情况的算法可以都用 自定义折扣费=会员折扣价*（1-自定义折扣率）
func (model *SaleOrderProduct) calcCustomDiscountFee() float64 {
	customDiscountRate := model.CustomDiscountRate
	if customDiscountRate == constant.NoDiscount {
		return 0
	}
	// 会员折扣价 = 商品销售价-会员折扣费。没有会员时，会员折扣费为0。
	memberDiscountPrice := decimal.NewFromFloat(model.calcSalePrice()).Sub(decimal.NewFromFloat(model.calcMemberDiscountFee()))
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
func (model *SaleOrderProduct) calcProductPriceNoneTax(price float64, taxFeeType int) float64 {
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
func (model *SaleOrderProduct) calcTaxFee(price float64, taxFeeType int) float64 {
	// 商品已含税时，消费税税费=商品销售价-商品未含税销售价
	if taxFeeType == constant.TaxFeeTypeTax {
		taxFee := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(model.calcProductPriceNoneTax(price, taxFeeType)))
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
func (model *SaleOrderProduct) calcOriginTaxFee(taxFeeType int) float64 {
	price := model.SalePrice
	return model.calcTaxFee(price, taxFeeType)
}

// 服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例
func (model *SaleOrderProduct) calcServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
	price := model.calcPrice()

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
		priceNoneTax := model.calcProductPriceNoneTax(price, taxFeeType) // decimal.NewFromFloat(model.calcPrice()).Sub(decimal.NewFromFloat(model.calcTaxFee(taxFeeType)))
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
func (model *SaleOrderProduct) calcOriginServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
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
		serviceFee := decimal.NewFromFloat(model.calcSalePrice()).Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.InexactFloat64()
	}
	// 已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		// 商品未含税价格=（销售价-商品原税费）
		priceNoneTax := decimal.NewFromFloat(model.calcSalePrice()).Sub(decimal.NewFromFloat(model.calcTaxFee(model.SalePrice, taxFeeType)))
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
func (model *SaleOrderProduct) calcServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 当服务费收费税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceTaxFee := decimal.NewFromFloat(model.calcServiceFee(serviceFeeRate, taxFeeType)).Mul(decimal.NewFromFloat(model.TaxRate))
		return serviceTaxFee.InexactFloat64()
	}
	return 0
}

// 计算订单商品的原服务费税费（打折前）。
// 当不收取服务费税费时，服务费税费为0
// 当收取服务费税费时，服务费税费=订单商品服务费（打折前）*商品消费税税率
func (model *SaleOrderProduct) calcOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 当服务费收费税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceTaxFee := decimal.NewFromFloat(model.calcOriginServiceFee(serviceFeeRate, taxFeeType)).Mul(decimal.NewFromFloat(model.TaxRate))
		return serviceTaxFee.InexactFloat64()
	}
	return 0
}

// 计算商品的总折扣费用。单个商品总折扣费用=会员折扣费用+自定义打折折扣费用
func (model *SaleOrderProduct) calcDiscountFee() float64 {
	// 会员折扣费用+自定义打折折扣费用
	discountFee := decimal.NewFromFloat(model.calcMemberDiscountFee()).Add(decimal.NewFromFloat(model.calcCustomDiscountFee()))
	return discountFee.InexactFloat64()
}

// 计算单个商品最终应收金额。
// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+总税费=最终价格（折后价）+服务费+（商品税费+服务费税费）
// 商品已含税时，单个商品最终应收金额=商品折后不含税价格+服务费+总税费=（最终价格（折扣价）-商品税费） + 服务费 + 总税费 = （最终价格（折后价）-商品税费） + 服务费 + （商品税费+服务费税费）= 最终价格（折后价）+ 服务费 + 服务费税费
// 总结，商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
func (model *SaleOrderProduct) calcTotalPrice(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+商品税费+服务费税费
	price := model.calcPrice()
	if taxFeeType == constant.TaxFeeTypeNoTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.calcServiceFee(serviceFeeRate, taxFeeType))).Add(
			decimal.NewFromFloat(model.calcTaxFee(price, taxFeeType))).Add(
			decimal.NewFromFloat(model.calcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
		return totalPrice.InexactFloat64()
	}
	// 当商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
	if taxFeeType == constant.TaxFeeTypeTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.calcServiceFee(serviceFeeRate, taxFeeType))).Add(
			decimal.NewFromFloat(model.calcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)))
		return totalPrice.InexactFloat64()
	}
	// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
	totalPrice := decimal.NewFromFloat(price).Add(
		decimal.NewFromFloat(model.calcServiceFee(serviceFeeRate, taxFeeType)))
	return totalPrice.InexactFloat64()
}

// 计算小料的价格。累计销售订单商品的所有小料的价格
func (model *SaleOrderProduct) calcSaucePrice() float64 {
	saucePrice := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if !bom.IsFlavor() {
			// 累加每个小料的价格
			saucePrice = saucePrice.Add(decimal.NewFromFloat(bom.Price))
		}
	}
	return saucePrice.InexactFloat64()
}

// 计算商品价格。某个规格商品价+小料价
// 当商品没有改价时,ProductPrice= 某个规格商品价+小料价
// 当商品改价时，ProductPrice= ProductPrice 。 改价后不会修改这个字段的值，只会修改salePrice的值
func (model *SaleOrderProduct) calcProductPrice() float64 {
	//if model.ChangePriceTime == constant.CustomPriceOn {
	//	return model.ProductPrice
	//}
	productPrice := decimal.NewFromFloat(model.FlavorPrice).Add(decimal.NewFromFloat(model.calcSaucePrice()))
	return productPrice.InexactFloat64()
}

func (model *SaleOrderProduct) IsBuffetProduct() bool {
	return model.IsBuffet == constant.SaleOrderProductIsBuffetYes
}

// 计算商品销售价。
// 如果商品改价，则直接修改SalePrice。
// 如果没有改价，销售价=ProductPrice
// 如果商品是自助餐商品，则销售价=0
func (model *SaleOrderProduct) calcSalePrice() float64 {
	if model.IsCustomPriceBool() {
		return model.SalePrice
	}
	if model.IsBuffetProduct() {
		return 0
	}
	return model.ProductPrice
}

// 计算商品的折扣率。 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
func (model *SaleOrderProduct) calcDiscountRate() float64 {
	rate := decimal.NewFromFloat(1)
	memberDiscountRate := model.MemberDiscountRate
	memberCardDiscountRate := model.MemberCardDiscountRate
	customDiscountRate := model.CustomDiscountRate
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(memberDiscountRate))
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(memberCardDiscountRate))
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(customDiscountRate))
	return rate.InexactFloat64()
}
