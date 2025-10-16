package model

import (
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

type CalcOption struct {
	IsLatestPrice         bool             // 是否获取最新价格
	SaleBillSetting       *SaleBillSetting // 新的销售账单设置
	CookingStatus         int              // 0-包括已送厨和未送厨 1-已送厨的 2-未送厨的
	H5OrderStatus         int              // 0-包括已下单和未下单 1-已接单的 2-未接单的
	IsOriginPrice         bool             // 是否计算原价. 默认计算折后价
	H5OrderUuid           uint64           // 指定一个h5订单uuid
	H5CheckLimit          bool             // 是否是在h5端检查限购的场景
	SaleOrderProductUuids []uint64         // 指定一个或多个销售订单商品,仅检查这个或这些销售订单商品是否超过限购
	ServiceFeeBase        uint             // 服务费计算基准, 0-商品惠后价 1-商品价格合计
	IsCanceled            bool             // 是否是取消订单的场景
}

const (
	CookingStatusAll              = iota // 0-包括已送厨和未送厨
	CookingStatusCooking                 // 1-已送厨的
	CookingStatusUnCooking               // 2-未送厨的
	CookingStatusAllAndOneH5Order        // 3-包括已送厨和未送厨和一个h5订单的商品
)

const (
	H5OrderStatusUnAccepted = 1 // 1-未接单的
	H5OrderStatusAccepted   = 2 // 2-已接单的
)

func WithCanceled() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.IsCanceled = true
	}
}

// WithLatestPrice 用最新的价格信息计算订单金额
func WithLatestPrice() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.IsLatestPrice = true
	}
}

// WithSaleBillSetting 使用新的销售账单设置
func WithSaleBillSetting(setting *SaleBillSetting) func(option *CalcOption) {
	return func(option *CalcOption) {
		option.SaleBillSetting = setting
	}
}

// WithCooking 获取已送厨的商品
func WithCooking() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.CookingStatus = CookingStatusCooking
	}
}

// WithUnCooking 获取未送厨的商品
func WithUnCooking() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.CookingStatus = CookingStatusUnCooking
	}
}

// WithH5Order 获取h5已下单的商品
func WithH5Order() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.H5OrderStatus = H5OrderStatusAccepted
	}
}

// WithH5Cart 获取h5购物车的商品（未下单的商品）
func WithH5Cart() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.H5OrderStatus = H5OrderStatusUnAccepted
	}
}

// WithAll 获取全部商品，包括已送厨和未送厨
func WithAll() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.CookingStatus = CookingStatusAll
	}
}

// WithAllAndOneH5Order 获取全部商品，包括已送厨和未送厨和某个h5订单的商品
func WithAllAndOneH5Order(h5OrderUuid uint64) func(option *CalcOption) {
	return func(option *CalcOption) {
		option.CookingStatus = CookingStatusAllAndOneH5Order
		option.H5OrderUuid = h5OrderUuid
	}
}

// WithOriginPrice 计算原价
func WithOriginPrice() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.IsOriginPrice = true
	}
}

func WithH5CheckLimit() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.H5CheckLimit = true
	}
}

func WithH5OrderUuid(h5OrderUuid uint64) func(option *CalcOption) {
	return func(option *CalcOption) {
		option.H5OrderUuid = h5OrderUuid
	}
}

func WithSaleOrderProductUuid(saleOrderProductUuids ...uint64) func(option *CalcOption) {
	return func(option *CalcOption) {
		option.SaleOrderProductUuids = saleOrderProductUuids
	}
}

// 设置服务费计算基准为商品价格合计
func WithServiceFeeBaseAmount() func(option *CalcOption) {
	return func(option *CalcOption) {
		option.ServiceFeeBase = constant.SaleBillSettingServiceFeeBaseAmount
	}
}

func (model *SaleOrderProduct) CalcSaleOrderProduct(setting SaleBillSetting, options ...func(option *CalcOption)) SaleOrderProductCalc {
	defer model.SetUpdate() // 标记该记录要更新
	serviceFeeRate := setting.GetServiceFeeRate()
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()

	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	// 如果服务费计算基准为商品价格合计，则需要使用商品价格合计计算服务费
	if setting.ServiceFeeBase == constant.SaleBillSettingServiceFeeBaseAmount {
		options = append(options, WithServiceFeeBaseAmount())
	}

	if option.IsLatestPrice {
		return model.calcLastestSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType, options...)
	}
	return model.calcSaleOrderProduct(serviceFeeRate, taxFeeType, serviceFeeType, options...)
}

