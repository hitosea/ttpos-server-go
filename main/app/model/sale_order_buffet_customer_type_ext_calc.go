package model

import (
	"fmt"
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

type SaleOrderBuffetCustomerCalc struct {
	Price             float64 `json:"price"`
	CustomDiscountFee float64 `json:"custom_discount_fee"`
	ServiceTaxFee     float64 `json:"service_tax_fee"`
	TaxFee            float64 `json:"tax_fee"`
	ServiceFee        float64 `json:"service_fee"`
	TotalPrice        float64 `json:"total_price"`
}

// 计算销售订单自助餐顾客类型
func (model *SaleOrderBuffetCustomerType) calcSaleOrderBuffetCustomerType(serviceFeeRate float64, taxFeeType int, serviceFeeType int) SaleOrderBuffetCustomerCalc {
	calc := SaleOrderBuffetCustomerCalc{}
	calc.Price = model.calcPrice()
	model.Price = calc.Price
	calc.CustomDiscountFee = model.calcCustomDiscountFee()
	model.CustomDiscountFee = calc.CustomDiscountFee
	calc.ServiceTaxFee = model.calcServiceTaxFee(serviceFeeRate, taxFeeType, serviceFeeType)
	model.ServiceTaxFee = calc.ServiceTaxFee
	calc.TaxFee = model.calcTaxFee(model.Price, taxFeeType)
	model.TaxFee = calc.TaxFee
	calc.ServiceFee = model.calcServiceFee(serviceFeeRate, taxFeeType)
	model.ServiceFee = calc.ServiceFee
	calc.TotalPrice = model.calcTotalPrice(serviceFeeRate, taxFeeType, serviceFeeType)
	model.TotalPrice = calc.TotalPrice
	fmt.Println("q444444")

	return calc
}

// 计算最终单价。最终单价=原始单价*自定义折扣率
func (model *SaleOrderBuffetCustomerType) calcPrice() float64 {
	// 最终单价=原始单价*自定义折扣率
	return decimal.NewFromFloat(model.SalePrice).Mul(
		decimal.NewFromFloat(model.CustomDiscountRate)).InexactFloat64()
}

// 计算自定义折扣金额，表示打折了多少钱。自定义折扣金额=原价-最终单价
func (model *SaleOrderBuffetCustomerType) calcCustomDiscountFee() float64 {
	// 自定义折扣金额=原价-最终单价
	return decimal.NewFromFloat(model.SalePrice).Sub(decimal.NewFromFloat(model.Price)).InexactFloat64()
}

// 计算服务费。服务费=最终单价*服务费比例
func (model *SaleOrderBuffetCustomerType) calcServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
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

// 计算服务费税费。服务费税费=服务费*税率
func (model *SaleOrderBuffetCustomerType) calcServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	// 服务费税费=服务费*税率
	// 当服务费收费税费时
	if serviceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		// 服务费税费=订单商品服务费*商品消费税税率
		serviceTaxFee := decimal.NewFromFloat(model.calcServiceFee(serviceFeeRate, taxFeeType)).Mul(decimal.NewFromFloat(model.TaxRate))
		return serviceTaxFee.InexactFloat64()
	}
	// 当服务费不收费税费时
	return 0
}

// 计算商品的税费（单商品的消费税税费，不含服务费的税费）
// 商品未含税时，商品原价=商品未含税原价。消费税税费=商品未含税原价*消费税税率
// 商品已含税时，商品原价=商品未含税原价+消费税税费；故，消费税税费=商品原价-商品未含税原价。
// 商品原价=商品未含税原价+（商品未含税原价*消费税税率）= 商品未含税原价*（1+消费税税率）
// 商品未含税原价=商品原价/(1+消费税税率)
func (model *SaleOrderBuffetCustomerType) calcTaxFee(taxFeeRate float64, taxFeeType int) float64 {
	// 计算自助餐顾客类型未含税原价的未含税原价
	productPriceNoneTax := model.calcProductPriceNoneTax(model.SalePrice, taxFeeType)
	// 自助餐顾客类型税费=自助餐顾客类型未含税原价*税率
	return decimal.NewFromFloat(productPriceNoneTax).Mul(decimal.NewFromFloat(taxFeeRate)).InexactFloat64()
}

// 计算商品的未含税原价。
// 商品原价=商品未含税原价+（商品未含税原价*消费税税率）= 商品未含税原价*（1+消费税税率）
// 商品未含税原价=商品原价/(1+消费税税率)
func (model *SaleOrderBuffetCustomerType) calcProductPriceNoneTax(price float64, taxFeeType int) float64 {
	// 不收取税费时、关闭税费时, 未含税原价=商品原价
	if taxFeeType == constant.TaxFeeTypeNone {
		return price
	}
	// 商品未含税, 未含税原价=商品原价
	if taxFeeType == constant.TaxFeeTypeNoTax {
		return price
	}
	// 商品已含税, 未含税原价= 商品原价/（1+消费税税率）
	// 商品原价=商品未含税原价+（商品未含税原价*消费税税率）= 商品未含税原价*（1+消费税税率）
	// 商品未含税原价=商品原价/(1+消费税税率)
	if taxFeeType == constant.TaxFeeTypeTax {
		// （1+消费税税率）
		percent := decimal.NewFromFloat(1).Add(decimal.NewFromFloat(model.TaxRate))
		// 商品原价/（1+消费税税率）
		productPriceNoneTax := decimal.NewFromFloat(price).Div(percent)
		return productPriceNoneTax.InexactFloat64()
	}
	// 不收取税费时
	return price
}

// 计算单个商品最终应收金额。
// 如果不收取税费时，单个商品最终应收金额=最终价格（折后价）+ 服务费
// 商品未含税时，单个商品最终应收金额=最终价格（折后价）+服务费+总税费=最终价格（折后价）+服务费+（商品税费+服务费税费）
// 商品已含税时，单个商品最终应收金额=商品折后不含税价格+服务费+总税费=（最终价格（折扣价）-商品税费） + 服务费 + 总税费 = （最终价格（折后价）-商品税费） + 服务费 + （商品税费+服务费税费）= 最终价格（折后价）+ 服务费 + 服务费税费
// 总结，商品已含税时，单个商品最终应收金额=最终价格（折后价）+ 服务费 + 服务费税费
func (model *SaleOrderBuffetCustomerType) calcTotalPrice(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
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

func (model *SaleOrderBuffetCustomerType) CalcSaleOrderBuffetCustomerType(setting SaleBillSetting) SaleOrderBuffetCustomerCalc {
	defer model.SetUpdate() // 标记该记录要更新
	serviceFeeRate := setting.GetServiceFeeRate()
	taxFeeType := setting.GetTaxFeeType()
	serviceFeeType := setting.GetServiceFeeType()
	fmt.Println("q3333333")

	return model.calcSaleOrderBuffetCustomerType(serviceFeeRate, taxFeeType, serviceFeeType)
}

// 获取价格变动前的信息
func (model *SaleOrderBuffetCustomerType) BeforeCalc() SaleOrderBuffetCustomerCalc {
	calc := SaleOrderBuffetCustomerCalc{}
	calc.Price = model.Price
	calc.CustomDiscountFee = model.CustomDiscountFee
	calc.ServiceTaxFee = model.ServiceTaxFee
	calc.TaxFee = model.TaxFee
	calc.ServiceFee = model.ServiceFee
	calc.TotalPrice = model.TotalPrice
	return calc
}
