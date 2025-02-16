package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"github.com/shopspring/decimal"
)

// 订单计算服务
type IOrderCalcSrv interface {
	CalcOrderProductAmount(saleBill model.SaleBill, saleOrder model.SaleOrder, orderProduct model.SaleOrderProduct) CalcOrderProductAmountResp // 计算订单商品金额
	CalcOrderProductDiscountAmount(req CalcOrderProductDiscountAmountReq) CalcOrderProductDiscountAmountResp                                   // 计算订单商品折扣金额
	CalcOrderProductTaxAmount(req CalcOrderProductTaxAmountReq) CalcOrderProductTaxAmountResp                                                  // 计算订单商品税费
	CalcOrderProductServiceAmount(req CalcOrderProductServiceAmountReq) CalcOrderProductServiceAmountResp                                      // 计算订单商品服务费
}

// 订单计算服务实现
type orderCalcSrv struct {
	dbm *database.DBManager
}

// NewOrderCalcSrv 创建订单计算服务
func NewOrderCalcSrv(dbm *database.DBManager) IOrderCalcSrv {
	return NewOrderCalcSrvImpl(dbm)
}

// NewOrderCalcSrvImpl 创建订单计算服务实现
func NewOrderCalcSrvImpl(dbm *database.DBManager) IOrderCalcSrv {
	return &orderCalcSrv{dbm: dbm}
}

// CalcOrderProductAmountResp 计算订单商品金额响应
type CalcOrderProductAmountResp struct {
	DiscountAmount CalcOrderProductDiscountAmountResp // 订单商品折扣金额
	TaxAmount      CalcOrderProductTaxAmountResp      // 订单商品税费
	ServiceAmount  CalcOrderProductServiceAmountResp  // 订单商品服务费
	TotalPrice     float64                            // 订单商品应收金额
}

// CalcOrderProductAmount 计算订单商品金额
func (o *orderCalcSrv) CalcOrderProductAmount(saleBill model.SaleBill, saleOrder model.SaleOrder, orderProduct model.SaleOrderProduct) CalcOrderProductAmountResp {
	// 计算订单商品折扣金额
	discountAmount := o.CalcOrderProductDiscountAmount(CalcOrderProductDiscountAmountReq{
		SalePrice:              orderProduct.SalePrice,
		IsOpenMemberDiscount:   orderProduct.IsOpenMemberDiscount,
		MemberDiscountRate:     saleOrder.MemberDiscountRate,
		MemberCardDiscountRate: saleOrder.MemberCardDiscountRate,
		CustomDiscountRate:     saleOrder.CustomDiscountRate,
	})

	// 计算订单商品税费
	taxAmount := o.CalcOrderProductTaxAmount(CalcOrderProductTaxAmountReq{
		FlavorPrice:  orderProduct.FlavorPrice,
		ProductPrice: orderProduct.SalePrice,
		TaxFeeType:   saleBill.SaleBillSetting.TaxFeeType,
		TaxRate:      orderProduct.TaxRate,
	})

	// 计算订单商品服务费
	serviceAmount := o.CalcOrderProductServiceAmount(CalcOrderProductServiceAmountReq{
		Price:           discountAmount.Price,
		TaxFeeType:      saleBill.SaleBillSetting.TaxFeeType,
		TaxFee:          taxAmount.TaxFee,
		ServiceFeeType:  saleBill.SaleBillSetting.ServiceFeeType,
		ServiceFeeValue: saleBill.SaleBillSetting.ServiceFeeValue,
	})

	// 计算订单商品应收金额, 应收金额(单商品)=最终单价+服务费+总税费
	totalPrice := decimal.NewFromFloat(discountAmount.Price).
		Add(decimal.NewFromFloat(serviceAmount.ServiceFee)).
		Add(decimal.NewFromFloat(taxAmount.TaxFee)).
		Add(decimal.NewFromFloat(serviceAmount.ServiceTaxFee)).
		InexactFloat64()

	return CalcOrderProductAmountResp{
		DiscountAmount: discountAmount,
		TaxAmount:      taxAmount,
		ServiceAmount:  serviceAmount,
		TotalPrice:     totalPrice,
	}
}

