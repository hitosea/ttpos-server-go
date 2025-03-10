package model

import (
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
)

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

// 获取所有自助餐名称
func (model *SaleOrder) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, buffet := range model.SaleOrderBuffetCustomerTypes {
		buffets = append(buffets, buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language))
	}
	return strings.Join(buffets, "+")
}

// 获取总的退款金额
func (model *SaleOrder) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, refundOrder := range model.ReturnOrders {
		refundAmount += refundOrder.RefundAmount
	}
	return refundAmount
}

// 返回销售订单商品
func (model *SaleOrder) GetSaleOrderProduct(saleOrderProductUuid uint64) (*SaleOrderProduct, int, error) {
	for i, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProductUuid == saleOrderProduct.Uuid {
			return saleOrderProduct, i, nil
		}
	}
	return nil, 0, errors.New("销售订单商品不存在")
}

// 获取销售订单应收金额
func (model *SaleOrder) GetAmount() float64 {
	// 整单改价金额大于等于0时，返回整单改价金额
	if model.CustomAmount >= 0 {
		return model.CustomAmount
	}
	// 默认返回订单总金额
	return model.Amount
}

// 根据sign获取销售订单商品
func (model *SaleOrder) GetSaleOrderProductBySign(sign string) *SaleOrderProduct {
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.Sign == sign {
			return saleOrderProduct
		}
	}
	return nil
}

// 设置为空。为了更新数据库数据时，不更新关联对象
func (model *SaleOrder) SetNil() {
	model.PaymentOrders = nil
	model.Member = nil
	model.SaleOrderProducts = nil
	model.ReturnOrders = nil
	model.SaleOrderBuffetCustomerTypes = nil
	model.SaleOrderBuffetDelayProducts = nil
}

// 设置会员折扣，并修改订单商品的折扣
func (model *SaleOrder) setMemberDiscount(memberUuid uint64, memberDiscount, cardDiscount float64) {
	// 修改订单的会员信息
	model.MemberDiscountRate = memberDiscount
	model.MemberCardDiscountRate = cardDiscount
	model.ConsumerUuid = memberUuid
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.SetMemberDiscountInfo(model.MemberDiscountRate, model.MemberCardDiscountRate)
		saleOrderProduct.SetUpdate()
	}
	// 对自助餐顾客进行打折. 顾客没有会员折扣
}

func (model *SaleOrder) SetMemberDiscount(member Member) {
	// 修改订单的会员信息
	model.setMemberDiscount(member.Uuid, member.GetMemberDiscountRate(), member.GetMemberCardDiscountRate())
}

// 设置取消会员折扣，并修改订单商品的折扣
func (model *SaleOrder) SetMemberDiscountCancel() {
	// 修改订单的会员信息
	discountRate := float64(1)                  // 无折扣，1乘任何价格都等于原价
	model.MemberDiscountRate = discountRate     // 会员折扣，无折扣
	model.MemberCardDiscountRate = discountRate // 会员卡折扣，无折扣
	model.ConsumerUuid = 0                      // 会员ID置空
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.SetMemberDiscountInfo(discountRate, discountRate)
		saleOrderProduct.SetUpdate()
	}
}

// 设置整单折扣，并修改订单商品的折扣
// 参数discount，表示给订单设置的打折率，统一使用百分比打折。比如八折，discount值为0.8；比如30% off，discount值为0.7。
// 注意：请在调用该方法时，就做好discount值的转化
func (model *SaleOrder) SetCustomDiscount(discount float64) {
	defer model.SetCustomAmountCancel() // 取消整单改价金额
	defer model.SetZeroRuleCancel()     // 取消订单抹零

	model.CustomDiscountRate = discount
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.CustomDiscountRate = discount
		saleOrderProduct.SetUpdate()
	}
	// 对自助餐顾客进行打折
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() {
			continue
		}
		buffetCustomer.CustomDiscountRate = discount
		buffetCustomer.SetUpdate()
	}
}

// 取消整单折扣
func (model *SaleOrder) SetCustomDiscountCancel() bool {
	isChange := false
	model.CustomDiscountRate = constant.NoDiscount
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		// 如果订单商品折扣不为100%，则修改折扣。确保如果原本就没有自定义折扣就不用更新数据库
		if saleOrderProduct.CustomDiscountRate != constant.NoDiscount {
			saleOrderProduct.CustomDiscountRate = constant.NoDiscount
			saleOrderProduct.SetUpdate()
			isChange = true
		}
	}
	// 取消自助餐顾客折扣
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() {
			continue
		}
		// 如果自助餐顾客折扣不为100%，则修改折扣。确保如果原本就没有自定义折扣就不用更新数据库
		if buffetCustomer.CustomDiscountRate != constant.NoDiscount {
			buffetCustomer.CustomDiscountRate = constant.NoDiscount
			buffetCustomer.SetUpdate()
			isChange = true
		}
	}
	return isChange
}

// 设置整单改价金额
func (model *SaleOrder) SetCustomAmount(amount float64) {
	defer model.SetZeroRuleCancel()       // 取消订单抹零
	defer model.SetCustomDiscountCancel() // 取消整单折扣
	model.CustomAmount = amount
}

// 取消整单改价金额
func (model *SaleOrder) SetCustomAmountCancel() bool {
	isChange := false
	model.CustomAmount = constant.SaleOrderCustomAmountCancel
	return isChange
}

// 设置订单抹零规则
func (model *SaleOrder) SetZeroRule(zeroRule int) {
	model.ZeroRule = uint8(zeroRule)
}

// 取消订单抹零
func (model *SaleOrder) SetZeroRuleCancel() bool {
	isChange := false
	// 将订单的抹零规则设置为实款实收
	if model.ZeroRule != constant.DiscountZeroRuleNone {
		model.ZeroRule = constant.DiscountZeroRuleNone
		isChange = true
	}
	return isChange
}

// 取消整单折扣
func (model *SaleOrder) SetAllDiscountCancel() bool {
	isChange := false
	isChange = model.SetZeroRuleCancel() || isChange
	isChange = model.SetCustomDiscountCancel() || isChange
	isChange = model.SetCustomAmountCancel() || isChange
	return isChange
}

// 是否存在折扣
func (model *SaleOrder) IsDiscount() bool {
	// custom_amount == -1 是没有进行订单改价
	// custom_discount_rate = 1 是没有折扣
	// zero_rule = 0 是没有去零
	return model.CustomAmount != -1 || model.CustomDiscountRate != 1 || model.ZeroRule != 0
}