// 计算销售订单商品的所有计算值字段
func (model *SaleOrderProduct) calcSaleOrderProduct(serviceFeeRate float64, taxFeeType int, serviceFeeType int, options ...func(option *CalcOption)) SaleOrderProductCalc {
	calc := SaleOrderProductCalc{}
	// 开始计算
	calc.SaucePrice = model.calcSaucePrice()
	model.SaucePrice = calc.SaucePrice
	calc.ProductPrice = model.calcProductPrice()
	model.ProductPrice = calc.ProductPrice
	calc.SalePrice = model.calcSalePrice()
	model.SalePrice = calc.SalePrice
	calc.SalePriceNoTax = model.calcProductPriceNoneTax(model.SalePrice, taxFeeType)
	model.SalePriceNoTax = calc.SalePriceNoTax
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
	calc.ServiceFee = model.calcServiceFee(model.Price, serviceFeeRate, taxFeeType, options...)
	model.ServiceFee = calc.ServiceFee
	calc.ServiceTaxFee = model.calcServiceTaxFee(model.Price, serviceFeeRate, taxFeeType, serviceFeeType, options...)
	model.ServiceTaxFee = calc.ServiceTaxFee
	calc.TotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType, options...)
	model.TotalPrice = calc.TotalPrice
	calc.OriginTotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType, WithOriginPrice())
	model.OriginTotalPrice = calc.OriginTotalPrice
	return calc
}

// 计算销售订单商品的所有计算值字段。从后台获取最新的价格
func (model *SaleOrderProduct) calcLastestSaleOrderProduct(serviceFeeRate float64, taxFeeType int, serviceFeeType int, options ...func(option *CalcOption)) SaleOrderProductCalc {
	// 设置最新的商品会员折扣设置，即是否开启会员折扣
	model.OpenMemberDiscount = model.ProductPackage.OpenDiscount
	model.OpenOverallDiscount = *model.ProductPackage.OpenOverallDiscount

	calc := SaleOrderProductCalc{}
	// 开始计算
	calc.SaucePrice = model.calcLastestSaucePrice() // 从后台获取最新的价格
	model.SaucePrice = calc.SaucePrice
	calc.ProductPrice = model.calcLastestProductPrice() // 从后台获取最新的价格
	model.ProductPrice = calc.ProductPrice
	calc.SalePrice = model.calcSalePrice()
	model.SalePrice = calc.SalePrice
	calc.SalePriceNoTax = model.calcProductPriceNoneTax(model.SalePrice, taxFeeType)
	model.SalePriceNoTax = calc.SalePriceNoTax
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
	calc.ServiceFee = model.calcServiceFee(model.Price, serviceFeeRate, taxFeeType, options...)
	model.ServiceFee = calc.ServiceFee
	calc.ServiceTaxFee = model.calcServiceTaxFee(model.Price, serviceFeeRate, taxFeeType, serviceFeeType, options...)
	model.ServiceTaxFee = calc.ServiceTaxFee
	calc.TotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType, options...)
	model.TotalPrice = calc.TotalPrice
	calc.OriginTotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType, WithOriginPrice())
	model.OriginTotalPrice = calc.OriginTotalPrice
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
	return price.Round(2).InexactFloat64()
}

// 计算会员折扣率。会员折扣率=会员等级折扣率*会员卡折扣率
// 如果商品不参与会员打折的话，会员折扣率=0
func (model *SaleOrderProduct) calcMemberDiscountRate() float64 {
	memberDiscountRate := decimal.NewFromFloat(model.GetMemberDiscountRate()).Mul(decimal.NewFromFloat(model.GetMemberCardDiscountRate()))
	return memberDiscountRate.InexactFloat64()
}

// 计算商品的会员折扣费用。会员折扣费用(单个商品)=(商品销售价-商品销售价*会员折扣率)  = 商品销售价 *（1-会员折扣率）
// 注意： 商品销售价 *（1-会员折扣率）  这个公式不能使用，会因为四舍五入导致计算错误。如：商品销售价=75.7，会员折扣率=0.85，则折后价为64.345，四舍五入后为64.35，再计算会员折扣费用时，结果为11.355，再四舍五入后为11.36，64.35+11.36=75.71，计算结果错误多1分钱。
// 正确计算公式： 会员折扣费用= 商品销售价-商品销售价*会员折扣率
func (model *SaleOrderProduct) calcMemberDiscountFee() float64 {
	// 当会员折扣率为0时，会员折扣费用=0
	memberDiscountRate := model.calcMemberDiscountRate()

	// 会员折扣率的取值范围是大于0、小于等于1。如果值大于或等于1，则强制解释为不打折，折扣费用为0
	if memberDiscountRate >= constant.NoDiscount {
		return 0
	}
	// 如果商品赠送，则不计算会员折扣费用
	if model.IsGiftProduct() {
		return 0
	}

	// 会员折扣费用= 商品销售价-商品销售价*会员折扣率
	price := model.calcSalePrice()
	discountPrice := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(memberDiscountRate)).Round(2) // 商品折后价
	memberDiscountFee := decimal.NewFromFloat(price).Sub(discountPrice)
	return memberDiscountFee.Round(2).InexactFloat64()
}