// CalcOrderProductDiscountAmountReq 计算订单商品折扣金额请求
type CalcOrderProductDiscountAmountReq struct {
	SalePrice              float64 // 销售价
	IsOpenMemberDiscount   uint    // 是否开启会员折扣
	MemberDiscountRate     float64 // 会员折扣率
	MemberCardDiscountRate float64 // 会员卡折扣率
	CustomDiscountRate     float64 // 自定义折扣率
}

// CalcOrderProductDiscountAmountResp 计算订单商品折扣金额响应
type CalcOrderProductDiscountAmountResp struct {
	Price                  float64 // 订单商品最终单价
	MemberDiscountFee      float64 // 会员折扣金额
	CustomDiscountFee      float64 // 自定义折扣金额
	DiscountFee            float64 // 订单商品打折金额
	MemberDiscountRate     float64 // 会员折扣率
	MemberCardDiscountRate float64 // 会员卡折扣率
	CustomDiscountRate     float64 // 自定义折扣率
}

// CalcOrderProductDiscountAmount 计算订单商品折扣金额
func (o *orderCalcSrv) CalcOrderProductDiscountAmount(req CalcOrderProductDiscountAmountReq) CalcOrderProductDiscountAmountResp {
	var (
		price             decimal.Decimal // 订单商品最终单价
		memberDiscountFee decimal.Decimal // 会员折扣金额
		customDiscountFee decimal.Decimal // 自定义折扣金额
		discountFee       decimal.Decimal // 订单商品打折金额
	)

	// 会员折扣率, 会员折扣率*会员卡折扣率, 保留3位小数
	var memberDiscountRate decimal.Decimal = decimal.NewFromUint64(1)
	if req.IsOpenMemberDiscount == 1 {
		memberDiscountRate = decimal.NewFromFloat(req.MemberDiscountRate).Div(decimal.NewFromUint64(100))
	}
	memberCardDiscountRate := decimal.NewFromFloat(req.MemberCardDiscountRate).Div(decimal.NewFromUint64(100))
	memberRate := memberDiscountRate.Mul(memberCardDiscountRate).Round(3)
	memberDiscountFee = decimal.NewFromFloat(req.SalePrice).Mul(
		decimal.NewFromUint64(1).Sub(memberRate),
	).Round(2)

	// 自定义折扣率, 会员折扣率*自定义折扣率, 保留2位小数
	customDiscountRate := decimal.NewFromFloat(req.CustomDiscountRate).Div(decimal.NewFromUint64(100))
	discountRate := memberRate.Mul(customDiscountRate).Round(2)

	// 计算订单商品最终单价, 销售价*折扣率
	price = decimal.NewFromFloat(req.SalePrice).Mul(discountRate).Round(2)

	// 计算自定义折扣金额, 销售价-最终单价（单商品）-会员折扣金额
	customDiscountFee = decimal.NewFromFloat(req.SalePrice).Sub(price).Sub(memberDiscountFee)

	// 计算订单商品打折金额, 销售价-最终单价
	discountFee = decimal.NewFromFloat(req.SalePrice).Sub(price)

	// 计算折扣率, 转换为百分比, 保存到销售订单商品表中
	memberDiscountRate = memberDiscountRate.Mul(decimal.NewFromUint64(100))
	memberCardDiscountRate = memberCardDiscountRate.Mul(decimal.NewFromUint64(100))
	customDiscountRate = customDiscountRate.Mul(decimal.NewFromUint64(100))

	return CalcOrderProductDiscountAmountResp{
		Price:                  price.InexactFloat64(),
		MemberDiscountFee:      memberDiscountFee.InexactFloat64(),
		CustomDiscountFee:      customDiscountFee.InexactFloat64(),
		DiscountFee:            discountFee.InexactFloat64(),
		MemberDiscountRate:     memberDiscountRate.InexactFloat64(),
		MemberCardDiscountRate: memberCardDiscountRate.InexactFloat64(),
		CustomDiscountRate:     customDiscountRate.InexactFloat64(),
	}
}

// CalcOrderProductTaxAmountReq 计算订单商品税费请求
type CalcOrderProductTaxAmountReq struct {
	FlavorPrice  float64 // 规格原价
	ProductPrice float64 // 原始单价
	TaxFeeType   uint    // 税费类型: 0-关闭消费税 1-商品未含税 2-商品已含税
	TaxRate      float64 // 税率
}

