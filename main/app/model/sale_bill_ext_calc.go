package model

import "github.com/shopspring/decimal"

func (model *SaleBill) calcPaymentCommissionFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.PaymentCommissionFee))
	}
	return amount.InexactFloat64()
}

func (model *SaleBill) calcPaymentAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.PaymentAmount))
	}
	return amount.InexactFloat64()
}

func (model *SaleBill) calcProductOriginalAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.ProductOriginalAmount))
	}
	return amount.InexactFloat64()
}

// 计算销售账单的总金额。总金额=销售订单的应收金额之和
func (model *SaleBill) calcAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsDelete() {
			continue
		}
		amount = amount.Add(decimal.NewFromFloat(saleOrder.GetAmount()))
	}
	return amount.Truncate(3).Round(2).InexactFloat64()
}

// 计算销售账单的原始金额。原始金额=销售订单的原始应收金额之和
func (model *SaleBill) calcOriginAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsDelete() {
			continue
		}
		amount = amount.Add(decimal.NewFromFloat(saleOrder.OriginAmount))
	}
	return amount.Truncate(3).Round(2).InexactFloat64()
}

// 计算销售订单的商品金额。商品金额=销售订单的商品金额之和
func (model *SaleBill) calcProductAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.ProductAmount))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的服务费。服务费=销售订单的服务费之和
func (model *SaleBill) calcServiceFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.ServiceFee))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的税费。税费=销售订单的税费之和
func (model *SaleBill) calcTaxFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.TaxFee))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的折扣费用。折扣费用=销售订单的折扣费用之和
func (model *SaleBill) calcDiscountFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.CustomDiscountFee))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的会员折扣费用。会员折扣费用=销售订单的会员折扣费用之和
func (model *SaleBill) calcMemberDiscountFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.MemberDiscountFee))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的赠菜金额。赠菜金额=销售订单的赠菜金额之和
func (model *SaleBill) calcGiftAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		// 赠菜金额=销售订单的赠菜金额之和 累加
		amount = amount.Add(decimal.NewFromFloat(saleOrder.GiftAmount))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的免单金额。免单金额=销售订单的免单金额之和
func (model *SaleBill) calcFreeAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() && saleOrder.IsFreeSaleOrder() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.GetAmount()))
		}
	}
	return amount.InexactFloat64()
}

// 计算销售订单的满减活动抵扣金额。活动抵扣金额=销售订单的活动抵扣金额之和
func (model *SaleBill) calcActivityAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if !saleOrder.IsDelete() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.ActivityAmount))
		}
	}
	return amount.InexactFloat64()
}

// 重新计算销售账单的金额
func (model *SaleBill) CalcAll(options ...func(option *CalcOption)) {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	setting := model.SaleBillSetting
	if option.SaleBillSetting != nil {
		setting = option.SaleBillSetting
	}
	for i := range model.SaleOrders {
		saleOrder := model.SaleOrders[i]
		for j := range saleOrder.SaleOrderProducts {
			saleOrderProduct := saleOrder.SaleOrderProducts[j]
			if saleOrderProduct == nil {
				continue
			}
			// 如果订单商品已删除或已取消，则不计算
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() {
				continue
			}
			_ = saleOrderProduct.BeforeCalc()
			_ = saleOrderProduct.CalcSaleOrderProduct(*setting, options...)
		}
		// 计算自助餐顾客价格之和
		for _, buffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
			if buffetCustomer.IsDelete() {
				continue
			}
			_ = buffetCustomer.CalcSaleOrderBuffetCustomerType(*setting, options...)
		}
		if option.H5OrderUuid == 0 {
			if option.IsCanceled {
				saleOrder.CalcSaleOrder(*setting, WithCanceled())
			} else {
				saleOrder.CalcSaleOrder(*setting)
			}
		} else {
			saleOrder.CalcSaleOrder(*setting, options...)
		}
	}
	model.CalcSaleBill()
}

// 重新计算销售订单的金额
func (model *SaleBill) CalcSaleBill() *SaleBillCalc {
	calc := SaleBillCalc{}
	calc.Amount = model.calcAmount()
	model.Amount = calc.Amount
	calc.OriginAmount = model.calcOriginAmount()
	model.OriginAmount = calc.OriginAmount
	calc.ProductAmount = model.calcProductAmount()
	model.ProductAmount = calc.ProductAmount
	calc.ServiceFee = model.calcServiceFee()
	model.ServiceFee = calc.ServiceFee
	calc.TaxFee = model.calcTaxFee()
	model.TaxFee = calc.TaxFee
	calc.DiscountFee = model.calcDiscountFee()
	model.CustomDiscountFee = calc.DiscountFee
	calc.MemberDiscountFee = model.calcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.GiftAmount = model.calcGiftAmount()
	model.GiftAmount = calc.GiftAmount
	calc.ActivityAmount = model.calcActivityAmount()
	model.ActivityAmount = calc.ActivityAmount
	calc.FreeAmount = model.calcFreeAmount()
	model.FreeAmount = calc.FreeAmount
	calc.ProductOriginalAmount = model.calcProductOriginalAmount()
	model.ProductOriginalAmount = calc.ProductOriginalAmount
	calc.PaymentAmount = model.calcPaymentAmount()
	model.PaymentAmount = calc.PaymentAmount
	calc.PaymentCommissionFee = model.calcPaymentCommissionFee()
	model.PaymentCommissionFee = calc.PaymentCommissionFee
	return &calc
}