// 当有会员折扣时，自定义折扣费  = 会员折扣价-会员折扣价*自定义折扣率 = 会员折扣价*（1-自定义折扣率）=（商品销售价-会员折扣费）*（1-自定义折扣率）；
// 当没有会员折扣时，自定义折扣费= 商品销售价- 商品销售价*自定义折扣率 = 商品销售价*（1-自定义折扣率）= （商品销售价-会员折扣费0）*（1-自定义折扣率）
// 当没有会员折扣时，会员折扣费为0，则两个情况的算法可以都用 自定义折扣费=会员折扣价*（1-自定义折扣率）
func (model *SaleOrderProduct) calcCustomDiscountFee() float64 {
	customDiscountRate := model.GetCustomDiscountRate()
	if customDiscountRate == constant.NoDiscount {
		return 0
	}

	// 如果商品赠送，则不计算自定义折扣费用. 但销售订单的优惠折扣费用中会另外计算加上赠菜金额
	if model.IsGiftProduct() {
		return 0
	}

	// 会员折扣价 = 商品销售价-会员折扣费。没有会员时，会员折扣费为0。
	memberDiscountPrice := decimal.NewFromFloat(model.calcSalePrice()).Sub(decimal.NewFromFloat(model.calcMemberDiscountFee()))
	//（1-自定义折扣率）
	discount := decimal.NewFromFloat(1).Sub(decimal.NewFromFloat(customDiscountRate))
	// 会员折扣价*（1-自定义折扣率）
	// TODO 不能用这种方式计算自定义折扣金额，因为会因为四舍五入导致计算错误。
	// 自定义折扣费 = 商品销售价-会员折扣价
	customDiscountFee := memberDiscountPrice.Mul(discount)
	return customDiscountFee.Truncate(3).Round(2).InexactFloat64()
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
	// 如果商品赠送，则不计算税费
	if model.IsGiftProduct() {
		return 0
	}
	// 商品已含税时，消费税税费=商品销售价-商品未含税销售价
	if taxFeeType == constant.TaxFeeTypeTax {
		taxFee := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(model.calcProductPriceNoneTax(price, taxFeeType)))
		return taxFee.Truncate(3).Round(2).InexactFloat64()
	} else if taxFeeType == constant.TaxFeeTypeNoTax {
		// 商品未含税时，消费税税费=商品销售价*消费税税率
		taxFee := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(model.TaxRate))
		return taxFee.Truncate(3).Round(2).InexactFloat64()
	}

	// 默认为商品关闭税费
	return 0
}

// 计算商品原税费。
// 商品未含税时，商品原税费= salePrice * 消费税税率
// 商品已含税时，商品原税费= 商品未含税金额 * 消费税税率
func (model *SaleOrderProduct) calcOriginTaxFee(taxFeeType int) float64 {
	price := model.SalePrice
	return model.calcTaxFee(price, taxFeeType)
}

// 服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例
func (model *SaleOrderProduct) calcServiceFee(price float64, serviceFeeRate float64, taxFeeType int, options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	// 如果服务费计算基准为商品惠后价，则赠品不计算服务费
	if option.ServiceFeeBase == constant.SaleBillSettingServiceFeeBasePrice {
		// 如果商品赠送，则不计算服务费
		if !option.IsOriginPrice { // 如果是计算原始价格，则赠品要计算服务费. 且服务费计算基准无论是什么都要计算服务费，按原价计算服务费
			if model.IsGiftProduct() {
				return 0
			}
		}
	}

	// 如果服务费计算基准为商品价格合计，则price价格应该是ProductPrice '原始单价（单商品）,规格原价+小料价'
	if !option.IsOriginPrice && option.ServiceFeeBase == constant.SaleBillSettingServiceFeeBaseAmount {
		price = model.ProductPrice
	}

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
		return serviceFee.Truncate(3).Round(2).InexactFloat64()
	}
	// 已含税时
	if taxFeeType == constant.TaxFeeTypeTax {
		// 商品未含税价格=（最终单价-商品税费）
		priceNoneTax := model.calcProductPriceNoneTax(price, taxFeeType) // decimal.NewFromFloat(model.calcPrice()).Sub(decimal.NewFromFloat(model.calcTaxFee(taxFeeType)))
		priceNoneTaxDecimal := decimal.NewFromFloat(priceNoneTax)
		//  服务费=（最终单价-商品税费）*服务费比例
		serviceFee := priceNoneTaxDecimal.Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.Truncate(3).Round(2).InexactFloat64()
	}

	// 关闭税费时
	if taxFeeType == constant.TaxFeeTypeNone {
		// 服务费=最终单价*服务费比例
		serviceFee := decimal.NewFromFloat(price).Mul(decimal.NewFromFloat(serviceFeeRate))
		return serviceFee.Truncate(3).Round(2).InexactFloat64()
	}
	// 默认商品不收取服务费
	return 0
}