// CalcOrderProductTaxAmountResp 计算订单商品税费响应
type CalcOrderProductTaxAmountResp struct {
	TaxFee float64 // 订单商品税费
}

// CalcOrderProductTaxAmount 计算订单商品税费
func (o *orderCalcSrv) CalcOrderProductTaxAmount(req CalcOrderProductTaxAmountReq) CalcOrderProductTaxAmountResp {
	var taxFee decimal.Decimal

	// 商品未含税时,税费=原始单价*税率
	if req.TaxFeeType == constant.SaleBillSettingTaxFeeTypePercent {
		taxFee = decimal.NewFromFloat(req.ProductPrice).Mul(
			decimal.NewFromFloat(req.TaxRate).Div(decimal.NewFromUint64(100)),
		)
	}

	// 商品已含税时,税费=规格原价*(1-1/(1+税率/100))
	if req.TaxFeeType == constant.SaleBillSettingTaxFeeTypePercentTax {
		taxFee = decimal.NewFromFloat(req.FlavorPrice).Mul(
			decimal.NewFromFloat(1).Sub(
				decimal.NewFromFloat(1).Div(
					decimal.NewFromFloat(1).Add(decimal.NewFromFloat(req.TaxRate).Div(decimal.NewFromUint64(100))),
				),
			),
		)
	}

	// 税费保留2位小数
	taxFee = taxFee.Round(2)

	return CalcOrderProductTaxAmountResp{
		TaxFee: taxFee.InexactFloat64(),
	}
}

// CalcOrderProductServiceAmountReq 计算订单商品服务费请求
type CalcOrderProductServiceAmountReq struct {
	Price           float64 // 订单商品最终单价
	TaxFeeType      uint    // 税费类型: 0-关闭消费税 1-商品未含税 2-商品已含税
	TaxFee          float64 // 商品税费
	TaxRate         float64 // 税率
	ServiceFeeType  uint    // 服务费类型: 服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费
	ServiceFeeValue float64 // 服务费比例: 服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例
}

// CalcOrderProductServiceAmountResp 计算订单商品服务费响应
type CalcOrderProductServiceAmountResp struct {
	ServiceFee    float64 // 订单商品服务费
	ServiceTaxFee float64 // 订单商品服务费税费
}

// CalcOrderProductServiceAmount 计算订单商品服务费
func (o *orderCalcSrv) CalcOrderProductServiceAmount(req CalcOrderProductServiceAmountReq) CalcOrderProductServiceAmountResp {
	var (
		serviceFee    decimal.Decimal
		serviceTaxFee decimal.Decimal
	)

	// 服务费类型为0或1时，服务费=0
	if req.ServiceFeeType > constant.SaleBillSettingServiceFeeTypeFixed {
		// 如果商品未含税
		// 服务费=最终单价*服务费比例
		if req.TaxFeeType == constant.SaleBillSettingTaxFeeTypeNone || req.TaxFeeType == constant.SaleBillSettingTaxFeeTypePercent {
			serviceFee = decimal.NewFromFloat(req.Price).Mul(decimal.NewFromFloat(req.ServiceFeeValue).Div(decimal.NewFromUint64(100)))
		}

		// 如果商品已含税
		// 服务费=(最终单价-商品税费)*服务费比例
		if req.TaxFeeType == constant.SaleBillSettingTaxFeeTypePercentTax {
			serviceFee = decimal.NewFromFloat(req.Price).Sub(decimal.NewFromFloat(req.TaxFee)).
				Mul(
					decimal.NewFromFloat(req.ServiceFeeValue).Div(decimal.NewFromUint64(100)),
				)
		}
	}

	// 服务费保留2位小数
	serviceFee = serviceFee.Round(2)

	// 按比例-收取税费时, 服务费税费=服务费*税率
	if req.ServiceFeeType == constant.SaleBillSettingServiceFeeTypePercentTax {
		serviceTaxFee = serviceFee.Mul(decimal.NewFromFloat(req.TaxRate).Div(decimal.NewFromUint64(100))).Round(2)
	}

	return CalcOrderProductServiceAmountResp{
		ServiceFee:    serviceFee.InexactFloat64(),
		ServiceTaxFee: serviceTaxFee.InexactFloat64(),
	}
}
