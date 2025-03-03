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
		buffets = append(buffets, buffet.BuffetPackageMultiLanguageName.GetNameByLang(language))
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
	model.Member = Member{}
	model.SaleOrderProducts = nil
	model.ReturnOrders = nil
	model.SaleOrderBuffetCustomerTypes = nil
	model.SaleOrderBuffetDelayProducts = nil
}

// 设置整单折扣，并修改订单商品的折扣
func (model *SaleOrder) SetCustomDiscount(discount float64) {
	defer model.SetCustomAmountCancel() // 取消整单改价金额
	defer model.SetZeroRuleCancel()     // 取消订单抹零

	model.CustomDiscountRate = discount
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.CustomDiscountRate = discount
		saleOrderProduct.SetUpdate()
	}
}

// 取消整单折扣
func (model *SaleOrder) SetCustomDiscountCancel() {
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
		}
	}
}

// 设置整单改价金额
func (model *SaleOrder) SetCustomAmount(amount float64) {
	defer model.SetZeroRuleCancel()       // 取消订单抹零
	defer model.SetCustomDiscountCancel() // 取消整单折扣
	model.CustomAmount = amount
}

// 取消整单改价金额
func (model *SaleOrder) SetCustomAmountCancel() {
	model.CustomAmount = constant.SaleOrderCustomAmountCancel
}

// 取消订单抹零
func (model *SaleOrder) SetZeroRuleCancel() {
	// 将订单的抹零规则设置为实款实收
	model.ZeroRule = constant.DiscountZeroRuleNone
}