// 计算订单商品的服务费税费。
// 当不收取服务费税费时，服务费税费为0
// 当收取服务费税费时，服务费税费=订单商品服务费*商品消费税税率
func (model *SaleOrderProduct) calcServiceTaxFee(price float64, serviceFeeRate float64, taxFeeType int, serviceFeeType int, options ...func(option *CalcOption)) float64 {
	// 当服务费收费税费且开启税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax && taxFeeType != constant.TaxFeeTypeNone {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceFee := model.calcServiceFee(price, serviceFeeRate, taxFeeType, options...) // 计算订单商品服务费。price可能是折后价，也可能是折前价。当计算订单金额（折前）时，price是折前价；当计算订单应收金额时，price是折后价。
		serviceTaxFee := decimal.NewFromFloat(serviceFee).Mul(decimal.NewFromFloat(model.TaxRate))
		// fmt.Println("订单商品服务费", serviceFee, "税费", serviceTaxFee.InexactFloat64(), serviceTaxFee.Truncate(3).Round(2).InexactFloat64())
		return serviceTaxFee.Truncate(3).Round(2).InexactFloat64()
	}
	return 0
}

// 计算商品的总折扣费用。单个商品总折扣费用=会员折扣费用+自定义打折折扣费用
func (model *SaleOrderProduct) calcDiscountFee() float64 {
	// 会员折扣费用+自定义打折折扣费用
	discountFee := decimal.NewFromFloat(model.calcMemberDiscountFee()).Add(decimal.NewFromFloat(model.calcCustomDiscountFee()))
	return discountFee.InexactFloat64()
}

// 计算单个商品最终应收金额（折后价）。
// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+总税费=最终价格（折后价）+服务费+（商品税费+服务费税费）
// 商品已含税时，单个商品最终应收金额=商品折后不含税价格+服务费+总税费=（最终价格（折扣价）-商品税费） + 服务费 + 总税费 = （最终价格（折后价）-商品税费） + 服务费 + （商品税费+服务费税费）= 最终价格（折后价）+ 服务费 + 服务费税费
// 总结，商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
func (model *SaleOrderProduct) calcTotalPrice(serviceFeeRate float64, taxFeeType int, serviceFeeType int, options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	price := model.calcPrice() // 折后价
	if option.IsOriginPrice {
		price = model.calcSalePrice() // 折前价
	}

	// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+商品税费+服务费税费
	if taxFeeType == constant.TaxFeeTypeNoTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.calcServiceFee(price, serviceFeeRate, taxFeeType, options...))).Add(
			decimal.NewFromFloat(model.calcTaxFee(price, taxFeeType))).Add(
			decimal.NewFromFloat(model.calcServiceTaxFee(price, serviceFeeRate, taxFeeType, serviceFeeType, options...)))
		return totalPrice.Truncate(3).Round(2).InexactFloat64()
	}
	// 当商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
	if taxFeeType == constant.TaxFeeTypeTax {
		totalPrice := decimal.NewFromFloat(price).Add(
			decimal.NewFromFloat(model.calcServiceFee(price, serviceFeeRate, taxFeeType, options...))).Add(
			decimal.NewFromFloat(model.calcServiceTaxFee(price, serviceFeeRate, taxFeeType, serviceFeeType, options...)))
		return totalPrice.Truncate(3).Round(2).InexactFloat64()
	}
	// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
	totalPrice := decimal.NewFromFloat(price).Add(
		decimal.NewFromFloat(model.calcServiceFee(price, serviceFeeRate, taxFeeType, options...)))
	return totalPrice.Truncate(3).Round(2).InexactFloat64()
}

// 计算小料的价格。累计销售订单商品的所有小料的价格
func (model *SaleOrderProduct) calcSaucePrice() float64 {
	saucePrice := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsDelete() {
			continue
		}
		if !bom.IsFlavor() {
			// 累加每个小料的价格
			price := decimal.NewFromFloat(bom.Price) // 加料单价
			// 如果要上浮价格的话
			if model.GetMemberOrderDiscountRate() != 1 {
				price = price.Mul(decimal.NewFromFloat(model.GetMemberOrderDiscountRate())).Round(2)
			}
			saucePrice = saucePrice.Add(price)
		}
	}
	return saucePrice.InexactFloat64()
}

// 计算商品的最新小料价格。从后台设置的小料价格中获取最新的价格
func (model *SaleOrderProduct) calcLastestSaucePrice() float64 {
	saucePrice := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsDelete() {
			continue
		}
		if !bom.IsFlavor() {
			// 累加每个小料的价格
			price := bom.ProductBom.Price
			bom.SetPrice(price)
			if price > 0 {
				saucePrice = saucePrice.Add(decimal.NewFromFloat(price))
			}
		}
	}
	return saucePrice.InexactFloat64()
}

// 计算商品价格。某个规格商品价+小料价
// 当商品没有改价时,ProductPrice= 某个规格商品价+小料价
// 当商品改价时，ProductPrice= ProductPrice 。 改价后不会修改这个字段的值，只会修改salePrice的值
func (model *SaleOrderProduct) calcProductPrice() float64 {
	productPrice := decimal.NewFromFloat(model.GetFlavorPrice()).Add(decimal.NewFromFloat(model.SaucePrice))
	return productPrice.InexactFloat64()
}

// 计算商品的最新价格。从后台设置的商品规格价格和小料价格中获取最新的价格
func (model *SaleOrderProduct) calcLastestProductPrice() float64 {
	flavorPrice := model.FlavorPrice
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsDelete() {
			continue
		}
		if bom.IsFlavor() {
			flavorPrice = bom.ProductBom.Price
			model.SetFlavorPrice(flavorPrice) // 并更新规格价格
			bom.SetPrice(flavorPrice)         // 并更新规格价格
		}
	}
	// 如果规格价格大于0，则使用规格价格
	productPrice := decimal.NewFromFloat(flavorPrice).Add(decimal.NewFromFloat(model.calcLastestSaucePrice()))
	return productPrice.InexactFloat64()
}

// 是否为自助餐商品
func (model *SaleOrderProduct) IsBuffetProduct() bool {
	return model.IsBuffet == constant.SaleOrderProductIsBuffetYes
}

// 是否为计量商品
func (model *SaleOrderProduct) IsMetering() bool {
	return model.NumType == constant.SaleOrderProductIsMeteringYes
}

// 计算商品销售价。
// 如果商品改价，则直接修改SalePrice。
// 如果没有改价，销售价=ProductPrice
// 如果商品是自助餐商品，则销售价=小料原价
func (model *SaleOrderProduct) calcSalePrice() float64 {
	if model.IsCustomPriceBool() {
		return model.SalePrice
	}
	if model.IsBuffetProduct() {
		return model.SaucePrice
	}
	return model.ProductPrice
}

// 计算商品的折扣率。 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
func (model *SaleOrderProduct) calcDiscountRate() float64 {
	rate := decimal.NewFromFloat(1)
	memberDiscountRate := model.GetMemberDiscountRate()
	memberCardDiscountRate := model.GetMemberCardDiscountRate()
	customDiscountRate := model.GetCustomDiscountRate()
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(memberDiscountRate))
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(memberCardDiscountRate))
	// 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
	rate = rate.Mul(decimal.NewFromFloat(customDiscountRate))
	return rate.InexactFloat64()
}

type SaleOrderProductCalc struct {
	SaucePrice        float64 `json:"sauce_price"`
	ProductPrice      float64 `json:"product_price"`
	SalePrice         float64 `json:"sale_price"`
	SalePriceNoTax    float64 `json:"sale_price_no_tax"`
	Price             float64 `json:"price"`
	MemberDiscountFee float64 `json:"member_discount_fee"`
	CustomDiscountFee float64 `json:"custom_discount_fee"`
	DiscountFee       float64 `json:"discount_fee"`
	TaxFee            float64 `json:"tax_fee"`
	ServiceFee        float64 `json:"service_fee"`
	ServiceTaxFee     float64 `json:"service_tax_fee"`
	TotalPrice        float64 `json:"total_price"`
	OriginTotalPrice  float64 `json:"origin_total_price"`
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
